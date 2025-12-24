// Package admin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 支付回调接口
package admin

import (
	"github.com/gogf/gf/v2/frame/g"
)

// NOWPaymentsCallbackReq NOWPayments充值回调请求
type NOWPaymentsCallbackReq struct {
	g.Meta `path:"/payment/nowpayments/callback" method:"post" tags:"支付回调" summary:"NOWPayments充值回调"`
}

type NOWPaymentsCallbackRes struct {
	Status string `json:"status"`
}

// NOWPaymentsPayoutCallbackReq NOWPayments提现回调请求
type NOWPaymentsPayoutCallbackReq struct {
	g.Meta `path:"/payment/nowpayments/payout-callback" method:"post" tags:"支付回调" summary:"NOWPayments提现回调"`
}

type NOWPaymentsPayoutCallbackRes struct {
	Status string `json:"status"`
}

