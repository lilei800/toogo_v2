// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingStrategyGroup 策略模板表
type TradingStrategyGroup struct {
	Id             int64       `json:"id"             orm:"id"               description:"主键ID"`
	GroupName      string      `json:"groupName"      orm:"group_name"       description:"模板名称"`
	GroupKey       string      `json:"groupKey"       orm:"group_key"        description:"模板标识"`
	Exchange       string      `json:"exchange"       orm:"exchange"         description:"交易平台"`
	Symbol         string      `json:"symbol"         orm:"symbol"           description:"交易对"`
	OrderType      string      `json:"orderType"      orm:"order_type"       description:"订单类型"`
	MarginMode     string      `json:"marginMode"     orm:"margin_mode"      description:"保证金模式"`
	IsOfficial     int         `json:"isOfficial"     orm:"is_official"      description:"是否官方模板"`
	FromOfficialId int64       `json:"fromOfficialId" orm:"from_official_id" description:"来源官方模板ID"`
	IsDefault      int         `json:"isDefault"      orm:"is_default"       description:"是否默认策略"`
	UserId         int64       `json:"userId"         orm:"user_id"          description:"创建用户ID"`
	Description    string      `json:"description"    orm:"description"      description:"模板描述"`
	IsActive       int         `json:"isActive"       orm:"is_active"        description:"是否启用"`
	Sort           int         `json:"sort"           orm:"sort"             description:"排序"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"       description:"创建时间"`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"       description:"更新时间"`
}

