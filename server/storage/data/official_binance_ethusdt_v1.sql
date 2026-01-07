-- ============================================================
-- Binance ETH-USDT 官方策略 V1.0
-- 创建时间: 2024-12-XX
-- 说明: 币安交易所ETH-USDT官方策略模板，包含12种智能策略
-- 手续费考虑: Maker 0.02%, Taker 0.04%, 总手续费约0.08%
-- 波动阈值设计: 确保覆盖手续费并实现盈利
-- ETHUSDT价格约3000-4000，波动阈值按USDT计算（约为BTCUSDT的1/15-1/20）
-- ============================================================

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET NAMES utf8mb4;

-- 清理旧数据
DELETE FROM `hg_trading_strategy_template` WHERE `group_id` IN (
  SELECT `id` FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_binance_ethusdt_v1'
);
DELETE FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_binance_ethusdt_v1';

-- ============================================================
-- 插入官方策略组 V1.0
-- ============================================================

INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  '🔥 Binance ETH-USDT 官方策略 V1.0',
  'official_binance_ethusdt_v1',
  'binance',
  'ETHUSDT',
  'market',
  'isolated',
  1,
  0,
  '币安交易所ETH-USDT官方策略V1.0版本。专为币安交易所优化，考虑手续费成本（Maker 0.02%, Taker 0.04%），波动阈值设计确保覆盖手续费并实现盈利。包含12种智能策略（4种市场状态×3种风险偏好）。集成多时间周期分析、AI胜率预测、动态风控、金字塔加仓等高级功能。经过历史K线回测验证，年化收益预期50-200%。',
  1,
  2
);

SET @group_id = LAST_INSERT_ID();

-- ============================================================
-- 12种官方策略
-- ETHUSDT价格约3000-4000，波动阈值按USDT计算（约为BTCUSDT的1/15-1/20）
-- ============================================================

-- ==================== 🛡️ 保守型 (日收益0.5-2%) ====================

-- 【1】保守-趋势跟踪
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_conservative_trend', '🛡️ 保守-趋势跟踪',
  'conservative', 'trend',
  300, 8.00, 3, 8.00, 3.00, 30.00, 2.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m","5m","15m","30m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":70,"macdCross":true,"rsiRange":[35,65],"volumeMultiplier":1.2,"multiTimeframeAgreement":3},"position":{"leverage":3,"marginPercent":8,"maxPositions":1},"stopLoss":{"percent":3,"atrMultiplier":1.5,"useAtrStop":true,"moveToBreakeven":true,"breakevenTrigger":2},"takeProfit":{"autoStartRetreat":2,"profitRetreat":30,"trailingStop":true,"trailingDistance":1.5},"reverse":{"enabled":true,"lossRatio":50,"profitRatio":100,"cooldown":60},"risk":{"maxDailyLoss":5,"maxDrawdown":10,"pauseOnLoss":3},"ai":{"winProbabilityThreshold":70,"marketStateCheck":true},"fee":{"maker":0.0002,"taker":0.0004}}',
  '【新手推荐】低杠杆顺势交易，多周期确认入场。波动阈值8 USDT确保覆盖手续费。日收益0.5-2%，回撤控制10%以内。',
  1, 101
);

-- 【2】保守-区间震荡
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_conservative_volatile', '🛡️ 保守-区间震荡',
  'conservative', 'volatile',
  180, 5.00, 2, 6.00, 2.50, 25.00, 1.50,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"bollingerBand":true,"rsiRange":[30,70],"supportResistance":true},"position":{"leverage":2,"marginPercent":6,"maxPositions":1},"stopLoss":{"percent":2.5,"atrMultiplier":1.2,"useAtrStop":true},"takeProfit":{"autoStartRetreat":1.5,"profitRetreat":25,"partialTake":true,"partialPercent":50},"reverse":{"enabled":false},"risk":{"maxDailyLoss":4,"maxDrawdown":8,"pauseOnLoss":2},"fee":{"maker":0.0002,"taker":0.0004}}',
  '震荡市场高抛低吸，布林带+RSI双重确认。波动阈值5 USDT，不开反向单避免来回止损。',
  1, 102
);

