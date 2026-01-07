// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"time"
	tradingapi "hotgo/api/admin/trading"
	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/exchange"
	"hotgo/internal/library/market"
	"hotgo/internal/logic/toogo"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type robotImpl struct{}

func canonicalPlatform(p string) string {
	return strings.ToLower(strings.TrimSpace(p))
}

// canonicalSymbol is the ONLY symbol format stored in DB / used as cache keys.
// It intentionally matches exchange.Formatter.NormalizeSymbol: "BTCUSDT".
func canonicalSymbol(s string) string {
	return exchange.Formatter.NormalizeSymbol(s)
}

// List 获取机器人列表
func (s *robotImpl) List(ctx context.Context, in *input.TradingRobotListInp) (list []*input.TradingRobotListModel, totalCount int, err error) {
	mod := dao.TradingRobot.Ctx(ctx)

	// 租户隔离
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		err = gerror.New("用户未登录")
		return
	}
	mod = mod.Where(dao.TradingRobot.Columns().UserId, memberId)

	// 条件筛选
	if in.Status > 0 {
		mod = mod.Where(dao.TradingRobot.Columns().Status, in.Status)
	}
	if in.Symbol != "" {
		mod = mod.Where(dao.TradingRobot.Columns().Symbol, in.Symbol)
	}
	if in.RiskPreference != "" {
		mod = mod.Where(dao.TradingRobot.Columns().RiskPreference, in.RiskPreference)
	}
	if in.RobotName != "" {
		mod = mod.WhereLike(dao.TradingRobot.Columns().RobotName, "%"+in.RobotName+"%")
	}

	// 软删除过滤
	mod = mod.WhereNull(dao.TradingRobot.Columns().DeletedAt)

	totalCount, err = mod.Count()
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return
	}

	err = mod.Page(in.Page, in.PageSize).
		Order(dao.TradingRobot.Columns().CreatedAt + " DESC").
		Scan(&list)

	if err != nil {
		return nil, 0, err
	}

	// ===== 列表实时市场状态（反推全局引擎是否正常产出）=====
	// 说明：
	// - DB 字段 trading_robot.market_state 在新架构下不会实时更新，容易“空/旧/被污染”
	// - 列表页展示更应取全局 MarketAnalyzer 的实时结果（内存读缓存，成本很低）
	// - 若分析未产出（nil/过期），则不返回 marketState（避免展示 DB 旧值误导）
	if len(list) > 0 {
		ma := market.GetMarketAnalyzer()
		now := time.Now()
		for _, item := range list {
			if item == nil || item.Exchange == "" || item.Symbol == "" {
				continue
			}
			// 不使用 DB 值（实时态不应依赖DB）；默认置空，只有实时分析可用时才填充
			item.MarketState = ""
			analysis := ma.GetAnalysis(item.Exchange, item.Symbol)
			if analysis == nil {
				continue
			}
			// 只使用较新的分析结果，避免展示过期状态
			if now.Sub(analysis.UpdatedAt) > 10*time.Second {
				continue
			}
			ms := string(analysis.MarketState)
			// 统一输出格式：trend/volatile/high_vol/low_vol
			if ms == "range" {
				ms = "volatile"
			}
			if ms == "high-volatility" {
				ms = "high_vol"
			}
			if ms == "low-volatility" {
				ms = "low_vol"
			}
			if ms != "" {
				item.MarketState = ms
			}
		}
	}

	// 解析当前策略JSON
	for _, item := range list {
		if item.CurrentStrategySnapshot != nil {
			// 策略已解析为JSON对象
		}
	}

	return
}

