-- ===========================================
-- Toogo 数据库性能优化索引
-- 执行前请备份数据库
-- ===========================================

-- ===========================================
-- 交易机器人表索引
-- ===========================================
-- 用户+状态复合索引（用于查询用户的运行中机器人）
ALTER TABLE hg_trading_robot ADD INDEX idx_user_status (user_id, status);

-- 策略组+状态索引（用于按策略查询机器人）
ALTER TABLE hg_trading_robot ADD INDEX idx_strategy_status (strategy_group_id, status);

-- 创建时间索引（用于时间范围查询）
ALTER TABLE hg_trading_robot ADD INDEX idx_created_at (created_at);

-- 更新时间索引（用于同步和排序）
ALTER TABLE hg_trading_robot ADD INDEX idx_updated_at (updated_at);


-- ===========================================
-- 交易订单表索引
-- ===========================================
-- 机器人+状态复合索引（查询机器人的订单）
ALTER TABLE hg_trading_order ADD INDEX idx_robot_status (robot_id, status);

-- 用户+状态索引（查询用户的订单）
ALTER TABLE hg_trading_order ADD INDEX idx_user_status (user_id, status);

-- 交易对+状态索引（按交易对查询）
ALTER TABLE hg_trading_order ADD INDEX idx_symbol_status (symbol, status);

-- 创建时间索引
ALTER TABLE hg_trading_order ADD INDEX idx_created_at (created_at);

-- 订单ID索引（用于快速查找）
ALTER TABLE hg_trading_order ADD INDEX idx_order_id (order_id);


-- ===========================================
-- Toogo用户表索引
-- ===========================================
-- 会员ID索引（关联查询）
ALTER TABLE hg_toogo_user ADD INDEX idx_member_id (member_id);

-- 邀请码索引（邀请注册查询）
ALTER TABLE hg_toogo_user ADD INDEX idx_invite_code (invite_code);

-- 父级用户索引（团队树查询）
ALTER TABLE hg_toogo_user ADD INDEX idx_parent_id (parent_id);

-- VIP等级索引
ALTER TABLE hg_toogo_user ADD INDEX idx_vip_level (vip_level);


-- ===========================================
-- Toogo钱包表索引
-- ===========================================
-- 用户ID索引
ALTER TABLE hg_toogo_wallet ADD INDEX idx_user_id (user_id);

-- 创建唯一索引（一个用户一个钱包）
ALTER TABLE hg_toogo_wallet ADD UNIQUE INDEX uk_user_id (user_id);


-- ===========================================
-- 钱包流水表索引
-- ===========================================
-- 用户+时间复合索引
ALTER TABLE hg_toogo_wallet_log ADD INDEX idx_user_time (user_id, created_at);

-- 用户+类型索引
ALTER TABLE hg_toogo_wallet_log ADD INDEX idx_user_type (user_id, type);

-- 关联订单索引
ALTER TABLE hg_toogo_wallet_log ADD INDEX idx_related_id (related_id);


-- ===========================================
-- 充值订单表索引
-- ===========================================
-- 用户+状态索引
ALTER TABLE hg_toogo_deposit ADD INDEX idx_user_status (user_id, status);

-- 订单号索引
ALTER TABLE hg_toogo_deposit ADD INDEX idx_order_no (order_no);

-- 支付ID索引（回调查询）
ALTER TABLE hg_toogo_deposit ADD INDEX idx_payment_id (payment_id);


-- ===========================================
-- 提现订单表索引
-- ===========================================
-- 用户+状态索引
ALTER TABLE hg_toogo_withdraw ADD INDEX idx_user_status (user_id, status);

-- 订单号索引
ALTER TABLE hg_toogo_withdraw ADD INDEX idx_order_no (order_no);

-- 审核时间索引
ALTER TABLE hg_toogo_withdraw ADD INDEX idx_audit_time (audit_time);


-- ===========================================
-- 订阅记录表索引
-- ===========================================
-- 用户+状态索引
ALTER TABLE hg_toogo_subscription ADD INDEX idx_user_status (user_id, status);

