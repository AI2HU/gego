# Gego - GEO System for your brand, working with all LLMs

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)

Gego is an open-source GEO (Generative Engine Optimization) tracker that schedules prompts across multiple Large Language Models (LLMs) and automatically extracts keywords from their responses. It helps you understand which keywords (brands, products, concepts) appear most frequently, which prompts generate the most mentions.

## Features

- 🤖 **Multi-LLM Support**: Works with OpenAI, Anthropic, Ollama, Google, Perplexity, and custom LLM providers
- 📊 **Hybrid Database**: SQLite for configuration data (LLMs, Schedules) and MongoDB for analytics data (Prompts, Responses)
- ⏰ **Flexible Scheduling**: Cron-based scheduler for automated prompt execution
- 📈 **Comprehensive Analytics**: Track keyword mentions, compare prompts and LLMs, view trends
- 💻 **User-Friendly CLI**: Interactive commands for all operations
- 🔌 **Pluggable Architecture**: Easy to add new LLM providers and database backends
- 🎯 **Automatic Keyword Extraction**: Intelligently extracts keywords from responses (no predefined list needed)
- 📉 **Performance Metrics**: Monitor latency, token usage, and error rates
- 🔄 **Retry Mechanism**: Automatic retry with 30-second delays for failed requests
- 📝 **Configurable Logging**: DEBUG, INFO, WARNING, ERROR levels with file output support
- **Personnas**: create your own personnas for more accurate metrics

## Use Cases

- **SEO/Marketing Research**: Track how brands and keywords are mentioned by AI assistants
- **Competitive Analysis**: Compare keyword visibility across different LLMs
- **Prompt Engineering**: Identify which prompts generate the most keyword mentions
- **Trend Analysis**: Monitor changes in keyword mentions over time

## Installation

### Prerequisites

- Go 1.21 or higher
- MongoDB (for analytics data)
- API keys for LLM providers (OpenAI, Anthropic, etc.)

### Build from Source

```bash
git clone https://github.com/AI2HU/gego.git
cd gego
go build -o gego cmd/gego/main.go
```

### Install

```bash
go install github.com/AI2HU/gego/cmd/gego@latest
```

## Quick Start

### 1. Initialize Configuration

```bash
gego init
```

This interactive wizard will guide you through:
- Database configuration
- Connection testing

Note: Gego automatically extracts keywords from responses - no predefined keyword list needed!

### 2. Add LLM Providers

```bash
gego llm add
```

Example providers:
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude)
- Ollama (Local models)
- Google (Gemini)
- Perplexity (Sonar)

### 3. Create Prompts

```bash
gego prompt add
```

Example prompts:
- "What are the best streaming services for movies?"
- "Which cloud providers offer the best value?"
- "What are popular social media platforms?"

### 4. Set Up Schedules

```bash
gego schedule add
```

Create schedules to run prompts automatically using cron expressions.

### 5. Start the Scheduler

```bash
gego run
```

The scheduler will execute your prompts according to their schedules and collect responses.

## Usage Examples

### View Statistics

```bash
# Top keywords by mentions
gego stats keywords --limit 20

# Statistics for a specific keyword
gego stats keyword Dior
```

### Manage LLMs

```bash
# List all LLMs
gego llm list

# Get LLM details
gego llm get <id>

# Enable/disable LLM
gego llm enable <id>
gego llm disable <id>

# Delete LLM
gego llm delete <id>
```

### Manage Prompts

```bash
# List all prompts
gego prompt list

# Get prompt details
gego prompt get <id>

# Enable/disable prompt
gego prompt enable <id>
gego prompt disable <id>

# Delete prompt
gego prompt delete <id>
```

### Manage Schedules

```bash
# List all schedules
gego schedule list

# Get schedule details
gego schedule get <id>

# Run schedule immediately
gego schedule run <id>

# Enable/disable schedule
gego schedule enable <id>
gego schedule disable <id>

# Delete schedule
gego schedule delete <id>
```

## Configuration

Configuration is stored in `~/.gego/config.yaml`:

```yaml
sql:
  provider: sqlite
  uri: ~/.gego/gego.db

nosql:
  provider: mongodb
  uri: mongodb://localhost:27017
  database: gego
```

**Database Architecture:**
- **SQLite**: Stores LLM configurations and schedules (lightweight, local)
- **MongoDB**: Stores prompts and responses with analytics (scalable, indexed)

Note: Keywords are automatically extracted from LLM responses. No predefined list needed!

## Logging

Gego includes a comprehensive logging system that allows you to control log levels and output destinations for better monitoring and debugging.

### Log Levels

- **DEBUG**: Detailed information for debugging (most verbose)
- **INFO**: General information about application flow (default)
- **WARNING**: Warning messages for potential issues
- **ERROR**: Error messages for failures (least verbose)

### Command Line Options

#### `--log-level`
Set the minimum log level to display:

```bash
# Show only errors
gego run --log-level ERROR

# Show warnings and errors
gego run --log-level WARNING

# Show info, warnings, and errors (default)
gego run --log-level INFO

# Show all messages including debug
gego run --log-level DEBUG
```

#### `--log-file`
Specify a file to write logs to instead of stdout:

```bash
# Log to a file
gego run --log-file /var/log/gego.log

# Log to file with debug level
gego run --log-level DEBUG --log-file /var/log/gego-debug.log
```

### Usage Examples

#### Production Deployment
```bash
# Log only errors to a file for production
gego run --log-level ERROR --log-file /var/log/gego/error.log
```

