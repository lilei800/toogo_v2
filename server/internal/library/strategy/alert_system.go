// Package strategy 策略分析库 - 预警日志系统
package strategy

import (
	"context"
	"fmt"
	"math"
	"time"
)

// AlertLevel 预警级别
type AlertLevel string

const (
	AlertInfo     AlertLevel = "INFO"     // 信息
	AlertWarning  AlertLevel = "WARNING"  // 警告
	AlertDanger   AlertLevel = "DANGER"   // 危险
	AlertCritical AlertLevel = "CRITICAL" // 紧急
)

// AlertType 预警类型
type AlertType string

const (
	AlertTypeMarket    AlertType = "MARKET"    // 市场状态预警
	AlertTypeRisk      AlertType = "RISK"      // 风险偏好预警
	AlertTypeDirection AlertType = "DIRECTION" // 下单方向预警
)

// BaseAlert 基础预警结构
type BaseAlert struct {
	ID        int64      `json:"id"`
	RobotID   int64      `json:"robotId"`
	Symbol    string     `json:"symbol"`
	AlertType AlertType  `json:"alertType"`
	Level     AlertLevel `json:"level"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Data      any        `json:"data"`      // 详细数据
	IsRead    bool       `json:"isRead"`    // 是否已读
	CreatedAt time.Time  `json:"createdAt"`
}

// MarketStateAlert 市场状态预警日志
type MarketStateAlert struct {
	BaseAlert
	PreviousState     MarketState `json:"previousState"`     // 之前的市场状态
	CurrentState      MarketState `json:"currentState"`      // 当前市场状态
	TrendScore        float64     `json:"trendScore"`        // 趋势评分
	Volatility        float64     `json:"volatility"`        // 波动率
	VolatilityLevel   string      `json:"volatilityLevel"`   // 波动等级
	Confidence        float64     `json:"confidence"`        // 置信度
	SuggestAction     string      `json:"suggestAction"`     // 建议操作
	TimeFrameSignals  map[string]string `json:"timeFrameSignals"` // 各周期信号
	TechnicalSummary  string      `json:"technicalSummary"`  // 技术面总结
	Recommendation    string      `json:"recommendation"`    // 操作建议
}

// RiskPreferenceAlert 风险偏好预警日志
type RiskPreferenceAlert struct {
	BaseAlert
	PreviousPreference RiskPreferenceType `json:"previousPreference"` // 之前的风险偏好
	CurrentPreference  RiskPreferenceType `json:"currentPreference"`  // 当前风险偏好
	WinProbability     float64            `json:"winProbability"`     // 胜算概率
	RiskScore          float64            `json:"riskScore"`          // 风险评分
	AccountHealth      float64            `json:"accountHealth"`      // 账户健康度
	SuggestLeverage    int                `json:"suggestLeverage"`    // 建议杠杆
	SuggestPosition    float64            `json:"suggestPosition"`    // 建议仓位
	SuggestStopLoss    float64            `json:"suggestStopLoss"`    // 建议止损
	Reasons            []string           `json:"reasons"`            // 原因
	ActionRequired     string             `json:"actionRequired"`     // 需要的操作
}

// OrderDirectionAlert 下单方向预警日志
type OrderDirectionAlert struct {
	BaseAlert
	Direction          string  `json:"direction"`          // 建议方向: LONG/SHORT/WAIT
	DirectionScore     float64 `json:"directionScore"`     // 方向评分 (-100 到 100)
	SignalStrength     float64 `json:"signalStrength"`     // 信号强度 (0-100)
	EntryPrice         float64 `json:"entryPrice"`         // 建议入场价
	StopLossPrice      float64 `json:"stopLossPrice"`      // 建议止损价
	TakeProfitPrice    float64 `json:"takeProfitPrice"`    // 建议止盈价
	RiskRewardRatio    float64 `json:"riskRewardRatio"`    // 风险收益比
	TimeWindow         string  `json:"timeWindow"`         // 时间窗口
	VolatilityPoints   float64 `json:"volatilityPoints"`   // 波动点数
	Confidence         float64 `json:"confidence"`         // 置信度
	MarketCondition    string  `json:"marketCondition"`    // 市场状况
	TechnicalSignals   []string `json:"technicalSignals"` // 技术信号
	Recommendation     string  `json:"recommendation"`     // 操作建议
}

// AlertSystem 预警系统
type AlertSystem struct {
	riskAnalyzer   *RiskPreferenceAnalyzer
	marketAnalyzer *MultiTimeFrameAnalyzer
	
	// 缓存上一次状态用于比较
	lastMarketState     map[int64]MarketState
	lastRiskPreference  map[int64]RiskPreferenceType
	
	// 回调函数
	onMarketAlert    func(ctx context.Context, alert *MarketStateAlert) error
	onRiskAlert      func(ctx context.Context, alert *RiskPreferenceAlert) error
	onDirectionAlert func(ctx context.Context, alert *OrderDirectionAlert) error
}

// NewAlertSystem 创建预警系统
func NewAlertSystem() *AlertSystem {
	return &AlertSystem{
		riskAnalyzer:       NewRiskPreferenceAnalyzer(),
		marketAnalyzer:     NewMultiTimeFrameAnalyzer(),
		lastMarketState:    make(map[int64]MarketState),
		lastRiskPreference: make(map[int64]RiskPreferenceType),
	}
}

// SetMarketAlertCallback 设置市场预警回调
func (s *AlertSystem) SetMarketAlertCallback(fn func(ctx context.Context, alert *MarketStateAlert) error) {
	s.onMarketAlert = fn
}

// SetRiskAlertCallback 设置风险预警回调
func (s *AlertSystem) SetRiskAlertCallback(fn func(ctx context.Context, alert *RiskPreferenceAlert) error) {
	s.onRiskAlert = fn
}

// SetDirectionAlertCallback 设置方向预警回调
func (s *AlertSystem) SetDirectionAlertCallback(fn func(ctx context.Context, alert *OrderDirectionAlert) error) {
	s.onDirectionAlert = fn
}

// AnalyzeAndAlert 分析并生成预警
func (s *AlertSystem) AnalyzeAndAlert(ctx context.Context, robotID int64, symbol string, 
	klineData map[TimeFrame][]KlineData, riskFactors RiskFactors, currentPrice float64) (*TradingSignal, error) {
	
	signal := &TradingSignal{
		RobotID:   robotID,
		Symbol:    symbol,
		Timestamp: time.Now(),
	}
	
	// 1. 市场状态分析
	marketResult := s.marketAnalyzer.Analyze(ctx, symbol, klineData)
	signal.MarketAnalysis = marketResult
	
	// 生成市场预警
	marketAlert := s.generateMarketAlert(robotID, symbol, marketResult)
	if marketAlert != nil && s.onMarketAlert != nil {
		s.onMarketAlert(ctx, marketAlert)
	}
	signal.MarketAlert = marketAlert
	
	// 更新市场波动率到风险因素
	riskFactors.MarketState = string(marketResult.FinalState)
	riskFactors.MarketVolatility = s.getAverageVolatility(marketResult)
	
	// 2. 风险偏好分析
	riskResult := s.riskAnalyzer.Analyze(ctx, riskFactors)
	signal.RiskAnalysis = riskResult
	
	// 生成风险预警
	riskAlert := s.generateRiskAlert(robotID, symbol, riskResult)
	if riskAlert != nil && s.onRiskAlert != nil {
		s.onRiskAlert(ctx, riskAlert)
	}
	signal.RiskAlert = riskAlert
	
	// 3. 综合分析下单方向
	directionAlert := s.generateDirectionAlert(robotID, symbol, currentPrice, marketResult, riskResult)
	if directionAlert != nil && s.onDirectionAlert != nil {
		s.onDirectionAlert(ctx, directionAlert)
	}
	signal.DirectionAlert = directionAlert
	
	// 4. 生成最终交易信号
	signal.FinalDirection = directionAlert.Direction
	signal.FinalConfidence = directionAlert.Confidence
	signal.ShouldTrade = s.shouldTrade(marketResult, riskResult, directionAlert)
	
	return signal, nil
}

// TradingSignal 交易信号综合结果
type TradingSignal struct {
	RobotID         int64                  `json:"robotId"`
	Symbol          string                 `json:"symbol"`
	Timestamp       time.Time              `json:"timestamp"`
	MarketAnalysis  *MarketAnalysisResult  `json:"marketAnalysis"`
	RiskAnalysis    *RiskPreferenceResult  `json:"riskAnalysis"`
	MarketAlert     *MarketStateAlert      `json:"marketAlert"`
	RiskAlert       *RiskPreferenceAlert   `json:"riskAlert"`
	DirectionAlert  *OrderDirectionAlert   `json:"directionAlert"`
	FinalDirection  string                 `json:"finalDirection"`  // LONG/SHORT/WAIT
	FinalConfidence float64                `json:"finalConfidence"` // 最终置信度
	ShouldTrade     bool                   `json:"shouldTrade"`     // 是否应该交易
}

// generateMarketAlert 生成市场状态预警
func (s *AlertSystem) generateMarketAlert(robotID int64, symbol string, result *MarketAnalysisResult) *MarketStateAlert {
	previousState := s.lastMarketState[robotID]
	currentState := result.FinalState
	
	// 更新缓存
	s.lastMarketState[robotID] = currentState
	
	// 确定预警级别
	level := s.determineMarketAlertLevel(previousState, currentState, result)
	
	alert := &MarketStateAlert{
		BaseAlert: BaseAlert{
			RobotID:   robotID,
			Symbol:    symbol,
			AlertType: AlertTypeMarket,
			Level:     level,
			Title:     s.getMarketAlertTitle(currentState),
			Message:   s.getMarketAlertMessage(previousState, currentState, result),
			CreatedAt: time.Now(),
		},
		PreviousState:    previousState,
		CurrentState:     currentState,
		TrendScore:       result.TrendScore,
		Volatility:       s.getAverageVolatility(result),
		VolatilityLevel:  result.VolatilityLevel,
		Confidence:       result.Confidence,
		SuggestAction:    result.SuggestAction,
		TimeFrameSignals: s.formatTimeFrameSignals(result.TimeFrameAnalysis),
		TechnicalSummary: s.generateTechnicalSummary(result),
		Recommendation:   s.getMarketRecommendation(result),
	}
	
	return alert
}

// generateRiskAlert 生成风险偏好预警
func (s *AlertSystem) generateRiskAlert(robotID int64, symbol string, result *RiskPreferenceResult) *RiskPreferenceAlert {
	previousPreference := s.lastRiskPreference[robotID]
	currentPreference := result.PreferenceType
	
	// 更新缓存
	s.lastRiskPreference[robotID] = currentPreference
	
	// 确定预警级别
	level := s.determineRiskAlertLevel(previousPreference, currentPreference, result)
	
	alert := &RiskPreferenceAlert{
		BaseAlert: BaseAlert{
			RobotID:   robotID,
			Symbol:    symbol,
			AlertType: AlertTypeRisk,
			Level:     level,
			Title:     s.getRiskAlertTitle(currentPreference),
			Message:   s.getRiskAlertMessage(previousPreference, currentPreference, result),
			CreatedAt: time.Now(),
		},
		PreviousPreference: previousPreference,
		CurrentPreference:  currentPreference,
		WinProbability:     result.WinProbability,
		RiskScore:          result.RiskScore,
		SuggestLeverage:    result.SuggestLeverage,
		SuggestPosition:    result.SuggestPosition,
		SuggestStopLoss:    result.SuggestStopLoss,
		Reasons:            result.Reasons,
		ActionRequired:     s.getRiskActionRequired(result),
	}
	
	return alert
}

// generateDirectionAlert 生成下单方向预警
func (s *AlertSystem) generateDirectionAlert(robotID int64, symbol string, currentPrice float64,
	marketResult *MarketAnalysisResult, riskResult *RiskPreferenceResult) *OrderDirectionAlert {
	
	// 综合分析方向
	direction, score := s.calculateDirection(marketResult, riskResult)
	signalStrength := s.calculateSignalStrength(marketResult, riskResult)
	confidence := s.calculateDirectionConfidence(marketResult, riskResult, signalStrength)
	
	// 计算入场点位
	stopLoss, takeProfit := s.calculateEntryPoints(currentPrice, direction, riskResult, marketResult)
	
	// 计算风险收益比
	riskRewardRatio := s.calculateRiskRewardRatio(currentPrice, stopLoss, takeProfit, direction)
	
	// 确定时间窗口
	timeWindow := s.determineTimeWindow(marketResult)
	
	// 计算波动点数
	volatilityPoints := s.calculateVolatilityPoints(marketResult, currentPrice)
	
	// 确定预警级别
	level := s.determineDirectionAlertLevel(signalStrength, confidence)
	
	alert := &OrderDirectionAlert{
		BaseAlert: BaseAlert{
			RobotID:   robotID,
			Symbol:    symbol,
			AlertType: AlertTypeDirection,
			Level:     level,
			Title:     s.getDirectionAlertTitle(direction, signalStrength),
			Message:   s.getDirectionAlertMessage(direction, score, confidence),
			CreatedAt: time.Now(),
		},
		Direction:        direction,
		DirectionScore:   score,
		SignalStrength:   signalStrength,
		EntryPrice:       currentPrice,
		StopLossPrice:    stopLoss,
		TakeProfitPrice:  takeProfit,
		RiskRewardRatio:  riskRewardRatio,
		TimeWindow:       timeWindow,
		VolatilityPoints: volatilityPoints,
		Confidence:       confidence,
		MarketCondition:  string(marketResult.FinalState),
		TechnicalSignals: s.collectTechnicalSignals(marketResult),
		Recommendation:   s.getDirectionRecommendation(direction, confidence, riskResult),
	}
	
	return alert
}

// calculateDirection 计算交易方向
func (s *AlertSystem) calculateDirection(market *MarketAnalysisResult, risk *RiskPreferenceResult) (string, float64) {
	score := market.TrendScore
	
	// 根据风险偏好调整
	if risk.PreferenceType == RiskConservative {
		// 保守型需要更强信号
		if score > 0 && score < 40 {
			return "WAIT", score
		}
		if score < 0 && score > -40 {
			return "WAIT", score
		}
	}
	
	// 高波动市场
	if market.VolatilityLevel == "EXTREME" {
		return "WAIT", score
	}
	
	// 确定方向
	if score > 30 {
		return "LONG", score
	} else if score < -30 {
		return "SHORT", score
	}
	return "WAIT", score
}

// calculateSignalStrength 计算信号强度
func (s *AlertSystem) calculateSignalStrength(market *MarketAnalysisResult, risk *RiskPreferenceResult) float64 {
	// 基础强度来自市场分析
	baseStrength := market.SignalStrength
	
	// 根据风险偏好调整
	switch risk.PreferenceType {
	case RiskAggressive:
		baseStrength *= 1.1 // 激进型提高信号强度
	case RiskConservative:
		baseStrength *= 0.8 // 保守型降低信号强度
	}
	
	// 根据胜算概率调整
	baseStrength *= (risk.WinProbability / 100)
	
	return math.Min(100, baseStrength)
}

// calculateDirectionConfidence 计算方向置信度
func (s *AlertSystem) calculateDirectionConfidence(market *MarketAnalysisResult, risk *RiskPreferenceResult, strength float64) float64 {
	// 综合市场置信度和风险置信度
	confidence := market.Confidence * 0.6 + risk.Confidence * 0.4
	
	// 信号强度影响
	confidence *= (strength / 100)
	
	return math.Min(0.95, confidence)
}

// calculateEntryPoints 计算入场点位
func (s *AlertSystem) calculateEntryPoints(currentPrice float64, direction string, 
	risk *RiskPreferenceResult, market *MarketAnalysisResult) (stopLoss, takeProfit float64) {
	
	stopLossPercent := risk.SuggestStopLoss / 100
	takeProfitPercent := risk.SuggestTakeProfit / 100
	
	// 根据波动率调整
	if market.VolatilityLevel == "HIGH" || market.VolatilityLevel == "EXTREME" {
		stopLossPercent *= 1.5
		takeProfitPercent *= 1.3
	}
	
	if direction == "LONG" {
		stopLoss = currentPrice * (1 - stopLossPercent)
		takeProfit = currentPrice * (1 + takeProfitPercent)
	} else if direction == "SHORT" {
		stopLoss = currentPrice * (1 + stopLossPercent)
		takeProfit = currentPrice * (1 - takeProfitPercent)
	} else {
		stopLoss = 0
		takeProfit = 0
	}
	
	return
}

// calculateRiskRewardRatio 计算风险收益比
func (s *AlertSystem) calculateRiskRewardRatio(currentPrice, stopLoss, takeProfit float64, direction string) float64 {
	if stopLoss == 0 || takeProfit == 0 {
		return 0
	}
	
	var risk, reward float64
	if direction == "LONG" {
		risk = currentPrice - stopLoss
		reward = takeProfit - currentPrice
	} else {
		risk = stopLoss - currentPrice
		reward = currentPrice - takeProfit
	}
	
	if risk == 0 {
		return 0
	}
	return reward / risk
}

// determineTimeWindow 确定时间窗口
func (s *AlertSystem) determineTimeWindow(market *MarketAnalysisResult) string {
	switch market.VolatilityLevel {
	case "EXTREME":
		return "1-5分钟"
	case "HIGH":
		return "5-15分钟"
	case "NORMAL":
		return "15-60分钟"
	default:
		return "1-4小时"
	}
}

// calculateVolatilityPoints 计算波动点数
func (s *AlertSystem) calculateVolatilityPoints(market *MarketAnalysisResult, currentPrice float64) float64 {
	// 使用15分钟ATR作为基准
	if analysis, ok := market.TimeFrameAnalysis[TimeFrame15m]; ok {
		return analysis.ATR
	}
	// 估算波动点数
	volatility := s.getAverageVolatility(market)
	return currentPrice * (volatility / 100)
}

// shouldTrade 判断是否应该交易
func (s *AlertSystem) shouldTrade(market *MarketAnalysisResult, risk *RiskPreferenceResult, direction *OrderDirectionAlert) bool {
	// 方向为等待则不交易
	if direction.Direction == "WAIT" {
		return false
	}
	
	// 信号强度太弱
	if direction.SignalStrength < 30 {
		return false
	}
	
	// 置信度太低
	if direction.Confidence < 0.5 {
		return false
	}
	
	// 风险收益比太低
	if direction.RiskRewardRatio < 1.5 {
		return false
	}
	
	// 保守型需要更高标准
	if risk.PreferenceType == RiskConservative {
		if direction.SignalStrength < 50 || direction.Confidence < 0.7 {
			return false
		}
	}
	
	return true
}

// 辅助方法

func (s *AlertSystem) getAverageVolatility(result *MarketAnalysisResult) float64 {
	total := 0.0
	count := 0
	for _, analysis := range result.TimeFrameAnalysis {
		total += analysis.Volatility
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func (s *AlertSystem) determineMarketAlertLevel(prev, curr MarketState, result *MarketAnalysisResult) AlertLevel {
	// 状态变化
	if prev != curr {
		if curr == MarketHighVolatility || curr == MarketStrongDowntrend {
			return AlertDanger
		}
		return AlertWarning
	}
	
	// 极端波动
	if result.VolatilityLevel == "EXTREME" {
		return AlertCritical
	}
	
	return AlertInfo
}

func (s *AlertSystem) determineRiskAlertLevel(prev, curr RiskPreferenceType, result *RiskPreferenceResult) AlertLevel {
	// 风险偏好变化
	if prev != curr {
		if curr == RiskConservative {
			return AlertWarning
		}
	}
	
	// 胜算概率过低
	if result.WinProbability < 40 {
		return AlertDanger
	}
	
	return AlertInfo
}

func (s *AlertSystem) determineDirectionAlertLevel(strength float64, confidence float64) AlertLevel {
	if strength > 70 && confidence > 0.8 {
		return AlertInfo // 强信号
	}
	if strength > 50 && confidence > 0.6 {
		return AlertInfo
	}
	if strength < 30 || confidence < 0.4 {
		return AlertWarning // 弱信号
	}
	return AlertInfo
}

func (s *AlertSystem) getMarketAlertTitle(state MarketState) string {
	titles := map[MarketState]string{
		MarketStrongUptrend:   "🚀 强势上涨",
		MarketMildUptrend:     "📈 温和上涨",
		MarketRanging:         "↔️ 震荡整理",
		MarketMildDowntrend:   "📉 温和下跌",
		MarketStrongDowntrend: "⚠️ 强势下跌",
		MarketHighVolatility:  "⚡ 高波动市场",
		MarketLowVolatility:   "😴 低波动市场",
	}
	return titles[state]
}

func (s *AlertSystem) getMarketAlertMessage(prev, curr MarketState, result *MarketAnalysisResult) string {
	if prev != curr && prev != "" {
		return fmt.Sprintf("市场状态从 %s 转变为 %s，趋势评分: %.1f，置信度: %.1f%%",
			prev, curr, result.TrendScore, result.Confidence*100)
	}
	return fmt.Sprintf("当前市场状态: %s，趋势评分: %.1f，置信度: %.1f%%",
		curr, result.TrendScore, result.Confidence*100)
}

func (s *AlertSystem) getRiskAlertTitle(pref RiskPreferenceType) string {
	titles := map[RiskPreferenceType]string{
		RiskConservative: "🛡️ 保守型策略",
		RiskBalanced:     "⚖️ 平衡型策略",
		RiskAggressive:   "🔥 激进型策略",
	}
	return titles[pref]
}

func (s *AlertSystem) getRiskAlertMessage(prev, curr RiskPreferenceType, result *RiskPreferenceResult) string {
	if prev != curr && prev != "" {
		return fmt.Sprintf("风险偏好从 %s 调整为 %s，胜算概率: %.1f%%，风险评分: %.1f",
			prev, curr, result.WinProbability, result.RiskScore)
	}
	return fmt.Sprintf("当前风险偏好: %s，胜算概率: %.1f%%，建议仓位: %.0f%%",
		curr, result.WinProbability, result.SuggestPosition*100)
}

func (s *AlertSystem) getRiskActionRequired(result *RiskPreferenceResult) string {
	switch result.PreferenceType {
	case RiskConservative:
		return "建议减仓或暂停交易，等待更好时机"
	case RiskBalanced:
		return "正常交易，按建议参数执行"
	case RiskAggressive:
		return "可适当加仓，但注意控制风险"
	}
	return ""
}

func (s *AlertSystem) getDirectionAlertTitle(direction string, strength float64) string {
	strengthText := ""
	if strength > 70 {
		strengthText = "强"
	} else if strength > 50 {
		strengthText = "中"
	} else {
		strengthText = "弱"
	}
	
	switch direction {
	case "LONG":
		return fmt.Sprintf("🟢 %s做多信号", strengthText)
	case "SHORT":
		return fmt.Sprintf("🔴 %s做空信号", strengthText)
	default:
		return "⏸️ 建议观望"
	}
}

func (s *AlertSystem) getDirectionAlertMessage(direction string, score, confidence float64) string {
	if direction == "WAIT" {
		return "当前信号不明确，建议等待更好的入场时机"
	}
	dirText := "做多"
	if direction == "SHORT" {
		dirText = "做空"
	}
	return fmt.Sprintf("建议%s，方向评分: %.1f，置信度: %.1f%%", dirText, score, confidence*100)
}

func (s *AlertSystem) getDirectionRecommendation(direction string, confidence float64, risk *RiskPreferenceResult) string {
	if direction == "WAIT" {
		return "暂不建议开仓，继续观察市场动态"
	}
	
	dirText := "做多"
	if direction == "SHORT" {
		dirText = "做空"
	}
	
	return fmt.Sprintf("建议%s，使用%dx杠杆，仓位%.0f%%，止损%.1f%%，止盈%.1f%%",
		dirText, risk.SuggestLeverage, risk.SuggestPosition*100,
		risk.SuggestStopLoss, risk.SuggestTakeProfit)
}

func (s *AlertSystem) formatTimeFrameSignals(analyses map[TimeFrame]*TimeFrameAnalysis) map[string]string {
	signals := make(map[string]string)
	for tf, analysis := range analyses {
		signal := "震荡"
		if analysis.TrendDirection > 0 {
			signal = "看多"
		} else if analysis.TrendDirection < 0 {
			signal = "看空"
		}
		signals[string(tf)] = signal
	}
	return signals
}

func (s *AlertSystem) generateTechnicalSummary(result *MarketAnalysisResult) string {
	summary := fmt.Sprintf("综合趋势评分: %.1f，", result.TrendScore)
	summary += fmt.Sprintf("波动等级: %s，", result.VolatilityLevel)
	summary += fmt.Sprintf("各周期一致性: %.0f%%", result.Confidence*100)
	return summary
}

func (s *AlertSystem) getMarketRecommendation(result *MarketAnalysisResult) string {
	switch result.SuggestAction {
	case "STRONG_BUY":
		return "强烈建议做多"
	case "BUY":
		return "建议做多"
	case "STRONG_SELL":
		return "强烈建议做空"
	case "SELL":
		return "建议做空"
	case "CAUTION":
		return "市场波动大，谨慎操作"
	default:
		return "建议观望，等待明确信号"
	}
}

func (s *AlertSystem) collectTechnicalSignals(result *MarketAnalysisResult) []string {
	signals := make([]string, 0)
	for tf, analysis := range result.TimeFrameAnalysis {
		if analysis.MACD.CrossUp {
			signals = append(signals, fmt.Sprintf("%s MACD金叉", tf))
		}
		if analysis.MACD.CrossDown {
			signals = append(signals, fmt.Sprintf("%s MACD死叉", tf))
		}
		if analysis.RSI > 70 {
			signals = append(signals, fmt.Sprintf("%s RSI超买(%.1f)", tf, analysis.RSI))
		}
		if analysis.RSI < 30 {
			signals = append(signals, fmt.Sprintf("%s RSI超卖(%.1f)", tf, analysis.RSI))
		}
		if analysis.MA.MABullish {
			signals = append(signals, fmt.Sprintf("%s 均线多头排列", tf))
		}
	}
	return signals
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

