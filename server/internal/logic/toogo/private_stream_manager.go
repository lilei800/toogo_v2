package toogo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"hotgo/internal/model/entity"
	"hotgo/internal/service"
	"hotgo/internal/websocket"
	"hotgo/utility/encrypt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// PrivateStreamManager 管理私有WS流（按 apiConfigId 复用，避免每个机器人开一条连接）
type PrivateStreamManager struct {
	mu sync.RWMutex

	streams map[string]*privateStreamEntry // key=platform:apiConfigId

	// robotDebounce: 避免同一robot在短时间内被多次事件触发导致goroutine风暴
	robotDebounce map[int64]time.Time
	// robotPosDeltaDebounce: positions/delta 推送去抖（避免 Binance 高频 position 事件导致前端/服务端风暴）
	robotPosDeltaDebounce map[int64]time.Time
	// tradeFillDebounce: 私有WS订单事件触发“成交流水落库”的去抖（按 platform+apiConfigId+symbol）
	tradeFillDebounce map[string]time.Time
}

type privateStreamEntry struct {
	platform   string
	apiConfig  *entity.TradingApiConfig
	stream     exchange.PrivateStream
	refCount   int
	symbolRefs map[string]int     // symbol -> refs
	robots     map[int64]struct{} // robotId set

	healthStopCh       chan struct{}
	lastSilentResyncAt time.Time
}

var (
	privateStreamManager     *PrivateStreamManager
	privateStreamManagerOnce sync.Once
)

func GetPrivateStreamManager() *PrivateStreamManager {
	privateStreamManagerOnce.Do(func() {
		privateStreamManager = &PrivateStreamManager{
			streams:       make(map[string]*privateStreamEntry),
			robotDebounce: make(map[int64]time.Time),
			robotPosDeltaDebounce: make(map[int64]time.Time),
			tradeFillDebounce:     make(map[string]time.Time),
		}
	})
	return privateStreamManager
}

func streamKey(platform string, apiConfigId int64) string {
	return strings.ToLower(strings.TrimSpace(platform)) + ":" + g.NewVar(apiConfigId).String()
}

