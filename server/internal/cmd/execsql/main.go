// Package main 执行 SQL 脚本
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()
	
	// 直接使用配置文件中的数据库连接
	db := g.DB()

	// 获取 SQL 文件路径
	sqlFile := "storage/data/add_strategy_group_id_to_robot.sql"
	if len(os.Args) > 1 {
		sqlFile = os.Args[1]
	}

	fmt.Printf("========== 执行 SQL 脚本: %s ==========\n\n", sqlFile)

	// 读取 SQL 文件
	sqlContent, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		fmt.Printf("❌ 读取 SQL 文件失败: %v\n", err)
		os.Exit(1)
	}

	// 分割 SQL 语句（按分号和换行）
	sqlStatements := splitSQL(string(sqlContent))

	// 执行每个 SQL 语句
	var successCount, skipCount, errorCount int

	for i, sql := range sqlStatements {
		sql = strings.TrimSpace(sql)
		if sql == "" || strings.HasPrefix(sql, "--") {
			continue
		}

		fmt.Printf("[%d/%d] 执行 SQL...\n", i+1, len(sqlStatements))
		if len(sql) > 100 {
			fmt.Printf("   SQL: %s...\n", sql[:100])
		} else {
			fmt.Printf("   SQL: %s\n", sql)
		}

		_, err := db.Exec(ctx, sql)
		if err != nil {
			// 检查是否是"已存在"的错误（可以忽略）
			errStr := err.Error()
			if strings.Contains(errStr, "Duplicate column name") ||
				strings.Contains(errStr, "Duplicate key name") ||
				strings.Contains(errStr, "already exists") {
				fmt.Printf("   ⚠️  跳过（已存在）: %v\n", err)
				skipCount++
			} else {
				fmt.Printf("   ❌ 执行失败: %v\n", err)
				errorCount++
			}
		} else {
			fmt.Printf("   ✅ 执行成功\n")
			successCount++
		}
		fmt.Println()
	}

	fmt.Println("========== 执行结果 ==========")
	fmt.Printf("✅ 成功: %d\n", successCount)
	fmt.Printf("⚠️  跳过: %d\n", skipCount)
	fmt.Printf("❌ 失败: %d\n", errorCount)
	fmt.Println("==============================")

	// 执行更新操作：从备注中提取策略组ID
	fmt.Println("\n========== 更新现有机器人的策略组ID ==========")
	updateRobotsFromRemark(ctx, db)
	
	// 查询结果
	fmt.Println("\n========== 查询结果 ==========")
	queryResults(ctx, db)
}

// splitSQL 分割 SQL 语句
func splitSQL(content string) []string {
	// 移除注释行
	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			cleanLines = append(cleanLines, line)
		}
	}
	content = strings.Join(cleanLines, "\n")

	// 按分号分割
	statements := strings.Split(content, ";")
	var result []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}
	return result
}

// updateRobotsFromRemark 从备注中提取策略组ID并更新
func updateRobotsFromRemark(ctx context.Context, db gdb.DB) {
	fmt.Println("从备注中提取策略组ID并更新机器人...")

	// 更新备注中包含"策略组ID: "的机器人
	result, err := db.Exec(ctx, `
		UPDATE hg_trading_robot 
		SET strategy_group_id = CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(remark, '策略组ID: ', -1), ' ', 1) AS UNSIGNED),
		    updated_at = NOW()
		WHERE remark LIKE '%策略组ID: %'
		  AND strategy_group_id IS NULL
		  AND deleted_at IS NULL
	`)

	if err != nil {
		fmt.Printf("❌ 更新失败: %v\n", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ 更新了 %d 个机器人的策略组ID\n", rowsAffected)
}

// queryResults 查询结果
func queryResults(ctx context.Context, db gdb.DB) {
	// 检查字段是否存在
	var columnExists int
	_, err := db.GetValue(ctx, `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'hg_trading_robot'
		  AND COLUMN_NAME = 'strategy_group_id'
	`, &columnExists)
	if err != nil {
		fmt.Printf("❌ 检查字段失败: %v\n", err)
		return
	}

	if columnExists == 0 {
		fmt.Println("❌ strategy_group_id 字段不存在")
		return
	}

	fmt.Println("✅ strategy_group_id 字段已存在")

	// 查询机器人策略组ID情况
	var robots []gdb.Record
	err = db.Model("hg_trading_robot r").
		LeftJoin("hg_trading_strategy_group g", "r.strategy_group_id = g.id").
		Fields("r.id, r.robot_name, r.strategy_group_id, g.group_name, r.remark").
		Where("r.deleted_at IS NULL").
		OrderDesc("r.id").
		Limit(20).
		Scan(&robots)

	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}

	fmt.Println("最近20个机器人的策略组ID情况:")
	fmt.Printf("%-8s %-20s %-15s %-30s\n", "ID", "机器人名称", "策略组ID", "策略组名称")
	fmt.Println(strings.Repeat("-", 80))

	for _, r := range robots {
		groupId := r["strategy_group_id"]
		groupName := r["group_name"]
		groupIdStr := "NULL"
		groupNameStr := "未设置"
		if groupId != nil {
			groupIdStr = fmt.Sprintf("%v", groupId)
		}
		if groupName != nil {
			groupNameStr = fmt.Sprintf("%v", groupName)
		}
		fmt.Printf("%-8s %-20s %-15s %-30s\n",
			fmt.Sprintf("%v", r["id"]),
			fmt.Sprintf("%v", r["robot_name"]),
			groupIdStr,
			groupNameStr)
	}

	// 统计未设置策略组ID的机器人数量
	var nullCount int
	_, err = db.GetValue(ctx, `
		SELECT COUNT(*) 
		FROM hg_trading_robot 
		WHERE deleted_at IS NULL 
		  AND strategy_group_id IS NULL
	`, &nullCount)
	if err != nil {
		fmt.Printf("❌ 统计失败: %v\n", err)
		return
	}

	fmt.Printf("\n未设置策略组ID的机器人数量: %d\n", nullCount)
}

