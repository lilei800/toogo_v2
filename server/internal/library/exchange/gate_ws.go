// Package exchange Gate.io WebSocket行情服务（公共行情）
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

// Gate USDT-Futures WebSocket (v4)
// 文档口径通常为 futures usdt：
// - URL: wss://fx-ws.gateio.ws/v4/ws/usdt
// - ticker channel: futures.tickers
// - candle channel: futures.candlesticks
const (
	GateWSFuturesUSDTURL = "wss://fx-ws.gateio.ws/v4/ws/usdt"
)

// GateWebSocket Gate WebSocket行情服务（tickers + candlesticks）
// 注意：Gate WS 字段可能随版本略有差异；这里以“能稳定拿到 last/bid/ask/24h high/low/volume”为主。
type GateWebSocket struct {
	mu   sync.RWMutex
	conn *WebSocketConnection

	// 行情数据缓存
	tickers map[string]*Ticker
	klines  map[string][]*Kline // symbol:interval -> klines

	// 回调管理
	tickerCallbacks map[string][]func(*Ticker)
	klineCallbacks  map[string][]func([]*Kline)

	// 订阅管理
	subscribed map[string]bool // key: streamKey

	// 状态
	running bool
	ctx     context.Context
	cancel  context.CancelFunc

	// 代理配置
	proxyDialer func(network, addr string) (net.Conn, error)
}

func NewGateWebSocket() *GateWebSocket {
	return &GateWebSocket{
		tickers:         make(map[string]*Ticker),
		klines:          make(map[string][]*Kline),
		tickerCallbacks: make(map[string][]func(*Ticker)),
		klineCallbacks:  make(map[string][]func([]*Kline)),
		subscribed:      make(map[string]bool),
	}
}

func (gt *GateWebSocket) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	gt.proxyDialer = dialer
}

func (gt *GateWebSocket) Start(ctx context.Context) error {
	gt.mu.Lock()
	if gt.running {
		gt.mu.Unlock()
		return nil
	}
	gt.running = true
	gt.ctx, gt.cancel = context.WithCancel(ctx)
	proxyDialer := gt.proxyDialer
	gt.mu.Unlock()

	cfg := DefaultWebSocketConfig()
	cfg.URL = GateWSFuturesUSDTURL
	cfg.PingInterval = 20 * time.Second
	cfg.ProxyDialer = proxyDialer

	gt.conn = NewWebSocketConnection(cfg)
	gt.conn.SetCallbacks(gt.onMessage, gt.onConnected, gt.onDisconnected)

	if err := gt.conn.Connect(gt.ctx); err != nil {
		gt.mu.Lock()
		gt.running = false
		gt.mu.Unlock()
		return err
	}

	g.Log().Info(ctx, "[GateWS] WebSocket服务已启动")
	return nil
}

func (gt *GateWebSocket) Stop() {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	if !gt.running {
		return
	}
	gt.running = false
	if gt.cancel != nil {
		gt.cancel()
	}
	if gt.conn != nil {
		gt.conn.Disconnect()
	}
	g.Log().Info(context.Background(), "[GateWS] WebSocket服务已停止")
}

func (gt *GateWebSocket) IsRunning() bool {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.running && gt.conn != nil && gt.conn.IsConnected()
}

// SubscribeTicker futures.tickers
func (gt *GateWebSocket) SubscribeTicker(symbol string, callback func(*Ticker)) error {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	contract := gateFormatContract(symbol)          // BTC_USDT
	normalizedSymbol := gateNormalizeSymbol(symbol) // BTCUSDT
	gt.tickerCallbacks[normalizedSymbol] = append(gt.tickerCallbacks[normalizedSymbol], callback)

	streamKey := "ticker:" + contract
	if gt.subscribed[streamKey] {
		return nil
	}

	// Gate WS subscribe format
	sub := map[string]interface{}{
		"time":    time.Now().Unix(),
		"channel": "futures.tickers",
		"event":   "subscribe",
		"payload": []string{contract},
	}

	if err := gt.conn.Send(sub); err != nil {
		return err
	}

	gt.subscribed[streamKey] = true
	gt.conn.SaveSubscription(streamKey, sub)
	g.Log().Infof(gt.ctx, "[GateWS] 订阅Ticker: %s (%s)", contract, normalizedSymbol)
	return nil
}

