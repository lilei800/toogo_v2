// Package strategy 策略分析库
package strategy

import (
	"context"
	"math"
	"time"
)

// RiskPreferenceType 风险偏好类型
type RiskPreferenceType string

const (
	RiskConservative RiskPreferenceType = "conservative" // 保守型
	RiskBalanced     RiskPreferenceType = "balanced"     // 平衡型
	RiskAggressive   RiskPreferenceType = "aggressive"   // 激进型
)

// RiskFactors 风险因素输入
type RiskFactors struct {
	AccountBalance    float64 // 账户余额
	AvailableBalance  float64 // 可用余额
	CurrentPnL        float64 // 当前持仓盈亏
	TodayPnL          float64 // 今日盈亏
	TotalPnL          float64 // 累计盈亏
	ProfitTarget      float64 // 盈利目标
	MaxLossLimit      float64 // 最大亏损限制
	WinCount          int     // 连续盈利次数
	LossCount         int     // 连续亏损次数
	TotalTrades       int     // 总交易次数
	WinTrades         int     // 盈利交易次数
	MarketVolatility  float64 // 市场波动率 (ATR/价格 百分比)
	MarketState       string  // 市场状态
	CurrentLeverage   int     // 当前杠杆
	MaxLeverage       int     // 最大杠杆
	PositionRatio     float64 // 当前仓位占比
}

// RiskPreferenceResult 风险偏好判断结果
type RiskPreferenceResult struct {
	PreferenceType    RiskPreferenceType `json:"preferenceType"`    // 风险偏好类型
	WinProbability    float64            `json:"winProbability"`    // 胜算概率 0-100
	SuggestLeverage   int                `json:"suggestLeverage"`   // 建议杠杆
	SuggestPosition   float64            `json:"suggestPosition"`   // 建议仓位比例 0-1
	SuggestStopLoss   float64            `json:"suggestStopLoss"`   // 建议止损比例
	SuggestTakeProfit float64            `json:"suggestTakeProfit"` // 建议止盈比例
	RiskScore         float64            `json:"riskScore"`         // 风险评分 0-100
	Confidence        float64            `json:"confidence"`        // 置信度 0-1
	Reasons           []string           `json:"reasons"`           // 判断理由
	Timestamp         time.Time          `json:"timestamp"`         // 时间戳
}

// RiskPreferenceAnalyzer 风险偏好分析器
type RiskPreferenceAnalyzer struct {
	// 权重配置
	weights struct {
		accountHealth    float64 // 账户健康度权重
		marketCondition  float64 // 市场状况权重
		tradingHistory   float64 // 交易历史权重
		profitProgress   float64 // 盈利进度权重
		volatilityFactor float64 // 波动率因子权重
	}
}

// NewRiskPreferenceAnalyzer 创建风险偏好分析器
func NewRiskPreferenceAnalyzer() *RiskPreferenceAnalyzer {
	analyzer := &RiskPreferenceAnalyzer{}
	// 设置默认权重
	analyzer.weights.accountHealth = 0.25
	analyzer.weights.marketCondition = 0.20
	analyzer.weights.tradingHistory = 0.20
	analyzer.weights.profitProgress = 0.20
	analyzer.weights.volatilityFactor = 0.15
	return analyzer
}

// Analyze 分析风险偏好
func (a *RiskPreferenceAnalyzer) Analyze(ctx context.Context, factors RiskFactors) *RiskPreferenceResult {
	result := &RiskPreferenceResult{
		Timestamp: time.Now(),
		Reasons:   make([]string, 0),
	}

	// 1. 计算账户健康度评分 (0-100)
	accountScore := a.calculateAccountHealthScore(factors)

	// 2. 计算市场状况评分 (0-100)
	marketScore := a.calculateMarketConditionScore(factors)

	// 3. 计算交易历史评分 (0-100)
	historyScore := a.calculateTradingHistoryScore(factors)

	// 4. 计算盈利进度评分 (0-100)
	progressScore := a.calculateProfitProgressScore(factors)

	// 5. 计算波动率因子评分 (0-100)
	volatilityScore := a.calculateVolatilityScore(factors)

	// 6. 综合评分
	totalScore := accountScore*a.weights.accountHealth +
		marketScore*a.weights.marketCondition +
		historyScore*a.weights.tradingHistory +
		progressScore*a.weights.profitProgress +
		volatilityScore*a.weights.volatilityFactor

	result.RiskScore = totalScore

	// 7. 计算胜算概率
	result.WinProbability = a.calculateWinProbability(factors, totalScore)

	// 8. 确定风险偏好类型
	result.PreferenceType, result.Confidence = a.determinePreferenceType(totalScore, factors)

	// 9. 生成建议参数
	a.generateSuggestions(result, factors)

	// 10. 生成判断理由
	a.generateReasons(result, factors, accountScore, marketScore, historyScore, progressScore, volatilityScore)

	return result
}

