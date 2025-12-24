// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoPlanDao is internal type for wrapping internal DAO implements.
type internalToogoPlanDao = *internal.ToogoPlanDao

// toogoPlanDao is the data access object for table hg_toogo_plan.
var toogoPlanDao = &toogoPlanDaoImpl{
	internal.NewToogoPlanDao(),
}

// ToogoPlan is the manager for table hg_toogo_plan.
var ToogoPlan = toogoPlanDao

type toogoPlanDaoImpl struct {
	internalToogoPlanDao
}

