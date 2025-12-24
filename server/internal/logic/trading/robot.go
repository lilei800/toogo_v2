// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"encoding/json"
	tradingapi "hotgo/api/admin/trading"
	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/library/contexts"
	"hotgo/internal/logic/toogo"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type robotImpl struct{}

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
		"exchange":           in.Exchange,
		"symbol":             in.Symbol,
		"use_monitor_signal": in.UseMonitorSignal,
		"current_strategy":   string(strategyJSON),
		"strategy_group_id":  in.StrategyGroupId,
		"auto_trade_enabled": autoTrade,
		"auto_close_enabled": autoClose,
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
	hasToggleUpdate := in.AutoTradeEnabled != nil || in.AutoCloseEnabled != nil || in.DualSidePosition != nil
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

	// 软删除
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().DeletedAt: gtime.Now(),
		}).
		Update()

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

	// TODO: 实际启动机器人监控进程
	// 这里需要启动goroutine进行行情监控和交易

	// 更新状态
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().Status:    2, // 运行中
			dao.TradingRobot.Columns().StartTime: gtime.Now(),
		}).
		Update()

	if err != nil {
		return err
	}

	g.Log().Infof(ctx, "机器人启动成功: ID=%d, 名称=%s", robot.Id, robot.RobotName)

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

	// 更新状态
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().Status:    3, // 暂停
			dao.TradingRobot.Columns().PauseTime: gtime.Now(),
		}).
		Update()

	if err != nil {
		return err
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
		return gerror.Newf("该机器人有%d笔持仓订单，请先平仓", count)
	}

	// TODO: 停止机器人监控进程

	// 更新状态
	_, err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingRobot.Columns().Status:   4, // 停用
			dao.TradingRobot.Columns().StopTime: gtime.Now(),
		}).
		Update()

	if err != nil {
		return err
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
