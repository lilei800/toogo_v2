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

	// orderId -> 本地订单映射
	orderLink := make(map[string]*tradeFillOrderLink)
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
		}
		var rows []*row
		q := dao.TradingOrder.Ctx(ctx).
			Fields("id", "user_id", "robot_id", "order_sn", "exchange_order_id", "close_order_id", "exchange", "symbol", "direction").
			Where("exchange_order_id IN (?) OR close_order_id IN (?)", orderIDs, orderIDs)
		if err := q.Scan(&rows); err != nil {
			return 0, 0, gerror.Wrap(err, "query trading_order for trade fill mapping failed")
		}
		for _, r := range rows {
			if r == nil {
				continue
			}
			link := &tradeFillOrderLink{
				OrderId:   r.Id,
				UserId:    r.UserId,
				RobotId:   r.RobotId,
				OrderSn:   r.OrderSn,
				Exchange:  r.Exchange,
				Symbol:    r.Symbol,
				Direction: r.Direction,
			}
			if strings.TrimSpace(r.ExchangeOrderId) != "" {
				orderLink[strings.TrimSpace(r.ExchangeOrderId)] = link
			}
			if strings.TrimSpace(r.CloseOrderId) != "" {
				orderLink[strings.TrimSpace(r.CloseOrderId)] = link
			}
		}
	}

	// fallback owner：用于解决“成交落库发生在 close_order_id 写入之前”导致无法通过订单ID匹配 owner 的情况。
	// 在本系统规则下：一个 api_config_id 通常只会被一个机器人使用（且机器人所属 user 唯一），因此可作为兜底 owner。
	var fallback tradeFillFallbackOwner
	{
		type rb struct {
			Id     int64 `orm:"id"`
			UserId int64 `orm:"user_id"`
		}
		var r rb
		_ = dao.TradingRobot.Ctx(ctx).
			Fields("id", "user_id").
			Where("api_config_id", apiConfigId).
			WhereNull("deleted_at").
			Limit(1).
			Scan(&r)
		if r.Id > 0 && r.UserId > 0 {
			fallback.RobotId = r.Id
			fallback.UserId = r.UserId
		}
	}

	now := gtime.Now()
	data := make([]g.Map, 0, len(trades))
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

		link := orderLink[oid]
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

		data = append(data, g.Map{
			"tenant_id":       0,
			"user_id":         uid,
			"robot_id":        rid,
			"session_id":      sessionId,
			"api_config_id":   apiConfigId,
			"exchange":        strings.TrimSpace(exchangeName),
			"symbol":          sym,
			"order_id":        oid,
			"client_order_id": "",
			"trade_id":        tradeId,
			"side":            normalizeUpper(t.Side),
			"price":           t.Price,
			"qty":             t.Quantity,
			"realized_pnl":    t.RealizedPnl,
			"fee":             fee,
			"fee_coin":        strings.TrimSpace(t.CommissionAsset),
			"ts":              tsMs,
			"created_at":      now,
			"updated_at":      now,
		})
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
