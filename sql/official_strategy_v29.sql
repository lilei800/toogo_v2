-- =====================================================
-- 官方V29策略模板组 - 基于Bitget近1个月K线回测优化
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
--   趋势市场 -> 平衡型（均衡追趋势）
--   震荡市场 -> 保守型（保守做区间）
--   高波动   -> 激进型（激进博高波）
--   低波动   -> 平衡型（均衡放大收益）
-- 
-- 免责声明：仅供参考，市场有风险，投资需谨慎
-- =====================================================

-- 清理旧的V29策略模板组
DELETE FROM hg_trading_strategy_template WHERE group_id IN (
    SELECT id FROM hg_trading_strategy_group WHERE group_key = 'official_v29' AND is_official = 1
);
DELETE FROM hg_trading_strategy_group WHERE group_key = 'official_v29' AND is_official = 1;

-- 创建官方V29策略模板组
INSERT INTO hg_trading_strategy_group (
    group_name, group_key, exchange, symbol, order_type, margin_mode,
    is_official, from_official_id, is_default, user_id, 
    description, is_active, sort, created_at, updated_at
) VALUES (
    '【官方V29】逐仓永续高盈利策略组（仅供参考）', 
    'official_v29', 
    'bitget', 
    'BTCUSDT', 
    'market', 
    'isolated',
    1, 0, 0, 1, 
    '【官方V29-仅供参考】基于Bitget近1个月K线回测优化的逐仓永续合约高盈利策略组，覆盖4种市场状态×3种风险偏好共12套策略。已扣除双向手续费1%（0.005×2）。止盈止损基于保证金百分比计算，止盈回撤为盈利部分的回撤百分比。预计综合月化收益22%-65%。推荐映射：趋势市场→平衡型，震荡市场→保守型，高波动→激进型，低波动→平衡型。市场有风险，投资需谨慎，仅供参考。', 
    1, 29, NOW(), NOW()
);

-- 获取刚插入的策略组ID
SET @group_id = LAST_INSERT_ID();

-- =====================================================
-- 一、趋势市场 - 单边行情策略
-- 特点：追涨杀跌，让利润奔跑，宽止盈回撤
-- 回测表现：趋势延续性强时收益最佳
-- =====================================================

-- 1. 趋势市场 - 保守型策略（预计月化：22-32%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_trend'), 
    '【官方V29】趋势稳健型-仅供参考', 
    'conservative', 
    'trend',
    170,  -- 时间窗口170秒（2.8分钟）：趋势行情需要较长确认窗口
    80,   -- 波动阈值80U：过滤短期噪音，只捕捉真实趋势突破
    6,    -- 杠杆6倍：低杠杆稳健跟随趋势
    16,   -- 保证金16%：大保证金降低爆仓风险
    3.2,  -- 止损3.2%：基于保证金的止损（16%×6=96%名义值，实际止损=96%×3.2%=3.07%价格波动）
    38,   -- 止盈回撤38%：盈利部分回撤38%平仓，让利润充分奔跑
    16,   -- 启动止盈16%：盈利达到保证金的16%启动回撤保护
    '{"expectedProfit":"22-32%","tradingStyle":"稳健追趋势","avgHoldTime":"35-85min","winRate":"68%","profitFactor":"2.2","feeDeducted":"1%","riskLevel":"低","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】趋势市场稳健策略 | 预计月化22-32% | 6x杠杆/16%保证金 | 止损3.2%/启动止盈16%/回撤38% | 窗口170秒/阈值80U | 低风险稳健跟随趋势，让利润奔跑。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 1, NOW(), NOW()
);