// Acquire 引用一个私有流（启动/复用）
func (m *PrivateStreamManager) Acquire(ctx context.Context, apiConfig *entity.TradingApiConfig, symbol string, robotId int64) error {
	if apiConfig == nil {
		return gerror.New("apiConfig is nil")
	}
	key := streamKey(apiConfig.Platform, apiConfig.Id)
	// 【强一致】内部 key 使用 BTCUSDT；对 WS 订阅入参使用平台格式（OKX/Gate 等不接受 BTCUSDT）
	canonical := exchange.Formatter.NormalizeSymbol(symbol)
	platform := strings.ToLower(strings.TrimSpace(apiConfig.Platform))
	subSymbol := canonical
	if canonical != "" {
		subSymbol = exchange.Formatter.FormatForPlatform(platform, canonical)
	}

	m.mu.Lock()
	entry := m.streams[key]
	if entry != nil {
		entry.refCount++
		if canonical != "" {
			entry.symbolRefs[canonical]++
			_ = entry.stream.AddSymbol(exchange.Formatter.FormatForPlatform(entry.platform, canonical))
		}
		if robotId > 0 {
			entry.robots[robotId] = struct{}{}
		}
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	g.Log().Infof(ctx, "[PrivateStreamManager] 创建新的私有WS: platform=%s, apiConfigId=%d, symbol=%s",
		platform, apiConfig.Id, canonical)

	// create new entry outside lock (but with a second check)
	cfg, err := buildExchangeConfigFromAPIConfig(ctx, apiConfig)
	if err != nil {
		g.Log().Warningf(ctx, "[PrivateStreamManager] buildExchangeConfigFromAPIConfig失败: platform=%s, err=%v", platform, err)
		return err
	}
	ps, err := exchange.NewPrivateStream(cfg)
	if err != nil {
		g.Log().Warningf(ctx, "[PrivateStreamManager] NewPrivateStream失败: platform=%s, err=%v", platform, err)
		return err
	}

	// proxy dialer（复用 RobotTaskManager 的全局代理配置）
	if dialer, err := getWebSocketDialer(ctx); err == nil && dialer != nil {
		ps.SetProxyDialer(dialer)
	}
	ps.SetOnEvent(func(ev *exchange.PrivateEvent) {
		if ev != nil {
			ev.ApiConfigId = apiConfig.Id
		}
		m.onEvent(ev)
	})

	if err := ps.Start(ctx); err != nil {
		g.Log().Warningf(ctx, "[PrivateStreamManager] 私有WS Start失败: platform=%s, err=%v", platform, err)
		return err
	}
	g.Log().Infof(ctx, "[PrivateStreamManager] 私有WS已启动: platform=%s, apiConfigId=%d", platform, apiConfig.Id)
	if canonical != "" {
		_ = ps.AddSymbol(subSymbol)
	}

	m.mu.Lock()
	// double check
	if existing := m.streams[key]; existing != nil {
		// someone created in between, close this one
		m.mu.Unlock()
		ps.Stop()
		return nil
	}
	newEntry := &privateStreamEntry{
		platform:     platform,
		apiConfig:    apiConfig,
		stream:       ps,
		refCount:     1,
		symbolRefs:   make(map[string]int),
		robots:       make(map[int64]struct{}),
		healthStopCh: make(chan struct{}),
	}
	if canonical != "" {
		newEntry.symbolRefs[canonical] = 1
	}
	if robotId > 0 {
		newEntry.robots[robotId] = struct{}{}
	}
	m.streams[key] = newEntry
	m.mu.Unlock()

	// Bitget 私有WS在部分环境下可能“连接在但不再推送/丢事件”，导致本地持仓/订单不同步。
	// 这里增加健康检测：当 WS 沉默超过阈值，触发一次轻量 resync（positions/openOrders 对账）恢复最终一致性。
	if platform == "bitget" {
		m.startPrivateStreamHealthLoop(key)
	}
	return nil
}

func (m *PrivateStreamManager) startPrivateStreamHealthLoop(key string) {
	// 防重复：如果已经有 loop 在跑，依赖 stopCh + key 存在性自然退出；这里不额外加全局 map。
	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			m.mu.RLock()
			entry := m.streams[key]
			m.mu.RUnlock()
			if entry == nil {
				return
			}
			select {
			case <-entry.healthStopCh:
				return
			case <-ticker.C:
			}

			// only for streams that provide status
			sp, ok := entry.stream.(exchange.PrivateStreamStatusProvider)
			if !ok {
				continue
			}
			if entry.stream == nil || !entry.stream.IsRunning() {
				continue
			}

			now := time.Now()
			lastMsgAt := sp.LastMessageAt()
			lastEventAt := sp.LastEventAt()

			// 尚未收到任何消息：给它一点启动时间
			if lastMsgAt.IsZero() && lastEventAt.IsZero() {
				continue
			}

			// 判定“沉默”：消息和业务事件都长时间不来（阈值偏保守，避免误触发）
			msgAge := now.Sub(lastMsgAt)
			evAge := now.Sub(lastEventAt)
			silent := msgAge >= 35*time.Second || evAge >= 45*time.Second
			if !silent {
				continue
			}

			// 节流：每个 private stream 10s 最多触发一次 resync
			m.mu.Lock()
			if !entry.lastSilentResyncAt.IsZero() && now.Sub(entry.lastSilentResyncAt) < 10*time.Second {
				m.mu.Unlock()
				continue
			}
			entry.lastSilentResyncAt = now
			// snapshot robot ids
			robotIds := make([]int64, 0, len(entry.robots))
			for rid := range entry.robots {
				robotIds = append(robotIds, rid)
			}
			m.mu.Unlock()

			if isOrderPositionSyncDebugEnabled(ctx) && shouldLogOrderPositionSync("ws_silent:"+key, 3*time.Second) {
				g.Log().Warningf(ctx, "[SyncDiag] bitget privateWS silent -> resync: key=%s msgAge=%v eventAge=%v robots=%d", key, msgAge, evAge, len(robotIds))
			}

			for _, rid := range robotIds {
				if eng := GetRobotTaskManager().GetEngine(rid); eng != nil {
					go eng.syncAccountDataIfNeeded(context.Background(), "ws_silent")
				}
				// 触发对账（positions/openOrders）兜底
				GetOrderStatusSyncService().TriggerRobotSync(rid)
			}
		}
	}()
}

// Release 释放引用，引用归零则停止流
func (m *PrivateStreamManager) Release(platform string, apiConfigId int64, symbol string, robotId int64) {
	key := streamKey(platform, apiConfigId)
	canonical := exchange.Formatter.NormalizeSymbol(symbol)

	m.mu.Lock()
	entry := m.streams[key]
	if entry == nil {
		m.mu.Unlock()
		return
	}
	if robotId > 0 {
		delete(entry.robots, robotId)
	}
	if canonical != "" {
		if n, ok := entry.symbolRefs[canonical]; ok {
			n--
			if n <= 0 {
				delete(entry.symbolRefs, canonical)
				_ = entry.stream.RemoveSymbol(exchange.Formatter.FormatForPlatform(entry.platform, canonical))
			} else {
				entry.symbolRefs[canonical] = n
			}
		}
	}
	entry.refCount--
	if entry.refCount <= 0 {
		delete(m.streams, key)
		// stop health loop
		if entry.healthStopCh != nil {
			close(entry.healthStopCh)
		}
		ps := entry.stream
		m.mu.Unlock()
		ps.Stop()
		return
	}
	m.mu.Unlock()
}

