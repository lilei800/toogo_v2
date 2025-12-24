// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingTradeFill is the golang structure of table hg_trading_trade_fill for DAO operations like Where/Data.
type TradingTradeFill struct {
	g.Meta        `orm:"table:hg_trading_trade_fill, do:true"`
	Id            any         // 主键ID
	TenantId      any         // 租户ID
	ApiConfigId   any         // API配置ID
	Exchange      any         // 交易所
	UserId        any         // 用户ID
	RobotId       any         // 机器人ID
	SessionId     any         // 运行区间ID(可选)
	Symbol        any         // 交易对
	OrderId       any         // 交易所订单ID
	ClientOrderId any         // 客户端订单ID(可选)
	TradeId       any         // 成交ID
	Side          any         // 方向
	Qty           any         // 成交数量
	Price         any         // 成交价格
	Fee           any         // 手续费
	FeeCoin       any         // 手续费币种
	RealizedPnl   any         // 已实现盈亏
	Ts            any         // 成交时间戳(毫秒)
	CreatedAt     *gtime.Time // 创建时间
	UpdatedAt     *gtime.Time // 更新时间
}