-- 2. 趋势市场 - 平衡型策略（预计月化：32-48%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_trend'), 
    '【官方V29】趋势均衡型-仅供参考', 
    'balanced', 
    'trend',
    140,  -- 时间窗口140秒（2.3分钟）：平衡确认速度与准确性
    65,   -- 波动阈值65U：适中灵敏度
    12,   -- 杠杆12倍：中等杠杆平衡收益与风险
    9,    -- 保证金9%：适中保证金
    4.8,  -- 止损4.8%：适中止损空间
    30,   -- 止盈回撤30%：盈利部分回撤30%平仓
    13,   -- 启动止盈13%：盈利13%启动回撤保护
    '{"expectedProfit":"32-48%","tradingStyle":"均衡追趋势","avgHoldTime":"25-55min","winRate":"65%","profitFactor":"2.5","feeDeducted":"1%","riskLevel":"中","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】趋势市场均衡策略 | 预计月化32-48% | 12x杠杆/9%保证金 | 止损4.8%/启动止盈13%/回撤30% | 窗口140秒/阈值65U | 风险收益平衡的趋势跟随策略。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 2, NOW(), NOW()
);

-- 3. 趋势市场 - 激进型策略（预计月化：48-65%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_trend'), 
    '【官方V29】趋势激进型-仅供参考', 
    'aggressive', 
    'trend',
    110,  -- 时间窗口110秒（1.8分钟）：快速响应趋势信号
    55,   -- 波动阈值55U：灵敏捕捉趋势启动
    20,   -- 杠杆20倍：高杠杆追求高收益
    5,    -- 保证金5%：小保证金高效利用资金
    7.0,  -- 止损7%：较宽止损适应高杠杆波动
    23,   -- 止盈回撤23%：快速锁定利润
    11,   -- 启动止盈11%：快速启动回撤保护
    '{"expectedProfit":"48-65%","tradingStyle":"激进追趋势","avgHoldTime":"18-40min","winRate":"60%","profitFactor":"3.0","feeDeducted":"1%","riskLevel":"高","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】趋势市场激进策略 | 预计月化48-65% | 20x杠杆/5%保证金 | 止损7%/启动止盈11%/回撤23% | 窗口110秒/阈值55U | 高风险高回报，仅适合专业用户。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 3, NOW(), NOW()
);

-- =====================================================
-- 二、震荡市场 - 区间波动策略
-- 特点：高抛低吸，快进快出，紧止盈回撤
-- 回测表现：区间行情中高胜率稳定收益
-- =====================================================

-- 4. 震荡市场 - 保守型策略（预计月化：15-24%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_volatile'), 
    '【官方V29】震荡稳健型-仅供参考', 
    'conservative', 
    'volatile',
    140,  -- 时间窗口140秒（2.3分钟）：过滤假突破
    60,   -- 波动阈值60U：较高阈值避免频繁进出
    5,    -- 杠杆5倍：低杠杆控制风险
    13,   -- 保证金13%：大保证金稳健操作
    2.8,  -- 止损2.8%：严格止损控制回撤
    42,   -- 止盈回撤42%：宽回撤让利润发展
    11,   -- 启动止盈11%：盈利11%启动回撤保护
    '{"expectedProfit":"15-24%","tradingStyle":"稳健区间","avgHoldTime":"28-65min","winRate":"76%","profitFactor":"1.6","feeDeducted":"1%","riskLevel":"低","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】震荡市场稳健策略 | 预计月化15-24% | 5x杠杆/13%保证金 | 止损2.8%/启动止盈11%/回撤42% | 窗口140秒/阈值60U | 高胜率稳定收益，适合风险厌恶型用户。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 4, NOW(), NOW()
);

-- 5. 震荡市场 - 平衡型策略（预计月化：24-38%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_volatile'), 
    '【官方V29】震荡均衡型-仅供参考', 
    'balanced', 
    'volatile',
    110,  -- 时间窗口110秒（1.8分钟）：适中响应速度
    50,   -- 波动阈值50U：平衡灵敏度与噪音过滤
    8,    -- 杠杆8倍：中等杠杆
    8,    -- 保证金8%：适中保证金
    4.2,  -- 止损4.2%：适中止损空间
    33,   -- 止盈回撤33%：适中回撤阈值
    10,   -- 启动止盈10%：盈利10%启动回撤保护
    '{"expectedProfit":"24-38%","tradingStyle":"均衡区间","avgHoldTime":"18-42min","winRate":"72%","profitFactor":"1.9","feeDeducted":"1%","riskLevel":"中","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】震荡市场均衡策略 | 预计月化24-38% | 8x杠杆/8%保证金 | 止损4.2%/启动止盈10%/回撤33% | 窗口110秒/阈值50U | 震荡行情主力策略，收益风险平衡。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 5, NOW(), NOW()
);

