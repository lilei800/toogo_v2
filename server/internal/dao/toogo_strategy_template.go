// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoStrategyTemplateDao is internal type for wrapping internal DAO implements.
type internalToogoStrategyTemplateDao = *internal.ToogoStrategyTemplateDao

// toogoStrategyTemplateDao is the data access object for table hg_toogo_strategy_template.
var toogoStrategyTemplateDao = &toogoStrategyTemplateDaoImpl{
	internal.NewToogoStrategyTemplateDao(),
}

// ToogoStrategyTemplate is the manager for table hg_toogo_strategy_template.
var ToogoStrategyTemplate = toogoStrategyTemplateDao

type toogoStrategyTemplateDaoImpl struct {
	internalToogoStrategyTemplateDao
}

