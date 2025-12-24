// Package market 止盈止损管理器
// 负责高级止盈止损逻辑：移动止损、追踪止损、分批止盈
package market

import (
	"math"
	"sync"
	"time"
)

// StopLossManager 止盈止损管理器
type StopLossManager struct {
	mu sync.RWMutex

	// 持仓追踪 key: robotId:positionSide
	trackers map[string]*PositionTracker
}

// PositionTracker 持仓跟踪器（增强版）
type PositionTracker struct {
	RobotId       int64
	PositionSide  string    // LONG/SHORT
	EntryPrice    float64   // 入场价格
	EntryMargin   float64   // 入场保证金
	EntryTime     time.Time // 入场时间
	Quantity      float64   // 持仓数量

	// 盈亏跟踪
	HighestPrice  float64 // 最高价格（做多）
	LowestPrice   float64 // 最低价格（做空）
	HighestProfit float64 // 最高盈利金额
	LowestProfit  float64 // 最低盈利金额
	CurrentProfit float64 // 当前盈利金额

	// 追踪止损状态
	TrailingStopEnabled  bool    // 是否启用追踪止损
	TrailingStopPrice    float64 // 追踪止损价格
	TrailingStopActivated bool   // 追踪止损是否已激活

	// 分批止盈状态
	PartialTakeProfitEnabled bool             // 是否启用分批止盈
	TakeProfitLevels         []TakeProfitLevel // 止盈档位配置
	ClosedLevels             map[int]bool      // 已执行的止盈档位
	TotalClosedRatio         float64           // 已平仓比例

	// 回撤止盈状态
	RetreatTakeProfitEnabled bool    // 回撤止盈是否启用
	RetreatActivated         bool    // 回撤止盈是否已激活
	RetreatThreshold         float64 // 回撤止盈启动阈值（%）
	RetreatPercent           float64 // 回撤百分比

	// 时间止损
	TimeStopEnabled  bool          // 是否启用时间止损
	MaxHoldDuration  time.Duration // 最大持仓时间
}

// CloseSignal 平仓信号
type CloseSignal struct {
	ShouldClose  bool
	CloseRatio   float64 // 平仓比例 0-1
	Reason       string
	Priority     int     // 优先级（越小越高）
	StopPrice    float64 // 触发价格
}

var (
	stopLossManager     *StopLossManager
	stopLossManagerOnce sync.Once
)

// GetStopLossManager 获取止盈止损管理器单例
func GetStopLossManager() *StopLossManager {
	stopLossManagerOnce.Do(func() {
		stopLossManager = &StopLossManager{
			trackers: make(map[string]*PositionTracker),
		}
	})
	return stopLossManager
}

// CreateTracker 创建持仓跟踪器
func (m *StopLossManager) CreateTracker(robotId int64, positionSide string, entryPrice, entryMargin, quantity float64, config *StrategyExtConfig) *PositionTracker {
	key := m.getKey(robotId, positionSide)

	tracker := &PositionTracker{
		RobotId:       robotId,
		PositionSide:  positionSide,
		EntryPrice:    entryPrice,
		EntryMargin:   entryMargin,
		EntryTime:     time.Now(),
		Quantity:      quantity,
		HighestPrice:  entryPrice,
		LowestPrice:   entryPrice,
		ClosedLevels:  make(map[int]bool),
	}

	// 应用配置
	if config != nil {
		tracker.TrailingStopEnabled = config.TrailingStopEnabled
		tracker.PartialTakeProfitEnabled = config.PartialTakeProfitEnabled
		tracker.TakeProfitLevels = config.TakeProfitLevels
	}

	m.mu.Lock()
	m.trackers[key] = tracker
	m.mu.Unlock()

	return tracker
}

// GetTracker 获取持仓跟踪器
func (m *StopLossManager) GetTracker(robotId int64, positionSide string) *PositionTracker {
	key := m.getKey(robotId, positionSide)

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.trackers[key]
}

// RemoveTracker 移除持仓跟踪器
func (m *StopLossManager) RemoveTracker(robotId int64, positionSide string) {
	key := m.getKey(robotId, positionSide)

	m.mu.Lock()
	delete(m.trackers, key)
	m.mu.Unlock()
}

