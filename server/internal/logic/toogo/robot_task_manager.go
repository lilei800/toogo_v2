// Package toogo 鏈哄櫒浜轰换鍔＄鐞嗗櫒
// 璐熻矗绠＄悊鏈哄櫒浜虹敓鍛藉懆鏈燂紝浣跨敤 RobotEngine 鏋舵瀯
package toogo

import (
	"context"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/library/config"
	"hotgo/internal/library/exchange"
	"hotgo/internal/library/market"
	"hotgo/internal/model/entity"
	"hotgo/utility/encrypt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"golang.org/x/net/proxy"
)

// RobotTaskManager 鏈哄櫒浜轰换鍔＄鐞嗗櫒
type RobotTaskManager struct {
	mu sync.RWMutex

	// 杩愯涓殑鏈哄櫒浜哄紩鎿?key: robotId
	engines map[int64]*RobotEngine

	// 杩愯鐘舵€?
	running bool
	stopCh  chan struct{}
}

var (
	robotTaskManager     *RobotTaskManager
	robotTaskManagerOnce sync.Once
)

// GetRobotTaskManager 鑾峰彇鏈哄櫒浜轰换鍔＄鐞嗗櫒鍗曚緥
func GetRobotTaskManager() *RobotTaskManager {
	robotTaskManagerOnce.Do(func() {
		robotTaskManager = &RobotTaskManager{
			engines: make(map[int64]*RobotEngine),
			stopCh:  make(chan struct{}),
		}
	})
	return robotTaskManager
}

// Start 鍚姩浠诲姟绠＄悊鍣?
func (m *RobotTaskManager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

	g.Log().Info(ctx, "[RobotTaskManager] 鏈哄櫒浜轰换鍔＄鐞嗗櫒鍚姩")

	// 鍚姩鍏ㄥ眬娉㈠姩鐜囬厤缃鐞嗗櫒
	if err := config.GetVolatilityConfigManager().Start(ctx); err != nil {
		return err
	}

	// 璁剧疆WebSocket浠ｇ悊閰嶇疆锛堢洿鎺ヨ幏鍙栵紝閬垮厤寰幆渚濊禆锛?
	wsDialer, err := getWebSocketDialer(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "[RobotTaskManager] 鑾峰彇WebSocket浠ｇ悊閰嶇疆澶辫触: %v", err)
	} else if wsDialer != nil {
		market.GetMarketServiceManager().SetProxyDialer(wsDialer)
	}

	// 鍚姩鍏ㄥ眬琛屾儏鏈嶅姟绠＄悊鍣?
	if err := market.GetMarketServiceManager().Start(ctx); err != nil {
		return err
	}

	// 鍚姩鍏ㄥ眬甯傚満鍒嗘瀽寮曟搸锛堟寜 platform+symbol 璁＄畻甯傚満鐘舵€侊紝鎵€鏈夋満鍣ㄤ汉鍏变韩锛?
	market.GetMarketAnalyzer().Start(ctx)

	// 鍚姩璁㈠崟鐘舵€佸悓姝ユ湇鍔?
	if err := GetOrderStatusSyncService().Start(ctx); err != nil {
		g.Log().Warningf(ctx, "[RobotTaskManager] 鍚姩璁㈠崟鍚屾鏈嶅姟澶辫触: %v", err)
	}

	// 鍚姩鍚屾浠诲姟
	go m.runSyncTask(ctx)

	return nil
}

// Stop 鍋滄浠诲姟绠＄悊鍣?
func (m *RobotTaskManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopCh)

	// 鍋滄鎵€鏈夋満鍣ㄤ汉寮曟搸
	for _, engine := range m.engines {
		engine.Stop()
	}
	m.engines = make(map[int64]*RobotEngine)

	// 鍋滄鍏ㄥ眬鏈嶅姟
	market.GetMarketServiceManager().Stop()
	config.GetVolatilityConfigManager().Stop()
	GetOrderStatusSyncService().Stop()

	g.Log().Info(context.Background(), "[RobotTaskManager] RobotTaskManager 已停止")
}

// IsRunning 妫€鏌ユ槸鍚﹁繍琛屼腑
func (m *RobotTaskManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// runSyncTask 鍚屾鏈哄櫒浜轰换鍔?
// 銆愬仴澹€т紭鍖栥€戞坊鍔?panic 鎭㈠锛岀‘淇濆悓姝ヤ换鍔″紓甯镐笉褰卞搷鍏ㄥ眬寮曟搸
func (m *RobotTaskManager) runSyncTask(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[RobotTaskManager] runSyncTask panic recovered: err=%v", r)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 绔嬪嵆鎵ц涓€娆?
	m.syncRobots(ctx)

	// 銆愬仴澹€т紭鍖栥€戝仴搴锋鏌icker锛堟瘡30绉掓鏌ヤ竴娆★紝杞婚噺绾э級
	healthTicker := time.NewTicker(30 * time.Second)
	defer healthTicker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.syncRobots(ctx)
		case <-healthTicker.C:
			// 銆愬仴澹€т紭鍖栥€戣交閲忕骇鍋ュ悍妫€鏌ワ紙涓嶉樆濉炰富娴佺▼锛?
			m.checkEnginesHealth(ctx)
		}
	}
}

