//go:build tools
// +build tools

// 交易问题诊断工具
// 使用方法：go run diagnose_trading.go
package main

import (
	"context"
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

	fmt.Println("===========================================")
	fmt.Println("交易预警不下单问题诊断工具")
	fmt.Println("===========================================\n")

	// 诊断1: 检查 is_processed 字段
	checkIsProcessedField(ctx)

	// 诊断2: 检查最近的预警记录
	checkRecentAlerts(ctx)

	// 诊断3: 检查机器人配置
	checkRobotConfig(ctx)

	// 诊断4: 检查执行日志
	checkExecutionLogs(ctx)

	// 诊断5: 问题分析
	analyzeProblems(ctx)

	fmt.Println("\n===========================================")
	fmt.Println("✅ 诊断完成！")
	fmt.Println("详细报告：预警不下单问题诊断报告.md")
	fmt.Println("===========================================")
}

// 检查 is_processed 字段是否存在
func checkIsProcessedField(ctx context.Context) {
	fmt.Println("【诊断1】检查 is_processed 字段")
	fmt.Println("-------------------------------------------")

	var count int
	err := g.DB().GetScan(ctx, &count, `
		SELECT COUNT(*) 
		FROM information_schema.columns
		WHERE table_name = 'hg_trading_signal_log' 
		  AND column_name = 'is_processed'
	`)

	if err != nil {
		fmt.Printf("   ✗ 查询失败: %v\n", err)
		return
	}

	if count > 0 {
		// 获取字段详细信息
		var fieldInfo struct {
			DataType      string `json:"data_type"`
			IsNullable    string `json:"is_nullable"`
			ColumnDefault string `json:"column_default"`
		}
		err = g.DB().GetScan(ctx, &fieldInfo, `
			SELECT data_type, is_nullable, column_default
			FROM information_schema.columns
			WHERE table_name = 'hg_trading_signal_log' 
			  AND column_name = 'is_processed'
		`)
		if err == nil {
			fmt.Printf("   ✓ is_processed 字段存在\n")
			fmt.Printf("     - 数据类型: %s\n", fieldInfo.DataType)
			fmt.Printf("     - 允许NULL: %s\n", fieldInfo.IsNullable)
			fmt.Printf("     - 默认值: %s\n", fieldInfo.ColumnDefault)
		} else {
			fmt.Printf("   ✓ is_processed 字段存在（但无法获取详细信息）\n")
		}
	} else {
		fmt.Printf("   ✗ is_processed 字段不存在\n")
		fmt.Printf("   → 解决方案：执行 fix_is_processed_postgresql.sql\n")
	}
	fmt.Println()
}

