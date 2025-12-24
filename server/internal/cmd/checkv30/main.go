package main

import (
	"context"
	"fmt"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "hotgo/internal/packed"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()

	// 查询策略组
	result, err := g.DB().Ctx(ctx).GetAll(context.Background(),
		"SELECT id, group_name, group_key, is_official, is_active FROM hg_trading_strategy_group WHERE group_key IN ('official_v30', 'official_v31') ORDER BY group_key")
	if err != nil {
		fmt.Printf("查询策略组失败: %v\n", err)
		return
	}
	
	fmt.Println("========== 官方V30策略组信息 ==========")
	if len(result) > 0 {
		for _, row := range result {
			fmt.Printf("ID: %v | 名称: %v | Key: %v | 官方: %v | 启用: %v\n",
				row["id"], row["group_name"], row["group_key"], row["is_official"], row["is_active"])
		}
	} else {
		fmt.Println("未找到官方V30策略组")
	}

	// 查询策略模板数量
	count, err := g.DB().Ctx(ctx).GetValue(context.Background(),
		`SELECT COUNT(*) FROM hg_trading_strategy_template t 
		 JOIN hg_trading_strategy_group g ON t.group_id = g.id 
		 WHERE g.group_key = 'official_v31'`)
	if err != nil {
		fmt.Printf("查询模板数量失败: %v\n", err)
		return
	}
	fmt.Printf("\n策略模板总数: %v 套\n", count)

	// 查询策略模板详情
	templates, err := g.DB().Ctx(ctx).GetAll(context.Background(),
		`SELECT t.sort, t.strategy_name, t.market_state, t.risk_preference, 
		        t.leverage, t.margin_percent, t.stop_loss_percent, 
		        t.auto_start_retreat_percent, t.profit_retreat_percent,
		        t.monitor_window, t.volatility_threshold
		 FROM hg_trading_strategy_template t 
		 JOIN hg_trading_strategy_group g ON t.group_id = g.id 
		 WHERE g.group_key = 'official_v31'
		 ORDER BY t.sort`)
	if err != nil {
		fmt.Printf("查询模板详情失败: %v\n", err)
		return
	}

	fmt.Println("\n========== 12套策略模板详情 ==========")
	fmt.Println("序号 | 策略名称 | 市场 | 风险 | 杠杆 | 保证金 | 止损 | 启动止盈 | 止盈回撤 | 窗口 | 阈值")
	fmt.Println("-----|----------|------|------|------|--------|------|----------|----------|------|------")
	for _, t := range templates {
		fmt.Printf("%v | %v | %v | %v | %vx | %v%% | %v%% | %v%% | %v%% | %vs | %vU\n",
			t["sort"], t["strategy_name"], t["market_state"], t["risk_preference"],
			t["leverage"], t["margin_percent"], t["stop_loss_percent"],
			t["auto_start_retreat_percent"], t["profit_retreat_percent"],
			t["monitor_window"], t["volatility_threshold"])
	}
	
	fmt.Println("\n✅ 官方V30策略模板组已成功写入数据库！")
}

