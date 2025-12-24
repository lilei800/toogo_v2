// Package dao 预警日志DAO
package dao

import (
	"hotgo/internal/dao/internal"
)

// TradingMarketStateLog 市场状态预警日志DAO
var TradingMarketStateLog = &tradingMarketStateLogDao{
	internal.NewTradingMarketStateLogDao(),
}

type tradingMarketStateLogDao struct {
	*internal.TradingMarketStateLogDao
}

// TradingRiskPreferenceLog 风险偏好预警日志DAO
var TradingRiskPreferenceLog = &tradingRiskPreferenceLogDao{
	internal.NewTradingRiskPreferenceLogDao(),
}

type tradingRiskPreferenceLogDao struct {
	*internal.TradingRiskPreferenceLogDao
}

// TradingDirectionLog 方向预警日志DAO
var TradingDirectionLog = &tradingDirectionLogDao{
	internal.NewTradingDirectionLogDao(),
}

type tradingDirectionLogDao struct {
	*internal.TradingDirectionLogDao
}

// TradingRobotRealtime 机器人实时状态DAO
var TradingRobotRealtime = &tradingRobotRealtimeDao{
	internal.NewTradingRobotRealtimeDao(),
}

type tradingRobotRealtimeDao struct {
	*internal.TradingRobotRealtimeDao
}