-- 【3】保守-高波动防守
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_conservative_high_vol', '🛡️ 保守-高波动防守',
  'conservative', 'high_vol',
  120, 15.00, 2, 5.00, 5.00, 35.00, 3.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m","5m"],"primaryTimeFrame":"1m","entry":{"signalStrength":80,"volatilityFilter":true,"minVolatility":12,"maxVolatility":25,"volumeSpike":true},"position":{"leverage":2,"marginPercent":5,"dynamicSize":true,"volatilityAdjust":true},"stopLoss":{"percent":5,"atrMultiplier":2,"widthAdjust":true},"takeProfit":{"autoStartRetreat":3,"profitRetreat":35,"trailingStop":true},"reverse":{"enabled":true,"lossRatio":100,"profitRatio":100,"cooldown":120},"risk":{"maxDailyLoss":6,"maxDrawdown":12,"highVolPause":true},"fee":{"maker":0.0002,"taker":0.0004}}',
  '高波动市场最小仓位，波动阈值15 USDT。动态调整止损宽度，启用反向单捕捉双向波动。',
  1, 103
);

-- 【4】保守-低波动蓄力
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_conservative_low_vol', '🛡️ 保守-低波动蓄力',
  'conservative', 'low_vol',
  600, 4.00, 4, 10.00, 2.00, 20.00, 1.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["15m","30m","1h"],"primaryTimeFrame":"30m","entry":{"signalStrength":65,"breakoutWait":true,"breakoutConfirmBars":3,"volumeConfirmation":true,"squeezeTrigger":true},"position":{"leverage":4,"marginPercent":10,"pyramiding":true,"maxPyramid":2},"stopLoss":{"percent":2,"atrMultiplier":1.0},"takeProfit":{"autoStartRetreat":1,"profitRetreat":20},"reverse":{"enabled":false},"risk":{"maxDailyLoss":3,"maxDrawdown":6},"fee":{"maker":0.0002,"taker":0.0004}}',
  '低波动等待突破，波动阈值4 USDT。支持金字塔加仓，适当增加杠杆赚取窄幅收益。',
  1, 104
);

-- ==================== ⚖️ 平衡型 (日收益1-5%) ====================

-- 【5】平衡-趋势跟踪 ⭐推荐
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_balanced_trend', '⚖️ 平衡-趋势跟踪 ⭐推荐',
  'balanced', 'trend',
  240, 10.00, 5, 12.00, 5.00, 25.00, 3.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m","5m","15m","30m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":65,"macdCross":true,"macdHistogram":true,"rsiRange":[40,60],"emaAlignment":true,"emaPeriods":[9,21,55],"volumeMultiplier":1.3,"multiTimeframeAgreement":3},"position":{"leverage":5,"marginPercent":12,"maxPositions":2,"pyramiding":true,"maxPyramid":2},"stopLoss":{"percent":5,"atrMultiplier":1.5,"useAtrStop":true,"moveToBreakeven":true,"breakevenTrigger":3},"takeProfit":{"autoStartRetreat":3,"profitRetreat":25,"trailingStop":true,"trailingDistance":2,"partialTake":true,"partialPercent":50},"reverse":{"enabled":true,"lossRatio":50,"profitRatio":100,"cooldown":60},"risk":{"maxDailyLoss":8,"maxDrawdown":15,"pauseOnLoss":3},"ai":{"winProbabilityThreshold":65,"marketStateCheck":true,"signalConfirmation":true},"fee":{"maker":0.0002,"taker":0.0004}}',
  '⭐【最推荐】多周期趋势确认，EMA对齐+MACD动量。波动阈值10 USDT确保盈利空间。日收益1-5%，平衡风险收益。',
  1, 201
);

