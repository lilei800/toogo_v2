// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoWithdrawDao is internal type for wrapping internal DAO implements.
type internalToogoWithdrawDao = *internal.ToogoWithdrawDao

// toogoWithdrawDao is the data access object for table hg_toogo_withdraw.
var toogoWithdrawDao = &toogoWithdrawDaoImpl{
	internal.NewToogoWithdrawDao(),
}

// ToogoWithdraw is the manager for table hg_toogo_withdraw.
var ToogoWithdraw = toogoWithdrawDao

type toogoWithdrawDaoImpl struct {
	internalToogoWithdrawDao
}

