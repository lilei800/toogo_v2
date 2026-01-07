package admin

import (
	"context"

	api "hotgo/api/admin/support_chat"
	"hotgo/internal/service"
)

var (
	SupportChat = cSupportChat{}
)

type cSupportChat struct{}

func (c *cSupportChat) AgentOnline(ctx context.Context, req *api.AgentOnlineReq) (res *api.AgentOnlineRes, err error) {
	err = service.SupportChat().AgentOnline(ctx, &req.SupportAgentOnlineInp)
	return
}

func (c *cSupportChat) SessionList(ctx context.Context, req *api.SessionListReq) (res *api.SessionListRes, err error) {
	list, total, err := service.SupportChat().SessionList(ctx, &req.SupportSessionListInp)
	if err != nil {
		return nil, err
	}
	res = new(api.SessionListRes)
	res.List = list
	res.PageRes.Pack(req, total)
	return
}

func (c *cSupportChat) Accept(ctx context.Context, req *api.AcceptReq) (res *api.AcceptRes, err error) {
	session, err := service.SupportChat().Accept(ctx, &req.SupportAcceptInp)
	if err != nil {
		return nil, err
	}
	res = new(api.AcceptRes)
	res.SupportSession = session
	return
}

func (c *cSupportChat) AcceptNext(ctx context.Context, _ *api.AcceptNextReq) (res *api.AcceptNextRes, err error) {
	session, err := service.SupportChat().AcceptNext(ctx)
	if err != nil {
		return nil, err
	}
	res = new(api.AcceptNextRes)
	res.SupportSession = session
	return
}

func (c *cSupportChat) Close(ctx context.Context, req *api.CloseReq) (res *api.CloseRes, err error) {
	err = service.SupportChat().Close(ctx, &req.SupportCloseInp)
	return
}

func (c *cSupportChat) Send(ctx context.Context, req *api.SendReq) (res *api.SendRes, err error) {
	msg, err := service.SupportChat().SendAgentMessage(ctx, &req.SupportSendInp)
	if err != nil {
		return nil, err
	}
	res = new(api.SendRes)
	res.SupportMessage = msg
	return
}

func (c *cSupportChat) MessageList(ctx context.Context, req *api.MessageListReq) (res *api.MessageListRes, err error) {
	list, total, err := service.SupportChat().MessageList(ctx, &req.SupportMessageListInp)
	if err != nil {
		return nil, err
	}
	res = new(api.MessageListRes)
	res.List = list
	res.PageRes.Pack(req, total)
	return
}

func (c *cSupportChat) Transfer(ctx context.Context, req *api.TransferReq) (res *api.TransferRes, err error) {
	session, err := service.SupportChat().Transfer(ctx, &req.SupportTransferInp)
	if err != nil {
		return nil, err
	}
	res = new(api.TransferRes)
	res.SupportSession = session
	return
}

func (c *cSupportChat) CannedList(ctx context.Context, req *api.CannedListReq) (res *api.CannedListRes, err error) {
	list, total, err := service.SupportChat().CannedList(ctx, &req.SupportCannedListInp)
	if err != nil {
		return nil, err
	}
	res = new(api.CannedListRes)
	res.List = list
	res.PageRes.Pack(req, total)
	return
}

func (c *cSupportChat) CannedEdit(ctx context.Context, req *api.CannedEditReq) (res *api.CannedEditRes, err error) {
	err = service.SupportChat().CannedEdit(ctx, &req.SupportCannedEditInp)
	return
}

func (c *cSupportChat) CannedDelete(ctx context.Context, req *api.CannedDeleteReq) (res *api.CannedDeleteRes, err error) {
	err = service.SupportChat().CannedDelete(ctx, &req.SupportCannedDeleteInp)
	return
}


