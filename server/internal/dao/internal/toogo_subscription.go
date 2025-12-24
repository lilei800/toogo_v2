// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoSubscriptionDao is the data access object for table hg_toogo_subscription.
type ToogoSubscriptionDao struct {
	table   string                   // table is the underlying table name of the DAO.
	group   string                   // group is the database configuration group name of current DAO.
	columns ToogoSubscriptionColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoSubscriptionColumns defines and stores column names for table hg_toogo_subscription.
type ToogoSubscriptionColumns struct {
	Id                string // 主键ID
	UserId            string // 用户ID
	PlanId            string // 套餐ID
	PlanCode          string // 套餐代码
	OrderSn           string // 订单号
	PeriodType        string // 订阅周期
	Amount            string // 订阅金额
	GiftPower         string // 赠送算力
	StartTime         string // 开始时间
	ExpireTime        string // 到期时间
	Days              string // 订阅天数
	Status            string // 状态
	PaidAt            string // 支付时间
	PayType           string // 支付方式
	InviterId         string // 邀请人ID
	CommissionSettled string // 佣金是否已结算
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
}

// toogoSubscriptionColumns holds the columns for table hg_toogo_subscription.
var toogoSubscriptionColumns = ToogoSubscriptionColumns{
	Id:                "id",
	UserId:            "user_id",
	PlanId:            "plan_id",
	PlanCode:          "plan_code",
	OrderSn:           "order_sn",
	PeriodType:        "period_type",
	Amount:            "amount",
	GiftPower:         "gift_power",
	StartTime:         "start_time",
	ExpireTime:        "expire_time",
	Days:              "days",
	Status:            "status",
	PaidAt:            "paid_at",
	PayType:           "pay_type",
	InviterId:         "inviter_id",
	CommissionSettled: "commission_settled",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewToogoSubscriptionDao creates and returns a new DAO object for table data access.
func NewToogoSubscriptionDao() *ToogoSubscriptionDao {
	return &ToogoSubscriptionDao{
		group:   "default",
		table:   "hg_toogo_subscription",
		columns: toogoSubscriptionColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoSubscriptionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoSubscriptionDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoSubscriptionDao) Columns() ToogoSubscriptionColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoSubscriptionDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoSubscriptionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoSubscriptionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

