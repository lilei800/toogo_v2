// Package exchange
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 交易所统一接口
package exchange

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Exchange 交易所接口
type Exchange interface {
	// GetName 获取交易所名称
	GetName() string
	// GetBalance 获取账户余额
	GetBalance(ctx context.Context) (*Balance, error)
	// GetTicker 获取行情
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)
	// GetKlines 获取K线数据
	GetKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error)
	// GetPositions 获取持仓
	GetPositions(ctx context.Context, symbol string) ([]*Position, error)
	// CreateOrder 创建订单
	CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error)
	// CancelOrder 取消订单
	CancelOrder(ctx context.Context, symbol, orderId string) (*Order, error)
	// ClosePosition 平仓
	ClosePosition(ctx context.Context, symbol, positionSide string, quantity float64) (*Order, error)
	// SetLeverage 设置杠杆
	SetLeverage(ctx context.Context, symbol string, leverage int) error
	// SetMarginType 设置保证金模式
	SetMarginType(ctx context.Context, symbol, marginType string) error
	// GetOpenOrders 获取当前挂单
	GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error)
	// GetOrderHistory 获取历史订单
	GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error)
}

// ExchangeAdvanced 高级交易所接口（可选实现）
type ExchangeAdvanced interface {
	Exchange

	// SetStopLoss 设置止损
	SetStopLoss(ctx context.Context, req *StopLossRequest) (*Order, error)
	// SetTakeProfit 设置止盈
	SetTakeProfit(ctx context.Context, req *TakeProfitRequest) (*Order, error)
	// SetStopLossAndTakeProfit 同时设置止损止盈
	SetStopLossAndTakeProfit(ctx context.Context, req *SLTPRequest) (*SLTPResponse, error)
	// CancelStopLoss 取消止损单
	CancelStopLoss(ctx context.Context, symbol, orderId string) error
	// CancelTakeProfit 取消止盈单
	CancelTakeProfit(ctx context.Context, symbol, orderId string) error
	// BatchClosePositions 批量平仓
	BatchClosePositions(ctx context.Context, symbols []string) ([]*CloseResult, error)
	// CloseAllPositions 一键平仓所有持仓
	CloseAllPositions(ctx context.Context) ([]*CloseResult, error)
	// GetAccountInfo 获取账户详情
	GetAccountInfo(ctx context.Context) (*AccountInfo, error)
	// GetSymbolInfo 获取交易对信息
	GetSymbolInfo(ctx context.Context, symbol string) (*SymbolInfo, error)
	// GetFundingRate 获取资金费率
	GetFundingRate(ctx context.Context, symbol string) (*FundingRate, error)
	// ModifyOrder 修改订单
	ModifyOrder(ctx context.Context, symbol, orderId string, price, quantity float64) (*Order, error)
	// GetTradeHistory 获取成交记录
	GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*Trade, error)
}

// Config 交易所配置
type Config struct {
	Platform   string       `json:"platform"`   // 平台: binance, bitget, okx, gate
	ApiKey     string       `json:"apiKey"`     // API Key
	SecretKey  string       `json:"secretKey"`  // Secret Key
	Passphrase string       `json:"passphrase"` // Passphrase (OKX/Bitget需要)
	IsTestnet  bool         `json:"isTestnet"`  // 是否测试网
	Proxy      *ProxyConfig `json:"proxy"`      // 代理配置
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled  bool   `json:"enabled"`  // 是否启用
	Type     string `json:"type"`     // 代理类型: socks5, http
	Host     string `json:"host"`     // 代理地址
	Port     int    `json:"port"`     // 代理端口
	Username string `json:"username"` // 用户名(可选)
	Password string `json:"password"` // 密码(可选)
}

// GetProxyURL 获取代理URL
func (p *ProxyConfig) GetProxyURL() string {
	if p == nil || !p.Enabled {
		return ""
	}
	if p.Username != "" && p.Password != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%d", p.Type, p.Username, p.Password, p.Host, p.Port)
	}
	return fmt.Sprintf("%s://%s:%d", p.Type, p.Host, p.Port)
}

