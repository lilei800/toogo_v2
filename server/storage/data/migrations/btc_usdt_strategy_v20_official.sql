-- ============================================================
-- BTC-USDT 官方策略组合 V20（稳健回撤版）
-- 交易所：Bitget
-- 交易对：BTCUSDT
-- 订单类型：市价单
-- 保证金模式：逐仓
--
-- 设计目标：
-- - 行情不好（震荡/高波动）尽量减少亏损“幅度”和“频率”（更高阈值、更低杠杆/更低保证金占比、更早锁盈）
-- - 行情好（趋势）提高盈利概率与盈亏比（允许利润奔跑：更高回撤阈值、更高杠杆/更低保证金占比）
--
-- 注意：无法承诺“必然盈利”。该模板组提供一个可回测、可迭代的基线。
-- 推荐默认映射（创建机器人 remark 的 marketRiskMapping）：
--   trend    -> aggressive
--   volatile -> balanced
--   high_vol -> conservative
--   low_vol  -> balanced
-- ============================================================

-- 1. 创建策略组（如果已存在则更新）
INSERT INTO `hg_trading_strategy_group` (
    `group_name`,
    `group_key`,
    `exchange`,
    `symbol`,
    `order_type`,
    `margin_mode`,
    `is_official`,
    `is_default`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    'BTC-USDT 官方策略组合 V20（稳健回撤版）',
    'btc_usdt_v20_official',
    'bitget',
    'BTCUSDT',
    'market',
    'isolated',
    1,
    0,
    '官方V20：12套策略模板（4种市场状态×3种风险偏好）。偏向稳健与降回撤：震荡/高波动降频降杠杆，趋势允许利润奔跑。推荐映射：trend->aggressive, volatile->balanced, high_vol->conservative, low_vol->balanced。',
    1,
    20,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    `group_name` = VALUES(`group_name`),
    `exchange` = VALUES(`exchange`),
    `symbol` = VALUES(`symbol`),
    `order_type` = VALUES(`order_type`),
    `margin_mode` = VALUES(`margin_mode`),
    `is_official` = VALUES(`is_official`),
    `is_default` = VALUES(`is_default`),
    `description` = VALUES(`description`),
    `is_active` = VALUES(`is_active`),
    `sort` = VALUES(`sort`),
    `updated_at` = NOW();

-- 获取策略组ID
SET @groupId = (SELECT id FROM hg_trading_strategy_group WHERE group_key = 'btc_usdt_v20_official' LIMIT 1);

-- 删除旧模板（如果存在）
DELETE FROM `hg_trading_strategy_template` WHERE `group_id` = @groupId;
DELETE FROM `hg_trading_strategy_template` WHERE `strategy_key` LIKE '%_v20';

-- ============================================================
-- 2. 创建12种策略模板（market_state 使用引擎规范化格式：trend/volatile/high_vol/low_vol）
-- ============================================================

-- ========== 低波动（low_vol）==========
-- 低波动：信号阈值较小但窗口偏长，避免过于频繁的“来回打脸”

INSERT INTO `hg_trading_strategy_template`
(`group_id`,`strategy_key`,`strategy_name`,`risk_preference`,`market_state`,`monitor_window`,`volatility_threshold`,`leverage`,`margin_percent`,`stop_loss_percent`,`profit_retreat_percent`,`auto_start_retreat_percent`,`description`,`is_active`,`sort`,`created_at`,`updated_at`)
VALUES
(@groupId,'conservative_low_vol_v20','低波动-保守','conservative','low_vol',300,40.0,6,12.0,2.0,1.8,3.0,'低波动保守：低杠杆+中保证金，严格止损与较早锁盈，尽量降低回撤',1,1,NOW(),NOW()),
(@groupId,'balanced_low_vol_v20','低波动-平衡','balanced','low_vol',300,40.0,8,10.0,2.2,2.0,2.8,'低波动平衡：适中杠杆与保证金，兼顾胜率与收益',1,2,NOW(),NOW()),
(@groupId,'aggressive_low_vol_v20','低波动-激进','aggressive','low_vol',240,35.0,12,8.0,2.4,2.2,2.5,'低波动激进：提高杠杆以放大收益，但保证金占比下调控制风险',1,3,NOW(),NOW());

-- ========== 震荡（volatile）==========
-- 震荡：提高阈值降噪，回撤阈值偏紧，尽快锁住小利润

INSERT INTO `hg_trading_strategy_template`
(`group_id`,`strategy_key`,`strategy_name`,`risk_preference`,`market_state`,`monitor_window`,`volatility_threshold`,`leverage`,`margin_percent`,`stop_loss_percent`,`profit_retreat_percent`,`auto_start_retreat_percent`,`description`,`is_active`,`sort`,`created_at`,`updated_at`)
VALUES
(@groupId,'conservative_volatile_v20','震荡-保守','conservative','volatile',240,80.0,4,8.0,2.6,2.0,3.8,'震荡保守：降频+降杠杆，回撤较紧，目标是减少亏损次数',1,4,NOW(),NOW()),
(@groupId,'balanced_volatile_v20','震荡-平衡','balanced','volatile',240,80.0,6,10.0,2.8,2.3,3.5,'震荡平衡：控制回撤的同时保持一定出手机会',1,5,NOW(),NOW()),
(@groupId,'aggressive_volatile_v20','震荡-激进','aggressive','volatile',180,70.0,10,12.0,3.0,2.6,3.2,'震荡激进：在可控风险下提高收益，但仍以较高阈值抑制噪声',1,6,NOW(),NOW());

-- ========== 趋势（trend）==========
-- 趋势：允许利润奔跑（更高回撤阈值），启动止盈稍高，避免过早被震出

INSERT INTO `hg_trading_strategy_template`
(`group_id`,`strategy_key`,`strategy_name`,`risk_preference`,`market_state`,`monitor_window`,`volatility_threshold`,`leverage`,`margin_percent`,`stop_loss_percent`,`profit_retreat_percent`,`auto_start_retreat_percent`,`description`,`is_active`,`sort`,`created_at`,`updated_at`)
VALUES
(@groupId,'conservative_trend_v20','趋势-保守','conservative','trend',300,110.0,6,10.0,2.5,4.0,4.0,'趋势保守：低杠杆稳健跟随，允许利润回撤更大以捕捉趋势',1,7,NOW(),NOW()),
(@groupId,'balanced_trend_v20','趋势-平衡','balanced','trend',300,110.0,10,10.0,2.8,4.5,3.8,'趋势平衡：风险与收益平衡，利润回撤阈值提高以提升盈利空间',1,8,NOW(),NOW()),
(@groupId,'aggressive_trend_v20','趋势-激进','aggressive','trend',240,100.0,15,8.0,3.0,6.0,3.5,'趋势激进：趋势行情放大利润，使用更高杠杆但降低保证金占比控制风险暴露',1,9,NOW(),NOW());

-- ========== 高波动（high_vol）==========
-- 高波动：显著提高阈值+降低杠杆+降低保证金占比，尽量避免高频被扫

INSERT INTO `hg_trading_strategy_template`
(`group_id`,`strategy_key`,`strategy_name`,`risk_preference`,`market_state`,`monitor_window`,`volatility_threshold`,`leverage`,`margin_percent`,`stop_loss_percent`,`profit_retreat_percent`,`auto_start_retreat_percent`,`description`,`is_active`,`sort`,`created_at`,`updated_at`)
VALUES
(@groupId,'conservative_high_vol_v20','高波动-保守','conservative','high_vol',180,160.0,3,6.0,3.0,2.0,4.5,'高波动保守：最小化风险暴露，减少下单频率，尽量保命',1,10,NOW(),NOW()),
(@groupId,'balanced_high_vol_v20','高波动-平衡','balanced','high_vol',180,160.0,5,8.0,3.2,2.2,4.0,'高波动平衡：仍然严格控风险，适度参与大波动机会',1,11,NOW(),NOW()),
(@groupId,'aggressive_high_vol_v20','高波动-激进','aggressive','high_vol',120,140.0,8,10.0,3.5,2.5,3.5,'高波动激进：仅适合强风控偏好用户，高阈值减少噪声，但仍可能高回撤',1,12,NOW(),NOW());

-- ============================================================
-- 3. 验证输出
-- ============================================================
SELECT
  g.id AS group_id, g.group_key, g.group_name, g.is_official, g.is_default, g.is_active,
  t.market_state, t.risk_preference,
  t.monitor_window, t.volatility_threshold,
  t.leverage, t.margin_percent,
  t.stop_loss_percent, t.auto_start_retreat_percent, t.profit_retreat_percent,
  t.strategy_key
FROM hg_trading_strategy_group g
JOIN hg_trading_strategy_template t ON t.group_id = g.id
WHERE g.group_key = 'btc_usdt_v20_official'
ORDER BY FIELD(t.market_state,'low_vol','volatile','trend','high_vol'), FIELD(t.risk_preference,'conservative','balanced','aggressive');











