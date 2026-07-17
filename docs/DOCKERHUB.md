# Gego

Open-source **GEO (Generative Engine Optimization)** tracker. Schedules prompts across LLMs, captures web-search citations, tracks brand mentions, and serves analytics through a built-in dashboard.

This image bundles the **API + pre-built UI** in a single container (port `8989`).

## Quick start

```bash
docker pull ouai2h/gego:latest

docker run -d \
  --name gego \
  -p 8989:8989 \
  -e GEGO_POSTGRES_URI=postgres://user:pass@your-postgres-host:5432/gego?sslmode=disable \
  -e GEGO_MONGODB_URI=mongodb://your-mongodb-host:27017 \
  -e GEGO_MONGODB_DATABASE=gego \
  -e GEGO_JWT_SECRET="your-secret-at-least-32-characters-long" \
  -e GEGO_BOOTSTRAP_ADMIN_PASSWORD="your-admin-password" \
  ouai2h/gego:latest
```

Then open:

- Dashboard: `http://localhost:8989`
- Health: `http://localhost:8989/api/v1/health`

## What you need

Gego expects these services alongside the container:

| Service | Role |
|---------|------|
| **PostgreSQL** | Config, auth, brands, schedules |
| **MongoDB** | Prompts, responses, analytics |
| **etcd** | Job queue for scheduled / manual runs |

For full server deployment steps, see [DEPLOYMENT.md](https://github.com/AI2HU/gego/blob/main/docs/DEPLOYMENT.md).

## Tags

| Tag | Meaning |
|-----|---------|
| `latest` / `main` | Latest build from `main` |
| `vX.Y.Z` | Release version |
| `X.Y` / `X` | Semver aliases |
| `sha-<commit>` | Exact commit build |

Multi-arch: `linux/amd64`, `linux/arm64`.

## Environment variables

| Variable | Description |
|----------|-------------|
| `GEGO_POSTGRES_URI` | PostgreSQL connection string |
| `GEGO_MONGODB_URI` | MongoDB connection string |
| `GEGO_MONGODB_DATABASE` | MongoDB database name (default: `gego`) |
| `GEGO_ETCD_ENDPOINTS` | Comma-separated etcd endpoints (default: `127.0.0.1:2379`) |
| `GEGO_JWT_SECRET` | JWT signing secret (min 32 characters, **required**) |
| `GEGO_BOOTSTRAP_ADMIN_USERNAME` | First admin username (default: `admin`) |
| `GEGO_BOOTSTRAP_ADMIN_PASSWORD` | First admin password (min 8 characters, **required** on first run) |
| `GEGO_COOKIE_SECURE` | Set to `true` to mark auth cookies as Secure |
| `GEGO_CONFIG_PATH` | Config file path (default: `/app/config/config.yaml`) |
| `GEGO_DATA_PATH` | Legacy SQLite data directory (default: `/app/data`) |
| `GEGO_LOG_PATH` | Log directory (default: `/app/logs`) |

## Features

- Multi-LLM: OpenAI, Anthropic, Ollama, Google, Perplexity, and custom providers
- Brand & citation tracking from LLM web-search results
- Keyword analytics with exclusion words
- Cron scheduler + workers via etcd
- JWT auth with roles (`admin`, `member`)
- Vue dashboard served from the same image

## Links

- Source: [github.com/AI2HU/gego](https://github.com/AI2HU/gego)
- Deployment guide: [docs/DEPLOYMENT.md](https://github.com/AI2HU/gego/blob/main/docs/DEPLOYMENT.md)
- License: [GPL-3.0](https://github.com/AI2HU/gego/blob/main/LICENSE)
