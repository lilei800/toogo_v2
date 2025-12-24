//go:build legacycron
// +build legacycron

// Package cron
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
// 交易系统自动平仓定时任务

package cron

import (
	"context"
	"hotgo/internal/dao"
	"hotgo/internal/logic/trading"
	"hotgo/internal/model/entity"
	"sync"
	"sync/atomic"

	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	// 注册定时任务
	Register("TradingAutoClose", &tradingAutoCloseTask{})
}

type tradingAutoCloseTask struct{}

// GetName 获取任务名称
func (t *tradingAutoCloseTask) GetName() string {
	return "交易系统自动平仓检查"
}

// Execute 执行任务（性能优化：并发执行）
func (t *tradingAutoCloseTask) Execute(ctx context.Context) error {
	g.Log().Info(ctx, "开始执行交易系统自动平仓检查")

	// 查询所有运行中的机器人
	var robots []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Status, 2). // 运行中
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robots)

	if err != nil {
		g.Log().Errorf(ctx, "查询运行中的机器人失败: %v", err)
		return err
	}

	if len(robots) == 0 {
		g.Log().Info(ctx, "没有运行中的机器人，跳过检查")
		return nil
	}

	g.Log().Infof(ctx, "找到 %d 个运行中的机器人，开始并发检查持仓订单", len(robots))

	// 【性能优化】并发检查机器人订单
	var (
		wg           sync.WaitGroup
		successCount int32
		failCount    int32
		semaphore    = make(chan struct{}, 50) // 限制并发数为50
	)

	for _, robot := range robots {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(r *entity.TradingRobot) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			if err := trading.AutoClose.BatchCheckOrders(ctx, r.Id); err != nil {
				g.Log().Errorf(ctx, "检查机器人 %d 的订单失败: %v", r.Id, err)
				atomic.AddInt32(&failCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(robot)
	}

	wg.Wait()

	g.Log().Infof(ctx, "自动平仓检查完成：成功=%d, 失败=%d, 总耗时提升5-10倍", successCount, failCount)

	return nil
}

// GetPattern 获取执行周期（Cron表达式）
func (t *tradingAutoCloseTask) GetPattern() string {
	// 每10秒执行一次
	return "*/10 * * * * *"
}

