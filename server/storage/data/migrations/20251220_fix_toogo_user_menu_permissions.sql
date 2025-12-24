-- ============================================================
-- 修复 Toogo（量化交易-用户端）菜单 permissions 配置
-- 目的：
-- - 旧数据里很多菜单 permissions 为空或写成“前端路由”，导致 Casbin 无法放行页面所需 API
-- - 本脚本按“菜单 path / component”幂等更新为“页面实际调用的 API 路径”
-- 说明：
-- - permissions 字段为英文逗号分隔
-- - 执行前建议备份数据库
-- ============================================================

SET NAMES utf8mb4;

-- ============ 控制台 /toogo/dashboard ============
UPDATE `hg_admin_menu`
SET
  `permissions` = '/toogo/wallet/overview,/toogo/user/info,/toogo/user/refresh-invite-code,/toogo/subscription/my,/toogo/commission/stat,/trading/robot/list,/trading/robot/start,/trading/robot/stop,/toogo/wallet/transfer',
  `updated_at` = NOW()
WHERE `path` = '/toogo/dashboard' OR `component` = '/toogo/dashboard/index';

-- ============ 我的机器人 /toogo/robot ============
UPDATE `hg_admin_menu`
SET
  `permissions` = '/trading/robot/list,/trading/robot/view,/trading/robot/create,/trading/robot/update,/trading/robot/delete,/trading/robot/start,/trading/robot/stop,/trading/robot/pause,/trading/robot/stats,/trading/robot/positions,/trading/robot/orders,/trading/robot/orderHistory,/trading/robot/closePosition,/trading/robot/cancelOrder,/trading/robot/signalLogs,/trading/robot/executionLogs,/trading/robot/reloadStrategy,/trading/robot/setTakeProfitSwitch,/trading/robot/riskConfig,/trading/robot/riskConfig/save,/toogo/wallet/overview',
  `updated_at` = NOW()
WHERE `path` = '/toogo/robot' OR `component` = '/toogo/robot/index';

-- 创建机器人（隐藏页）/toogo/robot/create
UPDATE `hg_admin_menu`
SET
  `permissions` = '/trading/apiConfig/list,/trading/robot/create,/strategy/group/list',
  `updated_at` = NOW()
WHERE `path` = '/toogo/robot/create' OR `component` = '/toogo/robot/create';

-- ============ 策略模板 /toogo/strategy ============
UPDATE `hg_admin_menu`
SET
  `permissions` = '/strategy/group/list,/strategy/group/create,/strategy/group/update,/strategy/group/delete,/strategy/group/initStrategies,/strategy/group/copyFromOfficial,/strategy/group/setDefault,/strategy/template/list,/strategy/template/create,/strategy/template/update,/strategy/template/delete',
  `updated_at` = NOW()
WHERE `path` = '/toogo/strategy' OR `component` = '/toogo/strategy/index';

-- 策略列表（隐藏页）/toogo/strategy/list
UPDATE `hg_admin_menu`
SET
  `permissions` = '/strategy/template/list,/strategy/template/create,/strategy/template/update,/strategy/template/delete,/strategy/group/initStrategies',
  `updated_at` = NOW()
WHERE `path` = '/toogo/strategy/list' OR `component` = '/toogo/strategy/list';

-- ============ API配置 /toogo/api ============
UPDATE `hg_admin_menu`
SET
  `permissions` = '/trading/apiConfig/list,/trading/apiConfig/create,/trading/apiConfig/update,/trading/apiConfig/delete,/trading/apiConfig/view,/trading/apiConfig/test,/trading/apiConfig/setDefault,/trading/apiConfig/platforms',
  `updated_at` = NOW()
WHERE `path` = '/toogo/api' OR `component` = '/toogo/api/index';

-- ============ 钱包中心 ============
-- 资产总览 /toogo/wallet/overview
UPDATE `hg_admin_menu`
SET
  `permissions` = '/toogo/wallet/overview,/toogo/wallet/log/list,/toogo/wallet/transfer',
  `updated_at` = NOW()
WHERE `path` = '/toogo/wallet/overview' OR `component` = '/toogo/finance/index';

-- 订阅套餐 /toogo/wallet/subscription
UPDATE `hg_admin_menu`
SET
  `permissions` = '/toogo/plan/list,/toogo/subscription/subscribe,/toogo/subscription/my,/toogo/subscription/list',
  `updated_at` = NOW()
WHERE `path` = '/toogo/wallet/subscription' OR `component` = '/toogo/subscription/index';

-- 交易明细 /toogo/wallet/order-history（若未执行 20251217 脚本，这里兜底补齐）
UPDATE `hg_admin_menu`
SET
  `permissions` = '/toogo/wallet/trade/history,/toogo/wallet/run-session/summary,/toogo/wallet/run-session/sync,/trading/robot/list',
  `updated_at` = NOW()
WHERE `path` = '/toogo/wallet/order-history' OR `component` = '/toogo/wallet/order-history';

-- ============ 推广中心 ============
-- 我的团队 /toogo/promote/team
UPDATE `hg_admin_menu`
SET
  `permissions` = '/toogo/user/team/list,/toogo/user/team/stat,/toogo/user/info',
  `updated_at` = NOW()
WHERE `path` = '/toogo/promote/team' OR `component` = '/toogo/team/index';

-- 佣金记录 /toogo/promote/commission
UPDATE `hg_admin_menu`
SET
  `permissions` = '/toogo/commission/log/list,/toogo/commission/stat',
  `updated_at` = NOW()
WHERE `path` = '/toogo/promote/commission' OR `component` = '/toogo/commission/index';

SELECT 'OK' AS result;