// checkEnginesHealth 妫€鏌ユ墍鏈夊紩鎿庣殑鍋ュ悍鐘舵€侊紙杞婚噺绾э紝楂樻晥锛?
// 銆愬仴澹€т紭鍖栥€戝畾鏈熸鏌ュ紩鎿庡仴搴凤紝鍙婃椂鍙戠幇寮傚父
func (m *RobotTaskManager) checkEnginesHealth(ctx context.Context) {
	engines := m.GetAllEngines()
	if len(engines) == 0 {
		return
	}

	// 銆愭晥鐜囦紭鍖栥€戝苟鍙戞鏌ワ紝浣嗛檺鍒跺苟鍙戞暟锛堥伩鍏嶈繃澶歡oroutine锛?
	const maxConcurrent = 10
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, engine := range engines {
		wg.Add(1)
		go func(e *RobotEngine) {
			defer wg.Done()
			sem <- struct{}{}        // 鑾峰彇淇″彿閲?
			defer func() { <-sem }() // 閲婃斁淇″彿閲?

			if err := e.HealthCheck(); err != nil {
				g.Log().Warningf(ctx, "[RobotTaskManager] 寮曟搸鍋ュ悍妫€鏌ュけ璐? robotId=%d, err=%v",
					e.Robot.Id, err)
				// 娉ㄦ剰锛氳繖閲屼笉鑷姩閲嶅惎锛岄伩鍏嶉绻侀噸鍚鑷撮棶棰?
				// 鍙互璁板綍鍒扮洃鎺х郴缁燂紝鐢辩鐞嗗憳鍐冲畾鏄惁閲嶅惎
			}
		}(engine)
	}

	wg.Wait()
}

// syncRobots 鍚屾鏈哄櫒浜虹姸鎬?
// 銆愭晥鐜囦紭鍖栥€戝噺灏戦攣鎸佹湁鏃堕棿锛屽厛鏌ヨ鏁版嵁搴撳啀鎸佹湁閿?
func (m *RobotTaskManager) syncRobots(ctx context.Context) {
	// 鍏堝鐞嗗畾鏃跺惎鍋?
	m.handleScheduledRobots(ctx)
	// 处理最大运行时长到期（max_runtime）自动停机：先标记暂停，再全平并停止引擎
	m.handleMaxRuntimeRobots(ctx)

	// 銆愭晥鐜囦紭鍖栥€戝厛鏌ヨ鏁版嵁搴擄紙涓嶆寔鏈夐攣锛夛紝鍑忓皯閿佹寔鏈夋椂闂?
	var robots []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where("status", 2).Scan(&robots)
	if err != nil {
		g.Log().Warning(ctx, "[RobotTaskManager] 鏌ヨ鏈哄櫒浜哄け璐?", err)
		return
	}

	// 銆愭晥鐜囦紭鍖栥€戞瀯寤烘椿璺僆D鏄犲皠锛堜笉鎸佹湁閿侊級
	activeIds := make(map[int64]bool, len(robots))
	robotsToUpdate := make(map[int64]*entity.TradingRobot, len(robots))
	robotsToCreate := make([]*entity.TradingRobot, 0)

	for _, robot := range robots {
		activeIds[robot.Id] = true

		// 銆愭晥鐜囦紭鍖栥€戝揩閫熸鏌ユ槸鍚﹀凡瀛樺湪锛堝彧璇绘搷浣滐紝涓嶆寔鏈夐攣锛?
		m.mu.RLock()
		_, exists := m.engines[robot.Id]
		m.mu.RUnlock()

		if exists {
			robotsToUpdate[robot.Id] = robot
		} else {
			robotsToCreate = append(robotsToCreate, robot)
		}
	}

	// 銆愭晥鐜囦紭鍖栥€戝厛鏇存柊宸插瓨鍦ㄧ殑鏈哄櫒浜猴紙蹇€熸搷浣滐級
	if len(robotsToUpdate) > 0 {
		m.mu.RLock()
		for robotId, robot := range robotsToUpdate {
			if engine, ok := m.engines[robotId]; ok {
				engine.UpdateRobot(robot)
			}
		}
		m.mu.RUnlock()
	}

	// 銆愭晥鐜囦紭鍖栥€戝垱寤烘柊寮曟搸锛堝彲鑳借€楁椂锛屽湪閿佸鎵ц锛?
	for _, robot := range robotsToCreate {
		engine, err := m.initRobotEngine(ctx, robot)
		if err != nil {
			g.Log().Warningf(ctx, "[RobotTaskManager] 鍒濆鍖栨満鍣ㄤ汉寮曟搸澶辫触: robotId=%d, err=%v", robot.Id, err)
			continue
		}

		// 鍚姩寮曟搸锛堝彲鑳借€楁椂锛?
		if err := engine.Start(ctx); err != nil {
			g.Log().Warningf(ctx, "[RobotTaskManager] 鍚姩鏈哄櫒浜哄紩鎿庡け璐? robotId=%d, err=%v", robot.Id, err)
			continue
		}

		// 銆愭晥鐜囦紭鍖栥€戝揩閫熸坊鍔犲埌寮曟搸鏄犲皠锛堟寔鏈夐攣鏃堕棿鏈€鐭級
		m.mu.Lock()
		m.engines[robot.Id] = engine
		m.mu.Unlock()

		g.Log().Infof(ctx, "[RobotTaskManager] 鏈哄櫒浜哄紩鎿庡凡鍚姩: robotId=%d, symbol=%s", robot.Id, robot.Symbol)
	}

	// 銆愭晥鐜囦紭鍖栥€戠Щ闄ゅ凡鍋滄鐨勬満鍣ㄤ汉锛堝揩閫熸搷浣滐級
	m.mu.Lock()
	enginesToStop := make([]*RobotEngine, 0)
	for robotId, engine := range m.engines {
		if !activeIds[robotId] {
			enginesToStop = append(enginesToStop, engine)
			delete(m.engines, robotId)
		}
	}
	m.mu.Unlock()

	// 銆愭晥鐜囦紭鍖栥€戝仠姝㈠紩鎿庯紙鍙兘鑰楁椂锛屽湪閿佸鎵ц锛?
	for _, engine := range enginesToStop {
		engine.Stop()
		g.Log().Infof(ctx, "[RobotTaskManager] 鏈哄櫒浜哄紩鎿庡凡鍋滄: robotId=%d", engine.Robot.Id)
	}

	// ===== AutoTrade fallback: process pending signal logs (server-side, no client dependency) =====
	m.processPendingAutoTradeSignals(ctx, robots)

}

