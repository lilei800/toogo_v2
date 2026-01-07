// Package exchange Gate.io WebSocket行情服务（公共行情）
package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
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

	// 先保存订阅信息：即使当前未连接/发送失败，也能在重连后自动恢复（onConnected 会重放 subscriptions）
	gt.subscribed[streamKey] = true
	if gt.conn != nil {
		gt.conn.SaveSubscription(streamKey, sub)
	}

	// 尝试立即发送（若未连接则忽略，等待 onConnected 重放）
	if gt.conn != nil && gt.conn.IsConnected() {
		if err := gt.conn.Send(sub); err != nil {
			g.Log().Warningf(gt.ctx, "[GateWS] 订阅Ticker发送失败(将等待重连恢复): contract=%s, err=%v", contract, err)
		}
	}
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
		// 兼容：调用方可能传 60m/1h 等同义，但 Gate 推送的 interval 可能是规范化后的 gateInterval。
		// 为避免“收到了K线但回调key对不上”，这里同时按 gateInterval 再挂一份回调。
		cbKey2 := normalizedSymbol + ":" + gateInterval
		if cbKey2 != cbKey {
			gt.klineCallbacks[cbKey2] = append(gt.klineCallbacks[cbKey2], callback)
		}
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

	// 先保存订阅信息：即使当前未连接/发送失败，也能在重连后自动恢复（onConnected 会重放 subscriptions）
	gt.subscribed[streamKey] = true
	if gt.conn != nil {
		gt.conn.SaveSubscription(streamKey, sub)
	}

	// 尝试立即发送（若未连接则忽略，等待 onConnected 重放）
	if gt.conn != nil && gt.conn.IsConnected() {
		if err := gt.conn.Send(sub); err != nil {
			g.Log().Warningf(gt.ctx, "[GateWS] 订阅K线发送失败(将等待重连恢复): contract=%s, interval=%s, err=%v", contract, gateInterval, err)
		}
	}
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
	interval = gateNormalizeInterval(interval)
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
			// 订阅确认：用于排查“WS连上但K线没数据”是否因为订阅根本没成功/参数不匹配。
			// 这里提升到 WARN 方便线上直接看到（订阅阶段量很小：每个 symbol * (ticker + 多周期K线) 一次）。
			g.Log().Warningf(gt.ctx, "[GateWS] subscribe ack: channel=%v payload=%v", data["channel"], data["payload"])
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

	// 低频诊断：确认 candlesticks 更新消息是否真实到达（不依赖 result/data 的解析分支）
	if channel == "futures.candlesticks" {
		if gateShouldLog("candlesticks_msg", 30*time.Second) {
			ev, _ := data["event"].(string)
			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			g.Log().Warningf(gt.ctx, "[GateWS] recv candlesticks msg: event=%s keys=%v", ev, keys)
		}
	}

	switch channel {
	case "futures.tickers":
		// Gate 不同实现里可能用 result 或 data 承载 payload，这里做兼容
		if v, ok := data["result"]; ok && v != nil {
			gt.handleTicker(v)
		} else if v, ok := data["data"]; ok && v != nil {
			gt.handleTicker(v)
		}
	case "futures.candlesticks":
		// 关键修复：contract/interval 有时在“顶层字段”，K线本体在 result/data。
		// 如果只把 result/data 传下去，会丢失 meta，导致回调/缓存 key 不一致。
		gt.handleKline(data)
	}
}

