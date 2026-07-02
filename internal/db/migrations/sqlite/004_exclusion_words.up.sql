CREATE TABLE IF NOT EXISTS exclusion_words (
    id TEXT PRIMARY KEY,
    word TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exclusion_words_word_lower ON exclusion_words (LOWER(word));
CREATE INDEX IF NOT EXISTS idx_exclusion_words_created_at ON exclusion_words (created_at);
