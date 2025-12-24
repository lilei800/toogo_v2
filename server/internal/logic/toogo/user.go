// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogo

import (
	"context"
	"fmt"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/adminin"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
)

type sToogoUser struct{}

func NewToogoUser() *sToogoUser {
	return &sToogoUser{}
}

func init() {
	service.RegisterToogoUser(NewToogoUser())
}

// GetOrCreate 获取或创建Toogo用户扩展信息
func (s *sToogoUser) GetOrCreate(ctx context.Context, memberId int64) (user *entity.ToogoUser, err error) {
	err = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, memberId).Scan(&user)
	if err != nil {
		return nil, gerror.Wrap(err, "获取用户信息失败")
	}

	if user == nil {
		// 生成邀请码
		inviteCode := s.GenerateInviteCode()
		expireHours := 24 // 默认24小时有效

		user = &entity.ToogoUser{
			MemberId:         memberId,
			VipLevel:         1,
			IsAgent:          0,
			AgentLevel:       0,
			InviteCode:       inviteCode,
			InviteCodeExpire: gtime.Now().Add(time.Duration(expireHours) * time.Hour),
			InviterId:        0,
			InviteCount:      0,
			TeamCount:        0,
			RobotLimit:       1, // 免费套餐默认1个机器人
			Status:           1,
			CreatedAt:        gtime.Now(),
			UpdatedAt:        gtime.Now(),
		}

		_, err = dao.ToogoUser.Ctx(ctx).Data(user).Insert()
		if err != nil {
			return nil, gerror.Wrap(err, "创建用户信息失败")
		}

		// 同时创建钱包
		_, err = service.ToogoWallet().GetOrCreate(ctx, memberId)
		if err != nil {
			g.Log().Warningf(ctx, "创建用户钱包失败: %v", err)
		}
	}
	return
}

// GenerateInviteCode 生成邀请码
// 格式：2位大写字母 + 4位数字（数字不含4）
func (s *sToogoUser) GenerateInviteCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const digits = "012356789" // 不含4

	// 生成2位字母
	letter1 := letters[grand.N(0, len(letters)-1)]
	letter2 := letters[grand.N(0, len(letters)-1)]

	// 生成4位数字（不含4）
	digit1 := digits[grand.N(0, len(digits)-1)]
	digit2 := digits[grand.N(0, len(digits)-1)]
	digit3 := digits[grand.N(0, len(digits)-1)]
	digit4 := digits[grand.N(0, len(digits)-1)]

	return string([]byte{letter1, letter2, digit1, digit2, digit3, digit4})
}

// UserInfo 获取用户详细信息
func (s *sToogoUser) UserInfo(ctx context.Context, in *toogoin.UserInfoInp) (res *toogoin.UserInfoModel, err error) {
	if in == nil || in.MemberId <= 0 {
		return nil, gerror.New("\u672a\u767b\u5f55\u6216\u7528\u6237ID\u65e0\u6548")
	}

	user, err := s.GetOrCreate(ctx, in.MemberId)
	if err != nil {
		return nil, err
	}

	// 获取基础用户信息
	var member *entity.AdminMember
	err = dao.AdminMember.Ctx(ctx).Where(g.Map{
		dao.AdminMember.Columns().Id: in.MemberId,
	}).Scan(&member)
	if err != nil {
		return nil, gerror.Wrap(err, "获取用户基础信息失败")
	}

	// 获取VIP等级名称
	var vipLevel *entity.ToogoVipLevel
	dao.ToogoVipLevel.Ctx(ctx).Where(dao.ToogoVipLevel.Columns().Level, user.VipLevel).Scan(&vipLevel)

	res = &toogoin.UserInfoModel{
		ToogoUser: user,
	}
	if member != nil {
		res.Username = member.Username
		res.RealName = member.RealName
		res.Avatar = member.Avatar
	}
	if vipLevel != nil {
		res.VipLevelName = vipLevel.LevelName
	}
	return
}

