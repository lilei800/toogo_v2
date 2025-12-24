-- ============================================================
-- 优化量化交易用户菜单（简化版）
-- 父菜单ID: 2000
-- ============================================================

-- 1. 更新控制台排序
UPDATE `hg_admin_menu` SET `sort` = 10 WHERE `id` = 2001;

-- 2. 更新我的机器人排序，取消隐藏
UPDATE `hg_admin_menu` SET `sort` = 20, `hidden` = 0 WHERE `id` = 2003;

-- 3. 更新策略管理
UPDATE `hg_admin_menu` SET 
    `sort` = 30, 
    `title` = '策略模板',
    `component` = '/toogo/strategy/index'
WHERE `id` = 2526;

-- 4. 更新API管理
UPDATE `hg_admin_menu` SET `sort` = 40, `title` = 'API配置' WHERE `id` = 2510;

-- 5. 创建钱包中心父菜单
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`,
    `component`, `hidden`, `sort`, `status`, `created_at`, `updated_at`
) VALUES (
    2000, '钱包中心', 'ToogoWallet', 'wallet', 'WalletOutlined', 1, '/toogo/wallet/overview',
    'LAYOUT', 0, 50, 1, NOW(), NOW()
) ON DUPLICATE KEY UPDATE `sort` = 50, `title` = '钱包中心';

-- 获取钱包中心ID
SET @wallet_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = 'wallet' AND `pid` = 2000 LIMIT 1);

-- 5.1 创建资产总览
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`,
    `component`, `hidden`, `sort`, `status`, `created_at`, `updated_at`
) VALUES (
    @wallet_id, '资产总览', 'WalletOverview', 'overview', 'DashboardOutlined', 2,
    '/toogo/finance/index', 0, 51, 1, NOW(), NOW()
) ON DUPLICATE KEY UPDATE `sort` = 51;

-- 5.2 创建订阅套餐
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`,
    `component`, `hidden`, `sort`, `status`, `created_at`, `updated_at`
) VALUES (
    @wallet_id, '订阅套餐', 'WalletSubscription', 'subscription', 'CreditCardOutlined', 2,
    '/toogo/subscription/index', 0, 52, 1, NOW(), NOW()
) ON DUPLICATE KEY UPDATE `sort` = 52;

-- 6. 创建推广中心父菜单
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`,
    `component`, `hidden`, `sort`, `status`, `created_at`, `updated_at`
) VALUES (
    2000, '推广中心', 'ToogoPromote', 'promote', 'TeamOutlined', 1, '/toogo/promote/team',
    'LAYOUT', 0, 60, 1, NOW(), NOW()
) ON DUPLICATE KEY UPDATE `sort` = 60, `title` = '推广中心';

-- 获取推广中心ID
SET @promote_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = 'promote' AND `pid` = 2000 LIMIT 1);

-- 6.1 创建我的团队
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`,
    `component`, `hidden`, `sort`, `status`, `created_at`, `updated_at`
) VALUES (
    @promote_id, '我的团队', 'PromoteTeam', 'team', 'TeamOutlined', 2,
    '/toogo/team/index', 0, 61, 1, NOW(), NOW()
) ON DUPLICATE KEY UPDATE `sort` = 61;

-- 6.2 创建佣金记录
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`,
    `component`, `hidden`, `sort`, `status`, `created_at`, `updated_at`
) VALUES (
    @promote_id, '佣金记录', 'PromoteCommission', 'commission', 'RiseOutlined', 2,
    '/toogo/commission/index', 0, 62, 1, NOW(), NOW()
) ON DUPLICATE KEY UPDATE `sort` = 62;

-- 7. 隐藏旧的一级菜单
UPDATE `hg_admin_menu` SET `hidden` = 1, `sort` = 999 WHERE `id` = 2002;  -- 订阅套餐
UPDATE `hg_admin_menu` SET `hidden` = 1, `sort` = 999 WHERE `id` = 2511;  -- 财务中心
UPDATE `hg_admin_menu` SET `hidden` = 1, `sort` = 999 WHERE `id` = 2005;  -- 我的团队
UPDATE `hg_admin_menu` SET `hidden` = 1, `sort` = 999 WHERE `id` = 2006;  -- 佣金记录
UPDATE `hg_admin_menu` SET `hidden` = 1, `sort` = 999 WHERE `id` = 2530;  -- 预警监控（如需要）

-- 8. 验证结果
SELECT id, pid, title, path, sort, hidden FROM `hg_admin_menu` WHERE pid = 2000 OR pid IN (SELECT id FROM `hg_admin_menu` WHERE pid = 2000) ORDER BY pid, sort;

