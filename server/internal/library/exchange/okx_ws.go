// Package exchange OKX WebSocket行情服务（公共行情）
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
	OKXWSPublicURL = "wss://ws.okx.com:8443/ws/v5/public"
	// OKXWSBusinessURL: OKX v5 business 公共行情端点（部分频道会要求在 business 下订阅）
	OKXWSBusinessURL = "wss://ws.okx.com:8443/ws/v5/business"
)

// OKXWebSocket OKX WebSocket行情服务（tickers + candles）
// 说明：
// - 目前只实现 USDT 永续（SWAP）公共行情
// - symbol 输入接受：BTCUSDT / BTC/USDT / BTC-USDT / BTC_USDT / BTC-USDT-SWAP
type OKXWebSocket struct {
	mu   sync.RWMutex
	conn *WebSocketConnection
	// klineConn: K线专用连接（business）。tickers/mark-price 仍走 public。
	klineConn *WebSocketConnection

	// 行情数据缓存
	tickers map[string]*Ticker  // symbol -> ticker（symbol为标准化后的 BTCUSDT）
	klines  map[string][]*Kline // symbol:interval -> klines

	// 回调管理
	tickerCallbacks map[string][]func(*Ticker)
	klineCallbacks  map[string][]func([]*Kline)

	// 订阅管理
	subscribed map[string]bool // key: streamKey (ticker:instId / kline:instId:interval)

	// 状态
	running bool
	ctx     context.Context
	cancel  context.CancelFunc

	// 代理配置
	proxyDialer func(network, addr string) (net.Conn, error)
}

func NewOKXWebSocket() *OKXWebSocket {
	return &OKXWebSocket{
		tickers:         make(map[string]*Ticker),
		klines:          make(map[string][]*Kline),
		tickerCallbacks: make(map[string][]func(*Ticker)),
		klineCallbacks:  make(map[string][]func([]*Kline)),
		subscribed:      make(map[string]bool),
	}
}

func (o *OKXWebSocket) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.proxyDialer = dialer
}

func (o *OKXWebSocket) Start(ctx context.Context) error {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return nil
	}
	o.running = true
	o.ctx, o.cancel = context.WithCancel(ctx)
	proxyDialer := o.proxyDialer
	o.mu.Unlock()

	cfg := DefaultWebSocketConfig()
	// tickers/mark-price 在 public 端点
	cfg.URL = OKXWSPublicURL
	cfg.PingInterval = 25 * time.Second
	cfg.ProxyDialer = proxyDialer

	o.conn = NewWebSocketConnection(cfg)
	o.conn.SetCallbacks(o.onMessage, o.onConnected, o.onDisconnected)

	if err := o.conn.Connect(o.ctx); err != nil {
		o.mu.Lock()
		o.running = false
		o.mu.Unlock()
		return err
	}

	// K线 best-effort：使用 business 端点
	bizCfg := DefaultWebSocketConfig()
	bizCfg.URL = OKXWSBusinessURL
	bizCfg.PingInterval = 25 * time.Second
	bizCfg.ProxyDialer = proxyDialer
	o.klineConn = NewWebSocketConnection(bizCfg)
	o.klineConn.SetCallbacks(o.onMessage, o.onKlineConnected, o.onKlineDisconnected)
	if err := o.klineConn.Connect(o.ctx); err != nil {
		g.Log().Warningf(o.ctx, "[OKXWS] business(K线) 连接失败(降级REST): %v", err)
		// 不 return：public 仍然可用
		o.klineConn = nil
	}
	g.Log().Info(ctx, "[OKXWS] WebSocket服务已启动")
	return nil
}

func (o *OKXWebSocket) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.running {
		return
	}
	o.running = false
	if o.cancel != nil {
		o.cancel()
	}
	if o.conn != nil {
		o.conn.Disconnect()
	}
	if o.klineConn != nil {
		o.klineConn.Disconnect()
	}
	g.Log().Info(context.Background(), "[OKXWS] WebSocket服务已停止")
}

func (o *OKXWebSocket) IsRunning() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.running && o.conn != nil && o.conn.IsConnected()
}