func (m *PrivateStreamManager) onEvent(ev *exchange.PrivateEvent) {
	if ev == nil {
		return
	}
	ctx := context.Background()

	// Binance: 实时落库成交流水（来自 ORDER_TRADE_UPDATE），解决“成交流水滞后/缺失”
	// - 用 trade_id 幂等去重（hg_trading_trade_fill.uk_api_exchange_trade）
	// - 归属 robot/user：优先通过 order_id 匹配本地订单；否则仅在 api_config_id+symbol 唯一时兜底归属
	if ev.Platform == "binance" && ev.Type == exchange.PrivateEventOrder {
		m.tryUpsertBinanceTradeFillFromOrderEvent(ctx, ev)
	}

	// 找到对应 stream entry，分发给关联 robot（按 apiConfigId 精准路由）
	key := streamKey(ev.Platform, ev.ApiConfigId)
	m.mu.RLock()
	entry := m.streams[key]
	if entry == nil {
		m.mu.RUnlock()
		return
	}
	// OKX/Gate: 私有WS订单事件不直接携带可用的“每笔成交fill + realizedPnl”数据，
	// 但事件出现通常意味着发生了成交/平仓。这里做两层处理：
	// 1) 尝试从 WS payload 直接解析 fill-level 信息并实时 upsert（更快、更省 API）
	// 2) 若解析失败/字段缺失，则节流后拉取最近N条成交并 upsert（兜底）
	if ev.Type == exchange.PrivateEventOrder && (ev.Platform == "okx" || ev.Platform == "gate") && strings.TrimSpace(ev.Symbol) != "" {
		ok := false
		needBackfill := false
		if ev.Platform == "okx" {
			ok, needBackfill = m.tryUpsertOKXTradeFillsFromOrderEvent(ctx, ev)
		} else if ev.Platform == "gate" {
			ok, needBackfill = m.tryUpsertGateTradeFillsFromOrderEvent(ctx, ev)
		}
		if !ok || needBackfill {
			m.tryStoreRecentTradeFillsFromOrderEvent(ctx, entry, ev)
		}
	}
	// 按 symbol 过滤：事件没有 symbol（account update）则不过滤
	if ev.Symbol != "" && len(entry.symbolRefs) > 0 {
		if _, ok := entry.symbolRefs[exchange.Formatter.NormalizeSymbol(ev.Symbol)]; !ok {
			m.mu.RUnlock()
			return
		}
	}
	targets := make([]int64, 0, len(entry.robots))
	for rid := range entry.robots {
		targets = append(targets, rid)
	}
	m.mu.RUnlock()

	if isOrderPositionSyncDebugEnabled(ctx) && shouldLogOrderPositionSync("psm_ev:"+key+":"+string(ev.Type), 800*time.Millisecond) {
		g.Log().Warningf(ctx, "[SyncDiag] privateWS event: platform=%s apiConfigId=%d type=%s symbol=%s targets=%d", ev.Platform, ev.ApiConfigId, ev.Type, ev.Symbol, len(targets))
	}

	now := time.Now()
	for _, robotId := range targets {
		// debounce 200ms
		m.mu.Lock()
		last := m.robotDebounce[robotId]
		if !last.IsZero() && now.Sub(last) < 200*time.Millisecond {
			m.mu.Unlock()
			continue
		}
		m.robotDebounce[robotId] = now
		m.mu.Unlock()

		engine := GetRobotTaskManager().GetEngine(robotId)
		if engine != nil {
			if isOrderPositionSyncDebugEnabled(ctx) && shouldLogOrderPositionSync("psm_dispatch:"+g.NewVar(robotId).String()+":"+string(ev.Type), 800*time.Millisecond) {
				g.Log().Warningf(ctx, "[SyncDiag] dispatch->syncAccountDataIfNeeded(after_trade): robotId=%d type=%s symbol=%s", robotId, ev.Type, ev.Symbol)
			}
			// 风暴控制：
			// - Binance ACCOUNT_UPDATE 是账户级广播：只更新余额/持仓缓存，不触发 after_trade/对账（避免把所有机器人打爆）
			// - Binance position 事件来源于 ACCOUNT_UPDATE：已直接写入引擎缓存，不需要再触发 after_trade（避免多余REST）
			if !(ev.Platform == "binance" && ev.Type == exchange.PrivateEventAccount) &&
				!(ev.Platform == "binance" && ev.Type == exchange.PrivateEventPosition) {
				go engine.syncAccountDataIfNeeded(context.Background(), "after_trade")
			}
			// Binance: position event is derived from ACCOUNT_UPDATE. Use it to refresh engine positions cache directly
			// so subsequent GetRobotPositions hits in-memory cache (no REST) and UI updates are immediate/stable.
			if ev.Platform == "binance" && ev.Type == exchange.PrivateEventPosition && strings.TrimSpace(ev.Symbol) != "" {
				if ps, ok := parseBinancePositionsFromAccountUpdate(ev.Raw, ev.Symbol); ok {
					engine.updatePositionsCacheFromPrivateWS(ps, ev.ReceivedAt)
				}
			}
			// account 事件：尽量事件驱动刷新余额（尤其 Gate 没有稳定的 account 推送时，解析成功可直接写缓存）
			if ev.Type == exchange.PrivateEventAccount {
				if bal, ok := exchange.ParseBalanceFromPrivateWS(ev.Platform, ev.Raw); ok && bal != nil {
					engine.updateBalanceCacheFromPrivateWS(bal, ev.ReceivedAt)
				} else {
					// 解析失败：兜底触发一次低频 REST 刷新（smart 内部会做 timeout/去重）
					// Binance 的 account_update 频率较高：额外加一层节流，避免解析失败时频繁打 REST
					if ev.Platform != "binance" {
						go engine.refreshBalanceCacheAfterTrade(context.Background(), "after_private_account_event")
					} else {
						engine.mu.RLock()
						lastBalAt := engine.LastBalanceUpdate
						engine.mu.RUnlock()
						if lastBalAt.IsZero() || time.Since(lastBalAt) > 5*time.Second {
							go engine.refreshBalanceCacheAfterTrade(context.Background(), "after_private_account_event")
						}
					}
				}
			}
		}
		// 【方案A】私有WS订单事件增量落库（挂单/订单事实表）
		// 说明：落库尽量轻量，失败不阻断后续对账；若表未创建，会在执行日志里体现。
		if ev.Type == exchange.PrivateEventOrder {
			go UpsertExchangeOrdersFromPrivateEvent(context.Background(), robotId, ev)
		}
		// 【阶段C】私有WS持仓事件：即时推送 positions snapshot（用于前端避免“持仓丢失/闪烁”）
		// 说明：
		// - 不替代 positions/subscribe 的定时快照；它是兜底
		// - 这里用于“事件驱动的立即刷新”：交易所推送了 position 变更，就主动推一帧最新快照给该用户
		if ev.Type == exchange.PrivateEventPosition {
			// positions/delta 去抖：同一 robot 300ms 内最多推一次
			m.mu.Lock()
			lastDelta := m.robotPosDeltaDebounce[robotId]
			if !lastDelta.IsZero() && now.Sub(lastDelta) < 300*time.Millisecond {
				m.mu.Unlock()
			} else {
				m.robotPosDeltaDebounce[robotId] = now
				m.mu.Unlock()
			go func(robotId int64, receivedAt int64) {
				ctx := context.Background()
				meta, err := getRobotMeta(ctx, robotId)
				if err != nil || meta.UserId <= 0 {
					return
				}
				list, err := service.ToogoRobot().GetRobotPositions(ctx, robotId)
				data := g.Map{
					"robotId": robotId,
					"list":    list,
					"error":   "",
					"stale":   false,
					"ts":      receivedAt,
				}
				if err != nil {
					// 失败则只推“错误通知”，避免推空覆盖前端
					data["error"] = err.Error()
					data["stale"] = true
				}
				websocket.SendToUser(meta.UserId, &websocket.WResponse{
					Event: "toogo/robot/positions/delta",
					Data:  data,
				})
			}(robotId, ev.ReceivedAt)
			}
		}
		// 触发DB对账（按robot去抖）
		if isOrderPositionSyncDebugEnabled(ctx) && shouldLogOrderPositionSync("psm_trigger_db:"+g.NewVar(robotId).String(), 800*time.Millisecond) {
			g.Log().Warningf(ctx, "[SyncDiag] trigger OrderStatusSyncService: robotId=%d", robotId)
		}
		// 风暴控制：account 事件不触发对账（特别是 Binance 的账户级广播）
		if ev.Type != exchange.PrivateEventAccount {
			GetOrderStatusSyncService().TriggerRobotSync(robotId)
		}
	}
}

