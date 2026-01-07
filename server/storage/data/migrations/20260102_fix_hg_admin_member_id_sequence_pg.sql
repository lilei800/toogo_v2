-- Fix for PostgreSQL: ensure hg_admin_member.id has auto-increment default.
-- Symptom:
--   pq: null value in column "id" of relation "hg_admin_member" violates not-null constraint
-- Cause:
--   Table exists but id column is NOT backed by a sequence/identity default.
--
-- This script is safe to run multiple times.

DO $$
DECLARE
  seq_name text := 'hg_admin_member_id_seq';
  has_default boolean := false;
BEGIN
  -- Detect whether id already has a default.
  SELECT (column_default IS NOT NULL)
    INTO has_default
  FROM information_schema.columns
  WHERE table_schema = 'public'
    AND table_name = 'hg_admin_member'
    AND column_name = 'id';

  IF NOT has_default THEN
    -- Create and bind sequence.
    EXECUTE format('CREATE SEQUENCE IF NOT EXISTS %I', seq_name);
    EXECUTE format('ALTER SEQUENCE %I OWNED BY hg_admin_member.id', seq_name);
    EXECUTE format('ALTER TABLE hg_admin_member ALTER COLUMN id SET DEFAULT nextval(%L)', seq_name);
  END IF;

  -- Advance sequence to max(id)+1 to avoid collisions.
  EXECUTE format(
    'SELECT setval(%L, COALESCE((SELECT MAX(id) FROM hg_admin_member), 0) + 1, false)',
    seq_name
  );
END $$;


