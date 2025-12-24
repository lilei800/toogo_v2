//go:build tools
// +build tools

// 修复tenant_id字段问题
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
	fmt.Println("修复 tenant_id 字段问题")
	fmt.Println("===========================================\n")

	tables := []string{
		"hg_trading_order",
		"hg_trading_execution_log",
		"hg_trading_signal_log",
		"hg_trading_position",
	}

	for _, tableName := range tables {
		fmt.Printf("【处理表】%s\n", tableName)
		
		// 检查表是否存在
		val, _ := g.DB().GetValue(ctx, fmt.Sprintf(`
			SELECT COUNT(*) 
			FROM information_schema.tables 
			WHERE table_name = '%s'
		`, tableName))
		
		if val.Int() == 0 {
			fmt.Printf("   ✗ 表不存在，跳过\n\n")
			continue
		}
		
		// 检查tenant_id字段是否存在
		val2, _ := g.DB().GetValue(ctx, fmt.Sprintf(`
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_name = '%s' AND column_name = 'tenant_id'
		`, tableName))
		
		if val2.Int() == 0 {
			// 字段不存在，添加
			fmt.Println("   添加 tenant_id 字段...")
			_, err := g.DB().Exec(ctx, fmt.Sprintf(`
				ALTER TABLE %s 
				ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 0
			`, tableName))
			if err != nil {
				fmt.Printf("   ✗ 添加字段失败: %v\n", err)
			} else {
				fmt.Println("   ✓ 字段添加成功")
			}
		} else {
			// 字段存在，检查是否有默认值
			var columnDefault string
			val3, _ := g.DB().GetValue(ctx, fmt.Sprintf(`
				SELECT column_default
				FROM information_schema.columns
				WHERE table_name = '%s' AND column_name = 'tenant_id'
			`, tableName))
			columnDefault = val3.String()
			
			if columnDefault == "" {
				fmt.Println("   设置 tenant_id 默认值...")
				_, err := g.DB().Exec(ctx, fmt.Sprintf(`
					ALTER TABLE %s 
					ALTER COLUMN tenant_id SET DEFAULT 0
				`, tableName))
				if err != nil {
					fmt.Printf("   ✗ 设置默认值失败: %v\n", err)
				} else {
					fmt.Println("   ✓ 默认值设置成功")
				}
			} else {
				fmt.Printf("   ✓ 字段已有默认值: %s\n", columnDefault)
			}
			
			// 检查是否允许NULL
			var isNullable string
			val4, _ := g.DB().GetValue(ctx, fmt.Sprintf(`
				SELECT is_nullable
				FROM information_schema.columns
				WHERE table_name = '%s' AND column_name = 'tenant_id'
			`, tableName))
			isNullable = val4.String()
			
			if isNullable == "NO" {
				// 先更新现有NULL值
				fmt.Println("   更新现有NULL值...")
				g.DB().Exec(ctx, fmt.Sprintf(`
					UPDATE %s SET tenant_id = 0 WHERE tenant_id IS NULL
				`, tableName))
				fmt.Println("   ✓ NULL值已更新")
			}
		}
		
		fmt.Println()
	}

	// 验证修复
	fmt.Println("【验证】测试插入 hg_trading_order...")
	testId, err := g.DB().Ctx(ctx).Raw(`
		INSERT INTO hg_trading_order (
			user_id, robot_id, symbol, direction, order_type, 
			exchange_side, price, quantity, status, created_at
		) VALUES (
			1, 35, 'BTCUSDT', 'test', 'MARKET', 
			'BUY', 87000, 0.001, 0, NOW()
		) RETURNING id
	`).Value()
	
	if err != nil {
		fmt.Printf("   ✗ 测试插入失败: %v\n", err)
	} else {
		fmt.Printf("   ✓ 测试插入成功，新ID=%v\n", testId)
		// 删除测试记录
		g.DB().Exec(ctx, `DELETE FROM hg_trading_order WHERE id = ?`, testId)
		fmt.Println("   ✓ 测试记录已删除")
	}

	fmt.Println("\n===========================================")
	fmt.Println("✅ 修复完成！")
	fmt.Println("===========================================")
	fmt.Println("\n不需要重启应用，等待下一个信号即可！")
}

