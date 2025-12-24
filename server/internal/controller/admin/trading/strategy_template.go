package trading

import (
	"context"
	"encoding/json"
	"fmt"

	"hotgo/api/admin/trading"
	"hotgo/internal/dao"
	"hotgo/internal/logic/toogo"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// StrategyTemplate 策略模板控制器
var StrategyTemplate = cStrategyTemplate{}

type cStrategyTemplate struct{}

// List 策略模板列表
func (c *cStrategyTemplate) List(ctx context.Context, req *trading.StrategyTemplateListReq) (res *trading.StrategyTemplateListRes, err error) {
	m := g.DB().Model("hg_trading_strategy_template").Safe()

	if req.GroupId > 0 {
		m = m.Where("group_id", req.GroupId)
	}
	if req.RiskPreference != "" {
		m = m.Where("risk_preference", req.RiskPreference)
	}
	if req.MarketState != "" {
		m = m.Where("market_state", req.MarketState)
	}

	total, _ := m.Count()

	// 使用 entity 结构体确保返回 camelCase 的 JSON 字段名
	var list []*entity.TradingStrategyTemplate
	err = m.Page(req.Page, req.PageSize).Order("sort ASC, id ASC").Scan(&list)
	if err != nil {
		return nil, err
	}

	res = &trading.StrategyTemplateListRes{
		List:  list,
		Total: total,
		Page:  req.Page,
	}
	return
}

// Create 创建策略模板
func (c *cStrategyTemplate) Create(ctx context.Context, req *trading.StrategyTemplateCreateReq) (res *trading.StrategyTemplateCreateRes, err error) {
	// 使用前端提交的完整 configJson（包含所有手动配置参数）
	configJsonStr := req.ConfigJson
	if configJsonStr == "" {
		// 如果前端没有提供 configJson，则构建一个基础的
		configJson := map[string]interface{}{}
		configJsonBytes, _ := json.Marshal(configJson)
		configJsonStr = string(configJsonBytes)
	}

	// 从 configJson 中解析出关键字段（用于直接字段）
	var configData map[string]interface{}
	leverage := req.LeverageMin // 默认使用 Min 值
	marginPercent := req.MarginPercentMin
	stopLossPercent := req.StopLossPercent
	autoStartRetreatPercent := req.AutoStartRetreatPercent
	profitRetreatPercent := req.ProfitRetreatPercent
	monitorWindow := req.MonitorWindow
	volatilityThreshold := req.VolatilityThreshold

	if err := json.Unmarshal([]byte(configJsonStr), &configData); err == nil {
		if lev, ok := configData["leverage"].(float64); ok {
			leverage = int(lev)
		}
		if margin, ok := configData["marginPercent"].(float64); ok {
			marginPercent = margin
		}
		if stopLoss, ok := configData["stopLossPercent"].(float64); ok {
			stopLossPercent = stopLoss
		}
		if autoStart, ok := configData["autoStartRetreatPercent"].(float64); ok {
			autoStartRetreatPercent = autoStart
		}
		if profitRetreat, ok := configData["profitRetreatPercent"].(float64); ok {
			profitRetreatPercent = profitRetreat
		}
		if monitor, ok := configData["monitorWindow"].(float64); ok {
			monitorWindow = int(monitor)
		}
		if volatility, ok := configData["volatilityThreshold"].(float64); ok {
			volatilityThreshold = volatility
		}
	}

	data := g.Map{
		"group_id":                   req.GroupId,
		"strategy_key":               req.StrategyKey,
		"strategy_name":              req.StrategyName,
		"risk_preference":            req.RiskPreference,
		"market_state":               req.MarketState,
		"monitor_window":             monitorWindow,
		"volatility_threshold":       volatilityThreshold,
		"leverage":                   leverage,
		"margin_percent":             marginPercent,
		"stop_loss_percent":          stopLossPercent,
		"profit_retreat_percent":     profitRetreatPercent,
		"auto_start_retreat_percent": autoStartRetreatPercent,
		"config_json":                configJsonStr,
		"description":                req.Description,
		"is_active":                  req.IsActive,
		"sort":                       req.Sort,
		"created_at":                 gtime.Now(),
		"updated_at":                 gtime.Now(),
	}

	_, err = g.DB().Model("hg_trading_strategy_template").Insert(data)
	if err != nil {
		return nil, err
	}

	res = &trading.StrategyTemplateCreateRes{}
	return
}

// Update 更新策略模板
func (c *cStrategyTemplate) Update(ctx context.Context, req *trading.StrategyTemplateUpdateReq) (res *trading.StrategyTemplateUpdateRes, err error) {
	// 检查是否有机器人绑定（如果未确认，则返回绑定信息）
	if !req.Confirmed {
		// 先获取策略模板的 group_id
		var template *entity.TradingStrategyTemplate
		err = dao.TradingStrategyTemplate.Ctx(ctx).Where(dao.TradingStrategyTemplate.Columns().Id, req.Id).Scan(&template)
		if err != nil {
			return nil, err
		}
		if template == nil {
			return nil, gerror.New("策略模板不存在")
		}
		
		if template.GroupId > 0 {
			// 检查该策略组是否有机器人绑定
			robotCount, err := g.DB().Model("hg_trading_robot").
				Where("strategy_group_id", template.GroupId).
				WhereNull("deleted_at").
				Count()
			if err != nil {
				return nil, gerror.Wrap(err, "检查机器人绑定失败")
			}
			if robotCount > 0 {
				// 获取绑定的机器人列表（最多10个，用于提示）
				type RobotInfo struct {
					Id       int64  `json:"id"`
					RobotName string `json:"robot_name"`
					Status   int    `json:"status"`
				}
				var robots []RobotInfo
				err = g.DB().Model("hg_trading_robot").
					Fields("id", "robot_name", "status").
					Where("strategy_group_id", template.GroupId).
					WhereNull("deleted_at").
					Limit(10).
					Scan(&robots)
				if err != nil {
					return nil, gerror.Wrap(err, "查询机器人列表失败")
				}
				
				robotNames := make([]string, 0, len(robots))
				for _, robot := range robots {
					if robot.RobotName != "" {
						robotNames = append(robotNames, robot.RobotName)
					}
				}
				
				// 返回需要确认的错误信息
				msg := fmt.Sprintf("该策略模板所属的策略组已被%d个机器人绑定", robotCount)
				if len(robotNames) > 0 {
					if len(robotNames) >= 10 {
						msg += fmt.Sprintf("，包括：%s等", robotNames[0])
					} else {
						msg += fmt.Sprintf("：%s", robotNames[0])
						if len(robotNames) > 1 {
							msg += fmt.Sprintf("、%s", robotNames[1])
						}
						if len(robotNames) > 2 {
							msg += "等"
						}
					}
				}
				msg += "，修改策略模板会影响这些机器人的运行。请确认是否继续修改？"
				return nil, gerror.New(msg)
			}
		}
	}

	// 使用前端提交的完整 configJson（包含所有手动配置参数）
	configJsonStr := req.ConfigJson
	if configJsonStr == "" {
		// 如果前端没有提供 configJson，则构建一个基础的
		configJson := map[string]interface{}{}
		configJsonBytes, _ := json.Marshal(configJson)
		configJsonStr = string(configJsonBytes)
	}

	// 从 configJson 中解析出关键字段（用于直接字段）
	var configData map[string]interface{}
	leverage := req.LeverageMin // 默认使用 Min 值
	marginPercent := req.MarginPercentMin
	stopLossPercent := req.StopLossPercent
	autoStartRetreatPercent := req.AutoStartRetreatPercent
	profitRetreatPercent := req.ProfitRetreatPercent
	monitorWindow := req.MonitorWindow
	volatilityThreshold := req.VolatilityThreshold

	if err := json.Unmarshal([]byte(configJsonStr), &configData); err == nil {
		if lev, ok := configData["leverage"].(float64); ok {
			leverage = int(lev)
		}
		if margin, ok := configData["marginPercent"].(float64); ok {
			marginPercent = margin
		}
		if stopLoss, ok := configData["stopLossPercent"].(float64); ok {
			stopLossPercent = stopLoss
		}
		if autoStart, ok := configData["autoStartRetreatPercent"].(float64); ok {
			autoStartRetreatPercent = autoStart
		}
		if profitRetreat, ok := configData["profitRetreatPercent"].(float64); ok {
			profitRetreatPercent = profitRetreat
		}
		if monitor, ok := configData["monitorWindow"].(float64); ok {
			monitorWindow = int(monitor)
		}
		if volatility, ok := configData["volatilityThreshold"].(float64); ok {
			volatilityThreshold = volatility
		}
	}

	data := g.Map{
		"strategy_name":              req.StrategyName,
		"risk_preference":            req.RiskPreference,
		"market_state":               req.MarketState,
		"monitor_window":             monitorWindow,
		"volatility_threshold":       volatilityThreshold,
		"leverage":                   leverage,
		"margin_percent":             marginPercent,
		"stop_loss_percent":          stopLossPercent,
		"profit_retreat_percent":     profitRetreatPercent,
		"auto_start_retreat_percent": autoStartRetreatPercent,
		"config_json":                configJsonStr,
		"description":                req.Description,
		"is_active":                  req.IsActive,
		"sort":                       req.Sort,
		"updated_at":                 gtime.Now(),
	}

	_, err = g.DB().Model("hg_trading_strategy_template").Where("id", req.Id).Update(data)
	if err != nil {
		return nil, err
	}

	// 获取策略模板的 group_id，用于刷新相关引擎缓存
	var updatedTemplate *entity.TradingStrategyTemplate
	err = dao.TradingStrategyTemplate.Ctx(ctx).Where(dao.TradingStrategyTemplate.Columns().Id, req.Id).Scan(&updatedTemplate)
	if err == nil && updatedTemplate != nil && updatedTemplate.GroupId > 0 {
		// 刷新所有使用该策略组的机器人引擎的策略参数缓存
		toogo.GetRobotTaskManager().RefreshStrategyParamsByGroupId(ctx, updatedTemplate.GroupId)
	}

	res = &trading.StrategyTemplateUpdateRes{}
	return
}

// Delete 删除策略模板
func (c *cStrategyTemplate) Delete(ctx context.Context, req *trading.StrategyTemplateDeleteReq) (res *trading.StrategyTemplateDeleteRes, err error) {
	_, err = g.DB().Model("hg_trading_strategy_template").Where("id", req.Id).Delete()
	if err != nil {
		return nil, err
	}
	res = &trading.StrategyTemplateDeleteRes{}
	return
}

// Apply 应用策略到机器人（完整复制所有手动配置参数）
func (c *cStrategyTemplate) Apply(ctx context.Context, req *trading.StrategyTemplateApplyReq) (res *trading.StrategyTemplateApplyRes, err error) {
	// 获取策略模板
	var strategy *entity.TradingStrategyTemplate
	err = dao.TradingStrategyTemplate.Ctx(ctx).Where(dao.TradingStrategyTemplate.Columns().Id, req.StrategyId).Scan(&strategy)
	if err != nil {
		return nil, err
	}
	if strategy == nil {
		return nil, gerror.New("策略不存在")
	}

	// 获取策略组信息（用于获取交易平台、交易对、订单类型、保证金模式）
	var group *entity.TradingStrategyGroup
	if strategy.GroupId > 0 {
		_ = g.DB().Model("hg_trading_strategy_group").Where("id", strategy.GroupId).Scan(&group)
	}

	// 解析 config_json 获取完整手动配置
	var configJson map[string]interface{}
	if strategy.ConfigJson != "" {
		_ = json.Unmarshal([]byte(strategy.ConfigJson), &configJson)
	}

	// 构建完整的机器人配置更新（包含所有手动配置参数）
	data := g.Map{
		// ===== 风险偏好和市场状态 =====
		"risk_preference": strategy.RiskPreference,
		"market_state":    strategy.MarketState,

		// ===== 止损止盈参数（从表字段读取） =====
		"stop_loss_percent":          strategy.StopLossPercent,
		"profit_retreat_percent":     strategy.ProfitRetreatPercent,
		"auto_start_retreat_percent": strategy.AutoStartRetreatPercent,

		"updated_at": gtime.Now(),
	}

	// ===== 从 config_json 读取完整手动配置 =====
	if configJson != nil {
		// 杠杆倍数
		if leverage, ok := configJson["leverage"]; ok {
			data["leverage"] = leverage
		} else {
			data["leverage"] = strategy.Leverage
		}

		// 保证金比例
		if marginPercent, ok := configJson["marginPercent"]; ok {
			data["margin_percent"] = marginPercent
		} else {
			data["margin_percent"] = strategy.MarginPercent
		}

		// 订单类型
		if orderType, ok := configJson["orderType"].(string); ok && orderType != "" {
			data["order_type"] = orderType
		}

		// 保证金模式
		if marginMode, ok := configJson["marginMode"].(string); ok && marginMode != "" {
			data["margin_mode"] = marginMode
		}


		// 覆盖止损止盈参数（如果 config_json 中有更精确的值）
		if stopLoss, ok := configJson["stopLossPercent"]; ok {
			data["stop_loss_percent"] = stopLoss
		}
		if autoStart, ok := configJson["autoStartRetreatPercent"]; ok {
			data["auto_start_retreat_percent"] = autoStart
		}
		if profitRetreat, ok := configJson["profitRetreatPercent"]; ok {
			data["profit_retreat_percent"] = profitRetreat
		}
	} else {
		// 如果没有 config_json，使用表字段的默认值
		data["leverage"] = strategy.Leverage
		data["margin_percent"] = strategy.MarginPercent
	}

	// ===== 从策略组读取交易平台和交易对配置 =====
	if group != nil {
		if group.Exchange != "" {
			data["exchange"] = group.Exchange
		}
		if group.Symbol != "" {
			data["symbol"] = group.Symbol
		}
		if group.OrderType != "" {
			// 策略组的订单类型作为默认值，config_json 中的优先级更高
			if _, exists := data["order_type"]; !exists {
				data["order_type"] = group.OrderType
			}
		}
		if group.MarginMode != "" {
			// 策略组的保证金模式作为默认值
			if _, exists := data["margin_mode"]; !exists {
				data["margin_mode"] = group.MarginMode
			}
		}
	}

	// ===== 保存完整策略配置到 current_strategy 字段 =====
	strategyConfig := map[string]interface{}{
		"strategy_id":                strategy.Id,
		"strategy_key":               strategy.StrategyKey,
		"strategy_name":              strategy.StrategyName,
		"risk_preference":            strategy.RiskPreference,
		"market_state":               strategy.MarketState,
		"leverage":                   strategy.Leverage,
		"margin_percent":             strategy.MarginPercent,
		"monitor_window":             strategy.MonitorWindow,
		"volatility_threshold":       strategy.VolatilityThreshold,
		"stop_loss_percent":          strategy.StopLossPercent,
		"profit_retreat_percent":     strategy.ProfitRetreatPercent,
		"auto_start_retreat_percent": strategy.AutoStartRetreatPercent,
		"config":                     configJson,
	}
	if group != nil {
		strategyConfig["group_id"] = group.Id
		strategyConfig["group_name"] = group.GroupName
		strategyConfig["exchange"] = group.Exchange
		strategyConfig["symbol"] = group.Symbol
	}
	strategyBytes, _ := json.Marshal(strategyConfig)
	data["current_strategy"] = string(strategyBytes)

	// 执行更新
	_, err = g.DB().Model("hg_trading_robot").Where("id", req.RobotId).Update(data)
	if err != nil {
		return nil, err
	}

	g.Log().Infof(ctx, "[Strategy] 策略应用成功: robotId=%d, strategyId=%d, strategyName=%s, leverage=%v, marginPercent=%v",
		req.RobotId, req.StrategyId, strategy.StrategyName, data["leverage"], data["margin_percent"])

	res = &trading.StrategyTemplateApplyRes{}
	return
}
