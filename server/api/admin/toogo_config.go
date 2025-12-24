// Package admin Toogo系统配置API
package admin

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoConfigListReq 配置列表请求
type ToogoConfigListReq struct {
	g.Meta `path:"/toogo/config/list" method:"get" tags:"Toogo系统配置" summary:"获取配置列表"`
	Group  string `json:"group" dc:"配置分组"`
}

type ToogoConfigListRes struct {
	List []*ToogoConfigItem `json:"list"`
}

// ToogoConfigItem 配置项
type ToogoConfigItem struct {
	Id          int64  `json:"id"`
	Group       string `json:"group"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

// ToogoConfigGroupsReq 获取配置分组
type ToogoConfigGroupsReq struct {
	g.Meta `path:"/toogo/config/groups" method:"get" tags:"Toogo系统配置" summary:"获取配置分组"`
}

type ToogoConfigGroupsRes struct {
	Groups []ConfigGroup `json:"groups"`
}

type ConfigGroup struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ToogoConfigUpdateReq 更新配置
type ToogoConfigUpdateReq struct {
	g.Meta `path:"/toogo/config/update" method:"post" tags:"Toogo系统配置" summary:"更新配置"`
	Items  []ConfigUpdateItem `json:"items" v:"required#配置项不能为空"`
}

type ConfigUpdateItem struct {
	Group string `json:"group" v:"required#分组不能为空"`
	Key   string `json:"key" v:"required#配置KEY不能为空"`
	Value string `json:"value"`
}

type ToogoConfigUpdateRes struct{}

// ToogoConfigGetReq 获取单个配置
type ToogoConfigGetReq struct {
	g.Meta `path:"/toogo/config/get" method:"get" tags:"Toogo系统配置" summary:"获取单个配置"`
	Group  string `json:"group" v:"required#分组不能为空"`
	Key    string `json:"key" v:"required#配置KEY不能为空"`
}

type ToogoConfigGetRes struct {
	*ToogoConfigItem
}