// SubscribeTicker 订阅 ticker（OKX tickers channel）
func (o *OKXWebSocket) SubscribeTicker(symbol string, callback func(*Ticker)) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	instId := okxFormatInstId(symbol)
	normalizedSymbol := okxNormalizeSymbol(symbol)

	o.tickerCallbacks[normalizedSymbol] = append(o.tickerCallbacks[normalizedSymbol], callback)

	streamKey := "ticker:" + instId
	if o.subscribed[streamKey] {
		return nil
	}

	sub := map[string]interface{}{
		"op": "subscribe",
		"args": []map[string]string{
			{
				"channel": "tickers",
				"instId":  instId,
			},
		},
	}

	if err := o.conn.Send(sub); err != nil {
		return err
	}
	o.subscribed[streamKey] = true
	o.conn.SaveSubscription(streamKey, sub)
	g.Log().Infof(o.ctx, "[OKXWS] 订阅Ticker: %s (%s)", instId, normalizedSymbol)

	// 同时订阅 mark-price（best-effort，用于统一风控/盈亏口径）
	_ = o.subscribeMarkPriceLocked(symbol)
	return nil
}

// SubscribeMarkPrice 订阅标记价格
func (o *OKXWebSocket) SubscribeMarkPrice(symbol string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.subscribeMarkPriceLocked(symbol)
}

func (o *OKXWebSocket) subscribeMarkPriceLocked(symbol string) error {
	instId := okxFormatInstId(symbol)
	streamKey := "mark:" + instId
	if o.subscribed[streamKey] {
		return nil
	}
	sub := map[string]interface{}{
		"op": "subscribe",
		"args": []map[string]string{
			{
				"channel": "mark-price",
				"instId":  instId,
			},
		},
	}
	if o.conn == nil {
		return nil
	}
	if err := o.conn.Send(sub); err != nil {
		return err
	}
	o.subscribed[streamKey] = true
	o.conn.SaveSubscription(streamKey, sub)
	return nil
}

// UnsubscribeMarkPrice 取消订阅标记价格
func (o *OKXWebSocket) UnsubscribeMarkPrice(symbol string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	instId := okxFormatInstId(symbol)
	streamKey := "mark:" + instId
	if !o.subscribed[streamKey] {
		return nil
	}
	unsub := map[string]interface{}{
		"op": "unsubscribe",
		"args": []map[string]string{
			{
				"channel": "mark-price",
				"instId":  instId,
			},
		},
	}
	if o.conn != nil {
		_ = o.conn.Send(unsub)
		o.conn.RemoveSubscription(streamKey)
	}
	delete(o.subscribed, streamKey)
	return nil
}

// UnsubscribeTicker 取消订阅 ticker
func (o *OKXWebSocket) UnsubscribeTicker(symbol string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	instId := okxFormatInstId(symbol)
	streamKey := "ticker:" + instId
	// 清理回调（按统一symbol）
	delete(o.tickerCallbacks, okxNormalizeSymbol(symbol))
	if !o.subscribed[streamKey] {
		return nil
	}

	unsub := map[string]interface{}{
		"op": "unsubscribe",
		"args": []map[string]string{
			{
				"channel": "tickers",
				"instId":  instId,
			},
		},
	}
	if o.conn != nil {
		_ = o.conn.Send(unsub)
	}
	delete(o.subscribed, streamKey)
	if o.conn != nil {
		o.conn.RemoveSubscription(streamKey)
	}
	return nil
}

