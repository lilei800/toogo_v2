// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"hotgo/internal/dao"
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/market"
	"hotgo/internal/model/do"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type orderImpl struct{}

// List 获取订单列表
func (s *orderImpl) List(ctx context.Context, in *input.TradingOrderListInp) (list []*input.TradingOrderListModel, totalCount int, err error) {
	mod := dao.TradingOrder.Ctx(ctx)

	// 租户隔离
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		err = gerror.New("用户未登录")
		return
	}
	mod = mod.Where(dao.TradingOrder.Columns().UserId, memberId)

	// 条件筛选
	if in.RobotId > 0 {
		mod = mod.Where(dao.TradingOrder.Columns().RobotId, in.RobotId)
	}
	if in.Symbol != "" {
		mod = mod.Where(dao.TradingOrder.Columns().Symbol, in.Symbol)
	}
	if in.Direction != "" {
		mod = mod.Where(dao.TradingOrder.Columns().Direction, in.Direction)
	}
	if in.Status > 0 {
		mod = mod.Where(dao.TradingOrder.Columns().Status, in.Status)
	}
	if in.OrderSn != "" {
		mod = mod.WhereLike(dao.TradingOrder.Columns().OrderSn, "%"+in.OrderSn+"%")
	}

	totalCount, err = mod.Count()
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return
	}

	// 查询订单列表
	var orders []*entity.TradingOrder
	err = mod.Page(in.Page, in.PageSize).
		Order(dao.TradingOrder.Columns().CreatedAt + " DESC").
		Scan(&orders)

	if err != nil {
		return nil, 0, err
	}

	// 获取机器人名称
	robotIds := make([]int64, 0)
	for _, order := range orders {
		robotIds = append(robotIds, order.RobotId)
	}

	var robots []*entity.TradingRobot
	if len(robotIds) > 0 {
		err = dao.TradingRobot.Ctx(ctx).
			WhereIn(dao.TradingRobot.Columns().Id, robotIds).
			Scan(&robots)

		if err != nil {
			return nil, 0, err
		}
	}

	// 创建机器人ID到名称的映射
	robotMap := make(map[int64]string)
	for _, robot := range robots {
		robotMap[robot.Id] = robot.RobotName
	}

	// 转换为输出模型
	list = make([]*input.TradingOrderListModel, 0, len(orders))
	for _, order := range orders {
		item := &input.TradingOrderListModel{
			Id:                   order.Id,
			RobotId:              order.RobotId,
			RobotName:            robotMap[order.RobotId],
			OrderSn:              order.OrderSn,
			ExchangeOrderId:      order.ExchangeOrderId,
			Symbol:               order.Symbol,
			Direction:            order.Direction,
			OpenPrice:            order.OpenPrice,
			ClosePrice:           order.ClosePrice,
			Quantity:             order.Quantity,
			Leverage:             order.Leverage,
			Margin:               order.Margin,
			RealizedProfit:       order.RealizedProfit,
			UnrealizedProfit:     order.UnrealizedProfit,
			HighestProfit:        order.HighestProfit,
			ProfitRetreatStarted: order.ProfitRetreatStarted,
			OpenTime:             order.OpenTime,
			CloseTime:            order.CloseTime,
			HoldDuration:         order.HoldDuration,
			Status:               order.Status,
			CloseReason:          order.CloseReason,
		}
		list = append(list, item)
	}

	return
}

