// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoWithdrawDao is the data access object for table hg_toogo_withdraw.
type ToogoWithdrawDao struct {
	table   string               // table is the underlying table name of the DAO.
	group   string               // group is the database configuration group name of current DAO.
	columns ToogoWithdrawColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoWithdrawColumns defines and stores column names for table hg_toogo_withdraw.
type ToogoWithdrawColumns struct {
	Id          string
	UserId      string
	OrderSn     string
	AccountType string
	Amount      string
	Fee         string
	RealAmount  string
	ToAddress   string
	Network     string
	TxHash      string
	Status      string
	AuditRemark string
	AuditedBy   string
	AuditedAt   string
	CompletedAt string
	Remark      string
	CreatedAt   string
	UpdatedAt   string
}

// toogoWithdrawColumns holds the columns for table hg_toogo_withdraw.
var toogoWithdrawColumns = ToogoWithdrawColumns{
	Id:          "id",
	UserId:      "user_id",
	OrderSn:     "order_sn",
	AccountType: "account_type",
	Amount:      "amount",
	Fee:         "fee",
	RealAmount:  "real_amount",
	ToAddress:   "to_address",
	Network:     "network",
	TxHash:      "tx_hash",
	Status:      "status",
	AuditRemark: "audit_remark",
	AuditedBy:   "audited_by",
	AuditedAt:   "audited_at",
	CompletedAt: "completed_at",
	Remark:      "remark",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewToogoWithdrawDao creates and returns a new DAO object for table data access.
func NewToogoWithdrawDao() *ToogoWithdrawDao {
	return &ToogoWithdrawDao{
		group:   "default",
		table:   "hg_toogo_withdraw",
		columns: toogoWithdrawColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoWithdrawDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoWithdrawDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoWithdrawDao) Columns() ToogoWithdrawColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoWithdrawDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO.
func (dao *ToogoWithdrawDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoWithdrawDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

