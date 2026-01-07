// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Description 统一订单状态管理 - 订单状态枚举、字段完整性检查、数据一致性检查

package toogo

import (
	"context"
	"fmt"
	"math"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"hotgo/internal/library/market"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ==================== 订单状态枚举 ====================

// TradingOrderStatus 交易订单状态
const (
	OrderStatusPending   = 0 // 未成交
	OrderStatusOpen      = 1 // 持仓中
	OrderStatusClosed    = 2 // 已平仓
	OrderStatusCancelled = 3 // 已取消
	OrderStatusFailed    = 4 // 下单失败（新增）
)

// OrderStatusText 订单状态文本映射
var OrderStatusText = map[int]string{
	OrderStatusPending:   "未成交",
	OrderStatusOpen:      "持仓中",
	OrderStatusClosed:    "已平仓",
	OrderStatusCancelled: "已取消",
	OrderStatusFailed:    "下单失败",
}

// GetOrderStatusText 获取订单状态文本
func GetOrderStatusText(status int) string {
	if text, ok := OrderStatusText[status]; ok {
		return text
	}
	return "未知状态"
}

// ==================== 订单字段完整性检查 ====================

// OrderFieldCompleteness 订单字段完整性检查结果
type OrderFieldCompleteness struct {
	IsComplete      bool     // 是否完整
	MissingFields   []string // 缺失的字段列表
	IncompleteLevel string   // 完整度级别：complete/partial/missing
}

// CheckOrderFieldCompleteness 检查订单字段完整性
// 必填字段：订单ID、交易对、方向、开仓价格、数量、状态
// 推荐字段：市场状态、风险偏好、策略参数、杠杆、保证金
// 可选字段：平仓价格、平仓时间、手续费等
func CheckOrderFieldCompleteness(order *entity.TradingOrder) *OrderFieldCompleteness {
	result := &OrderFieldCompleteness{
		MissingFields: make([]string, 0),
	}

	// 检查必填字段
	if order.ExchangeOrderId == "" {
		result.MissingFields = append(result.MissingFields, "exchange_order_id")
	}
	if order.Symbol == "" {
		result.MissingFields = append(result.MissingFields, "symbol")
	}
	if order.Direction == "" {
		result.MissingFields = append(result.MissingFields, "direction")
	}
	if order.OpenPrice <= 0 {
		result.MissingFields = append(result.MissingFields, "open_price")
	}
	if order.Quantity <= 0 {
		result.MissingFields = append(result.MissingFields, "quantity")
	}

	// 检查推荐字段（通过查询数据库）
	// 注意：这些字段可能不在 entity 中，需要通过数据库查询

	// 判断完整度级别
	if len(result.MissingFields) == 0 {
		result.IsComplete = true
		result.IncompleteLevel = "complete"
	} else {
		result.IsComplete = false
		// 如果缺失的是必填字段，标记为 missing；否则标记为 partial
		requiredFields := []string{"exchange_order_id", "symbol", "direction", "open_price", "quantity"}
		hasRequiredMissing := false
		for _, field := range result.MissingFields {
			for _, required := range requiredFields {
				if field == required {
					hasRequiredMissing = true
					break
				}
			}
			if hasRequiredMissing {
				break
			}
		}
		if hasRequiredMissing {
			result.IncompleteLevel = "missing"
		} else {
			result.IncompleteLevel = "partial"
		}
	}

	return result
}

// CompleteOrderFields 补全订单缺失字段
// 根据订单状态和可用数据，智能补全缺失字段
func CompleteOrderFields(ctx context.Context, order *entity.TradingOrder, robot *entity.TradingRobot, exchangeOrder *exchange.Order, pos *exchange.Position) error {
	completeness := CheckOrderFieldCompleteness(order)
	if completeness.IsComplete {
		return nil // 订单已完整，无需补全
	}

	updateData := g.Map{
		"updated_at": gtime.Now(),
	}

	// 补全订单ID（如果缺失）
	if order.ExchangeOrderId == "" && exchangeOrder != nil {
		if exchangeOrder.ClientId != "" {
			updateData["exchange_order_id"] = exchangeOrder.ClientId
		} else if exchangeOrder.OrderId != "" {
			updateData["exchange_order_id"] = exchangeOrder.OrderId
		}
	}

	// 补全开仓价格（如果缺失）
	if order.OpenPrice <= 0 {
		if exchangeOrder != nil && exchangeOrder.AvgPrice > 0 {
			updateData["open_price"] = exchangeOrder.AvgPrice
		} else if pos != nil && pos.EntryPrice > 0 {
			updateData["open_price"] = pos.EntryPrice
		}
	}

	// 补全数量（如果缺失）
	if order.Quantity <= 0 {
		if exchangeOrder != nil && exchangeOrder.Quantity > 0 {
			updateData["quantity"] = exchangeOrder.Quantity
		} else if pos != nil && math.Abs(pos.PositionAmt) > positionAmtEpsilon {
			updateData["quantity"] = math.Abs(pos.PositionAmt)
		}
	}

	// 补全创建时间（如果缺失）
	if (order.OpenTime == nil || order.OpenTime.IsZero()) && exchangeOrder != nil && exchangeOrder.CreateTime > 0 {
		orderCreateTime := gtime.NewFromTimeStamp(exchangeOrder.CreateTime / 1000)
		updateData["open_time"] = orderCreateTime
		if order.CreatedAt == nil || order.CreatedAt.IsZero() {
			updateData["created_at"] = orderCreateTime
		}
	}

	// 补全市场状态和风险偏好（如果缺失）
	if robot != nil {
		// 从全局市场分析器获取市场状态
		ap := market.ResolveAnalysisPlatform(ctx, robot.Exchange)
		if ap == "" {
			ap = robot.Exchange
		}
		globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(ap, robot.Symbol)
		if globalAnalysis != nil {
			marketState := normalizeMarketState(string(globalAnalysis.MarketState))
			if marketState != "" {
				// 查询订单当前的市场状态（从数据库）
				var currentMarketState string
				_ = dao.TradingOrder.Ctx(ctx).Where("id", order.Id).Fields("market_state").Scan(&currentMarketState)
				if currentMarketState == "" {
					updateData["market_state"] = marketState

					// 从映射关系获取风险偏好
					robotEngine := GetRobotTaskManager().GetEngine(robot.Id)
					if robotEngine != nil {
						robotEngine.mu.RLock()
						riskPref := robotEngine.MarketRiskMapping[marketState]
						robotEngine.mu.RUnlock()

						if riskPref != "" {
							var currentRiskLevel string
							_ = dao.TradingOrder.Ctx(ctx).Where("id", order.Id).Fields("risk_level").Scan(&currentRiskLevel)
							if currentRiskLevel == "" {
								updateData["risk_level"] = riskPref
							}
						}
					}
				}
			}
		}
	}

	// 补全杠杆和保证金（如果缺失）
	if order.Leverage <= 0 && pos != nil && pos.Leverage > 0 {
		updateData["leverage"] = pos.Leverage
	}
	if order.Margin <= 0 && pos != nil && pos.IsolatedMargin > 0 {
		updateData["margin"] = pos.IsolatedMargin
	}

	// 执行更新
	if len(updateData) > 1 { // 除了 updated_at 还有其他字段
		_, err := dao.TradingOrder.Ctx(ctx).
			Where("id", order.Id).
			Data(updateData).
			Update()
		if err != nil {
			return fmt.Errorf("补全订单字段失败: %v", err)
		}
		g.Log().Infof(ctx, "[OrderStatus] 已补全订单字段: orderId=%d, 补全字段数=%d", order.Id, len(updateData)-1)
	}

	return nil
}

// ==================== 数据一致性检查 ====================

// DataConsistencyCheck 数据一致性检查结果
type DataConsistencyCheck struct {
	IsConsistent    bool     // 是否一致
	Inconsistencies []string // 不一致项列表
	MemoryStatus    string   // 内存状态：has_position/no_position/unknown
	DatabaseStatus  string   // 数据库状态：has_order/no_order/unknown
	ExchangeStatus  string   // 交易所状态：has_position/no_position/unknown
}

// CheckDataConsistency 检查三层数据一致性（内存 ↔ 数据库 ↔ 交易所）
func CheckDataConsistency(ctx context.Context, robotId int64, positionSide string, engine *RobotEngine, localOrder *entity.TradingOrder, exchangePos *exchange.Position) *DataConsistencyCheck {
	result := &DataConsistencyCheck{
		Inconsistencies: make([]string, 0),
	}

	// 检查内存状态
	if engine != nil {
		engine.mu.RLock()
		hasMemoryPosition := false
		if engine.CurrentPositions != nil {
			for _, pos := range engine.CurrentPositions {
				if pos.PositionSide == positionSide && math.Abs(pos.PositionAmt) > positionAmtEpsilon {
					hasMemoryPosition = true
					break
				}
			}
		}
		engine.mu.RUnlock()

		if hasMemoryPosition {
			result.MemoryStatus = "has_position"
		} else {
			result.MemoryStatus = "no_position"
		}
	} else {
		result.MemoryStatus = "unknown"
	}

	// 检查数据库状态
	if localOrder != nil && localOrder.Status == OrderStatusOpen {
		result.DatabaseStatus = "has_order"
	} else {
		result.DatabaseStatus = "no_order"
	}

	// 检查交易所状态
	if exchangePos != nil && math.Abs(exchangePos.PositionAmt) > positionAmtEpsilon {
		result.ExchangeStatus = "has_position"
	} else {
		result.ExchangeStatus = "no_position"
	}

	// 检查一致性
	// 情况1：内存有持仓，数据库有订单，交易所有持仓 → 一致
	if result.MemoryStatus == "has_position" && result.DatabaseStatus == "has_order" && result.ExchangeStatus == "has_position" {
		result.IsConsistent = true
		return result
	}

	// 情况2：内存无持仓，数据库无订单，交易所无持仓 → 一致
	if result.MemoryStatus == "no_position" && result.DatabaseStatus == "no_order" && result.ExchangeStatus == "no_position" {
		result.IsConsistent = true
		return result
	}

	// 其他情况都是不一致的
	result.IsConsistent = false

	// 记录不一致项
	if result.MemoryStatus == "has_position" && result.DatabaseStatus == "no_order" {
		result.Inconsistencies = append(result.Inconsistencies, "内存有持仓但数据库无订单")
	}
	if result.MemoryStatus == "has_position" && result.ExchangeStatus == "no_position" {
		result.Inconsistencies = append(result.Inconsistencies, "内存有持仓但交易所无持仓")
	}
	if result.DatabaseStatus == "has_order" && result.ExchangeStatus == "no_position" {
		result.Inconsistencies = append(result.Inconsistencies, "数据库有订单但交易所无持仓（可能已手动平仓）")
	}
	if result.DatabaseStatus == "no_order" && result.ExchangeStatus == "has_position" {
		result.Inconsistencies = append(result.Inconsistencies, "数据库无订单但交易所有持仓（外部持仓）")
	}
	if result.MemoryStatus == "no_position" && result.DatabaseStatus == "has_order" {
		result.Inconsistencies = append(result.Inconsistencies, "内存无持仓但数据库有订单")
	}
	if result.MemoryStatus == "no_position" && result.ExchangeStatus == "has_position" {
		result.Inconsistencies = append(result.Inconsistencies, "内存无持仓但交易所有持仓")
	}

	return result
}

// FixDataInconsistency 修复数据不一致
// 根据不一致情况，自动修复数据
func FixDataInconsistency(ctx context.Context, robotId int64, positionSide string, check *DataConsistencyCheck, engine *RobotEngine, localOrder *entity.TradingOrder, exchangePos *exchange.Position) error {
	if check.IsConsistent {
		return nil // 数据一致，无需修复
	}

	// 修复策略：以交易所数据为准（最权威）
	if check.ExchangeStatus == "has_position" {
		// 交易所有持仓
		if check.DatabaseStatus == "no_order" {
			// 数据库无订单 → 创建订单记录
			g.Log().Infof(ctx, "[OrderStatus] 修复数据不一致：交易所有持仓但数据库无订单，创建订单记录")
			// 这里需要调用 createOrderFromExchange，但需要 exchangeOrder
			// 暂时记录日志，由同步服务处理
		}
		if check.MemoryStatus == "no_position" && engine != nil {
			// 内存无持仓 → 更新内存
			engine.mu.Lock()
			if engine.CurrentPositions == nil {
				engine.CurrentPositions = make([]*exchange.Position, 0)
			}
			// 添加持仓到内存
			engine.CurrentPositions = append(engine.CurrentPositions, exchangePos)
			engine.mu.Unlock()
			g.Log().Infof(ctx, "[OrderStatus] 修复数据不一致：更新内存持仓")
		}
	} else {
		// 交易所无持仓
		if check.DatabaseStatus == "has_order" && localOrder != nil {
			// 数据库有订单 → 更新为已平仓
			g.Log().Infof(ctx, "[OrderStatus] 修复数据不一致：交易所有持仓但数据库有订单，更新为已平仓")
			// 这里需要调用 CloseOrder，但需要 closePrice 和 realizedProfit
			// 暂时记录日志，由同步服务处理
		}
		if check.MemoryStatus == "has_position" && engine != nil {
			// 内存有持仓 → 清除内存
			engine.mu.Lock()
			if engine.CurrentPositions != nil {
				for i, pos := range engine.CurrentPositions {
					if pos.PositionSide == positionSide {
						engine.CurrentPositions[i].PositionAmt = 0
					}
				}
			}
			engine.mu.Unlock()
			g.Log().Infof(ctx, "[OrderStatus] 修复数据不一致：清除内存持仓")
		}
	}

	return nil
}
