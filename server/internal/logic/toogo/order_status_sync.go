// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Description 订单状态实时同步服务 - 同步交易所订单状态到本地数据库

package toogo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"hotgo/internal/library/market"
	"hotgo/internal/model/entity"
	"hotgo/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
)

// tryCloseInfoFromTradeHistory 尝试从交易所成交记录中推断“平仓价/已实现盈亏/手续费/平仓时间”
// 适用场景：检测到交易所已无持仓，但本地订单仍为 OPEN（疑似交易所/外部手动平仓）。
//
// 返回：
// - closePrice: 平仓均价（按成交数量加权）
// - realizedProfit: 已实现盈亏（交易所返回的 realizedPnl 汇总）
// - closeFee: 平仓手续费（交易所成交记录 commission 汇总）
// - closeFeeCoin: 平仓手续费币种（若可识别）
// - closeOrderId: 平仓订单ID（交易所订单ID，用于对账）
// - ts: 平仓时间戳（毫秒）
// - ok: 是否成功从成交记录中获得信息
func tryCloseInfoFromTradeHistory(
	ctx context.Context,
	ex exchange.Exchange,
	symbol string,
	positionSide string,
	openTime *gtime.Time,
) (closePrice float64, realizedProfit float64, closeFee float64, closeFeeCoin string, closeOrderId string, ts int64, ok bool) {
	// 轻量接口：只要求实现 GetTradeHistory（避免强依赖 ExchangeAdvanced 全家桶方法）
	type tradeHistoryProvider interface {
		GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*exchange.Trade, error)
	}
	p, okP := ex.(tradeHistoryProvider)
	if !okP {
		return 0, 0, 0, "", "", 0, false
	}

	// 基于开仓时间过滤（避免抓到更早的历史成交）
	startMs := int64(0)
	if openTime != nil && !openTime.IsZero() {
		startMs = openTime.UnixMilli()
	}

	ps := strings.TrimSpace(positionSide)
	// 平仓成交方向：
	// - LONG 平仓通常为 SELL
	// - SHORT 平仓通常为 BUY
	closeSide := "SELL"
	if strings.EqualFold(ps, "SHORT") {
		closeSide = "BUY"
	}

	trades, err := p.GetTradeHistory(ctx, symbol, 200)
	if err != nil || len(trades) == 0 {
		return 0, 0, 0, "", "", 0, false
	}

	// 过滤候选成交：同交易对 + 同持仓方向 + 平仓方向 + 时间在开仓之后
	candidates := make([]*exchange.Trade, 0, len(trades))
	for _, t := range trades {
		if t == nil {
			continue
		}
		if startMs > 0 && t.Time > 0 && t.Time < startMs {
			continue
		}
		if t.PositionSide != "" && !strings.EqualFold(strings.TrimSpace(t.PositionSide), ps) {
			continue
		}
		if t.Side != "" && !strings.EqualFold(strings.TrimSpace(t.Side), closeSide) {
			continue
		}
		candidates = append(candidates, t)
	}
	if len(candidates) == 0 {
		return 0, 0, 0, "", "", 0, false
	}

	// 找出最新的一笔成交（用于锁定平仓订单ID）
	latest := candidates[0]
	for _, t := range candidates[1:] {
		if t.Time > latest.Time {
			latest = t
		}
	}

	// 如果有 orderId，则汇总该 orderId 的多笔成交，得到加权均价/总PnL
	targetOrderID := strings.TrimSpace(latest.OrderId)
	var (
		sumQty      float64
		sumPriceQty float64
		sumPnl      float64
		sumFee      float64
		feeCoin     string
		maxTs       int64
	)
	for _, t := range candidates {
		if targetOrderID != "" && strings.TrimSpace(t.OrderId) != targetOrderID {
			continue
		}
		if t.Quantity > 0 && t.Price > 0 {
			sumQty += t.Quantity
			sumPriceQty += t.Price * t.Quantity
		}
		sumPnl += t.RealizedPnl
		sumFee += t.Commission
		if feeCoin == "" && strings.TrimSpace(t.CommissionAsset) != "" {
			feeCoin = strings.TrimSpace(t.CommissionAsset)
		}
		if t.Time > maxTs {
			maxTs = t.Time
		}
	}

	if sumQty > 0 {
		closePrice = sumPriceQty / sumQty
	} else {
		closePrice = latest.Price
	}
	realizedProfit = sumPnl
	closeFee = sumFee
	closeFeeCoin = feeCoin
	closeOrderId = targetOrderID
	ts = maxTs

	// 基础有效性判断
	if closePrice <= 0 || ts <= 0 {
		return 0, 0, 0, "", "", 0, false
	}
	return closePrice, realizedProfit, closeFee, closeFeeCoin, closeOrderId, ts, true
}

// tradeAggByOrderId 成交汇总（按交易所 orderId 聚合）
// 说明：用于把“成交(fill)级别”汇总为订单级别的 平均价/数量/已实现盈亏/手续费/时间。
type tradeAggByOrderId struct {
	OrderId      string
	SumQty       float64
	SumPriceQty  float64
	AvgPrice     float64
	RealizedPnl  float64
	Commission   float64
	FeeCoin      string
	MinTs        int64
	MaxTs        int64
	HasAnyRecord bool
}

func (a *tradeAggByOrderId) add(t *exchange.Trade) {
	if t == nil {
		return
	}
	if strings.TrimSpace(t.OrderId) == "" {
		return
	}
	a.HasAnyRecord = true
	if a.OrderId == "" {
		a.OrderId = strings.TrimSpace(t.OrderId)
	}
	if t.Quantity > 0 && t.Price > 0 {
		a.SumQty += t.Quantity
		a.SumPriceQty += t.Price * t.Quantity
	}
	a.RealizedPnl += t.RealizedPnl
	// 统一手续费口径：很多交易所 fee/commission 返回为负数（扣费），这里统一按“正数金额”汇总
	// 这样上层在展示“手续费(USDT)”时不会因为符号问题误判为缺失。
	a.Commission += math.Abs(t.Commission)
	if a.FeeCoin == "" && strings.TrimSpace(t.CommissionAsset) != "" {
		a.FeeCoin = strings.TrimSpace(t.CommissionAsset)
	}
	if t.Time > 0 {
		if a.MinTs == 0 || t.Time < a.MinTs {
			a.MinTs = t.Time
		}
		if t.Time > a.MaxTs {
			a.MaxTs = t.Time
		}
	}
}

func (a *tradeAggByOrderId) finalize() {
	if a.SumQty > 0 {
		a.AvgPrice = a.SumPriceQty / a.SumQty
	}
}

// tryAggFromTradeHistoryByOrderID 从交易所成交记录中按 orderId 汇总信息（更精确：不靠“方向+时间”猜测）
func tryAggFromTradeHistoryByOrderID(
	ctx context.Context,
	ex exchange.Exchange,
	symbol string,
	orderID string,
	limit int,
) (agg tradeAggByOrderId, ok bool) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return tradeAggByOrderId{}, false
	}

	type tradeHistoryProvider interface {
		GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*exchange.Trade, error)
	}
	p, okP := ex.(tradeHistoryProvider)
	if !okP {
		return tradeAggByOrderId{}, false
	}
	if limit <= 0 {
		limit = 500
	}

	trades, err := p.GetTradeHistory(ctx, symbol, limit)
	if err != nil || len(trades) == 0 {
		return tradeAggByOrderId{}, false
	}

	for _, t := range trades {
		if t == nil {
			continue
		}
		if strings.TrimSpace(t.OrderId) != orderID {
			continue
		}
		agg.add(t)
	}
	if !agg.HasAnyRecord {
		return tradeAggByOrderId{}, false
	}
	agg.finalize()
	return agg, true
}

// OrderStatusSyncService 订单状态同步服务
type OrderStatusSyncService struct {
	mu            sync.RWMutex
	running       bool
	stopCh        chan struct{}
	syncingRobots map[int64]bool // 正在同步的机器人ID映射，避免并发同步
	syncMutex     sync.Mutex     // 保护 syncingRobots 的互斥锁

	// triggerCh: 事件驱动的“按robot触发同步”请求（来自私有WS/关键业务事件）
	triggerCh chan int64
	// lastTriggerAt: 简单去抖（避免同一机器人短时间触发太多次）
	lastTriggerAt map[int64]time.Time

	// lastOpenOrdersSyncAt: openOrders REST 兜底对账的节流（避免页面频繁刷新仍打交易所）
	openOrdersMu         sync.Mutex
	lastOpenOrdersSyncAt map[int64]time.Time
}

var (
	orderStatusSyncService     *OrderStatusSyncService
	orderStatusSyncServiceOnce sync.Once
)

// GetOrderStatusSyncService 获取订单状态同步服务单例
func GetOrderStatusSyncService() *OrderStatusSyncService {
	orderStatusSyncServiceOnce.Do(func() {
		orderStatusSyncService = &OrderStatusSyncService{
			stopCh:        make(chan struct{}),
			syncingRobots: make(map[int64]bool),
			triggerCh:     make(chan int64, 1024),
			lastTriggerAt: make(map[int64]time.Time),
			lastOpenOrdersSyncAt: make(map[int64]time.Time),
		}
	})
	return orderStatusSyncService
}

// Start 启动订单状态同步服务
func (s *OrderStatusSyncService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	g.Log().Info(ctx, "[OrderStatusSync] 订单状态同步服务启动")

	// 启动定时同步任务
	go s.runSyncLoop(ctx)

	return nil
}

// Stop 停止同步服务
func (s *OrderStatusSyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)
	g.Log().Info(context.Background(), "[OrderStatusSync] 订单状态同步服务已停止")
}

