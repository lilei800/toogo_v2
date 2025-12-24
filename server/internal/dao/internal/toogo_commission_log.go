// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoCommissionLogDao is the data access object for table hg_toogo_commission_log.
type ToogoCommissionLogDao struct {
	table   string                    // table is the underlying table name of the DAO.
	group   string                    // group is the database configuration group name of current DAO.
	columns ToogoCommissionLogColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoCommissionLogColumns defines and stores column names for table hg_toogo_commission_log.
type ToogoCommissionLogColumns struct {
	Id               string // 主键ID
	UserId           string // 获得佣金用户ID
	FromUserId       string // 来源用户ID
	CommissionType   string // 佣金类型
	Level            string // 层级
	BaseAmount       string // 基础金额
	CommissionRate   string // 佣金比例
	CommissionAmount string // 佣金金额
	SettleType       string // 结算类型
	Status           string // 状态
	RelatedId        string // 关联ID
	RelatedType      string // 关联类型
	OrderSn          string // 关联订单号
	Remark           string // 备注
	CreatedAt        string // 创建时间
}

// toogoCommissionLogColumns holds the columns for table hg_toogo_commission_log.
var toogoCommissionLogColumns = ToogoCommissionLogColumns{
	Id:               "id",
	UserId:           "user_id",
	FromUserId:       "from_user_id",
	CommissionType:   "commission_type",
	Level:            "level",
	BaseAmount:       "base_amount",
	CommissionRate:   "commission_rate",
	CommissionAmount: "commission_amount",
	SettleType:       "settle_type",
	Status:           "status",
	RelatedId:        "related_id",
	RelatedType:      "related_type",
	OrderSn:          "order_sn",
	Remark:           "remark",
	CreatedAt:        "created_at",
}

// NewToogoCommissionLogDao creates and returns a new DAO object for table data access.
func NewToogoCommissionLogDao() *ToogoCommissionLogDao {
	return &ToogoCommissionLogDao{
		group:   "default",
		table:   "hg_toogo_commission_log",
		columns: toogoCommissionLogColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoCommissionLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoCommissionLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoCommissionLogDao) Columns() ToogoCommissionLogColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoCommissionLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoCommissionLogDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoCommissionLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

