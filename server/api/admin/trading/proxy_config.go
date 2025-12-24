// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/model/input"
)

// ProxyConfigGetReq 获取代理配置请求
type ProxyConfigGetReq struct {
	g.Meta `path:"/trading/proxyConfig/get" method:"get" tags:"交易管理" summary:"获取代理配置" dc:"获取当前用户的代理配置"`
	input.TradingProxyConfigGetInp
}

// ProxyConfigGetRes 获取代理配置响应
type ProxyConfigGetRes struct {
	*input.TradingProxyConfigModel
}

// ProxyConfigSaveReq 保存代理配置请求
type ProxyConfigSaveReq struct {
	g.Meta `path:"/trading/proxyConfig/save" method:"post" tags:"交易管理" summary:"保存代理配置" dc:"保存或更新代理配置"`
	input.TradingProxyConfigSaveInp
}

// ProxyConfigSaveRes 保存代理配置响应
type ProxyConfigSaveRes struct{}

// ProxyConfigTestReq 测试代理连接请求
type ProxyConfigTestReq struct {
	g.Meta `path:"/trading/proxyConfig/test" method:"post" tags:"交易管理" summary:"测试代理连接" dc:"测试代理配置是否可用"`
	input.TradingProxyConfigTestInp
}

// ProxyConfigTestRes 测试代理连接响应
type ProxyConfigTestRes struct {
	Success    bool   `json:"success" dc:"是否成功"`
	Message    string `json:"message" dc:"消息"`
	ExternalIP string `json:"externalIp" dc:"外网IP"`
	Latency    int    `json:"latency" dc:"延迟(ms)"`
}

// ProxyConfigToggleReq 切换启用状态请求
type ProxyConfigToggleReq struct {
	g.Meta `path:"/trading/proxyConfig/toggle" method:"post" tags:"交易管理" summary:"切换代理状态" dc:"启用或禁用代理"`
	input.TradingProxyConfigToggleInp
}

// ProxyConfigToggleRes 切换启用状态响应
type ProxyConfigToggleRes struct{}