// IsRunning 检查是否运行中
func (s *OrderStatusSyncService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// runSyncLoop 同步循环（定期同步整个订单信息）
// 【简化策略】整个订单每1秒同步一次，包括：订单状态、持仓、未实现盈亏、开仓价格等完整信息
// 【事件驱动优化】订单未实现盈亏已改为事件驱动（价格更新时立即更新）
// 【事件驱动优化】手动平仓等事件已改为事件驱动（检测到时立即同步）
// 同时定期检测外部持仓（系统外部创建的持仓，不属于事件）
func (s *OrderStatusSyncService) runSyncLoop(ctx context.Context) {
	// 【健壮性优化】添加 panic 恢复机制，确保同步循环异常不会导致服务停止
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] runSyncLoop panic recovered: err=%v", r)
			// 重启同步循环（延迟重启，避免快速循环）
			time.Sleep(5 * time.Second)
			go s.runSyncLoop(ctx)
		}
	}()

	// 【事件驱动优化】大幅降低同步频率，减少API请求：
	// - 订单历史同步：每60秒一次（作为兜底，主要依赖事件驱动）
	// - 外部持仓检测：已禁用（意义不大，30秒检测一次无法及时响应）
	// 主要同步时机：开仓前、平仓后（事件驱动）
	orderSyncTicker := time.NewTicker(60 * time.Second)
	defer orderSyncTicker.Stop()

	// 【优化】外部持仓检测不再定期轮询
	// 改为在下单前检查：如果检测到外部持仓，自动创建订单记录
	// 这样既能发现外部下单，又不浪费API调用

	// 启动时执行一次同步（获取初始状态）
	s.SyncAllOrders(ctx)

	for {
		select {
		case <-s.stopCh:
			return
		case robotId := <-s.triggerCh:
			s.syncRobotByIdTriggered(ctx, robotId)
		case <-orderSyncTicker.C:
			// 每60秒同步一次订单历史（兜底同步，确保最终一致性）
			s.SyncAllOrders(ctx)
		}
	}
}

// TriggerRobotSync 触发单个机器人订单/持仓同步（事件驱动）
// - 非阻塞：队列满则丢弃（避免WS风暴拖垮服务）
// - 去抖：同一 robotId 1 秒内只触发一次
func (s *OrderStatusSyncService) TriggerRobotSync(robotId int64) {
	if robotId <= 0 {
		return
	}
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()
	if !running {
		return
	}

	now := time.Now()
	s.syncMutex.Lock()
	last := s.lastTriggerAt[robotId]
	if !last.IsZero() && now.Sub(last) < 1*time.Second {
		s.syncMutex.Unlock()
		return
	}
	s.lastTriggerAt[robotId] = now
	s.syncMutex.Unlock()

	select {
	case s.triggerCh <- robotId:
	default:
		// drop
	}
}

func (s *OrderStatusSyncService) syncRobotByIdTriggered(ctx context.Context, robotId int64) {
	if robotId <= 0 {
		return
	}

	// 防并发：同一robot不允许并发sync
	s.syncMutex.Lock()
	if s.syncingRobots[robotId] {
		s.syncMutex.Unlock()
		return
	}
	s.syncingRobots[robotId] = true
	s.syncMutex.Unlock()

	defer func() {
		s.syncMutex.Lock()
		delete(s.syncingRobots, robotId)
		s.syncMutex.Unlock()
	}()

	var robot *entity.TradingRobot
	if err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot); err != nil || robot == nil {
		return
	}
	// 只同步运行/有意义的机器人（状态过滤可按需调整）
	s.syncRobotOrders(ctx, robot)
}

// SyncAllOrders 同步有持仓的机器人的订单历史（兜底同步）
// 【优化】只同步有持仓订单的机器人，减少API调用
// 无持仓的机器人不需要同步订单历史
func (s *OrderStatusSyncService) SyncAllOrders(ctx context.Context) {
	// 【健壮性优化】添加 panic 恢复机制，确保同步失败不影响后续同步
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] SyncAllOrders panic recovered: err=%v", r)
		}
	}()

	// 【优化】只查询有持仓订单的机器人（通过 trading_order 表关联）
	// 这样可以大幅减少需要同步的机器人数量
	var robots []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where("status", 2). // 运行中
		Where("id IN (?)",
			dao.TradingOrder.Ctx(ctx).
				Fields("DISTINCT robot_id").
				Where("status", 1)). // 有持仓中订单
		Scan(&robots)
	if err != nil {
		// 如果子查询失败，降级为查询所有运行中的机器人
		err = dao.TradingRobot.Ctx(ctx).Where("status", 2).Scan(&robots)
		if err != nil {
			g.Log().Warningf(ctx, "[OrderStatusSync] 查询机器人失败: %v", err)
			return
		}
	}

	if len(robots) == 0 {
		return
	}

	// 【统一同步】同步所有运行中的机器人的整个订单信息（每秒一次，确保数据实时性）
	var wg sync.WaitGroup
	for _, robot := range robots {
		wg.Add(1)
		go func(robot *entity.TradingRobot) {
			defer wg.Done()
			// 【健壮性优化】每个机器人的同步都有独立的 panic 恢复
			defer func() {
				if p := recover(); p != nil {
					g.Log().Errorf(ctx, "[OrderStatusSync] syncRobotOrders panic recovered: robotId=%d, err=%v", robot.Id, p)
				}
			}()
			s.syncRobotOrders(ctx, robot)
		}(robot)
	}
	wg.Wait()
}

// SyncExternalPositions 检测并同步外部持仓（定期执行）
// 【事件驱动优化】只检测系统外部创建的持仓（不属于事件，需要定期检测）
// 包括：检测交易所持仓但本地无订单记录的情况，创建订单记录
func (s *OrderStatusSyncService) SyncExternalPositions(ctx context.Context) {
	// 【健壮性优化】添加 panic 恢复机制，确保同步失败不影响后续同步
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] SyncExternalPositions panic recovered: err=%v", r)
		}
	}()

	// 查询所有运行中的机器人
	var robots []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where("status", 2).Scan(&robots)
	if err != nil {
		g.Log().Warningf(ctx, "[OrderStatusSync] 查询机器人失败: %v", err)
		return
	}

	if len(robots) == 0 {
		return
	}

	// 【优化】只检测外部持仓：交易所持仓但本地无订单记录的情况
	var wg sync.WaitGroup
	for _, robot := range robots {
		wg.Add(1)
		go func(robot *entity.TradingRobot) {
			defer wg.Done()
			// 【健壮性优化】每个机器人的同步都有独立的 panic 恢复
			defer func() {
				if p := recover(); p != nil {
					g.Log().Errorf(ctx, "[OrderStatusSync] syncExternalPositions panic recovered: robotId=%d, err=%v", robot.Id, p)
				}
			}()
			s.syncExternalPositionsForRobot(ctx, robot)
		}(robot)
	}
	wg.Wait()
}

// syncExternalPositionsForRobot 检测单个机器人的外部持仓
// 【优化】只检测外部持仓：交易所持仓但本地无订单记录的情况
func (s *OrderStatusSyncService) syncExternalPositionsForRobot(ctx context.Context, robot *entity.TradingRobot) {
	// 获取交易所实例
	engine := GetRobotTaskManager().GetEngine(robot.Id)
	var ex exchange.Exchange

	if engine != nil && engine.Exchange != nil {
		ex = engine.Exchange
	} else {
		// RobotEngine 未运行时，临时创建 Exchange 实例
		var apiConfig *entity.TradingApiConfig
		err := dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig)
		if err != nil || apiConfig == nil {
			return
		}

		var exErr error
		ex, exErr = GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
		if exErr != nil {
			return
		}
	}

	// 【性能优化】获取交易所持仓（轻量级API调用）
	positions, err := ex.GetPositions(ctx, robot.Symbol)
	if err != nil {
		return
	}

	// 构建交易所持仓映射
	exchangePositions := make(map[string]*exchange.Position)
	for _, pos := range positions {
		if math.Abs(pos.PositionAmt) > 0.0001 {
			exchangePositions[pos.PositionSide] = pos
		}
	}

	// 【性能优化】无持仓时快速返回，避免不必要的数据库查询
	if len(exchangePositions) == 0 {
		return
	}

	// 查询本地"持仓中"的订单
	var localOrders []*entity.TradingOrder
	err = dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("status", OrderStatusOpen). // 持仓中（使用统一的订单状态常量）
		Scan(&localOrders)
	if err != nil {
		return
	}

	// 构建本地订单方向映射
	localOrderSides := make(map[string]bool)
	for _, order := range localOrders {
		positionSide := "LONG"
		if order.Direction == "short" {
			positionSide = "SHORT"
		}
		localOrderSides[positionSide] = true
	}

	// 检测外部持仓：交易所有持仓但本地无订单记录
	for positionSide, exchangePos := range exchangePositions {
		if !localOrderSides[positionSide] {
			// 检测到外部持仓，创建订单记录
			g.Log().Infof(ctx, "[OrderStatusSync] 【定期检测】检测到外部持仓: robotId=%d, positionSide=%s, 创建订单记录", robot.Id, positionSide)

			// 获取历史订单，查找匹配的开仓订单
			var historyOrders []*exchange.Order
			// 优先使用引擎缓存（避免重复打平台API）
			if engine != nil {
				historyOrders, _ = engine.GetCachedOrderHistory()
				engine.mu.RLock()
				last := engine.LastOrderHistoryUpdate
				engine.mu.RUnlock()
				if time.Since(last) >= 120*time.Second {
					historyOrders, _ = engine.GetOrderHistorySmart(ctx, 0, 50)
				}
			} else {
				var err error
				historyOrders, err = ex.GetOrderHistory(ctx, robot.Symbol, 50)
				if err != nil {
					g.Log().Debugf(ctx, "[OrderStatusSync] 获取历史订单失败: robotId=%d, err=%v", robot.Id, err)
					continue
				}
			}

			// 查找匹配的开仓订单
			var matchedOrder *exchange.Order
			for _, histOrder := range historyOrders {
				isOpenOrder := strings.Contains(histOrder.Type, "开仓") || strings.Contains(strings.ToLower(histOrder.Type), "open")
				orderPositionSide := ""
				if isOpenOrder {
					if histOrder.Side == "BUY" {
						orderPositionSide = "LONG"
					} else if histOrder.Side == "SELL" {
						orderPositionSide = "SHORT"
					}
				}
				if histOrder.PositionSide != "" {
					orderPositionSide = histOrder.PositionSide
				}

				if orderPositionSide == positionSide && isOpenOrder && histOrder.Status == "FILLED" {
					orderIdToCheck := histOrder.ClientId
					if orderIdToCheck == "" {
						orderIdToCheck = histOrder.OrderId
					}

					exists, _ := dao.TradingOrder.Ctx(ctx).
						Where("robot_id", robot.Id).
						Where("exchange_order_id", orderIdToCheck).
						Count()
					if exists == 0 {
						matchedOrder = histOrder
						matchedOrder.PositionSide = orderPositionSide
						break
					}
				}
			}

			if matchedOrder != nil {
				s.createOrderFromExchange(ctx, robot, matchedOrder, exchangePos)
			}
		}
	}
}

