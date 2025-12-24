-- 创建BTCUSDT和ETHUSDT的波动率配置
-- 执行前请确保表 toogo_volatility_config 存在

-- BTCUSDT配置（比特币，市值最大，流动性最好，波动相对稳定）
INSERT INTO `toogo_volatility_config` (
  `symbol`,
  `high_volatility_threshold`,
  `low_volatility_threshold`,
  `trend_strength_threshold`,
  `range_volatility_threshold`,
  `weight_1m`,
  `weight_5m`,
  `weight_15m`,
  `weight_30m`,
  `weight_1h`,
  `is_active`,
  `created_at`,
  `updated_at`
) VALUES (
  'BTCUSDT',           -- 交易对
  2.0,                 -- 高波动阈值：2.0%（比特币波动相对稳定）
  0.4,                 -- 低波动阈值：0.4%（比特币波动较小）
  0.35,                -- 趋势阈值：0.35
  0.0,                 -- 震荡市场波动率阈值（暂不使用）
  0.10,                -- 1分钟权重：10%
  0.20,                -- 5分钟权重：20%
  0.25,                -- 15分钟权重：25%
  0.25,                -- 30分钟权重：25%
  0.20,                -- 1小时权重：20%
  1,                   -- 启用状态：1=启用
  NOW(),               -- 创建时间
  NOW()                -- 更新时间
) ON DUPLICATE KEY UPDATE
  `high_volatility_threshold` = 2.0,
  `low_volatility_threshold` = 0.4,
  `trend_strength_threshold` = 0.35,
  `weight_1m` = 0.10,
  `weight_5m` = 0.20,
  `weight_15m` = 0.25,
  `weight_30m` = 0.25,
  `weight_1h` = 0.20,
  `is_active` = 1,
  `updated_at` = NOW();

-- ETHUSDT配置（以太坊，市值第二，流动性好，波动可能比BTC稍大）
INSERT INTO `toogo_volatility_config` (
  `symbol`,
  `high_volatility_threshold`,
  `low_volatility_threshold`,
  `trend_strength_threshold`,
  `range_volatility_threshold`,
  `weight_1m`,
  `weight_5m`,
  `weight_15m`,
  `weight_30m`,
  `weight_1h`,
  `is_active`,
  `created_at`,
  `updated_at`
) VALUES (
  'ETHUSDT',           -- 交易对
  2.5,                 -- 高波动阈值：2.5%（以太坊波动稍大）
  0.5,                 -- 低波动阈值：0.5%（以太坊波动稍大）
  0.35,                -- 趋势阈值：0.35
  0.0,                 -- 震荡市场波动率阈值（暂不使用）
  0.10,                -- 1分钟权重：10%
  0.20,                -- 5分钟权重：20%
  0.25,                -- 15分钟权重：25%
  0.25,                -- 30分钟权重：25%
  0.20,                -- 1小时权重：20%
  1,                   -- 启用状态：1=启用
  NOW(),               -- 创建时间
  NOW()                -- 更新时间
) ON DUPLICATE KEY UPDATE
  `high_volatility_threshold` = 2.5,
  `low_volatility_threshold` = 0.5,
  `trend_strength_threshold` = 0.35,
  `weight_1m` = 0.10,
  `weight_5m` = 0.20,
  `weight_15m` = 0.25,
  `weight_30m` = 0.25,
  `weight_1h` = 0.20,
  `is_active` = 1,
  `updated_at` = NOW();

-- 配置说明
-- BTCUSDT: 高波动2.0%, 低波动0.4% - 适合市值最大、流动性最好、波动相对稳定的主流币种
-- ETHUSDT: 高波动2.5%, 低波动0.5% - 适合市值第二、流动性好、波动可能比BTC稍大的主流币种

