// Package crons
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description Toogo机器人引擎定时任务 V2 - 基于新架构
package crons

import (
	"context"
	"hotgo/internal/library/cron"
	"hotgo/internal/logic/toogo"
)

func init() {
	cron.Register(ToogoRobotEngineV2)
}

// ToogoRobotEngineV2 机器人引擎 V2
// 新架构：任务管理器模式，独立管理多个任务
// 此定时任务只负责启动任务管理器，实际执行由任务管理器内部循环完成
var ToogoRobotEngineV2 = &cToogoRobotEngineV2{name: "toogo_robot_engine_v2"}

type cToogoRobotEngineV2 struct {
	name    string
	started bool
}

func (c *cToogoRobotEngineV2) GetName() string {
	return c.name
}

// Execute 执行任务
// 此任务只需执行一次，启动任务管理器后由内部循环接管
func (c *cToogoRobotEngineV2) Execute(ctx context.Context, parser *cron.Parser) (err error) {
	if c.started {
		return nil
	}

	err = toogo.GetRobotTaskManager().Start(ctx)
	if err != nil {
		parser.Logger.Warning(ctx, "cron ToogoRobotEngineV2 Start err:%+v", err)
		return err
	}

	c.started = true
	parser.Logger.Info(ctx, "cron ToogoRobotEngineV2 已启动")
	return nil
}

