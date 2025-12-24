//go:build tools
// +build tools

// 修复PostgreSQL主键自增问题
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
	fmt.Println("修复PostgreSQL主键自增问题")
	fmt.Println("===========================================\n")

	// 1. 修复 hg_trading_order 表
	fmt.Println("【1】修复 hg_trading_order 表...")
	
	// 检查表是否存在
	var orderTableExists int
	val, _ := g.DB().GetValue(ctx, `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_name = 'hg_trading_order'
	`)
	orderTableExists = val.Int()
	
	if orderTableExists > 0 {
		// 检查序列是否存在
		var seqExists int
		val, _ := g.DB().GetValue(ctx, `
			SELECT COUNT(*)
			FROM information_schema.sequences
			WHERE sequence_name = 'hg_trading_order_id_seq'
		`)
		seqExists = val.Int()
		
		if seqExists == 0 {
			fmt.Println("   创建序列 hg_trading_order_id_seq...")
			_, err := g.DB().Exec(ctx, `
				CREATE SEQUENCE IF NOT EXISTS hg_trading_order_id_seq
			`)
			if err != nil {
				fmt.Printf("   ✗ 创建序列失败: %v\n", err)
			} else {
				fmt.Println("   ✓ 序列创建成功")
			}
		}
		
		// 设置默认值
		fmt.Println("   设置 id 字段默认值...")
		_, err := g.DB().Exec(ctx, `
			ALTER TABLE hg_trading_order 
			ALTER COLUMN id SET DEFAULT nextval('hg_trading_order_id_seq')
		`)
		if err != nil {
			fmt.Printf("   ✗ 设置默认值失败: %v\n", err)
		} else {
			fmt.Println("   ✓ 默认值设置成功")
		}
		
		// 同步序列值
		fmt.Println("   同步序列当前值...")
		_, err = g.DB().Exec(ctx, `
			SELECT setval('hg_trading_order_id_seq', COALESCE((SELECT MAX(id) FROM hg_trading_order), 0) + 1, false)
		`)
		if err != nil {
			fmt.Printf("   ✗ 同步序列失败: %v\n", err)
		} else {
			fmt.Println("   ✓ 序列同步成功")
		}
	} else {
		fmt.Println("   ✗ hg_trading_order 表不存在")
	}

	// 2. 修复 hg_trading_execution_log 表
	fmt.Println("\n【2】修复 hg_trading_execution_log 表...")
	
	var execLogTableExists int
	val2, _ := g.DB().GetValue(ctx, `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_name = 'hg_trading_execution_log'
	`)
	execLogTableExists = val2.Int()
	
	if execLogTableExists > 0 {
		// 检查序列是否存在
		var seqExists int
		val3, _ := g.DB().GetValue(ctx, `
			SELECT COUNT(*)
			FROM information_schema.sequences
			WHERE sequence_name = 'hg_trading_execution_log_id_seq'
		`)
		seqExists = val3.Int()
		
		if seqExists == 0 {
			fmt.Println("   创建序列 hg_trading_execution_log_id_seq...")
			_, err := g.DB().Exec(ctx, `
				CREATE SEQUENCE IF NOT EXISTS hg_trading_execution_log_id_seq
			`)
			if err != nil {
				fmt.Printf("   ✗ 创建序列失败: %v\n", err)
			} else {
				fmt.Println("   ✓ 序列创建成功")
			}
		}
		
		// 设置默认值
		fmt.Println("   设置 id 字段默认值...")
		_, err := g.DB().Exec(ctx, `
			ALTER TABLE hg_trading_execution_log 
			ALTER COLUMN id SET DEFAULT nextval('hg_trading_execution_log_id_seq')
		`)
		if err != nil {
			fmt.Printf("   ✗ 设置默认值失败: %v\n", err)
		} else {
			fmt.Println("   ✓ 默认值设置成功")
		}
		
		// 同步序列值
		fmt.Println("   同步序列当前值...")
		_, err = g.DB().Exec(ctx, `
			SELECT setval('hg_trading_execution_log_id_seq', COALESCE((SELECT MAX(id) FROM hg_trading_execution_log), 0) + 1, false)
		`)
		if err != nil {
			fmt.Printf("   ✗ 同步序列失败: %v\n", err)
		} else {
			fmt.Println("   ✓ 序列同步成功")
		}
	} else {
		fmt.Println("   ✗ hg_trading_execution_log 表不存在")
	}

	// 3. 验证修复
	fmt.Println("\n【3】验证修复...")
	
	// 测试插入 hg_trading_order
	fmt.Println("   测试插入 hg_trading_order...")
	testOrderId, err := g.DB().Ctx(ctx).Raw(`
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
		fmt.Printf("   ✓ 测试插入成功，新ID=%v\n", testOrderId)
		// 删除测试记录
		g.DB().Exec(ctx, `DELETE FROM hg_trading_order WHERE id = ?`, testOrderId)
		fmt.Println("   ✓ 测试记录已删除")
	}

	fmt.Println("\n===========================================")
	fmt.Println("✅ 修复完成！")
	fmt.Println("===========================================")
	fmt.Println("\n不需要重启应用，等待下一个信号即可！")
}

