CREATE TABLE IF NOT EXISTS password_invites (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_invites_user_id ON password_invites(user_id);
CREATE INDEX IF NOT EXISTS idx_password_invites_token_hash ON password_invites(token_hash);
CREATE INDEX IF NOT EXISTS idx_password_invites_expires_at ON password_invites(expires_at);
