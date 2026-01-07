-- =============================================================================
-- Support Chat Menu (PostgreSQL)
-- 作用：
-- 1) 修复 hg_admin_menu.id 没有默认自增导致 insert id=null 的问题（若已配置则跳过）
-- 2) 在“量化管理”下创建：客服管理（目录）-> 实时对话（菜单：/supportChat/workbench）
--
-- 使用方式：
-- - Navicat：打开本文件直接执行
-- - psql：psql "postgresql://user:pass@host:5432/db?sslmode=disable" -f support_chat_menu_pg.sql
-- =============================================================================

DO $$
DECLARE
  has_default BOOLEAN;
  parent_id BIGINT;
  parent_level INT;
  parent_tree TEXT;

  kefu_dir_id BIGINT;
  dialog_menu_id BIGINT;
  now_ts TIMESTAMP := NOW();

  kefu_dir_name TEXT := 'trading_support_mgmt';
  kefu_dir_title TEXT := '客服管理';

  dialog_menu_name TEXT := 'trading_support_realtime_chat';
  dialog_menu_title TEXT := '实时对话';
BEGIN
  -- ---------------------------------------------------------------------------
  -- 1) 确保 hg_admin_menu.id 有默认自增（否则 insert 会触发 id=null 报错）
  -- ---------------------------------------------------------------------------
  SELECT (column_default IS NOT NULL)
  INTO has_default
  FROM information_schema.columns
  WHERE table_schema = 'public' AND table_name = 'hg_admin_menu' AND column_name = 'id';

  IF NOT has_default THEN
    -- 1.1 创建序列（若不存在）
    IF NOT EXISTS (
      SELECT 1 FROM pg_class c
      JOIN pg_namespace n ON n.oid = c.relnamespace
      WHERE c.relkind = 'S' AND c.relname = 'hg_admin_menu_id_seq' AND n.nspname = 'public'
    ) THEN
      EXECUTE 'CREATE SEQUENCE public.hg_admin_menu_id_seq';
    END IF;

    -- 1.2 拨到当前最大id之后
    EXECUTE format(
      'SELECT setval(''public.hg_admin_menu_id_seq'', %s, false)',
      COALESCE((SELECT MAX(id) FROM public.hg_admin_menu), 0) + 1
    );

    -- 1.3 设置默认 nextval
    EXECUTE 'ALTER TABLE public.hg_admin_menu ALTER COLUMN id SET DEFAULT nextval(''public.hg_admin_menu_id_seq'')';
  END IF;

  -- ---------------------------------------------------------------------------
  -- 2) 找父级“量化管理”
  -- 说明：你们项目可能叫“量化管理/量化交易/Toogo”等，做多条件兜底匹配。
  -- 找不到会直接报错，避免插错位置。
  -- ---------------------------------------------------------------------------
  WITH parent_candidates AS (
    SELECT id, level, tree
    FROM public.hg_admin_menu
    WHERE status = 1
      AND type = 1
      AND (
        title IN ('量化管理', '量化交易', 'ToogoAdmin', 'Toogo')
        OR name  IN ('trading', 'ToogoAdmin', 'ToogoRoot', 'toogo')
        OR path  IN ('/trading', '/toogo-admin', '/toogo')
      )
    ORDER BY id DESC
    LIMIT 1
  )
  SELECT id, level, tree INTO parent_id, parent_level, parent_tree
  FROM parent_candidates;

  IF parent_id IS NULL THEN
    RAISE EXCEPTION '未找到父级菜单【量化管理】。请先确认 hg_admin_menu 中量化管理的 title/name/path，并修改本脚本 parent_candidates 条件。';
  END IF;

  -- ---------------------------------------------------------------------------
  -- 3) 创建/复用 “客服管理”(目录)
  -- tree 规则：child.tree = parent.tree || 'tr_' || parent.id || ' '
  -- ---------------------------------------------------------------------------
  SELECT id INTO kefu_dir_id
  FROM public.hg_admin_menu
  WHERE pid = parent_id AND name = kefu_dir_name
  ORDER BY id DESC
  LIMIT 1;

  IF kefu_dir_id IS NULL THEN
    INSERT INTO public.hg_admin_menu (
      pid, level, tree, title, name, path, icon, type, redirect,
      permissions, permission_name, component, always_show, active_menu,
      is_root, is_frame, frame_src, keep_alive, hidden, affix, sort,
      remark, status, updated_at, created_at
    ) VALUES (
      parent_id,
      parent_level + 1,
      COALESCE(parent_tree, '') || 'tr_' || parent_id::TEXT || ' ',
      kefu_dir_title,
      kefu_dir_name,
      'support',
      'CustomerServiceOutlined',
      1,
      '/supportChat/workbench',
      '',
      '',
      'ParentLayout',
      1,
      '',
      0,
      0,
      '',
      0,
      0,
      0,
      90,
      '量化管理-客服管理',
      1,
      now_ts,
      now_ts
    )
    RETURNING id INTO kefu_dir_id;
  END IF;

  -- ---------------------------------------------------------------------------
  -- 4) 创建/复用 “实时对话”(菜单)
  -- - path: /supportChat/workbench
  -- - component: /supportChat/index （对应 web/src/views/supportChat/index.vue）
  -- ---------------------------------------------------------------------------
  SELECT id INTO dialog_menu_id
  FROM public.hg_admin_menu
  WHERE pid = kefu_dir_id AND name = dialog_menu_name
  ORDER BY id DESC
  LIMIT 1;

  IF dialog_menu_id IS NULL THEN
    INSERT INTO public.hg_admin_menu (
      pid, level, tree, title, name, path, icon, type, redirect,
      permissions, permission_name, component, always_show, active_menu,
      is_root, is_frame, frame_src, keep_alive, hidden, affix, sort,
      remark, status, updated_at, created_at
    ) VALUES (
      kefu_dir_id,
      parent_level + 2,
      COALESCE(parent_tree, '') || 'tr_' || parent_id::TEXT || ' ' || 'tr_' || kefu_dir_id::TEXT || ' ',
      dialog_menu_title,
      dialog_menu_name,
      '/supportChat/workbench',
      '',
      2,
      '',
      '/supportChat/sessionList,/supportChat/accept,/supportChat/acceptNext,/supportChat/messageList,/supportChat/send,/supportChat/transfer,/supportChat/close,/supportChat/canned/list,/supportChat/canned/edit,/supportChat/canned/delete,/supportChat/agentOnline',
      '客服实时对话',
      '/supportChat/index',
      0,
      '',
      0,
      0,
      '',
      1,
      0,
      0,
      10,
      '客服工作台入口',
      1,
      now_ts,
      now_ts
    )
    RETURNING id INTO dialog_menu_id;
  END IF;

  RAISE NOTICE 'OK: parent_id=%, kefu_dir_id=%, dialog_menu_id=%', parent_id, kefu_dir_id, dialog_menu_id;
END $$;