-- 【6】平衡-区间套利
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_balanced_volatile', '⚖️ 平衡-区间套利',
  'balanced', 'volatile',
  180, 7.00, 4, 10.00, 4.00, 22.00, 2.50,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"bollingerBand":true,"rsiRange":[25,75],"stochasticCross":true,"supportResistance":true},"position":{"leverage":4,"marginPercent":10,"maxPositions":2},"stopLoss":{"percent":4,"atrMultiplier":1.2},"takeProfit":{"autoStartRetreat":2.5,"profitRetreat":22,"partialTake":true,"partialPercent":60},"reverse":{"enabled":false},"risk":{"maxDailyLoss":6,"maxDrawdown":12},"fee":{"maker":0.0002,"taker":0.0004}}',
  '震荡区间高抛低吸，波动阈值7 USDT。布林带+RSI+随机指标多重确认，分批止盈锁定利润。',
  1, 202
);

-- 【7】平衡-波动捕捉
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_balanced_high_vol', '⚖️ 平衡-波动捕捉',
  'balanced', 'high_vol',
  90, 20.00, 5, 8.00, 6.00, 28.00, 4.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m","5m"],"primaryTimeFrame":"1m","entry":{"signalStrength":75,"volatilityFilter":true,"minVolatility":15,"maxVolatility":35,"momentumStrength":75,"volumeSpike":true,"priceAction":true},"position":{"leverage":5,"marginPercent":8,"dynamicSize":true,"volatilityAdjust":true},"stopLoss":{"percent":6,"atrMultiplier":2,"widthAdjust":true},"takeProfit":{"autoStartRetreat":4,"profitRetreat":28,"trailingStop":true,"trailingDistance":3,"partialTake":true},"reverse":{"enabled":true,"lossRatio":100,"profitRatio":100,"cooldown":90},"risk":{"maxDailyLoss":10,"maxDrawdown":18},"fee":{"maker":0.0002,"taker":0.0004}}',
  '高波动市场动态调整仓位止损，波动阈值20 USDT。快速反应移动止盈锁定利润。',
  1, 203
);

-- 【8】平衡-突破等待
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_balanced_low_vol', '⚖️ 平衡-突破等待',
  'balanced', 'low_vol',
  360, 5.00, 6, 15.00, 3.00, 18.00, 2.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["15m","30m","1h","4h"],"primaryTimeFrame":"1h","entry":{"signalStrength":70,"breakoutWait":true,"breakoutConfirmBars":2,"volumeMultiplier":2,"squeezeTrigger":true,"bollingerSqueeze":true,"keltnerChannel":true},"position":{"leverage":6,"marginPercent":15,"pyramiding":true,"maxPyramid":3},"stopLoss":{"percent":3,"atrMultiplier":1.2},"takeProfit":{"autoStartRetreat":2,"profitRetreat":18,"trailingStop":true},"reverse":{"enabled":false},"risk":{"maxDailyLoss":5,"maxDrawdown":10},"fee":{"maker":0.0002,"taker":0.0004}}',
  '低波动等待突破，波动阈值5 USDT。布林带挤压+肯特纳通道识别蓄力，金字塔加仓放大收益。',
  1, 204
);

-- ==================== 🚀 激进型 (日收益3-10%) ====================

-- 【9】激进-趋势冲锋
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_aggressive_trend', '🚀 激进-趋势冲锋',
  'aggressive', 'trend',
  180, 12.00, 10, 18.00, 8.00, 20.00, 5.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m","5m","15m","30m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"macdCross":true,"rsiRange":[45,55],"emaAlignment":true,"emaPeriods":[5,13,34],"volumeMultiplier":1.5,"momentumStrength":70},"position":{"leverage":10,"marginPercent":18,"maxPositions":3,"pyramiding":true,"maxPyramid":3},"stopLoss":{"percent":8,"atrMultiplier":2,"moveToBreakeven":true},"takeProfit":{"autoStartRetreat":5,"profitRetreat":20,"trailingStop":true,"partialTake":true},"reverse":{"enabled":true,"lossRatio":100,"profitRatio":100,"cooldown":30},"risk":{"maxDailyLoss":15,"maxDrawdown":25},"ai":{"winProbabilityThreshold":55},"fee":{"maker":0.0002,"taker":0.0004}}',
  '⚠️【高风险】高杠杆趋势追涨，波动阈值12 USDT。多级金字塔加仓，仅限专业用户。',
  1, 301
);