// processPendingAutoTradeSignals processes unhandled signal logs and triggers auto trade.
// Why: prevent "signal exists but no auto order" when goroutine didn't run / service restarted / is_processed was NULL.
// Scope: only window_weighted source, executed=0, (is_processed=0 OR NULL), recent 10 minutes.
func (m *RobotTaskManager) processPendingAutoTradeSignals(ctx context.Context, robots []*entity.TradingRobot) {
	if len(robots) == 0 {
		return
	}
	robotIds := make([]int64, 0, len(robots))
	for _, r := range robots {
		if r == nil {
			continue
		}
		robotIds = append(robotIds, r.Id)
	}
	if len(robotIds) == 0 {
		return
	}

	type pendingSignal struct {
		Id             int64       `json:"id"`
		RobotId        int64       `json:"robotId"`
		SignalType     string      `json:"signalType"`
		SignalSource   string      `json:"signalSource"`
		SignalStrength float64     `json:"signalStrength"`
		CurrentPrice   float64     `json:"currentPrice"`
		WindowMin      float64     `json:"windowMin"`
		WindowMax      float64     `json:"windowMax"`
		Threshold      float64     `json:"threshold"`
		Reason         string      `json:"reason"`
		CreatedAt      *gtime.Time `json:"createdAt"`
	}

	since := time.Now().Add(-10 * time.Minute)
	var rows []*pendingSignal
	err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).
		Fields("id,robot_id,signal_type,signal_source,signal_strength,current_price,window_min_price,window_max_price,threshold,reason,created_at").
		WhereIn("robot_id", robotIds).
		Where("executed", 0).
		Where("(is_processed = 0 OR is_processed IS NULL)").
		Where("created_at >= ?", since).
		Where("signal_source", "window_weighted").
		WhereIn("signal_type", []string{"LONG", "SHORT", "long", "short", "???", "???"}).
		OrderDesc("id").
		Limit(200).
		Scan(&rows)
	if err != nil || len(rows) == 0 {
		return
	}

	latestByRobot := make(map[int64]*pendingSignal, len(robots))
	for _, r := range rows {
		if r == nil || r.Id == 0 || r.RobotId == 0 {
			continue
		}
		if _, exists := latestByRobot[r.RobotId]; exists {
			continue
		}
		latestByRobot[r.RobotId] = r
	}
	if len(latestByRobot) == 0 {
		return
	}

	const maxConcurrent = 20
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for robotId, rec := range latestByRobot {
		engine := m.GetEngine(robotId)
		if engine == nil {
			continue
		}

		// only process when auto trade enabled
		engine.mu.RLock()
		autoTradeEnabled := 0
		if engine.Robot != nil {
			autoTradeEnabled = engine.Robot.AutoTradeEnabled
		}
		engine.mu.RUnlock()
		if autoTradeEnabled != 1 {
			continue
		}

		direction := strings.ToUpper(strings.TrimSpace(rec.SignalType))
		if direction == "???" {
			direction = "LONG"
		} else if direction == "???" {
			direction = "SHORT"
		}
		if direction != "LONG" && direction != "SHORT" {
			continue
		}

		action := "OPEN_LONG"
		if direction == "SHORT" {
			action = "OPEN_SHORT"
		}

		sig := &RobotSignal{
			Timestamp:       time.Now(),
			Direction:       direction,
			Strength:        rec.SignalStrength,
			Confidence:      100,
			Action:          action,
			Reason:          rec.Reason,
			WindowMinPrice:  rec.WindowMin,
			WindowMaxPrice:  rec.WindowMax,
			CurrentPrice:    rec.CurrentPrice,
			SignalThreshold: rec.Threshold,
			SignalType:      "window",
		}

		wg.Add(1)
		go func(e *RobotEngine, logId int64, s *RobotSignal) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			// TryAutoTradeAndUpdate is idempotent via atomic is_processed update
			e.Trader.TryAutoTradeAndUpdate(ctx, s, logId)
		}(engine, rec.Id, sig)
	}

	wg.Wait()
}

// initRobotEngine 鍒濆鍖栨満鍣ㄤ汉寮曟搸
func (m *RobotTaskManager) initRobotEngine(ctx context.Context, robot *entity.TradingRobot) (*RobotEngine, error) {
	// 鑾峰彇API閰嶇疆
	var apiConfig *entity.TradingApiConfig
	err := dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig)
	if err != nil || apiConfig == nil {
		return nil, gerror.Newf("API閰嶇疆涓嶅瓨鍦? apiConfigId=%d", robot.ApiConfigId)
	}

	// 鑾峰彇浜ゆ槗鎵€瀹炰緥
	ex, err := GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
	if err != nil {
		return nil, err
	}

	// 鍒涘缓鏈哄櫒浜哄紩鎿?
	engine := NewRobotEngine(ctx, robot, apiConfig, ex)
	return engine, nil
}

// GetEngine 鑾峰彇鏈哄櫒浜哄紩鎿?
func (m *RobotTaskManager) GetEngine(robotId int64) *RobotEngine {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.engines[robotId]
}

// GetAllEngines 鑾峰彇鎵€鏈夋満鍣ㄤ汉寮曟搸
func (m *RobotTaskManager) GetAllEngines() []*RobotEngine {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*RobotEngine, 0, len(m.engines))
	for _, engine := range m.engines {
		result = append(result, engine)
	}
	return result
}