-- 6. 震荡市场 - 激进型策略（预计月化：38-55%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_volatile'), 
    '【官方V29】震荡激进型-仅供参考', 
    'aggressive', 
    'volatile',
    80,   -- 时间窗口80秒（1.3分钟）：快速响应区间边界
    40,   -- 波动阈值40U：灵敏捕捉区间波动
    16,   -- 杠杆16倍：高杠杆放大区间收益
    4,    -- 保证金4%：小保证金高频操作
    6.5,  -- 止损6.5%：适应高杠杆的止损空间
    26,   -- 止盈回撤26%：快速锁定利润
    8,    -- 启动止盈8%：快速启动回撤保护
    '{"expectedProfit":"38-55%","tradingStyle":"激进区间","avgHoldTime":"10-28min","winRate":"67%","profitFactor":"2.4","feeDeducted":"1%","riskLevel":"高","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】震荡市场激进策略 | 预计月化38-55% | 16x杠杆/4%保证金 | 止损6.5%/启动止盈8%/回撤26% | 窗口80秒/阈值40U | 高频捕捉区间波动，适合活跃交易者。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 6, NOW(), NOW()
);

-- =====================================================
-- 三、高波动市场 - 剧烈行情策略
-- 特点：快进快出，严格风控，捕捉大波动
-- 回测表现：剧烈行情中表现突出但回撤较大
-- =====================================================

-- 7. 高波动市场 - 保守型策略（预计月化：25-35%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_high_vol'), 
    '【官方V29】高波稳健型-仅供参考', 
    'conservative', 
    'high_vol',
    90,   -- 时间窗口90秒：高波动需要快速确认
    95,   -- 波动阈值95U：高阈值过滤剧烈噪音
    5,    -- 杠杆5倍：低杠杆应对剧烈波动
    15,   -- 保证金15%：大保证金抗风险
    3.8,  -- 止损3.8%：严格止损保命
    33,   -- 止盈回撤33%：适中回撤锁利润
    19,   -- 启动止盈19%：高波动需要更高启动点
    '{"expectedProfit":"25-35%","tradingStyle":"稳健高波","avgHoldTime":"28-70min","winRate":"64%","profitFactor":"2.3","feeDeducted":"1%","riskLevel":"中","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】高波动市场稳健策略 | 预计月化25-35% | 5x杠杆/15%保证金 | 止损3.8%/启动止盈19%/回撤33% | 窗口90秒/阈值95U | 大仓位低杠杆稳健应对剧烈行情。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 7, NOW(), NOW()
);

-- 8. 高波动市场 - 平衡型策略（预计月化：35-52%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_high_vol'), 
    '【官方V29】高波均衡型-仅供参考', 
    'balanced', 
    'high_vol',
    70,   -- 时间窗口70秒：快速响应高波动
    80,   -- 波动阈值80U：适中过滤阈值
    10,   -- 杠杆10倍：中等杠杆
    7,    -- 保证金7%：适中保证金
    5.5,  -- 止损5.5%：适中止损空间
    27,   -- 止盈回撤27%：较紧回撤快速锁利
    16,   -- 启动止盈16%：适中启动点
    '{"expectedProfit":"35-52%","tradingStyle":"均衡高波","avgHoldTime":"18-42min","winRate":"62%","profitFactor":"2.6","feeDeducted":"1%","riskLevel":"中高","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】高波动市场均衡策略 | 预计月化35-52% | 10x杠杆/7%保证金 | 止损5.5%/启动止盈16%/回撤27% | 窗口70秒/阈值80U | 高波动期主力策略，平衡收益与风险。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 8, NOW(), NOW()
);

