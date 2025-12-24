// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Description 订单事件记录 - 追踪订单生命周期中的每个节点事件

package toogo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ==================== 订单事件类型 ====================

const (
	// OrderEventSignalGenerated 信号生成
	OrderEventSignalGenerated = "signal_generated"
	// OrderEventCheckStarted 开仓检查开始
	OrderEventCheckStarted = "check_started"
	// OrderEventPreCreated 预创建订单记录
	OrderEventPreCreated = "pre_created"
	// OrderEventExchangeOrdered 交易所下单
	OrderEventExchangeOrdered = "exchange_ordered"
	// OrderEventOrderFilled 订单成交
	OrderEventOrderFilled = "order_filled"
	// OrderEventPositionUpdated 持仓更新
	OrderEventPositionUpdated = "position_updated"
	// OrderEventOrderClosed 订单平仓
	OrderEventOrderClosed = "order_closed"
	// OrderEventOrderFailed 订单失败
	OrderEventOrderFailed = "order_failed"
)

// OrderEventStatus 订单事件状态
const (
	OrderEventStatusSuccess = "success"
	OrderEventStatusFailed  = "failed"
	OrderEventStatusPending = "pending"
)

// ==================== 订单事件记录 ====================

// RecordOrderEvent 记录订单事件
// orderId: 订单ID（本地订单ID）
// eventType: 事件类型
// eventStatus: 事件状态（success/failed/pending）
// eventData: 事件数据（任意结构，会被序列化为JSON）
// eventMessage: 事件消息（人类可读的描述）
func RecordOrderEvent(ctx context.Context, orderId int64, exchangeOrderId string, eventType, eventStatus string, eventData interface{}, eventMessage string) error {
	// 序列化事件数据
	var eventDataJSON []byte
	var err error
	if eventData != nil {
		eventDataJSON, err = json.Marshal(eventData)
		if err != nil {
			g.Log().Warningf(ctx, "[OrderEvent] 序列化事件数据失败: orderId=%d, eventType=%s, err=%v", orderId, eventType, err)
			eventDataJSON = []byte("{}")
		}
	} else {
		eventDataJSON = []byte("{}")
	}

	// 构建插入数据
	insertData := g.Map{
		"order_id":         orderId,
		"exchange_order_id": exchangeOrderId,
		"event_type":       eventType,
		"event_status":     eventStatus,
		"event_data":       string(eventDataJSON),
		"event_message":    eventMessage,
		"created_at":       gtime.Now(),
	}

	// 插入事件记录
	_, err = g.DB().Model("hg_trading_order_event").Ctx(ctx).Insert(insertData)
	if err != nil {
		g.Log().Errorf(ctx, "[OrderEvent] 记录订单事件失败: orderId=%d, eventType=%s, err=%v", orderId, eventType, err)
		return err
	}

	g.Log().Debugf(ctx, "[OrderEvent] 订单事件已记录: orderId=%d, eventType=%s, eventStatus=%s", orderId, eventType, eventStatus)
	return nil
}

// RecordSignalGenerated 记录信号生成事件
func RecordSignalGenerated(ctx context.Context, robotId int64, signal *RobotSignal) error {
	eventData := map[string]interface{}{
		"robot_id":   robotId,
		"direction":  signal.Direction,
		"action":     signal.Action,
		"strength":   signal.Strength,
		"confidence": signal.Confidence,
		"timestamp":  signal.Timestamp,
	}
	return RecordOrderEvent(ctx, 0, "", OrderEventSignalGenerated, OrderEventStatusSuccess, eventData, fmt.Sprintf("信号生成: direction=%s, strength=%.2f, confidence=%.2f", signal.Direction, signal.Strength, signal.Confidence))
}

