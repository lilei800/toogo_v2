// Package toogoin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogoin

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// CommissionLogListInp 佣金记录列表输入
type CommissionLogListInp struct {
	form.PageReq
	UserId         int64    `json:"userId" description:"用户ID"`
	CommissionType string   `json:"commissionType" description:"佣金类型"`
	Level          int      `json:"level" description:"层级"`
	CreatedAt      []string `json:"createdAt" description:"创建时间"`
}

// CommissionLogListModel 佣金记录列表返回
type CommissionLogListModel struct {
	*entity.ToogoCommissionLog
	FromUsername string `json:"fromUsername" description:"来源用户名"`
}

// CommissionStatInp 佣金统计输入
type CommissionStatInp struct {
	UserId int64 `json:"userId" v:"required" description:"用户ID"`
}

// CommissionStatModel 佣金统计返回
type CommissionStatModel struct {
	TotalCommission   float64 `json:"totalCommission" description:"累计获得佣金"`
	TodayCommission   float64 `json:"todayCommission" description:"今日佣金"`
	WeekCommission    float64 `json:"weekCommission" description:"本周佣金"`
	MonthCommission   float64 `json:"monthCommission" description:"本月佣金"`
	InviteReward      float64 `json:"inviteReward" description:"邀请奖励(积分)"`
	SubscribeCommission float64 `json:"subscribeCommission" description:"订阅佣金"`
}

// SettleCommissionInp 结算佣金输入(内部使用)
type SettleCommissionInp struct {
	FromUserId     int64   `json:"fromUserId" description:"来源用户ID"`
	CommissionType string  `json:"commissionType" description:"佣金类型"`
	BaseAmount     float64 `json:"baseAmount" description:"基础金额"`
	RelatedId      int64   `json:"relatedId" description:"关联ID"`
	RelatedType    string  `json:"relatedType" description:"关联类型"`
	OrderSn        string  `json:"orderSn" description:"关联订单号"`
}

// AgentLevelListInp 代理商等级列表输入
type AgentLevelListInp struct {
	form.PageReq
	Status int `json:"status" description:"状态"`
}

// AgentLevelListModel 代理商等级列表返回
type AgentLevelListModel struct {
	*entity.ToogoAgentLevel
}

// AgentLevelEditInp 编辑代理商等级输入
type AgentLevelEditInp struct {
	Id                   int64   `json:"id" description:"ID"`
	Level                int     `json:"level" v:"required|min:1|max:5" description:"等级"`
	LevelName            string  `json:"levelName" v:"required" description:"等级名称"`
	RequireTeamCount     int     `json:"requireTeamCount" description:"需要团队人数"`
	RequireTeamSubscribe float64 `json:"requireTeamSubscribe" description:"需要团队订阅额"`
	SubscribeRate1       float64 `json:"subscribeRate1" description:"订阅佣金比例(一级)"`
	SubscribeRate2       float64 `json:"subscribeRate2" description:"订阅佣金比例(二级)"`
	SubscribeRate3       float64 `json:"subscribeRate3" description:"订阅佣金比例(三级)"`
	PowerRate1           float64 `json:"powerRate1" description:"算力消耗佣金比例(一级)"`
	PowerRate2           float64 `json:"powerRate2" description:"算力消耗佣金比例(二级)"`
	PowerRate3           float64 `json:"powerRate3" description:"算力消耗佣金比例(三级)"`
	Description          string  `json:"description" description:"等级描述"`
	Sort                 int     `json:"sort" description:"排序"`
	Status               int     `json:"status" description:"状态"`
}

// ApplyAgentInp 申请成为代理商输入
type ApplyAgentInp struct {
	MemberId       int64   `json:"memberId" v:"required" description:"用户ID"`
	Remark         string  `json:"remark" description:"申请备注/理由"`
	SubscribeRate  float64 `json:"subscribeRate" v:"required|min:0.01|max:100" description:"申请订阅返佣比例(%)"`
}

