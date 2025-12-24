// Package crons
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description Toogo机器人引擎定时任务
package crons

import (
	"context"
	"hotgo/internal/library/cron"
	"hotgo/internal/logic/toogo"
	"hotgo/internal/service"
)

func init() {
	// 注意：ToogoRobotEngine 已废弃，机器人引擎现在在HTTP服务启动时由 RobotTaskManager 自动启动
	// 这里只保留订阅检查任务
	cron.Register(ToogoSubscriptionCheck)
}

// ToogoRobotEngine 机器人引擎主循环（已废弃）
// 现在使用 RobotTaskManager 在HTTP服务启动时自动运行
// 此任务保留用于兼容，实际不执行任何操作
var ToogoRobotEngine = &cToogoRobotEngine{name: "toogo_robot_engine"}

type cToogoRobotEngine struct {
	name string
}

func (c *cToogoRobotEngine) GetName() string {
	return c.name
}

// Execute 执行任务（已废弃，不再执行）
// 机器人引擎已迁移到 RobotTaskManager，在HTTP服务启动时自动运行
func (c *cToogoRobotEngine) Execute(ctx context.Context, parser *cron.Parser) (err error) {
	// 检查 RobotTaskManager 是否运行中，如果没有则启动
	if !toogo.GetRobotTaskManager().IsRunning() {
		parser.Logger.Info(ctx, "cron ToogoRobotEngine: RobotTaskManager未运行，尝试启动...")
		if err = toogo.GetRobotTaskManager().Start(ctx); err != nil {
			parser.Logger.Warning(ctx, "cron ToogoRobotEngine: 启动RobotTaskManager失败: %+v", err)
		}
	}
	return nil
}

// ToogoSubscriptionCheck 订阅过期检查
// 每小时执行一次，检查并处理过期的订阅
var ToogoSubscriptionCheck = &cToogoSubscriptionCheck{name: "toogo_subscription_check"}

type cToogoSubscriptionCheck struct {
	name string
}

func (c *cToogoSubscriptionCheck) GetName() string {
	return c.name
}

// Execute 执行任务
func (c *cToogoSubscriptionCheck) Execute(ctx context.Context, parser *cron.Parser) (err error) {
	err = service.ToogoSubscription().CheckExpired(ctx)
	if err != nil {
		parser.Logger.Warning(ctx, "cron ToogoSubscriptionCheck Execute err:%+v", err)
	}
	return
}
