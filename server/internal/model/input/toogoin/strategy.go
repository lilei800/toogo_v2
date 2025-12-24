// Package toogoin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogoin

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// StrategyTemplateListInp 策略模板列表输入
type StrategyTemplateListInp struct {
	form.PageReq
	RiskPreference string `json:"riskPreference" description:"风险偏好"`
	MarketState    string `json:"marketState" description:"市场状态"`
	IsOfficial     int    `json:"isOfficial" description:"是否官方推荐"`
	IsActive       int    `json:"isActive" description:"是否激活"`
}

// StrategyTemplateListModel 策略模板列表返回
type StrategyTemplateListModel struct {
	*entity.ToogoStrategyTemplate
}

// StrategyTemplateEditInp 编辑策略模板输入
type StrategyTemplateEditInp struct {
	Id                   int64   `json:"id" description:"ID"`
	StrategyKey          string  `json:"strategyKey" v:"required" description:"策略KEY"`
	StrategyName         string  `json:"strategyName" v:"required" description:"策略名称"`
	RiskPreference       string  `json:"riskPreference" v:"required|in:conservative,balanced,aggressive" description:"风险偏好"`
	MarketState          string  `json:"marketState" v:"required|in:trend,volatile,high_volatility,low_volatility" description:"市场状态"`
	TimeWindow           int     `json:"timeWindow" v:"required|min:60" description:"时间窗口(秒)"`
	VolatilityPoints     float64 `json:"volatilityPoints" v:"required|min:1" description:"波动点数"`
	LeverageMin          int     `json:"leverageMin" v:"required|min:1|max:50" description:"杠杆最小值"`
	LeverageMax          int     `json:"leverageMax" v:"required|min:1|max:50" description:"杠杆最大值"`
	MarginPercentMin     float64 `json:"marginPercentMin" v:"required|min:1|max:100" description:"保证金比例最小值"`
	MarginPercentMax     float64 `json:"marginPercentMax" v:"required|min:1|max:100" description:"保证金比例最大值"`
	StopLossPercent      float64 `json:"stopLossPercent" v:"required|min:1|max:50" description:"止损百分比"`
	ProfitRetreatPercent float64 `json:"profitRetreatPercent" v:"required|min:1|max:100" description:"止盈回撤百分比"`
	StartRetreatPercent  float64 `json:"startRetreatPercent" v:"required|min:1|max:50" description:"启动回撤百分比"`
	VolatilityConfig     string  `json:"volatilityConfig" description:"波动率配置(JSON)"`
	Description          string  `json:"description" description:"策略描述"`
	IsOfficial           int     `json:"isOfficial" description:"是否官方推荐"`
	IsActive             int     `json:"isActive" description:"是否激活"`
	Sort                 int     `json:"sort" description:"排序"`
}

// StrategyTemplateDeleteInp 删除策略模板输入
type StrategyTemplateDeleteInp struct {
	Id int64 `json:"id" v:"required" description:"ID"`
}

// GetStrategyByConditionInp 根据条件获取策略输入
type GetStrategyByConditionInp struct {
	RiskPreference string `json:"riskPreference" v:"required" description:"风险偏好"`
	MarketState    string `json:"marketState" v:"required" description:"市场状态"`
}

// GetStrategyByConditionModel 根据条件获取策略返回
type GetStrategyByConditionModel struct {
	*entity.ToogoStrategyTemplate
}

// PowerConsumeListInp 算力消耗记录列表输入
type PowerConsumeListInp struct {
	form.PageReq
	UserId    int64    `json:"userId" description:"用户ID"`
	RobotId   int64    `json:"robotId" description:"机器人ID"`
	CreatedAt []string `json:"createdAt" description:"创建时间"`
}

// PowerConsumeListModel 算力消耗记录列表返回
type PowerConsumeListModel struct {
	*entity.ToogoPowerConsume
	RobotName string `json:"robotName" description:"机器人名称"`
}

// PowerConsumeStatInp 算力消耗统计输入
type PowerConsumeStatInp struct {
	UserId int64 `json:"userId" v:"required" description:"用户ID"`
}

// PowerConsumeStatModel 算力消耗统计返回
type PowerConsumeStatModel struct {
	TotalConsume   float64 `json:"totalConsume" description:"累计消耗算力"`
	TodayConsume   float64 `json:"todayConsume" description:"今日消耗算力"`
	WeekConsume    float64 `json:"weekConsume" description:"本周消耗算力"`
	MonthConsume   float64 `json:"monthConsume" description:"本月消耗算力"`
	TotalProfit    float64 `json:"totalProfit" description:"累计盈利金额"`
	AvgConsumeRate float64 `json:"avgConsumeRate" description:"平均消耗比例"`
}

