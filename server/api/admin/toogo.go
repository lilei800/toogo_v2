// Package admin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package admin

import (
	"hotgo/internal/model/input/toogoin"

	"github.com/gogf/gf/v2/frame/g"
)

// ========== 钱包管理 ==========

// ToogoWalletOverviewReq 钱包概览请求
type ToogoWalletOverviewReq struct {
	g.Meta `path:"/toogo/wallet/overview" method:"get" tags:"Toogo钱包" summary:"钱包概览"`
	UserId int64 `json:"userId" dc:"用户ID，不传则获取当前登录用户"`
}

type ToogoWalletOverviewRes struct {
	*toogoin.WalletOverviewModel
}

// ToogoWalletLogListReq 钱包流水列表请求
type ToogoWalletLogListReq struct {
	g.Meta `path:"/toogo/wallet/log/list" method:"get" tags:"Toogo钱包" summary:"钱包流水列表"`
	toogoin.WalletLogListInp
}

type ToogoWalletLogListRes struct {
	List       []*toogoin.WalletLogListModel `json:"list"`
	TotalCount int                           `json:"totalCount"`
}

// ToogoTransferReq 账户互转请求
type ToogoTransferReq struct {
	g.Meta `path:"/toogo/wallet/transfer" method:"post" tags:"Toogo钱包" summary:"账户互转"`
	toogoin.TransferInp
}

type ToogoTransferRes struct {
	*toogoin.TransferModel
}

// ToogoAdminRechargeReq 管理员充值请求
type ToogoAdminRechargeReq struct {
	g.Meta      `path:"/toogo/wallet/adminRecharge" method:"post" tags:"Toogo钱包" summary:"管理员充值算力"`
	UserId      int64   `json:"userId" v:"required#请选择充值用户" dc:"用户ID"`
	AccountType string  `json:"accountType" v:"required|in:power,gift_power,balance#请选择账户类型|账户类型只能是power/gift_power/balance" dc:"账户类型：power=算力,gift_power=积分,balance=余额"`
	Amount      float64 `json:"amount" v:"required|min:0.01#请输入充值金额|充值金额必须大于0" dc:"充值金额"`
	Remark      string  `json:"remark" dc:"充值备注"`
}

type ToogoAdminRechargeRes struct {
	BeforeAmount float64 `json:"beforeAmount" dc:"充值前余额"`
	AfterAmount  float64 `json:"afterAmount" dc:"充值后余额"`
}

// ToogoUserWalletListReq 用户钱包列表请求（管理员）
type ToogoUserWalletListReq struct {
	g.Meta   `path:"/toogo/wallet/userList" method:"get" tags:"Toogo钱包" summary:"用户钱包列表"`
	Username string `json:"username" dc:"用户名搜索"`
	Mobile   string `json:"mobile" dc:"手机号搜索"`
	Page     int    `json:"page" d:"1" dc:"页码"`
	PerPage  int    `json:"perPage" d:"20" dc:"每页数量"`
}

type ToogoUserWalletListRes struct {
	List       []*toogoin.UserWalletListModel `json:"list"`
	TotalCount int                            `json:"totalCount"`
}

// ToogoOrderHistoryListReq 历史交易订单列表请求
type ToogoOrderHistoryListReq struct {
	g.Meta `path:"/toogo/wallet/order/history" method:"get" tags:"Toogo钱包" summary:"历史交易订单列表"`
	toogoin.OrderHistoryListInp
}

type ToogoOrderHistoryListRes struct {
	List       []*toogoin.OrderHistoryModel `json:"list"`
	TotalCount int                          `json:"totalCount"`
}

// ToogoTradeHistoryListReq 成交流水列表请求
type ToogoTradeHistoryListReq struct {
	g.Meta `path:"/toogo/wallet/trade/history" method:"get" tags:"Toogo钱包" summary:"成交流水列表（交易所成交明细）"`
	toogoin.TradeHistoryListInp
}

type ToogoTradeHistoryListRes struct {
	List       []*toogoin.TradeFillModel `json:"list"`
	TotalCount int                       `json:"totalCount"`
	Summary    *toogoin.TradeFillSummary `json:"summary"`
}

// ToogoRunSessionSummaryListReq 运行区间汇总列表请求
type ToogoRunSessionSummaryListReq struct {
	g.Meta `path:"/toogo/wallet/run-session/summary" method:"get" tags:"Toogo钱包" summary:"运行区间盈亏汇总"`
	toogoin.RunSessionSummaryListInp
}

type ToogoRunSessionSummaryListRes struct {
	List       []*toogoin.RunSessionSummaryModel `json:"list"`
	TotalCount int                               `json:"totalCount"`
	Summary    *toogoin.RunSessionTotalSummary   `json:"summary"`
}

