package toogo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"hotgo/internal/consts"
	"hotgo/internal/library/contexts"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

	// StrategyGroupService 策略模板服务
type StrategyGroupService struct{}

const (
	// StrategyGroupOrderTypeMarket: direct market order (legacy/default in older versions)
	StrategyGroupOrderTypeMarket = "market"
	// StrategyGroupOrderTypeLimitThenMarket: place a loose limit order first; if not filled quickly, cancel and fallback to market order.
	StrategyGroupOrderTypeLimitThenMarket = "limit_then_market"
)

func normalizeStrategyGroupOrderType(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "":
		// new default
		return StrategyGroupOrderTypeLimitThenMarket
	case "market", "市价", "市價":
		return StrategyGroupOrderTypeMarket
	case "limit_then_market", "先限价再市价", "先限價再市價", "限价再市价", "限價再市價":
		return StrategyGroupOrderTypeLimitThenMarket
	default:
		// keep safe legacy behavior for unknown values
		return StrategyGroupOrderTypeMarket
	}
}

// NewStrategyGroupService 创建策略模板服务实例
func NewStrategyGroupService() *StrategyGroupService {
	return &StrategyGroupService{}
}

// List 获取策略模板列表
func (s *StrategyGroupService) List(ctx context.Context, in *toogoin.StrategyGroupListInp) (*toogoin.StrategyGroupListModel, error) {
	m := g.DB().Model("hg_trading_strategy_group").Safe()

	if in.Exchange != "" {
		m = m.Where("exchange", in.Exchange)
	}
	if in.Symbol != "" {
		m = m.WhereLike("symbol", "%"+in.Symbol+"%")
	}
	if in.IsActive != nil {
		m = m.Where("is_active", *in.IsActive)
	}
	
	// NonPersonal=1: only system/public groups (user_id=0 or NULL)
	if in.NonPersonal != nil && *in.NonPersonal == 1 {
		m = m.Where("(user_id = 0 OR user_id IS NULL)")
	}
	// 支持筛选官方/非官方策略：
	// - 官方策略（is_official=1）为公用资源，不限制 user_id
	// - 我的策略（is_official=0）为用户私有资源，必须限制 user_id=当前登录用户（超级管理员不限制）
	userId := contexts.GetUserId(ctx)
	roleKey := contexts.GetRoleKey(ctx)
	isSuper := roleKey == consts.SuperRoleKey
	if in.IsOfficial != nil {
		m = m.Where("is_official", *in.IsOfficial)
		if (in.NonPersonal == nil || *in.NonPersonal != 1) && !isSuper && *in.IsOfficial == 0 {
			m = m.Where("user_id", userId)
		}
	} else if !isSuper {
		// Default for non-super: official + my groups (skip when NonPersonal=1)
		if in.NonPersonal == nil || *in.NonPersonal != 1 {
			m = m.Where("(is_official = 1 OR (is_official = 0 AND user_id = ?))", userId)
		}
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var groups []*entity.TradingStrategyGroup
	// 官方策略优先排在前面
	err = m.Page(in.Page, in.PageSize).Order("is_official DESC, sort ASC, id ASC").Scan(&groups)
	if err != nil {
		return nil, err
	}

	// 获取每个模板下的策略数量
	var list []*toogoin.StrategyGroupItem
	for _, group := range groups {
		count, _ := g.DB().Model("hg_trading_strategy_template").Where("group_id", group.Id).Count()
		list = append(list, &toogoin.StrategyGroupItem{
			TradingStrategyGroup: *group,
			StrategyCount:        count,
		})
	}

	return &toogoin.StrategyGroupListModel{
		List:  list,
		Page:  in.Page,
		Total: total,
	}, nil
}

// Create 创建策略模板
func (s *StrategyGroupService) Create(ctx context.Context, in *toogoin.StrategyGroupCreateInp) error {
	// 检查key是否重复
	count, err := g.DB().Model("hg_trading_strategy_group").Where("group_key", in.GroupKey).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("模板标识已存在")
	}

	data := g.Map{
		"group_name":  in.GroupName,
		"group_key":   in.GroupKey,
		"exchange":    in.Exchange,
		"symbol":      in.Symbol,
		"order_type":  normalizeStrategyGroupOrderType(in.OrderType),
		"margin_mode": in.MarginMode,
		"description": in.Description,
		"is_official": 0,
		"user_id":     contexts.GetUserId(ctx),
		"is_active":   1,
		"sort":        in.Sort,
		"created_at":  gtime.Now(),
		"updated_at":  gtime.Now(),
	}

	// admin overrides (super only)
	if contexts.GetRoleKey(ctx) == consts.SuperRoleKey {
		if in.IsOfficial != nil {
			data["is_official"] = *in.IsOfficial
			if *in.IsOfficial == 1 {
				data["user_id"] = int64(0)
			}
		}
		if in.UserId != nil {
			data["user_id"] = *in.UserId
		}
		if in.IsActive != nil {
			data["is_active"] = *in.IsActive
		}
		if in.IsVisible != nil {
			data["is_visible"] = *in.IsVisible
		}

	}

	// default is handled by normalizeStrategyGroupOrderType
	if data["margin_mode"] == "" {
		data["margin_mode"] = "isolated"
	}

	_, err = g.DB().Model("hg_trading_strategy_group").Insert(data)
	return err
}