// Create 创建机器人
func (s *robotImpl) Create(ctx context.Context, in *input.TradingRobotCreateInp) (id int64, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return 0, gerror.New("用户未登录")
	}

	// 验证API配置是否存在且属于当前用户
	var apiConfig *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.ApiConfigId).
		Where(dao.TradingApiConfig.Columns().UserId, memberId).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
		Scan(&apiConfig)

	if err != nil {
		return 0, err
	}
	if apiConfig == nil {
		return 0, gerror.New("API配置不存在或无权限")
	}

	platform := canonicalPlatform(apiConfig.Platform)
	if platform == "" {
		return 0, gerror.New("API配置平台为空，请检查API配置")
	}

	// 统一Symbol存储口径：DB 内只存 BTCUSDT（无分隔符）
	symbol := canonicalSymbol(in.Symbol)
	if symbol == "" {
		return 0, gerror.New("交易对不能为空")
	}

	// 校验策略组：平台/币对必须与机器人一致（每机器人只绑定一个平台API账户）
	var group *entity.TradingStrategyGroup
	_ = g.DB().Model("hg_trading_strategy_group").Ctx(ctx).
		Where("id", in.StrategyGroupId).
		Scan(&group)
	if group == nil || group.Id == 0 {
		return 0, gerror.New("策略组不存在，请重新选择")
	}
	if canonicalPlatform(group.Exchange) != platform {
		return 0, gerror.Newf("策略组平台(%s)与API平台(%s)不一致，无法创建机器人", group.Exchange, platform)
	}
	if canonicalSymbol(group.Symbol) != symbol {
		return 0, gerror.Newf("策略组交易对(%s)与机器人交易对(%s)不一致，无法创建机器人", group.Symbol, symbol)
	}
	if group.IsActive == 0 {
		return 0, gerror.New("该策略组已禁用，无法创建机器人")
	}

	// 【新增】校验：每个API配置只能绑定一个未删除的机器人
	var existingRobot *entity.TradingRobot
	err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().ApiConfigId, in.ApiConfigId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&existingRobot)
	
	if err != nil {
		return 0, gerror.Wrap(err, "检查API配置绑定失败")
	}
	if existingRobot != nil {
		return 0, gerror.Newf("该API配置已绑定机器人【%s】，每个API配置只能绑定一个机器人", existingRobot.RobotName)
	}

	// v2 规则：创建时只绑定策略组ID + 市场状态→风险偏好映射；交易参数/止盈止损运行时从策略模板加载
	if in.StrategyGroupId == 0 {
		return 0, gerror.New("策略组ID不能为空，请选择策略组")
	}

	// current_strategy 仅保存 groupId（运行时按 groupId+市场状态+风险偏好加载模板）
	strategyConfig := map[string]interface{}{
		"groupId": in.StrategyGroupId,
	}
	strategyJSON, err := json.Marshal(strategyConfig)
	if err != nil {
		return 0, gerror.Wrap(err, "策略JSON转换失败")
	}

	// remark 存储“市场状态→风险偏好”映射（若未传则落默认映射）
	mapping := in.MarketRiskMapping
	if mapping == nil || len(mapping) == 0 {
		mapping = map[string]string{
			"trend":    "balanced",
			"volatile": "balanced",
			"high_vol": "aggressive",
			"low_vol":  "conservative",
		}
	}
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return 0, gerror.Wrap(err, "映射关系JSON转换失败")
	}

	// 开关默认值
	autoTrade := 0
	if in.AutoTradeEnabled != nil {
		autoTrade = *in.AutoTradeEnabled
	}
	autoClose := 1
	if in.AutoCloseEnabled != nil {
		autoClose = *in.AutoCloseEnabled
	}
	// 锁定盈利开关默认开启（止盈启动后禁止自动开新仓）
	profitLock := 1
	dualSide := 1
	if in.DualSidePosition != nil {
		dualSide = *in.DualSidePosition
	}

	insertData := g.Map{
		"user_id":            memberId,
		"robot_name":         in.RobotName,
		"api_config_id":      in.ApiConfigId,
		"max_profit_target":  in.MaxProfitTarget,
		"max_loss_amount":    in.MaxLossAmount,
		"max_runtime":        in.MaxRuntime,
		"auto_market_state":  in.AutoMarketState,
		// 【强一致】platform 以 api_config.platform 为准；robot.exchange 仅存平台标识（binance/okx/bitget/gate）
		"exchange": platform,
		// 【强一致】symbol 统一存 BTCUSDT
		"symbol":             symbol,
		"use_monitor_signal": in.UseMonitorSignal,
		"current_strategy":   string(strategyJSON),
		"strategy_group_id":  in.StrategyGroupId,
		"auto_trade_enabled": autoTrade,
		"auto_close_enabled": autoClose,
		"profit_lock_enabled": profitLock,
		"dual_side_position": dualSide,
		"status":             1, // 未启动
		"remark":             string(mappingJSON),
	}

	// 定时开关（可选）
	if in.ScheduleStart != "" {
		if t, e := gtime.StrToTime(in.ScheduleStart); e == nil {
			insertData["schedule_start"] = t
		}
	}
	if in.ScheduleStop != "" {
		if t, e := gtime.StrToTime(in.ScheduleStop); e == nil {
			insertData["schedule_stop"] = t
		}
	}

	// 备注（注意：remark 已用于映射JSON；如果还想保留用户备注，需要单独字段。这里不覆盖）
	_ = in.Remark

	// 【PostgreSQL 兼容】InsertAndGetId() 不支持 PostgreSQL，改用事务 + LASTVAL()
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return 0, gerror.Wrap(err, "开启事务失败")
	}
	defer tx.Rollback()
	
	_, err = tx.Model("hg_trading_robot").Ctx(ctx).Data(insertData).Insert()
	if err != nil {
		return 0, gerror.Wrap(err, "创建机器人失败")
	}
	
	val, err := tx.GetValue("SELECT LASTVAL()")
	if err != nil {
		return 0, gerror.Wrap(err, "获取机器人ID失败")
	}
	id = val.Int64()
	
	err = tx.Commit()
	if err != nil {
		return 0, gerror.Wrap(err, "提交事务失败")
	}
	
	return id, nil
}

