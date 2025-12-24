// Package market 策略配置管理
// 负责解析和管理策略模板配置
package market

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
)

// StrategyConfigManager 策略配置管理器（单例）
type StrategyConfigManager struct {
	mu sync.RWMutex

	// 策略模板缓存 key: strategy_key
	templates map[string]*StrategyTemplate

	// 机器人策略缓存 key: robotId
	robotStrategies map[int64]*RobotStrategy

	// 缓存更新时间
	lastUpdate time.Time
}

// StrategyTemplate 策略模板
type StrategyTemplate struct {
	Id                     int64   `json:"id"`
	GroupId                int64   `json:"groupId"`
	StrategyKey            string  `json:"strategyKey"`
	StrategyName           string  `json:"strategyName"`
	RiskPreference         string  `json:"riskPreference"`  // conservative/balanced/aggressive
	MarketState            string  `json:"marketState"`     // trend/volatile/all
	MonitorWindow          int     `json:"monitorWindow"`   // 监控窗口（秒）
	VolatilityThreshold    float64 `json:"volatilityThreshold"` // 波动点数阈值
	Leverage               int     `json:"leverage"`        // 杠杆倍数
	MarginPercent          float64 `json:"marginPercent"`   // 保证金比例
	StopLossPercent        float64 `json:"stopLossPercent"`
	ProfitRetreatPercent   float64 `json:"profitRetreatPercent"`
	AutoStartRetreatPercent float64 `json:"autoStartRetreatPercent"`
	
	// 扩展配置（从config_json解析）
	ExtConfig *StrategyExtConfig `json:"extConfig"`
}

// StrategyExtConfig 策略扩展配置
type StrategyExtConfig struct {
	Exchange           string  `json:"exchange"`           // 交易所
	Symbol             string  `json:"symbol"`             // 交易对
	OrderType          string  `json:"orderType"`          // 订单类型
	MarginMode         string  `json:"marginMode"`         // 保证金模式
	
	// 高级止盈止损配置
	TrailingStopEnabled    bool    `json:"trailingStopEnabled"`    // 启用追踪止损
	TrailingStopPercent    float64 `json:"trailingStopPercent"`    // 追踪止损比例
	PartialTakeProfitEnabled bool  `json:"partialTakeProfitEnabled"` // 启用分批止盈
	TakeProfitLevels       []TakeProfitLevel `json:"takeProfitLevels"` // 止盈档位
	
	// 信号过滤配置
	MinSignalStrength      float64 `json:"minSignalStrength"`      // 最小信号强度
	MinSignalConfidence    float64 `json:"minSignalConfidence"`    // 最小信号置信度
	RequireMultiTimeframe  bool    `json:"requireMultiTimeframe"`  // 要求多周期确认
	MinAlignedTimeframes   int     `json:"minAlignedTimeframes"`   // 最小一致周期数
	
	// 仓位管理配置
	MaxPositions           int     `json:"maxPositions"`           // 最大持仓数
	MaxDailyTrades         int     `json:"maxDailyTrades"`         // 每日最大交易次数
	CooldownSeconds        int     `json:"cooldownSeconds"`        // 交易冷却时间（秒）
}

// TakeProfitLevel 止盈档位
type TakeProfitLevel struct {
	Percent    float64 `json:"percent"`    // 止盈比例
	CloseRatio float64 `json:"closeRatio"` // 平仓比例
}

// RobotStrategy 机器人策略配置
type RobotStrategy struct {
	RobotId          int64
	Template         *StrategyTemplate
	CustomConfig     *StrategyExtConfig // 机器人自定义配置（覆盖模板）
	CurrentLeverage  int                // 当前使用的杠杆
	CurrentMargin    float64            // 当前使用的保证金比例
	LastTradeTime    time.Time          // 最后交易时间
	TodayTradeCount  int                // 今日交易次数
}

var (
	strategyConfigManager     *StrategyConfigManager
	strategyConfigManagerOnce sync.Once
)

// GetStrategyConfigManager 获取策略配置管理器单例
func GetStrategyConfigManager() *StrategyConfigManager {
	strategyConfigManagerOnce.Do(func() {
		strategyConfigManager = &StrategyConfigManager{
			templates:       make(map[string]*StrategyTemplate),
			robotStrategies: make(map[int64]*RobotStrategy),
		}
	})
	return strategyConfigManager
}

