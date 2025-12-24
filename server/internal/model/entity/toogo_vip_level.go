// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoVipLevel is the golang structure for table hg_toogo_vip_level.
type ToogoVipLevel struct {
	Id                  int64       `json:"id"                  orm:"id"                    description:"主键ID"`
	Level               int         `json:"level"               orm:"level"                 description:"等级: 1-10"`
	LevelName           string      `json:"levelName"           orm:"level_name"            description:"等级名称"`
	RequireInviteCount  int         `json:"requireInviteCount"  orm:"require_invite_count"  description:"需要邀请人数"`
	RequireConsumePower float64     `json:"requireConsumePower" orm:"require_consume_power" description:"需要消耗算力"`
	RequireTeamConsume  float64     `json:"requireTeamConsume"  orm:"require_team_consume"  description:"需要团队消耗算力"`
	PowerDiscount       float64     `json:"powerDiscount"       orm:"power_discount"        description:"算力折扣(5-30%)"`
	InviteRewardPower   float64     `json:"inviteRewardPower"   orm:"invite_reward_power"   description:"邀请奖励算力"`
	Description         string      `json:"description"         orm:"description"           description:"等级描述"`
	Icon                string      `json:"icon"                orm:"icon"                  description:"等级图标"`
	Sort                int         `json:"sort"                orm:"sort"                  description:"排序"`
	Status              int         `json:"status"              orm:"status"                description:"状态: 1=启用, 2=禁用"`
	CreatedAt           *gtime.Time `json:"createdAt"           orm:"created_at"            description:"创建时间"`
	UpdatedAt           *gtime.Time `json:"updatedAt"           orm:"updated_at"            description:"更新时间"`
}

