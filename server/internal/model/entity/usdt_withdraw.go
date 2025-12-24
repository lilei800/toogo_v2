// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UsdtWithdraw is the golang structure for table usdt_withdraw.
type UsdtWithdraw struct {
	Id            int64       `json:"id"            orm:"id"             description:"ID"`
	OrderSn       string      `json:"orderSn"       orm:"order_sn"       description:"订单号"`
	UserId        int64       `json:"userId"        orm:"user_id"        description:"用户ID"`
	Amount        float64     `json:"amount"        orm:"amount"         description:"提现金额"`
	Fee           float64     `json:"fee"           orm:"fee"            description:"手续费"`
	RealAmount    float64     `json:"realAmount"    orm:"real_amount"    description:"实际到账"`
	ToAddress     string      `json:"toAddress"     orm:"to_address"     description:"提现地址"`
	TxHash        string      `json:"txHash"        orm:"tx_hash"        description:"交易哈希"`
	Status        int         `json:"status"        orm:"status"         description:"状态:1待审核,2审核通过,3审核拒绝,4已汇出,5失败"`
	Network       string      `json:"network"       orm:"network"        description:"网络类型"`
	AuditRemark   string      `json:"auditRemark"   orm:"audit_remark"   description:"审核备注"`
	AuditTime     *gtime.Time `json:"auditTime"     orm:"audit_time"     description:"审核时间"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:"更新时间"`
}

