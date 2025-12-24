// Package engine 机器人引擎模块 - 市场分析器调试工具
package engine

import (
	"fmt"
)

// GetMarketStateDebugInfo 获取市场状态判断的详细信息（用于诊断）
func (a *RobotAnalyzer) GetMarketStateDebugInfo() map[string]interface{} {
	if a.engine.LastAnalysis == nil {
		return map[string]interface{}{
			"error": "暂无市场分析数据",
		}
	}

	analysis := a.engine.LastAnalysis
	config := a.engine.VolatilityConfig

	// 配置信息
	configInfo := map[string]interface{}{
		"symbol":                  "未知",
		"highVolatilityThreshold": 2.0,
		"lowVolatilityThreshold":  0.5,
		"trendStrengthThreshold":  0.35,
	}
	if config != nil {
		configInfo["symbol"] = config.Symbol
		configInfo["highVolatilityThreshold"] = config.HighVolatilityThreshold
		configInfo["lowVolatilityThreshold"] = config.LowVolatilityThreshold
		configInfo["trendStrengthThreshold"] = config.TrendStrengthThreshold
	}

	// 各周期详情
	timeframeDetails := make([]map[string]interface{}, 0)
	var trendCount, highVolCount, lowVolCount, rangeCount int

	for _, tf := range []string{"5m", "15m", "1h"} {
		if score, ok := analysis.TimeframeScores[tf]; ok {
			detail := map[string]interface{}{
				"timeframe":     tf,
				"direction":     score.Direction,
				"trendStrength": fmt.Sprintf("%.3f", score.TrendStrength),
				"volatility":    fmt.Sprintf("%.3f%%", score.Volatility),
				"marketState":   score.MarketState,
			}

			// 判断逻辑解释
			reasons := []string{}
			if score.MarketState == "trend" {
				reasons = append(reasons, fmt.Sprintf("✅ 趋势强度%.3f > %.2f", score.TrendStrength, configInfo["trendStrengthThreshold"]))
				reasons = append(reasons, fmt.Sprintf("✅ 波动率%.3f 在合理范围", score.Volatility))
				trendCount++
			} else if score.MarketState == "high_vol" {
				reasons = append(reasons, fmt.Sprintf("⚠️ 波动率%.3f%% >= %.2f%% (高波动)", score.Volatility, configInfo["highVolatilityThreshold"]))
				highVolCount++
			} else if score.MarketState == "low_vol" {
				reasons = append(reasons, fmt.Sprintf("😴 波动率%.3f%% <= %.2f%% (低波动)", score.Volatility, configInfo["lowVolatilityThreshold"]))
				lowVolCount++
			} else {
				if score.TrendStrength <= configInfo["trendStrengthThreshold"].(float64) {
					reasons = append(reasons, fmt.Sprintf("❌ 趋势强度%.3f <= %.2f (趋势不明显)", score.TrendStrength, configInfo["trendStrengthThreshold"]))
				}
				reasons = append(reasons, fmt.Sprintf("📊 波动率%.3f%% 在中等区间 (%.2f-%.2f)", 
					score.Volatility, 
					configInfo["lowVolatilityThreshold"], 
					configInfo["highVolatilityThreshold"]))
				rangeCount++
			}
			detail["reasons"] = reasons

			timeframeDetails = append(timeframeDetails, detail)
		}
	}

	// 综合判断逻辑
	finalDecision := map[string]interface{}{
		"marketState": analysis.MarketState,
		"confidence":  fmt.Sprintf("%.1f%%", analysis.MarketStateConf*100),
	}

	decisionReasons := []string{}
	if highVolCount >= 2 {
		decisionReasons = append(decisionReasons, fmt.Sprintf("✅ %d个周期判定为高波动 (≥2) → high_vol", highVolCount))
	} else if trendCount >= 2 {
		decisionReasons = append(decisionReasons, fmt.Sprintf("✅ %d个周期判定为趋势 (≥2) → trend", trendCount))
	} else if lowVolCount >= 2 {
		decisionReasons = append(decisionReasons, fmt.Sprintf("✅ %d个周期判定为低波动 (≥2) → low_vol", lowVolCount))
	} else {
		decisionReasons = append(decisionReasons, fmt.Sprintf("📊 趋势:%d, 高波动:%d, 低波动:%d, 震荡:%d", 
			trendCount, highVolCount, lowVolCount, rangeCount))
		decisionReasons = append(decisionReasons, "❌ 没有任何状态达到≥2个周期 → range (震荡)")
	}
	finalDecision["reasons"] = decisionReasons

	// 统计计数
	voteSummary := map[string]interface{}{
		"trend":    trendCount,
		"high_vol": highVolCount,
		"low_vol":  lowVolCount,
		"range":    rangeCount,
		"total":    len(analysis.TimeframeScores),
	}

	return map[string]interface{}{
		"config":            configInfo,
		"timeframeDetails":  timeframeDetails,
		"voteSummary":       voteSummary,
		"finalDecision":     finalDecision,
		"overallVolatility": fmt.Sprintf("%.3f%%", analysis.Volatility),
		"trendDirection":    analysis.TrendDirection,
		"trendStrength":     fmt.Sprintf("%.1f", analysis.TrendStrength),
	}
}

