// Package input
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package input

import (
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/os/gtime"
)

// TradingOrderListInp 订单列表查询输入
type TradingOrderListInp struct {
	Page      int    `json:"page" v:"required|min:1" dc:"页码"`
	PageSize  int    `json:"pageSize" v:"required|min:1|max:100" dc:"每页数量"`
	RobotId   int64  `json:"robotId" dc:"机器人ID筛选"`
	Symbol    string `json:"symbol" dc:"交易对筛选"`
	Direction string `json:"direction" dc:"方向筛选：long/short"`
	Status    int    `json:"status" dc:"状态筛选：1=持仓中,2=已平仓,3=已取消"`
	OrderSn   string `json:"orderSn" dc:"订单号搜索"`
}

// TradingOrderListModel 订单列表输出
type TradingOrderListModel struct {
	Id                   int64       `json:"id" dc:"ID"`
	RobotId              int64       `json:"robotId" dc:"机器人ID"`
	RobotName            string      `json:"robotName" dc:"机器人名称"`
	OrderSn              string      `json:"orderSn" dc:"订单号"`
	ExchangeOrderId      string      `json:"exchangeOrderId" dc:"交易所订单ID"`
	Symbol               string      `json:"symbol" dc:"交易对"`
	Direction            string      `json:"direction" dc:"方向"`
	OpenPrice            float64     `json:"openPrice" dc:"开仓价格"`
	ClosePrice           float64     `json:"closePrice" dc:"平仓价格"`
	Quantity             float64     `json:"quantity" dc:"数量"`
	Leverage             int         `json:"leverage" dc:"杠杆倍数"`
	Margin               float64     `json:"margin" dc:"保证金"`
	RealizedProfit       float64     `json:"realizedProfit" dc:"已实现盈亏"`
	UnrealizedProfit     float64     `json:"unrealizedProfit" dc:"未实现盈亏"`
	HighestProfit        float64     `json:"highestProfit" dc:"最高盈利"`
	ProfitRetreatStarted int         `json:"profitRetreatStarted" dc:"止盈回撤已启动"`
	OpenTime             *gtime.Time `json:"openTime" dc:"开仓时间"`
	CloseTime            *gtime.Time `json:"closeTime" dc:"平仓时间"`
	HoldDuration         int         `json:"holdDuration" dc:"持仓时长(秒)"`
	Status               int         `json:"status" dc:"状态"`
	CloseReason          string      `json:"closeReason" dc:"平仓原因"`
}

// TradingOrderViewInp 查看订单详情输入
type TradingOrderViewInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingOrderViewModel 订单详情输出
type TradingOrderViewModel struct {
	entity.TradingOrder
	RobotName         string  `json:"robotName" dc:"机器人名称"`
	CurrentPrice      float64 `json:"currentPrice" dc:"当前价格"`
	ProfitPercent     float64 `json:"profitPercent" dc:"盈利百分比"`
	StopLossProgress  float64 `json:"stopLossProgress" dc:"止损进度"`
	RetreatProgress   float64 `json:"retreatProgress" dc:"回撤进度"`
	CanManualClose    bool    `json:"canManualClose" dc:"是否可手动平仓"`
}

// TradingOrderPositionsInp 获取持仓订单输入
type TradingOrderPositionsInp struct {
	RobotId int64 `json:"robotId" v:"required" dc:"机器人ID"`
}

// TradingOrderPositionsModel 持仓订单输出
type TradingOrderPositionsModel struct {
	Id                   int64       `json:"id" dc:"ID"`
	OrderSn              string      `json:"orderSn" dc:"订单号"`
	Symbol               string      `json:"symbol" dc:"交易对"`
	Direction            string      `json:"direction" dc:"方向"`
	OpenPrice            float64     `json:"openPrice" dc:"开仓价格"`
	CurrentPrice         float64     `json:"currentPrice" dc:"当前价格"`
	Quantity             float64     `json:"quantity" dc:"数量"`
	Leverage             int         `json:"leverage" dc:"杠杆倍数"`
	Margin               float64     `json:"margin" dc:"保证金"`
	UnrealizedProfit     float64     `json:"unrealizedProfit" dc:"未实现盈亏"`
	HighestProfit        float64     `json:"highestProfit" dc:"最高盈利"`
	ProfitPercent        float64     `json:"profitPercent" dc:"盈利百分比"`
	StopLossPrice        float64     `json:"stopLossPrice" dc:"止损价格"`
	ProfitRetreatStarted int         `json:"profitRetreatStarted" dc:"止盈回撤已启动"`
	ProfitRetreatPercent float64     `json:"profitRetreatPercent" dc:"止盈回撤百分比"`
	OpenTime             *gtime.Time `json:"openTime" dc:"开仓时间"`
	HoldDuration         int         `json:"holdDuration" dc:"持仓时长(秒)"`
	StopLossProgress     float64     `json:"stopLossProgress" dc:"止损进度(%)"`
	RetreatProgress      float64     `json:"retreatProgress" dc:"回撤进度(%)"`
}

