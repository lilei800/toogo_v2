-- BTC-USDT 官方策略 V5.0
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET NAMES utf8mb4;

-- 清理旧数据
DELETE FROM `hg_trading_strategy_template` WHERE `group_id` IN (
  SELECT `id` FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_btc_usdt_v5'
);
DELETE FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_btc_usdt_v5';

-- 插入官方策略组 V5.0
INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  'BTC-USDT 官方策略 V5.0',
  'official_btc_usdt_v5',
  'bitget',
  'BTCUSDT',
  'market',
  'isolated',
  1,
  0,
  'Toogo AI量化团队精心打造的BTC-USDT官方策略V5.0版本。支持Binance/Bitget/OKX/Gate多交易所，包含12种智能策略。',
  1,
  1
);

SET @group_id = LAST_INSERT_ID();

-- 1. 保守-趋势跟踪
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_conservative_trend', '[保守] 趋势跟踪',
  'conservative', 'trend',
  300, 80.00, 3, 5, 5.00, 10.00, 3.00, 30.00, 2.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m","30m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":70,"macdCross":true,"rsiRange":[35,65],"volumeMultiplier":1.2,"multiTimeframeAgreement":3},"position":{"leverage":4,"marginPercent":8,"maxPositions":1},"stopLoss":{"percent":3,"atrMultiplier":1.5,"useAtrStop":true,"moveToBreakeven":true,"breakevenTrigger":2},"takeProfit":{"autoStartRetreat":2,"profitRetreat":30,"trailingStop":true,"trailingDistance":1.5},"reverse":{"enabled":true,"lossRatio":50,"profitRatio":100,"cooldown":60},"risk":{"maxDailyLoss":5,"maxDrawdown":10,"pauseOnLoss":3},"ai":{"winProbabilityThreshold":70,"marketStateCheck":true}}',
  '新手推荐：低杠杆顺势交易，多周期确认入场。日收益0.5-2%',
  1, 101
);

-- 2. 保守-区间震荡
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_conservative_volatile', '[保守] 区间震荡',
  'conservative', 'volatile',
  180, 50.00, 2, 4, 4.00, 8.00, 2.50, 25.00, 1.50,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"bollingerBand":true,"rsiRange":[30,70],"supportResistance":true},"position":{"leverage":3,"marginPercent":6,"maxPositions":1},"stopLoss":{"percent":2.5,"atrMultiplier":1.2,"useAtrStop":true},"takeProfit":{"autoStartRetreat":1.5,"profitRetreat":25,"partialTake":true,"partialPercent":50},"reverse":{"enabled":false},"risk":{"maxDailyLoss":4,"maxDrawdown":8,"pauseOnLoss":2}}',
  '震荡市场高抛低吸，布林带+RSI双重确认',
  1, 102
);

-- 3. 保守-高波动防守
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_conservative_high_vol', '[保守] 高波动防守',
  'conservative', 'high_vol',
  120, 150.00, 2, 3, 3.00, 6.00, 5.00, 35.00, 3.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m"],"primaryTimeFrame":"1m","entry":{"signalStrength":80,"volatilityFilter":true,"minVolatility":100,"maxVolatility":200,"volumeSpike":true},"position":{"leverage":2,"marginPercent":5,"dynamicSize":true,"volatilityAdjust":true},"stopLoss":{"percent":5,"atrMultiplier":2,"widthAdjust":true},"takeProfit":{"autoStartRetreat":3,"profitRetreat":35,"trailingStop":true},"reverse":{"enabled":true,"lossRatio":100,"profitRatio":100,"cooldown":120},"risk":{"maxDailyLoss":6,"maxDrawdown":12,"highVolPause":true}}',
  '高波动市场最小仓位，动态调整止损宽度',
  1, 103
);

-- 4. 保守-低波动蓄力
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_conservative_low_vol', '[保守] 低波动蓄力',
  'conservative', 'low_vol',
  600, 30.00, 4, 6, 6.00, 12.00, 2.00, 20.00, 1.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["15m","30m","1h"],"primaryTimeFrame":"30m","entry":{"signalStrength":65,"breakoutWait":true,"breakoutConfirmBars":3,"volumeConfirmation":true,"squeezeTrigger":true},"position":{"leverage":5,"marginPercent":10,"pyramiding":true,"maxPyramid":2},"stopLoss":{"percent":2,"atrMultiplier":1.0},"takeProfit":{"autoStartRetreat":1,"profitRetreat":20},"reverse":{"enabled":false},"risk":{"maxDailyLoss":3,"maxDrawdown":6}}',
  '低波动等待突破，支持金字塔加仓',
  1, 104
);

