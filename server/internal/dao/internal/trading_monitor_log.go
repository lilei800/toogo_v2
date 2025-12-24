// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingMonitorLogDao is the data access object for the table hg_trading_monitor_log.
type TradingMonitorLogDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  TradingMonitorLogColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// TradingMonitorLogColumns defines and stores column names for the table hg_trading_monitor_log.
type TradingMonitorLogColumns struct {
	Id             string // 主键ID
	TenantId       string // 租户ID
	UserId         string // 用户ID
	RobotId        string // 机器人ID
	Symbol         string // 交易对
	CurrentPrice   string // 当前价格
	WindowHigh     string // 窗口最高价
	WindowLow      string // 窗口最低价
	Volatility     string // 波动值
	SignalType     string // 信号类型：buy/sell/hold
	SignalStrength string // 信号强度(0-100)
	MarketState    string // 市场状态
	SignalDetail   string // 信号详情(JSON)
	CreatedAt      string // 创建时间
}

// tradingMonitorLogColumns holds the columns for the table hg_trading_monitor_log.
var tradingMonitorLogColumns = TradingMonitorLogColumns{
	Id:             "id",
	TenantId:       "tenant_id",
	UserId:         "user_id",
	RobotId:        "robot_id",
	Symbol:         "symbol",
	CurrentPrice:   "current_price",
	WindowHigh:     "window_high",
	WindowLow:      "window_low",
	Volatility:     "volatility",
	SignalType:     "signal_type",
	SignalStrength: "signal_strength",
	MarketState:    "market_state",
	SignalDetail:   "signal_detail",
	CreatedAt:      "created_at",
}

// NewTradingMonitorLogDao creates and returns a new DAO object for table data access.
func NewTradingMonitorLogDao(handlers ...gdb.ModelHandler) *TradingMonitorLogDao {
	return &TradingMonitorLogDao{
		group:    "default",
		table:    "hg_trading_monitor_log",
		columns:  tradingMonitorLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingMonitorLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingMonitorLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingMonitorLogDao) Columns() TradingMonitorLogColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingMonitorLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingMonitorLogDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TradingMonitorLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

