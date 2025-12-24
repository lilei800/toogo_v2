// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogo

import (
	"context"
	"fmt"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type sToogoCommission struct{}

func NewToogoCommission() *sToogoCommission {
	return &sToogoCommission{}
}

func init() {
	service.RegisterToogoCommission(NewToogoCommission())
}

// CommissionLogList 佣金记录列表
func (s *sToogoCommission) CommissionLogList(ctx context.Context, in *toogoin.CommissionLogListInp) (list []*toogoin.CommissionLogListModel, totalCount int, err error) {
	mod := dao.ToogoCommissionLog.Ctx(ctx)
	cols := dao.ToogoCommissionLog.Columns()

	if in.UserId > 0 {
		mod = mod.Where(cols.UserId, in.UserId)
	}
	if in.CommissionType != "" {
		mod = mod.Where(cols.CommissionType, in.CommissionType)
	}
	if in.Level > 0 {
		mod = mod.Where(cols.Level, in.Level)
	}
	if len(in.CreatedAt) == 2 {
		mod = mod.WhereBetween(cols.CreatedAt, in.CreatedAt[0], in.CreatedAt[1])
	}

	var logs []*entity.ToogoCommissionLog
	err = mod.OrderDesc(cols.Id).Page(in.Page, in.PerPage).ScanAndCount(&logs, &totalCount, true)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "获取佣金记录失败")
	}

	for _, log := range logs {
		item := &toogoin.CommissionLogListModel{
			ToogoCommissionLog: log,
		}

		// 获取来源用户名
		var member *entity.AdminMember
		dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, log.FromUserId).Scan(&member)
		if member != nil {
			item.FromUsername = member.Username
		}

		list = append(list, item)
	}
	return
}

// CommissionStat 佣金统计
func (s *sToogoCommission) CommissionStat(ctx context.Context, in *toogoin.CommissionStatInp) (res *toogoin.CommissionStatModel, err error) {
	res = &toogoin.CommissionStatModel{}
	cols := dao.ToogoCommissionLog.Columns()

	// 累计佣金（仅USDT佣金）
	total, _ := dao.ToogoCommissionLog.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Where(cols.SettleType, "usdt").
		Sum(cols.CommissionAmount)
	res.TotalCommission = total

	// 今日佣金（仅USDT佣金）
	today := gtime.Now().Format("Y-m-d")
	todayTotal, _ := dao.ToogoCommissionLog.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Where(cols.SettleType, "usdt").
		WhereGTE(cols.CreatedAt, today+" 00:00:00").
		Sum(cols.CommissionAmount)
	res.TodayCommission = todayTotal

	// 本周佣金（仅USDT佣金）
	weekStart := gtime.Now().StartOfWeek().Format("Y-m-d H:i:s")
	weekTotal, _ := dao.ToogoCommissionLog.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Where(cols.SettleType, "usdt").
		WhereGTE(cols.CreatedAt, weekStart).
		Sum(cols.CommissionAmount)
	res.WeekCommission = weekTotal

	// 本月佣金（仅USDT佣金）
	monthStart := gtime.Now().StartOfMonth().Format("Y-m-d H:i:s")
	monthTotal, _ := dao.ToogoCommissionLog.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Where(cols.SettleType, "usdt").
		WhereGTE(cols.CreatedAt, monthStart).
		Sum(cols.CommissionAmount)
	res.MonthCommission = monthTotal

	// 分类统计
	inviteTotal, _ := dao.ToogoCommissionLog.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Where(cols.CommissionType, "invite_reward").
		Sum(cols.CommissionAmount)
	res.InviteReward = inviteTotal

	subscribeTotal, _ := dao.ToogoCommissionLog.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Where(cols.CommissionType, "subscribe").
		Where(cols.SettleType, "usdt").
		Sum(cols.CommissionAmount)
	res.SubscribeCommission = subscribeTotal

	return
}

