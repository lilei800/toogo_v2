// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogo

import (
	"context"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"
)

type sToogoStrategy struct{}

func NewToogoStrategy() *sToogoStrategy {
	return &sToogoStrategy{}
}

func init() {
	service.RegisterToogoStrategy(NewToogoStrategy())
}

// TemplateList 策略模板列表
func (s *sToogoStrategy) TemplateList(ctx context.Context, in *toogoin.StrategyTemplateListInp) (list []*toogoin.StrategyTemplateListModel, totalCount int, err error) {
	mod := dao.ToogoStrategyTemplate.Ctx(ctx)
	cols := dao.ToogoStrategyTemplate.Columns()

	if in.RiskPreference != "" {
		mod = mod.Where(cols.RiskPreference, in.RiskPreference)
	}
	if in.MarketState != "" {
		mod = mod.Where(cols.MarketState, in.MarketState)
	}
	if in.IsOfficial >= 0 {
		mod = mod.Where(cols.IsOfficial, in.IsOfficial)
	}
	if in.IsActive >= 0 {
		mod = mod.Where(cols.IsActive, in.IsActive)
	}

	err = mod.OrderAsc(cols.Sort).Page(in.Page, in.PerPage).ScanAndCount(&list, &totalCount, true)
	if err != nil {
		err = gerror.Wrap(err, "获取策略模板列表失败")
	}
	return
}

// TemplateEdit 编辑策略模板
func (s *sToogoStrategy) TemplateEdit(ctx context.Context, in *toogoin.StrategyTemplateEditInp) (err error) {
	cols := dao.ToogoStrategyTemplate.Columns()

	// 解析波动率配置JSON
	var volatilityConfig *gjson.Json
	if in.VolatilityConfig != "" {
		volatilityConfig = gjson.New(in.VolatilityConfig)
	}

	data := g.Map{
		cols.StrategyKey:          in.StrategyKey,
		cols.StrategyName:         in.StrategyName,
		cols.RiskPreference:       in.RiskPreference,
		cols.MarketState:          in.MarketState,
		cols.TimeWindow:           in.TimeWindow,
		cols.VolatilityPoints:     in.VolatilityPoints,
		cols.LeverageMin:          in.LeverageMin,
		cols.LeverageMax:          in.LeverageMax,
		cols.MarginPercentMin:     in.MarginPercentMin,
		cols.MarginPercentMax:     in.MarginPercentMax,
		cols.StopLossPercent:      in.StopLossPercent,
		cols.ProfitRetreatPercent: in.ProfitRetreatPercent,
		cols.StartRetreatPercent:  in.StartRetreatPercent,
		cols.VolatilityConfig:     volatilityConfig,
		cols.Description:          in.Description,
		cols.IsOfficial:           in.IsOfficial,
		cols.IsActive:             in.IsActive,
		cols.Sort:                 in.Sort,
	}

	if in.Id > 0 {
		_, err = dao.ToogoStrategyTemplate.Ctx(ctx).Where(dao.ToogoStrategyTemplate.Columns().Id, in.Id).Data(data).Update()
	} else {
		data[cols.CreatedAt] = gtime.Now()
		_, err = dao.ToogoStrategyTemplate.Ctx(ctx).Data(data).Insert()
	}

	if err != nil {
		err = gerror.Wrap(err, "保存策略模板失败")
	}
	return
}

// TemplateDelete 删除策略模板
func (s *sToogoStrategy) TemplateDelete(ctx context.Context, in *toogoin.StrategyTemplateDeleteInp) (err error) {
	// 检查是否是官方策略
	var template *entity.ToogoStrategyTemplate
	err = dao.ToogoStrategyTemplate.Ctx(ctx).Where(dao.ToogoStrategyTemplate.Columns().Id, in.Id).Scan(&template)
	if err != nil {
		return gerror.Wrap(err, "获取策略模板失败")
	}
	if template == nil {
		return gerror.New("策略模板不存在")
	}
	if template.IsOfficial == 1 {
		return gerror.New("官方策略模板不能删除")
	}

	_, err = dao.ToogoStrategyTemplate.Ctx(ctx).Where(dao.ToogoStrategyTemplate.Columns().Id, in.Id).Delete()
	if err != nil {
		err = gerror.Wrap(err, "删除策略模板失败")
	}
	return
}

// GetByCondition 根据条件获取策略
func (s *sToogoStrategy) GetByCondition(ctx context.Context, in *toogoin.GetStrategyByConditionInp) (res *toogoin.GetStrategyByConditionModel, err error) {
	cols := dao.ToogoStrategyTemplate.Columns()

	var template *entity.ToogoStrategyTemplate
	err = dao.ToogoStrategyTemplate.Ctx(ctx).
		Where(cols.RiskPreference, in.RiskPreference).
		Where(cols.MarketState, in.MarketState).
		Where(cols.IsActive, 1).
		OrderAsc(cols.Sort).
		Scan(&template)
	if err != nil {
		return nil, gerror.Wrap(err, "获取策略模板失败")
	}
	if template == nil {
		return nil, gerror.Newf("未找到匹配的策略模板: %s-%s", in.RiskPreference, in.MarketState)
	}

	res = &toogoin.GetStrategyByConditionModel{
		ToogoStrategyTemplate: template,
	}
	return
}

