# Database layer

Gego uses a hybrid storage model:

| Store | Role | Status |
|-------|------|--------|
| **PostgreSQL** | LLMs, schedules, users, sessions, exclusion words | Active — all new SQL features go here |
| **MongoDB** | Prompts, responses, analytics queries | Active |
| **SQLite** | Same tables as PostgreSQL for older deployments | **Legacy only** |

## SQLite is legacy

SQLite remains in the codebase only for existing installations that have not migrated to PostgreSQL yet (`gego db upgrade-from-sqlite` or the admin upgrade flow).

**Do not add or change SQLite code when implementing new features.**

When working on SQL-backed functionality:

1. Add migrations under `migrations/postgres/` only — not `migrations/sqlite/`.
2. Implement data access in `internal/db/postgres/` only.
3. Extend `SQLDatabase` in `sql.go` and forward methods from `hybrid.go` to the SQL backend.
4. Leave `internal/db/sqlite/` and `migrations/sqlite/` unchanged unless you are explicitly fixing a migration-blocking bug for legacy users.

New deployments and all active development paths use PostgreSQL.
