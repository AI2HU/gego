-- Migration: 003_user_sessions.up.sql
-- Description: Add user_sessions table for refresh token rotation
-- Author: AI2HU

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS user_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_hash ON user_sessions(token_hash);

CREATE TRIGGER IF NOT EXISTS trigger_user_sessions_updated_at
    AFTER UPDATE ON user_sessions
    FOR EACH ROW
    BEGIN
        UPDATE user_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
