// Package config 全局波动率配置管理器
// 为所有机器人提供统一的波动率配置服务，避免重复查询和缓存
package config

import (
	"context"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// VolatilityConfigManager 全局波动率配置管理器（单例）
type VolatilityConfigManager struct {
	mu sync.RWMutex

	// 配置缓存 key: symbol (空字符串表示全局配置)
	configs map[string]*VolatilityConfig

	// 全局默认配置
	defaultConfig *VolatilityConfig

	// 最后更新时间
	lastUpdateTime map[string]time.Time

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// VolatilityConfig 波动率配置（适配新算法）
type VolatilityConfig struct {
	Symbol                  string    // 交易对（空字符串表示全局）
	HighVolatilityThreshold float64   // 高波动阈值 (HighV)
	LowVolatilityThreshold  float64   // 低波动阈值 (LowV)
	TrendStrengthThreshold  float64   // 趋势阈值 (TrendV)
	DThreshold              float64   // 方向一致性阈值 (DThreshold)
	Delta1m                 float64   // 1分钟周期delta
	Delta5m                 float64   // 5分钟周期delta
	Delta15m                float64   // 15分钟周期delta
	Delta30m                float64   // 30分钟周期delta
	Delta1h                 float64   // 1小时周期delta
	Weight1m                float64   // 1分钟权重
	Weight5m                float64   // 5分钟权重
	Weight15m               float64   // 15分钟权重
	Weight30m               float64   // 30分钟权重
	Weight1h                float64   // 1小时权重
	IsActive                int       // 是否启用
	UpdatedAt               time.Time // 配置更新时间
}

var (
	volatilityConfigManager     *VolatilityConfigManager
	volatilityConfigManagerOnce sync.Once
)

// GetVolatilityConfigManager 获取全局配置管理器单例
func GetVolatilityConfigManager() *VolatilityConfigManager {
	volatilityConfigManagerOnce.Do(func() {
		volatilityConfigManager = &VolatilityConfigManager{
			configs:        make(map[string]*VolatilityConfig),
			lastUpdateTime: make(map[string]time.Time),
			stopCh:         make(chan struct{}),
			defaultConfig: &VolatilityConfig{
				Symbol:                  "",
				HighVolatilityThreshold: 2.0,  // HighV = 2.0 (BTCUSDT默认值)
				LowVolatilityThreshold:  0.9,  // LowV = 0.9 (BTCUSDT默认值)
				TrendStrengthThreshold:  1.2,  // TrendV = 1.2 (BTCUSDT默认值)
				DThreshold:              0.7,  // DThreshold = 0.7 (BTCUSDT默认值)
				Delta1m:                2.0,
				Delta5m:                2.0,
				Delta15m:               3.0,
				Delta30m:               3.0,
				Delta1h:                5.0,
				Weight1m:                0.18,  // BTCUSDT权重: 1m=0.18
				Weight5m:                0.25,  // BTCUSDT权重: 5m=0.25
				Weight15m:               0.27,  // BTCUSDT权重: 15m=0.27
				Weight30m:               0.20,  // BTCUSDT权重: 30m=0.20
				Weight1h:                0.10,  // BTCUSDT权重: 1h=0.10
				IsActive:                1,
			},
		}
	})
	return volatilityConfigManager
}

// Start 启动配置管理器
func (m *VolatilityConfigManager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

	g.Log().Info(ctx, "[VolatilityConfigManager] 全局波动率配置管理器启动")

	// 立即加载一次配置
	m.LoadAllConfigs(ctx)

	// 启动定时刷新任务（每5分钟刷新一次）
	go m.runRefreshTask(ctx)

	return nil
}

// Stop 停止配置管理器
func (m *VolatilityConfigManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopCh)
}

// GetConfig 获取指定交易对的配置（只读，无锁竞争）
// 优先级：交易对特定配置 > 全局配置 > 默认配置
func (m *VolatilityConfigManager) GetConfig(symbol string) *VolatilityConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 1. 查找交易对特定配置
	if config, ok := m.configs[symbol]; ok && config.IsActive == 1 {
		return config
	}

	// 2. 查找全局配置
	if config, ok := m.configs[""]; ok && config.IsActive == 1 {
		return config
	}

	// 3. 返回默认配置
	return m.defaultConfig
}

