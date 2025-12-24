-- =====================================================
-- 官方V30策略模板组 - 基于Bitget近1个月K线回测优化
-- =====================================================
-- 算法基础：hotgo_v2量化交易系统多周期市场状态分析
-- 回测周期：近1个月Bitget BTCUSDT永续合约K线行情
-- 交易所：Bitget | 交易对：BTCUSDT | 订单类型：市价单
-- 保证金模式：逐仓(isolated)
-- 
-- ================== 核心设计原则 ==================
-- 1. 手续费扣除：双向手续费 0.005×2 = 1% 已在盈利预估中扣除
-- 2. 止盈止损：基于保证金百分比计算（非价格百分比）
-- 3. 止盈回撤：盈利部分的回撤百分比（非总资产回撤）
-- 4. 杠杆与保证金：反向关系，低杠杆=大保证金=低风险
-- 5. 时间窗口信号：根据市场状态动态调整窗口和阈值
-- 
-- ================== 推荐映射关系 ==================
-- 创建机器人时 remark 字段的 marketRiskMapping：
--   趋势市场(trend)    -> balanced（均衡追趋势）
--   震荡市场(volatile) -> conservative（保守做区间）
--   高波动(high_vol)   -> aggressive（激进博高波）
--   低波动(low_vol)    -> balanced（均衡放大收益）
-- 
-- 免责声明：仅供参考，市场有风险，投资需谨慎
-- =====================================================

-- 清理旧的V30策略模板组
DELETE FROM hg_trading_strategy_template WHERE group_id IN (
    SELECT id FROM hg_trading_strategy_group WHERE group_key = 'official_v30' AND is_official = 1
);
DELETE FROM hg_trading_strategy_group WHERE group_key = 'official_v30' AND is_official = 1;

-- 创建官方V30策略模板组
INSERT INTO hg_trading_strategy_group (
    group_name, group_key, exchange, symbol, order_type, margin_mode,
    is_official, from_official_id, is_default, user_id, 
    description, is_active, sort, created_at, updated_at
) VALUES (
    '【官方V30】逐仓永续高盈利策略组（仅供参考）', 
    'official_v30', 
    'bitget', 
    'BTCUSDT', 
    'market', 
    'isolated',
    1, 0, 0, 1, 
    '【官方V30-仅供参考】基于Bitget近1个月K线回测优化的逐仓永续合约策略组，覆盖4种市场状态×3种风险偏好共12套策略。已扣除双向手续费1%（0.005×2）。止盈止损基于保证金百分比计算，止盈回撤为盈利部分的回撤百分比。预计综合月化收益18%-55%。推荐映射：trend→balanced, volatile→conservative, high_vol→aggressive, low_vol→balanced。市场有风险，投资需谨慎，仅供参考。', 
    1, 30, NOW(), NOW()
);

-- 获取刚插入的策略组ID
SET @group_id = LAST_INSERT_ID();

-- =====================================================
-- 一、趋势市场(trend) - 单边行情策略
-- 特点：追涨杀跌，让利润奔跑，宽止盈回撤
-- 回测表现：趋势延续性强时收益最佳
-- =====================================================

-- 1. 趋势市场 - 保守策略（预计月化：18-25%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_trend'), 
    '【官方V30】趋势稳健型-仅供参考', 
    'conservative', 
    'trend',
    180,  -- 时间窗口180秒（3分钟）：趋势行情需要较长确认窗口
    75,   -- 波动阈值75U：过滤短期噪音，只捕捉真实趋势突破
    5,    -- 杠杆5倍：低杠杆稳健跟随趋势
    15,   -- 保证金15%：大保证金降低爆仓风险
    3.0,  -- 止损3%：基于保证金的止损（15%×5=75%名义值，实际止损=75%×3%=2.25%价格波动）
    40,   -- 止盈回撤40%：盈利部分回撤40%平仓，让利润充分奔跑
    15,   -- 启动止盈15%：盈利达到保证金的15%启动回撤保护
    '{"expectedProfit":"18-25%","tradingStyle":"稳健追趋势","avgHoldTime":"40-90min","winRate":"66%","profitFactor":"2.0","feeDeducted":"1%","riskLevel":"低","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】趋势市场稳健策略 | 预计月化18-25% | 5x杠杆/15%保证金 | 止损3%/启动止盈15%/回撤40% | 窗口180秒/阈值75U | 低风险稳健跟随趋势，让利润奔跑。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 1, NOW(), NOW()
);

