// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoPowerConsumeDao is the data access object for table hg_toogo_power_consume.
type ToogoPowerConsumeDao struct {
	table   string                   // table is the underlying table name of the DAO.
	group   string                   // group is the database configuration group name of current DAO.
	columns ToogoPowerConsumeColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoPowerConsumeColumns defines and stores column names for table hg_toogo_power_consume.
type ToogoPowerConsumeColumns struct {
	Id            string // 主键ID
	UserId        string // 用户ID
	RobotId       string // 机器人ID
	OrderId       string // 交易订单ID
	OrderSn       string // 订单号
	ProfitAmount  string // 盈利金额
	ConsumeRate   string // 消耗比例
	ConsumePower  string // 消耗算力
	FromPower     string // 从算力账户扣除
	VipLevel      string // 用户VIP等级
	DiscountRate  string // 折扣比例
	OriginalPower string // 原始消耗算力
	CreatedAt     string // 创建时间
}

// toogoPowerConsumeColumns holds the columns for table hg_toogo_power_consume.
var toogoPowerConsumeColumns = ToogoPowerConsumeColumns{
	Id:            "id",
	UserId:        "user_id",
	RobotId:       "robot_id",
	OrderId:       "order_id",
	OrderSn:       "order_sn",
	ProfitAmount:  "profit_amount",
	ConsumeRate:   "consume_rate",
	ConsumePower:  "consume_power",
	FromPower:     "from_power",
	VipLevel:      "vip_level",
	DiscountRate:  "discount_rate",
	OriginalPower: "original_power",
	CreatedAt:     "created_at",
}

// NewToogoPowerConsumeDao creates and returns a new DAO object for table data access.
func NewToogoPowerConsumeDao() *ToogoPowerConsumeDao {
	return &ToogoPowerConsumeDao{
		group:   "default",
		table:   "hg_toogo_power_consume",
		columns: toogoPowerConsumeColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoPowerConsumeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoPowerConsumeDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoPowerConsumeDao) Columns() ToogoPowerConsumeColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoPowerConsumeDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoPowerConsumeDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoPowerConsumeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
