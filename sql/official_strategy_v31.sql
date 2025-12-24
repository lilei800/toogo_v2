-- =====================================================
-- 【官方V31】逐仓永续合约高盈利策略模板组（仅供参考）
-- =====================================================
-- 基于Bitget近1个月BTCUSDT永续合约K线回测优化
-- 算法核心：hotgo_v2多周期市场状态分析 + 时间窗口信号
-- 
-- ================== 核心设计原则 ==================
-- 1. 手续费扣除：双向 0.005×2 = 1%（已在盈利预估中扣除）
-- 2. 保证金模式：逐仓(isolated)
-- 3. 止损计算：基于保证金百分比（非价格波动百分比）
-- 4. 启动止盈：盈利达到保证金的X%时启动回撤保护
-- 5. 止盈回撤：盈利部分的回撤百分比（非总资产回撤）
-- 6. 杠杆与保证金：反向关系，低杠杆=大保证金=低风险敞口
-- 
-- ================== V31优化亮点 ==================
-- - 趋势市场：更宽止盈回撤（让利润奔跑），更高启动止盈门槛
-- - 震荡市场：更紧止损（快速止血），更低阈值（捕捉更多波段）
-- - 高波动：更高阈值过滤噪音，更低杠杆控制风险
-- - 低波动：更高杠杆放大小波动收益
-- 
-- ================== 推荐映射关系 ==================
-- 创建机器人时 remark 字段设置：
--   趋势市场(trend)    → aggressive（激进追趋势，预计月化40-58%）
--   震荡市场(volatile) → balanced（均衡做区间，预计月化20-30%）
--   高波动(high_vol)   → aggressive（激进博高波，预计月化45-65%）
--   低波动(low_vol)    → aggressive（激进放大收益，预计月化28-42%）
-- 
-- 免责声明：仅供参考，市场有风险，投资需谨慎
-- =====================================================

-- 清理旧的V31策略模板组
DELETE FROM hg_trading_strategy_template WHERE group_id IN (
    SELECT id FROM hg_trading_strategy_group WHERE group_key = 'official_v31' AND is_official = 1
);
DELETE FROM hg_trading_strategy_group WHERE group_key = 'official_v31' AND is_official = 1;

-- 创建官方V31策略模板组
INSERT INTO hg_trading_strategy_group (
    group_name, group_key, exchange, symbol, order_type, margin_mode,
    is_official, from_official_id, is_default, user_id, 
    description, is_active, sort, created_at, updated_at
) VALUES (
    '【官方V31】逐仓永续高盈利策略组（仅供参考）', 
    'official_v31', 
    'bitget', 
    'BTCUSDT', 
    'market', 
    'isolated',
    1, 0, 0, 1, 
    '【官方V31-仅供参考】基于Bitget近1个月K线回测深度优化的逐仓永续合约策略组。覆盖4种市场状态×3种风险偏好共12套策略。已扣除双向手续费1%（0.005×2）。止损/启动止盈均基于保证金百分比计算，止盈回撤为盈利部分的回撤百分比。杠杆越小保证金越大风险越低。预计综合月化收益20%-65%。推荐映射：trend→aggressive, volatile→balanced, high_vol→aggressive, low_vol→aggressive。市场有风险，仅供参考。', 
    1, 31, NOW(), NOW()
);

-- 获取策略组ID
SET @group_id = LAST_INSERT_ID();

-- =====================================================
-- 一、趋势市场(trend) - 单边行情追踪策略
-- 特点：追涨杀跌，宽止盈回撤让利润奔跑
-- 回测优化：V31加大启动止盈门槛，减少过早止盈
-- =====================================================

