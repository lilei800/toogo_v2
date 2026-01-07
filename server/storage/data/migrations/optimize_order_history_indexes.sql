-- 优化订单历史查询索引
-- 目标：优化钱包-交易明细页面查询性能，减少数据库扫描
-- 执行时间：2025-12-25

-- ========== 订单表索引优化 ==========

-- 1. 用户+状态+平仓时间组合索引（交易明细页核心查询）
-- 适用于: WHERE user_id=? AND status=? ORDER BY close_time
ALTER TABLE `hg_trading_order` 
ADD INDEX IF NOT EXISTS `idx_user_status_closetime` (`user_id`, `status`, `close_time`);

-- 2. 机器人+状态组合索引（按机器人筛选）
-- 适用于: WHERE robot_id=? AND status=?
ALTER TABLE `hg_trading_order` 
ADD INDEX IF NOT EXISTS `idx_robot_status` (`robot_id`, `status`);

-- 3. 交易所+交易对组合索引（按交易对筛选）
-- 适用于: WHERE exchange=? AND symbol=?
ALTER TABLE `hg_trading_order` 
ADD INDEX IF NOT EXISTS `idx_exchange_symbol` (`exchange`, `symbol`);

-- 4. 创建时间索引（按时间范围查询）
-- 适用于: WHERE created_at BETWEEN ? AND ?
ALTER TABLE `hg_trading_order` 
ADD INDEX IF NOT EXISTS `idx_created_at` (`created_at`);

-- ========== 成交流水表索引优化 ==========

-- 5. 用户+机器人+时间戳组合索引（成交明细核心查询）
-- 适用于: WHERE user_id=? AND robot_id=? ORDER BY ts
ALTER TABLE `hg_trading_trade_fill` 
ADD INDEX IF NOT EXISTS `idx_user_robot_ts` (`user_id`, `robot_id`, `ts`);

-- 6. 订单ID+时间戳组合索引（按订单查询成交）
-- 适用于: WHERE order_id=? ORDER BY ts
ALTER TABLE `hg_trading_trade_fill` 
ADD INDEX IF NOT EXISTS `idx_orderid_ts` (`order_id`, `ts`);

-- 7. 运行区间+时间戳组合索引（按区间统计）
-- 适用于: WHERE session_id=? AND ts BETWEEN ? AND ?
ALTER TABLE `hg_trading_trade_fill` 
ADD INDEX IF NOT EXISTS `idx_session_ts` (`session_id`, `ts`);

-- 8. API配置+交易对+时间戳组合索引（按API配置查询）
-- 适用于: WHERE api_config_id=? AND symbol=? ORDER BY ts
ALTER TABLE `hg_trading_trade_fill` 
ADD INDEX IF NOT EXISTS `idx_api_symbol_ts` (`api_config_id`, `symbol`, `ts`);

-- ========== 验证索引创建情况 ==========
-- 执行以下命令查看索引：
-- SHOW INDEX FROM hg_trading_order;
-- SHOW INDEX FROM hg_trading_trade_fill;

-- ========== 预期优化效果 ==========
-- 1. 交易明细页面查询速度提升 50-80%
-- 2. 减少全表扫描，降低CPU和IO消耗
-- 3. 支持高并发查询场景