// RecordCheckStarted 记录开仓检查开始事件
func RecordCheckStarted(ctx context.Context, robotId int64, direction string, checkResult map[string]interface{}) error {
	eventData := map[string]interface{}{
		"robot_id":     robotId,
		"direction":    direction,
		"check_result": checkResult,
	}
	message := fmt.Sprintf("开仓检查开始: direction=%s", direction)
	if result, ok := checkResult["has_position"]; ok {
		message += fmt.Sprintf(", has_position=%v", result)
	}
	return RecordOrderEvent(ctx, 0, "", OrderEventCheckStarted, OrderEventStatusSuccess, eventData, message)
}

// RecordPreCreated 记录预创建订单事件
func RecordPreCreated(ctx context.Context, orderId int64, orderData map[string]interface{}) error {
	eventData := map[string]interface{}{
		"order_id":  orderId,
		"order_data": orderData,
	}
	return RecordOrderEvent(ctx, orderId, "", OrderEventPreCreated, OrderEventStatusSuccess, eventData, fmt.Sprintf("预创建订单记录: orderId=%d, status=PENDING", orderId))
}

// RecordExchangeOrdered 记录交易所下单事件
func RecordExchangeOrdered(ctx context.Context, orderId int64, exchangeOrderId string, requestData, responseData map[string]interface{}, success bool) error {
	eventData := map[string]interface{}{
		"order_id":      orderId,
		"request_data":  requestData,
		"response_data": responseData,
	}
	status := OrderEventStatusSuccess
	message := fmt.Sprintf("交易所下单成功: orderId=%d, exchangeOrderId=%s", orderId, exchangeOrderId)
	if !success {
		status = OrderEventStatusFailed
		message = fmt.Sprintf("交易所下单失败: orderId=%d", orderId)
	}
	return RecordOrderEvent(ctx, orderId, exchangeOrderId, OrderEventExchangeOrdered, status, eventData, message)
}

// RecordOrderFilled 记录订单成交事件
func RecordOrderFilled(ctx context.Context, orderId int64, exchangeOrderId string, fillData map[string]interface{}) error {
	eventData := map[string]interface{}{
		"order_id":  orderId,
		"fill_data": fillData,
	}
	return RecordOrderEvent(ctx, orderId, exchangeOrderId, OrderEventOrderFilled, OrderEventStatusSuccess, eventData, fmt.Sprintf("订单成交: orderId=%d, exchangeOrderId=%s", orderId, exchangeOrderId))
}

// RecordPositionUpdated 记录持仓更新事件
func RecordPositionUpdated(ctx context.Context, orderId int64, exchangeOrderId string, updateData map[string]interface{}) error {
	eventData := map[string]interface{}{
		"order_id":    orderId,
		"update_data": updateData,
	}
	return RecordOrderEvent(ctx, orderId, exchangeOrderId, OrderEventPositionUpdated, OrderEventStatusSuccess, eventData, fmt.Sprintf("持仓更新: orderId=%d, unrealizedPnl=%.4f", orderId, updateData["unrealized_profit"]))
}

// RecordOrderClosed 记录订单平仓事件
func RecordOrderClosed(ctx context.Context, orderId int64, exchangeOrderId string, closeData map[string]interface{}) error {
	eventData := map[string]interface{}{
		"order_id":   orderId,
		"close_data": closeData,
	}
	return RecordOrderEvent(ctx, orderId, exchangeOrderId, OrderEventOrderClosed, OrderEventStatusSuccess, eventData, fmt.Sprintf("订单平仓: orderId=%d, realizedProfit=%.4f", orderId, closeData["realized_profit"]))
}

// RecordOrderFailed 记录订单失败事件
func RecordOrderFailed(ctx context.Context, orderId int64, exchangeOrderId string, errorMsg string, errorData map[string]interface{}) error {
	eventData := map[string]interface{}{
		"order_id":   orderId,
		"error_data": errorData,
		"error_msg":  errorMsg,
	}
	return RecordOrderEvent(ctx, orderId, exchangeOrderId, OrderEventOrderFailed, OrderEventStatusFailed, eventData, fmt.Sprintf("订单失败: orderId=%d, error=%s", orderId, errorMsg))
}