// Balance 账户余额
type Balance struct {
	TotalBalance     float64 `json:"totalBalance"`     // 总余额
	AvailableBalance float64 `json:"availableBalance"` // 可用余额
	FrozenBalance    float64 `json:"frozenBalance"`    // 冻结余额
	UnrealizedPnl    float64 `json:"unrealizedPnl"`    // 未实现盈亏
	Currency         string  `json:"currency"`         // 币种
}

// Ticker 行情数据
type Ticker struct {
	Symbol             string  `json:"symbol"`             // 交易对
	LastPrice          float64 `json:"lastPrice"`          // 最新价
	MarkPrice          float64 `json:"markPrice"`          // 标记价格（期货/永续优先用于风控/止盈止损/浮动盈亏）
	IndexPrice         float64 `json:"indexPrice"`         // 指数价格（可选）
	BidPrice           float64 `json:"bidPrice"`           // 买一价
	AskPrice           float64 `json:"askPrice"`           // 卖一价
	High24h            float64 `json:"high24h"`            // 24h最高
	Low24h             float64 `json:"low24h"`             // 24h最低
	Volume24h          float64 `json:"volume24h"`          // 24h成交量(基础货币)
	QuoteVolume24h     float64 `json:"quoteVolume24h"`     // 24h成交额(计价货币)
	Change24h          float64 `json:"change24h"`          // 24h涨跌幅
	PriceChangePercent float64 `json:"priceChangePercent"` // 24h涨跌百分比
	Timestamp          int64   `json:"timestamp"`          // 时间戳
}

// EffectiveMarkPrice 标准化“用于风控/盈亏”的价格口径：
// - 优先使用 MarkPrice（更符合交易所强平/止盈止损的口径）
// - 若 MarkPrice 缺失，则降级为 LastPrice
func (t *Ticker) EffectiveMarkPrice() float64 {
	if t == nil {
		return 0
	}
	if t.MarkPrice > 0 {
		return t.MarkPrice
	}
	return t.LastPrice
}

// Kline K线数据
type Kline struct {
	OpenTime  int64   `json:"openTime"`  // 开盘时间
	Open      float64 `json:"open"`      // 开盘价
	High      float64 `json:"high"`      // 最高价
	Low       float64 `json:"low"`       // 最低价
	Close     float64 `json:"close"`     // 收盘价
	Volume    float64 `json:"volume"`    // 成交量
	CloseTime int64   `json:"closeTime"` // 收盘时间
}

// Position 持仓信息
type Position struct {
	Symbol           string  `json:"symbol"`           // 交易对
	PositionSide     string  `json:"positionSide"`     // 持仓方向: LONG/SHORT
	PositionAmt      float64 `json:"positionAmt"`      // 持仓数量
	EntryPrice       float64 `json:"entryPrice"`       // 开仓均价
	MarkPrice        float64 `json:"markPrice"`        // 标记价格
	UnrealizedPnl    float64 `json:"unrealizedPnl"`    // 未实现盈亏
	Leverage         int     `json:"leverage"`         // 杠杆倍数
	Margin           float64 `json:"margin"`           // 保证金 = 持仓价值 / 杠杆
	MarginType       string  `json:"marginType"`       // 保证金模式
	IsolatedMargin   float64 `json:"isolatedMargin"`   // 逐仓保证金
	LiquidationPrice float64 `json:"liquidationPrice"` // 强平价格
}

// OrderRequest 下单请求
type OrderRequest struct {
	Symbol       string  `json:"symbol"`                 // 交易对
	Side         string  `json:"side"`                   // 买卖方向: BUY/SELL
	PositionSide string  `json:"positionSide,omitempty"` // 持仓方向: LONG/SHORT
	Type         string  `json:"type"`                   // 订单类型: MARKET/LIMIT
	Quantity     float64 `json:"quantity"`               // 数量
	Price        float64 `json:"price,omitempty"`        // 价格(限价单)
	ReduceOnly   bool    `json:"reduceOnly,omitempty"`   // 只减仓
	StopPrice    float64 `json:"stopPrice,omitempty"`    // 止损价
	TakeProfit   float64 `json:"takeProfit,omitempty"`   // 止盈价
}

