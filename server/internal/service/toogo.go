// Package service
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package service

import (
	"context"

	"hotgo/internal/library/exchange"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"
)

// IToogoWallet Toogo钱包服务接口
type IToogoWallet interface {
	// GetOrCreate 获取或创建钱包
	GetOrCreate(ctx context.Context, userId int64) (*entity.ToogoWallet, error)
	// GetOverview 获取钱包概览
	GetOverview(ctx context.Context, in *toogoin.WalletOverviewInp) (*toogoin.WalletOverviewModel, error)
	// ChangeBalance 变更余额
	ChangeBalance(ctx context.Context, in *toogoin.ChangeBalanceInp) error
	// WalletLogList 钱包流水列表
	WalletLogList(ctx context.Context, in *toogoin.WalletLogListInp) ([]*toogoin.WalletLogListModel, int, error)
	// Transfer 账户互转
	Transfer(ctx context.Context, in *toogoin.TransferInp) (*toogoin.TransferModel, error)
	// AdminRecharge 管理员手动充值
	AdminRecharge(ctx context.Context, userId int64, accountType string, amount float64, remark string) (beforeAmount, afterAmount float64, err error)
	// UserWalletList 用户钱包列表（管理员用）
	UserWalletList(ctx context.Context, username, mobile string, page, perPage int) ([]*toogoin.UserWalletListModel, int, error)
	// ConsumePower 消耗算力
	ConsumePower(ctx context.Context, userId int64, robotId int64, orderId int64, orderSn string, profitAmount float64) error
	// OrderHistoryList 历史交易订单列表
	OrderHistoryList(ctx context.Context, in *toogoin.OrderHistoryListInp) ([]*toogoin.OrderHistoryModel, int, error)
	// TradeHistoryList 成交流水列表（交易所成交明细）
	TradeHistoryList(ctx context.Context, in *toogoin.TradeHistoryListInp) ([]*toogoin.TradeFillModel, int, *toogoin.TradeFillSummary, error)
	// RunSessionSummaryList 运行区间盈亏汇总列表
	RunSessionSummaryList(ctx context.Context, in *toogoin.RunSessionSummaryListInp) ([]*toogoin.RunSessionSummaryModel, int, *toogoin.RunSessionTotalSummary, error)
	// SyncRunSession 同步/重算运行区间盈亏数据
	// - calcOnly=true: 仅按本地成交流水(trading_trade_fill)时间窗重算并写回 run_session（不调用交易所）
	// - calcOnly=false: 先从交易所拉取成交落库，再按本地成交流水时间窗重算写回 run_session
	SyncRunSession(ctx context.Context, sessionId int64, calcOnly bool) (totalPnl, totalFee float64, tradeCount int, err error)
	// StartTradeFillSyncTask 启动成交流水后台同步任务
	StartTradeFillSyncTask(ctx context.Context)
}

var localToogoWallet IToogoWallet

func ToogoWallet() IToogoWallet {
	if localToogoWallet == nil {
		panic("implement not found for interface IToogoWallet, forgot register?")
	}
	return localToogoWallet
}

func RegisterToogoWallet(i IToogoWallet) {
	localToogoWallet = i
}

// IToogoUser Toogo用户服务接口
type IToogoUser interface {
	// GetOrCreate 获取或创建用户
	GetOrCreate(ctx context.Context, memberId int64) (*entity.ToogoUser, error)
	// GenerateInviteCode 生成邀请码
	GenerateInviteCode() string
	// UserInfo 获取用户信息
	UserInfo(ctx context.Context, in *toogoin.UserInfoInp) (*toogoin.UserInfoModel, error)
	// UserList 用户列表
	UserList(ctx context.Context, in *toogoin.UserListInp) ([]*toogoin.UserListModel, int, error)
	// RefreshInviteCode 刷新邀请码
	RefreshInviteCode(ctx context.Context, in *toogoin.InviteCodeRefreshInp) (*toogoin.InviteCodeRefreshModel, error)
	// RegisterWithInvite 使用邀请码注册关联
	RegisterWithInvite(ctx context.Context, in *toogoin.RegisterWithInviteInp) error
	// UpdateTeamCount 更新团队人数
	UpdateTeamCount(ctx context.Context, memberId int64, delta int) error
	// TeamList 团队列表
	TeamList(ctx context.Context, in *toogoin.TeamListInp) ([]*toogoin.TeamListModel, int, error)
	// TeamStat 团队统计
	TeamStat(ctx context.Context, in *toogoin.TeamStatInp) (*toogoin.TeamStatModel, error)
	// VipLevelList VIP等级列表
	VipLevelList(ctx context.Context, in *toogoin.VipLevelListInp) ([]*toogoin.VipLevelListModel, int, error)
	// VipLevelEdit 编辑VIP等级
	VipLevelEdit(ctx context.Context, in *toogoin.VipLevelEditInp) error
	// CheckVipUpgrade 检查VIP升级
	CheckVipUpgrade(ctx context.Context, in *toogoin.CheckVipUpgradeInp) (*toogoin.CheckVipUpgradeModel, error)
}

