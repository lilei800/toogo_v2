// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoWalletDao is the data access object for table hg_toogo_wallet.
type ToogoWalletDao struct {
	table   string             // table is the underlying table name of the DAO.
	group   string             // group is the database configuration group name of current DAO.
	columns ToogoWalletColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoWalletColumns defines and stores column names for table hg_toogo_wallet.
type ToogoWalletColumns struct {
	Id                string // 主键ID
	UserId            string // 用户ID(member_id)
	Balance           string // 余额(USDT)
	FrozenBalance     string // 冻结余额
	Power             string // 算力余额
	FrozenPower       string // 冻结算力
	GiftPower         string // 赠送算力余额
	Commission        string // 佣金余额(USDT)
	FrozenCommission  string // 冻结佣金
	TotalDeposit      string // 累计充值
	TotalWithdraw     string // 累计提现
	TotalPowerConsume string // 累计消耗算力
	TotalCommission   string // 累计获得佣金
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
}

// toogoWalletColumns holds the columns for table hg_toogo_wallet.
var toogoWalletColumns = ToogoWalletColumns{
	Id:                "id",
	UserId:            "user_id",
	Balance:           "balance",
	FrozenBalance:     "frozen_balance",
	Power:             "power",
	FrozenPower:       "frozen_power",
	GiftPower:         "gift_power",
	Commission:        "commission",
	FrozenCommission:  "frozen_commission",
	TotalDeposit:      "total_deposit",
	TotalWithdraw:     "total_withdraw",
	TotalPowerConsume: "total_power_consume",
	TotalCommission:   "total_commission",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewToogoWalletDao creates and returns a new DAO object for table data access.
func NewToogoWalletDao() *ToogoWalletDao {
	return &ToogoWalletDao{
		group:   "default",
		table:   "hg_toogo_wallet",
		columns: toogoWalletColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoWalletDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoWalletDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoWalletDao) Columns() ToogoWalletColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoWalletDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoWalletDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoWalletDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