// UserList 用户列表
func (s *sToogoUser) UserList(ctx context.Context, in *toogoin.UserListInp) (list []*toogoin.UserListModel, totalCount int, err error) {
	mod := dao.ToogoUser.Ctx(ctx)
	cols := dao.ToogoUser.Columns()

	if in.VipLevel > 0 {
		mod = mod.Where(cols.VipLevel, in.VipLevel)
	}
	if in.IsAgent >= 0 {
		mod = mod.Where(cols.IsAgent, in.IsAgent)
	}
	if in.Status > 0 {
		mod = mod.Where(cols.Status, in.Status)
	}
	if len(in.CreatedAt) == 2 {
		mod = mod.WhereBetween(cols.CreatedAt, in.CreatedAt[0], in.CreatedAt[1])
	}

	var users []*entity.ToogoUser
	err = mod.OrderDesc(cols.Id).Page(in.Page, in.PerPage).ScanAndCount(&users, &totalCount, true)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "获取用户列表失败")
	}

	// 获取关联信息
	for _, user := range users {
		item := &toogoin.UserListModel{
			ToogoUser: user,
		}

		// 获取用户名
		var member *entity.AdminMember
		_ = dao.AdminMember.Ctx(ctx).Where(g.Map{
			dao.AdminMember.Columns().Id: user.MemberId,
		}).Scan(&member)
		if member != nil {
			item.Username = member.Username
			item.RealName = member.RealName
			item.Avatar = member.Avatar
		}

		// 获取VIP等级名称
		var vipLevel *entity.ToogoVipLevel
		dao.ToogoVipLevel.Ctx(ctx).Where(dao.ToogoVipLevel.Columns().Level, user.VipLevel).Scan(&vipLevel)
		if vipLevel != nil {
			item.VipLevelName = vipLevel.LevelName
		}

		// 获取钱包信息（钱包表的user_id对应member_id）
		var wallet *entity.ToogoWallet
		dao.ToogoWallet.Ctx(ctx).Where(dao.ToogoWallet.Columns().UserId, user.MemberId).Scan(&wallet)
		if wallet != nil {
			item.Balance = wallet.Balance
			item.Power = wallet.Power
			item.GiftPower = wallet.GiftPower
			item.Commission = wallet.Commission
		}

		list = append(list, item)
	}
	return
}

// RefreshInviteCode 刷新邀请码
func (s *sToogoUser) RefreshInviteCode(ctx context.Context, in *toogoin.InviteCodeRefreshInp) (res *toogoin.InviteCodeRefreshModel, err error) {
	if in == nil || in.MemberId <= 0 {
		return nil, gerror.New("\u672a\u767b\u5f55\u6216\u7528\u6237ID\u65e0\u6548")
	}

	user, err := s.GetOrCreate(ctx, in.MemberId)
	if err != nil {
		return nil, err
	}

	// 生成新邀请码
	newCode := s.GenerateInviteCode()
	expireTime := gtime.Now().Add(gtime.H * 24) // 24小时有效

	_, err = dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().MemberId, in.MemberId).
		Data(g.Map{
			dao.ToogoUser.Columns().InviteCode:       newCode,
			dao.ToogoUser.Columns().InviteCodeExpire: expireTime,
		}).
		Update()
	if err != nil {
		return nil, gerror.Wrap(err, "刷新邀请码失败")
	}

	res = &toogoin.InviteCodeRefreshModel{
		InviteCode:       newCode,
		InviteCodeExpire: expireTime.String(),
		InviteUrl:        fmt.Sprintf("/register?inviteCode=%s", newCode),
	}
	_ = user
	return
}

