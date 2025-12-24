// Package strategy 策略分析库 - 多时间周期市场状态分析
package strategy

import (
	"context"
	"math"
	"time"
)

// MarketState 市场状态
type MarketState string

const (
	MarketStrongUptrend   MarketState = "STRONG_UPTREND"   // 强势上涨
	MarketMildUptrend     MarketState = "MILD_UPTREND"     // 温和上涨
	MarketRanging         MarketState = "RANGING"          // 震荡
	MarketMildDowntrend   MarketState = "MILD_DOWNTREND"   // 温和下跌
	MarketStrongDowntrend MarketState = "STRONG_DOWNTREND" // 强势下跌
	MarketHighVolatility  MarketState = "HIGH_VOLATILITY"  // 高波动
	MarketLowVolatility   MarketState = "LOW_VOLATILITY"   // 低波动
)

// TimeFrame 时间周期
type TimeFrame string

const (
	TimeFrame1m  TimeFrame = "1m"
	TimeFrame5m  TimeFrame = "5m"
	TimeFrame15m TimeFrame = "15m"
	TimeFrame30m TimeFrame = "30m"
	TimeFrame1h  TimeFrame = "1h"
	TimeFrame4h  TimeFrame = "4h"
	TimeFrame1d  TimeFrame = "1d"
)

// KlineData K线数据
type KlineData struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// TimeFrameAnalysis 单周期分析结果
type TimeFrameAnalysis struct {
	TimeFrame      TimeFrame   `json:"timeFrame"`
	TrendDirection int         `json:"trendDirection"` // -2强空, -1弱空, 0震荡, 1弱多, 2强多
	TrendStrength  float64     `json:"trendStrength"`  // 趋势强度 0-100
	Volatility     float64     `json:"volatility"`     // 波动率
	RSI            float64     `json:"rsi"`            // RSI值
	MACD           MACDResult  `json:"macd"`           // MACD
	MA             MAResult    `json:"ma"`             // 均线
	ATR            float64     `json:"atr"`            // ATR
	Volume         VolumeState `json:"volume"`         // 成交量状态
	State          MarketState `json:"state"`          // 该周期市场状态
	Score          float64     `json:"score"`          // 综合评分 -100 到 100
}

// MACDResult MACD结果
type MACDResult struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
	CrossUp   bool    `json:"crossUp"`   // 金叉
	CrossDown bool    `json:"crossDown"` // 死叉
}

// MAResult 均线结果
type MAResult struct {
	MA5        float64 `json:"ma5"`
	MA10       float64 `json:"ma10"`
	MA20       float64 `json:"ma20"`
	MA60       float64 `json:"ma60"`
	PriceAbove bool    `json:"priceAbove"` // 价格在均线上方
	MABullish  bool    `json:"maBullish"`  // 均线多头排列
}

// VolumeState 成交量状态
type VolumeState struct {
	Current     float64 `json:"current"`
	Average     float64 `json:"average"`
	Ratio       float64 `json:"ratio"`       // 当前/平均
	IsIncreased bool    `json:"isIncreased"` // 是否放量
}

// MarketAnalysisResult 市场分析综合结果
type MarketAnalysisResult struct {
	Symbol            string                        `json:"symbol"`
	FinalState        MarketState                   `json:"finalState"`        // 最终市场状态
	FinalDirection    int                           `json:"finalDirection"`    // 最终方向 -2到2
	TrendScore        float64                       `json:"trendScore"`        // 趋势评分 -100到100
	VolatilityLevel   string                        `json:"volatilityLevel"`   // 波动等级
	Confidence        float64                       `json:"confidence"`        // 置信度 0-1
	TimeFrameAnalysis map[TimeFrame]*TimeFrameAnalysis `json:"timeFrameAnalysis"` // 各周期分析
	SignalStrength    float64                       `json:"signalStrength"`    // 信号强度 0-100
	SuggestAction     string                        `json:"suggestAction"`     // 建议操作
	Reasons           []string                      `json:"reasons"`           // 分析理由
	Timestamp         time.Time                     `json:"timestamp"`
}

