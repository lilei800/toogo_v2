// =================================================================================
// This file is auto-generated for the table hg_trading_order_status_history.
// =================================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingOrderStatusHistoryDao is the data access object for table hg_trading_order_status_history.
type TradingOrderStatusHistoryDao struct {
	table   string
	group   string
	columns TradingOrderStatusHistoryColumns
}

// TradingOrderStatusHistoryColumns defines and stores column names for table hg_trading_order_status_history.
type TradingOrderStatusHistoryColumns struct {
	Id                      string // 主键ID
	OrderId                string // 订单ID
	OrderSn                string // 订单号
	ExchangeOrderId        string // 交易所订单ID
	Status                 string // 订单状态
	StatusText             string // 状态文本
	NodeType               string // 节点类型
	NodeDescription        string // 节点描述
	Quantity               string // 数量
	Price                  string // 委托价格
	AvgPrice               string // 成交均价
	OpenPrice              string // 开仓价格
	FilledQty              string // 已成交数量
	Leverage               string // 杠杆倍数
	Margin                 string // 保证金
	OpenMargin             string // 开仓保证金
	MarkPrice              string // 标记价格
	UnrealizedProfit       string // 未实现盈亏
	HighestProfit          string // 最高盈利
	Fee                    string // 手续费
	FeeCoin                string // 手续费币种
	ClosePrice             string // 平仓价格
	RealizedProfit         string // 已实现盈亏
	CloseReason            string // 平仓原因
	MarketState            string // 市场状态
	RiskLevel              string // 风险偏好
	StopLossPercent        string // 止损百分比
	AutoStartRetreatPercent string // 启动止盈百分比
	ProfitRetreatPercent   string // 止盈回撤百分比
	NodeTime               string // 节点时间
	CreatedAt              string // 创建时间
}

var (
	// TradingOrderStatusHistory is globally public accessible object for table hg_trading_order_status_history operations.
	TradingOrderStatusHistory = TradingOrderStatusHistoryDao{
		group:   "default",
		table:   "hg_trading_order_status_history",
		columns: TradingOrderStatusHistoryColumns{
			Id:                      "id",
			OrderId:                "order_id",
			OrderSn:                "order_sn",
			ExchangeOrderId:        "exchange_order_id",
			Status:                 "status",
			StatusText:             "status_text",
			NodeType:               "node_type",
			NodeDescription:        "node_description",
			Quantity:               "quantity",
			Price:                  "price",
			AvgPrice:               "avg_price",
			OpenPrice:              "open_price",
			FilledQty:              "filled_qty",
			Leverage:               "leverage",
			Margin:                 "margin",
			OpenMargin:             "open_margin",
			MarkPrice:              "mark_price",
			UnrealizedProfit:       "unrealized_profit",
			HighestProfit:          "highest_profit",
			Fee:                    "fee",
			FeeCoin:                "fee_coin",
			ClosePrice:             "close_price",
			RealizedProfit:         "realized_profit",
			CloseReason:            "close_reason",
			MarketState:            "market_state",
			RiskLevel:              "risk_level",
			StopLossPercent:        "stop_loss_percent",
			AutoStartRetreatPercent: "auto_start_retreat_percent",
			ProfitRetreatPercent:   "profit_retreat_percent",
			NodeTime:               "node_time",
			CreatedAt:              "created_at",
		},
	}
)

// Ctx is a chaining function, which creates and returns a new DB that is a shallow copy
// of current DB object and with given context in it.
// Note that this returned DB object can be used only once, so do not assign it to
// a global or package variable for long using.
func (d *TradingOrderStatusHistoryDao) Ctx(ctx context.Context) *gdb.Model {
	return g.DB(d.group).Model(d.table).Safe().Ctx(ctx)
}

// NewTradingOrderStatusHistoryDao creates and returns a new DAO object for table data access.
func NewTradingOrderStatusHistoryDao() *TradingOrderStatusHistoryDao {
	return &TradingOrderStatusHistoryDao{
		group:   "default",
		table:   "hg_trading_order_status_history",
		columns: TradingOrderStatusHistory.columns,
	}
}