// Order 订单信息
type Order struct {
	OrderId      string  `json:"orderId"`      // 订单ID
	ClientId     string  `json:"clientId"`     // 客户端订单ID
	Symbol       string  `json:"symbol"`       // 交易对
	Side         string  `json:"side"`         // 买卖方向
	PositionSide string  `json:"positionSide"` // 持仓方向
	Type         string  `json:"type"`         // 订单类型
	Price        float64 `json:"price"`        // 价格
	Quantity     float64 `json:"quantity"`     // 数量
	FilledQty    float64 `json:"filledQty"`    // 已成交数量
	AvgPrice     float64 `json:"avgPrice"`     // 成交均价
	Status       string  `json:"status"`       // 状态
	TradeScope   string  `json:"tradeScope"`   // 流动性方向 (maker/taker)
	Fee          float64 `json:"fee"`          // 手续费
	FeeCoin      string  `json:"feeCoin"`      // 手续费币种
	CreateTime   int64   `json:"createTime"`   // 创建时间
	UpdateTime   int64   `json:"updateTime"`   // 更新时间
}

// SymbolInfo 交易对信息
type SymbolInfo struct {
	Symbol          string  `json:"symbol"`          // 交易对
	BaseCoin        string  `json:"baseCoin"`        // 基础币
	QuoteCoin       string  `json:"quoteCoin"`       // 计价币
	PricePrecision  int     `json:"pricePrecision"`  // 价格精度
	QtyPrecision    int     `json:"qtyPrecision"`    // 数量精度
	MinQty          float64 `json:"minQty"`          // 最小数量
	MaxLeverage     int     `json:"maxLeverage"`     // 最大杠杆
	ContractSize    float64 `json:"contractSize"`    // 合约乘数
	MinNotionalUSDT float64 `json:"minNotionalUSDT"` // 最小名义价值
}

// StopLossRequest 止损请求
type StopLossRequest struct {
	Symbol       string  `json:"symbol"`       // 交易对
	PositionSide string  `json:"positionSide"` // 持仓方向: LONG/SHORT
	StopPrice    float64 `json:"stopPrice"`    // 止损触发价
	Quantity     float64 `json:"quantity"`     // 数量(0表示全部)
	OrderType    string  `json:"orderType"`    // STOP_MARKET/STOP_LIMIT
	Price        float64 `json:"price"`        // 限价价格(STOP_LIMIT时使用)
}

// TakeProfitRequest 止盈请求
type TakeProfitRequest struct {
	Symbol       string  `json:"symbol"`       // 交易对
	PositionSide string  `json:"positionSide"` // 持仓方向: LONG/SHORT
	TakePrice    float64 `json:"takePrice"`    // 止盈触发价
	Quantity     float64 `json:"quantity"`     // 数量(0表示全部)
	OrderType    string  `json:"orderType"`    // TAKE_PROFIT_MARKET/TAKE_PROFIT_LIMIT
	Price        float64 `json:"price"`        // 限价价格(LIMIT时使用)
}

// SLTPRequest 止损止盈请求
type SLTPRequest struct {
	Symbol          string  `json:"symbol"`          // 交易对
	PositionSide    string  `json:"positionSide"`    // 持仓方向
	StopLossPrice   float64 `json:"stopLossPrice"`   // 止损价
	TakeProfitPrice float64 `json:"takeProfitPrice"` // 止盈价
	Quantity        float64 `json:"quantity"`        // 数量
}

// SLTPResponse 止损止盈响应
type SLTPResponse struct {
	StopLossOrder   *Order `json:"stopLossOrder"`   // 止损订单
	TakeProfitOrder *Order `json:"takeProfitOrder"` // 止盈订单
}

// CloseResult 平仓结果
type CloseResult struct {
	Symbol       string  `json:"symbol"`       // 交易对
	PositionSide string  `json:"positionSide"` // 持仓方向
	Quantity     float64 `json:"quantity"`     // 平仓数量
	Price        float64 `json:"price"`        // 成交价格
	RealizedPnl  float64 `json:"realizedPnl"`  // 实现盈亏
	Success      bool    `json:"success"`      // 是否成功
	Error        string  `json:"error"`        // 错误信息
	Order        *Order  `json:"order"`        // 订单信息
}