// Update 更新机器人
func (s *robotImpl) Update(ctx context.Context, in *input.TradingRobotUpdateInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 验证所属
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)

	if err != nil {
		return err
	}
	if robot == nil {
		return gerror.New("机器人不存在或无权限")
	}

	// 允许运行中仅更新开关；其他字段变更需要先暂停
	hasToggleUpdate := in.AutoTradeEnabled != nil || in.AutoCloseEnabled != nil || in.ProfitLockEnabled != nil || in.DualSidePosition != nil
	hasOtherUpdate := in.RobotName != "" ||
		in.MaxProfitTarget != 0 || in.MaxLossAmount != 0 || in.MaxRuntime != 0 ||
		in.RiskPreference != "" || in.MarketState != "" ||
		in.Leverage != 0 || in.MarginPercent != 0 ||
		in.StopLossPercent != 0 || in.ProfitRetreatPercent != 0 || in.AutoStartRetreatPercent != 0 ||
		in.UseMonitorSignal != 0 || in.AutoMarketState != 0 ||
		in.Remark != ""

	if robot.Status == 2 && hasOtherUpdate {
		return gerror.New("运行中的机器人不允许修改配置，只能切换自动下单/自动平仓/双向开单开关")
	}

	data := g.Map{}
	if in.RobotName != "" {
		data["robot_name"] = in.RobotName
	}
	if in.MaxProfitTarget != 0 {
		data["max_profit_target"] = in.MaxProfitTarget
	}
	if in.MaxLossAmount != 0 {
		data["max_loss_amount"] = in.MaxLossAmount
	}
	if in.MaxRuntime != 0 {
		data["max_runtime"] = in.MaxRuntime
	}
	if in.AutoMarketState != 0 {
		data["auto_market_state"] = in.AutoMarketState
	}
	if in.UseMonitorSignal != 0 {
		data["use_monitor_signal"] = in.UseMonitorSignal
	}
	// 这些字段运行时会被策略模板覆盖，但保留更新入口以兼容老数据/管理操作
	if in.RiskPreference != "" {
		data["risk_preference"] = in.RiskPreference
	}
	if in.MarketState != "" {
		data["market_state"] = in.MarketState
	}
	if in.Leverage != 0 {
		data["leverage"] = in.Leverage
	}
	if in.MarginPercent != 0 {
		data["margin_percent"] = in.MarginPercent
	}
	if in.StopLossPercent != 0 {
		data["stop_loss_percent"] = in.StopLossPercent
	}
	if in.ProfitRetreatPercent != 0 {
		data["profit_retreat_percent"] = in.ProfitRetreatPercent
	}
	if in.AutoStartRetreatPercent != 0 {
		data["auto_start_retreat_percent"] = in.AutoStartRetreatPercent
	}

	// 开关（可选）
	if in.AutoTradeEnabled != nil {
		data["auto_trade_enabled"] = *in.AutoTradeEnabled
	}
	if in.AutoCloseEnabled != nil {
		data["auto_close_enabled"] = *in.AutoCloseEnabled
	}
	if in.ProfitLockEnabled != nil {
		data["profit_lock_enabled"] = *in.ProfitLockEnabled
	}
	if in.DualSidePosition != nil {
		data["dual_side_position"] = *in.DualSidePosition
	}

	// remark：v2 remark 用于映射JSON，不建议直接写入普通备注；但保留管理员覆盖入口
	if in.Remark != "" {
		data["remark"] = in.Remark
	}

	if len(data) == 0 && hasToggleUpdate {
		// 只更新开关但 data 构造为空（理论上不会发生）
		return nil
	}
	if len(data) == 0 && !hasToggleUpdate {
		return nil
	}

	_, err = dao.TradingRobot.Ctx(ctx).
		Where("id", in.Id).
		Update(data)
	if err != nil {
		return err
	}

	// 若更新了开关，同步更新运行中的 toogo 引擎配置（如果引擎正在跑）
	if hasToggleUpdate {
		var updated *entity.TradingRobot
		_ = dao.TradingRobot.Ctx(ctx).Where("id", in.Id).Scan(&updated)
		if updated != nil {
			toogo.GetRobotTaskManager().UpdateRobot(updated)
		}
	}
	return nil
}

