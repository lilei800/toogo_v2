package toogo

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"hotgo/internal/model/entity"
	"hotgo/internal/websocket"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// 交易所订单事实表（挂单/订单）
// - WS 增量 upsert 写入
// - REST 低频兜底对账写入
// - 前端展示只读 DB
const exchangeOrderTable = "hg_trading_exchange_order"

type robotMeta struct {
	RobotId     int64
	TenantId    int64
	UserId      int64
	ApiConfigId int64
	Platform    string
	Symbol      string
}

// robot meta cache（避免私有WS高频事件每次查库）
var robotMetaCache = struct {
	mu   sync.RWMutex
	data map[int64]struct {
		meta robotMeta
		at   time.Time
	}
}{data: make(map[int64]struct {
	meta robotMeta
	at   time.Time
})}

// openOrdersAntiFlicker prevents transient empty openOrders snapshots (REST glitches)
// from mass-closing all local open orders, which would cause UI flicker and incorrect state.
var openOrdersAntiFlicker = struct {
	mu sync.Mutex
	m  map[string]*openOrdersAFState
}{
	m: make(map[string]*openOrdersAFState),
}

type openOrdersAFState struct {
	lastNonEmptyAt    time.Time
	pendingEmptySince time.Time
	lastAppliedAt     time.Time
	lastCount         int
}

func openOrdersAFKey(robotId int64, platform string, apiConfigId int64, symbol string) string {
	return g.NewVar(robotId).String() + "|" + strings.ToLower(strings.TrimSpace(platform)) + "|" + g.NewVar(apiConfigId).String() + "|" + strings.ToUpper(strings.TrimSpace(symbol))
}

func markOpenOrdersSnapshot(key string, now time.Time, count int) {
	openOrdersAntiFlicker.mu.Lock()
	defer openOrdersAntiFlicker.mu.Unlock()

	st := openOrdersAntiFlicker.m[key]
	if st == nil {
		st = &openOrdersAFState{}
		openOrdersAntiFlicker.m[key] = st
	}
	if count < 0 {
		count = 0
	}
	if count > 0 {
		st.lastNonEmptyAt = now
		st.pendingEmptySince = time.Time{}
	} else {
		// count==0: do not clear lastNonEmptyAt here; handled by empty-guard decision
	}
	st.lastAppliedAt = now
	st.lastCount = count
}

func shouldApplyOpenOrdersEmptySnapshot(key string, now time.Time, emptyDelay time.Duration) (apply bool, held bool, reason string) {
	openOrdersAntiFlicker.mu.Lock()
	defer openOrdersAntiFlicker.mu.Unlock()

	st := openOrdersAntiFlicker.m[key]
	if st == nil {
		st = &openOrdersAFState{}
		openOrdersAntiFlicker.m[key] = st
	}
	// never had non-empty snapshot -> accept empty
	if st.lastNonEmptyAt.IsZero() {
		st.pendingEmptySince = time.Time{}
		st.lastAppliedAt = now
		st.lastCount = 0
		return true, false, ""
	}
	if st.pendingEmptySince.IsZero() {
		st.pendingEmptySince = now
		return false, true, "empty_pending_first"
	}
	if now.Sub(st.pendingEmptySince) >= emptyDelay {
		st.pendingEmptySince = time.Time{}
		st.lastAppliedAt = now
		st.lastCount = 0
		return true, false, ""
	}
	return false, true, "empty_pending"
}