// View 查看订单详情
func (s *orderImpl) View(ctx context.Context, in *input.TradingOrderViewInp) (out *input.TradingOrderViewModel, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	var order *entity.TradingOrder
	err = dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().Id, in.Id).
		Where(dao.TradingOrder.Columns().UserId, memberId).
		Scan(&order)

	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, gerror.New("订单不存在或无权限")
	}

	// 获取机器人名称
	var robot *entity.TradingRobot
	err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, order.RobotId).
		Scan(&robot)

	if err != nil {
		return nil, err
	}

	out = &input.TradingOrderViewModel{
		TradingOrder: *order,
	}

	if robot != nil {
		out.RobotName = robot.RobotName
	}

	// 如果是持仓中的订单，计算实时数据
	if order.Status == 1 {
		// 当前实时价格：优先使用行情服务的 MarkPrice（缺失则LastPrice），拿不到则回退开仓价
		out.CurrentPrice = order.OpenPrice
		if robot != nil {
			symbol := order.Symbol
			if symbol == "" {
				symbol = robot.Symbol
			}
			if ticker := market.GetMarketServiceManager().GetTicker(robot.Exchange, symbol); ticker != nil {
				if p := ticker.EffectiveMarkPrice(); p > 0 {
					out.CurrentPrice = p
				}
			}
		}

		// 计算盈利百分比
		if order.Margin > 0 {
			out.ProfitPercent = (order.UnrealizedProfit / order.Margin) * 100
		}

		// 计算止损进度（距离触发还有多少空间）
		// TODO: 基于机器人的止损百分比计算

		// 计算回撤进度
		if order.ProfitRetreatStarted == 1 && order.HighestProfit > 0 {
			out.RetreatProgress = ((order.HighestProfit - order.UnrealizedProfit) / order.HighestProfit) * 100
		}

		out.CanManualClose = true
	}

	return
}

// GetPositions 获取持仓订单
func (s *orderImpl) GetPositions(ctx context.Context, in *input.TradingOrderPositionsInp) (list []*input.TradingOrderPositionsModel, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	// 验证机器人所属
	var robot *entity.TradingRobot
	err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, in.RobotId).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)

	if err != nil {
		return nil, err
	}
	if robot == nil {
		return nil, gerror.New("机器人不存在或无权限")
	}

	// 查询持仓订单
	var orders []*entity.TradingOrder
	err = dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().RobotId, in.RobotId).
		Where(dao.TradingOrder.Columns().Status, 1). // 持仓中
		Order(dao.TradingOrder.Columns().OpenTime + " DESC").
		Scan(&orders)

	if err != nil {
		return nil, err
	}

	// 转换为输出模型
	list = make([]*input.TradingOrderPositionsModel, 0, len(orders))
	for _, order := range orders {
		// 当前实时价格：优先使用行情服务的 MarkPrice（缺失则LastPrice），拿不到则回退开仓价
		currentPrice := order.OpenPrice
		if robot != nil {
			symbol := order.Symbol
			if symbol == "" {
				symbol = robot.Symbol
			}
			if ticker := market.GetMarketServiceManager().GetTicker(robot.Exchange, symbol); ticker != nil {
				if p := ticker.EffectiveMarkPrice(); p > 0 {
					currentPrice = p
				}
			}
		}

		// 计算盈利百分比
		profitPercent := 0.0
		if order.Margin > 0 {
			profitPercent = (order.UnrealizedProfit / order.Margin) * 100
		}

		// 计算止损进度
		stopLossProgress := 0.0
		// TODO: 基于机器人配置计算

		// 计算回撤进度
		retreatProgress := 0.0
		if order.ProfitRetreatStarted == 1 && order.HighestProfit > 0 {
			retreatProgress = ((order.HighestProfit - order.UnrealizedProfit) / order.HighestProfit) * 100
		}

		item := &input.TradingOrderPositionsModel{
			Id:                   order.Id,
			OrderSn:              order.OrderSn,
			Symbol:               order.Symbol,
			Direction:            order.Direction,
			OpenPrice:            order.OpenPrice,
			CurrentPrice:         currentPrice,
			Quantity:             order.Quantity,
			Leverage:             order.Leverage,
			Margin:               order.Margin,
			UnrealizedProfit:     order.UnrealizedProfit,
			HighestProfit:        order.HighestProfit,
			ProfitPercent:        profitPercent,
			StopLossPrice:        order.StopLossPrice,
			ProfitRetreatStarted: order.ProfitRetreatStarted,
			ProfitRetreatPercent: order.ProfitRetreatPercent,
			OpenTime:             order.OpenTime,
			HoldDuration:         order.HoldDuration,
			StopLossProgress:     stopLossProgress,
			RetreatProgress:      retreatProgress,
		}
		list = append(list, item)
	}

	return
}

