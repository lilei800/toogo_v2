package entity

import "github.com/gogf/gf/v2/os/gtime"

// SupportCannedReply is the golang structure for table hg_support_canned_reply.
type SupportCannedReply struct {
	Id        int64       `json:"id"        orm:"id"         description:"记录ID"`
	AgentId   int64       `json:"agentId"   orm:"agent_id"   description:"所属客服ID(0表示全局)"`
	Title     string      `json:"title"     orm:"title"      description:"标题"`
	Content   string      `json:"content"   orm:"content"    description:"内容"`
	Sort      int         `json:"sort"      orm:"sort"       description:"排序"`
	Status    int         `json:"status"    orm:"status"     description:"状态"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
}