// ToogoSyncRunSessionReq 同步运行区间盈亏请求
type ToogoSyncRunSessionReq struct {
	g.Meta    `path:"/toogo/wallet/run-session/sync" method:"post" tags:"Toogo钱包" summary:"同步运行区间盈亏数据"`
	SessionId int64 `json:"sessionId" v:"required" description:"区间ID"`
	// CalcOnly=1：仅根据本地成交流水(trading_trade_fill)按时间窗重算并写回 run_session，不调用交易所接口
	CalcOnly int `json:"calcOnly" description:"1=仅重算写回(不拉交易所)，0=先拉交易所再重算写回"`
}

type ToogoSyncRunSessionRes struct {
	TotalPnl   float64 `json:"totalPnl" description:"同步后总盈亏"`
	TotalFee   float64 `json:"totalFee" description:"同步后总手续费"`
	TradeCount int     `json:"tradeCount" description:"成交笔数"`
}

// ========== 订阅管理 ==========

// ToogoPlanListReq 套餐列表请求
type ToogoPlanListReq struct {
	g.Meta `path:"/toogo/plan/list" method:"get" tags:"Toogo订阅" summary:"套餐列表"`
	toogoin.PlanListInp
}

type ToogoPlanListRes struct {
	List       []*toogoin.PlanListModel `json:"list"`
	TotalCount int                      `json:"totalCount"`
}

// ToogoPlanEditReq 编辑套餐请求
type ToogoPlanEditReq struct {
	g.Meta `path:"/toogo/plan/edit" method:"post" tags:"Toogo订阅" summary:"编辑套餐"`
	toogoin.PlanEditInp
}

type ToogoPlanEditRes struct{}

// ToogoPlanDeleteReq 删除套餐请求
type ToogoPlanDeleteReq struct {
	g.Meta `path:"/toogo/plan/delete" method:"post" tags:"Toogo订阅" summary:"删除套餐"`
	toogoin.PlanDeleteInp
}

type ToogoPlanDeleteRes struct{}

// ToogoSubscribeReq 订阅套餐请求
type ToogoSubscribeReq struct {
	g.Meta `path:"/toogo/subscription/subscribe" method:"post" tags:"Toogo订阅" summary:"订阅套餐"`
	toogoin.SubscribeInp
}

type ToogoSubscribeRes struct {
	*toogoin.SubscribeModel
}

// ToogoSubscriptionListReq 订阅记录列表请求
type ToogoSubscriptionListReq struct {
	g.Meta `path:"/toogo/subscription/list" method:"get" tags:"Toogo订阅" summary:"订阅记录列表"`
	toogoin.SubscriptionListInp
}

type ToogoSubscriptionListRes struct {
	List       []*toogoin.SubscriptionListModel `json:"list"`
	TotalCount int                              `json:"totalCount"`
}

// ToogoMySubscriptionReq 我的订阅请求
type ToogoMySubscriptionReq struct {
	g.Meta `path:"/toogo/subscription/my" method:"get" tags:"Toogo订阅" summary:"我的订阅"`
}

type ToogoMySubscriptionRes struct {
	*toogoin.MySubscriptionModel
}

// ========== 用户管理 ==========

// ToogoUserInfoReq 用户信息请求
type ToogoUserInfoReq struct {
	g.Meta   `path:"/toogo/user/info" method:"get" tags:"Toogo用户" summary:"用户信息"`
	MemberId int64 `json:"memberId" dc:"用户ID，不传则获取当前登录用户"`
}

type ToogoUserInfoRes struct {
	*toogoin.UserInfoModel
}

// ToogoUserListReq 用户列表请求
type ToogoUserListReq struct {
	g.Meta `path:"/toogo/user/list" method:"get" tags:"Toogo用户" summary:"用户列表"`
	toogoin.UserListInp
}

type ToogoUserListRes struct {
	List       []*toogoin.UserListModel `json:"list"`
	TotalCount int                      `json:"totalCount"`
}

// ToogoRefreshInviteCodeReq 刷新邀请码请求
type ToogoRefreshInviteCodeReq struct {
	g.Meta `path:"/toogo/user/refresh-invite-code" method:"post" tags:"Toogo用户" summary:"刷新邀请码"`
}

type ToogoRefreshInviteCodeRes struct {
	*toogoin.InviteCodeRefreshModel
}

// ToogoTeamListReq 团队列表请求
type ToogoTeamListReq struct {
	g.Meta `path:"/toogo/user/team/list" method:"get" tags:"Toogo用户" summary:"团队列表"`
	toogoin.TeamListInp
}

type ToogoTeamListRes struct {
	List       []*toogoin.TeamListModel `json:"list"`
	TotalCount int                      `json:"totalCount"`
}

// ToogoTeamStatReq 团队统计请求
type ToogoTeamStatReq struct {
	g.Meta `path:"/toogo/user/team/stat" method:"get" tags:"Toogo用户" summary:"团队统计"`
}

