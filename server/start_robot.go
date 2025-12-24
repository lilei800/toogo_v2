//go:build tools
// +build tools

// 启动机器人
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
	fmt.Println("启动机器人")
	fmt.Println("===========================================\n")

	// 查询机器人当前状态
	fmt.Println("【1】查询机器人状态...")
	robot, _ := g.DB().Ctx(ctx).Raw(`
		SELECT id, auto_trade_enabled, dual_side_position, status
		FROM hg_trading_robot
		WHERE id = 35
	`).One()

	if robot == nil {
		fmt.Println("   ✗ 机器人不存在")
		return
	}

	fmt.Printf("   当前状态:\n")
	fmt.Printf("     - 自动交易: %v\n", robot["auto_trade_enabled"])
	fmt.Printf("     - 持仓模式: %v (0=单向, 1=双向)\n", robot["dual_side_position"])
	fmt.Printf("     - 状态: %v (0=停止, 1=运行)\n", robot["status"])

	// 启动机器人
	if robot["status"].Int() == 0 {
		fmt.Println("\n【2】启动机器人...")
		result, err := g.DB().Exec(ctx, `
			UPDATE hg_trading_robot
			SET status = 1
			WHERE id = 35
		`)
		if err != nil {
			fmt.Printf("   ✗ 启动失败: %v\n", err)
			return
		}
		rows, _ := result.RowsAffected()
		if rows > 0 {
			fmt.Println("   ✓ 机器人已启动")
		}
	} else {
		fmt.Println("\n【2】机器人已经在运行中")
	}

	// 确保自动交易开关开启
	if robot["auto_trade_enabled"].Int() == 0 {
		fmt.Println("\n【3】开启自动交易...")
		_, err := g.DB().Exec(ctx, `
			UPDATE hg_trading_robot
			SET auto_trade_enabled = 1
			WHERE id = 35
		`)
		if err != nil {
			fmt.Printf("   ✗ 开启失败: %v\n", err)
		} else {
			fmt.Println("   ✓ 自动交易已开启")
		}
	} else {
		fmt.Println("\n【3】自动交易已经开启")
	}

	// 验证最终状态
	fmt.Println("\n【4】验证最终状态...")
	finalRobot, _ := g.DB().Ctx(ctx).Raw(`
		SELECT id, auto_trade_enabled, dual_side_position, status
		FROM hg_trading_robot
		WHERE id = 35
	`).One()

	if finalRobot != nil {
		autoTrade := "✗ 未开启"
		if finalRobot["auto_trade_enabled"].Int() == 1 {
			autoTrade = "✓ 已开启"
		}
		status := "✗ 停止"
		if finalRobot["status"].Int() == 1 {
			status = "✓ 运行中"
		}
		posMode := "单向"
		if finalRobot["dual_side_position"].Int() == 1 {
			posMode = "双向"
		}

		fmt.Printf("   机器人 #%v:\n", finalRobot["id"])
		fmt.Printf("     - 自动交易: %s\n", autoTrade)
		fmt.Printf("     - 持仓模式: %s\n", posMode)
		fmt.Printf("     - 状态: %s\n", status)

		if finalRobot["auto_trade_enabled"].Int() == 1 && finalRobot["status"].Int() == 1 {
			fmt.Println("\n===========================================")
			fmt.Println("✅ 机器人配置正确！")
			fmt.Println("===========================================")
			fmt.Println("\n【重要】需要重启应用服务才能生效！")
			fmt.Println("1. 停止当前运行的服务")
			fmt.Println("2. 重新启动: .\\main.exe http")
			fmt.Println("3. 等待下一个交易信号")
			fmt.Println("4. 观察日志中是否有下单记录")
		} else {
			fmt.Println("\n⚠️ 配置可能不正确，请检查")
		}
	}
}

