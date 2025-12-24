// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoConfigDao is the data access object for table hg_toogo_config.
type ToogoConfigDao struct {
	table   string             // table is the underlying table name of the DAO.
	group   string             // group is the database configuration group name of current DAO.
	columns ToogoConfigColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoConfigColumns defines and stores column names for table hg_toogo_config.
type ToogoConfigColumns struct {
	Id          string // 主键ID
	Group       string // 配置分组
	Key         string // 配置KEY
	Value       string // 配置值
	Type        string // 值类型
	Name        string // 配置名称
	Description string // 配置描述
	Sort        string // 排序
	CreatedAt   string // 创建时间
	UpdatedAt   string // 更新时间
}

// toogoConfigColumns holds the columns for table hg_toogo_config.
var toogoConfigColumns = ToogoConfigColumns{
	Id:          "id",
	Group:       "group",
	Key:         "key",
	Value:       "value",
	Type:        "type",
	Name:        "name",
	Description: "description",
	Sort:        "sort",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewToogoConfigDao creates and returns a new DAO object for table data access.
func NewToogoConfigDao() *ToogoConfigDao {
	return &ToogoConfigDao{
		group:   "default",
		table:   "hg_toogo_config",
		columns: toogoConfigColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoConfigDao) Columns() ToogoConfigColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO.
func (dao *ToogoConfigDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