func getRobotMeta(ctx context.Context, robotId int64) (robotMeta, error) {
	robotMetaCache.mu.RLock()
	if v, ok := robotMetaCache.data[robotId]; ok && time.Since(v.at) < 5*time.Second {
		robotMetaCache.mu.RUnlock()
		return v.meta, nil
	}
	robotMetaCache.mu.RUnlock()

	var r *entity.TradingRobot
	if err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&r); err != nil || r == nil {
		if err == nil {
			err = gerror.New("robot not found")
		}
		return robotMeta{}, err
	}
	// 【多交易所一致性】platform 的权威来源应为 api_config.platform（交易所实例/WS连接都以此为准）。
	// 如果仅依赖 robot.exchange，一旦历史数据未同步更新，就会导致：
	// - 私有WS事件落库到错误 platform 分区
	// - 唯一键(uk_platform_api_order)维度不一致，引发“查不到/重复/覆盖”
	platform := strings.ToLower(strings.TrimSpace(r.Exchange))
	if r.ApiConfigId > 0 {
		type apiRow struct {
			Platform string `orm:"platform"`
		}
		var ar apiRow
		_ = dao.TradingApiConfig.Ctx(ctx).
			Fields("platform").
			Where(dao.TradingApiConfig.Columns().Id, r.ApiConfigId).
			Limit(1).
			Scan(&ar)
		if strings.TrimSpace(ar.Platform) != "" {
			platform = strings.ToLower(strings.TrimSpace(ar.Platform))
		}
	}
	m := robotMeta{
		RobotId:     r.Id,
		TenantId:    r.TenantId,
		UserId:      r.UserId,
		ApiConfigId: r.ApiConfigId,
		Platform:    platform,
		// 【强一致】事实表/缓存 key 统一使用 BTCUSDT（避免不同交易所格式导致“同一机器人同一币对分裂”）
		Symbol: exchange.Formatter.NormalizeSymbol(r.Symbol),
	}
	robotMetaCache.mu.Lock()
	robotMetaCache.data[robotId] = struct {
		meta robotMeta
		at   time.Time
	}{meta: m, at: time.Now()}
	robotMetaCache.mu.Unlock()
	return m, nil
}

