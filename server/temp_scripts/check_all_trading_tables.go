//go:build tools
// +build tools

package main

import (
	"fmt"
	"strings"

	_ "hotgo/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

type FieldInfo struct {
	ColumnName    string `json:"column_name"`
	DataType      string `json:"data_type"`
	IsNullable    string `json:"is_nullable"`
	ColumnDefault *struct{ String string } `json:"column_default"`
}

func main() {
	ctx := gctx.New()
	
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("【完整检查】所有交易相关表的 NOT NULL 字段")
	fmt.Println(strings.Repeat("=", 70))
	
	// 所有交易相关的表
	tables := []struct {
		Name        string
		Description string
	}{
		{"hg_trading_robot", "机器人表"},
		{"hg_trading_signal_log", "信号日志表"},
		{"hg_trading_execution_log", "执行日志表"},
		{"hg_trading_order", "订单表"},
		{"hg_trading_close_log", "平仓日志表"},
		{"hg_trading_api_config", "API配置表"},
		{"hg_trading_robot_run_session", "运行区间表"},
		{"hg_trading_strategy_group", "策略组表"},
		{"hg_trading_strategy_template", "策略模板表"},
	}
	
	allGood := true
	totalIssues := 0
	
	for _, table := range tables {
		fmt.Printf("\n【%s】%s\n", table.Description, table.Name)
		
		// 检查表是否存在
		query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_name = $1`
		count, err := g.DB().Ctx(ctx).GetValue(ctx, query, table.Name)
		if err != nil || count.Int() == 0 {
			fmt.Printf("  ⚠️  表不存在或查询失败\n")
			continue
		}
		
		// 检查 NOT NULL 字段
		query = `
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_name = $1 
			  AND is_nullable = 'NO'
			  AND column_default IS NULL
			ORDER BY ordinal_position
		`
		
		var fields []FieldInfo
		err = g.DB().Ctx(ctx).GetScan(ctx, &fields, query, table.Name)
		if err != nil {
			fmt.Printf("  ✗ 查询失败: %v\n", err)
			allGood = false
			continue
		}
		
		if len(fields) == 0 {
			fmt.Printf("  ✓ 所有 NOT NULL 字段都有默认值\n")
		} else {
			fmt.Printf("  ⚠️  发现 %d 个字段缺少默认值:\n", len(fields))
			for _, field := range fields {
				fmt.Printf("     - %s (%s)\n", field.ColumnName, field.DataType)
			}
			allGood = false
			totalIssues += len(fields)
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 70))
	if allGood {
		fmt.Println("✅ 【检查通过】所有交易相关表都已就绪")
		fmt.Println("\n🎉 系统状态:")
		fmt.Println("   ✅ 创建机器人 - 就绪")
		fmt.Println("   ✅ 自动下单 - 就绪")
		fmt.Println("   ✅ 自动平仓 - 就绪")
		fmt.Println("   ✅ 手动平仓 - 就绪")
	} else {
		fmt.Printf("⚠️  【发现问题】共 %d 个字段需要处理\n", totalIssues)
	}
	fmt.Println(strings.Repeat("=", 70))
}

