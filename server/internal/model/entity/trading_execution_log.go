// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingExecutionLog is the golang structure for table trading_execution_log.
type TradingExecutionLog struct {
	Id          int64       `json:"id"          orm:"id"           description:"主键ID"`
	SignalLogId int64       `json:"signalLogId" orm:"signal_log_id" description:"关联的预警日志ID（可选）"`
	RobotId     int64       `json:"robotId"     orm:"robot_id"      description:"机器人ID"`
	OrderId     int64       `json:"orderId"     orm:"order_id"      description:"关联的订单ID（可选）"`
	EventType   string      `json:"eventType"   orm:"event_type"    description:"事件类型"`
	EventData   string      `json:"eventData"   orm:"event_data"    description:"事件数据（JSON格式）"`
	Status      string      `json:"status"      orm:"status"        description:"状态：pending/success/failed"`
	Message     string      `json:"message"     orm:"message"       description:"消息（详细说明）"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"    description:"创建时间"`
}

