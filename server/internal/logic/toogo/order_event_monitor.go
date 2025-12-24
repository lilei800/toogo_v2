// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Description 订单事件监控和分析 - 监控订单事件失败率、分析订单各阶段耗时等

package toogo

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// OrderEventStats 订单事件统计
type OrderEventStats struct {
	TotalEvents      int64   `json:"total_events"`
	SuccessEvents    int64   `json:"success_events"`
	FailedEvents     int64   `json:"failed_events"`
	PendingEvents    int64   `json:"pending_events"`
	FailureRate      float64 `json:"failure_rate"`      // 失败率（百分比）
	AvgProcessTime   float64 `json:"avg_process_time"`   // 平均处理时间（秒）
	EventTypeStats   map[string]int64 `json:"event_type_stats"` // 各事件类型统计
}

// OrderLifecycleStats 订单生命周期统计
type OrderLifecycleStats struct {
	OrderId           int64   `json:"order_id"`
	SignalToPreCreate float64 `json:"signal_to_pre_create"` // 信号生成到预创建耗时（秒）
	PreCreateToOrder  float64 `json:"pre_create_to_order"` // 预创建到下单耗时（秒）
	OrderToFilled     float64 `json:"order_to_filled"`      // 下单到成交耗时（秒）
	TotalLifecycle    float64 `json:"total_lifecycle"`       // 总生命周期耗时（秒）
	HasFailed         bool    `json:"has_failed"`           // 是否有失败事件
}

// GetOrderEventStats 获取订单事件统计（最近N小时）
func GetOrderEventStats(ctx context.Context, hours int) (*OrderEventStats, error) {
	stats := &OrderEventStats{
		EventTypeStats: make(map[string]int64),
	}

	// 计算时间范围
	startTime := gtime.Now().Add(-time.Duration(hours) * time.Hour)

	// 查询总事件数
	totalCount, err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Where("created_at >= ?", startTime).
		Count()
	if err != nil {
		return nil, err
	}
	stats.TotalEvents = int64(totalCount)

	// 查询成功事件数
	successCount, err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Where("created_at >= ?", startTime).
		Where("event_status", OrderEventStatusSuccess).
		Count()
	if err != nil {
		return nil, err
	}
	stats.SuccessEvents = int64(successCount)

	// 查询失败事件数
	failedCount, err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Where("created_at >= ?", startTime).
		Where("event_status", OrderEventStatusFailed).
		Count()
	if err != nil {
		return nil, err
	}
	stats.FailedEvents = int64(failedCount)

	// 查询待处理事件数
	pendingCount, err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Where("created_at >= ?", startTime).
		Where("event_status", OrderEventStatusPending).
		Count()
	if err != nil {
		return nil, err
	}
	stats.PendingEvents = int64(pendingCount)

	// 计算失败率
	if stats.TotalEvents > 0 {
		stats.FailureRate = float64(stats.FailedEvents) / float64(stats.TotalEvents) * 100
	}

	// 统计各事件类型
	typeStats, err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Fields("event_type, COUNT(*) as count").
		Where("created_at >= ?", startTime).
		Group("event_type").
		All()
	if err == nil && typeStats != nil {
		for _, row := range typeStats {
			eventType := row["event_type"].String()
			count := row["count"].Int64()
			stats.EventTypeStats[eventType] = count
		}
	}

	return stats, nil
}

// GetOrderLifecycleStats 获取订单生命周期统计
func GetOrderLifecycleStats(ctx context.Context, orderId int64) (*OrderLifecycleStats, error) {
	stats := &OrderLifecycleStats{
		OrderId: orderId,
	}

	// 查询该订单的所有事件
	var events []map[string]interface{}
	err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Where("order_id", orderId).
		OrderAsc("created_at").
		Scan(&events)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return stats, nil
	}

	// 记录各阶段的时间戳
	var signalTime, preCreateTime, orderTime, filledTime *gtime.Time
	var hasFailed bool

	for _, event := range events {
		eventType := ""
		eventStatus := ""
		var createdAt *gtime.Time
		
		// 类型转换
		if et, ok := event["event_type"]; ok {
			eventType = g.NewVar(et).String()
		}
		if es, ok := event["event_status"]; ok {
			eventStatus = g.NewVar(es).String()
		}
		if ca, ok := event["created_at"]; ok {
			if gt, ok := ca.(*gtime.Time); ok {
				createdAt = gt
			} else if gt, ok := ca.(gtime.Time); ok {
				createdAt = &gt
			}
		}

		if eventStatus == OrderEventStatusFailed {
			hasFailed = true
		}

		if createdAt == nil {
			continue
		}

		switch eventType {
		case OrderEventSignalGenerated:
			signalTime = createdAt
		case OrderEventPreCreated:
			preCreateTime = createdAt
		case OrderEventExchangeOrdered:
			orderTime = createdAt
		case OrderEventOrderFilled:
			filledTime = createdAt
		}
	}

	stats.HasFailed = hasFailed

	// 计算各阶段耗时
	if signalTime != nil && preCreateTime != nil {
		stats.SignalToPreCreate = preCreateTime.Time.Sub(signalTime.Time).Seconds()
	}
	if preCreateTime != nil && orderTime != nil {
		stats.PreCreateToOrder = orderTime.Time.Sub(preCreateTime.Time).Seconds()
	}
	if orderTime != nil && filledTime != nil {
		stats.OrderToFilled = filledTime.Time.Sub(orderTime.Time).Seconds()
	}
	if signalTime != nil && filledTime != nil {
		stats.TotalLifecycle = filledTime.Time.Sub(signalTime.Time).Seconds()
	}

	return stats, nil
}

