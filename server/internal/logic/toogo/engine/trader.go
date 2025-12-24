// Package engine 机器人引擎模块 - 交易执行（占位实现）
//
// 注意：当前项目主流程使用的是 `internal/logic/toogo/robot_engine.go` 的实现，
// 该 `engine` 子包暂未接入运行链路，但仍会参与编译；因此这里提供最小可编译实现，
// 以避免空文件导致的构建失败。
package engine

import "context"

// RobotTrader 交易执行模块（最小实现）
type RobotTrader struct {
	engine *RobotEngine
}

// NewRobotTrader 创建交易模块
func NewRobotTrader(engine *RobotEngine) *RobotTrader {
	return &RobotTrader{engine: engine}
}

// CheckAndOpenPosition 检查是否应该开仓（占位：不执行交易）
func (t *RobotTrader) CheckAndOpenPosition(ctx context.Context) {
	_ = ctx
	_ = t
}
