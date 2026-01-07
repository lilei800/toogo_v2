// Package toogo Toogo机器人服务 (简化版)
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
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/exchange"
	"hotgo/internal/library/market"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"
	"hotgo/internal/websocket"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
)

// sToogoRobot 机器人服务
type sToogoRobot struct{}

func init() {
	service.RegisterToogoRobot(NewToogoRobot())
}

// ===================== “空持仓强制直拉”节流（不保留旧内容，只是限频） =====================

// emptyPositionsForceFetchAt: when positions are empty, we periodically force a direct exchange fetch
// to catch manual opens / missed private WS events.
var emptyPositionsForceFetchAt sync.Map // key: int64(robotId) -> time.Time

func shouldForceFetchEmptyPositions(robotId int64, every time.Duration) bool {
	if robotId <= 0 {
		return false
	}
	now := time.Now()
	if v, ok := emptyPositionsForceFetchAt.Load(robotId); ok {
		if t0, ok2 := v.(time.Time); ok2 && now.Sub(t0) < every {
			return false
		}
	}
	emptyPositionsForceFetchAt.Store(robotId, now)
	return true
}

func invalidateRobotPositionsCache(robotId int64) {
	// 历史上这里用于“positions UI 去抖缓存”的失效。
	// 该缓存已移除（避免展示保留旧内容），保留空实现以兼容各处调用点。
	_ = robotId
}

// NewToogoRobot 创建机器人服务
func NewToogoRobot() *sToogoRobot {
	return &sToogoRobot{}
}

// StartRobot 启动机器人
func (s *sToogoRobot) StartRobot(ctx context.Context, in *toogoin.StartRobotInp) error {
	// 查询机器人
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, in.RobotId).Scan(&robot)
	if err != nil {
		return gerror.Wrap(err, "查询机器人失败")
	}
	if robot == nil {
		return gerror.New("机器人不存在")
	}

	if robot.Status == 2 {
		return gerror.New("机器人已在运行中")
	}

	// 检查用户机器人配额
	var toogoUser *entity.ToogoUser
	err = dao.ToogoUser.Ctx(ctx).Where("member_id", robot.UserId).Scan(&toogoUser)
	if err != nil {
		return gerror.Wrap(err, "查询用户信息失败")
	}
	if toogoUser == nil {
		return gerror.New("用户信息不存在")
	}

	if toogoUser.ActiveRobotCount >= toogoUser.RobotLimit {
		return gerror.Newf("机器人数量已达上限(%d/%d)，请升级套餐", toogoUser.ActiveRobotCount, toogoUser.RobotLimit)
	}

	// 检查算力余额
	var wallet *entity.ToogoWallet
	err = dao.ToogoWallet.Ctx(ctx).Where("user_id", robot.UserId).Scan(&wallet)
	if err != nil {
		return gerror.Wrap(err, "查询钱包失败")
	}
	if wallet == nil {
		return gerror.New("钱包不存在")
	}

	totalPower := wallet.Power + wallet.GiftPower
	if totalPower < 10 { // 最少需要10算力
		return gerror.Newf("算力不足，当前算力: %.2f，请充值", totalPower)
	}

	// 更新机器人状态和启动时间
	now := gtime.Now()
	_, err = dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, in.RobotId).Data(g.Map{
		"status":     2,   // 运行中
		"start_time": now, // 记录启动时间
	}).Update()
	if err != nil {
		return gerror.Wrap(err, "更新机器人状态失败")
	}

	// 创建运行区间记录
	_, err = dao.TradingRobotRunSession.Ctx(ctx).Data(g.Map{
		"robot_id":   robot.Id,
		"user_id":    robot.UserId,
		"exchange":   robot.Exchange,
		"symbol":     robot.Symbol,
		"start_time": now,
		// end_time 为 NULL 表示运行中
		// runtime_seconds, total_pnl, total_fee 等在停止或同步时更新
	}).Insert()
	if err != nil {
		g.Log().Warningf(ctx, "创建运行区间记录失败: robotId=%d, err=%v", robot.Id, err)
		// 不影响机器人启动，继续执行
	} else {
		g.Log().Infof(ctx, "创建运行区间记录成功: robotId=%d, startTime=%s", robot.Id, now.Format("Y-m-d H:i:s"))
	}

	// 更新用户活跃机器人数量
	_, err = dao.ToogoUser.Ctx(ctx).
		Where("member_id", robot.UserId).
		Increment("active_robot_count", 1)
	if err != nil {
		return gerror.Wrap(err, "更新活跃机器人数量失败")
	}

	g.Log().Infof(ctx, "机器人启动成功: robotId=%d, userId=%d", robot.Id, robot.UserId)
	// UI 去抖缓存：启动后清理一次，避免页面短暂拿到旧缓存
	invalidateRobotPositionsCache(in.RobotId)
	return nil
}

// StopRobot 停止机器人
func (s *sToogoRobot) StopRobot(ctx context.Context, in *toogoin.StopRobotInp) error {
	// 查询机器人
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, in.RobotId).Scan(&robot)
	if err != nil {
		return gerror.Wrap(err, "查询机器人失败")
	}
	if robot == nil {
		return gerror.New("机器人不存在")
	}

	if robot.Status != 2 {
		return gerror.New("机器人未在运行中")
	}

	// 更新机器人状态
	_, err = dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, in.RobotId).Data(g.Map{
		"status": 3, // 暂停
	}).Update()
	if err != nil {
		return gerror.Wrap(err, "更新机器人状态失败")
	}

	// 【核心修复】更新运行区间的结束时间和运行时长
	// 查询当前运行中的区间（end_time 为 NULL）
	var session *entity.TradingRobotRunSession
	err = dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().RobotId, in.RobotId).
		Where(dao.TradingRobotRunSession.Columns().UserId, robot.UserId).
		WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
		OrderDesc(dao.TradingRobotRunSession.Columns().Id).
		Scan(&session)

	if err == nil && session != nil && session.StartTime != nil {
		// 计算运行时长（秒）
		now := gtime.Now()
		runtimeSeconds := int(now.Sub(session.StartTime).Seconds())

		// 更新运行区间
		_, updateErr := dao.TradingRobotRunSession.Ctx(ctx).
			Where(dao.TradingRobotRunSession.Columns().Id, session.Id).
			Data(g.Map{
				"end_time":        now,
				"end_reason":      "stop", // 手动停止
				"runtime_seconds": runtimeSeconds,
			}).Update()

		if updateErr != nil {
			g.Log().Warningf(ctx, "更新运行区间失败: robotId=%d, sessionId=%d, err=%v", robot.Id, session.Id, updateErr)
		} else {
			g.Log().Infof(ctx, "运行区间已结束: robotId=%d, sessionId=%d, runtime=%ds", robot.Id, session.Id, runtimeSeconds)
		}
	}

	// 更新用户活跃机器人数量
	_, err = dao.ToogoUser.Ctx(ctx).
		Where("member_id", robot.UserId).
		Decrement("active_robot_count", 1)
	if err != nil {
		return gerror.Wrap(err, "更新活跃机器人数量失败")
	}

	g.Log().Infof(ctx, "机器人停止成功: robotId=%d, userId=%d", robot.Id, robot.UserId)
	// UI 去抖缓存：停止后清理一次，避免页面短暂拿到旧缓存
	invalidateRobotPositionsCache(in.RobotId)
	return nil
}

// RunRobotEngine 机器人运行引擎 (供定时任务调用)
// 已废弃：机器人引擎现在由 RobotTaskManager 在HTTP服务启动时自动运行
func (s *sToogoRobot) RunRobotEngine(ctx context.Context) error {
	// 检查 RobotTaskManager 是否运行中
	if !GetRobotTaskManager().IsRunning() {
		return GetRobotTaskManager().Start(ctx)
	}
	return nil
}

// RobotList 机器人列表
func (s *sToogoRobot) RobotList(ctx context.Context, in *toogoin.RobotListInp) ([]*toogoin.RobotListModel, int, error) {
	var list []*entity.TradingRobot

	model := dao.TradingRobot.Ctx(ctx)

	if in.UserId > 0 {
		model = model.Where("user_id", in.UserId)
	}
	if in.Status > 0 {
		model = model.Where("status", in.Status)
	}

	// 软删除过滤
	model = model.WhereNull("deleted_at")

	// 分页
	count, err := model.Count()
	if err != nil {
		return nil, 0, gerror.Wrap(err, "查询机器人数量失败")
	}

	err = model.Page(in.Page, in.PerPage).OrderDesc("id").Scan(&list)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "查询机器人列表失败")
	}

	// 转换结果
	result := make([]*toogoin.RobotListModel, len(list))

	// 获取今天的开始时间（用于计算今日盈亏）
	now := gtime.Now()
	todayStart := gtime.NewFromStr(now.Format("Y-m-d") + " 00:00:00")

	for i, robot := range list {
		statusText := "未启动"
		switch robot.Status {
		case 1:
			statusText = "未启动"
		case 2:
			statusText = "运行中"
		case 3:
			statusText = "已暂停"
		case 4:
			statusText = "已停用"
		}

		// 统计消耗算力
		consumedPower, _ := dao.ToogoPowerConsume.Ctx(ctx).Where("robot_id", robot.Id).Sum("consume_power")

		// 统计今日盈亏（当天已平仓订单的已实现盈亏总和）
		todayPnl, _ := dao.TradingOrder.Ctx(ctx).
			Where("robot_id", robot.Id).
			Where("status", 2). // 已平仓
			WhereGTE("close_time", todayStart).
			Sum("realized_profit")

		result[i] = &toogoin.RobotListModel{
			TradingRobot:  robot,
			ConsumedPower: consumedPower,
			TotalPnl:      robot.TotalProfit,
			TodayPnl:      todayPnl,
			StatusText:    statusText,
		}
	}

	return result, count, nil
}

