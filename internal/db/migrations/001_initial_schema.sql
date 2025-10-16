-- Migration: 001_initial_schema.sql
-- Description: Initial database schema creation for Gego
-- Created: 2024-01-01
-- Author: Gego System

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Create migrations table to track applied migrations
CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checksum TEXT NOT NULL
);

-- Create LLMs table for storing LLM provider configurations
CREATE TABLE IF NOT EXISTS llms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN ('openai', 'anthropic', 'ollama', 'google', 'perplexity')),
    model TEXT NOT NULL,
    api_key TEXT,
    base_url TEXT,
    config TEXT DEFAULT '{}', -- JSON string for additional provider-specific config
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create Schedules table for storing scheduler configurations
CREATE TABLE IF NOT EXISTS schedules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    prompt_ids TEXT NOT NULL DEFAULT '[]', -- JSON array of prompt IDs
    llm_ids TEXT NOT NULL DEFAULT '[]',    -- JSON array of LLM IDs
    cron_expr TEXT NOT NULL,
    temperature REAL DEFAULT 0.7 CHECK (temperature >= 0.0 AND temperature <= 1.0),
    enabled BOOLEAN NOT NULL DEFAULT 1,
    last_run DATETIME,
    next_run DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for optimal query performance
CREATE INDEX IF NOT EXISTS idx_llms_provider ON llms(provider);
CREATE INDEX IF NOT EXISTS idx_llms_enabled ON llms(enabled);
CREATE INDEX IF NOT EXISTS idx_llms_created_at ON llms(created_at);
CREATE INDEX IF NOT EXISTS idx_llms_updated_at ON llms(updated_at);

CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);
CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run);
CREATE INDEX IF NOT EXISTS idx_schedules_created_at ON schedules(created_at);
CREATE INDEX IF NOT EXISTS idx_schedules_updated_at ON schedules(updated_at);
CREATE INDEX IF NOT EXISTS idx_schedules_cron_expr ON schedules(cron_expr);

-- Create triggers to automatically update the updated_at timestamp
CREATE TRIGGER IF NOT EXISTS trigger_llms_updated_at 
    AFTER UPDATE ON llms
    FOR EACH ROW
    BEGIN
        UPDATE llms SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS trigger_schedules_updated_at 
    AFTER UPDATE ON schedules
    FOR EACH ROW
    BEGIN
        UPDATE schedules SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- Insert migration record
INSERT INTO migrations (version, description, checksum) 
VALUES ('001', 'Initial database schema creation', 'sha256:initial_schema_v1');

-- Create views for common queries
CREATE VIEW IF NOT EXISTS v_enabled_llms AS
SELECT 
    id,
    name,
    provider,
    model,
    base_url,
    config,
    created_at,
    updated_at
FROM llms 
WHERE enabled = 1
ORDER BY created_at DESC;

CREATE VIEW IF NOT EXISTS v_enabled_schedules AS
SELECT 
    id,
    name,
    prompt_ids,
    llm_ids,
    cron_expr,
    temperature,
    last_run,
    next_run,
    created_at,
    updated_at
FROM schedules 
WHERE enabled = 1
ORDER BY next_run ASC;

-- Create view for schedule statistics
CREATE VIEW IF NOT EXISTS v_schedule_stats AS
SELECT 
    s.id,
    s.name,
    s.cron_expr,
    s.enabled,
    s.last_run,
    s.next_run,
    COUNT(DISTINCT json_extract(s.prompt_ids, '$[' || i.value || ']')) as prompt_count,
    COUNT(DISTINCT json_extract(s.llm_ids, '$[' || j.value || ']')) as llm_count
FROM schedules s
LEFT JOIN json_each(s.prompt_ids) i ON 1=1
LEFT JOIN json_each(s.llm_ids) j ON 1=1
GROUP BY s.id, s.name, s.cron_expr, s.enabled, s.last_run, s.next_run;

-- Add comments for documentation
-- LLMs table stores configuration for different LLM providers
-- Each LLM has a unique ID, name, provider type, and model specification
-- API keys and base URLs are stored for authentication and connection
-- Additional configuration can be stored as JSON in the config field

-- Schedules table stores cron-based scheduling configurations
-- Each schedule references multiple prompts and LLMs via JSON arrays
-- Temperature controls the randomness of LLM responses (0.0-1.0)
-- last_run and next_run track execution timing
-- Cron expressions follow standard cron format

-- Indexes are created for common query patterns:
-- - Provider-based LLM filtering
-- - Enabled status filtering
-- - Time-based sorting and filtering
-- - Schedule execution timing queries
