// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// usdtWithdrawDao is the data access object for table hg_usdt_withdraw.
type usdtWithdrawDao struct {
	*internal.UsdtWithdrawDao
}

var (
	// UsdtWithdraw is the globally accessible object for table hg_usdt_withdraw operations.
	UsdtWithdraw = usdtWithdrawDao{internal.NewUsdtWithdrawDao()}
)

