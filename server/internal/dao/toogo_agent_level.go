// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoAgentLevelDao is internal type for wrapping internal DAO implements.
type internalToogoAgentLevelDao = *internal.ToogoAgentLevelDao

// toogoAgentLevelDao is the data access object for table hg_toogo_agent_level.
var toogoAgentLevelDao = &toogoAgentLevelDaoImpl{
	internal.NewToogoAgentLevelDao(),
}

// ToogoAgentLevel is the manager for table hg_toogo_agent_level.
var ToogoAgentLevel = toogoAgentLevelDao

type toogoAgentLevelDaoImpl struct {
	internalToogoAgentLevelDao
}

