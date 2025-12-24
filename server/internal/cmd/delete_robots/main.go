// Package main 删除机器人工具
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()
	db := g.DB()

	// 查询所有机器人
	var robots []gdb.Record
	err := db.Model("hg_trading_robot").
		Fields("id, robot_name, status, strategy_group_id, created_at, deleted_at").
		OrderDesc("id").
		Scan(&robots)

	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		os.Exit(1)
	}

	if len(robots) == 0 {
		fmt.Println("数据库中没有机器人")
		return
	}

	// 显示机器人列表
	fmt.Println("========== 机器人列表 ==========")
	fmt.Printf("%-8s %-30s %-10s %-15s %-20s %-10s\n", "ID", "机器人名称", "状态", "策略组ID", "创建时间", "已删除")
	fmt.Println(strings.Repeat("-", 100))

	for _, r := range robots {
		status := getStatusText(r["status"])
		groupId := "NULL"
		if r["strategy_group_id"] != nil {
			groupId = fmt.Sprintf("%v", r["strategy_group_id"])
		}
		deleted := "否"
		if r["deleted_at"] != nil {
			deleted = "是"
		}
		createdAt := ""
		if r["created_at"] != nil {
			createdAt = fmt.Sprintf("%v", r["created_at"])[:19]
		}

		fmt.Printf("%-8s %-30s %-10s %-15s %-20s %-10s\n",
			fmt.Sprintf("%v", r["id"]),
			fmt.Sprintf("%v", r["robot_name"]),
			status,
			groupId,
			createdAt,
			deleted)
	}

	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("总计: %d 个机器人\n\n", len(robots))

	// 获取命令行参数
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  删除单个机器人: go run internal/cmd/delete_robots/main.go <机器人ID>")
		fmt.Println("  删除多个机器人: go run internal/cmd/delete_robots/main.go <ID1> <ID2> <ID3>...")
		fmt.Println("  删除所有机器人: go run internal/cmd/delete_robots/main.go all")
		fmt.Println("  软删除（推荐）: go run internal/cmd/delete_robots/main.go soft <ID>")
		fmt.Println("  硬删除: go run internal/cmd/delete_robots/main.go hard <ID>")
		return
	}

	// 解析参数
	arg := os.Args[1]
	var ids []int64

	if arg == "all" {
		// 删除所有机器人
		fmt.Println("⚠️  警告：将删除所有机器人！")
		fmt.Print("确认删除所有机器人？(yes/no): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("已取消")
			return
		}
		for _, r := range robots {
			if r["deleted_at"] == nil { // 只删除未删除的
				id, _ := strconv.ParseInt(fmt.Sprintf("%v", r["id"]), 10, 64)
				ids = append(ids, id)
			}
		}
	} else if arg == "soft" || arg == "hard" {
		// 软删除或硬删除指定ID
		if len(os.Args) < 3 {
			fmt.Println("❌ 请指定机器人ID")
			return
		}
		id, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Printf("❌ 无效的机器人ID: %s\n", os.Args[2])
			return
		}
		ids = []int64{id}
		if arg == "soft" {
			softDeleteRobots(ctx, db, ids)
		} else {
			hardDeleteRobots(ctx, db, ids)
		}
		return
	} else {
		// 删除指定的ID列表
		for i := 1; i < len(os.Args); i++ {
			id, err := strconv.ParseInt(os.Args[i], 10, 64)
			if err != nil {
				fmt.Printf("❌ 无效的机器人ID: %s\n", os.Args[i])
				continue
			}
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		fmt.Println("❌ 没有有效的机器人ID")
		return
	}

	// 确认删除
	fmt.Printf("\n⚠️  将删除 %d 个机器人 (ID: %v)\n", len(ids), ids)
	fmt.Print("确认删除？(yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("已取消")
		return
	}

	// 执行软删除（推荐）
	softDeleteRobots(ctx, db, ids)
}

func softDeleteRobots(ctx gctx.Ctx, db gdb.DB, ids []int64) {
	fmt.Println("\n========== 执行软删除 ==========")
	
	for _, id := range ids {
		// 检查机器人是否存在
		count, err := db.Model("hg_trading_robot").
			Where("id", id).
			WhereNull("deleted_at").
			Count()
		if err != nil || count == 0 {
			fmt.Printf("⚠️  机器人 ID %d 不存在或已删除，跳过\n", id)
			continue
		}

		// 软删除（设置 deleted_at）
		result, err := db.Model("hg_trading_robot").
			Where("id", id).
			WhereNull("deleted_at").
			Data(gdb.Map{
				"deleted_at": gdb.Raw("NOW()"),
				"updated_at": gdb.Raw("NOW()"),
			}).
			Update()

		if err != nil {
			fmt.Printf("❌ 删除机器人 ID %d 失败: %v\n", id, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("✅ 已软删除机器人 ID %d\n", id)
		} else {
			fmt.Printf("⚠️  机器人 ID %d 未找到或已删除\n", id)
		}
	}

	fmt.Println("========== 删除完成 ==========")
}

func hardDeleteRobots(ctx gctx.Ctx, db gdb.DB, ids []int64) {
	fmt.Println("\n========== 执行硬删除 ==========")
	fmt.Println("⚠️  警告：硬删除将永久删除数据，无法恢复！")
	fmt.Print("确认硬删除？(yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("已取消")
		return
	}

	for _, id := range ids {
		// 硬删除（物理删除）
		result, err := db.Model("hg_trading_robot").
			Where("id", id).
			Delete()

		if err != nil {
			fmt.Printf("❌ 删除机器人 ID %d 失败: %v\n", id, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("✅ 已硬删除机器人 ID %d\n", id)
		} else {
			fmt.Printf("⚠️  机器人 ID %d 未找到\n", id)
		}
	}

	fmt.Println("========== 删除完成 ==========")
}

func getStatusText(status interface{}) string {
	if status == nil {
		return "未知"
	}
	s := fmt.Sprintf("%v", status)
	switch s {
	case "1":
		return "未启动"
	case "2":
		return "运行中"
	case "3":
		return "暂停"
	case "4":
		return "停用"
	default:
		return s
	}
}