-- 5. 平衡-趋势跟踪 推荐
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_balanced_trend', '[平衡] 趋势跟踪 *推荐*',
  'balanced', 'trend',
  240, 100.00, 5, 10, 8.00, 15.00, 5.00, 25.00, 3.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m","30m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":65,"macdCross":true,"macdHistogram":true,"rsiRange":[40,60],"emaAlignment":true,"emaPeriods":[9,21,55],"volumeMultiplier":1.3,"multiTimeframeAgreement":3},"position":{"leverage":8,"marginPercent":12,"maxPositions":2,"pyramiding":true,"maxPyramid":2},"stopLoss":{"percent":5,"atrMultiplier":1.5,"useAtrStop":true,"moveToBreakeven":true,"breakevenTrigger":3},"takeProfit":{"autoStartRetreat":3,"profitRetreat":25,"trailingStop":true,"trailingDistance":2,"partialTake":true,"partialPercent":50},"reverse":{"enabled":true,"lossRatio":50,"profitRatio":100,"cooldown":60},"risk":{"maxDailyLoss":8,"maxDrawdown":15,"pauseOnLoss":3},"ai":{"winProbabilityThreshold":65,"marketStateCheck":true,"signalConfirmation":true}}',
  '【最推荐】多周期趋势确认，EMA对齐+MACD动量。日收益1-5%',
  1, 201
);

-- 6. 平衡-区间套利
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_balanced_volatile', '[平衡] 区间套利',
  'balanced', 'volatile',
  180, 60.00, 5, 8, 6.00, 12.00, 4.00, 22.00, 2.50,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"bollingerBand":true,"rsiRange":[25,75],"stochasticCross":true,"supportResistance":true},"position":{"leverage":6,"marginPercent":10,"maxPositions":2},"stopLoss":{"percent":4,"atrMultiplier":1.2},"takeProfit":{"autoStartRetreat":2.5,"profitRetreat":22,"partialTake":true,"partialPercent":60},"reverse":{"enabled":false},"risk":{"maxDailyLoss":6,"maxDrawdown":12}}',
  '震荡区间高抛低吸，布林带+RSI+随机指标多重确认',
  1, 202
);

-- 7. 平衡-波动捕捉
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_balanced_high_vol', '[平衡] 波动捕捉',
  'balanced', 'high_vol',
  90, 180.00, 4, 7, 5.00, 10.00, 6.00, 28.00, 4.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m"],"primaryTimeFrame":"1m","entry":{"signalStrength":75,"volatilityFilter":true,"minVolatility":120,"maxVolatility":250,"momentumStrength":75,"volumeSpike":true,"priceAction":true},"position":{"leverage":5,"marginPercent":8,"dynamicSize":true,"volatilityAdjust":true},"stopLoss":{"percent":6,"atrMultiplier":2,"widthAdjust":true},"takeProfit":{"autoStartRetreat":4,"profitRetreat":28,"trailingStop":true,"trailingDistance":3,"partialTake":true},"reverse":{"enabled":true,"lossRatio":100,"profitRatio":100,"cooldown":90},"risk":{"maxDailyLoss":10,"maxDrawdown":18}}',
  '高波动市场动态调整仓位止损，快速反应移动止盈',
  1, 203
);

-- 8. 平衡-突破等待
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_balanced_low_vol', '[平衡] 突破等待',
  'balanced', 'low_vol',
  360, 40.00, 6, 10, 10.00, 18.00, 3.00, 18.00, 2.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["15m","30m","1h","4h"],"primaryTimeFrame":"1h","entry":{"signalStrength":70,"breakoutWait":true,"breakoutConfirmBars":2,"volumeMultiplier":2,"squeezeTrigger":true,"bollingerSqueeze":true,"keltnerChannel":true},"position":{"leverage":8,"marginPercent":15,"pyramiding":true,"maxPyramid":3},"stopLoss":{"percent":3,"atrMultiplier":1.2},"takeProfit":{"autoStartRetreat":2,"profitRetreat":18,"trailingStop":true},"reverse":{"enabled":false},"risk":{"maxDailyLoss":5,"maxDrawdown":10}}',
  '低波动等待突破，布林带挤压+肯特纳通道识别蓄力',
  1, 204
);