// GetRobotPositions 获取机器人当前持仓
func (s *sToogoRobot) GetRobotPositions(ctx context.Context, robotId int64) ([]*toogoin.PositionModel, error) {
	// 查询机器人
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot)
	if err != nil {
		return nil, gerror.Wrap(err, "查询机器人失败")
	}
	if robot == nil {
		return nil, gerror.New("机器人不存在")
	}

	// 优先使用运行中引擎的交易所实例（避免重复创建/解密配置）
	engine := GetRobotTaskManager().GetEngine(robotId)
	var ex exchange.Exchange
	if engine != nil && engine.Exchange != nil {
		ex = engine.Exchange
	} else {
		// 获取API配置（后续获取订单历史需要）
		var apiConfig *entity.TradingApiConfig
		err = dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig)
		if err != nil || apiConfig == nil {
			return nil, gerror.New("API配置不存在")
		}
		// 创建交易所实例
		ex, err = GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
		if err != nil {
			return nil, err
		}
	}

	// positions 优先走“引擎缓存 + 行情估值”：
	// - 目标：让前端 positions 实时刷新时尽量不打交易所 GetPositions
	// - 口径：使用 MarketServiceManager 的 ticker.MarkPrice（缺失则LastPrice）估算 MarkPrice，并由 EntryPrice/Qty 推导 UnrealizedPnl
	// - 最终一致性：仍由引擎/同步服务低频对账或交易事件触发刷新交易所持仓
	var positions []*exchange.Position
	if engine != nil {
		engine.mu.RLock()
		cachedPositions := engine.CurrentPositions
		lastUpdate := engine.LastPositionUpdate
		engine.mu.RUnlock()

		// 如果有缓存持仓且不太陈旧，直接用缓存（并做行情估值覆盖），避免API
		// 关键优化：空持仓缓存不能缓存太久，否则用户在交易所手动开仓会“长时间不同步”。
		// - 非空：60s 认为可信（减少API调用）
		// - 空：仅 3s 认为可信（快速发现“手动开仓/私有WS漏事件”的变化）
		cacheTTL := 60 * time.Second
		if cachedPositions != nil && len(cachedPositions) == 0 {
			cacheTTL = 3 * time.Second
		}
		if cachedPositions != nil && time.Since(lastUpdate) < cacheTTL {
			positions = cachedPositions
			// 行情估值：有ticker就覆盖 markPrice/unrealizedPnl（不修改引擎内存，避免竞态）
			if engine.Platform != "" && robot.Symbol != "" {
				if tk := market.GetMarketServiceManager().GetTicker(engine.Platform, robot.Symbol); tk != nil && tk.EffectiveMarkPrice() > 0 {
					price := tk.EffectiveMarkPrice()
					valued := make([]*exchange.Position, 0, len(positions))
					for _, p := range positions {
						if p == nil {
							continue
						}
						cp := *p // copy struct
						cp.MarkPrice = price
						amt := math.Abs(cp.PositionAmt)
						if cp.EntryPrice > 0 && amt > 0 {
							if strings.ToUpper(strings.TrimSpace(cp.PositionSide)) == "SHORT" {
								cp.UnrealizedPnl = (cp.EntryPrice - cp.MarkPrice) * amt
							} else {
								cp.UnrealizedPnl = (cp.MarkPrice - cp.EntryPrice) * amt
							}
						}
						valued = append(valued, &cp)
					}
					positions = valued
				}
			}
			g.Log().Debugf(ctx, "[GetRobotPositions] 使用引擎缓存(<=%v) + 行情估值: robotId=%d, cacheAge=%v, cachedCount=%d", cacheTTL, robotId, time.Since(lastUpdate), len(cachedPositions))
		}
	}

	// 如果没有缓存或缓存超过2秒，从交易所获取
	if positions == nil {
		// 运行中引擎：优先使用智能缓存（singleflight），避免同一时刻重复打交易所
		if engine != nil {
			positions, _ = engine.GetPositionsSmart(ctx, 3*time.Second)
		}
		// 引擎未运行或缓存仍为空：才直接调用交易所API（兜底）
		if positions == nil {
			positions, err = ex.GetPositions(ctx, robot.Symbol)
			if err != nil {
				return nil, gerror.Wrap(err, "获取持仓失败")
			}
			// 同步更新引擎缓存（若引擎存在）
			if engine != nil {
				engine.mu.Lock()
				engine.CurrentPositions = positions
				engine.LastPositionUpdate = time.Now()
				engine.mu.Unlock()
			}
		}
	}

	// 如果拿到的是“空持仓”，但用户可能在交易所手动开仓（尤其 Bitget/Gate 私有WS 丢事件/订阅异常时），
	// 则做一次低频强制直连交易所获取，避免长时间看不到。
	if positions != nil && len(positions) == 0 {
		// 10 秒最多强制一次，避免刷接口
		if shouldForceFetchEmptyPositions(robotId, 10*time.Second) {
			plat := strings.ToLower(strings.TrimSpace(robot.Exchange))
			if engine != nil && strings.TrimSpace(engine.Platform) != "" {
				plat = strings.ToLower(strings.TrimSpace(engine.Platform))
			}
			g.Log().Warningf(ctx, "[GetRobotPositions] empty positions -> force fetch exchange: robotId=%d platform=%s symbol=%s", robotId, plat, robot.Symbol)
			pos2, err2 := ex.GetPositions(ctx, robot.Symbol)
			if err2 != nil {
				g.Log().Warningf(ctx, "[GetRobotPositions] force fetch exchange failed: robotId=%d err=%v", robotId, err2)
			} else {
				positions = pos2
				if engine != nil {
					engine.mu.Lock()
					engine.CurrentPositions = positions
					engine.LastPositionUpdate = time.Now()
					engine.mu.Unlock()
				}
			}
		}
	}

	// 【重要】记录交易所返回的所有持仓，用于调试
	// 【静默优化】避免在 Debug 模式下每次请求刷屏，保留一条汇总即可
	if len(positions) > 0 {
		first := positions[0]
		g.Log().Debugf(ctx, "[GetRobotPositions] 持仓汇总: robotId=%d count=%d first={symbol=%s side=%s amt=%.6f upl=%.4f}",
			robotId, len(positions), first.Symbol, first.PositionSide, first.PositionAmt, first.UnrealizedPnl)
	} else {
		g.Log().Debugf(ctx, "[GetRobotPositions] 持仓汇总: robotId=%d count=0", robotId)
	}

	// ===== 关键规范化：统一 PositionSide 大小写，避免 OKX/Gate 返回 long/short 导致内存 tracker 查不到 =====
	for _, p := range positions {
		if p == nil {
			continue
		}
		p.PositionSide = strings.ToUpper(strings.TrimSpace(p.PositionSide))
		p.Symbol = strings.ToUpper(strings.TrimSpace(p.Symbol))
	}

	// 【优化】使用引擎缓存的订单历史，避免重复调用API
	var historyOrders []*exchange.Order
	if engine := GetRobotTaskManager().GetEngine(robotId); engine != nil {
		cachedOrders, lastUpdate := engine.GetCachedOrderHistory()
		// 如果缓存在120秒内，使用缓存
		if cachedOrders != nil && time.Since(lastUpdate) < 120*time.Second {
			historyOrders = cachedOrders
			g.Log().Debugf(ctx, "[GetRobotPositions] 使用缓存的订单历史: robotId=%d, 缓存时间=%v", robotId, lastUpdate)
		}
	}
	// 如果没有缓存，才调用API
	if historyOrders == nil {
		historyOrders, err = ex.GetOrderHistory(ctx, robot.Symbol, 50)
		if err != nil {
			g.Log().Warningf(ctx, "[GetRobotPositions] 获取订单历史失败: robotId=%d, err=%v", robotId, err)
			historyOrders = []*exchange.Order{} // 设置为空切片，避免nil
		}
	}

	// 【优化】构建订单映射表，根据持仓方向和交易对匹配订单
	// key: "symbol_positionSide"，value: 最新的开仓订单
	orderMap := make(map[string]*exchange.Order)
	for _, order := range historyOrders {
		// 只处理开仓订单（持仓方向匹配）
		if order.PositionSide == "" {
			continue
		}
		ps := strings.ToUpper(strings.TrimSpace(order.PositionSide))
		sym := strings.ToUpper(strings.TrimSpace(order.Symbol))
		key := fmt.Sprintf("%s_%s", sym, ps)
		// 如果已存在订单，保留创建时间最新的
		if existing, exists := orderMap[key]; !exists || order.CreateTime > existing.CreateTime {
			orderMap[key] = order
		}
	}

	// 【内存优化】从引擎获取监控数据，不查询数据库

	// ===== 冻结策略参数：优先使用内存 tracker；若引擎未运行/参数未加载，则从DB读取一次兜底 =====
	type frozenParams struct {
		StopLossPercent         float64
		AutoStartRetreatPercent float64
		ProfitRetreatPercent    float64
		ProfitRetreatStarted    int
		HighestProfit           float64
		MarginPercent           float64
		Margin                  float64
		Leverage                int
		MarketState             string
		RiskPreference          string
	}
	frozenBySide := map[string]frozenParams{}
	needFrozenFromDB := engine == nil
	if !needFrozenFromDB {
		// 如果某个方向 tracker 不存在或未加载，也需要兜底
		for _, pos := range positions {
			if engine.GetPositionTracker(pos.PositionSide) == nil {
				needFrozenFromDB = true
				break
			}
			if tr := engine.GetPositionTracker(pos.PositionSide); tr != nil && !tr.ParamsLoaded {
				needFrozenFromDB = true
				break
			}
		}
	}
	if needFrozenFromDB {
		// 只取持仓中订单（最多两条：long/short），轻量查询
		var rows []struct {
			Direction               string  `json:"direction"`
			StopLossPercent         float64 `json:"stopLossPercent"`
			AutoStartRetreatPercent float64 `json:"autoStartRetreatPercent"`
			ProfitRetreatPercent    float64 `json:"profitRetreatPercent"`
			ProfitRetreatStarted    int     `json:"profitRetreatStarted"`
			HighestProfit           float64 `json:"highestProfit"`
			MarginPercent           float64 `json:"marginPercent"`
			Margin                  float64 `json:"margin"`
			Leverage                int     `json:"leverage"`
			MarketState             string  `json:"marketState"`
			RiskPreference          string  `json:"riskPreference"`
		}
		_ = dao.TradingOrder.Ctx(ctx).
			Where("robot_id", robotId).
			Where("status", OrderStatusOpen).
			Fields(
				"direction",
				"margin",
				"leverage",
				"stop_loss_percent",
				"auto_start_retreat_percent",
				"profit_retreat_percent",
				"profit_retreat_started",
				"highest_profit",
				"margin_percent",
				"market_state",
				"risk_preference",
			).
			Scan(&rows)
		for _, r := range rows {
			side := "LONG"
			if strings.ToLower(strings.TrimSpace(r.Direction)) == "short" {
				side = "SHORT"
			}
			frozenBySide[side] = frozenParams{
				StopLossPercent:         r.StopLossPercent,
				AutoStartRetreatPercent: r.AutoStartRetreatPercent,
				ProfitRetreatPercent:    r.ProfitRetreatPercent,
				ProfitRetreatStarted:    r.ProfitRetreatStarted,
				HighestProfit:           r.HighestProfit,
				MarginPercent:           r.MarginPercent,
				Margin:                  r.Margin,
				Leverage:                r.Leverage,
				MarketState:             r.MarketState,
				RiskPreference:          r.RiskPreference,
			}
		}
	}

	// 转换结果 (exchange.Position 字段映射到 PositionModel)
	// 注意：必须返回空切片而不是 nil，避免 WS/HTTP 序列化为 null 导致前端误判
	result := make([]*toogoin.PositionModel, 0)
	filteredByZero := 0
	for _, pos := range positions {
		if pos == nil {
			continue
		}
		// 过滤已平仓/无效持仓（避免平仓后短暂返回 PositionAmt≈0 的残留对象导致前端仍显示）
		// 注意：这里不能用 0.0001 这种较大阈值，否则小仓位（例如 0.0001 BTC）会被误过滤，导致“交易所有两条仓位但页面只显示一条”。
		//
		// 【Bitget/Gate 兼容兜底】少数情况下交易所返回 qty=0 但 margin/leverage/entryPrice 已就绪，
		// 这会导致前端“有仓位但不显示”。这里按保证金口径反推一个近似数量用于展示/风控：
		// qty ≈ margin * leverage / entryPrice
		if math.Abs(pos.PositionAmt) <= positionAmtEpsilon {
			entry := pos.EntryPrice
			lev := pos.Leverage
			if lev <= 0 && robot != nil && robot.Leverage > 0 {
				lev = robot.Leverage
			}
			m := pos.Margin
			if m <= 0 && pos.IsolatedMargin > 0 {
				m = pos.IsolatedMargin
			}
			if entry > 0 && lev > 0 && m > 0 {
				derived := (m * float64(lev)) / entry
				if strings.EqualFold(strings.TrimSpace(pos.PositionSide), "SHORT") {
					derived = -derived
				}
				// 只在推导结果足够可信时覆盖（避免把真实空仓误判为有仓）
				if math.Abs(derived) > positionAmtEpsilon {
					pos.PositionAmt = derived
					if isOrderPositionSyncDebugEnabled(ctx) && shouldLogOrderPositionSync("pos_qty_derived:"+g.NewVar(robotId).String(), 3*time.Second) {
						g.Log().Warningf(ctx, "[SyncDiag] derived PositionAmt for display: robotId=%d symbol=%s posSide=%s entry=%.6f leverage=%d margin=%.6f qty=%.8f",
							robotId, pos.Symbol, pos.PositionSide, entry, lev, m, derived)
					}
				}
			}
		}
		if math.Abs(pos.PositionAmt) <= positionAmtEpsilon {
			filteredByZero++
			continue
		}

		g.Log().Debugf(ctx, "[GetRobotPositions] 处理持仓: symbol=%s, PositionSide=%s, PositionAmt=%.6f",
			pos.Symbol, pos.PositionSide, pos.PositionAmt)

		// 【内存优化】从内存获取止盈回撤状态和最高盈利
		var maxProfitReached float64
		var takeProfitEnabled bool
		var stopLossPercent, autoStartRetreatPercent, profitRetreatPercent, marginPercent *float64
		entryMargin := 0.0
		entryLeverage := 0
		var marketState, riskPreference string

		if engine != nil {
			tracker := engine.GetPositionTracker(pos.PositionSide)
			if tracker != nil {
				maxProfitReached = tracker.HighestProfit
				takeProfitEnabled = tracker.TakeProfitEnabled
				if tracker.EntryMargin > 0 {
					entryMargin = tracker.EntryMargin
				}
				// 冻结参数（开仓时确定）
				if tracker.ParamsLoaded {
					if tracker.StopLossPercent > 0 {
						v := tracker.StopLossPercent
						stopLossPercent = &v
					}
					if tracker.AutoStartRetreatPercent > 0 {
						v := tracker.AutoStartRetreatPercent
						autoStartRetreatPercent = &v
					}
					if tracker.ProfitRetreatPercent > 0 {
						v := tracker.ProfitRetreatPercent
						profitRetreatPercent = &v
					}
					if tracker.MarginPercent > 0 {
						v := tracker.MarginPercent
						marginPercent = &v
					}
					marketState = tracker.MarketState
					riskPreference = tracker.RiskPreference
				}
			}
		}
		// 兜底：引擎未运行或 tracker 未加载时，从DB冻结参数补齐
		if stopLossPercent == nil && autoStartRetreatPercent == nil && profitRetreatPercent == nil {
			if fp, ok := frozenBySide[pos.PositionSide]; ok {
				if fp.StopLossPercent > 0 {
					v := fp.StopLossPercent
					stopLossPercent = &v
				}
				if fp.AutoStartRetreatPercent > 0 {
					v := fp.AutoStartRetreatPercent
					autoStartRetreatPercent = &v
				}
				if fp.ProfitRetreatPercent > 0 {
					v := fp.ProfitRetreatPercent
					profitRetreatPercent = &v
				}
				if fp.MarginPercent > 0 {
					v := fp.MarginPercent
					marginPercent = &v
				}
				if entryMargin <= 0 && fp.Margin > 0 {
					entryMargin = fp.Margin
				}
				if entryLeverage <= 0 && fp.Leverage > 0 {
					entryLeverage = fp.Leverage
				}
				if marketState == "" {
					marketState = fp.MarketState
				}
				if riskPreference == "" {
					riskPreference = fp.RiskPreference
				}

				// 【按需求调整】HighestProfit / TakeProfitEnabled 的读取保持“原来口径”（仅从内存 tracker 读取）。
				// DB 仍会持久化 highest_profit / profit_retreat_started 用于审计与兜底恢复，
				// 但接口返回不再用 DB 覆盖/回灌这两个状态，避免改变前端读取口径。
			}
		}

		// 【重要】最高盈利就是订单的未实现盈亏的最大值，只增加不减少
		// 如果当前未实现盈亏为正且大于记录的最高盈利，更新最高盈利
		if pos.UnrealizedPnl > 0 && pos.UnrealizedPnl > maxProfitReached {
			maxProfitReached = pos.UnrealizedPnl
			// 同步更新内存中的最高盈利
			if engine != nil {
				tracker := engine.GetOrCreatePositionTracker(pos.PositionSide, pos.Margin)
				if tracker != nil && maxProfitReached > tracker.HighestProfit {
					tracker.HighestProfit = maxProfitReached
				}
			}
		}

		// 匹配订单信息
		orderKey := fmt.Sprintf("%s_%s", pos.Symbol, pos.PositionSide)
		matchedOrder := orderMap[orderKey]

		var orderId, clientOrderId, orderType, orderSide string
		var orderQuantity, orderAvgPrice float64
		var orderCreateTime int64

		if matchedOrder != nil {
			orderId = matchedOrder.OrderId
			clientOrderId = matchedOrder.ClientId
			orderType = matchedOrder.Type
			orderSide = matchedOrder.Side
			orderQuantity = matchedOrder.Quantity
			orderAvgPrice = matchedOrder.AvgPrice
			// 【修复】只有当CreateTime有效时才赋值（大于0）
			if matchedOrder.CreateTime > 0 {
				orderCreateTime = matchedOrder.CreateTime
			}
			g.Log().Debugf(ctx, "[GetRobotPositions] 匹配到订单: symbol=%s, positionSide=%s, orderId=%s, clientOrderId=%s, createTime=%d",
				pos.Symbol, pos.PositionSide, orderId, clientOrderId, orderCreateTime)
		}
		// 前端通常会用 orderId 作为 rowKey/渲染关键字段。
		// 对于“交易所有持仓但本地/历史订单无法匹配”的场景（手动开仓/WS丢事件/缓存未就绪），
		// 如果 orderId 为空，前端可能直接不展示该行。
		// 这里生成一个稳定的虚拟ID，保证 UI 可渲染且不随刷新抖动。
		if strings.TrimSpace(orderId) == "" {
			symKey := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(pos.Symbol), "/", ""))
			ps := strings.ToUpper(strings.TrimSpace(pos.PositionSide))
			if ps == "" {
				ps = "UNKNOWN"
			}
			orderId = fmt.Sprintf("POS-%d-%s-%s", robotId, symKey, ps)
			// best-effort：补齐订单侧字段（用于前端展示，不参与对账）
			orderType = "POSITION"
			if ps == "LONG" {
				orderSide = "BUY"
			} else if ps == "SHORT" {
				orderSide = "SELL"
			}
			orderQuantity = math.Abs(pos.PositionAmt)
			orderAvgPrice = pos.EntryPrice
			if orderCreateTime == 0 {
				orderCreateTime = time.Now().UnixMilli()
			}
		}

		// ===== 保证金口径（规则6）：优先使用“订单表冻结的保证金”，缺失时才用交易所/估算兜底 =====
		displayMargin := entryMargin
		if displayMargin <= 0 {
			// fallback: exchange snapshot
			displayMargin = pos.Margin
			if displayMargin <= 0 && pos.IsolatedMargin > 0 {
				displayMargin = pos.IsolatedMargin
			}
			if displayMargin <= 0 && pos.EntryPrice > 0 && math.Abs(pos.PositionAmt) > 0 {
				lev := pos.Leverage
				if lev <= 0 {
					lev = entryLeverage
				}
				if lev <= 0 && robot != nil && robot.Leverage > 0 {
					lev = robot.Leverage
				}
				if lev > 0 {
					displayMargin = math.Abs(pos.PositionAmt) * pos.EntryPrice / float64(lev)
				}
			}
		}

		// 直接使用交易所API返回的数据（部分字段做了上面的显示兜底）
		// ===== 后端统一血条/进度口径（前端只展示）=====
		// 1) 实时盈利百分比 = 未实现盈亏 / 保证金 × 100%
		realTimeProfitPercent := 0.0
		if displayMargin > 0 {
			realTimeProfitPercent = (pos.UnrealizedPnl / displayMargin) * 100.0
		}
		// 2) 启动止盈血条：达到阈值后锁定为100%（不可关闭原则）
		takeProfitStartProgress := 0.0
		if autoStartRetreatPercent != nil && *autoStartRetreatPercent > 0 {
			if takeProfitEnabled {
				takeProfitStartProgress = 100.0
			} else if realTimeProfitPercent > 0 {
				takeProfitStartProgress = (realTimeProfitPercent / *autoStartRetreatPercent) * 100.0
				if takeProfitStartProgress < 0 {
					takeProfitStartProgress = 0
				}
				if takeProfitStartProgress > 100 {
					takeProfitStartProgress = 100
				}
			}
		}
		// 3) 止盈回撤：回撤百分比=(最高盈利-未实现盈亏)/最高盈利×100%；血条默认100回撤到0触发止盈
		takeProfitRetreatPercentNow := 0.0
		takeProfitRetreatBar := 0.0
		if takeProfitEnabled {
			// 默认展示 100%（刚启动但参数/最高盈利未就绪时不抖动）
			takeProfitRetreatBar = 100.0
			if profitRetreatPercent != nil && *profitRetreatPercent > 0 && maxProfitReached > 0 {
				takeProfitRetreatPercentNow = ((maxProfitReached - pos.UnrealizedPnl) / maxProfitReached) * 100.0
				if takeProfitRetreatPercentNow < 0 {
					takeProfitRetreatPercentNow = 0
				}
				takeProfitRetreatBar = 100.0 - (takeProfitRetreatPercentNow/(*profitRetreatPercent))*100.0
				if takeProfitRetreatBar < 0 {
					takeProfitRetreatBar = 0
				}
				if takeProfitRetreatBar > 100 {
					takeProfitRetreatBar = 100
				}
			}
		}
		// 4) 止损血条：|未实现盈亏| / (保证金×止损%) × 100%
		stopLossProgress := 0.0
		if stopLossPercent != nil && *stopLossPercent > 0 && displayMargin > 0 && pos.UnrealizedPnl < 0 {
			stopLossAmount := displayMargin * (*stopLossPercent / 100.0)
			if stopLossAmount > 0 {
				stopLossProgress = (math.Abs(pos.UnrealizedPnl) / stopLossAmount) * 100.0
				if stopLossProgress < 0 {
					stopLossProgress = 0
				}
				// 展示层限制为100%，触发阈值由后端风控执行
				if stopLossProgress > 100 {
					stopLossProgress = 100
				}
			}
		}

		positionModel := &toogoin.PositionModel{
			Symbol:                   pos.Symbol,
			PositionSide:             pos.PositionSide,
			PositionAmt:              pos.PositionAmt,
			EntryPrice:               pos.EntryPrice, // 使用交易所返回的开仓价格
			MarkPrice:                pos.MarkPrice,
			UnrealizedPnl:            pos.UnrealizedPnl,
			Leverage:                 pos.Leverage, // 使用交易所返回的杠杆
			Margin:                   displayMargin,
			MarginType:               pos.MarginType,
			IsolatedMargin:           pos.IsolatedMargin,
			LiquidationPrice:         pos.LiquidationPrice,
			MaxProfitReached:         maxProfitReached,  // 运行时状态
			TakeProfitEnabled:        takeProfitEnabled, // 运行时状态
			RealTimeProfitPercent:    realTimeProfitPercent,
			TakeProfitStartProgress:  takeProfitStartProgress,
			TakeProfitRetreatPercent: takeProfitRetreatPercentNow,
			TakeProfitRetreatBar:     takeProfitRetreatBar,
			StopLossProgress:         stopLossProgress,
			// 冻结策略参数（开仓时确定，持仓期间不随市场状态变化）
			StopLossPercent:         stopLossPercent,
			AutoStartRetreatPercent: autoStartRetreatPercent,
			ProfitRetreatPercent:    profitRetreatPercent,
			MarginPercent:           marginPercent,
			MarketState:             marketState,
			RiskPreference:          riskPreference,
			CreateTime:              0, // 不使用数据库数据
			// 订单信息（从交易所API获取）
			OrderId:       orderId,
			ClientOrderId: clientOrderId,
			OrderType:     orderType,
			OrderSide:     orderSide,
			OrderQuantity: orderQuantity,
			OrderAvgPrice: orderAvgPrice,
		}

		// 【修复】只有当orderCreateTime有效时才设置（避免前端显示错误时间）
		if orderCreateTime > 0 {
			positionModel.CreateTime = orderCreateTime
			positionModel.OrderCreateTime = orderCreateTime
		}

		result = append(result, positionModel)
	}
	// 关键排障：交易所返回持仓对象非空，但全部被过滤为 0 → 前端会“完全不显示持仓”
	if len(positions) > 0 && len(result) == 0 && filteredByZero > 0 {
		first := positions[0]
		if first != nil && isOrderPositionSyncDebugEnabled(ctx) && shouldLogOrderPositionSync("pos_all_filtered:"+g.NewVar(robotId).String(), 5*time.Second) {
			g.Log().Warningf(ctx, "[SyncDiag] positions all filtered (qty~0): robotId=%d symbol=%s first={symbol=%s side=%s amt=%.10f entry=%.6f mark=%.6f lev=%d margin=%.6f iso=%.6f} filteredByZero=%d",
				robotId, robot.Symbol, first.Symbol, first.PositionSide, first.PositionAmt, first.EntryPrice, first.MarkPrice, first.Leverage, first.Margin, first.IsolatedMargin, filteredByZero)
		}
	}
	return result, nil
}

