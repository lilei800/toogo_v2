// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UsdtBalance is the golang structure for table usdt_balance.
type UsdtBalance struct {
	UserId        int64       `json:"userId"        orm:"user_id"        description:"用户ID"`
	Balance       float64     `json:"balance"       orm:"balance"        description:"可用余额"`
	FrozenBalance float64     `json:"frozenBalance" orm:"frozen_balance" description:"冻结余额"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:"更新时间"`
}

