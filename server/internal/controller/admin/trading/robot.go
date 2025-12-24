// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"hotgo/api/admin/trading"
	tradingLogic "hotgo/internal/logic/trading"
	"hotgo/internal/service"
)

var Robot = cRobot{}

type cRobot struct{}

// List 获取机器人列表
func (c *cRobot) List(ctx context.Context, req *trading.RobotListReq) (res *trading.RobotListRes, err error) {
	list, totalCount, err := tradingLogic.Robot.List(ctx, &req.TradingRobotListInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotListRes{
		List:       list,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	return
}

// Create 创建机器人
func (c *cRobot) Create(ctx context.Context, req *trading.RobotCreateReq) (res *trading.RobotCreateRes, err error) {
	id, err := tradingLogic.Robot.Create(ctx, &req.TradingRobotCreateInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotCreateRes{Id: id}
	return
}

// Update 更新机器人
func (c *cRobot) Update(ctx context.Context, req *trading.RobotUpdateReq) (res *trading.RobotUpdateRes, err error) {
	err = tradingLogic.Robot.Update(ctx, &req.TradingRobotUpdateInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotUpdateRes{}
	return
}

// Delete 删除机器人
func (c *cRobot) Delete(ctx context.Context, req *trading.RobotDeleteReq) (res *trading.RobotDeleteRes, err error) {
	err = tradingLogic.Robot.Delete(ctx, &req.TradingRobotDeleteInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotDeleteRes{}
	return
}

// View 查看机器人详情
func (c *cRobot) View(ctx context.Context, req *trading.RobotViewReq) (res *trading.RobotViewRes, err error) {
	out, err := tradingLogic.Robot.View(ctx, &req.TradingRobotViewInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotViewRes{
		TradingRobotViewModel: out,
	}
	return
}

// Start 启动机器人
func (c *cRobot) Start(ctx context.Context, req *trading.RobotStartReq) (res *trading.RobotStartRes, err error) {
	err = tradingLogic.Robot.Start(ctx, &req.TradingRobotStartInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotStartRes{}
	return
}

// Pause 暂停机器人
func (c *cRobot) Pause(ctx context.Context, req *trading.RobotPauseReq) (res *trading.RobotPauseRes, err error) {
	err = tradingLogic.Robot.Pause(ctx, &req.TradingRobotPauseInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotPauseRes{}
	return
}

// Stop 停止机器人
func (c *cRobot) Stop(ctx context.Context, req *trading.RobotStopReq) (res *trading.RobotStopRes, err error) {
	err = tradingLogic.Robot.Stop(ctx, &req.TradingRobotStopInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotStopRes{}
	return
}

// Stats 获取运行统计
func (c *cRobot) Stats(ctx context.Context, req *trading.RobotStatsReq) (res *trading.RobotStatsRes, err error) {
	out, err := tradingLogic.Robot.GetStats(ctx, &req.TradingRobotStatsInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotStatsRes{
		TradingRobotStatsModel: out,
	}
	return
}

// RecommendStrategy 推荐策略
func (c *cRobot) RecommendStrategy(ctx context.Context, req *trading.RobotRecommendStrategyReq) (res *trading.RobotRecommendStrategyRes, err error) {
	out, err := tradingLogic.Robot.RecommendStrategy(ctx, &req.TradingRobotRecommendStrategyInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotRecommendStrategyRes{
		TradingRobotRecommendStrategyModel: out,
	}
	return
}

// Positions 获取机器人持仓
func (c *cRobot) Positions(ctx context.Context, req *trading.RobotPositionsReq) (res *trading.RobotPositionsRes, err error) {
	list, err := service.ToogoRobot().GetRobotPositions(ctx, req.RobotId)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotPositionsRes{
		List: list,
	}
	return
}

// Orders 获取机器人挂单
func (c *cRobot) Orders(ctx context.Context, req *trading.RobotOrdersReq) (res *trading.RobotOrdersRes, err error) {
	list, err := service.ToogoRobot().GetRobotOpenOrders(ctx, req.RobotId)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotOrdersRes{
		List: list,
	}
	return
}

// OrderHistory 获取机器人历史订单
func (c *cRobot) OrderHistory(ctx context.Context, req *trading.RobotOrderHistoryReq) (res *trading.RobotOrderHistoryRes, err error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	list, err := service.ToogoRobot().GetRobotOrderHistory(ctx, req.RobotId, limit)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotOrderHistoryRes{
		List: list,
	}
	return
}

// ClosePosition 手动平仓
func (c *cRobot) ClosePosition(ctx context.Context, req *trading.RobotClosePositionReq) (res *trading.RobotClosePositionRes, err error) {
	err = service.ToogoRobot().CloseRobotPosition(ctx, &req.ClosePositionInp)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotClosePositionRes{}
	return
}

// CancelOrder 撤销挂单
func (c *cRobot) CancelOrder(ctx context.Context, req *trading.RobotCancelOrderReq) (res *trading.RobotCancelOrderRes, err error) {
	err = service.ToogoRobot().CancelRobotOrder(ctx, req.RobotId, req.OrderId)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotCancelOrderRes{}
	return
}

// SetTakeProfitSwitch 设置止盈回撤开关
func (c *cRobot) SetTakeProfitSwitch(ctx context.Context, req *trading.RobotSetTakeProfitSwitchReq) (res *trading.RobotSetTakeProfitSwitchRes, err error) {
	err = service.ToogoRobot().SetTakeProfitRetreatSwitch(ctx, req.RobotId, req.PositionSide, req.Enabled)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotSetTakeProfitSwitchRes{}
	return
}

// RiskConfig 获取风险配置
func (c *cRobot) RiskConfig(ctx context.Context, req *trading.RobotRiskConfigReq) (res *trading.RobotRiskConfigRes, err error) {
	config, err := tradingLogic.Robot.GetRiskConfig(ctx, req.RobotId)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotRiskConfigRes{
		Config: config,
	}
	return
}

// SaveRiskConfig 保存风险配置
func (c *cRobot) SaveRiskConfig(ctx context.Context, req *trading.RobotSaveRiskConfigReq) (res *trading.RobotSaveRiskConfigRes, err error) {
	err = tradingLogic.Robot.SaveRiskConfig(ctx, req.RobotId, req.Config)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotSaveRiskConfigRes{}
	return
}

// SignalLogs 获取方向预警日志
func (c *cRobot) SignalLogs(ctx context.Context, req *trading.RobotSignalLogsReq) (res *trading.RobotSignalLogsRes, err error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	list, err := tradingLogic.Robot.GetSignalLogs(ctx, req.RobotId, limit)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotSignalLogsRes{
		List: list,
	}
	return
}

// ExecutionLogs 获取交易执行日志
func (c *cRobot) ExecutionLogs(ctx context.Context, req *trading.RobotExecutionLogsReq) (res *trading.RobotExecutionLogsRes, err error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	list, err := tradingLogic.Robot.GetExecutionLogs(ctx, req.RobotId, limit)
	if err != nil {
		return nil, err
	}

	res = &trading.RobotExecutionLogsRes{
		List: list,
	}
	return
}

// ReloadStrategy 重新加载策略配置
func (c *cRobot) ReloadStrategy(ctx context.Context, req *trading.RobotReloadStrategyReq) (res *trading.RobotReloadStrategyRes, err error) {
	// 调用机器人任务管理器重新加载策略
	err = service.Toogo().GetRobotTaskManager().ReloadRobotStrategy(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &trading.RobotReloadStrategyRes{}, nil
}