// AgentWithRate 代理及其佣金比例（用于级差计算）
type AgentWithRate struct {
	UserId           int64
	SubscribeRate    float64
	IsAgent          int
	AgentUnlockLevel int // 层级解锁: 0=仅一级佣金, 1=无限级佣金
}

// GetAgentChainWithRates 获取完整的代理链及佣金比例（从消费者往上）
func (s *sToogoCommission) GetAgentChainWithRates(ctx context.Context, userId int64) []*AgentWithRate {
	var result []*AgentWithRate
	currentId := userId
	visited := make(map[int64]bool) // 防止循环引用

	for {
		if visited[currentId] {
			break
		}
		visited[currentId] = true

		var user *entity.ToogoUser
		err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, currentId).Scan(&user)
		if err != nil || user == nil || user.InviterId == 0 {
			break
		}

		// 获取上级信息
		var inviter *entity.ToogoUser
		err = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, user.InviterId).Scan(&inviter)
		if err != nil || inviter == nil {
			break
		}

		result = append(result, &AgentWithRate{
			UserId:           inviter.MemberId,
			SubscribeRate:    inviter.SubscribeRate,
			IsAgent:          inviter.IsAgent,
			AgentUnlockLevel: inviter.AgentUnlockLevel,
		})

		currentId = inviter.MemberId
	}

	return result
}

// SettleSubscribeCommission 结算订阅佣金（级差制）
// 规则：
// - 未解锁层级(AgentUnlockLevel=0)：只能获得一级佣金（直推下级的消费）
// - 已解锁层级(AgentUnlockLevel=1)：可获得无限级佣金（按级差计算）
func (s *sToogoCommission) SettleSubscribeCommission(ctx context.Context, fromUserId int64, amount float64, subscriptionId int64, orderSn string) error {
	// 获取完整代理链
	agentChain := s.GetAgentChainWithRates(ctx, fromUserId)
	if len(agentChain) == 0 {
		return nil
	}

	// 级差制计算：每个代理获得 (自己比例 - 下级比例) × 金额
	// 下级比例指的是紧邻的下一个代理的比例（离消费者更近的）
	prevRate := 0.0 // 下级比例，初始为0（消费者本身没有佣金比例）

	for level, agent := range agentChain {
		// 只有代理才能获得佣金
		if agent.IsAgent != 1 || agent.SubscribeRate <= 0 {
			// 非代理或无比例，跳过但更新prevRate
			continue
		}

		// 【核心逻辑】检查层级限制：
		// - level=0 表示一级（直推下级），所有代理都可获得
		// - level>0 表示二级及以上，需要解锁层级才能获得
		if level > 0 && agent.AgentUnlockLevel == 0 {
			// 未解锁层级，只能获得一级佣金，跳过更深层级
			g.Log().Debugf(ctx, "[SettleSubscribeCommission] 代理 %d 未解锁层级，跳过第%d级佣金", agent.UserId, level+1)
			// 仍需更新prevRate以便上级正确计算级差
			prevRate = agent.SubscribeRate
			continue
		}

		// 计算级差
		rateDiff := agent.SubscribeRate - prevRate
		if rateDiff <= 0 {
			// 没有级差，跳过
			prevRate = agent.SubscribeRate // 更新为当前比例供上级计算
			continue
		}

		// 计算佣金
		commissionAmount := amount * (rateDiff / 100)

		// 记录佣金
		err := s.AddCommission(ctx, agent.UserId, fromUserId, "subscribe", level+1, amount, rateDiff, commissionAmount, subscriptionId, "subscription", orderSn)
		if err != nil {
			g.Log().Warningf(ctx, "记录订阅佣金失败: %v", err)
		}

		// 更新下级比例为当前代理的比例
		prevRate = agent.SubscribeRate
	}

	return nil
}