// GetActiveCount 鑾峰彇娲昏穬鏈哄櫒浜烘暟閲?
func (m *RobotTaskManager) GetActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.engines)
}

// GetEngineStatus 鑾峰彇鏈哄櫒浜哄紩鎿庣姸鎬?
func (m *RobotTaskManager) GetEngineStatus(robotId int64) *RobotEngineStatus {
	engine := m.GetEngine(robotId)
	if engine == nil {
		return nil
	}
	return engine.GetStatus()
}

// GetAllEngineStatuses 鑾峰彇鎵€鏈夋満鍣ㄤ汉寮曟搸鐘舵€?
func (m *RobotTaskManager) GetAllEngineStatuses() []*RobotEngineStatus {
	engines := m.GetAllEngines()
	result := make([]*RobotEngineStatus, 0, len(engines))
	for _, engine := range engines {
		result = append(result, engine.GetStatus())
	}
	return result
}

// UpdateRobot 鏇存柊杩愯涓殑鏈哄櫒浜洪厤缃?
func (m *RobotTaskManager) UpdateRobot(robot *entity.TradingRobot) {
	m.mu.RLock()
	engine := m.engines[robot.Id]
	m.mu.RUnlock()

	if engine != nil {
		engine.UpdateRobot(robot)
		g.Log().Infof(context.Background(), "[RobotTaskManager] 鏈哄櫒浜洪厤缃凡鏇存柊: robotId=%d", robot.Id)
	}
}

// positionAmtEpsilon 鎸佷粨鏁伴噺鍒ゆ柇鐨勬瀬灏忛槇鍊硷紙鐢ㄤ簬杩囨护娴偣璇樊/鏋佸皬娈嬬暀浠撲綅锛?
// 鏌愪簺浜ゆ槗鎵€鍙兘杩斿洖 1e-12 绾у埆娈嬬暀鏁伴噺锛屼笉搴旇涓衡€滀粛鏈夋寔浠撯€濄€?
const positionAmtEpsilon = 1e-9

// CloseAllAndWait 鍋滄鍓?杈炬爣鍚庡己鍒跺叏閮ㄦ挙鍗?骞充粨锛堣嫢寮曟搸鏈繍琛屽垯鐩存帴鐢ㄤ氦鏄撴墍API鎵ц锛?
func (m *RobotTaskManager) CloseAllAndWait(ctx context.Context, robotId int64, reason string, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// 浼樺厛璧拌繍琛屼腑寮曟搸锛堜繚璇佽兘鍐欏叆鎵ц鏃ュ織銆佹洿鏂板唴瀛樹笌璁㈠崟鐘舵€侊級
	engine := m.GetEngine(robotId)
	if engine != nil {
		return engine.CloseAllPositionsAndCancelOrders(ctx, reason, timeout)
	}

	// 鍏滃簳锛氬紩鎿庢湭杩愯鏃讹紝鐩存帴鐢ㄤ氦鏄撴墍API杩涜鎾ゅ崟+骞充粨锛堟棤娉曟洿鏂板紩鎿庡唴瀛橈紝浣嗚兘淇濊瘉"鍋滄鍓嶅叏骞?锛?
	var robot *entity.TradingRobot
	if err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot); err != nil {
		return err
	}
	if robot == nil {
		return gerror.Newf("鏈哄櫒浜轰笉瀛樺湪: robotId=%d", robotId)
	}
	var apiConfig *entity.TradingApiConfig
	if err := dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig); err != nil {
		return err
	}
	if apiConfig == nil {
		return gerror.New("API閰嶇疆涓嶅瓨鍦紝鏃犳硶鎵ц鍏ㄥ钩")
	}
	ex, err := GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
	if err != nil {
		return err
	}

	start := time.Now()
	// 鎾ゅ崟
	openOrders, _ := ex.GetOpenOrders(ctx, robot.Symbol)
	for _, o := range openOrders {
		if o == nil || o.OrderId == "" {
			continue
		}
		_, _ = ex.CancelOrder(ctx, robot.Symbol, o.OrderId)
	}

	// 骞充粨
	positions, err := ex.GetPositions(ctx, robot.Symbol)
	if err != nil {
		return err
	}
	for _, pos := range positions {
		if pos == nil || math.Abs(pos.PositionAmt) <= positionAmtEpsilon {
			continue
		}
		_, _ = ex.ClosePosition(ctx, robot.Symbol, pos.PositionSide, math.Abs(pos.PositionAmt))
	}

	// 绛夊緟褰掗浂
	deadline := start.Add(timeout)
	for time.Now().Before(deadline) {
		ps, _ := ex.GetPositions(ctx, robot.Symbol)
		any := false
		for _, p := range ps {
			if p != nil && math.Abs(p.PositionAmt) > positionAmtEpsilon {
				any = true
				break
			}
		}
		if !any {
			return nil
		}
		time.Sleep(400 * time.Millisecond)
	}
	return gerror.New("全部平仓超时，仍检测到持仓未归零")
}

