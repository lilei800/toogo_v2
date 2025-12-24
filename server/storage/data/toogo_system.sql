-- =============================================
-- Toogo.Ai 土狗 - 全自动虚拟货币量化交易系统
-- 完整数据库设计 v1.0
-- 创建时间: 2024-11-28
-- =============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";

-- =============================================
-- 一、用户体系相关表
-- =============================================

-- -------------------------------------------------
-- 1. 用户扩展表 (扩展admin_member)
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_user`;
CREATE TABLE `hg_toogo_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `member_id` bigint(20) NOT NULL COMMENT '关联admin_member.id',
  
  -- 身份信息
  `vip_level` tinyint(2) NOT NULL DEFAULT '1' COMMENT '身份等级: V1-V10',
  `is_agent` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否代理商: 0=否, 1=是',
  `agent_level` tinyint(2) DEFAULT '0' COMMENT '代理商等级',
  
  -- 邀请关系
  `invite_code` varchar(20) NOT NULL COMMENT '邀请码',
  `invite_code_expire` datetime DEFAULT NULL COMMENT '邀请码过期时间',
  `inviter_id` bigint(20) DEFAULT '0' COMMENT '邀请人ID',
  `invite_count` int(11) NOT NULL DEFAULT '0' COMMENT '直接邀请人数',
  `team_count` int(11) NOT NULL DEFAULT '0' COMMENT '团队总人数',
  
  -- 消费统计
  `total_consume_power` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '总消耗算力',
  `team_consume_power` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '团队消耗算力',
  
  -- 订阅信息
  `current_plan_id` bigint(20) DEFAULT NULL COMMENT '当前订阅套餐ID',
  `plan_expire_time` datetime DEFAULT NULL COMMENT '套餐到期时间',
  `robot_limit` int(11) NOT NULL DEFAULT '0' COMMENT '机器人数量限制',
  `active_robot_count` int(11) NOT NULL DEFAULT '0' COMMENT '运行中机器人数量',
  
  -- 算力折扣(根据VIP等级)
  `power_discount` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '算力消耗折扣(%)',
  
  `status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '状态: 1=正常, 2=禁用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_member_id` (`member_id`),
  UNIQUE KEY `uk_invite_code` (`invite_code`),
  KEY `idx_inviter_id` (`inviter_id`),
  KEY `idx_vip_level` (`vip_level`),
  KEY `idx_is_agent` (`is_agent`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Toogo用户扩展表';

-- -------------------------------------------------
-- 2. VIP等级配置表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_vip_level`;
CREATE TABLE `hg_toogo_vip_level` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `level` tinyint(2) NOT NULL COMMENT '等级: 1-10',
  `level_name` varchar(50) NOT NULL COMMENT '等级名称',
  
  -- 升级条件
  `require_invite_count` int(11) NOT NULL DEFAULT '0' COMMENT '需要邀请人数',
  `require_consume_power` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '需要消耗算力',
  `require_team_consume` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '需要团队消耗算力',
  
  -- 权益
  `power_discount` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '算力折扣(5-30%)',
  `invite_reward_power` decimal(10,4) NOT NULL DEFAULT '30.0000' COMMENT '邀请奖励算力',
  
  `description` varchar(500) DEFAULT NULL COMMENT '等级描述',
  `icon` varchar(255) DEFAULT NULL COMMENT '等级图标',
  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '状态: 1=启用, 2=禁用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='VIP等级配置表';

-- 插入默认VIP等级
INSERT INTO `hg_toogo_vip_level` (`level`, `level_name`, `require_invite_count`, `require_consume_power`, `require_team_consume`, `power_discount`, `invite_reward_power`, `description`, `sort`) VALUES
(1, 'V1-入门会员', 0, 0, 0, 0, 30, '新用户默认等级', 1),
(2, 'V2-初级会员', 5, 100, 500, 5, 35, '邀请5人或消耗100算力', 2),
(3, 'V3-中级会员', 15, 500, 2000, 8, 40, '邀请15人或消耗500算力', 3),
(4, 'V4-高级会员', 30, 1500, 5000, 10, 45, '邀请30人或消耗1500算力', 4),
(5, 'V5-白银会员', 50, 3000, 10000, 12, 50, '邀请50人或消耗3000算力', 5),
(6, 'V6-黄金会员', 100, 6000, 25000, 15, 55, '邀请100人或消耗6000算力', 6),
(7, 'V7-白金会员', 200, 12000, 50000, 18, 60, '邀请200人或消耗12000算力', 7),
(8, 'V8-钻石会员', 400, 25000, 100000, 22, 70, '邀请400人或消耗25000算力', 8),
(9, 'V9-皇冠会员', 800, 50000, 250000, 26, 80, '邀请800人或消耗50000算力', 9),
(10, 'V10-至尊会员', 1500, 100000, 500000, 30, 100, '顶级会员，最高权益', 10);

-- =============================================
-- 二、订阅套餐相关表
-- =============================================

-- -------------------------------------------------
-- 3. 订阅套餐表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_plan`;
CREATE TABLE `hg_toogo_plan` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `plan_name` varchar(50) NOT NULL COMMENT '套餐名称',
  `plan_code` varchar(20) NOT NULL COMMENT '套餐代码: FREE/A/B/C/D',
  
  -- 机器人配置
  `robot_limit` int(11) NOT NULL COMMENT '支持机器人数量',
  
  -- 价格配置 (USDT)
  `price_daily` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '日价格',
  `price_monthly` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '月价格',
  `price_quarterly` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '季价格',
  `price_half_year` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '半年价格',
  `price_yearly` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '年价格',
  
  -- 赠送算力
  `gift_power_monthly` decimal(10,2) DEFAULT '0.00' COMMENT '月订阅赠送算力',
  `gift_power_quarterly` decimal(10,2) DEFAULT '0.00' COMMENT '季订阅赠送算力',
  `gift_power_half_year` decimal(10,2) DEFAULT '0.00' COMMENT '半年订阅赠送算力',
  `gift_power_yearly` decimal(10,2) DEFAULT '0.00' COMMENT '年订阅赠送算力',
  
  `description` text COMMENT '套餐描述',
  `features` text COMMENT '套餐特性(JSON)',
  `is_default` tinyint(1) DEFAULT '0' COMMENT '是否默认套餐(免费)',
  `sort` int(11) DEFAULT '0' COMMENT '排序',
  `status` tinyint(2) DEFAULT '1' COMMENT '状态: 1=上架, 2=下架',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plan_code` (`plan_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订阅套餐表';

-- 插入默认套餐
INSERT INTO `hg_toogo_plan` (`plan_name`, `plan_code`, `robot_limit`, `price_daily`, `price_monthly`, `price_quarterly`, `price_half_year`, `price_yearly`, `gift_power_monthly`, `gift_power_quarterly`, `gift_power_half_year`, `gift_power_yearly`, `is_default`, `sort`, `description`) VALUES
('免费套餐', 'FREE', 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, '新用户免费体验，支持1个机器人'),
('A套餐', 'A', 1, 1, 27, 72, 126, 180, 10, 30, 60, 100, 0, 1, '入门套餐，支持1个云机器人'),
('B套餐', 'B', 3, 3, 60, 160, 240, 360, 30, 80, 150, 250, 0, 2, '进阶套餐，支持3个云机器人'),
('C套餐', 'C', 5, 5, 90, 240, 370, 520, 50, 130, 250, 400, 0, 3, '专业套餐，支持5个云机器人'),
('D套餐', 'D', 10, 10, 150, 400, 580, 880, 100, 250, 450, 700, 0, 4, '旗舰套餐，支持10个云机器人');

-- -------------------------------------------------
-- 4. 用户订阅记录表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_subscription`;
CREATE TABLE `hg_toogo_subscription` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID(member_id)',
  `plan_id` bigint(20) NOT NULL COMMENT '套餐ID',
  `plan_code` varchar(20) NOT NULL COMMENT '套餐代码',
  
  -- 订阅信息
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  `period_type` varchar(20) NOT NULL COMMENT '订阅周期: daily/monthly/quarterly/half_year/yearly',
  `amount` decimal(10,2) NOT NULL COMMENT '订阅金额(USDT)',
  `gift_power` decimal(10,2) DEFAULT '0.00' COMMENT '赠送算力',
  
  -- 时间
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `expire_time` datetime NOT NULL COMMENT '到期时间',
  `days` int(11) NOT NULL COMMENT '订阅天数',
  
  -- 状态
  `status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '状态: 1=待支付, 2=生效中, 3=已过期, 4=已取消',
  `paid_at` datetime DEFAULT NULL COMMENT '支付时间',
  `pay_type` varchar(20) DEFAULT NULL COMMENT '支付方式: balance/crypto',
  
  -- 佣金信息
  `inviter_id` bigint(20) DEFAULT '0' COMMENT '邀请人ID',
  `commission_settled` tinyint(1) DEFAULT '0' COMMENT '佣金是否已结算',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户订阅记录表';

-- =============================================
-- 三、财务账户相关表
-- =============================================

-- -------------------------------------------------
-- 5. 用户钱包账户表(多账户)
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_wallet`;
CREATE TABLE `hg_toogo_wallet` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID(member_id)',
  
  -- 余额账户(可充值、可提现)
  `balance` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '余额(USDT)',
  `frozen_balance` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '冻结余额',
  
  -- 算力账户(可消耗、可获得佣金)
  `power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '算力余额',
  `frozen_power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '冻结算力',
  
  -- 赠送算力账户(只能消耗、无佣金)
  `gift_power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '赠送算力余额',
  
  -- 佣金账户(可提现、可转算力)
  `commission` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '佣金余额(USDT)',
  `frozen_commission` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '冻结佣金',
  
  -- 累计统计
  `total_deposit` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '累计充值',
  `total_withdraw` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '累计提现',
  `total_power_consume` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '累计消耗算力',
  `total_commission` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '累计获得佣金',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户钱包账户表';

-- -------------------------------------------------
-- 6. 账户流水记录表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_wallet_log`;
CREATE TABLE `hg_toogo_wallet_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID(member_id)',
  
  -- 账户类型: balance=余额, power=算力, gift_power=赠送算力, commission=佣金
  `account_type` varchar(20) NOT NULL COMMENT '账户类型',
  
  -- 变动类型
  `change_type` varchar(30) NOT NULL COMMENT '变动类型',
  `change_amount` decimal(20,8) NOT NULL COMMENT '变动金额',
  `before_amount` decimal(20,8) NOT NULL COMMENT '变动前余额',
  `after_amount` decimal(20,8) NOT NULL COMMENT '变动后余额',
  
  -- 关联信息
  `related_id` bigint(20) DEFAULT NULL COMMENT '关联ID',
  `related_type` varchar(30) DEFAULT NULL COMMENT '关联类型',
  `order_sn` varchar(64) DEFAULT NULL COMMENT '关联订单号',
  
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_account_type` (`account_type`),
  KEY `idx_change_type` (`change_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账户流水记录表';

-- -------------------------------------------------
-- 7. 充值订单表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_deposit`;
CREATE TABLE `hg_toogo_deposit` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID(member_id)',
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  
  -- 充值信息
  `amount` decimal(20,8) NOT NULL COMMENT '充值金额(USDT)',
  `real_amount` decimal(20,8) DEFAULT NULL COMMENT '实际到账金额',
  `network` varchar(32) NOT NULL COMMENT '网络: TRC20/ERC20/BEP20',
  `to_address` varchar(128) NOT NULL COMMENT '充值地址',
  
  -- 第三方信息
  `payment_channel` varchar(50) DEFAULT NULL COMMENT '支付渠道',
  `payment_id` varchar(128) DEFAULT NULL COMMENT '第三方支付ID',
  `tx_hash` varchar(128) DEFAULT NULL COMMENT '交易哈希',
  
  -- 状态
  `status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '状态: 1=待支付, 2=已完成, 3=已超时, 4=已取消',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `paid_at` datetime DEFAULT NULL COMMENT '支付时间',
  
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='充值订单表';

-- -------------------------------------------------
-- 8. 提现订单表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_withdraw`;
CREATE TABLE `hg_toogo_withdraw` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID(member_id)',
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  
  -- 提现信息
  `account_type` varchar(20) NOT NULL COMMENT '账户类型: balance/commission',
  `amount` decimal(20,8) NOT NULL COMMENT '提现金额(USDT)',
  `fee` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '手续费',
  `real_amount` decimal(20,8) NOT NULL COMMENT '实际到账金额',
  `to_address` varchar(128) NOT NULL COMMENT '提现地址',
  `network` varchar(32) NOT NULL COMMENT '网络: TRC20/ERC20/BEP20',
  
  -- 交易信息
  `tx_hash` varchar(128) DEFAULT NULL COMMENT '交易哈希',
  
  -- 审核信息
  `status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '状态: 1=待审核, 2=审核通过, 3=审核拒绝, 4=已完成, 5=已取消',
  `audit_remark` varchar(500) DEFAULT NULL COMMENT '审核备注',
  `audited_by` bigint(20) DEFAULT NULL COMMENT '审核人ID',
  `audited_at` datetime DEFAULT NULL COMMENT '审核时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提现订单表';

-- -------------------------------------------------
-- 9. 账户互转记录表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_transfer`;
CREATE TABLE `hg_toogo_transfer` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID(member_id)',
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  
  -- 转账信息
  `from_account` varchar(20) NOT NULL COMMENT '转出账户: balance/commission',
  `to_account` varchar(20) NOT NULL COMMENT '转入账户: power',
  `amount` decimal(20,8) NOT NULL COMMENT '转账金额(USDT)',
  `power_amount` decimal(20,8) NOT NULL COMMENT '获得算力',
  `rate` decimal(10,4) NOT NULL DEFAULT '1.0000' COMMENT '兑换比率',
  
  `status` tinyint(2) NOT NULL DEFAULT '2' COMMENT '状态: 1=处理中, 2=已完成, 3=失败',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账户互转记录表';

-- =============================================
-- 四、代理商体系相关表
-- =============================================

-- -------------------------------------------------
-- 10. 代理商等级配置表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_agent_level`;
CREATE TABLE `hg_toogo_agent_level` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `level` tinyint(2) NOT NULL COMMENT '等级: 1-5',
  `level_name` varchar(50) NOT NULL COMMENT '等级名称',
  
  -- 升级条件
  `require_team_count` int(11) NOT NULL DEFAULT '0' COMMENT '需要团队人数',
  `require_team_subscribe` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '需要团队订阅额(USDT)',
  
  -- 佣金比例 (三级分销)
  `subscribe_rate_1` decimal(5,4) NOT NULL DEFAULT '0.0000' COMMENT '订阅佣金比例(一级)',
  `subscribe_rate_2` decimal(5,4) NOT NULL DEFAULT '0.0000' COMMENT '订阅佣金比例(二级)',
  `subscribe_rate_3` decimal(5,4) NOT NULL DEFAULT '0.0000' COMMENT '订阅佣金比例(三级)',
  `power_rate_1` decimal(5,4) NOT NULL DEFAULT '0.0000' COMMENT '算力消耗佣金比例(一级)',
  `power_rate_2` decimal(5,4) NOT NULL DEFAULT '0.0000' COMMENT '算力消耗佣金比例(二级)',
  `power_rate_3` decimal(5,4) NOT NULL DEFAULT '0.0000' COMMENT '算力消耗佣金比例(三级)',
  
  `description` varchar(500) DEFAULT NULL COMMENT '等级描述',
  `sort` int(11) DEFAULT '0' COMMENT '排序',
  `status` tinyint(2) DEFAULT '1' COMMENT '状态: 1=启用, 2=禁用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='代理商等级配置表';

-- 插入默认代理商等级
INSERT INTO `hg_toogo_agent_level` (`level`, `level_name`, `require_team_count`, `require_team_subscribe`, `subscribe_rate_1`, `subscribe_rate_2`, `subscribe_rate_3`, `power_rate_1`, `power_rate_2`, `power_rate_3`, `description`, `sort`) VALUES
(1, '初级代理', 10, 500, 0.10, 0.05, 0.02, 0.05, 0.02, 0.01, '入门代理，团队10人', 1),
(2, '中级代理', 50, 3000, 0.12, 0.06, 0.03, 0.06, 0.03, 0.015, '中级代理，团队50人', 2),
(3, '高级代理', 200, 15000, 0.15, 0.08, 0.04, 0.08, 0.04, 0.02, '高级代理，团队200人', 3),
(4, '超级代理', 500, 50000, 0.18, 0.10, 0.05, 0.10, 0.05, 0.025, '超级代理，团队500人', 4),
(5, '合伙人', 1000, 150000, 0.20, 0.12, 0.06, 0.12, 0.06, 0.03, '战略合伙人，最高权益', 5);

-- -------------------------------------------------
-- 11. 佣金记录表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_commission_log`;
CREATE TABLE `hg_toogo_commission_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) NOT NULL COMMENT '获得佣金用户ID',
  `from_user_id` bigint(20) NOT NULL COMMENT '来源用户ID',
  
  -- 佣金类型: invite_reward=邀请奖励, subscribe=订阅佣金, power_consume=算力消耗佣金
  `commission_type` varchar(30) NOT NULL COMMENT '佣金类型',
  `level` tinyint(2) NOT NULL DEFAULT '1' COMMENT '层级: 1=一级, 2=二级, 3=三级',
  
  -- 佣金信息
  `base_amount` decimal(20,8) NOT NULL COMMENT '基础金额(订阅额/算力消耗)',
  `commission_rate` decimal(5,4) NOT NULL COMMENT '佣金比例',
  `commission_amount` decimal(20,8) NOT NULL COMMENT '佣金金额',
  
  -- 结算信息: power=算力, usdt=USDT
  `settle_type` varchar(20) NOT NULL COMMENT '结算类型',
  `status` tinyint(2) NOT NULL DEFAULT '2' COMMENT '状态: 1=待结算, 2=已结算',
  
  -- 关联信息
  `related_id` bigint(20) DEFAULT NULL COMMENT '关联ID',
  `related_type` varchar(30) DEFAULT NULL COMMENT '关联类型',
  `order_sn` varchar(64) DEFAULT NULL COMMENT '关联订单号',
  
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_from_user_id` (`from_user_id`),
  KEY `idx_commission_type` (`commission_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='佣金记录表';

