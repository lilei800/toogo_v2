// Package admin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package admin

import (
	"context"
	"hotgo/api/admin"
	"hotgo/internal/library/contexts"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"
)

var Toogo = cToogo{}

type cToogo struct{}

// ========== 钱包管理 ==========

// WalletOverview 钱包概览
func (c *cToogo) WalletOverview(ctx context.Context, req *admin.ToogoWalletOverviewReq) (res *admin.ToogoWalletOverviewRes, err error) {
	userId := req.UserId
	if userId == 0 {
		userId = contexts.GetUserId(ctx)
	}

	data, err := service.ToogoWallet().GetOverview(ctx, &toogoin.WalletOverviewInp{UserId: userId})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoWalletOverviewRes{WalletOverviewModel: data}
	return
}

// WalletLogList 钱包流水列表
func (c *cToogo) WalletLogList(ctx context.Context, req *admin.ToogoWalletLogListReq) (res *admin.ToogoWalletLogListRes, err error) {
	if req.UserId == 0 {
		req.UserId = contexts.GetUserId(ctx)
	}

	list, totalCount, err := service.ToogoWallet().WalletLogList(ctx, &req.WalletLogListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoWalletLogListRes{List: list, TotalCount: totalCount}
	return
}

// Transfer 账户互转
func (c *cToogo) Transfer(ctx context.Context, req *admin.ToogoTransferReq) (res *admin.ToogoTransferRes, err error) {
	if req.UserId == 0 {
		req.UserId = contexts.GetUserId(ctx)
	}

	data, err := service.ToogoWallet().Transfer(ctx, &req.TransferInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoTransferRes{TransferModel: data}
	return
}

// AdminRecharge 管理员充值（算力/积分/余额）
func (c *cToogo) AdminRecharge(ctx context.Context, req *admin.ToogoAdminRechargeReq) (res *admin.ToogoAdminRechargeRes, err error) {
	beforeAmount, afterAmount, err := service.ToogoWallet().AdminRecharge(ctx, req.UserId, req.AccountType, req.Amount, req.Remark)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoAdminRechargeRes{
		BeforeAmount: beforeAmount,
		AfterAmount:  afterAmount,
	}
	return
}

// UserWalletList 用户钱包列表（管理员）
func (c *cToogo) UserWalletList(ctx context.Context, req *admin.ToogoUserWalletListReq) (res *admin.ToogoUserWalletListRes, err error) {
	list, totalCount, err := service.ToogoWallet().UserWalletList(ctx, req.Username, req.Mobile, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoUserWalletListRes{
		List:       list,
		TotalCount: totalCount,
	}
	return
}

// OrderHistoryList 历史交易订单列表
func (c *cToogo) OrderHistoryList(ctx context.Context, req *admin.ToogoOrderHistoryListReq) (res *admin.ToogoOrderHistoryListRes, err error) {
	list, totalCount, err := service.ToogoWallet().OrderHistoryList(ctx, &req.OrderHistoryListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoOrderHistoryListRes{
		List:       list,
		TotalCount: totalCount,
	}
	return
}

// TradeHistoryList 成交流水列表
func (c *cToogo) TradeHistoryList(ctx context.Context, req *admin.ToogoTradeHistoryListReq) (res *admin.ToogoTradeHistoryListRes, err error) {
	list, totalCount, summary, err := service.ToogoWallet().TradeHistoryList(ctx, &req.TradeHistoryListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoTradeHistoryListRes{
		List:       list,
		TotalCount: totalCount,
		Summary:    summary,
	}
	return
}

// RunSessionSummaryList 运行区间盈亏汇总列表
func (c *cToogo) RunSessionSummaryList(ctx context.Context, req *admin.ToogoRunSessionSummaryListReq) (res *admin.ToogoRunSessionSummaryListRes, err error) {
	list, totalCount, summary, err := service.ToogoWallet().RunSessionSummaryList(ctx, &req.RunSessionSummaryListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoRunSessionSummaryListRes{
		List:       list,
		TotalCount: totalCount,
		Summary:    summary,
	}
	return
}

// SyncRunSession 同步运行区间盈亏数据
func (c *cToogo) SyncRunSession(ctx context.Context, req *admin.ToogoSyncRunSessionReq) (res *admin.ToogoSyncRunSessionRes, err error) {
	totalPnl, totalFee, tradeCount, err := service.ToogoWallet().SyncRunSession(ctx, req.SessionId, req.CalcOnly == 1)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoSyncRunSessionRes{
		TotalPnl:   totalPnl,
		TotalFee:   totalFee,
		TradeCount: tradeCount,
	}
	return
}

// ========== 订阅管理 ==========

// PlanList 套餐列表
func (c *cToogo) PlanList(ctx context.Context, req *admin.ToogoPlanListReq) (res *admin.ToogoPlanListRes, err error) {
	list, totalCount, err := service.ToogoSubscription().PlanList(ctx, &req.PlanListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoPlanListRes{List: list, TotalCount: totalCount}
	return
}

// PlanEdit 编辑套餐
func (c *cToogo) PlanEdit(ctx context.Context, req *admin.ToogoPlanEditReq) (res *admin.ToogoPlanEditRes, err error) {
	err = service.ToogoSubscription().PlanEdit(ctx, &req.PlanEditInp)
	return
}

// PlanDelete 删除套餐
func (c *cToogo) PlanDelete(ctx context.Context, req *admin.ToogoPlanDeleteReq) (res *admin.ToogoPlanDeleteRes, err error) {
	err = service.ToogoSubscription().PlanDelete(ctx, &req.PlanDeleteInp)
	return
}

// Subscribe 订阅套餐
func (c *cToogo) Subscribe(ctx context.Context, req *admin.ToogoSubscribeReq) (res *admin.ToogoSubscribeRes, err error) {
	if req.UserId == 0 {
		req.UserId = contexts.GetUserId(ctx)
	}

	data, err := service.ToogoSubscription().Subscribe(ctx, &req.SubscribeInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoSubscribeRes{SubscribeModel: data}
	return
}

// SubscriptionList 订阅记录列表
func (c *cToogo) SubscriptionList(ctx context.Context, req *admin.ToogoSubscriptionListReq) (res *admin.ToogoSubscriptionListRes, err error) {
	// 安全控制：如果没有指定userId，默认查询当前用户的订阅记录
	// 如果指定了userId且不是当前用户，需要有相应权限（这里简化处理，允许查询）
	// 前端普通用户页面应该主动传递当前用户ID，确保只查询自己的数据
	currentUserId := contexts.GetUserId(ctx)

	if req.UserId == 0 {
		// 如果没有指定userId，使用当前用户ID
		req.UserId = currentUserId
	}

	list, totalCount, err := service.ToogoSubscription().SubscriptionList(ctx, &req.SubscriptionListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoSubscriptionListRes{List: list, TotalCount: totalCount}
	return
}

// MySubscription 我的订阅
func (c *cToogo) MySubscription(ctx context.Context, req *admin.ToogoMySubscriptionReq) (res *admin.ToogoMySubscriptionRes, err error) {
	userId := contexts.GetUserId(ctx)
	data, err := service.ToogoSubscription().MySubscription(ctx, &toogoin.MySubscriptionInp{UserId: userId})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoMySubscriptionRes{MySubscriptionModel: data}
	return
}

// ========== 用户管理 ==========

// UserInfo 用户信息
func (c *cToogo) UserInfo(ctx context.Context, req *admin.ToogoUserInfoReq) (res *admin.ToogoUserInfoRes, err error) {
	memberId := req.MemberId
	if memberId == 0 {
		memberId = contexts.GetUserId(ctx)
	}

	data, err := service.ToogoUser().UserInfo(ctx, &toogoin.UserInfoInp{MemberId: memberId})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoUserInfoRes{UserInfoModel: data}
	return
}

// UserList 用户列表
func (c *cToogo) UserList(ctx context.Context, req *admin.ToogoUserListReq) (res *admin.ToogoUserListRes, err error) {
	list, totalCount, err := service.ToogoUser().UserList(ctx, &req.UserListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoUserListRes{List: list, TotalCount: totalCount}
	return
}

// RefreshInviteCode 刷新邀请码
func (c *cToogo) RefreshInviteCode(ctx context.Context, req *admin.ToogoRefreshInviteCodeReq) (res *admin.ToogoRefreshInviteCodeRes, err error) {
	memberId := contexts.GetUserId(ctx)
	data, err := service.ToogoUser().RefreshInviteCode(ctx, &toogoin.InviteCodeRefreshInp{MemberId: memberId})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoRefreshInviteCodeRes{InviteCodeRefreshModel: data}
	return
}

// TeamList 团队列表
func (c *cToogo) TeamList(ctx context.Context, req *admin.ToogoTeamListReq) (res *admin.ToogoTeamListRes, err error) {
	if req.MemberId == 0 {
		req.MemberId = contexts.GetUserId(ctx)
	}

	list, totalCount, err := service.ToogoUser().TeamList(ctx, &req.TeamListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoTeamListRes{List: list, TotalCount: totalCount}
	return
}

// TeamStat 团队统计
func (c *cToogo) TeamStat(ctx context.Context, req *admin.ToogoTeamStatReq) (res *admin.ToogoTeamStatRes, err error) {
	memberId := contexts.GetUserId(ctx)
	data, err := service.ToogoUser().TeamStat(ctx, &toogoin.TeamStatInp{MemberId: memberId})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoTeamStatRes{TeamStatModel: data}
	return
}

// VipLevelList VIP等级列表
func (c *cToogo) VipLevelList(ctx context.Context, req *admin.ToogoVipLevelListReq) (res *admin.ToogoVipLevelListRes, err error) {
	list, totalCount, err := service.ToogoUser().VipLevelList(ctx, &req.VipLevelListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoVipLevelListRes{List: list, TotalCount: totalCount}
	return
}

// VipLevelEdit 编辑VIP等级
func (c *cToogo) VipLevelEdit(ctx context.Context, req *admin.ToogoVipLevelEditReq) (res *admin.ToogoVipLevelEditRes, err error) {
	err = service.ToogoUser().VipLevelEdit(ctx, &req.VipLevelEditInp)
	return
}

// CheckVipUpgrade 检查VIP升级
func (c *cToogo) CheckVipUpgrade(ctx context.Context, req *admin.ToogoCheckVipUpgradeReq) (res *admin.ToogoCheckVipUpgradeRes, err error) {
	memberId := contexts.GetUserId(ctx)
	data, err := service.ToogoUser().CheckVipUpgrade(ctx, &toogoin.CheckVipUpgradeInp{MemberId: memberId})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoCheckVipUpgradeRes{CheckVipUpgradeModel: data}
	return
}

// ========== 佣金管理 ==========

// CommissionLogList 佣金记录列表
func (c *cToogo) CommissionLogList(ctx context.Context, req *admin.ToogoCommissionLogListReq) (res *admin.ToogoCommissionLogListRes, err error) {
	if req.UserId == 0 {
		req.UserId = contexts.GetUserId(ctx)
	}

	list, totalCount, err := service.ToogoCommission().CommissionLogList(ctx, &req.CommissionLogListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoCommissionLogListRes{List: list, TotalCount: totalCount}
	return
}

// CommissionStat 佣金统计
func (c *cToogo) CommissionStat(ctx context.Context, req *admin.ToogoCommissionStatReq) (res *admin.ToogoCommissionStatRes, err error) {
	userId := contexts.GetUserId(ctx)
	data, err := service.ToogoCommission().CommissionStat(ctx, &toogoin.CommissionStatInp{UserId: userId})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoCommissionStatRes{CommissionStatModel: data}
	return
}

// AgentLevelList 代理商等级列表
func (c *cToogo) AgentLevelList(ctx context.Context, req *admin.ToogoAgentLevelListReq) (res *admin.ToogoAgentLevelListRes, err error) {
	list, totalCount, err := service.ToogoCommission().AgentLevelList(ctx, &req.AgentLevelListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoAgentLevelListRes{List: list, TotalCount: totalCount}
	return
}

// AgentLevelEdit 编辑代理商等级
func (c *cToogo) AgentLevelEdit(ctx context.Context, req *admin.ToogoAgentLevelEditReq) (res *admin.ToogoAgentLevelEditRes, err error) {
	err = service.ToogoCommission().AgentLevelEdit(ctx, &req.AgentLevelEditInp)
	return
}

// ApplyAgent 申请成为代理商
func (c *cToogo) ApplyAgent(ctx context.Context, req *admin.ToogoApplyAgentReq) (res *admin.ToogoApplyAgentRes, err error) {
	memberId := contexts.GetUserId(ctx)
	data, err := service.ToogoCommission().ApplyAgent(ctx, &toogoin.ApplyAgentInp{
		MemberId:      memberId,
		Remark:        req.Remark,
		SubscribeRate: req.SubscribeRate,
	})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoApplyAgentRes{ApplyAgentModel: data}
	return
}

// ApplyAgentForSub 代下级提交代理申请（仅直属下级）
func (c *cToogo) ApplyAgentForSub(ctx context.Context, req *admin.ToogoApplyAgentForSubReq) (res *admin.ToogoApplyAgentForSubRes, err error) {
	agentId := contexts.GetUserId(ctx)
	err = service.ToogoCommission().ApplyAgentForSub(ctx, &toogoin.ApplyAgentForSubInp{
		AgentId:       agentId,
		SubUserId:     req.SubUserId,
		Remark:        req.Remark,
		SubscribeRate: req.SubscribeRate,
	})
	return
}

// ApproveAgent 审批代理商申请（管理员）
func (c *cToogo) ApproveAgent(ctx context.Context, req *admin.ToogoApproveAgentReq) (res *admin.ToogoApproveAgentRes, err error) {
	operatorId := contexts.GetUserId(ctx)
	data, err := service.ToogoCommission().ApproveAgent(ctx, &req.ApproveAgentInp, operatorId)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoApproveAgentRes{ApproveAgentModel: data}
	return
}

// UpdateAgent 更新代理商信息（管理员）
func (c *cToogo) UpdateAgent(ctx context.Context, req *admin.ToogoUpdateAgentReq) (res *admin.ToogoUpdateAgentRes, err error) {
	err = service.ToogoCommission().UpdateAgent(ctx, &req.UpdateAgentInp)
	return
}

// SetSubAgentRate 设置下级代理佣金比例
func (c *cToogo) SetSubAgentRate(ctx context.Context, req *admin.ToogoSetSubAgentRateReq) (res *admin.ToogoSetSubAgentRateRes, err error) {
	agentId := contexts.GetUserId(ctx)
	err = service.ToogoCommission().SetSubAgentRate(ctx, &toogoin.SetSubAgentRateInp{
		AgentId:       agentId,
		SubUserId:     req.SubUserId,
		SubscribeRate: req.SubscribeRate,
	})
	return
}

// GetAgentInfo 获取代理信息
func (c *cToogo) GetAgentInfo(ctx context.Context, req *admin.ToogoGetAgentInfoReq) (res *admin.ToogoGetAgentInfoRes, err error) {
	memberId := contexts.GetUserId(ctx)
	data, err := service.ToogoCommission().GetAgentInfo(ctx, memberId)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoGetAgentInfoRes{AgentInfoModel: data}
	return
}

// ========== 策略管理 ==========

// StrategyTemplateList 策略模板列表
func (c *cToogo) StrategyTemplateList(ctx context.Context, req *admin.ToogoStrategyTemplateListReq) (res *admin.ToogoStrategyTemplateListRes, err error) {
	list, totalCount, err := service.ToogoStrategy().TemplateList(ctx, &req.StrategyTemplateListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoStrategyTemplateListRes{List: list, TotalCount: totalCount}
	return
}

// StrategyTemplateEdit 编辑策略模板
func (c *cToogo) StrategyTemplateEdit(ctx context.Context, req *admin.ToogoStrategyTemplateEditReq) (res *admin.ToogoStrategyTemplateEditRes, err error) {
	err = service.ToogoStrategy().TemplateEdit(ctx, &req.StrategyTemplateEditInp)
	return
}

// StrategyTemplateDelete 删除策略模板
func (c *cToogo) StrategyTemplateDelete(ctx context.Context, req *admin.ToogoStrategyTemplateDeleteReq) (res *admin.ToogoStrategyTemplateDeleteRes, err error) {
	err = service.ToogoStrategy().TemplateDelete(ctx, &req.StrategyTemplateDeleteInp)
	return
}

// GetStrategy 获取策略
func (c *cToogo) GetStrategy(ctx context.Context, req *admin.ToogoGetStrategyReq) (res *admin.ToogoGetStrategyRes, err error) {
	data, err := service.ToogoStrategy().GetByCondition(ctx, &req.GetStrategyByConditionInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoGetStrategyRes{GetStrategyByConditionModel: data}
	return
}

// PowerConsumeList 算力消耗记录列表
func (c *cToogo) PowerConsumeList(ctx context.Context, req *admin.ToogoPowerConsumeListReq) (res *admin.ToogoPowerConsumeListRes, err error) {
	if req.UserId == 0 {
		req.UserId = contexts.GetUserId(ctx)
	}

	list, totalCount, err := service.ToogoStrategy().PowerConsumeList(ctx, &req.PowerConsumeListInp)
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoPowerConsumeListRes{List: list, TotalCount: totalCount}
	return
}

// PowerConsumeStat 算力消耗统计
func (c *cToogo) PowerConsumeStat(ctx context.Context, req *admin.ToogoPowerConsumeStatReq) (res *admin.ToogoPowerConsumeStatRes, err error) {
	userId := contexts.GetUserId(ctx)
	data, err := service.ToogoStrategy().PowerConsumeStat(ctx, &toogoin.PowerConsumeStatInp{UserId: userId})
	if err != nil {
		return nil, err
	}
	res = &admin.ToogoPowerConsumeStatRes{PowerConsumeStatModel: data}
	return
}

// ========== 管理员操作 ==========

// AdminRechargePower 管理员手动充值算力
func (c *cToogo) AdminRechargePower(ctx context.Context, req *admin.ToogoAdminRechargePowerReq) (res *admin.ToogoAdminRechargePowerRes, err error) {
	// 充值算力
	err = service.ToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
		UserId:      req.UserId,
		AccountType: "power",
		ChangeType:  "admin_recharge",
		Amount:      req.Amount,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}

	// 获取用户钱包信息
	wallet, err := service.ToogoWallet().GetOrCreate(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	res = &admin.ToogoAdminRechargePowerRes{
		AdminRechargePowerModel: &toogoin.AdminRechargePowerModel{
			UserId:     req.UserId,
			Amount:     req.Amount,
			NewBalance: wallet.Power,
		},
	}
	return
}

// AdminRechargeBalance 管理员手动充值余额
func (c *cToogo) AdminRechargeBalance(ctx context.Context, req *admin.ToogoAdminRechargeBalanceReq) (res *admin.ToogoAdminRechargeBalanceRes, err error) {
	// 充值余额
	err = service.ToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
		UserId:      req.UserId,
		AccountType: "balance",
		ChangeType:  "admin_recharge",
		Amount:      req.Amount,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}

	// 获取用户钱包信息
	wallet, err := service.ToogoWallet().GetOrCreate(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	res = &admin.ToogoAdminRechargeBalanceRes{
		AdminRechargeBalanceModel: &toogoin.AdminRechargeBalanceModel{
			UserId:     req.UserId,
			Amount:     req.Amount,
			NewBalance: wallet.Balance,
		},
	}
	return
}

// AdminRechargePoints 管理员手动充值积分
func (c *cToogo) AdminRechargePoints(ctx context.Context, req *admin.ToogoAdminRechargePointsReq) (res *admin.ToogoAdminRechargePointsRes, err error) {
	// 充值积分
	err = service.ToogoWallet().ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
		UserId:      req.UserId,
		AccountType: "gift_power",
		ChangeType:  "admin_recharge",
		Amount:      req.Amount,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}

	// 获取用户钱包信息
	wallet, err := service.ToogoWallet().GetOrCreate(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	res = &admin.ToogoAdminRechargePointsRes{
		AdminRechargePointsModel: &toogoin.AdminRechargePointsModel{
			UserId:     req.UserId,
			Amount:     req.Amount,
			NewBalance: wallet.GiftPower,
		},
	}
	return
}
