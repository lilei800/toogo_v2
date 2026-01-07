-- ============================================================
-- PostgreSQL: 修复 hg_trading_signal_log 表的 is_processed 字段
-- 解决预警"已读"但未自动下单的问题
-- ============================================================

-- 步骤1: 检查字段是否存在
DO $$
DECLARE
    field_exists INTEGER;
BEGIN
    SELECT COUNT(*) INTO field_exists
    FROM information_schema.columns
    WHERE table_name = 'hg_trading_signal_log' 
      AND column_name = 'is_processed';
    
    IF field_exists > 0 THEN
        RAISE NOTICE '✓ is_processed 字段已存在';
    ELSE
        RAISE NOTICE '✗ is_processed 字段不存在，需要添加';
    END IF;
END $$;

-- 步骤2: 添加 is_processed 字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'hg_trading_signal_log' 
          AND column_name = 'is_processed'
    ) THEN
        -- 添加字段
        ALTER TABLE hg_trading_signal_log 
        ADD COLUMN is_processed SMALLINT NOT NULL DEFAULT 0;
        
        -- 添加注释
        COMMENT ON COLUMN hg_trading_signal_log.is_processed 
        IS '已读标识：0=未处理，1=已处理（用于防止重复下单）';
        
        RAISE NOTICE '✓ 成功添加 is_processed 字段';
    ELSE
        RAISE NOTICE '✓ is_processed 字段已存在，跳过添加';
    END IF;
END $$;

-- 步骤3: 确保字段类型正确
DO $$
DECLARE
    current_type TEXT;
BEGIN
    SELECT data_type INTO current_type
    FROM information_schema.columns
    WHERE table_name = 'hg_trading_signal_log' 
      AND column_name = 'is_processed';
    
    IF current_type IS NULL THEN
        RAISE NOTICE '✗ is_processed 字段不存在';
    ELSIF current_type NOT IN ('smallint', 'integer', 'bigint', 'boolean') THEN
        RAISE NOTICE '⚠ is_processed 字段类型为 %，建议修改为 smallint', current_type;
        -- 修改类型（如果需要）
        ALTER TABLE hg_trading_signal_log 
        ALTER COLUMN is_processed TYPE SMALLINT USING is_processed::SMALLINT;
        RAISE NOTICE '✓ 已将 is_processed 字段类型修改为 smallint';
    ELSE
        RAISE NOTICE '✓ is_processed 字段类型正确：%', current_type;
    END IF;
END $$;

-- 步骤4: 确保字段有默认值
DO $$
BEGIN
    ALTER TABLE hg_trading_signal_log 
    ALTER COLUMN is_processed SET DEFAULT 0;
    
    RAISE NOTICE '✓ 已设置 is_processed 默认值为 0';
END $$;

-- 步骤5: 确保字段非空
DO $$
BEGIN
    -- 先将 NULL 值更新为 0
    UPDATE hg_trading_signal_log 
    SET is_processed = 0 
    WHERE is_processed IS NULL;
    
    -- 然后设置非空约束
    ALTER TABLE hg_trading_signal_log 
    ALTER COLUMN is_processed SET NOT NULL;
    
    RAISE NOTICE '✓ 已设置 is_processed 非空约束';
END $$;

-- 步骤6: 添加其他可能缺失的字段
DO $$
BEGIN
    -- window_min_price
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'hg_trading_signal_log' 
          AND column_name = 'window_min_price'
    ) THEN
        ALTER TABLE hg_trading_signal_log 
        ADD COLUMN window_min_price NUMERIC(20,8) NOT NULL DEFAULT 0;
        RAISE NOTICE '✓ 已添加 window_min_price 字段';
    END IF;
    
    -- window_max_price
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'hg_trading_signal_log' 
          AND column_name = 'window_max_price'
    ) THEN
        ALTER TABLE hg_trading_signal_log 
        ADD COLUMN window_max_price NUMERIC(20,8) NOT NULL DEFAULT 0;
        RAISE NOTICE '✓ 已添加 window_max_price 字段';
    END IF;
    
    -- threshold
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'hg_trading_signal_log' 
          AND column_name = 'threshold'
    ) THEN
        ALTER TABLE hg_trading_signal_log 
        ADD COLUMN threshold NUMERIC(20,8) NOT NULL DEFAULT 0;
        RAISE NOTICE '✓ 已添加 threshold 字段';
    END IF;
    
    -- market_state
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'hg_trading_signal_log' 
          AND column_name = 'market_state'
    ) THEN
        ALTER TABLE hg_trading_signal_log 
        ADD COLUMN market_state VARCHAR(50) NOT NULL DEFAULT '';
        RAISE NOTICE '✓ 已添加 market_state 字段';
    END IF;
END $$;

