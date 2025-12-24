// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingTradeFillDao is the data access object for the table hg_trading_trade_fill.
type TradingTradeFillDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  TradingTradeFillColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// TradingTradeFillColumns defines and stores column names for the table hg_trading_trade_fill.
type TradingTradeFillColumns struct {
	Id            string // 主键ID
	TenantId      string // 租户ID
	ApiConfigId   string // API配置ID
	Exchange      string // 交易所
	UserId        string // 用户ID
	RobotId       string // 机器人ID
	SessionId     string // 运行区间ID(可选)
	Symbol        string // 交易对
	OrderId       string // 交易所订单ID
	ClientOrderId string // 客户端订单ID(可选)
	TradeId       string // 成交ID
	Side          string // 方向
	Qty           string // 成交数量
	Price         string // 成交价格
	Fee           string // 手续费
	FeeCoin       string // 手续费币种
	RealizedPnl   string // 已实现盈亏
	Ts            string // 成交时间戳(毫秒)
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
}

var tradingTradeFillColumns = TradingTradeFillColumns{
	Id:            "id",
	TenantId:      "tenant_id",
	ApiConfigId:   "api_config_id",
	Exchange:      "exchange",
	UserId:        "user_id",
	RobotId:       "robot_id",
	SessionId:     "session_id",
	Symbol:        "symbol",
	OrderId:       "order_id",
	ClientOrderId: "client_order_id",
	TradeId:       "trade_id",
	Side:          "side",
	Qty:           "qty",
	Price:         "price",
	Fee:           "fee",
	FeeCoin:       "fee_coin",
	RealizedPnl:   "realized_pnl",
	Ts:            "ts",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewTradingTradeFillDao creates and returns a new DAO object for table data access.
func NewTradingTradeFillDao(handlers ...gdb.ModelHandler) *TradingTradeFillDao {
	return &TradingTradeFillDao{
		group:    "default",
		table:    "hg_trading_trade_fill",
		columns:  tradingTradeFillColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingTradeFillDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingTradeFillDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingTradeFillDao) Columns() TradingTradeFillColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingTradeFillDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingTradeFillDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TradingTradeFillDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