// calculateAccountHealthScore 计算账户健康度评分
func (a *RiskPreferenceAnalyzer) calculateAccountHealthScore(factors RiskFactors) float64 {
	score := 50.0

	// 可用余额占比
	if factors.AccountBalance > 0 {
		availableRatio := factors.AvailableBalance / factors.AccountBalance
		score += (availableRatio - 0.5) * 40 // 50%可用为中性
	}

	// 当前盈亏影响
	if factors.AccountBalance > 0 {
		pnlRatio := factors.CurrentPnL / factors.AccountBalance
		if pnlRatio > 0 {
			score += math.Min(pnlRatio*100, 20) // 盈利加分，最多20分
		} else {
			score += math.Max(pnlRatio*100, -30) // 亏损减分，最多减30分
		}
	}

	// 仓位占比影响
	if factors.PositionRatio > 0.8 {
		score -= 20 // 仓位过重减分
	} else if factors.PositionRatio < 0.3 {
		score += 10 // 仓位较轻加分
	}

	return math.Max(0, math.Min(100, score))
}

// calculateMarketConditionScore 计算市场状况评分
func (a *RiskPreferenceAnalyzer) calculateMarketConditionScore(factors RiskFactors) float64 {
	score := 50.0

	// 根据市场状态调整
	switch factors.MarketState {
	case "STRONG_UPTREND", "STRONG_DOWNTREND":
		score += 20 // 强趋势有利于顺势交易
	case "MILD_UPTREND", "MILD_DOWNTREND":
		score += 10
	case "RANGING":
		score -= 10 // 震荡市场风险较高
	case "HIGH_VOLATILITY":
		score -= 20 // 高波动风险高
	case "LOW_VOLATILITY":
		score += 5 // 低波动相对安全
	}

	return math.Max(0, math.Min(100, score))
}

// calculateTradingHistoryScore 计算交易历史评分
func (a *RiskPreferenceAnalyzer) calculateTradingHistoryScore(factors RiskFactors) float64 {
	score := 50.0

	// 胜率影响
	if factors.TotalTrades > 0 {
		winRate := float64(factors.WinTrades) / float64(factors.TotalTrades)
		score += (winRate - 0.5) * 60 // 50%胜率为中性
	}

	// 连续盈亏影响
	if factors.WinCount > 3 {
		score += math.Min(float64(factors.WinCount)*3, 15) // 连胜加分
	}
	if factors.LossCount > 2 {
		score -= math.Min(float64(factors.LossCount)*5, 25) // 连亏减分
	}

	return math.Max(0, math.Min(100, score))
}

// calculateProfitProgressScore 计算盈利进度评分
func (a *RiskPreferenceAnalyzer) calculateProfitProgressScore(factors RiskFactors) float64 {
	score := 50.0

	if factors.ProfitTarget > 0 {
		// 今日盈利进度
		todayProgress := factors.TodayPnL / factors.ProfitTarget
		if todayProgress >= 1 {
			score += 30 // 已达成目标
		} else if todayProgress >= 0.5 {
			score += 20 // 完成50%以上
		} else if todayProgress >= 0 {
			score += todayProgress * 20
		} else {
			score += math.Max(todayProgress*40, -30) // 亏损减分
		}
	}

	// 距离最大亏损的距离
	if factors.MaxLossLimit > 0 && factors.TodayPnL < 0 {
		lossRatio := math.Abs(factors.TodayPnL) / factors.MaxLossLimit
		if lossRatio > 0.8 {
			score -= 30 // 接近最大亏损，大幅减分
		} else if lossRatio > 0.5 {
			score -= 15
		}
	}

	return math.Max(0, math.Min(100, score))
}

// calculateVolatilityScore 计算波动率评分
func (a *RiskPreferenceAnalyzer) calculateVolatilityScore(factors RiskFactors) float64 {
	score := 50.0

	// 波动率影响 (正常波动率约为1-3%)
	if factors.MarketVolatility < 1 {
		score += 15 // 低波动
	} else if factors.MarketVolatility < 2 {
		score += 5 // 正常波动
	} else if factors.MarketVolatility < 4 {
		score -= 10 // 较高波动
	} else {
		score -= 25 // 极高波动
	}

	return math.Max(0, math.Min(100, score))
}