// GetRobotOpenOrders 获取机器人当前挂单
func (s *sToogoRobot) GetRobotOpenOrders(ctx context.Context, robotId int64) ([]*toogoin.OrderModel, error) {
	// 查询机器人
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot)
	if err != nil {
		return nil, gerror.Wrap(err, "查询机器人失败")
	}
	if robot == nil {
		return nil, gerror.New("机器人不存在")
	}

	// 【方案A】挂单列表只读 DB（交易所订单事实表）
	// - 私有WS事件增量写入 hg_trading_exchange_order
	// - OrderStatusSyncService 低频REST兜底对账
	type row struct {
		ExchangeOrderId string  `json:"exchangeOrderId" orm:"exchange_order_id"`
		ClientOrderId   string  `json:"clientOrderId" orm:"client_order_id"`
		Symbol          string  `json:"symbol" orm:"symbol"`
		Side            string  `json:"side" orm:"side"`
		PositionSide    string  `json:"positionSide" orm:"position_side"`
		OrderType       string  `json:"orderType" orm:"order_type"`
		Price           float64 `json:"price" orm:"price"`
		Quantity        float64 `json:"quantity" orm:"quantity"`
		FilledQty       float64 `json:"filledQty" orm:"filled_qty"`
		AvgPrice        float64 `json:"avgPrice" orm:"avg_price"`
		Status          string  `json:"status" orm:"status"`
		RawStatus       string  `json:"rawStatus" orm:"raw_status"`
		CreateTime      int64   `json:"createTime" orm:"create_time"`
		UpdateTime      int64   `json:"updateTime" orm:"update_time"`
		LastEventTime   int64   `json:"lastEventTime" orm:"last_event_time"`
	}
	var rows []row
	q := g.DB().Model(exchangeOrderTable).Ctx(ctx).
		Where("robot_id", robotId).
		Where("symbol", robot.Symbol).
		Where("is_open", 1).
		Fields(
			"exchange_order_id",
			"client_order_id",
			"symbol",
			"side",
			"position_side",
			"order_type",
			"price",
			"quantity",
			"filled_qty",
			"avg_price",
			"status",
			"raw_status",
			"create_time",
			"update_time",
			"last_event_time",
		).
		OrderDesc("COALESCE(update_time, last_event_time, create_time)")
	if err := q.Scan(&rows); err != nil {
		// 表不存在时降级为旧逻辑（避免上线先跑服务再跑SQL直接报错）
		if isTableMissingErr(err) {
			g.Log().Warningf(ctx, "[GetRobotOpenOrders] %s 表不存在，已降级为直连交易所获取（请先执行迁移SQL）: err=%v", exchangeOrderTable, err)
			return s.getRobotOpenOrdersFromExchange(ctx, robot)
		}
		return nil, gerror.Wrap(err, "查询挂单失败")
	}

	result := make([]*toogoin.OrderModel, 0, len(rows))
	for _, r := range rows {
		ct := r.CreateTime
		if ct <= 0 {
			ct = r.LastEventTime
		}
		ut := r.UpdateTime
		if ut <= 0 {
			ut = r.LastEventTime
		}
		status := r.Status
		if strings.TrimSpace(status) == "" {
			status = r.RawStatus
		}
		result = append(result, &toogoin.OrderModel{
			OrderId:      r.ExchangeOrderId,
			ClientId:     r.ClientOrderId,
			Symbol:       r.Symbol,
			Side:         r.Side,
			PositionSide: r.PositionSide,
			Type:         r.OrderType,
			Price:        r.Price,
			Quantity:     r.Quantity,
			FilledQty:    r.FilledQty,
			AvgPrice:     r.AvgPrice,
			Status:       status,
			CreateTime:   ct,
			UpdateTime:   ut,
		})
	}
	return result, nil
}

