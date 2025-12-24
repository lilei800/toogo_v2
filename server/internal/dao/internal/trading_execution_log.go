// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingExecutionLogDao is the data access object for the table hg_trading_execution_log.
type TradingExecutionLogDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  TradingExecutionLogColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// TradingExecutionLogColumns defines and stores column names for the table hg_trading_execution_log.
type TradingExecutionLogColumns struct {
	Id          string // 主键ID
	SignalLogId string // 关联的预警日志ID（可选）
	RobotId     string // 机器人ID
	OrderId     string // 关联的订单ID（可选）
	EventType   string // 事件类型：signal_detected/order_attempt/order_success/order_failed/position_monitor/position_close/stop_loss/take_profit
	EventData   string // 事件数据（JSON格式，包含详细信息）
	Status      string // 状态：pending/success/failed
	Message     string // 消息（详细说明，TEXT类型）
	CreatedAt   string // 创建时间
}

// tradingExecutionLogColumns holds the columns for the table hg_trading_execution_log.
var tradingExecutionLogColumns = TradingExecutionLogColumns{
	Id:          "id",
	SignalLogId: "signal_log_id",
	RobotId:     "robot_id",
	OrderId:     "order_id",
	EventType:   "event_type",
	EventData:   "event_data",
	Status:      "status",
	Message:     "message",
	CreatedAt:   "created_at",
}

// NewTradingExecutionLogDao creates and returns a new DAO object for table data access.
func NewTradingExecutionLogDao(handlers ...gdb.ModelHandler) *TradingExecutionLogDao {
	return &TradingExecutionLogDao{
		group:    "default",
		table:    "hg_trading_execution_log",
		columns:  tradingExecutionLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingExecutionLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingExecutionLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingExecutionLogDao) Columns() TradingExecutionLogColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingExecutionLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingExecutionLogDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TradingExecutionLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

