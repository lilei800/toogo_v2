// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoTransfer is the golang structure for table hg_toogo_transfer.
type ToogoTransfer struct {
	Id          int64       `json:"id"          orm:"id"           description:"主键ID"`
	UserId      int64       `json:"userId"      orm:"user_id"      description:"用户ID(member_id)"`
	OrderSn     string      `json:"orderSn"     orm:"order_sn"     description:"订单号"`
	FromAccount string      `json:"fromAccount" orm:"from_account" description:"转出账户: balance/commission"`
	ToAccount   string      `json:"toAccount"   orm:"to_account"   description:"转入账户: power"`
	Amount      float64     `json:"amount"      orm:"amount"       description:"转账金额(USDT)"`
	PowerAmount float64     `json:"powerAmount" orm:"power_amount" description:"获得算力"`
	Rate        float64     `json:"rate"        orm:"rate"         description:"兑换比率"`
	Status      int         `json:"status"      orm:"status"       description:"状态: 1=处理中, 2=已完成, 3=失败"`
	Remark      string      `json:"remark"      orm:"remark"       description:"备注"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:"创建时间"`
}

