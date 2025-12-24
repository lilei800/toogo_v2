// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoVipLevelDao is internal type for wrapping internal DAO implements.
type internalToogoVipLevelDao = *internal.ToogoVipLevelDao

// toogoVipLevelDao is the data access object for table hg_toogo_vip_level.
var toogoVipLevelDao = &toogoVipLevelDaoImpl{
	internal.NewToogoVipLevelDao(),
}

// ToogoVipLevel is the manager for table hg_toogo_vip_level.
var ToogoVipLevel = toogoVipLevelDao

type toogoVipLevelDaoImpl struct {
	internalToogoVipLevelDao
}

