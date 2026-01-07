package support_chat

import (
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/adminin"
	"hotgo/internal/model/input/form"
)

// AgentOnlineReq 客服上下线
type AgentOnlineReq struct {
	g.Meta `path:"/supportChat/agentOnline" method:"post" tags:"客服" summary:"客服上下线"`
	adminin.SupportAgentOnlineInp
}

type AgentOnlineRes struct{}

// SessionListReq 会话列表
type SessionListReq struct {
	g.Meta `path:"/supportChat/sessionList" method:"get" tags:"客服" summary:"会话列表"`
	adminin.SupportSessionListInp
}

type SessionListRes struct {
	List []*adminin.SupportSessionListModel `json:"list" dc:"数据列表"`
	form.PageRes
}

// AcceptReq 接线
type AcceptReq struct {
	g.Meta `path:"/supportChat/accept" method:"post" tags:"客服" summary:"接线(指定会话)"`
	adminin.SupportAcceptInp
}

type AcceptRes struct {
	*entity.SupportSession `json:"session" dc:"会话"`
}

// AcceptNextReq 接线(下一单)
type AcceptNextReq struct {
	g.Meta `path:"/supportChat/acceptNext" method:"post" tags:"客服" summary:"接线(下一单)"`
}

type AcceptNextRes struct {
	*entity.SupportSession `json:"session" dc:"会话"`
}

// CloseReq 关闭会话
type CloseReq struct {
	g.Meta `path:"/supportChat/close" method:"post" tags:"客服" summary:"关闭会话"`
	adminin.SupportCloseInp
}

type CloseRes struct{}

// SendReq 发送消息（客服）
type SendReq struct {
	g.Meta `path:"/supportChat/send" method:"post" tags:"客服" summary:"发送消息(客服)"`
	adminin.SupportSendInp
}

type SendRes struct {
	*entity.SupportMessage `json:"message" dc:"消息"`
}

// MessageListReq 消息列表
type MessageListReq struct {
	g.Meta `path:"/supportChat/messageList" method:"get" tags:"客服" summary:"消息列表"`
	adminin.SupportMessageListInp
}

type MessageListRes struct {
	List []*entity.SupportMessage `json:"list" dc:"数据列表"`
	form.PageRes
}

// TransferReq 转接会话
type TransferReq struct {
	g.Meta `path:"/supportChat/transfer" method:"post" tags:"客服" summary:"转接会话(指定客服)"`
	adminin.SupportTransferInp
}

type TransferRes struct {
	*entity.SupportSession `json:"session" dc:"会话"`
}

// CannedListReq 常用语列表
type CannedListReq struct {
	g.Meta `path:"/supportChat/canned/list" method:"get" tags:"客服" summary:"常用语列表"`
	adminin.SupportCannedListInp
}

type CannedListRes struct {
	List []*adminin.SupportCannedListModel `json:"list" dc:"数据列表"`
	form.PageRes
}

// CannedEditReq 常用语新增/编辑
type CannedEditReq struct {
	g.Meta `path:"/supportChat/canned/edit" method:"post" tags:"客服" summary:"常用语新增/编辑"`
	adminin.SupportCannedEditInp
}

type CannedEditRes struct{}

// CannedDeleteReq 常用语删除
type CannedDeleteReq struct {
	g.Meta `path:"/supportChat/canned/delete" method:"post" tags:"客服" summary:"常用语删除"`
	adminin.SupportCannedDeleteInp
}

type CannedDeleteRes struct{}


