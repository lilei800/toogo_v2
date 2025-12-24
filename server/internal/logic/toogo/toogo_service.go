// Package toogo Toogo总服务实现
package toogo

import (
	"context"
	"hotgo/internal/service"
)

type sToogo struct{}

func init() {
	service.RegisterToogo(NewToogo())
}

func NewToogo() *sToogo {
	return &sToogo{}
}

// GetRobotTaskManager 获取机器人任务管理器
func (s *sToogo) GetRobotTaskManager() service.IToogoTaskManager {
	return &sRobotTaskManagerWrapper{}
}

// sRobotTaskManagerWrapper RobotTaskManager包装器，实现IToogoTaskManager接口
type sRobotTaskManagerWrapper struct{}

// ReloadRobotStrategy 重新加载机器人策略配置
func (s *sRobotTaskManagerWrapper) ReloadRobotStrategy(ctx context.Context, robotId int64) error {
	return GetRobotTaskManager().ReloadRobotStrategy(ctx, robotId)
}

// Start 启动任务管理器
func (s *sRobotTaskManagerWrapper) Start(ctx context.Context) error {
	return GetRobotTaskManager().Start(ctx)
}

// Stop 停止任务管理器
func (s *sRobotTaskManagerWrapper) Stop() {
	GetRobotTaskManager().Stop()
}

// IsRunning 检查是否运行中
func (s *sRobotTaskManagerWrapper) IsRunning() bool {
	return GetRobotTaskManager().IsRunning()
}

