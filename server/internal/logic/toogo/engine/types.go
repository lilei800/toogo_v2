// Package engine 机器人引擎模块 - 类型定义
// 将原robot_engine.go中的类型定义拆分到此文件
package engine

import (
	"sync"
	"time"

	"hotgo/internal/library/exchange"
	"hotgo/internal/library/market"
	"hotgo/internal/model/entity"
)

// ==================== 常量定义 ====================

const (
	// 波动率阈值
	HighVolatilityThreshold = 2.0
	LowVolatilityThreshold  = 0.5
	TrendStrengthThreshold  = 0.35

	// 价格窗口限制
	MaxPriceWindowSize   = 1000
	MaxSignalHistorySize = 100

	// 循环间隔
	AnalysisInterval = 1 * time.Second
	RiskInterval     = 3 * time.Second
	SignalInterval   = 1 * time.Second
	TradingInterval  = 500 * time.Millisecond
	SyncInterval     = 5 * time.Second
)

// 周期权重配置（【短线优化】提高短期周期权重，降低长期周期权重，更灵敏）
var TimeframeWeights = map[string]float64{
	"1m":  0.30, // 超短期：30%（捕捉短期波动）
	"5m":  0.40, // 短期：40%（短期趋势，最重要）
	"15m": 0.20, // 中期：20%（中期趋势，重要）
	"30m": 0.07, // 中期：7%（中期参考）
	"1h":  0.03, // 长期：3%（长期参考）
}

// ==================== 核心结构体 ====================

// RobotEngine 机器人核心引擎
type RobotEngine struct {
	Mu sync.RWMutex

	// 基础配置
	Robot     *entity.TradingRobot
	APIConfig *entity.TradingApiConfig
	Platform  string
	Exchange  exchange.Exchange

	// 模块 (已移除 RiskManager，直接使用策略模板参数)
	Analyzer  *RobotAnalyzer
	SignalGen *RobotSignalGen
	Trader    *RobotTrader
	PriceWin  *PriceWindow

	// 状态缓存
	LastTicker       *exchange.Ticker
	LastKlines       *market.KlineCache
	LastAnalysis     *MarketAnalysis
	LastSignal       *Signal
	CurrentPositions []*exchange.Position
	AccountBalance   *exchange.Balance

	// 时间记录
	LastTickerUpdate   time.Time
	LastAnalysisUpdate time.Time
	LastSignalUpdate   time.Time
	LastPositionUpdate time.Time
	LastBalanceUpdate  time.Time

	// 持仓跟踪
	PositionTrackers map[string]*PositionTracker

	// 市场状态与策略配置
	LastMarketState       string
	MarketRiskMapping     map[string]string
	CurrentStrategyParams *StrategyParams
	VolatilityConfig      *VolatilityConfig

	// 运行状态
	Running bool
	StopCh  chan struct{}

	// 锁
	OrderLock sync.Mutex
	CloseLock sync.Mutex
}

// ==================== 市场分析相关 ====================

// MarketAnalysis 市场分析结果
type MarketAnalysis struct {
	Timestamp time.Time

	// 市场状态
	MarketState     string  // trend/volatile/high_vol/low_vol
	MarketStateConf float64 // 置信度
	TrendDirection  string  // up/down/neutral
	TrendStrength   float64 // 趋势强度 0-100
	Volatility      float64 // 波动率
	VolatilityLevel string  // low/normal/high

	// 多周期分析
	TimeframeScores map[string]*TimeframeScore

	// 技术指标
	Indicators *TechnicalIndicators
}

// TimeframeScore 单周期评分
type TimeframeScore struct {
	Timeframe     string
	Direction     string  // up/down/neutral
	Strength      float64 // 方向强度 0-100
	TrendStrength float64 // 趋势强度 0-1
	Volatility    float64 // 波动率
	MarketState   string  // trend/volatile/high_vol/low_vol
	MACD          float64
	EMA12         float64
	EMA26         float64
	KlinesCount   int
}

// TechnicalIndicators 技术指标汇总
type TechnicalIndicators struct {
	TrendScore      float64 // 趋势综合评分 -100 ~ 100
	VolatilityScore float64 // 波动评分 0-100
}

// ==================== 信号相关 ====================

// Signal 方向信号
type Signal struct {
	Timestamp time.Time

	// 方向 LONG/SHORT/NEUTRAL
	Direction string

	// 信号强度 0-100
	Strength float64

	// 置信度 0-100
	Confidence float64

	// 建议操作
	Action string

	// 多周期对齐数
	AlignedTimeframes int

	// 信号原因
	Reason string

	// 窗口信号相关
	WindowMaxPrice  float64
	WindowMinPrice  float64
	CurrentPrice    float64
	DistanceFromMin float64
	DistanceFromMax float64
	SignalThreshold float64
	SignalProgress  float64
	SignalType      string // window/analysis
}

// SignalHistoryItem 信号历史记录
type SignalHistoryItem struct {
	Timestamp int64  `json:"timestamp"`
	Signal    string `json:"signal"`
}

