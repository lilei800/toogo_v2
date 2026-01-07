package support_chat

import (
	"context"
	"strings"

	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/library/contexts"
	"hotgo/internal/model"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/adminin"
	"hotgo/internal/service"
	"hotgo/internal/websocket"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type sSupportChat struct{}

func NewSupportChat() *sSupportChat { return &sSupportChat{} }

func init() {
	service.RegisterSupportChat(NewSupportChat())
}

func (s *sSupportChat) mustUser(ctx context.Context) (*model.Identity, error) {
	u := contexts.GetUser(ctx)
	if u == nil {
		return nil, gerror.New("获取用户信息失败！")
	}
	return u, nil
}

func (s *sSupportChat) AgentOnline(ctx context.Context, in *adminin.SupportAgentOnlineInp) (err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return err
	}
	now := gtime.Now()
	online := 0
	if in.Online {
		online = 1
	}
	_, err = dao.SupportAgentPresence.Ctx(ctx).
		Data(g.Map{
			dao.SupportAgentPresence.Columns().AgentId:    u.Id,
			dao.SupportAgentPresence.Columns().Online:     online,
			dao.SupportAgentPresence.Columns().LastSeenAt: now,
			dao.SupportAgentPresence.Columns().UpdatedAt:  now,
		}).
		Save()
	return
}

func (s *sSupportChat) SessionList(ctx context.Context, in *adminin.SupportSessionListInp) (list []*adminin.SupportSessionListModel, totalCount int, err error) {
	mod := dao.SupportSession.Ctx(ctx)
	if in.Status > 0 {
		mod = mod.Where(dao.SupportSession.Columns().Status, in.Status)
	}

	totalCount, err = mod.Count()
	if err != nil || totalCount == 0 {
		return
	}
	err = mod.Page(in.Page, in.PerPage).Order(dao.SupportSession.Columns().Id + " desc").Scan(&list)
	return
}

func (s *sSupportChat) Accept(ctx context.Context, in *adminin.SupportAcceptInp) (session *entity.SupportSession, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, err
	}
	now := gtime.Now()
	r, err := dao.SupportSession.Ctx(ctx).
		Where(dao.SupportSession.Columns().Id, in.SessionId).
		Where(dao.SupportSession.Columns().Status, consts.SupportSessionStatusWaiting).
		Where(dao.SupportSession.Columns().AgentId, 0).
		Data(g.Map{
			dao.SupportSession.Columns().AgentId:   u.Id,
			dao.SupportSession.Columns().Status:    consts.SupportSessionStatusActive,
			dao.SupportSession.Columns().UpdatedAt: now,
		}).
		Update()
	if err != nil {
		return nil, err
	}
	affected, _ := r.RowsAffected()
	if affected == 0 {
		return nil, gerror.New("会话已被接线或状态不允许")
	}

	session = new(entity.SupportSession)
	if err = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, in.SessionId).Scan(session); err != nil {
		return nil, err
	}

	s.pushSessionUpdated(ctx, session)
	return
}

func (s *sSupportChat) AcceptNext(ctx context.Context) (session *entity.SupportSession, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, err
	}
	now := gtime.Now()
	// 乐观并发：挑一个最早的排队会话，然后尝试抢占；失败则重试。
	for i := 0; i < 5; i++ {
		var picked entity.SupportSession
		if err = dao.SupportSession.Ctx(ctx).
			Where(dao.SupportSession.Columns().Status, consts.SupportSessionStatusWaiting).
			Where(dao.SupportSession.Columns().AgentId, 0).
			Order(dao.SupportSession.Columns().Id + " asc").
			Limit(1).
			Scan(&picked); err != nil {
			return nil, err
		}
		if picked.Id == 0 {
			return nil, gerror.New("暂无排队会话")
		}

		r, err2 := dao.SupportSession.Ctx(ctx).
			Where(dao.SupportSession.Columns().Id, picked.Id).
			Where(dao.SupportSession.Columns().Status, consts.SupportSessionStatusWaiting).
			Where(dao.SupportSession.Columns().AgentId, 0).
			Data(g.Map{
				dao.SupportSession.Columns().AgentId:   u.Id,
				dao.SupportSession.Columns().Status:    consts.SupportSessionStatusActive,
				dao.SupportSession.Columns().UpdatedAt: now,
			}).
			Update()
		if err2 != nil {
			return nil, err2
		}
		affected, _ := r.RowsAffected()
		if affected == 0 {
			continue
		}

		session = new(entity.SupportSession)
		if err = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, picked.Id).Scan(session); err != nil {
			return nil, err
		}
		s.pushSessionUpdated(ctx, session)
		return
	}
	return nil, gerror.New("接线失败，请重试")
}

