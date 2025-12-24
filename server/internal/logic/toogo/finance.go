// Package toogo Toogo财务模块 - 充值提现 (简化版)
package toogo

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"
)

// ToogoFinance 财务服务
type ToogoFinance struct{}

var financeService = &ToogoFinance{}

// GetFinance 获取财务服务单例
func GetFinance() *ToogoFinance {
	return financeService
}

// genOrderSn 生成订单号
func genOrderSn(prefix string) string {
	// 【修复】使用标准 Go time 包的正确格式，或使用 gtime
	return prefix + time.Now().Format("20060102150405") + grand.S(6)
}

// CreateDeposit 创建充值订单
func (f *ToogoFinance) CreateDeposit(ctx context.Context, in *toogoin.CreateDepositInp) (*toogoin.CreateDepositModel, error) {
	// 生成订单号
	orderSn := genOrderSn("D")

	// 创建订单记录
	deposit := &entity.ToogoDeposit{
		UserId:  in.UserId,
		OrderSn: orderSn,
		Amount:  in.Amount,
		Network: in.Network,
		Status:  1, // 待支付
	}

	_, err := dao.ToogoDeposit.Ctx(ctx).Data(deposit).Insert()
	if err != nil {
		return nil, gerror.Wrap(err, "创建充值订单失败")
	}

	return &toogoin.CreateDepositModel{
		OrderSn:  orderSn,
		Amount:   in.Amount,
		Currency: in.Currency,
		Network:  in.Network,
	}, nil
}

// DepositCallback 充值回调
func (f *ToogoFinance) DepositCallback(ctx context.Context, in *toogoin.DepositCallbackInp) error {
	// 查询订单
	var deposit *entity.ToogoDeposit
	err := dao.ToogoDeposit.Ctx(ctx).Where("order_sn", in.OrderSn).Scan(&deposit)
	if err != nil {
		return gerror.Wrap(err, "查询订单失败")
	}
	if deposit == nil {
		return gerror.New("订单不存在")
	}

	if deposit.Status != 1 {
		return gerror.New("订单状态异常")
	}

	// 更新订单状态
	_, err = dao.ToogoDeposit.Ctx(ctx).Where(dao.ToogoDeposit.Columns().Id, deposit.Id).Data(g.Map{
		"status":     2, // 已完成
		"tx_hash":    in.TxHash,
		"paid_at":    gtime.Now(),
		"updated_at": gtime.Now(),
	}).Update()
	if err != nil {
		return gerror.Wrap(err, "更新订单状态失败")
	}

	// 增加用户余额 - 使用 ChangeBalance 记录流水
	err = NewToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
		UserId:      deposit.UserId,
		AccountType: "balance",
		ChangeType:  "deposit",
		Amount:      in.Amount,
		OrderSn:     deposit.OrderSn,
		Remark:      "充值到账",
	})
	if err != nil {
		return gerror.Wrap(err, "增加用户余额失败")
	}

	// 更新累计充值
	_, err = dao.ToogoWallet.Ctx(ctx).Where("user_id", deposit.UserId).Data(g.Map{
		"total_deposit": g.DB().Raw(fmt.Sprintf("total_deposit + %f", in.Amount)),
		"updated_at":    gtime.Now(),
	}).Update()
	if err != nil {
		return gerror.Wrap(err, "更新累计充值失败")
	}

	return nil
}

