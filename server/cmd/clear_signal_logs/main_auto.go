//go:build tools
// +build tools

package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"hotgo/internal/logic/toogo"
)

func main() {
	ctx := context.Background()

	// 默认清空所有预警记录
	fmt.Println("⚠️  警告：此操作将清空所有预警记录，不可恢复！")
	fmt.Println("正在清空所有预警记录...")

	robotService := toogo.NewToogoRobot()
	err := robotService.ClearSignalLogs(ctx, 0, false)
	if err != nil {
		fmt.Printf("❌ 清空失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 清空成功！所有预警记录已删除。")
}

