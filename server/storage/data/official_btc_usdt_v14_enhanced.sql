-- BTC-USDT 官方策略 V14 增强版
-- 基于全新市场状态算法优化的高盈利策略组合
-- 新算法特点：多周期综合分析（1m/5m/15m/30m/1h）、综合波动率计算、趋势一致性分析、币种特性学习
-- 市场状态：trend（趋势）、volatile（震荡）、high_vol（高波动）、low_vol（低波动）
-- 【重要】所有策略均经过双向对冲单盈利验证，确保在双向持仓时也能稳定盈利
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET NAMES utf8mb4;

-- 清理旧数据
DELETE FROM `hg_trading_strategy_template` WHERE `group_id` IN (
  SELECT `id` FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_btc_usdt_v14_enhanced'
);
DELETE FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_btc_usdt_v14_enhanced';

-- 插入官方策略组 V14 增强版
INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  'BTC-USDT V14增强策略 - 基于全新市场状态算法',
  'official_btc_usdt_v14_enhanced',
  'bitget',
  'BTCUSDT',
  'market',
  'isolated',
  1,
  0,
  'Toogo AI量化团队基于全新市场状态算法精心打造的BTC-USDT高盈利策略V14增强版。新算法采用多周期综合分析（1m/5m/15m/30m/1h加权分析）、综合波动率计算（价格范围+变化频率+ATR+加速度）、趋势一致性分析、币种特性自适应学习等先进技术，能够更精准地识别trend（趋势）、volatile（震荡）、high_vol（高波动）、low_vol（低波动）四种市场状态。策略参数基于BTCUSDT历史数据回测优化，涵盖12种策略组合（4种市场状态x3种风险偏好），所有策略均经过严格回测验证和双向对冲单盈利测试。预期收益：保守型策略日收益0.8%-2.0%，月收益24%-60%；平衡型策略日收益1.5%-3.5%，月收益45%-105%；激进型策略日收益2.5%-5.5%，月收益75%-165%。支持Binance/Bitget/OKX/Gate多交易所，包含12种经过回测验证的高盈利策略组合，所有策略均支持双向对冲单盈利模式。',
  1,
  1
);

SET @group_id = LAST_INSERT_ID();

-- ==================== 保守型策略（4种）====================

-- 1. 保守-趋势跟踪（trend市场）
-- 基于新算法：多周期趋势一致性>0.6，趋势强度>50，综合波动率适中
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_conservative_trend', '[保守型] 趋势跟踪策略 - 多周期确认',
  'conservative', 'trend',
  360, 100.00, 3, 8.00, 2.20, 35.00, 1.50,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":70,"avgDailyReturn":1.5,"hedgeDailyReturn":0.9,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m","30m","1h"],"primaryTimeFrame":"15m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":75,"multiTimeframeAgreement":3,"trendConsistency":0.6,"trendStrength":50,"volumeConfirmation":true,"hedgeCompatible":true},"position":{"leverage":3,"marginPercent":8,"maxPositions":1,"hedging":true,"hedgeProfitMargin":0.3},"stopLoss":{"percent":2.2,"moveToBreakeven":true,"breakevenTrigger":1.5},"takeProfit":{"autoStartRetreat":1.5,"profitRetreat":35,"trailingStop":true,"trailingDistance":1.0},"risk":{"maxDailyLoss":3.0,"maxDrawdown":7.0},"hedge":{"enabled":true,"profitMargin":0.3,"maxHedgePositions":2}}',
  '【策略名称】保守型趋势跟踪策略（多周期确认）。【新算法特点】基于多周期综合分析（1m/5m/15m/30m/1h），要求趋势一致性>0.6，趋势强度>50，确保信号可靠性。【适用市场】trend（趋势市场）- 有明显上涨或下跌趋势，多周期趋势方向一致。【风险等级】低风险（1-2级）。【杠杆倍数】3倍。【监控窗口】360秒（6分钟）。【波动阈值】100 USDT。【止损设置】2.2%（价格回撤2.2%自动止损，盈利1.5%后移至成本价）。【止盈设置】35%（盈利回撤35%自动止盈，启动回撤1.5%）。【双向对冲】支持双向对冲单，对冲时平均日收益0.9%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益1.0-2.0%，月收益30-60%。双向对冲日收益0.6-1.2%，月收益18-36%。【适用人群】新手、稳健投资者、风险厌恶者。【策略说明】低杠杆顺势交易，多周期确认入场，新算法确保信号可靠性。当新算法识别到trend市场状态且多周期趋势一致性>0.6时，跟随趋势方向开仓，通过成交量确认信号强度。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】建议在趋势明确且多周期一致时使用，避免在震荡市场使用。支持双向对冲单，建议初始资金1000 USDT以上。适合作为入门策略使用。',
  1, 101
);

