CREATE TABLE IF NOT EXISTS llms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN ('openai', 'anthropic', 'ollama', 'google', 'perplexity')),
    model TEXT NOT NULL,
    api_key TEXT,
    base_url TEXT,
    config JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    prompt_ids JSONB NOT NULL DEFAULT '[]',
    llm_ids JSONB NOT NULL DEFAULT '[]',
    cron_expr TEXT NOT NULL,
    temperature DOUBLE PRECISION DEFAULT 0.7 CHECK (temperature >= 0.0 AND temperature <= 1.0),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_run TIMESTAMPTZ,
    next_run TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_llms_provider ON llms(provider);
CREATE INDEX IF NOT EXISTS idx_llms_enabled ON llms(enabled);
CREATE INDEX IF NOT EXISTS idx_llms_created_at ON llms(created_at);
CREATE INDEX IF NOT EXISTS idx_llms_updated_at ON llms(updated_at);

CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);
CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run);
CREATE INDEX IF NOT EXISTS idx_schedules_created_at ON schedules(created_at);
CREATE INDEX IF NOT EXISTS idx_schedules_updated_at ON schedules(updated_at);
CREATE INDEX IF NOT EXISTS idx_schedules_cron_expr ON schedules(cron_expr);