// SubscribeKline 订阅K线（OKX candle channel）
// interval: 1m/5m/15m/30m/1h/4h/1d（映射为 OKX candle1m/candle5m/...）
func (o *OKXWebSocket) SubscribeKline(symbol, interval string, callback func([]*Kline)) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	instId := okxFormatInstId(symbol)
	normalizedSymbol := okxNormalizeSymbol(symbol)
	okxInterval := okxFormatInterval(interval)

	cbKey := normalizedSymbol + ":" + interval
	if callback != nil {
		o.klineCallbacks[cbKey] = append(o.klineCallbacks[cbKey], callback)
	}

	streamKey := "kline:" + instId + ":" + okxInterval
	if o.subscribed[streamKey] {
		return nil
	}

	sub := map[string]interface{}{
		"op": "subscribe",
		"args": []map[string]string{
			{
				"channel": "candle" + okxInterval,
				"instId":  instId,
			},
		},
	}
	// 优先走 business 连接订阅 K线
	if o.klineConn != nil && o.klineConn.IsConnected() {
		if err := o.klineConn.Send(sub); err != nil {
			return err
		}
		o.klineConn.SaveSubscription(streamKey, sub)
	} else {
		if o.conn == nil {
			return fmt.Errorf("OKXWS conn is nil")
		}
		if err := o.conn.Send(sub); err != nil {
			return err
		}
		o.conn.SaveSubscription(streamKey, sub)
	}
	o.subscribed[streamKey] = true
	g.Log().Infof(o.ctx, "[OKXWS] 订阅K线: %s candle%s", instId, okxInterval)
	return nil
}

// UnsubscribeKline 取消订阅K线
func (o *OKXWebSocket) UnsubscribeKline(symbol, interval string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	instId := okxFormatInstId(symbol)
	okxInterval := okxFormatInterval(interval)
	streamKey := "kline:" + instId + ":" + okxInterval

	// 清理回调（即使没订阅成功也不阻断）
	delete(o.klineCallbacks, okxNormalizeSymbol(symbol)+":"+interval)

	if !o.subscribed[streamKey] {
		return nil
	}

	unsub := map[string]interface{}{
		"op": "unsubscribe",
		"args": []map[string]string{
			{
				"channel": "candle" + okxInterval,
				"instId":  instId,
			},
		},
	}
	// best-effort：优先 business 退订
	if o.klineConn != nil {
		_ = o.klineConn.Send(unsub)
		o.klineConn.RemoveSubscription(streamKey)
	}
	if o.conn != nil {
		_ = o.conn.Send(unsub)
		o.conn.RemoveSubscription(streamKey)
	}
	delete(o.subscribed, streamKey)
	return nil
}

func (o *OKXWebSocket) GetTicker(symbol string) *Ticker {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.tickers[okxNormalizeSymbol(symbol)]
}

func (o *OKXWebSocket) GetKlines(symbol, interval string) []*Kline {
	o.mu.RLock()
	defer o.mu.RUnlock()
	key := okxNormalizeSymbol(symbol) + ":" + interval
	return o.klines[key]
}

// ============ 消息处理 ============

func (o *OKXWebSocket) onConnected() {
	g.Log().Info(o.ctx, "[OKXWS] 连接成功，恢复订阅...")
	if o.conn == nil {
		return
	}
	subs := o.conn.GetSubscriptions()
	for _, sub := range subs {
		_ = o.conn.Send(sub)
	}
}

func (o *OKXWebSocket) onDisconnected(err error) {
	g.Log().Warningf(o.ctx, "[OKXWS] 连接断开: %v", err)
}

func (o *OKXWebSocket) onKlineConnected() {
	g.Log().Info(o.ctx, "[OKXWS] business(K线) 连接成功，恢复订阅...")
	if o.klineConn == nil {
		return
	}
	subs := o.klineConn.GetSubscriptions()
	for _, sub := range subs {
		_ = o.klineConn.Send(sub)
	}
}

func (o *OKXWebSocket) onKlineDisconnected(err error) {
	g.Log().Warningf(o.ctx, "[OKXWS] business(K线) 连接断开: %v", err)
}
func (o *OKXWebSocket) onMessage(msg []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		return
	}

	// event: subscribe/unsubscribe/error
	if ev, ok := data["event"].(string); ok {
		if ev == "subscribe" || ev == "unsubscribe" {
			return
		}
		if ev == "error" {
			g.Log().Warningf(o.ctx, "[OKXWS] error msg: %s", string(msg))
			return
		}
	}

	arg, _ := data["arg"].(map[string]interface{})
	if arg == nil {
		return
	}
	channel, _ := arg["channel"].(string)
	instId, _ := arg["instId"].(string)
	if channel == "" || instId == "" {
		return
	}

	switch {
	case channel == "tickers":
		o.handleTicker(instId, data["data"])
	case channel == "mark-price":
		o.handleMarkPrice(instId, data["data"])
	case strings.HasPrefix(channel, "candles"):
		interval := strings.TrimPrefix(channel, "candles") // e.g. 1m/1H
		o.handleKline(instId, okxParseInterval(interval), data["data"])
	case strings.HasPrefix(channel, "candle"):
		// 兼容旧实现/异常情况：若服务端仍推送单数前缀，仍可解析
		interval := strings.TrimPrefix(channel, "candle") // e.g. 1m/1H
		o.handleKline(instId, okxParseInterval(interval), data["data"])
	}
}

