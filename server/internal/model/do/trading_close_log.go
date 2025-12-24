// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingCloseLog is the golang structure of table hg_trading_close_log for DAO operations like Where/Data.
type TradingCloseLog struct {
	g.Meta            `orm:"table:hg_trading_close_log, do:true"`
	Id                any         // 主键ID
	TenantId          any         // 租户ID
	UserId            any         // 用户ID
	RobotId           any         // 机器人ID
	OrderId           any         // 订单ID
	OrderSn           any         // 订单号
	Symbol            any         // 交易对
	Direction         any         // 方向：long/short
	OpenPrice         any         // 开仓价格
	ClosePrice        any         // 平仓价格
	Quantity          any         // 数量
	Leverage          any         // 杠杆倍数
	Margin            any         // 保证金(USDT)
	RealizedProfit    any         // 已实现盈亏
	HighestProfit     any         // 最高盈利
	ProfitPercent     any         // 盈利百分比
	CloseReason       any         // 平仓原因
	CloseDetail       any         // 平仓详情(JSON)
	OpenFee           any         // 开仓费用
	HoldFee           any         // 持仓费用
	CloseFee          any         // 平仓费用
	TotalFee          any         // 总费用
	CommissionAmount  any         // 佣金金额
	CommissionPercent any         // 佣金比例
	NetProfit         any         // 净利润
	OpenTime          *gtime.Time // 开仓时间
	CloseTime         *gtime.Time // 平仓时间
	HoldDuration      any         // 持仓时长(秒)
	CreatedAt         *gtime.Time // 创建时间
}

