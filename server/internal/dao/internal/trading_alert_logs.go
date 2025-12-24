// Package internal DAO内部实现
package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingMarketStateLogDao 市场状态预警日志DAO
type TradingMarketStateLogDao struct {
	table   string
	group   string
	columns TradingMarketStateLogColumns
}

type TradingMarketStateLogColumns struct {
	Id            string
	Platform      string
	Symbol        string
	PrevState     string
	NewState      string
	Confidence    string
	TrendStrength string
	Volatility    string
	TrendScore    string
	MomentumScore string
	Reason        string
	Indicators    string
	CreatedAt     string
}

var tradingMarketStateLogColumns = TradingMarketStateLogColumns{
	Id:            "id",
	Platform:      "platform",
	Symbol:        "symbol",
	PrevState:     "prev_state",
	NewState:      "new_state",
	Confidence:    "confidence",
	TrendStrength: "trend_strength",
	Volatility:    "volatility",
	TrendScore:    "trend_score",
	MomentumScore: "momentum_score",
	Reason:        "reason",
	Indicators:    "indicators",
	CreatedAt:     "created_at",
}

func NewTradingMarketStateLogDao() *TradingMarketStateLogDao {
	return &TradingMarketStateLogDao{
		table:   "hg_trading_market_state_log",
		group:   "default",
		columns: tradingMarketStateLogColumns,
	}
}

func (d *TradingMarketStateLogDao) DB() gdb.DB                { return g.DB(d.group) }
func (d *TradingMarketStateLogDao) Table() string             { return d.table }
func (d *TradingMarketStateLogDao) Columns() TradingMarketStateLogColumns { return d.columns }
func (d *TradingMarketStateLogDao) Group() string             { return d.group }
func (d *TradingMarketStateLogDao) Ctx(ctx context.Context) *gdb.Model { return d.DB().Model(d.table).Safe().Ctx(ctx) }

// TradingRiskPreferenceLogDao 风险偏好预警日志DAO
type TradingRiskPreferenceLogDao struct {
	table   string
	group   string
	columns TradingRiskPreferenceLogColumns
}

type TradingRiskPreferenceLogColumns struct {
	Id                     string
	RobotId                string
	UserId                 string
	Platform               string
	Symbol                 string
	PrevPreference         string
	NewPreference          string
	WinProbability         string
	MarketScore            string
	TechnicalScore         string
	AccountScore           string
	HistoryScore           string
	VolatilityRisk         string
	SuggestedLeverage      string
	SuggestedMarginPercent string
	SuggestedStopLoss      string
	SuggestedTakeProfit    string
	Reason                 string
	Factors                string
	CreatedAt              string
}

var tradingRiskPreferenceLogColumns = TradingRiskPreferenceLogColumns{
	Id:                     "id",
	RobotId:                "robot_id",
	UserId:                 "user_id",
	Platform:               "platform",
	Symbol:                 "symbol",
	PrevPreference:         "prev_preference",
	NewPreference:          "new_preference",
	WinProbability:         "win_probability",
	MarketScore:            "market_score",
	TechnicalScore:         "technical_score",
	AccountScore:           "account_score",
	HistoryScore:           "history_score",
	VolatilityRisk:         "volatility_risk",
	SuggestedLeverage:      "suggested_leverage",
	SuggestedMarginPercent: "suggested_margin_percent",
	SuggestedStopLoss:      "suggested_stop_loss",
	SuggestedTakeProfit:    "suggested_take_profit",
	Reason:                 "reason",
	Factors:                "factors",
	CreatedAt:              "created_at",
}

func NewTradingRiskPreferenceLogDao() *TradingRiskPreferenceLogDao {
	return &TradingRiskPreferenceLogDao{
		table:   "hg_trading_risk_preference_log",
		group:   "default",
		columns: tradingRiskPreferenceLogColumns,
	}
}

func (d *TradingRiskPreferenceLogDao) DB() gdb.DB                { return g.DB(d.group) }
func (d *TradingRiskPreferenceLogDao) Table() string             { return d.table }
func (d *TradingRiskPreferenceLogDao) Columns() TradingRiskPreferenceLogColumns { return d.columns }
func (d *TradingRiskPreferenceLogDao) Group() string             { return d.group }
func (d *TradingRiskPreferenceLogDao) Ctx(ctx context.Context) *gdb.Model { return d.DB().Model(d.table).Safe().Ctx(ctx) }

// TradingDirectionLogDao 方向预警日志DAO
type TradingDirectionLogDao struct {
	table   string
	group   string
	columns TradingDirectionLogColumns
}

