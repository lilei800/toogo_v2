// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"encoding/json"
	"hotgo/internal/dao"
	"hotgo/internal/model/do"
	"hotgo/internal/model/entity"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type autoCloseImpl struct{}

var AutoClose = &autoCloseImpl{}

// CloseStrategy 平仓策略配置
type CloseStrategy struct {
	StopLossPercent         float64 `json:"stopLossPercent"`         // 止损百分比
	ProfitRetreatPercent    float64 `json:"profitRetreatPercent"`    // 止盈回撤百分比
	AutoStartRetreatPercent float64 `json:"autoStartRetreatPercent"` // 启动回撤百分比
}

// CloseDecision 平仓决策结果
type CloseDecision struct {
	ShouldClose  bool    `json:"shouldClose"`  // 是否应该平仓
	CloseReason  string  `json:"closeReason"`  // 平仓原因
	ClosePrice   float64 `json:"closePrice"`   // 平仓价格
	ReasonDetail string  `json:"reasonDetail"` // 原因详情
}

// CheckOrder 检查订单是否需要平仓
func (s *autoCloseImpl) CheckOrder(ctx context.Context, order *entity.TradingOrder, currentPrice float64, strategy *CloseStrategy) (*CloseDecision, error) {
	decision := &CloseDecision{
		ShouldClose: false,
		ClosePrice:  currentPrice,
	}

	// 计算当前盈亏
	var currentProfit float64
	if order.Direction == "long" {
		currentProfit = (currentPrice - order.OpenPrice) * order.Quantity * float64(order.Leverage)
	} else {
		currentProfit = (order.OpenPrice - currentPrice) * order.Quantity * float64(order.Leverage)
	}

	// 计算盈亏百分比
	profitPercent := (currentProfit / order.Margin) * 100

	// 1. 检查止损
	if profitPercent <= -strategy.StopLossPercent {
		decision.ShouldClose = true
		decision.CloseReason = "stop_loss"
		detailData := map[string]interface{}{
			"type":          "止损",
			"profitPercent": profitPercent,
			"stopLossLine":  -strategy.StopLossPercent,
			"currentProfit": currentProfit,
		}
		detailBytes, _ := json.Marshal(detailData)
		decision.ReasonDetail = string(detailBytes)
		return decision, nil
	}

	// 2. 检查止盈回撤
	// 首先判断是否达到启动回撤的条件
	if profitPercent >= strategy.AutoStartRetreatPercent {
		// 更新最高盈利
		if currentProfit > order.HighestProfit {
			err := s.updateHighestProfit(ctx, order.Id, currentProfit, profitPercent)
			if err != nil {
				g.Log().Errorf(ctx, "更新最高盈利失败: %v", err)
			}
		}

		// 启动止盈回撤
		if order.ProfitRetreatStarted != 1 {
			err := s.startProfitRetreat(ctx, order.Id)
			if err != nil {
				g.Log().Errorf(ctx, "启动止盈回撤失败: %v", err)
			}
		}

		// 检查是否触发回撤平仓
		if order.HighestProfit > 0 {
			retreatPercent := ((order.HighestProfit - currentProfit) / order.HighestProfit) * 100
			if retreatPercent >= strategy.ProfitRetreatPercent {
				decision.ShouldClose = true
				decision.CloseReason = "profit_retreat"
				detailData := map[string]interface{}{
					"type":           "止盈回撤",
					"highestProfit":  order.HighestProfit,
					"currentProfit":  currentProfit,
					"retreatPercent": retreatPercent,
					"retreatLine":    strategy.ProfitRetreatPercent,
				}
				detailBytes, _ := json.Marshal(detailData)
				decision.ReasonDetail = string(detailBytes)
				return decision, nil
			}
		}
	}

	return decision, nil
}

