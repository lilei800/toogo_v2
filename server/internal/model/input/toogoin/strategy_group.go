package toogoin

import (
    "hotgo/internal/model/entity"
)

// StrategyGroupListInp strategy group list input
// isOfficial: 0=my, 1=official, nil=default(all official + my)
// nonPersonal: 1=only system/public groups (user_id=0 or NULL)
type StrategyGroupListInp struct {
    Page        int    `json:"page" dc:"page"`
    PageSize    int    `json:"pageSize" dc:"page size"`
    Exchange    string `json:"exchange" dc:"exchange"`
    Symbol      string `json:"symbol" dc:"symbol"`
    IsOfficial  *int   `json:"isOfficial" dc:"0=my,1=official,nil=default"`
    NonPersonal *int   `json:"nonPersonal" dc:"1=only non-personal groups (user_id=0 or NULL)"`
    IsActive    *int   `json:"isActive" dc:"0=disabled,1=enabled,nil=all"`
}

// StrategyGroupListModel list output
type StrategyGroupListModel struct {
    List  []*StrategyGroupItem `json:"list" dc:"list"`
    Page  int                  `json:"page" dc:"page"`
    Total int                  `json:"total" dc:"total"`
}

// StrategyGroupItem list item
type StrategyGroupItem struct {
    entity.TradingStrategyGroup
    StrategyCount int `json:"strategyCount" dc:"strategy count"`
}

// StrategyGroupCreateInp create input
type StrategyGroupCreateInp struct {
    GroupName   string `json:"groupName" v:"required#groupName required" dc:"group name"`
    GroupKey    string `json:"groupKey" v:"required#groupKey required" dc:"group key"`
    Exchange    string `json:"exchange" v:"required#exchange required" dc:"exchange"`
    Symbol      string `json:"symbol" v:"required#symbol required" dc:"symbol"`
    OrderType   string `json:"orderType" dc:"order type: market=市价, limit_then_market=先限价再市价(default)"`
    MarginMode  string `json:"marginMode" dc:"margin mode"`
    Description string `json:"description" dc:"description"`
    Sort        int    `json:"sort" dc:"sort"`

    // admin-only optional fields
    IsOfficial *int   `json:"isOfficial" dc:"0/1 (admin only)"`
    UserId     *int64 `json:"userId" dc:"owner user id (admin only)"`
    IsActive   *int   `json:"isActive" dc:"0/1 (admin only)"`
    IsVisible  *int   `json:"isVisible" dc:"0/1 (admin only)"`
}

// StrategyGroupUpdateInp update input
type StrategyGroupUpdateInp struct {
    Id          int64  `json:"id" v:"required#id required" dc:"id"`
    GroupName   string `json:"groupName" dc:"group name"`
    Exchange    string `json:"exchange" dc:"exchange"`
    Symbol      string `json:"symbol" dc:"symbol"`
    OrderType   string `json:"orderType" dc:"order type: market=市价, limit_then_market=先限价再市价(default)"`
    MarginMode  string `json:"marginMode" dc:"margin mode"`
    Description string `json:"description" dc:"description"`
    Sort        int    `json:"sort" dc:"sort"`

    IsVisible  *int `json:"isVisible" dc:"0/1 (admin only)"`
    IsOfficial *int `json:"isOfficial" dc:"0/1 (admin only)"`
    IsActive   *int `json:"isActive" dc:"0/1 (admin only)"`

    Confirmed bool `json:"confirmed" dc:"confirmed if group is bound to robots"`
}

// StrategyGroupDeleteInp delete input
type StrategyGroupDeleteInp struct {
    Id int64 `json:"id" v:"required#id required" dc:"id"`
}

// StrategyGroupInitInp init input
type StrategyGroupInitInp struct {
    GroupId    int64 `json:"groupId" v:"required#groupId required" dc:"groupId"`
    UseDefault bool  `json:"useDefault" dc:"use default params"`
}
