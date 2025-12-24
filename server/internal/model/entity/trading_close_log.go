// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingCloseLog is the golang structure for table trading_close_log.
type TradingCloseLog struct {
	Id                int64       `json:"id"                orm:"id"                  description:"主键ID"`
	TenantId          int64       `json:"tenantId"          orm:"tenant_id"           description:"租户ID"`
	UserId            int64       `json:"userId"            orm:"user_id"             description:"用户ID"`
	RobotId           int64       `json:"robotId"           orm:"robot_id"            description:"机器人ID"`
	OrderId           int64       `json:"orderId"           orm:"order_id"            description:"订单ID"`
	OrderSn           string      `json:"orderSn"           orm:"order_sn"            description:"订单号"`
	Symbol            string      `json:"symbol"            orm:"symbol"              description:"交易对"`
	Direction         string      `json:"direction"         orm:"direction"           description:"方向：long/short"`
	OpenPrice         float64     `json:"openPrice"         orm:"open_price"          description:"开仓价格"`
	ClosePrice        float64     `json:"closePrice"        orm:"close_price"         description:"平仓价格"`
	Quantity          float64     `json:"quantity"          orm:"quantity"            description:"数量"`
	Leverage          int         `json:"leverage"          orm:"leverage"            description:"杠杆倍数"`
	Margin            float64     `json:"margin"            orm:"margin"              description:"保证金(USDT)"`
	RealizedProfit    float64     `json:"realizedProfit"    orm:"realized_profit"     description:"已实现盈亏"`
	HighestProfit     float64     `json:"highestProfit"     orm:"highest_profit"      description:"最高盈利"`
	ProfitPercent     float64     `json:"profitPercent"     orm:"profit_percent"      description:"盈利百分比"`
	CloseReason       string      `json:"closeReason"       orm:"close_reason"        description:"平仓原因"`
	CloseDetail       string      `json:"closeDetail"       orm:"close_detail"        description:"平仓详情(JSON)"`
	OpenFee           float64     `json:"openFee"           orm:"open_fee"            description:"开仓费用"`
	HoldFee           float64     `json:"holdFee"           orm:"hold_fee"            description:"持仓费用"`
	CloseFee          float64     `json:"closeFee"          orm:"close_fee"           description:"平仓费用"`
	TotalFee          float64     `json:"totalFee"          orm:"total_fee"           description:"总费用"`
	CommissionAmount  float64     `json:"commissionAmount"  orm:"commission_amount"   description:"佣金金额"`
	CommissionPercent float64     `json:"commissionPercent" orm:"commission_percent"  description:"佣金比例"`
	NetProfit         float64     `json:"netProfit"         orm:"net_profit"          description:"净利润"`
	OpenTime          *gtime.Time `json:"openTime"          orm:"open_time"           description:"开仓时间"`
	CloseTime         *gtime.Time `json:"closeTime"         orm:"close_time"          description:"平仓时间"`
	HoldDuration      int         `json:"holdDuration"      orm:"hold_duration"       description:"持仓时长(秒)"`
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"          description:"创建时间"`
}