func (gt *GateWebSocket) handleTicker(result interface{}) {
	applyOne := func(m map[string]interface{}) {
		// result 常见为 map: { contract, last, highest_bid, lowest_ask, high_24h, low_24h, volume_24h, time }
		// 但不同实现字段名可能有差异，这里做容错解析。
		contract, _ := m["contract"].(string)
		if contract == "" {
			// 兜底：部分实现可能使用 s / symbol 字段
			if s, ok := m["symbol"].(string); ok && s != "" {
				contract = s
			} else if s2, ok := m["s"].(string); ok && s2 != "" {
				contract = s2
			}
		}
		if contract == "" {
			return
		}
		symbol := gateNormalizeSymbol(contract)

		// 解析字段（多别名）
		last := parseFloatAny(m["last"])
		if last <= 0 {
			last = parseFloatAny(m["last_price"])
		}
		if last <= 0 {
			last = parseFloatAny(m["lastPrice"])
		}

		bid := parseFloatAny(m["highest_bid"])
		if bid <= 0 {
			bid = parseFloatAny(m["bid"])
		}
		if bid <= 0 {
			bid = parseFloatAny(m["bidPrice"])
		}

		ask := parseFloatAny(m["lowest_ask"])
		if ask <= 0 {
			ask = parseFloatAny(m["ask"])
		}
		if ask <= 0 {
			ask = parseFloatAny(m["askPrice"])
		}

		high24 := parseFloatAny(m["high_24h"])
		if high24 <= 0 {
			high24 = parseFloatAny(m["high24h"])
		}
		low24 := parseFloatAny(m["low_24h"])
		if low24 <= 0 {
			low24 = parseFloatAny(m["low24h"])
		}
		vol24 := parseFloatAny(m["volume_24h"])
		if vol24 <= 0 {
			vol24 = parseFloatAny(m["vol_24h"])
		}
		if vol24 <= 0 {
			vol24 = parseFloatAny(m["volume24h"])
		}

		mark := parseFloatAny(m["mark_price"])
		if mark <= 0 {
			mark = parseFloatAny(m["markPrice"])
		}

		changePct := parseFloatAny(m["change_percentage"])
		if changePct == 0 {
			changePct = parseFloatAny(m["changePercent"])
		}
		if changePct == 0 {
			changePct = parseFloatAny(m["change_pct"])
		}

		// time: 秒级时间戳（若有则使用）
		ts := parseIntAny(m["time"])
		if ts <= 0 {
			ts = parseIntAny(m["t"])
		}
		if ts > 0 && ts < 1e12 {
			ts = ts * 1000
		}
		if ts <= 0 {
			ts = time.Now().UnixMilli()
		}

		gt.mu.Lock()
		t := gt.tickers[symbol]
		if t == nil {
			t = &Ticker{Symbol: symbol}
			gt.tickers[symbol] = t
		}
		// 只要 last>0，就认为行情已就绪（连接判断只需要 lastPrice）
		if last > 0 {
			t.LastPrice = last
		}
		if bid > 0 {
			t.BidPrice = bid
		}
		if ask > 0 {
			t.AskPrice = ask
		}
		if high24 > 0 {
			t.High24h = high24
		}
		if low24 > 0 {
			t.Low24h = low24
		}
		if vol24 > 0 {
			t.Volume24h = vol24
		}
		if mark > 0 {
			t.MarkPrice = mark
		}
		if changePct != 0 {
			t.Change24h = changePct
			t.PriceChangePercent = changePct
		}
		t.Timestamp = ts

		cbs := append([]func(*Ticker){}, gt.tickerCallbacks[symbol]...)
		gt.mu.Unlock()

		for _, cb := range cbs {
			if cb != nil {
				go cb(t)
			}
		}
	}

	// 1) 单条 map
	if m, ok := result.(map[string]interface{}); ok {
		applyOne(m)
		return
	}
	// 2) 批量数组：[{...},{...}] 或 [[...], ...]
	if arr, ok := result.([]interface{}); ok {
		for _, it := range arr {
			if m, ok2 := it.(map[string]interface{}); ok2 {
				applyOne(m)
				continue
			}
			// 少数实现可能是二维数组/行格式，这里不处理（避免误解析）；需要时再扩展
		}
		return
	}
}