// RegisterWithInvite 使用邀请码注册关联
func (s *sToogoUser) RegisterWithInvite(ctx context.Context, in *toogoin.RegisterWithInviteInp) error {
	// 允许使用两种邀请码：
	// 1) Toogo用户邀请码（toogo_user.invite_code）
	// 2) 基础邀请码（admin_member.invite_code）——用于统一邀请码入口，触发Toogo奖励
	var inviter *entity.ToogoUser
	foundByToogoCode := true
	err := dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().InviteCode, in.InviteCode).
		Where(dao.ToogoUser.Columns().Status, 1).
		Scan(&inviter)
	if err != nil {
		return gerror.Wrap(err, "查询邀请码失败")
	}

	if inviter == nil {
		foundByToogoCode = false
		// 回退：按基础邀请码查邀请人ID
		pmb, err := service.AdminMember().GetIdByCode(ctx, &adminin.GetIdByCodeInp{Code: in.InviteCode})
		if err != nil {
			return gerror.Wrap(err, "查询基础邀请码失败")
		}
		if pmb == nil || pmb.Id <= 0 {
			return gerror.New("邀请码无效或已过期")
		}

		// 确保邀请人存在Toogo扩展信息
		inviter, err = s.GetOrCreate(ctx, pmb.Id)
		if err != nil {
			return err
		}
		if inviter == nil || inviter.Status != 1 {
			return gerror.New("邀请人信息不存在或不可用")
		}
	}

	if inviter.MemberId == in.MemberId {
		return gerror.New("不能使用自己的邀请码")
	}

	// Toogo邀请码才检查过期；基础邀请码不做过期限制
	if foundByToogoCode {
		if inviter.InviteCodeExpire != nil && inviter.InviteCodeExpire.Before(gtime.Now()) {
			return gerror.New("邀请码已过期")
		}
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 创建新用户的Toogo信息
		newUser, err := s.GetOrCreate(ctx, in.MemberId)
		if err != nil {
			return err
		}

		// 设置邀请关系
		_, err = dao.ToogoUser.Ctx(ctx).
			Where(dao.ToogoUser.Columns().MemberId, in.MemberId).
			Data(g.Map{
				dao.ToogoUser.Columns().InviterId: inviter.MemberId,
			}).
			Update()
		if err != nil {
			return gerror.Wrap(err, "设置邀请关系失败")
		}

		// 更新邀请人的邀请数量
		_, err = dao.ToogoUser.Ctx(ctx).
			Where(dao.ToogoUser.Columns().MemberId, inviter.MemberId).
			Increment(dao.ToogoUser.Columns().InviteCount, 1)
		if err != nil {
			return gerror.Wrap(err, "更新邀请数量失败")
		}

		// 更新邀请人的团队数量 (需要递归更新整个上级链)
		err = s.UpdateTeamCount(ctx, inviter.MemberId, 1)
		if err != nil {
			g.Log().Warningf(ctx, "更新团队数量失败: %v", err)
		}

		// 发放邀请奖励
		err = service.ToogoCommission().SettleInviteReward(ctx, inviter.MemberId, newUser.MemberId)
		if err != nil {
			g.Log().Warningf(ctx, "发放邀请奖励失败: %v", err)
		}

		return nil
	})
}

// UpdateTeamCount 更新团队人数 (递归更新上级链)
func (s *sToogoUser) UpdateTeamCount(ctx context.Context, memberId int64, delta int) error {
	var user *entity.ToogoUser
	err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, memberId).Scan(&user)
	if err != nil || user == nil {
		return err
	}

	// 更新当前用户的团队数量
	_, err = dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().MemberId, memberId).
		Increment(dao.ToogoUser.Columns().TeamCount, delta)
	if err != nil {
		return err
	}

	// 如果有上级，递归更新
	if user.InviterId > 0 {
		return s.UpdateTeamCount(ctx, user.InviterId, delta)
	}
	return nil
}