type TradingDirectionLogColumns struct {
	Id               string
	Platform         string
	Symbol           string
	PrevDirection    string
	NewDirection     string
	Strength         string
	Confidence       string
	Action           string
	TrendSignal      string
	MomentumSignal   string
	PatternSignal    string
	NearSupport      string
	NearResistance   string
	EntryPrice       string
	StopLoss         string
	TakeProfit1      string
	TakeProfit2      string
	Reason           string
	TimeframeSignals string
	Indicators       string
	CreatedAt        string
}

var tradingDirectionLogColumns = TradingDirectionLogColumns{
	Id:               "id",
	Platform:         "platform",
	Symbol:           "symbol",
	PrevDirection:    "prev_direction",
	NewDirection:     "new_direction",
	Strength:         "strength",
	Confidence:       "confidence",
	Action:           "action",
	TrendSignal:      "trend_signal",
	MomentumSignal:   "momentum_signal",
	PatternSignal:    "pattern_signal",
	NearSupport:      "near_support",
	NearResistance:   "near_resistance",
	EntryPrice:       "entry_price",
	StopLoss:         "stop_loss",
	TakeProfit1:      "take_profit_1",
	TakeProfit2:      "take_profit_2",
	Reason:           "reason",
	TimeframeSignals: "timeframe_signals",
	Indicators:       "indicators",
	CreatedAt:        "created_at",
}

func NewTradingDirectionLogDao() *TradingDirectionLogDao {
	return &TradingDirectionLogDao{
		table:   "hg_trading_direction_log",
		group:   "default",
		columns: tradingDirectionLogColumns,
	}
}

func (d *TradingDirectionLogDao) DB() gdb.DB             { return g.DB(d.group) }
func (d *TradingDirectionLogDao) Table() string          { return d.table }
func (d *TradingDirectionLogDao) Columns() TradingDirectionLogColumns { return d.columns }
func (d *TradingDirectionLogDao) Group() string          { return d.group }
func (d *TradingDirectionLogDao) Ctx(ctx context.Context) *gdb.Model { return d.DB().Model(d.table).Safe().Ctx(ctx) }

// TradingRobotRealtimeDao 机器人实时状态DAO
type TradingRobotRealtimeDao struct {
	table   string
	group   string
	columns TradingRobotRealtimeColumns
}

type TradingRobotRealtimeColumns struct {
	Id                  string
	RobotId             string
	UserId              string
	Platform            string
	Symbol              string
	CurrentPrice        string
	PriceChange24h      string
	MarketState         string
	MarketStateConf     string
	TrendStrength       string
	Volatility          string
	RiskPreference      string
	WinProbability      string
	RiskLevel           string
	Direction           string
	DirectionStrength   string
	DirectionConfidence string
	SuggestedAction     string
	HasPosition         string
	PositionSide        string
	PositionAmt         string
	PositionPnl         string
	PositionPnlPercent  string
	AccountBalance      string
	AvailableBalance    string
	UpdatedAt           string
}

var tradingRobotRealtimeColumns = TradingRobotRealtimeColumns{
	Id:                  "id",
	RobotId:             "robot_id",
	UserId:              "user_id",
	Platform:            "platform",
	Symbol:              "symbol",
	CurrentPrice:        "current_price",
	PriceChange24h:      "price_change_24h",
	MarketState:         "market_state",
	MarketStateConf:     "market_state_conf",
	TrendStrength:       "trend_strength",
	Volatility:          "volatility",
	RiskPreference:      "risk_preference",
	WinProbability:      "win_probability",
	RiskLevel:           "risk_level",
	Direction:           "direction",
	DirectionStrength:   "direction_strength",
	DirectionConfidence: "direction_confidence",
	SuggestedAction:     "suggested_action",
	HasPosition:         "has_position",
	PositionSide:        "position_side",
	PositionAmt:         "position_amt",
	PositionPnl:         "position_pnl",
	PositionPnlPercent:  "position_pnl_percent",
	AccountBalance:      "account_balance",
	AvailableBalance:    "available_balance",
	UpdatedAt:           "updated_at",
}

func NewTradingRobotRealtimeDao() *TradingRobotRealtimeDao {
	return &TradingRobotRealtimeDao{
		table:   "hg_trading_robot_realtime",
		group:   "default",
		columns: tradingRobotRealtimeColumns,
	}
}

func (d *TradingRobotRealtimeDao) DB() gdb.DB          { return g.DB(d.group) }
func (d *TradingRobotRealtimeDao) Table() string       { return d.table }
func (d *TradingRobotRealtimeDao) Columns() TradingRobotRealtimeColumns { return d.columns }
func (d *TradingRobotRealtimeDao) Group() string       { return d.group }
func (d *TradingRobotRealtimeDao) Ctx(ctx context.Context) *gdb.Model { return d.DB().Model(d.table).Safe().Ctx(ctx) }
