//go:build tools
// +build tools

// 验证修复结果
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
	_ = pgsql.Driver{}
	ctx := gctx.New()

	fmt.Println("===========================================")
	fmt.Println("验证修复结果")
	fmt.Println("===========================================\n")

	// 1. 检查 is_processed 字段
	fmt.Println("【1】is_processed 字段:")
	countVal, _ := g.DB().GetValue(ctx, `
		SELECT COUNT(*) 
		FROM information_schema.columns
		WHERE table_name = 'hg_trading_signal_log' 
		  AND column_name = 'is_processed'
	`)
	if countVal.Int() > 0 {
		fmt.Println("   ✓ 字段存在")
	} else {
		fmt.Println("   ✗ 字段不存在")
	}

	// 2. 检查执行日志表
	fmt.Println("\n【2】hg_trading_execution_log 表:")
	tableVal, _ := g.DB().GetValue(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_name = 'hg_trading_execution_log'
	`)
	if tableVal.Int() > 0 {
		fmt.Println("   ✓ 表存在")
	} else {
		fmt.Println("   ✗ 表不存在")
	}

	// 3. 查看最近的预警记录
	fmt.Println("\n【3】最近5条预警记录:")
	records, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			robot_id,
			signal_type,
			ROUND(current_price::NUMERIC, 2) as price,
			is_processed,
			executed,
			TO_CHAR(created_at, 'HH24:MI:SS') as time
		FROM hg_trading_signal_log
		ORDER BY id DESC
		LIMIT 5
	`).All()

	for i, rec := range records {
		fmt.Printf("   %d. ID=%v | 机器人=%v | 方向=%v | 价格=%v | is_processed=%v | executed=%v | 时间=%v\n",
			i+1,
			rec["id"],
			rec["robot_id"],
			rec["signal_type"],
			rec["price"],
			rec["is_processed"],
			rec["executed"],
			rec["time"])
	}

	// 4. 统计数据
	fmt.Println("\n【4】统计数据（最近1小时）:")
	stats, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN is_processed = 0 THEN 1 ELSE 0 END) as unprocessed,
			SUM(CASE WHEN is_processed = 1 THEN 1 ELSE 0 END) as processed,
			SUM(CASE WHEN executed = 1 THEN 1 ELSE 0 END) as executed
		FROM hg_trading_signal_log
		WHERE created_at >= NOW() - INTERVAL '1 hour'
	`).One()

	if stats != nil {
		fmt.Printf("   总数: %v\n", stats["total"])
		fmt.Printf("   未处理 (is_processed=0): %v\n", stats["unprocessed"])
		fmt.Printf("   已处理 (is_processed=1): %v\n", stats["processed"])
		fmt.Printf("   已执行 (executed=1): %v\n", stats["executed"])
	}

	// 5. 检查机器人配置
	fmt.Println("\n【5】机器人配置:")
	robots, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			auto_trade_enabled,
			dual_side_position,
			status
		FROM hg_trading_robot
		WHERE id IN (
			SELECT DISTINCT robot_id 
			FROM hg_trading_signal_log 
			WHERE created_at >= NOW() - INTERVAL '1 hour'
		)
		LIMIT 3
	`).All()

	for _, robot := range robots {
		autoTrade := "✗ 未开启"
		if robot["auto_trade_enabled"].Int() == 1 {
			autoTrade = "✓ 已开启"
		}
		posMode := "单向"
		if robot["dual_side_position"].Int() == 1 {
			posMode = "双向"
		}
		status := "停止"
		if robot["status"].Int() == 1 {
			status = "运行中"
		}
		fmt.Printf("   机器人 #%v: 自动交易=%s | 模式=%s | 状态=%s\n",
			robot["id"], autoTrade, posMode, status)
	}

	fmt.Println("\n===========================================")
	fmt.Println("✅ 验证完成！")
	fmt.Println("===========================================")
	fmt.Println("\n【结论】")
	if countVal.Int() > 0 && tableVal.Int() > 0 {
		fmt.Println("✓ 数据库修复成功！")
		fmt.Println("\n下一步:")
		fmt.Println("1. 重启应用服务（如果正在运行）")
		fmt.Println("2. 等待下一个交易信号")
		fmt.Println("3. 观察是否正常下单")
		fmt.Println("4. 查看执行日志: SELECT * FROM hg_trading_execution_log ORDER BY id DESC LIMIT 10;")
	} else {
		fmt.Println("✗ 修复未完成，请检查错误信息")
	}
}

