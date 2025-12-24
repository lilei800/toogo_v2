// Package market 风险评估服务
// 负责计算胜算概率和风险偏好判定
package market

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// RiskEvaluator 风险评估器（单例）
type RiskEvaluator struct {
	mu sync.RWMutex

	// 风险评估缓存 key: robotId
	evaluations map[int64]*RiskEvaluation

	// 风险偏好预警日志
	riskAlerts map[int64][]*RiskPreferenceAlert

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// RiskEvaluation 风险评估结果
type RiskEvaluation struct {
	RobotId   int64
	UpdatedAt time.Time

	// 胜算概率 0-100
	WinProbability float64

	// 风险偏好
	RiskPreference RiskPreference

	// 评估维度分数
	MarketScore   float64 // 市场状态评分 0-100
	TechnicalScore float64 // 技术指标评分 0-100
	AccountScore  float64 // 账户状况评分 0-100
	HistoryScore  float64 // 历史表现评分 0-100
	VolatilityRisk float64 // 波动风险 0-100

	// 建议参数
	SuggestedLeverage     int     // 建议杠杆
	SuggestedMarginPercent float64 // 建议保证金比例
	SuggestedStopLoss     float64 // 建议止损比例
	SuggestedTakeProfit   float64 // 建议止盈比例

	// 风险等级 1-5
	RiskLevel int

	// 评估说明
	Reason string
}

// RiskPreference 风险偏好枚举
type RiskPreference string

const (
	RiskConservative RiskPreference = "conservative" // 保守型
	RiskBalanced     RiskPreference = "balanced"     // 平衡型
	RiskAggressive   RiskPreference = "aggressive"   // 激进型
)

// RiskPreferenceAlert 风险偏好预警
type RiskPreferenceAlert struct {
	Timestamp      time.Time
	RobotId        int64
	PrevPreference RiskPreference
	NewPreference  RiskPreference
	WinProbability float64
	Reason         string
	Factors        map[string]float64
}

// RobotContext 机器人上下文（用于风险评估）
type RobotContext struct {
	RobotId       int64
	UserId        int64
	Platform      string
	Symbol        string
	
	// 账户信息
	AccountBalance    float64
	AvailableBalance  float64
	UnrealizedPnl     float64
	
	// 机器人配置
	MaxProfitTarget float64
	MaxLossAmount   float64
	CurrentProfit   float64
	
	// 历史表现
	TotalTrades   int
	WinTrades     int
	LossTrades    int
	AvgProfit     float64
	AvgLoss       float64
	MaxDrawdown   float64
	
	// 当前持仓
	HasPosition   bool
	PositionSide  string
	PositionPnl   float64
}

var (
	riskEvaluator     *RiskEvaluator
	riskEvaluatorOnce sync.Once
)

// GetRiskEvaluator 获取风险评估器单例
func GetRiskEvaluator() *RiskEvaluator {
	riskEvaluatorOnce.Do(func() {
		riskEvaluator = &RiskEvaluator{
			evaluations: make(map[int64]*RiskEvaluation),
			riskAlerts:  make(map[int64][]*RiskPreferenceAlert),
			stopCh:      make(chan struct{}),
		}
	})
	return riskEvaluator
}

// EvaluateRisk 评估风险（主入口）
func (r *RiskEvaluator) EvaluateRisk(ctx context.Context, robotCtx *RobotContext) *RiskEvaluation {
	// 获取市场分析结果
	analysis := GetMarketAnalyzer().GetAnalysis(robotCtx.Platform, robotCtx.Symbol)
	if analysis == nil {
		// 没有市场分析数据，返回保守评估
		return r.defaultConservativeEvaluation(robotCtx)
	}

	eval := &RiskEvaluation{
		RobotId:   robotCtx.RobotId,
		UpdatedAt: time.Now(),
	}

	// 1. 评估市场状态得分
	eval.MarketScore = r.evaluateMarketScore(analysis)

	// 2. 评估技术指标得分
	eval.TechnicalScore = r.evaluateTechnicalScore(analysis)

	// 3. 评估账户状况得分
	eval.AccountScore = r.evaluateAccountScore(robotCtx)

	// 4. 评估历史表现得分
	eval.HistoryScore = r.evaluateHistoryScore(robotCtx)

	// 5. 评估波动风险
	eval.VolatilityRisk = r.evaluateVolatilityRisk(analysis)

	// 6. 计算综合胜算概率
	eval.WinProbability = r.calculateWinProbability(eval)

	// 7. 判定风险偏好
	eval.RiskPreference = r.determineRiskPreference(eval, robotCtx, analysis)

	// 8. 生成建议参数
	r.generateSuggestedParams(eval, robotCtx, analysis)

	// 9. 生成评估说明
	eval.Reason = r.generateEvaluationReason(eval, robotCtx, analysis)

	// 保存评估结果
	r.mu.Lock()
	oldEval := r.evaluations[robotCtx.RobotId]
	r.evaluations[robotCtx.RobotId] = eval
	
	// 检查风险偏好变化，生成预警
	if oldEval != nil && oldEval.RiskPreference != eval.RiskPreference {
		r.generateRiskAlert(ctx, robotCtx.RobotId, oldEval.RiskPreference, eval)
	}
	r.mu.Unlock()

	return eval
}

// evaluateMarketScore 评估市场状态得分
func (r *RiskEvaluator) evaluateMarketScore(analysis *MarketAnalysis) float64 {
	score := 50.0 // 基础分

	// 根据市场状态调整
	switch analysis.MarketState {
	case MarketStateTrend:
		// 趋势市场得分较高
		score += 30 * analysis.MarketStateConf
	case MarketStateVolatile:
		// 震荡市场中等
		score += 10
	case MarketStateHighVol:
		// 高波动风险较高，得分降低
		score -= 20
	case MarketStateLowVol:
		// 低波动较安全
		score += 20
	}

	// 趋势一致性加分
	consistency := GetMarketAnalyzer().calculateTrendConsistency(analysis.TimeframeAnalysis)
	score += consistency * 20

	return math.Min(100, math.Max(0, score))
}

// evaluateTechnicalScore 评估技术指标得分（精简版）
func (r *RiskEvaluator) evaluateTechnicalScore(analysis *MarketAnalysis) float64 {
	if analysis.Indicators == nil {
		return 50
	}

	score := 50.0

	// 趋势评分影响（明确趋势有利于交易）
	trendScore := analysis.Indicators.TrendScore
	if trendScore > 50 {
		score += 20 // 明确上涨趋势
	} else if trendScore < -50 {
		score += 15 // 明确下跌趋势（有利于做空）
	} else if math.Abs(trendScore) < 20 {
		score -= 10 // 趋势不明确，风险增加
	}

	return math.Min(100, math.Max(0, score))
}

// evaluateAccountScore 评估账户状况得分
func (r *RiskEvaluator) evaluateAccountScore(ctx *RobotContext) float64 {
	score := 50.0

	// 可用余额比例
	if ctx.AccountBalance > 0 {
		availableRatio := ctx.AvailableBalance / ctx.AccountBalance
		if availableRatio > 0.7 {
			score += 30
		} else if availableRatio > 0.5 {
			score += 15
		} else if availableRatio < 0.3 {
			score -= 20
		}
	}

	// 当前盈亏状况
	if ctx.MaxProfitTarget > 0 {
		progressToTarget := ctx.CurrentProfit / ctx.MaxProfitTarget
		if progressToTarget > 0.8 {
			score += 20 // 接近盈利目标
		} else if progressToTarget < -0.5 {
			score -= 30 // 亏损较大
		}
	}

	// 未实现盈亏
	if ctx.UnrealizedPnl < 0 && ctx.AccountBalance > 0 {
		lossRatio := math.Abs(ctx.UnrealizedPnl) / ctx.AccountBalance
		if lossRatio > 0.1 {
			score -= 30
		}
	}

	return math.Min(100, math.Max(0, score))
}

// evaluateHistoryScore 评估历史表现得分
func (r *RiskEvaluator) evaluateHistoryScore(ctx *RobotContext) float64 {
	if ctx.TotalTrades < 5 {
		return 50 // 交易次数不足，返回中性分数
	}

	score := 50.0

	// 胜率影响
	winRate := float64(ctx.WinTrades) / float64(ctx.TotalTrades)
	if winRate > 0.6 {
		score += 30
	} else if winRate > 0.5 {
		score += 15
	} else if winRate < 0.4 {
		score -= 20
	}

	// 盈亏比影响
	if ctx.AvgLoss != 0 {
		profitFactor := math.Abs(ctx.AvgProfit) / math.Abs(ctx.AvgLoss)
		if profitFactor > 2 {
			score += 20
		} else if profitFactor > 1.5 {
			score += 10
		} else if profitFactor < 1 {
			score -= 15
		}
	}

	// 最大回撤影响
	if ctx.MaxDrawdown > 20 {
		score -= 20
	} else if ctx.MaxDrawdown > 10 {
		score -= 10
	}

	return math.Min(100, math.Max(0, score))
}

// evaluateVolatilityRisk 评估波动风险
func (r *RiskEvaluator) evaluateVolatilityRisk(analysis *MarketAnalysis) float64 {
	// 基于ATR和波动率计算风险
	risk := 30.0 // 基础风险

	if analysis.Volatility > 2 {
		risk += 40
	} else if analysis.Volatility > 1 {
		risk += 20
	} else if analysis.Volatility < 0.5 {
		risk -= 10
	}

	// 高波动市场状态
	if analysis.MarketState == MarketStateHighVol {
		risk += 20
	}

	return math.Min(100, math.Max(0, risk))
}

// calculateWinProbability 计算胜算概率
func (r *RiskEvaluator) calculateWinProbability(eval *RiskEvaluation) float64 {
	// 加权计算胜算概率
	weights := map[string]float64{
		"market":     0.25,
		"technical":  0.25,
		"account":    0.20,
		"history":    0.20,
		"volatility": 0.10, // 波动风险是负向影响
	}

	probability := eval.MarketScore*weights["market"] +
		eval.TechnicalScore*weights["technical"] +
		eval.AccountScore*weights["account"] +
		eval.HistoryScore*weights["history"] -
		eval.VolatilityRisk*weights["volatility"]

	return math.Min(100, math.Max(0, probability))
}

// determineRiskPreference 判定风险偏好
func (r *RiskEvaluator) determineRiskPreference(eval *RiskEvaluation, ctx *RobotContext, analysis *MarketAnalysis) RiskPreference {
	// 基于胜算概率判定
	if eval.WinProbability >= 70 {
		// 高胜算，可以激进
		eval.RiskLevel = 1
		return RiskAggressive
	} else if eval.WinProbability >= 50 {
		// 中等胜算，平衡操作
		eval.RiskLevel = 2
		return RiskBalanced
	} else {
		// 低胜算，保守操作
		eval.RiskLevel = 3
		return RiskConservative
	}
}

// generateSuggestedParams 生成建议参数
func (r *RiskEvaluator) generateSuggestedParams(eval *RiskEvaluation, ctx *RobotContext, analysis *MarketAnalysis) {
	switch eval.RiskPreference {
	case RiskAggressive:
		eval.SuggestedLeverage = 20
		eval.SuggestedMarginPercent = 15
		eval.SuggestedStopLoss = 5
		eval.SuggestedTakeProfit = 15
	case RiskBalanced:
		eval.SuggestedLeverage = 10
		eval.SuggestedMarginPercent = 10
		eval.SuggestedStopLoss = 8
		eval.SuggestedTakeProfit = 12
	case RiskConservative:
		eval.SuggestedLeverage = 5
		eval.SuggestedMarginPercent = 5
		eval.SuggestedStopLoss = 10
		eval.SuggestedTakeProfit = 8
	}

	// 根据波动率调整
	if eval.VolatilityRisk > 60 {
		eval.SuggestedLeverage = int(float64(eval.SuggestedLeverage) * 0.7)
		eval.SuggestedMarginPercent *= 0.8
		eval.SuggestedStopLoss *= 1.2
	}
}

// generateEvaluationReason 生成评估说明
func (r *RiskEvaluator) generateEvaluationReason(eval *RiskEvaluation, ctx *RobotContext, analysis *MarketAnalysis) string {
	reasons := []string{}

	// 市场状态
	switch analysis.MarketState {
	case MarketStateTrend:
		reasons = append(reasons, "市场处于趋势状态，适合顺势操作")
	case MarketStateVolatile:
		reasons = append(reasons, "市场处于震荡状态，建议区间操作")
	case MarketStateHighVol:
		reasons = append(reasons, "市场高波动，建议降低仓位")
	case MarketStateLowVol:
		reasons = append(reasons, "市场低波动，可适当增加仓位")
	}

	// 胜算概率
	if eval.WinProbability >= 70 {
		reasons = append(reasons, "综合胜算较高")
	} else if eval.WinProbability < 40 {
		reasons = append(reasons, "综合胜算偏低，建议谨慎")
	}

	// 账户状况
	if eval.AccountScore < 40 {
		reasons = append(reasons, "账户风险敞口较大")
	}

	if len(reasons) == 0 {
		return "市场环境正常，建议按标准参数操作"
	}

	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += "；" + reasons[i]
	}
	return result
}

