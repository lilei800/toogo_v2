// Package engine 机器人引擎模块 - 策略加载器
package engine

import (
	"context"
	"encoding/json"

	"hotgo/internal/dao"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// LoadStrategyParams 加载策略参数
// 简化版：优先从数据库加载，否则使用默认值
func LoadStrategyParams(ctx context.Context, robot *entity.TradingRobot, marketState, riskPreference string) *StrategyParams {
	params := &StrategyParams{}

	// 1. 尝试从机器人配置获取策略组ID
	var groupId int64 = 0
	if robot.CurrentStrategy != "" {
		var strategyData map[string]interface{}
		if err := json.Unmarshal([]byte(robot.CurrentStrategy), &strategyData); err == nil {
			if gid, ok := strategyData["group_id"].(float64); ok {
				groupId = int64(gid)
			}
		}
	}

	// 2. 从策略模板表查询
	if groupId > 0 {
		var strategy *entity.TradingStrategyTemplate
		err := dao.TradingStrategyTemplate.Ctx(ctx).
			Where("group_id", groupId).
			Where(dao.TradingStrategyTemplate.Columns().MarketState, marketState).
			Where(dao.TradingStrategyTemplate.Columns().RiskPreference, riskPreference).
			Where(dao.TradingStrategyTemplate.Columns().IsActive, 1).
			Scan(&strategy)

		if err == nil && strategy != nil {
			params.Window = strategy.MonitorWindow
			params.Threshold = strategy.VolatilityThreshold
			params.LeverageMin = strategy.Leverage
			params.LeverageMax = strategy.Leverage
			params.MarginPercentMin = strategy.MarginPercent
			params.MarginPercentMax = strategy.MarginPercent
			params.StopLossPercent = strategy.StopLossPercent
			params.ProfitRetreatPercent = strategy.ProfitRetreatPercent
			params.AutoStartRetreatPercent = strategy.AutoStartRetreatPercent

			g.Log().Debugf(ctx, "[StrategyLoader] robotId=%d 从策略模板加载参数: market=%s, risk=%s",
				robot.Id, marketState, riskPreference)
			return params
		}
	}

	// 3. 使用默认参数
	params = getDefaultParams(riskPreference)

	// 时间窗口和波动值从默认策略获取
	if marketParams, ok := DefaultStrategyParams[marketState]; ok {
		if riskParams, ok := marketParams[riskPreference]; ok {
			params.Window = riskParams.Window
			params.Threshold = riskParams.Threshold
		}
	}

	if params.Window == 0 {
		params.Window = 60
	}
	if params.Threshold == 0 {
		params.Threshold = 15
	}

	g.Log().Debugf(ctx, "[StrategyLoader] robotId=%d 使用默认策略参数: market=%s, risk=%s",
		robot.Id, marketState, riskPreference)
	return params
}

// getDefaultParams 获取默认参数
func getDefaultParams(riskPreference string) *StrategyParams {
	params := &StrategyParams{}

	switch riskPreference {
	case "aggressive":
		params.LeverageMin = 10
		params.LeverageMax = 20
		params.MarginPercentMin = 10
		params.MarginPercentMax = 20
		params.StopLossPercent = 8
		params.ProfitRetreatPercent = 20
		params.AutoStartRetreatPercent = 5
	case "balanced":
		params.LeverageMin = 5
		params.LeverageMax = 10
		params.MarginPercentMin = 8
		params.MarginPercentMax = 15
		params.StopLossPercent = 5
		params.ProfitRetreatPercent = 25
		params.AutoStartRetreatPercent = 3
	default: // conservative
		params.LeverageMin = 3
		params.LeverageMax = 5
		params.MarginPercentMin = 5
		params.MarginPercentMax = 10
		params.StopLossPercent = 3
		params.ProfitRetreatPercent = 30
		params.AutoStartRetreatPercent = 2
	}

	return params
}

// GetFloatFromMap 从map中安全获取float64值
func GetFloatFromMap(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return defaultValue
}

