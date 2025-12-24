// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoConfig is the golang structure for table hg_toogo_config.
type ToogoConfig struct {
	Id          int64       `json:"id"          orm:"id"          description:"主键ID"`
	Group       string      `json:"group"       orm:"group"       description:"配置分组"`
	Key         string      `json:"key"         orm:"key"         description:"配置KEY"`
	Value       string      `json:"value"       orm:"value"       description:"配置值"`
	Type        string      `json:"type"        orm:"type"        description:"值类型: string/number/boolean/json"`
	Name        string      `json:"name"        orm:"name"        description:"配置名称"`
	Description string      `json:"description" orm:"description" description:"配置描述"`
	Sort        int         `json:"sort"        orm:"sort"        description:"排序"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"  description:"创建时间"`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"  description:"更新时间"`
}

