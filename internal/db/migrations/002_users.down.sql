-- Migration: 002_users.down.sql
-- Description: Rollback users table for JWT authentication
-- Author: AI2HU

DROP TRIGGER IF EXISTS trigger_users_updated_at;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;