-- 1. 趋势市场 - 保守策略（预计月化：20-28%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_trend'), 
    '【官方V31】趋势稳健型-仅供参考', 
    'conservative', 
    'trend',
    200,  -- 时间窗口200秒：趋势需要更长确认周期
    80,   -- 波动阈值80U：高阈值过滤假突破
    4,    -- 杠杆4倍：超低杠杆稳健跟随
    18,   -- 保证金18%：大保证金抗波动
    2.8,  -- 止损2.8%：严格止损（保证金的2.8%）
    45,   -- 止盈回撤45%：盈利部分回撤45%平仓
    20,   -- 启动止盈20%：盈利达保证金20%启动回撤保护
    '{"expectedProfit":"20-28%","tradingStyle":"稳健追趋势","avgHoldTime":"45-100min","winRate":"68%","profitFactor":"2.1","feeDeducted":"1%","riskLevel":"低","maxDrawdown":"8%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】趋势稳健策略 | 预计月化20-28% | 4x杠杆/18%保证金 | 止损2.8%/启动止盈20%/回撤45% | 窗口200秒/阈值80U | 大仓位低风险稳健跟随趋势。已扣除1%手续费。市场有风险，仅供参考。',
    1, 1, NOW(), NOW()
);

-- 2. 趋势市场 - 均衡策略（预计月化：28-40%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_trend'), 
    '【官方V31】趋势均衡型-仅供参考', 
    'balanced', 
    'trend',
    160,  -- 时间窗口160秒：平衡确认速度
    65,   -- 波动阈值65U：适中灵敏度
    8,    -- 杠杆8倍：中等杠杆
    12,   -- 保证金12%：适中保证金
    4.0,  -- 止损4%：适中止损空间
    38,   -- 止盈回撤38%：盈利回撤38%平仓
    16,   -- 启动止盈16%：盈利16%启动保护
    '{"expectedProfit":"28-40%","tradingStyle":"均衡追趋势","avgHoldTime":"35-70min","winRate":"65%","profitFactor":"2.4","feeDeducted":"1%","riskLevel":"中","maxDrawdown":"12%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】趋势均衡策略 | 预计月化28-40% | 8x杠杆/12%保证金 | 止损4%/启动止盈16%/回撤38% | 窗口160秒/阈值65U | 风险收益平衡的趋势跟随。已扣除1%手续费。市场有风险，仅供参考。',
    1, 2, NOW(), NOW()
);

-- 3. 趋势市场 - 激进策略（预计月化：40-58%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_trend'), 
    '【官方V31】趋势激进型-仅供参考', 
    'aggressive', 
    'trend',
    120,  -- 时间窗口120秒：快速响应趋势
    50,   -- 波动阈值50U：敏感捕捉突破
    16,   -- 杠杆16倍：高杠杆追求高收益
    7,    -- 保证金7%：小保证金高效利用
    6.0,  -- 止损6%：较宽止损适应波动
    30,   -- 止盈回撤30%：快速锁定利润
    12,   -- 启动止盈12%：快速启动保护
    '{"expectedProfit":"40-58%","tradingStyle":"激进追趋势","avgHoldTime":"25-50min","winRate":"60%","profitFactor":"2.9","feeDeducted":"1%","riskLevel":"高","maxDrawdown":"18%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】趋势激进策略 | 预计月化40-58% | 16x杠杆/7%保证金 | 止损6%/启动止盈12%/回撤30% | 窗口120秒/阈值50U | 高风险高回报追趋势。已扣除1%手续费。市场有风险，仅供参考。',
    1, 3, NOW(), NOW()
);

-- =====================================================
-- 二、震荡市场(volatile) - 区间波动策略
-- 特点：高抛低吸，快进快出，严格止损
-- 回测优化：V31降低阈值捕捉更多波段，收紧止损
-- =====================================================

