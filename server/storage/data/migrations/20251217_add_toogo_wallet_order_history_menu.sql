-- ============================================================
-- 增加「钱包中心-交易明细」菜单（用户端）
-- 说明：
-- - 前端路由已存在：/toogo/wallet/order-history -> /views/toogo/wallet/order-history.vue
-- - 但生产环境通常以数据库菜单为准（权限管理-菜单管理），需要补齐该菜单项
-- ============================================================

SET NAMES utf8mb4;

-- 获取量化交易父菜单ID
SET @toogo_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo' LIMIT 1);

-- 获取钱包中心父菜单ID（若不存在，可先执行 optimize_toogo_user_menu.sql 创建）
SET @wallet_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo/wallet' LIMIT 1);

-- 插入「交易明细」菜单（幂等）
INSERT INTO `hg_admin_menu` (
  `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
  `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
  `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
  `remark`, `status`, `created_at`, `updated_at`
)
SELECT
  @wallet_parent_id AS pid,
  '交易明细' AS title,
  'WalletOrderHistory' AS name,
  '/toogo/wallet/order-history' AS path,
  'FileTextOutlined' AS icon,
  2 AS type,
  '' AS redirect,
  -- 页面会调用的接口（用于 Casbin）
  '/toogo/wallet/trade/history,/toogo/wallet/run-session/summary,/toogo/wallet/run-session/sync,/trading/robot/list' AS permissions,
  '交易明细' AS permission_name,
  '/toogo/wallet/order-history' AS component,
  0 AS always_show,
  '' AS active_menu,
  0 AS is_root,
  0 AS is_frame,
  '' AS frame_src,
  1 AS keep_alive,
  0 AS hidden,
  0 AS affix,
  53 AS sort,
  '钱包中心-用户交易明细（闭环主表 hg_trading_order）' AS remark,
  1 AS status,
  NOW() AS created_at,
  NOW() AS updated_at
WHERE
  @wallet_parent_id IS NOT NULL
  AND @wallet_parent_id > 0
  AND NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/wallet/order-history' LIMIT 1);

-- 若已存在该菜单（之前插入为“历史订单”），统一修正为“交易明细”
UPDATE `hg_admin_menu`
SET
  `title` = '交易明细',
  `permission_name` = '交易明细',
  `component` = '/toogo/wallet/order-history',
  `icon` = 'FileTextOutlined',
  `permissions` = '/toogo/wallet/trade/history,/toogo/wallet/run-session/summary,/toogo/wallet/run-session/sync,/trading/robot/list',
  `updated_at` = NOW()
WHERE `path` = '/toogo/wallet/order-history';

SELECT 'OK' AS result, @toogo_parent_id AS toogo_parent_id, @wallet_parent_id AS wallet_parent_id;


