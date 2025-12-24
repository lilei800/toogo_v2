// Package toogo
// @Description 运行区间盈亏“事件驱动”实时汇总（尽量不打交易所接口）
package toogo

import (
	"context"
	"strings"

	"hotgo/internal/dao"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// refreshCurrentRunSessionSummaryByRobot 根据“当前运行区间 + 本地订单表(已同步交易所口径)”实时刷新汇总
//
// 设计目标：
// - 触发点：自动平仓/手动平仓/开仓手续费补齐等“订单事件”
// - 不依赖交易所 API（避免频繁 GetTradeHistory 消耗资源/限流）
// - 幂等：采用“重算”而不是“累加”，避免重复触发导致重复统计
func refreshCurrentRunSessionSummaryByRobot(ctx context.Context, userId, robotId int64, exchange, symbol string) {
	if userId <= 0 || robotId <= 0 {
		return
	}
	exchange = strings.TrimSpace(exchange)
	symbol = strings.TrimSpace(symbol)
	if exchange == "" || symbol == "" {
		return
	}

	// 找到当前运行区间（end_time 为空）
	var sess *entity.TradingRobotRunSession
	_ = dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().UserId, userId).
		Where(dao.TradingRobotRunSession.Columns().RobotId, robotId).
		Where(dao.TradingRobotRunSession.Columns().Exchange, exchange).
		Where(dao.TradingRobotRunSession.Columns().Symbol, symbol).
		WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
		OrderDesc(dao.TradingRobotRunSession.Columns().Id).
		Limit(1).
		Scan(&sess)
	if sess == nil || sess.StartTime == nil || sess.StartTime.IsZero() {
		return
	}

	now := gtime.Now()

	// 从本地订单表重算（订单表字段由平仓时的交易所成交汇总写入/补齐）
	// 统计口径：
	// - total_pnl：sum(realized_profit)（仅 status=2 已平仓）
	// - total_fee：sum(open_fee+close_fee)（仅当 fee coin 为 USDT/空时计入；避免币种不一致）
	// - trade_count：此处按“平仓订单数”统计（更稳定）；成交(fill)笔数若需要可走强制同步接口补齐
	type aggRow struct {
		Pnl   float64 `json:"pnl"   orm:"pnl"`
		Fee   float64 `json:"fee"   orm:"fee"`
		Count int     `json:"count" orm:"count"`
	}
	var agg aggRow

	// 注意：字段名在不同环境可能存在（历史迁移），这里用 SQL 表达式并容错 0 值
	// fee 仅统计 USDT（或空）口径
	//
	// 说明：
	// - MySQL 的 IF/IFNULL 在 PostgreSQL 不可用，因此统一用 CASE/COALESCE
	// - UPPER/IS NULL/'' 判断在两边都可用
	feeExpr := `
SUM(
  CASE WHEN (open_fee_coin IS NULL OR open_fee_coin='' OR UPPER(open_fee_coin)='USDT') THEN open_fee ELSE 0 END
  +
  CASE WHEN (close_fee_coin IS NULL OR close_fee_coin='' OR UPPER(close_fee_coin)='USDT') THEN close_fee ELSE 0 END
) AS fee`

	_ = dao.TradingOrder.Ctx(ctx).
		Fields("COALESCE(SUM(realized_profit),0) AS pnl", feeExpr, "COUNT(1) AS count").
		Where("user_id", userId).
		Where("robot_id", robotId).
		Where("exchange", exchange).
		Where("symbol", symbol).
		Where("status", OrderStatusClosed).
		WhereGTE("close_time", sess.StartTime).
		WhereLTE("close_time", now).
		Scan(&agg)

	// 写回区间汇总
	// total_pnl/total_fee 允许为 0（有效值），使用指针字段以便前端能区分"未同步/无数据"
	pnl := agg.Pnl
	fee := agg.Fee
	_, err := dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().Id, sess.Id).
		Data(g.Map{
			"total_pnl":   pnl,
			"total_fee":   fee,
			"trade_count": agg.Count,
			"synced_at":   now,
			"updated_at":  now,
		}).
		Update()
	if err != nil {
		g.Log().Warningf(ctx, "[RunSession] 刷新区间汇总失败: sessionId=%d, robotId=%d, err=%v", sess.Id, robotId, err)
	}
}