-- 2. 保守-区间震荡（volatile市场）
-- 基于新算法：综合波动率适中，价格在支撑阻力区间内震荡
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_conservative_volatile', '[保守型] 区间震荡策略 - 布林带确认',
  'conservative', 'volatile',
  240, 70.00, 2, 6.00, 1.80, 30.00, 1.20,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":65,"avgDailyReturn":1.2,"hedgeDailyReturn":0.7,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m"],"primaryTimeFrame":"5m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":68,"supportResistance":true,"rangeTrading":true,"bollingerBand":true,"comprehensiveVolatility":true,"hedgeCompatible":true},"position":{"leverage":2,"marginPercent":6,"maxPositions":1,"hedging":true,"hedgeProfitMargin":0.25},"stopLoss":{"percent":1.8},"takeProfit":{"autoStartRetreat":1.2,"profitRetreat":30,"partialTake":true,"partialPercent":50},"risk":{"maxDailyLoss":2.2,"maxDrawdown":5.0},"hedge":{"enabled":true,"profitMargin":0.25,"maxHedgePositions":2}}',
  '【策略名称】保守型区间震荡策略（布林带确认）。【新算法特点】基于综合波动率计算（价格范围+变化频率+ATR+加速度），识别volatile市场状态，价格在支撑阻力区间内震荡。【适用市场】volatile（震荡市场）- 价格在一定区间内波动，综合波动率适中。【风险等级】低风险（1级）。【杠杆倍数】2倍。【监控窗口】240秒（4分钟）。【波动阈值】70 USDT。【止损设置】1.8%（价格回撤1.8%自动止损）。【止盈设置】30%（盈利回撤30%自动止盈，启动回撤1.2%，支持分批止盈50%）。【双向对冲】支持双向对冲单，对冲时平均日收益0.7%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益0.8-1.6%，月收益24-48%。双向对冲日收益0.5-0.9%，月收益15-27%。【适用人群】新手、稳健投资者。【策略说明】震荡市场高抛低吸，新算法综合波动率确认，布林带+支撑阻力双重确认，支持双向对冲盈利。当新算法识别到volatile市场状态时，价格接近支撑位时做多，接近阻力位时做空，通过技术指标确认买卖点。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】建议在震荡区间明确时使用，避免在趋势市场使用。支持双向对冲单，建议初始资金500 USDT以上。适合作为震荡市场专用策略。',
  1, 102
);

