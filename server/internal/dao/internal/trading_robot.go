// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingRobotDao is the data access object for the table hg_trading_robot.
type TradingRobotDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  TradingRobotColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// TradingRobotColumns defines and stores column names for the table hg_trading_robot.
type TradingRobotColumns struct {
	Id                      string // 主键ID
	TenantId                string // 租户ID
	UserId                  string // 用户ID
	RobotName               string // 机器人名称
	ApiConfigId             string // API接口ID
	MaxProfitTarget         string // 最大盈利目标(USDT)
	MaxLossAmount           string // 最大亏损额(USDT)
	MaxRuntime              string // 最大运行时长(秒)
	RiskPreference          string // 风险偏好：conservative/balanced/aggressive
	AutoRiskPreference      string // 自动风险偏好：0=手动,1=自动
	MarketState             string // 市场状态：trend/volatile/high-volatility/low-volatility
	AutoMarketState         string // 自动市场状态：0=手动,1=自动
	Exchange                string // 交易所
	Symbol                  string // 交易对
	OrderType               string // 订单类型：market/limit
	MarginMode              string // 保证金模式：isolated/cross
	Leverage                string // 杠杆倍数
	MarginPercent           string // 使用保证金比例(%)
	UseMonitorSignal        string // 采用方向预警信号：0=否,1=是
	StopLossPercent         string // 止损百分比(%)
	ProfitRetreatPercent    string // 止盈回撤百分比(%)
	AutoStartRetreatPercent string // 启动回撤百分比(%)
	CurrentStrategy         string // 当前策略配置(JSON)
	Status                  string // 状态：1=未启动,2=运行中,3=暂停,4=停用
	StartTime               string // 启动时间
	PauseTime               string // 暂停时间
	StopTime                string // 停止时间
	LongCount               string // 多单数
	ShortCount              string // 空单数
	TotalProfit             string // 总盈亏(USDT)
	RuntimeSeconds          string // 已运行时长(秒)
	AutoTradeEnabled        string // 全自动下单：0=否,1=是
	AutoCloseEnabled        string // 全自动平仓：0=否,1=是
	ProfitLockEnabled       string // 锁定盈利开关：0=关闭,1=开启（止盈启动后禁止自动开新仓）
	DualSidePosition        string // 双向开单：0=单向,1=双向
	Remark                  string // 备注
	CreatedAt               string // 创建时间
	UpdatedAt               string // 更新时间
	DeletedAt               string // 删除时间
}

// tradingRobotColumns holds the columns for the table hg_trading_robot.
var tradingRobotColumns = TradingRobotColumns{
	Id:                      "id",
	TenantId:                "tenant_id",
	UserId:                  "user_id",
	RobotName:               "robot_name",
	ApiConfigId:             "api_config_id",
	MaxProfitTarget:         "max_profit_target",
	MaxLossAmount:           "max_loss_amount",
	MaxRuntime:              "max_runtime",
	RiskPreference:          "risk_preference",
	AutoRiskPreference:      "auto_risk_preference",
	MarketState:             "market_state",
	AutoMarketState:         "auto_market_state",
	Exchange:                "exchange",
	Symbol:                  "symbol",
	OrderType:               "order_type",
	MarginMode:              "margin_mode",
	Leverage:                "leverage",
	MarginPercent:           "margin_percent",
	UseMonitorSignal:        "use_monitor_signal",
	StopLossPercent:         "stop_loss_percent",
	ProfitRetreatPercent:    "profit_retreat_percent",
	AutoStartRetreatPercent: "auto_start_retreat_percent",
	CurrentStrategy:         "current_strategy",
	Status:                  "status",
	StartTime:               "start_time",
	PauseTime:               "pause_time",
	StopTime:                "stop_time",
	LongCount:               "long_count",
	ShortCount:              "short_count",
	TotalProfit:             "total_profit",
	RuntimeSeconds:          "runtime_seconds",
	AutoTradeEnabled:        "auto_trade_enabled",
	AutoCloseEnabled:        "auto_close_enabled",
	ProfitLockEnabled:       "profit_lock_enabled",
	DualSidePosition:        "dual_side_position",
	Remark:                  "remark",
	CreatedAt:               "created_at",
	UpdatedAt:               "updated_at",
	DeletedAt:               "deleted_at",
}

// NewTradingRobotDao creates and returns a new DAO object for table data access.
func NewTradingRobotDao(handlers ...gdb.ModelHandler) *TradingRobotDao {
	return &TradingRobotDao{
		group:    "default",
		table:    "hg_trading_robot",
		columns:  tradingRobotColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingRobotDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingRobotDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingRobotDao) Columns() TradingRobotColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingRobotDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingRobotDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TradingRobotDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
