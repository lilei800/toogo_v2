// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoTransferDao is the data access object for table hg_toogo_transfer.
type ToogoTransferDao struct {
	table   string               // table is the underlying table name of the DAO.
	group   string               // group is the database configuration group name of current DAO.
	columns ToogoTransferColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoTransferColumns defines and stores column names for table hg_toogo_transfer.
type ToogoTransferColumns struct {
	Id          string
	UserId      string
	OrderSn     string
	FromAccount string
	ToAccount   string
	Amount      string
	PowerAmount string
	Rate        string
	Status      string
	Remark      string
	CreatedAt   string
}

// toogoTransferColumns holds the columns for table hg_toogo_transfer.
var toogoTransferColumns = ToogoTransferColumns{
	Id:          "id",
	UserId:      "user_id",
	OrderSn:     "order_sn",
	FromAccount: "from_account",
	ToAccount:   "to_account",
	Amount:      "amount",
	PowerAmount: "power_amount",
	Rate:        "rate",
	Status:      "status",
	Remark:      "remark",
	CreatedAt:   "created_at",
}

// NewToogoTransferDao creates and returns a new DAO object for table data access.
func NewToogoTransferDao() *ToogoTransferDao {
	return &ToogoTransferDao{
		group:   "default",
		table:   "hg_toogo_transfer",
		columns: toogoTransferColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoTransferDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoTransferDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoTransferDao) Columns() ToogoTransferColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoTransferDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO.
func (dao *ToogoTransferDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoTransferDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

