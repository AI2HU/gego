# Contributing to Gego

Thank you for your interest in contributing to Gego! We welcome contributions from the community.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/AI2HU/gego.git`
3. Create a branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Test your changes
6. Commit: `git commit -m "Add your feature"`
7. Push: `git push origin feature/your-feature-name`
8. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL (active SQL store for development)
- MongoDB (prompts, responses, analytics)
- Make (optional but recommended)

### Setup

```bash
# Clone the repository
git clone https://github.com/AI2HU/gego.git
cd gego

# Install dependencies
go mod download

# Copy dev environment variables (database URIs, auth, fixtures flag)
cp .env.dev.example .env.dev

# Build
make build

# Run
./build/gego --help
```

### Running databases for development

Gego uses a hybrid store: **PostgreSQL** for LLMs, schedules, users, brands, and exclusion words; **MongoDB** for prompts and responses.

```bash
# PostgreSQL (example with Docker)
docker run -d -p 5432:5432 \
  -e POSTGRES_USER=gego \
  -e POSTGRES_PASSWORD=gego \
  -e POSTGRES_DB=gego \
  --name gego-postgres postgres:16

# MongoDB
docker run -d -p 27017:27017 --name gego-mongo mongo:latest
```

Adjust `GEGO_POSTGRES_URI` and `GEGO_MONGODB_URI` in `.env.dev` if your connection strings differ. See [`.env.dev.example`](.env.dev.example).

### Local development with fixtures

For dashboard and API work, use the **dev fixture system** to get predictable sample data without manual setup or real LLM API keys.

#### Quick start

```bash
cp .env.dev.example .env.dev   # first time only
make dev                       # build UI, load fixtures, start API on :8989
```

Sign in at http://localhost:8989 with the bootstrap admin from `.env.dev` (default: `admin` / `admin1234`).

`make dev` runs, in order:

1. `make build` — compile the CLI
2. `make ui-build` — build the dashboard
3. `make fixtures-dev` — clean and seed PostgreSQL + MongoDB
4. Start the API with the embedded UI

#### What fixtures provide

| Store | Data |
|-------|------|
| PostgreSQL | 3 LLMs, 2 schedules, 5 brands (with aliases), 10 exclusion words |
| MongoDB | 5 prompts (with tags), ~60 generated responses (30-day spread, search URLs) |

This is enough to exercise keyword stats, brand trends, domain citations, tag filters, and the scheduler UI without running the worker or etcd.

#### How fixtures work

- **Entry point:** [`cmd/fixtures-dev/main.go`](cmd/fixtures-dev/main.go) — invoked only from Make, not a `gego` subcommand
- **Loader:** [`internal/fixtures/`](internal/fixtures/) — reset, YAML load, synthetic response generation
- **YAML files:** [`internal/fixtures/dev/`](internal/fixtures/dev/) — stable IDs for cross-store references (`llms.yaml`, `prompts.yaml`, etc.)
- **Safety guard:** fixtures run only when `GEGO_FIXTURES=dev` is set (the Makefile sets this)

Before loading, fixtures **fully clean** both databases:

- **PostgreSQL:** truncates all application tables (`users`, `llms`, `schedules`, `brands`, …); schema migrations are preserved
- **MongoDB:** drops the configured database and recreates indexes

The API recreates the bootstrap admin on startup if no users exist.

#### Makefile targets

| Target | Fixtures | Use when |
|--------|----------|----------|
| `make dev` | Yes (reset + load) | Full local run with dashboard data |
| `make fixtures-dev` | Yes (reset + load) | Reload data without starting the API |
| `make dev-api` | No | API only; use your own data or load fixtures separately |
| `make dev-worker` | No | Worker only |

#### Editing fixture data

**Static entities** — edit YAML under `internal/fixtures/dev/`:

```yaml
# internal/fixtures/dev/prompts.yaml
- id: prompt-fashion-001
  template: "What are the top luxury fashion brands in 2025?"
  tags: [fashion, luxury]
  enabled: true
```

Use stable `id` values when referencing prompts or LLMs from schedules. Brand aliases must be unique when lowercased (PostgreSQL enforces `LOWER(alias)` uniqueness).

**Responses** — generated in [`internal/fixtures/responses.go`](internal/fixtures/responses.go) (~60 rows, varied timestamps, brand mentions, `search_urls`). Change this file when you need different trend shapes or volume.

After editing fixtures, run:

```bash
make fixtures-dev   # or make dev
```

YAML is embedded at build time via `//go:embed`; no separate install step is required.

#### Environment variables

| Variable | Purpose |
|----------|---------|
| `GEGO_POSTGRES_URI` | PostgreSQL connection string |
| `GEGO_MONGODB_URI` | MongoDB connection string |
| `GEGO_MONGODB_DATABASE` | MongoDB database name (default: `gego`) |
| `GEGO_FIXTURES` | Must be `dev` for fixture loading (set by Makefile) |
| `GEGO_BOOTSTRAP_ADMIN_USERNAME` | Admin username recreated after clean |
| `GEGO_BOOTSTRAP_ADMIN_PASSWORD` | Admin password recreated after clean |

You do not need `gego init` if `.env.dev` provides database URIs and auth variables.

### Running MongoDB for Development