var localToogoUser IToogoUser

func ToogoUser() IToogoUser {
	if localToogoUser == nil {
		panic("implement not found for interface IToogoUser, forgot register?")
	}
	return localToogoUser
}

func RegisterToogoUser(i IToogoUser) {
	localToogoUser = i
}

// IToogoSubscription Toogo订阅服务接口
type IToogoSubscription interface {
	// PlanList 套餐列表
	PlanList(ctx context.Context, in *toogoin.PlanListInp) ([]*toogoin.PlanListModel, int, error)
	// PlanEdit 编辑套餐
	PlanEdit(ctx context.Context, in *toogoin.PlanEditInp) error
	// PlanDelete 删除套餐
	PlanDelete(ctx context.Context, in *toogoin.PlanDeleteInp) error
	// Subscribe 订阅套餐
	Subscribe(ctx context.Context, in *toogoin.SubscribeInp) (*toogoin.SubscribeModel, error)
	// SubscriptionList 订阅记录列表
	SubscriptionList(ctx context.Context, in *toogoin.SubscriptionListInp) ([]*toogoin.SubscriptionListModel, int, error)
	// MySubscription 我的订阅
	MySubscription(ctx context.Context, in *toogoin.MySubscriptionInp) (*toogoin.MySubscriptionModel, error)
	// CheckExpired 检查并处理过期订阅
	CheckExpired(ctx context.Context) error
}

var localToogoSubscription IToogoSubscription

func ToogoSubscription() IToogoSubscription {
	if localToogoSubscription == nil {
		panic("implement not found for interface IToogoSubscription, forgot register?")
	}
	return localToogoSubscription
}

func RegisterToogoSubscription(i IToogoSubscription) {
	localToogoSubscription = i
}

// IToogoCommission Toogo佣金服务接口
type IToogoCommission interface {
	// CommissionLogList 佣金记录列表
	CommissionLogList(ctx context.Context, in *toogoin.CommissionLogListInp) ([]*toogoin.CommissionLogListModel, int, error)
	// CommissionStat 佣金统计
	CommissionStat(ctx context.Context, in *toogoin.CommissionStatInp) (*toogoin.CommissionStatModel, error)
	// SettleSubscribeCommission 结算订阅佣金（级差制）
	SettleSubscribeCommission(ctx context.Context, fromUserId int64, amount float64, subscriptionId int64, orderSn string) error
	// SettleInviteReward 发放邀请奖励
	SettleInviteReward(ctx context.Context, inviterId int64, inviteeId int64) error
	// AgentLevelList 代理商等级列表（已废弃）
	AgentLevelList(ctx context.Context, in *toogoin.AgentLevelListInp) ([]*toogoin.AgentLevelListModel, int, error)
	// AgentLevelEdit 编辑代理商等级（已废弃）
	AgentLevelEdit(ctx context.Context, in *toogoin.AgentLevelEditInp) error
	// ApplyAgent 申请成为代理商
	ApplyAgent(ctx context.Context, in *toogoin.ApplyAgentInp) (*toogoin.ApplyAgentModel, error)
	// ApplyAgentForSub 代下级提交代理申请（仅直属下级）
	ApplyAgentForSub(ctx context.Context, in *toogoin.ApplyAgentForSubInp) error
	// ApproveAgent 审批代理商申请（管理员）
	ApproveAgent(ctx context.Context, in *toogoin.ApproveAgentInp, operatorId int64) (*toogoin.ApproveAgentModel, error)
	// UpdateAgent 更新代理商信息（管理员）
	UpdateAgent(ctx context.Context, in *toogoin.UpdateAgentInp) error
	// SetSubAgentRate 设置下级代理佣金比例
	SetSubAgentRate(ctx context.Context, in *toogoin.SetSubAgentRateInp) error
	// GetAgentInfo 获取代理信息
	GetAgentInfo(ctx context.Context, memberId int64) (*toogoin.AgentInfoModel, error)
}

