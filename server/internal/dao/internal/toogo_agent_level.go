// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoAgentLevelDao is the data access object for table hg_toogo_agent_level.
type ToogoAgentLevelDao struct {
	table   string                 // table is the underlying table name of the DAO.
	group   string                 // group is the database configuration group name of current DAO.
	columns ToogoAgentLevelColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoAgentLevelColumns defines and stores column names for table hg_toogo_agent_level.
type ToogoAgentLevelColumns struct {
	Id                   string // 主键ID
	Level                string // 等级
	LevelName            string // 等级名称
	RequireTeamCount     string // 需要团队人数
	RequireTeamSubscribe string // 需要团队订阅额
	SubscribeRate1       string // 订阅佣金比例(一级)
	SubscribeRate2       string // 订阅佣金比例(二级)
	SubscribeRate3       string // 订阅佣金比例(三级)
	PowerRate1           string // 算力消耗佣金比例(一级)
	PowerRate2           string // 算力消耗佣金比例(二级)
	PowerRate3           string // 算力消耗佣金比例(三级)
	Description          string // 等级描述
	Sort                 string // 排序
	Status               string // 状态
	CreatedAt            string // 创建时间
	UpdatedAt            string // 更新时间
}

// toogoAgentLevelColumns holds the columns for table hg_toogo_agent_level.
var toogoAgentLevelColumns = ToogoAgentLevelColumns{
	Id:                   "id",
	Level:                "level",
	LevelName:            "level_name",
	RequireTeamCount:     "require_team_count",
	RequireTeamSubscribe: "require_team_subscribe",
	SubscribeRate1:       "subscribe_rate_1",
	SubscribeRate2:       "subscribe_rate_2",
	SubscribeRate3:       "subscribe_rate_3",
	PowerRate1:           "power_rate_1",
	PowerRate2:           "power_rate_2",
	PowerRate3:           "power_rate_3",
	Description:          "description",
	Sort:                 "sort",
	Status:               "status",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
}

// NewToogoAgentLevelDao creates and returns a new DAO object for table data access.
func NewToogoAgentLevelDao() *ToogoAgentLevelDao {
	return &ToogoAgentLevelDao{
		group:   "default",
		table:   "hg_toogo_agent_level",
		columns: toogoAgentLevelColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoAgentLevelDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoAgentLevelDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoAgentLevelDao) Columns() ToogoAgentLevelColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoAgentLevelDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoAgentLevelDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoAgentLevelDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