func (o *OKXWebSocket) handleTicker(instId string, payload interface{}) {
	arr, ok := payload.([]interface{})
	if !ok || len(arr) == 0 {
		return
	}
	item, ok := arr[0].(map[string]interface{})
	if !ok {
		return
	}

	// OKX tickers fields: last, bidPx, askPx, high24h, low24h, vol24h, ts
	symbol := okxNormalizeSymbol(instId)

	o.mu.Lock()
	t := o.tickers[symbol]
	if t == nil {
		t = &Ticker{Symbol: symbol}
		o.tickers[symbol] = t
	}
	t.LastPrice = parseFloatAny(item["last"])
	t.BidPrice = parseFloatAny(item["bidPx"])
	t.AskPrice = parseFloatAny(item["askPx"])
	t.High24h = parseFloatAny(item["high24h"])
	t.Low24h = parseFloatAny(item["low24h"])
	t.Volume24h = parseFloatAny(item["vol24h"])
	// 24h涨跌幅（%）：OKX tickers 通常包含 open24h / sodUtc0
	open24h := parseFloatAny(item["open24h"])
	if open24h <= 0 {
		open24h = parseFloatAny(item["sodUtc0"])
	}
	if open24h > 0 && t.LastPrice > 0 {
		changePercent := (t.LastPrice-open24h)/open24h*100.0
		t.Change24h = changePercent
		t.PriceChangePercent = changePercent
	}
	t.Timestamp = parseIntAny(item["ts"])

	cbs := append([]func(*Ticker){}, o.tickerCallbacks[symbol]...)
	o.mu.Unlock()

	for _, cb := range cbs {
		if cb != nil {
			go cb(t)
		}
	}
}

func (o *OKXWebSocket) handleMarkPrice(instId string, payload interface{}) {
	arr, ok := payload.([]interface{})
	if !ok || len(arr) == 0 {
		return
	}
	item, ok := arr[0].(map[string]interface{})
	if !ok {
		return
	}

	// OKX mark-price fields: markPx, instId, ts
	symbol := okxNormalizeSymbol(instId)
	mark := parseFloatAny(item["markPx"])
	if mark <= 0 {
		return
	}

	o.mu.Lock()
	t := o.tickers[symbol]
	if t == nil {
		t = &Ticker{Symbol: symbol}
		o.tickers[symbol] = t
	}
	t.MarkPrice = mark
	if ts := parseIntAny(item["ts"]); ts > 0 {
		t.Timestamp = ts
	}
	cbs := append([]func(*Ticker){}, o.tickerCallbacks[symbol]...)
	o.mu.Unlock()

	for _, cb := range cbs {
		if cb != nil {
			go cb(t)
		}
	}
}

func (o *OKXWebSocket) handleKline(instId, interval string, payload interface{}) {
	arr, ok := payload.([]interface{})
	if !ok || len(arr) == 0 {
		return
	}

	normalizedSymbol := okxNormalizeSymbol(instId)
	key := normalizedSymbol + ":" + interval

	var klines []*Kline
	for _, it := range arr {
		row, ok := it.([]interface{})
		if !ok || len(row) < 6 {
			continue
		}
		// OKX candle: [ts, o, h, l, c, vol, ...]
		openTime := parseIntAny(row[0])
		k := &Kline{
			OpenTime:  openTime,
			Open:      parseFloatAny(row[1]),
			High:      parseFloatAny(row[2]),
			Low:       parseFloatAny(row[3]),
			Close:     parseFloatAny(row[4]),
			Volume:    parseFloatAny(row[5]),
			CloseTime: openTime,
		}
		klines = append(klines, k)
	}
	if len(klines) == 0 {
		return
	}

	o.mu.Lock()
	existing := o.klines[key]
	if len(existing) > 0 && len(klines) == 1 {
		last := klines[0]
		if existing[len(existing)-1].OpenTime == last.OpenTime {
			existing[len(existing)-1] = last
		} else {
			existing = append(existing, last)
			if len(existing) > 500 {
				existing = existing[len(existing)-500:]
			}
		}
		o.klines[key] = existing
	} else {
		o.klines[key] = klines
	}
	cbKey := normalizedSymbol + ":" + interval
	cbs := append([]func([]*Kline){}, o.klineCallbacks[cbKey]...)
	copyK := append([]*Kline(nil), o.klines[key]...)
	o.mu.Unlock()

	for _, cb := range cbs {
		if cb != nil {
			go cb(copyK)
		}
	}
}