// SettleInviteReward 发放邀请奖励（积分）
func (s *sToogoCommission) SettleInviteReward(ctx context.Context, inviterId int64, inviteeId int64) error {
	// 获取邀请奖励积分配置
	rewardPower := 30.0 // 默认30积分

	// 获取邀请人VIP等级对应的奖励
	var inviter *entity.ToogoUser
	dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, inviterId).Scan(&inviter)
	if inviter != nil {
		var vipLevel *entity.ToogoVipLevel
		dao.ToogoVipLevel.Ctx(ctx).Where(dao.ToogoVipLevel.Columns().Level, inviter.VipLevel).Scan(&vipLevel)
		if vipLevel != nil && vipLevel.InviteRewardPower > 0 {
			rewardPower = vipLevel.InviteRewardPower
		}
	}

	// 给邀请人发放积分
	err := service.ToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
		UserId:      inviterId,
		AccountType: "gift_power",
		ChangeType:  "invite_reward",
		Amount:      rewardPower,
		RelatedId:   inviteeId,
		RelatedType: "invite",
		Remark:      "邀请新用户奖励积分",
	})
	if err != nil {
		g.Log().Warningf(ctx, "发放邀请人奖励失败: %v", err)
	}

	// 给被邀请人发放积分
	err = service.ToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
		UserId:      inviteeId,
		AccountType: "gift_power",
		ChangeType:  "invited_reward",
		Amount:      rewardPower,
		RelatedId:   inviterId,
		RelatedType: "invite",
		Remark:      "注册奖励积分",
	})
	if err != nil {
		g.Log().Warningf(ctx, "发放被邀请人奖励失败: %v", err)
	}

	// 记录佣金日志 (邀请人)
	s.AddCommission(ctx, inviterId, inviteeId, "invite_reward", 1, 0, 0, rewardPower, inviteeId, "user", "")

	return nil
}

// AddCommission 添加佣金记录
func (s *sToogoCommission) AddCommission(ctx context.Context, userId, fromUserId int64, commissionType string, level int, baseAmount, rate, amount float64, relatedId int64, relatedType, orderSn string) error {
	log := &entity.ToogoCommissionLog{
		UserId:           userId,
		FromUserId:       fromUserId,
		CommissionType:   commissionType,
		Level:            level,
		BaseAmount:       baseAmount,
		CommissionRate:   rate,
		CommissionAmount: amount,
		SettleType:       "usdt",
		Status:           2, // 已结算
		RelatedId:        relatedId,
		RelatedType:      relatedType,
		OrderSn:          orderSn,
		CreatedAt:        gtime.Now(),
	}

	// 邀请奖励结算为积分
	if commissionType == "invite_reward" {
		log.SettleType = "power"
	}

	_, err := dao.ToogoCommissionLog.Ctx(ctx).Data(log).Insert()
	if err != nil {
		return gerror.Wrap(err, "添加佣金记录失败")
	}

	// 如果是USDT佣金，增加佣金账户余额
	if log.SettleType == "usdt" && amount > 0 {
		err = service.ToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
			UserId:      userId,
			AccountType: "commission",
			ChangeType:  commissionType,
			Amount:      amount,
			RelatedId:   relatedId,
			RelatedType: relatedType,
			OrderSn:     orderSn,
			Remark:      fmt.Sprintf("级差佣金(级差%.2f%%)", rate),
		})
		if err != nil {
			g.Log().Warningf(ctx, "增加佣金余额失败: %v", err)
		}

		// 更新累计佣金
		dao.ToogoWallet.Ctx(ctx).
			Where(dao.ToogoWallet.Columns().UserId, userId).
			Increment(dao.ToogoWallet.Columns().TotalCommission, amount)
	}

	return nil
}

