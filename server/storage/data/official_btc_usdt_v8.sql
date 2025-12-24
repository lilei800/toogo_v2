-- BTC-USDT 官方策略 V8.0
-- 基于半年K线数据和市场波动算法优化的官方推荐策略
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET NAMES utf8mb4;

-- 清理旧数据
DELETE FROM `hg_trading_strategy_template` WHERE `group_id` IN (
  SELECT `id` FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_btc_usdt_v8'
);
DELETE FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_btc_usdt_v8';

-- 插入官方策略组 V8.0
INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  'BTCUSDT-v8官方推荐策略 - 半年数据优化版',
  'official_btc_usdt_v8',
  'bitget',
  'BTCUSDT',
  'market',
  'isolated',
  1,
  0,
  '【策略组说明】Toogo AI量化团队基于半年K线数据和市场波动算法精心打造的BTC-USDT官方推荐策略V8.0版本。\n\n【数据基础】策略参数基于2024年6个月BTCUSDT历史K线数据回测优化，涵盖趋势、震荡、高波动、低波动四种市场状态。\n\n【适用场景】支持Binance/Bitget/OKX/Gate多交易所，适合不同风险偏好的交易者。包含12种智能策略，覆盖保守、平衡、激进三种风险偏好。\n\n【策略特点】\n- 保守型：低杠杆(2-4x)，小止损(2-5%)，适合新手和稳健投资者\n- 平衡型：中杠杆(5-8x)，平衡止损止盈，适合有一定经验的交易者\n- 激进型：高杠杆(10-20x)，大止损(5-10%)，适合专业交易者\n\n【使用建议】根据自身风险承受能力选择合适的策略，建议新手从保守型开始，逐步提升风险偏好。',
  1,
  1
);

SET @group_id = LAST_INSERT_ID();

-- ==================== 保守型策略（4种） ====================

-- 1. 保守-趋势跟踪
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_conservative_trend', '[保守型] 趋势跟踪策略 - 顺势而为',
  'conservative', 'trend',
  300, 85.00, 4, 8.00, 3.00, 30.00, 2.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":70,"multiTimeframeAgreement":3,"trendConfirmation":true},"position":{"leverage":3,"marginPercent":8,"maxPositions":1},"stopLoss":{"percent":3,"moveToBreakeven":true,"breakevenTrigger":2},"takeProfit":{"autoStartRetreat":2,"profitRetreat":30,"trailingStop":true},"risk":{"maxDailyLoss":5,"maxDrawdown":10}}',
  '【策略名称】保守型趋势跟踪策略\n【适用市场】趋势市场（有明显上涨或下跌趋势）\n【风险等级】低风险（1-2级）\n【杠杆倍数】3-4倍\n【监控窗口】300秒（5分钟）\n【波动阈值】85 USDT\n【止损设置】3%（价格回撤3%自动止损）\n【止盈设置】30%（盈利回撤30%自动止盈，启动回撤2%）\n【预期收益】日收益0.5-2%，月收益15-60%\n【适用人群】新手、稳健投资者、风险厌恶者\n【策略说明】低杠杆顺势交易，多周期确认入场。当市场出现明显趋势时，跟随趋势方向开仓，通过多周期技术指标确认信号强度。适合BTCUSDT在趋势市场中的稳健盈利。\n【使用建议】建议在趋势明确时使用，避免在震荡市场使用。建议初始资金1000 USDT以上。',
  1, 101
);

-- 2. 保守-区间震荡
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_conservative_volatile', '[保守型] 区间震荡策略 - 高抛低吸',
  'conservative', 'volatile',
  180, 55.00, 2, 6.00, 2.50, 25.00, 1.50,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"supportResistance":true,"rangeTrading":true},"position":{"leverage":2,"marginPercent":6,"maxPositions":1},"stopLoss":{"percent":2.5},"takeProfit":{"autoStartRetreat":1.5,"profitRetreat":25,"partialTake":true},"risk":{"maxDailyLoss":4,"maxDrawdown":8}}',
  '【策略名称】保守型区间震荡策略\n【适用市场】震荡市场（价格在一定区间内波动）\n【风险等级】低风险（1级）\n【杠杆倍数】2倍\n【监控窗口】180秒（3分钟）\n【波动阈值】55 USDT\n【止损设置】2.5%（价格回撤2.5%自动止损）\n【止盈设置】25%（盈利回撤25%自动止盈，启动回撤1.5%）\n【预期收益】日收益0.5-1.5%，月收益15-45%\n【适用人群】新手、稳健投资者\n【策略说明】震荡市场高抛低吸，布林带+支撑阻力双重确认。当价格接近支撑位时做多，接近阻力位时做空，通过技术指标确认买卖点。适合BTCUSDT在震荡市场中的稳定盈利。\n【使用建议】建议在震荡区间明确时使用，避免在趋势市场使用。建议初始资金500 USDT以上。',
  1, 102
);