// MultiTimeFrameAnalyzer 多时间周期分析器
type MultiTimeFrameAnalyzer struct {
	// 时间周期权重
	weights map[TimeFrame]float64
	// 技术指标权重
	indicatorWeights struct {
		trend      float64 // 趋势权重 (MA)
		momentum   float64 // 动量权重 (RSI, MACD)
		volatility float64 // 波动权重 (ATR, 布林带)
		volume     float64 // 成交量权重
	}
}

// NewMultiTimeFrameAnalyzer 创建多时间周期分析器
func NewMultiTimeFrameAnalyzer() *MultiTimeFrameAnalyzer {
	analyzer := &MultiTimeFrameAnalyzer{
		weights: map[TimeFrame]float64{
			TimeFrame1m:  0.10, // 1分钟 10%
			TimeFrame5m:  0.15, // 5分钟 15%
			TimeFrame15m: 0.25, // 15分钟 25%
			TimeFrame30m: 0.25, // 30分钟 25%
			TimeFrame1h:  0.25, // 1小时 25%
		},
	}
	analyzer.indicatorWeights.trend = 0.35
	analyzer.indicatorWeights.momentum = 0.30
	analyzer.indicatorWeights.volatility = 0.20
	analyzer.indicatorWeights.volume = 0.15
	return analyzer
}

// Analyze 分析市场状态
func (a *MultiTimeFrameAnalyzer) Analyze(ctx context.Context, symbol string, klineData map[TimeFrame][]KlineData) *MarketAnalysisResult {
	result := &MarketAnalysisResult{
		Symbol:            symbol,
		TimeFrameAnalysis: make(map[TimeFrame]*TimeFrameAnalysis),
		Reasons:           make([]string, 0),
		Timestamp:         time.Now(),
	}

	// 1. 分析各个时间周期
	totalScore := 0.0
	totalWeight := 0.0
	avgVolatility := 0.0

	for tf, klines := range klineData {
		if len(klines) < 60 { // 至少需要60根K线
			continue
		}

		analysis := a.analyzeTimeFrame(tf, klines)
		result.TimeFrameAnalysis[tf] = analysis

		weight := a.weights[tf]
		totalScore += analysis.Score * weight
		totalWeight += weight
		avgVolatility += analysis.Volatility * weight
	}

	if totalWeight > 0 {
		result.TrendScore = totalScore / totalWeight
		avgVolatility = avgVolatility / totalWeight
	}

	// 2. 综合判断最终市场状态
	result.FinalState, result.FinalDirection = a.determineFinalState(result.TrendScore, avgVolatility)

	// 3. 计算置信度
	result.Confidence = a.calculateConfidence(result.TimeFrameAnalysis)

	// 4. 确定波动等级
	result.VolatilityLevel = a.determineVolatilityLevel(avgVolatility)

	// 5. 计算信号强度
	result.SignalStrength = a.calculateSignalStrength(result)

	// 6. 生成建议
	result.SuggestAction = a.generateSuggestAction(result)

	// 7. 生成分析理由
	a.generateReasons(result)

	return result
}

// analyzeTimeFrame 分析单个时间周期
func (a *MultiTimeFrameAnalyzer) analyzeTimeFrame(tf TimeFrame, klines []KlineData) *TimeFrameAnalysis {
	analysis := &TimeFrameAnalysis{
		TimeFrame: tf,
	}

	closes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	volumes := make([]float64, len(klines))

	for i, k := range klines {
		closes[i] = k.Close
		highs[i] = k.High
		lows[i] = k.Low
		volumes[i] = k.Volume
	}

	currentPrice := closes[len(closes)-1]

	// 计算技术指标
	analysis.RSI = a.calculateRSI(closes, 14)
	analysis.MACD = a.calculateMACD(closes, 12, 26, 9)
	analysis.MA = a.calculateMA(closes, currentPrice)
	analysis.ATR = a.calculateATR(highs, lows, closes, 14)
	analysis.Volatility = (analysis.ATR / currentPrice) * 100
	analysis.Volume = a.calculateVolumeState(volumes)

	// 计算趋势评分
	trendScore := a.calculateTrendScore(analysis, currentPrice)
	momentumScore := a.calculateMomentumScore(analysis)
	volatilityScore := a.calculateVolatilityInfluence(analysis.Volatility)
	volumeScore := a.calculateVolumeScore(analysis.Volume)

	// 综合评分
	analysis.Score = trendScore*a.indicatorWeights.trend +
		momentumScore*a.indicatorWeights.momentum +
		volatilityScore*a.indicatorWeights.volatility +
		volumeScore*a.indicatorWeights.volume

	// 确定趋势方向和强度
	analysis.TrendDirection, analysis.TrendStrength = a.determineTrend(analysis.Score)

	// 确定该周期市场状态
	analysis.State = a.determineTimeFrameState(analysis)

	return analysis
}

