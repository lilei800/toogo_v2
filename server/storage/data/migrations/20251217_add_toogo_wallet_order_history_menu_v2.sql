-- ============================================================
-- 增加「钱包中心-交易明细」菜单（用户端）
-- ============================================================

SET NAMES utf8mb4;

-- 获取钱包中心父菜单ID
SET @wallet_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo/wallet' LIMIT 1);

-- 如果已存在该菜单，则更新
UPDATE `hg_admin_menu`
SET
  `title` = '交易明细',
  `permission_name` = '交易明细',
  `component` = '/toogo/wallet/order-history',
  `icon` = 'FileTextOutlined',
  `permissions` = '/toogo/wallet/trade/history,/toogo/wallet/run-session/summary,/toogo/wallet/run-session/sync,/trading/robot/list',
  `updated_at` = NOW()
WHERE `path` = '/toogo/wallet/order-history';

-- 如果不存在，则插入（仅当钱包中心父菜单存在时）
INSERT INTO `hg_admin_menu` (
  `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
  `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
  `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
  `remark`, `status`, `created_at`, `updated_at`
)
SELECT
  @wallet_parent_id,
  '交易明细',
  'WalletOrderHistory',
  '/toogo/wallet/order-history',
  'FileTextOutlined',
  2,
  '',
  '/toogo/wallet/trade/history,/toogo/wallet/run-session/summary,/toogo/wallet/run-session/sync,/trading/robot/list',
  '交易明细',
  '/toogo/wallet/order-history',
  0,
  '',
  0,
  0,
  '',
  1,
  0,
  0,
  53,
  '钱包中心-用户交易明细',
  1,
  NOW(),
  NOW()
FROM DUAL
WHERE @wallet_parent_id IS NOT NULL
  AND @wallet_parent_id > 0
  AND NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/wallet/order-history' LIMIT 1);

SELECT 'OK' AS result;
