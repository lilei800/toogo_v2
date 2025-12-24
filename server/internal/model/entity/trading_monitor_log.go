// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingMonitorLog is the golang structure for table trading_monitor_log.
type TradingMonitorLog struct {
	Id             int64       `json:"id"             orm:"id"                description:"主键ID"`
	TenantId       int64       `json:"tenantId"       orm:"tenant_id"         description:"租户ID"`
	UserId         int64       `json:"userId"         orm:"user_id"           description:"用户ID"`
	RobotId        int64       `json:"robotId"        orm:"robot_id"          description:"机器人ID"`
	Symbol         string      `json:"symbol"         orm:"symbol"            description:"交易对"`
	CurrentPrice   float64     `json:"currentPrice"   orm:"current_price"     description:"当前价格"`
	WindowHigh     float64     `json:"windowHigh"     orm:"window_high"       description:"窗口最高价"`
	WindowLow      float64     `json:"windowLow"      orm:"window_low"        description:"窗口最低价"`
	Volatility     float64     `json:"volatility"     orm:"volatility"        description:"波动值"`
	SignalType     string      `json:"signalType"     orm:"signal_type"       description:"信号类型：buy/sell/hold"`
	SignalStrength float64     `json:"signalStrength" orm:"signal_strength"   description:"信号强度(0-100)"`
	MarketState    string      `json:"marketState"    orm:"market_state"      description:"市场状态"`
	SignalDetail   string      `json:"signalDetail"   orm:"signal_detail"     description:"信号详情(JSON)"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"        description:"创建时间"`
}

