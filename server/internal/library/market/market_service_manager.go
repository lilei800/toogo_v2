// Package market 全局行情服务管理器
// 每个交易所一个行情服务实例，统一管理所有交易所的实时行情数据
package market

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"hotgo/internal/library/exchange"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func normalizePlatform(platform string) string {
	return NormalizePlatform(platform)
}

func normalizeSymbol(symbol string) string {
	// 仅做轻量规范化：去空格 + 大写。避免破坏诸如 OKX 的 instId 格式（若业务层直接传 instId）。
	return NormalizeSymbol(symbol)
}

// MarketServiceManager 全局行情服务管理器（单例）
// 管理每个交易所的独立行情服务
type MarketServiceManager struct {
	mu sync.RWMutex

	// 每个交易所一个行情服务 key: platform (binance/okx/gate)
	services map[string]*ExchangeMarketService

	// WebSocket服务（优先使用）
	wsEnabled bool
	// wsOnly: 强制“只使用WebSocket数据源”（ticker/klines均不做REST兜底/轮询）。
	// 适用场景：需要彻底隔离行情链路与HTTP/交易所REST限流，并允许“未就绪则为空”的严格语义。
	wsOnly    bool
	binanceWS *exchange.BinanceWebSocket
	okxWS     *exchange.OKXWebSocket
	gateWS    *exchange.GateWebSocket

	// 代理配置
	proxyDialer func(network, addr string) (net.Conn, error)

	// 【新增】价格更新回调（用于实时触发引擎检查）
	// key: platform:symbol, value: 回调函数列表
	priceCallbacks map[string][]func(*exchange.Ticker)

	// 【新增】回调队列：将“实时报价（WS -> 缓存写入）”与“策略/订单/风控回调”彻底隔离
	// - WS 线程只做“写缓存 + 非阻塞入队”，绝不执行慢逻辑
	// - 回调侧用容量=1的队列做 coalesce，只保留最新 tick，避免 goroutine 风暴/CPU 抢占
	// key: platform:symbol
	callbackQueues map[string]chan *exchange.Ticker

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// ExchangeMarketService 单个交易所的行情服务
type ExchangeMarketService struct {
	mu sync.RWMutex

	Platform string            // 交易所名称
	Exchange exchange.Exchange // 交易所API实例

	// WSOnly: 强制只使用WS数据源（禁用REST初始拉取/轮询兜底）
	WSOnly bool

	// 行情数据缓存 key: symbol
	Tickers    map[string]*TickerCache
	Klines     map[string]*KlineCache
	OrderBooks map[string]*OrderBookCache

	// 订阅的交易对 key: symbol, value: 引用计数
	Subscriptions map[string]int

	// markPriceFallbackAt: MarkPrice 的低频REST兜底限流（当WS ticker 没有 markPrice 时使用）
	// key: symbol
	markPriceFallbackAt map[string]time.Time

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// TickerCache Ticker缓存（带时间戳）
type TickerCache struct {
	Data      *exchange.Ticker
	UpdatedAt time.Time
}

// OrderBookCache 订单簿缓存
type OrderBookCache struct {
	Bids      [][2]float64
	Asks      [][2]float64
	UpdatedAt time.Time
}

var (
	marketServiceManager     *MarketServiceManager
	marketServiceManagerOnce sync.Once
)

// GetMarketServiceManager 获取全局行情服务管理器单例
func GetMarketServiceManager() *MarketServiceManager {
	marketServiceManagerOnce.Do(func() {
		marketServiceManager = &MarketServiceManager{
			services:       make(map[string]*ExchangeMarketService),
			priceCallbacks: make(map[string][]func(*exchange.Ticker)),
			callbackQueues: make(map[string]chan *exchange.Ticker),
			stopCh:         make(chan struct{}),
		}
	})
	return marketServiceManager
}

// Start 启动行情服务管理器
func (m *MarketServiceManager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		g.Log().Warning(ctx, "[MarketServiceManager] 已经在运行中，跳过重复启动")
		return nil
	}
	m.running = true
	m.mu.Unlock()

	g.Log().Warning(ctx, "[MarketServiceManager] 🚀 开始启动全局行情服务管理器...")

	// 尝试启动WebSocket服务（非阻塞，失败不影响主流程）
	m.startWebSocketServices(ctx)

	g.Log().Warning(ctx, "[MarketServiceManager] ✅ 全局行情服务管理器启动完成")
	return nil
}

// SetProxyDialer 设置代理拨号器（应在 Start 之前调用）
func (m *MarketServiceManager) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.proxyDialer = dialer
	if dialer != nil {
		g.Log().Info(context.Background(), "[MarketServiceManager] 已设置WebSocket代理")
	}
}