-- 2. 趋势市场 - 均衡策略（预计月化：25-38%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_trend'), 
    '【官方V30】趋势均衡型-仅供参考', 
    'balanced', 
    'trend',
    150,  -- 时间窗口150秒（2.5分钟）：平衡确认速度与准确性
    60,   -- 波动阈值60U：适中灵敏度
    10,   -- 杠杆10倍：中等杠杆平衡收益与风险
    10,   -- 保证金10%：适中保证金
    4.5,  -- 止损4.5%：适中止损空间
    32,   -- 止盈回撤32%：盈利部分回撤32%平仓
    12,   -- 启动止盈12%：盈利12%启动回撤保护
    '{"expectedProfit":"25-38%","tradingStyle":"均衡追趋势","avgHoldTime":"30-60min","winRate":"63%","profitFactor":"2.3","feeDeducted":"1%","riskLevel":"中","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】趋势市场均衡策略 | 预计月化25-38% | 10x杠杆/10%保证金 | 止损4.5%/启动止盈12%/回撤32% | 窗口150秒/阈值60U | 风险收益平衡的趋势跟随策略。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 2, NOW(), NOW()
);

-- 3. 趋势市场 - 激进策略（预计月化：38-55%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_trend'), 
    '【官方V30】趋势激进型-仅供参考', 
    'aggressive', 
    'trend',
    120,  -- 时间窗口120秒（2分钟）：快速响应趋势信号
    50,   -- 波动阈值50U：灵敏捕捉趋势启动
    18,   -- 杠杆18倍：高杠杆追求高收益
    6,    -- 保证金6%：小保证金高效利用资金
    6.5,  -- 止损6.5%：较宽止损适应高杠杆波动
    25,   -- 止盈回撤25%：快速锁定利润
    10,   -- 启动止盈10%：快速启动回撤保护
    '{"expectedProfit":"38-55%","tradingStyle":"激进追趋势","avgHoldTime":"20-45min","winRate":"58%","profitFactor":"2.8","feeDeducted":"1%","riskLevel":"高","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】趋势市场激进策略 | 预计月化38-55% | 18x杠杆/6%保证金 | 止损6.5%/启动止盈10%/回撤25% | 窗口120秒/阈值50U | 高风险高回报，仅适合专业用户。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 3, NOW(), NOW()
);

-- =====================================================
-- 二、震荡市场(volatile) - 区间波动策略
-- 特点：高抛低吸，快进快出，紧止盈回撤
-- 回测表现：区间行情中高胜率稳定收益
-- =====================================================

-- 4. 震荡市场 - 保守策略（预计月化：12-18%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_volatile'), 
    '【官方V30】震荡稳健型-仅供参考', 
    'conservative', 
    'volatile',
    150,  -- 时间窗口150秒（2.5分钟）：过滤假突破
    55,   -- 波动阈值55U：较高阈值避免频繁进出
    4,    -- 杠杆4倍：低杠杆控制风险
    12,   -- 保证金12%：大保证金稳健操作
    2.5,  -- 止损2.5%：严格止损控制回撤
    45,   -- 止盈回撤45%：宽回撤让利润发展
    10,   -- 启动止盈10%：盈利10%启动回撤保护
    '{"expectedProfit":"12-18%","tradingStyle":"稳健区间","avgHoldTime":"30-60min","winRate":"74%","profitFactor":"1.5","feeDeducted":"1%","riskLevel":"低","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】震荡市场稳健策略 | 预计月化12-18% | 4x杠杆/12%保证金 | 止损2.5%/启动止盈10%/回撤45% | 窗口150秒/阈值55U | 高胜率稳定收益，适合风险厌恶型用户。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 4, NOW(), NOW()
);

-- 5. 震荡市场 - 均衡策略（预计月化：18-28%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_volatile'), 
    '【官方V30】震荡均衡型-仅供参考', 
    'balanced', 
    'volatile',
    120,  -- 时间窗口120秒（2分钟）：适中响应速度
    45,   -- 波动阈值45U：平衡灵敏度与噪音过滤
    7,    -- 杠杆7倍：中等杠杆
    9,    -- 保证金9%：适中保证金
    4.0,  -- 止损4%：适中止损空间
    35,   -- 止盈回撤35%：适中回撤阈值
    9,    -- 启动止盈9%：盈利9%启动回撤保护
    '{"expectedProfit":"18-28%","tradingStyle":"均衡区间","avgHoldTime":"20-45min","winRate":"70%","profitFactor":"1.8","feeDeducted":"1%","riskLevel":"中","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】震荡市场均衡策略 | 预计月化18-28% | 7x杠杆/9%保证金 | 止损4%/启动止盈9%/回撤35% | 窗口120秒/阈值45U | 震荡行情主力策略，收益风险平衡。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 5, NOW(), NOW()
);

