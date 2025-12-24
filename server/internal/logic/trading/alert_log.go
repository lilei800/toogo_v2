// Package trading 预警日志服务实现
package trading

import (
	"context"

	"hotgo/api/admin/trading"
	"hotgo/internal/dao"
	"hotgo/internal/library/market"
	"hotgo/internal/logic/toogo"
	"hotgo/internal/model/entity"
	"hotgo/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
)

type sAlertLog struct{}

func init() {
	service.RegisterAlertLog(NewAlertLog())
}

func NewAlertLog() *sAlertLog {
	return &sAlertLog{}
}

// List 预警日志列表
func (s *sAlertLog) List(ctx context.Context, in *trading.AlertLogListReq) (*trading.AlertLogListRes, error) {
	switch in.Type {
	case "market_state":
		return s.listMarketStateLogs(ctx, in)
	case "risk_preference":
		return s.listRiskPreferenceLogs(ctx, in)
	case "direction":
		return s.listDirectionLogs(ctx, in)
	default:
		return nil, gerror.New("无效的日志类型")
	}
}

func (s *sAlertLog) listMarketStateLogs(ctx context.Context, in *trading.AlertLogListReq) (*trading.AlertLogListRes, error) {
	var list []*entity.TradingMarketStateLog

	model := dao.TradingMarketStateLog.Ctx(ctx)
	if in.Platform != "" {
		model = model.Where("platform", in.Platform)
	}
	if in.Symbol != "" {
		model = model.Where("symbol", in.Symbol)
	}

	count, err := model.Count()
	if err != nil {
		return nil, err
	}

	err = model.Page(in.Page, in.PerPage).OrderDesc("id").Scan(&list)
	if err != nil {
		return nil, err
	}

	return &trading.AlertLogListRes{
		List:  list,
		Page:  in.Page,
		Total: count,
	}, nil
}

func (s *sAlertLog) listRiskPreferenceLogs(ctx context.Context, in *trading.AlertLogListReq) (*trading.AlertLogListRes, error) {
	var list []*entity.TradingRiskPreferenceLog

	model := dao.TradingRiskPreferenceLog.Ctx(ctx)
	if in.RobotId > 0 {
		model = model.Where("robot_id", in.RobotId)
	}
	if in.Platform != "" {
		model = model.Where("platform", in.Platform)
	}
	if in.Symbol != "" {
		model = model.Where("symbol", in.Symbol)
	}

	count, err := model.Count()
	if err != nil {
		return nil, err
	}

	err = model.Page(in.Page, in.PerPage).OrderDesc("id").Scan(&list)
	if err != nil {
		return nil, err
	}

	return &trading.AlertLogListRes{
		List:  list,
		Page:  in.Page,
		Total: count,
	}, nil
}

func (s *sAlertLog) listDirectionLogs(ctx context.Context, in *trading.AlertLogListReq) (*trading.AlertLogListRes, error) {
	var list []*entity.TradingDirectionLog

	model := dao.TradingDirectionLog.Ctx(ctx)
	if in.Platform != "" {
		model = model.Where("platform", in.Platform)
	}
	if in.Symbol != "" {
		model = model.Where("symbol", in.Symbol)
	}

	count, err := model.Count()
	if err != nil {
		return nil, err
	}

	err = model.Page(in.Page, in.PerPage).OrderDesc("id").Scan(&list)
	if err != nil {
		return nil, err
	}

	return &trading.AlertLogListRes{
		List:  list,
		Page:  in.Page,
		Total: count,
	}, nil
}

// MarketAnalysis 获取市场分析数据
func (s *sAlertLog) MarketAnalysis(ctx context.Context, in *trading.MarketAnalysisReq) (*trading.MarketAnalysisRes, error) {
	analysis := market.GetMarketAnalyzer().GetAnalysis(in.Platform, in.Symbol)
	if analysis == nil {
		return nil, gerror.New("暂无市场分析数据，请确认交易对已被订阅")
	}

	res := &trading.MarketAnalysisRes{
		Platform:        analysis.Platform,
		Symbol:          analysis.Symbol,
		CurrentPrice:    analysis.CurrentPrice,
		MarketState:     string(analysis.MarketState),
		MarketStateConf: analysis.MarketStateConf,
		TrendStrength:   analysis.TrendStrength,
		Volatility:      analysis.Volatility,
		SupportLevel:    analysis.SupportLevel,
		ResistanceLevel: analysis.ResistanceLevel,
		TimeframeData:   make(map[string]*trading.TimeframeInfo),
	}

	if analysis.Indicators != nil {
		res.Indicators = &trading.TechnicalIndicators{
			TrendScore:      analysis.Indicators.TrendScore,
			VolatilityScore: analysis.Indicators.VolatilityScore,
		}
	}

	for interval, tf := range analysis.TimeframeAnalysis {
		res.TimeframeData[interval] = &trading.TimeframeInfo{
			Interval:      tf.Interval,
			Trend:         tf.Trend,
			TrendStrength: tf.TrendStrength,
			MACD:          tf.MACD,
			EMA12:         tf.EMA12,
			EMA26:         tf.EMA26,
		}
	}

	return res, nil
}

