// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingRobot is the golang structure of table hg_trading_robot for DAO operations like Where/Data.
type TradingRobot struct {
	g.Meta                  `orm:"table:hg_trading_robot, do:true"`
	Id                      any         // 主键ID
	TenantId                any         // 租户ID
	UserId                  any         // 用户ID
	RobotName               any         // 机器人名称
	ApiConfigId             any         // API接口ID
	MaxProfitTarget         any         // 最大盈利目标(USDT)
	MaxLossAmount           any         // 最大亏损额(USDT)
	MaxRuntime              any         // 最大运行时长(秒)
	RiskPreference          any         // 风险偏好：conservative/balanced/aggressive
	AutoRiskPreference      any         // 自动风险偏好：0=手动,1=自动
	MarketState             any         // 市场状态：trend/volatile/high-volatility/low-volatility
	AutoMarketState         any         // 自动市场状态：0=手动,1=自动
	Exchange                any         // 交易所
	Symbol                  any         // 交易对
	OrderType               any         // 订单类型：market/limit
	MarginMode              any         // 保证金模式：isolated/cross
	Leverage                any         // 杠杆倍数
	MarginPercent           any         // 使用保证金比例(%)
	UseMonitorSignal        any         // 采用方向预警信号：0=否,1=是
	StopLossPercent         any         // 止损百分比(%)
	ProfitRetreatPercent    any         // 止盈回撤百分比(%)
	AutoStartRetreatPercent any         // 启动回撤百分比(%)
	CurrentStrategy         any         // 当前策略配置(JSON)
	Status                  any         // 状态：1=未启动,2=运行中,3=暂停,4=停用
	StartTime               *gtime.Time // 启动时间
	PauseTime               *gtime.Time // 暂停时间
	StopTime                *gtime.Time // 停止时间
	LongCount               any         // 多单数
	ShortCount              any         // 空单数
	TotalProfit             any         // 总盈亏(USDT)
	RuntimeSeconds          any         // 已运行时长(秒)
	AutoTradeEnabled        any         // 全自动下单：0=否,1=是
	AutoCloseEnabled        any         // 全自动平仓：0=否,1=是
	ProfitLockEnabled       any         // 锁定盈利开关：0=关闭,1=开启（止盈启动后禁止自动开新仓）
	Remark                  any         // 备注
	CreatedAt               *gtime.Time // 创建时间
	UpdatedAt               *gtime.Time // 更新时间
	DeletedAt               *gtime.Time // 删除时间
}

