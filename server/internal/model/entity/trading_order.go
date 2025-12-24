// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingOrder is the golang structure for table trading_order.
type TradingOrder struct {
	Id                   int64       `json:"id"                   orm:"id"                       description:"主键ID"`
	TenantId             int64       `json:"tenantId"             orm:"tenant_id"                description:"租户ID"`
	UserId               int64       `json:"userId"               orm:"user_id"                  description:"用户ID"`
	RobotId              int64       `json:"robotId"              orm:"robot_id"                 description:"机器人ID"`
	OrderSn              string      `json:"orderSn"              orm:"order_sn"                 description:"订单号"`
	ExchangeOrderId      string      `json:"exchangeOrderId"      orm:"exchange_order_id"        description:"交易所订单ID"`
	Symbol               string      `json:"symbol"               orm:"symbol"                   description:"交易对"`
	Direction            string      `json:"direction"            orm:"direction"                description:"方向：long/short"`
	OpenPrice            float64     `json:"openPrice"            orm:"open_price"               description:"开仓价格"`
	ClosePrice           float64     `json:"closePrice"           orm:"close_price"              description:"平仓价格"`
	Quantity             float64     `json:"quantity"             orm:"quantity"                 description:"数量"`
	Leverage             int         `json:"leverage"             orm:"leverage"                 description:"杠杆倍数"`
	Margin               float64     `json:"margin"               orm:"margin"                   description:"保证金(USDT)"`
	RealizedProfit       float64     `json:"realizedProfit"       orm:"realized_profit"          description:"已实现盈亏"`
	UnrealizedProfit     float64     `json:"unrealizedProfit"     orm:"unrealized_profit"        description:"未实现盈亏"`
	HighestProfit        float64     `json:"highestProfit"        orm:"highest_profit"           description:"最高盈利"`
	StopLossPrice        float64     `json:"stopLossPrice"        orm:"stop_loss_price"          description:"止损价格"`
	ProfitRetreatStarted int         `json:"profitRetreatStarted" orm:"profit_retreat_started"   description:"止盈回撤已启动"`
	ProfitRetreatPercent float64     `json:"profitRetreatPercent" orm:"profit_retreat_percent"   description:"止盈回撤百分比"`
	OpenTime             *gtime.Time `json:"openTime"             orm:"open_time"                description:"开仓时间"`
	CloseTime            *gtime.Time `json:"closeTime"            orm:"close_time"               description:"平仓时间"`
	HoldDuration         int         `json:"holdDuration"         orm:"hold_duration"            description:"持仓时长(秒)"`
	Status               int         `json:"status"               orm:"status"                   description:"状态：1=持仓中,2=已平仓,3=已取消"`
	CloseReason          string      `json:"closeReason"          orm:"close_reason"             description:"平仓原因：stop_loss/take_profit/manual/timeout"`
	Remark               string      `json:"remark"               orm:"remark"                   description:"备注"`
	CreatedAt            *gtime.Time `json:"createdAt"            orm:"created_at"               description:"创建时间"`
	UpdatedAt            *gtime.Time `json:"updatedAt"            orm:"updated_at"               description:"更新时间"`
}

