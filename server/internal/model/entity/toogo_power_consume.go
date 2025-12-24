// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoPowerConsume is the golang structure for table hg_toogo_power_consume.
type ToogoPowerConsume struct {
	Id            int64       `json:"id"            orm:"id"             description:"主键ID"`
	UserId        int64       `json:"userId"        orm:"user_id"        description:"用户ID(member_id)"`
	RobotId       int64       `json:"robotId"       orm:"robot_id"       description:"机器人ID"`
	OrderId       int64       `json:"orderId"       orm:"order_id"       description:"交易订单ID"`
	OrderSn       string      `json:"orderSn"       orm:"order_sn"       description:"订单号"`
	ProfitAmount  float64     `json:"profitAmount"  orm:"profit_amount"  description:"盈利金额(USDT)"`
	ConsumeRate   float64     `json:"consumeRate"   orm:"consume_rate"   description:"消耗比例"`
	ConsumePower  float64     `json:"consumePower"  orm:"consume_power"  description:"消耗算力"`
	FromPower     float64     `json:"fromPower"     orm:"from_power"     description:"从算力账户扣除"`
	VipLevel      int         `json:"vipLevel"      orm:"vip_level"      description:"用户VIP等级"`
	DiscountRate  float64     `json:"discountRate"  orm:"discount_rate"  description:"折扣比例(%)"`
	OriginalPower float64     `json:"originalPower" orm:"original_power" description:"原始消耗算力(未折扣)"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:"创建时间"`
}
