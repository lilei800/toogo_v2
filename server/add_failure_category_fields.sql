-- 为执行日志表增加失败分类字段
-- 用于优化失败原因展示

-- 1. 增加失败分类字段
ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_category VARCHAR(50) DEFAULT NULL 
COMMENT '失败分类：config/balance/position/exchange/strategy/system';

-- 2. 增加结构化失败原因字段
ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_reason TEXT DEFAULT NULL 
COMMENT '失败原因详情（结构化文本，包含解决建议）';

-- 3. 创建索引（提升查询性能）
CREATE INDEX IF NOT EXISTS idx_failure_category 
ON hg_trading_execution_log(failure_category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_robot_status_category 
ON hg_trading_execution_log(robot_id, status, failure_category, created_at DESC);

-- 4. 验证字段已添加
SELECT column_name, data_type, character_maximum_length, column_default, is_nullable
FROM information_schema.columns
WHERE table_name = 'hg_trading_execution_log'
  AND column_name IN ('failure_category', 'failure_reason')
ORDER BY ordinal_position;

-- 5. 查看表结构
\d hg_trading_execution_log;

-- 完成！可以继续部署代码优化。