// UnsubscribeTicker 取消订阅Ticker
func (gt *GateWebSocket) UnsubscribeTicker(symbol string) error {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	contract := gateFormatContract(symbol)
	streamKey := "ticker:" + contract
	delete(gt.tickerCallbacks, gateNormalizeSymbol(symbol))
	if !gt.subscribed[streamKey] {
		return nil
	}

	unsub := map[string]interface{}{
		"time":    time.Now().Unix(),
		"channel": "futures.tickers",
		"event":   "unsubscribe",
		"payload": []string{contract},
	}
	if gt.conn != nil {
		_ = gt.conn.Send(unsub)
	}
	delete(gt.subscribed, streamKey)
	if gt.conn != nil {
		gt.conn.RemoveSubscription(streamKey)
	}
	return nil
}

// SubscribeKline futures.candlesticks
// interval: 1m/5m/15m/30m/1h/4h/1d
func (gt *GateWebSocket) SubscribeKline(symbol, interval string, callback func([]*Kline)) error {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	contract := gateFormatContract(symbol)
	normalizedSymbol := gateNormalizeSymbol(symbol)
	gateInterval := gateFormatInterval(interval)

	cbKey := normalizedSymbol + ":" + interval
	if callback != nil {
		gt.klineCallbacks[cbKey] = append(gt.klineCallbacks[cbKey], callback)
	}

	streamKey := "kline:" + contract + ":" + gateInterval
	if gt.subscribed[streamKey] {
		return nil
	}

	// payload: [contract, interval]
	sub := map[string]interface{}{
		"time":    time.Now().Unix(),
		"channel": "futures.candlesticks",
		"event":   "subscribe",
		"payload": []string{contract, gateInterval},
	}
	if err := gt.conn.Send(sub); err != nil {
		return err
	}
	gt.subscribed[streamKey] = true
	gt.conn.SaveSubscription(streamKey, sub)
	g.Log().Infof(gt.ctx, "[GateWS] 订阅K线: %s %s", contract, gateInterval)
	return nil
}

// UnsubscribeKline 取消订阅K线
func (gt *GateWebSocket) UnsubscribeKline(symbol, interval string) error {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	contract := gateFormatContract(symbol)
	gateInterval := gateFormatInterval(interval)
	streamKey := "kline:" + contract + ":" + gateInterval

	delete(gt.klineCallbacks, gateNormalizeSymbol(symbol)+":"+interval)
	if !gt.subscribed[streamKey] {
		return nil
	}

	unsub := map[string]interface{}{
		"time":    time.Now().Unix(),
		"channel": "futures.candlesticks",
		"event":   "unsubscribe",
		"payload": []string{contract, gateInterval},
	}
	if gt.conn != nil {
		_ = gt.conn.Send(unsub)
	}
	delete(gt.subscribed, streamKey)
	if gt.conn != nil {
		gt.conn.RemoveSubscription(streamKey)
	}
	return nil
}

func (gt *GateWebSocket) GetTicker(symbol string) *Ticker {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.tickers[gateNormalizeSymbol(symbol)]
}

func (gt *GateWebSocket) GetKlines(symbol, interval string) []*Kline {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	key := gateNormalizeSymbol(symbol) + ":" + interval
	return gt.klines[key]
}

// ============ 消息处理 ============

func (gt *GateWebSocket) onConnected() {
	g.Log().Info(gt.ctx, "[GateWS] 连接成功，恢复订阅...")
	if gt.conn == nil {
		return
	}
	subs := gt.conn.GetSubscriptions()
	for _, sub := range subs {
		_ = gt.conn.Send(sub)
	}
}

func (gt *GateWebSocket) onDisconnected(err error) {
	g.Log().Warningf(gt.ctx, "[GateWS] 连接断开: %v", err)
}

func (gt *GateWebSocket) onMessage(msg []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		return
	}

	// subscribe ack/error
	if ev, ok := data["event"].(string); ok {
		if ev == "subscribe" {
			return
		}
		if ev == "error" {
			g.Log().Warningf(gt.ctx, "[GateWS] error msg: %s", string(msg))
			return
		}
	}

	channel, _ := data["channel"].(string)
	if channel == "" {
		return
	}

	switch channel {
	case "futures.tickers":
		gt.handleTicker(data["result"])
	case "futures.candlesticks":
		gt.handleKline(data["result"])
	}
}