// ReloadRobotStrategy 閲嶆柊鍔犺浇鏈哄櫒浜虹瓥鐣ラ厤缃紙杩愯涓敓鏁堬級
func (m *RobotTaskManager) ReloadRobotStrategy(ctx context.Context, robotId int64) error {
	// 鑾峰彇鏈哄櫒浜哄紩鎿?
	engine := m.GetEngine(robotId)
	if engine == nil {
		return gerror.Newf("鏈哄櫒浜哄紩鎿庝笉瀛樺湪鎴栨湭杩愯: robotId=%d", robotId)
	}

	// 閲嶆柊鍔犺浇鏈哄櫒浜洪厤缃?
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot)
	if err != nil {
		return gerror.Wrap(err, "获取机器人配置失败")
	}
	if robot == nil {
		return gerror.New("鏈哄櫒浜轰笉瀛樺湪")
	}

	// 鏇存柊寮曟搸鐨勬満鍣ㄤ汉閰嶇疆
	engine.UpdateRobot(robot)

	// 閲嶆柊鍔犺浇椋庨櫓閰嶇疆鏄犲皠
	engine.loadRiskConfigFromRobot(ctx)

	// 瑙﹀彂甯傚満鐘舵€侀噸鏂拌瘎浼帮紝寮哄埗鍔犺浇鏈€鏂扮殑绛栫暐鍙傛暟
	if engine.LastAnalysis != nil {
		// 銆愪紭鍖栥€戜粠鍏ㄥ眬甯傚満鍒嗘瀽鍣ㄨ幏鍙栧競鍦虹姸鎬侊紝瑙﹀彂绛栫暐鏇存柊
		globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(engine.Platform, robot.Symbol)
		if globalAnalysis != nil {
			marketState := normalizeMarketState(string(globalAnalysis.MarketState))
			if marketState != "" {
				engine.checkAndUpdateStrategyConfig(ctx, marketState)
			}
		}
	}

	g.Log().Infof(ctx, "[RobotTaskManager] 鏈哄櫒浜虹瓥鐣ラ厤缃凡閲嶆柊鍔犺浇: robotId=%d", robotId)
	return nil
}

// RefreshStrategyParamsByGroupId 鍒锋柊鎸囧畾绛栫暐缁勭殑鎵€鏈夋満鍣ㄤ汉寮曟搸鐨勭瓥鐣ュ弬鏁扮紦瀛?
// 褰撶瓥鐣ユā鏉挎垨绛栫暐缁勮淇敼鏃惰皟鐢ㄦ鏂规硶锛屽己鍒舵墍鏈夌浉鍏冲紩鎿庨噸鏂板姞杞芥渶鏂板弬鏁?
func (m *RobotTaskManager) RefreshStrategyParamsByGroupId(ctx context.Context, groupId int64) error {
	// 鏌ヨ鎵€鏈変娇鐢ㄨ绛栫暐缁勭殑杩愯涓満鍣ㄤ汉
	var robots []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where("strategy_group_id", groupId).
		Where("status", 2). // 鍙埛鏂拌繍琛屼腑鐨勬満鍣ㄤ汉
		Scan(&robots)
	if err != nil {
		return gerror.Wrap(err, "查询机器人失败")
	}

	if len(robots) == 0 {
		g.Log().Debugf(ctx, "[RobotTaskManager] 绛栫暐缁処D=%d 娌℃湁杩愯涓殑鏈哄櫒浜猴紝鏃犻渶鍒锋柊缂撳瓨", groupId)
		return nil
	}

	// 鑾峰彇鎵€鏈夊紩鎿?
	engines := m.GetAllEngines()
	refreshedCount := 0

	for _, engine := range engines {
		// 妫€鏌ュ紩鎿庢槸鍚︿娇鐢ㄨ绛栫暐缁?
		engine.mu.RLock()
		engineGroupId := engine.Robot.StrategyGroupId
		engine.mu.RUnlock()

		if engineGroupId == groupId {
			// 鍒锋柊璇ュ紩鎿庣殑绛栫暐鍙傛暟缂撳瓨
			if err := engine.RefreshStrategyParams(ctx); err != nil {
				g.Log().Warningf(ctx, "[RobotTaskManager] 鍒锋柊鏈哄櫒浜篒D=%d鐨勭瓥鐣ュ弬鏁板け璐? %v", engine.Robot.Id, err)
				continue
			}
			refreshedCount++
		}
	}

	g.Log().Infof(ctx, "[RobotTaskManager] 绛栫暐缁処D=%d 鐨勭瓥鐣ュ弬鏁扮紦瀛樺凡鍒锋柊锛屽叡鍒锋柊 %d 涓満鍣ㄤ汉寮曟搸", groupId, refreshedCount)
	return nil
}

// StartRobot 鎵嬪姩鍚姩鏈哄櫒浜?
func (m *RobotTaskManager) StartRobot(ctx context.Context, robotId int64) error {
	// 鏇存柊鏁版嵁搴撶姸鎬?
	_, err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Data(g.Map{
		// 注意：hg_trading_robot 表字段为 start_time/pause_time/stop_time，不存在 started_at/stopped_at
		"status":     2,          // 运行中
		"start_time": time.Now(), // 启动时间
	}).Update()
	if err != nil {
		return err
	}

	// 绔嬪嵆鍚屾
	m.syncRobots(ctx)
	return nil
}

// StopRobot 鎵嬪姩鍋滄鏈哄櫒浜?
func (m *RobotTaskManager) StopRobot(ctx context.Context, robotId int64, reason string) error {
	// 更新数据库状态（hg_trading_robot 无 stop_reason/stopped_at 字段，避免写入不存在列导致 PG 报错）
	_, err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Data(g.Map{
		"status":     3,          // 暂停
		"pause_time": time.Now(), // 暂停时间
	}).Update()
	if err != nil {
		return err
	}

	// 鍋滄寮曟搸
	m.mu.Lock()
	if engine, ok := m.engines[robotId]; ok {
		engine.Stop()
		delete(m.engines, robotId)
	}
	m.mu.Unlock()

	return nil
}

