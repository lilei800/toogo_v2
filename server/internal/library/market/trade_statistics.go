// Package market 交易统计和绩效分析
// 负责记录和分析机器人的交易绩效
package market

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/dao"
)

// TradeStatisticsManager 交易统计管理器（单例）
type TradeStatisticsManager struct {
	mu sync.RWMutex

	// 机器人统计缓存 key: robotId
	robotStats map[int64]*RobotStatistics

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// RobotStatistics 机器人统计数据
type RobotStatistics struct {
	RobotId   int64
	UpdatedAt time.Time

	// 交易计数
	TotalTrades     int // 总交易次数
	WinTrades       int // 盈利次数
	LossTrades      int // 亏损次数
	BreakEvenTrades int // 平盘次数

	// 盈亏统计
	TotalProfit    float64 // 总盈利
	TotalLoss      float64 // 总亏损
	NetProfit      float64 // 净盈亏
	GrossProfit    float64 // 毛利润
	MaxProfit      float64 // 单笔最大盈利
	MaxLoss        float64 // 单笔最大亏损
	AvgProfit      float64 // 平均盈利
	AvgLoss        float64 // 平均亏损

	// 方向统计
	LongTrades      int     // 做多次数
	ShortTrades     int     // 做空次数
	LongWinRate     float64 // 做多胜率
	ShortWinRate    float64 // 做空胜率
	LongProfit      float64 // 做多盈亏
	ShortProfit     float64 // 做空盈亏

	// 比率指标
	WinRate        float64 // 胜率
	ProfitFactor   float64 // 盈亏比（总盈利/总亏损）
	ExpectedValue  float64 // 期望值
	RiskRewardRatio float64 // 风险收益比

	// 风险指标
	MaxDrawdown        float64 // 最大回撤
	MaxDrawdownPercent float64 // 最大回撤百分比
	MaxConsecutiveLoss int     // 最大连续亏损次数
	MaxConsecutiveWin  int     // 最大连续盈利次数
	CurrentStreak      int     // 当前连续（正数盈利，负数亏损）

	// 时间统计
	AvgHoldDuration    time.Duration // 平均持仓时间
	TotalHoldDuration  time.Duration // 总持仓时间
	LongestWinHold     time.Duration // 最长盈利持仓
	LongestLossHold    time.Duration // 最长亏损持仓

	// 日内统计
	TodayTrades   int     // 今日交易次数
	TodayProfit   float64 // 今日盈亏
	TodayWinRate  float64 // 今日胜率

	// 峰值跟踪
	PeakEquity    float64 // 权益峰值
	CurrentEquity float64 // 当前权益
}

// TradeRecord 交易记录
type TradeRecord struct {
	RobotId       int64
	OrderId       string
	Symbol        string
	Side          string // BUY/SELL
	PositionSide  string // LONG/SHORT
	OpenPrice     float64
	ClosePrice    float64
	Quantity      float64
	Profit        float64
	ProfitPercent float64
	OpenTime      time.Time
	CloseTime     time.Time
	HoldDuration  time.Duration
	CloseReason   string
	MarketState   string
	RiskPreference string
	SignalStrength float64
}

// PerformanceReport 绩效报告
type PerformanceReport struct {
	RobotId      int64
	Period       string // daily/weekly/monthly/all
	StartTime    time.Time
	EndTime      time.Time
	Statistics   *RobotStatistics
	
	// 分析评价
	OverallRating   string  // A/B/C/D/F
	RiskRating      string  // 风险评级
	ConsistencyRating string // 稳定性评级
	
	// 建议
	Recommendations []string
}

var (
	tradeStatisticsManager     *TradeStatisticsManager
	tradeStatisticsManagerOnce sync.Once
)

// GetTradeStatisticsManager 获取交易统计管理器单例
func GetTradeStatisticsManager() *TradeStatisticsManager {
	tradeStatisticsManagerOnce.Do(func() {
		tradeStatisticsManager = &TradeStatisticsManager{
			robotStats: make(map[int64]*RobotStatistics),
			stopCh:     make(chan struct{}),
		}
	})
	return tradeStatisticsManager
}

// Start 启动统计管理器
func (m *TradeStatisticsManager) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	g.Log().Info(ctx, "[TradeStatisticsManager] 交易统计管理器启动")

	// 启动每日重置任务
	go m.runDailyResetTask(ctx)
}

