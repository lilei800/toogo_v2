// Package main 检查策略模板数据
package main

import (
	"fmt"

	_ "hotgo/internal/packed"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"hotgo/internal/dao"
	"hotgo/internal/global"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.GetInitCtx()
	global.Init(ctx)

	fmt.Println("========== 检查策略模板数据 ==========")

	// 1. 查找策略组 "BTC-USDT 官方策略 V5.0 (我的副本)"
	fmt.Println("1. 查找策略组:")
	var groups []*entity.TradingStrategyGroup
	err := g.DB().Model("hg_trading_strategy_group").
		Where("group_name LIKE ?", "%BTC-USDT 官方策略 V5.0%").
		WhereOr("group_name LIKE ?", "%V5.0%我的副本%").
		OrderDesc("created_at").
		Scan(&groups)
	if err != nil {
		fmt.Printf("   查询失败: %v\n", err)
		return
	}

	if len(groups) == 0 {
		fmt.Println("   未找到策略组")
		return
	}

	for _, g := range groups {
		fmt.Printf("   ID: %d, 名称: %s, 官方: %d, 激活: %d\n",
			g.Id, g.GroupName, g.IsOfficial, g.IsActive)
	}

	// 2. 检查 volatile + balanced 的策略模板
	fmt.Println("\n2. 检查 volatile + balanced 的策略模板:")
	for _, group := range groups {
		fmt.Printf("\n   策略组: %s (ID: %d)\n", group.GroupName, group.Id)

		var templates []*entity.TradingStrategyTemplate
		err := dao.TradingStrategyTemplate.Ctx(ctx).
			Where("group_id", group.Id).
			Where("market_state", "volatile").
			Where("risk_preference", "balanced").
			Where("is_active", 1).
			Scan(&templates)

		if err != nil {
			fmt.Printf("   查询失败: %v\n", err)
			continue
		}

		if len(templates) == 0 {
			fmt.Println("   ❌ 未找到匹配的策略模板")
			// 检查是否有其他市场状态名称
			var allTemplates []*entity.TradingStrategyTemplate
			dao.TradingStrategyTemplate.Ctx(ctx).
				Where("group_id", group.Id).
				Where("risk_preference", "balanced").
				Where("is_active", 1).
				Scan(&allTemplates)
			if len(allTemplates) > 0 {
				fmt.Println("   找到其他市场状态的策略模板:")
				for _, t := range allTemplates {
					fmt.Printf("     - ID: %d, 市场状态: %s, 杠杆: %d, 保证金: %.1f%%\n",
						t.Id, t.MarketState, t.Leverage, t.MarginPercent)
				}
			}
			continue
		}

		for _, t := range templates {
			fmt.Printf("   ✅ 找到策略模板 ID: %d\n", t.Id)
			fmt.Printf("      市场状态: %s\n", t.MarketState)
			fmt.Printf("      风险偏好: %s\n", t.RiskPreference)
			fmt.Printf("      时间窗口: %d秒\n", t.MonitorWindow)
			fmt.Printf("      波动值: %.1f USDT\n", t.VolatilityThreshold)
			fmt.Printf("      杠杆: %dx\n", t.Leverage)
			fmt.Printf("      保证金: %.1f%%\n", t.MarginPercent)
			fmt.Printf("      止损: %.1f%%\n", t.StopLossPercent)
			fmt.Printf("      启动止盈: %.1f%%\n", t.AutoStartRetreatPercent)
			fmt.Printf("      止盈回撤: %.1f%%\n", t.ProfitRetreatPercent)
			if t.ConfigJson != "" {
				fmt.Printf("      其他配置: %s\n", t.ConfigJson)
			}
			
			// 检查参数是否正确
			fmt.Println("\n      参数检查:")
			if t.Leverage != 10 {
				fmt.Printf("      ⚠️  杠杆错误: 期望 10x, 实际 %dx\n", t.Leverage)
			} else {
				fmt.Printf("      ✅ 杠杆正确: %dx\n", t.Leverage)
			}
			if t.MarginPercent != 10.0 {
				fmt.Printf("      ⚠️  保证金错误: 期望 10.0%%, 实际 %.1f%%\n", t.MarginPercent)
			} else {
				fmt.Printf("      ✅ 保证金正确: %.1f%%\n", t.MarginPercent)
			}
			if t.MonitorWindow != 60 {
				fmt.Printf("      ⚠️  时间窗口错误: 期望 60秒, 实际 %d秒\n", t.MonitorWindow)
			} else {
				fmt.Printf("      ✅ 时间窗口正确: %d秒\n", t.MonitorWindow)
			}
			if t.VolatilityThreshold != 50.0 {
				fmt.Printf("      ⚠️  波动值错误: 期望 50.0 USDT, 实际 %.1f USDT\n", t.VolatilityThreshold)
			} else {
				fmt.Printf("      ✅ 波动值正确: %.1f USDT\n", t.VolatilityThreshold)
			}
			if t.StopLossPercent != 10.0 {
				fmt.Printf("      ⚠️  止损错误: 期望 10.0%%, 实际 %.1f%%\n", t.StopLossPercent)
			} else {
				fmt.Printf("      ✅ 止损正确: %.1f%%\n", t.StopLossPercent)
			}
			if t.AutoStartRetreatPercent != 5.0 {
				fmt.Printf("      ⚠️  启动止盈错误: 期望 5.0%%, 实际 %.1f%%\n", t.AutoStartRetreatPercent)
			} else {
				fmt.Printf("      ✅ 启动止盈正确: %.1f%%\n", t.AutoStartRetreatPercent)
			}
			if t.ProfitRetreatPercent != 10.0 {
				fmt.Printf("      ⚠️  止盈回撤错误: 期望 10.0%%, 实际 %.1f%%\n", t.ProfitRetreatPercent)
			} else {
				fmt.Printf("      ✅ 止盈回撤正确: %.1f%%\n", t.ProfitRetreatPercent)
			}
		}
	}

	// 3. 检查机器人的策略组ID配置
	fmt.Println("\n3. 检查机器人的策略组ID配置:")
	var robots []gdb.Record
	g.DB().Model("hg_trading_robot r").
		LeftJoin("hg_trading_strategy_group g", "r.strategy_group_id = g.id").
		Fields("r.id, r.robot_name, r.strategy_group_id, g.group_name").
		Where("g.group_name LIKE ?", "%V5.0%我的副本%").
		WhereOr("g.group_name LIKE ?", "%BTC-USDT 官方策略 V5.0%").
		OrderDesc("r.id").
		Limit(10).
		Scan(&robots)

	if len(robots) == 0 {
		fmt.Println("   未找到使用该策略组的机器人")
	} else {
		for _, r := range robots {
			fmt.Printf("   机器人ID: %d, 名称: %s, 策略组ID: %d, 策略组: %s\n",
				r["id"].Int64(), r["robot_name"].String(), r["strategy_group_id"].Int64(), r["group_name"].String())
		}
	}

	// 4. 检查市场状态值分布
	fmt.Println("\n4. 检查市场状态值分布:")
	var marketStates []gdb.Record
	g.DB().Model("hg_trading_strategy_template").
		Fields("market_state, COUNT(*) as count").
		Group("market_state").
		OrderDesc("count").
		Scan(&marketStates)

	for _, ms := range marketStates {
		fmt.Printf("   市场状态: %s, 数量: %d\n", ms["market_state"].String(), ms["count"].Int64())
	}

	fmt.Println("\n========== 检查完成 ==========")
}