// TeamList 团队列表
func (s *sToogoUser) TeamList(ctx context.Context, in *toogoin.TeamListInp) (list []*toogoin.TeamListModel, totalCount int, err error) {
	// 根据层级查询
	var memberIds []int64

	// 是否已解锁无限层级（解锁后：默认“全部”=无限级；未解锁：默认“全部”=直推）
	allowUnlimited := false
	var self *entity.ToogoUser
	_ = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, in.MemberId).Scan(&self)
	if self != nil && self.IsAgent == 1 && self.AgentUnlockLevel == 1 {
		allowUnlimited = true
	}

	switch in.Level {
	case 1: // 直推
		memberIds, err = s.GetDirectMembers(ctx, in.MemberId)
	case 2: // 二级
		memberIds, err = s.GetLevel2Members(ctx, in.MemberId)
	case 3: // 三级
		memberIds, err = s.GetLevel3Members(ctx, in.MemberId)
	default: // 全部
		if allowUnlimited {
			memberIds, err = s.GetAllTeamMembersUnlimited(ctx, in.MemberId)
		} else {
			// 未解锁：仅展示直推
			memberIds, err = s.GetDirectMembers(ctx, in.MemberId)
		}
	}

	if err != nil {
		return nil, 0, err
	}

	totalCount = len(memberIds)
	if totalCount == 0 {
		return
	}

	// 分页
	start := (in.Page - 1) * in.PerPage
	end := start + in.PerPage
	if start >= totalCount {
		return
	}
	if end > totalCount {
		end = totalCount
	}
	pageIds := memberIds[start:end]

	// 获取详细信息
	for _, mid := range pageIds {
		item := &toogoin.TeamListModel{
			MemberId: mid,
			Level:    s.GetMemberLevel(ctx, in.MemberId, mid),
		}

		// 获取用户信息
		var member *entity.AdminMember
		_ = dao.AdminMember.Ctx(ctx).Where(g.Map{
			dao.AdminMember.Columns().Id: mid,
		}).Scan(&member)
		if member != nil {
			item.Username = member.Username
			item.Avatar = member.Avatar
			item.RegisterTime = member.CreatedAt.String()
		}

		// 获取Toogo用户信息
		var toogoUser *entity.ToogoUser
		dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, mid).Scan(&toogoUser)
		if toogoUser != nil {
			item.VipLevel = toogoUser.VipLevel
			item.TotalConsume = toogoUser.TotalConsumePower
			item.IsAgent = toogoUser.IsAgent
			item.AgentStatus = toogoUser.AgentStatus
			item.SubscribeRate = toogoUser.SubscribeRate
			item.InviteCount = toogoUser.InviteCount
			item.TeamCount = toogoUser.TeamCount
			item.CurrentPlanId = toogoUser.CurrentPlanId
			item.PlanExpireTime = toogoUser.PlanExpireTime
			item.RobotLimit = toogoUser.RobotLimit
			item.ActiveRobotCount = toogoUser.ActiveRobotCount

			// 获取VIP等级名称
			var vipLevel *entity.ToogoVipLevel
			dao.ToogoVipLevel.Ctx(ctx).Where(dao.ToogoVipLevel.Columns().Level, toogoUser.VipLevel).Scan(&vipLevel)
			if vipLevel != nil {
				item.VipLevelName = vipLevel.LevelName
			}
		}

		list = append(list, item)
	}
	return
}

// GetAllTeamMembersUnlimited 获取所有团队成员（无限级）
// 使用 BFS 按 inviter_id 向下遍历，返回所有下级 member_id（不包含自己）
func (s *sToogoUser) GetAllTeamMembersUnlimited(ctx context.Context, memberId int64) ([]int64, error) {
	visited := make(map[int64]struct{})
	result := make([]int64, 0, 64)

	current := []int64{memberId}
	for len(current) > 0 {
		var users []*entity.ToogoUser
		err := dao.ToogoUser.Ctx(ctx).
			WhereIn(dao.ToogoUser.Columns().InviterId, current).
			Fields(dao.ToogoUser.Columns().MemberId).
			Scan(&users)
		if err != nil {
			return nil, err
		}

		next := make([]int64, 0, len(users))
		for _, u := range users {
			if u == nil {
				continue
			}
			id := u.MemberId
			if id <= 0 || id == memberId {
				continue
			}
			if _, ok := visited[id]; ok {
				continue
			}
			visited[id] = struct{}{}
			result = append(result, id)
			next = append(next, id)
		}
		current = next
	}
	return result, nil
}

// GetDirectMembers 获取直推成员
func (s *sToogoUser) GetDirectMembers(ctx context.Context, memberId int64) ([]int64, error) {
	var users []*entity.ToogoUser
	err := dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().InviterId, memberId).
		Scan(&users)
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, u := range users {
		ids = append(ids, u.MemberId)
	}
	return ids, nil
}

// GetLevel2Members 获取二级成员
func (s *sToogoUser) GetLevel2Members(ctx context.Context, memberId int64) ([]int64, error) {
	directIds, err := s.GetDirectMembers(ctx, memberId)
	if err != nil || len(directIds) == 0 {
		return nil, err
	}

	var users []*entity.ToogoUser
	err = dao.ToogoUser.Ctx(ctx).
		WhereIn(dao.ToogoUser.Columns().InviterId, directIds).
		Scan(&users)
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, u := range users {
		ids = append(ids, u.MemberId)
	}
	return ids, nil
}

