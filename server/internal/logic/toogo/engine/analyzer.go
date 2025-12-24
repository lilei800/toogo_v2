// Package engine 机器人引擎模块 - 市场分析器
package engine

import (
	"context"
	"math"
	"time"

	"hotgo/internal/library/exchange"
)

// RobotAnalyzer 机器人市场分析器
type RobotAnalyzer struct {
	engine *RobotEngine
}

// NewRobotAnalyzer 创建分析器
func NewRobotAnalyzer(engine *RobotEngine) *RobotAnalyzer {
	return &RobotAnalyzer{engine: engine}
}

// Analyze 执行市场分析
func (a *RobotAnalyzer) Analyze(ctx context.Context) *MarketAnalysis {
	klines := a.engine.LastKlines
	if klines == nil {
		return nil
	}

	analysis := &MarketAnalysis{
		Timestamp:       time.Now(),
		TimeframeScores: make(map[string]*TimeframeScore),
		Indicators:      &TechnicalIndicators{},
	}

	// 分析3个核心周期
	timeframes := map[string][]*exchange.Kline{
		"5m":  klines.Klines5m,
		"15m": klines.Klines15m,
		"1h":  klines.Klines1h,
	}

	for tf, data := range timeframes {
		if len(data) < 26 {
			continue
		}
		score := a.analyzeTimeframe(data)
		score.Timeframe = tf
		analysis.TimeframeScores[tf] = score
	}

	// 计算综合指标
	a.calculateOverallIndicators(analysis)

	// 判断市场状态
	a.determineMarketState(analysis)

	return analysis
}

// analyzeTimeframe 分析单周期
func (a *RobotAnalyzer) analyzeTimeframe(klines []*exchange.Kline) *TimeframeScore {
	score := &TimeframeScore{
		KlinesCount: len(klines),
	}

	if len(klines) < 26 {
		return score
	}

	// 计算收盘价序列
	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}

	// 计算EMA和MACD
	score.EMA12 = a.calculateEMA(closes, 12)
	score.EMA26 = a.calculateEMA(closes, 26)
	score.MACD = score.EMA12 - score.EMA26

	// 计算趋势强度
	score.TrendStrength = a.calculateTrendStrength(klines)

	// 计算波动率
	score.Volatility = a.calculateVolatility(klines)

	// 判断方向和强度
	if score.EMA12 > score.EMA26 && score.MACD > 0 {
		score.Direction = "up"
		score.Strength = math.Min(100, 50+score.TrendStrength*50)
	} else if score.EMA12 < score.EMA26 && score.MACD < 0 {
		score.Direction = "down"
		score.Strength = math.Min(100, 50+score.TrendStrength*50)
	} else {
		score.Direction = "neutral"
		score.Strength = 30
	}

	// 判断市场状态
	score.MarketState = a.determineTimeframeMarketState(score.TrendStrength, score.Volatility)

	return score
}

// calculateTrendStrength 计算趋势强度
func (a *RobotAnalyzer) calculateTrendStrength(klines []*exchange.Kline) float64 {
	if len(klines) < 10 {
		return 0
	}

	n := len(klines)
	var sumX, sumY, sumXY, sumX2 float64
	for i, k := range klines {
		x := float64(i)
		y := k.Close
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := float64(n)*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}
	slope := (float64(n)*sumXY - sumX*sumY) / denominator

	avgPrice := sumY / float64(n)
	if avgPrice == 0 {
		return 0
	}
	normalizedSlope := math.Abs(slope) / avgPrice * 100

	return math.Min(1, normalizedSlope)
}

// calculateVolatility 计算波动率
func (a *RobotAnalyzer) calculateVolatility(klines []*exchange.Kline) float64 {
	if len(klines) < 10 {
		return 1.0
	}

	var atr float64
	for i := 1; i < len(klines); i++ {
		high := klines[i].High
		low := klines[i].Low
		prevClose := klines[i-1].Close

		tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		atr += tr
	}
	atr /= float64(len(klines) - 1)

	lastPrice := klines[len(klines)-1].Close
	if lastPrice > 0 {
		return (atr / lastPrice) * 100
	}
	return 1.0
}

// determineTimeframeMarketState 判断单周期市场状态
func (a *RobotAnalyzer) determineTimeframeMarketState(trendStrength, volatility float64) string {
	a.engine.Mu.RLock()
	config := a.engine.VolatilityConfig
	a.engine.Mu.RUnlock()

	if config == nil {
		config = &VolatilityConfig{
			HighVolatilityThreshold: HighVolatilityThreshold,
			LowVolatilityThreshold:  LowVolatilityThreshold,
			TrendStrengthThreshold:  TrendStrengthThreshold,
		}
	}

	highThreshold := config.HighVolatilityThreshold
	lowThreshold := config.LowVolatilityThreshold
	trendThreshold := config.TrendStrengthThreshold
	
	// 如果配置中的趋势阈值为0，使用默认值
	if trendThreshold == 0 {
		trendThreshold = TrendStrengthThreshold
	}

	// 优先判断趋势市场
	if trendStrength > trendThreshold && volatility >= lowThreshold && volatility <= highThreshold*1.5 {
		return "trend"
	}

	if volatility >= highThreshold {
		return "high_vol"
	} else if volatility <= lowThreshold {
		return "low_vol"
	}
	return "range"
}

