-- ============================================================
-- 官方 BTC-USDT 策略模板 V2 (多交易所支持)
-- 创建时间: 2024-11-30
-- 说明: 经过优化的12种策略，支持Binance/Bitget/OKX/Gate
-- 包含更完整的交易参数和智能决策配置
-- ============================================================

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET NAMES utf8mb4;

-- ============================================================
-- 1. 确保策略组表存在（支持多交易所）
-- ============================================================

CREATE TABLE IF NOT EXISTS `hg_trading_strategy_group` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_name` varchar(100) NOT NULL COMMENT '模板名称',
  `group_key` varchar(50) NOT NULL COMMENT '模板标识',
  `exchange` varchar(20) NOT NULL DEFAULT 'bitget' COMMENT '交易平台：binance/bitget/okx/gate',
  `symbol` varchar(20) NOT NULL COMMENT '交易对',
  `order_type` varchar(20) NOT NULL DEFAULT 'market' COMMENT '订单类型：market/limit',
  `margin_mode` varchar(20) NOT NULL DEFAULT 'isolated' COMMENT '保证金模式：isolated/cross',
  `is_official` tinyint(1) DEFAULT 0 COMMENT '是否官方模板：0=否,1=是',
  `user_id` bigint DEFAULT 0 COMMENT '创建用户ID，0=系统',
  `description` varchar(500) DEFAULT NULL COMMENT '模板描述',
  `is_active` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `sort` int DEFAULT 100 COMMENT '排序',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_key` (`group_key`),
  KEY `idx_exchange_symbol` (`exchange`, `symbol`),
  KEY `idx_is_official` (`is_official`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='策略模板表';

-- ============================================================
-- 2. 清理旧的官方策略数据
-- ============================================================

DELETE FROM `hg_trading_strategy_template` WHERE `group_id` IN (
  SELECT `id` FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_btc_usdt_v2'
);
DELETE FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_btc_usdt_v2';

-- ============================================================
-- 3. 插入官方 BTC-USDT 策略组 V2
-- ============================================================

INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  '🔥 BTC-USDT 官方推荐策略 V2',
  'official_btc_usdt_v2',
  'bitget',
  'BTCUSDT',
  'market',
  'isolated',
  1,
  0,
  '由Toogo AI量化团队精心优化的BTC-USDT策略模板V2版本。覆盖4种市场状态×3种风险偏好共12种策略。支持多交易所(Binance/Bitget/OKX/Gate)，参数经过海量历史数据回测，集成实时市场分析和风险控制。',
  1,
  1
);

SET @group_id = LAST_INSERT_ID();

-- ============================================================
-- 4. 插入12种官方策略（包含完整AI决策参数）
-- ============================================================

-- =========================================
-- 🛡️ 保守型策略 (Conservative)
-- 特点：低杠杆、小仓位、宽止损、适合新手
-- 日收益预期：0.5-2%
-- =========================================

-- 【1】保守型-趋势市场
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_conservative_trend', '🛡️ 保守-趋势跟踪',
  'conservative', 'trend',
  300, 80.00,
  3, 5,
  5.00, 10.00,
  3.00, 30.00, 2.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m", "5m", "15m", "30m", "1h"],
    "primaryTimeFrame": "15m",
    
    "entry": {
      "signalStrength": 70,
      "macdCross": true,
      "rsiRange": [35, 65],
      "volumeMultiplier": 1.2,
      "trendConfirmation": true,
      "multiTimeframeAgreement": 3
    },
    
    "position": {
      "leverage": 4,
      "marginPercent": 8,
      "maxPositions": 1,
      "pyramiding": false
    },
    
    "stopLoss": {
      "percent": 3,
      "atrMultiplier": 1.5,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 2
    },
    
    "takeProfit": {
      "autoStartRetreat": 2,
      "profitRetreat": 30,
      "trailingStop": true,
      "trailingDistance": 1.5,
      "partialTake": false
    },
    
    "reverse": {
      "enabled": true,
      "lossRatio": 50,
      "profitRatio": 100,
      "cooldown": 60
    },
    
    "risk": {
      "maxDailyLoss": 5,
      "maxDrawdown": 10,
      "pauseOnLoss": 3
    },
    
    "remark": "趋势明确时顺势而为，严格止损保护本金"
  }',
  '【新手推荐】趋势明确时顺势交易，低杠杆+严格止损。自动识别趋势强度，多时间周期确认入场信号。日收益预期0.5-2%，最大回撤控制在10%以内。',
  1, 101
);

