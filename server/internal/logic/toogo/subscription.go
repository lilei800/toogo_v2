// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogo

import (
	"context"
	"fmt"
	"time"
	
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"
)

type sToogoSubscription struct{}

func NewToogoSubscription() *sToogoSubscription {
	return &sToogoSubscription{}
}

func init() {
	service.RegisterToogoSubscription(NewToogoSubscription())
}

// PlanList 套餐列表
func (s *sToogoSubscription) PlanList(ctx context.Context, in *toogoin.PlanListInp) (list []*toogoin.PlanListModel, totalCount int, err error) {
	mod := dao.ToogoPlan.Ctx(ctx)

	if in.Status > 0 {
		mod = mod.Where(dao.ToogoPlan.Columns().Status, in.Status)
	}

	err = mod.OrderAsc(dao.ToogoPlan.Columns().Sort).Page(in.Page, in.PerPage).ScanAndCount(&list, &totalCount, true)
	if err != nil {
		err = gerror.Wrap(err, "获取套餐列表失败")
	}
	return
}

// PlanEdit 编辑套餐
func (s *sToogoSubscription) PlanEdit(ctx context.Context, in *toogoin.PlanEditInp) (err error) {
	cols := dao.ToogoPlan.Columns()

	data := g.Map{
		cols.PlanName:           in.PlanName,
		cols.PlanCode:           in.PlanCode,
		cols.RobotLimit:         in.RobotLimit,
		cols.PriceDaily:         in.PriceDaily,
		cols.PriceMonthly:       in.PriceMonthly,
		cols.PriceQuarterly:     in.PriceQuarterly,
		cols.PriceHalfYear:      in.PriceHalfYear,
		cols.PriceYearly:        in.PriceYearly,
		cols.DefaultPeriod:      in.DefaultPeriod,
		cols.PurchaseLimit:      in.PurchaseLimit,
		cols.PurchaseLimitDaily: in.PurchaseLimitDaily,
		cols.PurchaseLimitMonthly: in.PurchaseLimitMonthly,
		cols.PurchaseLimitQuarterly: in.PurchaseLimitQuarterly,
		cols.PurchaseLimitHalfYear: in.PurchaseLimitHalfYear,
		cols.PurchaseLimitYearly: in.PurchaseLimitYearly,
		cols.GiftPowerMonthly:   in.GiftPowerMonthly,
		cols.GiftPowerQuarterly: in.GiftPowerQuarterly,
		cols.GiftPowerHalfYear:  in.GiftPowerHalfYear,
		cols.GiftPowerYearly:    in.GiftPowerYearly,
		cols.Description:        in.Description,
		cols.Features:           in.Features,
		cols.IsDefault:          in.IsDefault,
		cols.Sort:               in.Sort,
		cols.Status:             in.Status,
	}

	if in.Id > 0 {
		_, err = dao.ToogoPlan.Ctx(ctx).Where(dao.ToogoPlan.Columns().Id, in.Id).Data(data).Update()
	} else {
		_, err = dao.ToogoPlan.Ctx(ctx).Data(data).Insert()
	}

	if err != nil {
		err = gerror.Wrap(err, "保存套餐失败")
	}
	return
}

// PlanDelete 删除套餐
func (s *sToogoSubscription) PlanDelete(ctx context.Context, in *toogoin.PlanDeleteInp) (err error) {
	// 检查是否有用户正在使用
	count, err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().CurrentPlanId, in.Id).Count()
	if err != nil {
		return gerror.Wrap(err, "检查套餐使用情况失败")
	}
	if count > 0 {
		return gerror.Newf("有%d个用户正在使用该套餐，无法删除", count)
	}

	_, err = dao.ToogoPlan.Ctx(ctx).Where(dao.ToogoPlan.Columns().Id, in.Id).Delete()
	if err != nil {
		err = gerror.Wrap(err, "删除套餐失败")
	}
	return
}

