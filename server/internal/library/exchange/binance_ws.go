// Package exchange Binance WebSocket行情服务
package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	BinanceWSFuturesURL = "wss://fstream.binance.com/ws"
)

// BinanceWebSocket Binance WebSocket行情服务
type BinanceWebSocket struct {
	mu   sync.RWMutex
	conn *WebSocketConnection

	// 行情数据缓存
	tickers map[string]*Ticker
	klines  map[string][]*Kline

	// 回调管理
	tickerCallbacks map[string][]func(*Ticker)
	klineCallbacks  map[string][]func([]*Kline)

	// 订阅管理
	subscribedStreams map[string]bool

	// 状态
	running bool
	ctx     context.Context
	cancel  context.CancelFunc

	// 代理配置
	proxyDialer func(network, addr string) (net.Conn, error)
}

// NewBinanceWebSocket 创建Binance WebSocket服务
func NewBinanceWebSocket() *BinanceWebSocket {
	return &BinanceWebSocket{
		tickers:           make(map[string]*Ticker),
		klines:            make(map[string][]*Kline),
		tickerCallbacks:   make(map[string][]func(*Ticker)),
		klineCallbacks:    make(map[string][]func([]*Kline)),
		subscribedStreams: make(map[string]bool),
	}
}

// SetProxyDialer 设置代理拨号器
func (b *BinanceWebSocket) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.proxyDialer = dialer
}

// Start 启动WebSocket服务
func (b *BinanceWebSocket) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = true
	b.ctx, b.cancel = context.WithCancel(ctx)
	proxyDialer := b.proxyDialer // 复制代理配置
	b.mu.Unlock()

	config := DefaultWebSocketConfig()
	config.URL = BinanceWSFuturesURL
	config.PingInterval = 3 * time.Minute // Binance的心跳间隔较长
	config.ProxyDialer = proxyDialer      // 设置代理

	b.conn = NewWebSocketConnection(config)
	b.conn.SetCallbacks(b.onMessage, b.onConnected, b.onDisconnected)

	if err := b.conn.Connect(b.ctx); err != nil {
		b.mu.Lock()
		b.running = false
		b.mu.Unlock()
		return err
	}

	g.Log().Info(ctx, "[BinanceWS] WebSocket服务已启动")
	return nil
}

// Stop 停止WebSocket服务
func (b *BinanceWebSocket) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}

	b.running = false
	if b.cancel != nil {
		b.cancel()
	}
	if b.conn != nil {
		b.conn.Disconnect()
	}

	g.Log().Info(context.Background(), "[BinanceWS] WebSocket服务已停止")
}

// IsRunning 检查是否运行中
func (b *BinanceWebSocket) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running && b.conn != nil && b.conn.IsConnected()
}

// SubscribeTicker 订阅Ticker行情
func (b *BinanceWebSocket) SubscribeTicker(symbol string, callback func(*Ticker)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	stream := strings.ToLower(symbol) + "@ticker"

	b.tickerCallbacks[symbol] = append(b.tickerCallbacks[symbol], callback)

	if b.subscribedStreams[stream] {
		return nil
	}

	// Binance使用SUBSCRIBE消息
	sub := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{stream},
		"id":     time.Now().UnixNano(),
	}

	if err := b.conn.Send(sub); err != nil {
		return err
	}

	b.subscribedStreams[stream] = true
	b.conn.SaveSubscription("ticker:"+symbol, sub)

	g.Log().Infof(b.ctx, "[BinanceWS] 订阅Ticker: %s", stream)

	// 同时订阅 markPrice，用于统一风控/盈亏口径（best-effort）
	_ = b.subscribeMarkPriceLocked(symbol)
	return nil
}

// SubscribeMarkPrice 订阅标记价格（markPriceUpdate）
func (b *BinanceWebSocket) SubscribeMarkPrice(symbol string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.subscribeMarkPriceLocked(symbol)
}

func (b *BinanceWebSocket) subscribeMarkPriceLocked(symbol string) error {
	// markPrice stream: <symbol>@markPrice@1s
	stream := strings.ToLower(symbol) + "@markPrice@1s"
	if b.subscribedStreams[stream] {
		return nil
	}
	sub := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{stream},
		"id":     time.Now().UnixNano(),
	}
	if err := b.conn.Send(sub); err != nil {
		return err
	}
	b.subscribedStreams[stream] = true
	b.conn.SaveSubscription("mark:"+symbol, sub)
	g.Log().Infof(b.ctx, "[BinanceWS] 订阅MarkPrice: %s", stream)
	return nil
}