type ToogoTeamStatRes struct {
	*toogoin.TeamStatModel
}

// ToogoVipLevelListReq VIP等级列表请求
type ToogoVipLevelListReq struct {
	g.Meta `path:"/toogo/vip-level/list" method:"get" tags:"Toogo用户" summary:"VIP等级列表"`
	toogoin.VipLevelListInp
}

type ToogoVipLevelListRes struct {
	List       []*toogoin.VipLevelListModel `json:"list"`
	TotalCount int                          `json:"totalCount"`
}

// ToogoVipLevelEditReq 编辑VIP等级请求
type ToogoVipLevelEditReq struct {
	g.Meta `path:"/toogo/vip-level/edit" method:"post" tags:"Toogo用户" summary:"编辑VIP等级"`
	toogoin.VipLevelEditInp
}

type ToogoVipLevelEditRes struct{}

// ToogoCheckVipUpgradeReq 检查VIP升级请求
type ToogoCheckVipUpgradeReq struct {
	g.Meta `path:"/toogo/user/check-vip-upgrade" method:"post" tags:"Toogo用户" summary:"检查VIP升级"`
}

type ToogoCheckVipUpgradeRes struct {
	*toogoin.CheckVipUpgradeModel
}

// ========== 佣金管理 ==========

// ToogoCommissionLogListReq 佣金记录列表请求
type ToogoCommissionLogListReq struct {
	g.Meta `path:"/toogo/commission/log/list" method:"get" tags:"Toogo佣金" summary:"佣金记录列表"`
	toogoin.CommissionLogListInp
}

type ToogoCommissionLogListRes struct {
	List       []*toogoin.CommissionLogListModel `json:"list"`
	TotalCount int                               `json:"totalCount"`
}

// ToogoCommissionStatReq 佣金统计请求
type ToogoCommissionStatReq struct {
	g.Meta `path:"/toogo/commission/stat" method:"get" tags:"Toogo佣金" summary:"佣金统计"`
}

type ToogoCommissionStatRes struct {
	*toogoin.CommissionStatModel
}

// ToogoAgentLevelListReq 代理商等级列表请求
type ToogoAgentLevelListReq struct {
	g.Meta `path:"/toogo/agent-level/list" method:"get" tags:"Toogo代理商" summary:"代理商等级列表"`
	toogoin.AgentLevelListInp
}

type ToogoAgentLevelListRes struct {
	List       []*toogoin.AgentLevelListModel `json:"list"`
	TotalCount int                            `json:"totalCount"`
}

// ToogoAgentLevelEditReq 编辑代理商等级请求
type ToogoAgentLevelEditReq struct {
	g.Meta `path:"/toogo/agent-level/edit" method:"post" tags:"Toogo代理商" summary:"编辑代理商等级"`
	toogoin.AgentLevelEditInp
}

type ToogoAgentLevelEditRes struct{}

// ToogoApplyAgentReq 申请成为代理商请求
type ToogoApplyAgentReq struct {
	g.Meta        `path:"/toogo/agent/apply" method:"post" tags:"Toogo代理商" summary:"申请成为代理商"`
	Remark        string  `json:"remark" dc:"申请备注/理由"`
	SubscribeRate float64 `json:"subscribeRate" v:"required|min:0.01|max:100" dc:"申请订阅返佣比例(%)"`
}

type ToogoApplyAgentRes struct {
	*toogoin.ApplyAgentModel
}

// ToogoApplyAgentForSubReq 代下级提交代理申请请求（用户端，仅直属下级）
type ToogoApplyAgentForSubReq struct {
	g.Meta        `path:"/toogo/agent/applyForSub" method:"post" tags:"Toogo代理商" summary:"代下级提交代理申请（仅直属下级）"`
	SubUserId     int64   `json:"subUserId" v:"required" dc:"下级用户ID"`
	Remark        string  `json:"remark" dc:"申请备注/理由"`
	SubscribeRate float64 `json:"subscribeRate" v:"required|min:0.01|max:100" dc:"申请订阅返佣比例(%)"`
}

type ToogoApplyAgentForSubRes struct{}

// ToogoApproveAgentReq 审批代理商申请请求（管理员）
type ToogoApproveAgentReq struct {
	g.Meta `path:"/toogo/agent/approve" method:"post" tags:"Toogo代理商" summary:"审批代理商申请（管理员）"`
	toogoin.ApproveAgentInp
}

type ToogoApproveAgentRes struct {
	*toogoin.ApproveAgentModel
}

// ToogoUpdateAgentReq 更新代理商信息请求（管理员）
type ToogoUpdateAgentReq struct {
	g.Meta `path:"/toogo/agent/update" method:"post" tags:"Toogo代理商" summary:"更新代理商信息（管理员）"`
	toogoin.UpdateAgentInp
}

