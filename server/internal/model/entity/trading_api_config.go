// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingApiConfig is the golang structure for table trading_api_config.
type TradingApiConfig struct {
	Id             int64       `json:"id"             orm:"id"                 description:"主键ID"`
	TenantId       int64       `json:"tenantId"       orm:"tenant_id"          description:"租户ID"`
	UserId         int64       `json:"userId"         orm:"user_id"            description:"用户ID"`
	ApiName        string      `json:"apiName"        orm:"api_name"           description:"API接口名称"`
	Platform       string      `json:"platform"       orm:"platform"           description:"平台名称：bitget/binance/okx"`
	BaseUrl        string      `json:"baseUrl"        orm:"base_url"           description:"API地址"`
	ApiKey         string      `json:"apiKey"         orm:"api_key"            description:"API Key（加密）"`
	SecretKey      string      `json:"secretKey"      orm:"secret_key"         description:"Secret Key（加密）"`
	Passphrase     string      `json:"passphrase"     orm:"passphrase"         description:"Passphrase（加密，可选）"`
	IsDefault      int         `json:"isDefault"      orm:"is_default"         description:"是否默认：0=否,1=是"`
	Status         int         `json:"status"         orm:"status"             description:"状态：1=正常,2=禁用"`
	LastVerifyTime *gtime.Time `json:"lastVerifyTime" orm:"last_verify_time"   description:"最后验证时间"`
	VerifyStatus   int         `json:"verifyStatus"   orm:"verify_status"      description:"验证状态：0=未验证,1=成功,2=失败"`
	VerifyMessage  string      `json:"verifyMessage"  orm:"verify_message"     description:"验证消息"`
	Remark         string      `json:"remark"         orm:"remark"             description:"备注"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"         description:"创建时间"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"         description:"更新时间"`
	DeletedAt      *gtime.Time `json:"deletedAt"      orm:"deleted_at"         description:"删除时间"`
}