// handleMaxRuntimeRobots 处理 max_runtime 到期的机器人：自动暂停 + 全平（不依赖客户端）
// 说明：
// - max_runtime 单位为秒
// - 只处理 status=2(运行中) 且 start_time 不为空的机器人
// - 先用“UPDATE ... WHERE status=2”原子抢占，避免并发重复处理
func (m *RobotTaskManager) handleMaxRuntimeRobots(ctx context.Context) {
	now := time.Now()

	// 查询候选机器人（轻量字段即可）
	type runtimeRobot struct {
		Id         int64       `json:"id"`
		UserId     int64       `json:"userId"`
		RobotName  string      `json:"robotName"`
		MaxRuntime int         `json:"maxRuntime"`
		StartTime  *gtime.Time `json:"startTime"`
	}
	var candidates []*runtimeRobot
	_ = dao.TradingRobot.Ctx(ctx).
		Fields("id,user_id,robot_name,max_runtime,start_time").
		Where("status", 2).
		WhereGT("max_runtime", 0).
		WhereNotNull("start_time").
		Scan(&candidates)

	if len(candidates) == 0 {
		return
	}

	for _, r := range candidates {
		if r == nil || r.Id == 0 || r.StartTime == nil || r.MaxRuntime <= 0 {
			continue
		}
		// 未到期
		if r.StartTime.Time.Add(time.Duration(r.MaxRuntime) * time.Second).After(now) {
			continue
		}

		// 原子抢占：把机器人从运行中标记为暂停（避免多个循环重复处理）
		res, err := dao.TradingRobot.Ctx(ctx).
			Where("id", r.Id).
			Where("status", 2).
			Data(g.Map{
				"status":     3,
				"pause_time": now,
			}).
			Update()
		if err != nil {
			g.Log().Warningf(ctx, "[RobotTaskManager] max_runtime 到期处理失败(更新状态): robotId=%d, err=%v", r.Id, err)
			continue
		}
		aff, _ := res.RowsAffected()
		if aff == 0 {
			// 已被其他协程/实例处理
			continue
		}

		// 维护用户活跃机器人数量（尽力而为）
		_, _ = dao.ToogoUser.Ctx(ctx).
			Where("member_id", r.UserId).
			Decrement("active_robot_count", 1)

		// 到期自动全平（不依赖客户端）
		if err := m.CloseAllAndWait(ctx, r.Id, "max_runtime_expired", 30*time.Second); err != nil {
			g.Log().Warningf(ctx, "[RobotTaskManager] max_runtime 到期自动全平失败: robotId=%d, err=%v", r.Id, err)
		}

		// 停止引擎并移除
		m.mu.Lock()
		if engine, ok := m.engines[r.Id]; ok {
			engine.Stop()
			delete(m.engines, r.Id)
		}
		m.mu.Unlock()

		g.Log().Infof(ctx, "[RobotTaskManager] max_runtime 到期自动暂停并全平: robotId=%d, robotName=%s, maxRuntime=%ds",
			r.Id, r.RobotName, r.MaxRuntime)
	}
}

// handleScheduledRobots 澶勭悊瀹氭椂鍚仠鏈哄櫒浜?
func (m *RobotTaskManager) handleScheduledRobots(ctx context.Context) {
	now := time.Now()

	// 1. 鏌ユ壘闇€瑕佸畾鏃跺惎鍔ㄧ殑鏈哄櫒浜?(status=1鏈惎鍔?涓?宸插埌鍚姩鏃堕棿)
	var toStart []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where("status", 1).
		WhereNotNull("schedule_start").
		WhereLTE("schedule_start", now).
		Scan(&toStart)
	if err == nil && len(toStart) > 0 {
		for _, robot := range toStart {
			g.Log().Infof(ctx, "[RobotTaskManager] 瀹氭椂鍚姩鏈哄櫒浜? robotId=%d, robotName=%s", robot.Id, robot.RobotName)
			// 鏇存柊鐘舵€佷负杩愯涓?
			_, _ = dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robot.Id).Data(g.Map{
				"status":         2,
				"start_time":     now,
				"schedule_start": nil, // 娓呯┖宸叉墽琛岀殑瀹氭椂鍚姩鏃堕棿
			}).Update()

			// 鏇存柊鐢ㄦ埛娲昏穬鏈哄櫒浜烘暟閲?
			_, _ = dao.ToogoUser.Ctx(ctx).
				Where("member_id", robot.UserId).
				Increment("active_robot_count", 1)
		}
	}

	// 2. 鏌ユ壘闇€瑕佸畾鏃跺仠姝㈢殑鏈哄櫒浜?(status=2杩愯涓?涓?宸插埌鍋滄鏃堕棿)
	var toStop []*entity.TradingRobot
	err = dao.TradingRobot.Ctx(ctx).
		Where("status", 2).
		WhereNotNull("schedule_stop").
		WhereLTE("schedule_stop", now).
		Scan(&toStop)
	if err == nil && len(toStop) > 0 {
		for _, robot := range toStop {
			g.Log().Infof(ctx, "[RobotTaskManager] 瀹氭椂鍋滄鏈哄櫒浜? robotId=%d, robotName=%s", robot.Id, robot.RobotName)
			// 先原子抢占为“暂停”，避免并发重复处理；并清空 schedule_stop
			res, uerr := dao.TradingRobot.Ctx(ctx).
				Where(dao.TradingRobot.Columns().Id, robot.Id).
				Where("status", 2).
				Data(g.Map{
					"status":        3,
					"pause_time":    now,
					"schedule_stop": nil, // 清空已执行的定时停止时间
				}).Update()
			if uerr != nil {
				g.Log().Warningf(ctx, "[RobotTaskManager] 定时停止更新状态失败: robotId=%d, err=%v", robot.Id, uerr)
				continue
			}
			aff, _ := res.RowsAffected()
			if aff == 0 {
				continue
			}

			// 鏇存柊鐢ㄦ埛娲昏穬鏈哄櫒浜烘暟閲?
			_, _ = dao.ToogoUser.Ctx(ctx).
				Where("member_id", robot.UserId).
				Decrement("active_robot_count", 1)

			// 定时到期：自动全平（不依赖客户端）
			if err := m.CloseAllAndWait(ctx, robot.Id, "schedule_stop", 30*time.Second); err != nil {
				g.Log().Warningf(ctx, "[RobotTaskManager] 定时停止自动全平失败: robotId=%d, err=%v", robot.Id, err)
			}

			// 鍋滄寮曟搸
			m.mu.Lock()
			if engine, ok := m.engines[robot.Id]; ok {
				engine.Stop()
				delete(m.engines, robot.Id)
			}
			m.mu.Unlock()
		}
	}
}

