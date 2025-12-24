// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingMonitorLog is the golang structure of table hg_trading_monitor_log for DAO operations like Where/Data.
type TradingMonitorLog struct {
	g.Meta         `orm:"table:hg_trading_monitor_log, do:true"`
	Id             any         // 主键ID
	TenantId       any         // 租户ID
	UserId         any         // 用户ID
	RobotId        any         // 机器人ID
	Symbol         any         // 交易对
	CurrentPrice   any         // 当前价格
	WindowHigh     any         // 窗口最高价
	WindowLow      any         // 窗口最低价
	Volatility     any         // 波动值
	SignalType     any         // 信号类型：buy/sell/hold
	SignalStrength any         // 信号强度(0-100)
	MarketState    any         // 市场状态
	SignalDetail   any         // 信号详情(JSON)
	CreatedAt      *gtime.Time // 创建时间
}

