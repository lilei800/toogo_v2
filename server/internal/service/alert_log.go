// Package service 预警日志服务
package service

import (
	"context"
	"hotgo/api/admin/trading"
)

// IAlertLog 预警日志服务接口
type IAlertLog interface {
	// List 预警日志列表
	List(ctx context.Context, in *trading.AlertLogListReq) (*trading.AlertLogListRes, error)
	// MarketAnalysis 获取市场分析数据
	MarketAnalysis(ctx context.Context, in *trading.MarketAnalysisReq) (*trading.MarketAnalysisRes, error)
	// DirectionSignal 获取方向信号
	DirectionSignal(ctx context.Context, in *trading.DirectionSignalReq) (*trading.DirectionSignalRes, error)
	// RobotRiskEval 获取机器人风险评估
	RobotRiskEval(ctx context.Context, in *trading.RobotRiskEvalReq) (*trading.RobotRiskEvalRes, error)
	// RobotStatus 获取机器人实时状态
	RobotStatus(ctx context.Context, in *trading.RobotStatusReq) (*trading.RobotStatusRes, error)
	// EngineStatus 获取引擎状态
	EngineStatus(ctx context.Context, in *trading.EngineStatusReq) (*trading.EngineStatusRes, error)
}

var localAlertLog IAlertLog

func AlertLog() IAlertLog {
	if localAlertLog == nil {
		panic("implement not found for interface IAlertLog, forgot register?")
	}
	return localAlertLog
}

func RegisterAlertLog(i IAlertLog) {
	localAlertLog = i
}

