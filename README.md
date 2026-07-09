# Gego - GEO System for your brand, working with all LLMs

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org/)

Gego is an open-source GEO (Generative Engine Optimization) tracker. It schedules prompts across multiple Large Language Models (LLMs), captures web-search citations from their responses, and surfaces keyword and domain analytics through a built-in dashboard and CLI.

## Features

- **Multi-LLM support**: OpenAI, Anthropic, Ollama, Google, Perplexity (Sonar), and pluggable custom providers
- **Built-in web dashboard** (`gego-ui`): Vue 3 dashboard for stats, search, models, prompts, scheduler, and error logs
- **Citation tracking**: Extracts cited URLs and domains from provider web-search results (OpenAI, Anthropic, Google, Perplexity)
- **Keyword analytics**: Automatic keyword extraction from responses with configurable exclusion lists
- **Hybrid database**: SQLite for configuration and auth; MongoDB for prompts, responses, and analytics
- **JWT authentication**: Role-based access (`admin`, `member`) with session refresh
- **Flexible scheduling**: Cron-based scheduler for automated prompt execution
- **Prompt generation**: AI-assisted prompt creation via the API
- **Tag-based filtering**: Organize prompts with tags and filter dashboard/search stats
- **Error logs**: Review failed scheduled LLM calls (rate limits, provider errors)
- **CLI and REST API**: Full management from the terminal or HTTP
- **Retry mechanism**: Automatic retry with delays for failed requests
- **Configurable logging**: DEBUG, INFO, WARNING, ERROR levels with optional file output
- **Docker-ready**: Single image bundles the API and pre-built UI

## Use Cases

- **GEO / brand visibility**: Track how your brand and competitors appear in AI-generated answers
- **Citation analysis**: See which domains and URLs LLMs cite most often for your prompts
- **SEO and marketing research**: Monitor keyword mentions across AI assistants
- **Competitive analysis**: Compare visibility across providers and models
- **Prompt engineering**: Identify which prompts drive the most mentions and citations

## Project Structure

```
gego/
├── cmd/gego/           # CLI entrypoint
├── gego-ui/            # Vue 3 + Vite dashboard (served by the API in production)
├── internal/
│   ├── api/            # Gin REST API and static UI serving
│   ├── auth/           # JWT middleware, permissions, sessions
│   ├── cli/            # Cobra commands
│   ├── db/             # SQLite + MongoDB hybrid layer and migrations
│   ├── llm/            # Provider implementations (openai, anthropic, google, …)
│   ├── models/         # Domain and API types
│   └── services/       # Business logic (scheduler, stats, search, auth, …)
├── docs/               # Deployment and usage guides
├── Dockerfile          # Production image (API + UI)
└── Makefile            # build, dev, ui-* targets
```

## Installation

### Prerequisites

- Go 1.25 or higher
- Node.js 20.19+ or 22.12+ (for UI development)
- MongoDB (analytics data)
- API keys for LLM providers (OpenAI, Anthropic, etc.)

### Build from Source

```bash
git clone https://github.com/AI2HU/gego.git
cd gego
make build
```

The binary is written to `build/gego`.

### Install via Go

```bash
go install github.com/AI2HU/gego/cmd/gego@latest
```

### Docker

The production image builds the UI and API into one container. The dashboard is served at the same port as the API.

> For deployment details, see [DEPLOYMENT.md](docs/DEPLOYMENT.md)

```bash
docker build -t gego:latest .

docker run -d \
  --name gego \
  -p 8989:8989 \
  -e MONGODB_URI=mongodb://your-mongodb-host:27017 \
  -e GEGO_JWT_SECRET="your-secret-at-least-32-characters-long" \
  -e GEGO_BOOTSTRAP_ADMIN_PASSWORD="your-admin-password" \
  gego:latest
```

**Docker environment variables**