// calculateRSI 计算RSI
func (a *MultiTimeFrameAnalyzer) calculateRSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 50
	}

	gains := 0.0
	losses := 0.0

	for i := len(closes) - period; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// calculateMACD 计算MACD
func (a *MultiTimeFrameAnalyzer) calculateMACD(closes []float64, fast, slow, signal int) MACDResult {
	result := MACDResult{}

	if len(closes) < slow+signal {
		return result
	}

	// 计算EMA
	emaFast := a.calculateEMA(closes, fast)
	emaSlow := a.calculateEMA(closes, slow)

	macdLine := make([]float64, len(closes))
	for i := 0; i < len(closes); i++ {
		macdLine[i] = emaFast[i] - emaSlow[i]
	}

	signalLine := a.calculateEMA(macdLine, signal)

	lastIdx := len(closes) - 1
	result.MACD = macdLine[lastIdx]
	result.Signal = signalLine[lastIdx]
	result.Histogram = result.MACD - result.Signal

	// 判断金叉死叉
	if lastIdx > 0 {
		prevMACD := macdLine[lastIdx-1]
		prevSignal := signalLine[lastIdx-1]
		result.CrossUp = prevMACD < prevSignal && result.MACD > result.Signal
		result.CrossDown = prevMACD > prevSignal && result.MACD < result.Signal
	}

	return result
}

