// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoPlan is the golang structure for table hg_toogo_plan.
type ToogoPlan struct {
	Id                int64       `json:"id"                orm:"id"                  description:"主键ID"`
	PlanName          string      `json:"planName"          orm:"plan_name"           description:"套餐名称"`
	PlanCode          string      `json:"planCode"          orm:"plan_code"           description:"套餐代码: FREE/A/B/C/D"`
	RobotLimit        int         `json:"robotLimit"        orm:"robot_limit"         description:"支持机器人数量"`
	PriceDaily        float64     `json:"priceDaily"        orm:"price_daily"         description:"日价格"`
	PriceMonthly      float64     `json:"priceMonthly"      orm:"price_monthly"       description:"月价格"`
	PriceQuarterly    float64     `json:"priceQuarterly"    orm:"price_quarterly"     description:"季价格"`
	PriceHalfYear     float64     `json:"priceHalfYear"     orm:"price_half_year"     description:"半年价格"`
	PriceYearly       float64     `json:"priceYearly"       orm:"price_yearly"        description:"年价格"`
	DefaultPeriod     string      `json:"defaultPeriod"     orm:"default_period"      description:"默认价格方案: daily/monthly/quarterly/half_year/yearly"`
	PurchaseLimit     int         `json:"purchaseLimit"     orm:"purchase_limit"      description:"购买次数限制(0为不限)"`
	PurchaseLimitDaily int        `json:"purchaseLimitDaily" orm:"purchase_limit_daily" description:"日付购买次数限制，0为不限"`
	PurchaseLimitMonthly int      `json:"purchaseLimitMonthly" orm:"purchase_limit_monthly" description:"月付购买次数限制，0为不限"`
	PurchaseLimitQuarterly int     `json:"purchaseLimitQuarterly" orm:"purchase_limit_quarterly" description:"季付购买次数限制，0为不限"`
	PurchaseLimitHalfYear int     `json:"purchaseLimitHalfYear" orm:"purchase_limit_half_year" description:"半年付购买次数限制，0为不限"`
	PurchaseLimitYearly int       `json:"purchaseLimitYearly" orm:"purchase_limit_yearly" description:"年付购买次数限制，0为不限"`
	GiftPowerMonthly  float64     `json:"giftPowerMonthly"  orm:"gift_power_monthly"  description:"月订阅赠送积分"`
	GiftPowerQuarterly float64    `json:"giftPowerQuarterly" orm:"gift_power_quarterly" description:"季订阅赠送积分"`
	GiftPowerHalfYear float64     `json:"giftPowerHalfYear" orm:"gift_power_half_year" description:"半年订阅赠送积分"`
	GiftPowerYearly   float64     `json:"giftPowerYearly"   orm:"gift_power_yearly"   description:"年订阅赠送积分"`
	Description       string      `json:"description"       orm:"description"         description:"套餐描述"`
	Features          string      `json:"features"          orm:"features"            description:"套餐特性(JSON)"`
	IsDefault         int         `json:"isDefault"         orm:"is_default"          description:"是否默认套餐(免费)"`
	Sort              int         `json:"sort"              orm:"sort"                description:"排序"`
	Status            int         `json:"status"            orm:"status"              description:"状态: 1=上架, 2=下架"`
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"          description:"创建时间"`
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"          description:"更新时间"`
}

