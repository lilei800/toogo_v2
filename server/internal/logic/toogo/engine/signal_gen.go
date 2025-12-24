// Package engine 机器人引擎模块 - 信号生成器
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// RobotSignalGen 机器人信号生成器
type RobotSignalGen struct {
	engine *RobotEngine
}

// NewRobotSignalGen 创建信号生成器
func NewRobotSignalGen(engine *RobotEngine) *RobotSignalGen {
	return &RobotSignalGen{engine: engine}
}

// Generate 生成方向信号
func (s *RobotSignalGen) Generate(ctx context.Context) *Signal {
	// 核心：评估窗口价格信号
	windowSignal := s.EvaluateWindowSignal(ctx)
	if windowSignal == nil {
		return &Signal{
			Timestamp:  time.Now(),
			Direction:  "NEUTRAL",
			SignalType: "none",
			Reason:     "等待数据...",
		}
	}

	// 窗口信号有明确方向时，直接使用
	if windowSignal.Direction != "NEUTRAL" {
		analysis := s.engine.LastAnalysis
		if analysis != nil && analysis.TrendDirection != "neutral" {
			// 方向一致，增强置信度
			if (windowSignal.Direction == "LONG" && analysis.TrendDirection == "up") ||
				(windowSignal.Direction == "SHORT" && analysis.TrendDirection == "down") {
				windowSignal.Confidence = math.Min(100, windowSignal.Confidence*1.1)
			}
		}
		return windowSignal
	}

	return windowSignal
}

