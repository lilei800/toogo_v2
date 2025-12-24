// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoDeposit is the golang structure for table hg_toogo_deposit.
type ToogoDeposit struct {
	Id             int64       `json:"id"             orm:"id"              description:"主键ID"`
	UserId         int64       `json:"userId"         orm:"user_id"         description:"用户ID(member_id)"`
	OrderSn        string      `json:"orderSn"        orm:"order_sn"        description:"订单号"`
	Amount         float64     `json:"amount"         orm:"amount"          description:"充值金额(USDT)"`
	RealAmount     float64     `json:"realAmount"     orm:"real_amount"     description:"实际到账金额"`
	Network        string      `json:"network"        orm:"network"         description:"网络: TRC20/ERC20/BEP20"`
	ToAddress      string      `json:"toAddress"      orm:"to_address"      description:"充值地址"`
	PaymentChannel string      `json:"paymentChannel" orm:"payment_channel" description:"支付渠道"`
	PaymentId      string      `json:"paymentId"      orm:"payment_id"      description:"第三方支付ID"`
	TxHash         string      `json:"txHash"         orm:"tx_hash"         description:"交易哈希"`
	Status         int         `json:"status"         orm:"status"          description:"状态: 1=待支付, 2=已完成, 3=已超时, 4=已取消"`
	ExpireTime     *gtime.Time `json:"expireTime"     orm:"expire_time"     description:"过期时间"`
	PaidAt         *gtime.Time `json:"paidAt"         orm:"paid_at"         description:"支付时间"`
	Remark         string      `json:"remark"         orm:"remark"          description:"备注"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"创建时间"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"更新时间"`
}

