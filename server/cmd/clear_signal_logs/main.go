package main

import (
	"context"
	"fmt"
	"os"

	"hotgo/internal/logic/toogo"
)

func main() {
	ctx := context.Background()

	// 询问用户确认
	fmt.Println("⚠️  警告：此操作将清空预警记录，不可恢复！")
	fmt.Println("请选择操作：")
	fmt.Println("1. 清空所有预警记录")
	fmt.Println("2. 只删除未执行的预警记录（保留已执行的）")
	fmt.Println("3. 取消操作")
	fmt.Print("请输入选项 (1/2/3): ")

	var choice int
	fmt.Scanf("%d", &choice)

	switch choice {
	case 1:
		// 清空所有预警记录
		fmt.Println("\n正在清空所有预警记录...")
		robotService := toogo.NewToogoRobot()
		err := robotService.ClearSignalLogs(ctx, 0, false)
		if err != nil {
			fmt.Printf("❌ 清空失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ 清空成功！所有预警记录已删除。")

	case 2:
		// 只删除未执行的预警记录
		fmt.Println("\n正在删除未执行的预警记录...")
		robotService := toogo.NewToogoRobot()
		err := robotService.ClearSignalLogs(ctx, 0, true)
		if err != nil {
			fmt.Printf("❌ 删除失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ 删除成功！未执行的预警记录已删除。")

	case 3:
		fmt.Println("\n操作已取消。")
		os.Exit(0)

	default:
		fmt.Println("\n❌ 无效的选项，操作已取消。")
		os.Exit(1)
	}
}

