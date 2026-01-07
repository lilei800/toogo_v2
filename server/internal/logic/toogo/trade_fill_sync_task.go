// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogo

import (
	"context"
	"sync"
	"time"

	"hotgo/internal/dao"
	exlib "hotgo/internal/library/exchange"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// TradeFillSyncTask 成交流水后台同步任务
// 目标：定期从交易所拉取成交数据并落库，避免用户查询交易明细时实时打API
type TradeFillSyncTask struct {
	interval time.Duration // 同步间隔（默认5分钟）
	limit    int           // 每次拉取成交数量（默认500）
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
	mu       sync.Mutex
}

var (
	tradeFillSyncTaskInstance *TradeFillSyncTask
	tradeFillSyncTaskOnce     sync.Once
)

// GetTradeFillSyncTask 获取成交流水同步任务单例
func GetTradeFillSyncTask() *TradeFillSyncTask {
	tradeFillSyncTaskOnce.Do(func() {
		tradeFillSyncTaskInstance = NewTradeFillSyncTask()
	})
	return tradeFillSyncTaskInstance
}

// NewTradeFillSyncTask 创建成交流水同步任务
func NewTradeFillSyncTask() *TradeFillSyncTask {
	return &TradeFillSyncTask{
		interval: 5 * time.Minute, // 默认5分钟同步一次
		limit:    500,             // 默认拉取500条成交记录（足够覆盖大部分场景）
	}
}

// Start 启动同步任务
func (t *TradeFillSyncTask) Start() {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		g.Log().Info(context.Background(), "[TradeFillSyncTask] 任务已在运行中，跳过启动")
		return
	}
	t.running = true
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.mu.Unlock()

	g.Log().Info(t.ctx, "[TradeFillSyncTask] 成交流水同步任务已启动，同步间隔=%v", t.interval)

	// 启动后立即执行一次同步
	go t.syncOnce()

	// 定时同步
	go t.run()
}

// Stop 停止同步任务
func (t *TradeFillSyncTask) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return
	}

	t.running = false
	if t.cancel != nil {
		t.cancel()
	}
	g.Log().Info(context.Background(), "[TradeFillSyncTask] 成交流水同步任务已停止")
}

// run 运行同步任务主循环
func (t *TradeFillSyncTask) run() {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.syncOnce()
		}
	}
}

// syncOnce 执行一次完整同步
func (t *TradeFillSyncTask) syncOnce() {
	ctx := t.ctx
	startTime := time.Now()

	g.Log().Debug(ctx, "[TradeFillSyncTask] 开始同步成交流水...")

	// 1. 查询所有运行中的机器人
	var robots []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().Status, 2). // 2=运行中
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Scan(&robots)

	if err != nil {
		g.Log().Warningf(ctx, "[TradeFillSyncTask] 查询运行中机器人失败: %v", err)
		return
	}

	if len(robots) == 0 {
		g.Log().Debug(ctx, "[TradeFillSyncTask] 没有运行中的机器人，跳过同步")
		return
	}

	g.Log().Infof(ctx, "[TradeFillSyncTask] 找到 %d 个运行中的机器人，开始同步成交数据", len(robots))

	// 2. 并发同步（限制并发数避免过载）
	concurrency := 5 // 最多5个并发
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var successCount, failCount int
	var mu sync.Mutex

	for _, robot := range robots {
		if robot == nil {
			continue
		}

		wg.Add(1)
		go func(r *entity.TradingRobot) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 同步单个机器人的成交数据
			if t.syncRobotTrades(ctx, r) {
				mu.Lock()
				successCount++
				mu.Unlock()
			} else {
				mu.Lock()
				failCount++
				mu.Unlock()
			}
		}(robot)
	}

	wg.Wait()

	duration := time.Since(startTime)
	g.Log().Infof(ctx, "[TradeFillSyncTask] 同步完成: 成功=%d, 失败=%d, 耗时=%v",
		successCount, failCount, duration)
}

