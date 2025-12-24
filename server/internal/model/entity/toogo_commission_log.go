// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoCommissionLog is the golang structure for table hg_toogo_commission_log.
type ToogoCommissionLog struct {
	Id               int64       `json:"id"               orm:"id"                description:"主键ID"`
	UserId           int64       `json:"userId"           orm:"user_id"           description:"获得佣金用户ID"`
	FromUserId       int64       `json:"fromUserId"       orm:"from_user_id"      description:"来源用户ID"`
	CommissionType   string      `json:"commissionType"   orm:"commission_type"   description:"佣金类型: invite_reward/subscribe/power_consume"`
	Level            int         `json:"level"            orm:"level"             description:"层级: 1=一级, 2=二级, 3=三级"`
	BaseAmount       float64     `json:"baseAmount"       orm:"base_amount"       description:"基础金额(订阅额/算力消耗)"`
	CommissionRate   float64     `json:"commissionRate"   orm:"commission_rate"   description:"佣金比例"`
	CommissionAmount float64     `json:"commissionAmount" orm:"commission_amount" description:"佣金金额"`
	SettleType       string      `json:"settleType"       orm:"settle_type"       description:"结算类型: power/usdt"`
	Status           int         `json:"status"           orm:"status"            description:"状态: 1=待结算, 2=已结算"`
	RelatedId        int64       `json:"relatedId"        orm:"related_id"        description:"关联ID"`
	RelatedType      string      `json:"relatedType"      orm:"related_type"      description:"关联类型"`
	OrderSn          string      `json:"orderSn"          orm:"order_sn"          description:"关联订单号"`
	Remark           string      `json:"remark"           orm:"remark"            description:"备注"`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"        description:"创建时间"`
}

