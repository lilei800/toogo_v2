// Package trading 预警日志控制器
package trading

import (
	"context"

	"hotgo/api/admin/trading"
	"hotgo/internal/service"
)

var AlertLog = cAlertLog{}

type cAlertLog struct{}

// List 预警日志列表
func (c *cAlertLog) List(ctx context.Context, req *trading.AlertLogListReq) (res *trading.AlertLogListRes, err error) {
	return service.AlertLog().List(ctx, req)
}

// MarketAnalysis 获取市场分析数据
func (c *cAlertLog) MarketAnalysis(ctx context.Context, req *trading.MarketAnalysisReq) (res *trading.MarketAnalysisRes, err error) {
	return service.AlertLog().MarketAnalysis(ctx, req)
}

// DirectionSignal 获取方向信号
func (c *cAlertLog) DirectionSignal(ctx context.Context, req *trading.DirectionSignalReq) (res *trading.DirectionSignalRes, err error) {
	return service.AlertLog().DirectionSignal(ctx, req)
}

// RobotRiskEval 获取机器人风险评估
func (c *cAlertLog) RobotRiskEval(ctx context.Context, req *trading.RobotRiskEvalReq) (res *trading.RobotRiskEvalRes, err error) {
	return service.AlertLog().RobotRiskEval(ctx, req)
}

// RobotStatus 获取机器人实时状态
func (c *cAlertLog) RobotStatus(ctx context.Context, req *trading.RobotStatusReq) (res *trading.RobotStatusRes, err error) {
	return service.AlertLog().RobotStatus(ctx, req)
}

// EngineStatus 获取引擎状态
func (c *cAlertLog) EngineStatus(ctx context.Context, req *trading.EngineStatusReq) (res *trading.EngineStatusRes, err error) {
	return service.AlertLog().EngineStatus(ctx, req)
}

