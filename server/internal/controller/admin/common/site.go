// Package common
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package common

import (
	"context"
	"fmt"
	"hotgo/api/admin/common"
	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/library/captcha"
	"hotgo/internal/library/token"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/sysin"
	"hotgo/internal/service"
	"hotgo/utility/charset"
	"hotgo/utility/simple"
	"hotgo/utility/validate"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/i18n/gi18n"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gmode"
	"github.com/gogf/gf/v2/util/grand"
)

var Site = cSite{}

type cSite struct{}

// Ping ping
func (c *cSite) Ping(_ context.Context, _ *common.SitePingReq) (res *common.SitePingRes, err error) {
	return
}

// Config 获取配置
func (c *cSite) Config(ctx context.Context, _ *common.SiteConfigReq) (res *common.SiteConfigRes, err error) {
	request := ghttp.RequestFromCtx(ctx)
	res = &common.SiteConfigRes{
		Version: consts.VersionApp,
		WsAddr:  c.getWsAddr(ctx, request),
		Domain:  c.getDomain(ctx, request),
		Mode:    gmode.Mode(),
	}
	return
}

func (c *cSite) getWsAddr(ctx context.Context, request *ghttp.Request) string {
	// 如果是本地IP访问，则认为是调试模式，走实际请求地址，否则走配置中的地址
	// 尝试读取hostname，兼容本地运行模式
	ip := ghttp.RequestFromCtx(ctx).GetHeader("hostname")
	if len(ip) == 0 {
		ip = ghttp.RequestFromCtx(ctx).GetHost()
	}

	if validate.IsLocalIPAddr(ip) {
		return "ws://" + ip + ":" + gstr.StrEx(request.Host, ":") + g.Cfg().MustGet(ctx, "router.websocket.prefix").String()
	}

	basic, err := service.SysConfig().GetBasic(ctx)
	if err != nil || basic == nil {
		return ""
	}
	return basic.WsAddr
}

func (c *cSite) getDomain(ctx context.Context, request *ghttp.Request) string {
	// 如果是本地IP访问，则认为是调试模式，走实际请求地址，否则走配置中的地址
	// 尝试读取hostname，兼容本地运行模式
	ip := ghttp.RequestFromCtx(ctx).GetHeader("hostname")
	if len(ip) == 0 {
		ip = ghttp.RequestFromCtx(ctx).GetHost()
	}

	if validate.IsLocalIPAddr(ip) {
		return "http://" + ip + ":" + gstr.StrEx(request.Host, ":")
	}

	basic, err := service.SysConfig().GetBasic(ctx)
	if err != nil || basic == nil {
		return ""
	}
	return basic.Domain
}

// LoginConfig 登录配置
func (c *cSite) LoginConfig(ctx context.Context, _ *common.SiteLoginConfigReq) (res *common.SiteLoginConfigRes, err error) {
	res = new(common.SiteLoginConfigRes)
	login, err := service.SysConfig().GetLogin(ctx)
	if err != nil {
		return
	}

	res.LoginConfig = login
	res.I18nSwitch = g.Cfg().MustGet(ctx, "system.i18n.switch", true).Bool()
	res.DefaultLanguage = g.Cfg().MustGet(ctx, "system.i18n.defaultLanguage", consts.SysDefaultLanguage).String()
	res.ProjectName = gi18n.T(ctx, "HotGo管理系统")
	return
}

// Captcha 登录验证码
func (c *cSite) Captcha(ctx context.Context, _ *common.LoginCaptchaReq) (res *common.LoginCaptchaRes, err error) {
	loginConf, err := service.SysConfig().GetLogin(ctx)
	if err != nil {
		return
	}
	cid, base64 := captcha.Generate(ctx, loginConf.CaptchaType)
	res = &common.LoginCaptchaRes{Cid: cid, Base64: base64}
	return
}

// Register 账号注册
func (c *cSite) Register(ctx context.Context, req *common.RegisterReq) (res *common.RegisterRes, err error) {
	err = service.AdminSite().Register(ctx, &req.RegisterInp)
	return
}

// SendRegisterEmail 发送注册邮箱验证码（无需登录）
func (c *cSite) SendRegisterEmail(ctx context.Context, req *common.SendRegisterEmailReq) (res *common.SendRegisterEmailRes, err error) {
	// 发送注册验证码
	err = service.SysEmsLog().Send(ctx, &sysin.SendEmsInp{
		Event: consts.EmsTemplateRegister,
		Email: req.Email,
	})
	return
}