// syncRobotTrades 同步单个机器人的成交数据
func (t *TradeFillSyncTask) syncRobotTrades(ctx context.Context, robot *entity.TradingRobot) bool {
	// 1. 获取机器人引擎（优先使用运行中的引擎实例）
	engine := GetRobotTaskManager().GetEngine(robot.Id)
	if engine == nil || engine.Exchange == nil {
		// 引擎未运行，尝试临时创建交易所实例
		return t.syncRobotTradesWithTempExchange(ctx, robot)
	}

	// 2. 类型断言检查是否支持成交历史查询
	type tradeHistoryProvider interface {
		GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*exlib.Trade, error)
	}
	p, ok := engine.Exchange.(tradeHistoryProvider)
	if !ok {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] 交易所不支持成交历史查询: robotId=%d, exchange=%s",
			robot.Id, robot.Exchange)
		return false
	}

	// 3. 拉取最近成交
	trades, err := p.GetTradeHistory(ctx, robot.Symbol, t.limit)
	if err != nil {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] 拉取成交失败: robotId=%d, symbol=%s, err=%v",
			robot.Id, robot.Symbol, err)
		return false
	}

	if len(trades) == 0 {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] 无成交记录: robotId=%d, symbol=%s",
			robot.Id, robot.Symbol)
		return true // 无成交也算成功
	}

	// 4. 落库成交数据（幂等去重）
	// 注意：后台同步只负责“缓存成交到DB”，不负责给成交打 session_id。
	// session_id 归属如果处理不当会污染运行区间统计（把区间外历史成交算进来）。
	inserted, updated, err := upsertTradeFillsFromTrades(
		ctx,
		robot.ApiConfigId,
		engine.Exchange.GetName(),
		robot.Symbol,
		trades,
		nil,
	)

	if err != nil {
		g.Log().Warningf(ctx, "[TradeFillSyncTask] 落库成交数据失败: robotId=%d, symbol=%s, err=%v",
			robot.Id, robot.Symbol, err)
		return false
	}

	if inserted > 0 || updated > 0 {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] 同步成功: robotId=%d, symbol=%s, 新增=%d, 更新=%d",
			robot.Id, robot.Symbol, inserted, updated)
	}

	return true
}

// syncRobotTradesWithTempExchange 使用临时交易所实例同步成交数据
func (t *TradeFillSyncTask) syncRobotTradesWithTempExchange(ctx context.Context, robot *entity.TradingRobot) bool {
	// 查询API配置
	var apiConfig *entity.TradingApiConfig
	err := dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).
		Scan(&apiConfig)

	if err != nil || apiConfig == nil {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] API配置不存在: robotId=%d, apiConfigId=%d",
			robot.Id, robot.ApiConfigId)
		return false
	}

	// 创建交易所实例
	ex, err := GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
	if err != nil {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] 创建交易所实例失败: robotId=%d, exchange=%s, err=%v",
			robot.Id, robot.Exchange, err)
		return false
	}

	// 类型断言检查是否支持成交历史查询
	type tradeHistoryProvider interface {
		GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*exlib.Trade, error)
	}
	p, ok := ex.(tradeHistoryProvider)
	if !ok {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] 交易所不支持成交历史查询: robotId=%d, exchange=%s",
			robot.Id, robot.Exchange)
		return false
	}

	// 拉取成交
	trades, err := p.GetTradeHistory(ctx, robot.Symbol, t.limit)
	if err != nil {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] 拉取成交失败: robotId=%d, symbol=%s, err=%v",
			robot.Id, robot.Symbol, err)
		return false
	}

	if len(trades) == 0 {
		return true // 无成交也算成功
	}

	// 落库
	inserted, updated, err := upsertTradeFillsFromTrades(
		ctx,
		robot.ApiConfigId,
		ex.GetName(),
		robot.Symbol,
		trades,
		nil,
	)

	if err != nil {
		g.Log().Warningf(ctx, "[TradeFillSyncTask] 落库成交数据失败: robotId=%d, symbol=%s, err=%v",
			robot.Id, robot.Symbol, err)
		return false
	}

	if inserted > 0 || updated > 0 {
		g.Log().Debugf(ctx, "[TradeFillSyncTask] 同步成功: robotId=%d, symbol=%s, 新增=%d, 更新=%d",
			robot.Id, robot.Symbol, inserted, updated)
	}

	return true
}

// SetInterval 设置同步间隔（需要重启任务才生效）
func (t *TradeFillSyncTask) SetInterval(interval time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.interval = interval
	g.Log().Infof(context.Background(), "[TradeFillSyncTask] 同步间隔已设置为: %v", interval)
}

// SetLimit 设置每次拉取成交数量（需要重启任务才生效）
func (t *TradeFillSyncTask) SetLimit(limit int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.limit = limit
	g.Log().Infof(context.Background(), "[TradeFillSyncTask] 每次拉取成交数量已设置为: %d", limit)
}

// IsRunning 检查任务是否运行中
func (t *TradeFillSyncTask) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running
}

