-- ========================================
-- 迁移脚本：统一市场状态值格式
-- 日期：2025-12-03
-- 说明：将旧格式的市场状态值更新为新格式
--   - volatile → range
--   - high_volatility → high_vol
--   - low_volatility → low_vol
-- ========================================

-- 开始事务
START TRANSACTION;

-- 1. 更新策略模板表中的市场状态
UPDATE hg_trading_strategy_template 
SET market_state = CASE 
  WHEN market_state = 'volatile' THEN 'range'
  WHEN market_state = 'high_volatility' THEN 'high_vol'
  WHEN market_state = 'low_volatility' THEN 'low_vol'
  ELSE market_state
END
WHERE market_state IN ('volatile', 'high_volatility', 'low_volatility');

-- 2. 更新机器人的 current_strategy JSON 字段中的 marketRiskMapping
-- 注意：这个更新比较复杂，因为需要处理 JSON 字段
-- 如果有机器人使用了旧格式的映射，需要更新
UPDATE hg_trading_robot
SET current_strategy = JSON_SET(
    current_strategy,
    '$.riskConfig.marketRiskMapping.range', 
    JSON_EXTRACT(current_strategy, '$.riskConfig.marketRiskMapping.volatile')
)
WHERE JSON_EXTRACT(current_strategy, '$.riskConfig.marketRiskMapping.volatile') IS NOT NULL;

-- 3. 删除旧的 volatile 键
UPDATE hg_trading_robot
SET current_strategy = JSON_REMOVE(current_strategy, '$.riskConfig.marketRiskMapping.volatile')
WHERE JSON_EXTRACT(current_strategy, '$.riskConfig.marketRiskMapping.volatile') IS NOT NULL;

-- 检查更新结果
SELECT 
    '策略模板更新统计' as 说明,
    COUNT(*) as 总数,
    SUM(CASE WHEN market_state = 'range' THEN 1 ELSE 0 END) as 震荡市场,
    SUM(CASE WHEN market_state = 'high_vol' THEN 1 ELSE 0 END) as 高波动,
    SUM(CASE WHEN market_state = 'low_vol' THEN 1 ELSE 0 END) as 低波动,
    SUM(CASE WHEN market_state IN ('volatile', 'high_volatility', 'low_volatility') THEN 1 ELSE 0 END) as 旧格式剩余
FROM hg_trading_strategy_template;

-- 提交事务
COMMIT;

-- ========================================
-- 迁移完成
-- ========================================

