// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UsdtWithdrawDao is the data access object for table hg_usdt_withdraw.
type UsdtWithdrawDao struct {
	table   string              // table is the underlying table name of the DAO.
	group   string              // group is the database configuration group name of current DAO.
	columns UsdtWithdrawColumns // columns contains all the column names of Table for convenient usage.
}

// UsdtWithdrawColumns defines and stores column names for table hg_usdt_withdraw.
type UsdtWithdrawColumns struct {
	Id          string // ID
	OrderSn     string // 订单号
	UserId      string // 用户ID
	Amount      string // 提现金额
	Fee         string // 手续费
	RealAmount  string // 实际到账
	ToAddress   string // 提现地址
	TxHash      string // 交易哈希
	Status      string // 状态:1待审核,2审核通过,3审核拒绝,4已汇出,5失败
	Network     string // 网络类型
	AuditRemark string // 审核备注
	AuditTime   string // 审核时间
	CreatedAt   string // 创建时间
	UpdatedAt   string // 更新时间
}

// usdtWithdrawColumns holds the columns for table hg_usdt_withdraw.
var usdtWithdrawColumns = UsdtWithdrawColumns{
	Id:          "id",
	OrderSn:     "order_sn",
	UserId:      "user_id",
	Amount:      "amount",
	Fee:         "fee",
	RealAmount:  "real_amount",
	ToAddress:   "to_address",
	TxHash:      "tx_hash",
	Status:      "status",
	Network:     "network",
	AuditRemark: "audit_remark",
	AuditTime:   "audit_time",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewUsdtWithdrawDao creates and returns a new DAO object for table data access.
func NewUsdtWithdrawDao() *UsdtWithdrawDao {
	return &UsdtWithdrawDao{
		group:   "default",
		table:   "hg_usdt_withdraw",
		columns: usdtWithdrawColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UsdtWithdrawDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current DAO.
func (dao *UsdtWithdrawDao) Table() string {
	return dao.table
}

// Columns returns all column names of current DAO.
func (dao *UsdtWithdrawDao) Columns() UsdtWithdrawColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current DAO.
func (dao *UsdtWithdrawDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UsdtWithdrawDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UsdtWithdrawDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

