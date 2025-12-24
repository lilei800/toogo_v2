-- =========================================================
-- PostgreSQL compatibility helpers for HotGo/Toogo
-- Purpose:
--   Make a small subset of MySQL-style expressions work on PG,
--   so we can do "downtime migration" without blocking on code refactors.
--
-- Safe to run repeatedly.
-- =========================================================

-- MySQL: IF(cond, a, b)
CREATE OR REPLACE FUNCTION public."IF"(cond boolean, t anyelement, f anyelement)
RETURNS anyelement
LANGUAGE plpgsql
IMMUTABLE
AS $$
BEGIN
  IF cond THEN
    RETURN t;
  END IF;
  RETURN f;
END;
$$;

-- MySQL: IFNULL(a, b)
CREATE OR REPLACE FUNCTION public."IFNULL"(a anyelement, b anyelement)
RETURNS anyelement
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT COALESCE($1, $2);
$$;

-- MySQL: YEAR(ts)
CREATE OR REPLACE FUNCTION public."YEAR"(ts timestamp)
RETURNS integer
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT EXTRACT(YEAR FROM $1)::int;
$$;

CREATE OR REPLACE FUNCTION public."YEAR"(ts timestamptz)
RETURNS integer
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT EXTRACT(YEAR FROM $1)::int;
$$;

CREATE OR REPLACE FUNCTION public."YEAR"(d date)
RETURNS integer
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT EXTRACT(YEAR FROM $1)::int;
$$;


