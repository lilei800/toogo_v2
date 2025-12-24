// Package input
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package input

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingMonitorLogListInp 监控日志列表查询输入
type TradingMonitorLogListInp struct {
	Page       int    `json:"page" v:"required|min:1" dc:"页码"`
	PageSize   int    `json:"pageSize" v:"required|min:1|max:100" dc:"每页数量"`
	RobotId    int64  `json:"robotId" dc:"机器人ID筛选"`
	SignalType string `json:"signalType" dc:"信号类型筛选"`
	Symbol     string `json:"symbol" dc:"交易对筛选"`
	StartDate  string `json:"startDate" dc:"开始日期"`
	EndDate    string `json:"endDate" dc:"结束日期"`
}

// TradingMonitorLogListModel 监控日志列表输出
type TradingMonitorLogListModel struct {
	Id              int64       `json:"id" dc:"ID"`
	RobotId         int64       `json:"robotId" dc:"机器人ID"`
	RobotName       string      `json:"robotName" dc:"机器人名称"`
	Symbol          string      `json:"symbol" dc:"交易对"`
	CurrentPrice    float64     `json:"currentPrice" dc:"当前价格"`
	SignalType      string      `json:"signalType" dc:"信号类型"`
	SignalStrength  float64     `json:"signalStrength" dc:"信号强度"`
	SignalDetail    string      `json:"signalDetail" dc:"信号详情"`
	ActionTaken     string      `json:"actionTaken" dc:"采取的行动"`
	ActionResult    string      `json:"actionResult" dc:"行动结果"`
	VolatilityIndex float64     `json:"volatilityIndex" dc:"波动率指数"`
	TrendIndex      float64     `json:"trendIndex" dc:"趋势指数"`
	CreateTime      *gtime.Time `json:"createTime" dc:"创建时间"`
}

// TradingMonitorTickerInp 获取实时行情输入
type TradingMonitorTickerInp struct {
	ApiConfigId int64  `json:"apiConfigId" v:"required" dc:"API配置ID"`
	Symbol      string `json:"symbol" v:"required" dc:"交易对"`
}

// TradingMonitorTickerModel 实时行情输出
type TradingMonitorTickerModel struct {
	Symbol    string  `json:"symbol" dc:"交易对"`
	LastPrice float64 `json:"lastPrice" dc:"最新价格"`
	High24h   float64 `json:"high24h" dc:"24小时最高价"`
	Low24h    float64 `json:"low24h" dc:"24小时最低价"`
	Volume24h float64 `json:"volume24h" dc:"24小时成交量"`
	Change24h float64 `json:"change24h" dc:"24小时涨跌幅"`
	Timestamp string  `json:"timestamp" dc:"时间戳"`
}

// TradingMonitorMarketStateInp 获取市场状态输入
type TradingMonitorMarketStateInp struct {
	ApiConfigId int64  `json:"apiConfigId" v:"required" dc:"API配置ID"`
	Symbol      string `json:"symbol" v:"required" dc:"交易对"`
	TimeWindow  int    `json:"timeWindow" dc:"时间窗口(分钟), 默认60"`
}

// TradingMonitorMarketStateModel 市场状态输出
type TradingMonitorMarketStateModel struct {
	Symbol           string  `json:"symbol" dc:"交易对"`
	CurrentPrice     float64 `json:"currentPrice" dc:"当前价格"`
	MarketState      string  `json:"marketState" dc:"市场状态：trend/volatile/high-volatility/low-volatility"`
	VolatilityIndex  float64 `json:"volatilityIndex" dc:"波动率指数"`
	TrendIndex       float64 `json:"trendIndex" dc:"趋势指数"`
	SignalType       string  `json:"signalType" dc:"信号类型：long/short/neutral"`
	SignalStrength   float64 `json:"signalStrength" dc:"信号强度 0-100"`
	RecommendedRisk  string  `json:"recommendedRisk" dc:"推荐风险偏好"`
	Analysis         string  `json:"analysis" dc:"市场分析"`
	Timestamp        string  `json:"timestamp" dc:"时间戳"`
}

