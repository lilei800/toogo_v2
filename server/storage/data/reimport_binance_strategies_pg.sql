-- ============================================================
-- 重新导入币安策略组（确保正确设置is_official=1）
-- 如果之前导入失败或is_official字段不对，执行此脚本
-- ============================================================

DO $$
DECLARE
    v_btc_group_id BIGINT;
    v_eth_group_id BIGINT;
BEGIN
    -- 清理可能存在的旧数据
    SELECT id INTO v_btc_group_id FROM hg_trading_strategy_group WHERE group_key = 'official_binance_btcusdt_v1';
    IF v_btc_group_id IS NOT NULL THEN
        DELETE FROM hg_trading_strategy_template WHERE group_id = v_btc_group_id;
        DELETE FROM hg_trading_strategy_group WHERE id = v_btc_group_id;
        RAISE NOTICE '已清理旧BTCUSDT策略组数据';
    END IF;
    
    SELECT id INTO v_eth_group_id FROM hg_trading_strategy_group WHERE group_key = 'official_binance_ethusdt_v1';
    IF v_eth_group_id IS NOT NULL THEN
        DELETE FROM hg_trading_strategy_template WHERE group_id = v_eth_group_id;
        DELETE FROM hg_trading_strategy_group WHERE id = v_eth_group_id;
        RAISE NOTICE '已清理旧ETHUSDT策略组数据';
    END IF;
END $$;

-- 现在执行导入（这里需要包含完整的导入SQL）
-- 由于文件较大，建议直接执行 official_binance_btcusdt_v1_pg.sql 和 official_binance_ethusdt_v1_pg.sql

-- 导入后验证
SELECT 
  id,
  group_name AS "策略组名称",
  group_key AS "标识",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方",
  is_active AS "是否启用"
FROM hg_trading_strategy_group 
WHERE group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1');

