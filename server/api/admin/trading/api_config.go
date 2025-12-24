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

// ApiConfigListReq API配置列表请求
type ApiConfigListReq struct {
	g.Meta `path:"/trading/apiConfig/list" method:"get" tags:"交易管理" summary:"API配置列表" dc:"获取用户的API配置列表"`
	input.TradingApiConfigListInp
}

// ApiConfigListRes API配置列表响应
type ApiConfigListRes struct {
	List       []*input.TradingApiConfigListModel `json:"list" dc:"列表数据"`
	TotalCount int                                `json:"totalCount" dc:"总数"`
	Page       int                                `json:"page" dc:"当前页"`
	PageSize   int                                `json:"pageSize" dc:"每页数量"`
}

// ApiConfigCreateReq 创建API配置请求
type ApiConfigCreateReq struct {
	g.Meta `path:"/trading/apiConfig/create" method:"post" tags:"交易管理" summary:"创建API配置" dc:"创建新的交易所API配置"`
	input.TradingApiConfigCreateInp
}

// ApiConfigCreateRes 创建API配置响应
type ApiConfigCreateRes struct {
	Id int64 `json:"id" dc:"配置ID"`
}

// ApiConfigUpdateReq 更新API配置请求
type ApiConfigUpdateReq struct {
	g.Meta `path:"/trading/apiConfig/update" method:"post" tags:"交易管理" summary:"更新API配置" dc:"更新已有的API配置"`
	input.TradingApiConfigUpdateInp
}

// ApiConfigUpdateRes 更新API配置响应
type ApiConfigUpdateRes struct{}

// ApiConfigDeleteReq 删除API配置请求
type ApiConfigDeleteReq struct {
	g.Meta `path:"/trading/apiConfig/delete" method:"post" tags:"交易管理" summary:"删除API配置" dc:"删除指定的API配置"`
	input.TradingApiConfigDeleteInp
}

// ApiConfigDeleteRes 删除API配置响应
type ApiConfigDeleteRes struct{}

// ApiConfigViewReq 查看API配置请求
type ApiConfigViewReq struct {
	g.Meta `path:"/trading/apiConfig/view" method:"get" tags:"交易管理" summary:"查看API配置详情" dc:"查看单个API配置的详细信息"`
	input.TradingApiConfigViewInp
}

// ApiConfigViewRes 查看API配置响应
type ApiConfigViewRes struct {
	*input.TradingApiConfigViewModel
}

// ApiConfigTestReq 测试API连接请求
type ApiConfigTestReq struct {
	g.Meta `path:"/trading/apiConfig/test" method:"post" tags:"交易管理" summary:"测试API连接" dc:"测试API配置是否可用"`
	input.TradingApiConfigTestInp
}

// ApiConfigTestRes 测试API连接响应
type ApiConfigTestRes struct {
	Success bool   `json:"success" dc:"是否成功"`
	Message string `json:"message" dc:"消息"`
	Balance string `json:"balance" dc:"账户余额"`
	Latency int    `json:"latency" dc:"延迟(ms)"`
}

// ApiConfigSetDefaultReq 设为默认请求
type ApiConfigSetDefaultReq struct {
	g.Meta `path:"/trading/apiConfig/setDefault" method:"post" tags:"交易管理" summary:"设为默认配置" dc:"将指定配置设为默认"`
	input.TradingApiConfigSetDefaultInp
}

// ApiConfigSetDefaultRes 设为默认响应
type ApiConfigSetDefaultRes struct{}

// ApiConfigPlatformsReq 获取支持的平台列表请求
type ApiConfigPlatformsReq struct {
	g.Meta `path:"/trading/apiConfig/platforms" method:"get" tags:"交易管理" summary:"获取支持的平台" dc:"获取系统支持的交易所平台列表"`
}

// ApiConfigPlatformsRes 获取支持的平台列表响应
type ApiConfigPlatformsRes struct {
	List []*input.TradingApiConfigPlatformsModel `json:"list" dc:"平台列表"`
}

