// Package toogoin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogoin

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// StartRobotInp 启动机器人输入
type StartRobotInp struct {
	RobotId int64 `json:"robotId" v:"required" description:"机器人ID"`
}

// StopRobotInp 停止机器人输入
type StopRobotInp struct {
	RobotId int64 `json:"robotId" v:"required" description:"机器人ID"`
}

// RobotListInp 机器人列表输入
type RobotListInp struct {
	form.PageReq
	UserId int64 `json:"userId" description:"用户ID"`
	Status int   `json:"status" description:"状态"`
}

// RobotListModel 机器人列表返回
type RobotListModel struct {
	*entity.TradingRobot
	ConsumedPower float64 `json:"consumedPower" description:"已消耗算力"`
	TotalPnl      float64 `json:"totalPnl" description:"总盈亏"`
	TodayPnl      float64 `json:"todayPnl" description:"今日盈亏"`
	StatusText    string  `json:"statusText" description:"状态文本"`
}

// RobotDetailInp 机器人详情输入
type RobotDetailInp struct {
	RobotId int64 `json:"robotId" v:"required" description:"机器人ID"`
}

// RobotDetailModel 机器人详情返回
type RobotDetailModel struct {
	*entity.TradingRobot
	ConsumedPower float64 `json:"consumedPower" description:"已消耗算力"`
	TotalPnl      float64 `json:"totalPnl" description:"总盈亏"`
	MarketState   string  `json:"marketState" description:"当前市场状态"`
	RiskPref      string  `json:"riskPref" description:"当前风险偏好"`
	OpenOrders    int     `json:"openOrders" description:"持仓订单数"`
	ClosedOrders  int     `json:"closedOrders" description:"已平仓订单数"`
}

// CreateRobotInp 创建机器人输入
type CreateRobotInp struct {
	UserId             int64   `json:"userId" v:"required" description:"用户ID"`
	RobotName          string  `json:"robotName" v:"required|length:2,30" description:"机器人名称"`
	ApiConfigId        int64   `json:"apiConfigId" v:"required" description:"API配置ID"`
	TradingPair        string  `json:"tradingPair" v:"required" description:"交易对"`
	Platform           string  `json:"platform" v:"required|in:binance,bitget,okx,gate" description:"交易平台"`
	TradeType          string  `json:"tradeType" d:"perpetual" description:"交易类型: perpetual=永续合约"`
	OrderType          string  `json:"orderType" d:"market" description:"订单类型: market=市价, limit=限价"`
	MarginMode         string  `json:"marginMode" d:"isolated" description:"保证金模式: isolated=逐仓, cross=全仓"`
	MaxProfit          float64 `json:"maxProfit" description:"最大盈利目标(USDT)"`
	MaxLoss            float64 `json:"maxLoss" description:"最大亏损额度(USDT)"`
	ScheduleStart      string  `json:"scheduleStart" description:"定时开始时间"`
	ScheduleStop       string  `json:"scheduleStop" description:"定时停止时间"`
	AutoAnalyzeMarket  int     `json:"autoAnalyzeMarket" d:"1" description:"自动分析行情: 0=关闭, 1=开启"`
	AutoSignalEnabled  int     `json:"autoSignalEnabled" d:"1" description:"自动下单信号: 0=关闭, 1=开启"`
	StrategyId         int64   `json:"strategyId" description:"策略模板ID"`
}

// CreateRobotModel 创建机器人返回
type CreateRobotModel struct {
	RobotId        int64   `json:"robotId" description:"机器人ID"`
	EstimatedPower float64 `json:"estimatedPower" description:"预计消耗算力"`
}

// UpdateRobotInp 更新机器人输入
type UpdateRobotInp struct {
	RobotId            int64   `json:"robotId" v:"required" description:"机器人ID"`
	RobotName          string  `json:"robotName" description:"机器人名称"`
	MaxProfit          float64 `json:"maxProfit" description:"最大盈利目标"`
	MaxLoss            float64 `json:"maxLoss" description:"最大亏损额度"`
	ScheduleStart      string  `json:"scheduleStart" description:"定时开始时间"`
	ScheduleStop       string  `json:"scheduleStop" description:"定时停止时间"`
	AutoAnalyzeMarket  int     `json:"autoAnalyzeMarket" description:"自动分析行情"`
	AutoSignalEnabled  int     `json:"autoSignalEnabled" description:"自动下单信号"`
	StrategyId         int64   `json:"strategyId" description:"策略模板ID"`
}

// DeleteRobotInp 删除机器人输入
type DeleteRobotInp struct {
	RobotId int64 `json:"robotId" v:"required" description:"机器人ID"`
}

// RobotOrderListInp 机器人订单列表输入
type RobotOrderListInp struct {
	form.PageReq
	RobotId int64 `json:"robotId" v:"required" description:"机器人ID"`
	Status  int   `json:"status" description:"订单状态"`
}

// RobotOrderListModel 机器人订单列表返回
type RobotOrderListModel struct {
	*entity.TradingOrder
}

// ManualCloseOrderInp 手动平仓输入
type ManualCloseOrderInp struct {
	OrderId int64 `json:"orderId" v:"required" description:"订单ID"`
}