// tryUpsertOKXTradeFillsFromOrderEvent parses OKX private WS "orders" payload and upserts fill-level trades if present.
// Returns true if at least one fill was upserted.
func (m *PrivateStreamManager) tryUpsertOKXTradeFillsFromOrderEvent(ctx context.Context, ev *exchange.PrivateEvent) (ok bool, needBackfill bool) {
	if ev == nil || ev.Platform != "okx" || ev.ApiConfigId <= 0 || len(ev.Raw) == 0 {
		return false, false
	}
	symbol := strings.TrimSpace(ev.Symbol)
	if symbol == "" {
		return false, false
	}

	j := gjson.New(string(ev.Raw))
	data := j.Get("data").Array()
	if len(data) == 0 {
		return false, false
	}

	abs := func(x float64) float64 {
		if x < 0 {
			return -x
		}
		return x
	}

	trades := make([]*exchange.Trade, 0, len(data))
	for _, it := range data {
		d := gjson.New(it)
		ordId := strings.TrimSpace(d.Get("ordId").String())
		if ordId == "" {
			continue
		}
		px := d.Get("fillPx").Float64()
		sz := abs(d.Get("fillSz").Float64())
		if px <= 0 || sz <= 0 {
			// OKX orders channel may include non-fill updates; skip those.
			continue
		}
		fee := abs(d.Get("fillFee").Float64())
		feeCcy := strings.TrimSpace(d.Get("feeCcy").String())
		if feeCcy == "" {
			feeCcy = strings.TrimSpace(d.Get("fillFeeCcy").String())
		}
		// pnl field variants on OKX private orders
		pnl := d.Get("fillPnl").Float64()
		if pnl == 0 {
			pnl = d.Get("pnl").Float64()
		}
		if pnl == 0 {
			pnl = d.Get("realizedPnl").Float64()
		}
		// time: prefer fillTime, then uTime/cTime
		ts := g.NewVar(d.Get("fillTime").String()).Int64()
		if ts <= 0 {
			ts = g.NewVar(d.Get("uTime").String()).Int64()
		}
		if ts <= 0 {
			ts = g.NewVar(d.Get("cTime").String()).Int64()
		}
		// ensure ms
		if ts > 0 && ts < 1e12 {
			ts *= 1000
		}
		tradeId := strings.TrimSpace(d.Get("fillId").String())
		if tradeId == "" {
			tradeId = strings.TrimSpace(d.Get("tradeId").String())
		}
		if tradeId == "" {
			tradeId = strings.TrimSpace(d.Get("billId").String())
		}

		side := strings.ToUpper(strings.TrimSpace(d.Get("side").String()))
		posSide := strings.ToUpper(strings.TrimSpace(d.Get("posSide").String()))

		tr := &exchange.Trade{
			TradeId:         tradeId,
			OrderId:         ordId,
			Symbol:          symbol,
			Side:            side,
			PositionSide:    posSide,
			Price:           px,
			Quantity:        sz, // NOTE: OKX fillSz is in contracts; REST sync will later correct qty by ctVal
			RealizedPnl:     pnl,
			Commission:      fee,
			CommissionAsset: feeCcy,
			Time:            ts,
		}
		// If fillId is missing, generate stable-ish id for idempotency.
		if strings.TrimSpace(tr.TradeId) == "" {
			tr.TradeId = fmt.Sprintf("%s-%d-%.8f-%.8f-%s", tr.OrderId, tr.Time, tr.Price, tr.Quantity, strings.ToUpper(strings.TrimSpace(tr.Side)))
		}
		trades = append(trades, tr)
	}
	if len(trades) == 0 {
		return false, false
	}
	_, _, _ = upsertTradeFillsFromTrades(ctx, ev.ApiConfigId, "okx", symbol, trades, nil)
	// OKX orders WS 通常已包含 fillFee/fillPnl；如缺失，可由后续定时/人工同步/外部检测链路补齐。
	// 这里默认不强制回填，避免每次 WS 事件都打 REST。
	return true, false
}

