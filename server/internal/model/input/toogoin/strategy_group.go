package toogoin

import (
	"hotgo/internal/model/entity"
)

// StrategyGroupListInp 策略模板列表输入
type StrategyGroupListInp struct {
	Page       int    `json:"page" dc:"页码"`
	PageSize   int    `json:"pageSize" dc:"每页数量"`
	Exchange   string `json:"exchange" dc:"交易平台"`
	Symbol     string `json:"symbol" dc:"交易对"`
	IsOfficial *int   `json:"isOfficial" dc:"是否官方 0-我的 1-官方 nil-全部"`
}

// StrategyGroupListModel 策略模板列表输出
type StrategyGroupListModel struct {
	List  []*StrategyGroupItem `json:"list" dc:"列表"`
	Page  int                  `json:"page" dc:"页码"`
	Total int                  `json:"total" dc:"总数"`
}

// StrategyGroupItem 策略模板项
type StrategyGroupItem struct {
	entity.TradingStrategyGroup
	StrategyCount int `json:"strategyCount" dc:"策略数量"`
}

// StrategyGroupCreateInp 创建策略模板输入
type StrategyGroupCreateInp struct {
	GroupName   string `json:"groupName" v:"required#请输入模板名称" dc:"模板名称"`
	GroupKey    string `json:"groupKey" v:"required#请输入模板标识" dc:"模板标识"`
	Exchange    string `json:"exchange" v:"required#请选择交易平台" dc:"交易平台"`
	Symbol      string `json:"symbol" v:"required#请选择交易对" dc:"交易对"`
	OrderType   string `json:"orderType" dc:"订单类型"`
	MarginMode  string `json:"marginMode" dc:"保证金模式"`
	Description string `json:"description" dc:"描述"`
	Sort        int    `json:"sort" dc:"排序"`
}

// StrategyGroupUpdateInp 更新策略模板输入
type StrategyGroupUpdateInp struct {
	Id          int64  `json:"id" v:"required#请指定模板ID" dc:"模板ID"`
	GroupName   string `json:"groupName" dc:"模板名称"`
	Exchange    string `json:"exchange" dc:"交易平台"`
	Symbol      string `json:"symbol" dc:"交易对"`
	OrderType   string `json:"orderType" dc:"订单类型"`
	MarginMode  string `json:"marginMode" dc:"保证金模式"`
	Description string `json:"description" dc:"描述"`
	Sort        int    `json:"sort" dc:"排序"`
	IsVisible   *int   `json:"isVisible" dc:"是否可见: 0=隐藏, 1=显示（仅官方策略模板管理使用）"`
	Confirmed   bool   `json:"confirmed" dc:"已确认修改（当策略组被机器人绑定时需要确认）"`
}

// StrategyGroupDeleteInp 删除策略模板输入
type StrategyGroupDeleteInp struct {
	Id int64 `json:"id" v:"required#请指定模板ID" dc:"模板ID"`
}

// StrategyGroupInitInp 初始化策略输入
type StrategyGroupInitInp struct {
	GroupId    int64 `json:"groupId" v:"required#请指定模板ID" dc:"模板ID"`
	UseDefault bool  `json:"useDefault" dc:"使用默认参数"`
}

