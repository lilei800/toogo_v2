// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingRobot is the golang structure for table trading_robot.
type TradingRobot struct {
	Id                      int64       `json:"id"                       orm:"id"                          description:"主键ID"`
	TenantId                int64       `json:"tenantId"                 orm:"tenant_id"                   description:"租户ID"`
	UserId                  int64       `json:"userId"                   orm:"user_id"                     description:"用户ID"`
	RobotName               string      `json:"robotName"                orm:"robot_name"                  description:"机器人名称"`
	ApiConfigId             int64       `json:"apiConfigId"              orm:"api_config_id"               description:"API接口ID"`
	MaxProfitTarget         float64     `json:"maxProfitTarget"          orm:"max_profit_target"           description:"最大盈利目标(USDT)"`
	MaxLossAmount           float64     `json:"maxLossAmount"            orm:"max_loss_amount"             description:"最大亏损额(USDT)"`
	MaxRuntime              int         `json:"maxRuntime"               orm:"max_runtime"                 description:"最大运行时长(秒)"`
	RiskPreference          string      `json:"riskPreference"           orm:"risk_preference"             description:"风险偏好：conservative/balanced/aggressive"`
	AutoRiskPreference      int         `json:"autoRiskPreference"       orm:"auto_risk_preference"        description:"自动风险偏好：0=手动,1=自动"`
	MarketState             string      `json:"marketState"              orm:"market_state"                description:"市场状态：trend/volatile/high-volatility/low-volatility"`
	AutoMarketState         int         `json:"autoMarketState"          orm:"auto_market_state"           description:"自动市场状态：0=手动,1=自动"`
	Exchange                string      `json:"exchange"                 orm:"exchange"                    description:"交易所"`
	Symbol                  string      `json:"symbol"                   orm:"symbol"                      description:"交易对"`
	OrderType               string      `json:"orderType"                orm:"order_type"                  description:"订单类型：market/limit"`
	MarginMode              string      `json:"marginMode"               orm:"margin_mode"                 description:"保证金模式：isolated/cross"`
	Leverage                int         `json:"leverage"                 orm:"leverage"                    description:"杠杆倍数"`
	MarginPercent           float64     `json:"marginPercent"            orm:"margin_percent"              description:"使用保证金比例(%)"`
	UseMonitorSignal        int         `json:"useMonitorSignal"         orm:"use_monitor_signal"          description:"采用方向预警信号：0=否,1=是"`
	StopLossPercent         float64     `json:"stopLossPercent"          orm:"stop_loss_percent"           description:"止损百分比(%)"`
	ProfitRetreatPercent    float64     `json:"profitRetreatPercent"     orm:"profit_retreat_percent"      description:"止盈回撤百分比(%)"`
	AutoStartRetreatPercent float64     `json:"autoStartRetreatPercent"  orm:"auto_start_retreat_percent"  description:"启动回撤百分比(%)"`
	CurrentStrategy         string      `json:"currentStrategy"          orm:"current_strategy"            description:"当前策略配置(JSON)"`
	StrategyGroupId         int64       `json:"strategyGroupId"          orm:"strategy_group_id"           description:"策略组ID"`
	Status                  int         `json:"status"                   orm:"status"                      description:"状态：1=未启动,2=运行中,3=暂停,4=停用"`
	StartTime               *gtime.Time `json:"startTime"                orm:"start_time"                  description:"启动时间"`
	PauseTime               *gtime.Time `json:"pauseTime"                orm:"pause_time"                  description:"暂停时间"`
	StopTime                *gtime.Time `json:"stopTime"                 orm:"stop_time"                   description:"停止时间"`
	LongCount               int         `json:"longCount"                orm:"long_count"                  description:"多单数"`
	ShortCount              int         `json:"shortCount"               orm:"short_count"                 description:"空单数"`
	TotalProfit             float64     `json:"totalProfit"              orm:"total_profit"                description:"总盈亏(USDT)"`
	RuntimeSeconds          int         `json:"runtimeSeconds"           orm:"runtime_seconds"             description:"已运行时长(秒)"`
	AutoTradeEnabled        int         `json:"autoTradeEnabled"         orm:"auto_trade_enabled"          description:"全自动下单：0=否,1=是"`
	AutoCloseEnabled        int         `json:"autoCloseEnabled"         orm:"auto_close_enabled"          description:"全自动平仓：0=否,1=是"`
	ProfitLockEnabled       int         `json:"profitLockEnabled"        orm:"profit_lock_enabled"         description:"锁定盈利开关：0=关闭,1=开启（止盈启动后禁止自动开新仓）"`
	DualSidePosition        int         `json:"dualSidePosition"         orm:"dual_side_position"          description:"双向开单：0=单向,1=双向"`
	ScheduleStart           *gtime.Time `json:"scheduleStart"            orm:"schedule_start"              description:"定时启动时间"`
	ScheduleStop            *gtime.Time `json:"scheduleStop"             orm:"schedule_stop"               description:"定时停止时间"`
	Remark                  string      `json:"remark"                   orm:"remark"                      description:"备注"`
	CreatedAt               *gtime.Time `json:"createdAt"                orm:"created_at"                  description:"创建时间"`
	UpdatedAt               *gtime.Time `json:"updatedAt"                orm:"updated_at"                  description:"更新时间"`
	DeletedAt               *gtime.Time `json:"deletedAt"                orm:"deleted_at"                  description:"删除时间"`
}