// syncRobotOrders 同步单个机器人的整个订单信息
// 【简化策略】每1秒同步一次，包括：订单状态、持仓、未实现盈亏、开仓价格等完整信息
// 【优化】优先使用 RobotEngine 缓存的数据，避免重复调用 API
func (s *OrderStatusSyncService) syncRobotOrders(ctx context.Context, robot *entity.TradingRobot) {
	// 【健壮性优化】添加 panic 恢复机制，确保单个机器人同步失败不影响其他机器人
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] syncRobotOrders panic recovered: robotId=%d, err=%v", robot.Id, r)
		}
		// 确保清除同步标记
		s.syncMutex.Lock()
		delete(s.syncingRobots, robot.Id)
		s.syncMutex.Unlock()
	}()

	// 检查是否正在同步该机器人，避免并发同步
	s.syncMutex.Lock()
	if s.syncingRobots[robot.Id] {
		s.syncMutex.Unlock()
		g.Log().Debugf(ctx, "[OrderStatusSync] 机器人正在同步中，跳过: robotId=%d", robot.Id)
		return
	}
	s.syncingRobots[robot.Id] = true
	s.syncMutex.Unlock()

	// 【修复】重新启用历史订单同步到数据库（用于历史交易记录展示）
	// 同步频率降低：统一使用 engine 的缓存/节流，避免重复打平台API
	engine := GetRobotTaskManager().GetEngine(robot.Id)
	if engine == nil {
		return
	}

	// 先拿缓存（用于后续持仓对账，避免重复 API）
	cachedPositions, _ := engine.GetCachedPositions()
	historyOrders, _ := engine.GetCachedOrderHistory()

	// 检查距离上次订单历史同步是否超过60秒
	engine.mu.RLock()
	lastSync := engine.LastOrderHistoryUpdate
	engine.mu.RUnlock()

	// 对齐 RobotEngine 的 120s 窗口：同一机器人最多 120 秒打一次历史订单接口
	if time.Since(lastSync) >= 120*time.Second {
		// 获取订单历史并同步到数据库（统一走引擎智能缓存 + singleflight）
		fetchedOrders, err := engine.GetOrderHistorySmart(ctx, 0, 50) // 进入此分支即视为需要强制刷新
		if err != nil {
			g.Log().Debugf(ctx, "[OrderStatusSync] 获取订单历史失败: robotId=%d, err=%v", robot.Id, err)
		} else {
			historyOrders = fetchedOrders
			// 同步到数据库
			if len(fetchedOrders) > 0 {
				if err := service.ToogoRobot().SyncOrderHistoryToDB(ctx, robot.Id, robot, fetchedOrders); err != nil {
					g.Log().Warningf(ctx, "[OrderStatusSync] 同步订单历史到数据库失败: robotId=%d, err=%v", robot.Id, err)
				}
			}
		}
	}

	// 【关键】同步持仓状态：手动平仓检测、未实现盈亏刷新、交易所持仓但本地无订单则补单
	s.syncPositionsWithCache(ctx, robot, engine.Exchange, cachedPositions, historyOrders)

	// 【兜底】从历史订单补全本地订单缺失字段（市场状态/风险偏好/策略参数等）
	s.syncLocalOrders(ctx, robot, engine.Exchange, historyOrders)

	// 【方案A兜底】open orders 对账写入事实表（供前端挂单列表只读DB）
	// 说明：优先依赖私有WS增量；这里仅用于 WS 漏包/断连后的最终一致性兜底。
	s.syncOpenOrdersToDBThrottled(ctx, robot, engine.Exchange)
}

func (s *OrderStatusSyncService) syncOpenOrdersToDBThrottled(ctx context.Context, robot *entity.TradingRobot, ex exchange.Exchange) {
	if robot == nil || ex == nil {
		return
	}
	// 只对运行中机器人做兜底对账
	if robot.Status != 2 {
		return
	}
	// 节流：同一机器人 10s 内最多打一次 openOrders
	s.openOrdersMu.Lock()
	last := s.lastOpenOrdersSyncAt[robot.Id]
	if !last.IsZero() && time.Since(last) < 10*time.Second {
		s.openOrdersMu.Unlock()
		return
	}
	s.lastOpenOrdersSyncAt[robot.Id] = time.Now()
	s.openOrdersMu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(ctx, "[OrderStatusSync] syncOpenOrdersToDB panic recovered: robotId=%d err=%v", robot.Id, r)
			}
		}()
		orders, err := ex.GetOpenOrders(ctx, robot.Symbol)
		if err != nil {
			g.Log().Debugf(ctx, "[OrderStatusSync] 获取 openOrders 失败(兜底对账): robotId=%d err=%v", robot.Id, err)
			return
		}
		_ = SyncExchangeOpenOrdersToDB(ctx, robot.Id, robot.Exchange, robot.ApiConfigId, robot.Symbol, orders)
	}()
}