// generateRiskAlert 生成风险偏好预警
func (r *RiskEvaluator) generateRiskAlert(ctx context.Context, robotId int64, prevPref RiskPreference, eval *RiskEvaluation) {
	alert := &RiskPreferenceAlert{
		Timestamp:      time.Now(),
		RobotId:        robotId,
		PrevPreference: prevPref,
		NewPreference:  eval.RiskPreference,
		WinProbability: eval.WinProbability,
		Reason:         eval.Reason,
		Factors: map[string]float64{
			"market_score":    eval.MarketScore,
			"technical_score": eval.TechnicalScore,
			"account_score":   eval.AccountScore,
			"history_score":   eval.HistoryScore,
			"volatility_risk": eval.VolatilityRisk,
		},
	}

	// 保存预警日志（保留最近50条）
	if len(r.riskAlerts[robotId]) >= 50 {
		r.riskAlerts[robotId] = r.riskAlerts[robotId][1:]
	}
	r.riskAlerts[robotId] = append(r.riskAlerts[robotId], alert)

	// 写入数据库
	GetAlertLogger().LogRiskPreference(&RiskPreferenceLogEntry{
		RobotId:        robotId,
		PrevPreference: string(prevPref),
		NewPreference:  string(eval.RiskPreference),
		WinProbability: eval.WinProbability,
		MarketScore:    eval.MarketScore,
		TechnicalScore: eval.TechnicalScore,
		AccountScore:   eval.AccountScore,
		HistoryScore:   eval.HistoryScore,
		VolatilityRisk: eval.VolatilityRisk,
		RiskLevel:      eval.RiskLevel,
		Reason:         eval.Reason,
	})

	g.Log().Infof(ctx, "[RiskPreferenceAlert] RobotId=%d %s->%s, 胜算=%.2f%%, 原因=%s",
		robotId, prevPref, eval.RiskPreference, eval.WinProbability, eval.Reason)
}

