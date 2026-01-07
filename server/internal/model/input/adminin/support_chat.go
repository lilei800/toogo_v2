package adminin

import (
	"context"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// SupportAgentOnlineInp 客服上下线
type SupportAgentOnlineInp struct {
	Online bool `json:"online" dc:"是否在线"`
}

func (in *SupportAgentOnlineInp) Filter(ctx context.Context) (err error) {
	return
}

// SupportSessionListInp 会话列表
type SupportSessionListInp struct {
	form.PageReq
	Status int `json:"status" dc:"状态：1排队 2进行中 3已关闭(0表示全部)"`
}

// SupportSessionListModel 会话列表行
type SupportSessionListModel struct {
	entity.SupportSession
}

// SupportAcceptInp 接线
type SupportAcceptInp struct {
	SessionId int64 `json:"sessionId" v:"required#会话ID不能为空" dc:"会话ID"`
}

// SupportCloseInp 关闭会话
type SupportCloseInp struct {
	SessionId int64 `json:"sessionId" v:"required#会话ID不能为空" dc:"会话ID"`
}

// SupportSendInp 发送消息（客服端）
type SupportSendInp struct {
	SessionId int64  `json:"sessionId" v:"required#会话ID不能为空" dc:"会话ID"`
	Content   string `json:"content" v:"required#消息内容不能为空" dc:"消息内容"`
}

// SupportMessageListInp 消息列表
type SupportMessageListInp struct {
	form.PageReq
	SessionId int64 `json:"sessionId" v:"required#会话ID不能为空" dc:"会话ID"`
}

// SupportTransferInp 转接会话（客服端）
type SupportTransferInp struct {
	SessionId int64 `json:"sessionId" v:"required#会话ID不能为空" dc:"会话ID"`
	ToAgentId int64 `json:"toAgentId" v:"required#目标客服ID不能为空" dc:"目标客服ID"`
}

// SupportCannedListInp 常用语列表
type SupportCannedListInp struct {
	form.PageReq
}

type SupportCannedListModel struct {
	entity.SupportCannedReply
}

// SupportCannedEditInp 新增/编辑常用语
type SupportCannedEditInp struct {
	entity.SupportCannedReply
}

// SupportCannedDeleteInp 删除常用语
type SupportCannedDeleteInp struct {
	Id int64 `json:"id" v:"required#ID不能为空" dc:"ID"`
}


