// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
// @Description 订单同步服务 - 防止手动平仓逃避算力消耗

package toogo

import (
	"context"

	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SyncClosedOrders 同步已平仓订单并补扣算力（已禁用）
// 说明：按产品需求不再扣除“盈利订单算力”，该任务保留空实现以兼容历史调度/接口。
func (s *sToogoRobot) SyncClosedOrders(ctx context.Context) error {
	g.Log().Info(ctx, "[OrderSync] 已禁用：不再对盈利订单补扣算力，跳过执行")
	return nil
}

// consumeOrderPower 为订单消耗算力
func (s *sToogoRobot) consumeOrderPower(ctx context.Context, order *entity.TradingOrder) (float64, error) {
	// 获取机器人信息
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where("id", order.RobotId).Scan(&robot)
	if err != nil {
		return 0, gerror.Wrap(err, "查询机器人失败")
	}
	if robot == nil {
		return 0, gerror.New("机器人不存在")
	}

	// 获取系统配置的消耗比例
	consumeRate := 0.10 // 默认10%
	configVal, err := g.Cfg().Get(ctx, "toogo.powerConsumePercent")
	if err == nil && !configVal.IsEmpty() {
		consumeRate = configVal.Float64() / 100
	}

	// 计算应消耗的算力（未考虑VIP折扣，因为订单已平仓）
	powerAmount := order.RealizedProfit * consumeRate

	// 调用钱包服务消耗算力
	err = service.ToogoWallet().ConsumePower(ctx,
		order.UserId,
		order.RobotId,
		order.Id,
		order.OrderSn,
		order.RealizedProfit,
	)

	if err != nil {
		// 【重要】算力扣除失败时记录日志，但不影响平仓操作
		// 注意：由于已修改 ChangeBalance 允许负数，这里理论上不应该失败
		// 即使失败，也不应该影响订单的平仓状态
		g.Log().Warningf(ctx, "[OrderSync] 用户 %d 订单 %s 扣除算力失败: %v（不影响平仓）",
			order.UserId, order.OrderSn, err)
		// 不再记录欠费，因为允许负算力
		// 返回成功，确保不影响平仓操作
		return powerAmount, nil
	}

	// 更新订单状态（标记算力已扣除）
	// 【重要】即使算力扣除失败，也尝试更新订单状态，不影响平仓操作
	_, err = dao.TradingOrder.Ctx(ctx).
		Where("id", order.Id).
		Data(g.Map{
			"power_consumed": 1,
			"power_amount":   powerAmount,
			"updated_at":     gtime.Now(),
		}).
		Update()

	if err != nil {
		// 更新订单状态失败时记录日志，但不影响平仓操作
		g.Log().Errorf(ctx, "[OrderSync] 更新订单状态失败（不影响平仓）: %v", err)
		// 不返回错误，确保不影响平仓操作
		return powerAmount, nil
	}

	g.Log().Infof(ctx, "[OrderSync] 订单 %s 补扣算力成功，盈利: %.4f USDT，消耗算力: %.4f",
		order.OrderSn, order.RealizedProfit, powerAmount)

	return powerAmount, nil
}

// recordDebt 记录欠费（可选实现）
func (s *sToogoRobot) recordDebt(ctx context.Context, userId int64, orderId int64, amount float64, reason string) {
	// TODO: 实现欠费记录表
	// 可以创建 toogo_debt 表记录用户欠费情况
	g.Log().Warningf(ctx, "[OrderSync] 用户 %d 产生欠费，金额: %.4f，原因: %s",
		userId, amount, reason)

	// 示例：插入欠费记录
	/*
		_, _ = dao.ToogoDebt.Ctx(ctx).Insert(g.Map{
			"user_id":    userId,
			"order_id":   orderId,
			"amount":     amount,
			"reason":     reason,
			"status":     0, // 0=未还清
			"created_at": gtime.Now(),
		})
	*/
}

// GetOrderSyncStats 获取订单同步统计
func (s *sToogoRobot) GetOrderSyncStats(ctx context.Context) (map[string]interface{}, error) {
	// 统计未消耗算力的盈利订单
	var stats struct {
		TotalOrders   int     `json:"totalOrders"`
		NotConsumed   int     `json:"notConsumed"`
		TotalProfit   float64 `json:"totalProfit"`
		EstimatePower float64 `json:"estimatePower"`
	}

	// 总订单数
	count, _ := dao.TradingOrder.Ctx(ctx).
		Where("status", 2).
		Where("realized_profit > 0").
		Count()
	stats.TotalOrders = count

	// 未消耗算力订单数
	notConsumedCount, _ := dao.TradingOrder.Ctx(ctx).
		Where("status", 2).
		Where("power_consumed", 0).
		Where("realized_profit > 0").
		Count()
	stats.NotConsumed = notConsumedCount

	// 未消耗订单的总盈利
	totalProfit, _ := dao.TradingOrder.Ctx(ctx).
		Where("status", 2).
		Where("power_consumed", 0).
		Where("realized_profit > 0").
		Sum("realized_profit")
	stats.TotalProfit = totalProfit

	// 预估应补扣算力
	stats.EstimatePower = totalProfit * 0.10 // 默认10%

	return g.Map{
		"totalOrders":   stats.TotalOrders,
		"notConsumed":   stats.NotConsumed,
		"totalProfit":   stats.TotalProfit,
		"estimatePower": stats.EstimatePower,
	}, nil
}