func (s *sSupportChat) Close(ctx context.Context, in *adminin.SupportCloseInp) (err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return err
	}
	now := gtime.Now()
	r, err := dao.SupportSession.Ctx(ctx).
		Where(dao.SupportSession.Columns().Id, in.SessionId).
		Where(dao.SupportSession.Columns().AgentId, u.Id).
		WhereIn(dao.SupportSession.Columns().Status, []int{consts.SupportSessionStatusActive, consts.SupportSessionStatusWaiting}).
		Data(g.Map{
			dao.SupportSession.Columns().Status:    consts.SupportSessionStatusClosed,
			dao.SupportSession.Columns().ClosedAt:  now,
			dao.SupportSession.Columns().UpdatedAt: now,
		}).
		Update()
	if err != nil {
		return err
	}
	affected, _ := r.RowsAffected()
	if affected == 0 {
		return gerror.New("会话不存在或无权限")
	}
	var session entity.SupportSession
	_ = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, in.SessionId).Scan(&session)
	s.pushSessionUpdated(ctx, &session)
	return nil
}

func (s *sSupportChat) SendAgentMessage(ctx context.Context, in *adminin.SupportSendInp) (msg *entity.SupportMessage, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, err
	}
	return s.sendMessage(ctx, in.SessionId, consts.SupportSenderRoleAgent, u.Id, in.Content)
}

func (s *sSupportChat) MessageList(ctx context.Context, in *adminin.SupportMessageListInp) (list []*entity.SupportMessage, totalCount int, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 仅允许该客服查看自己接线的会话（排队会话不允许偷窥）
	var session entity.SupportSession
	if err = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, in.SessionId).Scan(&session); err != nil {
		return
	}
	if session.Id == 0 || (session.AgentId != 0 && session.AgentId != u.Id) {
		return nil, 0, gerror.New("会话不存在或无权限")
	}
	if session.Status == consts.SupportSessionStatusWaiting && session.AgentId == 0 {
		return nil, 0, gerror.New("会话未接线")
	}

	// 已读：客服查看消息列表时清空 unread_agent
	if session.AgentId == u.Id && session.UnreadAgent > 0 {
		now := gtime.Now()
		_, _ = dao.SupportSession.Ctx(ctx).
			Where(dao.SupportSession.Columns().Id, session.Id).
			Where(dao.SupportSession.Columns().AgentId, u.Id).
			Data(g.Map{
				dao.SupportSession.Columns().UnreadAgent: 0,
				dao.SupportSession.Columns().UpdatedAt:   now,
			}).
			Update()
		session.UnreadAgent = 0
		session.UpdatedAt = now
		s.pushSessionUpdated(ctx, &session)
	}

	mod := dao.SupportMessage.Ctx(ctx).Where(dao.SupportMessage.Columns().SessionId, in.SessionId)
	totalCount, err = mod.Count()
	if err != nil || totalCount == 0 {
		return
	}
	err = mod.Page(in.Page, in.PerPage).Order(dao.SupportMessage.Columns().Id + " desc").Scan(&list)
	return
}