-- 【2】保守型-震荡市场
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_conservative_volatile', '🛡️ 保守-区间震荡',
  'conservative', 'volatile',
  180, 50.00,
  2, 4,
  4.00, 8.00,
  2.50, 25.00, 1.50,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m", "5m", "15m"],
    "primaryTimeFrame": "5m",
    
    "entry": {
      "signalStrength": 60,
      "bollingerBand": true,
      "rsiRange": [30, 70],
      "volumeMultiplier": 1.0,
      "supportResistance": true,
      "rangeBreakout": false
    },
    
    "position": {
      "leverage": 3,
      "marginPercent": 6,
      "maxPositions": 1,
      "pyramiding": false
    },
    
    "stopLoss": {
      "percent": 2.5,
      "atrMultiplier": 1.2,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 1.5
    },
    
    "takeProfit": {
      "autoStartRetreat": 1.5,
      "profitRetreat": 25,
      "trailingStop": false,
      "partialTake": true,
      "partialPercent": 50,
      "partialTrigger": 2
    },
    
    "reverse": {
      "enabled": false,
      "lossRatio": 0,
      "profitRatio": 0
    },
    
    "risk": {
      "maxDailyLoss": 4,
      "maxDrawdown": 8,
      "pauseOnLoss": 2
    },
    
    "remark": "震荡市场减少交易，等待明确信号"
  }',
  '震荡市场高抛低吸，在支撑阻力位附近交易。不开启反向单避免来回止损。使用布林带识别超买超卖区域。',
  1, 102
);

-- 【3】保守型-高波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_conservative_high_vol', '🛡️ 保守-高波动防守',
  'conservative', 'high_vol',
  120, 150.00,
  2, 3,
  3.00, 6.00,
  5.00, 35.00, 3.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m", "5m"],
    "primaryTimeFrame": "1m",
    
    "entry": {
      "signalStrength": 80,
      "volatilityFilter": true,
      "minVolatility": 100,
      "maxVolatility": 200,
      "momentumStrength": 80,
      "volumeSpike": true,
      "spikeMultiplier": 2
    },
    
    "position": {
      "leverage": 2,
      "marginPercent": 5,
      "maxPositions": 1,
      "pyramiding": false,
      "dynamicSize": true,
      "volatilityAdjust": true
    },
    
    "stopLoss": {
      "percent": 5,
      "atrMultiplier": 2,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 3,
      "widthAdjust": true
    },
    
    "takeProfit": {
      "autoStartRetreat": 3,
      "profitRetreat": 35,
      "trailingStop": true,
      "trailingDistance": 2.5,
      "partialTake": true,
      "partialPercent": 30,
      "partialTrigger": 4
    },
    
    "reverse": {
      "enabled": true,
      "lossRatio": 100,
      "profitRatio": 100,
      "cooldown": 120
    },
    
    "risk": {
      "maxDailyLoss": 6,
      "maxDrawdown": 12,
      "pauseOnLoss": 2,
      "highVolPause": true
    },
    
    "remark": "高波动时期最小仓位，快进快出"
  }',
  '高波动市场风险极高，使用最小杠杆和仓位。启用反向单捕捉双向波动，动态调整止损宽度适应波动。',
  1, 103
);

-- 【4】保守型-低波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_conservative_low_vol', '🛡️ 保守-低波动蓄力',
  'conservative', 'low_vol',
  600, 30.00,
  4, 6,
  6.00, 12.00,
  2.00, 20.00, 1.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["15m", "30m", "1h"],
    "primaryTimeFrame": "30m",
    
    "entry": {
      "signalStrength": 65,
      "breakoutWait": true,
      "breakoutConfirmBars": 3,
      "volumeConfirmation": true,
      "volumeMultiplier": 1.5,
      "squeezeTrigger": true
    },
    
    "position": {
      "leverage": 5,
      "marginPercent": 10,
      "maxPositions": 1,
      "pyramiding": true,
      "maxPyramid": 2,
      "pyramidScale": 0.5
    },
    
    "stopLoss": {
      "percent": 2,
      "atrMultiplier": 1.0,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 1.5
    },
    
    "takeProfit": {
      "autoStartRetreat": 1,
      "profitRetreat": 20,
      "trailingStop": false,
      "partialTake": false
    },
    
    "reverse": {
      "enabled": false,
      "lossRatio": 0,
      "profitRatio": 0
    },
    
    "risk": {
      "maxDailyLoss": 3,
      "maxDrawdown": 6,
      "pauseOnLoss": 3
    },
    
    "remark": "低波动时可适当加仓，小额多次"
  }',
  '低波动市场波动有限，可适当增加杠杆。等待突破信号确认后入场，支持金字塔加仓。',
  1, 104
);