-- =============================================
-- 五、策略模板表(合并波动率配置)
-- =============================================

-- -------------------------------------------------
-- 12. 策略模板表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_strategy_template`;
CREATE TABLE `hg_toogo_strategy_template` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  
  -- 策略标识
  `strategy_key` varchar(50) NOT NULL COMMENT '策略KEY',
  `strategy_name` varchar(100) NOT NULL COMMENT '策略名称',
  `risk_preference` varchar(20) NOT NULL COMMENT '风险偏好: conservative/balanced/aggressive',
  `market_state` varchar(20) NOT NULL COMMENT '市场状态: trend/volatile/high_volatility/low_volatility',
  
  -- 下单参数
  `time_window` int(11) NOT NULL DEFAULT '300' COMMENT '时间窗口(秒)',
  `volatility_points` decimal(10,2) NOT NULL DEFAULT '100.00' COMMENT '波动点数(USDT)',
  `leverage_min` int(11) NOT NULL DEFAULT '5' COMMENT '杠杆倍数最小值',
  `leverage_max` int(11) NOT NULL DEFAULT '15' COMMENT '杠杆倍数最大值',
  `margin_percent_min` decimal(5,2) NOT NULL DEFAULT '20.00' COMMENT '保证金比例最小值(%)',
  `margin_percent_max` decimal(5,2) NOT NULL DEFAULT '45.00' COMMENT '保证金比例最大值(%)',
  
  -- 平仓参数
  `stop_loss_percent` decimal(5,2) NOT NULL DEFAULT '10.00' COMMENT '止损百分比(%)',
  `profit_retreat_percent` decimal(5,2) NOT NULL DEFAULT '18.00' COMMENT '止盈回撤百分比(%)',
  `start_retreat_percent` decimal(5,2) NOT NULL DEFAULT '8.00' COMMENT '启动回撤百分比(%)',
  
  -- 反方向下单策略配置
  `reverse_loss_retreat` decimal(5,2) NOT NULL DEFAULT '50.00' COMMENT '反向-亏损订单回撤百分比',
  `reverse_profit_retreat` decimal(5,2) NOT NULL DEFAULT '100.00' COMMENT '反向-盈利订单回撤百分比',
  
  -- 波动率判断阈值(内嵌JSON)
  `volatility_config` JSON COMMENT '多周期波动率配置',
  
  `description` varchar(500) DEFAULT NULL COMMENT '策略描述',
  `is_official` tinyint(1) DEFAULT '0' COMMENT '是否官方推荐: 0=否, 1=是',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否激活: 0=否, 1=是',
  `sort` int(11) DEFAULT '100' COMMENT '排序',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_strategy_key` (`strategy_key`),
  KEY `idx_risk_market` (`risk_preference`, `market_state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='策略模板表';

-- 插入默认策略模板数据 (12种组合)
INSERT INTO `hg_toogo_strategy_template` 
(`strategy_key`, `strategy_name`, `risk_preference`, `market_state`, 
 `time_window`, `volatility_points`, `leverage_min`, `leverage_max`, 
 `margin_percent_min`, `margin_percent_max`, `stop_loss_percent`, 
 `profit_retreat_percent`, `start_retreat_percent`,
 `reverse_loss_retreat`, `reverse_profit_retreat`,
 `volatility_config`, `description`, `is_official`, `sort`) 
VALUES
-- 保守型策略 (4种)
('conservative_trend', '保守型-趋势市场', 'conservative', 'trend',
 180, 100.00, 6, 8, 20.00, 25.00, 10.00, 20.00, 8.00,
 50.00, 100.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '适合新手，稳健盈利，风险较低', 1, 10),

('conservative_volatile', '保守型-震荡市场', 'conservative', 'volatile',
 300, 120.00, 6, 8, 20.00, 25.00, 10.00, 22.00, 8.00,
 0.00, 0.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '震荡市场保守策略，降低交易频率', 1, 11),

('conservative_high_vol', '保守型-高波动', 'conservative', 'high_volatility',
 360, 150.00, 5, 7, 18.00, 22.00, 12.00, 25.00, 10.00,
 100.00, 100.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '高波动市场保守策略，风控优先', 1, 12),

('conservative_low_vol', '保守型-低波动', 'conservative', 'low_volatility',
 120, 80.00, 7, 9, 22.00, 27.00, 8.00, 18.00, 6.00,
 0.00, 0.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '低波动市场保守策略，小额多次', 1, 13),

-- 平衡型策略 (4种)
('balanced_trend', '平衡型-趋势市场', 'balanced', 'trend',
 300, 75.00, 10, 10, 30.00, 35.00, 8.00, 18.00, 8.00,
 50.00, 100.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '平衡风险收益，适合大多数用户', 1, 20),

('balanced_volatile', '平衡型-震荡市场', 'balanced', 'volatile',
 240, 90.00, 10, 10, 30.00, 35.00, 8.00, 20.00, 8.00,
 0.00, 0.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '震荡市场平衡策略，灵活应对', 1, 21),

('balanced_high_vol', '平衡型-高波动', 'balanced', 'high_volatility',
 300, 120.00, 8, 10, 28.00, 33.00, 10.00, 22.00, 10.00,
 100.00, 100.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '高波动市场平衡策略', 1, 22),

('balanced_low_vol', '平衡型-低波动', 'balanced', 'low_volatility',
 180, 60.00, 10, 12, 32.00, 37.00, 7.00, 15.00, 6.00,
 0.00, 0.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '低波动市场平衡策略', 1, 23),

-- 激进型策略 (4种)
('aggressive_trend', '激进型-趋势市场', 'aggressive', 'trend',
 120, 50.00, 12, 15, 40.00, 45.00, 8.00, 15.00, 5.00,
 50.00, 100.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '高风险高收益，适合专业用户', 1, 30),

('aggressive_volatile', '激进型-震荡市场', 'aggressive', 'volatile',
 180, 70.00, 12, 15, 40.00, 45.00, 8.00, 18.00, 5.00,
 0.00, 0.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '震荡市场激进策略，快进快出', 1, 31),

('aggressive_high_vol', '激进型-高波动', 'aggressive', 'high_volatility',
 240, 100.00, 10, 13, 35.00, 40.00, 10.00, 20.00, 8.00,
 100.00, 100.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '高波动市场激进策略，控制风险', 1, 32),

('aggressive_low_vol', '激进型-低波动', 'aggressive', 'low_volatility',
 90, 40.00, 13, 16, 42.00, 48.00, 6.00, 12.00, 4.00,
 0.00, 0.00,
 '{"timeframes":[{"frame":"1m","weight":0.10,"trend_threshold":0.15,"volatile_threshold":0.08},{"frame":"5m","weight":0.15,"trend_threshold":0.20,"volatile_threshold":0.10},{"frame":"15m","weight":0.25,"trend_threshold":0.30,"volatile_threshold":0.15},{"frame":"30m","weight":0.25,"trend_threshold":0.40,"volatile_threshold":0.20},{"frame":"1h","weight":0.25,"trend_threshold":0.50,"volatile_threshold":0.25}]}',
 '低波动市场激进策略，最大化收益', 1, 33);

-- =============================================
-- 六、算力消耗相关表
-- =============================================

-- -------------------------------------------------
-- 13. 算力消耗记录表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_power_consume`;
CREATE TABLE `hg_toogo_power_consume` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID(member_id)',
  `robot_id` bigint(20) NOT NULL COMMENT '机器人ID',
  `order_id` bigint(20) NOT NULL COMMENT '交易订单ID',
  
  -- 消耗信息
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  `profit_amount` decimal(20,8) NOT NULL COMMENT '盈利金额(USDT)',
  `consume_rate` decimal(5,4) NOT NULL COMMENT '消耗比例',
  `consume_power` decimal(20,8) NOT NULL COMMENT '消耗算力',
  
  -- 消耗来源
  `from_power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '从算力账户扣除',
  
  -- 折扣信息
  `vip_level` tinyint(2) NOT NULL DEFAULT '1' COMMENT '用户VIP等级',
  `discount_rate` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '折扣比例(%)',
  `original_power` decimal(20,8) NOT NULL COMMENT '原始消耗算力(未折扣)',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_robot_id` (`robot_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='算力消耗记录表';

-- =============================================
-- 七、系统配置表
-- =============================================

-- -------------------------------------------------
-- 14. Toogo系统配置表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_config`;
CREATE TABLE `hg_toogo_config` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group` varchar(50) NOT NULL COMMENT '配置分组',
  `key` varchar(100) NOT NULL COMMENT '配置KEY',
  `value` text NOT NULL COMMENT '配置值',
  `type` varchar(20) NOT NULL DEFAULT 'string' COMMENT '值类型: string/number/boolean/json',
  `name` varchar(100) NOT NULL COMMENT '配置名称',
  `description` varchar(500) DEFAULT NULL COMMENT '配置描述',
  `sort` int(11) DEFAULT '0' COMMENT '排序',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_key` (`group`, `key`),
  KEY `idx_group` (`group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Toogo系统配置表';

-- 插入默认配置
INSERT INTO `hg_toogo_config` (`group`, `key`, `value`, `type`, `name`, `description`, `sort`) VALUES
-- 基础配置
('basic', 'site_name', 'Toogo.Ai', 'string', '站点名称', '系统名称', 1),
('basic', 'site_logo', '/images/toogo-logo.png', 'string', '站点Logo', 'Logo图片路径', 2),
('basic', 'site_description', '全自动虚拟货币量化交易系统', 'string', '站点描述', '', 3),

-- 算力配置
('power', 'consume_rate', '0.10', 'number', '算力消耗比例', '云机器人消耗比例(默认10%)', 1),
('power', 'min_consume', '0.01', 'number', '最小消耗算力', '单笔最小消耗', 2),
('power', 'exchange_rate', '1.00', 'number', 'USDT兑算力比率', '1 USDT = ? 算力', 3),

-- 佣金配置(普通用户)
('commission', 'subscribe_rate_1', '0.10', 'number', '订阅一级佣金', '普通用户订阅一级佣金比例', 1),
('commission', 'subscribe_rate_2', '0.05', 'number', '订阅二级佣金', '普通用户订阅二级佣金比例', 2),
('commission', 'power_rate_1', '0.05', 'number', '算力一级佣金', '普通用户算力消耗一级佣金比例', 3),
('commission', 'power_rate_2', '0.02', 'number', '算力二级佣金', '普通用户算力消耗二级佣金比例', 4),

-- 提现配置
('withdraw', 'min_amount', '10', 'number', '最低提现金额', 'USDT', 1),
('withdraw', 'fee_rate', '0.02', 'number', '提现手续费比例', '', 2),
('withdraw', 'daily_limit', '10000', 'number', '每日提现限额', 'USDT', 3),

-- 邀请配置
('invite', 'code_expire_hours', '24', 'number', '邀请码有效期', '小时', 1),
('invite', 'register_reward', '30', 'number', '注册奖励算力', '邀请人和被邀请人各获得', 2),
('invite', 'need_invite_code', '1', 'boolean', '注册需要邀请码', '1=是,0=否', 3),

-- 机器人默认配置
('robot', 'default_leverage', '10', 'number', '默认杠杆倍数', '', 1),
('robot', 'default_margin_percent', '30', 'number', '默认保证金比例', '%', 2),
('robot', 'default_stop_loss', '10', 'number', '默认止损比例', '%', 3),
('robot', 'default_profit_retreat', '18', 'number', '默认止盈回撤比例', '%', 4),
('robot', 'default_start_retreat', '8', 'number', '默认启动回撤比例', '%', 5);

-- =============================================
-- 八、扩展现有交易机器人表
-- =============================================

-- 为现有机器人表添加Toogo相关字段
-- 注意：如果字段已存在会报错，可以忽略

-- 使用存储过程安全添加列
DELIMITER //
DROP PROCEDURE IF EXISTS add_column_if_not_exists//
CREATE PROCEDURE add_column_if_not_exists()
BEGIN
    -- plan_id
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hg_trading_robot' AND column_name = 'plan_id') THEN
        ALTER TABLE `hg_trading_robot` ADD COLUMN `plan_id` bigint(20) DEFAULT NULL COMMENT '套餐ID' AFTER `user_id`;
    END IF;
    
    -- consumed_power
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hg_trading_robot' AND column_name = 'consumed_power') THEN
        ALTER TABLE `hg_trading_robot` ADD COLUMN `consumed_power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '已消耗算力';
    END IF;
    
    -- estimated_power
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hg_trading_robot' AND column_name = 'estimated_power') THEN
        ALTER TABLE `hg_trading_robot` ADD COLUMN `estimated_power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '预计消耗算力';
    END IF;
    
    -- schedule_start
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hg_trading_robot' AND column_name = 'schedule_start') THEN
        ALTER TABLE `hg_trading_robot` ADD COLUMN `schedule_start` datetime DEFAULT NULL COMMENT '定时开启时间';
    END IF;
    
    -- schedule_stop
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hg_trading_robot' AND column_name = 'schedule_stop') THEN
        ALTER TABLE `hg_trading_robot` ADD COLUMN `schedule_stop` datetime DEFAULT NULL COMMENT '定时关闭时间';
    END IF;
    
    -- auto_analyze_market
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hg_trading_robot' AND column_name = 'auto_analyze_market') THEN
        ALTER TABLE `hg_trading_robot` ADD COLUMN `auto_analyze_market` tinyint(1) DEFAULT '0' COMMENT '全自动分析行情: 0=手动, 1=自动';
    END IF;
    
    -- auto_signal_enabled
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hg_trading_robot' AND column_name = 'auto_signal_enabled') THEN
        ALTER TABLE `hg_trading_robot` ADD COLUMN `auto_signal_enabled` tinyint(1) DEFAULT '1' COMMENT '全自动方向预警信号: 0=关, 1=开';
    END IF;
    
    -- trade_type
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hg_trading_robot' AND column_name = 'trade_type') THEN
        ALTER TABLE `hg_trading_robot` ADD COLUMN `trade_type` varchar(20) DEFAULT 'perpetual' COMMENT '交易类: perpetual=永续合约';
    END IF;
END//
DELIMITER ;

-- 执行存储过程
CALL add_column_if_not_exists();

-- 删除存储过程
DROP PROCEDURE IF EXISTS add_column_if_not_exists;

-- =============================================
-- 九、AI学习记录表(可选)
-- =============================================

-- -------------------------------------------------
-- 15. AI学习记录表
-- -------------------------------------------------
DROP TABLE IF EXISTS `hg_toogo_ai_learning`;
CREATE TABLE `hg_toogo_ai_learning` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `symbol` varchar(20) NOT NULL COMMENT '交易对',
  
  -- 学习数据
  `time_frame` varchar(10) NOT NULL COMMENT '时间周期: 1m/5m/15m/30m/1h',
  `market_state` varchar(20) NOT NULL COMMENT '市场状态',
  `risk_preference` varchar(20) NOT NULL COMMENT '风险偏好',
  
  -- 权重分析
  `price_weight` decimal(5,4) NOT NULL DEFAULT '0.3000' COMMENT '价格权重',
  `volume_weight` decimal(5,4) NOT NULL DEFAULT '0.2000' COMMENT '成交量权重',
  `trend_weight` decimal(5,4) NOT NULL DEFAULT '0.2500' COMMENT '趋势权重',
  `volatility_weight` decimal(5,4) NOT NULL DEFAULT '0.2500' COMMENT '波动率权重',
  
  -- 准确率统计
  `total_signals` int(11) NOT NULL DEFAULT '0' COMMENT '总信号数',
  `correct_signals` int(11) NOT NULL DEFAULT '0' COMMENT '正确信号数',
  `accuracy_rate` decimal(5,4) NOT NULL DEFAULT '0.0000' COMMENT '准确率',
  
  -- 收益统计
  `total_profit` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '总收益',
  `avg_profit` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '平均收益',
  
  `last_update` datetime DEFAULT NULL COMMENT '最后更新时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_symbol_frame_state_risk` (`symbol`, `time_frame`, `market_state`, `risk_preference`),
  KEY `idx_symbol` (`symbol`),
  KEY `idx_accuracy` (`accuracy_rate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI学习记录表';

-- =============================================
-- 完成
-- =============================================

SET FOREIGN_KEY_CHECKS = 1;

SELECT '✅ Toogo.Ai 数据库表创建完成！' as result;

