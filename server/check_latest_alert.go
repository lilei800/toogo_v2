//go:build tools
// +build tools

// 检查最新预警为什么没有下单
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
	fmt.Println("检查最新预警：15:11:55 做空 87459.90")
	fmt.Println("===========================================\n")

	// 1. 查找这条预警记录
	fmt.Println("【1】查找预警记录...")
	alert, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			robot_id,
			signal_type,
			current_price,
			is_processed,
			executed,
			reason,
			TO_CHAR(created_at, 'HH24:MI:SS') as time,
			created_at
		FROM hg_trading_signal_log
		WHERE signal_type = 'SHORT'
		  AND ROUND(current_price::NUMERIC, 2) = 87459.90
		  AND created_at >= NOW() - INTERVAL '1 hour'
		ORDER BY id DESC
		LIMIT 1
	`).One()

	if alert == nil {
		fmt.Println("   ✗ 未找到该预警记录")
		
		// 显示最近的做空预警
		fmt.Println("\n   最近的做空预警:")
		alerts, _ := g.DB().Ctx(ctx).Raw(`
			SELECT 
				id,
				ROUND(current_price::NUMERIC, 2) as price,
				is_processed,
				executed,
				TO_CHAR(created_at, 'HH24:MI:SS') as time
			FROM hg_trading_signal_log
			WHERE signal_type = 'SHORT'
			ORDER BY id DESC
			LIMIT 5
		`).All()
		for i, a := range alerts {
			fmt.Printf("     %d. ID=%v | 价格=%v | is_processed=%v | executed=%v | 时间=%v\n",
				i+1, a["id"], a["price"], a["is_processed"], a["executed"], a["time"])
		}
		return
	}

	fmt.Printf("   ✓ 找到预警记录\n")
	fmt.Printf("     - ID: %v\n", alert["id"])
	fmt.Printf("     - 机器人ID: %v\n", alert["robot_id"])
	fmt.Printf("     - 信号类型: %v\n", alert["signal_type"])
	fmt.Printf("     - 价格: %v\n", alert["current_price"])
	fmt.Printf("     - is_processed: %v (%s)\n", alert["is_processed"], 
		map[string]string{"0": "未处理", "1": "已处理"}[alert["is_processed"].String()])
	fmt.Printf("     - executed: %v (%s)\n", alert["executed"],
		map[string]string{"0": "未执行", "1": "已执行"}[alert["executed"].String()])
	fmt.Printf("     - 时间: %v\n", alert["time"])

	logId := alert["id"].Int64()
	robotId := alert["robot_id"].Int64()

	// 2. 查看执行日志
	fmt.Println("\n【2】查看执行日志...")
	logs, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			event_type,
			status,
			message,
			event_data,
			TO_CHAR(created_at, 'HH24:MI:SS') as time
		FROM hg_trading_execution_log
		WHERE signal_log_id = ?
		ORDER BY id ASC
	`, logId).All()

	if len(logs) == 0 {
		fmt.Println("   ✗ 没有找到执行日志")
		fmt.Println("   → 这说明 TryAutoTradeAndUpdate 函数可能没有被调用")
		fmt.Println("   → 或者在调用前就被拦截了")
	} else {
		fmt.Printf("   ✓ 找到 %d 条执行日志:\n", len(logs))
		for i, log := range logs {
			statusIcon := "✗"
			if log["status"].String() == "success" {
				statusIcon = "✓"
			}
			fmt.Printf("     %d. %s [%v] %v - %v | 时间=%v\n",
				i+1, statusIcon, log["status"], log["event_type"], log["message"], log["time"])
			if log["event_data"].String() != "" && log["event_data"].String() != "{}" {
				fmt.Printf("        详情: %v\n", log["event_data"])
			}
		}
	}

	// 3. 检查机器人状态
	fmt.Println("\n【3】检查机器人状态...")
	robot, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			auto_trade_enabled,
			dual_side_position,
			status
		FROM hg_trading_robot
		WHERE id = ?
	`, robotId).One()

	if robot == nil {
		fmt.Println("   ✗ 机器人不存在")
	} else {
		autoTrade := robot["auto_trade_enabled"].Int() == 1
		status := robot["status"].Int()
		dualSide := robot["dual_side_position"].Int() == 1
		
		fmt.Printf("   机器人 #%v:\n", robot["id"])
		fmt.Printf("     - 自动交易: %v (%s)\n", robot["auto_trade_enabled"], 
			map[bool]string{true: "✓ 已开启", false: "✗ 未开启"}[autoTrade])
		fmt.Printf("     - 持仓模式: %v (%s)\n", robot["dual_side_position"],
			map[bool]string{true: "双向", false: "单向"}[dualSide])
		fmt.Printf("     - 状态: %v\n", status)
		
		if !autoTrade {
			fmt.Println("\n   ⚠️ 问题：自动交易未开启！")
		}
	}

	// 4. 检查持仓情况
	fmt.Println("\n【4】检查持仓情况...")
	positions, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			position_side,
			position_amt,
			entry_price,
			status
		FROM hg_trading_position
		WHERE robot_id = ?
		  AND status = 1
		ORDER BY id DESC
	`, robotId).All()

	if len(positions) == 0 {
		fmt.Println("   ✓ 没有持仓")
	} else {
		fmt.Printf("   ⚠️ 发现 %d 个持仓:\n", len(positions))
		for i, pos := range positions {
			fmt.Printf("     %d. ID=%v | 方向=%v | 数量=%v | 入场价=%v | 状态=%v\n",
				i+1, pos["id"], pos["position_side"], pos["position_amt"], 
				pos["entry_price"], pos["status"])
		}
		
		// 分析持仓冲突
		hasSameDirection := false
		hasAnyPosition := true
		for _, pos := range positions {
			if pos["position_side"].String() == "SHORT" {
				hasSameDirection = true
			}
		}
		
		if robot != nil {
			dualSide := robot["dual_side_position"].Int() == 1
			if !dualSide && hasAnyPosition {
				fmt.Println("\n   ⚠️ 问题：单向模式下已有持仓，不能开新仓")
			} else if dualSide && hasSameDirection {
				fmt.Println("\n   ⚠️ 问题：双向模式下，做空方向已有持仓，禁止加仓")
			}
		}
	}

	// 5. 检查订单记录
	fmt.Println("\n【5】检查订单记录...")
	orders, _ := g.DB().Ctx(ctx).Raw(`
		SELECT 
			id,
			order_type,
			side,
			position_side,
			status,
			TO_CHAR(created_at, 'HH24:MI:SS') as time
		FROM hg_trading_order
		WHERE robot_id = ?
		  AND created_at >= NOW() - INTERVAL '1 hour'
		ORDER BY id DESC
		LIMIT 5
	`, robotId).All()

	if len(orders) == 0 {
		fmt.Println("   ✗ 最近1小时没有订单记录")
	} else {
		fmt.Printf("   找到 %d 条订单:\n", len(orders))
		for i, order := range orders {
			fmt.Printf("     %d. ID=%v | 类型=%v | 方向=%v | 持仓=%v | 状态=%v | 时间=%v\n",
				i+1, order["id"], order["order_type"], order["side"], 
				order["position_side"], order["status"], order["time"])
		}
	}

	// 6. 分析结论
	fmt.Println("\n===========================================")
	fmt.Println("【诊断结论】")
	fmt.Println("===========================================")

	if len(logs) == 0 {
		fmt.Println("\n❌ 主要问题：没有执行日志")
		fmt.Println("\n可能原因：")
		fmt.Println("1. saveSignalAlertSimple 返回的 logId 为 0（保存失败）")
		fmt.Println("2. 应用服务未重启，还在使用旧代码")
		fmt.Println("3. TryAutoTradeAndUpdate 函数未被调用")
		fmt.Println("4. 异步goroutine执行失败")
		
		fmt.Println("\n建议操作：")
		fmt.Println("1. 重启应用服务（如果还未重启）")
		fmt.Println("2. 查看应用日志：")
		fmt.Println("   Get-Content logs\\app.log -Tail 100 | Select-String -Pattern \"预警|下单|RobotTrader\"")
		fmt.Println("3. 检查是否有 panic 或 error")
	} else {
		// 分析执行日志
		hasFailLog := false
		failReason := ""
		for _, log := range logs {
			if log["status"].String() == "failed" {
				hasFailLog = true
				failReason = log["message"].String()
				break
			}
		}
		
		if hasFailLog {
			fmt.Printf("\n❌ 下单失败：%s\n", failReason)
			fmt.Println("\n根据失败原因采取相应措施")
		} else {
			fmt.Println("\n✓ 有执行日志但未下单，查看具体日志了解原因")
		}
	}

	fmt.Println("\n===========================================")
}

