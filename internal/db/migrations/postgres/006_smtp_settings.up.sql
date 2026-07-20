CREATE TABLE IF NOT EXISTS smtp_settings (
    id TEXT PRIMARY KEY DEFAULT 'default',
    host TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 587,
    username TEXT NOT NULL DEFAULT '',
    password TEXT NOT NULL DEFAULT '',
    from_email TEXT NOT NULL DEFAULT '',
    from_name TEXT NOT NULL DEFAULT '',
    use_tls BOOLEAN NOT NULL DEFAULT TRUE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT smtp_settings_singleton CHECK (id = 'default')
);

INSERT INTO smtp_settings (id)
VALUES ('default')
ON CONFLICT (id) DO NOTHING;
