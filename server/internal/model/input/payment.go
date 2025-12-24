// Package input
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package input

import "github.com/gogf/gf/v2/os/gtime"

// DepositCreateInp 创建充值订单输入
type DepositCreateInp struct {
	Amount  float64 `json:"amount" v:"required|min:0.01#充值金额不能为空|充值金额最少0.01" dc:"充值金额"`
	Network string  `json:"network" v:"required#网络不能为空" dc:"网络(TRC20/ERC20)"`
}

// DepositListInp 充值订单列表输入
type DepositListInp struct {
	PageReq
	Status  int    `json:"status" dc:"状态：1待支付 2已完成 3已超时 4已退款 5已取消"`
	OrderSn string `json:"orderSn" dc:"订单号"`
	Network string `json:"network" dc:"网络"`
}

// WithdrawApplyInp 申请提现输入
type WithdrawApplyInp struct {
	Amount    float64 `json:"amount" v:"required|min:0.01#提现金额不能为空|提现金额最少0.01" dc:"提现金额"`
	ToAddress string  `json:"toAddress" v:"required#提现地址不能为空" dc:"提现地址"`
	Network   string  `json:"network" v:"required#网络不能为空" dc:"网络(TRC20/ERC20)"`
}

// WithdrawListInp 提现订单列表输入
type WithdrawListInp struct {
	PageReq
	Status    int    `json:"status" dc:"状态：1待审核 2审核通过 3审核拒绝 4已完成 5已取消"`
	OrderSn   string `json:"orderSn" dc:"订单号"`
	ToAddress string `json:"toAddress" dc:"提现地址"`
	Network   string `json:"network" dc:"网络"`
}

// WithdrawAuditInp 审核提现输入
type WithdrawAuditInp struct {
	Id     int64  `json:"id" v:"required#ID不能为空" dc:"提现订单ID"`
	Status int    `json:"status" v:"required|in:2,3#状态不能为空|状态必须为2或3" dc:"状态：2审核通过 3审核拒绝"`
	Remark string `json:"remark" dc:"审核备注"`
}

// BalanceLogListInp 资金流水列表输入
type BalanceLogListInp struct {
	PageReq
	Type    int       `json:"type" dc:"类型：1充值 2提现 3支付 4退款"`
	OrderSn string    `json:"orderSn" dc:"订单号"`
	StartAt *gtime.Time `json:"startAt" dc:"开始时间"`
	EndAt   *gtime.Time `json:"endAt" dc:"结束时间"`
}

// PageReq 分页请求
type PageReq struct {
	Page     int `json:"page" v:"required|min:1#页码不能为空|页码最小为1" dc:"页码"`
	PageSize int `json:"pageSize" v:"required|min:1|max:100#每页数量不能为空|每页数量最小为1|每页数量最大为100" dc:"每页数量"`
}