#### Development/Debugging
```bash
# Show all debug information on stdout
gego run --log-level DEBUG
```

#### Monitoring
```bash
# Log info and above to a file for monitoring
gego run --log-level INFO --log-file /var/log/gego/app.log
```

### Log Format

Logs are formatted with timestamps and level prefixes:

```
[INFO] 2024-01-15 10:30:45 Logging initialized - Level: INFO
[INFO] 2024-01-15 10:30:45 🚀 Starting Gego Scheduler
[DEBUG] 2024-01-15 10:30:45 Getting prompt: prompt-123
[INFO] 2024-01-15 10:30:45 Found 3 prompts and 2 enabled LLMs
[WARNING] 2024-01-15 10:30:46 ❌ Attempt 1/3 failed: connection timeout
[INFO] 2024-01-15 10:30:46 ⏳ Waiting 30s before retry attempt 2...
[ERROR] 2024-01-15 10:30:47 💥 All 3 attempts failed. Last error: service unavailable
```

### Retry Mechanism

Gego automatically retries failed prompt requests with the following behavior:

- **Maximum Retries**: 3 attempts total
- **Retry Delay**: 30 seconds between each attempt
- **Automatic Recovery**: Handles temporary network issues and API rate limits
- **Detailed Logging**: Comprehensive retry attempt tracking

Example retry log:
```
[WARNING] ❌ Attempt 1/3 failed for prompt 'What are the best streaming services...' with LLM 'GPT-4': connection timeout
[INFO] ⏳ Waiting 30s before retry attempt 2...
[INFO] ✅ Prompt execution succeeded on attempt 2 after 1 previous failures
```

### Integration with System Logging

For production deployments, you can integrate with system logging:

```bash
# Use systemd journal
gego run --log-level INFO | systemd-cat -t gego

# Use syslog
gego run --log-level WARNING --log-file /dev/log
```

### Monitoring Commands

```bash
# Monitor retry attempts
gego run --log-level DEBUG | grep "Attempt"

# Monitor retry failures
gego run --log-level WARNING | grep "❌"

# Monitor successful retries
gego run --log-level INFO | grep "✅"
```

## Architecture

### Hybrid Database Schema

Gego uses a hybrid database architecture optimized for different data types:

**SQLite (Configuration Data):**
- `llms`: LLM provider configurations (id, name, provider, model, api_key, base_url, config, enabled, timestamps)
- `schedules`: Execution schedules (id, name, prompt_ids, llm_ids, cron_expr, enabled, last_run, next_run, timestamps)

**MongoDB (Analytics Data):**
- `prompts`: Prompt templates (id, template, tags, enabled, timestamps)
- `responses`: LLM responses with metadata (id, prompt_id, llm_id, response_text, tokens_used, latency_ms, timestamps)

**Key Indexes:**
- **SQLite**: `idx_llms_provider`, `idx_llms_enabled`, `idx_schedules_enabled`, `idx_schedules_next_run`
- **MongoDB**: `(prompt_id, created_at)`, `(created_at)` for responses

### Components

```
┌─────────────────┐
│   CLI (Cobra)   │
└────────┬────────┘
         │
    ┌────┴────┐
    │  Core   │
    └────┬────┘
         │
    ┌────┴─────────────────────┐
    │                          │
┌───┴───┐              ┌───────┴────────┐
│Hybrid │              │   LLM Registry │
│  DB   │              │                │
└───┬───┘              └───────┬────────┘
    │                          │
┌───┴────┐            ┌────────┴─────────┐
│SQLite  │            │ OpenAI│Anthropic │
│MongoDB │            │ Ollama│Custom... │
└────────┘            └──────────────────┘
         │                    │
         └─────┬──────────────┘
               │
        ┌──────┴────────┐
        │   Scheduler   │
        └───────────────┘
```

## Adding Custom LLM Providers

Implement the `llm.Provider` interface:

```go
type Provider interface {
    Name() string
    Generate(ctx context.Context, prompt string, config map[string]interface{}) (*Response, error)
    Validate(config map[string]string) error
}
```

Register your provider in the registry:

```go
registry.Register(myProvider)
```

## Performance Optimization

Gego uses several strategies for optimal performance:

1. **Hybrid Database**: SQLite for fast configuration queries, MongoDB for scalable analytics
2. **On-demand Statistics**: Keyword statistics are calculated dynamically from response data
3. **Indexed Queries**: All common queries are backed by database indexes
4. **Concurrent Execution**: Prompts are executed in parallel across LLMs
5. **Caching**: Keyword extraction patterns are compiled once and reused

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [ ] Persona embedding to simulate Chat version of models
- [ ] System prompt to simulate Chat version of models
- [ ] Schedules cost forecast
- [ ] Prompts batches to optimize costs
- [ ] Prompts threading per provider for speed
- [ ] Additional NoSQL database support (Cassandra, etc.)
- [ ] Web dashboard for visualizations (another repo)
- [ ] Export statistics to CSV/JSON
- [ ] Webhook notifications
- [ ] Custom keyword extraction rules and patterns
- [ ] Time-series trend analysis
- [ ] Multi-tenancy support
- [ ] Docker support

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) for analytics database
- [SQLite3](https://github.com/mattn/go-sqlite3) for configuration database
- [Cron](https://github.com/robfig/cron) for scheduling

## Support

- 📧 Email: jonathan@blocs.fr
- 🐛 Issues: [GitHub Issues](https://github.com/AI2HU/gego/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/AI2HU/gego/discussions)

---

Made with ❤️ for the open-source community
