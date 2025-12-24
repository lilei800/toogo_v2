// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoAgentLevel is the golang structure for table hg_toogo_agent_level.
type ToogoAgentLevel struct {
	Id                   int64       `json:"id"                   orm:"id"                     description:"主键ID"`
	Level                int         `json:"level"                orm:"level"                  description:"等级: 1-5"`
	LevelName            string      `json:"levelName"            orm:"level_name"             description:"等级名称"`
	RequireTeamCount     int         `json:"requireTeamCount"     orm:"require_team_count"     description:"需要团队人数"`
	RequireTeamSubscribe float64     `json:"requireTeamSubscribe" orm:"require_team_subscribe" description:"需要团队订阅额(USDT)"`
	SubscribeRate1       float64     `json:"subscribeRate1"       orm:"subscribe_rate_1"       description:"订阅佣金比例(一级)"`
	SubscribeRate2       float64     `json:"subscribeRate2"       orm:"subscribe_rate_2"       description:"订阅佣金比例(二级)"`
	SubscribeRate3       float64     `json:"subscribeRate3"       orm:"subscribe_rate_3"       description:"订阅佣金比例(三级)"`
	PowerRate1           float64     `json:"powerRate1"           orm:"power_rate_1"           description:"算力消耗佣金比例(一级)"`
	PowerRate2           float64     `json:"powerRate2"           orm:"power_rate_2"           description:"算力消耗佣金比例(二级)"`
	PowerRate3           float64     `json:"powerRate3"           orm:"power_rate_3"           description:"算力消耗佣金比例(三级)"`
	Description          string      `json:"description"          orm:"description"            description:"等级描述"`
	Sort                 int         `json:"sort"                 orm:"sort"                   description:"排序"`
	Status               int         `json:"status"               orm:"status"                 description:"状态: 1=启用, 2=禁用"`
	CreatedAt            *gtime.Time `json:"createdAt"            orm:"created_at"             description:"创建时间"`
	UpdatedAt            *gtime.Time `json:"updatedAt"            orm:"updated_at"             description:"更新时间"`
}