// getRobotOpenOrdersFromExchange 旧逻辑：直接查询交易所挂单（仅用于迁移期间的兜底）
func (s *sToogoRobot) getRobotOpenOrdersFromExchange(ctx context.Context, robot *entity.TradingRobot) ([]*toogoin.OrderModel, error) {
	if robot == nil {
		return nil, gerror.New("robot is nil")
	}
	// 获取API配置
	var apiConfig *entity.TradingApiConfig
	err := dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig)
	if err != nil || apiConfig == nil {
		return nil, gerror.New("API配置不存在")
	}
	ex, err := GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
	if err != nil {
		return nil, err
	}
	orders, err := ex.GetOpenOrders(ctx, robot.Symbol)
	if err != nil {
		return nil, gerror.Wrap(err, "获取挂单失败")
	}
	result := make([]*toogoin.OrderModel, 0, len(orders))
	for _, order := range orders {
		result = append(result, &toogoin.OrderModel{
			OrderId:      order.OrderId,
			ClientId:     order.ClientId,
			Symbol:       order.Symbol,
			Side:         order.Side,
			PositionSide: order.PositionSide,
			Type:         order.Type,
			Price:        order.Price,
			Quantity:     order.Quantity,
			FilledQty:    order.FilledQty,
			AvgPrice:     order.AvgPrice,
			Status:       order.Status,
			CreateTime:   order.CreateTime,
			UpdateTime:   order.UpdateTime,
		})
	}
	return result, nil
}

