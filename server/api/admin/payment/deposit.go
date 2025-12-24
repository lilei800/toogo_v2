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

// DepositCreateReq 创建充值订单请求
type DepositCreateReq struct {
	g.Meta `path:"/payment/deposit/create" method:"post" tags:"USDT充值" summary:"创建充值订单"`
	input.DepositCreateInp
}

type DepositCreateRes struct {
	OrderSn string `json:"orderSn" dc:"订单号"`
}

// DepositListReq 充值订单列表请求
type DepositListReq struct {
	g.Meta `path:"/payment/deposit/list" method:"get" tags:"USDT充值" summary:"充值订单列表"`
	input.DepositListInp
}

type DepositListRes struct {
	List       []*entity.UsdtDeposit `json:"list" dc:"列表"`
	TotalCount int                   `json:"totalCount" dc:"总数"`
	Page       int                   `json:"page" dc:"页码"`
	PageSize   int                   `json:"pageSize" dc:"每页数量"`
}

// DepositViewReq 查看充值订单请求
type DepositViewReq struct {
	g.Meta `path:"/payment/deposit/view" method:"get" tags:"USDT充值" summary:"查看充值订单"`
	Id     int64 `json:"id" v:"required#ID不能为空" dc:"订单ID"`
}

type DepositViewRes struct {
	*entity.UsdtDeposit
}

// DepositCheckReq 检查充值状态请求
type DepositCheckReq struct {
	g.Meta `path:"/payment/deposit/check" method:"post" tags:"USDT充值" summary:"检查充值状态"`
	Id     int64 `json:"id" v:"required#ID不能为空" dc:"订单ID"`
}

type DepositCheckRes struct{}

// DepositCancelReq 取消充值订单请求
type DepositCancelReq struct {
	g.Meta `path:"/payment/deposit/cancel" method:"post" tags:"USDT充值" summary:"取消充值订单"`
	Id     int64 `json:"id" v:"required#ID不能为空" dc:"订单ID"`
}

type DepositCancelRes struct{}



