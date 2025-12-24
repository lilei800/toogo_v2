// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingOrderDao is the data access object for the table hg_trading_order.
type TradingOrderDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  TradingOrderColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// TradingOrderColumns defines and stores column names for the table hg_trading_order.
type TradingOrderColumns struct {
	Id                   string // 主键ID
	TenantId             string // 租户ID
	UserId               string // 用户ID
	RobotId              string // 机器人ID
	OrderSn              string // 订单号
	ExchangeOrderId      string // 交易所订单ID
	Symbol               string // 交易对
	Direction            string // 方向：long/short
	OpenPrice            string // 开仓价格
	ClosePrice           string // 平仓价格
	Quantity             string // 数量
	Leverage             string // 杠杆倍数
	Margin               string // 保证金(USDT)
	RealizedProfit       string // 已实现盈亏
	UnrealizedProfit     string // 未实现盈亏
	HighestProfit        string // 最高盈利
	StopLossPrice        string // 止损价格
	ProfitRetreatStarted string // 止盈回撤已启动
	ProfitRetreatPercent string // 止盈回撤百分比
	OpenTime             string // 开仓时间
	CloseTime            string // 平仓时间
	HoldDuration         string // 持仓时长(秒)
	Status               string // 状态：1=持仓中,2=已平仓,3=已取消
	CloseReason          string // 平仓原因：stop_loss/take_profit/manual/timeout
	Remark               string // 备注
	CreatedAt            string // 创建时间
	UpdatedAt            string // 更新时间
}

// tradingOrderColumns holds the columns for the table hg_trading_order.
var tradingOrderColumns = TradingOrderColumns{
	Id:                   "id",
	TenantId:             "tenant_id",
	UserId:               "user_id",
	RobotId:              "robot_id",
	OrderSn:              "order_sn",
	ExchangeOrderId:      "exchange_order_id",
	Symbol:               "symbol",
	Direction:            "direction",
	OpenPrice:            "open_price",
	ClosePrice:           "close_price",
	Quantity:             "quantity",
	Leverage:             "leverage",
	Margin:               "margin",
	RealizedProfit:       "realized_profit",
	UnrealizedProfit:     "unrealized_profit",
	HighestProfit:        "highest_profit",
	StopLossPrice:        "stop_loss_price",
	ProfitRetreatStarted: "profit_retreat_started",
	ProfitRetreatPercent: "profit_retreat_percent",
	OpenTime:             "open_time",
	CloseTime:            "close_time",
	HoldDuration:         "hold_duration",
	Status:               "status",
	CloseReason:          "close_reason",
	Remark:               "remark",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
}

// NewTradingOrderDao creates and returns a new DAO object for table data access.
func NewTradingOrderDao(handlers ...gdb.ModelHandler) *TradingOrderDao {
	return &TradingOrderDao{
		group:    "default",
		table:    "hg_trading_order",
		columns:  tradingOrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingOrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingOrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingOrderDao) Columns() TradingOrderColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingOrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingOrderDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TradingOrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

