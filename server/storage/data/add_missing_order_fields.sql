-- ============================================================
-- 为 hg_trading_order 表添加缺失的字段
-- 开仓订单和平仓订单二合一方案优化
-- ============================================================

-- 开仓相关字段
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `order_type_detail` VARCHAR(50) DEFAULT NULL COMMENT '订单类型详情：market_open_long/market_open_short等' AFTER `order_type`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `exchange_side` VARCHAR(10) DEFAULT NULL COMMENT '交易所买卖方向：BUY/SELL' AFTER `order_type_detail`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `price` DECIMAL(15,4) DEFAULT NULL COMMENT '委托价格' AFTER `open_price`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `avg_price` DECIMAL(15,4) DEFAULT NULL COMMENT '平均成交价格' AFTER `price`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `filled_qty` DECIMAL(15,8) DEFAULT NULL COMMENT '已成交数量' AFTER `quantity`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `open_margin` DECIMAL(15,4) DEFAULT NULL COMMENT '开仓保证金(USDT)' AFTER `margin`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `open_fee` DECIMAL(15,8) DEFAULT NULL COMMENT '开仓手续费' AFTER `open_margin`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `open_fee_coin` VARCHAR(20) DEFAULT NULL COMMENT '开仓手续费币种' AFTER `open_fee`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `mark_price` DECIMAL(15,4) DEFAULT NULL COMMENT '标记价格（用于计算未实现盈亏）' AFTER `open_price`;

-- 平仓相关字段
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_order_id` VARCHAR(100) DEFAULT NULL COMMENT '平仓订单ID（交易所）' AFTER `close_price`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_client_order_id` VARCHAR(100) DEFAULT NULL COMMENT '平仓客户端订单ID' AFTER `close_order_id`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_fee` DECIMAL(15,8) DEFAULT NULL COMMENT '平仓手续费' AFTER `close_time`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_fee_coin` VARCHAR(20) DEFAULT NULL COMMENT '平仓手续费币种' AFTER `close_fee`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_market_state` VARCHAR(50) DEFAULT NULL COMMENT '平仓时的市场状态' AFTER `close_fee_coin`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_mark_price` DECIMAL(15,4) DEFAULT NULL COMMENT '平仓时的标记价格' AFTER `close_market_state`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_leverage` INT DEFAULT NULL COMMENT '平仓时的杠杆倍数' AFTER `close_mark_price`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_margin_mode` VARCHAR(20) DEFAULT NULL COMMENT '平仓时的保证金模式：isolated/cross' AFTER `close_leverage`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_quantity` DECIMAL(15,8) DEFAULT NULL COMMENT '平仓数量' AFTER `close_margin_mode`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_unrealized_profit` DECIMAL(15,4) DEFAULT NULL COMMENT '平仓时的未实现盈亏' AFTER `close_quantity`;
ALTER TABLE `hg_trading_order` ADD COLUMN IF NOT EXISTS `close_highest_profit` DECIMAL(15,4) DEFAULT NULL COMMENT '平仓时的最高盈利' AFTER `close_unrealized_profit`;

