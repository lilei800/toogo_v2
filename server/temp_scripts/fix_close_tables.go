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

// 根据数据类型获取默认值
func getDefaultValueForType(dataType string) string {
	dataType = strings.ToLower(dataType)
	
	if strings.Contains(dataType, "int") || strings.Contains(dataType, "serial") {
		return "0"
	}
	if strings.Contains(dataType, "numeric") || strings.Contains(dataType, "decimal") || 
	   strings.Contains(dataType, "float") || strings.Contains(dataType, "double") || 
	   strings.Contains(dataType, "real") {
		return "0"
	}
	if strings.Contains(dataType, "char") || strings.Contains(dataType, "text") {
		return "''"
	}
	if strings.Contains(dataType, "bool") {
		return "false"
	}
	if strings.Contains(dataType, "timestamp") || strings.Contains(dataType, "date") || 
	   strings.Contains(dataType, "time") {
		return "CURRENT_TIMESTAMP"
	}
	if strings.Contains(dataType, "json") {
		return "'{}'"
	}
	
	return "UNKNOWN"
}

func main() {
	ctx := gctx.New()
	
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("【修复平仓相关表】PostgreSQL NOT NULL 字段默认值")
	fmt.Println(strings.Repeat("=", 60))
	
	// 1. 修复 hg_trading_close_log 表
	fmt.Println("\n【修复表】hg_trading_close_log")
	
	// 先确保 id 字段有序列
	fmt.Println("  1. 检查并创建 id 序列...")
	createSeqSQL := `
		DO $$
		BEGIN
			-- 创建序列（如果不存在）
			IF NOT EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'hg_trading_close_log_id_seq') THEN
				CREATE SEQUENCE hg_trading_close_log_id_seq;
				RAISE NOTICE '序列已创建';
			ELSE
				RAISE NOTICE '序列已存在';
			END IF;
			
			-- 设置序列默认值
			EXECUTE 'ALTER TABLE hg_trading_close_log ALTER COLUMN id SET DEFAULT nextval(''hg_trading_close_log_id_seq'')';
			
			-- 同步序列值
			PERFORM setval('hg_trading_close_log_id_seq', COALESCE((SELECT MAX(id) FROM hg_trading_close_log), 0) + 1, false);
			
			RAISE NOTICE 'id 序列配置完成';
		END
		$$;
	`
	_, err := g.DB().Ctx(ctx).Exec(ctx, createSeqSQL)
	if err != nil {
		fmt.Printf("     ✗ 配置序列失败: %v\n", err)
	} else {
		fmt.Printf("     ✓ id 序列配置完成\n")
	}
	
	// 2. 修复其他字段的默认值
	fmt.Println("  2. 设置字段默认值...")
	closeLogFields := map[string]string{
		"tenant_id":        "0",
		"user_id":          "0",
		"robot_id":         "0",
		"order_id":         "0",
		"order_sn":         "''",
		"symbol":           "''",
		"direction":        "''",
		"open_price":       "0",
		"close_price":      "0",
		"quantity":         "0",
		"leverage":         "0",
		"margin":           "0",
		"realized_profit":  "0",
		"close_reason":     "''",
		"open_time":        "CURRENT_TIMESTAMP",
		"close_time":       "CURRENT_TIMESTAMP",
	}
	
	for field, defaultValue := range closeLogFields {
		sql := fmt.Sprintf(`ALTER TABLE "hg_trading_close_log" ALTER COLUMN "%s" SET DEFAULT %s`, field, defaultValue)
		_, err := g.DB().Ctx(ctx).Exec(ctx, sql)
		if err != nil {
			fmt.Printf("     ✗ %s: %v\n", field, err)
		} else {
			fmt.Printf("     ✓ %s = %s\n", field, defaultValue)
		}
	}
	
	// 3. 修复 hg_trading_order 表的 open_time 字段
	fmt.Println("\n【修复表】hg_trading_order")
	fmt.Println("  设置 open_time 默认值...")
	sql := `ALTER TABLE "hg_trading_order" ALTER COLUMN "open_time" SET DEFAULT CURRENT_TIMESTAMP`
	_, err = g.DB().Ctx(ctx).Exec(ctx, sql)
	if err != nil {
		fmt.Printf("     ✗ open_time: %v\n", err)
	} else {
		fmt.Printf("     ✓ open_time = CURRENT_TIMESTAMP\n")
	}
	
	// 4. 验证修复结果
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("【验证修复结果】")
	fmt.Println(strings.Repeat("=", 60))
	
	tables := []string{"hg_trading_close_log", "hg_trading_order"}
	allGood := true
	
	for _, tableName := range tables {
		fmt.Printf("\n【验证表】%s\n", tableName)
		
		query := `
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_name = $1 
			  AND is_nullable = 'NO'
			  AND column_default IS NULL
			ORDER BY ordinal_position
		`
		
		var fields []FieldInfo
		err := g.DB().Ctx(ctx).GetScan(ctx, &fields, query, tableName)
		if err != nil {
			fmt.Printf("✗ 查询失败: %v\n", err)
			allGood = false
			continue
		}
		
		if len(fields) == 0 {
			fmt.Printf("✓ 所有 NOT NULL 字段都有默认值\n")
		} else {
			fmt.Printf("⚠️  仍有 %d 个字段缺少默认值:\n", len(fields))
			for _, field := range fields {
				fmt.Printf("   - %s (%s)\n", field.ColumnName, field.DataType)
			}
			allGood = false
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	if allGood {
		fmt.Println("✅ 修复完成：所有平仓相关表已就绪")
		fmt.Println("\n【提示】现在可以正常执行平仓操作了！")
	} else {
		fmt.Println("⚠️  修复完成但有警告，请检查上述信息")
	}
	fmt.Println(strings.Repeat("=", 60))
}