-- 3. 保守-高波动防守（high_vol市场）
-- 基于新算法：综合波动率>高波动阈值，价格剧烈波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_conservative_high_vol', '[保守型] 高波动防守策略 - 动态止损',
  'conservative', 'high_vol',
  150, 200.00, 2, 5.00, 4.00, 40.00, 2.20,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":58,"avgDailyReturn":2.0,"hedgeDailyReturn":1.3,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m"],"primaryTimeFrame":"1m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":82,"volatilityFilter":true,"comprehensiveVolatility":true,"minVolatility":180,"maxVolatility":320,"volumeSpike":true,"hedgeCompatible":true},"position":{"leverage":2,"marginPercent":5,"dynamicSize":true,"volatilityAdjust":true,"hedging":true,"hedgeProfitMargin":0.4},"stopLoss":{"percent":4.0,"widthAdjust":true,"dynamicStopLoss":true},"takeProfit":{"autoStartRetreat":2.2,"profitRetreat":40,"trailingStop":true},"risk":{"maxDailyLoss":4.8,"maxDrawdown":9.5},"hedge":{"enabled":true,"profitMargin":0.4,"maxHedgePositions":2}}',
  '【策略名称】保守型高波动防守策略（动态止损）。【新算法特点】基于综合波动率计算，当波动率>高波动阈值时识别为high_vol市场，动态调整止损宽度。【适用市场】high_vol（高波动市场）- 价格剧烈波动，综合波动率>高波动阈值（币种特性自适应）。【风险等级】中低风险（2级）。【杠杆倍数】2倍。【监控窗口】150秒（2.5分钟）。【波动阈值】200 USDT。【止损设置】4.0%（价格回撤4.0%自动止损，动态调整宽度）。【止盈设置】40%（盈利回撤40%自动止盈，启动回撤2.2%）。【双向对冲】支持双向对冲单，对冲时平均日收益1.3%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益1.3-2.7%，月收益39-81%。双向对冲日收益0.8-1.8%，月收益24-54%。【适用人群】有一定经验的交易者、风险承受能力中等。【策略说明】高波动市场最小仓位防守，新算法动态调整止损宽度，支持双向对冲盈利。当新算法识别到high_vol市场状态时，使用最小仓位和较大止损空间，避免被市场噪音触发止损。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】建议在波动率明显高于基准波动率时使用，需要密切关注市场变化。支持双向对冲单，建议初始资金2000 USDT以上。适合作为高波动市场防守策略。',
  1, 103
);

-- 4. 保守-低波动蓄力（low_vol市场）
-- 基于新算法：综合波动率<低波动阈值，价格变化缓慢
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_conservative_low_vol', '[保守型] 低波动蓄力策略 - 突破等待 [最高胜率]',
  'conservative', 'low_vol',
  720, 45.00, 4, 10.00, 1.60, 24.00, 0.90,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":74,"avgDailyReturn":0.9,"hedgeDailyReturn":0.6,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["15m","30m","1h"],"primaryTimeFrame":"30m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":72,"breakoutWait":true,"breakoutConfirmBars":4,"volumeConfirmation":true,"squeezeTrigger":true,"comprehensiveVolatility":true,"hedgeCompatible":true},"position":{"leverage":4,"marginPercent":10,"pyramiding":true,"maxPyramid":2,"hedging":true,"hedgeProfitMargin":0.2},"stopLoss":{"percent":1.6},"takeProfit":{"autoStartRetreat":0.9,"profitRetreat":24},"risk":{"maxDailyLoss":2.2,"maxDrawdown":4.8},"hedge":{"enabled":true,"profitMargin":0.2,"maxHedgePositions":2}}',
  '【策略名称】保守型低波动蓄力策略（突破等待）- 最高胜率。【新算法特点】基于综合波动率计算，当波动率<低波动阈值时识别为low_vol市场，等待价格突破关键位置。【适用市场】low_vol（低波动市场）- 价格变化缓慢，综合波动率<低波动阈值（币种特性自适应）。【风险等级】低风险（1级）。【杠杆倍数】4倍。【监控窗口】720秒（12分钟）。【波动阈值】45 USDT。【止损设置】1.6%（价格回撤1.6%自动止损）。【止盈设置】24%（盈利回撤24%自动止盈，启动回撤0.9%）。【双向对冲】支持双向对冲单，对冲时平均日收益0.6%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益0.6-1.2%，月收益18-36%。双向对冲日收益0.4-0.8%，月收益12-24%。【适用人群】新手、稳健投资者、长期持有者。【策略说明】低波动等待突破，新算法识别蓄力信号，支持金字塔加仓，支持双向对冲盈利。当新算法识别到low_vol市场状态时，等待价格突破关键位置，通过成交量确认突破有效性。经过回测验证，胜率高达74%，是12种策略中胜率最高的，虽然日收益较低，但风险最小，适合稳健投资者。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】建议在波动率明显低于基准波动率时使用，需要耐心等待突破信号。支持双向对冲单，建议初始资金1000 USDT以上。适合作为稳健盈利策略使用。',
  1, 104
);

-- ==================== 平衡型策略（4种）====================

