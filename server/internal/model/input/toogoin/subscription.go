// Package toogoin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogoin

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// PlanListInp 套餐列表输入
type PlanListInp struct {
	form.PageReq
	Status int `json:"status" description:"状态"`
}

// PlanListModel 套餐列表返回
type PlanListModel struct {
	*entity.ToogoPlan
}

// PlanEditInp 编辑套餐输入
type PlanEditInp struct {
	Id                 int64   `json:"id" description:"套餐ID"`
	PlanName           string  `json:"planName" v:"required" description:"套餐名称"`
	PlanCode           string  `json:"planCode" v:"required" description:"套餐代码"`
	RobotLimit         int     `json:"robotLimit" v:"required|min:1" description:"机器人数量"`
	PriceDaily         float64 `json:"priceDaily" description:"日价格"`
	PriceMonthly       float64 `json:"priceMonthly" description:"月价格"`
	PriceQuarterly     float64 `json:"priceQuarterly" description:"季价格"`
	PriceHalfYear      float64 `json:"priceHalfYear" description:"半年价格"`
	PriceYearly        float64 `json:"priceYearly" description:"年价格"`
	DefaultPeriod      string  `json:"defaultPeriod" description:"默认价格方案: daily/monthly/quarterly/half_year/yearly"`
	PurchaseLimit      int     `json:"purchaseLimit" description:"购买次数限制(0为不限)"`
	PurchaseLimitDaily int     `json:"purchaseLimitDaily" description:"日付购买次数限制，0为不限"`
	PurchaseLimitMonthly int    `json:"purchaseLimitMonthly" description:"月付购买次数限制，0为不限"`
	PurchaseLimitQuarterly int  `json:"purchaseLimitQuarterly" description:"季付购买次数限制，0为不限"`
	PurchaseLimitHalfYear int   `json:"purchaseLimitHalfYear" description:"半年付购买次数限制，0为不限"`
	PurchaseLimitYearly int     `json:"purchaseLimitYearly" description:"年付购买次数限制，0为不限"`
	GiftPowerMonthly   float64 `json:"giftPowerMonthly" description:"月赠送积分"`
	GiftPowerQuarterly float64 `json:"giftPowerQuarterly" description:"季赠送积分"`
	GiftPowerHalfYear  float64 `json:"giftPowerHalfYear" description:"半年赠送积分"`
	GiftPowerYearly    float64 `json:"giftPowerYearly" description:"年赠送积分"`
	Description        string  `json:"description" description:"套餐描述"`
	Features           string  `json:"features" description:"套餐特性"`
	IsDefault          int     `json:"isDefault" description:"是否默认"`
	Sort               int     `json:"sort" description:"排序"`
	Status             int     `json:"status" description:"状态"`
}

// PlanDeleteInp 删除套餐输入
type PlanDeleteInp struct {
	Id int64 `json:"id" v:"required" description:"套餐ID"`
}

// SubscribeInp 订阅套餐输入
type SubscribeInp struct {
	UserId     int64  `json:"userId" description:"用户ID（不传则从上下文获取）"`
	PlanId     int64  `json:"planId" v:"required" description:"套餐ID"`
	PeriodType string `json:"periodType" v:"required|in:daily,monthly,quarterly,half_year,yearly" description:"订阅周期"`
	PayType    string `json:"payType" v:"required|in:balance,crypto" description:"支付方式"`
	UsePoints  bool   `json:"usePoints" description:"是否使用积分抵扣"`
}

// SubscribeModel 订阅套餐返回
type SubscribeModel struct {
	OrderSn      string  `json:"orderSn" description:"订单号"`
	PlanName     string  `json:"planName" description:"套餐名称"`
	Amount       float64 `json:"amount" description:"订阅原价"`
	PointsDeduct float64 `json:"pointsDeduct" description:"积分抵扣金额"`
	BalancePaid  float64 `json:"balancePaid" description:"余额支付金额"`
	Days         int     `json:"days" description:"订阅天数"`
	ExpireTime   string  `json:"expireTime" description:"到期时间"`
}

// SubscriptionListInp 订阅记录列表输入
type SubscriptionListInp struct {
	form.PageReq
	UserId    int64    `json:"userId" description:"用户ID"`
	Status    int      `json:"status" description:"状态"`
	CreatedAt []string `json:"createdAt" description:"创建时间"`
}

// SubscriptionListModel 订阅记录列表返回
type SubscriptionListModel struct {
	*entity.ToogoSubscription
	PlanName string `json:"planName" description:"套餐名称"`
}

// MySubscriptionInp 我的订阅输入
type MySubscriptionInp struct {
	UserId int64 `json:"userId" description:"用户ID"`
}

// MySubscriptionModel 我的订阅返回
type MySubscriptionModel struct {
	HasSubscription bool    `json:"hasSubscription" description:"是否有有效订阅"`
	PlanId          int64   `json:"planId" description:"套餐ID"`
	PlanName        string  `json:"planName" description:"套餐名称"`
	PlanCode        string  `json:"planCode" description:"套餐代码"`
	RobotLimit      int     `json:"robotLimit" description:"机器人限制"`
	ActiveRobots    int     `json:"activeRobots" description:"运行中机器人"`
	ExpireTime      string  `json:"expireTime" description:"到期时间"`
	RemainingDays   int     `json:"remainingDays" description:"剩余天数"`
}

