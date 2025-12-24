// Package common
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package common

import (
	"hotgo/internal/model"
	"hotgo/internal/model/input/adminin"

	"github.com/gogf/gf/v2/frame/g"
)

// LoginLogoutReq 注销登录
type LoginLogoutReq struct {
	g.Meta `path:"/site/logout" method:"post" tags:"后台基础" summary:"注销登录"`
}

type LoginLogoutRes struct{}

// RegisterReq 提交账号注册
type RegisterReq struct {
	g.Meta `path:"/site/register" method:"post" tags:"后台基础" summary:"账号注册"`
	adminin.RegisterInp
}

type RegisterRes struct {
	*adminin.LoginModel
}

// SendRegisterEmailReq 发送注册邮箱验证码（无需登录）
type SendRegisterEmailReq struct {
	g.Meta `path:"/site/sendRegisterEmail" tags:"后台基础" method:"post" summary:"发送注册邮箱验证码"`
	Email  string `json:"email" v:"required|email#邮箱不能为空|邮箱格式不正确" dc:"邮箱地址"`
}

type SendRegisterEmailRes struct{}

// SendResetPwdEmailReq 发送重置密码邮件（无需登录）
type SendResetPwdEmailReq struct {
	g.Meta `path:"/site/sendResetPwdEmail" tags:"后台基础" method:"post" summary:"发送重置密码邮件"`
	Email  string `json:"email" v:"required|email#邮箱不能为空|邮箱格式不正确" dc:"邮箱地址"`
}

type SendResetPwdEmailRes struct{}

// PasswordResetReq 重置密码（通过token）
type PasswordResetReq struct {
	g.Meta `path:"/site/passwordReset" tags:"后台基础" method:"post" summary:"重置密码(通过token)"`
	Token  string `json:"token" v:"required#token不能为空" dc:"重置token"`
	// 前端会对password做AES加密（与注册/登录一致），后端在Filter里解密校验
	Password string `json:"password" v:"required#密码不能为空" dc:"新密码"`
}

type PasswordResetRes struct{}

// LoginCaptchaReq 获取登录验证码
type LoginCaptchaReq struct {
	g.Meta `path:"/site/captcha" method:"get" tags:"后台基础" summary:"获取登录验证码"`
}

type LoginCaptchaRes struct {
	Cid    string `json:"cid" dc:"验证码ID"`
	Base64 string `json:"base64" dc:"验证码"`
}

// AccountLoginReq 提交账号登录
type AccountLoginReq struct {
	g.Meta `path:"/site/accountLogin" method:"post" tags:"后台基础" summary:"账号登录"`
	adminin.AccountLoginInp
}

type AccountLoginRes struct {
	*adminin.LoginModel
}

// MobileLoginReq 提交手机号登录
type MobileLoginReq struct {
	g.Meta `path:"/site/mobileLogin" method:"post" tags:"后台基础" summary:"手机号登录"`
	adminin.MobileLoginInp
}

type MobileLoginRes struct {
	*adminin.LoginModel
}

// SiteConfigReq 获取配置
type SiteConfigReq struct {
	g.Meta `path:"/site/config" method:"get" tags:"后台基础" summary:"获取配置"`
}

type SiteConfigRes struct {
	Version string `json:"version"        dc:"系统版本"`
	WsAddr  string `json:"wsAddr"         dc:"客户端websocket地址"`
	Domain  string `json:"domain"         dc:"对外域名"`
	Mode    string `json:"mode"           dc:"运行模式"`
}

// SiteLoginConfigReq 获取登录配置
type SiteLoginConfigReq struct {
	g.Meta `path:"/site/loginConfig" method:"get" tags:"后台基础" summary:"获取登录配置"`
}

type SiteLoginConfigRes struct {
	*model.LoginConfig
	I18nSwitch      bool   `json:"i18nSwitch" dc:"国际化开关"`
	DefaultLanguage string `json:"defaultLanguage" dc:"默认语言设置"`
	ProjectName     string `json:"projectName" dc:"项目名称"`
}

// SitePingReq ping
type SitePingReq struct {
	g.Meta `path:"/site/ping" method:"get" tags:"后台基础" summary:"ping"`
}

type SitePingRes struct{}
