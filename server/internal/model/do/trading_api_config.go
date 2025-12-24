// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingApiConfig is the golang structure of table hg_trading_api_config for DAO operations like Where/Data.
type TradingApiConfig struct {
	g.Meta         `orm:"table:hg_trading_api_config, do:true"`
	Id             any         // 主键ID
	TenantId       any         // 租户ID
	UserId         any         // 用户ID
	ApiName        any         // API接口名称
	Platform       any         // 平台名称：bitget/binance/okx
	BaseUrl        any         // API地址
	ApiKey         any         // API Key（加密）
	SecretKey      any         // Secret Key（加密）
	Passphrase     any         // Passphrase（加密，可选）
	IsDefault      any         // 是否默认：0=否,1=是
	Status         any         // 状态：1=正常,2=禁用
	LastVerifyTime *gtime.Time // 最后验证时间
	VerifyStatus   any         // 验证状态：0=未验证,1=成功,2=失败
	VerifyMessage  any         // 验证消息
	Remark         any         // 备注
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 更新时间
	DeletedAt      *gtime.Time // 删除时间
}

