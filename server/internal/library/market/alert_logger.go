// Package market 预警日志持久化服务
// 负责将预警日志写入数据库
package market

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/dao"
)

// AlertLogger 预警日志记录器（单例）
type AlertLogger struct {
	mu sync.Mutex

	// 日志缓冲队列（批量写入）
	marketStateBuffer     []*MarketStateLogEntry
	riskPreferenceBuffer  []*RiskPreferenceLogEntry
	directionBuffer       []*DirectionLogEntry

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// MarketStateLogEntry 市场状态日志条目
type MarketStateLogEntry struct {
	Platform     string
	Symbol       string
	PrevState    string
	NewState     string
	Confidence   float64
	TrendStrength float64
	Volatility   float64
	TrendScore   float64
	MomentumScore float64
	SupportLevel float64
	ResistanceLevel float64
	Reason       string
	CreatedAt    time.Time
}

// RiskPreferenceLogEntry 风险偏好日志条目
type RiskPreferenceLogEntry struct {
	RobotId        int64
	UserId         int64
	PrevPreference string
	NewPreference  string
	WinProbability float64
	MarketScore    float64
	TechnicalScore float64
	AccountScore   float64
	HistoryScore   float64
	VolatilityRisk float64
	RiskLevel      int
	Reason         string
	CreatedAt      time.Time
}

// DirectionLogEntry 方向日志条目
type DirectionLogEntry struct {
	RobotId       int64
	UserId        int64
	Platform      string
	Symbol        string
	PrevDirection string
	NewDirection  string
	Strength      float64
	Confidence    float64
	Action        string
	EntryPrice    float64
	StopLoss      float64
	TakeProfit    float64
	TrendSignal   string
	MomentumSignal string
	PatternSignal string
	Reason        string
	CreatedAt     time.Time
}

var (
	alertLogger     *AlertLogger
	alertLoggerOnce sync.Once
)

// GetAlertLogger 获取预警日志记录器单例
func GetAlertLogger() *AlertLogger {
	alertLoggerOnce.Do(func() {
		alertLogger = &AlertLogger{
			marketStateBuffer:    make([]*MarketStateLogEntry, 0, 100),
			riskPreferenceBuffer: make([]*RiskPreferenceLogEntry, 0, 100),
			directionBuffer:      make([]*DirectionLogEntry, 0, 100),
			stopCh:               make(chan struct{}),
		}
	})
	return alertLogger
}

// Start 启动日志记录器
func (l *AlertLogger) Start(ctx context.Context) {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return
	}
	l.running = true
	l.mu.Unlock()

	g.Log().Info(ctx, "[AlertLogger] 预警日志记录器启动")

	// 启动定时刷新任务（每5秒批量写入数据库）
	go l.runFlushLoop(ctx)
}

// Stop 停止日志记录器
func (l *AlertLogger) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return
	}
	l.running = false
	close(l.stopCh)
}

// runFlushLoop 定时刷新循环
func (l *AlertLogger) runFlushLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopCh:
			// 停止前最后刷新一次
			l.flushAll(ctx)
			return
		case <-ticker.C:
			l.flushAll(ctx)
		}
	}
}

// flushAll 刷新所有缓冲
func (l *AlertLogger) flushAll(ctx context.Context) {
	l.flushMarketStateLogs(ctx)
	l.flushRiskPreferenceLogs(ctx)
	l.flushDirectionLogs(ctx)
}

// LogMarketState 记录市场状态变化
func (l *AlertLogger) LogMarketState(entry *MarketStateLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry.CreatedAt = time.Now()
	l.marketStateBuffer = append(l.marketStateBuffer, entry)

	// 缓冲区满时立即刷新
	if len(l.marketStateBuffer) >= 50 {
		go l.flushMarketStateLogs(context.Background())
	}
}

// LogRiskPreference 记录风险偏好变化
func (l *AlertLogger) LogRiskPreference(entry *RiskPreferenceLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry.CreatedAt = time.Now()
	l.riskPreferenceBuffer = append(l.riskPreferenceBuffer, entry)

	if len(l.riskPreferenceBuffer) >= 50 {
		go l.flushRiskPreferenceLogs(context.Background())
	}
}

// LogDirection 记录方向变化
func (l *AlertLogger) LogDirection(entry *DirectionLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry.CreatedAt = time.Now()
	l.directionBuffer = append(l.directionBuffer, entry)

	if len(l.directionBuffer) >= 50 {
		go l.flushDirectionLogs(context.Background())
	}
}