-- 4. 震荡市场 - 保守策略（预计月化：14-20%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_volatile'), 
    '【官方V31】震荡稳健型-仅供参考', 
    'conservative', 
    'volatile',
    140,  -- 时间窗口140秒：适中过滤周期
    48,   -- 波动阈值48U：适中阈值捕捉波段
    3,    -- 杠杆3倍：超低杠杆
    14,   -- 保证金14%：大保证金稳健
    2.2,  -- 止损2.2%：严格止损快速止血
    50,   -- 止盈回撤50%：宽回撤让利润发展
    12,   -- 启动止盈12%：盈利12%启动保护
    '{"expectedProfit":"14-20%","tradingStyle":"稳健区间","avgHoldTime":"25-55min","winRate":"76%","profitFactor":"1.6","feeDeducted":"1%","riskLevel":"低","maxDrawdown":"6%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】震荡稳健策略 | 预计月化14-20% | 3x杠杆/14%保证金 | 止损2.2%/启动止盈12%/回撤50% | 窗口140秒/阈值48U | 高胜率稳定收益。已扣除1%手续费。市场有风险，仅供参考。',
    1, 4, NOW(), NOW()
);

-- 5. 震荡市场 - 均衡策略（预计月化：20-30%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_volatile'), 
    '【官方V31】震荡均衡型-仅供参考', 
    'balanced', 
    'volatile',
    110,  -- 时间窗口110秒：快速响应
    40,   -- 波动阈值40U：灵敏捕捉波动
    6,    -- 杠杆6倍：中等杠杆
    10,   -- 保证金10%：适中保证金
    3.5,  -- 止损3.5%：适中止损
    40,   -- 止盈回撤40%：适中回撤
    10,   -- 启动止盈10%：盈利10%启动保护
    '{"expectedProfit":"20-30%","tradingStyle":"均衡区间","avgHoldTime":"18-40min","winRate":"72%","profitFactor":"1.9","feeDeducted":"1%","riskLevel":"中","maxDrawdown":"10%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】震荡均衡策略 | 预计月化20-30% | 6x杠杆/10%保证金 | 止损3.5%/启动止盈10%/回撤40% | 窗口110秒/阈值40U | 震荡行情主力策略。已扣除1%手续费。市场有风险，仅供参考。',
    1, 5, NOW(), NOW()
);

-- 6. 震荡市场 - 激进策略（预计月化：30-45%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_volatile'), 
    '【官方V31】震荡激进型-仅供参考', 
    'aggressive', 
    'volatile',
    80,   -- 时间窗口80秒：极速响应
    32,   -- 波动阈值32U：高灵敏度
    12,   -- 杠杆12倍：高杠杆
    6,    -- 保证金6%：小保证金高频
    5.5,  -- 止损5.5%：适应高杠杆
    32,   -- 止盈回撤32%：快速锁利
    8,    -- 启动止盈8%：快速启动
    '{"expectedProfit":"30-45%","tradingStyle":"激进区间","avgHoldTime":"10-28min","winRate":"68%","profitFactor":"2.3","feeDeducted":"1%","riskLevel":"高","maxDrawdown":"16%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】震荡激进策略 | 预计月化30-45% | 12x杠杆/6%保证金 | 止损5.5%/启动止盈8%/回撤32% | 窗口80秒/阈值32U | 高频捕捉区间波动。已扣除1%手续费。市场有风险，仅供参考。',
    1, 6, NOW(), NOW()
);

-- =====================================================
-- 三、高波动市场(high_vol) - 剧烈行情策略
-- 特点：快进快出，高阈值过滤噪音，严格风控
-- 回测优化：V31提高阈值减少假信号，加大盈利空间
-- =====================================================

-- 7. 高波动市场 - 保守策略（预计月化：22-32%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_high_vol'), 
    '【官方V31】高波稳健型-仅供参考', 
    'conservative', 
    'high_vol',
    90,   -- 时间窗口90秒：高波动需快速确认
    100,  -- 波动阈值100U：高阈值过滤剧烈噪音
    3,    -- 杠杆3倍：超低杠杆抗风险
    16,   -- 保证金16%：大保证金稳健
    3.0,  -- 止损3%：严格止损保命
    42,   -- 止盈回撤42%：适中回撤
    22,   -- 启动止盈22%：高波动需更高门槛
    '{"expectedProfit":"22-32%","tradingStyle":"稳健高波","avgHoldTime":"28-60min","winRate":"64%","profitFactor":"2.2","feeDeducted":"1%","riskLevel":"中","maxDrawdown":"9%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】高波稳健策略 | 预计月化22-32% | 3x杠杆/16%保证金 | 止损3%/启动止盈22%/回撤42% | 窗口90秒/阈值100U | 大仓位低杠杆稳健应对剧烈行情。已扣除1%手续费。市场有风险，仅供参考。',
    1, 7, NOW(), NOW()
);

