-- =============================================================================
-- Support Chat Tables (PostgreSQL)
-- 仅创建客服聊天体系所需的 4 张表 + 索引
-- 使用方式：Navicat 选中 hotgo 库执行本文件
-- =============================================================================

-- 会话表
CREATE TABLE IF NOT EXISTS hg_support_session (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,
    subject VARCHAR(255) DEFAULT '',
    last_msg TEXT DEFAULT '',
    last_msg_at TIMESTAMP,
    unread_user INTEGER NOT NULL DEFAULT 0,
    unread_agent INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    closed_at TIMESTAMP
);

COMMENT ON TABLE hg_support_session IS '客服_会话';
COMMENT ON COLUMN hg_support_session.user_id IS '用户ID';
COMMENT ON COLUMN hg_support_session.agent_id IS '客服ID';
COMMENT ON COLUMN hg_support_session.status IS '状态：1排队 2进行中 3已关闭';

CREATE INDEX IF NOT EXISTS idx_support_session_user ON hg_support_session (user_id);
CREATE INDEX IF NOT EXISTS idx_support_session_agent ON hg_support_session (agent_id);
CREATE INDEX IF NOT EXISTS idx_support_session_status ON hg_support_session (status);

-- 消息表
CREATE TABLE IF NOT EXISTS hg_support_message (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL,
    sender_role SMALLINT NOT NULL,
    sender_id BIGINT NOT NULL,
    msg_type SMALLINT NOT NULL DEFAULT 1,
    content TEXT NOT NULL,
    created_at TIMESTAMP
);

COMMENT ON TABLE hg_support_message IS '客服_消息';
COMMENT ON COLUMN hg_support_message.session_id IS '会话ID';
COMMENT ON COLUMN hg_support_message.sender_role IS '发送方角色：1用户 2客服 3系统';

CREATE INDEX IF NOT EXISTS idx_support_message_session ON hg_support_message (session_id);

-- 客服在线状态
CREATE TABLE IF NOT EXISTS hg_support_agent_presence (
    agent_id BIGINT PRIMARY KEY,
    online SMALLINT NOT NULL DEFAULT 0,
    last_seen_at TIMESTAMP,
    updated_at TIMESTAMP
);

COMMENT ON TABLE hg_support_agent_presence IS '客服_在线状态';

-- 客服常用语
CREATE TABLE IF NOT EXISTS hg_support_canned_reply (
    id BIGSERIAL PRIMARY KEY,
    agent_id BIGINT NOT NULL DEFAULT 0,
    title VARCHAR(128) NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    sort INTEGER NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

COMMENT ON TABLE hg_support_canned_reply IS '客服_常用语';

CREATE INDEX IF NOT EXISTS idx_support_canned_reply_agent ON hg_support_canned_reply (agent_id);

-- 验证：检查表是否存在
-- SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename LIKE 'hg_support_%' ORDER BY tablename;