// startWebSocketServices 启动WebSocket服务
func (m *MarketServiceManager) startWebSocketServices(ctx context.Context) {
	// 读取配置决定是否启用WebSocket
	wsEnabled, _ := g.Cfg().Get(ctx, "toogo.websocketEnabled")
	g.Log().Warningf(ctx, "[MarketServiceManager] 检查WebSocket配置: wsEnabled=%v, IsEmpty=%v, Bool=%v",
		wsEnabled.Val(), wsEnabled.IsEmpty(), wsEnabled.Bool())

	if wsEnabled.IsEmpty() || !wsEnabled.Bool() {
		g.Log().Warning(ctx, "[MarketServiceManager] WebSocket未启用，使用HTTP轮询模式")
		return
	}

	g.Log().Warning(ctx, "[MarketServiceManager] 准备启动WebSocket服务...")

	// WebSocket-only 模式：强制行情/多周期K线只走WS（不做REST兜底/轮询）
	wsOnlyVal, _ := g.Cfg().Get(ctx, "toogo.websocketOnly")

	m.mu.Lock()
	m.wsEnabled = true
	m.wsOnly = (!wsOnlyVal.IsEmpty() && wsOnlyVal.Bool())
	proxyDialer := m.proxyDialer // 复制代理配置
	m.mu.Unlock()

	// 统一启动流程：减少重复代码
	successCount := 0
	totalCount := 3

	// 启动各个交易所WebSocket
	startWS := func(name string, getter func() interface{}, setter func(interface{})) {
		ws := getter()

		// 设置代理
		type proxySettable interface {
			SetProxyDialer(func(string, string) (net.Conn, error))
		}
		if proxyDialer != nil {
			if p, ok := ws.(proxySettable); ok {
				p.SetProxyDialer(proxyDialer)
			}
		}

		// 启动WebSocket
		type startable interface {
			Start(context.Context) error
		}
		if s, ok := ws.(startable); ok {
			if err := s.Start(ctx); err != nil {
				g.Log().Warningf(ctx, "[MarketServiceManager] %s WebSocket启动失败: %v", name, err)
			} else {
				g.Log().Warningf(ctx, "[MarketServiceManager] ✅ %s WebSocket已启动", name)
				setter(ws)
				successCount++
			}
		}
	}

	// 启动顺序：Gate -> OKX -> Binance
	startWS("Gate", func() interface{} { return exchange.GetGateWebSocket() }, func(ws interface{}) { m.gateWS = ws.(*exchange.GateWebSocket) })
	startWS("OKX", func() interface{} { return exchange.GetOKXWebSocket() }, func(ws interface{}) { m.okxWS = ws.(*exchange.OKXWebSocket) })
	startWS("Binance", func() interface{} { return exchange.GetBinanceWebSocket() }, func(ws interface{}) { m.binanceWS = ws.(*exchange.BinanceWebSocket) })

	g.Log().Warningf(ctx, "[MarketServiceManager] WebSocket服务启动完成: 成功=%d/%d", successCount, totalCount)
}

// Stop 停止行情服务管理器
func (m *MarketServiceManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopCh)

	// 停止所有WebSocket服务（统一循环处理）
	type stoppable interface {
		Stop()
	}

	wsClients := []struct {
		name   string
		client stoppable
	}{
		{"Binance", m.binanceWS},
		{"OKX", m.okxWS},
		{"Gate", m.gateWS},
	}

	for _, ws := range wsClients {
		if ws.client != nil {
			ws.client.Stop()
			g.Log().Infof(context.Background(), "[MarketServiceManager] %s WebSocket已停止", ws.name)
		}
	}

	// 停止所有HTTP轮询服务
	for _, svc := range m.services {
		svc.Stop()
	}

	g.Log().Info(context.Background(), "[MarketServiceManager] 行情服务管理器已停止")
}

// IsRunning 检查是否运行中
func (m *MarketServiceManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetOrCreateService 获取或创建交易所行情服务
func (m *MarketServiceManager) GetOrCreateService(ctx context.Context, platform string, ex exchange.Exchange) *ExchangeMarketService {
	platform = normalizePlatform(platform)
	m.mu.Lock()
	defer m.mu.Unlock()

	if svc, ok := m.services[platform]; ok {
		return svc
	}

	// 创建新的交易所行情服务
	svc := &ExchangeMarketService{
		Platform:            platform,
		Exchange:            ex,
		WSOnly:              m.wsOnly,
		Tickers:             make(map[string]*TickerCache),
		Klines:              make(map[string]*KlineCache),
		OrderBooks:          make(map[string]*OrderBookCache),
		Subscriptions:       make(map[string]int),
		markPriceFallbackAt: make(map[string]time.Time),
		stopCh:              make(chan struct{}),
	}

	// 启动服务
	svc.Start(ctx)
	m.services[platform] = svc

	g.Log().Infof(ctx, "[MarketServiceManager] 创建交易所行情服务: %s", platform)
	return svc
}

// GetService 获取交易所行情服务
func (m *MarketServiceManager) GetService(platform string) *ExchangeMarketService {
	platform = normalizePlatform(platform)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.services[platform]
}

// Subscribe 订阅交易对行情
func (m *MarketServiceManager) Subscribe(ctx context.Context, platform, symbol string, ex exchange.Exchange) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	// 先确保 HTTP/缓存服务存在（并触发首轮 fetchInitialData），避免 WS 回调早到导致写缓存时 svc 为空
	svc := m.GetOrCreateService(ctx, platform, ex)
	svc.Subscribe(ctx, symbol)

	// 再订阅WebSocket（如果启用）
	m.subscribeWebSocket(ctx, platform, symbol)
}

// SubscribeQuoteOnly 仅订阅报价（ticker/mark price），不订阅K线。
// 适用场景：执行平台只需要报价/风控口径，K线/市场状态来自其它分析平台。
func (m *MarketServiceManager) SubscribeQuoteOnly(ctx context.Context, platform, symbol string, ex exchange.Exchange) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	svc := m.GetOrCreateService(ctx, platform, ex)
	svc.Subscribe(ctx, symbol)
	m.subscribeWebSocketQuoteOnly(ctx, platform, symbol)
}