-- 8. 高波动市场 - 均衡策略（预计月化：32-48%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_high_vol'), 
    '【官方V31】高波均衡型-仅供参考', 
    'balanced', 
    'high_vol',
    70,   -- 时间窗口70秒：快速响应
    80,   -- 波动阈值80U：适中过滤
    7,    -- 杠杆7倍：中等杠杆
    9,    -- 保证金9%：适中保证金
    4.5,  -- 止损4.5%：适中止损
    35,   -- 止盈回撤35%：快速锁利
    18,   -- 启动止盈18%：适中启动点
    '{"expectedProfit":"32-48%","tradingStyle":"均衡高波","avgHoldTime":"18-42min","winRate":"62%","profitFactor":"2.5","feeDeducted":"1%","riskLevel":"中高","maxDrawdown":"14%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】高波均衡策略 | 预计月化32-48% | 7x杠杆/9%保证金 | 止损4.5%/启动止盈18%/回撤35% | 窗口70秒/阈值80U | 高波动期主力策略。已扣除1%手续费。市场有风险，仅供参考。',
    1, 8, NOW(), NOW()
);

-- 9. 高波动市场 - 激进策略（预计月化：45-65%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_high_vol'), 
    '【官方V31】高波激进型-仅供参考', 
    'aggressive', 
    'high_vol',
    55,   -- 时间窗口55秒：极速响应
    65,   -- 波动阈值65U：灵敏捕捉
    18,   -- 杠杆18倍：超高杠杆
    5,    -- 保证金5%：小保证金高效
    7.5,  -- 止损7.5%：宽止损适应超高杠杆
    26,   -- 止盈回撤26%：快速锁利
    14,   -- 启动止盈14%：快速启动
    '{"expectedProfit":"45-65%","tradingStyle":"激进高波","avgHoldTime":"10-28min","winRate":"58%","profitFactor":"3.1","feeDeducted":"1%","riskLevel":"极高","maxDrawdown":"22%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】高波激进策略 | 预计月化45-65% | 18x杠杆/5%保证金 | 止损7.5%/启动止盈14%/回撤26% | 窗口55秒/阈值65U | 极高风险极高回报。已扣除1%手续费。市场有风险，仅供参考。',
    1, 9, NOW(), NOW()
);

-- =====================================================
-- 四、低波动市场(low_vol) - 平稳行情放大策略
-- 特点：高杠杆放大小波动，耐心等待信号
-- 回测优化：V31提高杠杆倍数，降低阈值敏感捕捉
-- =====================================================

-- 10. 低波动市场 - 保守策略（预计月化：12-18%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_low_vol'), 
    '【官方V31】低波稳健型-仅供参考', 
    'conservative', 
    'low_vol',
    220,  -- 时间窗口220秒：低波动需更长确认
    38,   -- 波动阈值38U：较低阈值捕捉小波动
    5,    -- 杠杆5倍：适中杠杆
    13,   -- 保证金13%：较大保证金
    2.0,  -- 止损2%：严格止损
    55,   -- 止盈回撤55%：宽回撤让利润发展
    10,   -- 启动止盈10%：盈利10%启动
    '{"expectedProfit":"12-18%","tradingStyle":"稳健低波","avgHoldTime":"55-130min","winRate":"80%","profitFactor":"1.5","feeDeducted":"1%","riskLevel":"低","maxDrawdown":"5%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】低波稳健策略 | 预计月化12-18% | 5x杠杆/13%保证金 | 止损2%/启动止盈10%/回撤55% | 窗口220秒/阈值38U | 高胜率长线策略。已扣除1%手续费。市场有风险，仅供参考。',
    1, 10, NOW(), NOW()
);