```bash
# Using Docker
docker run -d -p 27017:27017 --name gego-mongo mongo:latest

# Or install MongoDB locally
# https://www.mongodb.com/docs/manual/installation/
```

> **Note:** PostgreSQL is also required for local development. See [Running databases for development](#running-databases-for-development) above.

## What to Contribute

### 🐛 Bug Fixes

Found a bug? Please:
1. Check if it's already reported in Issues
2. If not, open a new issue with:
   - Description of the bug
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment (OS, Go version, etc.)
3. Submit a PR with the fix

### ✨ New Features

Want to add a feature? Please:
1. Open an issue first to discuss the feature
2. Wait for approval from maintainers
3. Implement the feature
4. Add tests if applicable
5. Update documentation
6. Submit a PR

### 📚 Documentation

Documentation improvements are always welcome:
- Fix typos
- Add examples
- Clarify explanations
- Add tutorials

### 🧪 Tests

Help us improve test coverage:
- Add unit tests
- Add integration tests
- Add end-to-end tests

## Areas Needing Contributions

### High Priority

1. **Cassandra Database Support**
   - Implement `internal/db/cassandra/` following the MongoDB pattern
   - Update CLI to support Cassandra configuration

2. **Additional LLM Providers**
   - Google PaLM
   - Cohere
   - Hugging Face Inference API
   - Local models (llama.cpp integration)

3. **Web Dashboard**
   - React/Vue frontend for visualizations
   - REST API for data access
   - Charts and graphs for trends

4. **Export Functionality**
   - CSV export
   - JSON export
   - PDF reports

### Medium Priority

5. **Advanced Brand Extraction**
   - NLP-based extraction
   - Custom regex patterns
   - Brand aliases/variants

6. **Webhook Notifications**
   - Notify when specific brands are mentioned
   - Alert on anomalies
   - Integration with Slack, Discord, etc.

7. **Docker Support**
   - Dockerfile
   - Docker Compose for full stack
   - Kubernetes manifests

8. **Performance Improvements**
   - Caching layer
   - Query optimization
   - Batch processing

### Low Priority

9. **Additional Database Backends**
   - ScyllaDB
   - DynamoDB

10. **More Statistics**
    - Sentiment analysis
    - Response length analysis
    - Time-series forecasting

## Coding Guidelines

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Use meaningful variable names
- Add comments for exported functions

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to create LLM: %w", err)
}

// Bad
if err != nil {
    return err
}
```

### CLI Output

- Use emojis sparingly for visual feedback
- Use tabwriter for formatted output
- Provide helpful error messages
- Show progress for long operations

### Database Operations

Gego's active SQL store is **PostgreSQL**. **SQLite is legacy** (kept only for unmigrated installations). See [`internal/db/README.md`](internal/db/README.md).

When adding or changing SQL-backed features:

- Add PostgreSQL migrations in `internal/db/migrations/postgres/`
- Implement access in `internal/db/postgres/`
- **Do not** update `internal/db/sqlite/` or `internal/db/migrations/sqlite/`

- Always use context.Context
- Handle errors gracefully
- Use transactions when appropriate
- Optimize queries with indexes

### Testing

```go
func TestSomething(t *testing.T) {
    // Setup
    // Execute
    // Assert
    // Cleanup
}
```

## Adding a New LLM Provider

1. Create directory: `internal/llm/providername/`
2. Create file: `providername.go`
3. Implement interface:

```go
package providername

import (
    "context"
    "github.com/AI2HU/gego/internal/llm"
)

type Provider struct {
    apiKey string
    // other fields
}

func New(apiKey string) *Provider {
    return &Provider{apiKey: apiKey}
}

func (p *Provider) Name() string {
    return "providername"
}

func (p *Provider) Generate(ctx context.Context, prompt string, config map[string]interface{}) (*llm.Response, error) {
    // Implementation
}

func (p *Provider) Validate(config map[string]string) error {
    // Validation
}
```

4. Register in `internal/cli/root.go`:

```go
case "providername":
    provider = providername.New(llmConfig.APIKey)
```

5. Add tests
6. Update documentation

## Adding a New CLI Command

1. Create/update file in `internal/cli/`
2. Follow existing patterns
3. Use Cobra framework
4. Add to root command
5. Test thoroughly
6. Document in README

Example:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Short description",
    Long:  "Long description",
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run specific package
go test ./internal/db/...

# Run with coverage
go test -cover ./...
```

### Writing Tests

- Test files: `*_test.go`
- Test functions: `TestXxx(*testing.T)`
- Use table-driven tests when applicable
- Mock external dependencies

### PR Title Format

- `feat: Add new feature`
- `fix: Fix bug in component`
- `docs: Update documentation`
- `test: Add tests for feature`
- `refactor: Refactor code`
- `chore: Update dependencies`

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring

## Testing
How has this been tested?

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests pass
```

## Code of Conduct

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

- Be respectful and inclusive
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards others

## Questions?

- Open an issue for questions
- Join discussions in GitHub Discussions
- Contact maintainers directly for sensitive matters

## Recognition

Contributors will be recognized in:
- README.md (Contributors section)
- Release notes
- Project documentation

Thank you for contributing to Gego! 🎉