// tryUpsertGateTradeFillsFromOrderEvent parses Gate private WS futures.orders payload and upserts fill-level trades if present.
// Returns true if at least one fill was upserted.
func (m *PrivateStreamManager) tryUpsertGateTradeFillsFromOrderEvent(ctx context.Context, ev *exchange.PrivateEvent) (ok bool, needBackfill bool) {
	if ev == nil || ev.Platform != "gate" || ev.ApiConfigId <= 0 || len(ev.Raw) == 0 {
		return false, false
	}
	symbol := strings.TrimSpace(ev.Symbol)
	if symbol == "" {
		return false, false
	}

	var msg map[string]any
	if err := json.Unmarshal(ev.Raw, &msg); err != nil {
		return false, false
	}
	result, _ := msg["result"].(map[string]any)
	if result == nil {
		return false, false
	}
	asString := func(v any) string {
		if v == nil {
			return ""
		}
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
		return strings.TrimSpace(g.NewVar(v).String())
	}
	asFloat := func(v any) float64 {
		if v == nil {
			return 0
		}
		return g.NewVar(v).Float64()
	}
	asInt64 := func(v any) int64 {
		if v == nil {
			return 0
		}
		return g.NewVar(v).Int64()
	}
	abs := func(x float64) float64 {
		if x < 0 {
			return -x
		}
		return x
	}

	orderID := asString(result["id"])
	if orderID == "" {
		orderID = asString(result["order_id"])
	}
	if orderID == "" {
		return false, false
	}
	fillPrice := asFloat(result["fill_price"])
	if fillPrice <= 0 {
		// Some updates may not include fill info.
		return false, false
	}
	size := asFloat(result["size"])
	left := asFloat(result["left"])
	filledContracts := abs(size) - abs(left)
	if filledContracts <= 0 {
		// no new filled qty in this update
		return false, false
	}

	fee := abs(asFloat(result["fee"]))
	pnl := asFloat(result["pnl"])
	if pnl == 0 {
		pnl = asFloat(result["realized_pnl"])
	}
	ts := asInt64(result["finish_time"])
	if ts <= 0 {
		ts = asInt64(msg["time"])
	}
	// Gate uses seconds in WS time fields
	if ts > 0 && ts < 1e12 {
		ts *= 1000
	}

	side := "BUY"
	if size < 0 {
		side = "SELL"
	}
	tradeId := asString(result["trade_id"])
	if tradeId == "" {
		// Gate WS does not always provide a distinct fill id; generate a stable-ish one.
		tradeId = fmt.Sprintf("%s-%d-%.8f-%.8f-%s", orderID, ts, fillPrice, filledContracts, side)
	}

	tr := &exchange.Trade{
		TradeId:         tradeId,
		OrderId:         orderID,
		Symbol:          symbol,
		Side:            side,
		PositionSide:    "", // Gate WS may omit; REST/account_book will fill pnl later if needed
		Price:           fillPrice,
		Quantity:        filledContracts, // NOTE: contracts; REST sync will later correct qty by multiplier
		RealizedPnl:     pnl,
		Commission:      fee,
		CommissionAsset: "USDT",
		Time:            ts,
	}

	_, _, _ = upsertTradeFillsFromTrades(ctx, ev.ApiConfigId, "gate", symbol, []*exchange.Trade{tr}, nil)
	// Gate WS 往往缺失 pnl/fee：触发一次节流的 REST+账本补齐（Gate.GetTradeHistory 内部会 account_book 优先回填）
	if tr.RealizedPnl == 0 || tr.Commission == 0 {
		return true, true
	}
	return true, false
}