func (gt *GateWebSocket) handleTicker(result interface{}) {
	// result 通常为 map: { contract, last, highest_bid, lowest_ask, high_24h, low_24h, volume_24h, time }
	m, ok := result.(map[string]interface{})
	if !ok {
		return
	}
	contract, _ := m["contract"].(string)
	if contract == "" {
		return
	}
	symbol := gateNormalizeSymbol(contract)

	gt.mu.Lock()
	t := gt.tickers[symbol]
	if t == nil {
		t = &Ticker{Symbol: symbol}
		gt.tickers[symbol] = t
	}
	t.LastPrice = parseFloatAny(m["last"])
	t.BidPrice = parseFloatAny(m["highest_bid"])
	t.AskPrice = parseFloatAny(m["lowest_ask"])
	t.High24h = parseFloatAny(m["high_24h"])
	t.Low24h = parseFloatAny(m["low_24h"])
	t.Volume24h = parseFloatAny(m["volume_24h"])
	// Gate futures tickers 通常可能包含 mark_price（如果有则填充）
	if mp := parseFloatAny(m["mark_price"]); mp > 0 {
		t.MarkPrice = mp
	} else if mp2 := parseFloatAny(m["markPrice"]); mp2 > 0 {
		t.MarkPrice = mp2
	}
	t.Timestamp = time.Now().UnixMilli()
	cbs := append([]func(*Ticker){}, gt.tickerCallbacks[symbol]...)
	gt.mu.Unlock()

	for _, cb := range cbs {
		if cb != nil {
			go cb(t)
		}
	}
}

func (gt *GateWebSocket) handleKline(result interface{}) {
	// Gate futures.candlesticks 的 result 在常见实现里为 map：
	// {
	//   "t":1700000000, "v":"123", "c":"...", "h":"...", "l":"...", "o":"...",
	//   "contract":"BTC_USDT", "interval":"1m"
	// }
	if m, ok := result.(map[string]interface{}); ok {
		contract, _ := m["contract"].(string)
		interval, _ := m["interval"].(string)
		if contract == "" || interval == "" {
			return
		}

		symbol := gateNormalizeSymbol(contract)
		interval = strings.ToLower(strings.TrimSpace(interval))
		cbKey := symbol + ":" + interval

		// t 多为秒，统一成毫秒
		ts := parseIntAny(m["t"])
		if ts > 0 && ts < 1e12 {
			ts = ts * 1000
		}

		k := &Kline{
			OpenTime:  ts,
			Open:      parseFloatAny(m["o"]),
			High:      parseFloatAny(m["h"]),
			Low:       parseFloatAny(m["l"]),
			Close:     parseFloatAny(m["c"]),
			Volume:    parseFloatAny(m["v"]),
			CloseTime: ts,
		}

		gt.mu.Lock()
		existing := gt.klines[cbKey]
		if len(existing) > 0 {
			if existing[len(existing)-1].OpenTime == k.OpenTime {
				existing[len(existing)-1] = k
			} else {
				existing = append(existing, k)
				if len(existing) > 500 {
					existing = existing[len(existing)-500:]
				}
			}
		} else {
			existing = []*Kline{k}
		}
		gt.klines[cbKey] = existing

		cbs := append([]func([]*Kline){}, gt.klineCallbacks[cbKey]...)
		copyK := append([]*Kline(nil), existing...)
		gt.mu.Unlock()

		for _, cb := range cbs {
			if cb != nil {
				go cb(copyK)
			}
		}
		return
	}

	// 兼容数组格式：
	// 文档/不同实现里可能返回：
	// 1) [t, v, c, h, l, o, contract, interval]（单根+元信息）
	// 2) [[t,v,c,h,l,o], contract, interval]（一组K线+元信息）
	if arr, ok := result.([]interface{}); ok {
		// case2: [[...], contract, interval]
		if len(arr) == 3 {
			contract, _ := arr[1].(string)
			interval, _ := arr[2].(string)
			if contract == "" || interval == "" {
				return
			}
			symbol := gateNormalizeSymbol(contract)
			interval = strings.ToLower(strings.TrimSpace(interval))
			cbKey := symbol + ":" + interval

			rows, ok := arr[0].([]interface{})
			if !ok || len(rows) == 0 {
				return
			}
			klines := make([]*Kline, 0, len(rows))
			for _, it := range rows {
				row, ok := it.([]interface{})
				if !ok || len(row) < 6 {
					continue
				}
				ts := parseIntAny(row[0])
				if ts > 0 && ts < 1e12 {
					ts = ts * 1000
				}
				// Gate 常见: [t, v, c, h, l, o]
				k := &Kline{
					OpenTime:  ts,
					Open:      parseFloatAny(row[5]),
					High:      parseFloatAny(row[3]),
					Low:       parseFloatAny(row[4]),
					Close:     parseFloatAny(row[2]),
					Volume:    parseFloatAny(row[1]),
					CloseTime: ts,
				}
				klines = append(klines, k)
			}
			if len(klines) == 0 {
				return
			}

			gt.mu.Lock()
			existing := gt.klines[cbKey]
			// 增量合并：只更新/追加最后一根
			if len(klines) == 1 && len(existing) > 0 {
				last := klines[0]
				if existing[len(existing)-1].OpenTime == last.OpenTime {
					existing[len(existing)-1] = last
				} else {
					existing = append(existing, last)
				}
				if len(existing) > 500 {
					existing = existing[len(existing)-500:]
				}
				gt.klines[cbKey] = existing
			} else {
				// 如果一次推多根，直接覆盖（WS返回通常已是最新快照）
				if len(klines) > 500 {
					klines = klines[len(klines)-500:]
				}
				gt.klines[cbKey] = klines
				existing = klines
			}

			cbs := append([]func([]*Kline){}, gt.klineCallbacks[cbKey]...)
			copyK := append([]*Kline(nil), existing...)
			gt.mu.Unlock()

			for _, cb := range cbs {
				if cb != nil {
					go cb(copyK)
				}
			}
			return
		}

		// case1: [t, v, c, h, l, o, contract, interval]
		if len(arr) >= 8 {
			contract, _ := arr[6].(string)
			interval, _ := arr[7].(string)
			if contract == "" || interval == "" {
				return
			}
			symbol := gateNormalizeSymbol(contract)
			interval = strings.ToLower(strings.TrimSpace(interval))
			cbKey := symbol + ":" + interval

			ts := parseIntAny(arr[0])
			if ts > 0 && ts < 1e12 {
				ts = ts * 1000
			}
			k := &Kline{
				OpenTime:  ts,
				Open:      parseFloatAny(arr[5]),
				High:      parseFloatAny(arr[3]),
				Low:       parseFloatAny(arr[4]),
				Close:     parseFloatAny(arr[2]),
				Volume:    parseFloatAny(arr[1]),
				CloseTime: ts,
			}

			gt.mu.Lock()
			existing := gt.klines[cbKey]
			if len(existing) > 0 && existing[len(existing)-1].OpenTime == k.OpenTime {
				existing[len(existing)-1] = k
			} else {
				existing = append(existing, k)
				if len(existing) > 500 {
					existing = existing[len(existing)-500:]
				}
			}
			gt.klines[cbKey] = existing
			cbs := append([]func([]*Kline){}, gt.klineCallbacks[cbKey]...)
			copyK := append([]*Kline(nil), existing...)
			gt.mu.Unlock()

			for _, cb := range cbs {
				if cb != nil {
					go cb(copyK)
				}
			}
			return
		}
	}
}