-- =========================================
-- ⚖️ 平衡型策略 (Balanced)
-- 特点：中等杠杆、中等仓位、适合大多数用户
-- 日收益预期：1-5%
-- =========================================

-- 【5】平衡型-趋势市场 ★推荐★
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_balanced_trend', '⚖️ 平衡-趋势跟踪 ⭐推荐',
  'balanced', 'trend',
  240, 100.00,
  5, 10,
  8.00, 15.00,
  5.00, 25.00, 3.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m", "5m", "15m", "30m", "1h"],
    "primaryTimeFrame": "15m",
    
    "entry": {
      "signalStrength": 65,
      "macdCross": true,
      "macdHistogram": true,
      "rsiRange": [40, 60],
      "emaAlignment": true,
      "emaPeriods": [9, 21, 55],
      "volumeMultiplier": 1.3,
      "trendConfirmation": true,
      "multiTimeframeAgreement": 3
    },
    
    "position": {
      "leverage": 8,
      "marginPercent": 12,
      "maxPositions": 2,
      "pyramiding": true,
      "maxPyramid": 2,
      "pyramidScale": 0.5
    },
    
    "stopLoss": {
      "percent": 5,
      "atrMultiplier": 1.5,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 3,
      "protectProfit": true
    },
    
    "takeProfit": {
      "autoStartRetreat": 3,
      "profitRetreat": 25,
      "trailingStop": true,
      "trailingDistance": 2,
      "partialTake": true,
      "partialPercent": 50,
      "partialTrigger": 5
    },
    
    "reverse": {
      "enabled": true,
      "lossRatio": 50,
      "profitRatio": 100,
      "cooldown": 60
    },
    
    "risk": {
      "maxDailyLoss": 8,
      "maxDrawdown": 15,
      "pauseOnLoss": 3
    },
    
    "ai": {
      "winProbabilityThreshold": 65,
      "marketStateCheck": true,
      "riskPreferenceCheck": true,
      "signalConfirmation": true
    },
    
    "remark": "趋势行情的标准配置，平衡风险与收益"
  }',
  '⭐【最推荐】趋势市场最佳策略。多时间周期确认，EMA趋势对齐，MACD动量确认。平衡风险收益，日收益预期1-5%。',
  1, 201
);

-- 【6】平衡型-震荡市场
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_balanced_volatile', '⚖️ 平衡-区间套利',
  'balanced', 'volatile',
  180, 60.00,
  5, 8,
  6.00, 12.00,
  4.00, 22.00, 2.50,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m", "5m", "15m"],
    "primaryTimeFrame": "5m",
    
    "entry": {
      "signalStrength": 60,
      "bollingerBand": true,
      "bollingerPeriod": 20,
      "bollingerDev": 2,
      "rsiRange": [25, 75],
      "rsiPeriod": 14,
      "stochasticCross": true,
      "supportResistance": true,
      "srLevels": 3
    },
    
    "position": {
      "leverage": 6,
      "marginPercent": 10,
      "maxPositions": 2,
      "pyramiding": false
    },
    
    "stopLoss": {
      "percent": 4,
      "atrMultiplier": 1.2,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 2.5
    },
    
    "takeProfit": {
      "autoStartRetreat": 2.5,
      "profitRetreat": 22,
      "trailingStop": false,
      "partialTake": true,
      "partialPercent": 60,
      "partialTrigger": 3
    },
    
    "reverse": {
      "enabled": false,
      "lossRatio": 0,
      "profitRatio": 0
    },
    
    "risk": {
      "maxDailyLoss": 6,
      "maxDrawdown": 12,
      "pauseOnLoss": 2
    },
    
    "remark": "震荡市场谨慎操作，等待突破"
  }',
  '震荡区间高抛低吸，使用布林带+RSI+随机指标多重确认。支撑阻力位识别，分批止盈锁定利润。',
  1, 202
);