// syncPositionsWithCache 使用缓存数据同步持仓状态
// 【优化】优先使用 RobotEngine 缓存的持仓数据，避免重复调用 GetPositions API
func (s *OrderStatusSyncService) syncPositionsWithCache(ctx context.Context, robot *entity.TradingRobot, ex exchange.Exchange, cachedPositions []*exchange.Position, historyOrders []*exchange.Order) {
	// 【健壮性优化】添加 panic 恢复机制
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] syncPositionsWithCache panic recovered: robotId=%d, err=%v", robot.Id, r)
		}
	}()

	var positions []*exchange.Position

	// 优先使用缓存的持仓数据
	if len(cachedPositions) > 0 {
		positions = cachedPositions
		g.Log().Debugf(ctx, "[OrderStatusSync] robotId=%d 使用缓存持仓数据，避免API调用", robot.Id)
	} else {
		// 缓存为空时才调用 API
		var err error
		positions, err = ex.GetPositions(ctx, robot.Symbol)
		if err != nil {
			g.Log().Debugf(ctx, "[OrderStatusSync] 获取持仓失败: robotId=%d, err=%v", robot.Id, err)
			return
		}
	}

	// 构建持仓映射 (交易所实际持仓)
	exchangePositions := make(map[string]*exchange.Position)
	for _, pos := range positions {
		if math.Abs(pos.PositionAmt) > 0.0001 {
			exchangePositions[pos.PositionSide] = pos
		}
	}

	// 查询本地"持仓中"的订单
	var localOrders []*entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("status", OrderStatusOpen). // 持仓中（使用统一的订单状态常量）
		Scan(&localOrders)
	if err != nil {
		return
	}

	// 检查本地订单是否在交易所仍有持仓
	for _, order := range localOrders {
		positionSide := ""
		if order.Direction == "long" {
			positionSide = "LONG"
		} else {
			positionSide = "SHORT"
		}

		exchangePos, hasPosition := exchangePositions[positionSide]

		if !hasPosition {
			// 交易所已无该方向持仓，但本地显示持仓中 → 订单已被手动平仓
			g.Log().Infof(ctx, "[OrderStatusSync] 检测到手动平仓: robotId=%d, orderId=%d, positionSide=%s",
				robot.Id, order.Id, positionSide)

			// 【优化】清除内存中的持仓状态（事件驱动优化）
			if robotEngine := GetRobotTaskManager().GetEngine(robot.Id); robotEngine != nil {
				robotEngine.ClearPosition(ctx, positionSide)
				g.Log().Debugf(ctx, "[OrderStatusSync] robotId=%d 已清除内存持仓: positionSide=%s", robot.Id, positionSide)
			}

			// 使用缓存行情或获取最新行情
			// Prefer exchange trade history for close info (realized PnL / price / time) to keep reconciliation consistent.
			// 【新增】同步时批量落库成交流水（幂等去重），用于后续交易明细/手续费/已实现盈亏展示
			// 说明：这里属于“订单状态同步/外部手动平仓检测”链路，尽量先把成交落库，页面查询只读DB。
			if saved, matched, ferr := fetchAndStoreTradeHistory(ctx, ex, robot.ApiConfigId, robot.Exchange, robot.Symbol, 800); ferr != nil {
				g.Log().Warningf(ctx, "[OrderStatusSync] 外部手动平仓检测落库成交流水失败(继续结算): robotId=%d, symbol=%s, err=%v",
					robot.Id, robot.Symbol, ferr)
			} else {
				g.Log().Debugf(ctx, "[OrderStatusSync] 外部手动平仓检测已落库成交流水: robotId=%d, symbol=%s, saved=%d, matched=%d",
					robot.Id, robot.Symbol, saved, matched)
			}

			closePrice := order.OpenPrice
			realizedProfit := 0.0
			var closeOrder *exchange.Order
			if cp, rp, fee, feeCoin, oid, ts, ok := tryCloseInfoFromTradeHistory(ctx, ex, robot.Symbol, positionSide, order.OpenTime); ok {
				closePrice = cp
				realizedProfit = rp
				closeOrder = &exchange.Order{OrderId: oid, Fee: fee, FeeCoin: feeCoin, CreateTime: ts, UpdateTime: ts}
			} else {
				// Fallback: estimate by latest price if trade history is unavailable.
				ticker, err := ex.GetTicker(ctx, robot.Symbol)
				closePrice = order.OpenPrice // default
				if err == nil && ticker != nil {
					closePrice = ticker.LastPrice
				}

				if order.Direction == "long" {
					realizedProfit = (closePrice - order.OpenPrice) * order.Quantity
				} else {
					realizedProfit = (order.OpenPrice - closePrice) * order.Quantity
				}
			}

			// Update local order as CLOSED
			s.CloseOrder(ctx, order, closePrice, realizedProfit, "手动平仓(同步检测)", closeOrder, nil)
		} else {
			// 交易所仍有持仓，更新未实现盈亏
			s.updateOrderUnrealizedPnl(ctx, order, exchangePos)
		}
	}

	// 【重要】检查交易所持仓，如果本地没有对应订单记录，创建订单记录
	for positionSide, exchangePos := range exchangePositions {
		// 检查本地是否已有该方向的持仓订单
		hasLocalOrder := false
		for _, order := range localOrders {
			localPositionSide := ""
			if order.Direction == "long" {
				localPositionSide = "LONG"
			} else {
				localPositionSide = "SHORT"
			}
			if localPositionSide == positionSide {
				hasLocalOrder = true
				break
			}
		}

		// 如果交易所有持仓但本地没有订单记录，创建订单记录
		if !hasLocalOrder {
			g.Log().Infof(ctx, "[OrderStatusSync] 检测到交易所持仓但本地无订单记录: robotId=%d, positionSide=%s, 创建订单记录",
				robot.Id, positionSide)

			created := false

			// 【优化】使用传入的历史订单，避免重复请求
			if len(historyOrders) > 0 {
				// 查找最近的开仓订单（匹配持仓方向）
				// 注意：GetOrderHistory 返回的是成交记录，需要根据 Type 字段判断是开仓还是平仓
				var matchedOrder *exchange.Order
				for _, histOrder := range historyOrders {
					// 判断是否是开仓订单：Type 字段包含 "开仓" 或 "open"
					isOpenOrder := strings.Contains(histOrder.Type, "开仓") || strings.Contains(strings.ToLower(histOrder.Type), "open")

					// 推断持仓方向：根据 Side 和 Type 判断
					// side="BUY" + tradeSide="open" -> LONG
					// side="SELL" + tradeSide="open" -> SHORT
					orderPositionSide := ""
					if isOpenOrder {
						if histOrder.Side == "BUY" {
							orderPositionSide = "LONG"
						} else if histOrder.Side == "SELL" {
							orderPositionSide = "SHORT"
						}
					}

					// 如果 PositionSide 字段已设置，优先使用
					if histOrder.PositionSide != "" {
						orderPositionSide = histOrder.PositionSide
					}

					// 匹配持仓方向和开仓订单
					if orderPositionSide == positionSide && isOpenOrder && histOrder.Status == "FILLED" {
						// 检查是否已有该订单ID的记录
						// 注意：GetOrderHistory 返回的是成交记录，OrderId 是成交ID，ClientId 是订单ID
						// 应该使用 ClientId（订单ID）来检查，因为同一个订单可能有多个成交记录
						orderIdToCheck := histOrder.ClientId // 订单ID
						if orderIdToCheck == "" {
							orderIdToCheck = histOrder.OrderId // 如果没有 ClientId，使用 OrderId
						}

						exists, _ := dao.TradingOrder.Ctx(ctx).
							Where("robot_id", robot.Id).
							Where("exchange_order_id", orderIdToCheck).
							Count()
						if exists == 0 {
							matchedOrder = histOrder
							// 设置 PositionSide，确保后续处理正确
							matchedOrder.PositionSide = orderPositionSide
							break
						}
					}
				}

				if matchedOrder != nil {
					// 创建订单记录
					s.createOrderFromExchange(ctx, robot, matchedOrder, exchangePos)
					created = true
				} else {
					// 【优化】未找到匹配的开仓订单时，只记录警告，不影响后续逻辑
					// 这种情况可能是：订单是在系统外创建的，或者历史订单已过期
					g.Log().Debugf(ctx, "[OrderStatusSync] 未找到匹配的开仓订单: robotId=%d, positionSide=%s, historyOrdersCount=%d（不影响后续同步）", robot.Id, positionSide, len(historyOrders))
				}
			} else {
				// 【优化】历史订单为空时，只记录调试日志，不影响后续逻辑
				g.Log().Debugf(ctx, "[OrderStatusSync] 历史订单为空，无法创建订单记录: robotId=%d, positionSide=%s（不影响后续同步）", robot.Id, positionSide)
			}

			// 【关键兜底】历史订单无法构建时，用持仓补建OPEN订单，保证DB不缺持仓中订单
			if !created {
				g.Log().Warningf(ctx, "[OrderStatusSync] 将使用持仓数据补建OPEN订单(兜底): robotId=%d, symbol=%s, positionSide=%s, entryPrice=%.4f, qty=%.6f, pnl=%.4f",
					robot.Id, exchangePos.Symbol, positionSide, exchangePos.EntryPrice, math.Abs(exchangePos.PositionAmt), exchangePos.UnrealizedPnl)
				s.createOrderFromPosition(ctx, robot, positionSide, exchangePos)
			}
		}
	}
}

// createOrderFromPosition 从交易所持仓创建本地订单记录（兜底对账单）
// 目标：保证当交易所确实有持仓时，本地一定存在对应的 status=持仓中 订单，避免自动止盈/止损被“无订单”挡住。
func (s *OrderStatusSyncService) createOrderFromPosition(ctx context.Context, robot *entity.TradingRobot, positionSide string, pos *exchange.Position) {
	// 【健壮性优化】panic 恢复
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] createOrderFromPosition panic recovered: robotId=%d, err=%v", robot.Id, r)
		}
	}()

	if robot == nil || pos == nil {
		return
	}
	if math.Abs(pos.PositionAmt) <= 0.0001 {
		return
	}

	// 确定方向
	positionSide = strings.ToUpper(strings.TrimSpace(positionSide))
	direction := "long"
	if positionSide == "SHORT" {
		direction = "short"
	}

	// 市场状态
	var marketState string
	globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(robot.Exchange, robot.Symbol)
	if globalAnalysis != nil {
		marketState = normalizeMarketState(string(globalAnalysis.MarketState))
	}
	if marketState == "" {
		if robot.AutoMarketState == 1 && robot.MarketState != "" {
			marketState = robot.MarketState
		} else {
			marketState = "trend"
		}
	}
	marketState = normalizeMarketState(marketState)

	// 风险偏好：优先用映射关系，其次用机器人当前配置，最后兜底 balanced
	riskPref := ""
	if engine := GetRobotTaskManager().GetEngine(robot.Id); engine != nil {
		engine.mu.RLock()
		if marketState != "" {
			riskPref = engine.MarketRiskMapping[marketState]
		}
		engine.mu.RUnlock()
	}
	if riskPref == "" {
		riskPref = robot.RiskPreference
	}
	if riskPref == "" {
		riskPref = "balanced"
	}

	// 策略参数（尽力从模板取，取不到用默认）
	stopLossPercent := 10.0
	autoStartRetreatPercent := 5.0
	profitRetreatPercent := 30.0
	var template *entity.TradingStrategyTemplate
	if robot.StrategyGroupId > 0 && marketState != "" && riskPref != "" {
		err := dao.TradingStrategyTemplate.Ctx(ctx).
			Where("group_id", robot.StrategyGroupId).
			Where("market_state", marketState).
			Where("risk_preference", riskPref).
			Scan(&template)
		if err == nil && template != nil {
			stopLossPercent = template.StopLossPercent
			autoStartRetreatPercent = template.AutoStartRetreatPercent
			profitRetreatPercent = template.ProfitRetreatPercent
		}
	}

	// 价格/数量/保证金/杠杆
	openPrice := pos.EntryPrice
	if openPrice <= 0 {
		openPrice = pos.MarkPrice
	}
	quantity := math.Abs(pos.PositionAmt)

	leverage := pos.Leverage
	if leverage <= 0 {
		leverage = robot.Leverage
	}
	margin := pos.Margin
	if margin <= 0 {
		margin = pos.IsolatedMargin
	}
	if margin <= 0 && leverage > 0 && openPrice > 0 {
		margin = quantity * openPrice / float64(leverage)
	}

	// 合成一个不会冲突的 exchange_order_id（对账占位）
	exchangeOrderId := fmt.Sprintf("POS-%d-%s-%s-%d", robot.Id, strings.ToUpper(strings.ReplaceAll(pos.Symbol, "/", "")), direction, time.Now().UnixNano())
	orderSn := fmt.Sprintf("TO%s%s", gtime.Now().Format("YmdHis"), grand.S(6))
	now := gtime.Now()

	orderData := g.Map{
		"user_id":                    robot.UserId,
		"robot_id":                   robot.Id,
		"exchange":                   robot.Exchange,
		"symbol":                     pos.Symbol,
		"order_sn":                   orderSn,
		"exchange_order_id":          exchangeOrderId,
		"direction":                  direction,
		"open_price":                 openPrice,
		"quantity":                   quantity,
		"leverage":                   leverage,
		"margin":                     margin,
		"unrealized_profit":          pos.UnrealizedPnl,
		"status":                     OrderStatusOpen,
		"stop_loss_percent":          stopLossPercent,
		"auto_start_retreat_percent": autoStartRetreatPercent,
		"profit_retreat_percent":     profitRetreatPercent,
		"market_state":               marketState,
		"risk_level":                 riskPref,
		"open_time":                  now,
		"created_at":                 now,
		"updated_at":                 now,
		"profit_retreat_started":     0,
		"highest_profit":             0,
	}

	// 【PostgreSQL 兼容】使用 InsertAndGetId() 而不是 Insert() + LastInsertId()
	newId, err := dao.TradingOrder.Ctx(ctx).Data(orderData).InsertAndGetId()
	if err != nil {
		g.Log().Errorf(ctx, "[OrderStatusSync] 补建OPEN订单失败: robotId=%d, positionSide=%s, err=%v", robot.Id, positionSide, err)
		return
	}
	g.Log().Warningf(ctx, "[OrderStatusSync] 已补建OPEN订单(兜底): robotId=%d, orderId=%d, positionSide=%s, exchangeOrderId=%s",
		robot.Id, newId, positionSide, exchangeOrderId)
}

