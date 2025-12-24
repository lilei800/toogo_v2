// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoCommissionLogDao is internal type for wrapping internal DAO implements.
type internalToogoCommissionLogDao = *internal.ToogoCommissionLogDao

// toogoCommissionLogDao is the data access object for table hg_toogo_commission_log.
var toogoCommissionLogDao = &toogoCommissionLogDaoImpl{
	internal.NewToogoCommissionLogDao(),
}

// ToogoCommissionLog is the manager for table hg_toogo_commission_log.
var ToogoCommissionLog = toogoCommissionLogDao

type toogoCommissionLogDaoImpl struct {
	internalToogoCommissionLogDao
}