-- 【7】平衡型-高波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_balanced_high_vol', '⚖️ 平衡-波动捕捉',
  'balanced', 'high_vol',
  90, 180.00,
  4, 7,
  5.00, 10.00,
  6.00, 28.00, 4.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m", "5m"],
    "primaryTimeFrame": "1m",
    
    "entry": {
      "signalStrength": 75,
      "volatilityFilter": true,
      "minVolatility": 120,
      "maxVolatility": 250,
      "momentumStrength": 75,
      "volumeSpike": true,
      "spikeMultiplier": 1.8,
      "priceAction": true,
      "candlePattern": true
    },
    
    "position": {
      "leverage": 5,
      "marginPercent": 8,
      "maxPositions": 2,
      "pyramiding": false,
      "dynamicSize": true,
      "volatilityAdjust": true
    },
    
    "stopLoss": {
      "percent": 6,
      "atrMultiplier": 2,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 4,
      "widthAdjust": true
    },
    
    "takeProfit": {
      "autoStartRetreat": 4,
      "profitRetreat": 28,
      "trailingStop": true,
      "trailingDistance": 3,
      "partialTake": true,
      "partialPercent": 40,
      "partialTrigger": 5
    },
    
    "reverse": {
      "enabled": true,
      "lossRatio": 100,
      "profitRatio": 100,
      "cooldown": 90
    },
    
    "risk": {
      "maxDailyLoss": 10,
      "maxDrawdown": 18,
      "pauseOnLoss": 2,
      "highVolPause": false
    },
    
    "remark": "高波动需快速反应，启用移动止盈"
  }',
  '高波动市场机会与风险并存，动态调整仓位和止损。快速反应，移动止盈锁定利润。',
  1, 203
);

-- 【8】平衡型-低波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_balanced_low_vol', '⚖️ 平衡-突破等待',
  'balanced', 'low_vol',
  360, 40.00,
  6, 10,
  10.00, 18.00,
  3.00, 18.00, 2.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["15m", "30m", "1h", "4h"],
    "primaryTimeFrame": "1h",
    
    "entry": {
      "signalStrength": 70,
      "breakoutWait": true,
      "breakoutConfirmBars": 2,
      "volumeConfirmation": true,
      "volumeMultiplier": 2,
      "squeezeTrigger": true,
      "bollingerSqueeze": true,
      "keltnerChannel": true
    },
    
    "position": {
      "leverage": 8,
      "marginPercent": 15,
      "maxPositions": 2,
      "pyramiding": true,
      "maxPyramid": 3,
      "pyramidScale": 0.6
    },
    
    "stopLoss": {
      "percent": 3,
      "atrMultiplier": 1.2,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 2
    },
    
    "takeProfit": {
      "autoStartRetreat": 2,
      "profitRetreat": 18,
      "trailingStop": true,
      "trailingDistance": 1.5,
      "partialTake": false
    },
    
    "reverse": {
      "enabled": false,
      "lossRatio": 0,
      "profitRatio": 0
    },
    
    "risk": {
      "maxDailyLoss": 5,
      "maxDrawdown": 10,
      "pauseOnLoss": 2
    },
    
    "remark": "低波动加大仓位，耐心等待行情"
  }',
  '低波动市场等待突破，使用布林带挤压+肯特纳通道识别蓄力状态。突破后加大仓位，金字塔加仓。',
  1, 204
);

-- =========================================
-- 🚀 激进型策略 (Aggressive)
-- 特点：高杠杆、大仓位、适合专业用户
-- 日收益预期：3-10%（风险也相应增大）
-- =========================================

