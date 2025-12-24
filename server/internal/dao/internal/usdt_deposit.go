// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UsdtDepositDao is the data access object for table hg_usdt_deposit.
type UsdtDepositDao struct {
	table   string             // table is the underlying table name of the DAO.
	group   string             // group is the database configuration group name of current DAO.
	columns UsdtDepositColumns // columns contains all the column names of Table for convenient usage.
}

// UsdtDepositColumns defines and stores column names for table hg_usdt_deposit.
type UsdtDepositColumns struct {
	Id        string // ID
	OrderSn   string // 订单号
	UserId    string // 用户ID
	Amount    string // 充值金额
	ToAddress string // 充值地址
	TxHash    string // 交易哈希
	Status    string // 状态:1待确认,2已完成,3已关闭
	Network   string // 网络类型(TRC20/ERC20)
	Remark    string // 备注
	CreatedAt string // 创建时间
	UpdatedAt string // 更新时间
}

// usdtDepositColumns holds the columns for table hg_usdt_deposit.
var usdtDepositColumns = UsdtDepositColumns{
	Id:        "id",
	OrderSn:   "order_sn",
	UserId:    "user_id",
	Amount:    "amount",
	ToAddress: "to_address",
	TxHash:    "tx_hash",
	Status:    "status",
	Network:   "network",
	Remark:    "remark",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewUsdtDepositDao creates and returns a new DAO object for table data access.
func NewUsdtDepositDao() *UsdtDepositDao {
	return &UsdtDepositDao{
		group:   "default",
		table:   "hg_usdt_deposit",
		columns: usdtDepositColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UsdtDepositDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current DAO.
func (dao *UsdtDepositDao) Table() string {
	return dao.table
}

// Columns returns all column names of current DAO.
func (dao *UsdtDepositDao) Columns() UsdtDepositColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current DAO.
func (dao *UsdtDepositDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UsdtDepositDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UsdtDepositDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