// Update 更新策略模板
func (s *StrategyGroupService) Update(ctx context.Context, in *toogoin.StrategyGroupUpdateInp) error {
	// 检查模板是否存在
	var group entity.TradingStrategyGroup
	err := g.DB().Model("hg_trading_strategy_group").Where("id", in.Id).Scan(&group)
	if err != nil {
		return err
	}
	if group.Id == 0 {
		return gerror.New("模板不存在")
	}
	// 【管理员权限】允许修改官方策略组（在官方策略组管理页面中）

	// 检查是否有机器人绑定（如果未确认，则返回绑定信息）
	if !in.Confirmed {
		robotCount, err := g.DB().Model("hg_trading_robot").
			Where("strategy_group_id", in.Id).
			WhereNull("deleted_at").
			Count()
		if err != nil {
			return gerror.Wrap(err, "检查机器人绑定失败")
		}
		if robotCount > 0 {
			// 获取绑定的机器人列表（最多10个，用于提示）
			var robots []map[string]interface{}
			g.DB().Model("hg_trading_robot").
				Fields("id", "robot_name", "status").
				Where("strategy_group_id", in.Id).
				WhereNull("deleted_at").
				Limit(10).
				Scan(&robots)

			robotNames := make([]string, 0, len(robots))
			for _, robot := range robots {
				if name, ok := robot["robot_name"].(string); ok {
					robotNames = append(robotNames, name)
				}
			}

			// 返回需要确认的错误信息
			msg := fmt.Sprintf("该策略组已被%d个机器人绑定", robotCount)
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
			msg += "，修改策略组会影响这些机器人的运行。请确认是否继续修改？"
			return gerror.New(msg)
		}
	}

	data := g.Map{
		"group_name":  in.GroupName,
		"exchange":    in.Exchange,
		"symbol":      in.Symbol,
		"order_type":  normalizeStrategyGroupOrderType(in.OrderType),
		"margin_mode": in.MarginMode,
		"description": in.Description,
		"sort":        in.Sort,
		"updated_at":  gtime.Now(),
	}
	// 如果提供了 is_visible，则更新
	if in.IsVisible != nil {
		data["is_visible"] = *in.IsVisible
	}

	// admin-only: allow toggling is_official / is_active
	if in.IsOfficial != nil || in.IsActive != nil {
		if contexts.GetRoleKey(ctx) != consts.SuperRoleKey {
			return gerror.New("permission denied")
		}
		if in.IsOfficial != nil {
			data["is_official"] = *in.IsOfficial
			if *in.IsOfficial == 1 {
				data["user_id"] = int64(0)
			}
		}
		if in.IsActive != nil {
			data["is_active"] = *in.IsActive
		}
	}

	_, err = g.DB().Model("hg_trading_strategy_group").Where("id", in.Id).Update(data)
	if err != nil {
		return err
	}

	// 刷新所有使用该策略组的机器人引擎的策略参数缓存
	GetRobotTaskManager().RefreshStrategyParamsByGroupId(ctx, in.Id)

	return nil
}