// calculateEMA 计算EMA
func (a *MultiTimeFrameAnalyzer) calculateEMA(data []float64, period int) []float64 {
	ema := make([]float64, len(data))
	multiplier := 2.0 / float64(period+1)

	// 初始EMA使用SMA
	sum := 0.0
	for i := 0; i < period && i < len(data); i++ {
		sum += data[i]
	}
	ema[period-1] = sum / float64(period)

	// 计算后续EMA
	for i := period; i < len(data); i++ {
		ema[i] = (data[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

// calculateMA 计算均线
func (a *MultiTimeFrameAnalyzer) calculateMA(closes []float64, currentPrice float64) MAResult {
	result := MAResult{}

	if len(closes) >= 5 {
		result.MA5 = a.calculateSMA(closes, 5)
	}
	if len(closes) >= 10 {
		result.MA10 = a.calculateSMA(closes, 10)
	}
	if len(closes) >= 20 {
		result.MA20 = a.calculateSMA(closes, 20)
	}
	if len(closes) >= 60 {
		result.MA60 = a.calculateSMA(closes, 60)
	}

	// 判断价格位置
	result.PriceAbove = currentPrice > result.MA20

	// 判断均线排列
	result.MABullish = result.MA5 > result.MA10 && result.MA10 > result.MA20

	return result
}

// calculateSMA 计算SMA
func (a *MultiTimeFrameAnalyzer) calculateSMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	sum := 0.0
	for i := len(data) - period; i < len(data); i++ {
		sum += data[i]
	}
	return sum / float64(period)
}

// calculateATR 计算ATR
func (a *MultiTimeFrameAnalyzer) calculateATR(highs, lows, closes []float64, period int) float64 {
	if len(highs) < period+1 {
		return 0
	}

	tr := make([]float64, len(highs))
	for i := 1; i < len(highs); i++ {
		hl := highs[i] - lows[i]
		hc := math.Abs(highs[i] - closes[i-1])
		lc := math.Abs(lows[i] - closes[i-1])
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	// 计算ATR
	sum := 0.0
	for i := len(tr) - period; i < len(tr); i++ {
		sum += tr[i]
	}
	return sum / float64(period)
}

// calculateVolumeState 计算成交量状态
func (a *MultiTimeFrameAnalyzer) calculateVolumeState(volumes []float64) VolumeState {
	state := VolumeState{}

	if len(volumes) < 20 {
		return state
	}

	state.Current = volumes[len(volumes)-1]

	// 计算20周期平均成交量
	sum := 0.0
	for i := len(volumes) - 20; i < len(volumes)-1; i++ {
		sum += volumes[i]
	}
	state.Average = sum / 19

	if state.Average > 0 {
		state.Ratio = state.Current / state.Average
		state.IsIncreased = state.Ratio > 1.5
	}

	return state
}

// calculateTrendScore 计算趋势评分
func (a *MultiTimeFrameAnalyzer) calculateTrendScore(analysis *TimeFrameAnalysis, currentPrice float64) float64 {
	score := 0.0

	// 均线位置 (-50 到 50)
	if analysis.MA.PriceAbove {
		score += 25
	} else {
		score -= 25
	}

	// 均线排列 (-25 到 25)
	if analysis.MA.MABullish {
		score += 25
	} else if analysis.MA.MA5 < analysis.MA.MA10 && analysis.MA.MA10 < analysis.MA.MA20 {
		score -= 25 // 空头排列
	}

	return score
}

// calculateMomentumScore 计算动量评分
func (a *MultiTimeFrameAnalyzer) calculateMomentumScore(analysis *TimeFrameAnalysis) float64 {
	score := 0.0

	// RSI评分 (-50 到 50)
	if analysis.RSI > 70 {
		score -= 30 // 超买
	} else if analysis.RSI > 50 {
		score += (analysis.RSI - 50) * 1.5
	} else if analysis.RSI < 30 {
		score += 30 // 超卖可能反弹
	} else {
		score -= (50 - analysis.RSI) * 1.5
	}

	// MACD评分 (-50 到 50)
	if analysis.MACD.CrossUp {
		score += 30
	} else if analysis.MACD.CrossDown {
		score -= 30
	} else if analysis.MACD.Histogram > 0 {
		score += 15
	} else {
		score -= 15
	}

	return score
}

// calculateVolatilityInfluence 计算波动率影响
func (a *MultiTimeFrameAnalyzer) calculateVolatilityInfluence(volatility float64) float64 {
	// 低波动返回中性，高波动可能带来机会但也有风险
	if volatility < 1 {
		return 10 // 低波动，稳定
	} else if volatility < 2 {
		return 0 // 正常波动
	} else if volatility < 4 {
		return -10 // 较高波动
	}
	return -25 // 极高波动，风险大
}

// calculateVolumeScore 计算成交量评分
func (a *MultiTimeFrameAnalyzer) calculateVolumeScore(volume VolumeState) float64 {
	if volume.IsIncreased {
		return 20 // 放量确认趋势
	}
	if volume.Ratio < 0.5 {
		return -10 // 缩量，趋势可能减弱
	}
	return 0
}

// determineTrend 确定趋势方向和强度
func (a *MultiTimeFrameAnalyzer) determineTrend(score float64) (direction int, strength float64) {
	strength = math.Abs(score)
	if score > 50 {
		return 2, strength // 强多
	} else if score > 20 {
		return 1, strength // 弱多
	} else if score < -50 {
		return -2, strength // 强空
	} else if score < -20 {
		return -1, strength // 弱空
	}
	return 0, strength // 震荡
}

// determineTimeFrameState 确定周期市场状态
func (a *MultiTimeFrameAnalyzer) determineTimeFrameState(analysis *TimeFrameAnalysis) MarketState {
	if analysis.Volatility > 4 {
		return MarketHighVolatility
	}
	if analysis.Volatility < 0.5 {
		return MarketLowVolatility
	}

	switch analysis.TrendDirection {
	case 2:
		return MarketStrongUptrend
	case 1:
		return MarketMildUptrend
	case -1:
		return MarketMildDowntrend
	case -2:
		return MarketStrongDowntrend
	default:
		return MarketRanging
	}
}

// determineFinalState 确定最终市场状态
func (a *MultiTimeFrameAnalyzer) determineFinalState(score float64, volatility float64) (MarketState, int) {
	// 先判断波动性
	if volatility > 4 {
		if score > 30 {
			return MarketHighVolatility, 1
		} else if score < -30 {
			return MarketHighVolatility, -1
		}
		return MarketHighVolatility, 0
	}

	if volatility < 0.5 {
		return MarketLowVolatility, 0
	}

	// 根据评分判断趋势
	if score > 50 {
		return MarketStrongUptrend, 2
	} else if score > 20 {
		return MarketMildUptrend, 1
	} else if score < -50 {
		return MarketStrongDowntrend, -2
	} else if score < -20 {
		return MarketMildDowntrend, -1
	}
	return MarketRanging, 0
}

// calculateConfidence 计算置信度
func (a *MultiTimeFrameAnalyzer) calculateConfidence(analyses map[TimeFrame]*TimeFrameAnalysis) float64 {
	if len(analyses) == 0 {
		return 0
	}

	// 计算各周期方向一致性
	directions := make([]int, 0)
	for _, analysis := range analyses {
		directions = append(directions, analysis.TrendDirection)
	}

	// 统计方向
	bullCount := 0
	bearCount := 0
	for _, d := range directions {
		if d > 0 {
			bullCount++
		} else if d < 0 {
			bearCount++
		}
	}

	total := len(directions)
	maxCount := bullCount
	if bearCount > maxCount {
		maxCount = bearCount
	}

	// 一致性越高，置信度越高
	consistency := float64(maxCount) / float64(total)
	return consistency
}

// determineVolatilityLevel 确定波动等级
func (a *MultiTimeFrameAnalyzer) determineVolatilityLevel(volatility float64) string {
	if volatility < 1 {
		return "LOW"
	} else if volatility < 2 {
		return "NORMAL"
	} else if volatility < 4 {
		return "HIGH"
	}
	return "EXTREME"
}

// calculateSignalStrength 计算信号强度
func (a *MultiTimeFrameAnalyzer) calculateSignalStrength(result *MarketAnalysisResult) float64 {
	// 综合趋势评分和置信度
	baseStrength := math.Abs(result.TrendScore)
	return baseStrength * result.Confidence
}

// generateSuggestAction 生成建议操作
func (a *MultiTimeFrameAnalyzer) generateSuggestAction(result *MarketAnalysisResult) string {
	if result.Confidence < 0.5 {
		return "WAIT" // 信号不明确，等待
	}

	if result.VolatilityLevel == "EXTREME" {
		return "CAUTION" // 极端波动，谨慎
	}

	switch result.FinalState {
	case MarketStrongUptrend:
		return "STRONG_BUY"
	case MarketMildUptrend:
		return "BUY"
	case MarketStrongDowntrend:
		return "STRONG_SELL"
	case MarketMildDowntrend:
		return "SELL"
	case MarketHighVolatility:
		return "CAUTION"
	default:
		return "WAIT"
	}
}

// generateReasons 生成分析理由
func (a *MultiTimeFrameAnalyzer) generateReasons(result *MarketAnalysisResult) {
	// 趋势分析
	switch result.FinalState {
	case MarketStrongUptrend:
		result.Reasons = append(result.Reasons, "多周期确认强势上涨趋势")
	case MarketMildUptrend:
		result.Reasons = append(result.Reasons, "市场呈温和上涨态势")
	case MarketStrongDowntrend:
		result.Reasons = append(result.Reasons, "多周期确认强势下跌趋势")
	case MarketMildDowntrend:
		result.Reasons = append(result.Reasons, "市场呈温和下跌态势")
	case MarketRanging:
		result.Reasons = append(result.Reasons, "市场处于震荡区间，无明显趋势")
	case MarketHighVolatility:
		result.Reasons = append(result.Reasons, "市场波动剧烈，风险较高")
	case MarketLowVolatility:
		result.Reasons = append(result.Reasons, "市场波动较低，可能即将变盘")
	}

	// 置信度分析
	if result.Confidence > 0.8 {
		result.Reasons = append(result.Reasons, "各周期方向高度一致，信号可靠")
	} else if result.Confidence < 0.5 {
		result.Reasons = append(result.Reasons, "各周期信号存在分歧，建议观望")
	}

	// 各周期详情
	for tf, analysis := range result.TimeFrameAnalysis {
		if analysis.MACD.CrossUp {
			result.Reasons = append(result.Reasons, string(tf)+"周期MACD金叉")
		}
		if analysis.MACD.CrossDown {
			result.Reasons = append(result.Reasons, string(tf)+"周期MACD死叉")
		}
		if analysis.RSI > 70 {
			result.Reasons = append(result.Reasons, string(tf)+"周期RSI超买")
		}
		if analysis.RSI < 30 {
			result.Reasons = append(result.Reasons, string(tf)+"周期RSI超卖")
		}
	}
}