// ApplyAgent 申请成为代理商
// 规则：用户提交申请后，状态变为"待审批"，需要管理员在后台手动审批通过
func (s *sToogoCommission) ApplyAgent(ctx context.Context, in *toogoin.ApplyAgentInp) (res *toogoin.ApplyAgentModel, err error) {
	// 获取用户信息
	toogoUser, err := service.ToogoUser().GetOrCreate(ctx, in.MemberId)
	if err != nil {
		return nil, err
	}

	// 已经是代理商
	if toogoUser.IsAgent == 1 && toogoUser.AgentStatus == 2 {
		res = &toogoin.ApplyAgentModel{
			Success:       true,
			AgentStatus:   2,
			Message:       "您已经是代理商",
			SubscribeRate: toogoUser.SubscribeRate,
		}
		return
	}

	// 检查当前申请状态
	if toogoUser.AgentStatus == 1 {
		// 已经在待审批状态
		res = &toogoin.ApplyAgentModel{
			Success:     false,
			AgentStatus: 1,
			Message:     "您的代理商申请正在审核中，请耐心等待",
		}
		return
	}

	if toogoUser.AgentStatus == 3 {
		// 之前被拒绝，允许重新申请
		g.Log().Infof(ctx, "[ApplyAgent] 用户 %d 之前被拒绝，重新提交申请", in.MemberId)
	}

	// 校验申请比例
	if in.SubscribeRate <= 0 || in.SubscribeRate > 100 {
		return nil, gerror.New("申请订阅返佣比例必须在0-100之间")
	}

	// 如有上级代理，则申请比例不能超过上级代理比例
	if toogoUser.InviterId > 0 {
		var inviter *entity.ToogoUser
		_ = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, toogoUser.InviterId).Scan(&inviter)
		if inviter != nil && inviter.IsAgent == 1 && inviter.AgentStatus == 2 && inviter.SubscribeRate > 0 {
			if in.SubscribeRate > inviter.SubscribeRate {
				return nil, gerror.Newf("申请订阅返佣比例不能超过上级代理比例 %.2f%%", inviter.SubscribeRate)
			}
		}
	}

	// 更新为待审批状态
	_, err = dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().MemberId, in.MemberId).
		Data(g.Map{
			dao.ToogoUser.Columns().AgentStatus:      1, // 待审批
			dao.ToogoUser.Columns().AgentApplyAt:     gtime.Now(),
			dao.ToogoUser.Columns().AgentApplyRemark: in.Remark,
			dao.ToogoUser.Columns().SubscribeRate:    in.SubscribeRate, // 记录申请比例（管理员可在审批时调整）
		}).
		Update()
	if err != nil {
		return nil, gerror.Wrap(err, "提交代理商申请失败")
	}

	res = &toogoin.ApplyAgentModel{
		Success:     true,
		AgentStatus: 1,
		Message:     "申请已提交，请等待管理员审核",
	}
	return
}

// ApplyAgentForSub 代下级提交代理申请（仅直属下级）
func (s *sToogoCommission) ApplyAgentForSub(ctx context.Context, in *toogoin.ApplyAgentForSubInp) error {
	// 当前用户必须是已通过的代理
	var agent *entity.ToogoUser
	err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, in.AgentId).Scan(&agent)
	if err != nil || agent == nil {
		return gerror.New("当前用户不存在")
	}
	if agent.IsAgent != 1 || agent.AgentStatus != 2 {
		return gerror.New("您尚未成为代理，无法代下级申请")
	}

	// 校验申请比例
	if in.SubscribeRate <= 0 || in.SubscribeRate > 100 {
		return gerror.New("申请订阅返佣比例必须在0-100之间")
	}
	if in.SubscribeRate > agent.SubscribeRate {
		return gerror.Newf("申请订阅返佣比例不能超过您的比例 %.2f%%", agent.SubscribeRate)
	}

	// 下级必须存在且为直属
	var sub *entity.ToogoUser
	err = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, in.SubUserId).Scan(&sub)
	if err != nil || sub == nil {
		return gerror.New("下级用户不存在")
	}
	if sub.InviterId != in.AgentId {
		return gerror.New("只能为直属下级提交申请")
	}
	if sub.IsAgent == 1 && sub.AgentStatus == 2 {
		return gerror.New("该用户已是代理，无需申请")
	}
	if sub.AgentStatus == 1 {
		return gerror.New("该用户已有待审批的代理申请")
	}

	_, err = dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().MemberId, in.SubUserId).
		Data(g.Map{
			dao.ToogoUser.Columns().AgentStatus:      1,
			dao.ToogoUser.Columns().AgentApplyAt:     gtime.Now(),
			dao.ToogoUser.Columns().AgentApplyRemark: in.Remark,
			dao.ToogoUser.Columns().SubscribeRate:    in.SubscribeRate, // 记录申请比例
		}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "提交下级代理申请失败")
	}

	g.Log().Infof(ctx, "[ApplyAgentForSub] 代理 %d 为下级 %d 提交申请: 订阅比例=%.2f%%", in.AgentId, in.SubUserId, in.SubscribeRate)
	return nil
}