| Variable | Description |
|----------|-------------|
| `GEGO_CONFIG_PATH` | Config file path (default: `/app/config/config.yaml`) |
| `GEGO_DATA_PATH` | SQLite data directory (default: `/app/data`) |
| `GEGO_LOG_PATH` | Log directory (default: `/app/logs`) |
| `GEGO_JWT_SECRET` | JWT signing secret (min 32 characters, **required**) |
| `GEGO_BOOTSTRAP_ADMIN_USERNAME` | First admin username (default: `admin`) |
| `GEGO_BOOTSTRAP_ADMIN_PASSWORD` | First admin password (min 8 characters, **required** on first run) |
| `GEGO_COOKIE_SECURE` | Set to `true` to mark auth cookies as Secure |

## Quick Start

### Production-like local run (API + UI on one port)

Copy [`.env.dev.example`](.env.dev.example) to `.env.dev`, start PostgreSQL and MongoDB, then:

```bash
make dev
```

This builds the UI, loads [dev database fixtures](CONTRIBUTING.md#local-development-with-fixtures) (sample LLMs, prompts, responses, brands), and starts the API with the embedded dashboard.

| Service | URL |
|---------|-----|
| Dashboard | http://localhost:8989 |
| API | http://localhost:8989/api/v1 |

Sign in with your bootstrap admin credentials (default username: `admin`; default dev password when using `make dev`: `admin1234`).

For manual setup without fixtures, or hot-reload UI development:

```bash
# 1. Initialize Gego (first time only)
./build/gego init

# 2. Build CLI and UI
make build
make ui-build

# 3. Set auth env vars (required for the API)
export GEGO_JWT_SECRET="your-secret-at-least-32-characters-long"
export GEGO_BOOTSTRAP_ADMIN_PASSWORD="your-admin-password"

# 4. Start API with embedded UI (no fixture reload)
make dev-api
```

### Hot-reload development (separate UI dev server)

```bash
make dev-api   # Terminal 1 — API on http://localhost:8989
make ui-dev    # Terminal 2 — Vite on http://localhost:5173 (proxies /api to the API)
```

See `gego-ui/.env.example` for optional UI environment variables.

## Web Dashboard

The dashboard lives in `gego-ui/` and is included in the repository.

| Page | Path | Description |
|------|------|-------------|
| Dashboard | `/` | Keyword trends, provider distribution, top cited domains |
| Search | `/search` | Full-text keyword search across stored responses |
| Models | `/admin/models` | Manage LLM provider configurations |
| Prompts | `/admin/prompts` | Create, tag, and filter prompt templates |
| Scheduler | `/admin/scheduler` | Cron schedules and background scheduler control |
| Logs | `/admin/logs` | Failed LLM executions from scheduled runs |

Admin pages require the `admin` role. Members can access dashboard and search.

## Authentication

The API uses JWT access tokens with refresh-token rotation stored in SQLite.

**Bootstrap the first admin** (when no users exist):

```bash
export GEGO_JWT_SECRET="your-secret-at-least-32-characters-long"
export GEGO_BOOTSTRAP_ADMIN_PASSWORD="your-admin-password"
./build/gego api
```

Or create users manually:

```bash
./build/gego user create --username admin --password "your-password" --role admin
./build/gego user list
```

**Roles**

| Role | Access |
|------|--------|
| `admin` | Full read/write on models, prompts, schedules, stats, search, logs |
| `member` | Read-only on models/prompts/schedules; dashboard, stats, and search |

**Auth endpoints** (public unless noted)

- `POST /api/v1/auth/login` — obtain tokens
- `POST /api/v1/auth/refresh` — rotate access token
- `POST /api/v1/auth/logout` — invalidate session
- `GET /api/v1/auth/me` — current user (authenticated)

All other `/api/v1/*` routes require a valid `Authorization: Bearer <token>` header.

## CLI Workflow

### 1. Initialize configuration

```bash
gego init
```

Interactive setup for SQLite and MongoDB connections.

### 2. Add LLM providers

```bash
gego llm add
```

Supported providers include OpenAI, Anthropic, Ollama, Google (Gemini), and Perplexity (Sonar).

### 3. Create prompts

```bash
gego prompt add
```

Prompts support tags for filtering dashboard and search results.

### 4. Set up schedules

```bash
gego schedule add
```

Cron expressions run selected prompts against selected models automatically.

### 5. Run prompts

```bash
gego run                  # Run all enabled prompts with all enabled LLMs once (direct execution)
```

### 6. Start etcd, workers, and scheduler

Schedule execution uses **etcd** for the job queue and **worker processes** for LLM calls.

```bash
make etcd-dev             # Terminal 1: local etcd on :2379
make dev-worker           # Terminal 2: job worker
gego scheduler start      # Terminal 3: cron enqueuer (requires etcd)
```

Or run everything with Docker:

```bash
make dev-up               # etcd + api + worker
```

Required env: `GEGO_ETCD_ENDPOINTS` (default `127.0.0.1:2379`).

### 7. Start the API server

```bash
gego api                          # Default: 0.0.0.0:8989
gego api --port 3000              # Custom port
gego api --cors-origin "https://myapp.com"
```

When `gego-ui/dist` exists (or `/app/ui` in Docker), the API also serves the dashboard as static files.

## API Reference

Base URL: `http://localhost:8989/api/v1`

### Health

- `GET /health` — health check (public)

### Auth

- `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`
- `GET /auth/me`

### Models and providers

- `GET /providers` — list supported LLM providers
- `GET /providers/:provider/api-keys` — list stored API keys for a provider
- `POST /providers/:provider/models` — list models available from a provider
- `GET /models`, `GET /models/:id`, `POST /models`, `PUT /models/:id`, `DELETE /models/:id`

### Prompts

- `GET /prompts`, `GET /prompts/:id`, `POST /prompts`, `PUT /prompts/:id`, `DELETE /prompts/:id`
- `POST /prompts/generate` — AI-assisted prompt generation

### Schedules and scheduler

- `GET /schedules`, `GET /schedules/:id`, `POST /schedules`, `PUT /schedules/:id`, `DELETE /schedules/:id`
- `POST /schedules/:id/run` — enqueue a schedule run (`202`, returns `run_id`)
- `GET /schedule-runs`, `GET /schedule-runs/:id`, `GET /schedule-runs/:id/jobs`
- `POST /schedule-runs/:id/cancel`, `POST /schedule-runs/:id/jobs/:job_id/retry`
- `GET /scheduler/status`, `GET /scheduler/health`, `POST /scheduler/start`, `POST /scheduler/stop`, `POST /scheduler/reload`

### Analytics and search

- `GET /stats` — dashboard stats (keywords, trends, provider breakdown)
- `GET /stats/urls` — top cited URLs and domains
- `GET /stats/query-urls` — URLs grouped by search query
- `GET /stats/keyword-domains` — keyword × domain matrix
- `POST /search` — keyword search across responses

### Logs

- `GET /logs/errors` — failed LLM calls from scheduled executions

**Example**

```bash
curl http://localhost:8989/api/v1/health

curl -X POST http://localhost:8989/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'

curl http://localhost:8989/api/v1/stats \
  -H "Authorization: Bearer <access_token>"
```

> More examples: [EXAMPLES.md](docs/EXAMPLES.md)

## Usage Examples

### View statistics

```bash
gego stats keywords --limit 20
gego stats keyword Dior
```

### Manage LLMs, prompts, schedules

```bash
gego llm list
gego prompt list
gego schedule list
gego schedule run <id>
gego scheduler status
```

## Configuration

Configuration is stored in `~/.gego/config.yaml`:

```yaml
sql_database:
  provider: sqlite
  uri: ~/.gego/gego.db
  database: gego

nosql_database:
  provider: mongodb
  uri: mongodb://localhost:27017
  database: gego

auth:
  issuer: gego-api
  audience: gego-api
  access_token_ttl: 15m
  refresh_token_ttl: 168h
```

**Database architecture**

| Store | Contents |
|-------|----------|
| SQLite | LLM configs, schedules, users, sessions |
| MongoDB | Prompts, responses (including `search_urls` citations), analytics |

### Keywords exclusion

Gego filters common words from keyword extraction. Customize the list in `~/.gego/keywords_exclusion` (one word per line; `#` for comments). Restart the application after changes.

### System instructions

Optional provider-specific system prompts in `config.yaml`:

- `gemini_system_instruction`
- `chatgpt_system_instruction`
- `claude_system_instruction`

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the CLI to `build/gego` |
| `make ui-install` | Install `gego-ui` npm dependencies |
| `make ui-build` | Build the dashboard to `gego-ui/dist` |
| `make ui-dev` | Vite dev server with hot reload (port 5173) |
| `make dev` | Build UI, load dev fixtures, start API with embedded static UI (port 8989) |
| `make fixtures-dev` | Reset and load dev database fixtures only (see [CONTRIBUTING.md](CONTRIBUTING.md#local-development-with-fixtures)) |
| `make dev-api` | Start API only (port 8989); does not load fixtures |
| `make test` | Run Go tests |
| `make build-all` | Cross-platform binaries |

## Logging

Control verbosity with `--log-level` (`DEBUG`, `INFO`, `WARNING`, `ERROR`) and optional `--log-file`:

```bash
gego run --log-level DEBUG
gego api --log-level INFO --log-file /var/log/gego/app.log
```

Failed prompt executions are retried up to 3 times with a 30-second delay between attempts.

## Architecture

```
┌──────────────┐     ┌─────────────────┐
│  gego-ui     │     │   CLI (Cobra)   │
│  (Vue 3)     │     └────────┬────────┘
└──────┬───────┘              │
       │                      │
       └──────────┬───────────┘
                  │
           ┌──────┴──────┐
           │  REST API   │
           │   (Gin)     │
           └──────┬──────┘
                  │
     ┌────────────┼────────────┐
     │            │            │
┌────┴────┐  ┌────┴────┐  ┌────┴─────────┐
│ SQLite  │  │ MongoDB │  │ LLM Registry │
│ users   │  │ prompts │  │ OpenAI       │
│ llms    │  │ responses│ │ Anthropic    │
│ schedules│ │ citations│ │ Google …     │
└─────────┘  └─────────┘  └──────────────┘
                  │
           ┌──────┴──────┐
           │  Scheduler  │
           └─────────────┘
```

LLM providers return `SearchURLs` alongside response text. Citations are stored on each response and aggregated for domain/URL stats.

## Adding Custom LLM Providers

Implement the `llm.Provider` interface:

```go
type Provider interface {
    Name() string
    Generate(ctx context.Context, prompt string, config Config) (*Response, error)
    Validate(config map[string]string) error
    ListModels(ctx context.Context, apiKey, baseURL string) ([]models.ModelInfo, error)
}
```

Register your provider in the LLM registry (see `internal/services/llm_registry.go`).

## Contributing

Contributions are welcome.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Roadmap

- [ ] Persona embedding to simulate chat-style model behavior
- [ ] System prompts per model for chat simulation
- [ ] Schedule run-time estimation and cost forecasting
- [ ] Prompt batches to optimize costs
- [ ] Provider-specific prompt threading for speed
- [ ] Additional NoSQL database support (Cassandra, etc.)
- [ ] Export statistics to CSV/JSON
- [ ] Webhook notifications
- [ ] Custom keyword extraction rules
- [ ] Time-series trend analysis

## License

This project is licensed under the GNU General Public License v3.0 — see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Gin](https://github.com/gin-gonic/gin) — HTTP API
- [Vue](https://vuejs.org/) — dashboard UI
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) — analytics database
- [go-sqlite3](https://github.com/mattn/go-sqlite3) — configuration database
- [robfig/cron](https://github.com/robfig/cron) — scheduling

## Support

- Email: jonathan@blocs.fr
- Issues: [GitHub Issues](https://github.com/AI2HU/gego/issues)
- Discussions: [GitHub Discussions](https://github.com/AI2HU/gego/discussions)

---

Made with care for the open-source community