// AccountInfo 账户信息
type AccountInfo struct {
	TotalWalletBalance    float64         `json:"totalWalletBalance"`    // 钱包总余额
	TotalUnrealizedProfit float64         `json:"totalUnrealizedProfit"` // 总未实现盈亏
	TotalMarginBalance    float64         `json:"totalMarginBalance"`    // 保证金总额
	AvailableBalance      float64         `json:"availableBalance"`      // 可用余额
	MaxWithdrawAmount     float64         `json:"maxWithdrawAmount"`     // 最大可提取金额
	FeeTier               int             `json:"feeTier"`               // 手续费等级
	CanTrade              bool            `json:"canTrade"`              // 是否可交易
	CanDeposit            bool            `json:"canDeposit"`            // 是否可充值
	CanWithdraw           bool            `json:"canWithdraw"`           // 是否可提现
	Positions             []*Position     `json:"positions"`             // 持仓列表
	Assets                []*AssetBalance `json:"assets"`                // 资产列表
}

// AssetBalance 资产余额
type AssetBalance struct {
	Asset              string  `json:"asset"`              // 资产名称
	WalletBalance      float64 `json:"walletBalance"`      // 钱包余额
	UnrealizedProfit   float64 `json:"unrealizedProfit"`   // 未实现盈亏
	MarginBalance      float64 `json:"marginBalance"`      // 保证金余额
	AvailableBalance   float64 `json:"availableBalance"`   // 可用余额
	CrossWalletBalance float64 `json:"crossWalletBalance"` // 全仓钱包余额
	CrossUnPnl         float64 `json:"crossUnPnl"`         // 全仓未实现盈亏
	MaxWithdrawAmount  float64 `json:"maxWithdrawAmount"`  // 最大可提取
}

// FundingRate 资金费率
type FundingRate struct {
	Symbol          string  `json:"symbol"`          // 交易对
	FundingRate     float64 `json:"fundingRate"`     // 当前资金费率
	FundingTime     int64   `json:"fundingTime"`     // 资金费率结算时间
	NextFundingTime int64   `json:"nextFundingTime"` // 下次结算时间
	MarkPrice       float64 `json:"markPrice"`       // 标记价格
	IndexPrice      float64 `json:"indexPrice"`      // 指数价格
}

// Trade 成交记录
type Trade struct {
	TradeId         string  `json:"tradeId"`         // 成交ID
	OrderId         string  `json:"orderId"`         // 订单ID
	Symbol          string  `json:"symbol"`          // 交易对
	Side            string  `json:"side"`            // 买卖方向
	PositionSide    string  `json:"positionSide"`    // 持仓方向
	Price           float64 `json:"price"`           // 成交价格
	Quantity        float64 `json:"quantity"`        // 成交数量
	RealizedPnl     float64 `json:"realizedPnl"`     // 实现盈亏
	Commission      float64 `json:"commission"`      // 手续费
	CommissionAsset string  `json:"commissionAsset"` // 手续费币种
	Time            int64   `json:"time"`            // 成交时间
}

// NewExchange 创建交易所实例
func NewExchange(config *Config) (Exchange, error) {
	switch config.Platform {
	case "binance":
		return NewBinance(config), nil
	case "bitget":
		return NewBitget(config), nil
	case "okx":
		return NewOKX(config), nil
	case "gate":
		return NewGate(config), nil
	default:
		return nil, gerror.Newf("不支持的交易所: %s", config.Platform)
	}
}

// FormatSymbol 格式化交易对
func FormatSymbol(platform, symbol string) string {
	// 统一格式: BTC/USDT -> 各平台格式
	switch platform {
	case "binance":
		// BTCUSDT
		return symbol[:len(symbol)-5] + symbol[len(symbol)-4:]
	case "okx":
		// BTC-USDT-SWAP
		return symbol[:len(symbol)-5] + "-" + symbol[len(symbol)-4:] + "-SWAP"
	default:
		return symbol
	}
}

// CalculateStopLossPrice 计算止损价格
// entryPrice: 入场价格
// stopLossPercent: 止损百分比 (如 5 表示 5%)
// positionSide: 持仓方向 LONG/SHORT
func CalculateStopLossPrice(entryPrice float64, stopLossPercent float64, positionSide string) float64 {
	ratio := stopLossPercent / 100
	if positionSide == "LONG" {
		return entryPrice * (1 - ratio)
	}
	return entryPrice * (1 + ratio)
}