// ==================== 鍏ㄥ眬寮曟搸鐘舵€?====================

// GlobalEngineStatus 鍏ㄥ眬寮曟搸鐘舵€?
type GlobalEngineStatus struct {
	Running      bool                  `json:"running"`
	StartTime    time.Time             `json:"startTime"`
	Uptime       int64                 `json:"uptime"`
	ActiveRobots int                   `json:"activeRobots"`
	Services     *GlobalServicesStatus `json:"services"`
	Robots       []*RobotEngineStatus  `json:"robots"`
}

// GlobalServicesStatus 鍏ㄥ眬鏈嶅姟鐘舵€?
type GlobalServicesStatus struct {
	MarketService *MarketServiceStatus `json:"marketService"`
}

// MarketServiceStatus 琛屾儏鏈嶅姟鐘舵€?
type MarketServiceStatus struct {
	Running       bool                              `json:"running"`
	ExchangeCount int                               `json:"exchangeCount"`
	Exchanges     map[string]*ExchangeServiceStatus `json:"exchanges"`
}

// ExchangeServiceStatus 浜ゆ槗鎵€鏈嶅姟鐘舵€?
type ExchangeServiceStatus struct {
	Platform          string         `json:"platform"`
	Running           bool           `json:"running"`
	SubscriptionCount int            `json:"subscriptionCount"`
	Subscriptions     map[string]int `json:"subscriptions"`
}

// GetGlobalStatus 鑾峰彇鍏ㄥ眬寮曟搸鐘舵€?
func (m *RobotTaskManager) GetGlobalStatus() *GlobalEngineStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &GlobalEngineStatus{
		Running:      m.running,
		ActiveRobots: len(m.engines),
		Services:     m.getServicesStatus(),
		Robots:       make([]*RobotEngineStatus, 0, len(m.engines)),
	}

	// 鑾峰彇鎵€鏈夋満鍣ㄤ汉鐘舵€?
	for _, engine := range m.engines {
		status.Robots = append(status.Robots, engine.GetStatus())
	}

	return status
}

// getServicesStatus 鑾峰彇鏈嶅姟鐘舵€?
func (m *RobotTaskManager) getServicesStatus() *GlobalServicesStatus {
	services := &GlobalServicesStatus{
		MarketService: &MarketServiceStatus{
			Running:   market.GetMarketServiceManager().IsRunning(),
			Exchanges: make(map[string]*ExchangeServiceStatus),
		},
	}

	// 鑾峰彇鍚勪氦鏄撴墍鏈嶅姟鐘舵€?
	allServices := market.GetMarketServiceManager().GetAllServices()
	services.MarketService.ExchangeCount = len(allServices)

	for platform, svc := range allServices {
		services.MarketService.Exchanges[platform] = &ExchangeServiceStatus{
			Platform:          platform,
			Running:           true,
			SubscriptionCount: svc.GetSubscriptionCount(),
			Subscriptions:     svc.GetAllSubscriptions(),
		}
	}

	return services
}

// ==================== 鍏煎鏃ф帴鍙?====================

// GetActiveRobotCount 鑾峰彇娲昏穬鏈哄櫒浜烘暟閲忥紙鍏煎鏃ф帴鍙ｏ級
func (m *RobotTaskManager) GetActiveRobotCount() int {
	return m.GetActiveCount()
}

// ManagedRobot 琚鐞嗙殑鏈哄櫒浜猴紙鍏煎鏃ф帴鍙ｏ級
type ManagedRobot struct {
	Robot            *entity.TradingRobot
	Exchange         exchange.Exchange
	APIConfig        *entity.TradingApiConfig
	Platform         string
	DirectionSignal  *market.DirectionSignal
	CurrentPositions []*exchange.Position
	AccountBalance   *exchange.Balance
}

// GetAllManagedRobots 鑾峰彇鎵€鏈夎绠＄悊鐨勬満鍣ㄤ汉锛堝吋瀹规棫鎺ュ彛锛?
func (m *RobotTaskManager) GetAllManagedRobots() []*ManagedRobot {
	engines := m.GetAllEngines()
	result := make([]*ManagedRobot, 0, len(engines))

	for _, engine := range engines {
		result = append(result, m.engineToManagedRobot(engine))
	}

	return result
}

// GetManagedRobot 鑾峰彇鍗曚釜琚鐞嗙殑鏈哄櫒浜猴紙鍏煎鏃ф帴鍙ｏ級
func (m *RobotTaskManager) GetManagedRobot(robotId int64) *ManagedRobot {
	engine := m.GetEngine(robotId)
	if engine == nil {
		return nil
	}
	return m.engineToManagedRobot(engine)
}