// DirectionSignal 获取方向信号
func (s *sAlertLog) DirectionSignal(ctx context.Context, in *trading.DirectionSignalReq) (*trading.DirectionSignalRes, error) {
	signal := market.GetDirectionSignalService().GetSignal(in.Platform, in.Symbol)
	if signal == nil {
		return nil, gerror.New("暂无方向信号数据")
	}

	return &trading.DirectionSignalRes{
		Platform:   signal.Platform,
		Symbol:     signal.Symbol,
		Direction:  string(signal.Direction),
		Strength:   signal.Strength,
		Confidence: signal.Confidence,
		Action:     string(signal.Action),
		EntryPrice: signal.EntryPrice,
		StopLoss:   signal.StopLoss,
		Reason:     signal.Reason,
	}, nil
}

// RobotRiskEval 获取机器人风险评估
func (s *sAlertLog) RobotRiskEval(ctx context.Context, in *trading.RobotRiskEvalReq) (*trading.RobotRiskEvalRes, error) {
	eval := market.GetRiskEvaluator().GetEvaluation(in.RobotId)
	if eval == nil {
		return nil, gerror.New("暂无风险评估数据，机器人可能未运行")
	}

	return &trading.RobotRiskEvalRes{
		RobotId:                eval.RobotId,
		WinProbability:         eval.WinProbability,
		RiskPreference:         string(eval.RiskPreference),
		MarketScore:            eval.MarketScore,
		TechnicalScore:         eval.TechnicalScore,
		AccountScore:           eval.AccountScore,
		HistoryScore:           eval.HistoryScore,
		VolatilityRisk:         eval.VolatilityRisk,
		SuggestedLeverage:      eval.SuggestedLeverage,
		SuggestedMarginPercent: eval.SuggestedMarginPercent,
		Reason:                 eval.Reason,
	}, nil
}

// RobotStatus 获取机器人实时状态
func (s *sAlertLog) RobotStatus(ctx context.Context, in *trading.RobotStatusReq) (*trading.RobotStatusRes, error) {
	managed := toogo.GetRobotTaskManager().GetManagedRobot(in.RobotId)
	if managed == nil {
		return nil, gerror.New("机器人未运行")
	}

	res := &trading.RobotStatusRes{
		RobotId: in.RobotId,
		Symbol:  managed.Robot.Symbol,
		Status:  managed.Robot.Status,
	}

	// 获取当前价格
	ticker := market.GetMarketServiceManager().GetTicker(managed.Platform, managed.Robot.Symbol)
	if ticker != nil {
		res.CurrentPrice = ticker.LastPrice
	}

	// 账户余额
	if managed.AccountBalance != nil {
		res.AccountBalance = managed.AccountBalance.TotalBalance
		res.AvailableBalance = managed.AccountBalance.AvailableBalance
	}

	// 持仓信息
	for _, pos := range managed.CurrentPositions {
		if pos.PositionAmt != 0 {
			res.Positions = append(res.Positions, &trading.PositionInfo{
				PositionSide:   pos.PositionSide,
				PositionAmt:    pos.PositionAmt,
				EntryPrice:     pos.EntryPrice,
				MarkPrice:      pos.MarkPrice,
				UnrealizedPnl:  pos.UnrealizedPnl,
				Leverage:       pos.Leverage,
				IsolatedMargin: pos.IsolatedMargin,
			})
		}
	}

	// 风险评估已移除（新架构直接使用策略模板参数）
	// RiskEvaluation 不再由 RiskManager 动态计算

	// 方向信号
	if managed.DirectionSignal != nil {
		res.DirectionSignal = &trading.DirectionSignalRes{
			Direction:  string(managed.DirectionSignal.Direction),
			Strength:   managed.DirectionSignal.Strength,
			Confidence: managed.DirectionSignal.Confidence,
			Action:     string(managed.DirectionSignal.Action),
			Reason:     managed.DirectionSignal.Reason,
		}
	}

	// 市场状态
	analysis := market.GetMarketAnalyzer().GetAnalysis(managed.Platform, managed.Robot.Symbol)
	if analysis != nil {
		res.MarketState = string(analysis.MarketState)
	}

	return res, nil
}

// EngineStatus 获取引擎状态
func (s *sAlertLog) EngineStatus(ctx context.Context, in *trading.EngineStatusReq) (*trading.EngineStatusRes, error) {
	taskManager := toogo.GetRobotTaskManager()

	return &trading.EngineStatusRes{
		Running:             true,
		ActiveRobots:        taskManager.GetActiveRobotCount(),
		ActiveSubscriptions: market.GetMarketServiceManager().GetActiveServiceCount(),
		EngineVersion:       "V2.0",
	}, nil
}