// CalculateTakeProfitPrice 计算止盈价格
// entryPrice: 入场价格
// takeProfitPercent: 止盈百分比 (如 10 表示 10%)
// positionSide: 持仓方向 LONG/SHORT
func CalculateTakeProfitPrice(entryPrice float64, takeProfitPercent float64, positionSide string) float64 {
	ratio := takeProfitPercent / 100
	if positionSide == "LONG" {
		return entryPrice * (1 + ratio)
	}
	return entryPrice * (1 - ratio)
}

// CalculatePnLPercent 计算盈亏百分比
func CalculatePnLPercent(entryPrice, currentPrice float64, positionSide string, leverage int) float64 {
	var pnlPercent float64
	if positionSide == "LONG" {
		pnlPercent = (currentPrice - entryPrice) / entryPrice * 100
	} else {
		pnlPercent = (entryPrice - currentPrice) / entryPrice * 100
	}
	return pnlPercent * float64(leverage)
}

// CalculateLiquidationPrice 估算强平价格
// entryPrice: 入场价格
// leverage: 杠杆倍数
// positionSide: 持仓方向
// marginType: 保证金模式 isolated/cross
func CalculateLiquidationPrice(entryPrice float64, leverage int, positionSide, marginType string) float64 {
	// 简化计算，实际强平价格需要考虑更多因素
	maintenanceMarginRate := 0.004 // 维持保证金率 0.4%
	if marginType == "cross" {
		maintenanceMarginRate = 0.005
	}

	if positionSide == "LONG" {
		return entryPrice * (1 - 1/float64(leverage) + maintenanceMarginRate)
	}
	return entryPrice * (1 + 1/float64(leverage) - maintenanceMarginRate)
}

// CalculatePositionValue 计算持仓价值
func CalculatePositionValue(quantity, price float64) float64 {
	return quantity * price
}

// CalculateRequiredMargin 计算所需保证金
func CalculateRequiredMargin(positionValue float64, leverage int) float64 {
	return positionValue / float64(leverage)
}

// ValidateOrderRequest 验证下单请求
func ValidateOrderRequest(req *OrderRequest) error {
	if req.Symbol == "" {
		return gerror.New("交易对不能为空")
	}
	if req.Side != "BUY" && req.Side != "SELL" {
		return gerror.Newf("无效的交易方向: %s", req.Side)
	}
	if req.Type != "MARKET" && req.Type != "LIMIT" {
		return gerror.Newf("无效的订单类型: %s", req.Type)
	}
	if req.Quantity <= 0 {
		return gerror.New("数量必须大于0")
	}
	if req.Type == "LIMIT" && req.Price <= 0 {
		return gerror.New("限价单必须指定价格")
	}
	return nil
}

// OrderStatus 订单状态常量
const (
	OrderStatusNew             = "NEW"              // 新建
	OrderStatusPartiallyFilled = "PARTIALLY_FILLED" // 部分成交
	OrderStatusFilled          = "FILLED"           // 完全成交
	OrderStatusCanceled        = "CANCELED"         // 已取消
	OrderStatusRejected        = "REJECTED"         // 已拒绝
	OrderStatusExpired         = "EXPIRED"          // 已过期
)

// PositionSide 持仓方向常量
const (
	PositionSideLong  = "LONG"
	PositionSideShort = "SHORT"
	PositionSideBoth  = "BOTH"
)

// MarginType 保证金类型常量
const (
	MarginTypeIsolated = "ISOLATED"
	MarginTypeCrossed  = "CROSSED"
)

// OrderType 订单类型常量
const (
	OrderTypeMarket           = "MARKET"
	OrderTypeLimit            = "LIMIT"
	OrderTypeStopMarket       = "STOP_MARKET"
	OrderTypeStopLimit        = "STOP"
	OrderTypeTakeProfitMarket = "TAKE_PROFIT_MARKET"
	OrderTypeTakeProfitLimit  = "TAKE_PROFIT"
)