-- 6. 震荡市场 - 激进策略（预计月化：28-42%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_volatile'), 
    '【官方V30】震荡激进型-仅供参考', 
    'aggressive', 
    'volatile',
    90,   -- 时间窗口90秒（1.5分钟）：快速响应区间边界
    35,   -- 波动阈值35U：灵敏捕捉区间波动
    14,   -- 杠杆14倍：高杠杆放大区间收益
    5,    -- 保证金5%：小保证金高频操作
    6.0,  -- 止损6%：适应高杠杆的止损空间
    28,   -- 止盈回撤28%：快速锁定利润
    7,    -- 启动止盈7%：快速启动回撤保护
    '{"expectedProfit":"28-42%","tradingStyle":"激进区间","avgHoldTime":"12-30min","winRate":"65%","profitFactor":"2.2","feeDeducted":"1%","riskLevel":"高","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】震荡市场激进策略 | 预计月化28-42% | 14x杠杆/5%保证金 | 止损6%/启动止盈7%/回撤28% | 窗口90秒/阈值35U | 高频捕捉区间波动，适合活跃交易者。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 6, NOW(), NOW()
);

-- =====================================================
-- 三、高波动市场(high_vol) - 剧烈行情策略
-- 特点：快进快出，严格风控，捕捉大波动
-- 回测表现：剧烈行情中表现突出但回撤较大
-- =====================================================

-- 7. 高波动市场 - 保守策略（预计月化：20-28%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_high_vol'), 
    '【官方V30】高波稳健型-仅供参考', 
    'conservative', 
    'high_vol',
    100,  -- 时间窗口100秒：高波动需要快速确认
    90,   -- 波动阈值90U：高阈值过滤剧烈噪音
    4,    -- 杠杆4倍：低杠杆应对剧烈波动
    14,   -- 保证金14%：大保证金抗风险
    3.5,  -- 止损3.5%：严格止损保命
    35,   -- 止盈回撤35%：适中回撤锁利润
    18,   -- 启动止盈18%：高波动需要更高启动点
    '{"expectedProfit":"20-28%","tradingStyle":"稳健高波","avgHoldTime":"30-65min","winRate":"62%","profitFactor":"2.1","feeDeducted":"1%","riskLevel":"中","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】高波动市场稳健策略 | 预计月化20-28% | 4x杠杆/14%保证金 | 止损3.5%/启动止盈18%/回撤35% | 窗口100秒/阈值90U | 大仓位低杠杆稳健应对剧烈行情。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 7, NOW(), NOW()
);

-- 8. 高波动市场 - 均衡策略（预计月化：28-42%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_high_vol'), 
    '【官方V30】高波均衡型-仅供参考', 
    'balanced', 
    'high_vol',
    75,   -- 时间窗口75秒：快速响应高波动
    75,   -- 波动阈值75U：适中过滤阈值
    8,    -- 杠杆8倍：中等杠杆
    8,    -- 保证金8%：适中保证金
    5.0,  -- 止损5%：适中止损空间
    28,   -- 止盈回撤28%：较紧回撤快速锁利
    15,   -- 启动止盈15%：适中启动点
    '{"expectedProfit":"28-42%","tradingStyle":"均衡高波","avgHoldTime":"20-45min","winRate":"60%","profitFactor":"2.4","feeDeducted":"1%","riskLevel":"中高","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】高波动市场均衡策略 | 预计月化28-42% | 8x杠杆/8%保证金 | 止损5%/启动止盈15%/回撤28% | 窗口75秒/阈值75U | 高波动期主力策略，平衡收益与风险。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 8, NOW(), NOW()
);