// SetSubAgentRate 设置下级代理的佣金比例
// 规则：必须解锁层级后才能设置下级代理的佣金比例
func (s *sToogoCommission) SetSubAgentRate(ctx context.Context, in *toogoin.SetSubAgentRateInp) error {
	// 获取当前代理信息
	var currentAgent *entity.ToogoUser
	err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, in.AgentId).Scan(&currentAgent)
	if err != nil || currentAgent == nil {
		return gerror.New("代理商不存在")
	}
	if currentAgent.IsAgent != 1 {
		return gerror.New("您不是代理商，无法设置下级比例")
	}

	// 【核心逻辑】检查是否已解锁层级
	if currentAgent.AgentUnlockLevel == 0 {
		return gerror.New("您尚未解锁层级权限，无法设置下级代理的佣金比例。请联系管理员开通")
	}

	// 获取下级用户信息
	var subUser *entity.ToogoUser
	err = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, in.SubUserId).Scan(&subUser)
	if err != nil || subUser == nil {
		return gerror.New("下级用户不存在")
	}

	// 验证是否是直属下级
	if subUser.InviterId != in.AgentId {
		return gerror.New("只能设置直属下级的佣金比例")
	}

	// 下级必须已通过代理（不能绕过管理员审批）
	if subUser.IsAgent != 1 || subUser.AgentStatus != 2 {
		return gerror.New("下级尚未成为代理，无法设置佣金比例")
	}

	// 验证比例不能超过自己的比例
	if in.SubscribeRate > currentAgent.SubscribeRate {
		return gerror.Newf("订阅返佣比例不能超过您的比例 %.2f%%", currentAgent.SubscribeRate)
	}

	// 验证比例范围
	if in.SubscribeRate < 0 || in.SubscribeRate > 100 {
		return gerror.New("订阅返佣比例必须在0-100之间")
	}

	// 更新下级用户的订阅返佣比例（仅更新比例，不改变代理状态）
	_, err = dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().MemberId, in.SubUserId).
		Data(g.Map{
			dao.ToogoUser.Columns().SubscribeRate: in.SubscribeRate,
		}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "设置下级佣金比例失败")
	}

	g.Log().Infof(ctx, "[SetSubAgentRate] 代理 %d 设置下级 %d 的比例: 订阅=%.2f%%",
		in.AgentId, in.SubUserId, in.SubscribeRate)

	return nil
}

