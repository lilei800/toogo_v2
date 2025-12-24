-- 验证Toogo数据导入
SELECT '=== Toogo Tables ===' AS section;
SELECT table_name, table_rows FROM information_schema.tables WHERE table_schema = 'hotgo' AND table_name LIKE 'hg_toogo_%' ORDER BY table_name;

SELECT '=== Menu Count ===' AS section;
SELECT COUNT(*) AS toogo_menus FROM hg_admin_menu WHERE id >= 2000 AND id < 2200;

SELECT '=== Cron Tasks ===' AS section;
SELECT id, title, name, status FROM hg_sys_cron WHERE name LIKE 'toogo_%';

SELECT '=== VIP Levels ===' AS section;
SELECT level, level_name, power_discount FROM hg_toogo_vip_level;

SELECT '=== Plans ===' AS section;
SELECT plan_code, plan_name, robot_limit, price_monthly FROM hg_toogo_plan;

SELECT '=== Strategy Templates ===' AS section;
SELECT strategy_key, risk_preference, market_state FROM hg_toogo_strategy_template;

SELECT '=== Config Groups ===' AS section;
SELECT DISTINCT `group` FROM hg_toogo_config;

