// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoDepositDao is the data access object for table hg_toogo_deposit.
type ToogoDepositDao struct {
	table   string              // table is the underlying table name of the DAO.
	group   string              // group is the database configuration group name of current DAO.
	columns ToogoDepositColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoDepositColumns defines and stores column names for table hg_toogo_deposit.
type ToogoDepositColumns struct {
	Id             string
	UserId         string
	OrderSn        string
	Amount         string
	RealAmount     string
	Network        string
	ToAddress      string
	PaymentChannel string
	PaymentId      string
	TxHash         string
	Status         string
	ExpireTime     string
	PaidAt         string
	Remark         string
	CreatedAt      string
	UpdatedAt      string
}

// toogoDepositColumns holds the columns for table hg_toogo_deposit.
var toogoDepositColumns = ToogoDepositColumns{
	Id:             "id",
	UserId:         "user_id",
	OrderSn:        "order_sn",
	Amount:         "amount",
	RealAmount:     "real_amount",
	Network:        "network",
	ToAddress:      "to_address",
	PaymentChannel: "payment_channel",
	PaymentId:      "payment_id",
	TxHash:         "tx_hash",
	Status:         "status",
	ExpireTime:     "expire_time",
	PaidAt:         "paid_at",
	Remark:         "remark",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
}

// NewToogoDepositDao creates and returns a new DAO object for table data access.
func NewToogoDepositDao() *ToogoDepositDao {
	return &ToogoDepositDao{
		group:   "default",
		table:   "hg_toogo_deposit",
		columns: toogoDepositColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoDepositDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoDepositDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoDepositDao) Columns() ToogoDepositColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoDepositDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO.
func (dao *ToogoDepositDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoDepositDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

