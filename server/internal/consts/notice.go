// Package consts
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package consts

// 公告类型
const (
	NoticeTypeNotify = 1 // 系统通知（可广播）
	NoticeTypeNotice = 2 // 公告（可广播）
	NoticeTypeLetter = 3 // 私信（定向）

	// 下面这些类型用于“消息提醒中心”（一般为定向消息，receiver 必填）
	NoticeTypeFinance         = 4 // 财务消息：转账/订阅/提现...
	NoticeTypeOrder           = 5 // 订单消息：下单/平仓...
	NoticeTypeCommission      = 6 // 佣金消息
	NoticeTypePromotion       = 7 // 推广消息
	NoticeTypeTicket          = 8 // 工单消息
	NoticeTypeCustomerService = 9 // 客服聊天
)
