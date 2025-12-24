// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoVipLevelDao is the data access object for table hg_toogo_vip_level.
type ToogoVipLevelDao struct {
	table   string               // table is the underlying table name of the DAO.
	group   string               // group is the database configuration group name of current DAO.
	columns ToogoVipLevelColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoVipLevelColumns defines and stores column names for table hg_toogo_vip_level.
type ToogoVipLevelColumns struct {
	Id                  string // 主键ID
	Level               string // 等级
	LevelName           string // 等级名称
	RequireInviteCount  string // 需要邀请人数
	RequireConsumePower string // 需要消耗算力
	RequireTeamConsume  string // 需要团队消耗算力
	PowerDiscount       string // 算力折扣
	InviteRewardPower   string // 邀请奖励算力
	Description         string // 等级描述
	Icon                string // 等级图标
	Sort                string // 排序
	Status              string // 状态
	CreatedAt           string // 创建时间
	UpdatedAt           string // 更新时间
}

// toogoVipLevelColumns holds the columns for table hg_toogo_vip_level.
var toogoVipLevelColumns = ToogoVipLevelColumns{
	Id:                  "id",
	Level:               "level",
	LevelName:           "level_name",
	RequireInviteCount:  "require_invite_count",
	RequireConsumePower: "require_consume_power",
	RequireTeamConsume:  "require_team_consume",
	PowerDiscount:       "power_discount",
	InviteRewardPower:   "invite_reward_power",
	Description:         "description",
	Icon:                "icon",
	Sort:                "sort",
	Status:              "status",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
}

// NewToogoVipLevelDao creates and returns a new DAO object for table data access.
func NewToogoVipLevelDao() *ToogoVipLevelDao {
	return &ToogoVipLevelDao{
		group:   "default",
		table:   "hg_toogo_vip_level",
		columns: toogoVipLevelColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoVipLevelDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoVipLevelDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoVipLevelDao) Columns() ToogoVipLevelColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoVipLevelDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoVipLevelDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoVipLevelDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