// EvaluateWindowSignal 评估窗口信号
func (s *RobotSignalGen) EvaluateWindowSignal(ctx context.Context) *Signal {
	pw := s.engine.PriceWin
	if pw == nil || pw.Config == nil {
		return nil
	}

	signal := &Signal{
		Timestamp:       time.Now(),
		SignalType:      "window",
		SignalThreshold: pw.Config.Threshold,
	}

	stats := pw.GetStats()
	if stats.DataCount == 0 {
		signal.Direction = "NEUTRAL"
		signal.Reason = "等待价格数据..."
		return signal
	}

	if stats.DataCount == 1 {
		signal.Direction = "NEUTRAL"
		signal.CurrentPrice = stats.CurrentPrice
		signal.Reason = fmt.Sprintf("已获取初始价格 %.2f，等待更多数据...", signal.CurrentPrice)
		return signal
	}

	minPrice := stats.MinPrice
	maxPrice := stats.MaxPrice
	currentPrice := stats.CurrentPrice
	threshold := pw.Config.Threshold

	// 计算距离
	distanceFromMax := maxPrice - currentPrice
	distanceFromMin := currentPrice - minPrice

	// 触发条件
	shortTriggered := distanceFromMax >= threshold
	longTriggered := distanceFromMin >= threshold

	// 填充信号数据
	signal.WindowMinPrice = minPrice
	signal.WindowMaxPrice = maxPrice
	signal.CurrentPrice = currentPrice
	signal.DistanceFromMin = distanceFromMin
	signal.DistanceFromMax = distanceFromMax

	// 检测窗口基准价变化
	minChanged, maxChanged := pw.CheckWindowChange(minPrice, maxPrice)
	if minChanged {
		pw.SetAlertedLong(0) // 重置
	}
	if maxChanged {
		pw.SetAlertedShort(0) // 重置
	}
	pw.UpdateLastWindow(minPrice, maxPrice)

	// 权重计算
	const (
		weightWindow = 0.50
		weightMACD   = 0.25
		weightEMA    = 0.25
	)

	var longScore, shortScore float64
	var reasonParts []string

	// 1. 窗口信号得分 (50%)
	if longTriggered {
		longScore += weightWindow * 100
		reasonParts = append(reasonParts, fmt.Sprintf("窗口↑(距低%.0f)", distanceFromMin))
	}
	if shortTriggered {
		shortScore += weightWindow * 100
		reasonParts = append(reasonParts, fmt.Sprintf("窗口↓(距高%.0f)", distanceFromMax))
	}

	// 2. MACD和EMA得分 (各25%)
	macdLong, macdShort, emaLong, emaShort := s.getIndicatorSignals()

	if macdLong {
		longScore += weightMACD * 100
		reasonParts = append(reasonParts, "MACD↑")
	} else if macdShort {
		shortScore += weightMACD * 100
		reasonParts = append(reasonParts, "MACD↓")
	}

	if emaLong {
		longScore += weightEMA * 100
		reasonParts = append(reasonParts, "EMA↑")
	} else if emaShort {
		shortScore += weightEMA * 100
		reasonParts = append(reasonParts, "EMA↓")
	}

	// 判断是否对冲
	windowBothTriggered := longTriggered && shortTriggered
	scoreBothValid := longScore >= 25 && shortScore >= 25
	scoreClose := math.Abs(longScore-shortScore) < 20
	isHedged := windowBothTriggered || (scoreBothValid && scoreClose)

	if isHedged {
		pw.ResetAlerts()
		signal.Direction = "NEUTRAL"
		if windowBothTriggered {
			signal.Reason = fmt.Sprintf("窗口双向触发(多%.0f/空%.0f)，重置预警标记", longScore, shortScore)
		} else {
			signal.Reason = fmt.Sprintf("权重对冲(多%.0f/空%.0f差%.0f)，继续监控", longScore, shortScore, math.Abs(longScore-shortScore))
		}
		signal.SignalProgress = 0
		pw.SetLastSignal("neutral")
		return signal
	}

	// 确定最终方向
	var newSignal string = "neutral"
	shouldAlert := false

	if longScore > shortScore && longScore >= 50 {
		newSignal = "long"
		alertedLong := pw.GetAlertedLong()
		if alertedLong == nil || math.Abs(minPrice-*alertedLong) > 0.0001 {
			shouldAlert = true
			pw.SetAlertedLong(minPrice)
		}
	} else if shortScore > longScore && shortScore >= 50 {
		newSignal = "short"
		alertedShort := pw.GetAlertedShort()
		if alertedShort == nil || math.Abs(maxPrice-*alertedShort) > 0.0001 {
			shouldAlert = true
			pw.SetAlertedShort(maxPrice)
		}
	}

	// 设置信号结果
	switch newSignal {
	case "long":
		signal.Direction = "LONG"
		signal.Strength = longScore
		signal.Confidence = longScore
		signal.Action = "OPEN_LONG"
		signal.Reason = fmt.Sprintf("做多(权重%.0f%%) | 价格%.2f-低%.2f=%.2f≥阈值%.0f | %s",
			longScore, currentPrice, minPrice, distanceFromMin, threshold, strings.Join(reasonParts, " "))
	case "short":
		signal.Direction = "SHORT"
		signal.Strength = shortScore
		signal.Confidence = shortScore
		signal.Action = "OPEN_SHORT"
		signal.Reason = fmt.Sprintf("做空(权重%.0f%%) | 高%.2f-价格%.2f=%.2f≥阈值%.0f | %s",
			shortScore, maxPrice, currentPrice, distanceFromMax, threshold, strings.Join(reasonParts, " "))
	default:
		signal.Direction = "NEUTRAL"
		signal.Strength = 0
		signal.Confidence = 0
		signal.Action = "HOLD"
		longProgress := (distanceFromMin / threshold) * 100
		shortProgress := (distanceFromMax / threshold) * 100
		if longProgress > shortProgress {
			signal.SignalProgress = math.Min(100, longProgress)
		} else {
			signal.SignalProgress = math.Min(100, shortProgress)
		}
		signal.Reason = fmt.Sprintf("监控中 | 距做多%.0f/%.0f(%.0f%%) | 距做空%.0f/%.0f(%.0f%%)",
			distanceFromMin, threshold, longProgress, distanceFromMax, threshold, shortProgress)
	}

	// 记录信号历史
	pw.AddSignalHistory(newSignal)

	// 如果信号方向变化且需要预警
	lastSignal := pw.GetLastSignal()
	if newSignal != "neutral" && newSignal != lastSignal && shouldAlert {
		signal.AlignedTimeframes = 1
		if (newSignal == "long" && longScore > 50) || (newSignal == "short" && shortScore > 50) {
			go s.saveSignalLog(ctx, signal, newSignal, longScore, shortScore, reasonParts)
		}
	}

	pw.SetLastSignal(newSignal)

	return signal
}