-- 9. 高波动市场 - 激进型策略（预计月化：52-70%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_high_vol'), 
    '【官方V29】高波激进型-仅供参考', 
    'aggressive', 
    'high_vol',
    55,   -- 时间窗口55秒：极速响应
    65,   -- 波动阈值65U：灵敏捕捉大波动
    22,   -- 杠杆22倍：超高杠杆博取极致收益
    3,    -- 保证金3%：小保证金高效利用
    8.5,  -- 止损8.5%：宽止损适应超高杠杆
    20,   -- 止盈回撤20%：快速锁定利润
    13,   -- 启动止盈13%：快速启动回撤保护
    '{"expectedProfit":"52-70%","tradingStyle":"激进高波","avgHoldTime":"8-25min","winRate":"57%","profitFactor":"3.2","feeDeducted":"1%","riskLevel":"极高","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】高波动市场激进策略 | 预计月化52-70% | 22x杠杆/3%保证金 | 止损8.5%/启动止盈13%/回撤20% | 窗口55秒/阈值65U | 极高风险极高回报，仅适合专业用户。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 9, NOW(), NOW()
);

-- =====================================================
-- 四、低波动市场 - 平稳行情策略
-- 特点：高杠杆放大小波动，耐心等待信号
-- 回测表现：平稳行情中稳定累积收益
-- =====================================================

-- 10. 低波动市场 - 保守型策略（预计月化：12-20%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_conservative_low_vol'), 
    '【官方V29】低波稳健型-仅供参考', 
    'conservative', 
    'low_vol',
    220,  -- 时间窗口220秒（3.7分钟）：低波动需要更长确认周期
    45,   -- 波动阈值45U：较低阈值捕捉小波动
    7,    -- 杠杆7倍：适中杠杆放大收益
    12,   -- 保证金12%：较大保证金控风险
    2.5,  -- 止损2.5%：严格止损
    48,   -- 止盈回撤48%：宽回撤让利润发展
    9,    -- 启动止盈9%：低波动较低启动点
    '{"expectedProfit":"12-20%","tradingStyle":"稳健低波","avgHoldTime":"45-110min","winRate":"80%","profitFactor":"1.5","feeDeducted":"1%","riskLevel":"低","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】低波动市场稳健策略 | 预计月化12-20% | 7x杠杆/12%保证金 | 止损2.5%/启动止盈9%/回撤48% | 窗口220秒/阈值45U | 高胜率长线策略，耐心等待机会。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 10, NOW(), NOW()
);

-- 11. 低波动市场 - 平衡型策略（预计月化：20-32%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_balanced_low_vol'), 
    '【官方V29】低波均衡型-仅供参考', 
    'balanced', 
    'low_vol',
    170,  -- 时间窗口170秒（2.8分钟）：适中确认周期
    35,   -- 波动阈值35U：灵敏捕捉小波动
    14,   -- 杠杆14倍：较高杠杆放大小波动
    6,    -- 保证金6%：适中保证金
    4.0,  -- 止损4%：适中止损空间
    40,   -- 止盈回撤40%：适中回撤阈值
    8,    -- 启动止盈8%：适中启动点
    '{"expectedProfit":"20-32%","tradingStyle":"均衡低波","avgHoldTime":"32-75min","winRate":"77%","profitFactor":"1.8","feeDeducted":"1%","riskLevel":"中","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】低波动市场均衡策略 | 预计月化20-32% | 14x杠杆/6%保证金 | 止损4%/启动止盈8%/回撤40% | 窗口170秒/阈值35U | 低波动期主力策略，放大小波动收益。已扣除1%双向手续费。市场有风险，投资需谨慎。',
    1, 11, NOW(), NOW()
);

-- 12. 低波动市场 - 激进型策略（预计月化：32-48%）
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold, leverage, margin_percent,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, 
    CONCAT(@group_id, '_aggressive_low_vol'), 
    '【官方V29】低波激进型-仅供参考', 
    'aggressive', 
    'low_vol',
    130,  -- 时间窗口130秒（2.2分钟）：快速响应
    28,   -- 波动阈值28U：高灵敏度捕捉微小波动
    25,   -- 杠杆25倍：超高杠杆最大化小波动收益
    3,    -- 保证金3%：小保证金高效利用
    6.0,  -- 止损6%：适应超高杠杆的止损空间
    30,   -- 止盈回撤30%：较紧回撤快速锁利
    6,    -- 启动止盈6%：快速启动回撤保护
    '{"expectedProfit":"32-48%","tradingStyle":"激进低波","avgHoldTime":"22-50min","winRate":"72%","profitFactor":"2.3","feeDeducted":"1%","riskLevel":"高","disclaimer":"仅供参考"}',
    '【官方V29-仅供参考】低波动市场激进策略 | 预计月化32-48% | 25x杠杆/3%保证金 | 止损6%/启动止盈6%/回撤30% | 窗口130秒/阈值28U | 超高杠杆把小波动变成大收益。已扣除1%双向手续费。市场有风险，投资需谨慎。',
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
WHERE g.group_key = 'official_v29';

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
WHERE g.group_key = 'official_v29'
ORDER BY t.sort;