-- 3. 保守-高波动防守
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_conservative_high_vol', '[保守型] 高波动防守策略 - 最小仓位',
  'conservative', 'high_vol',
  120, 160.00, 2, 5.00, 5.00, 35.00, 3.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m"],"primaryTimeFrame":"1m","entry":{"signalStrength":80,"volatilityFilter":true,"minVolatility":120,"maxVolatility":250},"position":{"leverage":2,"marginPercent":5,"dynamicSize":true},"stopLoss":{"percent":5,"widthAdjust":true},"takeProfit":{"autoStartRetreat":3,"profitRetreat":35,"trailingStop":true},"risk":{"maxDailyLoss":6,"maxDrawdown":12}}',
  '【策略名称】保守型高波动防守策略\n【适用市场】高波动市场（价格剧烈波动，波动率>2%）\n【风险等级】中低风险（2级）\n【杠杆倍数】2倍\n【监控窗口】120秒（2分钟）\n【波动阈值】160 USDT\n【止损设置】5%（价格回撤5%自动止损，动态调整）\n【止盈设置】35%（盈利回撤35%自动止盈，启动回撤3%）\n【预期收益】日收益1-3%，月收益30-90%\n【适用人群】有一定经验的交易者、风险承受能力中等\n【策略说明】高波动市场最小仓位防守，动态调整止损宽度。当市场波动剧烈时，使用最小仓位和较大止损空间，避免被市场噪音触发止损。适合BTCUSDT在高波动市场中的防守型盈利。\n【使用建议】建议在波动率明显高于平时时使用，需要密切关注市场变化。建议初始资金2000 USDT以上。',
  1, 103
);

-- 4. 保守-低波动蓄力
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_conservative_low_vol', '[保守型] 低波动蓄力策略 - 等待突破',
  'conservative', 'low_vol',
  600, 32.00, 4, 10.00, 2.00, 20.00, 1.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["15m","30m","1h"],"primaryTimeFrame":"30m","entry":{"signalStrength":65,"breakoutWait":true,"breakoutConfirmBars":3,"volumeConfirmation":true},"position":{"leverage":4,"marginPercent":10,"pyramiding":true,"maxPyramid":2},"stopLoss":{"percent":2},"takeProfit":{"autoStartRetreat":1,"profitRetreat":20},"risk":{"maxDailyLoss":3,"maxDrawdown":6}}',
  '【策略名称】保守型低波动蓄力策略\n【适用市场】低波动市场（价格变化缓慢，波动率<0.5%）\n【风险等级】低风险（1级）\n【杠杆倍数】4倍\n【监控窗口】600秒（10分钟）\n【波动阈值】32 USDT\n【止损设置】2%（价格回撤2%自动止损）\n【止盈设置】20%（盈利回撤20%自动止盈，启动回撤1%）\n【预期收益】日收益0.3-1%，月收益9-30%\n【适用人群】新手、稳健投资者、长期持有者\n【策略说明】低波动等待突破，支持金字塔加仓。当市场波动较小时，等待价格突破关键位置，通过成交量确认突破有效性。适合BTCUSDT在低波动市场中的稳健盈利。\n【使用建议】建议在波动率明显低于平时时使用，需要耐心等待突破信号。建议初始资金1000 USDT以上。',
  1, 104
);

-- ==================== 平衡型策略（4种） ====================