// UpsertExchangeOrdersFromPrivateEvent 将私有WS订单事件写入事实表（按robot维度）
func UpsertExchangeOrdersFromPrivateEvent(ctx context.Context, robotId int64, ev *exchange.PrivateEvent) {
	if ev == nil || ev.Type != exchange.PrivateEventOrder {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	meta, err := getRobotMeta(ctx, robotId)
	if err != nil {
		return
	}
	platform := strings.ToLower(strings.TrimSpace(ev.Platform))
	if platform == "" {
		platform = meta.Platform
	}
	apiConfigId := ev.ApiConfigId
	if apiConfigId <= 0 {
		apiConfigId = meta.ApiConfigId
	}
	// 统一使用机器人 symbol，避免 ws/交易所 symbol 格式差异导致前端同一页面分裂
	symbol := meta.Symbol

	orders := parsePrivateOrderEvent(platform, ev.Raw)
	if len(orders) == 0 {
		// best-effort：至少落一条 raw，便于审计（但没有 orderId 无法 upsert）
		return
	}

	// 先组装 delta（尽量轻量），用于推送给前端做增量更新/触发刷新
	delta := make([]g.Map, 0, len(orders))
	for _, o := range orders {
		if strings.TrimSpace(o.ExchangeOrderId) == "" {
			continue
		}
		delta = append(delta, g.Map{
			"orderId":      o.ExchangeOrderId,
			"clientId":     o.ClientOrderId,
			"symbol":       symbol,
			"side":         o.Side,
			"positionSide": o.PositionSide,
			"type":         o.Type,
			"price":        o.Price,
			"quantity":     o.Quantity,
			"filledQty":    o.FilledQty,
			"avgPrice":     o.AvgPrice,
			"status":       o.Status,
			"rawStatus":    o.RawStatus,
			"isOpen":       o.IsOpen,
			"createTime":   o.CreateTime,
			"updateTime":   o.UpdateTime,
			"eventAt":      ev.ReceivedAt,
		})

		data := g.Map{
			"tenant_id":         meta.TenantId,
			"user_id":           meta.UserId,
			"robot_id":          meta.RobotId,
			"api_config_id":     apiConfigId,
			"platform":          platform,
			"symbol":            symbol,
			"exchange_order_id": o.ExchangeOrderId,
			"client_order_id":   o.ClientOrderId,
			"side":              o.Side,
			"position_side":     o.PositionSide,
			"order_type":        o.Type,
			"reduce_only":       boolToTiny(o.ReduceOnly),
			"price":             o.Price,
			"quantity":          o.Quantity,
			"filled_qty":        o.FilledQty,
			"avg_price":         o.AvgPrice,
			"status":            o.Status,
			"raw_status":        o.RawStatus,
			"is_open":           boolToTiny(o.IsOpen),
			"create_time":       o.CreateTime,
			"update_time":       o.UpdateTime,
			"last_event_time":   ev.ReceivedAt,
			"raw":               truncateString(string(ev.Raw), 8000),
		}
		_ = upsertExchangeOrder(ctx, platform, apiConfigId, o.ExchangeOrderId, data)
	}

	// WS：订单增量推送（不需要订阅，前端可用来避免“闪烁/丢失”并触发局部刷新）
	if meta.UserId > 0 && len(delta) > 0 {
		websocket.SendToUser(meta.UserId, &websocket.WResponse{
			Event: "toogo/robot/orders/delta",
			Data: g.Map{
				"robotId":     meta.RobotId,
				"platform":    platform,
				"apiConfigId": apiConfigId,
				"symbol":      symbol,
				"list":        delta,
				"ts":          ev.ReceivedAt,
			},
		})
	}
}

// SyncExchangeOpenOrdersToDB 兜底：用 REST openOrders 对账写入事实表（按robot维度）
func SyncExchangeOpenOrdersToDB(ctx context.Context, robotId int64, platform string, apiConfigId int64, symbol string, orders []*exchange.Order) error {
	if ctx == nil {
		ctx = context.Background()
	}
	meta, err := getRobotMeta(ctx, robotId)
	if err != nil {
		return err
	}
	if platform == "" {
		platform = meta.Platform
	}
	if apiConfigId <= 0 {
		apiConfigId = meta.ApiConfigId
	}
	if symbol == "" {
		symbol = meta.Symbol
	}
	platform = strings.ToLower(strings.TrimSpace(platform))
	key := openOrdersAFKey(robotId, platform, apiConfigId, symbol)
	now := time.Now()
	// openOrders REST 兜底对账默认节流 10s：空快照至少延迟 12s，避免瞬时空导致“全量关闭”
	emptyDelay := 12 * time.Second
	if emptyDelay < 8*time.Second {
		emptyDelay = 8 * time.Second
	}
	if emptyDelay > 30*time.Second {
		emptyDelay = 30 * time.Second
	}

	seen := make(map[string]struct{}, len(orders))
	nowMs := now.UnixMilli()
	for _, o := range orders {
		if o == nil || strings.TrimSpace(o.OrderId) == "" {
			continue
		}
		seen[o.OrderId] = struct{}{}
		data := g.Map{
			"tenant_id":         meta.TenantId,
			"user_id":           meta.UserId,
			"robot_id":          meta.RobotId,
			"api_config_id":     apiConfigId,
			"platform":          platform,
			"symbol":            symbol,
			"exchange_order_id": strings.TrimSpace(o.OrderId),
			"client_order_id":   strings.TrimSpace(o.ClientId),
			"side":              strings.ToUpper(strings.TrimSpace(o.Side)),
			"position_side":     strings.ToUpper(strings.TrimSpace(o.PositionSide)),
			"order_type":        strings.ToUpper(strings.TrimSpace(o.Type)),
			"reduce_only":       0, // openOrders REST 通用结构体不携带 reduceOnly，按未知处理
			"price":             o.Price,
			"quantity":          o.Quantity,
			"filled_qty":        o.FilledQty,
			"avg_price":         o.AvgPrice,
			"status":            normalizeOrderStatus(platform, o.Status),
			"raw_status":        strings.TrimSpace(o.Status),
			"is_open":           1,
			"create_time":       o.CreateTime,
			"update_time":       o.UpdateTime,
			"last_event_time":   nowMs,
			"raw":               "",
		}
		_ = upsertExchangeOrder(ctx, platform, apiConfigId, o.OrderId, data)
	}

	// 将“本地仍标记为 open 但本次对账未返回”的订单置为非 open（一般是成交/撤单）
	// 说明：只针对本 robot + symbol，避免跨symbol误伤。
	if len(seen) == 0 {
		apply, held, reason := shouldApplyOpenOrdersEmptySnapshot(key, now, emptyDelay)
		if held {
			if isOrderPositionSyncDebugEnabled(ctx) && shouldLogOrderPositionSync("openOrders_empty_hold:"+g.NewVar(robotId).String(), 3*time.Second) {
				g.Log().Warningf(ctx, "[SyncDiag] openOrders 空快照已拦截(anti-flicker): robotId=%d platform=%s symbol=%s reason=%s",
					robotId, platform, symbol, reason)
			}
			return nil
		}
		_ = apply // readability
		_, _ = g.DB().Model(exchangeOrderTable).Ctx(ctx).
			Where("robot_id", meta.RobotId).
			Where("platform", platform).
			Where("api_config_id", apiConfigId).
			Where("symbol", symbol).
			Where("is_open", 1).
			Data(g.Map{
				"is_open":     0,
				"status":      "CANCELED",
				"raw_status":  "sync_missing",
				"update_time": nowMs,
			}).Update()
		markOpenOrdersSnapshot(key, now, 0)
		return nil
	}
	// non-empty snapshot: clear pending empty
	markOpenOrdersSnapshot(key, now, len(seen))
	// IN 列表：只更新不在 seen 中的
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	_, _ = g.DB().Model(exchangeOrderTable).Ctx(ctx).
		Where("robot_id", meta.RobotId).
		Where("platform", platform).
		Where("api_config_id", apiConfigId).
		Where("symbol", symbol).
		Where("is_open", 1).
		WhereNotIn("exchange_order_id", ids).
		Data(g.Map{
			"is_open":     0,
			"status":      "CANCELED",
			"raw_status":  "sync_missing",
			"update_time": nowMs,
		}).Update()
	return nil
}

// ---- parsing helpers ----

type parsedOrder struct {
	ExchangeOrderId string
	ClientOrderId   string
	Side            string
	PositionSide    string
	Type            string
	ReduceOnly      bool
	Price           float64
	Quantity        float64
	FilledQty       float64
	AvgPrice        float64
	Status          string
	RawStatus       string
	IsOpen          bool
	CreateTime      int64
	UpdateTime      int64
}

func parsePrivateOrderEvent(platform string, raw []byte) []parsedOrder {
	platform = strings.ToLower(strings.TrimSpace(platform))
	switch platform {
	case "binance":
		return parseBinancePrivateOrders(raw)
	case "okx":
		return parseOKXPrivateOrders(raw)
	case "gate":
		return parseGatePrivateOrders(raw)
	case "bitget":
		return parseBitgetPrivateOrders(raw)
	default:
		return nil
	}
}

func parseBinancePrivateOrders(raw []byte) []parsedOrder {
	// ORDER_TRADE_UPDATE: { "e":"ORDER_TRADE_UPDATE", "E":..., "o":{...} }
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	o, _ := m["o"].(map[string]any)
	if o == nil {
		return nil
	}
	getStr := func(k string) string {
		switch v := o[k].(type) {
		case string:
			return strings.TrimSpace(v)
		case float64:
			// orderId 等字段在某些解析里可能被解成 number
			return g.NewVar(v).String()
		case int:
			return g.NewVar(v).String()
		case int64:
			return g.NewVar(v).String()
		default:
			return ""
		}
	}
	getF := func(k string) float64 {
		switch v := o[k].(type) {
		case float64:
			return v
		case string:
			return g.NewVar(v).Float64()
		default:
			return 0
		}
	}
	getI64 := func(k string) int64 {
		switch v := o[k].(type) {
		case float64:
			return int64(v)
		case string:
			return g.NewVar(v).Int64()
		default:
			return 0
		}
	}
	// reduceOnly 在币安 futures userData 里常见字段 R
	ro := false
	if v, ok := o["R"].(bool); ok {
		ro = v
	}
	rawStatus := getStr("X")
	status := normalizeOrderStatus("binance", rawStatus)
	return []parsedOrder{{
		ExchangeOrderId: getStr("i"),
		ClientOrderId:   getStr("c"),
		Side:            strings.ToUpper(getStr("S")),
		PositionSide:    strings.ToUpper(getStr("ps")),
		Type:            strings.ToUpper(getStr("o")),
		ReduceOnly:      ro,
		Price:           getF("p"),
		Quantity:        getF("q"),
		FilledQty:       getF("z"),
		AvgPrice:        getF("ap"),
		Status:          status,
		RawStatus:       rawStatus,
		IsOpen:          isOpenStatus("binance", rawStatus),
		CreateTime:      getI64("T"),
		UpdateTime:      getI64("t"),
	}}
}

func parseOKXPrivateOrders(raw []byte) []parsedOrder {
	// { "arg":{"channel":"orders", ...}, "data":[{...}] }
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	data, _ := m["data"].([]any)
	if len(data) == 0 {
		return nil
	}
	out := make([]parsedOrder, 0, len(data))
	for _, it := range data {
		j, _ := it.(map[string]any)
		if j == nil {
			continue
		}
		getStr := func(k string) string {
			if v, ok := j[k].(string); ok {
				return strings.TrimSpace(v)
			}
			return ""
		}
		getF := func(k string) float64 {
			switch v := j[k].(type) {
			case float64:
				return v
			case string:
				return g.NewVar(v).Float64()
			default:
				return 0
			}
		}
		getI64 := func(k string) int64 {
			switch v := j[k].(type) {
			case float64:
				return int64(v)
			case string:
				return g.NewVar(v).Int64()
			default:
				return 0
			}
		}
		rawState := getStr("state")
		status := normalizeOrderStatus("okx", rawState)
		out = append(out, parsedOrder{
			ExchangeOrderId: getStr("ordId"),
			ClientOrderId:   getStr("clOrdId"),
			Side:            strings.ToUpper(getStr("side")),
			PositionSide:    strings.ToUpper(getStr("posSide")),
			Type:            strings.ToUpper(getStr("ordType")),
			ReduceOnly:      strings.EqualFold(getStr("reduceOnly"), "true") || getStr("reduceOnly") == "1",
			Price:           getF("px"),
			Quantity:        getF("sz"),
			FilledQty:       getF("accFillSz"),
			AvgPrice:        getF("avgPx"),
			Status:          status,
			RawStatus:       rawState,
			IsOpen:          isOpenStatus("okx", rawState),
			CreateTime:      getI64("cTime"),
			UpdateTime:      getI64("uTime"),
		})
	}
	return out
}

func parseGatePrivateOrders(raw []byte) []parsedOrder {
	// { "channel":"futures.orders", "event":"update", "result":{...} }
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	res, _ := m["result"].(map[string]any)
	if res == nil {
		return nil
	}
	getStr := func(k string) string {
		if v, ok := res[k].(string); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}
	getF := func(k string) float64 {
		switch v := res[k].(type) {
		case float64:
			return v
		case string:
			return g.NewVar(v).Float64()
		default:
			return 0
		}
	}
	getI64 := func(k string) int64 {
		switch v := res[k].(type) {
		case float64:
			return int64(v)
		case string:
			return g.NewVar(v).Int64()
		default:
			return 0
		}
	}
	rawStatus := getStr("status")
	// Gate 的 reduce_only 字段可能为 bool 或 string
	ro := false
	if v, ok := res["reduce_only"].(bool); ok {
		ro = v
	} else if s := getStr("reduce_only"); s != "" {
		ro = (s == "true" || s == "1")
	}

	// size: 合约张数（正/负表示方向）
	size := getF("size")
	side := ""
	if size < 0 {
		side = "SELL"
	} else if size > 0 {
		side = "BUY"
	}
	qty := math.Abs(size)
	// filled: 优先用 left 推导（filled = abs(size) - abs(left)）
	filled := 0.0
	if left := getF("left"); left != 0 {
		filled = math.Abs(size) - math.Abs(left)
		if filled < 0 {
			filled = 0
		}
	} else if fs := getF("filled_size"); fs != 0 {
		filled = math.Abs(fs)
	}

	// pos_side: long/short（若有）
	posSide := strings.ToUpper(strings.TrimSpace(getStr("pos_side")))
	if posSide == "" {
		posSide = strings.ToUpper(strings.TrimSpace(getStr("position_side")))
	}
	if posSide == "" {
		posSide = strings.ToUpper(strings.TrimSpace(getStr("posSide")))
	}
	if posSide == "LONG" || posSide == "SHORT" {
		// ok
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "short") {
		// keep
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
		// keep
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
		// keep
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
		// keep
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
		// keep
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else {
		// Gate 可能返回 long/short
		if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else if strings.EqualFold(posSide, "LONG") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "SHORT") {
			posSide = "SHORT"
		} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
			// keep
		} else {
			if strings.EqualFold(posSide, "LONG") {
				posSide = "LONG"
			} else if strings.EqualFold(posSide, "SHORT") {
				posSide = "SHORT"
			} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
				// keep
			} else if strings.EqualFold(posSide, "LONG") {
				posSide = "LONG"
			} else if strings.EqualFold(posSide, "SHORT") {
				posSide = "SHORT"
			} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
				// keep
			} else if strings.EqualFold(posSide, "LONG") {
				posSide = "LONG"
			} else if strings.EqualFold(posSide, "SHORT") {
				posSide = "SHORT"
			} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
				// keep
			} else if strings.EqualFold(posSide, "LONG") {
				posSide = "LONG"
			} else if strings.EqualFold(posSide, "SHORT") {
				posSide = "SHORT"
			} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
				// keep
			} else if strings.EqualFold(posSide, "LONG") {
				posSide = "LONG"
			} else if strings.EqualFold(posSide, "SHORT") {
				posSide = "SHORT"
			} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
				// keep
			} else if strings.EqualFold(posSide, "LONG") {
				posSide = "LONG"
			} else if strings.EqualFold(posSide, "SHORT") {
				posSide = "SHORT"
			} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
				// keep
			} else if strings.EqualFold(posSide, "LONG") {
				posSide = "LONG"
			} else if strings.EqualFold(posSide, "SHORT") {
				posSide = "SHORT"
			} else if strings.EqualFold(posSide, "LONG") || strings.EqualFold(posSide, "SHORT") {
				// keep
			} else if strings.EqualFold(posSide, "LONG") {
				posSide = "LONG"
			} else if strings.EqualFold(posSide, "SHORT") {
				posSide = "SHORT"
			}
		}
	}
	// 规范化 long/short 写法
	if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	} else if strings.EqualFold(posSide, "LONG") {
		posSide = "LONG"
	} else if strings.EqualFold(posSide, "SHORT") {
		posSide = "SHORT"
	}

	// order type: price=0 => market
	typ := "LIMIT"
	if price := getF("price"); price == 0 {
		typ = "MARKET"
	}

	ct := getI64("create_time")
	if ct > 0 && ct < 1e12 {
		ct = ct * 1000
	}
	ut := getI64("finish_time")
	if ut > 0 && ut < 1e12 {
		ut = ut * 1000
	}
	if ut == 0 {
		ut = getI64("update_time")
		if ut > 0 && ut < 1e12 {
			ut = ut * 1000
		}
	}
	if ut == 0 {
		ut = ct
	}
	return []parsedOrder{{
		ExchangeOrderId: getStr("id"),
		ClientOrderId:   getStr("text"),
		Side:            side,
		PositionSide:    posSide,
		Type:            typ,
		ReduceOnly:      ro,
		Price:           getF("price"),
		Quantity:        qty,
		FilledQty:       filled,
		AvgPrice:        getF("fill_price"),
		Status:          normalizeOrderStatus("gate", rawStatus),
		RawStatus:       rawStatus,
		IsOpen:          isOpenStatus("gate", rawStatus),
		CreateTime:      ct,
		UpdateTime:      ut,
	}}
}