// SubscribeWithCallback 订阅交易对行情并注册价格更新回调
// 【新增】用于实时触发引擎的止损止盈检查
func (m *MarketServiceManager) SubscribeWithCallback(ctx context.Context, platform, symbol string, ex exchange.Exchange, callback func(*exchange.Ticker)) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	// 先执行标准订阅
	m.Subscribe(ctx, platform, symbol, ex)

	// 注册回调
	if callback != nil {
		key := platform + ":" + symbol
		m.mu.Lock()
		m.priceCallbacks[key] = append(m.priceCallbacks[key], callback)
		// 确保回调队列与 worker 存在（用于隔离“报价更新”与“回调慢逻辑”）
		if _, ok := m.callbackQueues[key]; !ok {
			m.callbackQueues[key] = make(chan *exchange.Ticker, 1) // coalesce：只保留最新
			go m.runPriceCallbackWorker(key)
		}
		m.mu.Unlock()

		g.Log().Debugf(ctx, "[MarketServiceManager] 注册价格更新回调: %s", key)

		// 【启动期优化】订阅后立刻尝试用“已有缓存ticker(WS/REST)”补一次回调，避免引擎/下单阶段 LastTicker 为空。
		// 注意：fetchInitialData 是异步的，所以这里做短暂重试（最多2秒）。
		go func() {
			defer func() { recover() }()
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				tk := m.GetTicker(platform, symbol)
				if tk != nil && tk.LastPrice > 0 {
					callback(tk)
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}
}

// SubscribeQuoteOnlyWithCallback 仅订阅报价并注册价格更新回调（不订阅K线）。
func (m *MarketServiceManager) SubscribeQuoteOnlyWithCallback(ctx context.Context, platform, symbol string, ex exchange.Exchange, callback func(*exchange.Ticker)) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	m.SubscribeQuoteOnly(ctx, platform, symbol, ex)

	// 注册回调（逻辑同 SubscribeWithCallback）
	if callback != nil {
		key := platform + ":" + symbol
		m.mu.Lock()
		m.priceCallbacks[key] = append(m.priceCallbacks[key], callback)
		if _, ok := m.callbackQueues[key]; !ok {
			m.callbackQueues[key] = make(chan *exchange.Ticker, 1)
			go m.runPriceCallbackWorker(key)
		}
		m.mu.Unlock()

		g.Log().Debugf(ctx, "[MarketServiceManager] 注册价格更新回调(QuoteOnly): %s", key)

		go func() {
			defer func() { recover() }()
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				tk := m.GetTicker(platform, symbol)
				if tk != nil && tk.LastPrice > 0 {
					callback(tk)
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}
}

// runPriceCallbackWorker 串行处理某个 (platform:symbol) 的价格回调。
// 设计目标：
// - WS ticker 更新只负责“写缓存 + 非阻塞入队”，不被任何慢逻辑拖慢
// - 回调天然可能慢（止盈止损/DB/策略计算/日志），用 coalesce 合并高频 tick，避免 goroutine 风暴
func (m *MarketServiceManager) runPriceCallbackWorker(key string) {
	// 最小触发间隔：避免回调过密导致 CPU 被策略/订单逻辑抢占，从而间接影响“报价实时性”
	const minInterval = 100 * time.Millisecond
	var lastAt time.Time

	for {
		// 队列可能在运行时被创建/复用，这里每轮都读取一次，保证安全
		m.mu.RLock()
		ch := m.callbackQueues[key]
		m.mu.RUnlock()
		if ch == nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		tk, ok := <-ch
		if !ok {
			return
		}
		if tk == nil {
			continue
		}

		// coalesce：尽可能把积压的 tick 合并为最后一条（只保留最新）
		for {
			select {
			case tk2 := <-ch:
				if tk2 != nil {
					tk = tk2
				}
			default:
				goto PROCESS
			}
		}

	PROCESS:
		// 节流（不影响报价缓存写入，只影响回调执行频率）
		if !lastAt.IsZero() {
			if d := time.Since(lastAt); d < minInterval {
				time.Sleep(minInterval - d)
			}
		}
		lastAt = time.Now()

		// 获取回调快照（避免持锁执行回调）
		m.mu.RLock()
		callbacks := append([]func(*exchange.Ticker){}, m.priceCallbacks[key]...)
		m.mu.RUnlock()

		for _, cb := range callbacks {
			if cb == nil {
				continue
			}
			func() {
				defer func() { recover() }()
				cb(tk)
			}()
		}
	}
}

// UnsubscribeCallback 取消订阅回调
func (m *MarketServiceManager) UnsubscribeCallback(platform, symbol string, callback func(*exchange.Ticker)) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	key := platform + ":" + symbol
	m.mu.Lock()
	defer m.mu.Unlock()

	// 移除指定的回调（如果传入nil，则清空所有回调）
	if callback == nil {
		delete(m.priceCallbacks, key)
		// 不关闭 callbackQueues：避免并发发送导致 panic。
		// worker 将阻塞在队列读取上，不消耗 CPU；后续重新注册回调可复用该队列。
		return
	}

	// 注意：Go中函数无法直接比较，这里只能清空所有回调
	// 如果需要精确移除，需要改用ID机制
	delete(m.priceCallbacks, key)
}

// subscribeWebSocket 订阅WebSocket行情
func (m *MarketServiceManager) subscribeWebSocket(ctx context.Context, platform, symbol string) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	m.mu.RLock()
	wsEnabled := m.wsEnabled
	m.mu.RUnlock()

	if !wsEnabled {
		return
	}

	// 将 WS K线数据写回 ExchangeMarketService.Klines，供 MarketAnalyzer 直接读取
	updateSvcKlines := func(interval string, klines []*exchange.Kline) {
		if len(klines) == 0 {
			return
		}
		svc := m.GetService(platform)
		if svc == nil {
			return
		}
		svc.mu.Lock()
		cache := svc.Klines[symbol]
		if cache == nil {
			cache = &KlineCache{UpdatedAt: time.Now()}
			svc.Klines[symbol] = cache
		}
		cache.UpdatedAt = time.Now()
		switch strings.ToLower(interval) {
		case "1m":
			cache.Klines1m = klines
		case "5m":
			cache.Klines5m = klines
		case "15m":
			cache.Klines15m = klines
		case "30m":
			cache.Klines30m = klines
		case "1h":
			cache.Klines1h = klines
		}
		svc.mu.Unlock()
	}

	switch platform {
	case "binance":
		if m.binanceWS != nil && m.binanceWS.IsRunning() {
			m.binanceWS.SubscribeTicker(symbol, func(ticker *exchange.Ticker) {
				if svc := m.GetService(platform); svc != nil {
					svc.mu.Lock()
					svc.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
					svc.mu.Unlock()
				}
				// 【新增】触发注册的回调函数
				m.triggerPriceCallbacks(platform, symbol, ticker)
			})
			// 标记价格（风控/盈亏口径）
			_ = m.binanceWS.SubscribeMarkPrice(symbol)
			for _, interval := range []string{"1m", "5m", "15m", "30m", "1h"} {
				_ = m.binanceWS.SubscribeKline(symbol, interval, func(klines []*exchange.Kline) {
					updateSvcKlines(interval, klines)
				})
			}
		}
	case "okx":
		if m.okxWS != nil && m.okxWS.IsRunning() {
			m.okxWS.SubscribeTicker(symbol, func(ticker *exchange.Ticker) {
				if svc := m.GetService(platform); svc != nil {
					svc.mu.Lock()
					svc.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
					svc.mu.Unlock()
				}
				// 【新增】触发注册的回调函数
				m.triggerPriceCallbacks(platform, symbol, ticker)
			})
			_ = m.okxWS.SubscribeMarkPrice(symbol)
			for _, interval := range []string{"1m", "5m", "15m", "30m", "1h"} {
				_ = m.okxWS.SubscribeKline(symbol, interval, func(klines []*exchange.Kline) {
					updateSvcKlines(interval, klines)
				})
			}
		}
	case "gate":
		// Gate WS 连接可能比其它交易所更慢（或短暂断线重连）。
		// 这里不要用 IsRunning() 做硬门槛，否则“订阅请求发生在连接完成之前”会被跳过，导致永远没有K线数据。
		// SubscribeKline/SubscribeTicker 内部会保存 subscriptions，连接恢复后 onConnected 会自动重放。
		if m.gateWS != nil {
			m.gateWS.SubscribeTicker(symbol, func(ticker *exchange.Ticker) {
				if svc := m.GetService(platform); svc != nil {
					svc.mu.Lock()
					svc.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
					svc.mu.Unlock()
				}
				// 【新增】触发注册的回调函数
				m.triggerPriceCallbacks(platform, symbol, ticker)
			})
			for _, interval := range []string{"1m", "5m", "15m", "30m", "1h"} {
				_ = m.gateWS.SubscribeKline(symbol, interval, func(klines []*exchange.Kline) {
					updateSvcKlines(interval, klines)
				})
			}

			// 兜底：如果 WS-only 模式下 Gate 长时间收不到 candlesticks（或解析异常），会导致 MarketAnalyzer 永远没有数据。
			// 这里做一次延迟检查：若仍无任何K线，则触发一次 REST 拉取补齐（仅一次，避免刷接口）。
			// 目的：让“市场状态/多周期播报”至少可用，即使 Gate WS 的 K线频道不稳定。
			m.mu.RLock()
			wsOnly := m.wsOnly
			m.mu.RUnlock()
			if wsOnly {
				p := platform
				sym := symbol
				svcPlatform := platform // capture
				go func() {
					defer func() {
						if r := recover(); r != nil {
							g.Log().Warningf(context.Background(),
								"[MarketServiceManager] Gate REST K线兜底 goroutine panic: platform=%s, symbol=%s, err=%v", p, sym, r)
						}
					}()
					// Gate candlesticks 可能“慢热”：新订阅后 10~30 秒才可能推第一根K线。
					// 这里不要在 5~6 秒内就判定“未就绪”，否则会误触发兜底并产生噪声日志。
					time.Sleep(35 * time.Second)
					svc := m.GetService(svcPlatform)
					if svc == nil {
						return
					}
				kc := svc.GetMultiTimeframeKlines(sym)
				if !klineCacheHasAnyData(kc) {
					// 【降级为Debug】Gate机器人通常使用OKX的市场状态和K线数据（analysisPlatform=okx），
					// 所以Gate自己的K线兜底日志只用于调试，不应该刷屏。
					g.Log().Debugf(context.Background(),
						"[MarketServiceManager] Gate WS K线未就绪，触发一次REST兜底补齐: platform=%s, symbol=%s", p, sym)
						// 注意：WS-only 模式下 ExchangeMarketService.fetchAllKlines 会被 WSOnly 短路，无法真正走 REST。
						// 因此这里直接使用“公共行情服务”拉取 Gate 的 candlesticks（不依赖用户API），并写回 KlineCache，确保 MarketAnalyzer 可产出数据。
						pms := exchange.GetPublicMarketService()
						type it struct {
							interval string
							limit    int
						}
						items := []it{
							{"1m", 100},
							{"5m", 100},
							{"15m", 100},
							{"30m", 50},
							{"1h", 50},
						}
						now := time.Now()
						var n1, n5, n15, n30, n1h int
						for _, item := range items {
							kl, err := pms.GetKlines(context.Background(), exchange.PlatformGate, sym, item.interval, item.limit)
							if err != nil || len(kl) == 0 {
						if err != nil {
								g.Log().Debugf(context.Background(),
									"[MarketServiceManager] Gate REST K线兜底失败: platform=%s, symbol=%s, interval=%s, err=%v", p, sym, item.interval, err)
							}
								continue
							}
							svc.mu.Lock()
							cache := svc.Klines[sym]
							if cache == nil {
								cache = &KlineCache{}
								svc.Klines[sym] = cache
							}
							cache.UpdatedAt = now
							switch item.interval {
							case "1m":
								cache.Klines1m = kl
								n1 = len(kl)
							case "5m":
								cache.Klines5m = kl
								n5 = len(kl)
							case "15m":
								cache.Klines15m = kl
								n15 = len(kl)
							case "30m":
								cache.Klines30m = kl
								n30 = len(kl)
							case "1h":
								cache.Klines1h = kl
								n1h = len(kl)
							}
							svc.mu.Unlock()
						}
						g.Log().Debugf(context.Background(),
							"[MarketServiceManager] Gate REST K线兜底完成: platform=%s, symbol=%s, 1m=%d, 5m=%d, 15m=%d, 30m=%d, 1h=%d",
							p, sym, n1, n5, n15, n30, n1h)
					}
				}()
			}
		}
	}
}