// syncLocalOrders 同步本地订单状态（从交易所历史订单更新）
func (s *OrderStatusSyncService) syncLocalOrders(ctx context.Context, robot *entity.TradingRobot, ex exchange.Exchange, historyOrders []*exchange.Order) {
	// 【健壮性优化】添加 panic 恢复机制，确保同步订单失败不影响后续逻辑
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] syncLocalOrders panic recovered: robotId=%d, err=%v", robot.Id, r)
		}
	}()

	// 【优化】使用传入的历史订单，避免重复请求
	if len(historyOrders) == 0 {
		return
	}

	// 查询本地该机器人的所有订单（用于匹配更新）
	var localOrders []*entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("status IN (?)", []int{OrderStatusOpen, OrderStatusPending}). // 持仓中 或 未成交（使用统一的订单状态常量）
		Scan(&localOrders)
	if err != nil {
		return
	}

	// 构建本地订单映射 (exchange_order_id -> order)
	localOrderMap := make(map[string]*entity.TradingOrder)
	for _, order := range localOrders {
		if order.ExchangeOrderId != "" {
			localOrderMap[order.ExchangeOrderId] = order
		}
	}

	// 更新匹配的订单，并补全缺失的字段
	for _, historyOrder := range historyOrders {
		// 【优化】尝试通过订单ID匹配（优先使用 OrderId，其次使用 ClientId）
		orderIdToMatch := historyOrder.OrderId
		if orderIdToMatch == "" {
			orderIdToMatch = historyOrder.ClientId
		}

		if localOrder, ok := localOrderMap[orderIdToMatch]; ok {
			// 检查并补全缺失的字段
			updateData := g.Map{
				"updated_at": gtime.Now(),
			}

			// 【优化】补全成交信息（只更新缺失字段，避免重复更新）
			if historyOrder.AvgPrice > 0 {
				// 只更新缺失的成交均价
				if localOrder.OpenPrice == 0 || localOrder.OpenPrice <= 0 {
					updateData["avg_price"] = historyOrder.AvgPrice
					updateData["open_price"] = historyOrder.AvgPrice
				} else if localOrder.OpenPrice > 0 && math.Abs(localOrder.OpenPrice-historyOrder.AvgPrice) > 0.01 {
					// 如果价格差异较大（超过0.01），可能是数据不一致，更新为交易所数据
					g.Log().Warningf(ctx, "[OrderStatusSync] 订单成交价格不一致: robotId=%d, orderId=%d, localPrice=%.4f, exchangePrice=%.4f, 更新为交易所价格",
						robot.Id, localOrder.Id, localOrder.OpenPrice, historyOrder.AvgPrice)
					updateData["avg_price"] = historyOrder.AvgPrice
					updateData["open_price"] = historyOrder.AvgPrice
				}
			}
			// 【优化】只更新缺失的已成交数量
			if historyOrder.FilledQty > 0 && (localOrder.Quantity == 0 || localOrder.Quantity <= 0) {
				updateData["filled_qty"] = historyOrder.FilledQty
			} else if historyOrder.FilledQty > 0 && math.Abs(localOrder.Quantity-historyOrder.FilledQty) > 0.0001 {
				// 如果数量差异较大，更新为交易所数据
				g.Log().Warningf(ctx, "[OrderStatusSync] 订单成交数量不一致: robotId=%d, orderId=%d, localQty=%.6f, exchangeQty=%.6f, 更新为交易所数量",
					robot.Id, localOrder.Id, localOrder.Quantity, historyOrder.FilledQty)
				updateData["filled_qty"] = historyOrder.FilledQty
			}

			// 补全订单ID（如果缺失）
			if localOrder.ExchangeOrderId == "" && historyOrder.OrderId != "" {
				updateData["exchange_order_id"] = historyOrder.OrderId
			}

			// 补全创建时间（如果缺失）
			if (localOrder.OpenTime == nil || localOrder.OpenTime.IsZero()) && historyOrder.CreateTime > 0 {
				orderCreateTime := gtime.NewFromTimeStamp(historyOrder.CreateTime / 1000)
				updateData["open_time"] = orderCreateTime
				if localOrder.CreatedAt == nil || localOrder.CreatedAt.IsZero() {
					updateData["created_at"] = orderCreateTime
				}
			}

			// 补全市场状态和风险偏好（如果缺失）
			// 注意：这些字段可能不在实体中，但数据库中存在，直接使用数据库字段名更新
			// 从全局市场分析器获取市场状态
			globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(robot.Exchange, robot.Symbol)
			if globalAnalysis != nil {
				marketState := normalizeMarketState(string(globalAnalysis.MarketState))
				if marketState != "" {
					// 查询订单当前的市场状态（从数据库）
					var currentMarketState string
					_ = dao.TradingOrder.Ctx(ctx).Where("id", localOrder.Id).Fields("market_state").Scan(&currentMarketState)
					if currentMarketState == "" {
						updateData["market_state"] = marketState
					}

					// 从映射关系获取风险偏好
					robotEngine := GetRobotTaskManager().GetEngine(robot.Id)
					if robotEngine != nil {
						robotEngine.mu.RLock()
						riskPref := robotEngine.MarketRiskMapping[marketState]
						robotEngine.mu.RUnlock()

						// 【严格模式】映射关系为空时直接报错
						if riskPref == "" {
							errMsg := fmt.Sprintf("机器人ID=%d 市场状态=%s 在映射关系中未找到对应的风险偏好，无法补全订单信息。请检查机器人的风险配置映射关系是否完整", robot.Id, marketState)
							g.Log().Errorf(ctx, "[OrderStatusSync] %s", errMsg)
							// 记录错误日志
							dao.TradingExecutionLog.Ctx(ctx).Insert(g.Map{
								"robot_id":   robot.Id,
								"event_type": "error",
								"event_msg":  errMsg,
								"event_data": fmt.Sprintf(`{"error":"映射关系不完整","orderId":%d,"marketState":"%s"}`, localOrder.Id, marketState),
								"created_at": gtime.Now(),
							})
							continue // 跳过该订单，不补全
						}

						// 查询订单当前的风险偏好（从数据库）
						var currentRiskLevel string
						_ = dao.TradingOrder.Ctx(ctx).Where("id", localOrder.Id).Fields("risk_level").Scan(&currentRiskLevel)
						if currentRiskLevel == "" {
							updateData["risk_level"] = riskPref
						}

						// 如果策略参数缺失，尝试加载并补全
						var currentStopLoss, currentAutoStart, currentProfitRetreat float64
						_ = dao.TradingOrder.Ctx(ctx).Where("id", localOrder.Id).
							Fields("stop_loss_percent", "auto_start_retreat_percent", "profit_retreat_percent").
							Scan(&struct {
								StopLossPercent         *float64 `json:"stopLossPercent"`
								AutoStartRetreatPercent *float64 `json:"autoStartRetreatPercent"`
								ProfitRetreatPercent    *float64 `json:"profitRetreatPercent"`
							}{&currentStopLoss, &currentAutoStart, &currentProfitRetreat})

						if currentStopLoss == 0 || currentAutoStart == 0 || currentProfitRetreat == 0 {
							// 加载策略参数
							strategyParams, err := s.loadStrategyParamsForOrder(ctx, robot, marketState, riskPref)
							if err == nil && strategyParams != nil {
								if currentStopLoss == 0 && strategyParams.StopLossPercent > 0 {
									updateData["stop_loss_percent"] = strategyParams.StopLossPercent
								}
								if currentAutoStart == 0 && strategyParams.AutoStartRetreatPercent > 0 {
									updateData["auto_start_retreat_percent"] = strategyParams.AutoStartRetreatPercent
								}
								if currentProfitRetreat == 0 && strategyParams.ProfitRetreatPercent > 0 {
									updateData["profit_retreat_percent"] = strategyParams.ProfitRetreatPercent
								}
								// 杠杆和保证金参数（如果数据库字段存在）
								if strategyParams.LeverageMin > 0 {
									updateData["leverage_min"] = strategyParams.LeverageMin
								}
								if strategyParams.LeverageMax > 0 {
									updateData["leverage_max"] = strategyParams.LeverageMax
								}
								if strategyParams.MarginPercentMin > 0 {
									updateData["margin_percent_min"] = strategyParams.MarginPercentMin
								}
								if strategyParams.MarginPercentMax > 0 {
									updateData["margin_percent_max"] = strategyParams.MarginPercentMax
								}
							}
						}
					} else {
						// RobotEngine 不存在，无法获取映射关系
						errMsg := fmt.Sprintf("机器人ID=%d 引擎不存在，无法补全订单的风险偏好。订单ID=%d", robot.Id, localOrder.Id)
						g.Log().Warningf(ctx, "[OrderStatusSync] %s", errMsg)
					}
				} else {
					// 无法获取市场状态
					g.Log().Warningf(ctx, "[OrderStatusSync] 机器人ID=%d 无法从全局市场分析器获取市场状态，无法补全订单信息。订单ID=%d", robot.Id, localOrder.Id)
				}
			}

			// 补全策略组ID（如果缺失，如果数据库字段存在）
			if robot.StrategyGroupId > 0 {
				var currentStrategyGroupId int64
				_ = dao.TradingOrder.Ctx(ctx).Where("id", localOrder.Id).Fields("strategy_group_id").Scan(&currentStrategyGroupId)
				if currentStrategyGroupId == 0 {
					updateData["strategy_group_id"] = robot.StrategyGroupId
				}
			}

			// 执行更新
			if len(updateData) > 1 { // 除了 updated_at 还有其他字段
				_, err := dao.TradingOrder.Ctx(ctx).
					Where("id", localOrder.Id).
					Data(updateData).
					Update()
				if err != nil {
					g.Log().Warningf(ctx, "[OrderStatusSync] 补全订单信息失败: robotId=%d, orderId=%d, err=%v", robot.Id, localOrder.Id, err)
				} else {
					g.Log().Infof(ctx, "[OrderStatusSync] 已补全订单信息: robotId=%d, orderId=%d, 补全字段数=%d", robot.Id, localOrder.Id, len(updateData)-1)
				}
			}
		}
	}
}

