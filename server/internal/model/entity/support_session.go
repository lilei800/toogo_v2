package entity

import "github.com/gogf/gf/v2/os/gtime"

// SupportSession is the golang structure for table hg_support_session.
type SupportSession struct {
	Id        int64       `json:"id"        orm:"id"         description:"会话ID"`
	UserId    int64       `json:"userId"    orm:"user_id"    description:"用户ID"`
	AgentId   int64       `json:"agentId"   orm:"agent_id"   description:"客服ID"`
	Status    int         `json:"status"    orm:"status"     description:"状态：1排队 2进行中 3已关闭"`
	Subject   string      `json:"subject"   orm:"subject"    description:"主题/摘要"`
	LastMsg   string      `json:"lastMsg"   orm:"last_msg"   description:"最后一条消息预览"`
	LastMsgAt *gtime.Time `json:"lastMsgAt" orm:"last_msg_at" description:"最后消息时间"`

	UnreadUser  int `json:"unreadUser"  orm:"unread_user"  description:"用户未读数"`
	UnreadAgent int `json:"unreadAgent" orm:"unread_agent" description:"客服未读数"`

	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`
	ClosedAt  *gtime.Time `json:"closedAt"  orm:"closed_at"  description:"关闭时间"`
}


