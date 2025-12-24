// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoWallet is the golang structure for table hg_toogo_wallet.
type ToogoWallet struct {
	Id                int64       `json:"id"                orm:"id"                  description:"主键ID"`
	UserId            int64       `json:"userId"            orm:"user_id"             description:"用户ID(member_id)"`
	Balance           float64     `json:"balance"           orm:"balance"             description:"余额(USDT)"`
	FrozenBalance     float64     `json:"frozenBalance"     orm:"frozen_balance"      description:"冻结余额"`
	Power             float64     `json:"power"             orm:"power"               description:"算力余额"`
	FrozenPower       float64     `json:"frozenPower"       orm:"frozen_power"        description:"冻结算力"`
	GiftPower         float64     `json:"giftPower"         orm:"gift_power"          description:"积分余额"`
	Commission        float64     `json:"commission"        orm:"commission"          description:"佣金余额(USDT)"`
	FrozenCommission  float64     `json:"frozenCommission"  orm:"frozen_commission"   description:"冻结佣金"`
	TotalDeposit      float64     `json:"totalDeposit"      orm:"total_deposit"       description:"累计充值"`
	TotalWithdraw     float64     `json:"totalWithdraw"     orm:"total_withdraw"      description:"累计提现"`
	TotalPowerConsume float64     `json:"totalPowerConsume" orm:"total_power_consume" description:"累计消耗算力"`
	TotalCommission   float64     `json:"totalCommission"   orm:"total_commission"    description:"累计获得佣金"`
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"          description:"创建时间"`
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"          description:"更新时间"`
}

