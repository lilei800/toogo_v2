// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoVolatilityConfigDao is internal type for wrapping internal DAO implements.
type internalToogoVolatilityConfigDao = *internal.ToogoVolatilityConfigDao

// toogoVolatilityConfigDao is the data access object for table hg_toogo_volatility_config.
var toogoVolatilityConfigDao = &toogoVolatilityConfigDaoImpl{
	internal.NewToogoVolatilityConfigDao(),
}

// ToogoVolatilityConfig is the manager for table hg_toogo_volatility_config.
var ToogoVolatilityConfig = toogoVolatilityConfigDao

type toogoVolatilityConfigDaoImpl struct {
	internalToogoVolatilityConfigDao
}