// GetLevel3Members 获取三级成员
func (s *sToogoUser) GetLevel3Members(ctx context.Context, memberId int64) ([]int64, error) {
	level2Ids, err := s.GetLevel2Members(ctx, memberId)
	if err != nil || len(level2Ids) == 0 {
		return nil, err
	}

	var users []*entity.ToogoUser
	err = dao.ToogoUser.Ctx(ctx).
		WhereIn(dao.ToogoUser.Columns().InviterId, level2Ids).
		Scan(&users)
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, u := range users {
		ids = append(ids, u.MemberId)
	}
	return ids, nil
}

// GetAllTeamMembers 获取所有团队成员
func (s *sToogoUser) GetAllTeamMembers(ctx context.Context, memberId int64) ([]int64, error) {
	var allIds []int64

	level1, _ := s.GetDirectMembers(ctx, memberId)
	level2, _ := s.GetLevel2Members(ctx, memberId)
	level3, _ := s.GetLevel3Members(ctx, memberId)

	allIds = append(allIds, level1...)
	allIds = append(allIds, level2...)
	allIds = append(allIds, level3...)

	return allIds, nil
}

// GetMemberLevel 获取成员层级
func (s *sToogoUser) GetMemberLevel(ctx context.Context, rootId, memberId int64) int {
	var user *entity.ToogoUser
	dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, memberId).Scan(&user)
	if user == nil {
		return 0
	}

	if user.InviterId == rootId {
		return 1
	}

	// 检查二级
	var inviter *entity.ToogoUser
	dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, user.InviterId).Scan(&inviter)
	if inviter != nil && inviter.InviterId == rootId {
		return 2
	}

	// 检查三级
	if inviter != nil {
		var inviter2 *entity.ToogoUser
		dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, inviter.InviterId).Scan(&inviter2)
		if inviter2 != nil && inviter2.InviterId == rootId {
			return 3
		}
	}

	return 0
}

// TeamStat 团队统计
func (s *sToogoUser) TeamStat(ctx context.Context, in *toogoin.TeamStatInp) (res *toogoin.TeamStatModel, err error) {
	level1, _ := s.GetDirectMembers(ctx, in.MemberId)
	level2, _ := s.GetLevel2Members(ctx, in.MemberId)
	level3, _ := s.GetLevel3Members(ctx, in.MemberId)

	// 是否已解锁无限层级：解锁后“推广总人数”=无限级人数；未解锁则只显示直推人数
	allowUnlimited := false
	var self *entity.ToogoUser
	_ = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, in.MemberId).Scan(&self)
	if self != nil && self.IsAgent == 1 && self.AgentUnlockLevel == 1 {
		allowUnlimited = true
	}

	totalCount := len(level1)
	var allIds []int64
	if allowUnlimited {
		allIds, _ = s.GetAllTeamMembersUnlimited(ctx, in.MemberId)
		totalCount = len(allIds)
	} else {
		allIds = level1
	}

	res = &toogoin.TeamStatModel{
		DirectCount: len(level1),
		Level2Count: len(level2),
		Level3Count: len(level3),
		TotalCount:  totalCount,
	}

	// 计算团队消耗
	if len(allIds) > 0 {
		// 统计团队总消耗算力
		value, err := dao.ToogoUser.Ctx(ctx).
			WhereIn(dao.ToogoUser.Columns().MemberId, allIds).
			Sum(dao.ToogoUser.Columns().TotalConsumePower)
		if err == nil {
			res.TotalConsume = value
		}

		// 统计团队总订阅金额
		subValue, err := dao.ToogoSubscription.Ctx(ctx).
			WhereIn(dao.ToogoSubscription.Columns().UserId, allIds).
			Where(dao.ToogoSubscription.Columns().Status, 2). // 生效中
			Sum(dao.ToogoSubscription.Columns().Amount)
		if err == nil {
			res.TotalSubscribe = subValue
		}
	}

	// 获取累计佣金
	commValue, err := dao.ToogoCommissionLog.Ctx(ctx).
		Where(dao.ToogoCommissionLog.Columns().UserId, in.MemberId).
		Sum(dao.ToogoCommissionLog.Columns().CommissionAmount)
	if err == nil {
		res.TotalCommission = commValue
	}

	return
}