type ToogoUpdateAgentRes struct{}

// ToogoSetSubAgentRateReq 设置下级代理佣金比例请求
type ToogoSetSubAgentRateReq struct {
	g.Meta        `path:"/toogo/agent/setSubRate" method:"post" tags:"Toogo代理商" summary:"设置下级代理佣金比例"`
	SubUserId     int64   `json:"subUserId" v:"required" dc:"下级用户ID"`
	SubscribeRate float64 `json:"subscribeRate" v:"required|min:0|max:100" dc:"订阅返佣比例(%)"`
}

type ToogoSetSubAgentRateRes struct{}

// ToogoGetAgentInfoReq 获取代理信息请求
type ToogoGetAgentInfoReq struct {
	g.Meta `path:"/toogo/agent/info" method:"get" tags:"Toogo代理商" summary:"获取代理信息"`
}

type ToogoGetAgentInfoRes struct {
	*toogoin.AgentInfoModel
}

// ========== 策略管理 ==========

// ToogoStrategyTemplateListReq 策略模板列表请求
type ToogoStrategyTemplateListReq struct {
	g.Meta `path:"/toogo/strategy/template/list" method:"get" tags:"Toogo策略" summary:"策略模板列表"`
	toogoin.StrategyTemplateListInp
}

type ToogoStrategyTemplateListRes struct {
	List       []*toogoin.StrategyTemplateListModel `json:"list"`
	TotalCount int                                  `json:"totalCount"`
}

// ToogoStrategyTemplateEditReq 编辑策略模板请求
type ToogoStrategyTemplateEditReq struct {
	g.Meta `path:"/toogo/strategy/template/edit" method:"post" tags:"Toogo策略" summary:"编辑策略模板"`
	toogoin.StrategyTemplateEditInp
}

type ToogoStrategyTemplateEditRes struct{}

// ToogoStrategyTemplateDeleteReq 删除策略模板请求
type ToogoStrategyTemplateDeleteReq struct {
	g.Meta `path:"/toogo/strategy/template/delete" method:"post" tags:"Toogo策略" summary:"删除策略模板"`
	toogoin.StrategyTemplateDeleteInp
}

type ToogoStrategyTemplateDeleteRes struct{}

// ToogoGetStrategyReq 获取策略请求
type ToogoGetStrategyReq struct {
	g.Meta `path:"/toogo/strategy/get" method:"get" tags:"Toogo策略" summary:"获取策略"`
	toogoin.GetStrategyByConditionInp
}

type ToogoGetStrategyRes struct {
	*toogoin.GetStrategyByConditionModel
}

// ToogoPowerConsumeListReq 算力消耗记录列表请求
type ToogoPowerConsumeListReq struct {
	g.Meta `path:"/toogo/power-consume/list" method:"get" tags:"Toogo策略" summary:"算力消耗记录列表"`
	toogoin.PowerConsumeListInp
}

type ToogoPowerConsumeListRes struct {
	List       []*toogoin.PowerConsumeListModel `json:"list"`
	TotalCount int                              `json:"totalCount"`
}

// ToogoPowerConsumeStatReq 算力消耗统计请求
type ToogoPowerConsumeStatReq struct {
	g.Meta `path:"/toogo/power-consume/stat" method:"get" tags:"Toogo策略" summary:"算力消耗统计"`
}

type ToogoPowerConsumeStatRes struct {
	*toogoin.PowerConsumeStatModel
}

// ========== 管理员操作 ==========

// ToogoAdminRechargePowerReq 管理员手动充值算力请求
type ToogoAdminRechargePowerReq struct {
	g.Meta `path:"/toogo/admin/recharge-power" method:"post" tags:"Toogo管理" summary:"手动充值算力"`
	toogoin.AdminRechargePowerInp
}

type ToogoAdminRechargePowerRes struct {
	*toogoin.AdminRechargePowerModel
}

// ToogoAdminRechargeBalanceReq 管理员手动充值余额请求
type ToogoAdminRechargeBalanceReq struct {
	g.Meta `path:"/toogo/admin/recharge-balance" method:"post" tags:"Toogo管理" summary:"手动充值余额"`
	toogoin.AdminRechargeBalanceInp
}

type ToogoAdminRechargeBalanceRes struct {
	*toogoin.AdminRechargeBalanceModel
}

// ToogoAdminRechargePointsReq 管理员手动充值积分请求
type ToogoAdminRechargePointsReq struct {
	g.Meta `path:"/toogo/admin/recharge-points" method:"post" tags:"Toogo管理" summary:"手动充值积分"`
	toogoin.AdminRechargePointsInp
}

type ToogoAdminRechargePointsRes struct {
	*toogoin.AdminRechargePointsModel
}
