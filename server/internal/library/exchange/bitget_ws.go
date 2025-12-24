// Package exchange Bitget WebSocket行情服务
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
	BitgetWSPublicURL = "wss://ws.bitget.com/v2/ws/public"
)

// BitgetWebSocket Bitget WebSocket行情服务
type BitgetWebSocket struct {
	mu   sync.RWMutex
	conn *WebSocketConnection

	// 行情数据缓存
	tickers map[string]*Ticker  // symbol -> ticker
	klines  map[string][]*Kline // symbol:interval -> klines

	// 回调管理
	tickerCallbacks map[string][]func(*Ticker)
	klineCallbacks  map[string][]func([]*Kline)

	// 订阅管理
	subscribedTickers map[string]bool
	subscribedKlines  map[string]bool // symbol:interval

	// 状态
	running bool
	ctx     context.Context
	cancel  context.CancelFunc

	// 代理配置
	proxyDialer func(network, addr string) (net.Conn, error)
}

// NewBitgetWebSocket 创建Bitget WebSocket服务
func NewBitgetWebSocket() *BitgetWebSocket {
	return &BitgetWebSocket{
		tickers:           make(map[string]*Ticker),
		klines:            make(map[string][]*Kline),
		tickerCallbacks:   make(map[string][]func(*Ticker)),
		klineCallbacks:    make(map[string][]func([]*Kline)),
		subscribedTickers: make(map[string]bool),
		subscribedKlines:  make(map[string]bool),
	}
}

// SetProxyDialer 设置代理拨号器
func (b *BitgetWebSocket) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.proxyDialer = dialer
}

// Start 启动WebSocket服务
func (b *BitgetWebSocket) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = true
	b.ctx, b.cancel = context.WithCancel(ctx)
	proxyDialer := b.proxyDialer // 复制代理配置
	b.mu.Unlock()

	// 创建WebSocket连接
	config := DefaultWebSocketConfig()
	config.URL = BitgetWSPublicURL
	config.PingInterval = 25 * time.Second // Bitget要求30秒内发送心跳
	config.ProxyDialer = proxyDialer       // 设置代理

	b.conn = NewWebSocketConnection(config)
	b.conn.SetCallbacks(b.onMessage, b.onConnected, b.onDisconnected)

	// 连接
	if err := b.conn.Connect(b.ctx); err != nil {
		b.mu.Lock()
		b.running = false
		b.mu.Unlock()
		return err
	}

	g.Log().Info(ctx, "[BitgetWS] WebSocket服务已启动")
	return nil
}

// Stop 停止WebSocket服务
func (b *BitgetWebSocket) Stop() {
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

	g.Log().Info(context.Background(), "[BitgetWS] WebSocket服务已停止")
}

// IsRunning 检查是否运行中
func (b *BitgetWebSocket) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running && b.conn != nil && b.conn.IsConnected()
}

// SubscribeTicker 订阅Ticker行情
func (b *BitgetWebSocket) SubscribeTicker(symbol string, callback func(*Ticker)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 格式化symbol
	wsSymbol := b.formatSymbol(symbol)

	// 保存回调
	b.tickerCallbacks[symbol] = append(b.tickerCallbacks[symbol], callback)

	// 如果已订阅，直接返回
	if b.subscribedTickers[wsSymbol] {
		return nil
	}

	// 发送订阅请求
	sub := map[string]interface{}{
		"op": "subscribe",
		"args": []map[string]string{
			{
				"instType": "USDT-FUTURES",
				"channel":  "ticker",
				"instId":   wsSymbol,
			},
		},
	}

	if err := b.conn.Send(sub); err != nil {
		return err
	}

	b.subscribedTickers[wsSymbol] = true
	b.conn.SaveSubscription("ticker:"+wsSymbol, sub)

	g.Log().Infof(b.ctx, "[BitgetWS] 订阅Ticker: %s", wsSymbol)
	return nil
}

// UnsubscribeTicker 取消订阅Ticker
func (b *BitgetWebSocket) UnsubscribeTicker(symbol string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	wsSymbol := b.formatSymbol(symbol)

	// 移除回调
	delete(b.tickerCallbacks, symbol)

	if !b.subscribedTickers[wsSymbol] {
		return nil
	}

	// 发送取消订阅请求
	unsub := map[string]interface{}{
		"op": "unsubscribe",
		"args": []map[string]string{
			{
				"instType": "USDT-FUTURES",
				"channel":  "ticker",
				"instId":   wsSymbol,
			},
		},
	}

	if err := b.conn.Send(unsub); err != nil {
		return err
	}

	delete(b.subscribedTickers, wsSymbol)
	b.conn.RemoveSubscription("ticker:" + wsSymbol)

	g.Log().Infof(b.ctx, "[BitgetWS] 取消订阅Ticker: %s", wsSymbol)
	return nil
}