-- 9. 激进-趋势冲锋
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_aggressive_trend', '[激进] 趋势冲锋',
  'aggressive', 'trend',
  180, 120.00, 10, 20, 10.00, 20.00, 8.00, 20.00, 5.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m","30m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"macdCross":true,"rsiRange":[45,55],"emaAlignment":true,"emaPeriods":[5,13,34],"volumeMultiplier":1.5,"momentumStrength":70},"position":{"leverage":15,"marginPercent":18,"maxPositions":3,"pyramiding":true,"maxPyramid":3},"stopLoss":{"percent":8,"atrMultiplier":2,"moveToBreakeven":true},"takeProfit":{"autoStartRetreat":5,"profitRetreat":20,"trailingStop":true,"partialTake":true},"reverse":{"enabled":true,"lossRatio":100,"profitRatio":100,"cooldown":30},"risk":{"maxDailyLoss":15,"maxDrawdown":25},"ai":{"winProbabilityThreshold":55}}',
  '【高风险】高杠杆趋势追涨，多级金字塔加仓。仅限专业用户',
  1, 301
);

-- 10. 激进-双向博弈
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_aggressive_volatile', '[激进] 双向博弈',
  'aggressive', 'volatile',
  120, 80.00, 8, 15, 8.00, 16.00, 6.00, 18.00, 4.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":55,"bollingerBand":true,"rsiRange":[20,80],"priceAction":true,"divergence":true},"position":{"leverage":12,"marginPercent":14,"hedging":true},"stopLoss":{"percent":6,"atrMultiplier":1.5},"takeProfit":{"autoStartRetreat":4,"profitRetreat":18,"partialTake":true},"reverse":{"enabled":true,"lossRatio":50,"profitRatio":50,"cooldown":45},"risk":{"maxDailyLoss":12,"maxDrawdown":20}}',
  '【高风险】震荡市场双向开单，支持对冲持仓',
  1, 302
);

-- 11. 激进-极速博弈
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_aggressive_high_vol', '[激进] 极速博弈',
  'aggressive', 'high_vol',
  60, 200.00, 8, 12, 6.00, 12.00, 10.00, 22.00, 6.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m"],"primaryTimeFrame":"1m","entry":{"signalStrength":70,"volatilityFilter":true,"minVolatility":150,"maxVolatility":300,"momentumStrength":80,"volumeSpike":true,"quickEntry":true,"entryTimeout":10},"position":{"leverage":10,"marginPercent":10,"dynamicSize":true,"hedging":true},"stopLoss":{"percent":10,"atrMultiplier":2.5,"quickStop":true},"takeProfit":{"autoStartRetreat":6,"profitRetreat":22,"trailingStop":true,"partialTake":true},"reverse":{"enabled":true,"lossRatio":100,"profitRatio":100,"cooldown":30,"quickReverse":true},"risk":{"maxDailyLoss":20,"maxDrawdown":30}}',
  '【极高风险】高波动快进快出，10秒内入场决策',
  1, 303
);

-- 12. 激进-突破狙击
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v5_aggressive_low_vol', '[激进] 突破狙击',
  'aggressive', 'low_vol',
  300, 50.00, 12, 20, 15.00, 25.00, 5.00, 15.00, 3.00,
  '{"version":"5.0","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m","30m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":60,"breakoutWait":true,"breakoutConfirmBars":1,"volumeMultiplier":2.5,"squeezeTrigger":true,"breakoutStrength":80,"fakeoutFilter":true},"position":{"leverage":18,"marginPercent":22,"pyramiding":true,"maxPyramid":4,"scaleInOnBreakout":true},"stopLoss":{"percent":5,"atrMultiplier":1.5,"protectProfit":true},"takeProfit":{"autoStartRetreat":3,"profitRetreat":15,"trailingStop":true,"partialTake":true},"reverse":{"enabled":false},"risk":{"maxDailyLoss":15,"maxDrawdown":25}}',
  '【高风险】低波动重仓等待大行情突破，4级金字塔加仓',
  1, 304
);

-- 验证结果
SELECT g.group_name, COUNT(s.id) as strategy_count
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key = 'official_btc_usdt_v5'
GROUP BY g.id;

