// Package payment
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package payment

import (
	"context"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"
	"hotgo/utility/simple"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
)

// Balance USDT余额管理
type sBalance struct{}

var Balance = &sBalance{}

// View 查看余额
func (s *sBalance) View(ctx context.Context) (res *entity.UsdtBalance, err error) {
	user := simple.User(ctx)
	if user == nil {
		return nil, gerror.New("用户未登录")
	}

	err = dao.UsdtBalance.Ctx(ctx).Where("user_id", user.Id).Scan(&res)
	if err != nil {
		return nil, err
	}

	// 如果用户余额记录不存在，创建一个
	if res == nil {
		res = &entity.UsdtBalance{
			UserId:        user.Id,
			Balance:       0,
			FrozenBalance: 0,
			CreatedAt:     gtime.Now(),
			UpdatedAt:     gtime.Now(),
		}
		_, err = dao.UsdtBalance.Ctx(ctx).Data(res).Insert()
		if err != nil {
			return nil, err
		}
	}

	return
}

// LogList 资金流水列表
func (s *sBalance) LogList(ctx context.Context, in *input.BalanceLogListInp) (list []*entity.UsdtBalanceLog, totalCount int, err error) {
	user := simple.User(ctx)
	if user == nil {
		return nil, 0, gerror.New("用户未登录")
	}

	mod := dao.UsdtBalanceLog.Ctx(ctx).Where("user_id", user.Id)

	// 类型筛选
	if in.Type > 0 {
		mod = mod.Where("type", in.Type)
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



