// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UsdtBalanceLog is the golang structure for table usdt_balance_log.
type UsdtBalanceLog struct {
	Id            int64       `json:"id"            orm:"id"             description:"ID"`
	UserId        int64       `json:"userId"        orm:"user_id"        description:"用户ID"`
	OrderSn       string      `json:"orderSn"       orm:"order_sn"       description:"关联订单号"`
	Type          int         `json:"type"          orm:"type"           description:"变动类型:1充值,2提现,3交易"`
	ChangeAmount  float64     `json:"changeAmount"  orm:"change_amount"  description:"变动金额"`
	BeforeBalance float64     `json:"beforeBalance" orm:"before_balance" description:"变动前余额"`
	AfterBalance  float64     `json:"afterBalance"  orm:"after_balance"  description:"变动后余额"`
	Remark        string      `json:"remark"        orm:"remark"         description:"备注"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:"更新时间"`
}