// subscribeWebSocketQuoteOnly 仅订阅报价（ticker/mark price），不订阅K线/不触发K线兜底。
func (m *MarketServiceManager) subscribeWebSocketQuoteOnly(ctx context.Context, platform, symbol string) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	m.mu.RLock()
	wsEnabled := m.wsEnabled
	m.mu.RUnlock()
	if !wsEnabled {
		return
	}

	switch platform {
	case "binance":
		if m.binanceWS != nil && m.binanceWS.IsRunning() {
			m.binanceWS.SubscribeTicker(symbol, func(ticker *exchange.Ticker) {
				if svc := m.GetService(platform); svc != nil {
					svc.mu.Lock()
					svc.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
					svc.mu.Unlock()
				}
				m.triggerPriceCallbacks(platform, symbol, ticker)
			})
			_ = m.binanceWS.SubscribeMarkPrice(symbol)
		}
	case "okx":
		if m.okxWS != nil && m.okxWS.IsRunning() {
			m.okxWS.SubscribeTicker(symbol, func(ticker *exchange.Ticker) {
				if svc := m.GetService(platform); svc != nil {
					svc.mu.Lock()
					svc.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
					svc.mu.Unlock()
				}
				m.triggerPriceCallbacks(platform, symbol, ticker)
			})
			_ = m.okxWS.SubscribeMarkPrice(symbol)
		}
	case "gate":
		// Gate quote-only: 只订阅 ticker（不订阅 candlesticks，因此不会打印 Gate K线兜底/未就绪日志）
		if m.gateWS != nil {
			m.gateWS.SubscribeTicker(symbol, func(ticker *exchange.Ticker) {
				if svc := m.GetService(platform); svc != nil {
					svc.mu.Lock()
					svc.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
					svc.mu.Unlock()
				}
				m.triggerPriceCallbacks(platform, symbol, ticker)
			})
		}
	}
}