// SubscribeKline 订阅K线
func (b *BitgetWebSocket) SubscribeKline(symbol, interval string, callback func([]*Kline)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	wsSymbol := b.formatSymbol(symbol)
	wsInterval := b.formatInterval(interval)
	key := wsSymbol + ":" + wsInterval

	// 保存回调
	cbKey := symbol + ":" + interval
	if callback != nil {
		b.klineCallbacks[cbKey] = append(b.klineCallbacks[cbKey], callback)
	}

	if b.subscribedKlines[key] {
		return nil
	}

	// 发送订阅请求
	sub := map[string]interface{}{
		"op": "subscribe",
		"args": []map[string]string{
			{
				"instType": "USDT-FUTURES",
				"channel":  "candle" + wsInterval,
				"instId":   wsSymbol,
			},
		},
	}

	if err := b.conn.Send(sub); err != nil {
		return err
	}

	b.subscribedKlines[key] = true
	b.conn.SaveSubscription("kline:"+key, sub)

	g.Log().Infof(b.ctx, "[BitgetWS] 订阅K线: %s %s", wsSymbol, wsInterval)
	return nil
}

// GetTicker 获取最新Ticker（从缓存）
func (b *BitgetWebSocket) GetTicker(symbol string) *Ticker {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.tickers[symbol]
}

// GetKlines 获取K线数据（从缓存）
func (b *BitgetWebSocket) GetKlines(symbol, interval string) []*Kline {
	b.mu.RLock()
	defer b.mu.RUnlock()
	key := symbol + ":" + interval
	return b.klines[key]
}

// onMessage 处理接收到的消息
func (b *BitgetWebSocket) onMessage(msg []byte) {
	// 解析消息
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		return
	}

	// 处理pong响应
	if event, ok := data["event"].(string); ok {
		if event == "pong" {
			return
		}
		if event == "subscribe" {
			g.Log().Debugf(b.ctx, "[BitgetWS] 订阅确认: %v", data)
			return
		}
		if event == "error" {
			g.Log().Warningf(b.ctx, "[BitgetWS] 错误响应: %v", data)
			return
		}
	}

	// 处理行情数据
	if action, ok := data["action"].(string); ok && action == "snapshot" || action == "update" {
		if arg, ok := data["arg"].(map[string]interface{}); ok {
			channel := arg["channel"].(string)
			instId := arg["instId"].(string)

			if channel == "ticker" {
				b.handleTickerData(instId, data["data"])
			} else if strings.HasPrefix(channel, "candle") {
				interval := strings.TrimPrefix(channel, "candle")
				b.handleKlineData(instId, interval, data["data"])
			}
		}
	}
}

// handleTickerData 处理Ticker数据
func (b *BitgetWebSocket) handleTickerData(instId string, data interface{}) {
	dataArr, ok := data.([]interface{})
	if !ok || len(dataArr) == 0 {
		return
	}

	tickerData, ok := dataArr[0].(map[string]interface{})
	if !ok {
		return
	}

	// 解析Ticker
	ticker := &Ticker{
		Symbol:             b.parseSymbol(instId),
		LastPrice:          parseFloat(tickerData["lastPr"]),
		BidPrice:           parseFloat(tickerData["bidPr"]),
		AskPrice:           parseFloat(tickerData["askPr"]),
		High24h:            parseFloat(tickerData["high24h"]),
		Low24h:             parseFloat(tickerData["low24h"]),
		Volume24h:          parseFloat(tickerData["baseVolume"]),
		Change24h:          parseFloat(tickerData["change24h"]) * 100,
		PriceChangePercent: parseFloat(tickerData["change24h"]) * 100,
		Timestamp:          parseInt(tickerData["ts"]),
	}

	symbol := ticker.Symbol

	// 更新缓存
	b.mu.Lock()
	b.tickers[symbol] = ticker
	callbacks := b.tickerCallbacks[symbol]
	b.mu.Unlock()

	// 触发回调
	for _, cb := range callbacks {
		if cb != nil {
			go cb(ticker)
		}
	}
}

// handleKlineData 处理K线数据
func (b *BitgetWebSocket) handleKlineData(instId, interval string, data interface{}) {
	dataArr, ok := data.([]interface{})
	if !ok || len(dataArr) == 0 {
		return
	}

	symbol := b.parseSymbol(instId)
	parsedInterval := b.parseInterval(interval)
	key := symbol + ":" + parsedInterval

	// 解析K线
	var klines []*Kline
	for _, item := range dataArr {
		klineArr, ok := item.([]interface{})
		if !ok || len(klineArr) < 6 {
			continue
		}

		kline := &Kline{
			OpenTime:  parseInt(klineArr[0]),
			Open:      parseFloat(klineArr[1]),
			High:      parseFloat(klineArr[2]),
			Low:       parseFloat(klineArr[3]),
			Close:     parseFloat(klineArr[4]),
			Volume:    parseFloat(klineArr[5]),
			CloseTime: parseInt(klineArr[0]) + b.getIntervalMs(parsedInterval),
		}
		klines = append(klines, kline)
	}

	if len(klines) == 0 {
		return
	}

	// 更新缓存（增量更新最新K线）
	b.mu.Lock()
	existing := b.klines[key]
	if len(existing) > 0 && len(klines) == 1 {
		// 更新最后一根K线
		lastKline := klines[0]
		if existing[len(existing)-1].OpenTime == lastKline.OpenTime {
			existing[len(existing)-1] = lastKline
		} else {
			existing = append(existing, lastKline)
			// 保持最多500根K线
			if len(existing) > 500 {
				existing = existing[len(existing)-500:]
			}
		}
		b.klines[key] = existing
	} else {
		b.klines[key] = klines
	}
	callbacks := b.klineCallbacks[key]
	klinesCopy := b.klines[key]
	b.mu.Unlock()

	// 触发回调
	for _, cb := range callbacks {
		if cb != nil {
			go cb(klinesCopy)
		}
	}
}