// ==================== 价格窗口相关 ====================

// PricePoint 价格数据点
type PricePoint struct {
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Symbol    string
	Window    int     // 窗口期（秒）
	Threshold float64 // 信号阈值
}

// ==================== 策略相关 ====================

// StrategyParams 策略参数
type StrategyParams struct {
	Window                  int     // 时间窗口(秒)
	Threshold               float64 // 波动阈值(USDT)
	LeverageMin             int     // 杠杆最小值
	LeverageMax             int     // 杠杆最大值
	MarginPercentMin        float64 // 保证金比例最小值
	MarginPercentMax        float64 // 保证金比例最大值
	StopLossPercent         float64 // 止损百分比
	ProfitRetreatPercent    float64 // 止盈回撤百分比
	AutoStartRetreatPercent float64 // 启动回撤百分比
}

// VolatilityConfig 波动率配置（市场状态阈值 + 5个时间周期权重）
type VolatilityConfig struct {
	HighVolatilityThreshold float64 // 高波动阈值（判断高波动市场）
	LowVolatilityThreshold  float64 // 低波动阈值（判断低波动市场）
	TrendStrengthThreshold  float64 // 趋势强度阈值（判断趋势市场）
	Weight1m                float64 // 1分钟周期权重
	Weight5m                float64 // 5分钟周期权重
	Weight15m               float64 // 15分钟周期权重
	Weight30m               float64 // 30分钟周期权重
	Weight1h                float64 // 1小时周期权重
	Symbol                  string  // 配置名称（交易对或"默认"/"全局"）
}

// ==================== 持仓跟踪 ====================

// PositionTracker 持仓跟踪器
type PositionTracker struct {
	PositionSide      string
	EntryMargin       float64
	EntryTime         time.Time
	HighestProfit     float64   // 最高盈利金额
	LowestProfit      float64   // 最低盈利金额（负数表示亏损）
	TakeProfitEnabled bool
}

// ==================== 状态相关 ====================

// EngineStatus 引擎状态
type EngineStatus struct {
	RobotId  int64  `json:"robotId"`
	Symbol   string `json:"symbol"`
	Platform string `json:"platform"`
	Running  bool   `json:"running"`

	// 连接状态
	Connected bool    `json:"connected"`
	LastPrice float64 `json:"lastPrice"`

	// 账户
	TotalBalance float64 `json:"totalBalance"`
	AvailBalance float64 `json:"availBalance"`

	// 市场分析
	MarketState    string  `json:"marketState"`
	TrendDirection string  `json:"trendDirection"`
	Volatility     float64 `json:"volatility"`

	// 方向信号
	SignalDirection  string  `json:"signalDirection"`
	SignalStrength   float64 `json:"signalStrength"`
	SignalConfidence float64 `json:"signalConfidence"`

	// 持仓
	HasPosition   bool    `json:"hasPosition"`
	PositionSide  string  `json:"positionSide"`
	PositionAmt   float64 `json:"positionAmt"`
	EntryPrice    float64 `json:"entryPrice"`
	UnrealizedPnl float64 `json:"unrealizedPnl"`

	// 价格窗口数据
	PriceWindowData    []PriceWindowPoint `json:"priceWindowData"`
	WindowMinPrice     float64            `json:"windowMinPrice"`
	WindowMaxPrice     float64            `json:"windowMaxPrice"`
	WindowCurrentPrice float64            `json:"windowCurrentPrice"`
	LongTriggerPrice   float64            `json:"longTriggerPrice"`
	ShortTriggerPrice  float64            `json:"shortTriggerPrice"`
	SignalProgress     float64            `json:"signalProgress"`
	SignalReason       string             `json:"signalReason"`

	// 策略配置
	StrategyWindow     int     `json:"strategyWindow"`
	StrategyThreshold  float64 `json:"strategyThreshold"`
	CurrentMarketState string  `json:"currentMarketState"`
	CurrentRiskPref    string  `json:"currentRiskPref"`
}

// PriceWindowPoint 价格窗口数据点（用于图表）
type PriceWindowPoint struct {
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
}

// ==================== 默认策略参数 ====================

// DefaultStrategyParams 默认策略参数配置
var DefaultStrategyParams = map[string]map[string]struct {
	Window    int
	Threshold float64
}{
	"trend": {
		"conservative": {Window: 120, Threshold: 15},
		"balanced":     {Window: 90, Threshold: 12},
		"aggressive":   {Window: 60, Threshold: 8},
	},
	"volatile": {
		"conservative": {Window: 90, Threshold: 20},
		"balanced":     {Window: 60, Threshold: 15},
		"aggressive":   {Window: 45, Threshold: 10},
	},
	"high_vol": {
		"conservative": {Window: 60, Threshold: 25},
		"balanced":     {Window: 45, Threshold: 20},
		"aggressive":   {Window: 30, Threshold: 15},
	},
	"low_vol": {
		"conservative": {Window: 180, Threshold: 8},
		"balanced":     {Window: 120, Threshold: 6},
		"aggressive":   {Window: 90, Threshold: 5},
	},
}

