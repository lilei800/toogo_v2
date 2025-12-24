// Package payment
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package payment

import (
	"context"
	"fmt"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"
	"hotgo/utility/simple"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Deposit USDT充值管理
type sDeposit struct{}

var Deposit = &sDeposit{}

// Create 创建充值订单
func (s *sDeposit) Create(ctx context.Context, in *input.DepositCreateInp) (orderSn string, err error) {
	// 获取当前用户信息
	user := simple.User(ctx)
	if user == nil {
		return "", gerror.New("用户未登录")
	}

	// 生成订单号
	orderSn = fmt.Sprintf("D%d%s", user.Id, gtime.Now().Format("YmdHis"))

	// 创建充值订单
	data := &entity.UsdtDeposit{
		UserId:    user.Id,
		OrderSn:   orderSn,
		Amount:    in.Amount,
		Network:   in.Network,
		Status:    1, // 待支付
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}

	_, err = dao.UsdtDeposit.Ctx(ctx).Data(data).Insert()
	if err != nil {
		return "", err
	}

	return orderSn, nil
}

// List 充值订单列表
func (s *sDeposit) List(ctx context.Context, in *input.DepositListInp) (list []*entity.UsdtDeposit, totalCount int, err error) {
	user := simple.User(ctx)
	if user == nil {
		return nil, 0, gerror.New("用户未登录")
	}

	mod := dao.UsdtDeposit.Ctx(ctx).Where("user_id", user.Id)

	// 状态筛选
	if in.Status > 0 {
		mod = mod.Where("status", in.Status)
	}

	// 订单号筛选
	if in.OrderSn != "" {
		mod = mod.Where("order_sn", in.OrderSn)
	}

	// 网络筛选
	if in.Network != "" {
		mod = mod.Where("network", in.Network)
	}

	// 总数
	totalCount, err = mod.Count()
	if err != nil {
		return
	}

	// 分页查询
	err = mod.Page(in.Page, in.PageSize).OrderDesc("id").Scan(&list)
	return
}

// View 查看充值订单详情
func (s *sDeposit) View(ctx context.Context, id int64) (res *entity.UsdtDeposit, err error) {
	user := simple.User(ctx)
	if user == nil {
		return nil, gerror.New("用户未登录")
	}

	err = dao.UsdtDeposit.Ctx(ctx).
		Where("id", id).
		Where("user_id", user.Id).
		Scan(&res)
	
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, gerror.New("充值订单不存在")
	}

	return
}

// Check 检查充值状态
func (s *sDeposit) Check(ctx context.Context, id int64) error {
	user := simple.User(ctx)
	if user == nil {
		return gerror.New("用户未登录")
	}

	deposit, err := s.View(ctx, id)
	if err != nil {
		return err
	}

	if deposit.Status != 1 {
		return gerror.New("订单状态不允许检查")
	}

	// TODO: 调用NOWPayments API检查支付状态
	// 这里简化处理，实际应该调用第三方接口

	return nil
}

// Cancel 取消充值订单
func (s *sDeposit) Cancel(ctx context.Context, id int64) error {
	user := simple.User(ctx)
	if user == nil {
		return gerror.New("用户未登录")
	}

	deposit, err := s.View(ctx, id)
	if err != nil {
		return err
	}

	if deposit.Status != 1 {
		return gerror.New("只能取消待支付订单")
	}

	// 更新状态为已取消
	_, err = dao.UsdtDeposit.Ctx(ctx).
		Where("id", id).
		Where("user_id", user.Id).
		Data(g.Map{
			"status":     5, // 已取消
			"updated_at": gtime.Now(),
		}).
		Update()

	return err
}

// UpdateStatus 更新充值订单状态（系统内部调用）
func (s *sDeposit) UpdateStatus(ctx context.Context, orderSn string, status int, paymentId string) error {
	return dao.UsdtDeposit.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 更新订单状态
		_, err := dao.UsdtDeposit.Ctx(ctx).
			Where("order_sn", orderSn).
			Data(g.Map{
				"status":     status,
				"payment_id": paymentId,
				"paid_at":    gtime.Now(),
				"updated_at": gtime.Now(),
			}).
			Update()
		if err != nil {
			return err
		}

		// 如果支付成功，增加用户余额
		if status == 2 { // 已完成
			var deposit *entity.UsdtDeposit
			err = dao.UsdtDeposit.Ctx(ctx).Where("order_sn", orderSn).Scan(&deposit)
			if err != nil {
				return err
			}

			// 增加余额
			_, err = dao.UsdtBalance.Ctx(ctx).
				Where("user_id", deposit.UserId).
				Increment("balance", deposit.Amount)
			if err != nil {
				return err
			}

			// 记录余额变动
			_, err = dao.UsdtBalanceLog.Ctx(ctx).Data(&entity.UsdtBalanceLog{
				UserId:       deposit.UserId,
				Type:         1, // 充值
				ChangeAmount: deposit.Amount,
				OrderSn:      orderSn,
				Remark:       "USDT充值",
				CreatedAt:    gtime.Now(),
			}).Insert()
		}

		return err
	})
}

