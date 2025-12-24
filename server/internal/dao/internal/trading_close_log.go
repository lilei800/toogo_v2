// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingCloseLogDao is the data access object for the table hg_trading_close_log.
type TradingCloseLogDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  TradingCloseLogColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// TradingCloseLogColumns defines and stores column names for the table hg_trading_close_log.
type TradingCloseLogColumns struct {
	Id                string // 主键ID
	TenantId          string // 租户ID
	UserId            string // 用户ID
	RobotId           string // 机器人ID
	OrderId           string // 订单ID
	OrderSn           string // 订单号
	Symbol            string // 交易对
	Direction         string // 方向：long/short
	OpenPrice         string // 开仓价格
	ClosePrice        string // 平仓价格
	Quantity          string // 数量
	Leverage          string // 杠杆倍数
	Margin            string // 保证金(USDT)
	RealizedProfit    string // 已实现盈亏
	HighestProfit     string // 最高盈利
	ProfitPercent     string // 盈利百分比
	CloseReason       string // 平仓原因
	CloseDetail       string // 平仓详情(JSON)
	OpenFee           string // 开仓费用
	HoldFee           string // 持仓费用
	CloseFee          string // 平仓费用
	TotalFee          string // 总费用
	CommissionAmount  string // 佣金金额
	CommissionPercent string // 佣金比例
	NetProfit         string // 净利润
	OpenTime          string // 开仓时间
	CloseTime         string // 平仓时间
	HoldDuration      string // 持仓时长(秒)
	CreatedAt         string // 创建时间
}

// tradingCloseLogColumns holds the columns for the table hg_trading_close_log.
var tradingCloseLogColumns = TradingCloseLogColumns{
	Id:                "id",
	TenantId:          "tenant_id",
	UserId:            "user_id",
	RobotId:           "robot_id",
	OrderId:           "order_id",
	OrderSn:           "order_sn",
	Symbol:            "symbol",
	Direction:         "direction",
	OpenPrice:         "open_price",
	ClosePrice:        "close_price",
	Quantity:          "quantity",
	Leverage:          "leverage",
	Margin:            "margin",
	RealizedProfit:    "realized_profit",
	HighestProfit:     "highest_profit",
	ProfitPercent:     "profit_percent",
	CloseReason:       "close_reason",
	CloseDetail:       "close_detail",
	OpenFee:           "open_fee",
	HoldFee:           "hold_fee",
	CloseFee:          "close_fee",
	TotalFee:          "total_fee",
	CommissionAmount:  "commission_amount",
	CommissionPercent: "commission_percent",
	NetProfit:         "net_profit",
	OpenTime:          "open_time",
	CloseTime:         "close_time",
	HoldDuration:      "hold_duration",
	CreatedAt:         "created_at",
}

// NewTradingCloseLogDao creates and returns a new DAO object for table data access.
func NewTradingCloseLogDao(handlers ...gdb.ModelHandler) *TradingCloseLogDao {
	return &TradingCloseLogDao{
		group:    "default",
		table:    "hg_trading_close_log",
		columns:  tradingCloseLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingCloseLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingCloseLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingCloseLogDao) Columns() TradingCloseLogColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingCloseLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingCloseLogDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TradingCloseLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