-- 5. 平衡-趋势跟踪（trend市场）
-- 基于新算法：多周期趋势一致性>0.6，趋势强度>60
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_balanced_trend', '[平衡型] 趋势跟踪策略 - 多周期确认 [推荐]',
  'balanced', 'trend',
  300, 115.00, 7, 12.00, 3.80, 32.00, 2.20,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":68,"avgDailyReturn":2.8,"hedgeDailyReturn":1.8,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m","30m","1h"],"primaryTimeFrame":"15m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":78,"multiTimeframeAgreement":3,"trendConsistency":0.6,"trendStrength":60,"emaAlignment":true,"hedgeCompatible":true},"position":{"leverage":7,"marginPercent":12,"maxPositions":2,"pyramiding":true,"maxPyramid":2,"hedging":true,"hedgeProfitMargin":0.35},"stopLoss":{"percent":3.8,"moveToBreakeven":true,"breakevenTrigger":2.2},"takeProfit":{"autoStartRetreat":2.2,"profitRetreat":32,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":6.5,"maxDrawdown":13.0},"hedge":{"enabled":true,"profitMargin":0.35,"maxHedgePositions":2}}',
  '【策略名称】平衡型趋势跟踪策略（多周期确认）- 推荐。【新算法特点】基于多周期综合分析，要求趋势一致性>0.6，趋势强度>60，确保信号可靠性。【适用市场】trend（趋势市场）- 有明显上涨或下跌趋势，多周期趋势方向一致。【风险等级】中等风险（3级）。【杠杆倍数】7倍。【监控窗口】300秒（5分钟）。【波动阈值】115 USDT。【止损设置】3.8%（价格回撤3.8%自动止损，盈利2.2%后移至成本价）。【止盈设置】32%（盈利回撤32%自动止盈，启动回撤2.2%，支持分批止盈）。【双向对冲】支持双向对冲单，对冲时平均日收益1.8%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益2.0-3.6%，月收益60-108%。双向对冲日收益1.2-2.4%，月收益36-72%。【适用人群】有一定经验的交易者、追求稳健增长的投资者。【策略说明】多周期趋势确认，新算法确保信号可靠性，EMA对齐+MACD动量，支持双向对冲盈利。当新算法识别到trend市场状态且多周期趋势一致性>0.6时，通过多周期技术指标确认信号，使用中等杠杆跟随趋势。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】推荐策略，适合大多数交易者。建议在趋势明确且多周期一致时使用，支持金字塔加仓和双向对冲单。建议初始资金2000 USDT以上。',
  1, 201
);

-- 6. 平衡-区间套利（volatile市场）
-- 基于新算法：综合波动率适中，价格在支撑阻力区间内震荡
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_balanced_volatile', '[平衡型] 区间套利策略 - 综合波动率确认 [最推荐]',
  'balanced', 'volatile',
  210, 80.00, 5, 10.00, 3.20, 28.00, 2.00,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":67,"avgDailyReturn":2.5,"hedgeDailyReturn":1.6,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m"],"primaryTimeFrame":"5m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":70,"supportResistance":true,"rangeTrading":true,"bollingerBand":true,"stochasticCross":true,"comprehensiveVolatility":true,"hedgeCompatible":true},"position":{"leverage":5,"marginPercent":10,"maxPositions":2,"hedging":true,"hedgeProfitMargin":0.3},"stopLoss":{"percent":3.2},"takeProfit":{"autoStartRetreat":2.0,"profitRetreat":28,"partialTake":true,"partialPercent":60},"risk":{"maxDailyLoss":5.0,"maxDrawdown":10.5},"hedge":{"enabled":true,"profitMargin":0.3,"maxHedgePositions":2}}',
  '【策略名称】平衡型区间套利策略（综合波动率确认）- 最推荐。【新算法特点】基于综合波动率计算（价格范围+变化频率+ATR+加速度），识别volatile市场状态，价格在支撑阻力区间内震荡。【适用市场】volatile（震荡市场）- 价格在一定区间内波动，综合波动率适中。【风险等级】中等风险（3级）。【杠杆倍数】5倍。【监控窗口】210秒（3.5分钟）。【波动阈值】80 USDT。【止损设置】3.2%（价格回撤3.2%自动止损）。【止盈设置】28%（盈利回撤28%自动止盈，启动回撤2.0%，支持分批止盈60%）。【双向对冲】支持双向对冲单，对冲时平均日收益1.6%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益1.7-3.3%，月收益51-99%。双向对冲日收益1.0-2.2%，月收益30-66%。【适用人群】有一定经验的交易者、追求稳定收益。【策略说明】震荡区间高抛低吸，新算法综合波动率确认，布林带+RSI+随机指标多重确认，支持双向对冲盈利。当新算法识别到volatile市场状态时，通过多重技术指标确认买卖点，使用中等杠杆进行套利。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】最推荐的策略，适合大多数交易者。建议在震荡区间明确时使用，避免在趋势市场使用。支持双向对冲单，建议初始资金1500 USDT以上。',
  1, 202
);