func (s *sSupportChat) Transfer(ctx context.Context, in *adminin.SupportTransferInp) (session *entity.SupportSession, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, err
	}
	if in.ToAgentId <= 0 {
		return nil, gerror.New("目标客服ID不正确")
	}
	if in.ToAgentId == u.Id {
		return nil, gerror.New("目标客服不能是自己")
	}

	// 校验目标客服存在且启用
	cnt, err := dao.AdminMember.Ctx(ctx).
		Where(dao.AdminMember.Columns().Id, in.ToAgentId).
		Where(dao.AdminMember.Columns().Status, consts.StatusEnabled).
		Count()
	if err != nil {
		return nil, err
	}
	if cnt == 0 {
		return nil, gerror.New("目标客服不存在或已禁用")
	}

	// 读取并校验会话归属
	var old entity.SupportSession
	if err = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, in.SessionId).Scan(&old); err != nil {
		return nil, err
	}
	if old.Id == 0 {
		return nil, gerror.New("会话不存在")
	}
	if old.Status != consts.SupportSessionStatusActive {
		return nil, gerror.New("仅进行中的会话允许转接")
	}
	if old.AgentId != u.Id {
		return nil, gerror.New("无权限：仅当前接线客服可转接")
	}

	now := gtime.Now()
	_, err = dao.SupportSession.Ctx(ctx).
		Where(dao.SupportSession.Columns().Id, old.Id).
		Where(dao.SupportSession.Columns().AgentId, u.Id).
		Data(g.Map{
			dao.SupportSession.Columns().AgentId:   in.ToAgentId,
			dao.SupportSession.Columns().UpdatedAt: now,
		}).
		Update()
	if err != nil {
		return nil, err
	}

	// 组装更新后的会话对象
	oldAgentId := old.AgentId
	old.AgentId = in.ToAgentId
	old.UpdatedAt = now
	session = &old

	// 通知原客服：该会话已转接（仅最小字段）
	websocket.SendToUser(oldAgentId, &websocket.WResponse{
		Event: consts.SupportWsEventSessionUpdated,
		Data: g.Map{
			"id":     session.Id,
			"userId": session.UserId,
			"agentId": session.AgentId,
			"status": session.Status,
		},
	})
	// 通知新客服 + 用户 + 在线客服队列
	s.pushSessionUpdated(ctx, session)
	return
}

func (s *sSupportChat) CannedList(ctx context.Context, in *adminin.SupportCannedListInp) (list []*adminin.SupportCannedListModel, totalCount int, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, 0, err
	}
	mod := dao.SupportCannedReply.Ctx(ctx).
		Where(dao.SupportCannedReply.Columns().Status, consts.StatusEnabled).
		WhereIn(dao.SupportCannedReply.Columns().AgentId, []int64{0, u.Id})
	totalCount, err = mod.Count()
	if err != nil || totalCount == 0 {
		return
	}
	err = mod.Page(in.Page, in.PerPage).Order(dao.SupportCannedReply.Columns().Sort+" desc,"+dao.SupportCannedReply.Columns().Id+" desc").Scan(&list)
	return
}

func (s *sSupportChat) CannedEdit(ctx context.Context, in *adminin.SupportCannedEditInp) (err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return err
	}
	now := gtime.Now()
	if in.Id > 0 {
		_, err = dao.SupportCannedReply.Ctx(ctx).
			Where(dao.SupportCannedReply.Columns().Id, in.Id).
			WhereIn(dao.SupportCannedReply.Columns().AgentId, []int64{0, u.Id}).
			Data(g.Map{
				dao.SupportCannedReply.Columns().Title:     in.Title,
				dao.SupportCannedReply.Columns().Content:   in.Content,
				dao.SupportCannedReply.Columns().Sort:      in.Sort,
				dao.SupportCannedReply.Columns().Status:    in.Status,
				dao.SupportCannedReply.Columns().UpdatedAt: now,
			}).
			Update()
		return
	}
	in.AgentId = u.Id
	in.CreatedAt = now
	in.UpdatedAt = now
	_, err = dao.SupportCannedReply.Ctx(ctx).Data(in).OmitEmptyData().Insert()
	return
}

func (s *sSupportChat) CannedDelete(ctx context.Context, in *adminin.SupportCannedDeleteInp) (err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return err
	}
	_, err = dao.SupportCannedReply.Ctx(ctx).
		Where(dao.SupportCannedReply.Columns().Id, in.Id).
		Where(dao.SupportCannedReply.Columns().AgentId, u.Id).
		Delete()
	return
}

