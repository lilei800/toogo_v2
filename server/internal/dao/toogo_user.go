// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoUserDao is internal type for wrapping internal DAO implements.
type internalToogoUserDao = *internal.ToogoUserDao

// toogoUserDao is the data access object for table hg_toogo_user.
var toogoUserDao = &toogoUserDaoImpl{
	internal.NewToogoUserDao(),
}

// ToogoUser is the manager for table hg_toogo_user.
var ToogoUser = toogoUserDao

type toogoUserDaoImpl struct {
	internalToogoUserDao
}