// GetRobotOrderHistory 获取机器人历史订单（从数据库读取，数据库数据会自动更新）
func (s *sToogoRobot) GetRobotOrderHistory(ctx context.Context, robotId int64, limit int) ([]*toogoin.OrderModel, error) {
	// 查询机器人
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot)
	if err != nil {
		return nil, gerror.Wrap(err, "查询机器人失败")
	}
	if robot == nil {
		return nil, gerror.New("机器人不存在")
	}

	if limit <= 0 {
		limit = 50
	}

	// 【优化】直接从数据库读取订单历史（数据库数据会自动更新）
	var orders []*entity.TradingOrder
	err = dao.TradingOrder.Ctx(ctx).
		Fields("id", "exchange_order_id", "symbol", "direction", "open_price", "close_price",
			"quantity", "leverage", "margin", "realized_profit", "status", "close_reason",
			"open_time", "close_time", "created_at", "updated_at").
		Where("robot_id", robotId).
		// 用户更关心“最新平仓/最新变更”的记录：
		// - 优先 close_time（若为空或历史占位值 2006 年，则使用 updated_at 兜底）
		// - 避免已平仓但 close_time 缺失的订单沉到后面
		// 兼容 MySQL/PG：用 CASE + EXTRACT 替代 IF/YEAR
		OrderDesc("CASE WHEN close_time IS NULL OR EXTRACT(YEAR FROM close_time)=2006 THEN updated_at ELSE close_time END").
		Limit(limit).
		Scan(&orders)
	if err != nil {
		return nil, gerror.Wrap(err, "查询订单历史失败")
	}

	// 转换结果
	var result []*toogoin.OrderModel
	for _, order := range orders {
		// 确定订单状态
		status := "NEW"
		if order.Status == 2 {
			status = "FILLED"
		} else if order.Status == 3 {
			status = "CANCELED"
		}

		// 确定持仓方向
		positionSide := "LONG"
		if order.Direction == "short" {
			positionSide = "SHORT"
		}

		// 确定买卖方向
		side := "BUY"
		if order.Direction == "short" {
			side = "SELL"
		}

		// 转换时间戳（优先使用open_time作为开仓时间）
		var createTime, updateTime int64
		if order.OpenTime != nil && !order.OpenTime.IsZero() {
			createTime = order.OpenTime.UnixMilli()
		} else if order.CreatedAt != nil {
			createTime = order.CreatedAt.UnixMilli()
		}
		if order.UpdatedAt != nil {
			updateTime = order.UpdatedAt.UnixMilli()
		}

		result = append(result, &toogoin.OrderModel{
			OrderId:      order.ExchangeOrderId,
			ClientId:     "", // 数据库中没有保存客户端订单ID
			Symbol:       order.Symbol,
			Side:         side,
			PositionSide: positionSide,
			Type:         "MARKET", // 默认市价单
			Price:        order.OpenPrice,
			Quantity:     order.Quantity,
			FilledQty:    order.Quantity,  // 已成交数量等于数量
			AvgPrice:     order.OpenPrice, // 使用开仓价格作为成交均价
			Status:       status,
			CreateTime:   createTime, // 使用open_time（开仓时间）
			UpdateTime:   updateTime,
		})
	}

	return result, nil
}

