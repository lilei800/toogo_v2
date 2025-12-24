// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoPowerConsumeDao is internal type for wrapping internal DAO implements.
type internalToogoPowerConsumeDao = *internal.ToogoPowerConsumeDao

// toogoPowerConsumeDao is the data access object for table hg_toogo_power_consume.
var toogoPowerConsumeDao = &toogoPowerConsumeDaoImpl{
	internal.NewToogoPowerConsumeDao(),
}

// ToogoPowerConsume is the manager for table hg_toogo_power_consume.
var ToogoPowerConsume = toogoPowerConsumeDao

type toogoPowerConsumeDaoImpl struct {
	internalToogoPowerConsumeDao
}

