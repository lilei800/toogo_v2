// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoStrategyTemplateDao is the data access object for table hg_toogo_strategy_template.
type ToogoStrategyTemplateDao struct {
	table   string                       // table is the underlying table name of the DAO.
	group   string                       // group is the database configuration group name of current DAO.
	columns ToogoStrategyTemplateColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoStrategyTemplateColumns defines and stores column names for table hg_toogo_strategy_template.
type ToogoStrategyTemplateColumns struct {
	Id                   string // 主键ID
	StrategyKey          string // 策略KEY
	StrategyName         string // 策略名称
	RiskPreference       string // 风险偏好
	MarketState          string // 市场状态
	TimeWindow           string // 时间窗口
	VolatilityPoints     string // 波动点数
	LeverageMin          string // 杠杆倍数最小值
	LeverageMax          string // 杠杆倍数最大值
	MarginPercentMin     string // 保证金比例最小值
	MarginPercentMax     string // 保证金比例最大值
	StopLossPercent      string // 止损百分比
	ProfitRetreatPercent string // 止盈回撤百分比
	StartRetreatPercent  string // 启动回撤百分比
	ReverseLossRetreat   string // 反向-亏损订单回撤百分比
	ReverseProfitRetreat string // 反向-盈利订单回撤百分比
	VolatilityConfig     string // 多周期波动率配置
	Description          string // 策略描述
	IsOfficial           string // 是否官方推荐
	IsActive             string // 是否激活
	Sort                 string // 排序
	CreatedAt            string // 创建时间
	UpdatedAt            string // 更新时间
}

// toogoStrategyTemplateColumns holds the columns for table hg_toogo_strategy_template.
var toogoStrategyTemplateColumns = ToogoStrategyTemplateColumns{
	Id:                   "id",
	StrategyKey:          "strategy_key",
	StrategyName:         "strategy_name",
	RiskPreference:       "risk_preference",
	MarketState:          "market_state",
	TimeWindow:           "time_window",
	VolatilityPoints:     "volatility_points",
	LeverageMin:          "leverage_min",
	LeverageMax:          "leverage_max",
	MarginPercentMin:     "margin_percent_min",
	MarginPercentMax:     "margin_percent_max",
	StopLossPercent:      "stop_loss_percent",
	ProfitRetreatPercent: "profit_retreat_percent",
	StartRetreatPercent:  "start_retreat_percent",
	ReverseLossRetreat:   "reverse_loss_retreat",
	ReverseProfitRetreat: "reverse_profit_retreat",
	VolatilityConfig:     "volatility_config",
	Description:          "description",
	IsOfficial:           "is_official",
	IsActive:             "is_active",
	Sort:                 "sort",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
}

// NewToogoStrategyTemplateDao creates and returns a new DAO object for table data access.
func NewToogoStrategyTemplateDao() *ToogoStrategyTemplateDao {
	return &ToogoStrategyTemplateDao{
		group:   "default",
		table:   "hg_toogo_strategy_template",
		columns: toogoStrategyTemplateColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoStrategyTemplateDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoStrategyTemplateDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoStrategyTemplateDao) Columns() ToogoStrategyTemplateColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoStrategyTemplateDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoStrategyTemplateDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoStrategyTemplateDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

