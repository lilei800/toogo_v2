-- ========================================
-- 钱包交易明细页面索引优化
-- 在Navicat中直接复制粘贴执行即可
-- ========================================

-- 订单表索引（4个）
CREATE INDEX IF NOT EXISTS idx_user_status_closetime ON hg_trading_order (user_id, status, close_time);
CREATE INDEX IF NOT EXISTS idx_robot_status ON hg_trading_order (robot_id, status);
CREATE INDEX IF NOT EXISTS idx_exchange_symbol ON hg_trading_order (exchange, symbol);
CREATE INDEX IF NOT EXISTS idx_created_at ON hg_trading_order (created_at);

-- 成交流水表索引（4个）
CREATE INDEX IF NOT EXISTS idx_user_robot_ts ON hg_trading_trade_fill (user_id, robot_id, ts);
CREATE INDEX IF NOT EXISTS idx_orderid_ts ON hg_trading_trade_fill (order_id, ts);
CREATE INDEX IF NOT EXISTS idx_session_ts ON hg_trading_trade_fill (session_id, ts);
CREATE INDEX IF NOT EXISTS idx_api_symbol_ts ON hg_trading_trade_fill (api_config_id, symbol, ts);

-- 验证索引创建成功
SELECT '订单表索引：' as info, indexname FROM pg_indexes WHERE tablename = 'hg_trading_order' AND indexname LIKE 'idx_%'
UNION ALL
SELECT '成交流水表索引：' as info, indexname FROM pg_indexes WHERE tablename = 'hg_trading_trade_fill' AND indexname LIKE 'idx_%'
ORDER BY info, indexname;