// ManualClose 手动平仓
func (s *orderImpl) ManualClose(ctx context.Context, in *input.TradingOrderManualCloseInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 获取订单信息
	var order *entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().Id, in.Id).
		Where(dao.TradingOrder.Columns().UserId, memberId).
		Scan(&order)

	if err != nil {
		return err
	}
	if order == nil {
		return gerror.New("订单不存在或无权限")
	}

	// 检查状态
	if order.Status != 1 {
		return gerror.New("只能平仓持仓中的订单")
	}

	// TODO: 调用交易所API执行平仓

	// 模拟平仓
	closePrice := order.OpenPrice * 1.01 // 模拟价格
	closeTime := gtime.Now()

	// 计算实际盈亏
	var realizedProfit float64
	if order.Direction == "long" {
		realizedProfit = (closePrice - order.OpenPrice) * order.Quantity * float64(order.Leverage)
	} else {
		realizedProfit = (order.OpenPrice - closePrice) * order.Quantity * float64(order.Leverage)
	}

	// 计算持仓时长
	holdDuration := int(closeTime.Sub(order.OpenTime).Seconds())

	// 开始事务
	err = dao.TradingOrder.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 更新订单状态
		_, err = dao.TradingOrder.Ctx(ctx).
			Where(dao.TradingOrder.Columns().Id, in.Id).
			Data(g.Map{
				dao.TradingOrder.Columns().ClosePrice:      closePrice,
				dao.TradingOrder.Columns().CloseTime:       closeTime,
				dao.TradingOrder.Columns().RealizedProfit:  realizedProfit,
				dao.TradingOrder.Columns().HoldDuration:    holdDuration,
				dao.TradingOrder.Columns().Status:          2, // 已平仓
				dao.TradingOrder.Columns().CloseReason:     "manual",
			}).
			Update()

		if err != nil {
			return err
		}

		// 记录平仓日志
		logData := &do.TradingCloseLog{
			TenantId:       order.TenantId,
			UserId:         order.UserId,
			RobotId:        order.RobotId,
			OrderId:        order.Id,
			OrderSn:        order.OrderSn,
			Symbol:         order.Symbol,
			Direction:      order.Direction,
			OpenPrice:      order.OpenPrice,
			ClosePrice:     closePrice,
			Quantity:       order.Quantity,
			Leverage:       order.Leverage,
			Margin:         order.Margin,
			RealizedProfit: realizedProfit,
			HighestProfit:  order.HighestProfit,
			ProfitPercent:  (realizedProfit / order.Margin) * 100,
			CloseReason:    "manual",
			CloseDetail:    `{"type":"手动平仓","operator":"user"}`,
			OpenFee:        0,   // TODO: 计算费用
			HoldFee:        0,   // TODO: 计算费用
			CloseFee:       0,   // TODO: 计算费用
			TotalFee:       0,   // TODO: 计算费用
			CommissionAmount: 0, // TODO: 计算佣金
			CommissionPercent: 0,
			NetProfit:      realizedProfit,
			OpenTime:       order.OpenTime,
			CloseTime:      closeTime,
			HoldDuration:   holdDuration,
		}

		_, err = dao.TradingCloseLog.Ctx(ctx).Data(logData).Insert()
		if err != nil {
			return err
		}

		// 更新机器人统计
		var robot *entity.TradingRobot
		err = dao.TradingRobot.Ctx(ctx).
			Where(dao.TradingRobot.Columns().Id, order.RobotId).
			Scan(&robot)

		if err != nil {
			return err
		}

		if robot != nil {
			newTotalProfit := robot.TotalProfit + realizedProfit

			_, err = dao.TradingRobot.Ctx(ctx).
				Where(dao.TradingRobot.Columns().Id, order.RobotId).
				Data(g.Map{
					dao.TradingRobot.Columns().TotalProfit: newTotalProfit,
				}).
				Update()

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	g.Log().Infof(ctx, "手动平仓成功: OrderID=%d, 盈亏=%.4f", order.Id, realizedProfit)

	return nil
}