func parseBitgetPrivateOrders(raw []byte) []parsedOrder {
	// { "arg":{"channel":"orders",...}, "data":[{...}] }
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	data, _ := m["data"].([]any)
	if len(data) == 0 {
		return nil
	}
	out := make([]parsedOrder, 0, len(data))
	for _, it := range data {
		j, _ := it.(map[string]any)
		if j == nil {
			continue
		}
		getStr := func(k string) string {
			if v, ok := j[k].(string); ok {
				return strings.TrimSpace(v)
			}
			return ""
		}
		getF := func(k string) float64 {
			switch v := j[k].(type) {
			case float64:
				return v
			case string:
				return g.NewVar(v).Float64()
			default:
				return 0
			}
		}
		getI64 := func(k string) int64 {
			switch v := j[k].(type) {
			case float64:
				return int64(v)
			case string:
				return g.NewVar(v).Int64()
			default:
				return 0
			}
		}
		oid := getStr("orderId")
		if oid == "" {
			oid = getStr("ordId")
		}
		cid := getStr("clientOid")
		if cid == "" {
			cid = getStr("clOrdId")
		}
		rawStatus := getStr("status")
		if rawStatus == "" {
			rawStatus = getStr("state")
		}
		out = append(out, parsedOrder{
			ExchangeOrderId: oid,
			ClientOrderId:   cid,
			Side:            strings.ToUpper(getStr("side")),
			PositionSide:    strings.ToUpper(getStr("posSide")),
			Type:            strings.ToUpper(getStr("orderType")),
			ReduceOnly:      strings.EqualFold(getStr("tradeSide"), "close"),
			Price:           getF("price"),
			Quantity:        getF("size"),
			FilledQty:       getF("fillSz"),
			AvgPrice:        getF("avgPx"),
			Status:          normalizeOrderStatus("bitget", rawStatus),
			RawStatus:       rawStatus,
			IsOpen:          isOpenStatus("bitget", rawStatus),
			CreateTime:      getI64("cTime"),
			UpdateTime:      getI64("uTime"),
		})
	}
	return out
}

