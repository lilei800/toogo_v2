// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// usdtBalanceLogDao is the data access object for table hg_usdt_balance_log.
type usdtBalanceLogDao struct {
	*internal.UsdtBalanceLogDao
}

var (
	// UsdtBalanceLog is the globally accessible object for table hg_usdt_balance_log operations.
	UsdtBalanceLog = usdtBalanceLogDao{internal.NewUsdtBalanceLogDao()}
)