func (s *sSupportChat) StartUserSession(ctx context.Context) (session *entity.SupportSession, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, err
	}
	userId := u.Id
	// 若存在未关闭会话，则复用
	session = new(entity.SupportSession)
	if err = dao.SupportSession.Ctx(ctx).
		Where(dao.SupportSession.Columns().UserId, userId).
		WhereIn(dao.SupportSession.Columns().Status, []int{consts.SupportSessionStatusWaiting, consts.SupportSessionStatusActive}).
		Order(dao.SupportSession.Columns().Id + " desc").
		Limit(1).
		Scan(session); err != nil {
		return nil, err
	}
	if session.Id > 0 {
		return session, nil
	}

	now := gtime.Now()
	id, err := dao.SupportSession.Ctx(ctx).Data(g.Map{
		dao.SupportSession.Columns().UserId:     userId,
		dao.SupportSession.Columns().AgentId:    0,
		dao.SupportSession.Columns().Status:     consts.SupportSessionStatusWaiting,
		dao.SupportSession.Columns().CreatedAt:  now,
		dao.SupportSession.Columns().UpdatedAt:  now,
		dao.SupportSession.Columns().UnreadUser: 0,
		dao.SupportSession.Columns().UnreadAgent: 0,
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	session.Id = id
	session.UserId = userId
	session.Status = consts.SupportSessionStatusWaiting
	session.CreatedAt = now
	session.UpdatedAt = now

	// 通知在线客服：有新排队会话（不包含消息内容，避免泄露给未接线客服）
	websocket.SendToTag(consts.SupportWsTagAgents, &websocket.WResponse{
		Event: consts.SupportWsEventSessionUpdated,
		Data: g.Map{
			"id":     session.Id,
			"userId": session.UserId,
			"agentId": session.AgentId,
			"status": session.Status,
		},
	})
	return session, nil
}

func (s *sSupportChat) SendUserMessage(ctx context.Context, sessionId int64, content string) (msg *entity.SupportMessage, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, err
	}
	return s.sendMessage(ctx, sessionId, consts.SupportSenderRoleUser, u.Id, content)
}

func (s *sSupportChat) UserMessageList(ctx context.Context, sessionId int64, page, perPage int) (list []*entity.SupportMessage, totalCount int, err error) {
	u, err := s.mustUser(ctx)
	if err != nil {
		return nil, 0, err
	}
	var session entity.SupportSession
	if err = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, sessionId).Scan(&session); err != nil {
		return
	}
	if session.Id == 0 || session.UserId != u.Id {
		return nil, 0, gerror.New("会话不存在或无权限")
	}

	// 已读：用户查看消息列表时清空 unread_user
	if session.UnreadUser > 0 {
		now := gtime.Now()
		_, _ = dao.SupportSession.Ctx(ctx).
			Where(dao.SupportSession.Columns().Id, session.Id).
			Where(dao.SupportSession.Columns().UserId, u.Id).
			Data(g.Map{
				dao.SupportSession.Columns().UnreadUser: 0,
				dao.SupportSession.Columns().UpdatedAt:  now,
			}).
			Update()
		session.UnreadUser = 0
		session.UpdatedAt = now
		s.pushSessionUpdated(ctx, &session)
	}

	if perPage <= 0 {
		perPage = 20
	}
	mod := dao.SupportMessage.Ctx(ctx).Where(dao.SupportMessage.Columns().SessionId, sessionId)
	totalCount, err = mod.Count()
	if err != nil || totalCount == 0 {
		return
	}
	err = mod.Page(page, perPage).Order(dao.SupportMessage.Columns().Id + " desc").Scan(&list)
	return
}