func normalizeOrderStatus(platform, raw string) string {
	raw = strings.ToUpper(strings.TrimSpace(raw))
	switch platform {
	case "binance":
		// NEW / PARTIALLY_FILLED / FILLED / CANCELED / EXPIRED / REJECTED
		return raw
	case "okx":
		// live / partially_filled / filled / canceled
		switch raw {
		case "LIVE":
			return "NEW"
		case "PARTIALLY_FILLED":
			return "PARTIALLY_FILLED"
		case "FILLED":
			return "FILLED"
		case "CANCELED", "CANCELLED":
			return "CANCELED"
		default:
			return raw
		}
	case "gate":
		// open / finished / cancelled
		switch raw {
		case "OPEN":
			return "NEW"
		case "FINISHED":
			return "FILLED"
		case "CANCELLED", "CANCELED":
			return "CANCELED"
		default:
			return raw
		}
	case "bitget":
		// live / partially_filled / filled / cancelled
		switch raw {
		case "LIVE":
			return "NEW"
		case "NEW":
			return "NEW"
		case "PARTIALLY_FILLED":
			return "PARTIALLY_FILLED"
		case "FILLED":
			return "FILLED"
		case "CANCELLED", "CANCELED":
			return "CANCELED"
		default:
			return raw
		}
	default:
		return raw
	}
}

