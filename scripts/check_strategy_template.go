package main

import (
	"fmt"
	"os"

	_ "hotgo/internal/packed"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	_ "hotgo/addons/modules"

	"hotgo/internal/dao"
	"hotgo/internal/global"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.GetInitCtx()
	global.Init(ctx)

	// 1. 查找策略组 "BTC-USDT 官方策略 V5.0 (我的副本)"
	fmt.Println("========== 1. 查找策略组 ==========")
	var groups []*entity.TradingStrategyGroup
	err := dao.TradingStrategyGroup.Ctx(ctx).
		Where("group_name LIKE ?", "%BTC-USDT 官方策略 V5.0%").
		Or("group_name LIKE ?", "%V5.0%我的副本%").
		OrderDesc("created_at").
		Scan(&groups)
	if err != nil {
		fmt.Printf("查询策略组失败: %v\n", err)
		return
	}

	if len(groups) == 0 {
		fmt.Println("未找到策略组")
		return
	}

	for _, g := range groups {
		fmt.Printf("策略组ID: %d, 名称: %s, 是否官方: %d, 是否激活: %d\n",
			g.Id, g.GroupName, g.IsOfficial, g.IsActive)
	}

	// 2. 检查 volatile + balanced 的策略模板
	fmt.Println("\n========== 2. 检查 volatile + balanced 的策略模板 ==========")
	for _, group := range groups {
		fmt.Printf("\n--- 策略组: %s (ID: %d) ---\n", group.GroupName, group.Id)

		var templates []*entity.TradingStrategyTemplate
		err := dao.TradingStrategyTemplate.Ctx(ctx).
			Where("group_id", group.Id).
			Where("market_state", "volatile").
			Where("risk_preference", "balanced").
			Where("is_active", 1).
			Scan(&templates)

		if err != nil {
			fmt.Printf("查询策略模板失败: %v\n", err)
			continue
		}

		if len(templates) == 0 {
			fmt.Println("未找到匹配的策略模板")
			// 检查是否有其他市场状态名称
			var allTemplates []*entity.TradingStrategyTemplate
			dao.TradingStrategyTemplate.Ctx(ctx).
				Where("group_id", group.Id).
				Where("risk_preference", "balanced").
				Where("is_active", 1).
				Scan(&allTemplates)
			if len(allTemplates) > 0 {
				fmt.Println("找到其他市场状态的策略模板:")
				for _, t := range allTemplates {
					fmt.Printf("  - ID: %d, 市场状态: %s, 杠杆: %d, 保证金: %.1f%%\n",
						t.Id, t.MarketState, t.Leverage, t.MarginPercent)
				}
			}
			continue
		}

		for _, t := range templates {
			fmt.Printf("策略模板ID: %d\n", t.Id)
			fmt.Printf("  市场状态: %s\n", t.MarketState)
			fmt.Printf("  风险偏好: %s\n", t.RiskPreference)
			fmt.Printf("  时间窗口: %d秒\n", t.MonitorWindow)
			fmt.Printf("  波动值: %.1f USDT\n", t.VolatilityThreshold)
			fmt.Printf("  杠杆: %dx\n", t.Leverage)
			fmt.Printf("  保证金: %.1f%%\n", t.MarginPercent)
			fmt.Printf("  止损: %.1f%%\n", t.StopLossPercent)
			fmt.Printf("  启动止盈: %.1f%%\n", t.AutoStartRetreatPercent)
			fmt.Printf("  止盈回撤: %.1f%%\n", t.ProfitRetreatPercent)
			fmt.Printf("  是否激活: %d\n", t.IsActive)
			if t.ConfigJson != "" {
				fmt.Printf("  其他配置: %s\n", t.ConfigJson)
			}
			fmt.Println()
		}
	}

	// 3. 检查所有策略组
	fmt.Println("\n========== 3. 检查所有包含 'V5.0' 或 '我的副本' 的策略组 ==========")
	var allGroups []*entity.TradingStrategyGroup
	dao.TradingStrategyGroup.Ctx(ctx).
		Where("group_name LIKE ?", "%V5.0%").
		Or("group_name LIKE ?", "%我的副本%").
		OrderDesc("created_at").
		Scan(&allGroups)

	for _, g := range allGroups {
		fmt.Printf("策略组ID: %d, 名称: %s\n", g.Id, g.GroupName)
	}

	// 4. 检查机器人的策略组ID配置
	fmt.Println("\n========== 4. 检查机器人的策略组ID配置 ==========")
	var robots []gdb.Record
	dao.TradingRobot.Ctx(ctx).
		LeftJoin("hg_trading_strategy_group g", "hg_trading_robot.strategy_group_id = g.id").
		Fields("hg_trading_robot.id, hg_trading_robot.robot_name, hg_trading_robot.strategy_group_id, g.group_name").
		Where("g.group_name LIKE ?", "%V5.0%我的副本%").
		Or("g.group_name LIKE ?", "%BTC-USDT 官方策略 V5.0%").
		OrderDesc("hg_trading_robot.id").
		Limit(10).
		Scan(&robots)

	if len(robots) == 0 {
		fmt.Println("未找到使用该策略组的机器人")
	} else {
		for _, r := range robots {
			fmt.Printf("机器人ID: %d, 名称: %s, 策略组ID: %d, 策略组名称: %s\n",
				r["id"], r["robot_name"], r["strategy_group_id"], r["group_name"])
		}
	}

	// 5. 检查市场状态值分布
	fmt.Println("\n========== 5. 检查市场状态值分布 ==========")
	var marketStates []gdb.Record
	dao.TradingStrategyTemplate.Ctx(ctx).
		Fields("market_state, COUNT(*) as count").
		Group("market_state").
		OrderDesc("count").
		Scan(&marketStates)

	for _, ms := range marketStates {
		fmt.Printf("市场状态: %s, 数量: %d\n", ms["market_state"], ms["count"])
	}

	os.Exit(0)
}

