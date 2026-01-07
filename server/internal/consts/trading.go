// Package consts
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 交易相关常量
package consts

// 默认交易配置
const (
	// 默认杠杆
	DefaultLeverage = 10
	// 最大杠杆
	MaxLeverage = 125
	// 默认保证金模式
	DefaultMarginType = "ISOLATED"
	// 默认订单类型
	DefaultOrderType = "MARKET"
)

// 风控配置
const (
	// 单笔最大金额
	MaxOrderAmount = 10000.0
	// 单日最大交易次数
	MaxDailyTrades = 1000
	// 最大持仓数量
	MaxPositions = 50
	// 默认止损百分比
	DefaultStopLossPercent = 5.0
	// 默认止盈百分比
	DefaultTakeProfitPercent = 10.0
	// 强制止损百分比
	ForceStopLossPercent = 50.0
	// 最小开仓金额
	MinOrderAmount = 5.0
	// 冷却时间(秒)
	CooldownSeconds = 10
)

// 机器人配置
const (
	// 引擎执行间隔(秒)
	RobotEngineInterval = 10
	// 行情检查间隔(秒)
	TickerCheckInterval = 5
	// 持仓检查间隔(秒)
	PositionCheckInterval = 30
	// 订单超时时间(秒)
	OrderTimeout = 60
	// 最大重试次数
	MaxRetries = 3
	// 重试间隔(秒)
	RetryInterval = 5
)

// 算力配置
const (
	// 算力消耗比例
	PowerConsumePercent = 10.0
	// 最低运行算力
	MinPowerRequired = 10.0
	// 每次消耗最低算力
	MinPowerConsume = 0.1
	// 新用户赠送积分
	FreePowerGift = 30.0
)

// API配置
const (
	// API请求超时(秒)
	ApiTimeout = 30
	// API请求重试次数
	ApiRetries = 3
	// API请求间隔(毫秒)
	ApiInterval = 100
	// API每分钟限制
	ApiRateLimit = 1200
)

// 策略限制
const (
	// 每组最大策略数
	MaxStrategiesPerGroup = 20
	// 每用户最大策略组数
	MaxStrategyGroupsPerUser = 10
)

// 交易方向
const (
	SideBuy  = "BUY"
	SideSell = "SELL"
)

// 持仓方向
const (
	PositionLong  = "LONG"
	PositionShort = "SHORT"
	PositionBoth  = "BOTH"
)

// 订单类型
const (
	OrderMarket           = "MARKET"
	OrderLimit            = "LIMIT"
	OrderStopMarket       = "STOP_MARKET"
	OrderStopLimit        = "STOP"
	OrderTakeProfitMarket = "TAKE_PROFIT_MARKET"
	OrderTakeProfitLimit  = "TAKE_PROFIT"
)

// 订单状态
const (
	OrderNew             = "NEW"
	OrderPartiallyFilled = "PARTIALLY_FILLED"
	OrderFilled          = "FILLED"
	OrderCanceled        = "CANCELED"
	OrderRejected        = "REJECTED"
	OrderExpired         = "EXPIRED"
)

// 机器人状态
const (
	RobotStatusStopped = 1 // 已停止
	RobotStatusRunning = 2 // 运行中
	RobotStatusPaused  = 3 // 已暂停
	RobotStatusError   = 4 // 错误
)

// 交易所平台
const (
	PlatformBinance = "binance"
	PlatformOKX     = "okx"
	PlatformGate    = "gate"
)

// 日志操作类型
const (
	OpOpen   = "OPEN"   // 开仓
	OpClose  = "CLOSE"  // 平仓
	OpModify = "MODIFY" // 修改
	OpCancel = "CANCEL" // 取消
)

// 信号类型
const (
	SignalOpenLong   = "OPEN_LONG"   // 开多
	SignalOpenShort  = "OPEN_SHORT"  // 开空
	SignalCloseLong  = "CLOSE_LONG"  // 平多
	SignalCloseShort = "CLOSE_SHORT" // 平空
	SignalAddLong    = "ADD_LONG"    // 加多
	SignalAddShort   = "ADD_SHORT"   // 加空
)

// 支持的交易对
var DefaultSymbols = []string{
	"BTCUSDT",
	"ETHUSDT",
	"BNBUSDT",
	"SOLUSDT",
	"XRPUSDT",
	"DOGEUSDT",
	"ADAUSDT",
	"AVAXUSDT",
	"DOTUSDT",
	"LINKUSDT",
}

// 高风险交易对
var HighRiskSymbols = []string{
	"DOGEUSDT",
	"SHIBUSDT",
	"PEPEUSDT",
	"FLOKIUSDT",
	"BONKUSDT",
}

