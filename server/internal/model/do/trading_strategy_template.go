// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingStrategyTemplate is the golang structure of table hg_trading_strategy_template for DAO operations like Where/Data.
type TradingStrategyTemplate struct {
	g.Meta                  `orm:"table:hg_trading_strategy_template, do:true"`
	Id                      any         // 主键ID
	StrategyKey             any         // 策略KEY：conservative_trend
	StrategyName            any         // 策略名称
	RiskPreference          any         // 风险偏好：conservative/balanced/aggressive
	MarketState             any         // 市场状态：trend/volatile/high-volatility/low-volatility
	MonitorWindow           any         // 监控时间窗口(秒)
	VolatilityThreshold     any         // 波动阈值(USDT)
	LeverageMin             any         // 杠杆倍数最小值
	LeverageMax             any         // 杠杆倍数最大值
	MarginPercentMin        any         // 保证金比例最小值(%)
	MarginPercentMax        any         // 保证金比例最大值(%)
	StopLossPercent         any         // 止损百分比(%)
	ProfitRetreatPercent    any         // 止盈回撤百分比(%)
	AutoStartRetreatPercent any         // 启动回撤百分比(%)
	ConfigJson              any         // 其他配置(JSON)
	Description             any         // 策略描述
	IsActive                any         // 是否激活
	Sort                    any         // 排序
	CreatedAt               *gtime.Time // 创建时间
	UpdatedAt               *gtime.Time // 更新时间
}