// ExecuteClose 执行平仓
func (s *autoCloseImpl) ExecuteClose(ctx context.Context, order *entity.TradingOrder, decision *CloseDecision) error {
	if !decision.ShouldClose {
		return gerror.New("不满足平仓条件")
	}

	// 计算实际盈亏
	var realizedProfit float64
	if order.Direction == "long" {
		realizedProfit = (decision.ClosePrice - order.OpenPrice) * order.Quantity * float64(order.Leverage)
	} else {
		realizedProfit = (order.OpenPrice - decision.ClosePrice) * order.Quantity * float64(order.Leverage)
	}

	// 计算持仓时长
	closeTime := gtime.Now()
	holdDuration := int(closeTime.Sub(order.OpenTime).Seconds())

	// 开始事务
	err := dao.TradingOrder.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 更新订单状态
		_, err := dao.TradingOrder.Ctx(ctx).
			Where(dao.TradingOrder.Columns().Id, order.Id).
			Data(g.Map{
				dao.TradingOrder.Columns().ClosePrice:     decision.ClosePrice,
				dao.TradingOrder.Columns().CloseTime:      closeTime,
				dao.TradingOrder.Columns().RealizedProfit: realizedProfit,
				dao.TradingOrder.Columns().HoldDuration:   holdDuration,
				dao.TradingOrder.Columns().Status:         2, // 已平仓
				dao.TradingOrder.Columns().CloseReason:    decision.CloseReason,
			}).
			Update()

		if err != nil {
			return err
		}

		// 记录平仓日志
		logData := &do.TradingCloseLog{
			TenantId:          order.TenantId,
			UserId:            order.UserId,
			RobotId:           order.RobotId,
			OrderId:           order.Id,
			OrderSn:           order.OrderSn,
			Symbol:            order.Symbol,
			Direction:         order.Direction,
			OpenPrice:         order.OpenPrice,
			ClosePrice:        decision.ClosePrice,
			Quantity:          order.Quantity,
			Leverage:          order.Leverage,
			Margin:            order.Margin,
			RealizedProfit:    realizedProfit,
			HighestProfit:     order.HighestProfit,
			ProfitPercent:     (realizedProfit / order.Margin) * 100,
			CloseReason:       decision.CloseReason,
			CloseDetail:       decision.ReasonDetail,
			OpenFee:           0, // TODO: 计算费用
			HoldFee:           0, // TODO: 计算费用
			CloseFee:          0, // TODO: 计算费用
			TotalFee:          0, // TODO: 计算费用
			CommissionAmount:  0, // TODO: 计算佣金
			CommissionPercent: 0,
			NetProfit:         realizedProfit,
			OpenTime:          order.OpenTime,
			CloseTime:         closeTime,
			HoldDuration:      holdDuration,
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

			// 检查是否达到盈利目标或止损限制
			if robot.MaxProfitTarget > 0 && newTotalProfit >= robot.MaxProfitTarget {
				// 停止机器人
				now := gtime.Now()
				_, _ = dao.TradingRobot.Ctx(ctx).
					Where(dao.TradingRobot.Columns().Id, order.RobotId).
					Data(g.Map{
						dao.TradingRobot.Columns().Status:   4, // 停用
						dao.TradingRobot.Columns().StopTime: now,
					}).
					Update()

				// 结束当前运行区间（auto_stop）
				var sess *entity.TradingRobotRunSession
				_ = dao.TradingRobotRunSession.Ctx(ctx).
					Where(dao.TradingRobotRunSession.Columns().RobotId, robot.Id).
					Where(dao.TradingRobotRunSession.Columns().UserId, robot.UserId).
					WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
					OrderDesc(dao.TradingRobotRunSession.Columns().Id).
					Limit(1).
					Scan(&sess)
				if sess != nil && sess.StartTime != nil {
					runtimeSeconds := int(time.Since(sess.StartTime.Time).Seconds())
					if runtimeSeconds < 0 {
						runtimeSeconds = 0
					}
					_, _ = dao.TradingRobotRunSession.Ctx(ctx).
						Where(dao.TradingRobotRunSession.Columns().Id, sess.Id).
						Data(g.Map{
							dao.TradingRobotRunSession.Columns().EndTime:        now,
							dao.TradingRobotRunSession.Columns().EndReason:      "auto_stop",
							dao.TradingRobotRunSession.Columns().RuntimeSeconds: runtimeSeconds,
							dao.TradingRobotRunSession.Columns().UpdatedAt:      now,
						}).
						Update()
				}

				g.Log().Infof(ctx, "机器人 %d 达到盈利目标，已自动停止", order.RobotId)
			}

			if robot.MaxLossAmount > 0 && newTotalProfit <= -robot.MaxLossAmount {
				// 停止机器人
				now := gtime.Now()
				_, _ = dao.TradingRobot.Ctx(ctx).
					Where(dao.TradingRobot.Columns().Id, order.RobotId).
					Data(g.Map{
						dao.TradingRobot.Columns().Status:   4, // 停用
						dao.TradingRobot.Columns().StopTime: now,
					}).
					Update()

				// 结束当前运行区间（auto_stop）
				var sess *entity.TradingRobotRunSession
				_ = dao.TradingRobotRunSession.Ctx(ctx).
					Where(dao.TradingRobotRunSession.Columns().RobotId, robot.Id).
					Where(dao.TradingRobotRunSession.Columns().UserId, robot.UserId).
					WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
					OrderDesc(dao.TradingRobotRunSession.Columns().Id).
					Limit(1).
					Scan(&sess)
				if sess != nil && sess.StartTime != nil {
					runtimeSeconds := int(time.Since(sess.StartTime.Time).Seconds())
					if runtimeSeconds < 0 {
						runtimeSeconds = 0
					}
					_, _ = dao.TradingRobotRunSession.Ctx(ctx).
						Where(dao.TradingRobotRunSession.Columns().Id, sess.Id).
						Data(g.Map{
							dao.TradingRobotRunSession.Columns().EndTime:        now,
							dao.TradingRobotRunSession.Columns().EndReason:      "auto_stop",
							dao.TradingRobotRunSession.Columns().RuntimeSeconds: runtimeSeconds,
							dao.TradingRobotRunSession.Columns().UpdatedAt:      now,
						}).
						Update()
				}

				g.Log().Infof(ctx, "机器人 %d 达到止损限制，已自动停止", order.RobotId)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	g.Log().Infof(ctx, "自动平仓成功: OrderID=%d, 原因=%s, 盈亏=%.4f", order.Id, decision.CloseReason, realizedProfit)

	return nil
}

// updateHighestProfit 更新最高盈利
func (s *autoCloseImpl) updateHighestProfit(ctx context.Context, orderId int64, highestProfit float64, profitPercent float64) error {
	_, err := dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().Id, orderId).
		Data(g.Map{
			dao.TradingOrder.Columns().HighestProfit:        highestProfit,
			dao.TradingOrder.Columns().ProfitRetreatPercent: profitPercent,
		}).
		Update()

	return err
}

// startProfitRetreat 启动止盈回撤
func (s *autoCloseImpl) startProfitRetreat(ctx context.Context, orderId int64) error {
	_, err := dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().Id, orderId).
		Data(g.Map{
			dao.TradingOrder.Columns().ProfitRetreatStarted: 1,
		}).
		Update()

	if err == nil {
		g.Log().Infof(ctx, "订单 %d 启动止盈回撤", orderId)
	}

	return err
}

