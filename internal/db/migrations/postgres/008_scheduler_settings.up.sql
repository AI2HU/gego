CREATE TABLE IF NOT EXISTS scheduler_settings (
    id TEXT PRIMARY KEY DEFAULT 'default',
    desired_running BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT scheduler_settings_singleton CHECK (id = 'default')
);

INSERT INTO scheduler_settings (id)
VALUES ('default')
ON CONFLICT (id) DO NOTHING;
