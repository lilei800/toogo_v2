// Package payment
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package payment

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"

	"github.com/gogf/gf/v2/frame/g"
)

// BalanceViewReq 查看余额请求
type BalanceViewReq struct {
	g.Meta `path:"/payment/balance/view" method:"get" tags:"USDT余额" summary:"查看余额"`
}

type BalanceViewRes struct {
	*entity.UsdtBalance
}

// BalanceLogListReq 资金流水列表请求
type BalanceLogListReq struct {
	g.Meta `path:"/payment/balance/logs" method:"get" tags:"USDT余额" summary:"资金流水列表"`
	input.BalanceLogListInp
}

type BalanceLogListRes struct {
	List       []*entity.UsdtBalanceLog `json:"list" dc:"列表"`
	TotalCount int                      `json:"totalCount" dc:"总数"`
	Page       int                      `json:"page" dc:"页码"`
	PageSize   int                      `json:"pageSize" dc:"每页数量"`
}