// GetFailedOrders 获取失败的订单列表（最近N小时）
func GetFailedOrders(ctx context.Context, hours int) ([]map[string]interface{}, error) {
	startTime := gtime.Now().Add(-time.Duration(hours) * time.Hour)

	var failedOrders []map[string]interface{}
	err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Fields("order_id, exchange_order_id, event_type, event_status, event_message, created_at").
		Where("created_at >= ?", startTime).
		Where("event_status", OrderEventStatusFailed).
		OrderDesc("created_at").
		Scan(&failedOrders)

	return failedOrders, err
}

// AnalyzeOrderPerformance 分析订单性能（批量分析）
func AnalyzeOrderPerformance(ctx context.Context, orderIds []int64) (map[int64]*OrderLifecycleStats, error) {
	results := make(map[int64]*OrderLifecycleStats)

	for _, orderId := range orderIds {
		stats, err := GetOrderLifecycleStats(ctx, orderId)
		if err != nil {
			g.Log().Warningf(ctx, "[OrderEventMonitor] 分析订单生命周期失败: orderId=%d, err=%v", orderId, err)
			continue
		}
		results[orderId] = stats
	}

	return results, nil
}

// MonitorOrderEvents 监控订单事件（定期检查失败率）
func MonitorOrderEvents(ctx context.Context) {
	// 获取最近1小时的事件统计
	stats, err := GetOrderEventStats(ctx, 1)
	if err != nil {
		g.Log().Errorf(ctx, "[OrderEventMonitor] 获取事件统计失败: err=%v", err)
		return
	}

	// 如果失败率超过阈值，记录警告
	if stats.FailureRate > 5.0 { // 失败率超过5%
		g.Log().Warningf(ctx, "[OrderEventMonitor] ⚠️ 订单事件失败率过高: 失败率=%.2f%%, 总事件数=%d, 失败事件数=%d",
			stats.FailureRate, stats.TotalEvents, stats.FailedEvents)
	}

	// 记录统计信息
	g.Log().Infof(ctx, "[OrderEventMonitor] 订单事件统计（最近1小时）: 总数=%d, 成功=%d, 失败=%d, 待处理=%d, 失败率=%.2f%%",
		stats.TotalEvents, stats.SuccessEvents, stats.FailedEvents, stats.PendingEvents, stats.FailureRate)

	// 记录各事件类型统计
	for eventType, count := range stats.EventTypeStats {
		g.Log().Debugf(ctx, "[OrderEventMonitor] 事件类型统计: %s=%d", eventType, count)
	}
}

// GetOrderEventSummary 获取订单事件摘要（用于API返回）
func GetOrderEventSummary(ctx context.Context, orderId int64) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// 查询订单的所有事件
	var events []map[string]interface{}
	err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Where("order_id", orderId).
		OrderAsc("created_at").
		Scan(&events)
	if err != nil {
		return nil, err
	}

	summary["order_id"] = orderId
	summary["total_events"] = len(events)
	summary["events"] = events

	// 获取生命周期统计
	lifecycleStats, err := GetOrderLifecycleStats(ctx, orderId)
	if err == nil && lifecycleStats != nil {
		summary["lifecycle_stats"] = lifecycleStats
	}

	return summary, nil
}

// CleanOldOrderEvents 清理旧的订单事件（保留最近N天）
func CleanOldOrderEvents(ctx context.Context, days int) error {
	cutoffTime := gtime.Now().Add(-time.Duration(days) * time.Duration(24) * time.Hour)

	result, err := g.DB().Model("hg_trading_order_event").
		Ctx(ctx).
		Where("created_at < ?", cutoffTime).
		Delete()

	if err != nil {
		return err
	}

	deletedCount, _ := result.RowsAffected()
	g.Log().Infof(ctx, "[OrderEventMonitor] 已清理 %d 条旧的订单事件记录（保留最近 %d 天）", deletedCount, days)

	return nil
}