// CreateWithdraw 创建提现申请
func (f *ToogoFinance) CreateWithdraw(ctx context.Context, in *toogoin.CreateWithdrawInp) (*toogoin.CreateWithdrawModel, error) {
	// 获取提现配置
	minAmount, _ := GetConfig().GetWithdrawMinAmount(ctx)
	if in.Amount < minAmount {
		return nil, gerror.Newf("最低提现金额为 %.2f USDT", minAmount)
	}

	// 获取手续费比例
	feeRate, _ := GetConfig().GetWithdrawFeeRate(ctx)
	fee := in.Amount * feeRate
	actualAmount := in.Amount - fee

	// 检查余额
	var wallet *entity.ToogoWallet
	err := dao.ToogoWallet.Ctx(ctx).Where("user_id", in.UserId).Scan(&wallet)
	if err != nil {
		return nil, gerror.Wrap(err, "获取钱包信息失败")
	}
	if wallet == nil {
		return nil, gerror.New("钱包不存在")
	}

	switch in.AccountType {
	case "balance":
		if wallet.Balance < in.Amount {
			return nil, gerror.New("余额不足")
		}
	case "commission":
		if wallet.Commission < in.Amount {
			return nil, gerror.New("佣金余额不足")
		}
	default:
		return nil, gerror.New("账户类型错误")
	}

	// 生成订单号
	orderSn := genOrderSn("W")

	// 创建提现记录
	withdraw := &entity.ToogoWithdraw{
		UserId:      in.UserId,
		OrderSn:     orderSn,
		AccountType: in.AccountType,
		Amount:      in.Amount,
		Fee:         fee,
		RealAmount:  actualAmount,
		ToAddress:   in.ToAddress,
		Network:     in.Network,
		Status:      1, // 待审核
	}

	_, err = dao.ToogoWithdraw.Ctx(ctx).Data(withdraw).Insert()
	if err != nil {
		return nil, gerror.Wrap(err, "创建提现申请失败")
	}

	// 冻结余额
	if in.AccountType == "balance" {
		_, err = dao.ToogoWallet.Ctx(ctx).Where("user_id", in.UserId).Data(g.Map{
			"balance":        g.DB().Raw(fmt.Sprintf("balance - %f", in.Amount)),
			"frozen_balance": g.DB().Raw(fmt.Sprintf("frozen_balance + %f", in.Amount)),
			"updated_at":     gtime.Now(),
		}).Update()
	} else {
		_, err = dao.ToogoWallet.Ctx(ctx).Where("user_id", in.UserId).Data(g.Map{
			"commission":        g.DB().Raw(fmt.Sprintf("commission - %f", in.Amount)),
			"frozen_commission": g.DB().Raw(fmt.Sprintf("frozen_commission + %f", in.Amount)),
			"updated_at":        gtime.Now(),
		}).Update()
	}
	if err != nil {
		return nil, gerror.Wrap(err, "冻结余额失败")
	}

	return &toogoin.CreateWithdrawModel{
		OrderSn:      orderSn,
		Amount:       in.Amount,
		Fee:          fee,
		ActualAmount: actualAmount,
		Status:       "pending",
	}, nil
}

// AuditWithdraw 审核提现
func (f *ToogoFinance) AuditWithdraw(ctx context.Context, in *toogoin.WithdrawAuditInp) error {
	// 查询提现记录
	var withdraw *entity.ToogoWithdraw
	err := dao.ToogoWithdraw.Ctx(ctx).Where(dao.ToogoWithdraw.Columns().Id, in.Id).Scan(&withdraw)
	if err != nil {
		return gerror.Wrap(err, "查询提现记录失败")
	}
	if withdraw == nil {
		return gerror.New("提现记录不存在")
	}

	if withdraw.Status != 1 {
		return gerror.New("提现状态异常，无法审核")
	}

	if in.Status == 2 {
		// 审核通过
		_, err = dao.ToogoWithdraw.Ctx(ctx).Where(dao.ToogoWithdraw.Columns().Id, in.Id).Data(g.Map{
			"status":       2, // 审核通过
			"audited_by":   in.AuditId,
			"audited_at":   gtime.Now(),
			"audit_remark": in.AuditNote,
			"updated_at":   gtime.Now(),
		}).Update()
		if err != nil {
			return gerror.Wrap(err, "更新审核状态失败")
		}

	} else if in.Status == 4 {
		// 审核拒绝
		_, err = dao.ToogoWithdraw.Ctx(ctx).Where(dao.ToogoWithdraw.Columns().Id, in.Id).Data(g.Map{
			"status":       3, // 审核拒绝
			"audited_by":   in.AuditId,
			"audited_at":   gtime.Now(),
			"audit_remark": in.AuditNote,
			"updated_at":   gtime.Now(),
		}).Update()
		if err != nil {
			return gerror.Wrap(err, "更新审核状态失败")
		}

		// 解冻余额并记录流水
		err = NewToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
			UserId:      withdraw.UserId,
			AccountType: withdraw.AccountType,
			ChangeType:  "withdraw_reject",
			Amount:      withdraw.Amount,
			OrderSn:     withdraw.OrderSn,
			Remark:      "提现审核拒绝，余额已退回",
		})
		if err != nil {
			return gerror.Wrap(err, "解冻余额失败")
		}

		// 更新冻结余额
		if withdraw.AccountType == "balance" {
			_, err = dao.ToogoWallet.Ctx(ctx).Where("user_id", withdraw.UserId).Data(g.Map{
				"frozen_balance": g.DB().Raw(fmt.Sprintf("frozen_balance - %f", withdraw.Amount)),
				"updated_at":     gtime.Now(),
			}).Update()
		} else {
			_, err = dao.ToogoWallet.Ctx(ctx).Where("user_id", withdraw.UserId).Data(g.Map{
				"frozen_commission": g.DB().Raw(fmt.Sprintf("frozen_commission - %f", withdraw.Amount)),
				"updated_at":        gtime.Now(),
			}).Update()
		}
		if err != nil {
			return gerror.Wrap(err, "更新冻结余额失败")
		}
	}

	return nil
}

