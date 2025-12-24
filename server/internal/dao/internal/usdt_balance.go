// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UsdtBalanceDao is the data access object for table hg_usdt_balance.
type UsdtBalanceDao struct {
	table   string             // table is the underlying table name of the DAO.
	group   string             // group is the database configuration group name of current DAO.
	columns UsdtBalanceColumns // columns contains all the column names of Table for convenient usage.
}

// UsdtBalanceColumns defines and stores column names for table hg_usdt_balance.
type UsdtBalanceColumns struct {
	UserId        string // 用户ID
	Balance       string // 可用余额
	FrozenBalance string // 冻结余额
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
}

// usdtBalanceColumns holds the columns for table hg_usdt_balance.
var usdtBalanceColumns = UsdtBalanceColumns{
	UserId:        "user_id",
	Balance:       "balance",
	FrozenBalance: "frozen_balance",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewUsdtBalanceDao creates and returns a new DAO object for table data access.
func NewUsdtBalanceDao() *UsdtBalanceDao {
	return &UsdtBalanceDao{
		group:   "default",
		table:   "hg_usdt_balance",
		columns: usdtBalanceColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UsdtBalanceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current DAO.
func (dao *UsdtBalanceDao) Table() string {
	return dao.table
}

// Columns returns all column names of current DAO.
func (dao *UsdtBalanceDao) Columns() UsdtBalanceColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current DAO.
func (dao *UsdtBalanceDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UsdtBalanceDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UsdtBalanceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