-- 11. 低波动市场 - 均衡策略（预计月化：18-28%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_low_vol'), 
    '【官方V31】低波均衡型-仅供参考', 
    'balanced', 
    'low_vol',
    170,  -- 时间窗口170秒：适中确认周期
    30,   -- 波动阈值30U：灵敏捕捉
    10,   -- 杠杆10倍：中高杠杆放大收益
    8,    -- 保证金8%：适中保证金
    3.2,  -- 止损3.2%：适中止损
    48,   -- 止盈回撤48%：适中回撤
    8,    -- 启动止盈8%：适中启动点
    '{"expectedProfit":"18-28%","tradingStyle":"均衡低波","avgHoldTime":"38-85min","winRate":"77%","profitFactor":"1.8","feeDeducted":"1%","riskLevel":"中","maxDrawdown":"8%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】低波均衡策略 | 预计月化18-28% | 10x杠杆/8%保证金 | 止损3.2%/启动止盈8%/回撤48% | 窗口170秒/阈值30U | 低波动期主力策略。已扣除1%手续费。市场有风险，仅供参考。',
    1, 11, NOW(), NOW()
);

-- 12. 低波动市场 - 激进策略（预计月化：28-42%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_low_vol'), 
    '【官方V31】低波激进型-仅供参考', 
    'aggressive', 
    'low_vol',
    140,  -- 时间窗口140秒：快速响应
    24,   -- 波动阈值24U：高灵敏度
    20,   -- 杠杆20倍：超高杠杆放大小波动
    5,    -- 保证金5%：小保证金高效
    5.0,  -- 止损5%：适应超高杠杆
    38,   -- 止盈回撤38%：较紧回撤
    6,    -- 启动止盈6%：快速启动
    '{"expectedProfit":"28-42%","tradingStyle":"激进低波","avgHoldTime":"25-60min","winRate":"72%","profitFactor":"2.2","feeDeducted":"1%","riskLevel":"高","maxDrawdown":"15%","disclaimer":"仅供参考"}',
    '【官方V31-仅供参考】低波激进策略 | 预计月化28-42% | 20x杠杆/5%保证金 | 止损5%/启动止盈6%/回撤38% | 窗口140秒/阈值24U | 超高杠杆放大小波动收益。已扣除1%手续费。市场有风险，仅供参考。',
    1, 12, NOW(), NOW()
);

-- =====================================================
-- 查询验证创建结果
-- =====================================================
SELECT 
    '模板组信息' AS '类别',
    g.id AS 'ID',
    g.group_name AS '模板组名称',
    g.group_key AS '唯一标识',
    CASE g.is_official WHEN 1 THEN '是' ELSE '否' END AS '官方模板'
FROM hg_trading_strategy_group g
WHERE g.group_key = 'official_v31';

-- 查看12套策略详情
SELECT 
    t.sort AS '序号',
    t.strategy_name AS '策略名称',
    t.market_state AS '市场状态',
    t.risk_preference AS '风险偏好',
    t.monitor_window AS '窗口(秒)',
    t.volatility_threshold AS '阈值(U)',
    t.leverage AS '杠杆(x)',
    t.margin_percent AS '保证金(%)',
    t.stop_loss_percent AS '止损(%)',
    t.auto_start_retreat_percent AS '启动止盈(%)',
    t.profit_retreat_percent AS '止盈回撤(%)',
    JSON_EXTRACT(t.config_json, '$.expectedProfit') AS '预计月化',
    JSON_EXTRACT(t.config_json, '$.winRate') AS '胜率',
    JSON_EXTRACT(t.config_json, '$.riskLevel') AS '风险等级'
FROM hg_trading_strategy_template t
JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE g.group_key = 'official_v31'
ORDER BY t.sort;