// calculateEMA 计算EMA
func (a *RobotAnalyzer) calculateEMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}

	multiplier := 2.0 / float64(period+1)
	ema := data[0]

	for i := 1; i < len(data); i++ {
		ema = (data[i]-ema)*multiplier + ema
	}

	return ema
}

// calculateOverallIndicators 计算综合指标
func (a *RobotAnalyzer) calculateOverallIndicators(analysis *MarketAnalysis) {
	var weightedTrendSum, totalWeight float64
	var avgVolatility float64
	var volatilityCount int

	// 获取配置中的权重
	a.engine.Mu.RLock()
	config := a.engine.VolatilityConfig
	a.engine.Mu.RUnlock()

	// 根据配置构建权重映射
	weights := a.getTimeframeWeights(config)

	for tf, score := range analysis.TimeframeScores {
		weight := weights[tf]
		if weight == 0 {
			weight = 0.2 // 兜底默认权重
		}
		totalWeight += weight

		if score.Direction == "up" {
			weightedTrendSum += score.Strength * weight
		} else if score.Direction == "down" {
			weightedTrendSum -= score.Strength * weight
		}

		if score.Volatility > 0 {
			avgVolatility += score.Volatility
			volatilityCount++
		}
	}

	if totalWeight > 0 {
		analysis.Indicators.TrendScore = weightedTrendSum / totalWeight
	}

	if volatilityCount > 0 {
		analysis.Volatility = avgVolatility / float64(volatilityCount)
	} else if klines := a.engine.LastKlines; klines != nil && len(klines.Klines5m) > 0 {
		analysis.Volatility = a.calculateVolatility(klines.Klines5m)
	}

	analysis.Indicators.VolatilityScore = math.Min(100, analysis.Volatility*20)
}

// getTimeframeWeights 从配置获取时间周期权重
func (a *RobotAnalyzer) getTimeframeWeights(config *VolatilityConfig) map[string]float64 {
	weights := make(map[string]float64)
	
	if config != nil {
		// 使用配置中的权重
		if config.Weight1m > 0 {
			weights["1m"] = config.Weight1m
		}
		if config.Weight5m > 0 {
			weights["5m"] = config.Weight5m
		}
		if config.Weight15m > 0 {
			weights["15m"] = config.Weight15m
		}
		if config.Weight30m > 0 {
			weights["30m"] = config.Weight30m
		}
		if config.Weight1h > 0 {
			weights["1h"] = config.Weight1h
		}
	}
	
	// 如果配置为空或权重都为0，使用全局默认权重
	if len(weights) == 0 {
		weights = TimeframeWeights
	}
	
	return weights
}

// determineMarketState 综合判断市场状态
func (a *RobotAnalyzer) determineMarketState(analysis *MarketAnalysis) {
	total := len(analysis.TimeframeScores)
	if total == 0 {
		analysis.MarketState = "range"
		analysis.VolatilityLevel = "normal"
		return
	}

	var upCount, downCount int
	var trendCount, highVolCount, lowVolCount int

	for _, score := range analysis.TimeframeScores {
		if score.Direction == "up" {
			upCount++
		} else if score.Direction == "down" {
			downCount++
		}

		switch score.MarketState {
		case "trend":
			trendCount++
		case "high_vol":
			highVolCount++
		case "low_vol":
			lowVolCount++
		}
	}

	// 判断趋势方向
	if upCount >= 2 {
		analysis.TrendDirection = "up"
		analysis.TrendStrength = float64(upCount) / float64(total) * 100
	} else if downCount >= 2 {
		analysis.TrendDirection = "down"
		analysis.TrendStrength = float64(downCount) / float64(total) * 100
	} else {
		analysis.TrendDirection = "neutral"
		analysis.TrendStrength = 30
	}

	// 判断市场状态
	if highVolCount >= 2 {
		analysis.MarketState = "high_vol"
		analysis.MarketStateConf = 0.8
		analysis.VolatilityLevel = "high"
		return
	}

	if trendCount >= 2 {
		analysis.MarketState = "trend"
		analysis.MarketStateConf = float64(trendCount) / float64(total)
		analysis.VolatilityLevel = "normal"
		return
	}

	if lowVolCount >= 2 {
		analysis.MarketState = "low_vol"
		analysis.MarketStateConf = 0.7
		analysis.VolatilityLevel = "low"
		return
	}

	analysis.MarketState = "range"
	analysis.MarketStateConf = 0.6
	analysis.VolatilityLevel = "normal"
}