// getIndicatorSignals 获取MACD和EMA指标信号
func (s *RobotSignalGen) getIndicatorSignals() (macdLong, macdShort, emaLong, emaShort bool) {
	analysis := s.engine.LastAnalysis
	if analysis == nil || analysis.TimeframeScores == nil {
		return false, false, false, false
	}

	var macdLongCount, macdShortCount int
	var emaLongCount, emaShortCount int
	var totalCount int

	for _, score := range analysis.TimeframeScores {
		totalCount++

		if score.MACD > 0 {
			macdLongCount++
		} else if score.MACD < 0 {
			macdShortCount++
		}

		if score.EMA12 > score.EMA26 {
			emaLongCount++
		} else if score.EMA12 < score.EMA26 {
			emaShortCount++
		}
	}

	if totalCount == 0 {
		return false, false, false, false
	}

	halfCount := totalCount / 2

	return macdLongCount > halfCount,
		macdShortCount > halfCount,
		emaLongCount > halfCount,
		emaShortCount > halfCount
}

// saveSignalLog 保存信号日志
func (s *RobotSignalGen) saveSignalLog(ctx context.Context, signal *Signal, direction string, longScore, shortScore float64, reasons []string) {
	// 方向文本：英文转中文
	directionCN := "做多"
	if direction == "short" || direction == "SHORT" {
		directionCN = "做空"
	}

	marketState := ""
	riskPreference := ""
	if s.engine.LastAnalysis != nil {
		marketState = s.engine.LastAnalysis.MarketState
	}
	// 风险偏好来自市场状态映射
	if marketState != "" {
		riskPreference = s.engine.MarketRiskMapping[marketState]
	}

	indicators := map[string]interface{}{
		"longScore":      longScore,
		"shortScore":     shortScore,
		"reasons":        reasons,
		"windowMin":      signal.WindowMinPrice,
		"windowMax":      signal.WindowMaxPrice,
		"currentPrice":   signal.CurrentPrice,
		"threshold":      signal.SignalThreshold,
		"marketState":    marketState,
		"riskPreference": riskPreference,
	}
	indicatorsJson, _ := json.Marshal(indicators)

	_, err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).Insert(g.Map{
		"robot_id":         s.engine.Robot.Id,
		"strategy_id":      0,
		"symbol":           s.engine.Robot.Symbol,
		"signal_type":      directionCN, // 使用中文：做多/做空
		"signal_source":    "signal_gen",
		"signal_strength":  signal.Strength,
		"current_price":    signal.CurrentPrice,
		"window_min_price": signal.WindowMinPrice,
		"window_max_price": signal.WindowMaxPrice,
		"threshold":        signal.SignalThreshold,
		"market_state":     marketState,
		"risk_preference":  riskPreference,
		"target_price":     0,
		"stop_loss":        0,
		"take_profit":      0,
		"executed":         0,
		"execute_result":   "",
		"is_processed":     1,
		"reason":           signal.Reason,
		"indicators":       string(indicatorsJson),
	})
	if err != nil {
		g.Log().Warningf(ctx, "[SignalGen] 保存信号日志失败: robotId=%d, err=%v", s.engine.Robot.Id, err)
	}
}
