// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoWalletLogDao is the data access object for table hg_toogo_wallet_log.
type ToogoWalletLogDao struct {
	table   string                // table is the underlying table name of the DAO.
	group   string                // group is the database configuration group name of current DAO.
	columns ToogoWalletLogColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoWalletLogColumns defines and stores column names for table hg_toogo_wallet_log.
type ToogoWalletLogColumns struct {
	Id           string
	UserId       string
	AccountType  string
	ChangeType   string
	ChangeAmount string
	BeforeAmount string
	AfterAmount  string
	RelatedId    string
	RelatedType  string
	OrderSn      string
	Remark       string
	CreatedAt    string
}

// toogoWalletLogColumns holds the columns for table hg_toogo_wallet_log.
var toogoWalletLogColumns = ToogoWalletLogColumns{
	Id:           "id",
	UserId:       "user_id",
	AccountType:  "account_type",
	ChangeType:   "change_type",
	ChangeAmount: "change_amount",
	BeforeAmount: "before_amount",
	AfterAmount:  "after_amount",
	RelatedId:    "related_id",
	RelatedType:  "related_type",
	OrderSn:      "order_sn",
	Remark:       "remark",
	CreatedAt:    "created_at",
}

// NewToogoWalletLogDao creates and returns a new DAO object for table data access.
func NewToogoWalletLogDao() *ToogoWalletLogDao {
	return &ToogoWalletLogDao{
		group:   "default",
		table:   "hg_toogo_wallet_log",
		columns: toogoWalletLogColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoWalletLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoWalletLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoWalletLogDao) Columns() ToogoWalletLogColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoWalletLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO.
func (dao *ToogoWalletLogDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoWalletLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