// PowerConsumeList 算力消耗记录列表
func (s *sToogoStrategy) PowerConsumeList(ctx context.Context, in *toogoin.PowerConsumeListInp) (list []*toogoin.PowerConsumeListModel, totalCount int, err error) {
	mod := dao.ToogoPowerConsume.Ctx(ctx)
	cols := dao.ToogoPowerConsume.Columns()

	if in.UserId > 0 {
		mod = mod.Where(cols.UserId, in.UserId)
	}
	if in.RobotId > 0 {
		mod = mod.Where(cols.RobotId, in.RobotId)
	}
	if len(in.CreatedAt) == 2 {
		mod = mod.WhereBetween(cols.CreatedAt, in.CreatedAt[0], in.CreatedAt[1])
	}

	var consumes []*entity.ToogoPowerConsume
	err = mod.OrderDesc(cols.Id).Page(in.Page, in.PerPage).ScanAndCount(&consumes, &totalCount, true)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "获取算力消耗记录失败")
	}

	for _, consume := range consumes {
		item := &toogoin.PowerConsumeListModel{
			ToogoPowerConsume: consume,
		}

		// 获取机器人名称
		var robot *entity.TradingRobot
		dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, consume.RobotId).Scan(&robot)
		if robot != nil {
			item.RobotName = robot.RobotName
		}

		list = append(list, item)
	}
	return
}

// PowerConsumeStat 算力消耗统计
func (s *sToogoStrategy) PowerConsumeStat(ctx context.Context, in *toogoin.PowerConsumeStatInp) (res *toogoin.PowerConsumeStatModel, err error) {
	res = &toogoin.PowerConsumeStatModel{}
	cols := dao.ToogoPowerConsume.Columns()

	// 累计消耗
	total, _ := dao.ToogoPowerConsume.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Sum(cols.ConsumePower)
	res.TotalConsume = total

	// 今日消耗
	today := gtime.Now().Format("Y-m-d")
	todayTotal, _ := dao.ToogoPowerConsume.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		WhereGTE(cols.CreatedAt, today+" 00:00:00").
		Sum(cols.ConsumePower)
	res.TodayConsume = todayTotal

	// 本周消耗
	weekStart := gtime.Now().StartOfWeek().Format("Y-m-d H:i:s")
	weekTotal, _ := dao.ToogoPowerConsume.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		WhereGTE(cols.CreatedAt, weekStart).
		Sum(cols.ConsumePower)
	res.WeekConsume = weekTotal

	// 本月消耗
	monthStart := gtime.Now().StartOfMonth().Format("Y-m-d H:i:s")
	monthTotal, _ := dao.ToogoPowerConsume.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		WhereGTE(cols.CreatedAt, monthStart).
		Sum(cols.ConsumePower)
	res.MonthConsume = monthTotal

	// 累计盈利
	profitTotal, _ := dao.ToogoPowerConsume.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Sum(cols.ProfitAmount)
	res.TotalProfit = profitTotal

	// 平均消耗比例
	count, _ := dao.ToogoPowerConsume.Ctx(ctx).
		Where(cols.UserId, in.UserId).
		Count()
	if count > 0 {
		avgRate, _ := dao.ToogoPowerConsume.Ctx(ctx).
			Where(cols.UserId, in.UserId).
			Avg(cols.ConsumeRate)
		res.AvgConsumeRate = avgRate
	}

	return
}

// AnalyzeMarketState 分析市场状态 (多周期综合分析)
func (s *sToogoStrategy) AnalyzeMarketState(ctx context.Context, symbol string, strategy *entity.ToogoStrategyTemplate) (marketState string, riskPreference string, err error) {
	// 解析波动率配置
	if strategy.VolatilityConfig == nil {
		// 使用默认配置
		marketState = "trend"
		riskPreference = "balanced"
		return
	}

	// 这里简化处理，实际需要接入交易所API获取实时行情数据
	// 根据多周期数据综合判断市场状态

	// 默认返回趋势市场和平衡型
	marketState = "trend"
	riskPreference = "balanced"

	g.Log().Debugf(ctx, "分析市场状态: symbol=%s, marketState=%s, riskPreference=%s", symbol, marketState, riskPreference)
	return
}

// GetOptimalStrategy 获取最优策略 (根据实时行情自动选择)
func (s *sToogoStrategy) GetOptimalStrategy(ctx context.Context, symbol string) (*entity.ToogoStrategyTemplate, error) {
	// 分析市场状态
	// 这里简化处理，实际应该调用AnalyzeMarketState
	marketState := "trend"
	riskPreference := "balanced"

	// 获取对应策略
	result, err := s.GetByCondition(ctx, &toogoin.GetStrategyByConditionInp{
		RiskPreference: riskPreference,
		MarketState:    marketState,
	})
	if err != nil {
		return nil, err
	}

	return result.ToogoStrategyTemplate, nil
}

