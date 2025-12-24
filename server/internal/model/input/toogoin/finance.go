// Package toogoin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogoin

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// ========== 充值相关 ==========

// CreateDepositInp 创建充值订单输入
type CreateDepositInp struct {
	UserId   int64   `json:"userId" v:"required" description:"用户ID"`
	Amount   float64 `json:"amount" v:"required|min:1" description:"充值金额"`
	Currency string  `json:"currency" v:"required" description:"币种: USDT"`
	Network  string  `json:"network" v:"required" description:"网络: TRC20, ERC20, BEP20"`
}

// CreateDepositModel 创建充值订单返回
type CreateDepositModel struct {
	OrderSn    string  `json:"orderSn" description:"订单号"`
	Amount     float64 `json:"amount" description:"充值金额"`
	Currency   string  `json:"currency" description:"币种"`
	Network    string  `json:"network" description:"网络"`
	ToAddress  string  `json:"toAddress" description:"充值地址"`
	ExpireAt   string  `json:"expireAt" description:"过期时间"`
	PaymentUrl string  `json:"paymentUrl" description:"支付页面URL"`
}

// DepositCallbackInp 充值回调输入
type DepositCallbackInp struct {
	OrderSn     string  `json:"orderSn" v:"required" description:"订单号"`
	Amount      float64 `json:"amount" v:"required" description:"实际到账金额"`
	TxHash      string  `json:"txHash" v:"required" description:"交易哈希"`
	FromAddress string  `json:"fromAddress" description:"来源地址"`
	Confirms    int     `json:"confirms" description:"确认数"`
}

// DepositListInp 充值记录列表输入
type DepositListInp struct {
	form.PageReq
	UserId    int64    `json:"userId" description:"用户ID"`
	Status    int      `json:"status" description:"状态"`
	CreatedAt []string `json:"createdAt" description:"创建时间"`
}

// DepositListModel 充值记录列表返回
type DepositListModel struct {
	*entity.ToogoDeposit
	Username string `json:"username" description:"用户名"`
}

// ========== 提现相关 ==========

// CreateWithdrawInp 创建提现申请输入
type CreateWithdrawInp struct {
	UserId      int64   `json:"userId" v:"required" description:"用户ID"`
	AccountType string  `json:"accountType" v:"required|in:balance,commission" description:"账户类型: balance=余额, commission=佣金"`
	Amount      float64 `json:"amount" v:"required|min:10" description:"提现金额"`
	Currency    string  `json:"currency" v:"required" description:"币种: USDT"`
	Network     string  `json:"network" v:"required" description:"网络: TRC20, ERC20, BEP20"`
	ToAddress   string  `json:"toAddress" v:"required" description:"提现地址"`
}

// CreateWithdrawModel 创建提现申请返回
type CreateWithdrawModel struct {
	OrderSn      string  `json:"orderSn" description:"订单号"`
	Amount       float64 `json:"amount" description:"提现金额"`
	Fee          float64 `json:"fee" description:"手续费"`
	ActualAmount float64 `json:"actualAmount" description:"实际到账金额"`
	Status       string  `json:"status" description:"状态"`
}

// WithdrawListInp 提现记录列表输入
type WithdrawListInp struct {
	form.PageReq
	UserId    int64    `json:"userId" description:"用户ID"`
	Status    int      `json:"status" description:"状态"`
	CreatedAt []string `json:"createdAt" description:"创建时间"`
}

// WithdrawListModel 提现记录列表返回
type WithdrawListModel struct {
	*entity.ToogoWithdraw
	Username string `json:"username" description:"用户名"`
}

// WithdrawAuditInp 提现审核输入
type WithdrawAuditInp struct {
	Id        int64  `json:"id" v:"required" description:"提现ID"`
	Status    int    `json:"status" v:"required|in:2,4" description:"审核状态: 2=通过, 4=拒绝"`
	AuditId   int64  `json:"auditId" description:"审核人ID"`
	AuditNote string `json:"auditNote" description:"审核备注"`
}

// WithdrawCompleteInp 提现完成回调输入
type WithdrawCompleteInp struct {
	OrderSn string `json:"orderSn" v:"required" description:"订单号"`
	TxHash  string `json:"txHash" v:"required" description:"交易哈希"`
}

