package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/apiin"
	"hotgo/internal/model/input/form"
)

// StartReq 发起/获取我的客服会话
type StartReq struct {
	g.Meta `path:"/supportChat/start" method:"post" tags:"客服" summary:"发起/获取我的客服会话"`
	apiin.SupportStartSessionInp
}

type StartRes struct {
	*entity.SupportSession `json:"session" dc:"会话"`
}

// SendReq 发送消息（用户）
type SendReq struct {
	g.Meta `path:"/supportChat/send" method:"post" tags:"客服" summary:"发送消息(用户)"`
	apiin.SupportSendInp
}

type SendRes struct {
	*entity.SupportMessage `json:"message" dc:"消息"`
}

// MessageListReq 消息列表（用户）
type MessageListReq struct {
	g.Meta `path:"/supportChat/messageList" method:"get" tags:"客服" summary:"消息列表(用户)"`
	apiin.SupportMessageListInp
}

type MessageListRes struct {
	List []*entity.SupportMessage `json:"list" dc:"数据列表"`
	form.PageRes
}