var localToogoCommission IToogoCommission

func ToogoCommission() IToogoCommission {
	if localToogoCommission == nil {
		panic("implement not found for interface IToogoCommission, forgot register?")
	}
	return localToogoCommission
}

func RegisterToogoCommission(i IToogoCommission) {
	localToogoCommission = i
}

// IToogoStrategy Toogo策略服务接口
type IToogoStrategy interface {
	// TemplateList 策略模板列表
	TemplateList(ctx context.Context, in *toogoin.StrategyTemplateListInp) ([]*toogoin.StrategyTemplateListModel, int, error)
	// TemplateEdit 编辑策略模板
	TemplateEdit(ctx context.Context, in *toogoin.StrategyTemplateEditInp) error
	// TemplateDelete 删除策略模板
	TemplateDelete(ctx context.Context, in *toogoin.StrategyTemplateDeleteInp) error
	// GetByCondition 根据条件获取策略
	GetByCondition(ctx context.Context, in *toogoin.GetStrategyByConditionInp) (*toogoin.GetStrategyByConditionModel, error)
	// PowerConsumeList 算力消耗记录列表
	PowerConsumeList(ctx context.Context, in *toogoin.PowerConsumeListInp) ([]*toogoin.PowerConsumeListModel, int, error)
	// PowerConsumeStat 算力消耗统计
	PowerConsumeStat(ctx context.Context, in *toogoin.PowerConsumeStatInp) (*toogoin.PowerConsumeStatModel, error)
	// AnalyzeMarketState 分析市场状态
	AnalyzeMarketState(ctx context.Context, symbol string, strategy *entity.ToogoStrategyTemplate) (marketState string, riskPreference string, err error)
	// GetOptimalStrategy 获取最优策略
	GetOptimalStrategy(ctx context.Context, symbol string) (*entity.ToogoStrategyTemplate, error)
}

var localToogoStrategy IToogoStrategy

func ToogoStrategy() IToogoStrategy {
	if localToogoStrategy == nil {
		panic("implement not found for interface IToogoStrategy, forgot register?")
	}
	return localToogoStrategy
}

func RegisterToogoStrategy(i IToogoStrategy) {
	localToogoStrategy = i
}

// IToogoFinance Toogo财务服务接口
type IToogoFinance interface {
	// CreateDeposit 创建充值订单
	CreateDeposit(ctx context.Context, in *toogoin.CreateDepositInp) (*toogoin.CreateDepositModel, error)
	// DepositCallback 充值回调
	DepositCallback(ctx context.Context, in *toogoin.DepositCallbackInp) error
	// DepositList 充值记录列表
	DepositList(ctx context.Context, in *toogoin.DepositListInp) ([]*toogoin.DepositListModel, int, error)
	// CreateWithdraw 创建提现申请
	CreateWithdraw(ctx context.Context, in *toogoin.CreateWithdrawInp) (*toogoin.CreateWithdrawModel, error)
	// WithdrawList 提现记录列表
	WithdrawList(ctx context.Context, in *toogoin.WithdrawListInp) ([]*toogoin.WithdrawListModel, int, error)
	// WithdrawAudit 提现审核
	WithdrawAudit(ctx context.Context, in *toogoin.WithdrawAuditInp) error
	// WithdrawComplete 提现完成回调
	WithdrawComplete(ctx context.Context, in *toogoin.WithdrawCompleteInp) error
	// HandleNOWPaymentsIPNCallback 处理NOWPayments充值回调
	HandleNOWPaymentsIPNCallback(ctx context.Context) error
	// HandleNOWPaymentsPayoutIPNCallback 处理NOWPayments提现回调
	HandleNOWPaymentsPayoutIPNCallback(ctx context.Context) error
}

var localToogoFinance IToogoFinance

func ToogoFinance() IToogoFinance {
	if localToogoFinance == nil {
		panic("implement not found for interface IToogoFinance, forgot register?")
	}
	return localToogoFinance
}

func RegisterToogoFinance(i IToogoFinance) {
	localToogoFinance = i
}

