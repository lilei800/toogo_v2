package entity

import "github.com/gogf/gf/v2/os/gtime"

// SupportAgentPresence is the golang structure for table hg_support_agent_presence.
type SupportAgentPresence struct {
	AgentId    int64       `json:"agentId"    orm:"agent_id"     description:"客服ID"`
	Online     int         `json:"online"     orm:"online"       description:"是否在线：1在线 0离线"`
	LastSeenAt *gtime.Time `json:"lastSeenAt" orm:"last_seen_at" description:"最后心跳时间"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"   description:"更新时间"`
}


