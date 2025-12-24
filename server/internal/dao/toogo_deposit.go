// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoDepositDao is internal type for wrapping internal DAO implements.
type internalToogoDepositDao = *internal.ToogoDepositDao

// toogoDepositDao is the data access object for table hg_toogo_deposit.
var toogoDepositDao = &toogoDepositDaoImpl{
	internal.NewToogoDepositDao(),
}

// ToogoDeposit is the manager for table hg_toogo_deposit.
var ToogoDeposit = toogoDepositDao

type toogoDepositDaoImpl struct {
	internalToogoDepositDao
}