// GetStats 获取订单统计
func (s *orderImpl) GetStats(ctx context.Context, in *input.TradingOrderStatsInp) (out *input.TradingOrderStatsModel, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	mod := dao.TradingOrder.Ctx(ctx).Where(dao.TradingOrder.Columns().UserId, memberId)

	// 条件筛选
	if in.RobotId > 0 {
		mod = mod.Where(dao.TradingOrder.Columns().RobotId, in.RobotId)
	}
	if in.StartDate != "" {
		mod = mod.WhereGTE(dao.TradingOrder.Columns().CreatedAt, in.StartDate)
	}
	if in.EndDate != "" {
		mod = mod.WhereLTE(dao.TradingOrder.Columns().CreatedAt, in.EndDate+" 23:59:59")
	}

	// 统计总数
	totalCount, err := mod.Count()
	if err != nil {
		return nil, err
	}

	// 统计多单数
	longCount, err := mod.Clone().Where(dao.TradingOrder.Columns().Direction, "long").Count()
	if err != nil {
		return nil, err
	}

	// 统计空单数
	shortCount, err := mod.Clone().Where(dao.TradingOrder.Columns().Direction, "short").Count()
	if err != nil {
		return nil, err
	}

	// 统计持仓中数量
	positionCount, err := mod.Clone().Where(dao.TradingOrder.Columns().Status, 1).Count()
	if err != nil {
		return nil, err
	}

	// 统计已平仓数量
	closedCount, err := mod.Clone().Where(dao.TradingOrder.Columns().Status, 2).Count()
	if err != nil {
		return nil, err
	}

	// 统计盈利和亏损
	type ProfitStats struct {
		ProfitCount int     `json:"profitCount"`
		LossCount   int     `json:"lossCount"`
		TotalProfit float64 `json:"totalProfit"`
		TotalLoss   float64 `json:"totalLoss"`
	}

	var profitStats ProfitStats
	err = mod.Clone().
		Where(dao.TradingOrder.Columns().Status, 2).
		Fields(
			"SUM(CASE WHEN realized_profit > 0 THEN 1 ELSE 0 END) as profit_count",
			"SUM(CASE WHEN realized_profit < 0 THEN 1 ELSE 0 END) as loss_count",
			"SUM(CASE WHEN realized_profit > 0 THEN realized_profit ELSE 0 END) as total_profit",
			"ABS(SUM(CASE WHEN realized_profit < 0 THEN realized_profit ELSE 0 END)) as total_loss",
		).
		Scan(&profitStats)

	if err != nil {
		return nil, err
	}

	// 计算平均持仓时长
	var avgDuration struct {
		AvgHoldDuration float64 `json:"avgHoldDuration"`
	}
	err = mod.Clone().
		Where(dao.TradingOrder.Columns().Status, 2).
		Fields("AVG(hold_duration) as avg_hold_duration").
		Scan(&avgDuration)

	if err != nil {
		return nil, err
	}

	out = &input.TradingOrderStatsModel{
		TotalCount:      totalCount,
		LongCount:       longCount,
		ShortCount:      shortCount,
		PositionCount:   positionCount,
		ClosedCount:     closedCount,
		ProfitCount:     profitStats.ProfitCount,
		LossCount:       profitStats.LossCount,
		TotalProfit:     profitStats.TotalProfit,
		TotalLoss:       profitStats.TotalLoss,
		NetProfit:       profitStats.TotalProfit - profitStats.TotalLoss,
		AvgHoldDuration: int(avgDuration.AvgHoldDuration),
	}

	// 计算胜率
	if closedCount > 0 {
		out.WinRate = (float64(profitStats.ProfitCount) / float64(closedCount)) * 100
	}

	// 计算平均盈利和亏损
	if profitStats.ProfitCount > 0 {
		out.AvgProfit = profitStats.TotalProfit / float64(profitStats.ProfitCount)
	}
	if profitStats.LossCount > 0 {
		out.AvgLoss = profitStats.TotalLoss / float64(profitStats.LossCount)
	}

	// 计算盈亏比
	if profitStats.TotalLoss > 0 {
		out.ProfitFactor = profitStats.TotalProfit / profitStats.TotalLoss
	}

	return
}

