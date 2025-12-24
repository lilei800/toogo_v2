// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingTradeFill is the golang structure for table trading_trade_fill.
type TradingTradeFill struct {
	Id            int64       `json:"id"            orm:"id"              description:"主键ID"`
	TenantId      int64       `json:"tenantId"      orm:"tenant_id"        description:"租户ID"`
	ApiConfigId   int64       `json:"apiConfigId"   orm:"api_config_id"    description:"API配置ID"`
	Exchange      string      `json:"exchange"      orm:"exchange"         description:"交易所"`
	UserId        int64       `json:"userId"        orm:"user_id"          description:"用户ID"`
	RobotId       int64       `json:"robotId"       orm:"robot_id"         description:"机器人ID"`
	SessionId     *int64      `json:"sessionId"     orm:"session_id"       description:"运行区间ID(可选)"`
	Symbol        string      `json:"symbol"        orm:"symbol"           description:"交易对"`
	OrderId       string      `json:"orderId"       orm:"order_id"         description:"交易所订单ID"`
	ClientOrderId string      `json:"clientOrderId" orm:"client_order_id"  description:"客户端订单ID(可选)"`
	TradeId       string      `json:"tradeId"       orm:"trade_id"         description:"成交ID"`
	Side          string      `json:"side"          orm:"side"             description:"方向"`
	Qty           float64     `json:"qty"           orm:"qty"              description:"成交数量"`
	Price         float64     `json:"price"         orm:"price"            description:"成交价格"`
	Fee           float64     `json:"fee"           orm:"fee"              description:"手续费"`
	FeeCoin       string      `json:"feeCoin"       orm:"fee_coin"         description:"手续费币种"`
	RealizedPnl   float64     `json:"realizedPnl"   orm:"realized_pnl"     description:"已实现盈亏"`
	Ts            int64       `json:"ts"            orm:"ts"               description:"成交时间戳(毫秒)"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"       description:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"       description:"更新时间"`
}
