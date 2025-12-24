// Package market 全局行情数据服务
// 负责统一管理所有交易所的实时行情数据，为所有机器人提供数据支持
package market

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/library/exchange"
)

// MarketDataService 全局行情数据服务（单例）
type MarketDataService struct {
	mu sync.RWMutex

	// 交易所连接管理
	exchanges map[string]exchange.Exchange // key: platform

	// 行情数据缓存 key: platform:symbol
	tickers     map[string]*TickerData
	klines      map[string]*KlineCache
	orderBooks  map[string]*OrderBookData

	// 订阅管理
	subscriptions map[string]*Subscription // key: platform:symbol

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// TickerData 行情数据（带时间戳）
type TickerData struct {
	Ticker    *exchange.Ticker
	UpdatedAt time.Time
}

// KlineCache K线缓存（多周期）
type KlineCache struct {
	Klines1m  []*exchange.Kline // 1分钟K线
	Klines5m  []*exchange.Kline // 5分钟K线
	Klines15m []*exchange.Kline // 15分钟K线
	Klines30m []*exchange.Kline // 30分钟K线
	Klines1h  []*exchange.Kline // 1小时K线
	Klines1d  []*exchange.Kline // 1天K线（【新增】短线需要长期趋势参考）
	UpdatedAt time.Time
}

// OrderBookData 订单簿数据
type OrderBookData struct {
	Bids      [][2]float64 // [价格, 数量]
	Asks      [][2]float64
	UpdatedAt time.Time
}

// Subscription 订阅信息
type Subscription struct {
	Platform   string
	Symbol     string
	RefCount   int       // 引用计数（多少机器人在用）
	LastAccess time.Time // 最后访问时间
}

var (
	marketDataService     *MarketDataService
	marketDataServiceOnce sync.Once
)

// GetMarketDataService 获取行情数据服务单例
func GetMarketDataService() *MarketDataService {
	marketDataServiceOnce.Do(func() {
		marketDataService = &MarketDataService{
			exchanges:     make(map[string]exchange.Exchange),
			tickers:       make(map[string]*TickerData),
			klines:        make(map[string]*KlineCache),
			orderBooks:    make(map[string]*OrderBookData),
			subscriptions: make(map[string]*Subscription),
			stopCh:        make(chan struct{}),
		}
	})
	return marketDataService
}

// Start 启动行情数据服务
func (s *MarketDataService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	g.Log().Info(ctx, "[MarketDataService] 行情数据服务启动")

	// 启动定时更新任务
	go s.runTickerUpdater(ctx)
	go s.runKlineUpdater(ctx)
	go s.runCleanupTask(ctx)

	return nil
}

// Stop 停止行情数据服务
func (s *MarketDataService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)
}

// Subscribe 订阅交易对行情
func (s *MarketDataService) Subscribe(ctx context.Context, platform, symbol string, ex exchange.Exchange) {
	key := platform + ":" + symbol

	s.mu.Lock()
	defer s.mu.Unlock()

	// 保存交易所连接
	if _, ok := s.exchanges[platform]; !ok {
		s.exchanges[platform] = ex
	}

	// 增加订阅引用
	if sub, ok := s.subscriptions[key]; ok {
		sub.RefCount++
		sub.LastAccess = time.Now()
	} else {
		s.subscriptions[key] = &Subscription{
			Platform:   platform,
			Symbol:     symbol,
			RefCount:   1,
			LastAccess: time.Now(),
		}
		// 立即获取一次数据
		go s.fetchInitialData(ctx, platform, symbol)
	}

	g.Log().Debugf(ctx, "[MarketDataService] 订阅行情: %s, 引用数=%d", key, s.subscriptions[key].RefCount)
}

// Unsubscribe 取消订阅
func (s *MarketDataService) Unsubscribe(platform, symbol string) {
	key := platform + ":" + symbol

	s.mu.Lock()
	defer s.mu.Unlock()

	if sub, ok := s.subscriptions[key]; ok {
		sub.RefCount--
		if sub.RefCount <= 0 {
			delete(s.subscriptions, key)
			delete(s.tickers, key)
			delete(s.klines, key)
		}
	}
}