// tryStoreRecentTradeFillsFromOrderEvent fetches recent trades from exchange and upserts into DB.
// It is designed for exchanges whose private WS order events do not contain sufficient fill-level pnl/fee details.
func (m *PrivateStreamManager) tryStoreRecentTradeFillsFromOrderEvent(ctx context.Context, entry *privateStreamEntry, ev *exchange.PrivateEvent) {
	if entry == nil || ev == nil || strings.TrimSpace(ev.Symbol) == "" || entry.apiConfig == nil {
		return
	}
	platform := strings.ToLower(strings.TrimSpace(ev.Platform))
	symbol := exchange.Formatter.NormalizeSymbol(ev.Symbol)
	if symbol == "" {
		return
	}

	// debounce by (platform+apiConfigId+symbol)
	debounceKey := platform + ":" + g.NewVar(ev.ApiConfigId).String() + ":" + symbol
	now := time.Now()
	m.mu.Lock()
	last := m.tradeFillDebounce[debounceKey]
	if !last.IsZero() && now.Sub(last) < 3*time.Second {
		m.mu.Unlock()
		return
	}
	m.tradeFillDebounce[debounceKey] = now
	m.mu.Unlock()

	go func(apiCfg *entity.TradingApiConfig, apiId int64, sym string) {
		tctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()

		ex, err := GetExchangeManager().GetExchangeFromConfig(tctx, apiCfg)
		if err != nil || ex == nil {
			return
		}
		// 轻量拉取：足够覆盖一次成交/平仓的 fills
		_, _, _ = fetchAndStoreTradeHistory(tctx, ex, apiId, ex.GetName(), sym, 200)
	}(entry.apiConfig, ev.ApiConfigId, symbol)
}

