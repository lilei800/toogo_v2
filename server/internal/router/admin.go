// Package router
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package router

import (
	"context"
	"hotgo/internal/consts"
	"hotgo/internal/controller/admin"
	adminCtrl "hotgo/internal/controller/admin/admin"
	"hotgo/internal/controller/admin/common"
	"hotgo/internal/controller/admin/pay"
	"hotgo/internal/controller/admin/payment"
	"hotgo/internal/controller/admin/sys"
	"hotgo/internal/controller/admin/trading"
	"hotgo/internal/router/genrouter"
	"hotgo/internal/service"
	"hotgo/utility/simple"

	"github.com/gogf/gf/v2/net/ghttp"
)

func Admin(ctx context.Context, group *ghttp.RouterGroup) {
	// 兼容后台登录入口
	group.ALL("/login", func(r *ghttp.Request) {
		r.Response.RedirectTo("/admin")
	})

	group.Group(simple.RouterPrefix(ctx, consts.AppAdmin), func(group *ghttp.RouterGroup) {
		group.Bind(
			common.Site, // 基础
		)
		group.Middleware(service.Middleware().AdminAuth)
		group.Bind(
			common.Console,       // 控制台
			common.Ems,           // 邮件
			common.Sms,           // 短信
			common.Upload,        // 上传
			common.Wechat,        // 微信授权
			sys.Config,           // 配置
			sys.DictType,         // 字典类型
			sys.DictData,         // 字典数据
			sys.Attachment,       // 附件
			sys.Provinces,        // 省市区
			sys.Cron,             // 定时任务
			sys.CronGroup,        // 定时任务分组
			sys.Blacklist,        // 黑名单
			sys.Log,              // 访问日志
			sys.LoginLog,         // 登录日志
			sys.ServeLog,         // 服务日志
			sys.SmsLog,           // 短信记录
			sys.ServeLicense,     // 服务许可证
			adminCtrl.Member,     // 用户
			adminCtrl.Monitor,    // 监控
			adminCtrl.Role,       // 路由
			adminCtrl.Dept,       // 部门
			adminCtrl.Menu,       // 菜单
			adminCtrl.Notice,     // 公告
			adminCtrl.SupportChat, // 客服聊天
			adminCtrl.Post,       // 岗位
			adminCtrl.Order,      // 充值订单
			adminCtrl.CreditsLog, // 资金变动
			adminCtrl.Cash,       // 提现
			pay.Refund,           // 交易退款
			// Trading模块
			trading.ApiConfig,        // Trading API配置
			trading.ProxyConfig,      // Trading 代理配置
			trading.Robot,            // Trading 机器人
			trading.Order,            // Trading 订单
			trading.Monitor,          // Trading 监控
			trading.StrategyGroup,    // Trading 策略模板
			trading.StrategyTemplate, // Trading 策略
			trading.VolatilityConfig, // Trading 波动率配置
			trading.PublicMarket,     // Trading 公共行情（无需API Key）
			trading.AlertController,  // Trading 预警日志
			// Payment模块
			payment.Deposit,  // USDT充值
			payment.Withdraw, // USDT提现
			payment.Balance,  // 余额管理
			// Toogo模块
			admin.Toogo,       // Toogo量化交易
			admin.ToogoConfig, // Toogo系统配置
		)

		group.Middleware(service.Middleware().Develop)
		group.Bind(
			sys.GenCodes, // 生成代码
			sys.Addons,   // 插件管理
		)
	})

	// 注册生成路由
	genrouter.Register(ctx, group)
}
