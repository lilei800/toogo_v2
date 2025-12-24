// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UsdtDeposit is the golang structure for table usdt_deposit.
type UsdtDeposit struct {
	Id            int64       `json:"id"            orm:"id"             description:"ID"`
	OrderSn       string      `json:"orderSn"       orm:"order_sn"       description:"订单号"`
	UserId        int64       `json:"userId"        orm:"user_id"        description:"用户ID"`
	Amount        float64     `json:"amount"        orm:"amount"         description:"充值金额"`
	ToAddress     string      `json:"toAddress"     orm:"to_address"     description:"充值地址"`
	TxHash        string      `json:"txHash"        orm:"tx_hash"        description:"交易哈希"`
	Status        int         `json:"status"        orm:"status"         description:"状态:1待确认,2已完成,3已关闭"`
	Network       string      `json:"network"       orm:"network"        description:"网络类型(TRC20/ERC20)"`
	Remark        string      `json:"remark"        orm:"remark"         description:"备注"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:"更新时间"`
}

