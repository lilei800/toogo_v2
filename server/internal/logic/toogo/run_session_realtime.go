// Package toogo
// @Description 运行区间盈亏“事件驱动”实时汇总（尽量不打交易所接口）
package toogo

import (
	"context"
	"sync"
	"time"

	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// refreshCurrentRunSessionSummaryByRobot 根据“当前运行区间 + 成交流水(trading_trade_fill)”实时刷新汇总（写回 run_session）。
//
// 设计目标：
// - 触发点：自动平仓/手动平仓/开仓手续费补齐等“订单事件”
// - 不依赖交易所 API（避免频繁 GetTradeHistory 消耗资源/限流）
// - 幂等：采用“重算”而不是“累加”，避免重复触发导致重复统计
func refreshCurrentRunSessionSummaryByRobot(ctx context.Context, userId, robotId int64, exchange, symbol string) {
	// 为保持口径一致，这里统一用 trade_fill 口径重算（而不是订单表口径）。
	refreshCurrentRunSessionSummaryByTradeFill(ctx, userId, robotId)
}

// ====== 成交流水(trade_fill)事件驱动写回 run_session ======

var runSessionTradeFillDebounce sync.Map // key -> time.Time

func shouldRunSessionRefresh(key string, interval time.Duration) bool {
	if key == "" {
		return false
	}
	now := time.Now()
	if v, ok := runSessionTradeFillDebounce.Load(key); ok {
		if t, ok2 := v.(time.Time); ok2 {
			if now.Sub(t) < interval {
				return false
			}
		}
	}
	runSessionTradeFillDebounce.Store(key, now)
	return true
}

// triggerRunSessionRefreshByTradeFill 由“成交落库”触发的写回入口（节流 + 异步）。
// 说明：只用 userId+robotId 定位 run_session，避免 exchange/symbol 历史数据不一致导致找不到区间。
func triggerRunSessionRefreshByTradeFill(ctx context.Context, userId, robotId int64) {
	if userId <= 0 || robotId <= 0 {
		return
	}
	key := g.NewVar(userId).String() + ":" + g.NewVar(robotId).String()
	if !shouldRunSessionRefresh(key, 3*time.Second) {
		return
	}
	go func() {
		// 不要用调用方 ctx（可能已取消）；这里走 background + timeout
		tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		refreshCurrentRunSessionSummaryByTradeFill(tctx, userId, robotId)
	}()
}

// refreshCurrentRunSessionSummaryByTradeFill 找到当前运行区间(end_time IS NULL)，按时间窗从 hg_trading_trade_fill 重算并写回 run_session。
func refreshCurrentRunSessionSummaryByTradeFill(ctx context.Context, userId, robotId int64) {
	if userId <= 0 || robotId <= 0 {
		return
	}

	// 找到当前运行区间（end_time 为空）
	var sess *entity.TradingRobotRunSession
	_ = dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().UserId, userId).
		Where(dao.TradingRobotRunSession.Columns().RobotId, robotId).
		WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
		OrderDesc(dao.TradingRobotRunSession.Columns().Id).
		Limit(1).
		Scan(&sess)
	if sess == nil {
		return
	}

	now := gtime.Now()

	// start_time 兜底：历史数据可能为 year=2006/空；优先用 robot.start_time，再用 sess.created_at
	effectiveStart := sess.StartTime
	if effectiveStart == nil || effectiveStart.IsZero() || effectiveStart.Year() == 2006 {
		var rb *entity.TradingRobot
		_ = dao.TradingRobot.Ctx(ctx).
			Where(dao.TradingRobot.Columns().Id, robotId).
			Where(dao.TradingRobot.Columns().UserId, userId).
			Scan(&rb)
		if rb != nil && rb.StartTime != nil && !rb.StartTime.IsZero() && rb.StartTime.Year() != 2006 {
			effectiveStart = rb.StartTime
		} else if sess.CreatedAt != nil && !sess.CreatedAt.IsZero() && sess.CreatedAt.Year() != 2006 {
			effectiveStart = sess.CreatedAt
		}
		// 写回修正 start_time，避免后续重复兜底
		if effectiveStart != nil && !effectiveStart.IsZero() && effectiveStart.Year() != 2006 {
			_, _ = dao.TradingRobotRunSession.Ctx(ctx).
				Where(dao.TradingRobotRunSession.Columns().Id, sess.Id).
				Data(g.Map{"start_time": effectiveStart, "updated_at": now}).
				Update()
		}
	}
	if effectiveStart == nil || effectiveStart.IsZero() || effectiveStart.Year() == 2006 {
		return
	}

	// 从成交流水表重算（口径与“运行区间列表展示”一致）
	type aggRow struct {
		TotalPnl   float64 `orm:"total_pnl"`
		TotalFee   float64 `orm:"total_fee"`
		TradeCount int     `orm:"trade_count"`
	}
	var agg aggRow

	// PG 兼容：timestamp without timezone 读出可能被当作 UTC，导致 epoch 偏移（常见为 +8h）
	// 这里将 “YYYY-MM-DD HH:mm:ss” 重新按 gtime 时区(Asia/Shanghai) 解析，再取 epoch-ms，确保与 trade_fill.ts 对齐。
	fixEpochMs := func(t *gtime.Time) int64 {
		if t == nil || t.IsZero() {
			return 0
		}
		if dao.TradingRobotRunSession.DB().GetConfig().Type == consts.DBPgsql {
			if tt, e := gtime.StrToTime(t.Format("Y-m-d H:i:s")); e == nil && tt != nil && !tt.IsZero() {
				return tt.UnixMilli()
			}
		}
		return t.UnixMilli()
	}

	startMs := fixEpochMs(effectiveStart)
	endMs := now.UnixMilli()
	if endMs < startMs {
		endMs = startMs
	}
	startSec := startMs / 1000
	endSec := endMs / 1000
	const tsMsThreshold int64 = 1000000000000 // 1e12
	_ = dao.TradingTradeFill.Ctx(ctx).
		Fields("COALESCE(SUM(realized_pnl),0) AS total_pnl", "COALESCE(SUM(fee),0) AS total_fee", "COUNT(1) AS trade_count").
		Where("user_id", userId).
		Where("robot_id", robotId).
		Where("((ts BETWEEN ? AND ?) OR (ts BETWEEN ? AND ? AND ts < ?))", startMs, endMs, startSec, endSec, tsMsThreshold).
		Scan(&agg)

	// 写回区间汇总
	runtimeSeconds := int((endMs - startMs) / 1000)
	if runtimeSeconds < 0 {
		runtimeSeconds = 0
	}
	_, err := dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().Id, sess.Id).
		Data(g.Map{
			"total_pnl":       agg.TotalPnl,
			"total_fee":       agg.TotalFee,
			"trade_count":     agg.TradeCount,
			"synced_at":   now,
			"updated_at":  now,
			"runtime_seconds": runtimeSeconds,
		}).
		Update()
	if err != nil {
		g.Log().Warningf(ctx, "[RunSession] 刷新区间汇总失败: sessionId=%d, robotId=%d, err=%v", sess.Id, robotId, err)
	}
}
