-- ============================================================
-- 增加「推广-我的邀请码」菜单（用户端）
-- ============================================================

SET NAMES utf8mb4;

-- 获取推广中心父菜单ID
SET @promotion_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo/promotion' LIMIT 1);

-- 如果推广中心父菜单不存在，先创建
INSERT INTO `hg_admin_menu` (
  `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
  `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
  `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
  `remark`, `status`, `created_at`, `updated_at`
)
SELECT
  0,
  '推广中心',
  'ToogoPromotion',
  '/toogo/promotion',
  'GiftOutlined',
  1,
  '',
  '',
  '推广中心',
  'Layout',
  1,
  '',
  0,
  0,
  '',
  0,
  0,
  0,
  40,
  'Toogo推广中心',
  1,
  NOW(),
  NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/promotion' LIMIT 1);

-- 重新获取推广中心父菜单ID
SET @promotion_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo/promotion' LIMIT 1);

-- 如果已存在"我的邀请码"菜单，则更新
UPDATE `hg_admin_menu`
SET
  `title` = '我的邀请码',
  `permission_name` = '我的邀请码',
  `component` = '/toogo/invite-code/index',
  `icon` = 'KeyOutlined',
  `permissions` = '/toogo/user/info,/toogo/user/refresh-invite-code',
  `updated_at` = NOW()
WHERE `path` = '/toogo/invite-code';

-- 如果不存在，则插入（仅当推广中心父菜单存在时）
INSERT INTO `hg_admin_menu` (
  `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
  `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
  `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
  `remark`, `status`, `created_at`, `updated_at`
)
SELECT
  @promotion_parent_id,
  '我的邀请码',
  'ToogoInviteCode',
  '/toogo/invite-code',
  'KeyOutlined',
  2,
  '',
  '/toogo/user/info,/toogo/user/refresh-invite-code',
  '我的邀请码',
  '/toogo/invite-code/index',
  0,
  '',
  0,
  0,
  '',
  1,
  0,
  0,
  10,
  '推广中心-我的邀请码',
  1,
  NOW(),
  NOW()
FROM DUAL
WHERE @promotion_parent_id IS NOT NULL
  AND @promotion_parent_id > 0
  AND NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/invite-code' LIMIT 1);

SELECT 'OK' AS result;