// PositionModel 持仓模型
type PositionModel struct {
	Symbol            string  `json:"symbol" description:"交易对"`
	PositionSide      string  `json:"positionSide" description:"持仓方向 LONG/SHORT"`
	PositionAmt       float64 `json:"positionAmt" description:"持仓数量"`
	EntryPrice        float64 `json:"entryPrice" description:"开仓价格"`
	MarkPrice         float64 `json:"markPrice" description:"标记价格"`
	UnrealizedPnl     float64 `json:"unrealizedPnl" description:"未实现盈亏"`
	Leverage          int     `json:"leverage" description:"杠杆倍数"`
	Margin            float64 `json:"margin" description:"保证金"`
	MarginType        string  `json:"marginType" description:"保证金模式"`
	IsolatedMargin    float64 `json:"isolatedMargin" description:"逐仓保证金"`
	LiquidationPrice  float64 `json:"liquidationPrice" description:"强平价格"`
	MaxProfitReached  float64 `json:"maxProfitReached" description:"最高盈利金额"`
	TakeProfitEnabled bool    `json:"takeProfitEnabled" description:"止盈是否已启动"`
	CreateTime        int64   `json:"createTime" description:"开仓时间(毫秒时间戳)"`
	// 订单创建时的策略参数（用于血条计算，优先使用订单创建时的参数）
	// 注意：不使用 omitempty，确保即使值为 nil 也会返回 null，前端可以检测到
	StopLossPercent         *float64 `json:"stopLossPercent" description:"止损百分比(%)"`
	AutoStartRetreatPercent *float64 `json:"autoStartRetreatPercent" description:"启动止盈百分比(%)"`
	ProfitRetreatPercent    *float64 `json:"profitRetreatPercent" description:"止盈回撤百分比(%)"`
	MarginPercent           *float64 `json:"marginPercent" description:"保证金比例(%)"`
	// 订单创建时的市场状态和风险偏好
	MarketState    string `json:"marketState" description:"市场状态（创建订单时）"`
	RiskPreference string `json:"riskPreference" description:"风险偏好（创建订单时）"`
	// 订单信息（从交易所API获取）
	OrderId      string  `json:"orderId" description:"交易所订单ID"`
	ClientOrderId string `json:"clientOrderId" description:"客户端订单ID"`
	OrderType    string  `json:"orderType" description:"订单类型"`
	OrderSide    string  `json:"orderSide" description:"买卖方向 BUY/SELL"`
	OrderQuantity float64 `json:"orderQuantity" description:"订单数量"`
	OrderAvgPrice float64 `json:"orderAvgPrice" description:"订单成交均价"`
	OrderCreateTime int64 `json:"orderCreateTime" description:"订单创建时间(毫秒时间戳)"`
}

// OrderModel 订单模型
type OrderModel struct {
	OrderId      string  `json:"orderId" description:"订单ID"`
	ClientId     string  `json:"clientId" description:"客户端订单ID"`
	Symbol       string  `json:"symbol" description:"交易对"`
	Side         string  `json:"side" description:"买卖方向 BUY/SELL"`
	PositionSide string  `json:"positionSide" description:"持仓方向 LONG/SHORT"`
	Type         string  `json:"type" description:"订单类型"`
	Price        float64 `json:"price" description:"价格"`
	Quantity     float64 `json:"quantity" description:"数量"`
	FilledQty    float64 `json:"filledQty" description:"已成交数量"`
	AvgPrice     float64 `json:"avgPrice" description:"成交均价"`
	Status       string  `json:"status" description:"订单状态"`
	CreateTime   int64   `json:"createTime" description:"创建时间"`
	UpdateTime   int64   `json:"updateTime" description:"更新时间"`
}

// ClosePositionInp 平仓输入
type ClosePositionInp struct {
	RobotId      int64   `json:"robotId" v:"required" description:"机器人ID"`
	Symbol       string  `json:"symbol" description:"交易对（可选，默认使用机器人配置的交易对）"`
	PositionSide string  `json:"positionSide" v:"required|in:LONG,SHORT" description:"持仓方向"`
	Quantity     float64 `json:"quantity" description:"平仓数量（可选，默认全部平仓）"`
}

// TickerModel 行情模型
type TickerModel struct {
	Symbol             string  `json:"symbol" description:"交易对"`
	LastPrice          float64 `json:"lastPrice" description:"最新价"`
	BidPrice           float64 `json:"bidPrice" description:"买一价"`
	AskPrice           float64 `json:"askPrice" description:"卖一价"`
	High24h            float64 `json:"high24h" description:"24小时最高"`
	Low24h             float64 `json:"low24h" description:"24小时最低"`
	Volume24h          float64 `json:"volume24h" description:"24小时成交量"`
	PriceChangePercent float64 `json:"changePercent" description:"24小时涨跌幅"`
	Timestamp          int64   `json:"timestamp" description:"时间戳"`
}

// GetTickerInp 获取行情输入
type GetTickerInp struct {
	ApiKeyId int64  `json:"apiKeyId" v:"required" description:"API密钥ID"`
	Symbol   string `json:"symbol" v:"required" description:"交易对"`
}

// GetPositionsInp 获取持仓输入
type GetPositionsInp struct {
	RobotId int64 `json:"robotId" v:"required" description:"机器人ID"`
}

// GetOrdersInp 获取订单输入
type GetOrdersInp struct {
	RobotId int64 `json:"robotId" v:"required" description:"机器人ID"`
	Limit   int   `json:"limit" d:"50" description:"数量限制"`
}