// GetAgentInfo 获取代理信息
func (s *sToogoCommission) GetAgentInfo(ctx context.Context, memberId int64) (*toogoin.AgentInfoModel, error) {
	var user *entity.ToogoUser
	err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, memberId).Scan(&user)
	if err != nil || user == nil {
		return nil, gerror.New("用户不存在")
	}

	// 获取下级列表
	var subUsers []*entity.ToogoUser
	dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().InviterId, memberId).
		Scan(&subUsers)

	subList := make([]*toogoin.SubAgentInfo, 0, len(subUsers))
	for _, sub := range subUsers {
		// 获取用户名
		var member *entity.AdminMember
		dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, sub.MemberId).Scan(&member)
		username := ""
		if member != nil {
			username = member.Username
		}

		subList = append(subList, &toogoin.SubAgentInfo{
			UserId:        sub.MemberId,
			Username:      username,
			IsAgent:       sub.IsAgent,
			SubscribeRate: sub.SubscribeRate,
		})
	}

	// 判断是否可以设置下级佣金比例（需要是代理商且已解锁层级）
	canSetSubRate := user.IsAgent == 1 && user.AgentUnlockLevel == 1

	return &toogoin.AgentInfoModel{
		IsAgent:          user.IsAgent,
		AgentStatus:      user.AgentStatus,
		AgentUnlockLevel: user.AgentUnlockLevel,
		SubscribeRate:    user.SubscribeRate,
		CanSetSubRate:    canSetSubRate,
		SubAgents:        subList,
	}, nil
}

// ApproveAgent 审批代理商申请（管理员使用）
func (s *sToogoCommission) ApproveAgent(ctx context.Context, in *toogoin.ApproveAgentInp, operatorId int64) (*toogoin.ApproveAgentModel, error) {
	// 获取用户信息
	var user *entity.ToogoUser
	err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, in.MemberId).Scan(&user)
	if err != nil || user == nil {
		return nil, gerror.New("用户不存在")
	}

	// 检查申请状态
	if user.AgentStatus != 1 {
		return nil, gerror.New("该用户没有待审批的代理商申请")
	}

	if in.Approved {
		// 审批通过
		if in.SubscribeRate <= 0 || in.SubscribeRate > 100 {
			return nil, gerror.New("订阅返佣比例必须在0-100之间")
		}

		_, err = dao.ToogoUser.Ctx(ctx).
			Where(dao.ToogoUser.Columns().MemberId, in.MemberId).
			Data(g.Map{
				dao.ToogoUser.Columns().IsAgent:           1,
				dao.ToogoUser.Columns().AgentStatus:       2, // 已通过
				dao.ToogoUser.Columns().AgentUnlockLevel:  in.AgentUnlockLevel,
				dao.ToogoUser.Columns().SubscribeRate:     in.SubscribeRate,
				dao.ToogoUser.Columns().AgentApprovedAt:   gtime.Now(),
				dao.ToogoUser.Columns().AgentApprovedBy:   operatorId,
			}).
			Update()
		if err != nil {
			return nil, gerror.Wrap(err, "审批代理商失败")
		}

		g.Log().Infof(ctx, "[ApproveAgent] 管理员 %d 审批通过用户 %d 的代理商申请, 订阅比例=%.2f%%, 解锁层级=%d",
			operatorId, in.MemberId, in.SubscribeRate, in.AgentUnlockLevel)

		return &toogoin.ApproveAgentModel{
			Success: true,
			Message: "审批通过，用户已成为代理商",
		}, nil
	} else {
		// 审批拒绝
		_, err = dao.ToogoUser.Ctx(ctx).
			Where(dao.ToogoUser.Columns().MemberId, in.MemberId).
			Data(g.Map{
				dao.ToogoUser.Columns().AgentStatus:     3, // 已拒绝
				dao.ToogoUser.Columns().AgentApprovedAt: gtime.Now(),
				dao.ToogoUser.Columns().AgentApprovedBy: operatorId,
			}).
			Update()
		if err != nil {
			return nil, gerror.Wrap(err, "拒绝代理商申请失败")
		}

		g.Log().Infof(ctx, "[ApproveAgent] 管理员 %d 拒绝用户 %d 的代理商申请, 原因: %s",
			operatorId, in.MemberId, in.RejectReason)

		return &toogoin.ApproveAgentModel{
			Success: true,
			Message: "已拒绝该用户的代理商申请",
		}, nil
	}
}