// TradingOrderManualCloseInp 手动平仓输入
type TradingOrderManualCloseInp struct {
	Id int64 `json:"id" v:"required" dc:"订单ID"`
}

// TradingOrderStatsInp 订单统计输入
type TradingOrderStatsInp struct {
	RobotId   int64  `json:"robotId" dc:"机器人ID"`
	StartDate string `json:"startDate" dc:"开始日期"`
	EndDate   string `json:"endDate" dc:"结束日期"`
}

// TradingOrderStatsModel 订单统计输出
type TradingOrderStatsModel struct {
	TotalCount      int     `json:"totalCount" dc:"总订单数"`
	LongCount       int     `json:"longCount" dc:"多单数"`
	ShortCount      int     `json:"shortCount" dc:"空单数"`
	PositionCount   int     `json:"positionCount" dc:"持仓中数量"`
	ClosedCount     int     `json:"closedCount" dc:"已平仓数量"`
	ProfitCount     int     `json:"profitCount" dc:"盈利订单数"`
	LossCount       int     `json:"lossCount" dc:"亏损订单数"`
	TotalProfit     float64 `json:"totalProfit" dc:"总盈利"`
	TotalLoss       float64 `json:"totalLoss" dc:"总亏损"`
	NetProfit       float64 `json:"netProfit" dc:"净利润"`
	WinRate         float64 `json:"winRate" dc:"胜率(%)"`
	AvgProfit       float64 `json:"avgProfit" dc:"平均盈利"`
	AvgLoss         float64 `json:"avgLoss" dc:"平均亏损"`
	AvgHoldDuration int     `json:"avgHoldDuration" dc:"平均持仓时长(秒)"`
	ProfitFactor    float64 `json:"profitFactor" dc:"盈亏比"`
}

// TradingOrderCloseLogListInp 平仓日志列表查询输入
type TradingOrderCloseLogListInp struct {
	Page        int    `json:"page" v:"required|min:1" dc:"页码"`
	PageSize    int    `json:"pageSize" v:"required|min:1|max:100" dc:"每页数量"`
	RobotId     int64  `json:"robotId" dc:"机器人ID筛选"`
	Symbol      string `json:"symbol" dc:"交易对筛选"`
	Direction   string `json:"direction" dc:"方向筛选"`
	CloseReason string `json:"closeReason" dc:"平仓原因筛选"`
	StartDate   string `json:"startDate" dc:"开始日期"`
	EndDate     string `json:"endDate" dc:"结束日期"`
}

// TradingOrderCloseLogListModel 平仓日志列表输出
type TradingOrderCloseLogListModel struct {
	Id                int64       `json:"id" dc:"ID"`
	RobotId           int64       `json:"robotId" dc:"机器人ID"`
	RobotName         string      `json:"robotName" dc:"机器人名称"`
	OrderSn           string      `json:"orderSn" dc:"订单号"`
	Symbol            string      `json:"symbol" dc:"交易对"`
	Direction         string      `json:"direction" dc:"方向"`
	OpenPrice         float64     `json:"openPrice" dc:"开仓价格"`
	ClosePrice        float64     `json:"closePrice" dc:"平仓价格"`
	Quantity          float64     `json:"quantity" dc:"数量"`
	Leverage          int         `json:"leverage" dc:"杠杆倍数"`
	Margin            float64     `json:"margin" dc:"保证金"`
	RealizedProfit    float64     `json:"realizedProfit" dc:"已实现盈亏"`
	HighestProfit     float64     `json:"highestProfit" dc:"最高盈利"`
	ProfitPercent     float64     `json:"profitPercent" dc:"盈利百分比"`
	CloseReason       string      `json:"closeReason" dc:"平仓原因"`
	TotalFee          float64     `json:"totalFee" dc:"总费用"`
	CommissionAmount  float64     `json:"commissionAmount" dc:"佣金金额"`
	NetProfit         float64     `json:"netProfit" dc:"净利润"`
	OpenTime          *gtime.Time `json:"openTime" dc:"开仓时间"`
	CloseTime         *gtime.Time `json:"closeTime" dc:"平仓时间"`
	HoldDuration      int         `json:"holdDuration" dc:"持仓时长(秒)"`
}

// TradingOrderCloseLogViewInp 查看平仓日志详情输入
type TradingOrderCloseLogViewInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingOrderCloseLogViewModel 平仓日志详情输出
type TradingOrderCloseLogViewModel struct {
	entity.TradingCloseLog
	RobotName string `json:"robotName" dc:"机器人名称"`
}

