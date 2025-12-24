-- ===========================================
-- 自动根据菜单名称匹配图标
-- 执行前请先备份: CREATE TABLE hg_admin_menu_backup AS SELECT * FROM hg_admin_menu;
-- ===========================================
SET NAMES utf8mb4;

-- ============ 仪表盘/控制台 ============
UPDATE hg_admin_menu SET icon = 'DashboardOutlined' WHERE (title LIKE '%仪表盘%' OR title LIKE '%Dashboard%' OR title LIKE '%控制台%' OR title LIKE '%Console%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'DesktopOutlined' WHERE (title LIKE '%主控台%' OR title LIKE '%工作台%' OR title LIKE '%Workplace%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 用户/会员/人员 ============
UPDATE hg_admin_menu SET icon = 'UserOutlined' WHERE (title LIKE '%用户%' OR title LIKE '%User%' OR title LIKE '%会员%' OR title LIKE '%Member%' OR title LIKE '%后台用户%' OR title LIKE '%在线用户%') AND title NOT LIKE '%团队%' AND title NOT LIKE '%管理员%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'TeamOutlined' WHERE (title LIKE '%团队%' OR title LIKE '%Team%' OR title LIKE '%岗位%' OR title LIKE '%Post%' OR title LIKE '%代理%' OR title LIKE '%Agent%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'ApartmentOutlined' WHERE (title LIKE '%部门%' OR title LIKE '%Dept%' OR title LIKE '%组织%' OR title LIKE '%Organization%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 机器人/量化交易 ============
UPDATE hg_admin_menu SET icon = 'RobotOutlined' WHERE (title LIKE '%机器人%' OR title LIKE '%Robot%' OR title LIKE '%自动%' OR title LIKE '%Auto%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'BarChartOutlined' WHERE (title LIKE '%量化%' OR title LIKE '%交易%' OR title LIKE '%Trading%' OR title LIKE '%Toogo%') AND title NOT LIKE '%配置%' AND title NOT LIKE '%记录%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'LineChartOutlined' WHERE (title LIKE '%波动%' OR title LIKE '%Volatility%' OR title LIKE '%K线%' OR title LIKE '%行情%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'SwapOutlined' WHERE (title LIKE '%交易对%' OR title LIKE '%Symbol%' OR title LIKE '%币种%' OR title LIKE '%货币%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 策略/模板 ============
UPDATE hg_admin_menu SET icon = 'BulbOutlined' WHERE (title LIKE '%策略%' OR title LIKE '%Strategy%') AND title NOT LIKE '%我的%' AND title NOT LIKE '%官方%' AND title NOT LIKE '%排行%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'FileTextOutlined' WHERE (title LIKE '%我的策略%' OR title LIKE '%模板%' OR title LIKE '%Template%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'StarOutlined' WHERE (title LIKE '%官方%' OR title LIKE '%Official%' OR title LIKE '%VIP%' OR title LIKE '%推荐%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'RiseOutlined' WHERE (title LIKE '%排行%' OR title LIKE '%Ranking%' OR title LIKE '%盈利%' OR title LIKE '%Profit%' OR title LIKE '%佣金%' OR title LIKE '%Commission%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ API/接口 ============
UPDATE hg_admin_menu SET icon = 'ApiOutlined' WHERE (title LIKE '%API%' OR title LIKE '%接口%' OR title LIKE '%Interface%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 财务/资金 ============
UPDATE hg_admin_menu SET icon = 'WalletOutlined' WHERE (title LIKE '%财务%' OR title LIKE '%Finance%' OR title LIKE '%提现%' OR title LIKE '%Withdraw%' OR title LIKE '%钱包%' OR title LIKE '%Wallet%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'DollarOutlined' WHERE (title LIKE '%充值%' OR title LIKE '%Deposit%' OR title LIKE '%付款%' OR title LIKE '%Payment%' OR title LIKE '%金额%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'CreditCardOutlined' WHERE (title LIKE '%套餐%' OR title LIKE '%Plan%' OR title LIKE '%订阅%' OR title LIKE '%Subscription%' OR title LIKE '%支付%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 权限/安全 ============
UPDATE hg_admin_menu SET icon = 'SafetyCertificateOutlined' WHERE (title LIKE '%权限%' OR title LIKE '%Permission%' OR title LIKE '%授权%') AND title NOT LIKE '%菜单%' AND title NOT LIKE '%角色%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'MenuOutlined' WHERE (title LIKE '%菜单%' OR title LIKE '%Menu%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'LockOutlined' WHERE (title LIKE '%角色%' OR title LIKE '%Role%' OR title LIKE '%密码%' OR title LIKE '%Password%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 系统设置 ============
UPDATE hg_admin_menu SET icon = 'SettingOutlined' WHERE (title LIKE '%设置%' OR title LIKE '%Setting%' OR title LIKE '%配置%' OR title LIKE '%Config%' OR title LIKE '%系统设置%') AND title NOT LIKE '%定时%' AND title NOT LIKE '%字典%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'BookOutlined' WHERE (title LIKE '%字典%' OR title LIKE '%Dict%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'ClockCircleOutlined' WHERE (title LIKE '%定时%' OR title LIKE '%Cron%' OR title LIKE '%任务%' OR title LIKE '%Task%' OR title LIKE '%调度%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'StopOutlined' WHERE (title LIKE '%黑名单%' OR title LIKE '%Blacklist%' OR title LIKE '%禁止%' OR title LIKE '%封禁%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 系统监控 ============
UPDATE hg_admin_menu SET icon = 'MonitorOutlined' WHERE (title LIKE '%监控%' OR title LIKE '%Monitor%' OR title LIKE '%服务监控%') AND title NOT LIKE '%日志%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'FileTextOutlined' WHERE (title LIKE '%日志%' OR title LIKE '%Log%' OR title LIKE '%记录%' OR title LIKE '%History%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'GlobalOutlined' WHERE (title LIKE '%在线服务%' OR title LIKE '%服务%' OR title LIKE '%Service%' OR title LIKE '%网络%' OR title LIKE '%地区%' OR title LIKE '%省市%') AND title NOT LIKE '%监控%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 系统应用 ============
UPDATE hg_admin_menu SET icon = 'AppstoreOutlined' WHERE (title LIKE '%应用%' OR title LIKE '%App%' OR title LIKE '%组织%' OR title LIKE '%Org%') AND title NOT LIKE '%附件%' AND title NOT LIKE '%公告%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'CloudUploadOutlined' WHERE (title LIKE '%附件%' OR title LIKE '%Attachment%' OR title LIKE '%上传%' OR title LIKE '%Upload%' OR title LIKE '%文件%' OR title LIKE '%File%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'NotificationOutlined' WHERE (title LIKE '%公告%' OR title LIKE '%Notice%' OR title LIKE '%通知%' OR title LIKE '%消息%' OR title LIKE '%Message%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 开发工具 ============
UPDATE hg_admin_menu SET icon = 'CodeOutlined' WHERE (title LIKE '%开发%' OR title LIKE '%Develop%' OR title LIKE '%代码%' OR title LIKE '%Code%' OR title LIKE '%生成%' OR title LIKE '%Generate%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'BlockOutlined' WHERE (title LIKE '%插件%' OR title LIKE '%Plugin%' OR title LIKE '%Addon%' OR title LIKE '%扩展%' OR title LIKE '%Extension%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'DatabaseOutlined' WHERE (title LIKE '%数据库%' OR title LIKE '%Database%' OR title LIKE '%数据%' OR title LIKE '%Data%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'ToolOutlined' WHERE (title LIKE '%工具%' OR title LIKE '%Tool%' OR title LIKE '%维护%' OR title LIKE '%Maintain%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'BugOutlined' WHERE (title LIKE '%调试%' OR title LIKE '%Debug%' OR title LIKE '%测试%' OR title LIKE '%Test%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 其他 ============
UPDATE hg_admin_menu SET icon = 'ThunderboltOutlined' WHERE (title LIKE '%算力%' OR title LIKE '%Power%' OR title LIKE '%性能%' OR title LIKE '%Performance%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'ProjectOutlined' WHERE (title LIKE '%关于%' OR title LIKE '%About%' OR title LIKE '%详情%' OR title LIKE '%Detail%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'InfoCircleOutlined' WHERE (title LIKE '%信息%' OR title LIKE '%Info%' OR title LIKE '%说明%' OR title LIKE '%帮助%' OR title LIKE '%Help%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'HomeOutlined' WHERE (title LIKE '%首页%' OR title LIKE '%Home%' OR title LIKE '%主页%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'ControlOutlined' WHERE (title LIKE '%管理%' OR title LIKE '%Admin%' OR title LIKE '%管理后台%') AND title NOT LIKE '%用户%' AND title NOT LIKE '%套餐%' AND title NOT LIKE '%策略%' AND title NOT LIKE '%提现%' AND title NOT LIKE '%充值%' AND title NOT LIKE '%菜单%' AND title NOT LIKE '%字典%' AND title NOT LIKE '%岗位%' AND title NOT LIKE '%部门%' AND title NOT LIKE '%附件%' AND title NOT LIKE '%插件%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'EyeOutlined' WHERE (title LIKE '%查看%' OR title LIKE '%View%' OR title LIKE '%预览%' OR title LIKE '%Preview%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'EditOutlined' WHERE (title LIKE '%编辑%' OR title LIKE '%Edit%' OR title LIKE '%修改%' OR title LIKE '%Update%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'PlusOutlined' WHERE (title LIKE '%新增%' OR title LIKE '%Add%' OR title LIKE '%创建%' OR title LIKE '%Create%' OR title LIKE '%新建%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'DeleteOutlined' WHERE (title LIKE '%删除%' OR title LIKE '%Delete%' OR title LIKE '%移除%' OR title LIKE '%Remove%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 个人中心 ============
UPDATE hg_admin_menu SET icon = 'UserOutlined' WHERE (title LIKE '%个人%' OR title LIKE '%Personal%' OR title LIKE '%账户%' OR title LIKE '%Account%' OR title LIKE '%我的%') AND title NOT LIKE '%策略%' AND title NOT LIKE '%团队%' AND title NOT LIKE '%机器人%' AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');
UPDATE hg_admin_menu SET icon = 'MailOutlined' WHERE (title LIKE '%邮件%' OR title LIKE '%Email%' OR title LIKE '%Mail%') AND (icon IS NULL OR icon = '' OR icon NOT LIKE '%Outlined%');

-- ============ 为空图标设置默认值 ============
UPDATE hg_admin_menu SET icon = 'AppstoreOutlined' WHERE (icon IS NULL OR icon = '') AND type = 1;

-- 查看更新结果
SELECT id, title, icon, path FROM hg_admin_menu WHERE status = 1 ORDER BY pid, sort;

