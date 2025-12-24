// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// usdtBalanceDao is the data access object for table hg_usdt_balance.
type usdtBalanceDao struct {
	*internal.UsdtBalanceDao
}

var (
	// UsdtBalance is the globally accessible object for table hg_usdt_balance operations.
	UsdtBalance = usdtBalanceDao{internal.NewUsdtBalanceDao()}
)

