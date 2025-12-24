// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// usdtDepositDao is the data access object for table hg_usdt_deposit.
type usdtDepositDao struct {
	*internal.UsdtDepositDao
}

var (
	// UsdtDeposit is the globally accessible object for table hg_usdt_deposit operations.
	UsdtDeposit = usdtDepositDao{internal.NewUsdtDepositDao()}
)

