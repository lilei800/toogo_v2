// Package config 全局波动率配置管理器 V2 - 支持配置变更推送
package config

import (
	"context"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// VolatilityConfigManagerV2 全局波动率配置管理器（支持配置变更推送）
type VolatilityConfigManagerV2 struct {
	mu sync.RWMutex

	// 配置缓存 key: symbol
	configs map[string]*VolatilityConfig

	// 全局默认配置
	defaultConfig *VolatilityConfig

	// 配置变更通知通道（观察者模式）
	changeNotifiers []chan ConfigChangeEvent

	// 最后更新时间
	lastUpdateTime map[string]time.Time

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	Symbol    string                  // 变更的交易对
	OldConfig *VolatilityConfig       // 旧配置
	NewConfig *VolatilityConfig       // 新配置
	ChangeType string                 // 变更类型: "create", "update", "delete"
	Timestamp  time.Time               // 变更时间
}

// RegisterChangeNotifier 注册配置变更监听器（机器人注册自己）
func (m *VolatilityConfigManagerV2) RegisterChangeNotifier() chan ConfigChangeEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	notifier := make(chan ConfigChangeEvent, 10) // 缓冲10个事件
	m.changeNotifiers = append(m.changeNotifiers, notifier)
	
	g.Log().Debugf(context.Background(), "[VolatilityConfigManager] 注册配置变更监听器, 当前监听器数量: %d", len(m.changeNotifiers))
	return notifier
}

// UnregisterChangeNotifier 取消注册配置变更监听器
func (m *VolatilityConfigManagerV2) UnregisterChangeNotifier(notifier chan ConfigChangeEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, n := range m.changeNotifiers {
		if n == notifier {
			// 关闭通道
			close(n)
			// 从列表中移除
			m.changeNotifiers = append(m.changeNotifiers[:i], m.changeNotifiers[i+1:]...)
			break
		}
	}
	
	g.Log().Debugf(context.Background(), "[VolatilityConfigManager] 取消注册配置变更监听器, 当前监听器数量: %d", len(m.changeNotifiers))
}

// notifyConfigChange 通知所有监听器配置变更
func (m *VolatilityConfigManagerV2) notifyConfigChange(event ConfigChangeEvent) {
	m.mu.RLock()
	notifiers := make([]chan ConfigChangeEvent, len(m.changeNotifiers))
	copy(notifiers, m.changeNotifiers)
	m.mu.RUnlock()

	// 异步通知所有监听器（非阻塞）
	for _, notifier := range notifiers {
		go func(ch chan ConfigChangeEvent) {
			select {
			case ch <- event:
				// 成功发送
			case <-time.After(100 * time.Millisecond):
				// 超时，防止阻塞
				g.Log().Warning(context.Background(), "[VolatilityConfigManager] 配置变更通知超时")
			}
		}(notifier)
	}
	
	g.Log().Infof(context.Background(), "[VolatilityConfigManager] 配置变更通知已发送: symbol=%s, type=%s, 监听器数量=%d", 
		event.Symbol, event.ChangeType, len(notifiers))
}

// ReloadConfigV2 重新加载配置（支持推送通知）
func (m *VolatilityConfigManagerV2) ReloadConfigV2(ctx context.Context, symbol string) error {
	// 保存旧配置
	m.mu.RLock()
	oldConfig := m.configs[symbol]
	m.mu.RUnlock()

	// 从数据库加载新配置
	config, err := loadConfigFromDB(ctx, symbol)
	if err != nil {
		return err
	}

	changeType := "update"
	if oldConfig == nil && config != nil {
		changeType = "create"
	} else if oldConfig != nil && config == nil {
		changeType = "delete"
	}

	// 更新缓存
	m.mu.Lock()
	if config != nil {
		m.configs[symbol] = config
		m.lastUpdateTime[symbol] = time.Now()
	} else {
		delete(m.configs, symbol)
	}
	m.mu.Unlock()

	// 推送配置变更事件
	event := ConfigChangeEvent{
		Symbol:     symbol,
		OldConfig:  oldConfig,
		NewConfig:  config,
		ChangeType: changeType,
		Timestamp:  time.Now(),
	}
	m.notifyConfigChange(event)

	g.Log().Infof(ctx, "[VolatilityConfigManager] 配置已更新并推送通知: symbol=%s, type=%s", symbol, changeType)
	return nil
}

// 辅助函数：从数据库加载配置
func loadConfigFromDB(ctx context.Context, symbol string) (*VolatilityConfig, error) {
	cols := dao.ToogoVolatilityConfig.Columns()
	
	var dbConfig *entity.ToogoVolatilityConfig
	err := dao.ToogoVolatilityConfig.Ctx(ctx).
		Where(cols.Symbol, symbol).
		Where(cols.IsActive, 1).
		Scan(&dbConfig)

	if err != nil || dbConfig == nil {
		return nil, err
	}

	configSymbol := ""
	if dbConfig.Symbol != nil {
		configSymbol = *dbConfig.Symbol
	}

	return &VolatilityConfig{
		Symbol:                  configSymbol,
		HighVolatilityThreshold: dbConfig.HighVolatilityThreshold,
		LowVolatilityThreshold:  dbConfig.LowVolatilityThreshold,
		TrendStrengthThreshold:  dbConfig.TrendStrengthThreshold,
		Weight1m:                dbConfig.Weight1m,
		Weight5m:                dbConfig.Weight5m,
		Weight15m:               dbConfig.Weight15m,
		Weight30m:               dbConfig.Weight30m,
		Weight1h:                dbConfig.Weight1h,
		IsActive:                dbConfig.IsActive,
		UpdatedAt:               dbConfig.UpdatedAt.Time,
	}, nil
}

