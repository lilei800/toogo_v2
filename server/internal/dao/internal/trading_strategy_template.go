// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingStrategyTemplateDao is the data access object for the table hg_trading_strategy_template.
type TradingStrategyTemplateDao struct {
	table    string                            // table is the underlying table name of the DAO.
	group    string                            // group is the database configuration group name of the current DAO.
	columns  TradingStrategyTemplateColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                // handlers for customized model modification.
}

// TradingStrategyTemplateColumns defines and stores column names for the table hg_trading_strategy_template.
type TradingStrategyTemplateColumns struct {
	Id                      string // 主键ID
	StrategyKey             string // 策略KEY：conservative_trend
	StrategyName            string // 策略名称
	RiskPreference          string // 风险偏好：conservative/balanced/aggressive
	MarketState             string // 市场状态：trend/volatile/high-volatility/low-volatility
	MonitorWindow           string // 监控时间窗口(秒)
	VolatilityThreshold     string // 波动阈值(USDT)
	LeverageMin             string // 杠杆倍数最小值
	LeverageMax             string // 杠杆倍数最大值
	MarginPercentMin        string // 保证金比例最小值(%)
	MarginPercentMax        string // 保证金比例最大值(%)
	StopLossPercent         string // 止损百分比(%)
	ProfitRetreatPercent    string // 止盈回撤百分比(%)
	AutoStartRetreatPercent string // 启动回撤百分比(%)
	ConfigJson              string // 其他配置(JSON)
	Description             string // 策略描述
	IsActive                string // 是否激活
	Sort                    string // 排序
	CreatedAt               string // 创建时间
	UpdatedAt               string // 更新时间
}

// tradingStrategyTemplateColumns holds the columns for the table hg_trading_strategy_template.
var tradingStrategyTemplateColumns = TradingStrategyTemplateColumns{
	Id:                      "id",
	StrategyKey:             "strategy_key",
	StrategyName:            "strategy_name",
	RiskPreference:          "risk_preference",
	MarketState:             "market_state",
	MonitorWindow:           "monitor_window",
	VolatilityThreshold:     "volatility_threshold",
	LeverageMin:             "leverage_min",
	LeverageMax:             "leverage_max",
	MarginPercentMin:        "margin_percent_min",
	MarginPercentMax:        "margin_percent_max",
	StopLossPercent:         "stop_loss_percent",
	ProfitRetreatPercent:    "profit_retreat_percent",
	AutoStartRetreatPercent: "auto_start_retreat_percent",
	ConfigJson:              "config_json",
	Description:             "description",
	IsActive:                "is_active",
	Sort:                    "sort",
	CreatedAt:               "created_at",
	UpdatedAt:               "updated_at",
}

// NewTradingStrategyTemplateDao creates and returns a new DAO object for table data access.
func NewTradingStrategyTemplateDao(handlers ...gdb.ModelHandler) *TradingStrategyTemplateDao {
	return &TradingStrategyTemplateDao{
		group:    "default",
		table:    "hg_trading_strategy_template",
		columns:  tradingStrategyTemplateColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingStrategyTemplateDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingStrategyTemplateDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingStrategyTemplateDao) Columns() TradingStrategyTemplateColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingStrategyTemplateDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingStrategyTemplateDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TradingStrategyTemplateDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

