// =================================================================================
// 机器人运行区间记录实体
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingRobotRunSession 机器人运行区间记录
type TradingRobotRunSession struct {
	Id             int64       `json:"id"             orm:"id"              description:"主键ID"`
	RobotId        int64       `json:"robotId"        orm:"robot_id"        description:"机器人ID"`
	UserId         int64       `json:"userId"         orm:"user_id"         description:"用户ID"`
	Exchange       string      `json:"exchange"       orm:"exchange"        description:"交易所"`
	Symbol         string      `json:"symbol"         orm:"symbol"          description:"交易对"`
	StartTime      *gtime.Time `json:"startTime"      orm:"start_time"      description:"启动时间"`
	EndTime        *gtime.Time `json:"endTime"        orm:"end_time"        description:"结束时间"`
	EndReason      string      `json:"endReason"      orm:"end_reason"      description:"结束原因"`
	RuntimeSeconds int         `json:"runtimeSeconds" orm:"runtime_seconds" description:"运行时长(秒)"`
	TotalPnl       *float64    `json:"totalPnl"       orm:"total_pnl"       description:"区间总盈亏(USDT)"`
	TotalFee       *float64    `json:"totalFee"       orm:"total_fee"       description:"区间总手续费(USDT)"`
	TradeCount     int         `json:"tradeCount"     orm:"trade_count"     description:"区间成交笔数"`
	SyncedAt       *gtime.Time `json:"syncedAt"       orm:"synced_at"       description:"最后同步时间"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"创建时间"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"更新时间"`
}
