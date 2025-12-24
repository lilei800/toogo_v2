// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Description 市场状态规范化工具函数

package trading

import (
	"github.com/gogf/gf/v2/frame/g"
)

// normalizeMarketState 规范化市场状态格式（与 toogo 包中的函数保持一致）
// 统一格式: trend, volatile, high_vol, low_vol
// 兼容旧格式: range → volatile, high-volatility → high_vol, low-volatility → low_vol
func normalizeMarketState(marketState string) string {
	if marketState == "" {
		return "trend" // 默认值
	}

	switch marketState {
	case "range":
		return "volatile"
	case "high-volatility":
		return "high_vol"
	case "low-volatility":
		return "low_vol"
	case "trend", "volatile", "high_vol", "low_vol":
		return marketState
	default:
		// 未知格式，返回默认值
		g.Log().Warningf(nil, "[Trading] 未知市场状态格式: %s，使用默认值 trend", marketState)
		return "trend"
	}
}