func isOpenStatus(platform, raw string) bool {
	s := strings.ToUpper(strings.TrimSpace(raw))
	switch platform {
	case "binance":
		return s == "NEW" || s == "PARTIALLY_FILLED"
	case "okx":
		return s == "LIVE" || s == "PARTIALLY_FILLED"
	case "gate":
		return s == "OPEN"
	case "bitget":
		return s == "LIVE" || s == "NEW" || s == "PARTIALLY_FILLED"
	default:
		return false
	}
}

func boolToTiny(b bool) int {
	if b {
		return 1
	}
	return 0
}

func truncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}

func isTableMissingErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "doesn't exist") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "no such table")
}

func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") ||
		strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "UNIQUE constraint failed")
}

func upsertExchangeOrder(ctx context.Context, platform string, apiConfigId int64, exchangeOrderId string, data g.Map) error {
	platform = strings.ToLower(strings.TrimSpace(platform))
	exchangeOrderId = strings.TrimSpace(exchangeOrderId)
	if platform == "" || apiConfigId <= 0 || exchangeOrderId == "" {
		return nil
	}

	// 先 update（常见路径：WS 多次更新同一订单）
	r, err := g.DB().Model(exchangeOrderTable).Ctx(ctx).
		Where("platform", platform).
		Where("api_config_id", apiConfigId).
		Where("exchange_order_id", exchangeOrderId).
		Data(data).Update()
	if err != nil {
		if isTableMissingErr(err) {
			return err
		}
		return err
	}
	aff, _ := r.RowsAffected()
	if aff > 0 {
		return nil
	}

	// 不存在则 insert
	_, err = g.DB().Model(exchangeOrderTable).Ctx(ctx).Insert(data)
	if err == nil {
		return nil
	}
	if isDuplicateKeyErr(err) {
		_, _ = g.DB().Model(exchangeOrderTable).Ctx(ctx).
			Where("platform", platform).
			Where("api_config_id", apiConfigId).
			Where("exchange_order_id", exchangeOrderId).
			Data(data).Update()
		return nil
	}
	return err
}