// Unsubscribe 取消订阅
func (m *MarketServiceManager) Unsubscribe(platform, symbol string) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	svc := m.GetService(platform)
	if svc != nil {
		removed := svc.Unsubscribe(symbol)
		// 引用计数归零时，同步退订WS，进一步降低订阅数与网络开销（失败不阻断）
		if removed {
			m.unsubscribeWebSocket(platform, symbol)
			// 同时清理回调
			m.UnsubscribeCallback(platform, symbol, nil)
		}
	}
}

// triggerPriceCallbacks 触发价格更新回调
// 【新增】用于实时触发引擎的止损止盈检查
func (m *MarketServiceManager) triggerPriceCallbacks(platform, symbol string, ticker *exchange.Ticker) {
	if ticker == nil {
		return
	}

	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	key := platform + ":" + symbol
	m.mu.RLock()
	ch := m.callbackQueues[key]
	hasCallbacks := len(m.priceCallbacks[key]) > 0
	m.mu.RUnlock()

	// 没有注册回调：直接返回（保证报价更新链路最短）
	if !hasCallbacks || ch == nil {
		return
	}

	// 非阻塞入队：coalesce，只保留最新 ticker
	select {
	case ch <- ticker:
	default:
		// 队列满：丢弃旧的，保留最新
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- ticker:
		default:
		}
	}
}

// unsubscribeWebSocket 取消订阅WebSocket行情（best-effort）
func (m *MarketServiceManager) unsubscribeWebSocket(platform, symbol string) {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	m.mu.RLock()
	wsEnabled := m.wsEnabled
	m.mu.RUnlock()
	if !wsEnabled {
		return
	}

	switch platform {
	case "binance":
		if m.binanceWS != nil && m.binanceWS.IsRunning() {
			_ = m.binanceWS.UnsubscribeTicker(symbol)
			_ = m.binanceWS.UnsubscribeMarkPrice(symbol)
		}
	case "okx":
		if m.okxWS != nil && m.okxWS.IsRunning() {
			_ = m.okxWS.UnsubscribeTicker(symbol)
			_ = m.okxWS.UnsubscribeMarkPrice(symbol)
			for _, interval := range []string{"1m", "5m", "15m", "30m", "1h"} {
				_ = m.okxWS.UnsubscribeKline(symbol, interval)
			}
		}
	case "gate":
		if m.gateWS != nil && m.gateWS.IsRunning() {
			_ = m.gateWS.UnsubscribeTicker(symbol)
			for _, interval := range []string{"1m", "5m", "15m", "30m", "1h"} {
				_ = m.gateWS.UnsubscribeKline(symbol, interval)
			}
		}
	}
}

// GetTicker 获取实时行情（优先WebSocket，降级HTTP）
func (m *MarketServiceManager) GetTicker(platform, symbol string) *exchange.Ticker {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	// 优先从WebSocket获取（如果启用且连接正常）
	m.mu.RLock()
	wsEnabled := m.wsEnabled
	wsOnly := m.wsOnly
	m.mu.RUnlock()

	if wsEnabled {
		ticker := m.getTickerFromWebSocket(platform, symbol)
		if ticker != nil {
			return ticker
		}
		// WS-only：不允许降级到HTTP缓存/REST
		if wsOnly {
			return nil
		}
	}

	// 降级到HTTP缓存
	svc := m.GetService(platform)
	if svc == nil {
		return nil
	}
	return svc.GetTicker(symbol)
}

