package service

import (
	"context"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/adminin"
)

type ISupportChat interface {
	// AgentOnline 客服上下线/心跳
	AgentOnline(ctx context.Context, in *adminin.SupportAgentOnlineInp) (err error)

	// SessionList 会话列表（客服端）
	SessionList(ctx context.Context, in *adminin.SupportSessionListInp) (list []*adminin.SupportSessionListModel, totalCount int, err error)
	// Accept 接线（指定会话）
	Accept(ctx context.Context, in *adminin.SupportAcceptInp) (session *entity.SupportSession, err error)
	// AcceptNext 接线（下一单）
	AcceptNext(ctx context.Context) (session *entity.SupportSession, err error)
	// Close 关闭会话
	Close(ctx context.Context, in *adminin.SupportCloseInp) (err error)
	// SendAgentMessage 客服发送消息
	SendAgentMessage(ctx context.Context, in *adminin.SupportSendInp) (msg *entity.SupportMessage, err error)
	// MessageList 消息列表（客服端）
	MessageList(ctx context.Context, in *adminin.SupportMessageListInp) (list []*entity.SupportMessage, totalCount int, err error)
	// Transfer 转接会话（客服端）
	Transfer(ctx context.Context, in *adminin.SupportTransferInp) (session *entity.SupportSession, err error)

	// CannedList 常用语列表
	CannedList(ctx context.Context, in *adminin.SupportCannedListInp) (list []*adminin.SupportCannedListModel, totalCount int, err error)
	// CannedEdit 常用语新增/编辑
	CannedEdit(ctx context.Context, in *adminin.SupportCannedEditInp) (err error)
	// CannedDelete 常用语删除
	CannedDelete(ctx context.Context, in *adminin.SupportCannedDeleteInp) (err error)

	// StartUserSession 用户端：发起/获取会话
	StartUserSession(ctx context.Context) (session *entity.SupportSession, err error)
	// SendUserMessage 用户端：发送消息
	SendUserMessage(ctx context.Context, sessionId int64, content string) (msg *entity.SupportMessage, err error)
	// UserMessageList 用户端：消息列表
	UserMessageList(ctx context.Context, sessionId int64, page, perPage int) (list []*entity.SupportMessage, totalCount int, err error)
}

var localSupportChat ISupportChat

func SupportChat() ISupportChat {
	if localSupportChat == nil {
		panic("implement not found for interface ISupportChat, forgot register?")
	}
	return localSupportChat
}

func RegisterSupportChat(i ISupportChat) {
	localSupportChat = i
}