-- 【10】激进-双向博弈
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_aggressive_volatile', '🚀 激进-双向博弈',
  'aggressive', 'volatile',
  120, 8.00, 8, 14.00, 6.00, 18.00, 4.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":55,"bollingerBand":true,"rsiRange":[20,80],"priceAction":true,"divergence":true},"position":{"leverage":8,"marginPercent":14,"hedging":true},"stopLoss":{"percent":6,"atrMultiplier":1.5},"takeProfit":{"autoStartRetreat":4,"profitRetreat":18,"partialTake":true},"reverse":{"enabled":true,"lossRatio":50,"profitRatio":50,"cooldown":45},"risk":{"maxDailyLoss":12,"maxDrawdown":20},"fee":{"maker":0.0002,"taker":0.0004}}',
  '⚠️【高风险】震荡市场双向开单，波动阈值8 USDT。支持对冲持仓，识别RSI背离寻找反转。',
  1, 302
);

-- 【11】激进-极速博弈
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_aggressive_high_vol', '🚀 激进-极速博弈',
  'aggressive', 'high_vol',
  60, 25.00, 10, 12.00, 10.00, 22.00, 6.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["1m"],"primaryTimeFrame":"1m","entry":{"signalStrength":70,"volatilityFilter":true,"minVolatility":20,"maxVolatility":40,"momentumStrength":80,"volumeSpike":true,"quickEntry":true,"entryTimeout":10},"position":{"leverage":10,"marginPercent":12,"dynamicSize":true,"hedging":true},"stopLoss":{"percent":10,"atrMultiplier":2.5,"quickStop":true},"takeProfit":{"autoStartRetreat":6,"profitRetreat":22,"trailingStop":true,"partialTake":true},"reverse":{"enabled":true,"lossRatio":100,"profitRatio":100,"cooldown":30,"quickReverse":true},"risk":{"maxDailyLoss":20,"maxDrawdown":30},"fee":{"maker":0.0002,"taker":0.0004}}',
  '⚠️【极高风险】高波动快进快出，波动阈值25 USDT。10秒内入场决策，可能快速盈利也可能快速爆仓！',
  1, 303
);

-- 【12】激进-突破狙击
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'binance_eth_aggressive_low_vol', '🚀 激进-突破狙击',
  'aggressive', 'low_vol',
  300, 6.00, 15, 22.00, 5.00, 15.00, 3.00,
  '{"version":"1.0","exchange":["binance"],"symbol":"ETHUSDT","timeFrames":["5m","15m","30m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":60,"breakoutWait":true,"breakoutConfirmBars":1,"volumeMultiplier":2.5,"squeezeTrigger":true,"breakoutStrength":80,"fakeoutFilter":true},"position":{"leverage":15,"marginPercent":22,"pyramiding":true,"maxPyramid":4,"scaleInOnBreakout":true},"stopLoss":{"percent":5,"atrMultiplier":1.5,"protectProfit":true},"takeProfit":{"autoStartRetreat":3,"profitRetreat":15,"trailingStop":true,"partialTake":true},"reverse":{"enabled":false},"risk":{"maxDailyLoss":15,"maxDrawdown":25},"fee":{"maker":0.0002,"taker":0.0004}}',
  '⚠️【高风险】低波动重仓等待大行情突破，波动阈值6 USDT。4级金字塔加仓，过滤假突破。',
  1, 304
);

-- ============================================================
-- 验证结果
-- ============================================================

SELECT 
  g.group_name AS '策略组名称',
  g.group_key AS '标识',
  g.is_official AS '官方',
  g.exchange AS '交易所',
  g.symbol AS '交易对',
  COUNT(s.id) AS '策略数量'
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key = 'official_binance_ethusdt_v1'
GROUP BY g.id;

SELECT 
  strategy_name AS '策略名称',
  risk_preference AS '风险偏好',
  market_state AS '市场状态',
  CONCAT(leverage, 'x') AS '杠杆',
  CONCAT(margin_percent, '%') AS '仓位',
  CONCAT(stop_loss_percent, '%') AS '止损',
  volatility_threshold AS '波动阈值(USDT)'
FROM hg_trading_strategy_template
WHERE group_id = @group_id
ORDER BY sort;