-- 5. 平衡-趋势跟踪（推荐）
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_balanced_trend', '[平衡型] 趋势跟踪策略 - 多周期确认 ⭐推荐',
  'balanced', 'trend',
  240, 100.00, 8, 12.00, 5.00, 25.00, 3.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":65,"multiTimeframeAgreement":3,"trendConfirmation":true,"emaAlignment":true},"position":{"leverage":8,"marginPercent":12,"maxPositions":2,"pyramiding":true,"maxPyramid":2},"stopLoss":{"percent":5,"moveToBreakeven":true,"breakevenTrigger":3},"takeProfit":{"autoStartRetreat":3,"profitRetreat":25,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":8,"maxDrawdown":15}}',
  '【策略名称】平衡型趋势跟踪策略（最推荐）\n【适用市场】趋势市场（有明显上涨或下跌趋势）\n【风险等级】中等风险（3级）\n【杠杆倍数】7-8倍\n【监控窗口】240秒（4分钟）\n【波动阈值】100 USDT\n【止损设置】5%（价格回撤5%自动止损，盈利后移至成本价）\n【止盈设置】25%（盈利回撤25%自动止盈，启动回撤3%，支持分批止盈）\n【预期收益】日收益1-5%，月收益30-150%\n【适用人群】有一定经验的交易者、追求稳健增长的投资者\n【策略说明】多周期趋势确认，EMA对齐+MACD动量。当市场出现明显趋势时，通过多周期技术指标确认信号，使用中等杠杆跟随趋势。适合BTCUSDT在趋势市场中的稳健增长。\n【使用建议】最推荐的策略，适合大多数交易者。建议在趋势明确时使用，支持金字塔加仓。建议初始资金2000 USDT以上。',
  1, 201
);

-- 6. 平衡-区间套利
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_balanced_volatile', '[平衡型] 区间套利策略 - 多重确认',
  'balanced', 'volatile',
  180, 65.00, 6, 10.00, 4.00, 22.00, 2.50,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"supportResistance":true,"rangeTrading":true,"bollingerBand":true},"position":{"leverage":6,"marginPercent":10,"maxPositions":2},"stopLoss":{"percent":4},"takeProfit":{"autoStartRetreat":2.5,"profitRetreat":22,"partialTake":true},"risk":{"maxDailyLoss":6,"maxDrawdown":12}}',
  '【策略名称】平衡型区间套利策略\n【适用市场】震荡市场（价格在一定区间内波动）\n【风险等级】中等风险（3级）\n【杠杆倍数】5-6倍\n【监控窗口】180秒（3分钟）\n【波动阈值】65 USDT\n【止损设置】4%（价格回撤4%自动止损）\n【止盈设置】22%（盈利回撤22%自动止盈，启动回撤2.5%，支持分批止盈）\n【预期收益】日收益1-3%，月收益30-90%\n【适用人群】有一定经验的交易者、追求稳定收益\n【策略说明】震荡区间高抛低吸，布林带+RSI+随机指标多重确认。当价格在震荡区间内时，通过多重技术指标确认买卖点，使用中等杠杆进行套利。适合BTCUSDT在震荡市场中的稳定盈利。\n【使用建议】建议在震荡区间明确时使用，避免在趋势市场使用。建议初始资金1500 USDT以上。',
  1, 202
);

-- 7. 平衡-波动捕捉
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_balanced_high_vol', '[平衡型] 波动捕捉策略 - 动态调整',
  'balanced', 'high_vol',
  90, 190.00, 5, 8.00, 6.00, 28.00, 4.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m"],"primaryTimeFrame":"1m","entry":{"signalStrength":75,"volatilityFilter":true,"minVolatility":150,"maxVolatility":300,"momentumStrength":75},"position":{"leverage":5,"marginPercent":8,"dynamicSize":true,"volatilityAdjust":true},"stopLoss":{"percent":6,"widthAdjust":true},"takeProfit":{"autoStartRetreat":4,"profitRetreat":28,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":10,"maxDrawdown":18}}',
  '【策略名称】平衡型波动捕捉策略\n【适用市场】高波动市场（价格剧烈波动，波动率>2%）\n【风险等级】中高风险（4级）\n【杠杆倍数】4-5倍\n【监控窗口】90秒（1.5分钟）\n【波动阈值】190 USDT\n【止损设置】6%（价格回撤6%自动止损，动态调整宽度）\n【止盈设置】28%（盈利回撤28%自动止盈，启动回撤4%，支持分批止盈）\n【预期收益】日收益2-6%，月收益60-180%\n【适用人群】有经验的交易者、能承受较大波动\n【策略说明】高波动市场动态调整仓位止损，快速反应移动止盈。当市场波动剧烈时，根据波动率动态调整仓位和止损宽度，快速捕捉波动带来的盈利机会。适合BTCUSDT在高波动市场中的积极盈利。\n【使用建议】建议在波动率明显高于平时时使用，需要密切关注市场变化。建议初始资金3000 USDT以上。',
  1, 203
);