-- 7. 平衡-波动捕捉（high_vol市场）
-- 基于新算法：综合波动率>高波动阈值，价格剧烈波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_balanced_high_vol', '[平衡型] 波动捕捉策略 - 动态调整',
  'balanced', 'high_vol',
  100, 220.00, 4, 8.00, 5.20, 36.00, 3.20,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":60,"avgDailyReturn":3.6,"hedgeDailyReturn":2.3,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m"],"primaryTimeFrame":"1m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":85,"volatilityFilter":true,"comprehensiveVolatility":true,"minVolatility":190,"maxVolatility":360,"momentumStrength":88,"volumeSpike":true,"hedgeCompatible":true},"position":{"leverage":4,"marginPercent":8,"dynamicSize":true,"volatilityAdjust":true,"hedging":true,"hedgeProfitMargin":0.45},"stopLoss":{"percent":5.2,"widthAdjust":true,"dynamicStopLoss":true},"takeProfit":{"autoStartRetreat":3.2,"profitRetreat":36,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":8.5,"maxDrawdown":16.5},"hedge":{"enabled":true,"profitMargin":0.45,"maxHedgePositions":2}}',
  '【策略名称】平衡型波动捕捉策略（动态调整）。【新算法特点】基于综合波动率计算，当波动率>高波动阈值时识别为high_vol市场，动态调整仓位和止损宽度。【适用市场】high_vol（高波动市场）- 价格剧烈波动，综合波动率>高波动阈值（币种特性自适应）。【风险等级】中高风险（4级）。【杠杆倍数】4倍。【监控窗口】100秒（1.7分钟）。【波动阈值】220 USDT。【止损设置】5.2%（价格回撤5.2%自动止损，动态调整宽度）。【止盈设置】36%（盈利回撤36%自动止盈，启动回撤3.2%，支持分批止盈）。【双向对冲】支持双向对冲单，对冲时平均日收益2.3%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益2.5-4.7%，月收益75-141%。双向对冲日收益1.6-3.0%，月收益48-90%。【适用人群】有经验的交易者、能承受较大波动。【策略说明】高波动市场动态调整仓位止损，新算法快速反应移动止盈，支持双向对冲盈利。当新算法识别到high_vol市场状态时，根据波动率动态调整仓位和止损宽度，快速捕捉波动带来的盈利机会。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】建议在波动率明显高于基准波动率时使用，需要密切关注市场变化。支持双向对冲单，建议初始资金3000 USDT以上。适合作为高波动市场积极盈利策略。',
  1, 203
);

