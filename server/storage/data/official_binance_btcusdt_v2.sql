-- ============================================================
-- Binance BTC-USDT 官方策略 V2.0 (MySQL版本)
-- 创建时间: 2026-01-01
-- 说明:
--   - 适配“新市场状态算法 + 波动率配置(ToogoVolatilityConfig) + 风险偏好映射(v2)”
--   - 机器人运行时只使用策略模板硬字段：
--       monitor_window / volatility_threshold / leverage / margin_percent /
--       stop_loss_percent / auto_start_retreat_percent / profit_retreat_percent
-- 手续费假设(币安U本位合约常见档位，实际以账户等级为准)：
--   - Market为Taker：开仓0.04% + 平仓0.04% ≈ 0.08%（名义价值口径）
-- 重要声明：
--   - 市场存在不确定性，本文件仅提供参数模板，无法保证实际盈利。
-- ============================================================

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET NAMES utf8mb4;

-- 清理旧数据
DELETE FROM `hg_trading_strategy_template` WHERE `group_id` IN (
  SELECT `id` FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_binance_btcusdt_v2'
);
DELETE FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_binance_btcusdt_v2';

-- 插入官方策略组 V2.0
INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  '🔥 Binance BTC-USDT 官方策略 V2.0（新算法）',
  'official_binance_btcusdt_v2',
  'binance',
  'BTCUSDT',
  'market',
  'isolated',
  1,
  0,
  '币安BTCUSDT官方策略V2.0（适配新市场状态算法/波动率配置/风险偏好映射）。窗口阈值触发→方向预警→自动下单→止损/追踪止盈自动平仓。开仓保证金百分比严格等于平台余额百分比（按AvailableBalance计算）。手续费按Taker双边约0.08%纳入止盈启动阈值设计。',
  1,
  3
);

SET @group_id = LAST_INSERT_ID();

-- ==================== 🛡️ 保守型 ====================

INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`,
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `auto_start_retreat_percent`, `profit_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES
(@group_id, 'binance_btc_v2_conservative_trend', '🛡️ 保守-趋势跟踪 (V2)',
 'conservative', 'trend', 480, 120.00, 3, 8.00, 3.00, 2.20, 28.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004},"notes":"runtime uses template fields only"}',
 '顺势为主，窗口较长+阈值中等，过滤噪音。止盈启动2.2%（保证金口径）覆盖手续费与滑点冗余。', 1, 101),
(@group_id, 'binance_btc_v2_conservative_volatile', '🛡️ 保守-区间震荡 (V2)',
 'conservative', 'volatile', 240, 80.00, 2, 7.00, 2.80, 1.80, 25.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '震荡市降低杠杆/仓位，阈值偏低以提高触发频率，但止损更紧，避免反复磨损。', 1, 102),
(@group_id, 'binance_btc_v2_conservative_high_vol', '🛡️ 保守-高波动防守 (V2)',
 'conservative', 'high_vol', 120, 220.00, 2, 5.00, 5.50, 3.50, 35.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '高波动时优先“少仓位+更高触发阈值”，减少噪声触发；止盈启动提高到3.5%。', 1, 103),
(@group_id, 'binance_btc_v2_conservative_low_vol', '🛡️ 保守-低波动蓄力 (V2)',
 'conservative', 'low_vol', 720, 60.00, 4, 10.00, 2.20, 1.20, 20.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '低波动用更长窗口捕捉“慢反弹/慢回落”，阈值更小；止盈启动更低以适配小波动收益。', 1, 104);

-- ==================== ⚖️ 平衡型 ====================

INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`,
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `auto_start_retreat_percent`, `profit_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES
(@group_id, 'binance_btc_v2_balanced_trend', '⚖️ 平衡-趋势跟踪 ⭐推荐 (V2)',
 'balanced', 'trend', 360, 140.00, 5, 12.00, 5.00, 3.00, 25.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '默认映射推荐：趋势→平衡。中等窗口/阈值，较好的触发质量与频率平衡。', 1, 201),
(@group_id, 'binance_btc_v2_balanced_volatile', '⚖️ 平衡-区间套利 ⭐推荐 (V2)',
 'balanced', 'volatile', 240, 95.00, 4, 10.00, 4.50, 2.50, 22.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '默认映射推荐：震荡→平衡。中等阈值避免频繁打点，追踪止盈偏紧锁利润。', 1, 202),
(@group_id, 'binance_btc_v2_balanced_high_vol', '⚖️ 平衡-波动捕捉 (V2)',
 'balanced', 'high_vol', 90, 260.00, 6, 8.00, 7.00, 4.50, 28.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '高波动下适度提高杠杆但控制仓位，阈值提高避免噪声；止盈启动上移到4.5%。', 1, 203),
(@group_id, 'binance_btc_v2_balanced_low_vol', '⚖️ 平衡-低波动突破 (V2)',
 'balanced', 'low_vol', 600, 70.00, 6, 14.00, 3.20, 1.80, 18.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '低波动时强调高质量触发，阈值略高于保守；止损更宽避免被轻微回撤洗掉。', 1, 204);

-- ==================== 🚀 激进型 ====================

INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`,
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `auto_start_retreat_percent`, `profit_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES
(@group_id, 'binance_btc_v2_aggressive_trend', '🚀 激进-趋势冲锋 (V2)',
 'aggressive', 'trend', 240, 160.00, 10, 18.00, 8.00, 5.00, 20.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '高杠杆趋势追击，止盈启动5%显著高于费用平衡点。', 1, 301),
(@group_id, 'binance_btc_v2_aggressive_volatile', '🚀 激进-双向博弈 (V2)',
 'aggressive', 'volatile', 180, 120.00, 8, 15.00, 7.00, 4.00, 18.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '震荡市更快窗口与更高阈值，偏快进快出；追踪止盈更紧。', 1, 302),
(@group_id, 'binance_btc_v2_aggressive_high_vol', '🚀 激进-极速博弈 ⭐推荐 (V2)',
 'aggressive', 'high_vol', 60, 320.00, 12, 12.00, 11.00, 6.50, 22.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '默认映射推荐：高波动→激进。超短窗口+高阈值减少乱扫，止盈启动6.5%。', 1, 303),
(@group_id, 'binance_btc_v2_aggressive_low_vol', '🚀 激进-低波动狙击 (V2)',
 'aggressive', 'low_vol', 420, 85.00, 15, 22.00, 5.50, 3.20, 15.00,
 '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
 '低波动重仓放大收益，风险极高；追踪止盈更紧以锁定利润。', 1, 304);