// Subscribe 订阅套餐（支持积分抵扣）
func (s *sToogoSubscription) Subscribe(ctx context.Context, in *toogoin.SubscribeInp) (res *toogoin.SubscribeModel, err error) {
	// 获取套餐信息
	var plan *entity.ToogoPlan
	err = dao.ToogoPlan.Ctx(ctx).Where(dao.ToogoPlan.Columns().Id, in.PlanId).Scan(&plan)
	if err != nil || plan == nil {
		return nil, gerror.New("套餐不存在")
	}
	if plan.Status != 1 {
		return nil, gerror.New("套餐已下架")
	}

	// 计算价格和天数（不再赠送积分）
	var amount float64
	var days int

	switch in.PeriodType {
	case "daily":
		amount = plan.PriceDaily
		days = 1
	case "monthly":
		amount = plan.PriceMonthly
		days = 30
	case "quarterly":
		amount = plan.PriceQuarterly
		days = 90
	case "half_year":
		amount = plan.PriceHalfYear
		days = 180
	case "yearly":
		amount = plan.PriceYearly
		days = 365
	default:
		return nil, gerror.New("无效的订阅周期")
	}

	// 购买次数限制校验（套餐级别：按“用户 + 套餐”限购；0 表示不限）
	// 注意：这里不做前端判断，统一由后端强制。
	if plan.PurchaseLimit > 0 {
		subCols := dao.ToogoSubscription.Columns()
		// 统计该用户购买该套餐的次数（包含：待支付/生效中/已过期；不含已取消）
		count, err := dao.ToogoSubscription.Ctx(ctx).
			Where(subCols.UserId, in.UserId).
			Where(subCols.PlanId, plan.Id).
			WhereIn(subCols.Status, []int{1, 2, 3}).
			Count()
		if err != nil {
			return nil, gerror.Wrap(err, "查询套餐购买次数限制失败")
		}
		if count >= plan.PurchaseLimit {
			return nil, gerror.Newf("该套餐已达到限购次数（%d次），请更换套餐后再试", plan.PurchaseLimit)
		}
	}

	// 获取用户信息
	toogoUser, err := service.ToogoUser().GetOrCreate(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	// 获取用户钱包
	wallet, err := service.ToogoWallet().GetOrCreate(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	// 计算积分抵扣金额（积分可抵扣订阅费用，1积分=1USDT）
	var pointsDeduct float64 = 0
	var balanceDeduct float64 = amount
	usePoints := in.UsePoints // 是否使用积分抵扣

	if usePoints && wallet.GiftPower > 0 && amount > 0 {
		if wallet.GiftPower >= amount {
			// 积分足够，全额抵扣
			pointsDeduct = amount
			balanceDeduct = 0
		} else {
			// 积分不足，部分抵扣
			pointsDeduct = wallet.GiftPower
			balanceDeduct = amount - pointsDeduct
		}
	}

	// 检查余额是否足够支付剩余金额
	if in.PayType == "balance" && balanceDeduct > 0 {
		if wallet.Balance < balanceDeduct {
			return nil, gerror.Newf("余额不足，需支付 %.2f USDT（已抵扣积分 %.2f）", balanceDeduct, pointsDeduct)
		}
	}

	// 生成订单号
	orderSn := fmt.Sprintf("SUB%s%s", gtime.Now().Format("YmdHis"), grand.S(6))

	// 计算开始和结束时间
	startTime := gtime.Now()
	// 如果当前有有效订阅，从过期时间开始
	if toogoUser.PlanExpireTime != nil && toogoUser.PlanExpireTime.After(gtime.Now()) {
		startTime = toogoUser.PlanExpireTime
	}
	expireTime := startTime.Add(time.Duration(days) * 24 * time.Hour)

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 创建订阅记录（不再记录赠送积分）
		subscription := &entity.ToogoSubscription{
			UserId:     in.UserId,
			PlanId:     plan.Id,
			PlanCode:   plan.PlanCode,
			OrderSn:    orderSn,
			PeriodType: in.PeriodType,
			Amount:     amount,
			GiftPower:  0, // 不再赠送积分
			StartTime:  startTime,
			ExpireTime: expireTime,
			Days:       days,
			Status:     1, // 待支付
			InviterId:  toogoUser.InviterId,
			CreatedAt:  gtime.Now(),
			UpdatedAt:  gtime.Now(),
		}

		subscriptionId, err := dao.ToogoSubscription.Ctx(ctx).Data(subscription).InsertAndGetId()
		if err != nil {
			return gerror.Wrap(err, "创建订阅记录失败")
		}

		// 如果是余额支付
		if in.PayType == "balance" {
			// 先扣除积分（如果使用积分抵扣）
			if pointsDeduct > 0 {
				err = service.ToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
					UserId:      in.UserId,
					AccountType: "gift_power",
					ChangeType:  "subscribe_deduct",
					Amount:      -pointsDeduct,
					RelatedId:   subscriptionId,
					RelatedType: "subscription",
					OrderSn:     orderSn,
					Remark:      fmt.Sprintf("订阅%s套餐积分抵扣", plan.PlanName),
				})
				if err != nil {
					return gerror.Wrap(err, "扣除积分失败")
				}
			}

			// 再扣除余额（如果还有剩余金额）
			if balanceDeduct > 0 {
				err = service.ToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
					UserId:      in.UserId,
					AccountType: "balance",
					ChangeType:  "subscribe",
					Amount:      -balanceDeduct,
					RelatedId:   subscriptionId,
					RelatedType: "subscription",
					OrderSn:     orderSn,
					Remark:      fmt.Sprintf("订阅%s套餐", plan.PlanName),
				})
				if err != nil {
					return err
				}
			}

			// 更新订阅状态为生效中
			_, err = dao.ToogoSubscription.Ctx(ctx).Where(dao.ToogoSubscription.Columns().Id, subscriptionId).Data(g.Map{
				dao.ToogoSubscription.Columns().Status:  2,
				dao.ToogoSubscription.Columns().PaidAt:  gtime.Now(),
				dao.ToogoSubscription.Columns().PayType: in.PayType,
			}).Update()
			if err != nil {
				return gerror.Wrap(err, "更新订阅状态失败")
			}

			// 更新用户套餐信息
			_, err = dao.ToogoUser.Ctx(ctx).
				Where(dao.ToogoUser.Columns().MemberId, in.UserId).
				Data(g.Map{
					dao.ToogoUser.Columns().CurrentPlanId:  plan.Id,
					dao.ToogoUser.Columns().PlanExpireTime: expireTime,
					dao.ToogoUser.Columns().RobotLimit:     plan.RobotLimit,
				}).
				Update()
			if err != nil {
				return gerror.Wrap(err, "更新用户套餐信息失败")
			}

			// 结算订阅佣金（只针对实际支付的余额部分）
			if balanceDeduct > 0 {
				err = service.ToogoCommission().SettleSubscribeCommission(ctx, in.UserId, balanceDeduct, subscriptionId, orderSn)
				if err != nil {
					g.Log().Warningf(ctx, "结算订阅佣金失败: %v", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	res = &toogoin.SubscribeModel{
		OrderSn:      orderSn,
		PlanName:     plan.PlanName,
		Amount:       amount,
		PointsDeduct: pointsDeduct,  // 积分抵扣金额
		BalancePaid:  balanceDeduct, // 余额支付金额
		Days:         days,
		ExpireTime:   expireTime.String(),
	}
	return
}

// SubscriptionList 订阅记录列表
func (s *sToogoSubscription) SubscriptionList(ctx context.Context, in *toogoin.SubscriptionListInp) (list []*toogoin.SubscriptionListModel, totalCount int, err error) {
	mod := dao.ToogoSubscription.Ctx(ctx)
	cols := dao.ToogoSubscription.Columns()

	if in.UserId > 0 {
		mod = mod.Where(cols.UserId, in.UserId)
	}
	if in.Status > 0 {
		mod = mod.Where(cols.Status, in.Status)
	}
	if len(in.CreatedAt) == 2 {
		mod = mod.WhereBetween(cols.CreatedAt, in.CreatedAt[0], in.CreatedAt[1])
	}

	var subscriptions []*entity.ToogoSubscription
	err = mod.OrderDesc(cols.Id).Page(in.Page, in.PerPage).ScanAndCount(&subscriptions, &totalCount, true)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "获取订阅记录失败")
	}

	for _, sub := range subscriptions {
		item := &toogoin.SubscriptionListModel{
			ToogoSubscription: sub,
		}

		// 获取套餐名称
		var plan *entity.ToogoPlan
		dao.ToogoPlan.Ctx(ctx).Where(dao.ToogoPlan.Columns().Id, sub.PlanId).Scan(&plan)
		if plan != nil {
			item.PlanName = plan.PlanName
		}

		list = append(list, item)
	}
	return
}