// SyncOrderHistoryToDB 同步订单历史到数据库（判断是否需要更新）
// 供 OrderStatusSyncService 调用，每60秒同步一次
func (s *sToogoRobot) SyncOrderHistoryToDB(ctx context.Context, robotId int64, robot *entity.TradingRobot, orders []*exchange.Order) error {
	if len(orders) == 0 {
		return nil
	}

	g.Log().Infof(ctx, "[syncOrderHistoryToDB] 开始同步订单历史: robotId=%d, 订单数量=%d", robotId, len(orders))

	// 查询数据库中已存在的订单（通过 exchange_order_id 或 client_order_id + create_time 匹配）
	existingOrderMap := make(map[string]*entity.TradingOrder)

	// 构建查询条件：获取该机器人的所有订单
	var existingOrders []*entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Fields("id", "exchange_order_id", "created_at", "updated_at").
		Where("robot_id", robotId).
		OrderDesc("created_at").
		Limit(1000). // 限制查询数量，避免查询过多
		Scan(&existingOrders)
	if err != nil {
		g.Log().Warningf(ctx, "[syncOrderHistoryToDB] 查询已存在订单失败: robotId=%d, err=%v", robotId, err)
	} else {
		// 构建映射表：key = exchange_order_id
		for _, existingOrder := range existingOrders {
			if existingOrder.ExchangeOrderId != "" {
				existingOrderMap[existingOrder.ExchangeOrderId] = existingOrder
			}
			// 如果创建时间存在，也使用 exchange_order_id + create_time 作为备用key
			if existingOrder.ExchangeOrderId != "" && existingOrder.CreatedAt != nil {
				key := fmt.Sprintf("%s_%d", existingOrder.ExchangeOrderId, existingOrder.CreatedAt.Unix())
				existingOrderMap[key] = existingOrder
			}
		}
	}

	// 处理每个订单
	insertCount := 0
	updateCount := 0
	for _, order := range orders {
		// 订单ID选择原则：
		// - exchange_order_id 优先使用交易所返回的 OrderId（更稳定/更适合作为幂等键）
		// - client_order_id 作为辅助字段保存（用于对账/排查）
		exchangeOrderId := strings.TrimSpace(order.OrderId)
		clientOrderId := strings.TrimSpace(order.ClientId)
		if exchangeOrderId == "" {
			// 极少数场景交易所可能不返回 OrderId，这时兜底用 ClientId
			exchangeOrderId = clientOrderId
		}
		if exchangeOrderId == "" {
			continue // 跳过没有订单ID的记录
		}

		// 判断订单是否已存在
		var existingOrder *entity.TradingOrder
		if existingOrder = existingOrderMap[exchangeOrderId]; existingOrder == nil {
			// 尝试使用 create_time 匹配
			if order.CreateTime > 0 {
				createTimeKey := fmt.Sprintf("%s_%d", exchangeOrderId, order.CreateTime/1000)
				existingOrder = existingOrderMap[createTimeKey]
			}
		}

		// 确定方向
		direction := "long"
		if order.PositionSide == "SHORT" {
			direction = "short"
		} else if order.Side == "SELL" {
			direction = "short"
		}

		// 确定订单状态
		// 【重要修复】交易所订单状态 FILLED 表示“订单成交”，不等于“持仓已平仓”。
		// 是否平仓必须结合 Type（开仓/平仓）来判断，否则会把开仓成交单误标记为已平仓，
		// 进而导致“平台手动平仓”无法通过持仓对账链路补全 close_time/close_price/close_reason。
		orderStatus := 1 // 默认持仓中
		statusUpper := strings.ToUpper(strings.TrimSpace(order.Status))
		typeLower := strings.ToLower(strings.TrimSpace(order.Type))
		isCloseOrder := strings.Contains(order.Type, "平仓") || strings.Contains(typeLower, "close")
		isOpenOrder := strings.Contains(order.Type, "开仓") || strings.Contains(typeLower, "open")

		if statusUpper == "CANCELED" || statusUpper == "CANCELLED" {
			orderStatus = 3 // 已取消
		} else if statusUpper == "FILLED" || statusUpper == "CLOSED" {
			// 只有明确是“平仓单”才标记为已平仓；开仓成交单仍然属于持仓中
			if isCloseOrder && !isOpenOrder {
				orderStatus = 2 // 已平仓
			} else {
				orderStatus = 1 // 持仓中（开仓成交/无法识别类型时默认按持仓中处理，避免误关闭）
			}
		}

		// 生成订单号（如果不存在）
		// 【修复】使用 PHP 风格格式 "YmdHis" 替代 Go 标准格式，确保 gtime.Format 正确工作
		orderSn := fmt.Sprintf("TO%s%s", gtime.NewFromTimeStamp(order.CreateTime/1000).Format("YmdHis"), grand.S(6))

		if existingOrder != nil {
			// 更新已存在的订单
			updateData := g.Map{
				"updated_at": gtime.Now(),
			}
			// 尝试补全 client_order_id（若交易所返回）
			if clientOrderId != "" {
				updateData["client_order_id"] = clientOrderId
			}

			// 更新字段（如果API返回的数据更新）
			if order.AvgPrice > 0 {
				updateData["open_price"] = order.AvgPrice
			}
			if order.FilledQty > 0 {
				updateData["quantity"] = order.FilledQty
			}
			if order.Status != "" {
				updateData["status"] = orderStatus
			}

			// 【修复】检测并修复错误的时间字段（包含 Go 格式模板字符串 "2006" 的错误数据）
			if order.CreateTime > 0 {
				correctTime := gtime.NewFromTimeStamp(order.CreateTime / 1000)
				// 检查 open_time 是否为空或错误
				if existingOrder.OpenTime == nil || existingOrder.OpenTime.IsZero() ||
					(existingOrder.OpenTime != nil && existingOrder.OpenTime.Year() == 2006) {
					updateData["open_time"] = correctTime
					g.Log().Infof(ctx, "[syncOrderHistoryToDB] 修复订单时间: orderId=%d, correctTime=%v", existingOrder.Id, correctTime)
				}
				// 检查 created_at 是否错误
				if existingOrder.CreatedAt != nil && existingOrder.CreatedAt.Year() == 2006 {
					updateData["created_at"] = correctTime
				}
			}

			// 只更新有变化的字段
			_, err := dao.TradingOrder.Ctx(ctx).
				Where(dao.TradingOrder.Columns().Id, existingOrder.Id).
				Update(updateData)
			// 兼容：部分环境可能尚未执行迁移脚本，缺少 client_order_id 字段；此时回退重试
			if err != nil && clientOrderId != "" &&
				strings.Contains(strings.ToLower(err.Error()), "client_order_id") &&
				(strings.Contains(strings.ToLower(err.Error()), "unknown column") || strings.Contains(strings.ToLower(err.Error()), "does not exist")) {
				delete(updateData, "client_order_id")
				_, err = dao.TradingOrder.Ctx(ctx).
					Where(dao.TradingOrder.Columns().Id, existingOrder.Id).
					Update(updateData)
			}
			if err != nil {
				g.Log().Warningf(ctx, "[syncOrderHistoryToDB] 更新订单失败: orderId=%d, exchangeOrderId=%s, err=%v",
					existingOrder.Id, exchangeOrderId, err)
			} else {
				updateCount++
			}
		} else {
			// 插入新订单
			// 确定开仓时间（优先使用交易所返回的CreateTime）
			var openTime *gtime.Time
			if order.CreateTime > 0 {
				openTime = gtime.NewFromTimeStamp(order.CreateTime / 1000)
			} else {
				openTime = gtime.Now()
			}

			orderData := g.Map{
				"user_id":           robot.UserId,
				"robot_id":          robotId,
				"exchange":          robot.Exchange,
				"symbol":            order.Symbol,
				"order_sn":          orderSn,
				"exchange_order_id": exchangeOrderId,
				"client_order_id":   clientOrderId,
				"direction":         direction,
				"open_price":        order.AvgPrice,
				"quantity":          order.FilledQty,
				"status":            orderStatus,
				"open_time":         openTime, // 保存交易所返回的订单创建时间作为开仓时间
				"created_at":        openTime, // 使用相同的开仓时间作为创建时间
				"updated_at":        gtime.Now(),
			}

			_, err := dao.TradingOrder.Ctx(ctx).Insert(orderData)
			// 兼容：部分环境可能尚未执行迁移脚本，缺少 client_order_id 字段；此时回退重试
			if err != nil && clientOrderId != "" &&
				strings.Contains(strings.ToLower(err.Error()), "client_order_id") &&
				(strings.Contains(strings.ToLower(err.Error()), "unknown column") || strings.Contains(strings.ToLower(err.Error()), "does not exist")) {
				delete(orderData, "client_order_id")
				_, err = dao.TradingOrder.Ctx(ctx).Insert(orderData)
			}
			if err != nil {
				g.Log().Warningf(ctx, "[syncOrderHistoryToDB] 插入订单失败: robotId=%d, exchangeOrderId=%s, err=%v",
					robotId, exchangeOrderId, err)
			} else {
				insertCount++
			}
		}
	}

	g.Log().Infof(ctx, "[syncOrderHistoryToDB] 同步完成: robotId=%d, 新增=%d, 更新=%d", robotId, insertCount, updateCount)
	return nil
}