// BatchCheckOrders 批量检查订单
func (s *autoCloseImpl) BatchCheckOrders(ctx context.Context, robotId int64) error {
	// 获取机器人配置
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Id, robotId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robot)

	if err != nil {
		return err
	}
	if robot == nil {
		return gerror.New("机器人不存在")
	}

	// 检查机器人状态
	if robot.Status != 2 { // 只处理运行中的机器人
		return nil
	}

	// 构建平仓策略
	strategy := &CloseStrategy{
		StopLossPercent:         robot.StopLossPercent,
		ProfitRetreatPercent:    robot.ProfitRetreatPercent,
		AutoStartRetreatPercent: robot.AutoStartRetreatPercent,
	}

	// 获取所有持仓订单
	var orders []*entity.TradingOrder
	err = dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().RobotId, robotId).
		Where(dao.TradingOrder.Columns().Status, 1). // 持仓中
		Scan(&orders)

	if err != nil {
		return err
	}

	if len(orders) == 0 {
		return nil
	}

	// 获取交易所实例
	exchange, err := ExchangeManager.GetExchange(ctx, robot.ApiConfigId)
	if err != nil {
		return err
	}

	// 批量检查订单
	for _, order := range orders {
		// 获取当前价格
		ticker, err := exchange.GetTicker(ctx, order.Symbol)
		if err != nil {
			g.Log().Errorf(ctx, "获取行情失败: %v", err)
			continue
		}

		// 检查是否需要平仓
		decision, err := s.CheckOrder(ctx, order, ticker.LastPrice, strategy)
		if err != nil {
			g.Log().Errorf(ctx, "检查订单失败: %v", err)
			continue
		}

		// 执行平仓
		if decision.ShouldClose {
			err = s.ExecuteClose(ctx, order, decision)
			if err != nil {
				g.Log().Errorf(ctx, "执行平仓失败: %v", err)
			}
		} else {
			// 更新未实现盈亏
			var unrealizedProfit float64
			if order.Direction == "long" {
				unrealizedProfit = (ticker.LastPrice - order.OpenPrice) * order.Quantity * float64(order.Leverage)
			} else {
				unrealizedProfit = (order.OpenPrice - ticker.LastPrice) * order.Quantity * float64(order.Leverage)
			}

			_, _ = dao.TradingOrder.Ctx(ctx).
				Where(dao.TradingOrder.Columns().Id, order.Id).
				Data(g.Map{
					dao.TradingOrder.Columns().UnrealizedProfit: unrealizedProfit,
				}).
				Update()
		}
	}

	return nil
}