// ApplyAgentForSubInp 代下级提交代理申请输入（仅直属下级）
type ApplyAgentForSubInp struct {
	AgentId        int64   `json:"agentId" v:"required" description:"当前用户ID"`
	SubUserId      int64   `json:"subUserId" v:"required" description:"下级用户ID"`
	Remark         string  `json:"remark" description:"申请备注/理由"`
	SubscribeRate  float64 `json:"subscribeRate" v:"required|min:0.01|max:100" description:"申请订阅返佣比例(%)"`
}

// ApplyAgentModel 申请成为代理商返回
type ApplyAgentModel struct {
	Success       bool    `json:"success" description:"是否成功"`
	AgentStatus   int     `json:"agentStatus" description:"代理商状态: 0=未申请, 1=待审批, 2=已通过, 3=已拒绝"`
	Message       string  `json:"message" description:"提示信息"`
	SubscribeRate float64 `json:"subscribeRate" description:"订阅返佣比例(%)"`
}

// ApproveAgentInp 审批代理商申请输入
type ApproveAgentInp struct {
	MemberId         int64   `json:"memberId" v:"required" description:"用户ID"`
	Approved         bool    `json:"approved" v:"required" description:"是否通过"`
	SubscribeRate    float64 `json:"subscribeRate" description:"订阅返佣比例(%)，通过时必填"`
	AgentUnlockLevel int     `json:"agentUnlockLevel" description:"层级解锁: 0=仅一级佣金, 1=无限级佣金"`
	RejectReason     string  `json:"rejectReason" description:"拒绝原因"`
}

// ApproveAgentModel 审批代理商申请返回
type ApproveAgentModel struct {
	Success bool   `json:"success" description:"是否成功"`
	Message string `json:"message" description:"提示信息"`
}

// UpdateAgentInp 更新代理商信息输入（管理员使用）
type UpdateAgentInp struct {
	MemberId         int64   `json:"memberId" v:"required" description:"用户ID"`
	IsAgent          int     `json:"isAgent" description:"是否代理商: 0=否, 1=是"`
	AgentStatus      int     `json:"agentStatus" description:"代理商状态"`
	AgentUnlockLevel int     `json:"agentUnlockLevel" description:"层级解锁: 0=仅一级, 1=无限级"`
	SubscribeRate    float64 `json:"subscribeRate" description:"订阅返佣比例(%)"`
}

// SetSubAgentRateInp 设置下级代理佣金比例输入
type SetSubAgentRateInp struct {
	AgentId       int64   `json:"agentId" v:"required" description:"当前代理ID"`
	SubUserId     int64   `json:"subUserId" v:"required" description:"下级用户ID"`
	SubscribeRate float64 `json:"subscribeRate" v:"required|min:0|max:100" description:"订阅返佣比例(%)"`
}

// SubAgentInfo 下级代理信息
type SubAgentInfo struct {
	UserId        int64   `json:"userId" description:"用户ID"`
	Username      string  `json:"username" description:"用户名"`
	IsAgent       int     `json:"isAgent" description:"是否代理"`
	SubscribeRate float64 `json:"subscribeRate" description:"订阅返佣比例(%)"`
}

// AgentInfoModel 代理信息返回
type AgentInfoModel struct {
	IsAgent          int             `json:"isAgent" description:"是否代理"`
	AgentStatus      int             `json:"agentStatus" description:"代理商状态: 0=未申请, 1=待审批, 2=已通过, 3=已拒绝"`
	AgentUnlockLevel int             `json:"agentUnlockLevel" description:"层级解锁: 0=仅一级佣金, 1=无限级佣金"`
	SubscribeRate    float64         `json:"subscribeRate" description:"订阅返佣比例(%)"`
	CanSetSubRate    bool            `json:"canSetSubRate" description:"是否可以设置下级佣金比例"`
	SubAgents        []*SubAgentInfo `json:"subAgents" description:"下级代理列表"`
}