// Stop 停止统计管理器
func (m *TradeStatisticsManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}
	m.running = false
	close(m.stopCh)
}

// runDailyResetTask 每日重置任务
func (m *TradeStatisticsManager) runDailyResetTask(ctx context.Context) {
	// 计算到明天0点的时间
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	duration := tomorrow.Sub(now)

	timer := time.NewTimer(duration)
	defer timer.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-timer.C:
			m.resetDailyStats()
			// 重置到下一个24小时
			timer.Reset(24 * time.Hour)
		}
	}
}

// resetDailyStats 重置每日统计
func (m *TradeStatisticsManager) resetDailyStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, stats := range m.robotStats {
		stats.TodayTrades = 0
		stats.TodayProfit = 0
		stats.TodayWinRate = 0
	}
}

// GetRobotStatistics 获取机器人统计数据
func (m *TradeStatisticsManager) GetRobotStatistics(robotId int64) *RobotStatistics {
	m.mu.RLock()
	stats, exists := m.robotStats[robotId]
	m.mu.RUnlock()

	if exists {
		return stats
	}

	// 从数据库加载
	return m.loadRobotStatistics(robotId)
}

// loadRobotStatistics 从数据库加载统计数据
func (m *TradeStatisticsManager) loadRobotStatistics(robotId int64) *RobotStatistics {
	stats := &RobotStatistics{
		RobotId:   robotId,
		UpdatedAt: time.Now(),
	}

	// 从订单表统计
	ctx := context.Background()

	// 统计总交易次数
	totalCount, _ := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robotId).
		Where("status", "FILLED").
		Count()
	stats.TotalTrades = totalCount

	// 统计盈利订单
	var profitSum float64
	profitOrders, _ := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robotId).
		Where("status", "FILLED").
		Where("realized_pnl > 0").
		Count()
	stats.WinTrades = profitOrders

	// 统计亏损订单
	lossOrders, _ := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robotId).
		Where("status", "FILLED").
		Where("realized_pnl < 0").
		Count()
	stats.LossTrades = lossOrders

	// 统计总盈亏
	profitSum, _ = dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robotId).
		Where("status", "FILLED").
		Sum("realized_pnl")
	stats.NetProfit = profitSum

	// 计算衍生指标
	m.calculateDerivedStats(stats)

	m.mu.Lock()
	m.robotStats[robotId] = stats
	m.mu.Unlock()

	return stats
}

// RecordTrade 记录交易
func (m *TradeStatisticsManager) RecordTrade(record *TradeRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.robotStats[record.RobotId]
	if !exists {
		stats = &RobotStatistics{
			RobotId:   record.RobotId,
			UpdatedAt: time.Now(),
		}
		m.robotStats[record.RobotId] = stats
	}

	// 更新交易计数
	stats.TotalTrades++
	stats.TodayTrades++

	// 更新方向统计
	if record.PositionSide == "LONG" {
		stats.LongTrades++
		stats.LongProfit += record.Profit
	} else {
		stats.ShortTrades++
		stats.ShortProfit += record.Profit
	}

	// 更新盈亏统计
	if record.Profit > 0 {
		stats.WinTrades++
		stats.TotalProfit += record.Profit
		if record.Profit > stats.MaxProfit {
			stats.MaxProfit = record.Profit
		}
		// 更新连续统计
		if stats.CurrentStreak < 0 {
			stats.CurrentStreak = 1
		} else {
			stats.CurrentStreak++
		}
		if stats.CurrentStreak > stats.MaxConsecutiveWin {
			stats.MaxConsecutiveWin = stats.CurrentStreak
		}
		if record.HoldDuration > stats.LongestWinHold {
			stats.LongestWinHold = record.HoldDuration
		}
	} else if record.Profit < 0 {
		stats.LossTrades++
		stats.TotalLoss += math.Abs(record.Profit)
		if math.Abs(record.Profit) > stats.MaxLoss {
			stats.MaxLoss = math.Abs(record.Profit)
		}
		// 更新连续统计
		if stats.CurrentStreak > 0 {
			stats.CurrentStreak = -1
		} else {
			stats.CurrentStreak--
		}
		if math.Abs(float64(stats.CurrentStreak)) > float64(stats.MaxConsecutiveLoss) {
			stats.MaxConsecutiveLoss = int(math.Abs(float64(stats.CurrentStreak)))
		}
		if record.HoldDuration > stats.LongestLossHold {
			stats.LongestLossHold = record.HoldDuration
		}
	} else {
		stats.BreakEvenTrades++
	}

	// 更新净盈亏和今日盈亏
	stats.NetProfit += record.Profit
	stats.TodayProfit += record.Profit

	// 更新持仓时间统计
	stats.TotalHoldDuration += record.HoldDuration

	// 更新权益峰值和回撤
	stats.CurrentEquity += record.Profit
	if stats.CurrentEquity > stats.PeakEquity {
		stats.PeakEquity = stats.CurrentEquity
	}
	if stats.PeakEquity > 0 {
		drawdown := (stats.PeakEquity - stats.CurrentEquity) / stats.PeakEquity * 100
		if drawdown > stats.MaxDrawdownPercent {
			stats.MaxDrawdownPercent = drawdown
			stats.MaxDrawdown = stats.PeakEquity - stats.CurrentEquity
		}
	}

	// 重新计算衍生指标
	m.calculateDerivedStats(stats)

	stats.UpdatedAt = time.Now()
}