// loadStrategyParamsForOrder 为订单补全加载策略参数（不依赖 RobotEngine）
//
//lint:ignore U1000 该方法作为订单补全的兜底能力保留（在不同部署/编译组合下可能由同步链路或外部调用触发）
func (s *OrderStatusSyncService) loadStrategyParamsForOrder(ctx context.Context, robot *entity.TradingRobot, marketState, riskPreference string) (*StrategyParams, error) { //nolint:unused
	params := &StrategyParams{}

	// 规范化市场状态（统一格式）
	normalizedMarketState := normalizeMarketState(marketState)

	// 1. 获取策略组ID（优先级：机器人.StrategyGroupId > CurrentStrategy.group_id）
	var groupId int64 = 0

	// 1.1 优先使用机器人绑定的策略组ID
	if robot.StrategyGroupId > 0 {
		groupId = robot.StrategyGroupId
		g.Log().Debugf(ctx, "[OrderStatusSync] robotId=%d 使用机器人绑定的策略组: groupId=%d", robot.Id, groupId)
	}

	// 1.2 其次从 CurrentStrategy JSON 中获取（兼容旧数据）
	if groupId == 0 && robot.CurrentStrategy != "" {
		var strategyData map[string]interface{}
		if err := json.Unmarshal([]byte(robot.CurrentStrategy), &strategyData); err == nil {
			// 支持 groupId 和 group_id 两种格式（兼容旧数据）
			if gid, ok := strategyData["groupId"].(float64); ok {
				groupId = int64(gid)
				g.Log().Debugf(ctx, "[OrderStatusSync] robotId=%d 从CurrentStrategy获取策略组: groupId=%d", robot.Id, groupId)
			} else if gid, ok := strategyData["group_id"].(float64); ok {
				groupId = int64(gid)
				g.Log().Debugf(ctx, "[OrderStatusSync] robotId=%d 从CurrentStrategy获取策略组: group_id=%d", robot.Id, groupId)
			}
		}
	}

	// 2. 检查是否有策略组ID
	if groupId == 0 {
		errMsg := fmt.Sprintf("机器人ID=%d 未绑定策略组ID，无法加载策略参数", robot.Id)
		g.Log().Errorf(ctx, "[OrderStatusSync] %s", errMsg)
		return nil, gerror.New(errMsg)
	}

	// 3. 从策略模板表中查询对应的策略（尝试多种市场状态名称，兼容旧数据）
	marketStatesToTry := []string{
		normalizedMarketState, // 标准格式（优先级最高）
	}

	// 如果原始格式与规范化格式不同，添加原始格式
	if normalizedMarketState != marketState {
		marketStatesToTry = append(marketStatesToTry, marketState)
	}

	// 添加兼容格式（仅当与标准格式不同时）
	if normalizedMarketState == "volatile" && marketState != "volatile" {
		marketStatesToTry = append(marketStatesToTry, "range") // 兼容旧格式
	}
	if normalizedMarketState == "high_vol" && marketState != "high_vol" {
		marketStatesToTry = append(marketStatesToTry, "high-volatility") // 兼容数据库格式
	}
	if normalizedMarketState == "low_vol" && marketState != "low_vol" {
		marketStatesToTry = append(marketStatesToTry, "low-volatility") // 兼容数据库格式
	}

	for _, ms := range marketStatesToTry {
		var strategy *entity.TradingStrategyTemplate
		err := dao.TradingStrategyTemplate.Ctx(ctx).
			Where("group_id", groupId).
			Where(dao.TradingStrategyTemplate.Columns().MarketState, ms).
			Where(dao.TradingStrategyTemplate.Columns().RiskPreference, riskPreference).
			Scan(&strategy)

		if err == nil && strategy != nil {
			params.Window = strategy.MonitorWindow
			params.Threshold = strategy.VolatilityThreshold
			params.LeverageMin = strategy.Leverage
			params.LeverageMax = strategy.Leverage
			params.MarginPercentMin = strategy.MarginPercent
			params.MarginPercentMax = strategy.MarginPercent
			params.StopLossPercent = strategy.StopLossPercent
			params.ProfitRetreatPercent = strategy.ProfitRetreatPercent
			params.AutoStartRetreatPercent = strategy.AutoStartRetreatPercent

			g.Log().Infof(ctx, "[OrderStatusSync] robotId=%d 从策略模板加载参数: market=%s(规范化=%s,查询=%s), risk=%s, 窗口=%d, 波动=%.1f, 杠杆=%d, 保证金=%.1f%%, 止损=%.1f%%, 启动止盈=%.1f%%, 止盈回撤=%.1f%%",
				robot.Id, marketState, normalizedMarketState, ms, riskPreference,
				params.Window, params.Threshold,
				params.LeverageMin, params.MarginPercentMin,
				params.StopLossPercent, params.AutoStartRetreatPercent, params.ProfitRetreatPercent)
			return params, nil
		}
	}

	// 4. 找不到策略模板，返回详细错误信息
	errMsg := fmt.Sprintf("机器人ID=%d 找不到策略模板: groupId=%d, marketState=%s/%s, riskPreference=%s",
		robot.Id, groupId, marketState, normalizedMarketState, riskPreference)
	g.Log().Errorf(ctx, "[OrderStatusSync] %s", errMsg)
	return nil, gerror.New(errMsg)
}

// updateOrderUnrealizedPnl 更新订单未实现盈亏
//
//lint:ignore U1000 该方法作为持仓同步/事件驱动更新的通用能力保留（在不同部署/编译组合下可能由同步链路或外部调用触发）
func (s *OrderStatusSyncService) updateOrderUnrealizedPnl(ctx context.Context, order *entity.TradingOrder, pos *exchange.Position) { //nolint:unused
	// 计算未实现盈亏
	unrealizedPnl := pos.UnrealizedPnl

	// 【重要】检查订单是否是刚创建的或刚变为OPEN的（5分钟内）
	// 如果是新订单，清除最高盈利，让每个订单独立计算自己的最高盈利
	isNewOrder := false
	if order.OpenTime != nil && !order.OpenTime.IsZero() {
		// 检查订单开仓时间是否在5分钟内
		timeSinceOpen := time.Since(order.OpenTime.Time)
		if timeSinceOpen < 5*time.Minute {
			isNewOrder = true
		}
	} else if order.UpdatedAt != nil && !order.UpdatedAt.IsZero() {
		// 如果没有开仓时间，检查更新时间（订单状态变为OPEN的时间）
		timeSinceUpdate := time.Since(order.UpdatedAt.Time)
		if timeSinceUpdate < 5*time.Minute {
			isNewOrder = true
		}
	}

	// 更新最高盈利（只增不减）
	highestProfit := order.HighestProfit

	// 【重要】如果是新订单，清除最高盈利，从0开始计算
	if isNewOrder && highestProfit > 0 {
		g.Log().Infof(ctx, "[OrderStatusSync] 检测到新订单，清除最高盈利: orderId=%d, oldHighestProfit=%.4f", order.Id, highestProfit)
		highestProfit = 0
	}

	highestProfitChanged := false
	if unrealizedPnl > highestProfit {
		highestProfit = unrealizedPnl
		highestProfitChanged = true
	}

	// 检查是否需要更新：未实现盈亏变化较大，或者最高盈利发生变化
	pnlChanged := math.Abs(order.UnrealizedProfit-unrealizedPnl) >= 0.01
	if !pnlChanged && !highestProfitChanged {
		return
	}

	_, err := dao.TradingOrder.Ctx(ctx).
		Where("id", order.Id).
		Data(g.Map{
			"unrealized_profit": unrealizedPnl,
			"highest_profit":    highestProfit,
			"mark_price":        pos.MarkPrice,
			"updated_at":        gtime.Now(),
		}).
		Update()

	if err != nil {
		g.Log().Warningf(ctx, "[OrderStatusSync] 更新未实现盈亏失败: orderId=%d, err=%v", order.Id, err)
	} else {
		// 【订单事件】记录持仓更新事件
		updateData := map[string]interface{}{
			"unrealized_profit": unrealizedPnl,
			"highest_profit":    highestProfit,
			"mark_price":        pos.MarkPrice,
			"position_amt":      pos.PositionAmt,
			"entry_price":       pos.EntryPrice,
		}
		RecordPositionUpdated(ctx, order.Id, order.ExchangeOrderId, updateData)
	}
}