-- 8. 平衡-突破等待
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_balanced_low_vol', '[平衡型] 突破等待策略 - 布林带挤压',
  'balanced', 'low_vol',
  360, 42.00, 8, 15.00, 3.00, 18.00, 2.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["15m","30m","1h"],"primaryTimeFrame":"1h","entry":{"signalStrength":70,"breakoutWait":true,"breakoutConfirmBars":2,"volumeMultiplier":2,"squeezeTrigger":true},"position":{"leverage":8,"marginPercent":15,"pyramiding":true,"maxPyramid":3},"stopLoss":{"percent":3},"takeProfit":{"autoStartRetreat":2,"profitRetreat":18,"trailingStop":true},"risk":{"maxDailyLoss":5,"maxDrawdown":10}}',
  '【策略名称】平衡型突破等待策略\n【适用市场】低波动市场（价格变化缓慢，波动率<0.5%）\n【风险等级】中等风险（3级）\n【杠杆倍数】7-8倍\n【监控窗口】360秒（6分钟）\n【波动阈值】42 USDT\n【止损设置】3%（价格回撤3%自动止损）\n【止盈设置】18%（盈利回撤18%自动止盈，启动回撤2%）\n【预期收益】日收益0.5-2%，月收益15-60%\n【适用人群】有经验的交易者、追求稳健增长\n【策略说明】低波动等待突破，布林带挤压+肯特纳通道识别蓄力。当市场波动较小时，通过布林带挤压和肯特纳通道识别蓄力信号，等待价格突破关键位置。适合BTCUSDT在低波动市场中的稳健盈利。\n【使用建议】建议在波动率明显低于平时时使用，需要耐心等待突破信号，支持金字塔加仓。建议初始资金2000 USDT以上。',
  1, 204
);

-- ==================== 激进型策略（4种） ====================

-- 9. 激进-趋势冲锋
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_aggressive_trend', '[激进型] 趋势冲锋策略 - 高杠杆追涨 ⚠️高风险',
  'aggressive', 'trend',
  180, 120.00, 15, 18.00, 8.00, 20.00, 5.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":60,"trendConfirmation":true,"momentumStrength":70},"position":{"leverage":15,"marginPercent":18,"maxPositions":3,"pyramiding":true,"maxPyramid":3},"stopLoss":{"percent":8,"moveToBreakeven":true},"takeProfit":{"autoStartRetreat":5,"profitRetreat":20,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":15,"maxDrawdown":25}}',
  '【策略名称】激进型趋势冲锋策略\n【适用市场】趋势市场（有明显上涨或下跌趋势）\n【风险等级】高风险（5级）\n【杠杆倍数】12-15倍\n【监控窗口】180秒（3分钟）\n【波动阈值】120 USDT\n【止损设置】8%（价格回撤8%自动止损，盈利后移至成本价）\n【止盈设置】20%（盈利回撤20%自动止盈，启动回撤5%，支持分批止盈）\n【预期收益】日收益3-10%，月收益90-300%\n【适用人群】专业交易者、能承受高风险高收益\n【策略说明】高杠杆趋势追涨，多级金字塔加仓。当市场出现明显趋势时，使用高杠杆快速追涨，通过多级金字塔加仓放大收益。适合BTCUSDT在趋势市场中的激进盈利。\n【使用建议】⚠️仅限专业用户使用，需要丰富的交易经验和强大的风险承受能力。建议在趋势非常明确时使用，支持多级金字塔加仓。建议初始资金5000 USDT以上。',
  1, 301
);

-- 10. 激进-双向博弈
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_aggressive_volatile', '[激进型] 双向博弈策略 - 对冲持仓 ⚠️高风险',
  'aggressive', 'volatile',
  120, 85.00, 12, 14.00, 6.00, 18.00, 4.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","entry":{"signalStrength":55,"priceAction":true,"divergence":true},"position":{"leverage":12,"marginPercent":14,"hedging":true},"stopLoss":{"percent":6},"takeProfit":{"autoStartRetreat":4,"profitRetreat":18,"partialTake":true},"risk":{"maxDailyLoss":12,"maxDrawdown":20}}',
  '【策略名称】激进型双向博弈策略\n【适用市场】震荡市场（价格在一定区间内波动）\n【风险等级】高风险（5级）\n【杠杆倍数】10-12倍\n【监控窗口】120秒（2分钟）\n【波动阈值】85 USDT\n【止损设置】6%（价格回撤6%自动止损）\n【止盈设置】18%（盈利回撤18%自动止盈，启动回撤4%，支持分批止盈）\n【预期收益】日收益2-8%，月收益60-240%\n【适用人群】专业交易者、能承受高风险高收益\n【策略说明】震荡市场双向开单，支持对冲持仓。当价格在震荡区间内时，通过价格行为和背离信号，使用高杠杆双向开单，通过对冲降低风险。适合BTCUSDT在震荡市场中的激进盈利。\n【使用建议】⚠️高风险策略，仅限专业用户使用。建议在震荡区间明确时使用，支持对冲持仓。建议初始资金5000 USDT以上。',
  1, 302
);