// parseBinancePositionsFromAccountUpdate extracts positions for a given symbol from Binance ACCOUNT_UPDATE payload.
// It returns a slice representing the latest snapshot for that symbol (can be empty slice for "no positions").
func parseBinancePositionsFromAccountUpdate(raw []byte, symbol string) ([]*exchange.Position, bool) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if len(raw) == 0 || symbol == "" {
		return nil, false
	}
	var msg map[string]any
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, false
	}
	a, _ := msg["a"].(map[string]any)
	if a == nil {
		return nil, false
	}
	ps, _ := a["P"].([]any)
	if len(ps) == 0 {
		// payload may omit positions; treat as "no update"
		return nil, false
	}

	asString := func(v any) string {
		if v == nil {
			return ""
		}
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
		return strings.TrimSpace(g.NewVar(v).String())
	}
	asFloat := func(v any) float64 {
		if v == nil {
			return 0
		}
		return g.NewVar(v).Float64()
	}
	asInt := func(v any) int {
		if v == nil {
			return 0
		}
		return g.NewVar(v).Int()
	}

	foundSymbol := false
	out := make([]*exchange.Position, 0, 2)
	for _, it := range ps {
		p, ok := it.(map[string]any)
		if !ok {
			continue
		}
		sym := strings.ToUpper(asString(p["s"]))
		if sym == "" || !strings.EqualFold(sym, symbol) {
			continue
		}
		foundSymbol = true

		amt := asFloat(p["pa"])
		ep := asFloat(p["ep"])
		up := asFloat(p["up"])
		mt := strings.ToUpper(asString(p["mt"])) // "isolated"/"cross"
		iw := asFloat(p["iw"])
		lev := asInt(p["l"])
		pside := strings.ToUpper(asString(p["ps"])) // LONG/SHORT/BOTH

		// normalize margin type
		if strings.EqualFold(mt, "isolated") {
			mt = exchange.MarginTypeIsolated
		} else if strings.EqualFold(mt, "cross") {
			mt = exchange.MarginTypeCrossed
		}

		// position side: BOTH means one-way mode; derive by sign
		if pside == "" || pside == "BOTH" {
			if amt < 0 {
				pside = "SHORT"
			} else {
				pside = "LONG"
			}
		}

		// ignore tiny noise; but still allow closing by returning empty slice
		if math.Abs(amt) <= positionAmtEpsilon {
			continue
		}

		// best-effort: ensure sign matches side for consistency
		if pside == "SHORT" && amt > 0 {
			amt = -amt
		}
		if pside == "LONG" && amt < 0 {
			amt = math.Abs(amt)
		}

		out = append(out, &exchange.Position{
			Symbol:         symbol,
			PositionSide:   pside,
			PositionAmt:    amt,
			EntryPrice:     ep,
			UnrealizedPnl:  up,
			Leverage:       lev,
			MarginType:     mt,
			IsolatedMargin: iw,
		})
	}
	if !foundSymbol {
		return nil, false
	}
	// Found symbol but no active positions -> clear snapshot to empty slice
	if len(out) == 0 {
		return []*exchange.Position{}, true
	}
	return out, true
}