// CloseOrder 关闭订单（更新为已平仓状态，保存完整的平仓信息）
// 【公开方法】用于手动平仓和自动平仓统一补全订单信息
// 【优化】只补全缺失字段，不覆盖已有数据；检查算力是否已扣除，避免重复扣除
func (s *OrderStatusSyncService) CloseOrder(ctx context.Context, order *entity.TradingOrder, closePrice, realizedProfit float64, reason string, closeOrder *exchange.Order, pos *exchange.Position) {
	// 【优化】先查询当前订单状态，只补全缺失字段
	var currentOrder *entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).Where(dao.TradingOrder.Columns().Id, order.Id).Scan(&currentOrder)
	if err != nil || currentOrder == nil {
		g.Log().Errorf(ctx, "[OrderStatusSync] 查询订单失败: orderId=%d, err=%v", order.Id, err)
		return
	}

	// 平仓时间尽量使用交易所回传时间（避免用“当前时间”造成历史数据不准确）
	closeTime := gtime.Now()
	if closeOrder != nil {
		ts := closeOrder.UpdateTime
		if ts <= 0 {
			ts = closeOrder.CreateTime
		}
		// 交易所时间戳通常为毫秒
		if ts > 0 {
			closeTime = gtime.NewFromTimeStamp(ts / 1000)
		}
	}
	holdDuration := 0
	if currentOrder.OpenTime != nil && !currentOrder.OpenTime.IsZero() {
		holdDuration = int(closeTime.Sub(currentOrder.OpenTime).Seconds())
	}

	// 【优化】只补全缺失字段，不覆盖已有数据
	closeData := g.Map{
		"updated_at": closeTime,
	}

	// 【必须】更新订单状态为已平仓（如果还不是已平仓状态）
	if currentOrder.Status != OrderStatusClosed {
		closeData["status"] = OrderStatusClosed
	}

	// 【补全】平仓价格（如果缺失）
	if currentOrder.ClosePrice == 0 && closePrice > 0 {
		closeData["close_price"] = closePrice
	}

	// 【补全】平仓时间（如果缺失/占位）
	// 注意：历史上 close_time 可能被写成 "2006-01-02 15:04:05" 这类占位值（非空非zero），也要视为无效并覆盖。
	if currentOrder.CloseTime == nil || currentOrder.CloseTime.IsZero() || currentOrder.CloseTime.Year() == 2006 {
		closeData["close_time"] = closeTime
	}

	// 【补全】已实现盈亏（如果缺失或为0，且新计算的盈亏不为0）
	if (currentOrder.RealizedProfit == 0 || math.Abs(currentOrder.RealizedProfit) < 0.01) && math.Abs(realizedProfit) >= 0.01 {
		closeData["realized_profit"] = realizedProfit
	}

	// 【补全】持仓时长（如果缺失）
	if currentOrder.HoldDuration == 0 && holdDuration > 0 {
		closeData["hold_duration"] = holdDuration
	}

	// 【补全】平仓原因（如果缺失）
	if currentOrder.CloseReason == "" && reason != "" {
		closeData["close_reason"] = reason
	}

	// 【补全】平仓订单ID（如果提供且缺失）
	// 【优化】通过查询当前订单的 close_order_id 字段来判断是否缺失
	if closeOrder != nil {
		var currentCloseOrderId string
		closeOrderData, _ := dao.TradingOrder.Ctx(ctx).
			Where("id", order.Id).
			Fields("close_order_id", "close_client_order_id", "close_fee", "close_fee_coin").
			One()
		if closeOrderData != nil && !closeOrderData.IsEmpty() {
			if val := closeOrderData["close_order_id"]; val != nil {
				currentCloseOrderId = val.String()
			}
		}

		// 只补全缺失的字段
		if closeOrder.OrderId != "" && currentCloseOrderId == "" {
			closeData["close_order_id"] = closeOrder.OrderId
		}
		if closeOrder.ClientId != "" {
			var currentCloseClientOrderId string
			if closeOrderData != nil && !closeOrderData.IsEmpty() {
				if val := closeOrderData["close_client_order_id"]; val != nil {
					currentCloseClientOrderId = val.String()
				}
			}
			if currentCloseClientOrderId == "" {
				closeData["close_client_order_id"] = closeOrder.ClientId
			}
		}
		// 【补全】平仓手续费（如果缺失）
		if closeOrder.Fee > 0 {
			var currentCloseFee float64
			if closeOrderData != nil && !closeOrderData.IsEmpty() {
				if val := closeOrderData["close_fee"]; val != nil {
					currentCloseFee = val.Float64()
				}
			}
			if currentCloseFee == 0 {
				closeData["close_fee"] = closeOrder.Fee
			}
		}
		if closeOrder.FeeCoin != "" {
			var currentCloseFeeCoin string
			if closeOrderData != nil && !closeOrderData.IsEmpty() {
				if val := closeOrderData["close_fee_coin"]; val != nil {
					currentCloseFeeCoin = val.String()
				}
			}
			if currentCloseFeeCoin == "" {
				closeData["close_fee_coin"] = closeOrder.FeeCoin
			}
		}
	}

	// 【补全】平仓时的市场状态（如果缺失）
	if pos != nil {
		// 查询当前订单的平仓相关字段
		var currentCloseMarketState, currentCloseMarkPrice string
		var currentCloseLeverage int
		var currentCloseQuantity float64
		closeInfoData, _ := dao.TradingOrder.Ctx(ctx).
			Where("id", order.Id).
			Fields("close_market_state", "close_mark_price", "close_leverage", "close_quantity").
			One()
		if closeInfoData != nil && !closeInfoData.IsEmpty() {
			if val := closeInfoData["close_market_state"]; val != nil {
				currentCloseMarketState = val.String()
			}
			if val := closeInfoData["close_mark_price"]; val != nil {
				currentCloseMarkPrice = val.String()
			}
			if val := closeInfoData["close_leverage"]; val != nil {
				currentCloseLeverage = val.Int()
			}
			if val := closeInfoData["close_quantity"]; val != nil {
				currentCloseQuantity = val.Float64()
			}
		}

		var robot *entity.TradingRobot
		_ = dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, order.RobotId).Scan(&robot)
		if robot != nil {
			globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(robot.Exchange, robot.Symbol)
			if globalAnalysis != nil {
				closeMarketState := normalizeMarketState(string(globalAnalysis.MarketState))
				if closeMarketState != "" && currentCloseMarketState == "" {
					closeData["close_market_state"] = closeMarketState
				}
			}
		}

		// 【补全】平仓时的标记价格（如果缺失）
		if pos.MarkPrice > 0 && currentCloseMarkPrice == "" {
			closeData["close_mark_price"] = pos.MarkPrice
		}

		// 【补全】平仓时的杠杆（如果缺失）
		if pos.Leverage > 0 && currentCloseLeverage == 0 {
			closeData["close_leverage"] = pos.Leverage
		}

		// 【补全】平仓数量（如果缺失）
		if math.Abs(pos.PositionAmt) > 0.0001 && currentCloseQuantity == 0 {
			closeData["close_quantity"] = math.Abs(pos.PositionAmt)
		}
	}

	// 【补全】平仓时的盈亏详情（如果缺失）
	var currentCloseUnrealizedProfit, currentCloseHighestProfit float64
	closePnlData, _ := dao.TradingOrder.Ctx(ctx).
		Where("id", order.Id).
		Fields("close_unrealized_profit", "close_highest_profit").
		One()
	if closePnlData != nil && !closePnlData.IsEmpty() {
		if val := closePnlData["close_unrealized_profit"]; val != nil {
			currentCloseUnrealizedProfit = val.Float64()
		}
		if val := closePnlData["close_highest_profit"]; val != nil {
			currentCloseHighestProfit = val.Float64()
		}
	}

	if currentOrder.UnrealizedProfit != 0 && currentCloseUnrealizedProfit == 0 {
		closeData["close_unrealized_profit"] = currentOrder.UnrealizedProfit
	}
	if currentOrder.HighestProfit > 0 && currentCloseHighestProfit == 0 {
		closeData["close_highest_profit"] = currentOrder.HighestProfit
	}

	// 执行更新（只更新缺失字段）
	if len(closeData) > 1 { // 除了 updated_at 还有其他字段需要更新
		_, err := dao.TradingOrder.Ctx(ctx).
			Where("id", order.Id).
			Data(closeData).
			Update()
		if err != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] 更新订单状态失败: orderId=%d, err=%v", order.Id, err)
			return
		}
		g.Log().Infof(ctx, "[OrderStatusSync] 已补全订单信息: orderId=%d, 补全字段数=%d", order.Id, len(closeData)-1)
	}

	// 【优化】更新机器人总盈亏（如果已实现盈亏有变化）
	if math.Abs(realizedProfit) >= 0.01 && math.Abs(realizedProfit-currentOrder.RealizedProfit) >= 0.01 {
		profitDiff := realizedProfit - currentOrder.RealizedProfit
		_, _ = dao.TradingRobot.Ctx(ctx).
			Where("id", order.RobotId).
			Increment("total_profit", profitDiff)
	}

	// 已移除：盈利订单消耗算力
	// 说明：按产品需求不再在平仓/同步时扣除算力，也不再写入 power_consumed/power_amount。

	g.Log().Infof(ctx, "[OrderStatusSync] 订单已同步平仓: orderId=%d, closePrice=%.4f, profit=%.4f, reason=%s",
		order.Id, closePrice, realizedProfit, reason)

	// 【事件驱动】实时刷新“运行区间盈亏汇总”
	// 说明：
	// - 不依赖交易所接口：直接从本地订单表(已按交易所口径写入 realized_profit/fee)重算区间汇总
	// - 幂等：即使 CloseOrder 被重复触发（同步兜底/重复事件），也不会重复累计
	var robot *entity.TradingRobot
	_ = dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, currentOrder.RobotId).Scan(&robot)
	if robot != nil {
		refreshCurrentRunSessionSummaryByRobot(ctx, currentOrder.UserId, currentOrder.RobotId, robot.Exchange, robot.Symbol)
	}

	// 【订单事件】记录订单平仓事件
	closeEventData := map[string]interface{}{
		"close_price":     closePrice,
		"realized_profit": realizedProfit,
		"reason":          reason,
		"close_time":      closeTime,
		"hold_duration":   holdDuration,
	}
	if closeOrder != nil {
		closeEventData["exchange_close_order_id"] = closeOrder.OrderId
		closeEventData["close_fee"] = closeOrder.Fee
	}
	RecordOrderClosed(ctx, order.Id, order.ExchangeOrderId, closeEventData)
}

