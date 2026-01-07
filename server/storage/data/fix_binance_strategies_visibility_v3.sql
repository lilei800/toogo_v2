-- ============================================================
-- 修复币安策略组可见性问题（V3）
-- 如果策略组已导入但后台看不到，执行此脚本修复
-- ============================================================

DO $$
DECLARE
    v_btc_group_id BIGINT;
BEGIN
    -- 查找BTCUSDT策略组
    SELECT id INTO v_btc_group_id FROM hg_trading_strategy_group WHERE group_key = 'official_binance_btcusdt_v3';

    -- 如果存在但is_official不是1，修复它
    IF v_btc_group_id IS NOT NULL THEN
        UPDATE hg_trading_strategy_group
        SET
            is_official = 1,
            user_id = 0,
            is_active = 1,
            updated_at = NOW()
        WHERE id = v_btc_group_id;
        RAISE NOTICE '已修复BTCUSDT策略组(V3): id=%, is_official=1', v_btc_group_id;
    ELSE
        RAISE NOTICE 'BTCUSDT策略组(V3)不存在，请先执行导入SQL';
    END IF;
END $$;

-- 验证修复结果
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
WHERE group_key IN ('official_binance_btcusdt_v3')
ORDER BY id;