// Delete 删除机器人（软删除）
func (s *robotImpl) Delete(ctx context.Context, in *input.TradingRobotDeleteInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 检查机器人状态
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)

	if err != nil {
		return err
	}
	if robot == nil {
		return gerror.New("机器人不存在或无权限")
	}

	// 运行中的机器人不能删除
	if robot.Status == 2 {
		return gerror.New("运行中的机器人不能删除，请先停止")
	}

	// 检查是否有持仓订单
	count, err := dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().RobotId, in.Id).
		Where(dao.TradingOrder.Columns().Status, 1). // 持仓中
		Count()

	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.Newf("该机器人有%d笔持仓订单，无法删除", count)
	}

	now := gtime.Now()
	// 软删除
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().DeletedAt: now,
		}).
		Update()

	// 【重要】删除机器人时关闭 open run session，避免钱包页“运行中”残留（best effort）
	{
		type sess struct {
			Id        int64       `orm:"id"`
			StartTime *gtime.Time `orm:"start_time"`
		}
		var session *sess
		_ = dao.TradingRobotRunSession.Ctx(ctx).
			Fields("id", "start_time").
			Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
			Where(dao.TradingRobotRunSession.Columns().RobotId, robot.Id).
			WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
			OrderDesc(dao.TradingRobotRunSession.Columns().Id).
			Scan(&session)
		if session != nil && session.Id > 0 {
			runtimeSeconds := 0
			if session.StartTime != nil && !session.StartTime.IsZero() {
				runtimeSeconds = int(now.Sub(session.StartTime).Seconds())
			}
			_, _ = dao.TradingRobotRunSession.Ctx(ctx).
				Where(dao.TradingRobotRunSession.Columns().Id, session.Id).
				WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
				Data(g.Map{
					"end_time":        now,
					"end_reason":      "delete",
					"runtime_seconds": runtimeSeconds,
					"updated_at":      now,
				}).Update()
		}
	}

	return err
}

// View 查看详情
func (s *robotImpl) View(ctx context.Context, in *input.TradingRobotViewInp) (out *input.TradingRobotViewModel, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&out)

	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, gerror.New("机器人不存在")
	}

	// 获取API配置名称
	var apiConfig *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, out.ApiConfigId).
		Scan(&apiConfig)

	if err == nil && apiConfig != nil {
		out.ApiConfigName = apiConfig.ApiName
	}

	return
}

// Start 启动机器人
func (s *robotImpl) Start(ctx context.Context, in *input.TradingRobotStartInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 获取机器人信息
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)

	if err != nil {
		return err
	}
	if robot == nil {
		return gerror.New("机器人不存在或无权限")
	}

	// 检查状态
	if robot.Status == 2 {
		return gerror.New("机器人已经在运行中")
	}
	if robot.Status == 4 {
		return gerror.New("已停用的机器人无法启动")
	}

	// 验证API配置是否可用
	var apiConfig *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).
		Where(dao.TradingApiConfig.Columns().Status, consts.StatusEnabled).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
		Scan(&apiConfig)

	if err != nil {
		return err
	}
	if apiConfig == nil {
		return gerror.New("API配置不可用")
	}

	// 【强一致修复】启动前强制对齐 robot.exchange/robot.symbol 到统一口径（避免平台/币对不一致导致全链路断）
	platform := canonicalPlatform(apiConfig.Platform)
	if platform == "" {
		return gerror.New("API配置平台为空，请检查API配置")
	}
	symbol := canonicalSymbol(robot.Symbol)
	if symbol == "" {
		return gerror.New("机器人交易对为空，请检查机器人配置")
	}

	// 启动前校验策略组：平台/交易对必须一致
	if robot.StrategyGroupId <= 0 {
		return gerror.New("机器人未绑定策略组，无法启动")
	}
	var group *entity.TradingStrategyGroup
	_ = g.DB().Model("hg_trading_strategy_group").Ctx(ctx).
		Where("id", robot.StrategyGroupId).
		Scan(&group)
	if group == nil || group.Id == 0 {
		return gerror.New("机器人绑定的策略组不存在，无法启动")
	}
	if canonicalPlatform(group.Exchange) != platform {
		return gerror.Newf("启动失败：策略组平台(%s)与API平台(%s)不一致", group.Exchange, platform)
	}
	if canonicalSymbol(group.Symbol) != symbol {
		return gerror.Newf("启动失败：策略组交易对(%s)与机器人交易对(%s)不一致", group.Symbol, symbol)
	}

	now := gtime.Now()
	// 更新状态
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().Status:    2, // 运行中
			dao.TradingRobot.Columns().StartTime: now,
			// 对齐口径（修复历史脏数据）
			"exchange": platform,
			"symbol":   symbol,
		}).
		Update()

	if err != nil {
		return err
	}

	// 【重要】admin Start 也要维护 run_session，否则钱包页区间可能缺失/错乱
	// 只在不存在 open session 时插入，避免重复
	{
		cnt, _ := dao.TradingRobotRunSession.Ctx(ctx).
			Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
			Where(dao.TradingRobotRunSession.Columns().RobotId, robot.Id).
			WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
			Count()
		if cnt == 0 {
			_, _ = dao.TradingRobotRunSession.Ctx(ctx).Data(g.Map{
				"robot_id":   robot.Id,
				"user_id":    memberId,
				"exchange":   platform,
				"symbol":     symbol,
				"start_time": now,
			}).Insert()
		}
	}

	g.Log().Infof(ctx, "机器人启动成功: ID=%d, 名称=%s", robot.Id, robot.RobotName)

	return nil
}

