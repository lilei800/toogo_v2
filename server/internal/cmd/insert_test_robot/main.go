// Package main 插入测试机器人
package main

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func main() {
	db := g.DB()

	// 获取策略组ID参数
	strategyGroupId := int64(18)
	if len(os.Args) > 1 {
		id, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err == nil {
			strategyGroupId = id
		}
	}

	// 获取用户ID（使用第一个用户，或从参数获取）
	userId := int64(1)
	if len(os.Args) > 2 {
		id, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err == nil {
			userId = id
		}
	}

	// 获取API配置ID（使用第一个API配置）
	var apiConfigId int64
	err := db.Model("hg_trading_api_config").
		Where("user_id", userId).
		WhereNull("deleted_at").
		OrderAsc("id").
		Limit(1).
		Fields("id").
		Scan(&apiConfigId)

	if err != nil || apiConfigId == 0 {
		fmt.Println("⚠️  警告：未找到用户的API配置，将使用默认值 1")
		apiConfigId = 1
	}

	fmt.Printf("========== 插入测试机器人 ==========\n")
	fmt.Printf("策略组ID: %d\n", strategyGroupId)
	fmt.Printf("用户ID: %d\n", userId)
	fmt.Printf("API配置ID: %d\n\n", apiConfigId)

	// 插入机器人数据
	data := gdb.Map{
		"user_id":                  userId,
		"robot_name":               fmt.Sprintf("测试机器人_%d", strategyGroupId),
		"api_config_id":            apiConfigId,
		"strategy_group_id":        strategyGroupId,
		"max_profit_target":        100.00,
		"max_loss_amount":          50.00,
		"max_runtime":              0,
		"risk_preference":          "balanced",
		"auto_risk_preference":     0,
		"market_state":             "volatile",
		"auto_market_state":        1,
		"exchange":                 "bitget",
		"symbol":                   "BTCUSDT",
		"order_type":               "market",
		"margin_mode":              "isolated",
		"leverage":                 10,
		"margin_percent":           10.00,
		"use_monitor_signal":       1,
		"enable_reverse_order":     1,
		"stop_loss_percent":        10.00,
		"profit_retreat_percent":  10.00,
		"auto_start_retreat_percent": 5.00,
		"status":                   1, // 未启动
		"auto_trade_enabled":       1,
		"auto_close_enabled":       1,
		"long_count":               0,
		"short_count":              0,
		"total_profit":             0.00,
		"runtime_seconds":          0,
		"remark":                   fmt.Sprintf("策略组ID: %d", strategyGroupId),
	}

	id, err := db.Model("hg_trading_robot").Data(data).InsertAndGetId()
	if err != nil {
		fmt.Printf("❌ 插入失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 成功插入机器人！\n")
	fmt.Printf("   机器人ID: %d\n", id)
	fmt.Printf("   机器人名称: %s\n", data["robot_name"])
	fmt.Printf("   策略组ID: %d\n", strategyGroupId)
	fmt.Printf("   状态: 未启动\n")
	fmt.Println("\n使用方法:")
	fmt.Println("  插入测试机器人: go run internal/cmd/insert_test_robot/main.go [策略组ID] [用户ID]")
	fmt.Println("  示例: go run internal/cmd/insert_test_robot/main.go 18 1")
}

