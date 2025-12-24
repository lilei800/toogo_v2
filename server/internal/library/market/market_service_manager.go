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
	return strings.ToLower(strings.TrimSpace(platform))
}

func normalizeSymbol(symbol string) string {
	// 仅做轻量规范化：去空格 + 大写。避免破坏诸如 OKX 的 instId 格式（若业务层直接传 instId）。
	return strings.ToUpper(strings.TrimSpace(symbol))
}

// MarketServiceManager 全局行情服务管理器（单例）
// 管理每个交易所的独立行情服务
type MarketServiceManager struct {
	mu sync.RWMutex

	// 每个交易所一个行情服务 key: platform (binance/bitget/okx/gate)
	services map[string]*ExchangeMarketService

	// WebSocket服务（优先使用）
	wsEnabled bool
	bitgetWS  *exchange.BitgetWebSocket
	binanceWS *exchange.BinanceWebSocket
	okxWS     *exchange.OKXWebSocket
	gateWS    *exchange.GateWebSocket

	// 代理配置
	proxyDialer func(network, addr string) (net.Conn, error)

	// 【新增】价格更新回调（用于实时触发引擎检查）
	// key: platform:symbol, value: 回调函数列表
	priceCallbacks map[string][]func(*exchange.Ticker)

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// ExchangeMarketService 单个交易所的行情服务
type ExchangeMarketService struct {
	mu sync.RWMutex

	Platform string            // 交易所名称
	Exchange exchange.Exchange // 交易所API实例

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

	m.mu.Lock()
	m.wsEnabled = true
	proxyDialer := m.proxyDialer // 复制代理配置
	m.mu.Unlock()

	// 统一启动流程：减少重复代码
	successCount := 0
	totalCount := 4

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

	startWS("Bitget", func() interface{} { return exchange.GetBitgetWebSocket() }, func(ws interface{}) { m.bitgetWS = ws.(*exchange.BitgetWebSocket) })
	startWS("Binance", func() interface{} { return exchange.GetBinanceWebSocket() }, func(ws interface{}) { m.binanceWS = ws.(*exchange.BinanceWebSocket) })
	startWS("OKX", func() interface{} { return exchange.GetOKXWebSocket() }, func(ws interface{}) { m.okxWS = ws.(*exchange.OKXWebSocket) })
	startWS("Gate", func() interface{} { return exchange.GetGateWebSocket() }, func(ws interface{}) { m.gateWS = ws.(*exchange.GateWebSocket) })

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
		{"Bitget", m.bitgetWS},
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
	// 订阅WebSocket（如果启用）
	m.subscribeWebSocket(ctx, platform, symbol)

	// 同时保留HTTP服务（作为降级方案）
	svc := m.GetOrCreateService(ctx, platform, ex)
	svc.Subscribe(ctx, symbol)
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
		m.mu.Unlock()

		g.Log().Debugf(ctx, "[MarketServiceManager] 注册价格更新回调: %s", key)
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
	case "bitget":
		if m.bitgetWS != nil && m.bitgetWS.IsRunning() {
			// 订阅Ticker
			m.bitgetWS.SubscribeTicker(symbol, func(ticker *exchange.Ticker) {
				// WebSocket数据更新回调 - 更新HTTP服务的缓存以保持一致
				if svc := m.GetService(platform); svc != nil {
					svc.mu.Lock()
					svc.Tickers[symbol] = &TickerCache{Data: ticker, UpdatedAt: time.Now()}
					svc.mu.Unlock()
				}
				// 【新增】触发注册的回调函数（用于实时平仓检查）
				m.triggerPriceCallbacks(platform, symbol, ticker)
			})
			// 订阅K线（多周期）
			for _, interval := range []string{"1m", "5m", "15m", "30m", "1h"} {
				_ = m.bitgetWS.SubscribeKline(symbol, interval, func(klines []*exchange.Kline) {
					updateSvcKlines(interval, klines)
				})
			}
		}
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
		if m.gateWS != nil && m.gateWS.IsRunning() {
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
	callbacks := m.priceCallbacks[key]
	m.mu.RUnlock()

	// 异步调用所有回调，避免阻塞WebSocket处理
	for _, cb := range callbacks {
		if cb != nil {
			go cb(ticker)
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
	case "bitget":
		// 当前 BitgetWS 有 UnsubscribeTicker，但 Kline 退订在现有实现里未暴露；这里先退订ticker（最关键）
		if m.bitgetWS != nil && m.bitgetWS.IsRunning() {
			_ = m.bitgetWS.UnsubscribeTicker(symbol)
		}
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
	m.mu.RUnlock()

	if wsEnabled {
		ticker := m.getTickerFromWebSocket(platform, symbol)
		if ticker != nil {
			return ticker
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
	case "bitget":
		if m.bitgetWS != nil && m.bitgetWS.IsRunning() {
			return m.bitgetWS.GetTicker(symbol)
		}
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
	m.mu.RUnlock()
	if wsEnabled {
		if kl := m.getKlinesFromWebSocket(platform, symbol, interval); len(kl) > 0 {
			return kl
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
	case "bitget":
		if m.bitgetWS != nil && m.bitgetWS.IsRunning() {
			return m.bitgetWS.GetKlines(symbol, interval)
		}
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
// 例如：bitget:BTCUSDT 和 binance:BTCUSDT 是两套独立的市场状态
// 【实时性优化】添加数据过期检查，确保使用最新的市场状态数据
func (m *MarketServiceManager) GetMarketState(platform, symbol string) string {
	platform = normalizePlatform(platform)
	symbol = normalizeSymbol(symbol)
	if platform == "" || symbol == "" {
		return ""
	}

	// 使用全局 MarketAnalyzer 获取市场状态（按 platform+symbol 唯一标识）
	analyzer := GetMarketAnalyzer()
	analysis := analyzer.GetAnalysis(platform, symbol)
	if analysis == nil {
		return ""
	}

	// 验证分析结果是否匹配请求的币种（确保数据一致性）
	if analysis.Platform != platform || analysis.Symbol != symbol {
		return ""
	}

	// 【实时性优化】检查数据是否过期（超过3秒认为过期，返回空表示数据不可用）
	// 超短线交易需要实时数据，过期数据可能导致错误决策
	if time.Since(analysis.UpdatedAt) > 3*time.Second {
		g.Log().Warningf(context.Background(),
			"[MarketServiceManager] 市场状态数据过期: platform=%s, symbol=%s, age=%v",
			platform, symbol, time.Since(analysis.UpdatedAt))
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

	// 启动定时更新任务
	go s.runTickerUpdater(ctx)
	go s.runKlineUpdater(ctx)

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
		// 首次订阅，立即获取数据
		go s.fetchInitialData(ctx, symbol)
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
			// WS优先：如果WS有数据，直接使用，避免REST请求
			if wsK := GetMarketServiceManager().getKlinesFromWebSocket(s.Platform, symbol, interval); len(wsK) > 0 {
				mu.Lock()
				*target = wsK
				mu.Unlock()
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
	BitgetStatus  *exchange.BitgetWSStatus  `json:"bitget"`
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

	if m.bitgetWS != nil {
		status.BitgetStatus = m.bitgetWS.GetStatus()
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

	if m.bitgetWS != nil {
		m.bitgetWS.Stop()
	}
	if m.binanceWS != nil {
		m.binanceWS.Stop()
	}

	g.Log().Info(context.Background(), "[MarketServiceManager] WebSocket已禁用")
}