// Restart 重启机器人：将“停用(4)”的机器人重新启动为“运行中(2)”
// 注意：Start 明确禁止停用机器人启动；Restart 是显式授权的操作入口。
func (s *robotImpl) Restart(ctx context.Context, in *input.TradingRobotStartInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 获取机器人信息
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)
	if err != nil {
		return err
	}
	if robot == nil {
		return gerror.New("机器人不存在或无权限")
	}

	// 仅允许停用状态重启
	if robot.Status == 2 {
		return gerror.New("机器人已经在运行中")
	}
	if robot.Status != 4 {
		return gerror.New("仅已停用的机器人可以重启")
	}

	// 验证API配置是否可用
	var apiConfig *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).
		Where(dao.TradingApiConfig.Columns().Status, consts.StatusEnabled).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
		Scan(&apiConfig)
	if err != nil {
		return err
	}
	if apiConfig == nil {
		return gerror.New("API配置不可用")
	}

	// 对齐口径（修复历史脏数据）
	platform := canonicalPlatform(apiConfig.Platform)
	if platform == "" {
		return gerror.New("API配置平台为空，请检查API配置")
	}
	symbol := canonicalSymbol(robot.Symbol)
	if symbol == "" {
		return gerror.New("机器人交易对为空，请检查机器人配置")
	}

	// 启动前校验策略组：平台/交易对必须一致
	if robot.StrategyGroupId <= 0 {
		return gerror.New("机器人未绑定策略组，无法重启")
	}
	var group *entity.TradingStrategyGroup
	_ = g.DB().Model("hg_trading_strategy_group").Ctx(ctx).
		Where("id", robot.StrategyGroupId).
		Scan(&group)
	if group == nil || group.Id == 0 {
		return gerror.New("机器人绑定的策略组不存在，无法重启")
	}
	if canonicalPlatform(group.Exchange) != platform {
		return gerror.Newf("重启失败：策略组平台(%s)与API平台(%s)不一致", group.Exchange, platform)
	}
	if canonicalSymbol(group.Symbol) != symbol {
		return gerror.Newf("重启失败：策略组交易对(%s)与机器人交易对(%s)不一致", group.Symbol, symbol)
	}

	now := gtime.Now()
	// 更新状态为运行中
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().Status:    2, // 运行中
			dao.TradingRobot.Columns().StartTime: now,
			"exchange":                          platform,
			"symbol":                            symbol,
		}).
		Update()
	if err != nil {
		return err
	}

	// 维护 run_session（与 Start 对齐）
	{
		cnt, _ := dao.TradingRobotRunSession.Ctx(ctx).
			Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
			Where(dao.TradingRobotRunSession.Columns().RobotId, robot.Id).
			WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
			Count()
		if cnt == 0 {
			_, _ = dao.TradingRobotRunSession.Ctx(ctx).Data(g.Map{
				"robot_id":   robot.Id,
				"user_id":    memberId,
				"exchange":   platform,
				"symbol":     symbol,
				"start_time": now,
			}).Insert()
		}
	}

	g.Log().Infof(ctx, "机器人重启成功: ID=%d, 名称=%s", robot.Id, robot.RobotName)
	return nil
}

// Pause 暂停机器人
func (s *robotImpl) Pause(ctx context.Context, in *input.TradingRobotPauseInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 获取机器人信息
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)

	if err != nil {
		return err
	}
	if robot == nil {
		return gerror.New("机器人不存在或无权限")
	}

	// 检查状态
	if robot.Status != 2 {
		return gerror.New("只能暂停运行中的机器人")
	}

	// TODO: 停止机器人监控进程

	now := gtime.Now()
	// 更新状态
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().Status:    3, // 暂停
			dao.TradingRobot.Columns().PauseTime: now,
		}).
		Update()

	if err != nil {
		return err
	}

	// close open run session (pause)
	{
		type sess struct {
			Id        int64       `orm:"id"`
			StartTime *gtime.Time `orm:"start_time"`
		}
		var session *sess
		_ = dao.TradingRobotRunSession.Ctx(ctx).
			Fields("id", "start_time").
			Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
			Where(dao.TradingRobotRunSession.Columns().RobotId, robot.Id).
			WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
			OrderDesc(dao.TradingRobotRunSession.Columns().Id).
			Scan(&session)
		if session != nil && session.Id > 0 {
			runtimeSeconds := 0
			if session.StartTime != nil && !session.StartTime.IsZero() {
				runtimeSeconds = int(now.Sub(session.StartTime).Seconds())
			}
			_, _ = dao.TradingRobotRunSession.Ctx(ctx).
				Where(dao.TradingRobotRunSession.Columns().Id, session.Id).
				WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
				Data(g.Map{
					"end_time":        now,
					"end_reason":      "pause",
					"runtime_seconds": runtimeSeconds,
					"updated_at":      now,
				}).Update()
		}
	}

	g.Log().Infof(ctx, "机器人暂停成功: ID=%d, 名称=%s", robot.Id, robot.RobotName)

	return nil
}

