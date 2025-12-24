// Package payment
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package payment

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"

	"github.com/gogf/gf/v2/frame/g"
)

// WithdrawApplyReq 申请提现请求
type WithdrawApplyReq struct {
	g.Meta `path:"/payment/withdraw/apply" method:"post" tags:"USDT提现" summary:"申请提现"`
	input.WithdrawApplyInp
}

type WithdrawApplyRes struct {
	OrderSn string `json:"orderSn" dc:"订单号"`
}

// WithdrawListReq 提现订单列表请求
type WithdrawListReq struct {
	g.Meta `path:"/payment/withdraw/list" method:"get" tags:"USDT提现" summary:"提现订单列表"`
	input.WithdrawListInp
}

type WithdrawListRes struct {
	List       []*entity.UsdtWithdraw `json:"list" dc:"列表"`
	TotalCount int                    `json:"totalCount" dc:"总数"`
	Page       int                    `json:"page" dc:"页码"`
	PageSize   int                    `json:"pageSize" dc:"每页数量"`
}

// WithdrawViewReq 查看提现订单请求
type WithdrawViewReq struct {
	g.Meta `path:"/payment/withdraw/view" method:"get" tags:"USDT提现" summary:"查看提现订单"`
	Id     int64 `json:"id" v:"required#ID不能为空" dc:"订单ID"`
}

type WithdrawViewRes struct {
	*entity.UsdtWithdraw
}

// WithdrawAuditReq 审核提现请求
type WithdrawAuditReq struct {
	g.Meta `path:"/payment/withdraw/audit" method:"post" tags:"USDT提现" summary:"审核提现"`
	input.WithdrawAuditInp
}

type WithdrawAuditRes struct{}

// WithdrawCancelReq 取消提现请求
type WithdrawCancelReq struct {
	g.Meta `path:"/payment/withdraw/cancel" method:"post" tags:"USDT提现" summary:"取消提现"`
	Id     int64 `json:"id" v:"required#ID不能为空" dc:"订单ID"`
}

type WithdrawCancelRes struct{}