// UnsubscribeTicker 取消订阅Ticker
func (b *BinanceWebSocket) UnsubscribeTicker(symbol string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	stream := strings.ToLower(symbol) + "@ticker"
	delete(b.tickerCallbacks, symbol)

	if !b.subscribedStreams[stream] {
		return nil
	}

	unsub := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": []string{stream},
		"id":     time.Now().UnixNano(),
	}

	if err := b.conn.Send(unsub); err != nil {
		return err
	}

	delete(b.subscribedStreams, stream)
	b.conn.RemoveSubscription("ticker:" + symbol)

	g.Log().Infof(b.ctx, "[BinanceWS] 取消订阅Ticker: %s", stream)

	// best-effort 同步退订markPrice
	_ = b.unsubscribeMarkPriceLocked(symbol)
	return nil
}

// UnsubscribeMarkPrice 取消订阅标记价格
func (b *BinanceWebSocket) UnsubscribeMarkPrice(symbol string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.unsubscribeMarkPriceLocked(symbol)
}

func (b *BinanceWebSocket) unsubscribeMarkPriceLocked(symbol string) error {
	stream := strings.ToLower(symbol) + "@markPrice@1s"
	if !b.subscribedStreams[stream] {
		return nil
	}
	unsub := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": []string{stream},
		"id":     time.Now().UnixNano(),
	}
	if err := b.conn.Send(unsub); err != nil {
		return err
	}
	delete(b.subscribedStreams, stream)
	b.conn.RemoveSubscription("mark:" + symbol)
	g.Log().Infof(b.ctx, "[BinanceWS] 取消订阅MarkPrice: %s", stream)
	return nil
}

// SubscribeKline 订阅K线
func (b *BinanceWebSocket) SubscribeKline(symbol, interval string, callback func([]*Kline)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	stream := strings.ToLower(symbol) + "@kline_" + interval
	key := symbol + ":" + interval

	if callback != nil {
		b.klineCallbacks[key] = append(b.klineCallbacks[key], callback)
	}

	if b.subscribedStreams[stream] {
		return nil
	}

	sub := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{stream},
		"id":     time.Now().UnixNano(),
	}

	if err := b.conn.Send(sub); err != nil {
		return err
	}

	b.subscribedStreams[stream] = true
	b.conn.SaveSubscription("kline:"+key, sub)

	g.Log().Infof(b.ctx, "[BinanceWS] 订阅K线: %s", stream)
	return nil
}

// GetTicker 获取最新Ticker
func (b *BinanceWebSocket) GetTicker(symbol string) *Ticker {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.tickers[symbol]
}

// GetKlines 获取K线数据
func (b *BinanceWebSocket) GetKlines(symbol, interval string) []*Kline {
	b.mu.RLock()
	defer b.mu.RUnlock()
	key := symbol + ":" + interval
	return b.klines[key]
}

// onMessage 处理消息
func (b *BinanceWebSocket) onMessage(msg []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		return
	}

	// 检查是否是订阅响应
	if _, ok := data["result"]; ok {
		return
	}

	// 获取事件类型
	eventType, ok := data["e"].(string)
	if !ok {
		return
	}

	switch eventType {
	case "24hrTicker":
		b.handleTickerData(data)
	case "markPriceUpdate":
		b.handleMarkPriceData(data)
	case "kline":
		b.handleKlineData(data)
	}
}

// handleTickerData 处理Ticker数据
func (b *BinanceWebSocket) handleTickerData(data map[string]interface{}) {
	symbol := data["s"].(string)

	priceChangePercent := parseFloat(data["P"])

	b.mu.Lock()
	ticker := b.tickers[symbol]
	if ticker == nil {
		ticker = &Ticker{Symbol: symbol}
		b.tickers[symbol] = ticker
	}
	ticker.LastPrice = parseFloat(data["c"])
	ticker.BidPrice = parseFloat(data["b"])
	ticker.AskPrice = parseFloat(data["a"])
	ticker.High24h = parseFloat(data["h"])
	ticker.Low24h = parseFloat(data["l"])
	ticker.Volume24h = parseFloat(data["v"])
	ticker.QuoteVolume24h = parseFloat(data["q"])
	ticker.Change24h = priceChangePercent
	ticker.PriceChangePercent = priceChangePercent
	ticker.Timestamp = parseInt(data["E"])
	callbacks := b.tickerCallbacks[symbol]
	b.mu.Unlock()

	for _, cb := range callbacks {
		if cb != nil {
			go cb(ticker)
		}
	}
}

