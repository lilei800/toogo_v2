-- 为策略组49创建策略模板（如果策略组49存在但没有模板）
-- 从V18策略组复制策略模板，或者使用V18的配置创建

-- 1. 检查策略组49是否存在
SET @group49Exists = (SELECT COUNT(*) FROM hg_trading_strategy_group WHERE id = 49);

-- 2. 如果策略组49存在但没有策略模板，从V18策略组复制
-- 获取V18策略组的ID
SET @v18GroupId = (SELECT id FROM hg_trading_strategy_group WHERE group_key = 'btc_usdt_v18_official' LIMIT 1);

-- 3. 如果V18策略组存在，复制其策略模板到策略组49
INSERT INTO hg_trading_strategy_template (
    group_id,
    strategy_key,
    strategy_name,
    risk_preference,
    market_state,
    monitor_window,
    volatility_threshold,
    leverage,
    margin_percent,
    stop_loss_percent,
    profit_retreat_percent,
    auto_start_retreat_percent,
    description,
    is_active,
    sort,
    created_at,
    updated_at
)
SELECT 
    49,  -- 目标策略组ID
    CONCAT(strategy_key, '_g49'),  -- 修改strategy_key避免冲突
    strategy_name,
    risk_preference,
    market_state,
    monitor_window,
    volatility_threshold,
    leverage,
    margin_percent,
    stop_loss_percent,
    profit_retreat_percent,
    auto_start_retreat_percent,
    description,
    is_active,
    sort,
    NOW(),
    NOW()
FROM hg_trading_strategy_template
WHERE group_id = @v18GroupId
  AND NOT EXISTS (
      SELECT 1 FROM hg_trading_strategy_template 
      WHERE group_id = 49 
        AND market_state = hg_trading_strategy_template.market_state 
        AND risk_preference = hg_trading_strategy_template.risk_preference
  );

-- 4. 验证策略组49的策略模板
SELECT 
    id,
    group_id,
    strategy_key,
    strategy_name,
    market_state,
    risk_preference,
    leverage,
    margin_percent,
    is_active
FROM hg_trading_strategy_template
WHERE group_id = 49
ORDER BY market_state, risk_preference;

