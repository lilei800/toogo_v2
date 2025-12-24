SET NAMES utf8mb4;

-- 更新V8策略组名称
UPDATE hg_trading_strategy_group SET group_name = 'BTCUSDT-v8官方推荐策略 - 半年数据优化版' WHERE group_key = 'official_btc_usdt_v8';

-- 更新V12策略组名称  
UPDATE hg_trading_strategy_group SET group_name = 'BTCUSDT-V12推荐策略，选择适合自己的 - 12个月回测验证版' WHERE group_key = 'official_btc_usdt_v12';

