//go:build tools
// +build tools

// 检查特定预警记录
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

	logId := int64(13959)

	fmt.Println("===========================================")
	fmt.Printf("检查预警记录 ID=%d\n", logId)
	fmt.Println("===========================================\n")

	// 1. 查看预警记录详情
	fmt.Println("【1】预警记录详情...")
	alert, _ := g.DB().Ctx(ctx).Raw(`
		SELECT *
		FROM hg_trading_signal_log
		WHERE id = ?
	`, logId).One()

	if alert == nil {
		fmt.Println("   ✗ 记录不存在")
		return
	}

	fmt.Printf("   ID: %v\n", alert["id"])
	fmt.Printf("   机器人ID: %v\n", alert["robot_id"])
	fmt.Printf("   信号类型: %v\n", alert["signal_type"])
	fmt.Printf("   当前价格: %v\n", alert["current_price"])
	fmt.Printf("   窗口最低价: %v\n", alert["window_min_price"])
	fmt.Printf("   窗口最高价: %v\n", alert["window_max_price"])
	fmt.Printf("   阈值: %v\n", alert["threshold"])
	fmt.Printf("   is_processed: %v\n", alert["is_processed"])
	fmt.Printf("   executed: %v\n", alert["executed"])
	fmt.Printf("   创建时间: %v\n", alert["created_at"])
	fmt.Printf("   原因: %v\n", alert["reason"])

	robotId := alert["robot_id"].Int64()

	// 2. 查看执行日志
	fmt.Println("\n【2】执行日志...")
	logs, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			event_type,
			status,
			message,
			event_data,
			created_at
		FROM hg_trading_execution_log
		WHERE signal_log_id = ?
		ORDER BY id ASC
	`, logId).All()

	if len(logs) == 0 {
		fmt.Println("   ✗ 没有执行日志！")
		fmt.Println("\n   这说明 TryAutoTradeAndUpdate 没有被调用或执行失败")
	} else {
		fmt.Printf("   找到 %d 条执行日志:\n\n", len(logs))
		for i, log := range logs {
			fmt.Printf("   日志 #%d:\n", i+1)
			fmt.Printf("     - 事件类型: %v\n", log["event_type"])
			fmt.Printf("     - 状态: %v\n", log["status"])
			fmt.Printf("     - 消息: %v\n", log["message"])
			fmt.Printf("     - 时间: %v\n", log["created_at"])
			if log["event_data"].String() != "" && log["event_data"].String() != "{}" {
				fmt.Printf("     - 详细数据: %v\n", log["event_data"])
			}
			fmt.Println()
		}
	}

	// 3. 检查机器人配置
	fmt.Println("【3】机器人配置...")
	robot, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			auto_trade_enabled,
			dual_side_position,
			status
		FROM hg_trading_robot
		WHERE id = ?
	`, robotId).One()

	if robot != nil {
		fmt.Printf("   机器人 #%v:\n", robot["id"])
		fmt.Printf("     - 自动交易: %v\n", robot["auto_trade_enabled"])
		fmt.Printf("     - 持仓模式: %v (0=单向, 1=双向)\n", robot["dual_side_position"])
		fmt.Printf("     - 状态: %v\n", robot["status"])
	}

	// 4. 检查持仓
	fmt.Println("\n【4】当前持仓...")
	positions, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			position_side,
			position_amt,
			entry_price,
			status,
			created_at
		FROM hg_trading_position
		WHERE robot_id = ?
		  AND status = 1
		ORDER BY id DESC
	`, robotId).All()

	if len(positions) == 0 {
		fmt.Println("   ✓ 没有持仓")
	} else {
		fmt.Printf("   ⚠️ 有 %d 个持仓:\n", len(positions))
		for i, pos := range positions {
			fmt.Printf("     %d. ID=%v | 方向=%v | 数量=%v | 价格=%v | 时间=%v\n",
				i+1, pos["id"], pos["position_side"], pos["position_amt"], 
				pos["entry_price"], pos["created_at"])
		}
	}

	// 5. 总结分析
	fmt.Println("\n===========================================")
	fmt.Println("【分析】")
	fmt.Println("===========================================\n")

	if len(logs) == 0 {
		fmt.Println("❌ 没有执行日志，说明：")
		fmt.Println("   1. saveSignalAlertSimple 可能返回了 logId=0（保存失败）")
		fmt.Println("   2. 或者 TryAutoTradeAndUpdate 的 goroutine 没有执行")
		fmt.Println("   3. 或者应用还在使用旧代码（未重启）")
		
		fmt.Println("\n建议：")
		fmt.Println("   1. 检查应用日志，搜索 'logId=13959'")
		fmt.Println("   2. 确认应用服务是否已重启")
		fmt.Println("   3. 查看是否有 panic 或 error")
	} else {
		failedLogs := []interface{}{}
		for _, log := range logs {
			if log["status"].String() == "failed" {
				failedLogs = append(failedLogs, log)
			}
		}
		
		if len(failedLogs) > 0 {
			fmt.Println("❌ 下单失败，原因已记录在执行日志中")
		} else {
			fmt.Println("✓ 有执行日志，请查看具体内容")
		}
	}

	// 6. 查看最近的订单
	fmt.Println("\n【5】最近的订单...")
	orders, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			order_type,
			side,
			position_side,
			status,
			created_at
		FROM hg_trading_order
		WHERE robot_id = ?
		ORDER BY id DESC
		LIMIT 3
	`, robotId).All()

	if len(orders) == 0 {
		fmt.Println("   ✗ 没有订单记录")
	} else {
		fmt.Printf("   最近 %d 条订单:\n", len(orders))
		for i, order := range orders {
			fmt.Printf("     %d. ID=%v | %v %v | 状态=%v | 时间=%v\n",
				i+1, order["id"], order["side"], order["position_side"], 
				order["status"], order["created_at"])
		}
	}

	fmt.Println("\n===========================================")
}