// GetEvaluation 获取风险评估结果
func (r *RiskEvaluator) GetEvaluation(robotId int64) *RiskEvaluation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.evaluations[robotId]
}

// GetRiskAlerts 获取风险偏好预警日志
func (r *RiskEvaluator) GetRiskAlerts(robotId int64, limit int) []*RiskPreferenceAlert {
	r.mu.RLock()
	defer r.mu.RUnlock()

	alerts := r.riskAlerts[robotId]
	if len(alerts) <= limit {
		return alerts
	}
	return alerts[len(alerts)-limit:]
}

// defaultConservativeEvaluation 默认保守评估
func (r *RiskEvaluator) defaultConservativeEvaluation(ctx *RobotContext) *RiskEvaluation {
	return &RiskEvaluation{
		RobotId:                ctx.RobotId,
		UpdatedAt:              time.Now(),
		WinProbability:         30,
		RiskPreference:         RiskConservative,
		MarketScore:            30,
		TechnicalScore:         30,
		AccountScore:           50,
		HistoryScore:           50,
		VolatilityRisk:         50,
		SuggestedLeverage:      5,
		SuggestedMarginPercent: 3,
		SuggestedStopLoss:      15,
		SuggestedTakeProfit:    10,
		RiskLevel:              4,
		Reason:                 "市场数据不足，采用保守策略",
	}
}