-- 步骤7: 创建索引（提升查询性能）
DO $$
BEGIN
    -- 检查索引是否存在
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE tablename = 'hg_trading_signal_log' 
          AND indexname = 'idx_is_processed_robot'
    ) THEN
        CREATE INDEX idx_is_processed_robot 
        ON hg_trading_signal_log(robot_id, is_processed, created_at);
        RAISE NOTICE '✓ 已创建索引 idx_is_processed_robot';
    ELSE
        RAISE NOTICE '✓ 索引 idx_is_processed_robot 已存在';
    END IF;
    
    -- 为 executed 字段创建索引（用于重试机制）
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE tablename = 'hg_trading_signal_log' 
          AND indexname = 'idx_executed_processed'
    ) THEN
        CREATE INDEX idx_executed_processed 
        ON hg_trading_signal_log(executed, is_processed, created_at);
        RAISE NOTICE '✓ 已创建索引 idx_executed_processed';
    ELSE
        RAISE NOTICE '✓ 索引 idx_executed_processed 已存在';
    END IF;
END $$;

-- 步骤8: 创建 hg_trading_execution_log 表（如果不存在）
CREATE TABLE IF NOT EXISTS hg_trading_execution_log (
    id BIGSERIAL PRIMARY KEY,
    signal_log_id BIGINT NOT NULL DEFAULT 0,
    robot_id BIGINT NOT NULL DEFAULT 0,
    order_id BIGINT NOT NULL DEFAULT 0,
    event_type VARCHAR(50) NOT NULL DEFAULT '',
    event_data TEXT,
    status VARCHAR(20) NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    -- 可选字段：新版代码会写入；旧库缺字段会自动降级写入
    failure_category VARCHAR(50) NOT NULL DEFAULT '',
    failure_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN hg_trading_execution_log.signal_log_id IS '预警记录ID（可选）';
COMMENT ON COLUMN hg_trading_execution_log.robot_id IS '机器人ID';
COMMENT ON COLUMN hg_trading_execution_log.order_id IS '关联订单ID（可选）';
COMMENT ON COLUMN hg_trading_execution_log.event_type IS '事件类型';
COMMENT ON COLUMN hg_trading_execution_log.event_data IS '事件数据JSON';
COMMENT ON COLUMN hg_trading_execution_log.status IS '状态：pending/success/failed';
COMMENT ON COLUMN hg_trading_execution_log.message IS '消息（详细说明）';
COMMENT ON COLUMN hg_trading_execution_log.failure_category IS '失败分类（用于前端展示）';
COMMENT ON COLUMN hg_trading_execution_log.failure_reason IS '结构化失败原因（用于前端展示）';
COMMENT ON COLUMN hg_trading_execution_log.created_at IS '创建时间';

-- 为 execution_log 创建索引
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'hg_trading_execution_log'
          AND indexname = 'idx_trading_execution_log_signal_log_id'
    ) THEN
        CREATE INDEX idx_trading_execution_log_signal_log_id
        ON hg_trading_execution_log(signal_log_id);
        RAISE NOTICE '✓ 已创建索引 idx_trading_execution_log_signal_log_id';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'hg_trading_execution_log'
          AND indexname = 'idx_trading_execution_log_robot_time'
    ) THEN
        CREATE INDEX idx_trading_execution_log_robot_time
        ON hg_trading_execution_log(robot_id, created_at);
        RAISE NOTICE '✓ 已创建索引 idx_trading_execution_log_robot_time';
    END IF;
END $$;

-- 步骤9: 验证修复结果
DO $$
DECLARE
    field_info RECORD;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE '修复完成！验证结果：';
    RAISE NOTICE '========================================';
    
    -- 显示 is_processed 字段信息
    SELECT 
        column_name, 
        data_type, 
        is_nullable, 
        column_default
    INTO field_info
    FROM information_schema.columns
    WHERE table_name = 'hg_trading_signal_log' 
      AND column_name = 'is_processed';
    
    IF FOUND THEN
        RAISE NOTICE '✓ is_processed 字段信息：';
        RAISE NOTICE '  - 字段名: %', field_info.column_name;
        RAISE NOTICE '  - 数据类型: %', field_info.data_type;
        RAISE NOTICE '  - 允许NULL: %', field_info.is_nullable;
        RAISE NOTICE '  - 默认值: %', field_info.column_default;
    ELSE
        RAISE NOTICE '✗ is_processed 字段不存在！';
    END IF;
    
    RAISE NOTICE '========================================';
END $$;

-- 步骤10: 显示最近的预警记录（用于测试）
DO $$
DECLARE
    rec_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO rec_count
    FROM hg_trading_signal_log
    WHERE created_at >= NOW() - INTERVAL '1 hour';
    
    RAISE NOTICE '';
    RAISE NOTICE '最近1小时的预警记录数: %', rec_count;
    RAISE NOTICE '';
    RAISE NOTICE '执行以下SQL查看详情：';
    RAISE NOTICE '  SELECT id, robot_id, signal_type, is_processed, executed, created_at';
    RAISE NOTICE '  FROM hg_trading_signal_log';
    RAISE NOTICE '  WHERE created_at >= NOW() - INTERVAL ''1 hour''';
    RAISE NOTICE '  ORDER BY id DESC LIMIT 10;';
    RAISE NOTICE '';
END $$;

-- ============================================================
-- 修复完成！
-- ============================================================

-- 可选：查看最近的预警记录
SELECT 
    id,
    robot_id,
    signal_type,
    current_price,
    is_processed,
    executed,
    LEFT(reason, 50) as reason_preview,
    created_at
FROM hg_trading_signal_log
ORDER BY id DESC
LIMIT 10;

