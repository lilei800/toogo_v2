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

// Withdraw USDT提现管理
type sWithdraw struct{}

var Withdraw = &sWithdraw{}

// Apply 申请提现
func (s *sWithdraw) Apply(ctx context.Context, in *input.WithdrawApplyInp) (orderSn string, err error) {
	user := simple.User(ctx)
	if user == nil {
		return "", gerror.New("用户未登录")
	}

	// 检查余额是否足够
	var balance *entity.UsdtBalance
	err = dao.UsdtBalance.Ctx(ctx).Where("user_id", user.Id).Scan(&balance)
	if err != nil {
		return "", err
	}
	if balance == nil || balance.Balance < in.Amount {
		return "", gerror.New("余额不足")
	}

	// 生成订单号
	orderSn = fmt.Sprintf("W%d%s", user.Id, gtime.Now().Format("YmdHis"))

	// 开启事务
	err = dao.UsdtWithdraw.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 创建提现订单
		data := &entity.UsdtWithdraw{
			UserId:    user.Id,
			OrderSn:   orderSn,
			Amount:    in.Amount,
			ToAddress: in.ToAddress,
			Network:   in.Network,
			Status:    1, // 待审核
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}

		_, err = dao.UsdtWithdraw.Ctx(ctx).Data(data).Insert()
		if err != nil {
			return err
		}

		// 冻结余额
		_, err = dao.UsdtBalance.Ctx(ctx).
			Where("user_id", user.Id).
			Data(g.Map{
				"balance":        gdb.Raw(fmt.Sprintf("balance - %v", in.Amount)),
				"frozen_balance": gdb.Raw(fmt.Sprintf("frozen_balance + %v", in.Amount)),
				"updated_at":     gtime.Now(),
			}).
			Update()

		return err
	})

	if err != nil {
		return "", err
	}

	return orderSn, nil
}

// List 提现订单列表
func (s *sWithdraw) List(ctx context.Context, in *input.WithdrawListInp) (list []*entity.UsdtWithdraw, totalCount int, err error) {
	user := simple.User(ctx)
	if user == nil {
		return nil, 0, gerror.New("用户未登录")
	}

	mod := dao.UsdtWithdraw.Ctx(ctx).Where("user_id", user.Id)

	// 状态筛选
	if in.Status > 0 {
		mod = mod.Where("status", in.Status)
	}

	// 订单号筛选
	if in.OrderSn != "" {
		mod = mod.Where("order_sn", in.OrderSn)
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

// View 查看提现订单详情
func (s *sWithdraw) View(ctx context.Context, id int64) (res *entity.UsdtWithdraw, err error) {
	user := simple.User(ctx)
	if user == nil {
		return nil, gerror.New("用户未登录")
	}

	err = dao.UsdtWithdraw.Ctx(ctx).
		Where("id", id).
		Where("user_id", user.Id).
		Scan(&res)
	
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, gerror.New("提现订单不存在")
	}

	return
}

// Audit 审核提现（管理员）
func (s *sWithdraw) Audit(ctx context.Context, in *input.WithdrawAuditInp) error {
	user := simple.User(ctx)
	if user == nil {
		return gerror.New("用户未登录")
	}

	// 查询提现订单
	var withdraw *entity.UsdtWithdraw
	err := dao.UsdtWithdraw.Ctx(ctx).Where("id", in.Id).Scan(&withdraw)
	if err != nil {
		return err
	}
	if withdraw == nil {
		return gerror.New("提现订单不存在")
	}

	if withdraw.Status != 1 {
		return gerror.New("只能审核待审核订单")
	}

	return dao.UsdtWithdraw.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 更新订单状态
		_, err = dao.UsdtWithdraw.Ctx(ctx).
			Where("id", in.Id).
			Data(g.Map{
				"status":      in.Status,
				"audit_remark": in.Remark,
				"audited_by":  user.Id,
				"audited_at":  gtime.Now(),
				"updated_at":  gtime.Now(),
			}).
			Update()
		if err != nil {
			return err
		}

		// 如果审核通过，处理余额
		if in.Status == 2 { // 审核通过
			// 扣除冻结余额
			_, err = dao.UsdtBalance.Ctx(ctx).
				Where("user_id", withdraw.UserId).
				Data(g.Map{
					"frozen_balance": gdb.Raw(fmt.Sprintf("frozen_balance - %v", withdraw.Amount)),
					"updated_at":     gtime.Now(),
				}).
				Update()
			if err != nil {
				return err
			}

			// 记录余额变动
			_, err = dao.UsdtBalanceLog.Ctx(ctx).Data(&entity.UsdtBalanceLog{
				UserId:       withdraw.UserId,
				Type:         2, // 提现
				ChangeAmount: -withdraw.Amount,
				OrderSn:      withdraw.OrderSn,
				Remark:       "USDT提现",
				CreatedAt:    gtime.Now(),
			}).Insert()

		} else if in.Status == 3 { // 审核拒绝
			// 解冻余额
			_, err = dao.UsdtBalance.Ctx(ctx).
				Where("user_id", withdraw.UserId).
				Data(g.Map{
					"balance":        gdb.Raw(fmt.Sprintf("balance + %v", withdraw.Amount)),
					"frozen_balance": gdb.Raw(fmt.Sprintf("frozen_balance - %v", withdraw.Amount)),
					"updated_at":     gtime.Now(),
				}).
				Update()
		}

		return err
	})
}

// Cancel 取消提现
func (s *sWithdraw) Cancel(ctx context.Context, id int64) error {
	user := simple.User(ctx)
	if user == nil {
		return gerror.New("用户未登录")
	}

	withdraw, err := s.View(ctx, id)
	if err != nil {
		return err
	}

	if withdraw.Status != 1 {
		return gerror.New("只能取消待审核订单")
	}

	return dao.UsdtWithdraw.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 更新状态为已取消
		_, err = dao.UsdtWithdraw.Ctx(ctx).
			Where("id", id).
			Where("user_id", user.Id).
			Data(g.Map{
				"status":     5, // 已取消
				"updated_at": gtime.Now(),
			}).
			Update()
		if err != nil {
			return err
		}

		// 解冻余额
		_, err = dao.UsdtBalance.Ctx(ctx).
			Where("user_id", user.Id).
			Data(g.Map{
				"balance":        gdb.Raw(fmt.Sprintf("balance + %v", withdraw.Amount)),
				"frozen_balance": gdb.Raw(fmt.Sprintf("frozen_balance - %v", withdraw.Amount)),
				"updated_at":     gtime.Now(),
			}).
			Update()

		return err
	})
}

// AdminList 管理员查看提现列表
func (s *sWithdraw) AdminList(ctx context.Context, in *input.WithdrawListInp) (list []*entity.UsdtWithdraw, totalCount int, err error) {
	mod := dao.UsdtWithdraw.Ctx(ctx)

	// 状态筛选
	if in.Status > 0 {
		mod = mod.Where("status", in.Status)
	}

	// 订单号筛选
	if in.OrderSn != "" {
		mod = mod.Where("order_sn", in.OrderSn)
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

