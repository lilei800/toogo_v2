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

var Balance = cBalance{}

type cBalance struct{}

// View 查看余额
func (c *cBalance) View(ctx context.Context, req *payment.BalanceViewReq) (res *payment.BalanceViewRes, err error) {
	balance, err := paymentLogic.Balance.View(ctx)
	if err != nil {
		return nil, err
	}

	res = &payment.BalanceViewRes{
		UsdtBalance: balance,
	}
	return
}

// Logs 资金流水列表
func (c *cBalance) Logs(ctx context.Context, req *payment.BalanceLogListReq) (res *payment.BalanceLogListRes, err error) {
	list, totalCount, err := paymentLogic.Balance.LogList(ctx, &req.BalanceLogListInp)
	if err != nil {
		return nil, err
	}

	res = &payment.BalanceLogListRes{
		List:       list,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	return
}



