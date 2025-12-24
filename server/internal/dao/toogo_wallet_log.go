// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoWalletLogDao is internal type for wrapping internal DAO implements.
type internalToogoWalletLogDao = *internal.ToogoWalletLogDao

// toogoWalletLogDao is the data access object for table hg_toogo_wallet_log.
var toogoWalletLogDao = &toogoWalletLogDaoImpl{
	internal.NewToogoWalletLogDao(),
}

// ToogoWalletLog is the manager for table hg_toogo_wallet_log.
var ToogoWalletLog = toogoWalletLogDao

type toogoWalletLogDaoImpl struct {
	internalToogoWalletLogDao
}

