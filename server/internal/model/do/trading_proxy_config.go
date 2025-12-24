// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingProxyConfig is the golang structure of table hg_trading_proxy_config for DAO operations like Where/Data.
type TradingProxyConfig struct {
	g.Meta       `orm:"table:hg_trading_proxy_config, do:true"`
	Id           any         // 主键ID
	TenantId     any         // 租户ID
	UserId       any         // 用户ID
	Enabled      any         // 是否启用：0=禁用,1=启用
	ProxyType    any         // 代理类型：socks5/http
	ProxyAddress any         // 代理地址
	AuthEnabled  any         // 是否需要认证
	Username     any         // 用户名
	Password     any         // 密码（加密）
	LastTestTime *gtime.Time // 最后测试时间
	TestStatus   any         // 测试状态：0=未测试,1=成功,2=失败
	TestMessage  any         // 测试消息
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
}