-- 过期时间索引（定时任务查询）
ALTER TABLE hg_toogo_subscription ADD INDEX idx_expire_time (expire_time);


-- ===========================================
-- 佣金记录表索引
-- ===========================================
-- 用户+时间索引
ALTER TABLE hg_toogo_commission_log ADD INDEX idx_user_time (user_id, created_at);

-- 来源用户索引
ALTER TABLE hg_toogo_commission_log ADD INDEX idx_from_user (from_user_id);

-- 类型索引
ALTER TABLE hg_toogo_commission_log ADD INDEX idx_type (type);


-- ===========================================
-- 策略模板表索引
-- ===========================================
-- 策略组索引
ALTER TABLE hg_toogo_strategy_template ADD INDEX idx_group_id (group_id);

-- 状态索引
ALTER TABLE hg_toogo_strategy_template ADD INDEX idx_status (status);

-- 用户索引（用户自定义策略）
ALTER TABLE hg_toogo_strategy_template ADD INDEX idx_user_id (user_id);


-- ===========================================
-- 算力消耗表索引
-- ===========================================
-- 用户+时间索引
ALTER TABLE hg_toogo_power_consume ADD INDEX idx_user_time (user_id, created_at);

-- 机器人索引
ALTER TABLE hg_toogo_power_consume ADD INDEX idx_robot_id (robot_id);


-- ===========================================
-- API配置表索引
-- ===========================================
-- 用户+平台索引
ALTER TABLE hg_trading_api_config ADD INDEX idx_user_platform (user_id, platform);

-- 默认配置索引
ALTER TABLE hg_trading_api_config ADD INDEX idx_user_default (user_id, is_default);


-- ===========================================
-- 策略组表索引
-- ===========================================
-- 用户索引
ALTER TABLE hg_trading_strategy_group ADD INDEX idx_user_id (user_id);

-- 类型索引
ALTER TABLE hg_trading_strategy_group ADD INDEX idx_type (type);

-- 状态索引
ALTER TABLE hg_trading_strategy_group ADD INDEX idx_status (status);


-- ===========================================
-- 系统日志表优化
-- ===========================================
-- 日志表分区（可选，针对大数据量）
-- 创建时间索引
ALTER TABLE hg_sys_log ADD INDEX idx_created_at (created_at);

-- 用户+时间索引
ALTER TABLE hg_sys_log ADD INDEX idx_member_time (member_id, created_at);


-- ===========================================
-- 查询优化视图（可选）
-- ===========================================

-- 用户资产概览视图
CREATE OR REPLACE VIEW v_user_asset_overview AS
SELECT 
    u.member_id,
    u.vip_level,
    u.active_robot_count,
    u.robot_limit,
    w.balance,
    w.frozen,
    w.power,
    w.gift_power,
    w.commission
FROM hg_toogo_user u
LEFT JOIN hg_toogo_wallet w ON u.member_id = w.user_id;

-- 机器人运行状态视图
CREATE OR REPLACE VIEW v_robot_status_overview AS
SELECT 
    r.id,
    r.name,
    r.user_id,
    r.status,
    r.symbol,
    r.leverage,
    r.total_profit,
    r.total_trades,
    r.win_rate,
    sg.name as strategy_name
FROM hg_trading_robot r
LEFT JOIN hg_trading_strategy_group sg ON r.strategy_group_id = sg.id;


-- ===========================================
-- 统计信息更新（定期执行）
-- ===========================================
-- 更新表统计信息以优化查询计划
ANALYZE TABLE hg_trading_robot;
ANALYZE TABLE hg_trading_order;
ANALYZE TABLE hg_toogo_user;
ANALYZE TABLE hg_toogo_wallet;
ANALYZE TABLE hg_toogo_wallet_log;
ANALYZE TABLE hg_toogo_deposit;
ANALYZE TABLE hg_toogo_withdraw;
ANALYZE TABLE hg_toogo_subscription;
ANALYZE TABLE hg_toogo_commission_log;


-- ===========================================
-- 完成提示
-- ===========================================
SELECT '性能索引创建完成！' AS message;