// getTickerFromWebSocket 从WebSocket获取Ticker
func (m *MarketServiceManager) getTickerFromWebSocket(platform, symbol string) *exchange.Ticker {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	switch platform {
	case "binance":
		if m.binanceWS != nil && m.binanceWS.IsRunning() {
			return m.binanceWS.GetTicker(symbol)
		}
	case "okx":
		if m.okxWS != nil && m.okxWS.IsRunning() {
			return m.okxWS.GetTicker(symbol)
		}
	case "gate":
		if m.gateWS != nil && m.gateWS.IsRunning() {
			return m.gateWS.GetTicker(symbol)
		}
	}
	return nil
}

// GetKlines 获取K线数据
func (m *MarketServiceManager) GetKlines(platform, symbol, interval string) []*exchange.Kline {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	// 优先从WebSocket获取（如果启用且连接正常）
	m.mu.RLock()
	wsEnabled := m.wsEnabled
	wsOnly := m.wsOnly
	m.mu.RUnlock()
	if wsEnabled {
		if kl := m.getKlinesFromWebSocket(platform, symbol, interval); len(kl) > 0 {
			return kl
		}
		// WS-only：不允许降级到HTTP缓存/REST
		if wsOnly {
			return nil
		}
	}
	svc := m.GetService(platform)
	if svc == nil {
		return nil
	}
	return svc.GetKlines(symbol, interval)
}

// getKlinesFromWebSocket 从WebSocket获取K线
func (m *MarketServiceManager) getKlinesFromWebSocket(platform, symbol, interval string) []*exchange.Kline {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	switch platform {
	case "binance":
		if m.binanceWS != nil && m.binanceWS.IsRunning() {
			return m.binanceWS.GetKlines(symbol, interval)
		}
	case "okx":
		if m.okxWS != nil && m.okxWS.IsRunning() {
			return m.okxWS.GetKlines(symbol, interval)
		}
	case "gate":
		if m.gateWS != nil && m.gateWS.IsRunning() {
			return m.gateWS.GetKlines(symbol, interval)
		}
	}
	return nil
}

// GetMultiTimeframeKlines 获取多周期K线
func (m *MarketServiceManager) GetMultiTimeframeKlines(platform, symbol string) *KlineCache {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	svc := m.GetService(platform)
	if svc == nil {
		return nil
	}
	return svc.GetMultiTimeframeKlines(symbol)
}

// GetMarketState 获取市场状态（全局共享，按 platform+symbol 缓存）
// 每个币种（platform+symbol）有独立的市场状态信号，所有交易该币种的机器人共享同一套信号
// 例如：okx:BTCUSDT 和 binance:BTCUSDT 是两套独立的市场状态
// 【实时性优化】添加数据过期检查，确保使用最新的市场状态数据
func (m *MarketServiceManager) GetMarketState(platform, symbol string) string {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	if platform == "" || symbol == "" {
		return ""
	}

	// 允许“执行所/分析所”解耦：仅市场状态(K线分析)可按配置覆写数据源平台。
	analysisPlatform := ResolveAnalysisPlatform(context.Background(), platform)
	if analysisPlatform == "" {
		analysisPlatform = platform
	}

	// 使用全局 MarketAnalyzer 获取市场状态（按 analysisPlatform+symbol 唯一标识）
	analyzer := GetMarketAnalyzer()
	analysis := analyzer.GetAnalysis(analysisPlatform, symbol)
	if analysis == nil {
		return ""
	}

	// 验证分析结果是否匹配请求的币种（确保数据一致性）
	if analysis.Platform != analysisPlatform || analysis.Symbol != symbol {
		return ""
	}

	// 【实时性优化】检查数据是否过期（超过3秒认为过期，返回空表示数据不可用）
	// 超短线交易需要实时数据，过期数据可能导致错误决策
	if time.Since(analysis.UpdatedAt) > 3*time.Second {
		g.Log().Warningf(context.Background(),
			"[MarketServiceManager] 市场状态数据过期: execPlatform=%s, analysisPlatform=%s, symbol=%s, age=%v",
			platform, analysisPlatform, symbol, time.Since(analysis.UpdatedAt))
		return "" // 返回空，表示数据不可用
	}

	// 规范化市场状态格式（range → volatile）
	marketState := string(analysis.MarketState)
	if marketState == "range" {
		marketState = "volatile"
	}
	return marketState
}

// IsDataFresh 检查数据是否新鲜
func (m *MarketServiceManager) IsDataFresh(platform, symbol string, maxAge time.Duration) bool {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	svc := m.GetService(platform)
	if svc == nil {
		return false
	}
	return svc.IsDataFresh(symbol, maxAge)
}

// GetAllServices 获取所有交易所服务
func (m *MarketServiceManager) GetAllServices() map[string]*ExchangeMarketService {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ExchangeMarketService, len(m.services))
	for k, v := range m.services {
		result[k] = v
	}
	return result
}

// GetActiveServiceCount 获取活跃服务数
func (m *MarketServiceManager) GetActiveServiceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.services)
}

// ==================== ExchangeMarketService 方法 ====================

// Start 启动交易所行情服务
func (s *ExchangeMarketService) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	// WS-only 模式：禁用所有REST轮询更新（WS回调会写入缓存）
	if !s.WSOnly {
		// 启动定时更新任务
		go s.runTickerUpdater(ctx)
		go s.runKlineUpdater(ctx)
	}

	g.Log().Infof(ctx, "[ExchangeMarketService] %s 行情服务启动", s.Platform)
}

// Stop 停止服务
func (s *ExchangeMarketService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)
}

