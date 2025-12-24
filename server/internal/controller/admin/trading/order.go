// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"hotgo/api/admin/trading"
	tradingLogic "hotgo/internal/logic/trading"
)

var Order = cOrder{}

type cOrder struct{}

// List 获取订单列表
func (c *cOrder) List(ctx context.Context, req *trading.OrderListReq) (res *trading.OrderListRes, err error) {
	list, totalCount, err := tradingLogic.Order.List(ctx, &req.TradingOrderListInp)
	if err != nil {
		return nil, err
	}

	res = &trading.OrderListRes{
		List:       list,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	return
}

// View 查看订单详情
func (c *cOrder) View(ctx context.Context, req *trading.OrderViewReq) (res *trading.OrderViewRes, err error) {
	out, err := tradingLogic.Order.View(ctx, &req.TradingOrderViewInp)
	if err != nil {
		return nil, err
	}

	res = &trading.OrderViewRes{
		TradingOrderViewModel: out,
	}
	return
}

// Positions 获取持仓订单
func (c *cOrder) Positions(ctx context.Context, req *trading.OrderPositionsReq) (res *trading.OrderPositionsRes, err error) {
	list, err := tradingLogic.Order.GetPositions(ctx, &req.TradingOrderPositionsInp)
	if err != nil {
		return nil, err
	}

	res = &trading.OrderPositionsRes{
		List: list,
	}
	return
}

// ManualClose 手动平仓
func (c *cOrder) ManualClose(ctx context.Context, req *trading.OrderManualCloseReq) (res *trading.OrderManualCloseRes, err error) {
	err = tradingLogic.Order.ManualClose(ctx, &req.TradingOrderManualCloseInp)
	if err != nil {
		return nil, err
	}

	res = &trading.OrderManualCloseRes{}
	return
}

// Stats 获取订单统计
func (c *cOrder) Stats(ctx context.Context, req *trading.OrderStatsReq) (res *trading.OrderStatsRes, err error) {
	out, err := tradingLogic.Order.GetStats(ctx, &req.TradingOrderStatsInp)
	if err != nil {
		return nil, err
	}

	res = &trading.OrderStatsRes{
		TradingOrderStatsModel: out,
	}
	return
}

// CloseLogs 获取平仓日志列表
func (c *cOrder) CloseLogs(ctx context.Context, req *trading.OrderCloseLogsReq) (res *trading.OrderCloseLogsRes, err error) {
	list, totalCount, err := tradingLogic.Order.GetCloseLogs(ctx, &req.TradingOrderCloseLogListInp)
	if err != nil {
		return nil, err
	}

	res = &trading.OrderCloseLogsRes{
		List:       list,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	return
}

// CloseLogView 查看平仓日志详情
func (c *cOrder) CloseLogView(ctx context.Context, req *trading.OrderCloseLogViewReq) (res *trading.OrderCloseLogViewRes, err error) {
	out, err := tradingLogic.Order.ViewCloseLog(ctx, &req.TradingOrderCloseLogViewInp)
	if err != nil {
		return nil, err
	}

	res = &trading.OrderCloseLogViewRes{
		TradingOrderCloseLogViewModel: out,
	}
	return
}