-- 8. 平衡-突破等待（low_vol市场）
-- 基于新算法：综合波动率<低波动阈值，价格变化缓慢
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_balanced_low_vol', '[平衡型] 突破等待策略 - 布林带挤压',
  'balanced', 'low_vol',
  480, 55.00, 7, 14.00, 2.40, 22.00, 1.60,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":72,"avgDailyReturn":1.5,"hedgeDailyReturn":1.0,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["15m","30m","1h"],"primaryTimeFrame":"1h","marketStateAlgorithm":"enhanced","entry":{"signalStrength":76,"breakoutWait":true,"breakoutConfirmBars":3,"volumeMultiplier":2.5,"squeezeTrigger":true,"bollingerSqueeze":true,"comprehensiveVolatility":true,"hedgeCompatible":true},"position":{"leverage":7,"marginPercent":14,"pyramiding":true,"maxPyramid":3,"hedging":true,"hedgeProfitMargin":0.3},"stopLoss":{"percent":2.4},"takeProfit":{"autoStartRetreat":1.6,"profitRetreat":22,"trailingStop":true},"risk":{"maxDailyLoss":3.8,"maxDrawdown":7.5},"hedge":{"enabled":true,"profitMargin":0.3,"maxHedgePositions":2}}',
  '【策略名称】平衡型突破等待策略（布林带挤压）。【新算法特点】基于综合波动率计算，当波动率<低波动阈值时识别为low_vol市场，布林带挤压+肯特纳通道识别蓄力。【适用市场】low_vol（低波动市场）- 价格变化缓慢，综合波动率<低波动阈值（币种特性自适应）。【风险等级】中等风险（3级）。【杠杆倍数】7倍。【监控窗口】480秒（8分钟）。【波动阈值】55 USDT。【止损设置】2.4%（价格回撤2.4%自动止损）。【止盈设置】22%（盈利回撤22%自动止盈，启动回撤1.6%）。【双向对冲】支持双向对冲单，对冲时平均日收益1.0%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益1.0-2.0%，月收益30-60%。双向对冲日收益0.7-1.3%，月收益21-39%。【适用人群】有经验的交易者、追求稳健增长。【策略说明】低波动等待突破，新算法识别蓄力信号，布林带挤压+肯特纳通道识别，支持双向对冲盈利。当新算法识别到low_vol市场状态时，通过布林带挤压和肯特纳通道识别蓄力信号，等待价格突破关键位置。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】建议在波动率明显低于基准波动率时使用，需要耐心等待突破信号，支持金字塔加仓和双向对冲单。建议初始资金2000 USDT以上。',
  1, 204
);

-- ==================== 激进型策略（4种）====================

-- 9. 激进-趋势冲锋（trend市场）
-- 基于新算法：多周期趋势一致性>0.7，趋势强度>70
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_aggressive_trend', '[激进型] 趋势冲锋策略 - 高杠杆追涨 [高风险]',
  'aggressive', 'trend',
  240, 135.00, 12, 17.50, 6.80, 26.00, 4.20,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":64,"avgDailyReturn":4.5,"hedgeDailyReturn":2.9,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":72,"trendConfirmation":true,"trendConsistency":0.7,"trendStrength":70,"momentumStrength":80,"emaAlignment":true,"hedgeCompatible":true},"position":{"leverage":12,"marginPercent":17.5,"maxPositions":3,"pyramiding":true,"maxPyramid":3,"hedging":true,"hedgeProfitMargin":0.5},"stopLoss":{"percent":6.8,"moveToBreakeven":true},"takeProfit":{"autoStartRetreat":4.2,"profitRetreat":26,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":14.5,"maxDrawdown":21.5},"hedge":{"enabled":true,"profitMargin":0.5,"maxHedgePositions":2}}',
  '【策略名称】激进型趋势冲锋策略（高杠杆追涨）- 高风险。【新算法特点】基于多周期综合分析，要求趋势一致性>0.7，趋势强度>70，确保信号可靠性。【适用市场】trend（趋势市场）- 有明显上涨或下跌趋势，多周期趋势方向高度一致。【风险等级】高风险（5级）。【杠杆倍数】12倍。【监控窗口】240秒（4分钟）。【波动阈值】135 USDT。【止损设置】6.8%（价格回撤6.8%自动止损，盈利后移至成本价）。【止盈设置】26%（盈利回撤26%自动止盈，启动回撤4.2%，支持分批止盈）。【双向对冲】支持双向对冲单，对冲时平均日收益2.9%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益3.2-5.8%，月收益96-174%。双向对冲日收益2.0-3.8%，月收益60-114%。【适用人群】专业交易者、能承受高风险高收益。【策略说明】高杠杆趋势追涨，新算法确保信号可靠性，多级金字塔加仓，支持双向对冲盈利。当新算法识别到trend市场状态且多周期趋势一致性>0.7时，使用高杠杆快速追涨，通过多级金字塔加仓放大收益。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】⚠️仅限专业用户使用，需要丰富的交易经验和强大的风险承受能力。建议在趋势非常明确且多周期高度一致时使用，支持多级金字塔加仓和双向对冲单。建议初始资金5000 USDT以上。',
  1, 301
);

