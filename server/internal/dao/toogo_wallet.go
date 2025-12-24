// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoWalletDao is internal type for wrapping internal DAO implements.
type internalToogoWalletDao = *internal.ToogoWalletDao

// toogoWalletDao is the data access object for table hg_toogo_wallet.
var toogoWalletDao = &toogoWalletDaoImpl{
	internal.NewToogoWalletDao(),
}

// ToogoWallet is the manager for table hg_toogo_wallet.
var ToogoWallet = toogoWalletDao

type toogoWalletDaoImpl struct {
	internalToogoWalletDao
}