// Stop 停止机器人
func (s *robotImpl) Stop(ctx context.Context, in *input.TradingRobotStopInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 获取机器人信息
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)

	if err != nil {
		return err
	}
	if robot == nil {
		return gerror.New("机器人不存在或无权限")
	}

	// 检查状态
	if robot.Status == 4 {
		return gerror.New("机器人已经停用")
	}

	// 检查是否有持仓
	count, err := dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().RobotId, in.Id).
		Where(dao.TradingOrder.Columns().Status, 1). // 持仓中
		Count()

	if err != nil {
		return err
	}
	if count > 0 {
		// 【修复】停止前做一次“快速对账”，避免本地残留 OPEN(1) 订单误拦停用
		// 典型现象：交易所已无真实持仓，但本地 hg_trading_order.status 仍为 1（多见于WS断连/重启/同步未运行）
		var apiConfig *entity.TradingApiConfig
		_ = dao.TradingApiConfig.Ctx(ctx).
			Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).
			WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
			Scan(&apiConfig)

		if apiConfig != nil {
			syncCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
			defer cancel()

			ex, exErr := toogo.GetExchangeManager().GetExchangeFromConfig(syncCtx, apiConfig)
			if exErr == nil && ex != nil {
				if positions, pErr := ex.GetPositions(syncCtx, robot.Symbol); pErr == nil {
					hasRealPos := false
					for _, p := range positions {
						if p == nil {
							continue
						}
						// 使用更小的 epsilon，避免小仓位被误判为“无持仓”从而错误修复本地OPEN订单
						if math.Abs(p.PositionAmt) > 1e-9 {
							hasRealPos = true
							break
						}
					}

					if hasRealPos {
						return gerror.New("该机器人在交易所仍有真实持仓，请先平仓后再停用")
					}

					// 交易所无真实持仓：修复本地残留 OPEN 订单为 CLOSED，允许停用
					now := gtime.Now()
					_, _ = dao.TradingOrder.Ctx(ctx).
						Where(dao.TradingOrder.Columns().RobotId, in.Id).
						Where(dao.TradingOrder.Columns().Status, toogo.OrderStatusOpen).
						Data(g.Map{
							"status":       toogo.OrderStatusClosed,
							"close_reason": "停用前快速对账：交易所无持仓，本地残留OPEN已自动修复",
							"close_time":   now,
							"updated_at":   now,
						}).
						Update()

					// 修复后重新计数（理论上应为0）
					count, _ = dao.TradingOrder.Ctx(ctx).
						Where(dao.TradingOrder.Columns().RobotId, in.Id).
						Where(dao.TradingOrder.Columns().Status, toogo.OrderStatusOpen).
						Count()
					if count == 0 {
						// 已修复为0，继续停用
					} else {
						return gerror.Newf("该机器人有%d笔持仓订单，请先平仓（已尝试快速对账修复，但本地仍存在OPEN订单）", count)
					}
				}
			}
		}

		// 交易所查询失败/无法对账时：保持原拦截，避免误放行
		return gerror.Newf("该机器人有%d笔持仓订单，请先平仓（如确认交易所无持仓，请先执行一次订单同步/对账修复）", count)
	}

	// TODO: 停止机器人监控进程

	now := gtime.Now()
	// 更新状态
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().Status:   4, // 停用
			dao.TradingRobot.Columns().StopTime: now,
		}).
		Update()

	if err != nil {
		return err
	}

	// close open run session (stop)
	{
		type sess struct {
			Id        int64       `orm:"id"`
			StartTime *gtime.Time `orm:"start_time"`
		}
		var session *sess
		_ = dao.TradingRobotRunSession.Ctx(ctx).
			Fields("id", "start_time").
			Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
			Where(dao.TradingRobotRunSession.Columns().RobotId, robot.Id).
			WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
			OrderDesc(dao.TradingRobotRunSession.Columns().Id).
			Scan(&session)
		if session != nil && session.Id > 0 {
			runtimeSeconds := 0
			if session.StartTime != nil && !session.StartTime.IsZero() {
				runtimeSeconds = int(now.Sub(session.StartTime).Seconds())
			}
			_, _ = dao.TradingRobotRunSession.Ctx(ctx).
				Where(dao.TradingRobotRunSession.Columns().Id, session.Id).
				WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
				Data(g.Map{
					"end_time":        now,
					"end_reason":      "stop",
					"runtime_seconds": runtimeSeconds,
					"updated_at":      now,
				}).Update()
		}
	}

	g.Log().Infof(ctx, "机器人停用成功: ID=%d, 名称=%s", robot.Id, robot.RobotName)

	return nil
}