// calculateWinProbability 计算胜算概率
func (a *RiskPreferenceAnalyzer) calculateWinProbability(factors RiskFactors, riskScore float64) float64 {
	baseProb := 50.0

	// 基于风险评分调整
	baseProb += (riskScore - 50) * 0.5

	// 基于历史胜率调整
	if factors.TotalTrades > 10 {
		historyWinRate := float64(factors.WinTrades) / float64(factors.TotalTrades) * 100
		baseProb = baseProb*0.6 + historyWinRate*0.4
	}

	// 基于市场状态调整
	switch factors.MarketState {
	case "STRONG_UPTREND", "STRONG_DOWNTREND":
		baseProb += 10 // 强趋势更容易判断方向
	case "HIGH_VOLATILITY":
		baseProb -= 15 // 高波动降低确定性
	}

	// 连续亏损降低概率
	if factors.LossCount > 3 {
		baseProb -= float64(factors.LossCount) * 3
	}

	return math.Max(20, math.Min(85, baseProb))
}

// determinePreferenceType 确定风险偏好类型
func (a *RiskPreferenceAnalyzer) determinePreferenceType(score float64, factors RiskFactors) (RiskPreferenceType, float64) {
	confidence := 0.7

	// 强制保守条件
	if factors.LossCount > 4 || 
		(factors.MaxLossLimit > 0 && factors.TodayPnL < -factors.MaxLossLimit*0.7) ||
		factors.PositionRatio > 0.9 {
		return RiskConservative, 0.9
	}

	// 强制激进条件（已接近盈利目标且状态良好）
	if factors.ProfitTarget > 0 && factors.TodayPnL >= factors.ProfitTarget*0.8 && score > 70 {
		return RiskAggressive, 0.8
	}

	// 根据评分判断
	if score >= 70 {
		confidence = 0.6 + (score-70)/100
		return RiskAggressive, math.Min(confidence, 0.95)
	} else if score >= 45 {
		confidence = 0.5 + math.Abs(score-57.5)/50
		return RiskBalanced, math.Min(confidence, 0.9)
	} else {
		confidence = 0.6 + (45-score)/100
		return RiskConservative, math.Min(confidence, 0.95)
	}
}

// generateSuggestions 生成建议参数
func (a *RiskPreferenceAnalyzer) generateSuggestions(result *RiskPreferenceResult, factors RiskFactors) {
	switch result.PreferenceType {
	case RiskConservative:
		result.SuggestLeverage = min(5, factors.MaxLeverage)
		result.SuggestPosition = 0.1
		result.SuggestStopLoss = 3.0
		result.SuggestTakeProfit = 20.0
	case RiskBalanced:
		result.SuggestLeverage = min(10, factors.MaxLeverage)
		result.SuggestPosition = 0.2
		result.SuggestStopLoss = 5.0
		result.SuggestTakeProfit = 30.0
	case RiskAggressive:
		result.SuggestLeverage = min(20, factors.MaxLeverage)
		result.SuggestPosition = 0.3
		result.SuggestStopLoss = 8.0
		result.SuggestTakeProfit = 50.0
	}

	// 根据波动率调整止损
	if factors.MarketVolatility > 3 {
		result.SuggestStopLoss *= 1.5
	}
}

// generateReasons 生成判断理由
func (a *RiskPreferenceAnalyzer) generateReasons(result *RiskPreferenceResult, factors RiskFactors,
	accountScore, marketScore, historyScore, progressScore, volatilityScore float64) {

	// 账户状况
	if accountScore >= 70 {
		result.Reasons = append(result.Reasons, "账户状况良好，可用资金充足")
	} else if accountScore < 40 {
		result.Reasons = append(result.Reasons, "账户可用资金紧张，建议降低风险")
	}

	// 市场状况
	if marketScore >= 70 {
		result.Reasons = append(result.Reasons, "市场趋势明显，适合顺势交易")
	} else if marketScore < 40 {
		result.Reasons = append(result.Reasons, "市场状态不稳定，建议谨慎操作")
	}

	// 交易历史
	if historyScore >= 70 {
		result.Reasons = append(result.Reasons, "近期交易表现良好，胜率较高")
	} else if historyScore < 40 {
		result.Reasons = append(result.Reasons, "近期交易表现不佳，建议减少交易频率")
	}

	// 连续亏损
	if factors.LossCount > 2 {
		result.Reasons = append(result.Reasons, "连续亏损中，建议暂停或减仓")
	}

	// 盈利进度
	if factors.ProfitTarget > 0 && factors.TodayPnL >= factors.ProfitTarget {
		result.Reasons = append(result.Reasons, "已达成今日盈利目标，建议止盈")
	}

	// 波动率
	if factors.MarketVolatility > 4 {
		result.Reasons = append(result.Reasons, "市场波动剧烈，建议扩大止损或观望")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