// Delete 删除策略模板
func (s *StrategyGroupService) Delete(ctx context.Context, in *toogoin.StrategyGroupDeleteInp) error {
	// 检查模板是否存在
	var group entity.TradingStrategyGroup
	err := g.DB().Model("hg_trading_strategy_group").Where("id", in.Id).Scan(&group)
	if err != nil {
		return err
	}
	if group.Id == 0 {
		return gerror.New("模板不存在")
	}
	// 【管理员权限】允许删除官方策略组（在官方策略组管理页面中）

	// 检查是否有机器人绑定
	robotCount, err := g.DB().Model("hg_trading_robot").
		Where("strategy_group_id", in.Id).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return gerror.Wrap(err, "检查机器人绑定失败")
	}
	if robotCount > 0 {
		// 获取绑定的机器人列表（最多5个，用于提示）
		var robots []map[string]interface{}
		g.DB().Model("hg_trading_robot").
			Fields("id", "robot_name", "status").
			Where("strategy_group_id", in.Id).
			WhereNull("deleted_at").
			Limit(5).
			Scan(&robots)

		robotNames := make([]string, 0, len(robots))
		for _, robot := range robots {
			if name, ok := robot["robot_name"].(string); ok {
				robotNames = append(robotNames, name)
			}
		}

		if len(robotNames) > 0 {
			msg := "该策略组已被以下机器人绑定，无法删除："
			if len(robotNames) >= 5 {
				msg += fmt.Sprintf("%s等%d个机器人", robotNames[0], robotCount)
			} else {
				msg += fmt.Sprintf("%s（共%d个）", robotNames[0], robotCount)
			}
			return gerror.New(msg)
		}
		return gerror.Newf("该策略组已被%d个机器人绑定，无法删除", robotCount)
	}

	// 删除关联的策略
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model("hg_trading_strategy_template").Where("group_id", in.Id).Delete()
		if err != nil {
			return err
		}
		_, err = tx.Model("hg_trading_strategy_group").Where("id", in.Id).Delete()
		return err
	})
	return err
}