// LoadTemplates 加载策略模板
func (m *StrategyConfigManager) LoadTemplates(ctx context.Context) error {
	var templates []*entity.TradingStrategyTemplate
	err := dao.TradingStrategyTemplate.Ctx(ctx).
		Where("is_active", 1).
		OrderAsc("sort").
		Scan(&templates)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range templates {
		template := &StrategyTemplate{
			Id:                     t.Id,
			GroupId:                t.GroupId,
			StrategyKey:            t.StrategyKey,
			StrategyName:           t.StrategyName,
			RiskPreference:         t.RiskPreference,
			MarketState:            t.MarketState,
			MonitorWindow:          t.MonitorWindow,
			VolatilityThreshold:    t.VolatilityThreshold,
			Leverage:               t.Leverage,
			MarginPercent:          t.MarginPercent,
			StopLossPercent:        t.StopLossPercent,
			ProfitRetreatPercent:   t.ProfitRetreatPercent,
			AutoStartRetreatPercent: t.AutoStartRetreatPercent,
		}

		// 解析扩展配置
		if t.ConfigJson != "" {
			var extConfig StrategyExtConfig
			if err := json.Unmarshal([]byte(t.ConfigJson), &extConfig); err == nil {
				template.ExtConfig = &extConfig
			}
		}

		// 设置默认值
		if template.ExtConfig == nil {
			template.ExtConfig = m.getDefaultExtConfig()
		}

		m.templates[t.StrategyKey] = template
	}

	m.lastUpdate = time.Now()
	g.Log().Infof(ctx, "[StrategyConfigManager] 加载%d个策略模板", len(templates))

	return nil
}

// getDefaultExtConfig 获取默认扩展配置
func (m *StrategyConfigManager) getDefaultExtConfig() *StrategyExtConfig {
	return &StrategyExtConfig{
		OrderType:              "market",
		MarginMode:             "isolated",
		TrailingStopEnabled:    false,
		TrailingStopPercent:    1.0,
		PartialTakeProfitEnabled: false,
		TakeProfitLevels: []TakeProfitLevel{
			{Percent: 5, CloseRatio: 0.3},
			{Percent: 10, CloseRatio: 0.5},
			{Percent: 15, CloseRatio: 1.0},
		},
		MinSignalStrength:      30,
		MinSignalConfidence:    50,
		RequireMultiTimeframe:  true,
		MinAlignedTimeframes:   3,
		MaxPositions:           1,
		MaxDailyTrades:         10,
		CooldownSeconds:        60,
	}
}

// GetTemplate 获取策略模板
func (m *StrategyConfigManager) GetTemplate(strategyKey string) *StrategyTemplate {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.templates[strategyKey]
}

// GetTemplateByRiskAndMarket 根据风险偏好和市场状态获取模板
func (m *StrategyConfigManager) GetTemplateByRiskAndMarket(riskPreference, marketState string) *StrategyTemplate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, t := range m.templates {
		if t.RiskPreference == riskPreference && (t.MarketState == marketState || t.MarketState == "all") {
			return t
		}
	}
	return nil
}

// GetRobotStrategy 获取机器人策略配置
func (m *StrategyConfigManager) GetRobotStrategy(ctx context.Context, robotId int64) *RobotStrategy {
	m.mu.RLock()
	rs, exists := m.robotStrategies[robotId]
	m.mu.RUnlock()

	if exists {
		return rs
	}

	// 从数据库加载
	return m.loadRobotStrategy(ctx, robotId)
}

// loadRobotStrategy 从数据库加载机器人策略
func (m *StrategyConfigManager) loadRobotStrategy(ctx context.Context, robotId int64) *RobotStrategy {
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot)
	if err != nil || robot == nil {
		return nil
	}

	rs := &RobotStrategy{
		RobotId:         robotId,
		CurrentLeverage: robot.Leverage,
		CurrentMargin:   robot.MarginPercent,
	}

	// 解析current_strategy字段
	if robot.CurrentStrategy != "" {
		var strategyConfig struct {
			StrategyKey string `json:"strategyKey"`
		}
		if err := json.Unmarshal([]byte(robot.CurrentStrategy), &strategyConfig); err == nil {
			if strategyConfig.StrategyKey != "" {
				rs.Template = m.GetTemplate(strategyConfig.StrategyKey)
			}
		}

		// 解析自定义配置
		var customConfig StrategyExtConfig
		if err := json.Unmarshal([]byte(robot.CurrentStrategy), &customConfig); err == nil {
			rs.CustomConfig = &customConfig
		}
	}

	// 如果没有模板，使用默认配置
	if rs.Template == nil {
		rs.Template = &StrategyTemplate{
			StrategyKey:            "default",
			StrategyName:           "默认策略",
			RiskPreference:         robot.RiskPreference,
			MonitorWindow:          300,
			VolatilityThreshold:    100,
			Leverage:               5,
			MarginPercent:          10,
			StopLossPercent:        robot.StopLossPercent,
			ProfitRetreatPercent:   robot.ProfitRetreatPercent,
			AutoStartRetreatPercent: robot.AutoStartRetreatPercent,
			ExtConfig:              m.getDefaultExtConfig(),
		}
	}

	m.mu.Lock()
	m.robotStrategies[robotId] = rs
	m.mu.Unlock()

	return rs
}

