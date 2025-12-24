// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoTransferDao is internal type for wrapping internal DAO implements.
type internalToogoTransferDao = *internal.ToogoTransferDao

// toogoTransferDao is the data access object for table hg_toogo_transfer.
var toogoTransferDao = &toogoTransferDaoImpl{
	internal.NewToogoTransferDao(),
}

// ToogoTransfer is the manager for table hg_toogo_transfer.
var ToogoTransfer = toogoTransferDao

type toogoTransferDaoImpl struct {
	internalToogoTransferDao
}