// IToogoRobot Toogo机器人服务接口
type IToogoRobot interface {
	// StartRobot 启动机器人
	StartRobot(ctx context.Context, in *toogoin.StartRobotInp) error
	// StopRobot 停止机器人
	StopRobot(ctx context.Context, in *toogoin.StopRobotInp) error
	// RunRobotEngine 机器人运行引擎
	RunRobotEngine(ctx context.Context) error
	// RobotList 机器人列表
	RobotList(ctx context.Context, in *toogoin.RobotListInp) ([]*toogoin.RobotListModel, int, error)
	// GetRobotPositions 获取机器人当前持仓
	GetRobotPositions(ctx context.Context, robotId int64) ([]*toogoin.PositionModel, error)
	// GetRobotOpenOrders 获取机器人当前挂单
	GetRobotOpenOrders(ctx context.Context, robotId int64) ([]*toogoin.OrderModel, error)
	// GetRobotOrderHistory 获取机器人历史订单
	GetRobotOrderHistory(ctx context.Context, robotId int64, limit int) ([]*toogoin.OrderModel, error)
	// CloseRobotPosition 手动平仓
	CloseRobotPosition(ctx context.Context, in *toogoin.ClosePositionInp) error
	// CancelRobotOrder 撤销挂单
	CancelRobotOrder(ctx context.Context, robotId int64, orderId string) error
	// SetTakeProfitRetreatSwitch 设置止盈回撤开关
	SetTakeProfitRetreatSwitch(ctx context.Context, robotId int64, positionSide string, enabled bool) error
	// SyncClosedOrders 同步已平仓订单并补扣算力
	SyncClosedOrders(ctx context.Context) error
	// SyncOrderHistoryToDB 同步订单历史到数据库
	SyncOrderHistoryToDB(ctx context.Context, robotId int64, robot *entity.TradingRobot, orders []*exchange.Order) error
}

var localToogoRobot IToogoRobot

func ToogoRobot() IToogoRobot {
	if localToogoRobot == nil {
		panic("implement not found for interface IToogoRobot, forgot register?")
	}
	return localToogoRobot
}

func RegisterToogoRobot(i IToogoRobot) {
	localToogoRobot = i
}

// IToogo Toogo总服务接口（用于获取子模块管理器）
type IToogo interface {
	// GetRobotTaskManager 获取机器人任务管理器
	GetRobotTaskManager() IToogoTaskManager
}

// IToogoTaskManager 机器人任务管理器接口
type IToogoTaskManager interface {
	// ReloadRobotStrategy 重新加载机器人策略配置
	ReloadRobotStrategy(ctx context.Context, robotId int64) error
	// Start 启动任务管理器
	Start(ctx context.Context) error
	// Stop 停止任务管理器
	Stop()
	// IsRunning 检查是否运行中
	IsRunning() bool
}

var localToogo IToogo

func Toogo() IToogo {
	if localToogo == nil {
		panic("implement not found for interface IToogo, forgot register?")
	}
	return localToogo
}

func RegisterToogo(i IToogo) {
	localToogo = i
}

// IToogoVolatilityConfig Toogo波动率配置服务接口（支持每个货币对独立配置）
type IToogoVolatilityConfig interface {
	// List 波动率配置列表（支持按货币对筛选）
	List(ctx context.Context, in *toogoin.VolatilityConfigListInp) ([]*entity.ToogoVolatilityConfig, int, error)
	// Edit 编辑波动率配置（支持全局配置和货币对特定配置）
	Edit(ctx context.Context, in *toogoin.VolatilityConfigEditInp) error
	// BatchEdit 批量编辑波动率配置（为多个货币对批量设置相同配置）
	BatchEdit(ctx context.Context, in *toogoin.VolatilityConfigBatchEditInp) error
	// Delete 删除波动率配置
	Delete(ctx context.Context, id int64) error
	// GetBySymbol 获取波动率配置（优先货币对特定配置，其次全局配置）
	GetBySymbol(ctx context.Context, symbol string) (*entity.ToogoVolatilityConfig, error)
	// GetAllSymbols 获取所有已配置的交易对列表
	GetAllSymbols(ctx context.Context) ([]string, error)
}

var localToogoVolatilityConfig IToogoVolatilityConfig

func ToogoVolatilityConfig() IToogoVolatilityConfig {
	if localToogoVolatilityConfig == nil {
		panic("implement not found for interface IToogoVolatilityConfig, forgot register?")
	}
	return localToogoVolatilityConfig
}

func RegisterToogoVolatilityConfig(i IToogoVolatilityConfig) {
	localToogoVolatilityConfig = i
}