// Subscribe 订阅交易对
func (s *ExchangeMarketService) Subscribe(ctx context.Context, symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Subscriptions[symbol]++
	if s.Subscriptions[symbol] == 1 {
		// 首次订阅：WS-only 模式不做REST初始拉取，等待WS回填缓存
		if !s.WSOnly {
			go s.fetchInitialData(ctx, symbol)
		}
	}

	g.Log().Debugf(ctx, "[ExchangeMarketService] %s 订阅 %s, 引用数=%d", s.Platform, symbol, s.Subscriptions[symbol])
}

// Unsubscribe 取消订阅
// Unsubscribe 取消订阅
// 返回值 removed: 是否在本次调用中将 refCount 归零并删除订阅（用于上层决定是否退订WS）
func (s *ExchangeMarketService) Unsubscribe(symbol string) (removed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if count, ok := s.Subscriptions[symbol]; ok {
		s.Subscriptions[symbol] = count - 1
		if s.Subscriptions[symbol] <= 0 {
			delete(s.Subscriptions, symbol)
			delete(s.Tickers, symbol)
			delete(s.Klines, symbol)
			return true
		}
	}
	return false
}

// GetTicker 获取Ticker（从缓存）
func (s *ExchangeMarketService) GetTicker(symbol string) *exchange.Ticker {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if cache, ok := s.Tickers[symbol]; ok {
		if time.Since(cache.UpdatedAt) < 10*time.Second {
			return cache.Data
		}
	}
	return nil
}

// GetKlines 获取K线数据
func (s *ExchangeMarketService) GetKlines(symbol, interval string) []*exchange.Kline {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if cache, ok := s.Klines[symbol]; ok {
		switch interval {
		case "1m":
			return cache.Klines1m
		case "5m":
			return cache.Klines5m
		case "15m":
			return cache.Klines15m
		case "30m":
			return cache.Klines30m
		case "1h":
			return cache.Klines1h
		}
	}
	return nil
}

// GetMultiTimeframeKlines 获取多周期K线
func (s *ExchangeMarketService) GetMultiTimeframeKlines(symbol string) *KlineCache {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Klines[symbol]
}

// IsDataFresh 检查数据是否新鲜
func (s *ExchangeMarketService) IsDataFresh(symbol string, maxAge time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if cache, ok := s.Tickers[symbol]; ok {
		return time.Since(cache.UpdatedAt) < maxAge
	}
	return false
}

// GetSubscriptionCount 获取订阅数
func (s *ExchangeMarketService) GetSubscriptionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Subscriptions)
}

// GetAllSubscriptions 获取所有订阅
func (s *ExchangeMarketService) GetAllSubscriptions() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int, len(s.Subscriptions))
	for k, v := range s.Subscriptions {
		result[k] = v
	}
	return result
}

// fetchInitialData 获取初始数据
func (s *ExchangeMarketService) fetchInitialData(ctx context.Context, symbol string) {
	if s.WSOnly {
		return
	}
	if s.Exchange == nil {
		return
	}

	// 获取Ticker
	ticker, err := s.Exchange.GetTicker(ctx, symbol)
	if err == nil {
		s.ensureMarkPrice(ctx, symbol, ticker)
		s.mu.Lock()
		s.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
		s.mu.Unlock()
	}

	// 获取K线
	s.fetchAllKlines(ctx, symbol)
}

// ensureMarkPrice 确保 ticker.MarkPrice 有值（风控/止盈止损/浮动盈亏统一口径）
// 策略：
// - 优先使用 ticker 自带 MarkPrice（来自WS/REST）
// - 若缺失且 ExchangeAdvanced 可用，则低频调用 GetFundingRate 补齐 MarkPrice/IndexPrice（按symbol限流）
// - 最终兜底：MarkPrice=LastPrice（保证 EffectiveMarkPrice 可用）
func (s *ExchangeMarketService) ensureMarkPrice(ctx context.Context, symbol string, ticker *exchange.Ticker) {
	if ticker == nil || ticker.MarkPrice > 0 {
		return
	}
	// 如果没有高级接口，直接用 LastPrice 兜底
	adv, ok := s.Exchange.(exchange.ExchangeAdvanced)
	if !ok {
		if ticker.LastPrice > 0 {
			ticker.MarkPrice = ticker.LastPrice
		}
		return
	}

	// 低频兜底（默认 5 秒）
	needFetch := false
	s.mu.Lock()
	lastAt := s.markPriceFallbackAt[symbol]
	if lastAt.IsZero() || time.Since(lastAt) >= 5*time.Second {
		s.markPriceFallbackAt[symbol] = time.Now()
		needFetch = true
	}
	s.mu.Unlock()

	if needFetch {
		if fr, err := adv.GetFundingRate(ctx, symbol); err == nil && fr != nil {
			if fr.MarkPrice > 0 {
				ticker.MarkPrice = fr.MarkPrice
			}
			if fr.IndexPrice > 0 {
				ticker.IndexPrice = fr.IndexPrice
			}
		}
	}

	if ticker.MarkPrice <= 0 && ticker.LastPrice > 0 {
		ticker.MarkPrice = ticker.LastPrice
	}
}

