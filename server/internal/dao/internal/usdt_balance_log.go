// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UsdtBalanceLogDao is the data access object for table hg_usdt_balance_log.
type UsdtBalanceLogDao struct {
	table   string                // table is the underlying table name of the DAO.
	group   string                // group is the database configuration group name of current DAO.
	columns UsdtBalanceLogColumns // columns contains all the column names of Table for convenient usage.
}

// UsdtBalanceLogColumns defines and stores column names for table hg_usdt_balance_log.
type UsdtBalanceLogColumns struct {
	Id            string // ID
	UserId        string // 用户ID
	OrderSn       string // 关联订单号
	Type          string // 变动类型:1充值,2提现,3交易
	ChangeAmount  string // 变动金额
	BeforeBalance string // 变动前余额
	AfterBalance  string // 变动后余额
	Remark        string // 备注
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
}

// usdtBalanceLogColumns holds the columns for table hg_usdt_balance_log.
var usdtBalanceLogColumns = UsdtBalanceLogColumns{
	Id:            "id",
	UserId:        "user_id",
	OrderSn:       "order_sn",
	Type:          "type",
	ChangeAmount:  "change_amount",
	BeforeBalance: "before_balance",
	AfterBalance:  "after_balance",
	Remark:        "remark",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewUsdtBalanceLogDao creates and returns a new DAO object for table data access.
func NewUsdtBalanceLogDao() *UsdtBalanceLogDao {
	return &UsdtBalanceLogDao{
		group:   "default",
		table:   "hg_usdt_balance_log",
		columns: usdtBalanceLogColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UsdtBalanceLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current DAO.
func (dao *UsdtBalanceLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of current DAO.
func (dao *UsdtBalanceLogDao) Columns() UsdtBalanceLogColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current DAO.
func (dao *UsdtBalanceLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UsdtBalanceLogDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UsdtBalanceLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