// CloseRobotPosition 手动平仓
func (s *sToogoRobot) CloseRobotPosition(ctx context.Context, in *toogoin.ClosePositionInp) error {
	// 验证用户权限
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 查询机器人（并验证权限）
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.RobotId).
		Where(dao.TradingRobot.Columns().UserId, memberId). // 验证权限：只能平仓自己的机器人
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)
	if err != nil {
		return gerror.Wrap(err, "查询机器人失败")
	}
	if robot == nil {
		return gerror.New("机器人不存在或无权限")
	}

	// 获取API配置
	var apiConfig *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig)
	if err != nil || apiConfig == nil {
		return gerror.New("API配置不存在")
	}

	// 创建交易所实例
	ex, err := GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
	if err != nil {
		return err
	}

	// 如果 Symbol 为空，使用机器人配置的 Symbol
	symbol := in.Symbol
	if symbol == "" {
		symbol = robot.Symbol
	}

	// 规范化 PositionSide（确保是大写）
	// 参考自动平仓逻辑：positionSide 表示持仓方向（LONG/SHORT），与订单 side（BUY/SELL）不同
	// LONG 多仓平仓 → 卖出(SELL)，SHORT 空仓平仓 → 买入(BUY)
	positionSide := strings.ToUpper(strings.TrimSpace(in.PositionSide))
	if positionSide != "LONG" && positionSide != "SHORT" {
		return gerror.Newf("持仓方向参数无效: '%s'，有效值为 LONG(做多) 或 SHORT(做空)", in.PositionSide)
	}

	g.Log().Infof(ctx, "[ClosePosition] 开始平仓: robotId=%d, inputSymbol=%s, robotSymbol=%s, finalSymbol=%s, inputPositionSide=%s, normalizedPositionSide=%s, inputQty=%.6f",
		in.RobotId, in.Symbol, robot.Symbol, symbol, in.PositionSide, positionSide, in.Quantity)

	// 【重要】优先检查内存中的持仓，而不是查询数据库
	// 如果内存中没有持仓，直接返回错误，不执行平仓
	robotEngine := GetRobotTaskManager().GetEngine(in.RobotId)
	if robotEngine == nil {
		return gerror.Newf("机器人引擎不存在，机器人可能未运行: robotId=%d", in.RobotId)
	}

	// 检查内存中是否有该方向的持仓
	if !robotEngine.HasActivePosition(positionSide) {
		// 获取内存中的持仓列表，用于错误提示
		cachedPositions, _ := robotEngine.GetCachedPositions()
		var availableSides []string
		for _, pos := range cachedPositions {
			if math.Abs(pos.PositionAmt) > positionAmtEpsilon {
				sideCN := "多单"
				if pos.PositionSide == "SHORT" {
					sideCN = "空单"
				}
				availableSides = append(availableSides, fmt.Sprintf("%s(%s, 数量:%.6f)", sideCN, pos.PositionSide, pos.PositionAmt))
			}
		}
		requestSideCN := "多单"
		if positionSide == "SHORT" {
			requestSideCN = "空单"
		}
		if len(availableSides) > 0 {
			return gerror.Newf("平仓失败: 内存中未找到 %s(%s) 方向的持仓。当前可平仓持仓: %v", requestSideCN, positionSide, availableSides)
		}
		return gerror.Newf("平仓失败: 内存中未找到 %s(%s) 方向的持仓，且无其他持仓", requestSideCN, positionSide)
	}

	// 从内存中获取持仓信息
	memoryPosition := robotEngine.GetPosition(positionSide)
	if memoryPosition == nil {
		return gerror.Newf("平仓失败: 无法从内存中获取 %s 方向的持仓信息", positionSide)
	}

	// 【重要】使用内存中的 symbol，而不是前端传递的 symbol
	// 因为内存中的 symbol 是从交易所获取的，格式是正确的
	// 这样可以避免前端传递的 symbol 格式（如 "BTCUSDT"）与交易所返回的格式（如 "BTC/USDT"）不匹配的问题
	symbol = memoryPosition.Symbol
	g.Log().Infof(ctx, "[ClosePosition] ✅ 使用内存中的 symbol: %s (原始请求: %s)", symbol, in.Symbol)

	g.Log().Infof(ctx, "[ClosePosition] ✅ 内存中确认有持仓: symbol=%s, positionSide=%s, positionAmt=%.6f, unrealizedPnl=%.4f",
		memoryPosition.Symbol, memoryPosition.PositionSide, memoryPosition.PositionAmt, memoryPosition.UnrealizedPnl)

	// 【重要】直接使用内存中的持仓数据，不需要再从交易所获取
	// 因为内存中的数据是从交易所获取的，格式是正确的，而且我们已经验证过内存中有持仓
	var foundPosition *exchange.Position
	var currentPnl float64
	var actualQuantity float64 = in.Quantity

	// 使用内存中的持仓数据
	foundPosition = memoryPosition
	currentPnl = memoryPosition.UnrealizedPnl

	// 如果传入的 Quantity 为 0 或未指定，使用实际持仓数量
	if actualQuantity <= 0 {
		actualQuantity = math.Abs(memoryPosition.PositionAmt)
	}
	// 确保平仓数量不超过实际持仓数量
	if actualQuantity > math.Abs(memoryPosition.PositionAmt) {
		actualQuantity = math.Abs(memoryPosition.PositionAmt)
		g.Log().Warningf(ctx, "[ClosePosition] 平仓数量超过实际持仓，调整为实际持仓数量: %.6f", actualQuantity)
	}

	g.Log().Infof(ctx, "[ClosePosition] ✅ 使用内存中的持仓数据: symbol=%s, positionSide=%s, positionAmt=%.6f, 平仓数量=%.6f, unrealizedPnl=%.4f",
		memoryPosition.Symbol, memoryPosition.PositionSide, memoryPosition.PositionAmt, actualQuantity, currentPnl)

	if actualQuantity <= 0 {
		return gerror.New("平仓数量必须大于0")
	}

	g.Log().Infof(ctx, "[ClosePosition] ⚠️ 准备执行平仓: symbol=%s, positionSide=%s, qty=%.6f, pnl=%.4f, foundPosition=%+v",
		symbol, positionSide, actualQuantity, currentPnl, foundPosition)

	// 【重要】再次确认 positionSide 正确（foundPosition 一定不为 nil，因为前面已经检查过）
	// 参考自动平仓逻辑：使用实际持仓的 PositionSide 进行 API 调用，确保方向一致
	foundPosSide := strings.ToUpper(strings.TrimSpace(foundPosition.PositionSide))
	if foundPosSide != positionSide {
		// 【修复】提供更清晰的错误提示，说明多空方向
		requestSideCN := "多单"
		if positionSide == "SHORT" {
			requestSideCN = "空单"
		}
		foundSideCN := "多单"
		if foundPosSide == "SHORT" {
			foundSideCN = "空单"
		}
		return gerror.Newf("平仓方向不匹配: 请求平仓 %s(%s)，但实际找到的持仓是 %s(%s)。请确认传入的方向参数是否正确", requestSideCN, positionSide, foundSideCN, foundPosSide)
	}

	// 执行平仓
	g.Log().Infof(ctx, "[ClosePosition] 调用交易所 ClosePosition API: symbol=%s, positionSide=%s, quantity=%.6f",
		symbol, positionSide, actualQuantity)
	// 【优化】为平仓增加硬超时，避免代理/网络/交易所偶发卡顿导致“平仓很久没响应”
	closeCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	order, err := ex.ClosePosition(closeCtx, symbol, positionSide, actualQuantity)
	if err != nil {
		g.Log().Errorf(ctx, "[ClosePosition] 平仓失败: robotId=%d, symbol=%s, side=%s, qty=%.6f, err=%v",
			in.RobotId, symbol, positionSide, actualQuantity, err)
		// 透传错误原因给前端，避免只显示“平仓失败”无法定位
		return gerror.Newf("平仓失败: %v", err)
	}

	g.Log().Infof(ctx, "手动平仓成功: robotId=%d, symbol=%s, side=%s, orderId=%s, qty=%.6f, pnl=%.4f",
		in.RobotId, symbol, positionSide, order.OrderId, actualQuantity, currentPnl)

	// UI 去抖缓存：手动平仓是强一致操作，成功后立即清掉缓存，避免页面短暂显示旧持仓
	invalidateRobotPositionsCache(in.RobotId)

	// 【新增】保存手动平仓日志
	s.saveManualCloseLog(ctx, in.RobotId, foundPosition, order, "")

	// 【优化】落库成交流水改为异步（不阻塞手动平仓接口响应）
	// 说明：trade fills 落库可能触发多次API请求（尤其 OKX 分页），同步执行会显著拖慢“手动平仓耗时”。
	{
		apiConfigId := robot.ApiConfigId
		robotId := in.RobotId
		exName := ex.GetName()
		go func(sym string) {
			tctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			if saved, matched, ferr := fetchAndStoreTradeHistory(tctx, ex, apiConfigId, exName, sym, 200); ferr != nil {
				g.Log().Debugf(tctx, "[ClosePosition] 异步落库成交流水失败(忽略): robotId=%d, symbol=%s, err=%v", robotId, sym, ferr)
			} else {
				g.Log().Debugf(tctx, "[ClosePosition] 异步已落库成交流水: robotId=%d, symbol=%s, saved=%d, matched=%d", robotId, sym, saved, matched)
			}
		}(symbol)
	}

	// 【重要】计算已实现盈亏和平仓价格（使用内存中的持仓数据和API返回的数据）
	var realizedProfit float64 = currentPnl
	var closePrice float64 = 0.0

	// 【API兼容性】优先使用API返回的成交均价，其次使用标记价格，最后使用开仓价格
	if order.AvgPrice > 0 {
		closePrice = order.AvgPrice
	} else if foundPosition.MarkPrice > 0 {
		closePrice = foundPosition.MarkPrice
	} else if foundPosition.EntryPrice > 0 {
		closePrice = foundPosition.EntryPrice
	} else if memoryPosition.EntryPrice > 0 {
		closePrice = memoryPosition.EntryPrice
	}

	// 如果没有价格，使用内存中的持仓价格
	if closePrice == 0 && memoryPosition.EntryPrice > 0 {
		closePrice = memoryPosition.EntryPrice
	}

	// 使用内存中的持仓数据计算盈亏（如果API没有返回盈亏）
	if realizedProfit == 0 && memoryPosition.UnrealizedPnl != 0 {
		realizedProfit = memoryPosition.UnrealizedPnl
	}

	// 【修复】立即更新数据库订单状态为已平仓（不依赖同步）
	// 统一走 OrderStatusSyncService.CloseOrder：
	// - 写入 close_time/close_price/realized_profit/close_order_id/close_fee 等字段（只补缺失字段，幂等）
	// - 处理算力扣除、运行区间汇总刷新
	// - 避免手动平仓里写入不存在字段导致 UPDATE 失败
	direction := "long"
	if positionSide == "SHORT" {
		direction = "short"
	}
	var localOrder *entity.TradingOrder
	_ = dao.TradingOrder.Ctx(ctx).
		Where("robot_id", in.RobotId).
		// 兼容历史数据：direction 可能为 LONG/SHORT/Long 等，统一按 lower(direction) 匹配
		Where("LOWER(direction) = ?", direction).
		// 兼容：平仓发生在本地订单仍为 pending/open 的窗口内
		Where("status IN (?)", []int{OrderStatusPending, OrderStatusOpen}).
		OrderDesc("id").
		Limit(1).
		Scan(&localOrder)
	if localOrder == nil {
		// 平仓已经在交易所成功，这里不返回错误；记录告警，后续可由同步服务/对账逻辑补齐
		g.Log().Warningf(ctx, "[ClosePosition] 平仓成功但未找到本地持仓订单记录，无法补全订单平仓信息: robotId=%d, direction=%s, closeOrderId=%s",
			in.RobotId, direction, order.OrderId)
	} else {
		GetOrderStatusSyncService().CloseOrder(ctx, localOrder, closePrice, realizedProfit, "manual", order, foundPosition)
	}

	// 【优化】使用统一的清除方法清除内存中的持仓信息，避免状态不一致
	if robotEngine := GetRobotTaskManager().GetEngine(in.RobotId); robotEngine != nil {
		robotEngine.ClearPosition(ctx, positionSide)
		robotEngine.ClearPositionTracker(positionSide)
		g.Log().Infof(ctx, "[ClosePosition] robotId=%d 手动平仓完成: 已更新数据库和内存", in.RobotId)
	}

	// 推送“平仓成功”事件给前端：用于详情弹窗订单列表秒级刷新（不依赖轮询/同步）
	if memberId > 0 {
		websocket.SendToUser(memberId, &websocket.WResponse{
			Event: "toogo/robot/trade/event",
			Data: g.Map{
				"type":           "close_success",
				"robotId":        in.RobotId,
				"symbol":         symbol,
				"positionSide":   positionSide,
				"direction":      strings.ToLower(positionSide), // long/short（与订单 direction 兼容）
				"closeOrderId":   order.OrderId,
				"closePrice":     closePrice,
				"realizedProfit": realizedProfit,
				"ts":             gtime.Now().TimestampMilli(),
			},
		})
	}

	return nil
}

