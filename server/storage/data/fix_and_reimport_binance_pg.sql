-- ============================================================
-- 修复并重新导入币安官方策略组
-- 确保is_official=1，group_key正确
-- ============================================================

DO $$
DECLARE
    v_btc_group_id BIGINT;
    v_eth_group_id BIGINT;
    v_existing_btc_id BIGINT;
    v_existing_eth_id BIGINT;
BEGIN
    -- 1. 检查是否存在我们创建的策略组
    SELECT id INTO v_existing_btc_id FROM hg_trading_strategy_group WHERE group_key = 'official_binance_btcusdt_v1';
    SELECT id INTO v_existing_eth_id FROM hg_trading_strategy_group WHERE group_key = 'official_binance_ethusdt_v1';
    
    -- 2. 如果存在但is_official不是1，修复它
    IF v_existing_btc_id IS NOT NULL THEN
        UPDATE hg_trading_strategy_group 
        SET 
            is_official = 1,
            user_id = 0,
            is_active = 1,
            updated_at = NOW()
        WHERE id = v_existing_btc_id;
        RAISE NOTICE '已修复BTCUSDT策略组: id=%, is_official=1', v_existing_btc_id;
    ELSE
        RAISE NOTICE 'BTCUSDT策略组不存在，需要重新导入';
    END IF;
    
    IF v_existing_eth_id IS NOT NULL THEN
        UPDATE hg_trading_strategy_group 
        SET 
            is_official = 1,
            user_id = 0,
            is_active = 1,
            updated_at = NOW()
        WHERE id = v_existing_eth_id;
        RAISE NOTICE '已修复ETHUSDT策略组: id=%, is_official=1', v_existing_eth_id;
    ELSE
        RAISE NOTICE 'ETHUSDT策略组不存在，需要重新导入';
    END IF;
    
    -- 3. 如果不存在，需要手动执行导入SQL文件
    IF v_existing_btc_id IS NULL OR v_existing_eth_id IS NULL THEN
        RAISE NOTICE '请执行以下SQL文件导入策略组：';
        RAISE NOTICE '1. official_binance_btcusdt_v1_pg.sql';
        RAISE NOTICE '2. official_binance_ethusdt_v1_pg.sql';
    END IF;
END $$;

-- 验证结果
SELECT 
  id,
  group_name AS "策略组名称",
  group_key AS "标识",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方",
  is_active AS "是否启用",
  user_id AS "用户ID"
FROM hg_trading_strategy_group 
WHERE group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1')
ORDER BY id;

