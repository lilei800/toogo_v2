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

var Deposit = cDeposit{}

type cDeposit struct{}

// Create 创建充值订单
func (c *cDeposit) Create(ctx context.Context, req *payment.DepositCreateReq) (res *payment.DepositCreateRes, err error) {
	orderSn, err := paymentLogic.Deposit.Create(ctx, &req.DepositCreateInp)
	if err != nil {
		return nil, err
	}

	res = &payment.DepositCreateRes{
		OrderSn: orderSn,
	}
	return
}

// List 充值订单列表
func (c *cDeposit) List(ctx context.Context, req *payment.DepositListReq) (res *payment.DepositListRes, err error) {
	list, totalCount, err := paymentLogic.Deposit.List(ctx, &req.DepositListInp)
	if err != nil {
		return nil, err
	}

	res = &payment.DepositListRes{
		List:       list,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	return
}

// View 查看充值订单
func (c *cDeposit) View(ctx context.Context, req *payment.DepositViewReq) (res *payment.DepositViewRes, err error) {
	deposit, err := paymentLogic.Deposit.View(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	res = &payment.DepositViewRes{
		UsdtDeposit: deposit,
	}
	return
}

// Check 检查充值状态
func (c *cDeposit) Check(ctx context.Context, req *payment.DepositCheckReq) (res *payment.DepositCheckRes, err error) {
	err = paymentLogic.Deposit.Check(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	res = &payment.DepositCheckRes{}
	return
}

// Cancel 取消充值订单
func (c *cDeposit) Cancel(ctx context.Context, req *payment.DepositCancelReq) (res *payment.DepositCancelRes, err error) {
	err = paymentLogic.Deposit.Cancel(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	res = &payment.DepositCancelRes{}
	return
}