// saveManualCloseLog 保存手动平仓日志
func (s *sToogoRobot) saveManualCloseLog(ctx context.Context, robotId int64, pos *exchange.Position, closeOrder *exchange.Order, errMsg string) {
	// 查询关联的本地订单（用于获取 orderId）
	var localOrderId int64 = 0
	direction := "long"
	if pos.PositionSide == "SHORT" {
		direction = "short"
	}
	var localOrder struct {
		Id int64
	}
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robotId).
		// 兼容历史数据：direction 可能为 LONG/SHORT/Long 等，统一按 lower(direction) 匹配
		Where("LOWER(direction) = ?", direction).
		Where("status IN (?)", []int{OrderStatusPending, OrderStatusOpen}).
		Fields("id").
		Scan(&localOrder)
	if err == nil && localOrder.Id > 0 {
		localOrderId = localOrder.Id
	}

	// 确定状态和消息
	status := "success"
	message := ""
	if errMsg != "" {
		status = "failed"
		message = errMsg
	} else {
		message = fmt.Sprintf("手动平仓成功: %s方向, 数量%.6f, 盈亏%.4f USDT", pos.PositionSide, math.Abs(pos.PositionAmt), pos.UnrealizedPnl)
	}

	// 构建事件数据
	eventData := map[string]interface{}{
		"close_type":     "manual",
		"symbol":         pos.Symbol,
		"position_side":  pos.PositionSide,
		"quantity":       math.Abs(pos.PositionAmt),
		"entry_price":    pos.EntryPrice,
		"unrealized_pnl": pos.UnrealizedPnl,
		"margin":         pos.Margin,
	}
	if closeOrder != nil {
		eventData["exchange_order_id"] = closeOrder.OrderId
		eventData["avg_price"] = closeOrder.AvgPrice
		eventData["filled_qty"] = closeOrder.FilledQty
	}

	// 序列化事件数据为JSON
	eventDataJSON := "{}"
	if len(eventData) > 0 {
		data, jsonErr := json.Marshal(eventData)
		if jsonErr == nil {
			eventDataJSON = string(data)
		}
	}

	// 写入交易日志
	_, insertErr := g.DB().Model("hg_trading_execution_log").Ctx(ctx).Insert(g.Map{
		"signal_log_id": 0, // 手动平仓不关联预警记录
		"robot_id":      robotId,
		"order_id":      localOrderId,
		"event_type":    "close_manual",
		"event_data":    eventDataJSON,
		"status":        status,
		"message":       message,
		"created_at":    time.Now(),
	})
	if insertErr != nil {
		g.Log().Warningf(ctx, "[ClosePosition] 保存手动平仓日志失败: robotId=%d, err=%v", robotId, insertErr)
	} else {
		g.Log().Debugf(ctx, "[ClosePosition] 手动平仓日志已保存: robotId=%d, status=%s", robotId, status)
	}
}

// SetTakeProfitRetreatSwitch 设置止盈回撤开关状态
// 【内存优化】直接操作内存中的 PositionTracker，不再依赖数据库
func (s *sToogoRobot) SetTakeProfitRetreatSwitch(ctx context.Context, robotId int64, positionSide string, enabled bool) error {
	// 验证用户权限
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 2026-01 规范：止盈由后端自动控制，前端只负责展示。
	// - 启动止盈：实时盈利百分比>=启动止盈阈值时由后端自动打开
	// - 不可关闭原则：一旦开启直到平仓不可关闭
	// 因此不再允许任何手动开关操作（兼容老前端调用，统一返回明确错误）。
	_ = robotId
	_ = positionSide
	_ = enabled
	return gerror.New("止盈回撤由系统自动控制，前端仅展示，不支持手动设置")
}

// CancelRobotOrder 撤销挂单
func (s *sToogoRobot) CancelRobotOrder(ctx context.Context, robotId int64, orderId string) error {
	// 查询机器人
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot)
	if err != nil {
		return gerror.Wrap(err, "查询机器人失败")
	}
	if robot == nil {
		return gerror.New("机器人不存在")
	}

	// 获取API配置
	var apiConfig *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig)
	if err != nil || apiConfig == nil {
		return gerror.New("API配置不存在")
	}

	// 创建交易所实例并撤单
	ex, err := GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
	if err != nil {
		return err
	}

	_, err = ex.CancelOrder(ctx, robot.Symbol, orderId)
	if err != nil {
		return gerror.Wrap(err, "撤单失败")
	}

	g.Log().Infof(ctx, "撤单成功: robotId=%d, symbol=%s, orderId=%s", robotId, robot.Symbol, orderId)
	// UI 去抖缓存：撤单后可能影响“持仓/挂单展示”，清理缓存让页面尽快刷新
	invalidateRobotPositionsCache(robotId)

	// OKX/Gate：撤单后 openOrders 在 DB 的兜底对账默认有 10s 节流，页面会出现“撤单成功但挂单列表还在”的体验问题。
	// 这里做一次轻量的“撤单后立即对账 openOrders”（带超时、异步，不阻塞接口响应）。
	// 仅用于改善用户体验（C），不影响交易所真实状态。
	plat := strings.ToLower(strings.TrimSpace(ex.GetName()))
	if plat == "okx" || plat == "gate" {
		go func() {
			defer func() { recover() }()
			callCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			orders, oerr := ex.GetOpenOrders(callCtx, robot.Symbol)
			if oerr != nil {
				g.Log().Debugf(callCtx, "[CancelRobotOrder] 撤单后立即对账 openOrders 失败(忽略): robotId=%d platform=%s symbol=%s err=%v", robotId, plat, robot.Symbol, oerr)
				return
			}
			_ = SyncExchangeOpenOrdersToDB(callCtx, robotId, ex.GetName(), robot.ApiConfigId, robot.Symbol, orders)
		}()
		// 同时触发一次订单/持仓事件驱动对账（不阻塞）
		GetOrderStatusSyncService().TriggerRobotSync(robotId)
	}

	// 推送“订单变更”事件给前端：用于详情弹窗挂单列表秒级刷新（不依赖10s轮询）
	if robot.UserId > 0 {
		websocket.SendToUser(robot.UserId, &websocket.WResponse{
			Event: "toogo/robot/trade/event",
			Data: g.Map{
				"type":            "order_delta",
				"action":          "cancel",
				"robotId":         robotId,
				"symbol":          robot.Symbol,
				"exchangeOrderId": orderId,
				"ts":              gtime.Now().TimestampMilli(),
			},
		})
	}

	return nil
}

// ClearSignalLogs 清空预警记录
// robotId: 机器人ID，0表示清空所有机器人的预警记录
// keepExecuted: 是否保留已执行的记录（true=只删除未执行的，false=删除所有）
func (s *sToogoRobot) ClearSignalLogs(ctx context.Context, robotId int64, keepExecuted bool) error {
	// 如果清空所有记录，使用 TRUNCATE（更快且不需要WHERE条件）
	if robotId == 0 && !keepExecuted {
		_, err := g.DB().Exec(ctx, "TRUNCATE TABLE hg_trading_signal_log")
		if err != nil {
			return gerror.Wrap(err, "清空预警记录失败")
		}
		g.Log().Infof(ctx, "清空预警记录成功: 已清空所有记录")
		return nil
	}

	// 有条件的删除，使用WHERE条件
	query := g.DB().Model("hg_trading_signal_log").Ctx(ctx)

	// 如果指定了机器人ID，只删除该机器人的记录
	if robotId > 0 {
		query = query.Where("robot_id", robotId)
	}

	// 如果保留已执行的记录，只删除未执行的
	if keepExecuted {
		query = query.Where("executed", 0)
	}

	// 执行删除
	result, err := query.Delete()
	if err != nil {
		return gerror.Wrap(err, "清空预警记录失败")
	}

	deletedCount, _ := result.RowsAffected()
	g.Log().Infof(ctx, "清空预警记录成功: robotId=%d, keepExecuted=%v, 删除数量=%d", robotId, keepExecuted, deletedCount)

	return nil
}