func (s *sSupportChat) sendMessage(ctx context.Context, sessionId int64, senderRole int, senderId int64, content string) (msg *entity.SupportMessage, err error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, gerror.New("消息内容不能为空")
	}
	if len(content) > 5000 {
		return nil, gerror.New("消息内容过长")
	}

	var session entity.SupportSession
	if err = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, sessionId).Scan(&session); err != nil {
		return nil, err
	}
	if session.Id == 0 {
		return nil, gerror.New("会话不存在")
	}
	if session.Status == consts.SupportSessionStatusClosed {
		return nil, gerror.New("会话已关闭")
	}
	switch senderRole {
	case consts.SupportSenderRoleUser:
		if session.UserId != senderId {
			return nil, gerror.New("无权限")
		}
	case consts.SupportSenderRoleAgent:
		if session.AgentId != senderId {
			return nil, gerror.New("会话未分配给当前客服")
		}
	}

	now := gtime.Now()
	msg = &entity.SupportMessage{
		SessionId:  sessionId,
		SenderRole: senderRole,
		SenderId:   senderId,
		MsgType:    1,
		Content:    content,
		CreatedAt:  now,
	}
	id, err := dao.SupportMessage.Ctx(ctx).Data(msg).OmitEmptyData().InsertAndGetId()
	if err != nil {
		return nil, err
	}
	msg.Id = id

	// 写入“消息提醒中心”(admin_notice)：客服聊天
	// - 用户发给客服：提醒客服
	// - 客服发给用户：提醒用户
	createNotice := func(receiverId int64) {
		if receiverId <= 0 || receiverId == senderId {
			return
		}
		previewRunes := []rune(content)
		if len(previewRunes) > 80 {
			previewRunes = previewRunes[:80]
		}
		notice := &entity.AdminNotice{
			Title:     "客服聊天",
			Type:      consts.NoticeTypeCustomerService,
			Content:   string(previewRunes),
			Receiver:  gjson.New([]int64{receiverId}),
			Remark:    "",
			Sort:      0,
			Status:    consts.StatusEnabled,
			CreatedBy: senderId,
			CreatedAt: now,
		}
		if _, e := dao.AdminNotice.Ctx(ctx).Data(notice).OmitEmptyData().Insert(); e == nil {
			websocket.SendToUser(receiverId, &websocket.WResponse{
				Event: "notice",
				Data:  notice,
			})
		}
	}
	if senderRole == consts.SupportSenderRoleUser {
		if session.AgentId > 0 {
			createNotice(session.AgentId)
		}
	} else if senderRole == consts.SupportSenderRoleAgent {
		createNotice(session.UserId)
	}

	// 更新会话摘要与未读数
	_, _ = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, sessionId).Data(g.Map{
		dao.SupportSession.Columns().LastMsg:   content,
		dao.SupportSession.Columns().LastMsgAt: now,
		dao.SupportSession.Columns().UpdatedAt: now,
	}).Update()
	if senderRole == consts.SupportSenderRoleUser {
		_, _ = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, sessionId).Increment(dao.SupportSession.Columns().UnreadAgent, 1)
	} else {
		_, _ = dao.SupportSession.Ctx(ctx).Where(dao.SupportSession.Columns().Id, sessionId).Increment(dao.SupportSession.Columns().UnreadUser, 1)
	}

	// 推送给双方
	websocket.SendToUser(session.UserId, &websocket.WResponse{
		Event: consts.SupportWsEventMessage,
		Data:  msg,
	})
	if session.AgentId > 0 {
		websocket.SendToUser(session.AgentId, &websocket.WResponse{
			Event: consts.SupportWsEventMessage,
			Data:  msg,
		})
	} else if senderRole == consts.SupportSenderRoleUser {
		// 未接线：只通知“有待接会话更新”，不广播消息内容/预览
		websocket.SendToTag(consts.SupportWsTagAgents, &websocket.WResponse{
			Event: consts.SupportWsEventSessionUpdated,
			Data: g.Map{
				"id":     session.Id,
				"userId": session.UserId,
				"agentId": session.AgentId,
				"status": session.Status,
			},
		})
	}
	return msg, nil
}

func (s *sSupportChat) pushSessionUpdated(ctx context.Context, session *entity.SupportSession) {
	if session == nil || session.Id == 0 {
		return
	}
	websocket.SendToUser(session.UserId, &websocket.WResponse{
		Event: consts.SupportWsEventSessionUpdated,
		Data:  session,
	})
	if session.AgentId > 0 {
		websocket.SendToUser(session.AgentId, &websocket.WResponse{
			Event: consts.SupportWsEventSessionUpdated,
			Data:  session,
		})
	}
	// 对在线客服仅推“会话状态/分配”信息，避免把客户消息预览泄露给未接线客服
	websocket.SendToTag(consts.SupportWsTagAgents, &websocket.WResponse{
		Event: consts.SupportWsEventSessionUpdated,
		Data: g.Map{
			"id":     session.Id,
			"userId": session.UserId,
			"agentId": session.AgentId,
			"status": session.Status,
		},
	})
}


