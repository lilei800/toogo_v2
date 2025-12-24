// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingProxyConfig is the golang structure for table trading_proxy_config.
type TradingProxyConfig struct {
	Id           int64       `json:"id"           orm:"id"               description:"主键ID"`
	TenantId     int64       `json:"tenantId"     orm:"tenant_id"        description:"租户ID"`
	UserId       int64       `json:"userId"       orm:"user_id"          description:"用户ID"`
	Enabled      int         `json:"enabled"      orm:"enabled"          description:"是否启用：0=禁用,1=启用"`
	ProxyType    string      `json:"proxyType"    orm:"proxy_type"       description:"代理类型：socks5/http"`
	ProxyAddress string      `json:"proxyAddress" orm:"proxy_address"    description:"代理地址"`
	AuthEnabled  int         `json:"authEnabled"  orm:"auth_enabled"     description:"是否需要认证"`
	Username     string      `json:"username"     orm:"username"         description:"用户名"`
	Password     string      `json:"password"     orm:"password"         description:"密码（加密）"`
	LastTestTime *gtime.Time `json:"lastTestTime" orm:"last_test_time"   description:"最后测试时间"`
	TestStatus   int         `json:"testStatus"   orm:"test_status"      description:"测试状态：0=未测试,1=成功,2=失败"`
	TestMessage  string      `json:"testMessage"  orm:"test_message"     description:"测试消息"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"       description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"       description:"更新时间"`
}

