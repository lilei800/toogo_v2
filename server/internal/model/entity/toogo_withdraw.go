// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoWithdraw is the golang structure for table hg_toogo_withdraw.
type ToogoWithdraw struct {
	Id          int64       `json:"id"          orm:"id"           description:"主键ID"`
	UserId      int64       `json:"userId"      orm:"user_id"      description:"用户ID(member_id)"`
	OrderSn     string      `json:"orderSn"     orm:"order_sn"     description:"订单号"`
	AccountType string      `json:"accountType" orm:"account_type" description:"账户类型: balance/commission"`
	Amount      float64     `json:"amount"      orm:"amount"       description:"提现金额(USDT)"`
	Fee         float64     `json:"fee"         orm:"fee"          description:"手续费"`
	RealAmount  float64     `json:"realAmount"  orm:"real_amount"  description:"实际到账金额"`
	ToAddress   string      `json:"toAddress"   orm:"to_address"   description:"提现地址"`
	Network     string      `json:"network"     orm:"network"      description:"网络: TRC20/ERC20/BEP20"`
	TxHash      string      `json:"txHash"      orm:"tx_hash"      description:"交易哈希"`
	Status      int         `json:"status"      orm:"status"       description:"状态: 1=待审核, 2=审核通过, 3=审核拒绝, 4=已完成, 5=已取消"`
	AuditRemark string      `json:"auditRemark" orm:"audit_remark" description:"审核备注"`
	AuditedBy   int64       `json:"auditedBy"   orm:"audited_by"   description:"审核人ID"`
	AuditedAt   *gtime.Time `json:"auditedAt"   orm:"audited_at"   description:"审核时间"`
	CompletedAt *gtime.Time `json:"completedAt" orm:"completed_at" description:"完成时间"`
	Remark      string      `json:"remark"      orm:"remark"       description:"备注"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:"创建时间"`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:"更新时间"`
}