// UpdatePrice 更新价格并检查平仓信号
func (m *StopLossManager) UpdatePrice(robotId int64, positionSide string, currentPrice, currentPnl float64, template *StrategyTemplate) *CloseSignal {
	tracker := m.GetTracker(robotId, positionSide)
	if tracker == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新最高最低价
	if currentPrice > tracker.HighestPrice {
		tracker.HighestPrice = currentPrice
	}
	if currentPrice < tracker.LowestPrice || tracker.LowestPrice == 0 {
		tracker.LowestPrice = currentPrice
	}

	// 更新最高最低盈亏
	if currentPnl > tracker.HighestProfit {
		tracker.HighestProfit = currentPnl
	}
	if currentPnl < tracker.LowestProfit {
		tracker.LowestProfit = currentPnl
	}
	tracker.CurrentProfit = currentPnl

	// 检查各种平仓条件
	signals := []*CloseSignal{}

	// 1. 检查固定止损
	if signal := m.checkFixedStopLoss(tracker, currentPrice, currentPnl, template); signal != nil {
		signals = append(signals, signal)
	}

	// 2. 检查追踪止损
	if signal := m.checkTrailingStop(tracker, currentPrice, template); signal != nil {
		signals = append(signals, signal)
	}

	// 3. 检查回撤止盈
	if signal := m.checkRetreatTakeProfit(tracker, currentPnl, template); signal != nil {
		signals = append(signals, signal)
	}

	// 4. 检查分批止盈
	if signal := m.checkPartialTakeProfit(tracker, currentPnl); signal != nil {
		signals = append(signals, signal)
	}

	// 5. 检查时间止损
	if signal := m.checkTimeStop(tracker); signal != nil {
		signals = append(signals, signal)
	}

	// 返回优先级最高的信号
	if len(signals) == 0 {
		return nil
	}

	highestPriority := signals[0]
	for _, s := range signals[1:] {
		if s.Priority < highestPriority.Priority {
			highestPriority = s
		}
	}

	return highestPriority
}

// checkFixedStopLoss 检查固定止损
func (m *StopLossManager) checkFixedStopLoss(tracker *PositionTracker, currentPrice, currentPnl float64, template *StrategyTemplate) *CloseSignal {
	if template == nil || template.StopLossPercent <= 0 {
		return nil
	}

	if tracker.EntryMargin <= 0 {
		return nil
	}

	// 计算亏损百分比
	lossPercent := math.Abs(currentPnl) / tracker.EntryMargin * 100

	if currentPnl < 0 && lossPercent >= template.StopLossPercent {
		return &CloseSignal{
			ShouldClose: true,
			CloseRatio:  1.0,
			Reason:      "触发固定止损",
			Priority:    1,
			StopPrice:   currentPrice,
		}
	}

	return nil
}

// checkTrailingStop 检查追踪止损
func (m *StopLossManager) checkTrailingStop(tracker *PositionTracker, currentPrice float64, template *StrategyTemplate) *CloseSignal {
	if !tracker.TrailingStopEnabled || template == nil {
		return nil
	}

	trailingPercent := 1.0 // 默认1%
	if template.ExtConfig != nil && template.ExtConfig.TrailingStopPercent > 0 {
		trailingPercent = template.ExtConfig.TrailingStopPercent
	}

	var shouldClose bool
	var stopPrice float64

	if tracker.PositionSide == "LONG" {
		// 做多：追踪最高价
		stopPrice = tracker.HighestPrice * (1 - trailingPercent/100)
		shouldClose = currentPrice <= stopPrice

		// 更新追踪止损价格
		if currentPrice > tracker.EntryPrice {
			tracker.TrailingStopActivated = true
			if stopPrice > tracker.TrailingStopPrice {
				tracker.TrailingStopPrice = stopPrice
			}
		}
	} else {
		// 做空：追踪最低价
		stopPrice = tracker.LowestPrice * (1 + trailingPercent/100)
		shouldClose = currentPrice >= stopPrice

		// 更新追踪止损价格
		if currentPrice < tracker.EntryPrice {
			tracker.TrailingStopActivated = true
			if tracker.TrailingStopPrice == 0 || stopPrice < tracker.TrailingStopPrice {
				tracker.TrailingStopPrice = stopPrice
			}
		}
	}

	if shouldClose && tracker.TrailingStopActivated {
		return &CloseSignal{
			ShouldClose: true,
			CloseRatio:  1.0,
			Reason:      "触发追踪止损",
			Priority:    2,
			StopPrice:   stopPrice,
		}
	}

	return nil
}

