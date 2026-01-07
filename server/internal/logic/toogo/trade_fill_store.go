package toogo

import (
	"context"
	"fmt"
	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"math"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type tradeFillOrderLink struct {
	OrderId   int64
	UserId    int64
	RobotId   int64
	OrderSn   string
	Exchange  string
	Symbol    string
	Direction string
	// 下面字段用于 Gate 等“成交明细不稳定提供已实现盈亏”的场景：用本地订单口径回填/分摊
	ExchangeOrderId string
	CloseOrderId    string
	RealizedProfit  float64
	OpenTime        *gtime.Time
	CloseTime       *gtime.Time
	// IsCloseKey: 该 link 是否由 close_order_id 匹配产生（即当前 trade.order_id == 本地 close_order_id）
	IsCloseKey bool
}

func normalizeTsMs(ts int64) int64 {
	if ts <= 0 {
		return 0
	}
	// 秒级
	if ts < 1e12 {
		return ts * 1000
	}
	return ts
}

func normalizeUpper(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

type tradeFillFallbackOwner struct {
	UserId  int64
	RobotId int64
}

// upsertTradeFillsFromTrades 将交易所成交(Trade/Fill)落库到 hg_trading_trade_fill。
//
// 关键点：
// - 用 (api_config_id, exchange, trade_id) 做去重（Save/ON DUPLICATE）
// - robot/user 通过 trade.OrderId 去匹配本地 hg_trading_order(exchange_order_id/close_order_id)
// - fee 统一按正数存（abs），避免前端/汇总口径混乱
func upsertTradeFillsFromTrades(ctx context.Context, apiConfigId int64, exchangeName string, symbol string, trades []*exchange.Trade, sessionId *int64) (saved int, matched int, err error) {
	if apiConfigId <= 0 {
		return 0, 0, gerror.New("apiConfigId is required")
	}
	// 统一交易所标识口径：全部使用小写 platform（binance/okx/bitget/gate）
	// 多交易所场景下，如果这里混入 robot.Exchange 的脏数据（大小写/别名），会导致：
	// - PG upsert 的冲突键(api_config_id, exchange, trade_id)不命中 → 重复插入/唯一键报错
	// - 订单/成交关联匹配困难（exchange 维度不一致）
	exchangeName = strings.ToLower(strings.TrimSpace(exchangeName))
	symbol = strings.TrimSpace(symbol)
	if len(trades) == 0 {
		return 0, 0, nil
	}

	// 收集 orderId
	orderIDs := make([]string, 0, len(trades))
	orderSeen := make(map[string]bool)
	for _, t := range trades {
		if t == nil {
			continue
		}
		oid := strings.TrimSpace(t.OrderId)
		if oid == "" || orderSeen[oid] {
			continue
		}
		orderSeen[oid] = true
		orderIDs = append(orderIDs, oid)
	}

	// orderId -> 本地订单候选映射（允许一对多，后续按时间/口径选择最佳候选）
	orderLinks := make(map[string][]*tradeFillOrderLink)
	if len(orderIDs) > 0 {
		// 注意：dao.TradingOrder.Columns() 可能缺字段，统一用字符串列名
		type row struct {
			Id              int64  `orm:"id"`
			UserId          int64  `orm:"user_id"`
			RobotId         int64  `orm:"robot_id"`
			OrderSn         string `orm:"order_sn"`
			ExchangeOrderId string `orm:"exchange_order_id"`
			CloseOrderId    string `orm:"close_order_id"`
			Exchange        string `orm:"exchange"`
			Symbol          string `orm:"symbol"`
			Direction       string `orm:"direction"`
			RealizedProfit  float64 `orm:"realized_profit"`
			OpenTime        *gtime.Time `orm:"open_time"`
			CloseTime       *gtime.Time `orm:"close_time"`
		}
		var rows []*row
		q := dao.TradingOrder.Ctx(ctx).
			Fields("id", "user_id", "robot_id", "order_sn", "exchange_order_id", "close_order_id", "exchange", "symbol", "direction", "realized_profit", "open_time", "close_time").
			Where("exchange_order_id IN (?) OR close_order_id IN (?)", orderIDs, orderIDs)
		if err := q.Scan(&rows); err != nil {
			return 0, 0, gerror.Wrap(err, "query trading_order for trade fill mapping failed")
		}
		for _, r := range rows {
			if r == nil {
				continue
			}
			base := &tradeFillOrderLink{
				OrderId:   r.Id,
				UserId:    r.UserId,
				RobotId:   r.RobotId,
				OrderSn:   r.OrderSn,
				Exchange:  r.Exchange,
				Symbol:    r.Symbol,
				Direction: r.Direction,
				ExchangeOrderId: strings.TrimSpace(r.ExchangeOrderId),
				CloseOrderId:    strings.TrimSpace(r.CloseOrderId),
				RealizedProfit:  r.RealizedProfit,
				OpenTime:        r.OpenTime,
				CloseTime:       r.CloseTime,
			}
			if base.ExchangeOrderId != "" {
				l := *base
				l.IsCloseKey = false
				orderLinks[base.ExchangeOrderId] = append(orderLinks[base.ExchangeOrderId], &l)
			}
			if base.CloseOrderId != "" {
				l := *base
				l.IsCloseKey = true
				orderLinks[base.CloseOrderId] = append(orderLinks[base.CloseOrderId], &l)
			}
		}
	}

	// Gate 专用：订单级已实现盈亏分摊（高效/稳定）
	// - 订单关联可能出现“一对多”（历史脏数据/多来源写入），这里不依赖 close_order_id 唯一性
	// - 先按 order_id 汇总本批 trades 的总成交数量，用于分摊
	closeOrderTotalQty := make(map[string]float64) // orderId -> sumQty (base qty)
	if exchangeName == "gate" {
		for _, t := range trades {
			if t == nil {
				continue
			}
			oid := strings.TrimSpace(t.OrderId)
			if oid == "" || t.Quantity <= 0 {
				continue
			}
			closeOrderTotalQty[oid] += t.Quantity
		}
	}

	absInt64 := func(x int64) int64 {
		if x < 0 {
			return -x
		}
		return x
	}
	absFloat64 := func(x float64) float64 {
		if x < 0 {
			return -x
		}
		return x
	}
	// chooseBestLink: 从多个候选本地订单中挑最可能匹配该成交的一个
	// 规则（按优先级）：
	// 1) 优先 IsCloseKey=true（trade.order_id 命中本地 close_order_id）
	// 2) 优先 abs(realized_profit) 更大（平仓单更容易有非0 realized_profit）
	// 3) 优先成交时间更接近（平仓候选看 close_time；开仓候选看 open_time）
	chooseBestLink := func(cands []*tradeFillOrderLink, tradeTsMs int64) *tradeFillOrderLink {
		if len(cands) == 0 {
			return nil
		}
		var best *tradeFillOrderLink
		var bestDt int64 = 1<<62
		for _, c := range cands {
			if c == nil {
				continue
			}
			if best == nil {
				best = c
				bestDt = 1<<62
				if tradeTsMs > 0 {
					if c.IsCloseKey && c.CloseTime != nil && !c.CloseTime.IsZero() {
						bestDt = absInt64(tradeTsMs - c.CloseTime.UnixMilli())
					} else if c.OpenTime != nil && !c.OpenTime.IsZero() {
						bestDt = absInt64(tradeTsMs - c.OpenTime.UnixMilli())
					}
				}
				continue
			}
			// 1) IsCloseKey
			if c.IsCloseKey != best.IsCloseKey {
				if c.IsCloseKey {
					best = c
					if tradeTsMs > 0 && c.CloseTime != nil && !c.CloseTime.IsZero() {
						bestDt = absInt64(tradeTsMs - c.CloseTime.UnixMilli())
					} else {
						bestDt = 1<<62
					}
				}
				continue
			}
			// 2) abs(realized_profit)
			ab := absFloat64(best.RealizedProfit)
			ac := absFloat64(c.RealizedProfit)
			if ac != ab {
				if ac > ab {
					best = c
					if tradeTsMs > 0 {
						if c.IsCloseKey && c.CloseTime != nil && !c.CloseTime.IsZero() {
							bestDt = absInt64(tradeTsMs - c.CloseTime.UnixMilli())
						} else if c.OpenTime != nil && !c.OpenTime.IsZero() {
							bestDt = absInt64(tradeTsMs - c.OpenTime.UnixMilli())
						} else {
							bestDt = 1<<62
						}
					}
				}
				continue
			}
			// 3) time distance
			if tradeTsMs > 0 {
				dt := int64(1 << 62)
				if c.IsCloseKey && c.CloseTime != nil && !c.CloseTime.IsZero() {
					dt = absInt64(tradeTsMs - c.CloseTime.UnixMilli())
				} else if c.OpenTime != nil && !c.OpenTime.IsZero() {
					dt = absInt64(tradeTsMs - c.OpenTime.UnixMilli())
				}
				if dt < bestDt {
					best = c
					bestDt = dt
				}
			}
		}
		return best
	}

	// fallback owner：用于解决“成交落库发生在 close_order_id 写入之前”导致无法通过订单ID匹配 owner 的情况。
	// 【多交易所/多机器人】同一 api_config_id 可能被多个机器人复用（不同symbol/不同策略），此时不能再“随便取一个机器人”做兜底，
	// 否则会把成交错误归属到其他机器人/用户。
	// 新策略：
	// - 若能通过 order_id 关联到本地订单：用关联结果（最准确）
	// - 若无法关联：仅当 (api_config_id + symbol) 能唯一定位到一个机器人时才兜底归属；否则不填 owner（user_id/robot_id=0）
	var fallback tradeFillFallbackOwner
	{
		type rb struct {
			Id     int64  `orm:"id"`
			UserId int64  `orm:"user_id"`
			Symbol string `orm:"symbol"`
		}
		var rs []*rb
		q := dao.TradingRobot.Ctx(ctx).
			Fields("id", "user_id", "symbol").
			Where("api_config_id", apiConfigId).
			WhereNull("deleted_at")
		if symbol != "" {
			q = q.Where("symbol", symbol)
		}
		_ = q.Limit(3).Scan(&rs)
		if len(rs) == 1 && rs[0] != nil && rs[0].Id > 0 && rs[0].UserId > 0 {
			fallback.RobotId = rs[0].Id
			fallback.UserId = rs[0].UserId
		}
	}

	now := gtime.Now()
	data := make([]g.Map, 0, len(trades))
	// 事件驱动：成交落库后，触发一次“当前运行区间”汇总写回（按 trade_fill 时间窗口径）
	// 只依赖 userId+robotId，避免 exchange/symbol 历史数据不一致导致找不到 run_session
	type rsKey struct {
		UserId  int64
		RobotId int64
	}
	rsKeys := make(map[string]*rsKey)
	for _, t := range trades {
		if t == nil {
			continue
		}
		oid := strings.TrimSpace(t.OrderId)
		if oid == "" {
			continue
		}
		tsMs := normalizeTsMs(t.Time)
		tradeId := strings.TrimSpace(t.TradeId)
		if tradeId == "" {
			// 兜底生成一个可去重的 trade_id（避免空 trade_id 导致唯一键失效）
			tradeId = fmt.Sprintf("%s-%d-%.8f-%.8f-%s", oid, tsMs, t.Price, t.Quantity, normalizeUpper(t.Side))
		}
		// symbol 优先使用调用方传入（通常更可信/已规范化），否则回退 t.Symbol
		sym := symbol
		if sym == "" {
			sym = strings.TrimSpace(t.Symbol)
		}

		link := chooseBestLink(orderLinks[oid], tsMs)
		uid := int64(0)
		rid := int64(0)
		if link != nil {
			uid = link.UserId
			rid = link.RobotId
			matched++
		} else if fallback.UserId > 0 && fallback.RobotId > 0 {
			uid = fallback.UserId
			rid = fallback.RobotId
		}

		fee := math.Abs(t.Commission)
		realizedPnl := t.RealizedPnl
		// Gate 专用：用本地订单 realized_profit 分摊回填成交流水 realized_pnl（仅平仓订单ID）
		if exchangeName == "gate" && realizedPnl == 0 && link != nil && link.IsCloseKey {
			if link.RealizedProfit != 0 {
				totalQty := closeOrderTotalQty[oid]
				if totalQty > 0 && t.Quantity > 0 {
					realizedPnl = link.RealizedProfit * (t.Quantity / totalQty)
				}
			}
		}

		data = append(data, g.Map{
			"tenant_id":       0,
			"user_id":         uid,
			"robot_id":        rid,
			"session_id":      sessionId,
			"api_config_id":   apiConfigId,
			"exchange":        exchangeName,
			"symbol":          sym,
			"order_id":        oid,
			"client_order_id": "",
			"trade_id":        tradeId,
			"side":            normalizeUpper(t.Side),
			"price":           t.Price,
			"qty":             t.Quantity,
			"realized_pnl":    realizedPnl,
			"fee":             fee,
			"fee_coin":        strings.TrimSpace(t.CommissionAsset),
			"ts":              tsMs,
			"created_at":      now,
			"updated_at":      now,
		})

		// collect unique keys for run-session refresh
		if uid > 0 && rid > 0 {
			k := g.NewVar(uid).String() + ":" + g.NewVar(rid).String()
			if _, ok := rsKeys[k]; !ok {
				rsKeys[k] = &rsKey{UserId: uid, RobotId: rid}
			}
		}
	}

	if len(data) == 0 {
		return 0, matched, nil
	}
	// MySQL: Save() 会生成 INSERT ... ON DUPLICATE KEY UPDATE（可命中任意 unique key）
	// PG: 需要显式 ON CONFLICT(uk) ...，否则可能重复插入/或违反唯一约束
	if dao.TradingTradeFill.DB().GetConfig().Type == consts.DBPgsql {
		if err := upsertTradingTradeFillPg(ctx, data); err != nil {
			return 0, matched, err
		}
	} else {
	_, err = dao.TradingTradeFill.Ctx(ctx).Data(data).Save()
	if err != nil {
		return 0, matched, gerror.Wrap(err, "save trading_trade_fill failed")
		}
	}

	// 触发 run_session 汇总写回（异步 + 去抖）
	if len(rsKeys) > 0 {
		for _, k := range rsKeys {
			if k == nil {
				continue
			}
			triggerRunSessionRefreshByTradeFill(ctx, k.UserId, k.RobotId)
		}
	}
	return len(data), matched, nil
}

// upsertTradingTradeFillPg 在 PG 上按 (api_config_id, exchange, trade_id) 进行 upsert。
// 依赖数据库中存在对应的唯一约束/唯一索引（Navicat 迁移 unique key 时通常会映射为 unique index/constraint）。
func upsertTradingTradeFillPg(ctx context.Context, rows []g.Map) error {
	if len(rows) == 0 {
		return nil
	}

	// 注意：PG 的 ON CONFLICT DO UPDATE 需要指定冲突列，这里与 MySQL 的 uk_api_exchange_trade 对齐
	const sql = `
INSERT INTO hg_trading_trade_fill
  (tenant_id, user_id, robot_id, session_id, api_config_id, exchange, symbol, order_id, client_order_id, trade_id, side, price, qty, realized_pnl, fee, fee_coin, ts, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (api_config_id, exchange, trade_id)
DO UPDATE SET
  tenant_id       = EXCLUDED.tenant_id,
  user_id         = EXCLUDED.user_id,
  robot_id        = EXCLUDED.robot_id,
  session_id      = EXCLUDED.session_id,
  symbol          = EXCLUDED.symbol,
  order_id        = EXCLUDED.order_id,
  client_order_id = EXCLUDED.client_order_id,
  side            = EXCLUDED.side,
  price           = EXCLUDED.price,
  qty             = EXCLUDED.qty,
  realized_pnl    = EXCLUDED.realized_pnl,
  fee            = EXCLUDED.fee,
  fee_coin        = EXCLUDED.fee_coin,
  ts              = EXCLUDED.ts,
  updated_at      = EXCLUDED.updated_at
`

	return dao.TradingTradeFill.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, r := range rows {
			if r == nil {
				continue
			}
			// GoFrame gdb.TX.Exec 不带 ctx 参数（ctx 已由 Transaction 传入并绑定到 tx）
			_, err := tx.Exec(sql,
				r["tenant_id"],
				r["user_id"],
				r["robot_id"],
				r["session_id"],
				r["api_config_id"],
				r["exchange"],
				r["symbol"],
				r["order_id"],
				r["client_order_id"],
				r["trade_id"],
				r["side"],
				r["price"],
				r["qty"],
				r["realized_pnl"],
				r["fee"],
				r["fee_coin"],
				r["ts"],
				r["created_at"],
				r["updated_at"],
			)
			if err != nil {
				return gerror.Wrap(err, "pgsql upsert trading_trade_fill failed")
			}
		}
		return nil
	})
}

// fetchAndStoreTradeHistory 拉取交易所成交历史并落库。
// 这是“离线/后台同步”与“兜底补齐”的基础能力。
func fetchAndStoreTradeHistory(ctx context.Context, ex exchange.Exchange, apiConfigId int64, exchangeName string, symbol string, limit int) (saved int, matched int, err error) {
	type tradeHistoryProvider interface {
		GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*exchange.Trade, error)
	}
	p, ok := ex.(tradeHistoryProvider)
	if !ok {
		return 0, 0, gerror.New("exchange does not support trade history")
	}
	if limit <= 0 {
		limit = 500
	}
	trades, err := p.GetTradeHistory(ctx, symbol, limit)
	if err != nil {
		return 0, 0, err
	}
	return upsertTradeFillsFromTrades(ctx, apiConfigId, exchangeName, symbol, trades, nil)
}
