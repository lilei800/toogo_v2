package support_chat

import (
	"context"

	v1 "hotgo/api/api/support_chat/v1"
	"hotgo/internal/service"
)

func (c *ControllerV1) Start(ctx context.Context, _ *v1.StartReq) (res *v1.StartRes, err error) {
	session, err := service.SupportChat().StartUserSession(ctx)
	if err != nil {
		return nil, err
	}
	res = new(v1.StartRes)
	res.SupportSession = session
	return
}

func (c *ControllerV1) Send(ctx context.Context, req *v1.SendReq) (res *v1.SendRes, err error) {
	msg, err := service.SupportChat().SendUserMessage(ctx, req.SessionId, req.Content)
	if err != nil {
		return nil, err
	}
	res = new(v1.SendRes)
	res.SupportMessage = msg
	return
}

func (c *ControllerV1) MessageList(ctx context.Context, req *v1.MessageListReq) (res *v1.MessageListRes, err error) {
	list, total, err := service.SupportChat().UserMessageList(ctx, req.SessionId, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}
	res = new(v1.MessageListRes)
	res.List = list
	res.PageRes.Pack(req, total)
	return
}


