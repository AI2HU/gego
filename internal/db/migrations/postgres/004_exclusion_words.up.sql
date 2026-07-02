CREATE TABLE IF NOT EXISTS exclusion_words (
    id TEXT PRIMARY KEY,
    word TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exclusion_words_word_lower ON exclusion_words (LOWER(word));
CREATE INDEX IF NOT EXISTS idx_exclusion_words_created_at ON exclusion_words (created_at);
