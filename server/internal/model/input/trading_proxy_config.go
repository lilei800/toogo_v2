// Package input
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package input

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingProxyConfigGetInp 获取代理配置输入
type TradingProxyConfigGetInp struct {
}

// TradingProxyConfigModel 代理配置输出
type TradingProxyConfigModel struct {
	Id           int64       `json:"id" dc:"ID"`
	Enabled      int         `json:"enabled" dc:"是否启用"`
	ProxyType    string      `json:"proxyType" dc:"代理类型"`
	ProxyAddress string      `json:"proxyAddress" dc:"代理地址"`
	AuthEnabled  int         `json:"authEnabled" dc:"是否需要认证"`
	Username     string      `json:"username" dc:"用户名"`
	LastTestTime *gtime.Time `json:"lastTestTime" dc:"最后测试时间"`
	TestStatus   int         `json:"testStatus" dc:"测试状态"`
	TestMessage  string      `json:"testMessage" dc:"测试消息"`
	CreatedAt    *gtime.Time `json:"createdAt" dc:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt" dc:"更新时间"`
}

// TradingProxyConfigSaveInp 保存代理配置输入
type TradingProxyConfigSaveInp struct {
	Enabled      int    `json:"enabled" v:"in:0,1" dc:"是否启用：0=禁用,1=启用"`
	ProxyType    string `json:"proxyType" v:"in:socks5,http" dc:"代理类型"`
	ProxyAddress string `json:"proxyAddress" v:"required" dc:"代理地址"`
	AuthEnabled  int    `json:"authEnabled" v:"in:0,1" dc:"是否需要认证"`
	Username     string `json:"username" dc:"用户名"`
	Password     string `json:"password" dc:"密码"`
}

// TradingProxyConfigTestInp 测试代理输入
type TradingProxyConfigTestInp struct {
	ProxyType    string `json:"proxyType" v:"required|in:socks5,http" dc:"代理类型"`
	ProxyAddress string `json:"proxyAddress" v:"required" dc:"代理地址"`
	AuthEnabled  int    `json:"authEnabled" v:"in:0,1" dc:"是否需要认证"`
	Username     string `json:"username" dc:"用户名"`
	Password     string `json:"password" dc:"密码"`
}

// TradingProxyConfigTestModel 测试代理输出
type TradingProxyConfigTestModel struct {
	Success    bool   `json:"success" dc:"是否成功"`
	Message    string `json:"message" dc:"消息"`
	ExternalIP string `json:"externalIp" dc:"外网IP"`
	Latency    int    `json:"latency" dc:"延迟(ms)"`
}

// TradingProxyConfigToggleInp 切换启用状态输入
type TradingProxyConfigToggleInp struct {
	Enabled int `json:"enabled" v:"required|in:0,1" dc:"是否启用"`
}
