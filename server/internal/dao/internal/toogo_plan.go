// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoPlanDao is the data access object for table hg_toogo_plan.
type ToogoPlanDao struct {
	table   string           // table is the underlying table name of the DAO.
	group   string           // group is the database configuration group name of current DAO.
	columns ToogoPlanColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoPlanColumns defines and stores column names for table hg_toogo_plan.
type ToogoPlanColumns struct {
	Id                 string // 主键ID
	PlanName           string // 套餐名称
	PlanCode           string // 套餐代码
	RobotLimit         string // 支持机器人数量
	PriceDaily         string // 日价格
	PriceMonthly       string // 月价格
	PriceQuarterly     string // 季价格
	PriceHalfYear      string // 半年价格
	PriceYearly        string // 年价格
	DefaultPeriod      string // 默认价格方案
	PurchaseLimit      string // 购买次数限制(0为不限)
	PurchaseLimitDaily string // 日付购买次数限制
	PurchaseLimitMonthly string // 月付购买次数限制
	PurchaseLimitQuarterly string // 季付购买次数限制
	PurchaseLimitHalfYear string // 半年付购买次数限制
	PurchaseLimitYearly string // 年付购买次数限制
	GiftPowerMonthly   string // 月订阅赠送算力
	GiftPowerQuarterly string // 季订阅赠送算力
	GiftPowerHalfYear  string // 半年订阅赠送算力
	GiftPowerYearly    string // 年订阅赠送算力
	Description        string // 套餐描述
	Features           string // 套餐特性(JSON)
	IsDefault          string // 是否默认套餐
	Sort               string // 排序
	Status             string // 状态
	CreatedAt          string // 创建时间
	UpdatedAt          string // 更新时间
}

// toogoPlanColumns holds the columns for table hg_toogo_plan.
var toogoPlanColumns = ToogoPlanColumns{
	Id:                 "id",
	PlanName:           "plan_name",
	PlanCode:           "plan_code",
	RobotLimit:         "robot_limit",
	PriceDaily:         "price_daily",
	PriceMonthly:       "price_monthly",
	PriceQuarterly:     "price_quarterly",
	PriceHalfYear:      "price_half_year",
	PriceYearly:        "price_yearly",
	DefaultPeriod:      "default_period",
	PurchaseLimit:      "purchase_limit",
	PurchaseLimitDaily: "purchase_limit_daily",
	PurchaseLimitMonthly: "purchase_limit_monthly",
	PurchaseLimitQuarterly: "purchase_limit_quarterly",
	PurchaseLimitHalfYear: "purchase_limit_half_year",
	PurchaseLimitYearly: "purchase_limit_yearly",
	GiftPowerMonthly:   "gift_power_monthly",
	GiftPowerQuarterly: "gift_power_quarterly",
	GiftPowerHalfYear:  "gift_power_half_year",
	GiftPowerYearly:    "gift_power_yearly",
	Description:        "description",
	Features:           "features",
	IsDefault:          "is_default",
	Sort:               "sort",
	Status:             "status",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewToogoPlanDao creates and returns a new DAO object for table data access.
func NewToogoPlanDao() *ToogoPlanDao {
	return &ToogoPlanDao{
		group:   "default",
		table:   "hg_toogo_plan",
		columns: toogoPlanColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoPlanDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoPlanDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoPlanDao) Columns() ToogoPlanColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoPlanDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoPlanDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoPlanDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