// engineToManagedRobot 灏嗗紩鎿庤浆鎹负 ManagedRobot (宸茬Щ闄iskEvaluation)
func (m *RobotTaskManager) engineToManagedRobot(engine *RobotEngine) *ManagedRobot {
	managed := &ManagedRobot{
		Robot:            engine.Robot,
		Exchange:         engine.Exchange,
		APIConfig:        engine.APIConfig,
		Platform:         engine.Platform,
		CurrentPositions: engine.CurrentPositions,
		AccountBalance:   engine.AccountBalance,
	}

	// 杞崲鏂瑰悜淇″彿
	if engine.LastSignal != nil {
		managed.DirectionSignal = &market.DirectionSignal{
			Platform:   engine.Platform,
			Symbol:     engine.Robot.Symbol,
			Direction:  market.Direction(engine.LastSignal.Direction),
			Strength:   engine.LastSignal.Strength,
			Confidence: engine.LastSignal.Confidence,
		}
	}

	return managed
}

// getWebSocketDialer 鑾峰彇WebSocket浠ｇ悊鎷ㄥ彿鍣紙閬垮厤寰幆渚濊禆锛岀洿鎺ュ湪鏈寘瀹炵幇锛?
func getWebSocketDialer(ctx context.Context) (func(network, addr string) (net.Conn, error), error) {
	// 鑾峰彇鍏ㄥ眬浠ｇ悊閰嶇疆
	var proxyConfig *entity.TradingProxyConfig
	err := dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 琛ㄧず鍏ㄥ眬閰嶇疆
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 琛ㄧず鍏ㄥ眬閰嶇疆
		Where(dao.TradingProxyConfig.Columns().Enabled, 1).
		Scan(&proxyConfig)

	if err != nil {
		return nil, err
	}

	// 濡傛灉娌℃湁鍚敤浠ｇ悊锛岃繑鍥瀗il锛堜娇鐢ㄩ粯璁ゆ嫧鍙峰櫒锛?
	if proxyConfig == nil {
		return nil, nil
	}

	// 瑙ｅ瘑瀵嗙爜
	password := ""
	if proxyConfig.AuthEnabled == 1 && proxyConfig.Password != "" {
		password, err = encrypt.AesDecrypt(proxyConfig.Password)
		if err != nil {
			return nil, gerror.Wrap(err, "瀵嗙爜瑙ｅ瘑澶辫触")
		}
	}

	// 鏍规嵁浠ｇ悊绫诲瀷鍒涘缓鎷ㄥ彿鍣?
	if proxyConfig.ProxyType == "http" || proxyConfig.ProxyType == "https" {
		// HTTP浠ｇ悊锛氫娇鐢–ONNECT鏂规硶寤虹珛闅ч亾
		return func(network, addr string) (net.Conn, error) {
			// 杩炴帴鍒颁唬鐞嗘湇鍔″櫒
			proxyConn, err := net.DialTimeout("tcp", proxyConfig.ProxyAddress, 30*time.Second)
			if err != nil {
				return nil, fmt.Errorf("杩炴帴HTTP浠ｇ悊澶辫触: %v", err)
			}

			// 鏋勫缓CONNECT璇锋眰
			connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)

			// 濡傛灉闇€瑕佽璇?
			if proxyConfig.AuthEnabled == 1 && proxyConfig.Username != "" {
				auth := proxyConfig.Username
				if password != "" {
					auth += ":" + password
				}
				encoded := base64EncodeSimple([]byte(auth))
				connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", encoded)
			}

			connectReq += "\r\n"

			// 鍙戦€丆ONNECT璇锋眰
			_, err = proxyConn.Write([]byte(connectReq))
			if err != nil {
				proxyConn.Close()
				return nil, fmt.Errorf("鍙戦€丆ONNECT璇锋眰澶辫触: %v", err)
			}

			// 璇诲彇鍝嶅簲
			response := make([]byte, 1024)
			n, err := proxyConn.Read(response)
			if err != nil {
				proxyConn.Close()
				return nil, fmt.Errorf("璇诲彇浠ｇ悊鍝嶅簲澶辫触: %v", err)
			}

			// 妫€鏌ュ搷搴旂姸鎬?
			respStr := string(response[:n])
			if !strings.Contains(respStr, "200") {
				proxyConn.Close()
				return nil, fmt.Errorf("HTTP浠ｇ悊CONNECT澶辫触: %s", respStr)
			}

			return proxyConn, nil
		}, nil
	} else if proxyConfig.ProxyType == "socks5" {
		// SOCKS5浠ｇ悊
		var auth *proxy.Auth
		if proxyConfig.AuthEnabled == 1 && proxyConfig.Username != "" {
			auth = &proxy.Auth{
				User:     proxyConfig.Username,
				Password: password,
			}
		}

		dialer, err := proxy.SOCKS5("tcp", proxyConfig.ProxyAddress, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		return dialer.Dial, nil
	}

	return nil, gerror.Newf("涓嶆敮鎸佺殑浠ｇ悊绫诲瀷: %s", proxyConfig.ProxyType)
}

// base64EncodeSimple 绠€鍗曠殑base64缂栫爜
func base64EncodeSimple(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		var n uint32
		var padding int

		n = uint32(data[i]) << 16
		if i+1 < len(data) {
			n |= uint32(data[i+1]) << 8
		} else {
			padding++
		}
		if i+2 < len(data) {
			n |= uint32(data[i+2])
		} else {
			padding++
		}

		result.WriteByte(base64Chars[(n>>18)&0x3F])
		result.WriteByte(base64Chars[(n>>12)&0x3F])
		if padding < 2 {
			result.WriteByte(base64Chars[(n>>6)&0x3F])
		} else {
			result.WriteByte('=')
		}
		if padding < 1 {
			result.WriteByte(base64Chars[n&0x3F])
		} else {
			result.WriteByte('=')
		}
	}

	return result.String()
}