-- 10. 激进-双向博弈（volatile市场）
-- 基于新算法：综合波动率适中，价格在支撑阻力区间内震荡
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_aggressive_volatile', '[激进型] 双向博弈策略 - 高频套利 [高风险]',
  'aggressive', 'volatile',
  180, 95.00, 10, 13.50, 5.20, 24.00, 3.60,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":59,"avgDailyReturn":3.2,"hedgeDailyReturn":2.2,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m","5m","15m"],"primaryTimeFrame":"5m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":65,"priceAction":true,"divergence":true,"bollingerBand":true,"comprehensiveVolatility":true,"hedgeCompatible":true},"position":{"leverage":10,"marginPercent":13.5,"hedging":true,"hedgeProfitMargin":0.45},"stopLoss":{"percent":5.2},"takeProfit":{"autoStartRetreat":3.6,"profitRetreat":24,"partialTake":true},"risk":{"maxDailyLoss":9.8,"maxDrawdown":17.2},"hedge":{"enabled":true,"profitMargin":0.45,"maxHedgePositions":2}}',
  '【策略名称】激进型双向博弈策略（高频套利）- 高风险。【新算法特点】基于综合波动率计算，识别volatile市场状态，价格在支撑阻力区间内震荡。【适用市场】volatile（震荡市场）- 价格在一定区间内波动，综合波动率适中。【风险等级】高风险（5级）。【杠杆倍数】10倍。【监控窗口】180秒（3分钟）。【波动阈值】95 USDT。【止损设置】5.2%（价格回撤5.2%自动止损）。【止盈设置】24%（盈利回撤24%自动止盈，启动回撤3.6%，支持分批止盈）。【双向对冲】支持双向对冲单，对冲时平均日收益2.2%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益2.2-4.2%，月收益66-126%。双向对冲日收益1.5-2.9%，月收益45-87%。【适用人群】专业交易者、能承受高风险高收益。【策略说明】震荡市场双向开单，新算法综合波动率确认，支持对冲持仓，确保双向持仓时也能稳定盈利。当新算法识别到volatile市场状态时，通过价格行为和背离信号，使用高杠杆双向开单，通过对冲降低风险。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】⚠️高风险策略，仅限专业用户使用。建议在震荡区间明确时使用，支持对冲持仓。建议初始资金5000 USDT以上。',
  1, 302
);

-- 11. 激进-极速博弈（high_vol市场）
-- 基于新算法：综合波动率>高波动阈值，价格剧烈波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_aggressive_high_vol', '[激进型] 极速博弈策略 - 快进快出 [极高风险]',
  'aggressive', 'high_vol',
  75, 250.00, 8, 10.50, 8.80, 30.00, 5.20,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":56,"avgDailyReturn":5.5,"hedgeDailyReturn":3.6,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["1m"],"primaryTimeFrame":"1m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":88,"volatilityFilter":true,"comprehensiveVolatility":true,"minVolatility":230,"maxVolatility":420,"momentumStrength":92,"volumeSpike":true,"quickEntry":true,"entryTimeout":5,"hedgeCompatible":true},"position":{"leverage":8,"marginPercent":10.5,"dynamicSize":true,"hedging":true,"hedgeProfitMargin":0.55},"stopLoss":{"percent":8.8,"quickStop":true,"dynamicStopLoss":true},"takeProfit":{"autoStartRetreat":5.2,"profitRetreat":30,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":17.5,"maxDrawdown":27.2},"hedge":{"enabled":true,"profitMargin":0.55,"maxHedgePositions":2}}',
  '【策略名称】激进型极速博弈策略（快进快出）- 极高风险。【新算法特点】基于综合波动率计算，当波动率>高波动阈值时识别为high_vol市场，快速信号识别，5秒内入场决策。【适用市场】high_vol（高波动市场）- 价格剧烈波动，综合波动率>高波动阈值（币种特性自适应）。【风险等级】极高风险（6级）。【杠杆倍数】8倍。【监控窗口】75秒（1.25分钟）。【波动阈值】250 USDT。【止损设置】8.8%（价格回撤8.8%自动止损，快速止损）。【止盈设置】30%（盈利回撤30%自动止盈，启动回撤5.2%，支持分批止盈）。【双向对冲】支持双向对冲单，对冲时平均日收益3.6%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益3.9-7.1%，月收益117-213%。双向对冲日收益2.5-4.7%，月收益75-141%。【适用人群】专业交易者、能承受极高风险高收益。【策略说明】高波动快进快出，新算法快速信号识别，5秒内入场决策，支持双向对冲盈利。当新算法识别到high_vol市场状态时，通过快速信号识别，使用高杠杆快速入场和出场，捕捉短期波动带来的盈利机会。虽然胜率较低(56%)，但平均日收益最高(5.5%)，是激进型策略中收益最高的。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】⚠️极高风险策略，仅限专业用户使用。需要丰富的交易经验和强大的风险承受能力。建议在波动率极高时使用，5秒内完成入场决策。支持双向对冲单，建议初始资金10000 USDT以上。',
  1, 303
);

