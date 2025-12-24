// Package websocket Toogo消息推送服务
package websocket

import (
	"time"
)

// ToogoPusher Toogo专用消息推送器
type ToogoPusher struct {
	hub *Hub
}

// NewToogoPusher 创建推送器
func NewToogoPusher() *ToogoPusher {
	return &ToogoPusher{
		hub: GetHub(),
	}
}

// TickerData 行情数据
type TickerData struct {
	Symbol    string  `json:"symbol"`     // 交易对
	Price     float64 `json:"price"`      // 最新价
	High24h   float64 `json:"high_24h"`   // 24h最高
	Low24h    float64 `json:"low_24h"`    // 24h最低
	Volume24h float64 `json:"volume_24h"` // 24h成交量
	Change24h float64 `json:"change_24h"` // 24h涨跌幅
}

// PositionData 持仓数据
type PositionData struct {
	RobotID      int64   `json:"robot_id"`       // 机器人ID
	Symbol       string  `json:"symbol"`         // 交易对
	Side         string  `json:"side"`           // 方向: long/short
	Size         float64 `json:"size"`           // 数量
	EntryPrice   float64 `json:"entry_price"`    // 开仓价
	MarkPrice    float64 `json:"mark_price"`     // 标记价
	UnrealizedPnL float64 `json:"unrealized_pnl"` // 未实现盈亏
	Leverage     int     `json:"leverage"`       // 杠杆
	Margin       float64 `json:"margin"`         // 保证金
	LiquidPrice  float64 `json:"liquid_price"`   // 强平价
}

// OrderData 订单数据
type OrderData struct {
	RobotID   int64   `json:"robot_id"`    // 机器人ID
	OrderID   string  `json:"order_id"`    // 订单ID
	Symbol    string  `json:"symbol"`      // 交易对
	Side      string  `json:"side"`        // 方向
	Type      string  `json:"type"`        // 类型
	Status    string  `json:"status"`      // 状态
	Price     float64 `json:"price"`       // 价格
	Size      float64 `json:"size"`        // 数量
	Filled    float64 `json:"filled"`      // 已成交
	PnL       float64 `json:"pnl"`         // 盈亏
	CreatedAt string  `json:"created_at"`  // 创建时间
}

// RobotStatusData 机器人状态数据
type RobotStatusData struct {
	RobotID       int64   `json:"robot_id"`        // 机器人ID
	Name          string  `json:"name"`            // 名称
	Status        int     `json:"status"`          // 状态
	TotalProfit   float64 `json:"total_profit"`    // 总盈亏
	ConsumedPower float64 `json:"consumed_power"`  // 已消耗算力
	OpenOrders    int     `json:"open_orders"`     // 持仓订单数
	MarketState   string  `json:"market_state"`    // 市场状态
	RiskLevel     string  `json:"risk_level"`      // 风险偏好
	Signal        string  `json:"signal"`          // 当前信号
}

// SignalData 交易信号数据
type SignalData struct {
	RobotID     int64   `json:"robot_id"`     // 机器人ID
	Symbol      string  `json:"symbol"`       // 交易对
	Signal      string  `json:"signal"`       // 信号: buy/sell/hold
	Confidence  float64 `json:"confidence"`   // 置信度 0-100
	MarketState string  `json:"market_state"` // 市场状态
	Reason      string  `json:"reason"`       // 原因描述
}

// PushTicker 推送行情数据
func (p *ToogoPusher) PushTicker(symbol string, data *TickerData) {
	channel := "ticker:" + symbol
	p.hub.SendToChannel(channel, &Message{
		Type:      MsgTypeTicker,
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushPosition 推送持仓更新(给指定用户)
func (p *ToogoPusher) PushPosition(userID int64, data *PositionData) {
	p.hub.SendToUser(userID, &Message{
		Type:      MsgTypePosition,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushPositions 批量推送持仓(给指定用户)
func (p *ToogoPusher) PushPositions(userID int64, positions []*PositionData) {
	p.hub.SendToUser(userID, &Message{
		Type:      MsgTypePosition,
		Data:      positions,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushOrder 推送订单更新(给指定用户)
func (p *ToogoPusher) PushOrder(userID int64, data *OrderData) {
	p.hub.SendToUser(userID, &Message{
		Type:      MsgTypeOrder,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushRobotStatus 推送机器人状态(给指定用户)
func (p *ToogoPusher) PushRobotStatus(userID int64, data *RobotStatusData) {
	p.hub.SendToUser(userID, &Message{
		Type:      MsgTypeRobot,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushSignal 推送交易信号(给指定用户)
func (p *ToogoPusher) PushSignal(userID int64, data *SignalData) {
	p.hub.SendToUser(userID, &Message{
		Type:      MsgTypeSignal,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushPnL 推送盈亏更新(给指定用户)
func (p *ToogoPusher) PushPnL(userID int64, robotID int64, pnl float64, consumedPower float64) {
	p.hub.SendToUser(userID, &Message{
		Type: MsgTypePnL,
		Data: map[string]interface{}{
			"robot_id":       robotID,
			"pnl":            pnl,
			"consumed_power": consumedPower,
		},
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushSystemNotice 推送系统通知(广播)
func (p *ToogoPusher) PushSystemNotice(title, content string) {
	p.hub.Broadcast(&Message{
		Type: MsgTypeSystem,
		Data: map[string]interface{}{
			"title":   title,
			"content": content,
		},
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushError 推送错误消息(给指定用户)
func (p *ToogoPusher) PushError(userID int64, code int, message string) {
	p.hub.SendToUser(userID, &Message{
		Type: MsgTypeError,
		Data: map[string]interface{}{
			"code":    code,
			"message": message,
		},
		Timestamp: time.Now().UnixMilli(),
	})
}

// GetOnlineCount 获取在线人数
func (p *ToogoPusher) GetOnlineCount() int {
	return p.hub.GetOnlineCount()
}

// IsUserOnline 检查用户是否在线
func (p *ToogoPusher) IsUserOnline(userID int64) bool {
	return p.hub.GetUserOnline(userID)
}

// 全局推送器实例
var defaultPusher *ToogoPusher

// GetPusher 获取推送器单例
func GetPusher() *ToogoPusher {
	if defaultPusher == nil {
		defaultPusher = NewToogoPusher()
	}
	return defaultPusher
}

