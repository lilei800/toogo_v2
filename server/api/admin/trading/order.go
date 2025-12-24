// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/model/input"
)

// OrderListReq 订单列表请求
type OrderListReq struct {
	g.Meta `path:"/trading/order/list" method:"get" tags:"交易管理" summary:"订单列表" dc:"获取交易订单列表"`
	input.TradingOrderListInp
}

// OrderListRes 订单列表响应
type OrderListRes struct {
	List       []*input.TradingOrderListModel `json:"list" dc:"列表数据"`
	TotalCount int                            `json:"totalCount" dc:"总数"`
	Page       int                            `json:"page" dc:"当前页"`
	PageSize   int                            `json:"pageSize" dc:"每页数量"`
}

// OrderViewReq 查看订单请求
type OrderViewReq struct {
	g.Meta `path:"/trading/order/view" method:"get" tags:"交易管理" summary:"查看订单详情" dc:"查看订单详细信息"`
	input.TradingOrderViewInp
}

// OrderViewRes 查看订单响应
type OrderViewRes struct {
	*input.TradingOrderViewModel
}

// OrderPositionsReq 获取持仓订单请求
type OrderPositionsReq struct {
	g.Meta `path:"/trading/order/positions" method:"get" tags:"交易管理" summary:"获取持仓订单" dc:"获取机器人的持仓订单"`
	input.TradingOrderPositionsInp
}

// OrderPositionsRes 获取持仓订单响应
type OrderPositionsRes struct {
	List []*input.TradingOrderPositionsModel `json:"list" dc:"持仓订单列表"`
}

// OrderManualCloseReq 手动平仓请求
type OrderManualCloseReq struct {
	g.Meta `path:"/trading/order/manualClose" method:"post" tags:"交易管理" summary:"手动平仓" dc:"手动平仓指定订单"`
	input.TradingOrderManualCloseInp
}

// OrderManualCloseRes 手动平仓响应
type OrderManualCloseRes struct{}

// OrderStatsReq 订单统计请求
type OrderStatsReq struct {
	g.Meta `path:"/trading/order/stats" method:"get" tags:"交易管理" summary:"订单统计" dc:"获取订单统计数据"`
	input.TradingOrderStatsInp
}

// OrderStatsRes 订单统计响应
type OrderStatsRes struct {
	*input.TradingOrderStatsModel
}

// OrderCloseLogsReq 平仓日志列表请求
type OrderCloseLogsReq struct {
	g.Meta `path:"/trading/order/closeLogs" method:"get" tags:"交易管理" summary:"平仓日志列表" dc:"获取平仓日志列表"`
	input.TradingOrderCloseLogListInp
}

// OrderCloseLogsRes 平仓日志列表响应
type OrderCloseLogsRes struct {
	List       []*input.TradingOrderCloseLogListModel `json:"list" dc:"列表数据"`
	TotalCount int                                    `json:"totalCount" dc:"总数"`
	Page       int                                    `json:"page" dc:"当前页"`
	PageSize   int                                    `json:"pageSize" dc:"每页数量"`
}

// OrderCloseLogViewReq 查看平仓日志请求
type OrderCloseLogViewReq struct {
	g.Meta `path:"/trading/order/closeLogView" method:"get" tags:"交易管理" summary:"查看平仓日志详情" dc:"查看平仓日志详细信息"`
	input.TradingOrderCloseLogViewInp
}

// OrderCloseLogViewRes 查看平仓日志响应
type OrderCloseLogViewRes struct {
	*input.TradingOrderCloseLogViewModel
}

