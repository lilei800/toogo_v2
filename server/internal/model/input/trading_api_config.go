// Package input
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
// @AutoGenerate Version 2.13.1

package input

import (
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/os/gtime"
)

// TradingApiConfigListInp 列表查询输入
type TradingApiConfigListInp struct {
	Page     int    `json:"page" v:"required|min:1" dc:"页码"`
	PageSize int    `json:"pageSize" v:"required|min:1|max:100" dc:"每页数量"`
	Platform string `json:"platform" dc:"平台筛选：bitget/binance/okx"`
	Status   int    `json:"status" dc:"状态筛选：1=正常,2=禁用"`
	ApiName  string `json:"apiName" dc:"API名称模糊搜索"`
}

// TradingApiConfigListModel 列表输出
type TradingApiConfigListModel struct {
	Id             int64       `json:"id" dc:"ID"`
	ApiName        string      `json:"apiName" dc:"API名称"`
	Platform       string      `json:"platform" dc:"平台"`
	BaseUrl        string      `json:"baseUrl" dc:"API地址"`
	ApiKey         string      `json:"apiKey" dc:"API Key（脱敏）"`
	IsDefault      int         `json:"isDefault" dc:"是否默认"`
	Status         int         `json:"status" dc:"状态"`
	LastVerifyTime *gtime.Time `json:"lastVerifyTime" dc:"最后验证时间"`
	VerifyStatus   int         `json:"verifyStatus" dc:"验证状态"`
	VerifyMessage  string      `json:"verifyMessage" dc:"验证消息"`
	Remark         string      `json:"remark" dc:"备注"`
	CreatedAt      *gtime.Time `json:"createdAt" dc:"创建时间"`
	UpdatedAt      *gtime.Time `json:"updatedAt" dc:"更新时间"`
}

// TradingApiConfigCreateInp 创建输入
type TradingApiConfigCreateInp struct {
	ApiName    string `json:"apiName" v:"required|length:1,100" dc:"API名称"`
	Platform   string `json:"platform" v:"required|in:bitget,binance,okx,gate" dc:"平台"`
	BaseUrl    string `json:"baseUrl" dc:"API地址（可选，自动填充）"`
	ApiKey     string `json:"apiKey" v:"required" dc:"API Key"`
	SecretKey  string `json:"secretKey" v:"required" dc:"Secret Key"`
	Passphrase string `json:"passphrase" dc:"Passphrase（OKX/Bitget必填）"`
	IsDefault  int    `json:"isDefault" dc:"是否默认"`
	Remark     string `json:"remark" dc:"备注"`
}

// TradingApiConfigUpdateInp 更新输入
type TradingApiConfigUpdateInp struct {
	Id         int64  `json:"id" v:"required" dc:"ID"`
	ApiName    string `json:"apiName" v:"required|length:1,100" dc:"API名称"`
	Platform   string `json:"platform" v:"required|in:bitget,binance,okx,gate" dc:"平台"`
	BaseUrl    string `json:"baseUrl" dc:"API地址（可选，自动填充）"`
	ApiKey     string `json:"apiKey" dc:"API Key（不修改则不传）"`
	SecretKey  string `json:"secretKey" dc:"Secret Key（不修改则不传）"`
	Passphrase string `json:"passphrase" dc:"Passphrase"`
	IsDefault  int    `json:"isDefault" dc:"是否默认"`
	Status     int    `json:"status" dc:"状态"`
	Remark     string `json:"remark" dc:"备注"`
}

// TradingApiConfigDeleteInp 删除输入
type TradingApiConfigDeleteInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingApiConfigViewInp 查看详情输入
type TradingApiConfigViewInp struct {
	Id int64 `json:"id" v:"required" dc:"ID"`
}

// TradingApiConfigViewModel 详情输出
type TradingApiConfigViewModel struct {
	entity.TradingApiConfig
}

// TradingApiConfigTestInp 测试连接输入
type TradingApiConfigTestInp struct {
	Id int64 `json:"id" v:"required" dc:"配置ID"`
}

// TradingApiConfigTestModel 测试连接输出
type TradingApiConfigTestModel struct {
	Success bool   `json:"success" dc:"是否成功"`
	Message string `json:"message" dc:"消息"`
	Balance string `json:"balance" dc:"账户余额"`
	Latency int    `json:"latency" dc:"延迟(ms)"`
}

// TradingApiConfigSetDefaultInp 设为默认输入
type TradingApiConfigSetDefaultInp struct {
	Id int64 `json:"id" v:"required" dc:"配置ID"`
}

// TradingApiConfigPlatformsModel 支持的平台列表
type TradingApiConfigPlatformsModel struct {
	Value    string `json:"value" dc:"平台值"`
	Label    string `json:"label" dc:"平台名称"`
	BaseUrl  string `json:"baseUrl" dc:"默认API地址"`
	NeedPass bool   `json:"needPass" dc:"是否需要Passphrase"`
}

