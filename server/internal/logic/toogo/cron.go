// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
// @Description 定时任务管理

package toogo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

// RegisterOrderSyncCron 注册订单同步定时任务
func RegisterOrderSyncCron(ctx context.Context) error {
	// 已禁用：不再补扣“盈利订单算力”，因此不再注册该定时任务。
	// 保留函数用于兼容历史调用。
	g.Log().Info(ctx, "[OrderSync] 已禁用：不再注册订单同步补扣算力定时任务")
	return nil
}

// RegisterAllCronTasks 注册所有定时任务
func RegisterAllCronTasks(ctx context.Context) error {
	// 1. 注册订单同步任务
	if err := RegisterOrderSyncCron(ctx); err != nil {
		return err
	}

	// 2. 其他定时任务可以在这里添加
	// ...

	g.Log().Info(ctx, "[Cron] 所有定时任务注册完成")
	return nil
}

// StopAllCronTasks 停止所有定时任务
func StopAllCronTasks(ctx context.Context) {
	gcron.Stop("OrderSyncTask")
	g.Log().Info(ctx, "[Cron] 所有定时任务已停止")
}
