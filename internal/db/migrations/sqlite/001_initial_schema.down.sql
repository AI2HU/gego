-- Migration: 001_initial_schema.down.sql
-- Description: Rollback initial database schema creation for Gego
-- Author: AI2HU

-- Drop views (order matters - drop dependent views first)
DROP VIEW IF EXISTS v_schedule_stats;
DROP VIEW IF EXISTS v_enabled_schedules;
DROP VIEW IF EXISTS v_enabled_llms;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_schedules_updated_at;
DROP TRIGGER IF EXISTS trigger_llms_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_schedules_cron_expr;
DROP INDEX IF EXISTS idx_schedules_updated_at;
DROP INDEX IF EXISTS idx_schedules_created_at;
DROP INDEX IF EXISTS idx_schedules_next_run;
DROP INDEX IF EXISTS idx_schedules_enabled;

DROP INDEX IF EXISTS idx_llms_updated_at;
DROP INDEX IF EXISTS idx_llms_created_at;
DROP INDEX IF EXISTS idx_llms_enabled;
DROP INDEX IF EXISTS idx_llms_provider;

-- Drop tables (order matters - drop dependent tables first)
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS llms;

