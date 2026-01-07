package entity

import "github.com/gogf/gf/v2/os/gtime"

// SupportMessage is the golang structure for table hg_support_message.
type SupportMessage struct {
	Id        int64       `json:"id"        orm:"id"         description:"消息ID"`
	SessionId int64       `json:"sessionId" orm:"session_id" description:"会话ID"`
	SenderRole int        `json:"senderRole" orm:"sender_role" description:"发送方角色：1用户 2客服 3系统"`
	SenderId  int64       `json:"senderId"  orm:"sender_id"  description:"发送方ID"`
	MsgType   int         `json:"msgType"   orm:"msg_type"   description:"消息类型：1文本"`
	Content   string      `json:"content"   orm:"content"    description:"内容"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
}