// calculateDerivedStats 计算衍生统计指标
func (m *TradeStatisticsManager) calculateDerivedStats(stats *RobotStatistics) {
	// 胜率
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinTrades) / float64(stats.TotalTrades) * 100
		stats.TodayWinRate = stats.WinRate // 简化，实际应单独计算
	}

	// 平均盈亏
	if stats.WinTrades > 0 {
		stats.AvgProfit = stats.TotalProfit / float64(stats.WinTrades)
	}
	if stats.LossTrades > 0 {
		stats.AvgLoss = stats.TotalLoss / float64(stats.LossTrades)
	}

	// 盈亏比
	if stats.TotalLoss > 0 {
		stats.ProfitFactor = stats.TotalProfit / stats.TotalLoss
	}

	// 期望值
	if stats.TotalTrades > 0 {
		winProb := float64(stats.WinTrades) / float64(stats.TotalTrades)
		lossProb := float64(stats.LossTrades) / float64(stats.TotalTrades)
		stats.ExpectedValue = winProb*stats.AvgProfit - lossProb*stats.AvgLoss
	}

	// 风险收益比
	if stats.AvgLoss > 0 {
		stats.RiskRewardRatio = stats.AvgProfit / stats.AvgLoss
	}

	// 平均持仓时间
	if stats.TotalTrades > 0 {
		stats.AvgHoldDuration = stats.TotalHoldDuration / time.Duration(stats.TotalTrades)
	}

	// 方向胜率
	if stats.LongTrades > 0 {
		longWins := 0 // 需要额外统计
		stats.LongWinRate = float64(longWins) / float64(stats.LongTrades) * 100
	}
	if stats.ShortTrades > 0 {
		shortWins := 0 // 需要额外统计
		stats.ShortWinRate = float64(shortWins) / float64(stats.ShortTrades) * 100
	}
}

// GeneratePerformanceReport 生成绩效报告
func (m *TradeStatisticsManager) GeneratePerformanceReport(robotId int64, period string) *PerformanceReport {
	stats := m.GetRobotStatistics(robotId)
	if stats == nil {
		return nil
	}

	report := &PerformanceReport{
		RobotId:    robotId,
		Period:     period,
		EndTime:    time.Now(),
		Statistics: stats,
	}

	// 设置时间范围
	switch period {
	case "daily":
		report.StartTime = time.Now().AddDate(0, 0, -1)
	case "weekly":
		report.StartTime = time.Now().AddDate(0, 0, -7)
	case "monthly":
		report.StartTime = time.Now().AddDate(0, -1, 0)
	default:
		report.StartTime = time.Time{}
	}

	// 评估总体评级
	report.OverallRating = m.calculateOverallRating(stats)
	report.RiskRating = m.calculateRiskRating(stats)
	report.ConsistencyRating = m.calculateConsistencyRating(stats)

	// 生成建议
	report.Recommendations = m.generateRecommendations(stats)

	return report
}