// GetCloseLogs 获取平仓日志列表
func (s *orderImpl) GetCloseLogs(ctx context.Context, in *input.TradingOrderCloseLogListInp) (list []*input.TradingOrderCloseLogListModel, totalCount int, err error) {
	mod := dao.TradingCloseLog.Ctx(ctx)

	// 租户隔离
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		err = gerror.New("用户未登录")
		return
	}
	mod = mod.Where(dao.TradingCloseLog.Columns().UserId, memberId)

	// 条件筛选
	if in.RobotId > 0 {
		mod = mod.Where(dao.TradingCloseLog.Columns().RobotId, in.RobotId)
	}
	if in.Symbol != "" {
		mod = mod.Where(dao.TradingCloseLog.Columns().Symbol, in.Symbol)
	}
	if in.Direction != "" {
		mod = mod.Where(dao.TradingCloseLog.Columns().Direction, in.Direction)
	}
	if in.CloseReason != "" {
		mod = mod.Where(dao.TradingCloseLog.Columns().CloseReason, in.CloseReason)
	}
	if in.StartDate != "" {
		mod = mod.WhereGTE(dao.TradingCloseLog.Columns().CloseTime, in.StartDate)
	}
	if in.EndDate != "" {
		mod = mod.WhereLTE(dao.TradingCloseLog.Columns().CloseTime, in.EndDate+" 23:59:59")
	}

	totalCount, err = mod.Count()
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return
	}

	// 查询日志列表
	var logs []*entity.TradingCloseLog
	err = mod.Page(in.Page, in.PageSize).
		Order(dao.TradingCloseLog.Columns().CloseTime + " DESC").
		Scan(&logs)

	if err != nil {
		return nil, 0, err
	}

	// 获取机器人名称
	robotIds := make([]int64, 0)
	for _, log := range logs {
		robotIds = append(robotIds, log.RobotId)
	}

	var robots []*entity.TradingRobot
	if len(robotIds) > 0 {
		err = dao.TradingRobot.Ctx(ctx).
			WhereIn(dao.TradingRobot.Columns().Id, robotIds).
			Scan(&robots)

		if err != nil {
			return nil, 0, err
		}
	}

	robotMap := make(map[int64]string)
	for _, robot := range robots {
		robotMap[robot.Id] = robot.RobotName
	}

	// 转换为输出模型
	list = make([]*input.TradingOrderCloseLogListModel, 0, len(logs))
	for _, log := range logs {
		item := &input.TradingOrderCloseLogListModel{
			Id:               log.Id,
			RobotId:          log.RobotId,
			RobotName:        robotMap[log.RobotId],
			OrderSn:          log.OrderSn,
			Symbol:           log.Symbol,
			Direction:        log.Direction,
			OpenPrice:        log.OpenPrice,
			ClosePrice:       log.ClosePrice,
			Quantity:         log.Quantity,
			Leverage:         log.Leverage,
			Margin:           log.Margin,
			RealizedProfit:   log.RealizedProfit,
			HighestProfit:    log.HighestProfit,
			ProfitPercent:    log.ProfitPercent,
			CloseReason:      log.CloseReason,
			TotalFee:         log.TotalFee,
			CommissionAmount: log.CommissionAmount,
			NetProfit:        log.NetProfit,
			OpenTime:         log.OpenTime,
			CloseTime:        log.CloseTime,
			HoldDuration:     log.HoldDuration,
		}
		list = append(list, item)
	}

	return
}

// ViewCloseLog 查看平仓日志详情
func (s *orderImpl) ViewCloseLog(ctx context.Context, in *input.TradingOrderCloseLogViewInp) (out *input.TradingOrderCloseLogViewModel, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	var log *entity.TradingCloseLog
	err = dao.TradingCloseLog.Ctx(ctx).
		Where(dao.TradingCloseLog.Columns().Id, in.Id).
		Where(dao.TradingCloseLog.Columns().UserId, memberId).
		Scan(&log)

	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, gerror.New("平仓日志不存在或无权限")
	}

	// 获取机器人名称
	var robot *entity.TradingRobot
	err = dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, log.RobotId).
		Scan(&robot)

	if err != nil {
		return nil, err
	}

	out = &input.TradingOrderCloseLogViewModel{
		TradingCloseLog: *log,
	}

	if robot != nil {
		out.RobotName = robot.RobotName
	}

	return
}

