// Package toogoin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogoin

import (
	"github.com/gogf/gf/v2/os/gtime"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// UserInfoInp 用户信息输入
type UserInfoInp struct {
	MemberId int64 `json:"memberId" description:"用户ID"`
}

// UserInfoModel 用户信息返回
type UserInfoModel struct {
	*entity.ToogoUser
	Username     string `json:"username" description:"用户名"`
	RealName     string `json:"realName" description:"真实姓名"`
	Avatar       string `json:"avatar" description:"头像"`
	VipLevelName string `json:"vipLevelName" description:"VIP等级名称"`
}

// UserListInp 用户列表输入
type UserListInp struct {
	form.PageReq
	Username  string   `json:"username" description:"用户名"`
	VipLevel  int      `json:"vipLevel" description:"VIP等级"`
	IsAgent   int      `json:"isAgent" d:"-1" description:"是否代理商: -1=全部, 0=否, 1=是"`
	Status    int      `json:"status" description:"状态"`
	CreatedAt []string `json:"createdAt" description:"创建时间"`
}

// UserListModel 用户列表返回
type UserListModel struct {
	*entity.ToogoUser
	Username     string  `json:"username" description:"用户名"`
	RealName     string  `json:"realName" description:"真实姓名"`
	Avatar       string  `json:"avatar" description:"头像"`
	VipLevelName string  `json:"vipLevelName" description:"VIP等级名称"`
	Balance      float64 `json:"balance" description:"余额(USDT)"`
	Power        float64 `json:"power" description:"算力余额"`
	GiftPower    float64 `json:"giftPower" description:"积分余额"`
	Commission   float64 `json:"commission" description:"佣金余额(USDT)"`
}

// InviteCodeRefreshInp 刷新邀请码输入
type InviteCodeRefreshInp struct {
	MemberId int64 `json:"memberId" v:"required" description:"用户ID"`
}

// InviteCodeRefreshModel 刷新邀请码返回
type InviteCodeRefreshModel struct {
	InviteCode       string `json:"inviteCode" description:"邀请码"`
	InviteCodeExpire string `json:"inviteCodeExpire" description:"过期时间"`
	InviteUrl        string `json:"inviteUrl" description:"邀请链接"`
}

// RegisterWithInviteInp 使用邀请码注册输入
type RegisterWithInviteInp struct {
	InviteCode string `json:"inviteCode" v:"required" description:"邀请码"`
	MemberId   int64  `json:"memberId" v:"required" description:"新用户ID"`
}

// TeamListInp 团队列表输入
type TeamListInp struct {
	form.PageReq
	MemberId int64 `json:"memberId" description:"用户ID（不传则取当前登录用户）"`
	Level    int   `json:"level" description:"层级: 1=直推, 2=二级, 3=三级"`
}

// TeamListModel 团队列表返回
type TeamListModel struct {
	MemberId         int64       `json:"memberId" description:"用户ID"`
	Username         string      `json:"username" description:"用户名"`
	Avatar           string      `json:"avatar" description:"头像"`
	VipLevel         int         `json:"vipLevel" description:"VIP等级"`
	VipLevelName     string      `json:"vipLevelName" description:"VIP等级名称"`
	Level            int         `json:"level" description:"层级"`
	IsAgent          int         `json:"isAgent" description:"是否代理: 0=否, 1=是"`
	AgentStatus      int         `json:"agentStatus" description:"代理商状态: 0=未申请, 1=待审批, 2=已通过, 3=已拒绝"`
	SubscribeRate    float64     `json:"subscribeRate" description:"订阅返佣比例(%)"`
	InviteCount      int         `json:"inviteCount" description:"推广人数（直推人数）"`
	TeamCount        int         `json:"teamCount" description:"团队总人数"`
	CurrentPlanId    int64       `json:"currentPlanId" description:"当前订阅套餐ID"`
	PlanExpireTime   *gtime.Time `json:"planExpireTime" description:"套餐到期时间"`
	RobotLimit       int         `json:"robotLimit" description:"机器人数量限制"`
	ActiveRobotCount int         `json:"activeRobotCount" description:"运行中机器人数量"`
	TotalConsume     float64     `json:"totalConsume" description:"累计消耗算力"`
	TotalSubscribe   float64     `json:"totalSubscribe" description:"累计订阅金额"`
	TotalCommission  float64     `json:"totalCommission" description:"累计贡献佣金"`
	RegisterTime     string      `json:"registerTime" description:"注册时间"`
}

// TeamStatInp 团队统计输入
type TeamStatInp struct {
	MemberId int64 `json:"memberId" v:"required" description:"用户ID"`
}

// TeamStatModel 团队统计返回
type TeamStatModel struct {
	DirectCount     int     `json:"directCount" description:"直推人数"`
	Level2Count     int     `json:"level2Count" description:"二级人数"`
	Level3Count     int     `json:"level3Count" description:"三级人数"`
	TotalCount      int     `json:"totalCount" description:"团队总人数"`
	TotalConsume    float64 `json:"totalConsume" description:"团队总消耗算力"`
	TotalSubscribe  float64 `json:"totalSubscribe" description:"团队总订阅金额"`
	TotalCommission float64 `json:"totalCommission" description:"累计获得佣金"`
}

// VipLevelListInp VIP等级列表输入
type VipLevelListInp struct {
	form.PageReq
	Status int `json:"status" description:"状态"`
}

// VipLevelListModel VIP等级列表返回
type VipLevelListModel struct {
	*entity.ToogoVipLevel
}

// VipLevelEditInp 编辑VIP等级输入
type VipLevelEditInp struct {
	Id                  int64   `json:"id" description:"ID"`
	Level               int     `json:"level" v:"required|min:1|max:10" description:"等级"`
	LevelName           string  `json:"levelName" v:"required" description:"等级名称"`
	RequireInviteCount  int     `json:"requireInviteCount" description:"需要邀请人数"`
	RequireConsumePower float64 `json:"requireConsumePower" description:"需要消耗算力"`
	RequireTeamConsume  float64 `json:"requireTeamConsume" description:"需要团队消耗算力"`
	PowerDiscount       float64 `json:"powerDiscount" v:"min:0|max:30" description:"算力折扣"`
	InviteRewardPower   float64 `json:"inviteRewardPower" description:"邀请奖励算力"`
	Description         string  `json:"description" description:"等级描述"`
	Icon                string  `json:"icon" description:"等级图标"`
	Sort                int     `json:"sort" description:"排序"`
	Status              int     `json:"status" description:"状态"`
}

// CheckVipUpgradeInp 检查VIP升级输入
type CheckVipUpgradeInp struct {
	MemberId int64 `json:"memberId" v:"required" description:"用户ID"`
}

// CheckVipUpgradeModel 检查VIP升级返回
type CheckVipUpgradeModel struct {
	CanUpgrade    bool   `json:"canUpgrade" description:"是否可升级"`
	CurrentLevel  int    `json:"currentLevel" description:"当前等级"`
	NextLevel     int    `json:"nextLevel" description:"下一等级"`
	NextLevelName string `json:"nextLevelName" description:"下一等级名称"`
	Progress      struct {
		InviteCount    int     `json:"inviteCount" description:"当前邀请人数"`
		RequireInvite  int     `json:"requireInvite" description:"需要邀请人数"`
		ConsumePower   float64 `json:"consumePower" description:"当前消耗算力"`
		RequireConsume float64 `json:"requireConsume" description:"需要消耗算力"`
		TeamConsume    float64 `json:"teamConsume" description:"当前团队消耗"`
		RequireTeam    float64 `json:"requireTeam" description:"需要团队消耗"`
	} `json:"progress" description:"升级进度"`
}