// handleMarkPriceData 处理标记价格更新
func (b *BinanceWebSocket) handleMarkPriceData(data map[string]interface{}) {
	symbolVar, ok := data["s"]
	if !ok {
		return
	}
	symbol, _ := symbolVar.(string)
	if symbol == "" {
		return
	}

	b.mu.Lock()
	ticker := b.tickers[symbol]
	if ticker == nil {
		ticker = &Ticker{Symbol: symbol}
		b.tickers[symbol] = ticker
	}
	ticker.MarkPrice = parseFloat(data["p"])
	ticker.IndexPrice = parseFloat(data["i"])
	// 尽量更新timestamp（E 为事件时间）
	if ts := parseInt(data["E"]); ts > 0 {
		ticker.Timestamp = ts
	}
	callbacks := b.tickerCallbacks[symbol]
	b.mu.Unlock()

	for _, cb := range callbacks {
		if cb != nil {
			go cb(ticker)
		}
	}
}

// handleKlineData 处理K线数据
func (b *BinanceWebSocket) handleKlineData(data map[string]interface{}) {
	klineData, ok := data["k"].(map[string]interface{})
	if !ok {
		return
	}

	symbol := klineData["s"].(string)
	interval := klineData["i"].(string)
	key := symbol + ":" + interval

	kline := &Kline{
		OpenTime:  parseInt(klineData["t"]),
		Open:      parseFloat(klineData["o"]),
		High:      parseFloat(klineData["h"]),
		Low:       parseFloat(klineData["l"]),
		Close:     parseFloat(klineData["c"]),
		Volume:    parseFloat(klineData["v"]),
		CloseTime: parseInt(klineData["T"]),
	}

	b.mu.Lock()
	existing := b.klines[key]
	if len(existing) > 0 {
		if existing[len(existing)-1].OpenTime == kline.OpenTime {
			existing[len(existing)-1] = kline
		} else {
			existing = append(existing, kline)
			if len(existing) > 500 {
				existing = existing[len(existing)-500:]
			}
		}
		b.klines[key] = existing
	} else {
		b.klines[key] = []*Kline{kline}
	}
	callbacks := b.klineCallbacks[key]
	klinesCopy := b.klines[key]
	b.mu.Unlock()

	for _, cb := range callbacks {
		if cb != nil {
			go cb(klinesCopy)
		}
	}
}

// onConnected 连接成功回调
func (b *BinanceWebSocket) onConnected() {
	g.Log().Info(b.ctx, "[BinanceWS] 连接成功，恢复订阅...")

	subs := b.conn.GetSubscriptions()
	for _, sub := range subs {
		if err := b.conn.Send(sub); err != nil {
			g.Log().Warningf(b.ctx, "[BinanceWS] 恢复订阅失败: %v", err)
		}
	}
}

// onDisconnected 断开连接回调
func (b *BinanceWebSocket) onDisconnected(err error) {
	g.Log().Warningf(b.ctx, "[BinanceWS] 连接断开: %v", err)
}

// GetSubscriptionCount 获取订阅数量
func (b *BinanceWebSocket) GetSubscriptionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribedStreams)
}

// GetConnectionState 获取连接状态
func (b *BinanceWebSocket) GetConnectionState() string {
	if b.conn == nil {
		return "disconnected"
	}
	switch b.conn.GetState() {
	case WSStateConnected:
		return "connected"
	case WSStateConnecting:
		return "connecting"
	case WSStateReconnecting:
		return "reconnecting"
	default:
		return "disconnected"
	}
}

// ============ 静态实例管理 ============

var (
	binanceWSInstance     *BinanceWebSocket
	binanceWSInstanceOnce sync.Once
)

// GetBinanceWebSocket 获取Binance WebSocket单例
func GetBinanceWebSocket() *BinanceWebSocket {
	binanceWSInstanceOnce.Do(func() {
		binanceWSInstance = NewBinanceWebSocket()
	})
	return binanceWSInstance
}

// BinanceWSStatus Binance WebSocket状态
type BinanceWSStatus struct {
	Running           bool              `json:"running"`
	ConnectionState   string            `json:"connectionState"`
	SubscriptionCount int               `json:"subscriptionCount"`
	TickerCount       int               `json:"tickerCount"`
	Tickers           map[string]string `json:"tickers"`
}

// GetStatus 获取状态
func (b *BinanceWebSocket) GetStatus() *BinanceWSStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tickers := make(map[string]string, len(b.tickers))
	for k, v := range b.tickers {
		tickers[k] = fmt.Sprintf("%.4f", v.LastPrice)
	}

	return &BinanceWSStatus{
		Running:           b.running,
		ConnectionState:   b.GetConnectionState(),
		SubscriptionCount: len(b.subscribedStreams),
		TickerCount:       len(b.tickers),
		Tickers:           tickers,
	}
}