// GetTicker 获取实时行情（从缓存）
func (s *MarketDataService) GetTicker(platform, symbol string) *exchange.Ticker {
	key := platform + ":" + symbol

	s.mu.RLock()
	defer s.mu.RUnlock()

	if data, ok := s.tickers[key]; ok {
		// 检查数据是否过期（超过10秒视为过期）
		if time.Since(data.UpdatedAt) < 10*time.Second {
			return data.Ticker
		}
	}
	return nil
}

// GetKlines 获取K线数据（从缓存）
func (s *MarketDataService) GetKlines(platform, symbol, interval string) []*exchange.Kline {
	key := platform + ":" + symbol

	s.mu.RLock()
	defer s.mu.RUnlock()

	if cache, ok := s.klines[key]; ok {
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

// GetMultiTimeframeKlines 获取多周期K线数据
func (s *MarketDataService) GetMultiTimeframeKlines(platform, symbol string) *KlineCache {
	key := platform + ":" + symbol

	s.mu.RLock()
	defer s.mu.RUnlock()

	if cache, ok := s.klines[key]; ok {
		return cache
	}
	return nil
}

// fetchInitialData 获取初始数据
func (s *MarketDataService) fetchInitialData(ctx context.Context, platform, symbol string) {
	ex := s.getExchange(platform)
	if ex == nil {
		return
	}

	key := platform + ":" + symbol

	// 获取Ticker
	ticker, err := ex.GetTicker(ctx, symbol)
	if err == nil {
		s.mu.Lock()
		s.tickers[key] = &TickerData{
			Ticker:    ticker,
			UpdatedAt: time.Now(),
		}
		s.mu.Unlock()
	}

	// 获取多周期K线
	s.fetchAllKlines(ctx, ex, platform, symbol)
}

// fetchAllKlines 获取所有周期K线（【优化】主动获取历史K线数据，不等待时间积累）
func (s *MarketDataService) fetchAllKlines(ctx context.Context, ex exchange.Exchange, platform, symbol string) {
	key := platform + ":" + symbol
	cache := &KlineCache{UpdatedAt: time.Now()}

	// 并行获取多周期K线
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 【优化】增加历史K线获取数量，确保有足够数据计算基准波动率
	// 从交易所API主动获取历史K线数据，而不是等待时间积累
	// 1h: 从50根增加到200根（约8.3天历史数据，足够计算基准波动率）
	// 15m: 从100根增加到200根（约2.1天历史数据）
	// 5m: 从100根增加到200根（约16.7小时历史数据）
	// 1m: 保持100根（约1.7小时，用于实时计算）
	intervals := []struct {
		interval string
		count    int
		target   *[]*exchange.Kline
	}{
		{"1m", 100, &cache.Klines1m},   // 1分钟：100根（约1.7小时）
		{"5m", 200, &cache.Klines5m},   // 5分钟：200根（约16.7小时）
		{"15m", 200, &cache.Klines15m}, // 15分钟：200根（约2.1天）
		{"30m", 100, &cache.Klines30m}, // 30分钟：100根（约2.1天）
		{"1h", 200, &cache.Klines1h},   // 1小时：200根（约8.3天，足够计算基准波动率）
		{"1d", 30, &cache.Klines1d},    // 1天：30根（约1个月，短线需要长期趋势参考）
	}

	for _, item := range intervals {
		wg.Add(1)
		go func(interval string, count int, target *[]*exchange.Kline) {
			defer wg.Done()
			klines, err := ex.GetKlines(ctx, symbol, interval, count)
			if err == nil {
				mu.Lock()
				*target = klines
				mu.Unlock()
			}
		}(item.interval, item.count, item.target)
	}

	wg.Wait()

	s.mu.Lock()
	s.klines[key] = cache
	s.mu.Unlock()

	// 【优化】记录获取的历史K线数量，便于调试
	g.Log().Debugf(ctx, "[MarketDataService] 已获取历史K线数据: platform=%s, symbol=%s, 1m=%d, 5m=%d, 15m=%d, 30m=%d, 1h=%d, 1d=%d",
		platform, symbol,
		len(cache.Klines1m), len(cache.Klines5m), len(cache.Klines15m),
		len(cache.Klines30m), len(cache.Klines1h), len(cache.Klines1d))
}

// RefreshKlines 主动刷新K线数据（【新增】供机器人引擎调用，确保有足够历史数据）
func (s *MarketDataService) RefreshKlines(ctx context.Context, platform, symbol string) error {
	ex := s.getExchange(platform)
	if ex == nil {
		return gerror.Newf("交易所 %s 未找到", platform)
	}

	// 主动获取历史K线数据
	s.fetchAllKlines(ctx, ex, platform, symbol)
	return nil
}

// runTickerUpdater 定时更新Ticker数据
func (s *MarketDataService) runTickerUpdater(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // 每秒更新一次
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

// runKlineUpdater 定时更新K线数据
func (s *MarketDataService) runKlineUpdater(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒更新一次
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

// runCleanupTask 清理过期数据
func (s *MarketDataService) runCleanupTask(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanupExpiredData()
		}
	}
}

// updateAllTickers 更新所有Ticker
func (s *MarketDataService) updateAllTickers(ctx context.Context) {
	s.mu.RLock()
	subs := make(map[string]*Subscription)
	for k, v := range s.subscriptions {
		subs[k] = v
	}
	s.mu.RUnlock()

	for key, sub := range subs {
		ex := s.getExchange(sub.Platform)
		if ex == nil {
			continue
		}

		ticker, err := ex.GetTicker(ctx, sub.Symbol)
		if err != nil {
			continue
		}

		s.mu.Lock()
		s.tickers[key] = &TickerData{
			Ticker:    ticker,
			UpdatedAt: time.Now(),
		}
		s.mu.Unlock()
	}
}

// updateAllKlines 更新所有K线
func (s *MarketDataService) updateAllKlines(ctx context.Context) {
	s.mu.RLock()
	subs := make(map[string]*Subscription)
	for k, v := range s.subscriptions {
		subs[k] = v
	}
	s.mu.RUnlock()

	for _, sub := range subs {
		ex := s.getExchange(sub.Platform)
		if ex == nil {
			continue
		}
		s.fetchAllKlines(ctx, ex, sub.Platform, sub.Symbol)
	}
}

// cleanupExpiredData 清理过期数据
func (s *MarketDataService) cleanupExpiredData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 清理超过5分钟未访问的无引用订阅
	for key, sub := range s.subscriptions {
		if sub.RefCount <= 0 && time.Since(sub.LastAccess) > 5*time.Minute {
			delete(s.subscriptions, key)
			delete(s.tickers, key)
			delete(s.klines, key)
		}
	}
}

// getExchange 获取交易所实例
func (s *MarketDataService) getExchange(platform string) exchange.Exchange {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.exchanges[platform]
}

// GetActiveSubscriptions 获取活跃订阅数
func (s *MarketDataService) GetActiveSubscriptions() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscriptions)
}

// IsDataFresh 检查数据是否新鲜
func (s *MarketDataService) IsDataFresh(platform, symbol string, maxAge time.Duration) bool {
	key := platform + ":" + symbol

	s.mu.RLock()
	defer s.mu.RUnlock()

	if data, ok := s.tickers[key]; ok {
		return time.Since(data.UpdatedAt) < maxAge
	}
	return false
}

// GetAllSubscriptions 获取所有订阅
func (s *MarketDataService) GetAllSubscriptions() map[string]*Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make(map[string]*Subscription, len(s.subscriptions))
	for k, v := range s.subscriptions {
		result[k] = v
	}
	return result
}