// tryUpsertBinanceTradeFillFromOrderEvent 从 Binance ORDER_TRADE_UPDATE 事件中抽取“本次成交(fill)”并实时落库。
// 说明：
// - Binance WS 订单事件可能只有状态变更而没有成交（lastFilledQty=0），此时不落库
// - 落库不依赖 robotId（通过 order_id 匹配本地订单；否则按 api_config_id+symbol 唯一性兜底）
func (m *PrivateStreamManager) tryUpsertBinanceTradeFillFromOrderEvent(ctx context.Context, ev *exchange.PrivateEvent) {
	if ev == nil || ev.Platform != "binance" || ev.ApiConfigId <= 0 || len(ev.Raw) == 0 {
		return
	}

	var msg map[string]any
	if err := json.Unmarshal(ev.Raw, &msg); err != nil {
		return
	}
	// Binance ORDER_TRADE_UPDATE payload is under "o"
	o, _ := msg["o"].(map[string]any)
	if o == nil {
		return
	}

	// helper converters
	asString := func(v any) string {
		if v == nil {
			return ""
		}
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
		return strings.TrimSpace(g.NewVar(v).String())
	}
	asFloat := func(v any) float64 {
		if v == nil {
			return 0
		}
		return g.NewVar(v).Float64()
	}
	asInt64 := func(v any) int64 {
		if v == nil {
			return 0
		}
		return g.NewVar(v).Int64()
	}

	// last filled quantity/price
	qty := asFloat(o["l"])
	price := asFloat(o["L"])
	if qty <= 0 || price <= 0 {
		return
	}

	orderID := asString(o["i"])
	if orderID == "" {
		// 没有 orderId 无法做关联归属，也无法作为去重辅助
		return
	}

	// time: trade time preferred, fallback to event time
	ts := asInt64(o["T"])
	if ts <= 0 {
		ts = asInt64(msg["E"])
	}

	// Build exchange.Trade (our normalized fill model)
	trade := &exchange.Trade{
		TradeId:         asString(o["t"]),
		OrderId:         orderID,
		Symbol:          strings.ToUpper(asString(o["s"])),
		Side:            strings.ToUpper(asString(o["S"])),  // BUY/SELL
		PositionSide:    strings.ToUpper(asString(o["ps"])), // LONG/SHORT
		Price:           price,
		Quantity:        qty,
		RealizedPnl:     asFloat(o["rp"]),
		Commission:      asFloat(o["n"]),
		CommissionAsset: asString(o["N"]),
		Time:            ts,
	}
	// Some Binance streams may omit tradeId. Generate a stable-ish fallback for idempotency.
	// NOTE: upsertTradeFillsFromTrades also has a fallback, but doing it here avoids empty trade_id reaching DB layer.
	if strings.TrimSpace(trade.TradeId) == "" {
		trade.TradeId = fmt.Sprintf("%s-%d-%.8f-%.8f-%s",
			trade.OrderId, trade.Time, trade.Price, trade.Quantity, strings.ToUpper(strings.TrimSpace(trade.Side)))
	}

	// symbol fallback: prefer ev.Symbol (already normalized for routing)
	symbol := strings.TrimSpace(ev.Symbol)
	if symbol == "" {
		symbol = trade.Symbol
	}
	if symbol == "" {
		return
	}

	// Upsert (idempotent). Ignore errors here to avoid blocking WS event flow.
	_, _, _ = upsertTradeFillsFromTrades(ctx, ev.ApiConfigId, "binance", symbol, []*exchange.Trade{trade}, nil)
}

// buildExchangeConfigFromAPIConfig 构建 exchange.Config（解密字段，补代理）
func buildExchangeConfigFromAPIConfig(ctx context.Context, apiConfig *entity.TradingApiConfig) (*exchange.Config, error) {
	if apiConfig == nil {
		return nil, gerror.New("apiConfig is nil")
	}

	apiKey, err := encrypt.AesDecrypt(apiConfig.ApiKey)
	if err != nil {
		apiKey = apiConfig.ApiKey
	}
	secretKey, err := encrypt.AesDecrypt(apiConfig.SecretKey)
	if err != nil {
		secretKey = apiConfig.SecretKey
	}
	passphrase := ""
	if apiConfig.Passphrase != "" {
		p, err := encrypt.AesDecrypt(apiConfig.Passphrase)
		if err != nil {
			passphrase = apiConfig.Passphrase
		} else {
			passphrase = p
		}
	}

	// 代理配置（全局）
	var proxyCfg *exchange.ProxyConfig
	var proxyEnt *entity.TradingProxyConfig
	_ = dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).
		Where(dao.TradingProxyConfig.Columns().TenantId, 0).
		Where(dao.TradingProxyConfig.Columns().Enabled, 1).
		Scan(&proxyEnt)
	if proxyEnt != nil {
		host := proxyEnt.ProxyAddress
		port := 0
		if idx := strings.Index(proxyEnt.ProxyAddress, ":"); idx > 0 {
			host = proxyEnt.ProxyAddress[:idx]
			port = g.NewVar(proxyEnt.ProxyAddress[idx+1:]).Int()
		}
		proxyCfg = &exchange.ProxyConfig{
			Enabled:  true,
			Type:     proxyEnt.ProxyType,
			Host:     host,
			Port:     port,
			Username: proxyEnt.Username,
			Password: "",
		}
		if proxyEnt.AuthEnabled == 1 && proxyEnt.Password != "" {
			if pwd, err := encrypt.AesDecrypt(proxyEnt.Password); err == nil {
				proxyCfg.Password = pwd
			}
		}
	}

	return &exchange.Config{
		Platform:   strings.ToLower(strings.TrimSpace(apiConfig.Platform)),
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		IsTestnet:  false,
		Proxy:      proxyCfg,
	}, nil
}