// checkRetreatTakeProfit 检查回撤止盈
func (m *StopLossManager) checkRetreatTakeProfit(tracker *PositionTracker, currentPnl float64, template *StrategyTemplate) *CloseSignal {
	if template == nil {
		return nil
	}

	autoStartPercent := template.AutoStartRetreatPercent
	retreatPercent := template.ProfitRetreatPercent

	if autoStartPercent <= 0 || retreatPercent <= 0 {
		return nil
	}

	if tracker.EntryMargin <= 0 {
		return nil
	}

	// 计算盈利百分比
	profitPercent := tracker.CurrentProfit / tracker.EntryMargin * 100

	// 检查是否达到启动阈值
	if profitPercent >= autoStartPercent {
		tracker.RetreatActivated = true
	}

	if !tracker.RetreatActivated {
		return nil
	}

	// 检查回撤是否达到阈值
	if tracker.HighestProfit <= 0 {
		return nil
	}

	retreat := (tracker.HighestProfit - tracker.CurrentProfit) / tracker.HighestProfit * 100

	if retreat >= retreatPercent {
		return &CloseSignal{
			ShouldClose: true,
			CloseRatio:  1.0,
			Reason:      "触发回撤止盈",
			Priority:    3,
		}
	}

	return nil
}

// checkPartialTakeProfit 检查分批止盈
func (m *StopLossManager) checkPartialTakeProfit(tracker *PositionTracker, currentPnl float64) *CloseSignal {
	if !tracker.PartialTakeProfitEnabled || len(tracker.TakeProfitLevels) == 0 {
		return nil
	}

	if tracker.EntryMargin <= 0 {
		return nil
	}

	// 计算盈利百分比
	profitPercent := currentPnl / tracker.EntryMargin * 100

	// 检查各止盈档位
	for i, level := range tracker.TakeProfitLevels {
		if tracker.ClosedLevels[i] {
			continue
		}

		if profitPercent >= level.Percent {
			// 计算剩余可平仓比例
			remainingRatio := 1.0 - tracker.TotalClosedRatio
			closeRatio := level.CloseRatio * remainingRatio

			if closeRatio > 0 && closeRatio <= remainingRatio {
				tracker.ClosedLevels[i] = true
				tracker.TotalClosedRatio += closeRatio

				return &CloseSignal{
					ShouldClose: true,
					CloseRatio:  closeRatio,
					Reason:      "触发分批止盈",
					Priority:    4,
				}
			}
		}
	}

	return nil
}

// checkTimeStop 检查时间止损
func (m *StopLossManager) checkTimeStop(tracker *PositionTracker) *CloseSignal {
	if !tracker.TimeStopEnabled || tracker.MaxHoldDuration <= 0 {
		return nil
	}

	if time.Since(tracker.EntryTime) >= tracker.MaxHoldDuration {
		return &CloseSignal{
			ShouldClose: true,
			CloseRatio:  1.0,
			Reason:      "触发时间止损",
			Priority:    5,
		}
	}

	return nil
}

// getKey 生成缓存键
func (m *StopLossManager) getKey(robotId int64, positionSide string) string {
	return string(rune(robotId)) + ":" + positionSide
}

// GetAllTrackers 获取所有跟踪器状态
func (m *StopLossManager) GetAllTrackers() map[string]*PositionTracker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*PositionTracker)
	for k, v := range m.trackers {
		result[k] = v
	}
	return result
}

// CalculateStopLossPrice 计算止损价格
func CalculateStopLossPrice(entryPrice float64, stopLossPercent float64, positionSide string, leverage int) float64 {
	// 考虑杠杆的止损价格
	effectivePercent := stopLossPercent / float64(leverage)

	if positionSide == "LONG" {
		return entryPrice * (1 - effectivePercent/100)
	}
	return entryPrice * (1 + effectivePercent/100)
}

// CalculateTakeProfitPrice 计算止盈价格
func CalculateTakeProfitPrice(entryPrice float64, takeProfitPercent float64, positionSide string, leverage int) float64 {
	effectivePercent := takeProfitPercent / float64(leverage)

	if positionSide == "LONG" {
		return entryPrice * (1 + effectivePercent/100)
	}
	return entryPrice * (1 - effectivePercent/100)
}

// CalculatePnLPercent 计算盈亏百分比
func CalculatePnLPercent(entryPrice, currentPrice float64, positionSide string, leverage int) float64 {
	var pnlPercent float64
	if positionSide == "LONG" {
		pnlPercent = (currentPrice - entryPrice) / entryPrice * 100
	} else {
		pnlPercent = (entryPrice - currentPrice) / entryPrice * 100
	}
	return pnlPercent * float64(leverage)
}