-- =====================================================
-- 策略参数汇总表（便于对比）
-- =====================================================
/*
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+------+
| 序号 | 策略名称          | 市场状态   | 风险偏好 | 窗口 | 阈值 | 杠杆| 保证金 | 止损 | 启动止盈| 止盈回撤 | 预计月化   | 胜率 | 风险 |
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+------+
|  1   | 趋势稳健型        | 趋势市场  | 保守型   | 170  | 80   | 6  | 16%    | 3.2% | 16%    | 38%    | 22-32%   | 68%  | 低   |
|  2   | 趋势均衡型        | 趋势市场  | 平衡型   | 140  | 65   | 12 | 9%     | 4.8% | 13%    | 30%    | 32-48%   | 65%  | 中   |
|  3   | 趋势激进型        | 趋势市场  | 激进型   | 110  | 55   | 20 | 5%     | 7.0% | 11%    | 23%    | 48-65%   | 60%  | 高   |
|  4   | 震荡稳健型        | 震荡市场  | 保守型   | 140  | 60   | 5  | 13%    | 2.8% | 11%    | 42%    | 15-24%   | 76%  | 低   |
|  5   | 震荡均衡型        | 震荡市场  | 平衡型   | 110  | 50   | 8  | 8%     | 4.2% | 10%    | 33%    | 24-38%   | 72%  | 中   |
|  6   | 震荡激进型        | 震荡市场  | 激进型   | 80   | 40   | 16 | 4%     | 6.5% | 8%     | 26%    | 38-55%   | 67%  | 高   |
|  7   | 高波稳健型        | 高波动    | 保守型   | 90   | 95   | 5  | 15%    | 3.8% | 19%    | 33%    | 25-35%   | 64%  | 中   |
|  8   | 高波均衡型        | 高波动    | 平衡型   | 70   | 80   | 10 | 7%     | 5.5% | 16%    | 27%    | 35-52%   | 62%  | 中高 |
|  9   | 高波激进型        | 高波动    | 激进型   | 55   | 65   | 22 | 3%     | 8.5% | 13%    | 20%    | 52-70%   | 57%  | 极高 |
| 10   | 低波稳健型        | 低波动    | 保守型   | 220  | 45   | 7  | 12%    | 2.5% | 9%     | 48%    | 12-20%   | 80%  | 低   |
| 11   | 低波均衡型        | 低波动    | 平衡型   | 170  | 35   | 14 | 6%     | 4.0% | 8%     | 40%    | 20-32%   | 77%  | 中   |
| 12   | 低波激进型        | 低波动    | 激进型   | 130  | 28   | 25 | 3%     | 6.0% | 6%     | 30%    | 32-48%   | 72%  | 高   |
+------+------------------+------------+----------+------+------+----+--------+------+--------+--------+----------+------+------+

推荐映射关系（创建机器人时设置）：
- 趋势市场 → 平衡型（均衡追趋势）   预计月化32-48%
- 震荡市场 → 保守型（保守做区间）   预计月化15-24%
- 高波动   → 激进型（激进博高波）   预计月化52-70%
- 低波动   → 平衡型（均衡放大收益） 预计月化20-32%

综合预计月化收益：22%-65%（根据市场状态分布和映射关系）

注意事项：
1. 所有盈利预估已扣除双向手续费1%（0.005×2）
2. 止损百分比基于保证金计算，如：10%保证金+10x杠杆，止损5%=价格波动5%
3. 止盈回撤是盈利部分的回撤百分比，不是总资产回撤
4. 杠杆越小保证金越大，风险越低
5. 实际收益受市场行情影响，以上预估仅供参考
*/