// flushMarketStateLogs 刷新市场状态日志
func (l *AlertLogger) flushMarketStateLogs(ctx context.Context) {
	l.mu.Lock()
	if len(l.marketStateBuffer) == 0 {
		l.mu.Unlock()
		return
	}
	logs := l.marketStateBuffer
	l.marketStateBuffer = make([]*MarketStateLogEntry, 0, 100)
	l.mu.Unlock()

	// 批量插入数据库
	for _, entry := range logs {
		indicators, _ := json.Marshal(map[string]float64{
			"trend_strength":  entry.TrendStrength,
			"volatility":      entry.Volatility,
			"trend_score":     entry.TrendScore,
			"momentum_score":  entry.MomentumScore,
			"support_level":   entry.SupportLevel,
			"resistance_level": entry.ResistanceLevel,
		})

		_, err := dao.TradingMarketStateLog.Ctx(ctx).Insert(g.Map{
			"platform":     entry.Platform,
			"symbol":       entry.Symbol,
			"prev_state":   entry.PrevState,
			"new_state":    entry.NewState,
			"confidence":   entry.Confidence,
			"indicators":   string(indicators),
			"reason":       entry.Reason,
			"created_at":   entry.CreatedAt,
		})
		if err != nil {
			g.Log().Warningf(ctx, "[AlertLogger] 写入市场状态日志失败: %v", err)
		}
	}

	g.Log().Debugf(ctx, "[AlertLogger] 写入%d条市场状态日志", len(logs))
}

// flushRiskPreferenceLogs 刷新风险偏好日志
func (l *AlertLogger) flushRiskPreferenceLogs(ctx context.Context) {
	l.mu.Lock()
	if len(l.riskPreferenceBuffer) == 0 {
		l.mu.Unlock()
		return
	}
	logs := l.riskPreferenceBuffer
	l.riskPreferenceBuffer = make([]*RiskPreferenceLogEntry, 0, 100)
	l.mu.Unlock()

	for _, entry := range logs {
		factors, _ := json.Marshal(map[string]float64{
			"market_score":    entry.MarketScore,
			"technical_score": entry.TechnicalScore,
			"account_score":   entry.AccountScore,
			"history_score":   entry.HistoryScore,
			"volatility_risk": entry.VolatilityRisk,
		})

		_, err := dao.TradingRiskPreferenceLog.Ctx(ctx).Insert(g.Map{
			"robot_id":        entry.RobotId,
			"user_id":         entry.UserId,
			"prev_preference": entry.PrevPreference,
			"new_preference":  entry.NewPreference,
			"win_probability": entry.WinProbability,
			"market_score":    entry.MarketScore,
			"technical_score": entry.TechnicalScore,
			"account_score":   entry.AccountScore,
			"history_score":   entry.HistoryScore,
			"volatility_risk": entry.VolatilityRisk,
			"factors":         string(factors),
			"reason":          entry.Reason,
			"created_at":      entry.CreatedAt,
		})
		if err != nil {
			g.Log().Warningf(ctx, "[AlertLogger] 写入风险偏好日志失败: %v", err)
		}
	}

	g.Log().Debugf(ctx, "[AlertLogger] 写入%d条风险偏好日志", len(logs))
}

// flushDirectionLogs 刷新方向日志
func (l *AlertLogger) flushDirectionLogs(ctx context.Context) {
	l.mu.Lock()
	if len(l.directionBuffer) == 0 {
		l.mu.Unlock()
		return
	}
	logs := l.directionBuffer
	l.directionBuffer = make([]*DirectionLogEntry, 0, 100)
	l.mu.Unlock()

	for _, entry := range logs {
		indicators, _ := json.Marshal(map[string]string{
			"trend_signal":    entry.TrendSignal,
			"momentum_signal": entry.MomentumSignal,
			"pattern_signal":  entry.PatternSignal,
		})

		_, err := dao.TradingDirectionLog.Ctx(ctx).Insert(g.Map{
			"platform":         entry.Platform,
			"symbol":           entry.Symbol,
			"prev_direction":   entry.PrevDirection,
			"new_direction":    entry.NewDirection,
			"strength":         entry.Strength,
			"confidence":       entry.Confidence,
			"action":           entry.Action,
			"trend_signal":     entry.TrendSignal,
			"momentum_signal":  entry.MomentumSignal,
			"pattern_signal":   entry.PatternSignal,
			"entry_price":      entry.EntryPrice,
			"stop_loss":        entry.StopLoss,
			"take_profit_1":    entry.TakeProfit,
			"indicators":       string(indicators),
			"reason":           entry.Reason,
			"created_at":       entry.CreatedAt,
		})
		if err != nil {
			g.Log().Warningf(ctx, "[AlertLogger] 写入方向日志失败: %v", err)
		}
	}

	g.Log().Debugf(ctx, "[AlertLogger] 写入%d条方向日志", len(logs))
}

// GetStats 获取缓冲统计
func (l *AlertLogger) GetStats() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return map[string]int{
		"market_state_buffer":    len(l.marketStateBuffer),
		"risk_preference_buffer": len(l.riskPreferenceBuffer),
		"direction_buffer":       len(l.directionBuffer),
	}
}