// GetStats 获取运行统计
func (s *robotImpl) GetStats(ctx context.Context, in *input.TradingRobotStatsInp) (out *input.TradingRobotStatsModel, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	var robot *entity.TradingRobot
	err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)

	if err != nil {
		return nil, err
	}
	if robot == nil {
		return nil, gerror.New("机器人不存在")
	}

	out = &input.TradingRobotStatsModel{
		Id:              robot.Id,
		RobotName:       robot.RobotName,
		Status:          robot.Status,
		RuntimeSeconds:  robot.RuntimeSeconds,
		MaxRuntime:      robot.MaxRuntime,
		LongCount:       robot.LongCount,
		ShortCount:      robot.ShortCount,
		TotalCount:      robot.LongCount + robot.ShortCount,
		TotalProfit:     robot.TotalProfit,
		MaxProfitTarget: robot.MaxProfitTarget,
		MaxLossAmount:   robot.MaxLossAmount,
		StartTime:       robot.StartTime,
	}

	// 计算盈利完成率
	if robot.MaxProfitTarget > 0 {
		out.ProfitRate = (robot.TotalProfit / robot.MaxProfitTarget) * 100
	}

	// 解析当前策略
	if robot.CurrentStrategy != "" {
		out.CurrentStrategy = gjson.New(robot.CurrentStrategy)
	}

	return
}

// RecommendStrategy 推荐策略
func (s *robotImpl) RecommendStrategy(ctx context.Context, in *input.TradingRobotRecommendStrategyInp) (out *input.TradingRobotRecommendStrategyModel, err error) {
	out, err = s.GetRecommendStrategy(ctx, in.RiskPreference, in.MarketState)
	return
}

// GetRecommendStrategy 获取推荐策略（内部方法）
func (s *robotImpl) GetRecommendStrategy(ctx context.Context, riskPreference, marketState string) (*input.TradingRobotRecommendStrategyModel, error) {
	var strategy *entity.TradingStrategyTemplate
	err := dao.TradingStrategyTemplate.Ctx(ctx).
		Where(dao.TradingStrategyTemplate.Columns().RiskPreference, riskPreference).
		Where(dao.TradingStrategyTemplate.Columns().MarketState, marketState).
		Where(dao.TradingStrategyTemplate.Columns().IsActive, 1).
		Scan(&strategy)

	if err != nil {
		return nil, err
	}
	if strategy == nil {
		return nil, gerror.Newf("未找到匹配的策略: %s + %s", riskPreference, marketState)
	}

	out := &input.TradingRobotRecommendStrategyModel{
		StrategyKey:         strategy.StrategyKey,
		StrategyName:        strategy.StrategyName,
		Description:         strategy.Description,
		MonitorWindow:       strategy.MonitorWindow,
		VolatilityThreshold: strategy.VolatilityThreshold,
		// v2 templates are single-value configs; keep API compatibility by setting Min=Max.
		LeverageMin:             strategy.Leverage,
		LeverageMax:             strategy.Leverage,
		MarginPercentMin:        strategy.MarginPercent,
		MarginPercentMax:        strategy.MarginPercent,
		StopLossPercent:         strategy.StopLossPercent,
		ProfitRetreatPercent:    strategy.ProfitRetreatPercent,
		AutoStartRetreatPercent: strategy.AutoStartRetreatPercent,
		GroupId:                 strategy.GroupId,
	}

	return out, nil
}

// GetRiskConfig returns robot risk config for admin UI.
// Robot.Remark is compatible with 2 formats:
// 1) legacy: a JSON object of map[string]string (MarketRiskMapping)
// 2) new: a JSON object with fields {marketRiskMapping, riskParams}
func (s *robotImpl) GetRiskConfig(ctx context.Context, robotId int64) (*tradingapi.RiskConfig, error) {
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, robotId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)
	if err != nil {
		return nil, err
	}
	if robot == nil {
		return nil, gerror.New("?????????")
	}

	if robot.Remark == "" {
		return &tradingapi.RiskConfig{
			MarketRiskMapping: map[string]string{},
			RiskParams:        map[string]tradingapi.RiskParams{},
		}, nil
	}

	// legacy mapping-only format
	var mapping map[string]string
	if err := json.Unmarshal([]byte(robot.Remark), &mapping); err == nil && mapping != nil {
		return &tradingapi.RiskConfig{
			MarketRiskMapping: mapping,
			RiskParams:        map[string]tradingapi.RiskParams{},
		}, nil
	}

	// new full config format
	var cfg tradingapi.RiskConfig
	if err := json.Unmarshal([]byte(robot.Remark), &cfg); err != nil {
		return nil, gerror.Wrap(err, "remark ??????JSON??????")
	}
	if cfg.MarketRiskMapping == nil {
		cfg.MarketRiskMapping = map[string]string{}
	}
	if cfg.RiskParams == nil {
		cfg.RiskParams = map[string]tradingapi.RiskParams{}
	}
	return &cfg, nil
}

