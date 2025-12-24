// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingOrder is the golang structure of table hg_trading_order for DAO operations like Where/Data.
type TradingOrder struct {
	g.Meta               `orm:"table:hg_trading_order, do:true"`
	Id                   any         // 主键ID
	TenantId             any         // 租户ID
	UserId               any         // 用户ID
	RobotId              any         // 机器人ID
	OrderSn              any         // 订单号
	ExchangeOrderId      any         // 交易所订单ID
	Symbol               any         // 交易对
	Direction            any         // 方向：long/short
	OpenPrice            any         // 开仓价格
	ClosePrice           any         // 平仓价格
	Quantity             any         // 数量
	Leverage             any         // 杠杆倍数
	Margin               any         // 保证金(USDT)
	RealizedProfit       any         // 已实现盈亏
	UnrealizedProfit     any         // 未实现盈亏
	HighestProfit        any         // 最高盈利
	StopLossPrice        any         // 止损价格
	ProfitRetreatStarted any         // 止盈回撤已启动
	ProfitRetreatPercent any         // 止盈回撤百分比
	OpenTime             *gtime.Time // 开仓时间
	CloseTime            *gtime.Time // 平仓时间
	HoldDuration         any         // 持仓时长(秒)
	Status               any         // 状态：1=持仓中,2=已平仓,3=已取消
	CloseReason          any         // 平仓原因：stop_loss/take_profit/manual/timeout
	Remark               any         // 备注
	CreatedAt            *gtime.Time // 创建时间
	UpdatedAt            *gtime.Time // 更新时间
}

