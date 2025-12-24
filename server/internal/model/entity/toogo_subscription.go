// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoSubscription is the golang structure for table hg_toogo_subscription.
type ToogoSubscription struct {
	Id                 int64       `json:"id"                 orm:"id"                   description:"主键ID"`
	UserId             int64       `json:"userId"             orm:"user_id"              description:"用户ID(member_id)"`
	PlanId             int64       `json:"planId"             orm:"plan_id"              description:"套餐ID"`
	PlanCode           string      `json:"planCode"           orm:"plan_code"            description:"套餐代码"`
	OrderSn            string      `json:"orderSn"            orm:"order_sn"             description:"订单号"`
	PeriodType         string      `json:"periodType"         orm:"period_type"          description:"订阅周期: daily/monthly/quarterly/half_year/yearly"`
	Amount             float64     `json:"amount"             orm:"amount"               description:"订阅金额(USDT)"`
	GiftPower          float64     `json:"giftPower"          orm:"gift_power"           description:"赠送积分"`
	StartTime          *gtime.Time `json:"startTime"          orm:"start_time"           description:"开始时间"`
	ExpireTime         *gtime.Time `json:"expireTime"         orm:"expire_time"          description:"到期时间"`
	Days               int         `json:"days"               orm:"days"                 description:"订阅天数"`
	Status             int         `json:"status"             orm:"status"               description:"状态: 1=待支付, 2=生效中, 3=已过期, 4=已取消"`
	PaidAt             *gtime.Time `json:"paidAt"             orm:"paid_at"              description:"支付时间"`
	PayType            string      `json:"payType"            orm:"pay_type"             description:"支付方式: balance/crypto"`
	InviterId          int64       `json:"inviterId"          orm:"inviter_id"           description:"邀请人ID"`
	CommissionSettled  int         `json:"commissionSettled"  orm:"commission_settled"   description:"佣金是否已结算"`
	CreatedAt          *gtime.Time `json:"createdAt"          orm:"created_at"           description:"创建时间"`
	UpdatedAt          *gtime.Time `json:"updatedAt"          orm:"updated_at"           description:"更新时间"`
}