// HandleNOWPaymentsIPNCallback 处理NOWPayments充值IPN回调
func (f *ToogoFinance) HandleNOWPaymentsIPNCallback(ctx context.Context) error {
	// 获取请求体
	request := g.RequestFromCtx(ctx)
	body := request.GetBodyString()
	g.Log().Infof(ctx, "[NOWPayments] IPN回调: %s", body)

	// 解析回调数据
	jsonData := gjson.New(body)
	paymentId := jsonData.Get("payment_id").String()
	paymentStatus := jsonData.Get("payment_status").String()
	orderSn := jsonData.Get("order_id").String()
	actualAmount := jsonData.Get("actually_paid").Float64()
	txHash := jsonData.Get("pay_in_tx_hash").String()

	if orderSn == "" {
		return gerror.New("订单号为空")
	}

	// 查询订单
	var deposit *entity.ToogoDeposit
	err := dao.ToogoDeposit.Ctx(ctx).Where("order_sn", orderSn).Scan(&deposit)
	if err != nil || deposit == nil {
		return gerror.Newf("订单不存在: %s", orderSn)
	}

	// 根据状态处理
	switch paymentStatus {
	case "finished", "confirmed":
		// 支付完成
		if deposit.Status != 1 {
			g.Log().Infof(ctx, "[NOWPayments] 订单已处理: %s, status=%d", orderSn, deposit.Status)
			return nil
		}

		// 更新订单状态
		_, err = dao.ToogoDeposit.Ctx(ctx).Where(dao.ToogoDeposit.Columns().Id, deposit.Id).Data(g.Map{
			"status":        2, // 已完成
			"payment_id":    paymentId,
			"tx_hash":       txHash,
			"actual_amount": actualAmount,
			"paid_at":       gtime.Now(),
			"updated_at":    gtime.Now(),
		}).Update()
		if err != nil {
			return gerror.Wrap(err, "更新订单状态失败")
		}

		// 增加用户余额 - 使用 ChangeBalance 记录流水
		err = NewToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
			UserId:      deposit.UserId,
			AccountType: "balance",
			ChangeType:  "deposit",
			Amount:      actualAmount,
			OrderSn:     orderSn,
			Remark:      fmt.Sprintf("NOWPayments充值到账，TxHash: %s", txHash),
		})
		if err != nil {
			return gerror.Wrap(err, "增加用户余额失败")
		}

		// 更新累计充值
		_, err = dao.ToogoWallet.Ctx(ctx).Where("user_id", deposit.UserId).Data(g.Map{
			"total_deposit": g.DB().Raw(fmt.Sprintf("total_deposit + %f", actualAmount)),
			"updated_at":    gtime.Now(),
		}).Update()
		if err != nil {
			return gerror.Wrap(err, "更新累计充值失败")
		}

		// 推送通知
		GetPusher().PushSystemNotice(ctx, deposit.UserId, "充值成功", 
			fmt.Sprintf("您的充值订单 %s 已完成，到账 %.4f USDT", orderSn, actualAmount), "success")

		g.Log().Infof(ctx, "[NOWPayments] 充值成功: orderSn=%s, userId=%d, amount=%.4f", 
			orderSn, deposit.UserId, actualAmount)

	case "partially_paid":
		// 部分支付
		_, _ = dao.ToogoDeposit.Ctx(ctx).Where(dao.ToogoDeposit.Columns().Id, deposit.Id).Data(g.Map{
			"payment_id":    paymentId,
			"actual_amount": actualAmount,
			"remark":        "部分支付",
			"updated_at":    gtime.Now(),
		}).Update()

	case "expired":
		// 订单过期
		_, _ = dao.ToogoDeposit.Ctx(ctx).Where(dao.ToogoDeposit.Columns().Id, deposit.Id).Data(g.Map{
			"status":     4, // 已过期
			"remark":     "订单已过期",
			"updated_at": gtime.Now(),
		}).Update()

	case "failed", "refunded":
		// 支付失败或已退款
		_, _ = dao.ToogoDeposit.Ctx(ctx).Where(dao.ToogoDeposit.Columns().Id, deposit.Id).Data(g.Map{
			"status":     3, // 失败
			"remark":     paymentStatus,
			"updated_at": gtime.Now(),
		}).Update()
	}

	return nil
}

