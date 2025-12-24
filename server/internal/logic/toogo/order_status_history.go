// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Description 订单状态历史记录 - 记录订单在不同节点的状态变更

package toogo

import (
	"context"
	"fmt"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// OrderStatusNodeType 订单状态节点类型
const (
	NodeTypePreCreate     = "pre_create"      // 预创建
	NodeTypeExchangeSubmit = "exchange_submit" // 交易所下单
	NodeTypeExchangeSuccess = "exchange_success" // 交易所成功
	NodeTypeSyncDetail    = "sync_detail"     // 同步详情
	NodeTypeSyncPnl       = "sync_pnl"        // 同步盈亏
	NodeTypeClose         = "close"           // 平仓
	NodeTypeFailed        = "failed"          // 下单失败
)

// RecordOrderStatusHistory 记录订单状态历史
// 【职责分离】每个节点创建独立的记录，不更新现有记录
func RecordOrderStatusHistory(ctx context.Context, orderId int64, orderSn string, nodeType string, status int, data map[string]interface{}) error {
	historyData := g.Map{
		"order_id":        orderId,
		"order_sn":        orderSn,
		"status":          status,
		"status_text":     GetOrderStatusText(status),
		"node_type":       nodeType,
		"node_time":       gtime.Now(),
		"created_at":      gtime.Now(),
	}

	// 根据节点类型设置描述
	switch nodeType {
	case NodeTypePreCreate:
		historyData["node_description"] = "预创建订单记录"
	case NodeTypeExchangeSubmit:
		historyData["node_description"] = "提交到交易所"
	case NodeTypeExchangeSuccess:
		historyData["node_description"] = "交易所下单成功"
	case NodeTypeSyncDetail:
		historyData["node_description"] = "同步订单详情"
	case NodeTypeSyncPnl:
		historyData["node_description"] = "同步未实现盈亏"
	case NodeTypeClose:
		historyData["node_description"] = "订单平仓"
	case NodeTypeFailed:
		historyData["node_description"] = "下单失败"
	default:
		historyData["node_description"] = fmt.Sprintf("状态变更: %s", nodeType)
	}

	// 合并传入的数据
	if data != nil {
		for k, v := range data {
			historyData[k] = v
		}
	}

	_, err := dao.TradingOrderStatusHistory.Ctx(ctx).Insert(historyData)
	if err != nil {
		g.Log().Errorf(ctx, "[OrderStatusHistory] 记录订单状态历史失败: orderId=%d, nodeType=%s, err=%v", orderId, nodeType, err)
		return err
	}

	g.Log().Debugf(ctx, "[OrderStatusHistory] 已记录订单状态历史: orderId=%d, orderSn=%s, nodeType=%s, status=%d", orderId, orderSn, nodeType, status)
	return nil
}

// RecordPreCreateNode 记录预创建节点
func RecordPreCreateNode(ctx context.Context, orderId int64, orderSn string, orderData map[string]interface{}) error {
	return RecordOrderStatusHistory(ctx, orderId, orderSn, NodeTypePreCreate, OrderStatusPending, orderData)
}

// RecordExchangeSubmitNode 记录交易所下单节点
func RecordExchangeSubmitNode(ctx context.Context, orderId int64, orderSn string, exchangeOrderId string) error {
	return RecordOrderStatusHistory(ctx, orderId, orderSn, NodeTypeExchangeSubmit, OrderStatusPending, g.Map{
		"exchange_order_id": exchangeOrderId,
	})
}

// RecordExchangeSuccessNode 记录交易所成功节点
func RecordExchangeSuccessNode(ctx context.Context, orderId int64, orderSn string, exchangeOrderId string, order *exchange.Order) error {
	data := g.Map{
		"exchange_order_id": exchangeOrderId,
		"status":           OrderStatusOpen,
	}

	if order != nil {
		if order.ClientId != "" {
			data["client_order_id"] = order.ClientId
		}
		if order.CreateTime > 0 {
			data["node_time"] = gtime.NewFromTimeStamp(order.CreateTime / 1000)
		}
	}

	return RecordOrderStatusHistory(ctx, orderId, orderSn, NodeTypeExchangeSuccess, OrderStatusOpen, data)
}

// RecordSyncDetailNode 记录同步详情节点
func RecordSyncDetailNode(ctx context.Context, orderId int64, orderSn string, historyOrder *exchange.Order) error {
	data := g.Map{
		"status": OrderStatusOpen,
	}

	if historyOrder != nil {
		if historyOrder.OrderId != "" {
			data["exchange_order_id"] = historyOrder.OrderId
		}
		if historyOrder.AvgPrice > 0 {
			data["avg_price"] = historyOrder.AvgPrice
			data["open_price"] = historyOrder.AvgPrice
		}
		if historyOrder.FilledQty > 0 {
			data["filled_qty"] = historyOrder.FilledQty
		}
		if historyOrder.Fee > 0 {
			data["fee"] = historyOrder.Fee
			data["fee_coin"] = historyOrder.FeeCoin
		}
		if historyOrder.CreateTime > 0 {
			data["node_time"] = gtime.NewFromTimeStamp(historyOrder.CreateTime / 1000)
		}
	}

	return RecordOrderStatusHistory(ctx, orderId, orderSn, NodeTypeSyncDetail, OrderStatusOpen, data)
}

// RecordSyncPnlNode 记录同步盈亏节点
func RecordSyncPnlNode(ctx context.Context, orderId int64, orderSn string, unrealizedPnl, highestProfit, markPrice float64) error {
	return RecordOrderStatusHistory(ctx, orderId, orderSn, NodeTypeSyncPnl, OrderStatusOpen, g.Map{
		"unrealized_profit": unrealizedPnl,
		"highest_profit":    highestProfit,
		"mark_price":        markPrice,
	})
}

// RecordCloseNode 记录平仓节点
func RecordCloseNode(ctx context.Context, orderId int64, orderSn string, closePrice, realizedProfit float64, closeReason string) error {
	return RecordOrderStatusHistory(ctx, orderId, orderSn, NodeTypeClose, OrderStatusClosed, g.Map{
		"close_price":     closePrice,
		"realized_profit": realizedProfit,
		"close_reason":    closeReason,
	})
}

// RecordFailedNode 记录下单失败节点
func RecordFailedNode(ctx context.Context, orderId int64, orderSn string, errorMsg string) error {
	return RecordOrderStatusHistory(ctx, orderId, orderSn, NodeTypeFailed, OrderStatusFailed, g.Map{
		"remark": errorMsg,
	})
}