-- 【9】激进型-趋势市场
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_aggressive_trend', '🚀 激进-趋势冲锋',
  'aggressive', 'trend',
  180, 120.00,
  10, 20,
  10.00, 20.00,
  8.00, 20.00, 5.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m", "5m", "15m", "30m"],
    "primaryTimeFrame": "5m",
    
    "entry": {
      "signalStrength": 60,
      "macdCross": true,
      "macdHistogram": true,
      "rsiRange": [45, 55],
      "emaAlignment": true,
      "emaPeriods": [5, 13, 34],
      "volumeMultiplier": 1.5,
      "trendConfirmation": true,
      "multiTimeframeAgreement": 2,
      "momentumStrength": 70
    },
    
    "position": {
      "leverage": 15,
      "marginPercent": 18,
      "maxPositions": 3,
      "pyramiding": true,
      "maxPyramid": 3,
      "pyramidScale": 0.7
    },
    
    "stopLoss": {
      "percent": 8,
      "atrMultiplier": 2,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 5,
      "protectProfit": true
    },
    
    "takeProfit": {
      "autoStartRetreat": 5,
      "profitRetreat": 20,
      "trailingStop": true,
      "trailingDistance": 2.5,
      "partialTake": true,
      "partialPercent": 30,
      "partialTrigger": 8
    },
    
    "reverse": {
      "enabled": true,
      "lossRatio": 100,
      "profitRatio": 100,
      "cooldown": 30
    },
    
    "risk": {
      "maxDailyLoss": 15,
      "maxDrawdown": 25,
      "pauseOnLoss": 3
    },
    
    "ai": {
      "winProbabilityThreshold": 55,
      "marketStateCheck": true,
      "riskPreferenceCheck": false,
      "signalConfirmation": true
    },
    
    "remark": "趋势明确时重仓出击，追求高收益"
  }',
  '⚠️【高风险】趋势明确时高杠杆追涨。支持多级金字塔加仓，快速移动止盈锁定利润。仅限专业用户。',
  1, 301
);

-- 【10】激进型-震荡市场
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_aggressive_volatile', '🚀 激进-双向博弈',
  'aggressive', 'volatile',
  120, 80.00,
  8, 15,
  8.00, 16.00,
  6.00, 18.00, 4.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m", "5m", "15m"],
    "primaryTimeFrame": "5m",
    
    "entry": {
      "signalStrength": 55,
      "bollingerBand": true,
      "bollingerPeriod": 20,
      "bollingerDev": 2.5,
      "rsiRange": [20, 80],
      "rsiPeriod": 7,
      "stochasticCross": true,
      "priceAction": true,
      "candlePattern": true,
      "divergence": true
    },
    
    "position": {
      "leverage": 12,
      "marginPercent": 14,
      "maxPositions": 2,
      "pyramiding": false,
      "hedging": true
    },
    
    "stopLoss": {
      "percent": 6,
      "atrMultiplier": 1.5,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 4
    },
    
    "takeProfit": {
      "autoStartRetreat": 4,
      "profitRetreat": 18,
      "trailingStop": false,
      "partialTake": true,
      "partialPercent": 50,
      "partialTrigger": 5
    },
    
    "reverse": {
      "enabled": true,
      "lossRatio": 50,
      "profitRatio": 50,
      "cooldown": 45
    },
    
    "risk": {
      "maxDailyLoss": 12,
      "maxDrawdown": 20,
      "pauseOnLoss": 2
    },
    
    "remark": "震荡市场双向操作，频繁交易"
  }',
  '⚠️【高风险】震荡市场双向开单，支持对冲持仓。识别RSI背离寻找反转点。需要较强市场判断能力。',
  1, 302
);

-- 【11】激进型-高波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_aggressive_high_vol', '🚀 激进-极速博弈',
  'aggressive', 'high_vol',
  60, 200.00,
  8, 12,
  6.00, 12.00,
  10.00, 22.00, 6.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["1m"],
    "primaryTimeFrame": "1m",
    
    "entry": {
      "signalStrength": 70,
      "volatilityFilter": true,
      "minVolatility": 150,
      "maxVolatility": 300,
      "momentumStrength": 80,
      "volumeSpike": true,
      "spikeMultiplier": 2.5,
      "priceAction": true,
      "candlePattern": true,
      "quickEntry": true,
      "entryTimeout": 10
    },
    
    "position": {
      "leverage": 10,
      "marginPercent": 10,
      "maxPositions": 2,
      "pyramiding": false,
      "dynamicSize": true,
      "volatilityAdjust": true,
      "hedging": true
    },
    
    "stopLoss": {
      "percent": 10,
      "atrMultiplier": 2.5,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 6,
      "widthAdjust": true,
      "quickStop": true
    },
    
    "takeProfit": {
      "autoStartRetreat": 6,
      "profitRetreat": 22,
      "trailingStop": true,
      "trailingDistance": 3.5,
      "partialTake": true,
      "partialPercent": 50,
      "partialTrigger": 8
    },
    
    "reverse": {
      "enabled": true,
      "lossRatio": 100,
      "profitRatio": 100,
      "cooldown": 30,
      "quickReverse": true
    },
    
    "risk": {
      "maxDailyLoss": 20,
      "maxDrawdown": 30,
      "pauseOnLoss": 2,
      "highVolPause": false
    },
    
    "remark": "高波动双向博弈，极高风险极高收益"
  }',
  '⚠️【极高风险】高波动市场快进快出，10秒内入场决策。双向对冲，极速止损止盈。可能快速盈利也可能快速爆仓！',
  1, 303
);

