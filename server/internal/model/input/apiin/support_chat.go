package apiin

import (
	"context"
	"hotgo/internal/model/input/form"
)

// SupportStartSessionInp 发起/获取我的客服会话
type SupportStartSessionInp struct{}

func (in *SupportStartSessionInp) Filter(ctx context.Context) (err error) { return }

// SupportSendInp 发送消息（用户端）
type SupportSendInp struct {
	SessionId int64  `json:"sessionId" v:"required#会话ID不能为空" dc:"会话ID"`
	Content   string `json:"content" v:"required#消息内容不能为空" dc:"消息内容"`
}

// SupportMessageListInp 消息列表（用户端）
type SupportMessageListInp struct {
	form.PageReq
	SessionId int64 `json:"sessionId" v:"required#会话ID不能为空" dc:"会话ID"`
}