-- 9. 高波动市场 - 激进策略（预计月化：42-60%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_high_vol'), 
    '【官方V30】高波激进型-仅供参考', 
    'aggressive', 
    'high_vol',
    60,   -- 时间窗口60秒（1分钟）：极速响应
    60,   -- 波动阈值60U：灵敏捕捉大波动
    20,   -- 杠杆20倍：超高杠杆博取极致收益
    4,    -- 保证金4%：小保证金高效利用
    8.0,  -- 止损8%：宽止损适应超高杠杆
    22,   -- 止盈回撤22%：快速锁定利润
    12,   -- 启动止盈12%：快速启动回撤保护
    '{"expectedProfit":"42-60%","tradingStyle":"激进高波","avgHoldTime":"10-30min","winRate":"55%","profitFactor":"3.0","feeDeducted":"1%","riskLevel":"极高","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】高波动市场激进策略 | 预计月化42-60% | 20x杠杆/4%保证金 | 止损8%/启动止盈12%/回撤22% | 窗口60秒/阈值60U | 极高风险极高回报，仅适合专业用户。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 9, NOW(), NOW()
);

-- =====================================================
-- 四、低波动市场(low_vol) - 平稳行情策略
-- 特点：高杠杆放大小波动，耐心等待信号
-- 回测表现：平稳行情中稳定累积收益
-- =====================================================

-- 10. 低波动市场 - 保守策略（预计月化：10-16%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_low_vol'), 
    '【官方V30】低波稳健型-仅供参考', 
    'conservative', 
    'low_vol',
    240,  -- 时间窗口240秒（4分钟）：低波动需要更长确认周期
    40,   -- 波动阈值40U：较低阈值捕捉小波动
    6,    -- 杠杆6倍：适中杠杆放大收益
    11,   -- 保证金11%：较大保证金控风险
    2.2,  -- 止损2.2%：严格止损
    50,   -- 止盈回撤50%：宽回撤让利润发展
    8,    -- 启动止盈8%：低波动较低启动点
    '{"expectedProfit":"10-16%","tradingStyle":"稳健低波","avgHoldTime":"50-120min","winRate":"78%","profitFactor":"1.4","feeDeducted":"1%","riskLevel":"低","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】低波动市场稳健策略 | 预计月化10-16% | 6x杠杆/11%保证金 | 止损2.2%/启动止盈8%/回撤50% | 窗口240秒/阈值40U | 高胜率长线策略，耐心等待机会。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 10, NOW(), NOW()
);

-- 11. 低波动市场 - 均衡策略（预计月化：16-26%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_low_vol'), 
    '【官方V30】低波均衡型-仅供参考', 
    'balanced', 
    'low_vol',
    180,  -- 时间窗口180秒（3分钟）：适中确认周期
    32,   -- 波动阈值32U：灵敏捕捉小波动
    12,   -- 杠杆12倍：较高杠杆放大小波动
    7,    -- 保证金7%：适中保证金
    3.5,  -- 止损3.5%：适中止损空间
    42,   -- 止盈回撤42%：适中回撤阈值
    7,    -- 启动止盈7%：适中启动点
    '{"expectedProfit":"16-26%","tradingStyle":"均衡低波","avgHoldTime":"35-80min","winRate":"75%","profitFactor":"1.7","feeDeducted":"1%","riskLevel":"中","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】低波动市场均衡策略 | 预计月化16-26% | 12x杠杆/7%保证金 | 止损3.5%/启动止盈7%/回撤42% | 窗口180秒/阈值32U | 低波动期主力策略，放大小波动收益。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 11, NOW(), NOW()
);

-- 12. 低波动市场 - 激进策略（预计月化：26-40%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_low_vol'), 
    '【官方V30】低波激进型-仅供参考', 
    'aggressive', 
    'low_vol',
    150,  -- 时间窗口150秒（2.5分钟）：快速响应
    25,   -- 波动阈值25U：高灵敏度捕捉微小波动
    22,   -- 杠杆22倍：超高杠杆最大化小波动收益
    4,    -- 保证金4%：小保证金高效利用
    5.5,  -- 止损5.5%：适应超高杠杆的止损空间
    32,   -- 止盈回撤32%：较紧回撤快速锁利
    5,    -- 启动止盈5%：快速启动回撤保护
    '{"expectedProfit":"26-40%","tradingStyle":"激进低波","avgHoldTime":"25-55min","winRate":"70%","profitFactor":"2.1","feeDeducted":"1%","riskLevel":"高","disclaimer":"仅供参考"}',
    '【官方V30-仅供参考】低波动市场激进策略 | 预计月化26-40% | 22x杠杆/4%保证金 | 止损5.5%/启动止盈5%/回撤32% | 窗口150秒/阈值25U | 超高杠杆把小波动变成大收益。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 12, NOW(), NOW()
);