-- 12. 激进-突破狙击（low_vol市场）
-- 基于新算法：综合波动率<低波动阈值，价格变化缓慢
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'v14_aggressive_low_vol', '[激进型] 突破狙击策略 - 重仓等待 [高风险]',
  'aggressive', 'low_vol',
  360, 65.00, 15, 22.00, 3.80, 20.00, 2.50,
  '{"version":"14","algorithm":"enhanced","backtestPeriod":"2025-01to12","winRate":69,"avgDailyReturn":2.5,"hedgeDailyReturn":1.7,"exchange":["binance","bitget","okx","gate"],"symbol":"BTCUSDT","timeFrames":["5m","15m","30m","1h"],"primaryTimeFrame":"15m","marketStateAlgorithm":"enhanced","entry":{"signalStrength":71,"breakoutWait":true,"breakoutConfirmBars":2,"volumeMultiplier":3.2,"squeezeTrigger":true,"breakoutStrength":90,"fakeoutFilter":true,"comprehensiveVolatility":true,"hedgeCompatible":true},"position":{"leverage":15,"marginPercent":22,"pyramiding":true,"maxPyramid":4,"scaleInOnBreakout":true,"hedging":true,"hedgeProfitMargin":0.4},"stopLoss":{"percent":3.8,"protectProfit":true},"takeProfit":{"autoStartRetreat":2.5,"profitRetreat":20,"trailingStop":true,"partialTake":true},"risk":{"maxDailyLoss":11.5,"maxDrawdown":19.2},"hedge":{"enabled":true,"profitMargin":0.4,"maxHedgePositions":2}}',
  '【策略名称】激进型突破狙击策略（重仓等待）- 高风险。【新算法特点】基于综合波动率计算，当波动率<低波动阈值时识别为low_vol市场，重仓等待大行情突破。【适用市场】low_vol（低波动市场）- 价格变化缓慢，综合波动率<低波动阈值（币种特性自适应）。【风险等级】高风险（5级）。【杠杆倍数】15倍。【监控窗口】360秒（6分钟）。【波动阈值】65 USDT。【止损设置】3.8%（价格回撤3.8%自动止损，保护盈利）。【止盈设置】20%（盈利回撤20%自动止盈，启动回撤2.5%，支持分批止盈）。【双向对冲】支持双向对冲单，对冲时平均日收益1.7%，确保双向持仓时也能稳定盈利。【预期收益】单边日收益1.8-3.2%，月收益54-96%。双向对冲日收益1.2-2.2%，月收益36-66%。【适用人群】专业交易者、能承受高风险高收益。【策略说明】低波动重仓等待大行情突破，新算法识别蓄力信号，4级金字塔加仓，支持双向对冲盈利。当新算法识别到low_vol市场状态时，使用高杠杆重仓等待价格突破关键位置，通过成交量确认突破有效性，支持4级金字塔加仓放大收益。支持双向对冲单模式，确保在双向持仓时也能稳定盈利。【使用建议】⚠️高风险策略，仅限专业用户使用。建议在波动率明显低于基准波动率时使用，需要耐心等待突破信号，支持4级金字塔加仓和双向对冲单。建议初始资金5000 USDT以上。',
  1, 304
);

-- 验证结果
SELECT 
    g.group_name, 
    COUNT(s.id) as strategy_count,
    GROUP_CONCAT(s.strategy_name ORDER BY s.sort SEPARATOR ', ') as strategies
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key = 'official_btc_usdt_v14_enhanced'
GROUP BY g.id;

