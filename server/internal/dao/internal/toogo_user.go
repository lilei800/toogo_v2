// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoUserDao is the data access object for table hg_toogo_user.
type ToogoUserDao struct {
	table   string           // table is the underlying table name of the DAO.
	group   string           // group is the database configuration group name of current DAO.
	columns ToogoUserColumns // columns contains all the column names of Table for convenient usage.
}

// ToogoUserColumns defines and stores column names for table hg_toogo_user.
type ToogoUserColumns struct {
	Id                string // 主键ID
	MemberId          string // 关联admin_member.id
	VipLevel          string // 身份等级: V1-V10
	IsAgent           string // 是否代理商: 0=否, 1=是
	AgentStatus       string // 代理商状态: 0=未申请, 1=待审批, 2=已通过, 3=已拒绝
	AgentUnlockLevel  string // 层级解锁: 0=仅一级佣金, 1=无限级佣金
	AgentLevel        string // 代理商等级(已废弃)
	SubscribeRate     string // 订阅返佣比例(%)
	PowerRate         string // 算力消耗佣金比例(%)
	InviteCode        string // 邀请码
	InviteCodeExpire  string // 邀请码过期时间
	InviterId         string // 邀请人ID
	InviteCount       string // 直接邀请人数
	TeamCount         string // 团队总人数
	TotalConsumePower string // 总消耗算力
	TeamConsumePower  string // 团队消耗算力
	CurrentPlanId     string // 当前订阅套餐ID
	PlanExpireTime    string // 套餐到期时间
	RobotLimit        string // 机器人数量限制
	ActiveRobotCount  string // 运行中机器人数量
	PowerDiscount     string // 算力消耗折扣(%)
	AgentApplyRemark  string // 代理商申请备注
	AgentApplyAt      string // 代理商申请时间
	AgentApprovedAt   string // 代理商审批时间
	AgentApprovedBy   string // 审批人ID
	Status            string // 状态: 1=正常, 2=禁用
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
}

// toogoUserColumns holds the columns for table hg_toogo_user.
var toogoUserColumns = ToogoUserColumns{
	Id:                "id",
	MemberId:          "member_id",
	VipLevel:          "vip_level",
	IsAgent:           "is_agent",
	AgentStatus:       "agent_status",
	AgentUnlockLevel:  "agent_unlock_level",
	AgentLevel:        "agent_level",
	SubscribeRate:     "subscribe_rate",
	PowerRate:         "power_rate",
	InviteCode:        "invite_code",
	InviteCodeExpire:  "invite_code_expire",
	InviterId:         "inviter_id",
	InviteCount:       "invite_count",
	TeamCount:         "team_count",
	TotalConsumePower: "total_consume_power",
	TeamConsumePower:  "team_consume_power",
	CurrentPlanId:     "current_plan_id",
	PlanExpireTime:    "plan_expire_time",
	RobotLimit:        "robot_limit",
	ActiveRobotCount:  "active_robot_count",
	PowerDiscount:     "power_discount",
	AgentApplyRemark:  "agent_apply_remark",
	AgentApplyAt:      "agent_apply_at",
	AgentApprovedAt:   "agent_approved_at",
	AgentApprovedBy:   "agent_approved_by",
	Status:            "status",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewToogoUserDao creates and returns a new DAO object for table data access.
func NewToogoUserDao() *ToogoUserDao {
	return &ToogoUserDao{
		group:   "default",
		table:   "hg_toogo_user",
		columns: toogoUserColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoUserDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoUserDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoUserDao) Columns() ToogoUserColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoUserDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoUserDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoUserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

