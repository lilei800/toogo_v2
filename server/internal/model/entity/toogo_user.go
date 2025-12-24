// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoUser is the golang structure for table hg_toogo_user.
type ToogoUser struct {
	Id                  int64       `json:"id"                  orm:"id"                    description:"主键ID"`
	MemberId            int64       `json:"memberId"            orm:"member_id"             description:"关联admin_member.id"`
	VipLevel            int         `json:"vipLevel"            orm:"vip_level"             description:"身份等级: V1-V10"`
	IsAgent             int         `json:"isAgent"             orm:"is_agent"              description:"是否代理商: 0=否, 1=是"`
	AgentStatus         int         `json:"agentStatus"         orm:"agent_status"          description:"代理商状态: 0=未申请, 1=待审批, 2=已通过, 3=已拒绝"`
	AgentUnlockLevel    int         `json:"agentUnlockLevel"    orm:"agent_unlock_level"    description:"层级解锁: 0=仅一级佣金, 1=无限级佣金"`
	AgentLevel          int         `json:"agentLevel"          orm:"agent_level"           description:"代理商等级(已废弃)"`
	SubscribeRate       float64     `json:"subscribeRate"       orm:"subscribe_rate"        description:"订阅返佣比例(%)"`
	PowerRate           float64     `json:"powerRate"           orm:"power_rate"            description:"算力消耗佣金比例(%)"`
	InviteCode          string      `json:"inviteCode"          orm:"invite_code"           description:"邀请码"`
	InviteCodeExpire    *gtime.Time `json:"inviteCodeExpire"    orm:"invite_code_expire"    description:"邀请码过期时间"`
	InviterId           int64       `json:"inviterId"           orm:"inviter_id"            description:"邀请人ID"`
	InviteCount         int         `json:"inviteCount"         orm:"invite_count"          description:"直接邀请人数"`
	TeamCount           int         `json:"teamCount"           orm:"team_count"            description:"团队总人数"`
	TotalConsumePower   float64     `json:"totalConsumePower"   orm:"total_consume_power"   description:"总消耗算力"`
	TeamConsumePower    float64     `json:"teamConsumePower"    orm:"team_consume_power"    description:"团队消耗算力"`
	CurrentPlanId       int64       `json:"currentPlanId"       orm:"current_plan_id"       description:"当前订阅套餐ID"`
	PlanExpireTime      *gtime.Time `json:"planExpireTime"      orm:"plan_expire_time"      description:"套餐到期时间"`
	RobotLimit          int         `json:"robotLimit"          orm:"robot_limit"           description:"机器人数量限制"`
	ActiveRobotCount    int         `json:"activeRobotCount"    orm:"active_robot_count"    description:"运行中机器人数量"`
	PowerDiscount       float64     `json:"powerDiscount"       orm:"power_discount"        description:"算力消耗折扣(%)"`
	AgentApplyRemark    string      `json:"agentApplyRemark"    orm:"agent_apply_remark"    description:"代理商申请备注"`
	AgentApplyAt        *gtime.Time `json:"agentApplyAt"        orm:"agent_apply_at"        description:"代理商申请时间"`
	AgentApprovedAt     *gtime.Time `json:"agentApprovedAt"     orm:"agent_approved_at"     description:"代理商审批时间"`
	AgentApprovedBy     int64       `json:"agentApprovedBy"     orm:"agent_approved_by"     description:"审批人ID"`
	Status              int         `json:"status"              orm:"status"                description:"状态: 1=正常, 2=禁用"`
	CreatedAt           *gtime.Time `json:"createdAt"           orm:"created_at"            description:"创建时间"`
	UpdatedAt           *gtime.Time `json:"updatedAt"           orm:"updated_at"            description:"更新时间"`
}

