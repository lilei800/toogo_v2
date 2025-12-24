// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoWalletLog is the golang structure for table hg_toogo_wallet_log.
type ToogoWalletLog struct {
	Id           int64       `json:"id"           orm:"id"            description:"主键ID"`
	UserId       int64       `json:"userId"       orm:"user_id"       description:"用户ID(member_id)"`
	AccountType  string      `json:"accountType"  orm:"account_type"  description:"账户类型: balance/power/gift_power/commission"`
	ChangeType   string      `json:"changeType"   orm:"change_type"   description:"变动类型"`
	ChangeAmount float64     `json:"changeAmount" orm:"change_amount" description:"变动金额"`
	BeforeAmount float64     `json:"beforeAmount" orm:"before_amount" description:"变动前余额"`
	AfterAmount  float64     `json:"afterAmount"  orm:"after_amount"  description:"变动后余额"`
	RelatedId    int64       `json:"relatedId"    orm:"related_id"    description:"关联ID"`
	RelatedType  string      `json:"relatedType"  orm:"related_type"  description:"关联类型"`
	OrderSn      string      `json:"orderSn"      orm:"order_sn"      description:"关联订单号"`
	Remark       string      `json:"remark"       orm:"remark"        description:"备注"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`
}

