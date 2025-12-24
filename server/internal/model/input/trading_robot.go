// Package input
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package input

import (
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingRobotListInp 机器人列表查询输入
type TradingRobotListInp struct {
	Page           int    `json:"page" v:"required|min:1" dc:"页码"`
	PageSize       int    `json:"pageSize" v:"required|min:1|max:100" dc:"每页数量"`
	Status         int    `json:"status" dc:"状态筛选"`
	Symbol         string `json:"symbol" dc:"交易对筛选"`
	RiskPreference string `json:"riskPreference" dc:"风险偏好筛选"`
	RobotName      string `json:"robotName" dc:"机器人名称模糊搜索"`
}

// TradingRobotListModel 机器人列表输出
type TradingRobotListModel struct {
	Id                      int64       `json:"id" dc:"ID"`
	RobotName               string      `json:"robotName" dc:"机器人名称"`
	ApiConfigId             int64       `json:"apiConfigId" dc:"API配置ID"`
	Symbol                  string      `json:"symbol" dc:"交易对"`
	Exchange                string      `json:"exchange" dc:"交易所"`
	Status                  int         `json:"status" dc:"状态"`
	RiskPreference          string      `json:"riskPreference" dc:"风险偏好"`
	AutoRiskPreference      int         `json:"autoRiskPreference" dc:"自动风险偏好"`
	MarketState             string      `json:"marketState" dc:"市场状态"`
	AutoMarketState         int         `json:"autoMarketState" dc:"自动市场状态"`
	Leverage                int         `json:"leverage" dc:"杠杆倍数"`
	MarginPercent           float64     `json:"marginPercent" dc:"保证金比例"`
	MaxProfitTarget         float64     `json:"maxProfitTarget" dc:"最大盈利目标"`
	MaxLossAmount           float64     `json:"maxLossAmount" dc:"最大亏损额"`
	LongCount               int         `json:"longCount" dc:"多单数"`
	ShortCount              int         `json:"shortCount" dc:"空单数"`
	TotalProfit             float64     `json:"totalProfit" dc:"总盈亏"`
	RuntimeSeconds          int         `json:"runtimeSeconds" dc:"运行时长"`
	AutoTradeEnabled        int         `json:"autoTradeEnabled" dc:"全自动下单"`
	AutoCloseEnabled        int         `json:"autoCloseEnabled" dc:"全自动平仓"`
	StartTime               *gtime.Time `json:"startTime" dc:"启动时间"`
	CreatedAt               *gtime.Time `json:"createdAt" dc:"创建时间"`
	StrategyGroupId         int64       `json:"strategyGroupId" dc:"策略组ID"`
	CurrentStrategySnapshot *gjson.Json `json:"currentStrategy" dc:"当前策略快照"`
	ScheduleStart           *gtime.Time `json:"scheduleStart" dc:"定时启动时间"`
	ScheduleStop            *gtime.Time `json:"scheduleStop" dc:"定时停止时间"`
}

// TradingRobotCreateInp 创建机器人输入
type TradingRobotCreateInp struct {
	// ①基础信息
	RobotName       string  `json:"robotName" v:"required|length:1,100" dc:"机器人名称"`
	ApiConfigId     int64   `json:"apiConfigId" v:"required" dc:"API接口ID"`
	MaxProfitTarget float64 `json:"maxProfitTarget" v:"min:0" dc:"最大盈利目标(USDT)"`
	MaxLossAmount   float64 `json:"maxLossAmount" v:"min:0" dc:"最大亏损额(USDT)"`
	MaxRuntime      int     `json:"maxRuntime" v:"min:0" dc:"最大运行时长(秒)"`

	// ②风险偏好（运行时根据映射关系动态获取，创建时不需要）
	// RiskPreference     string `json:"riskPreference" v:"required|in:conservative,balanced,aggressive" dc:"风险偏好"`
	// AutoRiskPreference int    `json:"autoRiskPreference" v:"in:0,1" dc:"自动风险偏好"`

	// ③市场行情（运行时实时分析，创建时不需要）
	// MarketState     string `json:"marketState" v:"required|in:trend,volatile,high-volatility,low-volatility" dc:"市场状态"`
	AutoMarketState int `json:"autoMarketState" v:"in:0,1" dc:"自动市场状态"`

	// ④下单配置（只保留必需字段）
	Exchange         string `json:"exchange" v:"required" dc:"交易所"`
	Symbol           string `json:"symbol" v:"required" dc:"交易对"`
	UseMonitorSignal int    `json:"useMonitorSignal" v:"in:0,1" dc:"采用方向预警信号"`
	// 以下字段运行时从策略模板加载，创建时不需要：
	// OrderType          string  `json:"orderType" v:"in:market,limit" dc:"订单类型"`
	// MarginMode         string  `json:"marginMode" v:"in:isolated,cross" dc:"保证金模式"`
	// Leverage           int     `json:"leverage" v:"required|between:1,125" dc:"杠杆倍数"`
	// MarginPercent      float64 `json:"marginPercent" v:"required|between:1,100" dc:"保证金比例(%)"`

	// ⑤自动平仓配置（运行时从策略模板加载，创建时不需要）
	// StopLossPercent         float64 `json:"stopLossPercent" v:"required|between:0.1,100" dc:"止损百分比(%)"`
	// ProfitRetreatPercent    float64 `json:"profitRetreatPercent" v:"required|between:0.1,100" dc:"止盈回撤百分比(%)"`
	// AutoStartRetreatPercent float64 `json:"autoStartRetreatPercent" v:"required|between:0.1,100" dc:"启动回撤百分比(%)"`

	// ⑥全自动交易开关
	AutoTradeEnabled *int `json:"autoTradeEnabled" v:"in:0,1" dc:"全自动下单：0=否,1=是（可选，nil表示不更新）"`
	AutoCloseEnabled *int `json:"autoCloseEnabled" v:"in:0,1" dc:"全自动平仓：0=否,1=是（可选，nil表示不更新）"`
	DualSidePosition *int `json:"dualSidePosition" v:"in:0,1" dc:"双向开单：0=单向,1=双向（可选，nil表示不更新，默认1=双向）"`

	// ⑦定时开关设置
	ScheduleStart string `json:"scheduleStart" dc:"定时启动时间"`
	ScheduleStop  string `json:"scheduleStop" dc:"定时停止时间"`

	// ⑧策略组ID（用于运行时加载策略模板）
	StrategyGroupId int64 `json:"strategyGroupId" dc:"策略组ID"`

	// ⑨市场状态与风险偏好映射（创建时静态配置，运行时根据映射关系匹配策略模板）
	MarketRiskMapping map[string]string `json:"marketRiskMapping" dc:"市场状态与风险偏好映射"`

	Remark string `json:"remark" dc:"备注"`
}

// TradingRobotUpdateInp 更新机器人输入
type TradingRobotUpdateInp struct {
	Id                      int64   `json:"id" v:"required" dc:"ID"`
	RobotName               string  `json:"robotName" v:"length:1,100" dc:"机器人名称"` // 改为可选，仅在完整更新时需要
	MaxProfitTarget         float64 `json:"maxProfitTarget" v:"min:0" dc:"最大盈利目标"`
	MaxLossAmount           float64 `json:"maxLossAmount" v:"min:0" dc:"最大亏损额"`
	MaxRuntime              int     `json:"maxRuntime" v:"min:0" dc:"最大运行时长"`
	RiskPreference          string  `json:"riskPreference" v:"in:conservative,balanced,aggressive" dc:"风险偏好"` // 改为可选
	AutoRiskPreference      int     `json:"autoRiskPreference" v:"in:0,1" dc:"自动风险偏好"`
	MarketState             string  `json:"marketState" v:"in:trend,volatile,high-volatility,low-volatility" dc:"市场状态"` // 改为可选
	AutoMarketState         int     `json:"autoMarketState" v:"in:0,1" dc:"自动市场状态"`
	Leverage                int     `json:"leverage" v:"between:1,125" dc:"杠杆倍数"`       // 改为可选
	MarginPercent           float64 `json:"marginPercent" v:"between:1,100" dc:"保证金比例"` // 改为可选
	UseMonitorSignal        int     `json:"useMonitorSignal" v:"in:0,1" dc:"采用方向预警信号"`
	StopLossPercent         float64 `json:"stopLossPercent" v:"between:0.1,100" dc:"止损百分比"`           // 改为可选
	ProfitRetreatPercent    float64 `json:"profitRetreatPercent" v:"between:0.1,100" dc:"止盈回撤百分比"`    // 改为可选
	AutoStartRetreatPercent float64 `json:"autoStartRetreatPercent" v:"between:0.1,100" dc:"启动回撤百分比"` // 改为可选
	AutoTradeEnabled        *int    `json:"autoTradeEnabled" dc:"全自动下单：0=否,1=是（可选，nil表示不更新）"`
	AutoCloseEnabled        *int    `json:"autoCloseEnabled" dc:"全自动平仓：0=否,1=是（可选，nil表示不更新）"`
	DualSidePosition        *int    `json:"dualSidePosition" dc:"双向开单：0=单向,1=双向（可选，nil表示不更新）"`
	Remark                  string  `json:"remark" dc:"备注"`
}

// TradingRobotDeleteInp 删除机器人输入
type TradingRobotDeleteInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingRobotViewInp 查看详情输入
type TradingRobotViewInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingRobotViewModel 详情输出
type TradingRobotViewModel struct {
	entity.TradingRobot
	ApiConfigName       string  `json:"apiConfigName" dc:"API配置名称"`
	MonitorWindow       int     `json:"monitorWindow" dc:"监控时间窗口(秒)"`
	VolatilityThreshold float64 `json:"volatilityThreshold" dc:"波动阈值(USDT)"`
}

// TradingRobotStartInp 启动机器人输入
type TradingRobotStartInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingRobotPauseInp 暂停机器人输入
type TradingRobotPauseInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingRobotStopInp 停止机器人输入
type TradingRobotStopInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingRobotStatusInp 更新状态输入
type TradingRobotStatusInp struct {
	Id     int64 `json:"id" v:"required" dc:"ID"`
	Status int   `json:"status" v:"required|in:1,2,3,4" dc:"状态：1=未启动,2=运行中,3=暂停,4=停用"`
}

// TradingRobotStatsInp 获取运行统计输入
type TradingRobotStatsInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingRobotStatsModel 运行统计输出
type TradingRobotStatsModel struct {
	Id              int64       `json:"id" dc:"ID"`
	RobotName       string      `json:"robotName" dc:"机器人名称"`
	Status          int         `json:"status" dc:"状态"`
	RuntimeSeconds  int         `json:"runtimeSeconds" dc:"已运行时长(秒)"`
	MaxRuntime      int         `json:"maxRuntime" dc:"最大运行时长(秒)"`
	LongCount       int         `json:"longCount" dc:"多单数"`
	ShortCount      int         `json:"shortCount" dc:"空单数"`
	TotalCount      int         `json:"totalCount" dc:"总订单数"`
	TotalProfit     float64     `json:"totalProfit" dc:"总盈亏"`
	MaxProfitTarget float64     `json:"maxProfitTarget" dc:"最大盈利目标"`
	MaxLossAmount   float64     `json:"maxLossAmount" dc:"最大亏损额"`
	ProfitRate      float64     `json:"profitRate" dc:"盈利完成率(%)"`
	StartTime       *gtime.Time `json:"startTime" dc:"启动时间"`
	CurrentStrategy *gjson.Json `json:"currentStrategy" dc:"当前策略"`
}

// TradingRobotRecommendStrategyInp 推荐策略输入
type TradingRobotRecommendStrategyInp struct {
	RiskPreference string `json:"riskPreference" v:"required|in:conservative,balanced,aggressive" dc:"风险偏好"`
	MarketState    string `json:"marketState" v:"required|in:trend,volatile,high-volatility,low-volatility" dc:"市场状态"`
}

// TradingRobotRecommendStrategyModel 推荐策略输出
type TradingRobotRecommendStrategyModel struct {
	StrategyKey             string  `json:"strategyKey" dc:"策略KEY"`
	StrategyName            string  `json:"strategyName" dc:"策略名称"`
	Description             string  `json:"description" dc:"策略描述"`
	MonitorWindow           int     `json:"monitorWindow" dc:"监控时间窗口(秒)"`
	VolatilityThreshold     float64 `json:"volatilityThreshold" dc:"波动阈值(USDT)"`
	LeverageMin             int     `json:"leverageMin" dc:"杠杆倍数最小值"`
	LeverageMax             int     `json:"leverageMax" dc:"杠杆倍数最大值"`
	MarginPercentMin        float64 `json:"marginPercentMin" dc:"保证金比例最小值(%)"`
	MarginPercentMax        float64 `json:"marginPercentMax" dc:"保证金比例最大值(%)"`
	StopLossPercent         float64 `json:"stopLossPercent" dc:"止损百分比(%)"`
	ProfitRetreatPercent    float64 `json:"profitRetreatPercent" dc:"止盈回撤百分比(%)"`
	AutoStartRetreatPercent float64 `json:"autoStartRetreatPercent" dc:"启动回撤百分比(%)"`
	GroupId                 int64   `json:"groupId" dc:"策略组ID"`
}