// HandleNOWPaymentsPayoutIPNCallback 处理NOWPayments提现IPN回调
func (f *ToogoFinance) HandleNOWPaymentsPayoutIPNCallback(ctx context.Context) error {
	// 获取请求体
	request := g.RequestFromCtx(ctx)
	body := request.GetBodyString()
	g.Log().Infof(ctx, "[NOWPayments] Payout IPN回调: %s", body)

	// 解析回调数据
	jsonData := gjson.New(body)
	payoutId := jsonData.Get("id").String()
	status := jsonData.Get("status").String()
	batchOrderSn := jsonData.Get("batch_withdrawal_id").String()
	txHash := jsonData.Get("hash").String()

	if batchOrderSn == "" {
		return gerror.New("批次号为空")
	}

	// 查询提现记录
	var withdraw *entity.ToogoWithdraw
	err := dao.ToogoWithdraw.Ctx(ctx).Where("batch_id", batchOrderSn).Scan(&withdraw)
	if err != nil || withdraw == nil {
		return gerror.Newf("提现记录不存在: %s", batchOrderSn)
	}

	// 根据状态处理
	switch status {
	case "FINISHED":
		// 提现完成
		if withdraw.Status == 3 {
			return nil // 已完成
		}

		_, err = dao.ToogoWithdraw.Ctx(ctx).Where(dao.ToogoWithdraw.Columns().Id, withdraw.Id).Data(g.Map{
			"status":       4, // 已完成
			"payout_id":    payoutId,
			"tx_hash":      txHash,
			"completed_at": gtime.Now(),
			"updated_at":   gtime.Now(),
		}).Update()
		if err != nil {
			return gerror.Wrap(err, "更新提现状态失败")
		}

		// 扣除冻结余额（提现完成，从冻结中扣除）
		frozenField := "frozen_balance"
		if withdraw.AccountType == "commission" {
			frozenField = "frozen_commission"
		}
		_, _ = dao.ToogoWallet.Ctx(ctx).Where("user_id", withdraw.UserId).Data(g.Map{
			frozenField:        g.DB().Raw(fmt.Sprintf("%s - %f", frozenField, withdraw.Amount)),
			"total_withdraw":   g.DB().Raw(fmt.Sprintf("total_withdraw + %f", withdraw.RealAmount)),
			"updated_at":       gtime.Now(),
		}).Update()

		// 记录提现完成流水
		_ = NewToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
			UserId:      withdraw.UserId,
			AccountType: withdraw.AccountType,
			ChangeType:  "withdraw_complete",
			Amount:      0, // 余额不变（之前已从冻结扣除）
			OrderSn:     withdraw.OrderSn,
			Remark:      fmt.Sprintf("提现成功，实际到账 %.4f USDT", withdraw.RealAmount),
		})

		// 推送通知
		GetPusher().PushSystemNotice(ctx, withdraw.UserId, "提现成功",
			fmt.Sprintf("您的提现订单 %s 已完成，实际到账 %.4f USDT", withdraw.OrderSn, withdraw.RealAmount), "success")

		g.Log().Infof(ctx, "[NOWPayments] 提现成功: orderSn=%s, userId=%d, amount=%.4f",
			withdraw.OrderSn, withdraw.UserId, withdraw.RealAmount)

	case "FAILED", "REJECTED":
		// 提现失败
		_, err = dao.ToogoWithdraw.Ctx(ctx).Where(dao.ToogoWithdraw.Columns().Id, withdraw.Id).Data(g.Map{
			"status":     5, // 已取消/失败
			"remark":     status,
			"updated_at": gtime.Now(),
		}).Update()
		if err != nil {
			return gerror.Wrap(err, "更新提现状态失败")
		}

		// 解冻余额并记录流水
		_ = NewToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
			UserId:      withdraw.UserId,
			AccountType: withdraw.AccountType,
			ChangeType:  "withdraw_fail",
			Amount:      withdraw.Amount, // 退回余额
			OrderSn:     withdraw.OrderSn,
			Remark:      "提现失败，余额已退回",
		})

		// 更新冻结余额
		frozenField := "frozen_balance"
		if withdraw.AccountType == "commission" {
			frozenField = "frozen_commission"
		}
		_, _ = dao.ToogoWallet.Ctx(ctx).Where("user_id", withdraw.UserId).Data(g.Map{
			frozenField:  g.DB().Raw(fmt.Sprintf("%s - %f", frozenField, withdraw.Amount)),
			"updated_at": gtime.Now(),
		}).Update()

		// 推送通知
		GetPusher().PushSystemNotice(ctx, withdraw.UserId, "提现失败",
			fmt.Sprintf("您的提现订单 %s 处理失败，已退回账户", withdraw.OrderSn), "error")

	case "PROCESSING", "SENDING":
		// 处理中
		_, _ = dao.ToogoWithdraw.Ctx(ctx).Where(dao.ToogoWithdraw.Columns().Id, withdraw.Id).Data(g.Map{
			"payout_id":  payoutId,
			"remark":     "处理中",
			"updated_at": gtime.Now(),
		}).Update()
	}

	return nil
}