// ============ 工具函数 ============

func okxFormatInstId(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if strings.Contains(s, "SWAP") && strings.Contains(s, "-") {
		return s
	}
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	if strings.HasSuffix(s, "USDT") {
		base := strings.TrimSuffix(s, "USDT")
		return base + "-USDT-SWAP"
	}
	return s
}

// okxNormalizeSymbol 统一输出为 BTCUSDT（便于与其他交易所缓存key一致）
func okxNormalizeSymbol(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	// instId -> BTCUSDT
	if strings.Contains(s, "-USDT") {
		s = strings.ReplaceAll(s, "-USDT-SWAP", "USDT")
		s = strings.ReplaceAll(s, "-USDT", "USDT")
		s = strings.ReplaceAll(s, "-", "")
	}
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func okxFormatInterval(interval string) string {
	switch strings.ToLower(interval) {
	case "1m":
		return "1m"
	case "3m":
		return "3m"
	case "5m":
		return "5m"
	case "15m":
		return "15m"
	case "30m":
		return "30m"
	case "1h", "60m":
		return "1H"
	case "2h":
		return "2H"
	case "4h":
		return "4H"
	case "6h":
		return "6H"
	case "12h":
		return "12H"
	case "1d":
		return "1D"
	default:
		return "1m"
	}
}

func okxParseInterval(okxInterval string) string {
	// OKX: 1H -> 1h etc
	switch okxInterval {
	case "1H":
		return "1h"
	case "2H":
		return "2h"
	case "4H":
		return "4h"
	case "6H":
		return "6h"
	case "12H":
		return "12h"
	case "1D":
		return "1d"
	default:
		return strings.ToLower(okxInterval)
	}
}

func parseFloatAny(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case string:
		return g.NewVar(x).Float64()
	case json.Number:
		f, _ := x.Float64()
		return f
	default:
		return 0
	}
}

func parseIntAny(v interface{}) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case string:
		return g.NewVar(x).Int64()
	case json.Number:
		i, _ := x.Int64()
		return i
	default:
		return 0
	}
}

// ============ 状态 ============

var (
	okxWSInstance     *OKXWebSocket
	okxWSInstanceOnce sync.Once
)

func GetOKXWebSocket() *OKXWebSocket {
	okxWSInstanceOnce.Do(func() {
		okxWSInstance = NewOKXWebSocket()
	})
	return okxWSInstance
}

type OKXWSStatus struct {
	Running           bool              `json:"running"`
	ConnectionState   string            `json:"connectionState"`
	SubscriptionCount int               `json:"subscriptionCount"`
	TickerCount       int               `json:"tickerCount"`
	Tickers           map[string]string `json:"tickers"`
}

func (o *OKXWebSocket) GetConnectionState() string {
	if o.conn == nil {
		return "disconnected"
	}
	switch o.conn.GetState() {
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

func (o *OKXWebSocket) GetStatus() *OKXWSStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	tickers := make(map[string]string, len(o.tickers))
	for k, v := range o.tickers {
		tickers[k] = fmt.Sprintf("%.4f", v.LastPrice)
	}
	return &OKXWSStatus{
		Running:           o.running,
		ConnectionState:   o.GetConnectionState(),
		SubscriptionCount: len(o.subscribed),
		TickerCount:       len(o.tickers),
		Tickers:           tickers,
	}
}