// fetchAllKlines 获取所有周期K线
func (s *ExchangeMarketService) fetchAllKlines(ctx context.Context, symbol string) {
	if s.Exchange == nil {
		return
	}

	cache := &KlineCache{UpdatedAt: time.Now()}
	var wg sync.WaitGroup
	var mu sync.Mutex

	intervals := []struct {
		interval string
		count    int
		target   *[]*exchange.Kline
	}{
		{"1m", 100, &cache.Klines1m},
		{"5m", 100, &cache.Klines5m},
		{"15m", 100, &cache.Klines15m},
		{"30m", 50, &cache.Klines30m},
		{"1h", 50, &cache.Klines1h},
	}

	for _, item := range intervals {
		wg.Add(1)
		go func(interval string, count int, target *[]*exchange.Kline) {
			defer wg.Done()
			// WS优先：如果WS有数据，直接使用
			if wsK := GetMarketServiceManager().getKlinesFromWebSocket(s.Platform, symbol, interval); len(wsK) > 0 {
				mu.Lock()
				*target = wsK
				mu.Unlock()
				return
			}

			// WS-only：不允许REST兜底
			if s.WSOnly {
				return
			}

			// REST兜底：拉取足够历史
			klines, err := s.Exchange.GetKlines(ctx, symbol, interval, count)
			if err == nil {
				mu.Lock()
				*target = klines
				mu.Unlock()
			} else {
				g.Log().Warningf(ctx, "[ExchangeMarketService] 获取K线失败: platform=%s, symbol=%s, interval=%s, error=%v", s.Platform, symbol, interval, err)
			}
		}(item.interval, item.count, item.target)
	}

	wg.Wait()

	s.mu.Lock()
	s.Klines[symbol] = cache
	s.mu.Unlock()

	// 记录获取结果
	g.Log().Infof(ctx, "[ExchangeMarketService] K线数据获取完成: platform=%s, symbol=%s, 1m=%d, 5m=%d, 15m=%d, 30m=%d, 1h=%d",
		s.Platform, symbol, len(cache.Klines1m), len(cache.Klines5m), len(cache.Klines15m), len(cache.Klines30m), len(cache.Klines1h))
}

// runTickerUpdater 定时更新Ticker
func (s *ExchangeMarketService) runTickerUpdater(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.updateAllTickers(ctx)
		}
	}
}

// runKlineUpdater 定时更新K线
func (s *ExchangeMarketService) runKlineUpdater(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.updateAllKlines(ctx)
		}
	}
}

// updateAllTickers 更新所有Ticker
func (s *ExchangeMarketService) updateAllTickers(ctx context.Context) {
	s.mu.RLock()
	symbols := make([]string, 0, len(s.Subscriptions))
	for symbol := range s.Subscriptions {
		symbols = append(symbols, symbol)
	}
	s.mu.RUnlock()

	if s.Exchange == nil {
		return
	}

	for _, symbol := range symbols {
		// WS优先：如果WS有ticker，直接回写缓存，避免REST调用
		if wsTicker := GetMarketServiceManager().getTickerFromWebSocket(s.Platform, symbol); wsTicker != nil {
			s.ensureMarkPrice(ctx, symbol, wsTicker)
			s.mu.Lock()
			s.Tickers[symbol] = &TickerCache{Data: wsTicker, UpdatedAt: time.Now()}
			s.mu.Unlock()
			continue
		}

		// WS-only：不允许REST兜底
		if s.WSOnly {
			continue
		}

		ticker, err := s.Exchange.GetTicker(ctx, symbol)
		if err != nil {
			continue
		}
		s.ensureMarkPrice(ctx, symbol, ticker)

		s.mu.Lock()
		s.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
		s.mu.Unlock()
	}
}

// updateAllKlines 更新所有K线
func (s *ExchangeMarketService) updateAllKlines(ctx context.Context) {
	// WS-only：K线由 WS 回调持续写入缓存，不做REST轮询
	if s.WSOnly {
		return
	}
	s.mu.RLock()
	symbols := make([]string, 0, len(s.Subscriptions))
	for symbol := range s.Subscriptions {
		symbols = append(symbols, symbol)
	}
	s.mu.RUnlock()

	for _, symbol := range symbols {
		s.fetchAllKlines(ctx, symbol)
	}
}

// FetchTickerDirect 直接获取Ticker（不经过缓存）
func (s *ExchangeMarketService) FetchTickerDirect(ctx context.Context, symbol string) (*exchange.Ticker, error) {
	if s.Exchange == nil {
		return nil, gerror.New("交易所实例未初始化")
	}
	return s.Exchange.GetTicker(ctx, symbol)
}

// ==================== WebSocket状态查询 ====================

// WebSocketStatus WebSocket状态
type WebSocketStatus struct {
	Enabled       bool                      `json:"enabled"`
	BinanceStatus *exchange.BinanceWSStatus `json:"binance"`
	OKXStatus     *exchange.OKXWSStatus     `json:"okx"`
	GateStatus    *exchange.GateWSStatus    `json:"gate"`
}

// GetWebSocketStatus 获取WebSocket状态
func (m *MarketServiceManager) GetWebSocketStatus() *WebSocketStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &WebSocketStatus{
		Enabled: m.wsEnabled,
	}

	if m.binanceWS != nil {
		status.BinanceStatus = m.binanceWS.GetStatus()
	}
	if m.okxWS != nil {
		status.OKXStatus = m.okxWS.GetStatus()
	}
	if m.gateWS != nil {
		status.GateStatus = m.gateWS.GetStatus()
	}

	return status
}

// IsWebSocketEnabled 检查WebSocket是否启用
func (m *MarketServiceManager) IsWebSocketEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.wsEnabled
}

// EnableWebSocket 运行时启用WebSocket
func (m *MarketServiceManager) EnableWebSocket(ctx context.Context) error {
	m.mu.Lock()
	if m.wsEnabled {
		m.mu.Unlock()
		return nil
	}
	m.wsEnabled = true
	m.mu.Unlock()

	m.startWebSocketServices(ctx)
	return nil
}

// DisableWebSocket 运行时禁用WebSocket
func (m *MarketServiceManager) DisableWebSocket() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.wsEnabled {
		return
	}

	m.wsEnabled = false
	if m.binanceWS != nil {
		m.binanceWS.Stop()
	}

	g.Log().Info(context.Background(), "[MarketServiceManager] WebSocket已禁用")
}