func (gt *GateWebSocket) handleKline(result interface{}) {
	// Gate futures.candlesticks 的 payload 形态不稳定：
	// - 有的实现：顶层带 contract/interval，K线本体在 result/data（通常是数组）
	// - 有的实现：result 本身就是 map（包含 t/o/h/l/c/v + contract/interval）
	// - 还有实现：result 为二维数组，顶层带 contract/interval
	//
	// 关键：内部存取 key 必须统一用 BTCUSDT + interval(通常 1m/5m/.../1h)。
	if m, ok := result.(map[string]interface{}); ok {
		contract, _ := m["contract"].(string)
		interval, _ := m["interval"].(string)
		contract = strings.TrimSpace(contract)
		interval = gateNormalizeInterval(interval)

		// 如果顶层没有 meta，尝试从 result(map) 中取
		if (contract == "" || interval == "") && m["result"] != nil {
			if rm, ok2 := m["result"].(map[string]interface{}); ok2 {
				if contract == "" {
					if s, ok3 := rm["contract"].(string); ok3 {
						contract = strings.TrimSpace(s)
					}
				}
				if interval == "" {
					if s, ok3 := rm["interval"].(string); ok3 {
						interval = gateNormalizeInterval(s)
					}
				}
			}
		}

		// 顶层只有 meta，但没有 OHLCV 字段：把 meta 注入到 result/data，再复用下面的数组解析逻辑
		if _, okT := m["t"]; !okT {
			var payload interface{}
			if v := m["result"]; v != nil {
				payload = v
			} else if v := m["data"]; v != nil {
				payload = v
			}
			if payload != nil && contract != "" && interval != "" {
				switch pv := payload.(type) {
				case []interface{}:
					// 1) 单根数组: [t,o,h,l,c,v] -> append meta
					if len(pv) > 0 {
						// 二维数组：[[...],[...]] -> 包成 case2 的结构
						if _, ok2 := pv[0].([]interface{}); ok2 {
							gt.handleKline([]interface{}{pv, contract, interval})
							return
						}
						gt.handleKline(append(append([]interface{}{}, pv...), contract, interval))
						return
					}
				case map[string]interface{}:
					// map payload：补齐 meta 后走 map 解析
					if pv["contract"] == nil {
						pv["contract"] = contract
					}
					if pv["interval"] == nil {
						pv["interval"] = interval
					}
					gt.handleKline(pv)
					return
				}
			}
		}

		if contract == "" || interval == "" {
			return
		}

		symbol := gateNormalizeSymbol(contract)
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

		// 低频诊断：确认是否真的收到 candlesticks，以及 interval/callback key 是否匹配
		if gateShouldLog("kline_"+cbKey, 30*time.Second) {
			g.Log().Warningf(gt.ctx, "[GateWS] kline update: cbKey=%s, interval=%s, klines=%d, callbacks=%d",
				cbKey, interval, len(copyK), len(cbs))
		}

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
		// 工具：从“某个数组”里构建 kline（兼容多种字段顺序）
		buildKlineFromRow := func(row []interface{}) *Kline {
			if len(row) < 6 {
				return nil
			}
			// 尝试解析出 6 个数字（t + 5个字段）
			t := parseIntAny(row[0])
			if t > 0 && t < 1e12 {
				t = t * 1000
			}
			// 可能的字段排列（Gate 不同实现差异较大）
			type cand struct {
				open, high, low, close, vol float64
			}
			cands := []cand{
				// [t, v, c, h, l, o]
				{open: parseFloatAny(row[5]), high: parseFloatAny(row[3]), low: parseFloatAny(row[4]), close: parseFloatAny(row[2]), vol: parseFloatAny(row[1])},
				// [t, o, c, h, l, v]
				{open: parseFloatAny(row[1]), high: parseFloatAny(row[3]), low: parseFloatAny(row[4]), close: parseFloatAny(row[2]), vol: parseFloatAny(row[5])},
				// [t, o, h, l, c, v]
				{open: parseFloatAny(row[1]), high: parseFloatAny(row[2]), low: parseFloatAny(row[3]), close: parseFloatAny(row[4]), vol: parseFloatAny(row[5])},
				// [t, c, h, l, o, v]
				{open: parseFloatAny(row[4]), high: parseFloatAny(row[2]), low: parseFloatAny(row[3]), close: parseFloatAny(row[1]), vol: parseFloatAny(row[5])},
			}
			choose := func(c cand) bool {
				if c.high <= 0 || c.low <= 0 || c.open <= 0 || c.close <= 0 {
					return false
				}
				if c.high < c.low {
					return false
				}
				// 高低价应包住开收
				if c.high < c.open || c.high < c.close {
					return false
				}
				if c.low > c.open || c.low > c.close {
					return false
				}
				return true
			}
			var picked *cand
			for i := range cands {
				if choose(cands[i]) {
					picked = &cands[i]
					break
				}
			}
			// 若都不满足，仍尽力按最常见格式落一份（避免“一直空”）
			if picked == nil {
				picked = &cands[0]
			}
			return &Kline{
				OpenTime:  t,
				Open:      picked.open,
				High:      picked.high,
				Low:       picked.low,
				Close:     picked.close,
				Volume:    picked.vol,
				CloseTime: t,
			}
		}

		// 工具：从数组里猜 contract 与 interval（某些实现并非固定位置）
		extractMeta := func(vs []interface{}) (contract string, interval string) {
			for _, it := range vs {
				s, ok := it.(string)
				if !ok {
					continue
				}
				ss := strings.TrimSpace(s)
				if ss == "" {
					continue
				}
				low := strings.ToLower(ss)
				// interval: 1m/5m/15m/30m/1h/4h/1d...
				if interval == "" {
					switch low {
					case "1m", "5m", "15m", "30m", "1h", "4h", "1d":
						interval = low
						continue
					}
				}
				// contract: BTC_USDT / ETH_USDT ...
				if contract == "" && strings.Contains(ss, "_") {
					contract = ss
					continue
				}
			}
			return
		}

		// case2: [[...], contract, interval]
		if len(arr) == 3 {
			contract, _ := arr[1].(string)
			interval, _ := arr[2].(string)
			contract = strings.TrimSpace(contract)
			interval = strings.ToLower(strings.TrimSpace(interval))
			if contract == "" || interval == "" {
				// 尝试兜底猜测
				c2, i2 := extractMeta(arr)
				if contract == "" {
					contract = c2
				}
				if interval == "" {
					interval = i2
				}
			}
			if contract == "" || interval == "" {
				return
			}
			symbol := gateNormalizeSymbol(contract)
			cbKey := symbol + ":" + interval

			// rows 可能是：
			// - []interface{}{ []interface{}{...}, []interface{}{...} }
			// - 或 [][]interface{}（在Go反序列化里仍会是 []interface{}）
			rows, ok := arr[0].([]interface{})
			if !ok || len(rows) == 0 {
				// 也可能直接是一根：arr[0] 就是 row
				if row, ok2 := arr[0].([]interface{}); ok2 {
					rows = []interface{}{row}
				} else {
					return
				}
			}
			klines := make([]*Kline, 0, len(rows))
			for _, it := range rows {
				row, ok := it.([]interface{})
				if !ok {
					continue
				}
				if k := buildKlineFromRow(row); k != nil {
					klines = append(klines, k)
				}
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

			if gateShouldLog("kline_"+cbKey, 30*time.Second) {
				g.Log().Warningf(gt.ctx, "[GateWS] kline update: cbKey=%s, interval=%s, klines=%d, callbacks=%d",
					cbKey, interval, len(copyK), len(cbs))
			}

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
			contract = strings.TrimSpace(contract)
			interval = strings.ToLower(strings.TrimSpace(interval))
			if contract == "" || interval == "" {
				// 尝试兜底猜测（有些实现把 meta 放在别的位置）
				c2, i2 := extractMeta(arr)
				if contract == "" {
					contract = c2
				}
				if interval == "" {
					interval = i2
				}
			}
			if contract == "" || interval == "" {
				return
			}
			symbol := gateNormalizeSymbol(contract)
			cbKey := symbol + ":" + interval

			// candle row：优先使用前6位（t + 5字段）
			k := buildKlineFromRow(arr[:6])
			if k == nil {
				return
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

			if gateShouldLog("kline_"+cbKey, 30*time.Second) {
				g.Log().Warningf(gt.ctx, "[GateWS] kline update: cbKey=%s, interval=%s, klines=%d, callbacks=%d",
					cbKey, interval, len(copyK), len(cbs))
			}

			for _, cb := range cbs {
				if cb != nil {
					go cb(copyK)
				}
			}
			return
		}

		// 兜底：有些实现可能是 [contract, interval, t, o, h, l, c, v] 或者其它顺序
		// 这里尝试在数组里找出 meta，并在剩余元素中提取一段长度>=6的 row
		{
			contract, interval := extractMeta(arr)
			if contract == "" || interval == "" {
				return
			}
			// 找到一个“看起来像 row”的片段：第一个元素可解析成时间戳
			var row []interface{}
			for i := 0; i+5 < len(arr); i++ {
				if parseIntAny(arr[i]) > 0 {
					row = arr[i : i+6]
					break
				}
			}
			if row == nil {
				return
			}
			k := buildKlineFromRow(row)
			if k == nil {
				return
			}
			symbol := gateNormalizeSymbol(contract)
			cbKey := symbol + ":" + interval

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

			if gateShouldLog("kline_"+cbKey, 30*time.Second) {
				g.Log().Warningf(gt.ctx, "[GateWS] kline update: cbKey=%s, interval=%s, klines=%d, callbacks=%d",
					cbKey, interval, len(copyK), len(cbs))
			}

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

// gateNormalizeInterval 将“系统/调用方可能传入的 interval”归一到 GateWS 内部一致的 interval key。
// 目前系统 1小时K线字段为 "1h"，但上层可能传 "60m"；这里统一映射为 "1h"。
func gateNormalizeInterval(interval string) string {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "60m":
		return "1h"
	default:
		return strings.ToLower(strings.TrimSpace(interval))
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
	KlineKeyCount     int               `json:"klineKeyCount"`
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
		KlineKeyCount:     len(gt.klines),
		Tickers:           tickers,
	}
}
