// Package payment
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package payment

import (
	"context"
	"hotgo/api/admin/payment"
	paymentLogic "hotgo/internal/logic/payment"
)

var Withdraw = cWithdraw{}

type cWithdraw struct{}

// Apply 申请提现
func (c *cWithdraw) Apply(ctx context.Context, req *payment.WithdrawApplyReq) (res *payment.WithdrawApplyRes, err error) {
	orderSn, err := paymentLogic.Withdraw.Apply(ctx, &req.WithdrawApplyInp)
	if err != nil {
		return nil, err
	}

	res = &payment.WithdrawApplyRes{
		OrderSn: orderSn,
	}
	return
}

// List 提现订单列表
func (c *cWithdraw) List(ctx context.Context, req *payment.WithdrawListReq) (res *payment.WithdrawListRes, err error) {
	list, totalCount, err := paymentLogic.Withdraw.List(ctx, &req.WithdrawListInp)
	if err != nil {
		return nil, err
	}

	res = &payment.WithdrawListRes{
		List:       list,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	return
}

// View 查看提现订单
func (c *cWithdraw) View(ctx context.Context, req *payment.WithdrawViewReq) (res *payment.WithdrawViewRes, err error) {
	withdraw, err := paymentLogic.Withdraw.View(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	res = &payment.WithdrawViewRes{
		UsdtWithdraw: withdraw,
	}
	return
}

// Audit 审核提现
func (c *cWithdraw) Audit(ctx context.Context, req *payment.WithdrawAuditReq) (res *payment.WithdrawAuditRes, err error) {
	err = paymentLogic.Withdraw.Audit(ctx, &req.WithdrawAuditInp)
	if err != nil {
		return nil, err
	}

	res = &payment.WithdrawAuditRes{}
	return
}

// Cancel 取消提现
func (c *cWithdraw) Cancel(ctx context.Context, req *payment.WithdrawCancelReq) (res *payment.WithdrawCancelRes, err error) {
	err = paymentLogic.Withdraw.Cancel(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	res = &payment.WithdrawCancelRes{}
	return
}



