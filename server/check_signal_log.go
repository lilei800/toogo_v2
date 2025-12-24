//go:build tools
// +build tools

package main

import (
	"fmt"
	_ "hotgo/internal/packed"

	"github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	_ "github.com/lib/pq"
)

func main() {
	// 初始化 PostgreSQL 驱动
	_ = pgsql.Driver{}

	ctx := gctx.New()

	fmt.Println("=====================================")
	fmt.Println("检查 hg_trading_signal_log 表最新记录")
	fmt.Println("=====================================\n")

	// 查询最新的 5 条记录
	fmt.Println("1. 查询最新的 5 条记录...")
	var records []map[string]interface{}
	err := g.DB().Model("hg_trading_signal_log").
		Where("robot_id", 35).
		OrderDesc("id").
		Limit(5).
		Scan(&records)
	if err != nil {
		fmt.Printf("   ✗ 查询失败: %v\n", err)
		return
	}

	fmt.Printf("   ✓ 找到 %d 条记录\n\n", len(records))

	if len(records) == 0 {
		fmt.Println("   没有找到记录\n")
	} else {
		for i, record := range records {
			fmt.Printf("   记录 #%d:\n", i+1)
			fmt.Printf("     ID: %v\n", record["id"])
			fmt.Printf("     机器人ID: %v\n", record["robot_id"])
			fmt.Printf("     信号类型: %v\n", record["signal_type"])
			fmt.Printf("     当前价格: %v\n", record["current_price"])
			fmt.Printf("     信号强度: %v\n", record["signal_strength"])
			fmt.Printf("     创建时间: %v\n", record["created_at"])
			fmt.Println()
		}
	}

	// 尝试插入一条测试记录
	fmt.Println("2. 尝试插入一条测试记录...")
	logId, err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).Data(g.Map{
		"robot_id":         35,
		"strategy_id":      0,
		"symbol":           "BTCUSDT",
		"signal_type":      "LONG",
		"signal_source":    "test",
		"signal_strength":  0.5,
		"current_price":    98000.00,
		"window_min_price": 97000.00,
		"window_max_price": 99000.00,
		"threshold":        100.0,
		"market_state":     "neutral",
		"risk_preference":  "",
		"target_price":     0,
		"stop_loss":        0,
		"take_profit":      0,
		"executed":         0,
		"execute_result":   "",
		"is_processed":     0,
		"reason":           "测试记录",
		"indicators":       "{}",
	}).InsertAndGetId()

	if err != nil {
		fmt.Printf("   ✗ 插入失败: %v\n", err)
		return
	}

	fmt.Printf("   ✓ 插入成功，新记录ID: %d\n\n", logId)

	// 删除测试记录
	fmt.Println("3. 删除测试记录...")
	_, err = g.DB().Model("hg_trading_signal_log").Where("id", logId).Delete()
	if err != nil {
		fmt.Printf("   ✗ 删除失败: %v\n", err)
	} else {
		fmt.Println("   ✓ 删除成功\n")
	}

	fmt.Println("=====================================")
	fmt.Println("✅ 检查完成！")
	fmt.Println("=====================================")
}
