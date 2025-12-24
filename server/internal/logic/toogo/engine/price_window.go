// Package engine 机器人引擎模块 - 价格窗口管理
package engine

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PriceWindow 价格窗口管理器
type PriceWindow struct {
	mu sync.RWMutex

	// 配置
	Config *MonitorConfig

	// 数据
	Points        []PricePoint
	SignalHistory []SignalHistoryItem

	// 状态
	LastAlertedLong  *float64
	LastAlertedShort *float64
	LastWindowMin    *float64
	LastWindowMax    *float64
	LastWindowSignal string
}

// WindowStats 窗口统计数据
type WindowStats struct {
	MinPrice     float64
	MaxPrice     float64
	CurrentPrice float64
	DataCount    int
}

// NewPriceWindow 创建价格窗口
func NewPriceWindow(symbol string, window int, threshold float64) *PriceWindow {
	return &PriceWindow{
		Config: &MonitorConfig{
			Symbol:    symbol,
			Window:    window,
			Threshold: threshold,
		},
		Points:           make([]PricePoint, 0, MaxPriceWindowSize),
		SignalHistory:    make([]SignalHistoryItem, 0, MaxSignalHistorySize),
		LastWindowSignal: "neutral",
	}
}

// AddPoint 添加价格数据点
func (pw *PriceWindow) AddPoint(price float64) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	now := time.Now().UnixMilli()

	// 容量限制
	if len(pw.Points) >= MaxPriceWindowSize {
		// 保留后半部分数据
		pw.Points = pw.Points[len(pw.Points)-MaxPriceWindowSize/2:]
	}

	pw.Points = append(pw.Points, PricePoint{
		Timestamp: now,
		Price:     price,
	})

	// 修剪窗口期外的数据
	pw.pruneOldData(now)
}

// pruneOldData 修剪窗口期外的价格数据
func (pw *PriceWindow) pruneOldData(now int64) {
	if pw.Config == nil || pw.Config.Window <= 0 {
		return
	}

	cutoff := now - int64(pw.Config.Window)*1000
	newPoints := make([]PricePoint, 0, len(pw.Points))
	for _, p := range pw.Points {
		if p.Timestamp >= cutoff {
			newPoints = append(newPoints, p)
		}
	}
	pw.Points = newPoints
}

// GetPoints 获取价格窗口数据
func (pw *PriceWindow) GetPoints() []PricePoint {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	result := make([]PricePoint, len(pw.Points))
	copy(result, pw.Points)
	return result
}

// GetStats 获取窗口统计数据
func (pw *PriceWindow) GetStats() *WindowStats {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	stats := &WindowStats{
		DataCount: len(pw.Points),
	}

	if stats.DataCount == 0 {
		return stats
	}

	stats.MinPrice = pw.Points[0].Price
	stats.MaxPrice = pw.Points[0].Price

	for _, p := range pw.Points {
		if p.Price < stats.MinPrice {
			stats.MinPrice = p.Price
		}
		if p.Price > stats.MaxPrice {
			stats.MaxPrice = p.Price
		}
	}

	stats.CurrentPrice = pw.Points[stats.DataCount-1].Price
	return stats
}

// GetSignalHistory 获取信号历史
func (pw *PriceWindow) GetSignalHistory() []SignalHistoryItem {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	result := make([]SignalHistoryItem, len(pw.SignalHistory))
	copy(result, pw.SignalHistory)
	return result
}

// AddSignalHistory 添加信号历史
func (pw *PriceWindow) AddSignalHistory(signal string) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.SignalHistory = append(pw.SignalHistory, SignalHistoryItem{
		Timestamp: time.Now().UnixMilli(),
		Signal:    signal,
	})

	// 限制历史数量
	if len(pw.SignalHistory) > MaxSignalHistorySize {
		pw.SignalHistory = pw.SignalHistory[len(pw.SignalHistory)-MaxSignalHistorySize:]
	}
}

// Clear 清空价格窗口
func (pw *PriceWindow) Clear() {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.Points = make([]PricePoint, 0, MaxPriceWindowSize)
	pw.LastAlertedLong = nil
	pw.LastAlertedShort = nil
	pw.LastWindowMin = nil
	pw.LastWindowMax = nil
	g.Log().Debugf(context.Background(), "[PriceWindow] 清空价格窗口数据")
}

// UpdateConfig 更新监控配置
func (pw *PriceWindow) UpdateConfig(window int, threshold float64) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if pw.Config == nil {
		pw.Config = &MonitorConfig{}
	}

	if window > 0 {
		pw.Config.Window = window
	}
	if threshold > 0 {
		pw.Config.Threshold = threshold
	}

	g.Log().Debugf(context.Background(), "[PriceWindow] 更新配置: window=%ds, threshold=%.4f",
		pw.Config.Window, pw.Config.Threshold)
}

// ResetAlerts 重置预警标记
func (pw *PriceWindow) ResetAlerts() {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.LastAlertedLong = nil
	pw.LastAlertedShort = nil
}

// UpdateLastWindow 更新上次窗口极值
func (pw *PriceWindow) UpdateLastWindow(minPrice, maxPrice float64) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	pw.LastWindowMin = &minPrice
	pw.LastWindowMax = &maxPrice
}

// CheckWindowChange 检查窗口基准价是否变化
func (pw *PriceWindow) CheckWindowChange(minPrice, maxPrice float64) (minChanged, maxChanged bool) {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	if pw.LastWindowMin != nil && minPrice != *pw.LastWindowMin {
		minChanged = true
	}
	if pw.LastWindowMax != nil && maxPrice != *pw.LastWindowMax {
		maxChanged = true
	}
	return
}

// SetLastSignal 设置上次信号
func (pw *PriceWindow) SetLastSignal(signal string) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	pw.LastWindowSignal = signal
}

// GetLastSignal 获取上次信号
func (pw *PriceWindow) GetLastSignal() string {
	pw.mu.RLock()
	defer pw.mu.RUnlock()
	return pw.LastWindowSignal
}

// SetAlertedLong 设置做多预警标记
func (pw *PriceWindow) SetAlertedLong(price float64) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	pw.LastAlertedLong = &price
}

// SetAlertedShort 设置做空预警标记
func (pw *PriceWindow) SetAlertedShort(price float64) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	pw.LastAlertedShort = &price
}

// GetAlertedLong 获取做多预警标记
func (pw *PriceWindow) GetAlertedLong() *float64 {
	pw.mu.RLock()
	defer pw.mu.RUnlock()
	return pw.LastAlertedLong
}

// GetAlertedShort 获取做空预警标记
func (pw *PriceWindow) GetAlertedShort() *float64 {
	pw.mu.RLock()
	defer pw.mu.RUnlock()
	return pw.LastAlertedShort
}