// ============ 工具函数 ============

func gateFormatContract(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	// 支持 BTCUSDT / BTC/USDT / BTC-USDT / BTC_USDT
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	if strings.HasSuffix(s, "USDT") {
		base := strings.TrimSuffix(s, "USDT")
		return base + "_USDT"
	}
	// 兜底：如果用户直接传了 BTC_USDT
	if strings.Contains(symbol, "_") {
		return strings.ToUpper(symbol)
	}
	return s
}

func gateNormalizeSymbol(symbol string) string {
	// 使用统一的Symbol格式化器
	return Formatter.NormalizeSymbol(symbol) // BTCUSDT
}

func gateFormatInterval(interval string) string {
	// Gate WS futures candlesticks interval: "1m","5m","15m","30m","1h","4h","1d"
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
		return "1h"
	case "2h":
		return "2h"
	case "4h":
		return "4h"
	case "6h":
		return "6h"
	case "12h":
		return "12h"
	case "1d":
		return "1d"
	default:
		return "1m"
	}
}

// ============ 状态 ============

var (
	gateWSInstance     *GateWebSocket
	gateWSInstanceOnce sync.Once
)

func GetGateWebSocket() *GateWebSocket {
	gateWSInstanceOnce.Do(func() {
		gateWSInstance = NewGateWebSocket()
	})
	return gateWSInstance
}

type GateWSStatus struct {
	Running           bool              `json:"running"`
	ConnectionState   string            `json:"connectionState"`
	SubscriptionCount int               `json:"subscriptionCount"`
	TickerCount       int               `json:"tickerCount"`
	Tickers           map[string]string `json:"tickers"`
}

func (gt *GateWebSocket) GetConnectionState() string {
	if gt.conn == nil {
		return "disconnected"
	}
	switch gt.conn.GetState() {
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

func (gt *GateWebSocket) GetStatus() *GateWSStatus {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	tickers := make(map[string]string, len(gt.tickers))
	for k, v := range gt.tickers {
		tickers[k] = fmt.Sprintf("%.4f", v.LastPrice)
	}
	return &GateWSStatus{
		Running:           gt.running,
		ConnectionState:   gt.GetConnectionState(),
		SubscriptionCount: len(gt.subscribed),
		TickerCount:       len(gt.tickers),
		Tickers:           tickers,
	}
}
