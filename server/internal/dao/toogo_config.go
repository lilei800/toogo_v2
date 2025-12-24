// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoConfigDao is internal type for wrapping internal DAO implements.
type internalToogoConfigDao = *internal.ToogoConfigDao

// toogoConfigDao is the data access object for table hg_toogo_config.
var toogoConfigDao = &toogoConfigDaoImpl{
	internal.NewToogoConfigDao(),
}

// ToogoConfig is the manager for table hg_toogo_config.
var ToogoConfig = toogoConfigDao

type toogoConfigDaoImpl struct {
	internalToogoConfigDao
}