// onConnected 连接成功回调
func (b *BitgetWebSocket) onConnected() {
	g.Log().Info(b.ctx, "[BitgetWS] 连接成功，恢复订阅...")

	// 恢复所有订阅
	subs := b.conn.GetSubscriptions()
	for _, sub := range subs {
		if err := b.conn.Send(sub); err != nil {
			g.Log().Warningf(b.ctx, "[BitgetWS] 恢复订阅失败: %v", err)
		}
	}
}

// onDisconnected 断开连接回调
func (b *BitgetWebSocket) onDisconnected(err error) {
	g.Log().Warningf(b.ctx, "[BitgetWS] 连接断开: %v", err)
}

// formatSymbol 格式化交易对（用于WebSocket订阅）
func (b *BitgetWebSocket) formatSymbol(symbol string) string {
	// 使用统一的Symbol格式化器
	return Formatter.FormatForBitget(symbol) // BTCUSDT
}

// parseSymbol 解析交易对（从WebSocket响应）
func (b *BitgetWebSocket) parseSymbol(instId string) string {
	// BTCUSDT -> BTCUSDT
	return instId
}

// formatInterval 格式化K线周期
func (b *BitgetWebSocket) formatInterval(interval string) string {
	mapping := map[string]string{
		"1m":  "1m",
		"5m":  "5m",
		"15m": "15m",
		"30m": "30m",
		"1h":  "1H",
		"4h":  "4H",
		"1d":  "1D",
	}
	if v, ok := mapping[interval]; ok {
		return v
	}
	return "5m"
}

// parseInterval 解析K线周期
func (b *BitgetWebSocket) parseInterval(interval string) string {
	mapping := map[string]string{
		"1m":  "1m",
		"5m":  "5m",
		"15m": "15m",
		"30m": "30m",
		"1H":  "1h",
		"4H":  "4h",
		"1D":  "1d",
	}
	if v, ok := mapping[interval]; ok {
		return v
	}
	return interval
}

// getIntervalMs 获取K线周期毫秒数
func (b *BitgetWebSocket) getIntervalMs(interval string) int64 {
	mapping := map[string]int64{
		"1m":  60000,
		"5m":  300000,
		"15m": 900000,
		"30m": 1800000,
		"1h":  3600000,
		"4h":  14400000,
		"1d":  86400000,
	}
	if v, ok := mapping[interval]; ok {
		return v
	}
	return 300000
}

// 辅助函数
func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		return g.NewVar(val).Float64()
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func parseInt(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case string:
		return g.NewVar(val).Int64()
	case int64:
		return val
	default:
		return 0
	}
}

// GetSubscriptionCount 获取订阅数量
func (b *BitgetWebSocket) GetSubscriptionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribedTickers) + len(b.subscribedKlines)
}

// GetAllTickers 获取所有Ticker缓存
func (b *BitgetWebSocket) GetAllTickers() map[string]*Ticker {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make(map[string]*Ticker, len(b.tickers))
	for k, v := range b.tickers {
		result[k] = v
	}
	return result
}

// SendPing 发送心跳
func (b *BitgetWebSocket) SendPing() error {
	return b.conn.Send("ping")
}

// GetConnectionState 获取连接状态
func (b *BitgetWebSocket) GetConnectionState() string {
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
	bitgetWSInstance     *BitgetWebSocket
	bitgetWSInstanceOnce sync.Once
)

// GetBitgetWebSocket 获取Bitget WebSocket单例
func GetBitgetWebSocket() *BitgetWebSocket {
	bitgetWSInstanceOnce.Do(func() {
		bitgetWSInstance = NewBitgetWebSocket()
	})
	return bitgetWSInstance
}

// BitgetWSStatus Bitget WebSocket状态
type BitgetWSStatus struct {
	Running           bool              `json:"running"`
	ConnectionState   string            `json:"connectionState"`
	SubscriptionCount int               `json:"subscriptionCount"`
	TickerCount       int               `json:"tickerCount"`
	Tickers           map[string]string `json:"tickers"` // symbol -> lastPrice
}

// GetStatus 获取状态
func (b *BitgetWebSocket) GetStatus() *BitgetWSStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tickers := make(map[string]string, len(b.tickers))
	for k, v := range b.tickers {
		tickers[k] = fmt.Sprintf("%.4f", v.LastPrice)
	}

	return &BitgetWSStatus{
		Running:           b.running,
		ConnectionState:   b.GetConnectionState(),
		SubscriptionCount: len(b.subscribedTickers) + len(b.subscribedKlines),
		TickerCount:       len(b.tickers),
		Tickers:           tickers,
	}
}