// Init 初始化12种策略
func (s *StrategyGroupService) Init(ctx context.Context, in *toogoin.StrategyGroupInitInp) error {
	// 获取模板信息
	var group entity.TradingStrategyGroup
	err := g.DB().Model("hg_trading_strategy_group").Where("id", in.GroupId).Scan(&group)
	if err != nil {
		return err
	}
	if group.Id == 0 {
		return gerror.New("模板不存在")
	}

	// 4种市场状态
	marketStates := []struct {
		Key    string
		Name   string
		Params map[string]interface{}
	}{
		{"trend", "趋势市场", map[string]interface{}{"monitorWindow": 300, "volatilityThreshold": 10}},
		{"volatile", "震荡市场", map[string]interface{}{"monitorWindow": 120, "volatilityThreshold": 5}},
		{"high_vol", "高波动", map[string]interface{}{"monitorWindow": 60, "volatilityThreshold": 20}},
		{"low_vol", "低波动", map[string]interface{}{"monitorWindow": 300, "volatilityThreshold": 2}},
	}

	// 3种风险偏好 × 4种市场状态 = 12种策略参数
	// 每种组合都有固定的杠杆和保证金比例
	strategyParams := map[string]map[string]map[string]interface{}{
		"conservative": {
			"trend":    {"leverage": 3, "marginPercent": 8.0},
			"volatile": {"leverage": 2, "marginPercent": 5.0},
			"high_vol": {"leverage": 2, "marginPercent": 5.0},
			"low_vol":  {"leverage": 5, "marginPercent": 10.0},
		},
		"balanced": {
			"trend":    {"leverage": 5, "marginPercent": 12.0},
			"volatile": {"leverage": 4, "marginPercent": 10.0},
			"high_vol": {"leverage": 5, "marginPercent": 12.0},
			"low_vol":  {"leverage": 8, "marginPercent": 15.0},
		},
		"aggressive": {
			"trend":    {"leverage": 10, "marginPercent": 20.0},
			"volatile": {"leverage": 8, "marginPercent": 15.0},
			"high_vol": {"leverage": 10, "marginPercent": 20.0},
			"low_vol":  {"leverage": 15, "marginPercent": 25.0},
		},
	}

	// 3种风险偏好的通用参数
	riskPreferences := []struct {
		Key    string
		Name   string
		Params map[string]interface{}
	}{
		{"conservative", "保守型", map[string]interface{}{
			"stopLossPercent": 3, "autoStartRetreatPercent": 2, "profitRetreatPercent": 30,
		}},
		{"balanced", "平衡型", map[string]interface{}{
			"stopLossPercent": 5, "autoStartRetreatPercent": 3, "profitRetreatPercent": 25,
		}},
		{"aggressive", "激进型", map[string]interface{}{
			"stopLossPercent": 8, "autoStartRetreatPercent": 5, "profitRetreatPercent": 20,
		}},
	}

	sort := 100
	for _, market := range marketStates {
		for _, risk := range riskPreferences {
			strategyKey := fmt.Sprintf("%d_%s_%s", group.Id, market.Key, risk.Key)

			// 检查是否已存在
			count, _ := g.DB().Model("hg_trading_strategy_template").Where("strategy_key", strategyKey).Count()
			if count > 0 {
				continue
			}

			// 合并参数
			configJson := map[string]interface{}{
				"exchange":   group.Exchange,
				"symbol":     group.Symbol,
				"orderType":  group.OrderType,
				"marginMode": group.MarginMode,
			}
			configJsonBytes, _ := json.Marshal(configJson)

			// 获取该组合的固定参数
			params := strategyParams[risk.Key][market.Key]

			data := g.Map{
				"group_id":                   group.Id,
				"strategy_key":               strategyKey,
				"strategy_name":              fmt.Sprintf("%s-%s", market.Name, risk.Name),
				"risk_preference":            risk.Key,
				"market_state":               market.Key,
				"monitor_window":             market.Params["monitorWindow"],
				"volatility_threshold":       market.Params["volatilityThreshold"],
				"leverage":                   params["leverage"],
				"margin_percent":             params["marginPercent"],
				"stop_loss_percent":          risk.Params["stopLossPercent"],
				"auto_start_retreat_percent": risk.Params["autoStartRetreatPercent"],
				"profit_retreat_percent":     risk.Params["profitRetreatPercent"],
				"config_json":                string(configJsonBytes),
				"description":                fmt.Sprintf("%s %s-%s策略", group.Symbol, market.Name, risk.Name),
				"is_active":                  1,
				"sort":                       sort,
				"created_at":                 gtime.Now(),
				"updated_at":                 gtime.Now(),
			}

			_, err := g.DB().Model("hg_trading_strategy_template").Insert(data)
			if err != nil {
				g.Log().Error(ctx, "初始化策略失败", err, data)
			}
			sort++
		}
	}

	return nil
}