// createOrderFromExchange 从交易所订单创建本地订单记录
func (s *OrderStatusSyncService) createOrderFromExchange(ctx context.Context, robot *entity.TradingRobot, exchangeOrder *exchange.Order, pos *exchange.Position) {
	// 【健壮性优化】添加 panic 恢复机制，确保创建订单失败不影响后续逻辑
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[OrderStatusSync] createOrderFromExchange panic recovered: robotId=%d, err=%v", robot.Id, r)
		}
	}()

	// 确定方向
	direction := "long"
	if exchangeOrder.PositionSide == "SHORT" {
		direction = "short"
	}

	// 【优化】从全局市场分析器获取实时市场状态（而不是从 RobotEngine）
	var marketState, riskPref string
	globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(robot.Exchange, robot.Symbol)
	if globalAnalysis != nil {
		// 从全局市场分析结果获取市场状态
		marketState = normalizeMarketState(string(globalAnalysis.MarketState))
		g.Log().Infof(ctx, "[OrderStatusSync] 从全局市场分析器获取市场状态: robotId=%d, platform=%s, symbol=%s, marketState=%s",
			robot.Id, robot.Exchange, robot.Symbol, marketState)
	}

	// 如果全局市场分析器没有数据，使用机器人配置的默认值
	if marketState == "" {
		if robot.AutoMarketState == 1 && robot.MarketState != "" {
			marketState = robot.MarketState
		} else {
			marketState = "trend" // 默认值
		}
	}
	marketState = normalizeMarketState(marketState)

	// 【重要】从机器人的映射关系获取风险偏好（映射关系是机器人级别的配置）
	robotEngine := GetRobotTaskManager().GetEngine(robot.Id)
	if robotEngine != nil {
		robotEngine.mu.RLock()
		if marketState != "" {
			riskPref = robotEngine.MarketRiskMapping[marketState]
		}
		robotEngine.mu.RUnlock()
	}

	// 如果映射关系中没有风险偏好：兜底用机器人当前配置，避免因为“映射缺失”导致无法对账建单
	if riskPref == "" {
		riskPref = robot.RiskPreference
	}
	if riskPref == "" {
		riskPref = "balanced"
	}

	// 获取策略参数（从策略模板）
	var strategyParams struct {
		StopLossPercent         float64
		AutoStartRetreatPercent float64
		ProfitRetreatPercent    float64
	}

	// 查询策略模板
	var template *entity.TradingStrategyTemplate
	err := dao.TradingStrategyTemplate.Ctx(ctx).
		Where("group_id", robot.StrategyGroupId).
		Where("market_state", marketState).
		Where("risk_preference", riskPref).
		Scan(&template)
	if err == nil && template != nil {
		strategyParams.StopLossPercent = template.StopLossPercent
		strategyParams.AutoStartRetreatPercent = template.AutoStartRetreatPercent
		strategyParams.ProfitRetreatPercent = template.ProfitRetreatPercent
	} else {
		// 使用默认值
		strategyParams.StopLossPercent = 10
		strategyParams.AutoStartRetreatPercent = 5
		strategyParams.ProfitRetreatPercent = 30
		g.Log().Warningf(ctx, "[OrderStatusSync] 未找到策略模板，使用默认值: robotId=%d, market=%s, risk=%s",
			robot.Id, marketState, riskPref)
	}

	// 构建订单数据
	openPrice := exchangeOrder.AvgPrice
	if openPrice <= 0 && pos.EntryPrice > 0 {
		openPrice = pos.EntryPrice
	}
	if openPrice <= 0 {
		openPrice = pos.MarkPrice // 最后兜底使用标记价格
	}

	// 【重要】GetOrderHistory 返回的是成交记录，OrderId 是成交ID，ClientId 是订单ID
	// 应该使用 ClientId（订单ID）作为 exchange_order_id，如果没有则使用 OrderId
	exchangeOrderId := exchangeOrder.ClientId
	if exchangeOrderId == "" {
		exchangeOrderId = exchangeOrder.OrderId
	}

	// 生成订单号：TO (Trading Order) + 时间戳 + 6位随机字符串
	orderSn := fmt.Sprintf("TO%s%s", gtime.Now().Format("YmdHis"), grand.S(6))

	// 【优化】使用交易所返回的创建时间（如果存在），否则使用本地时间
	orderCreateTime := gtime.Now()
	if exchangeOrder.CreateTime > 0 {
		// Order.CreateTime 是毫秒时间戳，转换为 gtime.Time
		orderCreateTime = gtime.NewFromTimeStamp(exchangeOrder.CreateTime / 1000)
	}

	orderData := g.Map{
		"user_id":                    robot.UserId,
		"robot_id":                   robot.Id,
		"exchange":                   robot.Exchange,
		"symbol":                     exchangeOrder.Symbol,
		"order_sn":                   orderSn,         // 订单号
		"exchange_order_id":          exchangeOrderId, // 使用订单ID而不是成交ID
		"direction":                  direction,
		"open_price":                 openPrice,
		"quantity":                   exchangeOrder.Quantity,
		"leverage":                   pos.Leverage,
		"margin":                     pos.IsolatedMargin,
		"unrealized_profit":          pos.UnrealizedPnl,
		"status":                     OrderStatusOpen, // 持仓中（使用统一的订单状态常量）
		"stop_loss_percent":          strategyParams.StopLossPercent,
		"auto_start_retreat_percent": strategyParams.AutoStartRetreatPercent,
		"profit_retreat_percent":     strategyParams.ProfitRetreatPercent,
		"market_state":               marketState,     // 【新增】市场状态
		"risk_level":                 riskPref,        // 【新增】风险偏好
		"open_time":                  orderCreateTime, // 【优化】使用交易所返回的创建时间（如果存在）
		"created_at":                 orderCreateTime, // 【优化】使用交易所返回的创建时间（如果存在）
		"updated_at":                 gtime.Now(),
		// 【关键】明确初始化止盈回撤相关字段，避免依赖数据库默认值
		"profit_retreat_started": 0, // 默认未开启止盈回撤
		"highest_profit":         0, // 初始最高盈利为0
	}

	// 插入订单记录
	// 【PostgreSQL 兼容】使用 InsertAndGetId() 而不是 Insert() + LastInsertId()
	orderId, err := dao.TradingOrder.Ctx(ctx).Data(orderData).InsertAndGetId()
	if err != nil {
		g.Log().Errorf(ctx, "[OrderStatusSync] 创建订单记录失败: robotId=%d, exchangeOrderId=%s, err=%v",
			robot.Id, exchangeOrder.OrderId, err)
		return
	}

	g.Log().Infof(ctx, "[OrderStatusSync] 已创建订单记录: robotId=%d, orderId=%d, exchangeOrderId=%s, direction=%s, price=%.2f",
		robot.Id, orderId, exchangeOrder.OrderId, direction, openPrice)
}

// SyncSingleRobot 已删除 - 使用每1秒自动同步，无需手动触发

// checkAndFixDataConsistency 已删除 - 订单不对接数据库，无需数据一致性检查
// 【已废弃】此函数已删除，订单同步不再更新数据库，只从交易所读取数据

// GetSyncStats 获取同步统计
func (s *OrderStatusSyncService) GetSyncStats(ctx context.Context) map[string]interface{} {
	// 统计各状态订单数量
	var stats struct {
		TotalOrders   int `json:"totalOrders"`
		OpenOrders    int `json:"openOrders"`
		ClosedOrders  int `json:"closedOrders"`
		PendingOrders int `json:"pendingOrders"`
	}

	stats.TotalOrders, _ = dao.TradingOrder.Ctx(ctx).Count()
	stats.OpenOrders, _ = dao.TradingOrder.Ctx(ctx).Where("status", OrderStatusOpen).Count()       // 使用统一的订单状态常量
	stats.ClosedOrders, _ = dao.TradingOrder.Ctx(ctx).Where("status", OrderStatusClosed).Count()   // 使用统一的订单状态常量
	stats.PendingOrders, _ = dao.TradingOrder.Ctx(ctx).Where("status", OrderStatusPending).Count() // 使用统一的订单状态常量

	return g.Map{
		"running":       s.IsRunning(),
		"totalOrders":   stats.TotalOrders,
		"openOrders":    stats.OpenOrders,
		"closedOrders":  stats.ClosedOrders,
		"pendingOrders": stats.PendingOrders,
	}
}