// 检查最近的预警记录
func checkRecentAlerts(ctx context.Context) {
	fmt.Println("【诊断2】检查最近1小时的预警记录")
	fmt.Println("-------------------------------------------")

	var stats struct {
		Total        int `json:"total"`
		Unprocessed  int `json:"unprocessed"`
		Processed    int `json:"processed"`
		Executed     int `json:"executed"`
		NotExecuted  int `json:"not_executed"`
	}

	err := g.DB().GetScan(ctx, &stats, `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN is_processed = 0 THEN 1 ELSE 0 END) as unprocessed,
			SUM(CASE WHEN is_processed = 1 THEN 1 ELSE 0 END) as processed,
			SUM(CASE WHEN executed = 1 THEN 1 ELSE 0 END) as executed,
			COUNT(*) - SUM(CASE WHEN executed = 1 THEN 1 ELSE 0 END) as not_executed
		FROM hg_trading_signal_log
		WHERE created_at >= NOW() - INTERVAL '1 hour'
	`)

	if err != nil {
		fmt.Printf("   ✗ 查询失败: %v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("   总预警数: %d\n", stats.Total)
	fmt.Printf("   未处理: %d\n", stats.Unprocessed)
	fmt.Printf("   已处理: %d\n", stats.Processed)
	fmt.Printf("   已执行: %d\n", stats.Executed)
	fmt.Printf("   未执行: %d\n", stats.NotExecuted)

	// 显示最近5条记录
	if stats.Total > 0 {
		fmt.Println("\n   最近5条预警记录:")
		var records []map[string]interface{}
		err = g.DB().Ctx(ctx).Model("hg_trading_signal_log").
			Fields("id, robot_id, signal_type, current_price, is_processed, executed, created_at").
			Where("created_at >= NOW() - INTERVAL '1 hour'").
			OrderDesc("id").
			Limit(5).
			Scan(&records)

		if err == nil && len(records) > 0 {
			for i, rec := range records {
				processStatus := "未处理"
				if rec["is_processed"] == 1 || rec["is_processed"] == int64(1) {
					processStatus = "已读"
				}
				executeStatus := "未执行"
				if rec["executed"] == 1 || rec["executed"] == int64(1) {
					executeStatus = "已执行"
				}
				fmt.Printf("     %d. ID=%v | 机器人=%v | 方向=%v | 价格=%v | 状态=%s | 执行=%s\n",
					i+1, rec["id"], rec["robot_id"], rec["signal_type"],
					rec["current_price"], processStatus, executeStatus)
			}
		}
	}
	fmt.Println()
}

// 检查机器人配置
func checkRobotConfig(ctx context.Context) {
	fmt.Println("【诊断3】检查机器人配置")
	fmt.Println("-------------------------------------------")

	// 获取最近有预警的机器人ID
	var robotIds []int64
	err := g.DB().Ctx(ctx).Model("hg_trading_signal_log").
		Fields("DISTINCT robot_id").
		Where("created_at >= NOW() - INTERVAL '2 hour'").
		Scan(&robotIds)

	if err != nil {
		fmt.Printf("   ✗ 查询失败: %v\n", err)
		fmt.Println()
		return
	}

	if len(robotIds) == 0 {
		fmt.Println("   没有找到最近有预警的机器人")
		fmt.Println()
		return
	}

	fmt.Printf("   找到 %d 个机器人\n\n", len(robotIds))

	var robots []map[string]interface{}
	err = g.DB().Ctx(ctx).Model("hg_trading_robot").
		Fields("id, name, symbol, auto_trade_enabled, dual_side_position, status").
		WhereIn("id", robotIds).
		Limit(5).
		Scan(&robots)

	if err != nil {
		fmt.Printf("   ✗ 查询机器人配置失败: %v\n", err)
		fmt.Println()
		return
	}

	for _, robot := range robots {
		autoTrade := "✗ 未开启"
		if robot["auto_trade_enabled"] == 1 || robot["auto_trade_enabled"] == int64(1) {
			autoTrade = "✓ 已开启"
		}

		positionMode := "单向模式"
		if robot["dual_side_position"] == 1 || robot["dual_side_position"] == int64(1) {
			positionMode = "双向模式"
		}

		status := "已停止"
		if robot["status"] == 1 || robot["status"] == int64(1) {
			status = "运行中"
		}

		fmt.Printf("   机器人 #%v: %v (%v)\n", robot["id"], robot["name"], robot["symbol"])
		fmt.Printf("     - 自动交易: %s\n", autoTrade)
		fmt.Printf("     - 持仓模式: %s\n", positionMode)
		fmt.Printf("     - 状态: %s\n", status)
		fmt.Println()
	}
}

// 检查执行日志
func checkExecutionLogs(ctx context.Context) {
	fmt.Println("【诊断4】检查交易执行日志")
	fmt.Println("-------------------------------------------")

	// 检查表是否存在
	var tableExists int
	err := g.DB().GetScan(ctx, &tableExists, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_name = 'hg_trading_execution_log'
	`)

	if err != nil || tableExists == 0 {
		fmt.Println("   ✗ hg_trading_execution_log 表不存在")
		fmt.Println("   → 需要执行 fix_is_processed_postgresql.sql 创建表")
		fmt.Println()
		return
	}

	fmt.Println("   ✓ hg_trading_execution_log 表存在")

	// 查询最近的执行日志
	var logs []map[string]interface{}
	err = g.DB().Ctx(ctx).Model("hg_trading_execution_log").
		Fields("id, signal_log_id, robot_id, event_type, status, message, created_at").
		Where("created_at >= NOW() - INTERVAL '2 hour'").
		OrderDesc("id").
		Limit(5).
		Scan(&logs)

	if err != nil {
		fmt.Printf("   ✗ 查询执行日志失败: %v\n", err)
		fmt.Println()
		return
	}

	if len(logs) == 0 {
		fmt.Println("   最近2小时没有执行日志")
	} else {
		fmt.Printf("   找到 %d 条执行日志（最近5条）:\n\n", len(logs))
		for i, log := range logs {
			statusIcon := "✓"
			if log["status"] != "success" {
				statusIcon = "✗"
			}
			fmt.Printf("     %d. %s [%v] 机器人=%v | 预警ID=%v | 类型=%v\n",
				i+1, statusIcon, log["status"], log["robot_id"],
				log["signal_log_id"], log["event_type"])
			fmt.Printf("        消息: %v\n", log["message"])
		}
	}
	fmt.Println()
}

// 问题分析
func analyzeProblems(ctx context.Context) {
	fmt.Println("【诊断5】问题分析")
	fmt.Println("-------------------------------------------")

	hasIsProcessed := false
	unprocessedCount := 0
	processedNotExecuted := 0
	autoTradeDisabled := 0

	// 检查 is_processed 字段
	var count int
	err := g.DB().GetScan(ctx, &count, `
		SELECT COUNT(*) 
		FROM information_schema.columns
		WHERE table_name = 'hg_trading_signal_log' 
		  AND column_name = 'is_processed'
	`)
	if err == nil && count > 0 {
		hasIsProcessed = true
	}

	// 统计未处理的预警
	_ = g.DB().GetScan(ctx, &unprocessedCount, `
		SELECT COUNT(*) 
		FROM hg_trading_signal_log
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		  AND (is_processed = 0 OR is_processed IS NULL)
	`)

	// 统计已处理但未执行的预警
	_ = g.DB().GetScan(ctx, &processedNotExecuted, `
		SELECT COUNT(*) 
		FROM hg_trading_signal_log
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		  AND is_processed = 1
		  AND executed = 0
	`)

	// 统计自动交易未开启的机器人
	_ = g.DB().GetScan(ctx, &autoTradeDisabled, `
		SELECT COUNT(*) 
		FROM hg_trading_robot
		WHERE id IN (
			SELECT DISTINCT robot_id 
			FROM hg_trading_signal_log 
			WHERE created_at >= NOW() - INTERVAL '1 hour'
		)
		AND auto_trade_enabled != 1
	`)

	// 分析结果
	if !hasIsProcessed {
		fmt.Println("   ❌ 严重问题: is_processed 字段不存在")
		fmt.Println("      → 解决方案: 执行 fix_is_processed_postgresql.sql")
	} else {
		fmt.Println("   ✓ is_processed 字段存在")
	}

	if unprocessedCount > 0 {
		fmt.Printf("   ⚠ 发现 %d 条未处理的预警记录\n", unprocessedCount)
		fmt.Println("      → 可能原因: 保存预警记录失败（logId=0）")
		fmt.Println("      → 或: 重试机制尚未处理")
	} else {
		fmt.Println("   ✓ 没有未处理的预警记录")
	}

	if processedNotExecuted > 0 {
		fmt.Printf("   ⚠ 发现 %d 条已处理但未执行的预警\n", processedNotExecuted)
		fmt.Println("      → 可能原因1: 自动交易开关未开启")
		fmt.Println("      → 可能原因2: 已有持仓，被持仓检查阻止")
		fmt.Println("      → 可能原因3: 获取锁失败")
		fmt.Println("      → 建议: 查看 hg_trading_execution_log 表的详细日志")
	} else {
		fmt.Println("   ✓ 没有已处理但未执行的预警")
	}

	if autoTradeDisabled > 0 {
		fmt.Printf("   ⚠ 发现 %d 个机器人的自动交易开关未开启\n", autoTradeDisabled)
		fmt.Println("      → 解决方案: 在机器人管理页面开启自动交易")
	} else {
		fmt.Println("   ✓ 所有相关机器人的自动交易开关已开启")
	}

	fmt.Println()

	// 建议操作
	fmt.Println("【建议操作】")
	fmt.Println("-------------------------------------------")
	if !hasIsProcessed {
		fmt.Println("   1. 【必须】执行修复脚本:")
		fmt.Println("      psql -U your_user -d your_db -f storage/data/fix_is_processed_postgresql.sql")
	} else if processedNotExecuted > 0 {
		fmt.Println("   1. 查看执行日志，了解具体失败原因")
		fmt.Println("   2. 检查机器人配置:")
		fmt.Println("      - 自动交易开关是否开启")
		fmt.Println("      - 是否已有持仓（单向/双向模式）")
		fmt.Println("      - 持仓方向是否冲突")
	} else if unprocessedCount > 0 {
		fmt.Println("   1. 等待重试机制处理（每分钟自动重试）")
		fmt.Println("   2. 或者重启服务，触发立即重试")
	} else {
		fmt.Println("   ✓ 系统运行正常，没有发现问题")
		fmt.Println("     等待下一个交易信号，观察是否正常下单")
	}
}