// UpdateAgent 更新代理商信息（管理员使用）
func (s *sToogoCommission) UpdateAgent(ctx context.Context, in *toogoin.UpdateAgentInp) error {
	// 获取用户信息
	var user *entity.ToogoUser
	err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, in.MemberId).Scan(&user)
	if err != nil || user == nil {
		return gerror.New("用户不存在")
	}

	// 验证比例范围
	if in.SubscribeRate < 0 || in.SubscribeRate > 100 {
		return gerror.New("订阅返佣比例必须在0-100之间")
	}

	updateData := g.Map{
		dao.ToogoUser.Columns().IsAgent:          in.IsAgent,
		dao.ToogoUser.Columns().AgentStatus:      in.AgentStatus,
		dao.ToogoUser.Columns().AgentUnlockLevel: in.AgentUnlockLevel,
		dao.ToogoUser.Columns().SubscribeRate:    in.SubscribeRate,
	}

	// 如果设为代理商且状态为已通过，更新审批时间
	if in.IsAgent == 1 && in.AgentStatus == 2 && user.AgentApprovedAt == nil {
		updateData[dao.ToogoUser.Columns().AgentApprovedAt] = gtime.Now()
	}

	_, err = dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().MemberId, in.MemberId).
		Data(updateData).
		Update()
	if err != nil {
		return gerror.Wrap(err, "更新代理商信息失败")
	}

	g.Log().Infof(ctx, "[UpdateAgent] 更新用户 %d 的代理商信息: isAgent=%d, status=%d, unlockLevel=%d, subscribeRate=%.2f%%",
		in.MemberId, in.IsAgent, in.AgentStatus, in.AgentUnlockLevel, in.SubscribeRate)

	return nil
}

// AgentLevelList 代理商等级列表（已废弃，保留兼容）
func (s *sToogoCommission) AgentLevelList(ctx context.Context, in *toogoin.AgentLevelListInp) (list []*toogoin.AgentLevelListModel, totalCount int, err error) {
	mod := dao.ToogoAgentLevel.Ctx(ctx)

	if in.Status > 0 {
		mod = mod.Where(dao.ToogoAgentLevel.Columns().Status, in.Status)
	}

	err = mod.OrderAsc(dao.ToogoAgentLevel.Columns().Level).Page(in.Page, in.PerPage).ScanAndCount(&list, &totalCount, true)
	if err != nil {
		err = gerror.Wrap(err, "获取代理商等级列表失败")
	}
	return
}

// AgentLevelEdit 编辑代理商等级（已废弃，保留兼容）
func (s *sToogoCommission) AgentLevelEdit(ctx context.Context, in *toogoin.AgentLevelEditInp) (err error) {
	cols := dao.ToogoAgentLevel.Columns()

	data := g.Map{
		cols.Level:                in.Level,
		cols.LevelName:            in.LevelName,
		cols.RequireTeamCount:     in.RequireTeamCount,
		cols.RequireTeamSubscribe: in.RequireTeamSubscribe,
		cols.SubscribeRate1:       in.SubscribeRate1,
		cols.SubscribeRate2:       in.SubscribeRate2,
		cols.SubscribeRate3:       in.SubscribeRate3,
		cols.PowerRate1:           in.PowerRate1,
		cols.PowerRate2:           in.PowerRate2,
		cols.PowerRate3:           in.PowerRate3,
		cols.Description:          in.Description,
		cols.Sort:                 in.Sort,
		cols.Status:               in.Status,
	}

	if in.Id > 0 {
		_, err = dao.ToogoAgentLevel.Ctx(ctx).Where(dao.ToogoAgentLevel.Columns().Id, in.Id).Data(data).Update()
	} else {
		_, err = dao.ToogoAgentLevel.Ctx(ctx).Data(data).Insert()
	}

	if err != nil {
		err = gerror.Wrap(err, "保存代理商等级失败")
	}
	return
}