// MySubscription 我的订阅
func (s *sToogoSubscription) MySubscription(ctx context.Context, in *toogoin.MySubscriptionInp) (res *toogoin.MySubscriptionModel, err error) {
	toogoUser, err := service.ToogoUser().GetOrCreate(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	res = &toogoin.MySubscriptionModel{
		HasSubscription: false,
	}

	// 检查是否有有效订阅
	if toogoUser.CurrentPlanId > 0 && toogoUser.PlanExpireTime != nil {
		if toogoUser.PlanExpireTime.After(gtime.Now()) {
			res.HasSubscription = true
			res.PlanId = toogoUser.CurrentPlanId
			res.RobotLimit = toogoUser.RobotLimit
			res.ActiveRobots = toogoUser.ActiveRobotCount
			res.ExpireTime = toogoUser.PlanExpireTime.String()

			// 计算剩余天数
			remaining := toogoUser.PlanExpireTime.Sub(gtime.Now())
			res.RemainingDays = int(remaining.Hours() / 24)

			// 获取套餐信息
			var plan *entity.ToogoPlan
			dao.ToogoPlan.Ctx(ctx).Where(dao.ToogoPlan.Columns().Id, toogoUser.CurrentPlanId).Scan(&plan)
			if plan != nil {
				res.PlanName = plan.PlanName
				res.PlanCode = plan.PlanCode
			}
		}
	}

	return
}

// CheckExpired 检查并处理过期订阅 (定时任务调用)
func (s *sToogoSubscription) CheckExpired(ctx context.Context) error {
	// 查找已过期但状态还是生效中的订阅
	var subscriptions []*entity.ToogoSubscription
	err := dao.ToogoSubscription.Ctx(ctx).
		Where(dao.ToogoSubscription.Columns().Status, 2). // 生效中
		WhereLT(dao.ToogoSubscription.Columns().ExpireTime, gtime.Now()).
		Scan(&subscriptions)
	if err != nil {
		return gerror.Wrap(err, "查询过期订阅失败")
	}

	for _, sub := range subscriptions {
		// 更新订阅状态为已过期
		_, err = dao.ToogoSubscription.Ctx(ctx).Where(dao.ToogoSubscription.Columns().Id, sub.Id).Data(g.Map{
			dao.ToogoSubscription.Columns().Status: 3, // 已过期
		}).Update()
		if err != nil {
			g.Log().Warningf(ctx, "更新订阅状态失败: %v", err)
			continue
		}

		// 检查用户是否还有其他有效订阅
		count, err := dao.ToogoSubscription.Ctx(ctx).
			Where(dao.ToogoSubscription.Columns().UserId, sub.UserId).
			Where(dao.ToogoSubscription.Columns().Status, 2).
			WhereGT(dao.ToogoSubscription.Columns().ExpireTime, gtime.Now()).
			Count()
		if err != nil {
			g.Log().Warningf(ctx, "检查用户订阅失败: %v", err)
			continue
		}

		// 如果没有其他有效订阅，需要停止用户的所有运行中机器人并重置套餐
		if count == 0 {
			// 停止用户的所有运行中机器人（会自动平仓所有订单）
			var runningRobots []*entity.TradingRobot
			err = dao.TradingRobot.Ctx(ctx).
				Where(dao.TradingRobot.Columns().UserId, sub.UserId).
				Where(dao.TradingRobot.Columns().Status, 2). // 运行中
				WhereNull(dao.TradingRobot.Columns().DeletedAt).
				Scan(&runningRobots)
			if err != nil {
				g.Log().Warningf(ctx, "查询用户运行中机器人失败: %v", err)
			} else {
				// 停止所有运行中的机器人
				for _, robot := range runningRobots {
					// 使用系统上下文（不需要用户登录）
					systemCtx := context.Background()
					// 调用机器人任务管理器平仓并停止
					if err := GetRobotTaskManager().CloseAllAndWait(systemCtx, robot.Id, "subscription_expired", 30*time.Second); err != nil {
						g.Log().Warningf(ctx, "停止机器人失败 (robotId=%d): %v", robot.Id, err)
						continue
					}
					// 更新机器人状态为未启动
					_, err = dao.TradingRobot.Ctx(ctx).
						Where(dao.TradingRobot.Columns().Id, robot.Id).
						Data(g.Map{
							dao.TradingRobot.Columns().Status:   1, // 未启动
							dao.TradingRobot.Columns().StopTime: gtime.Now(),
						}).
						Update()
					if err != nil {
						g.Log().Warningf(ctx, "更新机器人状态失败 (robotId=%d): %v", robot.Id, err)
					} else {
						g.Log().Infof(ctx, "订阅到期自动停止机器人: userId=%d, robotId=%d, robotName=%s", sub.UserId, robot.Id, robot.RobotName)
					}
				}
			}

			// 获取免费套餐
			var freePlan *entity.ToogoPlan
			dao.ToogoPlan.Ctx(ctx).Where(dao.ToogoPlan.Columns().IsDefault, 1).Scan(&freePlan)

			robotLimit := 1
			if freePlan != nil {
				robotLimit = freePlan.RobotLimit
			}

			// 重置用户套餐为免费套餐
			_, err = dao.ToogoUser.Ctx(ctx).
				Where(dao.ToogoUser.Columns().MemberId, sub.UserId).
				Data(g.Map{
					dao.ToogoUser.Columns().CurrentPlanId:  0,
					dao.ToogoUser.Columns().PlanExpireTime: nil,
					dao.ToogoUser.Columns().RobotLimit:     robotLimit,
				}).
				Update()
			if err != nil {
				g.Log().Warningf(ctx, "重置用户套餐失败: %v", err)
			}
		}
	}

	g.Log().Infof(ctx, "处理过期订阅完成，共处理 %d 条", len(subscriptions))
	return nil
}