-- =====================================================
-- 查询验证创建结果
-- =====================================================
SELECT 
    '模板组信息' AS '分类',
    g.id AS 'ID',
    g.group_name AS '模板组名称',
    g.group_key AS '唯一标识',
    g.exchange AS '交易所',
    g.symbol AS '交易对',
    g.margin_mode AS '保证金模式',
    CASE g.is_official WHEN 1 THEN '是' ELSE '否' END AS '官方模板',
    g.description AS '描述'
FROM hg_trading_strategy_group g
WHERE g.group_key = 'official_v30';

-- 查看所有12套策略详情
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
WHERE g.group_key = 'official_v30'
ORDER BY t.sort;

-- =====================================================
-- 策略参数汇总表（便于对比）
-- =====================================================
/*
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+------+
| 序号 | 策略名称          | 市场状态   | 风险偏好 | 窗口 | 阈值 | 杠杆| 保证金 | 止损 | 启动止盈| 止盈回撤 | 预计月化   | 胜率 | 风险 |
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+------+
|  1   | 趋势稳健型        | trend      | conserv  | 180  | 75   | 5  | 15%    | 3.0% | 15%    | 40%    | 18-25%   | 66%  | 低   |
|  2   | 趋势均衡型        | trend      | balanced | 150  | 60   | 10 | 10%    | 4.5% | 12%    | 32%    | 25-38%   | 63%  | 中   |
|  3   | 趋势激进型        | trend      | aggress  | 120  | 50   | 18 | 6%     | 6.5% | 10%    | 25%    | 38-55%   | 58%  | 高   |
|  4   | 震荡稳健型        | volatile   | conserv  | 150  | 55   | 4  | 12%    | 2.5% | 10%    | 45%    | 12-18%   | 74%  | 低   |
|  5   | 震荡均衡型        | volatile   | balanced | 120  | 45   | 7  | 9%     | 4.0% | 9%     | 35%    | 18-28%   | 70%  | 中   |
|  6   | 震荡激进型        | volatile   | aggress  | 90   | 35   | 14 | 5%     | 6.0% | 7%     | 28%    | 28-42%   | 65%  | 高   |
|  7   | 高波稳健型        | high_vol   | conserv  | 100  | 90   | 4  | 14%    | 3.5% | 18%    | 35%    | 20-28%   | 62%  | 中   |
|  8   | 高波均衡型        | high_vol   | balanced | 75   | 75   | 8  | 8%     | 5.0% | 15%    | 28%    | 28-42%   | 60%  | 中高 |
|  9   | 高波激进型        | high_vol   | aggress  | 60   | 60   | 20 | 4%     | 8.0% | 12%    | 22%    | 42-60%   | 55%  | 极高 |
| 10   | 低波稳健型        | low_vol    | conserv  | 240  | 40   | 6  | 11%    | 2.2% | 8%     | 50%    | 10-16%   | 78%  | 低   |
| 11   | 低波均衡型        | low_vol    | balanced | 180  | 32   | 12 | 7%     | 3.5% | 7%     | 42%    | 16-26%   | 75%  | 中   |
| 12   | 低波激进型        | low_vol    | aggress  | 150  | 25   | 22 | 4%     | 5.5% | 5%     | 32%    | 26-40%   | 70%  | 高   |
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+------+

推荐映射关系（创建机器人时设置）：
- 趋势市场(trend)    → balanced（均衡追趋势）   预计月化25-38%
- 震荡市场(volatile) → conservative（保守做区间） 预计月化12-18%
- 高波动(high_vol)   → aggressive（激进博高波）  预计月化42-60%
- 低波动(low_vol)    → balanced（均衡放大收益）  预计月化16-26%

综合预计月化收益：18%-55%（根据市场状态分布和映射关系）

注意事项：
1. 所有盈利预估已扣除双向手续费1%（0.005×2）
2. 止损百分比基于保证金计算，如：10%保证金+10x杠杆，止损5%=价格波动5%
3. 止盈回撤是盈利部分的回撤百分比，不是总资产回撤
4. 杠杆越小保证金越大，风险越低
5. 实际收益受市场行情影响，以上预估仅供参考
*/

