//go:build tools
// +build tools

// 一次性修复所有NOT NULL但没有默认值的字段
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
	fmt.Println("一次性修复所有字段默认值问题")
	fmt.Println("===========================================\n")

	// 需要修复的表
	tables := []string{
		"hg_trading_order",
		"hg_trading_execution_log",
		"hg_trading_signal_log",
	}

	// 常见字段的默认值映射
	defaultValues := map[string]string{
		// ID相关
		"id":                "nextval序列",
		"tenant_id":         "0",
		"user_id":           "0",
		"robot_id":          "0",
		
		// 字符串字段
		"exchange":          "''",
		"symbol":            "''",
		"direction":         "''",
		"order_type":        "''",
		"order_sn":          "''",
		"exchange_order_id": "''",
		"market_state":      "''",
		"risk_level":        "''",
		"signal_type":       "''",
		"signal_source":     "''",
		"event_type":        "''",
		"status_str":        "''",
		"message":           "''",
		
		// 数字字段
		"status":            "0",
		"quantity":          "0",
		"price":             "0",
		"open_price":        "0",
		"avg_price":         "0",
		"mark_price":        "0",
		"leverage":          "1",
		"margin":            "0",
		"open_margin":       "0",
		"open_fee":          "0",
		"power_consumed":    "0",  // 新发现的字段
		"executed":          "0",
		"is_processed":      "0",
		"order_id":          "0",
		"signal_log_id":     "0",
		"strategy_id":       "0",
		"strategy_group_id": "0",
		
		// 小数字段
		"signal_strength":   "0",
		"threshold":         "0",
		"current_price":     "0",
		"window_min_price":  "0",
		"window_max_price":  "0",
		"target_price":      "0",
		"stop_loss":         "0",
		"take_profit":       "0",
		
		// 百分比字段
		"margin_percent":             "0",
		"margin_percent_min":         "0",
		"margin_percent_max":         "0",
		"profit_retreat_percent":     "0",
		"stop_loss_percent":          "0",
		"auto_start_retreat_percent": "0",
		
		// 时间字段 - 不设置默认值，因为通常由应用程序处理
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
		
		// 查找所有NOT NULL但没有默认值的字段
		rows, err := g.DB().Ctx(ctx).Raw(fmt.Sprintf(`
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns
			WHERE table_name = '%s'
			  AND is_nullable = 'NO'
			  AND column_default IS NULL
			  AND column_name NOT LIKE '%%_at'
			ORDER BY ordinal_position
		`, tableName)).All()
		
		if err != nil {
			fmt.Printf("   ✗ 查询失败: %v\n\n", err)
			continue
		}
		
		if len(rows) == 0 {
			fmt.Println("   ✓ 所有NOT NULL字段都有默认值")
		} else {
			fmt.Printf("   发现 %d 个缺少默认值的字段:\n", len(rows))
			
			for _, row := range rows {
				columnName := row["column_name"].String()
				dataType := row["data_type"].String()
				
				// 确定默认值
				var defaultValue string
				if dv, ok := defaultValues[columnName]; ok {
					if dv == "nextval序列" {
						// ID字段，跳过（已经单独处理）
						continue
					}
					defaultValue = dv
				} else {
					// 根据数据类型猜测默认值
					switch {
					case dataType == "character varying" || dataType == "text":
						defaultValue = "''"
					case dataType == "smallint" || dataType == "integer" || dataType == "bigint":
						defaultValue = "0"
					case dataType == "numeric" || dataType == "double precision":
						defaultValue = "0"
					case dataType == "boolean":
						defaultValue = "false"
					default:
						fmt.Printf("     - %s (%s) - ⚠️ 未知类型，跳过\n", columnName, dataType)
						continue
					}
				}
				
				// 设置默认值
				sql := fmt.Sprintf(`
					ALTER TABLE %s 
					ALTER COLUMN %s SET DEFAULT %s
				`, tableName, columnName, defaultValue)
				
				_, err := g.DB().Exec(ctx, sql)
				if err != nil {
					fmt.Printf("     - %s (%s) - ✗ 失败: %v\n", columnName, dataType, err)
				} else {
					fmt.Printf("     - %s (%s) - ✓ 默认值: %s\n", columnName, dataType, defaultValue)
				}
			}
		}
		
		fmt.Println()
	}

	// 验证修复
	fmt.Println("【验证】测试插入 hg_trading_order...")
	testId, err := g.DB().Ctx(ctx).Raw(`
		INSERT INTO hg_trading_order (
			user_id, robot_id, symbol, direction, order_type, 
			exchange_side, price, quantity, status
		) VALUES (
			1, 35, 'BTCUSDT', 'test', 'MARKET', 
			'BUY', 87000, 0.001, 0
		) RETURNING id
	`).Value()
	
	if err != nil {
		fmt.Printf("   ✗ 测试插入失败: %v\n", err)
		fmt.Println("\n可能还有其他缺失的字段，请查看错误信息")
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
	fmt.Println("\n提示：如果还有字段报错，可以再次运行此工具")
}

