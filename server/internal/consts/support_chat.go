// Package consts
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package consts

const (
	// SupportSessionStatusWaiting 用户排队中（未分配客服）
	SupportSessionStatusWaiting = 1
	// SupportSessionStatusActive 会话进行中（已接线）
	SupportSessionStatusActive = 2
	// SupportSessionStatusClosed 会话已关闭
	SupportSessionStatusClosed = 3
)

const (
	// SupportSenderRoleUser 用户
	SupportSenderRoleUser = 1
	// SupportSenderRoleAgent 客服
	SupportSenderRoleAgent = 2
	// SupportSenderRoleSystem 系统
	SupportSenderRoleSystem = 3
)

const (
	// SupportWsTagAgents 客服在线标签（WS join 后用于广播排队/新会话）
	SupportWsTagAgents = "support_agents"
)

const (
	// SupportWsEventSessionUpdated 会话状态/分配更新
	SupportWsEventSessionUpdated = "support/session/updated"
	// SupportWsEventMessage 新消息
	SupportWsEventMessage = "support/message"
)