// SaveRiskConfig saves robot risk config.
func (s *robotImpl) SaveRiskConfig(ctx context.Context, robotId int64, cfg *tradingapi.RiskConfig) error {
	if cfg == nil {
		return gerror.New("?????????")
	}
	if cfg.MarketRiskMapping == nil {
		return gerror.New("marketRiskMapping ??????")
	}

	// normalize market state keys for engine compatibility
	normalized := make(map[string]string, len(cfg.MarketRiskMapping))
	for k, v := range cfg.MarketRiskMapping {
		normalized[normalizeMarketState(k)] = v
	}
	cfg.MarketRiskMapping = normalized

	required := []string{"trend", "volatile", "high_vol", "low_vol"}
	var missing []string
	for _, st := range required {
		if _, ok := cfg.MarketRiskMapping[st]; !ok {
			missing = append(missing, st)
		}
	}
	if len(missing) > 0 {
		return gerror.Newf("?????????????????????? %v", missing)
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return gerror.Wrap(err, "??????JSON??????")
	}

	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, robotId).
		Update(g.Map{dao.TradingRobot.Columns().Remark: string(b)})
	if err != nil {
		return err
	}

	// sync into running engine if any
	var updated *entity.TradingRobot
	_ = dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&updated)
	if updated != nil {
		toogo.GetRobotTaskManager().UpdateRobot(updated)
	}
	return nil
}

// GetSignalLogs returns latest signal logs for a robot.
func (s *robotImpl) GetSignalLogs(ctx context.Context, robotId int64, limit int) ([]*tradingapi.SignalLogItem, error) {
	if limit <= 0 {
		limit = 20
	}

	model := g.DB().Model("hg_trading_signal_log").Ctx(ctx)
	if robotId > 0 {
		model = model.Where("robot_id", robotId)
	}

	records, err := model.OrderDesc("id").Limit(limit).All()
	if err != nil {
		return nil, err
	}

	list := make([]*tradingapi.SignalLogItem, 0, len(records))
	for _, r := range records {
		list = append(list, &tradingapi.SignalLogItem{
			Id:             r["id"].Int64(),
			RobotId:        r["robot_id"].Int64(),
			Symbol:         r["symbol"].String(),
			SignalType:     r["signal_type"].String(),
			SignalStrength: r["signal_strength"].Float64(),
			CurrentPrice:   r["current_price"].Float64(),
			WindowMinPrice: r["window_min_price"].Float64(),
			WindowMaxPrice: r["window_max_price"].Float64(),
			Threshold:      r["threshold"].Float64(),
			Reason:         r["reason"].String(),
			MarketState:    r["market_state"].String(),
			RiskPreference: r["risk_preference"].String(),
			Executed:       r["executed"].Int() != 0,
			ExecuteResult:  r["execute_result"].String(),
			IsProcessed:    r["is_processed"].Int() != 0,
			CreatedAt:      r["created_at"].String(),
		})
	}
	return list, nil
}

// GetExecutionLogs returns latest execution logs for a robot.
func (s *robotImpl) GetExecutionLogs(ctx context.Context, robotId int64, limit int) ([]*tradingapi.ExecutionLogItem, error) {
	if limit <= 0 {
		limit = 20
	}

	var logs []*entity.TradingExecutionLog
	err := dao.TradingExecutionLog.Ctx(ctx).
		Where(dao.TradingExecutionLog.Columns().RobotId, robotId).
		OrderDesc(dao.TradingExecutionLog.Columns().Id).
		Limit(limit).
		Scan(&logs)
	if err != nil {
		return nil, err
	}

	list := make([]*tradingapi.ExecutionLogItem, 0, len(logs))
	for _, l := range logs {
		createdAt := ""
		if l.CreatedAt != nil {
			createdAt = l.CreatedAt.String()
		}
		list = append(list, &tradingapi.ExecutionLogItem{
			Id:          l.Id,
			SignalLogId: l.SignalLogId,
			RobotId:     l.RobotId,
			OrderId:     l.OrderId,
			EventType:   l.EventType,
			EventData:   l.EventData,
			Status:      l.Status,
			Message:     l.Message,
			CreatedAt:   createdAt,
		})
	}
	return list, nil
}