-- 11. 激进-极速博弈
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_aggressive_high_vol', '[激进型] 极速博弈策略 - 快进快出 ⚠️极高风险',
  'aggressive', 'high_vol',
  60, 210.00, 10, 10.00, 10.00, 22.00, 6.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m"],"primaryTimeFrame":"1m","entry":{"signalStrength":70,"volatilityFilter":true,"minVolatility":180,"maxVolatility":350,"momentumStrength":80,"quickEntry":true,"entryTimeout":10},"position":{"leverage":10,"marginPercent":10,"dynamicSize":true,"hedging":true},"stopLoss":{"percent":10,"quickStop":true},"takeProfit":{"autoStartRetreat":6,"profitRetreat":22,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":20,"maxDrawdown":30}}',
  '【策略名称】激进型极速博弈策略\n【适用市场】高波动市场（价格剧烈波动，波动率>2%）\n【风险等级】极高风险（6级）\n【杠杆倍数】8-10倍\n【监控窗口】60秒（1分钟）\n【波动阈值】210 USDT\n【止损设置】10%（价格回撤10%自动止损，快速止损）\n【止盈设置】22%（盈利回撤22%自动止盈，启动回撤6%，支持分批止盈）\n【预期收益】日收益5-15%，月收益150-450%\n【适用人群】专业交易者、能承受极高风险高收益\n【策略说明】高波动快进快出，10秒内入场决策。当市场波动剧烈时，通过快速信号识别，使用高杠杆快速入场和出场，捕捉短期波动带来的盈利机会。适合BTCUSDT在高波动市场中的极速盈利。\n【使用建议】⚠️极高风险策略，仅限专业用户使用。需要丰富的交易经验和强大的风险承受能力。建议在波动率极高时使用，10秒内完成入场决策。建议初始资金10000 USDT以上。',
  1, 303
);

-- 12. 激进-突破狙击
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v8_aggressive_low_vol', '[激进型] 突破狙击策略 - 重仓等待 ⚠️高风险',
  'aggressive', 'low_vol',
  300, 52.00, 18, 22.00, 5.00, 15.00, 3.00,
  '{"version":"8.0","dataSource":"6months","exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m","30m","1h"],"primaryTimeFrame":"15m","entry":{"signalStrength":60,"breakoutWait":true,"breakoutConfirmBars":1,"volumeMultiplier":2.5,"breakoutStrength":80,"fakeoutFilter":true},"position":{"leverage":18,"marginPercent":22,"pyramiding":true,"maxPyramid":4,"scaleInOnBreakout":true},"stopLoss":{"percent":5,"protectProfit":true},"takeProfit":{"autoStartRetreat":3,"profitRetreat":15,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":15,"maxDrawdown":25}}',
  '【策略名称】激进型突破狙击策略\n【适用市场】低波动市场（价格变化缓慢，波动率<0.5%）\n【风险等级】高风险（5级）\n【杠杆倍数】15-18倍\n【监控窗口】300秒（5分钟）\n【波动阈值】52 USDT\n【止损设置】5%（价格回撤5%自动止损，保护盈利）\n【止盈设置】15%（盈利回撤15%自动止盈，启动回撤3%，支持分批止盈）\n【预期收益】日收益1-5%，月收益30-150%\n【适用人群】专业交易者、能承受高风险高收益\n【策略说明】低波动重仓等待大行情突破，4级金字塔加仓。当市场波动较小时，使用高杠杆重仓等待价格突破关键位置，通过成交量确认突破有效性，支持4级金字塔加仓放大收益。适合BTCUSDT在低波动市场中的激进盈利。\n【使用建议】⚠️高风险策略，仅限专业用户使用。建议在波动率明显低于平时时使用，需要耐心等待突破信号，支持4级金字塔加仓。建议初始资金5000 USDT以上。',
  1, 304
);

-- 验证结果
SELECT 
    g.group_name, 
    COUNT(s.id) as strategy_count,
    GROUP_CONCAT(s.strategy_name ORDER BY s.sort SEPARATOR ', ') as strategies
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key = 'official_btc_usdt_v8'
GROUP BY g.id;
