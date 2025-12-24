// Package admin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 支付回调控制器
package admin

import (
	"context"
	"hotgo/api/admin"
	"hotgo/internal/service"
)

var PaymentCallback = cPaymentCallback{}

type cPaymentCallback struct{}

// NOWPaymentsCallback NOWPayments充值回调
func (c *cPaymentCallback) NOWPaymentsCallback(ctx context.Context, req *admin.NOWPaymentsCallbackReq) (res *admin.NOWPaymentsCallbackRes, err error) {
	err = service.ToogoFinance().HandleNOWPaymentsIPNCallback(ctx)
	if err != nil {
		return &admin.NOWPaymentsCallbackRes{Status: "error"}, nil
	}
	return &admin.NOWPaymentsCallbackRes{Status: "ok"}, nil
}

// NOWPaymentsPayoutCallback NOWPayments提现回调
func (c *cPaymentCallback) NOWPaymentsPayoutCallback(ctx context.Context, req *admin.NOWPaymentsPayoutCallbackReq) (res *admin.NOWPaymentsPayoutCallbackRes, err error) {
	err = service.ToogoFinance().HandleNOWPaymentsPayoutIPNCallback(ctx)
	if err != nil {
		return &admin.NOWPaymentsPayoutCallbackRes{Status: "error"}, nil
	}
	return &admin.NOWPaymentsPayoutCallbackRes{Status: "ok"}, nil
}