// UpdateRobotStrategy 更新机器人策略
func (m *StrategyConfigManager) UpdateRobotStrategy(robotId int64, leverage int, margin float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rs, exists := m.robotStrategies[robotId]; exists {
		rs.CurrentLeverage = leverage
		rs.CurrentMargin = margin
	}
}

// RecordTrade 记录交易
func (m *StrategyConfigManager) RecordTrade(robotId int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rs, exists := m.robotStrategies[robotId]; exists {
		rs.LastTradeTime = time.Now()
		rs.TodayTradeCount++
	}
}

// CanTrade 检查是否可以交易
func (m *StrategyConfigManager) CanTrade(robotId int64) (bool, string) {
	m.mu.RLock()
	rs, exists := m.robotStrategies[robotId]
	m.mu.RUnlock()

	if !exists {
		return true, ""
	}

	extConfig := rs.Template.ExtConfig
	if extConfig == nil {
		return true, ""
	}

	// 检查冷却时间
	if extConfig.CooldownSeconds > 0 {
		cooldown := time.Duration(extConfig.CooldownSeconds) * time.Second
		if time.Since(rs.LastTradeTime) < cooldown {
			return false, "交易冷却中"
		}
	}

	// 检查每日交易次数
	if extConfig.MaxDailyTrades > 0 && rs.TodayTradeCount >= extConfig.MaxDailyTrades {
		return false, "今日交易次数已达上限"
	}

	return true, ""
}

// ResetDailyCounters 重置每日计数器
func (m *StrategyConfigManager) ResetDailyCounters() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rs := range m.robotStrategies {
		rs.TodayTradeCount = 0
	}
}

// ClearRobotStrategy 清除机器人策略缓存
func (m *StrategyConfigManager) ClearRobotStrategy(robotId int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.robotStrategies, robotId)
}

// GetOptimalLeverage 根据风险评估获取最优杠杆
func (m *StrategyConfigManager) GetOptimalLeverage(template *StrategyTemplate, riskEval *RiskEvaluation) int {
	if template == nil || riskEval == nil || template.Leverage <= 0 {
		return 5 // 默认杠杆
	}

	optimalLeverage := template.Leverage
	
	// 根据波动风险微调（高波动降低杠杆）
	if riskEval.VolatilityRisk > 70 {
		optimalLeverage = int(float64(optimalLeverage) * 0.7) // 极高波动降低30%
	} else if riskEval.VolatilityRisk > 50 {
		optimalLeverage = int(float64(optimalLeverage) * 0.85) // 高波动降低15%
	}

	// 确保至少为1
	if optimalLeverage < 1 {
		optimalLeverage = 1
	}

	return optimalLeverage
}

// GetOptimalMarginPercent 根据风险评估获取最优保证金比例
func (m *StrategyConfigManager) GetOptimalMarginPercent(template *StrategyTemplate, riskEval *RiskEvaluation) float64 {
	if template == nil || riskEval == nil || template.MarginPercent <= 0 {
		return 5.0 // 默认保证金比例
	}

	optimalMargin := template.MarginPercent
	
	// 根据账户状况微调
	if riskEval.AccountScore < 40 {
		optimalMargin *= 0.7 // 账户状况差，降低仓位
	} else if riskEval.AccountScore > 80 {
		optimalMargin *= 1.1 // 账户状况好，可适当提高
	}

	// 根据波动风险微调
	if riskEval.VolatilityRisk > 60 {
		optimalMargin *= 0.8 // 高波动降低仓位
	}

	// 确保在合理范围内（1%-50%）
	if optimalMargin < 1.0 {
		optimalMargin = 1.0
	}
	if optimalMargin > 50.0 {
		optimalMargin = 50.0
	}

	return optimalMargin
}

