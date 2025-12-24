// Package trading
package trading

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 策略模板 API ====================

type StrategyGroupListReq struct {
	g.Meta     `path:"/strategy/group/list" method:"get" tags:"策略模板" summary:"策略模板列表"`
	Page       int    `json:"page" d:"1"`
	PageSize   int    `json:"pageSize" d:"10"`
	Exchange   string `json:"exchange"`
	Symbol     string `json:"symbol"`
	IsOfficial *int   `json:"isOfficial"` // 0-我的策略 1-官方策略 nil-全部
}

type StrategyGroupListRes struct {
	List  interface{} `json:"list"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
}

type StrategyGroupCreateReq struct {
	g.Meta      `path:"/strategy/group/create" method:"post" tags:"策略模板" summary:"创建策略模板"`
	GroupName   string `json:"groupName" v:"required#请输入模板名称"`
	GroupKey    string `json:"groupKey" v:"required#请输入模板标识"`
	Exchange    string `json:"exchange" v:"required#请选择交易平台"`
	Symbol      string `json:"symbol" v:"required#请选择交易对"`
	OrderType   string `json:"orderType"`
	MarginMode  string `json:"marginMode"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

type StrategyGroupCreateRes struct{}

type StrategyGroupUpdateReq struct {
	g.Meta      `path:"/strategy/group/update" method:"post" tags:"策略模板" summary:"更新策略模板"`
	Id          int64  `json:"id" v:"required#请指定模板ID"`
	GroupName   string `json:"groupName"`
	Exchange    string `json:"exchange"`
	Symbol      string `json:"symbol"`
	OrderType   string `json:"orderType"`
	MarginMode  string `json:"marginMode"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	IsVisible   *int   `json:"isVisible" dc:"是否可见: 0=隐藏, 1=显示（仅官方策略模板管理使用）"`
	Confirmed   bool   `json:"confirmed" dc:"已确认修改（当策略组被机器人绑定时需要确认）"`
}

type StrategyGroupUpdateRes struct{}

type StrategyGroupDeleteReq struct {
	g.Meta `path:"/strategy/group/delete" method:"post" tags:"策略模板" summary:"删除策略模板"`
	Id     int64 `json:"id" v:"required#请指定模板ID"`
}

type StrategyGroupDeleteRes struct{}

type StrategyGroupInitReq struct {
	g.Meta     `path:"/strategy/group/initStrategies" method:"post" tags:"策略模板" summary:"初始化策略"`
	GroupId    int64 `json:"groupId" v:"required#请指定模板ID"`
	UseDefault bool  `json:"useDefault"`
}

type StrategyGroupInitRes struct{}

type StrategyGroupCopyReq struct {
	g.Meta          `path:"/strategy/group/copyFromOfficial" method:"post" tags:"策略模板" summary:"从官方复制策略模板"`
	OfficialGroupId int64 `json:"officialGroupId" v:"required#请指定官方模板ID"`
}

type StrategyGroupCopyRes struct {
	Id int64 `json:"id" dc:"复制后的策略组ID"`
}

type StrategyGroupSetDefaultReq struct {
	g.Meta `path:"/strategy/group/setDefault" method:"post" tags:"策略模板" summary:"设置默认策略模板"`
	Id     int64 `json:"id" v:"required#请指定模板ID"`
}

type StrategyGroupSetDefaultRes struct{}

// ==================== 策略 API ====================

type StrategyTemplateListReq struct {
	g.Meta         `path:"/strategy/template/list" method:"get" tags:"策略" summary:"策略列表"`
	Page           int    `json:"page" d:"1"`
	PageSize       int    `json:"pageSize" d:"20"`
	GroupId        int64  `json:"groupId"`
	RiskPreference string `json:"riskPreference"`
	MarketState    string `json:"marketState"`
}

type StrategyTemplateListRes struct {
	List  interface{} `json:"list"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
}

type StrategyTemplateCreateReq struct {
	g.Meta                  `path:"/strategy/template/create" method:"post" tags:"策略" summary:"创建策略"`
	GroupId                 int64   `json:"groupId"`
	StrategyKey             string  `json:"strategyKey" v:"required#请输入策略标识"`
	StrategyName            string  `json:"strategyName" v:"required#请输入策略名称"`
	RiskPreference          string  `json:"riskPreference" v:"required#请选择风险偏好"`
	MarketState             string  `json:"marketState" v:"required#请选择市场状态"`
	MonitorWindow           int     `json:"monitorWindow" d:"300"`
	VolatilityThreshold     float64 `json:"volatilityThreshold" d:"10"`
	LeverageMin             int     `json:"leverageMin" v:"required"`
	LeverageMax             int     `json:"leverageMax" v:"required"`
	MarginPercentMin        float64 `json:"marginPercentMin" v:"required"`
	MarginPercentMax        float64 `json:"marginPercentMax" v:"required"`
	StopLossPercent         float64 `json:"stopLossPercent" v:"required"`
	ProfitRetreatPercent    float64 `json:"profitRetreatPercent" v:"required"`
	AutoStartRetreatPercent float64 `json:"autoStartRetreatPercent" v:"required"`
	ConfigJson              string  `json:"configJson"`
	Description             string  `json:"description"`
	IsActive                int     `json:"isActive" d:"1"`
	Sort                    int     `json:"sort" d:"100"`
}

type StrategyTemplateCreateRes struct{}

type StrategyTemplateUpdateReq struct {
	g.Meta                  `path:"/strategy/template/update" method:"post" tags:"策略" summary:"更新策略"`
	Id                      int64   `json:"id" v:"required#请指定策略ID"`
	StrategyName            string  `json:"strategyName"`
	RiskPreference          string  `json:"riskPreference"`
	MarketState             string  `json:"marketState"`
	MonitorWindow           int     `json:"monitorWindow"`
	VolatilityThreshold     float64 `json:"volatilityThreshold"`
	LeverageMin             int     `json:"leverageMin"`
	LeverageMax             int     `json:"leverageMax"`
	MarginPercentMin        float64 `json:"marginPercentMin"`
	MarginPercentMax        float64 `json:"marginPercentMax"`
	StopLossPercent         float64 `json:"stopLossPercent"`
	ProfitRetreatPercent    float64 `json:"profitRetreatPercent"`
	AutoStartRetreatPercent float64 `json:"autoStartRetreatPercent"`
	ReverseEnabled          bool    `json:"reverseEnabled"`
	ReverseLossRatio        int     `json:"reverseLossRatio"`
	ReverseProfitRatio      int     `json:"reverseProfitRatio"`
	ConfigJson              string  `json:"configJson"`
	Description             string  `json:"description"`
	IsActive                int     `json:"isActive"`
	Sort                    int     `json:"sort"`
	Confirmed               bool    `json:"confirmed" dc:"已确认修改（当策略组被机器人绑定时需要确认）"`
}

type StrategyTemplateUpdateRes struct{}

type StrategyTemplateDeleteReq struct {
	g.Meta `path:"/strategy/template/delete" method:"post" tags:"策略" summary:"删除策略"`
	Id     int64 `json:"id" v:"required#请指定策略ID"`
}

type StrategyTemplateDeleteRes struct{}

type StrategyTemplateApplyReq struct {
	g.Meta     `path:"/strategy/template/apply" method:"post" tags:"策略" summary:"应用策略到机器人"`
	StrategyId int64 `json:"strategyId" v:"required#请选择策略"`
	RobotId    int64 `json:"robotId" v:"required#请选择机器人"`
}

type StrategyTemplateApplyRes struct{}

