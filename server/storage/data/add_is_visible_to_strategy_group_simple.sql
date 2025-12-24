-- 为策略组表添加 is_visible 字段（用于控制策略组是否对用户可见）
-- 如果字段已存在则跳过（MySQL 语法）

-- 检查字段是否存在，如果不存在则添加
ALTER TABLE hg_trading_strategy_group 
ADD COLUMN IF NOT EXISTS is_visible TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否可见: 0=隐藏, 1=显示';

-- 为现有记录设置默认值（可见）
UPDATE hg_trading_strategy_group SET is_visible = 1 WHERE is_visible IS NULL OR is_visible = 0;