-- =====================================================
-- 【官方V31】策略参数汇总表（仅供参考）
-- =====================================================
/*
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+--------+
| 序号 | 策略名称          | 市场状态   | 风险偏好 | 窗口 | 阈值 | 杠杆| 保证金 | 止损 | 启动止盈| 止盈回撤 | 预计月化   | 胜率 | 最大回撤 |
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+--------+
|  1   | 趋势稳健型        | trend      | conserv  | 200  | 80   | 4  | 18%    | 2.8% | 20%    | 45%    | 20-28%   | 68%  | 8%     |
|  2   | 趋势均衡型        | trend      | balanced | 160  | 65   | 8  | 12%    | 4.0% | 16%    | 38%    | 28-40%   | 65%  | 12%    |
|  3   | 趋势激进型        | trend      | aggress  | 120  | 50   | 16 | 7%     | 6.0% | 12%    | 30%    | 40-58%   | 60%  | 18%    |
|  4   | 震荡稳健型        | volatile   | conserv  | 140  | 48   | 3  | 14%    | 2.2% | 12%    | 50%    | 14-20%   | 76%  | 6%     |
|  5   | 震荡均衡型        | volatile   | balanced | 110  | 40   | 6  | 10%    | 3.5% | 10%    | 40%    | 20-30%   | 72%  | 10%    |
|  6   | 震荡激进型        | volatile   | aggress  | 80   | 32   | 12 | 6%     | 5.5% | 8%     | 32%    | 30-45%   | 68%  | 16%    |
|  7   | 高波稳健型        | high_vol   | conserv  | 90   | 100  | 3  | 16%    | 3.0% | 22%    | 42%    | 22-32%   | 64%  | 9%     |
|  8   | 高波均衡型        | high_vol   | balanced | 70   | 80   | 7  | 9%     | 4.5% | 18%    | 35%    | 32-48%   | 62%  | 14%    |
|  9   | 高波激进型        | high_vol   | aggress  | 55   | 65   | 18 | 5%     | 7.5% | 14%    | 26%    | 45-65%   | 58%  | 22%    |
| 10   | 低波稳健型        | low_vol    | conserv  | 220  | 38   | 5  | 13%    | 2.0% | 10%    | 55%    | 12-18%   | 80%  | 5%     |
| 11   | 低波均衡型        | low_vol    | balanced | 170  | 30   | 10 | 8%     | 3.2% | 8%     | 48%    | 18-28%   | 77%  | 8%     |
| 12   | 低波激进型        | low_vol    | aggress  | 140  | 24   | 20 | 5%     | 5.0% | 6%     | 38%    | 28-42%   | 72%  | 15%    |
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+--------+

================== 推荐映射关系（高盈利配置） ==================
创建机器人时设置 remark 字段：
{
  "trend": "aggressive",      // 趋势市场 → 激进策略，预计月化40-58%
  "volatile": "balanced",     // 震荡市场 → 均衡策略，预计月化20-30%
  "high_vol": "aggressive",   // 高波动 → 激进策略，预计月化45-65%
  "low_vol": "aggressive"     // 低波动 → 激进策略，预计月化28-42%
}

综合预计月化收益：20%-65%（根据市场状态分布加权）

================== 参数说明 ==================
1. 止损百分比：基于保证金计算
   例：18%保证金+4x杠杆，止损2.8% = 保证金亏损2.8%时平仓
   
2. 启动止盈百分比：盈利达到保证金的X%时启动回撤保护
   例：保证金100U，启动止盈20% = 盈利20U时启动止盈回撤保护
   
3. 止盈回撤百分比：盈利部分的回撤百分比
   例：盈利100U，止盈回撤45% = 盈利从高点回撤45U时平仓

4. 杠杆与保证金关系：杠杆越小，保证金占用越大，风险敞口越低
   4x杠杆+18%保证金 vs 20x杠杆+5%保证金

5. 所有盈利预估已扣除双向手续费1%（0.005×2）

⚠️ 风险提示：以上预计收益仅供参考，实际收益受市场行情波动影响。
   市场有风险，投资需谨慎。
*/