// calculateOverallRating 计算总体评级
func (m *TradeStatisticsManager) calculateOverallRating(stats *RobotStatistics) string {
	score := 0.0

	// 胜率评分 (0-25分)
	if stats.WinRate >= 60 {
		score += 25
	} else if stats.WinRate >= 50 {
		score += 20
	} else if stats.WinRate >= 40 {
		score += 15
	} else {
		score += 10
	}

	// 盈亏比评分 (0-25分)
	if stats.ProfitFactor >= 2 {
		score += 25
	} else if stats.ProfitFactor >= 1.5 {
		score += 20
	} else if stats.ProfitFactor >= 1 {
		score += 15
	} else {
		score += 5
	}

	// 最大回撤评分 (0-25分)
	if stats.MaxDrawdownPercent <= 10 {
		score += 25
	} else if stats.MaxDrawdownPercent <= 20 {
		score += 20
	} else if stats.MaxDrawdownPercent <= 30 {
		score += 15
	} else {
		score += 5
	}

	// 期望值评分 (0-25分)
	if stats.ExpectedValue > 0 {
		score += 25
	} else {
		score += 10
	}

	// 评级
	if score >= 90 {
		return "A"
	} else if score >= 75 {
		return "B"
	} else if score >= 60 {
		return "C"
	} else if score >= 45 {
		return "D"
	}
	return "F"
}

// calculateRiskRating 计算风险评级
func (m *TradeStatisticsManager) calculateRiskRating(stats *RobotStatistics) string {
	if stats.MaxDrawdownPercent <= 10 && stats.MaxConsecutiveLoss <= 3 {
		return "低风险"
	} else if stats.MaxDrawdownPercent <= 20 && stats.MaxConsecutiveLoss <= 5 {
		return "中等风险"
	} else if stats.MaxDrawdownPercent <= 30 {
		return "较高风险"
	}
	return "高风险"
}

// calculateConsistencyRating 计算稳定性评级
func (m *TradeStatisticsManager) calculateConsistencyRating(stats *RobotStatistics) string {
	// 根据连续亏损和胜率波动评估稳定性
	if stats.MaxConsecutiveLoss <= 3 && stats.WinRate >= 50 {
		return "高稳定性"
	} else if stats.MaxConsecutiveLoss <= 5 && stats.WinRate >= 40 {
		return "中等稳定性"
	}
	return "低稳定性"
}

// generateRecommendations 生成建议
func (m *TradeStatisticsManager) generateRecommendations(stats *RobotStatistics) []string {
	recommendations := []string{}

	// 胜率建议
	if stats.WinRate < 40 {
		recommendations = append(recommendations, "胜率偏低，建议提高信号过滤阈值，只在高置信度信号时交易")
	}

	// 盈亏比建议
	if stats.ProfitFactor < 1 {
		recommendations = append(recommendations, "盈亏比小于1，建议调整止盈止损比例，提高盈亏比")
	} else if stats.ProfitFactor < 1.5 {
		recommendations = append(recommendations, "盈亏比可以优化，建议适当扩大止盈目标或收紧止损")
	}

	// 回撤建议
	if stats.MaxDrawdownPercent > 20 {
		recommendations = append(recommendations, "最大回撤较大，建议降低杠杆或减少单笔保证金比例")
	}

	// 连续亏损建议
	if stats.MaxConsecutiveLoss > 5 {
		recommendations = append(recommendations, "出现过连续5次以上亏损，建议增加交易冷却时间，优化入场时机")
	}

	// 方向偏好建议
	if stats.LongTrades > 0 && stats.ShortTrades > 0 {
		if stats.LongProfit > stats.ShortProfit*2 {
			recommendations = append(recommendations, "做多盈利明显优于做空，可考虑增加做多权重")
		} else if stats.ShortProfit > stats.LongProfit*2 {
			recommendations = append(recommendations, "做空盈利明显优于做多，可考虑增加做空权重")
		}
	}

	// 持仓时间建议
	if stats.AvgHoldDuration > 2*time.Hour {
		recommendations = append(recommendations, "平均持仓时间较长，注意隔夜风险和资金费率成本")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "策略运行良好，继续保持当前参数设置")
	}

	return recommendations
}

// ClearRobotStats 清除机器人统计缓存
func (m *TradeStatisticsManager) ClearRobotStats(robotId int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.robotStats, robotId)
}

