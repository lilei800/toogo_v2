//go:build tools
// +build tools

// 数据库修复工具 - 添加 is_processed 字段
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
	fmt.Println("数据库修复工具 - 添加 is_processed 字段")
	fmt.Println("===========================================\n")

	// 步骤1: 检查字段是否存在
	fmt.Println("【步骤1】检查 is_processed 字段...")
	countVal, err := g.DB().GetValue(ctx, `
		SELECT COUNT(*) 
		FROM information_schema.columns
		WHERE table_name = 'hg_trading_signal_log' 
		  AND column_name = 'is_processed'
	`)
	
	if err != nil {
		fmt.Printf("   ✗ 检查失败: %v\n", err)
		return
	}

	count := countVal.Int()
	if count > 0 {
		fmt.Println("   ✓ is_processed 字段已存在，无需添加")
	} else {
		fmt.Println("   ✗ is_processed 字段不存在，开始添加...")
		
		// 添加 is_processed 字段
		_, err = g.DB().Exec(ctx, `
			ALTER TABLE hg_trading_signal_log 
			ADD COLUMN IF NOT EXISTS is_processed SMALLINT NOT NULL DEFAULT 0
		`)
		if err != nil {
			fmt.Printf("   ✗ 添加字段失败: %v\n", err)
			return
		}
		fmt.Println("   ✓ is_processed 字段添加成功")
	}

	// 步骤2: 添加其他必需字段
	fmt.Println("\n【步骤2】检查其他必需字段...")
	
	fields := map[string]string{
		"window_min_price": "NUMERIC(20,8) NOT NULL DEFAULT 0",
		"window_max_price": "NUMERIC(20,8) NOT NULL DEFAULT 0",
		"threshold":        "NUMERIC(20,8) NOT NULL DEFAULT 0",
		"market_state":     "VARCHAR(50) NOT NULL DEFAULT ''",
	}

	for fieldName, fieldDef := range fields {
		fieldCountVal, err := g.DB().GetValue(ctx, fmt.Sprintf(`
			SELECT COUNT(*) 
			FROM information_schema.columns
			WHERE table_name = 'hg_trading_signal_log' 
			  AND column_name = '%s'
		`, fieldName))
		
		if err == nil && fieldCountVal.Int() == 0 {
			sql := fmt.Sprintf("ALTER TABLE hg_trading_signal_log ADD COLUMN IF NOT EXISTS %s %s", fieldName, fieldDef)
			_, err = g.DB().Exec(ctx, sql)
			if err != nil {
				fmt.Printf("   ⚠ 添加 %s 字段失败: %v\n", fieldName, err)
			} else {
				fmt.Printf("   ✓ 添加 %s 字段成功\n", fieldName)
			}
		}
	}

	// 步骤3: 创建执行日志表
	fmt.Println("\n【步骤3】创建 hg_trading_execution_log 表...")
	
	_, err = g.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS hg_trading_execution_log (
			id BIGSERIAL PRIMARY KEY,
			signal_log_id BIGINT NOT NULL DEFAULT 0,
			robot_id BIGINT NOT NULL DEFAULT 0,
			order_id BIGINT NOT NULL DEFAULT 0,
			event_type VARCHAR(50) NOT NULL DEFAULT '',
			event_data TEXT,
			status VARCHAR(20) NOT NULL DEFAULT '',
			message VARCHAR(500) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Printf("   ✗ 创建表失败: %v\n", err)
	} else {
		fmt.Println("   ✓ hg_trading_execution_log 表创建成功")
	}

	// 步骤4: 创建索引
	fmt.Println("\n【步骤4】创建索引...")
	
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_is_processed_robot ON hg_trading_signal_log(robot_id, is_processed, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_executed_processed ON hg_trading_signal_log(executed, is_processed, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_signal_log_id ON hg_trading_execution_log(signal_log_id)",
		"CREATE INDEX IF NOT EXISTS idx_robot_time ON hg_trading_execution_log(robot_id, created_at)",
	}

	for _, sql := range indexes {
		_, err = g.DB().Exec(ctx, sql)
		if err != nil {
			fmt.Printf("   ⚠ 创建索引失败: %v\n", err)
		}
	}
	fmt.Println("   ✓ 索引创建完成")

	// 步骤5: 更新现有记录
	fmt.Println("\n【步骤5】更新现有记录...")
	
	result, err := g.DB().Exec(ctx, `
		UPDATE hg_trading_signal_log 
		SET is_processed = 0 
		WHERE is_processed IS NULL
	`)
	if err != nil {
		fmt.Printf("   ⚠ 更新失败: %v\n", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✓ 更新了 %d 条记录\n", rows)
	}

	// 步骤6: 验证修复
	fmt.Println("\n【步骤6】验证修复结果...")
	
	type FieldInfo struct {
		ColumnName    string `json:"column_name"`
		DataType      string `json:"data_type"`
		IsNullable    string `json:"is_nullable"`
		ColumnDefault string `json:"column_default"`
	}
	
	var fieldInfo FieldInfo
	err = g.DB().Ctx(ctx).Raw(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = 'hg_trading_signal_log' 
		  AND column_name = 'is_processed'
	`).Scan(&fieldInfo)
	
	if err != nil {
		fmt.Printf("   ✗ 验证失败: %v\n", err)
	} else {
		fmt.Println("   ✓ is_processed 字段信息:")
		fmt.Printf("     - 字段名: %s\n", fieldInfo.ColumnName)
		fmt.Printf("     - 数据类型: %s\n", fieldInfo.DataType)
		fmt.Printf("     - 允许NULL: %s\n", fieldInfo.IsNullable)
		fmt.Printf("     - 默认值: %s\n", fieldInfo.ColumnDefault)
	}

	fmt.Println("\n===========================================")
	fmt.Println("✅ 修复完成！")
	fmt.Println("===========================================")
	fmt.Println("\n建议下一步:")
	fmt.Println("1. 运行诊断工具验证: go run diagnose_trading.go")
	fmt.Println("2. 重启应用服务")
	fmt.Println("3. 等待下一个交易信号，观察是否正常下单")
}