// CopyFromOfficial 从官方策略复制到我的策略
// 返回复制后的策略组ID（如果已存在则返回已存在的ID）
func (s *StrategyGroupService) CopyFromOfficial(ctx context.Context, officialGroupId int64) (int64, error) {
	userId := contexts.GetUserId(ctx)
	// 获取官方模板
	var officialGroup entity.TradingStrategyGroup
	err := g.DB().Model("hg_trading_strategy_group").Ctx(ctx).Where("id", officialGroupId).Scan(&officialGroup)
	if err != nil {
		return 0, gerror.Wrap(err, "查询官方模板失败")
	}
	if officialGroup.Id == 0 {
		return 0, gerror.New("官方模板不存在")
	}
	if officialGroup.IsOfficial != 1 {
		return 0, gerror.New("只能复制官方模板")
	}

	// 检查是否已经存在从该官方模板复制的版本
	count, err := g.DB().Model("hg_trading_strategy_group").Ctx(ctx).
		Where("from_official_id", officialGroupId).
		Where("is_official", 0).
		Where("user_id", userId).
		Count()
	if err != nil {
		g.Log().Errorf(ctx, "[CopyFromOfficial] 检查已添加状态失败: %v, officialGroupId=%d", err, officialGroupId)
		return 0, gerror.Wrap(err, "检查已添加状态失败")
	}
	if count > 0 {
		// 已存在，查询并返回已存在的策略组ID
		var existingGroup entity.TradingStrategyGroup
		err = g.DB().Model("hg_trading_strategy_group").Ctx(ctx).
			Where("from_official_id", officialGroupId).
			Where("is_official", 0).
			Where("user_id", userId).
			Scan(&existingGroup)
		if err != nil {
			g.Log().Errorf(ctx, "[CopyFromOfficial] 查询已存在策略组失败: %v, officialGroupId=%d", err, officialGroupId)
			return 0, gerror.Wrap(err, "查询已存在策略组失败")
		}
		if existingGroup.Id > 0 {
			g.Log().Infof(ctx, "官方策略模板 %d 已添加到我的策略（ID: %d），跳过重复添加", officialGroupId, existingGroup.Id)
			// ✅ 自动初始化：补齐缺失的 12 种策略模板（只补缺，不覆盖已有）
			if err := s.Init(ctx, &toogoin.StrategyGroupInitInp{
				GroupId:    existingGroup.Id,
				UseDefault: true,
			}); err != nil {
				return 0, gerror.Wrap(err, "自动初始化策略失败")
			}
			return existingGroup.Id, nil
		}
	}

	// 创建新的策略组
	// 使用 UnixNano 并做一次唯一性检查，避免并发点击/同秒冲突
	newGroupKey := ""
	for i := 0; i < 5; i++ {
		candidate := fmt.Sprintf("copy_%d_%d", officialGroupId, gtime.Now().UnixNano())
		c, e := g.DB().Model("hg_trading_strategy_group").Ctx(ctx).Where("group_key", candidate).Count()
		if e != nil {
			return 0, gerror.Wrap(e, "生成策略组标识失败")
		}
		if c == 0 {
			newGroupKey = candidate
			break
		}
	}
	if newGroupKey == "" {
		return 0, gerror.New("生成策略组标识失败")
	}
	newGroupData := g.Map{
		"group_name":       officialGroup.GroupName + " (我的副本)",
		"group_key":        newGroupKey,
		"exchange":         officialGroup.Exchange,
		"symbol":           officialGroup.Symbol,
		"order_type":       officialGroup.OrderType,
		"margin_mode":      officialGroup.MarginMode,
		"description":      officialGroup.Description,
		"is_official":      0,
		"from_official_id": officialGroup.Id, // 记录来源
		"user_id":          userId,
		"is_active":        1,
		"sort":             100,
		"created_at":       gtime.Now(),
		"updated_at":       gtime.Now(),
	}

	var newGroupId int64
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 【关键修复】不使用 InsertAndGetId/LastInsertId，避免 PostgreSQL 驱动不支持导致“第一次点击报错”
		if _, e := tx.Model("hg_trading_strategy_group").Data(newGroupData).Insert(); e != nil {
			return e
		}
		v, e := tx.Model("hg_trading_strategy_group").Where("group_key", newGroupKey).Value("id")
		if e != nil {
			return e
		}
		newGroupId = v.Int64()
		if newGroupId == 0 {
			return gerror.New("创建策略组失败")
		}

		// 复制策略（同事务，避免只插入组但没插入模板）
		var strategies []*entity.TradingStrategyTemplate
		if e := tx.Model("hg_trading_strategy_template").Where("group_id", officialGroupId).Scan(&strategies); e != nil {
			return gerror.Wrap(e, "查询官方策略列表失败")
		}
		if len(strategies) == 0 {
			return gerror.Newf("官方模板 %d 下没有策略", officialGroupId)
		}

		now := gtime.Now()
		for _, strategy := range strategies {
			if strategy == nil {
				continue
			}
			newStrategyKey := fmt.Sprintf("%d_%s_%s", newGroupId, strategy.MarketState, strategy.RiskPreference)
			newStrategyData := g.Map{
				"group_id":                   newGroupId,
				"strategy_key":               newStrategyKey,
				"strategy_name":              strategy.StrategyName,
				"risk_preference":            strategy.RiskPreference,
				"market_state":               strategy.MarketState,
				"monitor_window":             strategy.MonitorWindow,
				"volatility_threshold":       strategy.VolatilityThreshold,
				"leverage":                   strategy.Leverage,
				"margin_percent":             strategy.MarginPercent,
				"stop_loss_percent":          strategy.StopLossPercent,
				"auto_start_retreat_percent": strategy.AutoStartRetreatPercent,
				"profit_retreat_percent":     strategy.ProfitRetreatPercent,
				"config_json":                strategy.ConfigJson,
				"description":                strategy.Description,
				"is_active":                  1,
				"sort":                       strategy.Sort,
				"created_at":                 now,
				"updated_at":                 now,
			}
			if _, e := tx.Model("hg_trading_strategy_template").Data(newStrategyData).Insert(); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	// ✅ 自动初始化：补齐缺失的 12 种策略模板（只补缺，不覆盖已有）
	if err := s.Init(ctx, &toogoin.StrategyGroupInitInp{
		GroupId:    newGroupId,
		UseDefault: true,
	}); err != nil {
		return 0, gerror.Wrap(err, "自动初始化策略失败")
	}

	return newGroupId, nil
}

// Clone 复制策略组（含策略模板）
// - 仅超级管理员可用
// - 会复制策略组记录，并复制其下所有策略模板
// - 自动生成新的 group_key，避免唯一键冲突
func (s *StrategyGroupService) Clone(ctx context.Context, groupId int64) (int64, error) {
	if contexts.GetRoleKey(ctx) != consts.SuperRoleKey {
		return 0, gerror.New("permission denied")
	}
	if groupId <= 0 {
		return 0, gerror.New("groupId invalid")
	}

	// 读取源策略组
	var src entity.TradingStrategyGroup
	if err := g.DB().Model("hg_trading_strategy_group").Ctx(ctx).
		Where("id", groupId).
		Scan(&src); err != nil {
		return 0, err
	}
	if src.Id == 0 {
		return 0, gerror.New("模板不存在")
	}

	// 生成新的 group_key（确保唯一）
	newKey := ""
	for i := 0; i < 5; i++ {
		newKey = fmt.Sprintf("%s_copy_%d", src.GroupKey, gtime.Now().UnixNano())
		cnt, err := g.DB().Model("hg_trading_strategy_group").Ctx(ctx).Where("group_key", newKey).Count()
		if err != nil {
			return 0, err
		}
		if cnt == 0 {
			break
		}
	}
	if newKey == "" {
		return 0, gerror.New("生成模板标识失败")
	}

	var newGroupId int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 复制策略组
		data := g.Map{
			"group_name":  fmt.Sprintf("%s (复制)", src.GroupName),
			"group_key":   newKey,
			"exchange":    src.Exchange,
			"symbol":      src.Symbol,
			"order_type":  src.OrderType,
			"margin_mode": src.MarginMode,
			"description": src.Description,
			"is_official": src.IsOfficial,
			"user_id":     src.UserId,
			"is_active":   src.IsActive,
			"sort":        src.Sort,
			"created_at":  gtime.Now(),
			"updated_at":  gtime.Now(),
		}

		// PostgreSQL 兼容：InsertAndGetId 在某些驱动实现下仍可能触发 LastInsertId 错误
		// 兜底策略：InsertAndGetId 失败且包含 LastInsertId → 普通 Insert + 通过唯一 group_key 反查 id
		id, err := tx.Model("hg_trading_strategy_group").Data(data).InsertAndGetId()
		if err != nil && strings.Contains(err.Error(), "LastInsertId is not supported") {
			if _, e := tx.Model("hg_trading_strategy_group").Data(data).Insert(); e != nil {
				return e
			}
			v, e := tx.Model("hg_trading_strategy_group").Where("group_key", newKey).Value("id")
			if e != nil {
				return e
			}
			id = v.Int64()
			err = nil
		}
		if err != nil {
			return err
		}
		newGroupId = id
		if newGroupId == 0 {
			return gerror.New("创建策略组失败")
		}

		// 复制策略模板
		var templates []*entity.TradingStrategyTemplate
		if err := tx.Model("hg_trading_strategy_template").Where("group_id", src.Id).Scan(&templates); err != nil {
			return err
		}
		// 仅复制“标准12套”（4种市场状态×3种风险偏好），避免源数据存在非标准组合导致复制后模板数量异常。
		allowedMarket := map[string]bool{"trend": true, "volatile": true, "high_vol": true, "low_vol": true}
		allowedRisk := map[string]bool{"conservative": true, "balanced": true, "aggressive": true}
		seen := map[string]bool{}
		for _, t := range templates {
			if t == nil {
				continue
			}
			ms := strings.TrimSpace(t.MarketState)
			rp := strings.TrimSpace(t.RiskPreference)
			if !allowedMarket[ms] || !allowedRisk[rp] {
				continue
			}
			k := ms + "|" + rp
			if seen[k] {
				continue
			}
			seen[k] = true

			newStrategyKey := fmt.Sprintf("%d_%s_%s", newGroupId, t.MarketState, t.RiskPreference)
			row := g.Map{
				"group_id":                   newGroupId,
				"strategy_key":               newStrategyKey,
				"strategy_name":              t.StrategyName,
				"risk_preference":            t.RiskPreference,
				"market_state":               t.MarketState,
				"monitor_window":             t.MonitorWindow,
				"volatility_threshold":       t.VolatilityThreshold,
				"leverage":                   t.Leverage,
				"margin_percent":             t.MarginPercent,
				"stop_loss_percent":          t.StopLossPercent,
				"auto_start_retreat_percent": t.AutoStartRetreatPercent,
				"profit_retreat_percent":     t.ProfitRetreatPercent,
				"config_json":                t.ConfigJson,
				"description":                t.Description,
				"is_active":                  t.IsActive,
				"sort":                       t.Sort,
				"created_at":                 gtime.Now(),
				"updated_at":                 gtime.Now(),
			}
			if _, err := tx.Model("hg_trading_strategy_template").Insert(row); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	// ✅ 保证 12 套模板：若复制后不足 12，则仅补齐缺失组合（不覆盖已复制的模板）
	// 注意：这里补齐的是“标准12套”，若源数据只有非标准组合，我们会复制不到任何模板，此时用默认值补齐。
	cnt, _ := g.DB().Model("hg_trading_strategy_template").Ctx(ctx).Where("group_id", newGroupId).Count()
	if cnt < 12 {
		if err := s.Init(ctx, &toogoin.StrategyGroupInitInp{
			GroupId:    newGroupId,
			UseDefault: true,
		}); err != nil {
			return 0, gerror.Wrap(err, "自动补齐12套策略模板失败")
		}
	}
	return newGroupId, nil
}

// SetDefault 设置默认策略模板
func (s *StrategyGroupService) SetDefault(ctx context.Context, groupId int64) error {
	// 检查模板是否存在
	var group entity.TradingStrategyGroup
	err := g.DB().Model("hg_trading_strategy_group").Where("id", groupId).Scan(&group)
	if err != nil {
		return err
	}
	if group.Id == 0 {
		return gerror.New("模板不存在")
	}
	if group.IsOfficial == 1 {
		return gerror.New("官方模板不能设为默认，请先添加到我的策略")
	}

	// 清除其他默认标记
	_, err = g.DB().Model("hg_trading_strategy_group").
		Where("is_official", 0).
		Where("is_default", 1).
		Update(g.Map{"is_default": 0, "updated_at": gtime.Now()})
	if err != nil {
		return err
	}

	// 设置当前为默认
	_, err = g.DB().Model("hg_trading_strategy_group").
		Where("id", groupId).
		Update(g.Map{"is_default": 1, "updated_at": gtime.Now()})

	return err
}