-- 【12】激进型-低波动
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_v2_aggressive_low_vol', '🚀 激进-突破狙击',
  'aggressive', 'low_vol',
  300, 50.00,
  12, 20,
  15.00, 25.00,
  5.00, 15.00, 3.00,
  '{
    "version": "2.0",
    "exchange": ["binance", "bitget", "okx", "gate"],
    "symbol": "BTCUSDT",
    "orderType": "market",
    "marginMode": "isolated",
    
    "timeFrames": ["5m", "15m", "30m", "1h"],
    "primaryTimeFrame": "15m",
    
    "entry": {
      "signalStrength": 60,
      "breakoutWait": true,
      "breakoutConfirmBars": 1,
      "volumeConfirmation": true,
      "volumeMultiplier": 2.5,
      "squeezeTrigger": true,
      "bollingerSqueeze": true,
      "keltnerChannel": true,
      "breakoutStrength": 80,
      "fakeoutFilter": true
    },
    
    "position": {
      "leverage": 18,
      "marginPercent": 22,
      "maxPositions": 3,
      "pyramiding": true,
      "maxPyramid": 4,
      "pyramidScale": 0.8,
      "scaleInOnBreakout": true
    },
    
    "stopLoss": {
      "percent": 5,
      "atrMultiplier": 1.5,
      "useAtrStop": true,
      "moveToBreakeven": true,
      "breakevenTrigger": 3,
      "protectProfit": true
    },
    
    "takeProfit": {
      "autoStartRetreat": 3,
      "profitRetreat": 15,
      "trailingStop": true,
      "trailingDistance": 2,
      "partialTake": true,
      "partialPercent": 25,
      "partialTrigger": 6
    },
    
    "reverse": {
      "enabled": false,
      "lossRatio": 0,
      "profitRatio": 0
    },
    
    "risk": {
      "maxDailyLoss": 15,
      "maxDrawdown": 25,
      "pauseOnLoss": 2
    },
    
    "remark": "低波动时重仓等待突破，博取大行情"
  }',
  '⚠️【高风险】低波动时布局等待大行情突破。超高杠杆+4级金字塔加仓。过滤假突破信号。',
  1, 304
);

-- ============================================================
-- 5. 验证插入结果
-- ============================================================

SELECT 
  g.id AS group_id,
  g.group_name,
  g.is_official AS '官方',
  g.exchange AS '交易平台',
  g.symbol AS '交易对',
  COUNT(s.id) AS '策略数'
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key = 'official_btc_usdt_v2'
GROUP BY g.id;

-- 显示所有策略的详细参数
SELECT 
  strategy_name AS '策略名称',
  risk_preference AS '风险偏好',
  market_state AS '市场状态',
  monitor_window AS '时间窗口(秒)',
  volatility_threshold AS '波动阈值',
  CONCAT(leverage_min, '-', leverage_max, 'x') AS '杠杆范围',
  CONCAT(margin_percent_min, '-', margin_percent_max, '%') AS '仓位范围',
  CONCAT(stop_loss_percent, '%') AS '止损',
  CONCAT(auto_start_retreat_percent, '%') AS '启动回撤',
  CONCAT(profit_retreat_percent, '%') AS '止盈回撤',
  IF(JSON_EXTRACT(config_json, '$.reverse.enabled') = true, '✅', '❌') AS '反向单'
FROM hg_trading_strategy_template
WHERE group_id = @group_id
ORDER BY sort;

-- ============================================================
-- 完成！
-- ============================================================