// LoadAllConfigs 从数据库加载所有配置
func (m *VolatilityConfigManager) LoadAllConfigs(ctx context.Context) error {
	cols := dao.ToogoVolatilityConfig.Columns()

	// 查询所有启用的配置
	var dbConfigs []*entity.ToogoVolatilityConfig
	err := dao.ToogoVolatilityConfig.Ctx(ctx).
		Where(cols.IsActive, 1).
		OrderAsc(cols.Symbol). // NULL值排在前面（全局配置）
		Scan(&dbConfigs)

	if err != nil {
		g.Log().Errorf(ctx, "[VolatilityConfigManager] 加载配置失败: %v", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空旧配置
	m.configs = make(map[string]*VolatilityConfig)

	// 加载新配置
	for _, dbConfig := range dbConfigs {
		symbol := ""
		if dbConfig.Symbol != nil {
			symbol = *dbConfig.Symbol
		}

		config := &VolatilityConfig{
			Symbol:                  symbol,
			HighVolatilityThreshold: dbConfig.HighVolatilityThreshold,
			LowVolatilityThreshold:  dbConfig.LowVolatilityThreshold,
			TrendStrengthThreshold:  dbConfig.TrendStrengthThreshold,
			DThreshold:              dbConfig.DThreshold,
			Delta1m:                 dbConfig.Delta1m,
			Delta5m:                 dbConfig.Delta5m,
			Delta15m:                dbConfig.Delta15m,
			Delta30m:                dbConfig.Delta30m,
			Delta1h:                 dbConfig.Delta1h,
			Weight1m:                dbConfig.Weight1m,
			Weight5m:                dbConfig.Weight5m,
			Weight15m:               dbConfig.Weight15m,
			Weight30m:               dbConfig.Weight30m,
			Weight1h:                dbConfig.Weight1h,
			IsActive:                dbConfig.IsActive,
			UpdatedAt:               dbConfig.UpdatedAt.Time,
		}

		m.configs[symbol] = config
		m.lastUpdateTime[symbol] = time.Now()
	}

	g.Log().Infof(ctx, "[VolatilityConfigManager] 加载配置完成: 共%d个配置", len(m.configs))
	return nil
}

// ReloadConfig 重新加载指定交易对的配置（支持热更新）
func (m *VolatilityConfigManager) ReloadConfig(ctx context.Context, symbol string) error {
	config, err := service.ToogoVolatilityConfig().GetBySymbol(ctx, symbol)
	if err != nil {
		return err
	}

	if config == nil {
		// 如果配置被删除，从缓存中移除
		m.mu.Lock()
		delete(m.configs, symbol)
		m.mu.Unlock()
		g.Log().Infof(ctx, "[VolatilityConfigManager] 配置已删除: symbol=%s", symbol)
		return nil
	}

	configSymbol := ""
	if config.Symbol != nil {
		configSymbol = *config.Symbol
	}

	volatilityConfig := &VolatilityConfig{
		Symbol:                  configSymbol,
		HighVolatilityThreshold: config.HighVolatilityThreshold,
		LowVolatilityThreshold:  config.LowVolatilityThreshold,
		TrendStrengthThreshold:  config.TrendStrengthThreshold,
		DThreshold:              config.DThreshold,
		Delta1m:                 config.Delta1m,
		Delta5m:                 config.Delta5m,
		Delta15m:                config.Delta15m,
		Delta30m:                config.Delta30m,
		Delta1h:                 config.Delta1h,
		Weight1m:                config.Weight1m,
		Weight5m:                config.Weight5m,
		Weight15m:               config.Weight15m,
		Weight30m:               config.Weight30m,
		Weight1h:                config.Weight1h,
		IsActive:                config.IsActive,
		UpdatedAt:               config.UpdatedAt.Time,
	}

	m.mu.Lock()
	m.configs[symbol] = volatilityConfig
	m.lastUpdateTime[symbol] = time.Now()
	m.mu.Unlock()

	g.Log().Infof(ctx, "[VolatilityConfigManager] 配置已更新: symbol=%s", symbol)
	return nil
}

// runRefreshTask 定时刷新任务
func (m *VolatilityConfigManager) runRefreshTask(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒刷新一次（兜底机制）
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// 定时刷新作为兜底，正常情况下配置修改会立即触发 ReloadConfig
			if err := m.LoadAllConfigs(ctx); err != nil {
				g.Log().Errorf(ctx, "[VolatilityConfigManager] 定时刷新配置失败: %v", err)
			} else {
				g.Log().Debugf(ctx, "[VolatilityConfigManager] 定时刷新配置完成")
			}
		}
	}
}

// GetAllConfigs 获取所有配置（用于监控和调试）
func (m *VolatilityConfigManager) GetAllConfigs() map[string]*VolatilityConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本，避免外部修改
	result := make(map[string]*VolatilityConfig, len(m.configs))
	for k, v := range m.configs {
		configCopy := *v
		result[k] = &configCopy
	}
	return result
}

// GetStats 获取统计信息
func (m *VolatilityConfigManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_configs":    len(m.configs),
		"running":          m.running,
		"last_update_time": m.lastUpdateTime,
	}
}