// VipLevelList VIP等级列表
func (s *sToogoUser) VipLevelList(ctx context.Context, in *toogoin.VipLevelListInp) (list []*toogoin.VipLevelListModel, totalCount int, err error) {
	mod := dao.ToogoVipLevel.Ctx(ctx)

	if in.Status > 0 {
		mod = mod.Where(dao.ToogoVipLevel.Columns().Status, in.Status)
	}

	err = mod.OrderAsc(dao.ToogoVipLevel.Columns().Level).Page(in.Page, in.PerPage).ScanAndCount(&list, &totalCount, true)
	if err != nil {
		err = gerror.Wrap(err, "获取VIP等级列表失败")
	}
	return
}

// VipLevelEdit 编辑VIP等级
func (s *sToogoUser) VipLevelEdit(ctx context.Context, in *toogoin.VipLevelEditInp) (err error) {
	cols := dao.ToogoVipLevel.Columns()

	data := g.Map{
		cols.Level:               in.Level,
		cols.LevelName:           in.LevelName,
		cols.RequireInviteCount:  in.RequireInviteCount,
		cols.RequireConsumePower: in.RequireConsumePower,
		cols.RequireTeamConsume:  in.RequireTeamConsume,
		cols.PowerDiscount:       in.PowerDiscount,
		cols.InviteRewardPower:   in.InviteRewardPower,
		cols.Description:         in.Description,
		cols.Icon:                in.Icon,
		cols.Sort:                in.Sort,
		cols.Status:              in.Status,
	}

	if in.Id > 0 {
		_, err = dao.ToogoVipLevel.Ctx(ctx).Where(dao.ToogoVipLevel.Columns().Id, in.Id).Data(data).Update()
	} else {
		_, err = dao.ToogoVipLevel.Ctx(ctx).Data(data).Insert()
	}

	if err != nil {
		err = gerror.Wrap(err, "保存VIP等级失败")
	}
	return
}

// CheckVipUpgrade 检查VIP升级
func (s *sToogoUser) CheckVipUpgrade(ctx context.Context, in *toogoin.CheckVipUpgradeInp) (res *toogoin.CheckVipUpgradeModel, err error) {
	if in == nil || in.MemberId <= 0 {
		return nil, gerror.New("\u672a\u767b\u5f55\u6216\u7528\u6237ID\u65e0\u6548")
	}

	user, err := s.GetOrCreate(ctx, in.MemberId)
	if err != nil {
		return nil, err
	}

	// 获取下一等级配置
	var nextLevel *entity.ToogoVipLevel
	err = dao.ToogoVipLevel.Ctx(ctx).
		Where(dao.ToogoVipLevel.Columns().Level, user.VipLevel+1).
		Where(dao.ToogoVipLevel.Columns().Status, 1).
		Scan(&nextLevel)
	if err != nil || nextLevel == nil {
		// 已是最高等级或没有下一等级
		res = &toogoin.CheckVipUpgradeModel{
			CanUpgrade:   false,
			CurrentLevel: user.VipLevel,
		}
		return
	}

	res = &toogoin.CheckVipUpgradeModel{
		CurrentLevel:  user.VipLevel,
		NextLevel:     nextLevel.Level,
		NextLevelName: nextLevel.LevelName,
	}

	res.Progress.InviteCount = user.InviteCount
	res.Progress.RequireInvite = nextLevel.RequireInviteCount
	res.Progress.ConsumePower = user.TotalConsumePower
	res.Progress.RequireConsume = nextLevel.RequireConsumePower
	res.Progress.TeamConsume = user.TeamConsumePower
	res.Progress.RequireTeam = nextLevel.RequireTeamConsume

	// 判断是否满足升级条件
	canUpgrade := (user.InviteCount >= nextLevel.RequireInviteCount) ||
		(user.TotalConsumePower >= nextLevel.RequireConsumePower) ||
		(user.TeamConsumePower >= nextLevel.RequireTeamConsume)
	res.CanUpgrade = canUpgrade

	// 如果满足条件，自动升级
	if canUpgrade {
		_, err = dao.ToogoUser.Ctx(ctx).
			Where(dao.ToogoUser.Columns().MemberId, in.MemberId).
			Data(g.Map{
				dao.ToogoUser.Columns().VipLevel:      nextLevel.Level,
				dao.ToogoUser.Columns().PowerDiscount: nextLevel.PowerDiscount,
			}).
			Update()
		if err != nil {
			g.Log().Warningf(ctx, "VIP升级失败: %v", err)
		}
	}

	return
}