// SendResetPwdEmail 发送重置密码邮件（无需登录）
func (c *cSite) SendResetPwdEmail(ctx context.Context, req *common.SendResetPwdEmailReq) (res *common.SendResetPwdEmailRes, err error) {
	// 为防止“邮箱枚举”，无论邮箱是否存在，都返回成功；实际发送失败会打日志
	var mb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Email, req.Email).Scan(&mb); err != nil {
		return
	}
	if mb == nil {
		return
	}

	// 生成token并落库
	resetToken := string(charset.RandomCreateBytes(32))
	_, err = dao.AdminMember.Ctx(ctx).
		Where(dao.AdminMember.Columns().Id, mb.Id).
		Data(g.Map{
			dao.AdminMember.Columns().PasswordResetToken: resetToken,
		}).
		Update()
	if err != nil {
		return
	}

	// 域名优先走配置，否则走当前请求域名（兼容本地调试）
	basic, _ := service.SysConfig().GetBasic(ctx)
	domain := ""
	if basic != nil && basic.Domain != "" {
		domain = basic.Domain
	} else {
		r := ghttp.RequestFromCtx(ctx)
		if r != nil {
			domain = "http://" + r.GetHost()
		}
	}

	resetLink := fmt.Sprintf("%s/admin/passwordReset?token=%s", domain, resetToken)
	_ = service.SysEmsLog().Send(ctx, &sysin.SendEmsInp{
		Event: consts.EmsTemplateResetPwd,
		Email: req.Email,
		TplData: g.Map{
			"username":          mb.Username,
			"passwordResetLink": resetLink,
		},
	})
	return
}

// PasswordReset 重置密码（通过token）
func (c *cSite) PasswordReset(ctx context.Context, req *common.PasswordResetReq) (res *common.PasswordResetRes, err error) {
	if req.Token == "" {
		err = gerror.New("链接无效或已过期")
		return
	}

	// 解密并校验密码长度
	password, err := simple.DecryptText(req.Password)
	if err != nil {
		return
	}
	if err = g.Validator().Data(password).Rules("password").Messages("密码长度在6~18之间").Run(ctx); err != nil {
		return
	}

	// 按token查用户
	var mb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().PasswordResetToken, req.Token).Scan(&mb); err != nil {
		return
	}
	if mb == nil {
		err = gerror.New("链接无效或已过期")
		return
	}

	// 复用注册的加盐MD5策略
	salt := grand.S(6)
	passwordHash := gmd5.MustEncryptString(password + salt)
	_, err = dao.AdminMember.Ctx(ctx).
		Where(dao.AdminMember.Columns().Id, mb.Id).
		Data(g.Map{
			dao.AdminMember.Columns().Salt:               salt,
			dao.AdminMember.Columns().PasswordHash:       passwordHash,
			dao.AdminMember.Columns().PasswordResetToken: "",
		}).
		Update()
	return
}

// AccountLogin 账号登录
func (c *cSite) AccountLogin(ctx context.Context, req *common.AccountLoginReq) (res *common.AccountLoginRes, err error) {
	login, err := service.SysConfig().GetLogin(ctx)
	if err != nil {
		return
	}

	if !req.IsLock && login.CaptchaSwitch == consts.StatusEnabled {
		// 校验 验证码
		if !captcha.Verify(req.Cid, req.Code) {
			err = gerror.New("验证码错误")
			return
		}
	}

	model, err := service.AdminSite().AccountLogin(ctx, &req.AccountLoginInp)
	if err != nil {
		return
	}

	err = gconv.Scan(model, &res)
	return
}

// MobileLogin 手机号登录
func (c *cSite) MobileLogin(ctx context.Context, req *common.MobileLoginReq) (res *common.MobileLoginRes, err error) {
	model, err := service.AdminSite().MobileLogin(ctx, &req.MobileLoginInp)
	if err != nil {
		return
	}

	err = gconv.Scan(model, &res)
	return
}

// Logout 注销登录
func (c *cSite) Logout(ctx context.Context, _ *common.LoginLogoutReq) (res *common.LoginLogoutRes, err error) {
	err = token.Logout(ghttp.RequestFromCtx(ctx))
	return
}
