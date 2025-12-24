-- 性能优化：数据库索引优化
-- 创建时间：2024-11-26
-- 说明：这些索引可以显著提升查询性能，建议在生产环境中执行

-- ==================== 机器人表索引 ====================

-- 优化：按状态和用户查询（机器人列表）
CREATE INDEX IF NOT EXISTS `idx_robot_status_user` ON `hg_trading_robot`(`status`, `user_id`, `tenant_id`);

-- 优化：按创建时间查询（统计分析）
CREATE INDEX IF NOT EXISTS `idx_robot_created` ON `hg_trading_robot`(`created_at` DESC);

-- 优化：联合查询（状态+风险偏好+市场状态）
CREATE INDEX IF NOT EXISTS `idx_robot_query` ON `hg_trading_robot`(`status`, `risk_preference`, `market_state`);

-- ==================== 订单表索引 ====================

-- 优化：按机器人ID和状态查询（持仓订单）
CREATE INDEX IF NOT EXISTS `idx_order_robot_status` ON `hg_trading_order`(`robot_id`, `status`);

-- 优化：按用户ID和状态查询（用户订单列表）
CREATE INDEX IF NOT EXISTS `idx_order_user_status` ON `hg_trading_order`(`user_id`, `status`, `tenant_id`);

-- 优化：按时间范围查询（历史订单）
CREATE INDEX IF NOT EXISTS `idx_order_time` ON `hg_trading_order`(`open_time` DESC, `close_time` DESC);

-- 优化：按交易对查询（行情分析）
CREATE INDEX IF NOT EXISTS `idx_order_symbol` ON `hg_trading_order`(`symbol`, `status`);

-- 优化：盈亏统计查询
CREATE INDEX IF NOT EXISTS `idx_order_profit` ON `hg_trading_order`(`user_id`, `status`, `realized_profit`);

-- ==================== 平仓日志表索引 ====================

-- 优化：按机器人ID查询（机器人历史）
CREATE INDEX IF NOT EXISTS `idx_closelog_robot` ON `hg_trading_close_log`(`robot_id`, `close_time` DESC);

-- 优化：按用户ID查询（用户历史）
CREATE INDEX IF NOT EXISTS `idx_closelog_user` ON `hg_trading_close_log`(`user_id`, `close_time` DESC);

-- 优化：按交易对查询（交易对统计）
CREATE INDEX IF NOT EXISTS `idx_closelog_symbol` ON `hg_trading_close_log`(`symbol`, `close_time` DESC);

-- 优化：按平仓原因统计
CREATE INDEX IF NOT EXISTS `idx_closelog_reason` ON `hg_trading_close_log`(`close_reason`, `close_time` DESC);

-- ==================== API配置表索引 ====================

-- 优化：按用户和平台查询
CREATE INDEX IF NOT EXISTS `idx_apiconfig_user_platform` ON `hg_trading_api_config`(`user_id`, `platform`, `status`);

-- 优化：查找默认配置
CREATE INDEX IF NOT EXISTS `idx_apiconfig_default` ON `hg_trading_api_config`(`user_id`, `is_default`, `status`);

-- ==================== 监控日志表索引 ====================

-- 优化：按机器人ID查询
CREATE INDEX IF NOT EXISTS `idx_monitorlog_robot` ON `hg_trading_monitor_log`(`robot_id`, `created_at` DESC);

-- 优化：按信号类型统计
CREATE INDEX IF NOT EXISTS `idx_monitorlog_signal` ON `hg_trading_monitor_log`(`signal_type`, `created_at` DESC);

-- ==================== 策略模板表索引 ====================

-- 优化：按类型查询
CREATE INDEX IF NOT EXISTS `idx_strategy_type` ON `hg_trading_strategy_template`(`risk_preference`, `market_state`, `status`);

-- ==================== 查看索引 ====================

-- 查看所有索引
SHOW INDEX FROM `hg_trading_robot`;
SHOW INDEX FROM `hg_trading_order`;
SHOW INDEX FROM `hg_trading_close_log`;
SHOW INDEX FROM `hg_trading_api_config`;
SHOW INDEX FROM `hg_trading_monitor_log`;
SHOW INDEX FROM `hg_trading_strategy_template`;

-- ==================== 性能分析 ====================

-- 分析表统计信息（执行后可以提升查询优化器的效率）
ANALYZE TABLE `hg_trading_robot`;
ANALYZE TABLE `hg_trading_order`;
ANALYZE TABLE `hg_trading_close_log`;
ANALYZE TABLE `hg_trading_api_config`;
ANALYZE TABLE `hg_trading_monitor_log`;
ANALYZE TABLE `hg_trading_strategy_template`;

-- ==================== 使用说明 ====================
/*
执行步骤：
1. 备份数据库
2. 在测试环境先执行
3. 观察索引创建时间（大表可能需要几分钟）
4. 验证查询性能提升
5. 在生产环境执行

预期效果：
- 订单列表查询：提升 10-50倍
- 机器人检查：提升 5-20倍
- 统计分析：提升 20-100倍
- 历史数据查询：提升 30-100倍

注意事项：
- 索引会占用额外的磁盘空间（约10-20%）
- 索引会略微降低写入性能（约5-10%）
- 但读取性能提升远大于写入性能损失
- 对于读多写少的交易系统，收益巨大

监控建议：
- 使用 EXPLAIN 分析查询计划
- 监控慢查询日志
- 定期执行 ANALYZE TABLE 更新统计信息
*/

