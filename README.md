# Gego - GEO System for your brand, working with all LLMs

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)

Gego is an open-source GEO (Generative Engine Optimization) tracker that schedules prompts across multiple Large Language Models (LLMs) and automatically extracts keywords from their responses. It helps you understand which keywords (brands, products, concepts) appear most frequently, which prompts generate the most mentions.

## Features

- 🤖 **Multi-LLM Support**: Works with OpenAI, Anthropic, Ollama, and custom LLM providers
- 📊 **NoSQL Database**: Agnostic design supporting MongoDB and Cassandra with optimized analytics
- ⏰ **Flexible Scheduling**: Cron-based scheduler for automated prompt execution
- 📈 **Comprehensive Analytics**: Track keyword mentions, compare prompts and LLMs, view trends
- 💻 **User-Friendly CLI**: Interactive commands for all operations
- 🔌 **Pluggable Architecture**: Easy to add new LLM providers and database backends
- 🎯 **Automatic Keyword Extraction**: Intelligently extracts keywords from responses (no predefined list needed)
- 📉 **Performance Metrics**: Monitor latency, token usage, and error rates
- **Personnas**: create your own personnas for more accurate metrics

## Use Cases

- **SEO/Marketing Research**: Track how brands and keywords are mentioned by AI assistants
- **Competitive Analysis**: Compare keyword visibility across different LLMs
- **Prompt Engineering**: Identify which prompts generate the most keyword mentions
- **Trend Analysis**: Monitor changes in keyword mentions over time

## Installation

### Prerequisites

- Go 1.21 or higher
- MongoDB (or Cassandra for production use)
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
database:
  provider: mongodb
  uri: mongodb://localhost:27017
  database: gego
```

Note: Keywords are automatically extracted from LLM responses. No predefined list needed!

## Architecture

### Database Schema (MongoDB)

Gego uses an optimized NoSQL schema for fast analytics:

**Collections:**
- `llms`: LLM provider configurations
- `prompts`: Prompt templates
- `schedules`: Execution schedules
- `responses`: LLM responses with extracted keywords
- `keyword_stats`: Pre-aggregated keyword statistics
- `prompt_stats`: Pre-aggregated prompt statistics
- `llm_stats`: Pre-aggregated LLM statistics

**Key Indexes:**
- `responses`: `(keywords.keyword, created_at)`, `(prompt_id, created_at)`, `(llm_id, created_at)`
- `keyword_stats`: `(total_mentions)`, `(last_seen)`

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
│  DB   │              │   LLM Registry │
│Interface             │                │
└───┬───┘              └───────┬────────┘
    │                          │
┌───┴────┐            ┌────────┴─────────┐
│MongoDB │            │ OpenAI│Anthropic │
│Cassandra            │ Ollama│Custom... │
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

1. **Pre-aggregated Statistics**: Statistics are updated asynchronously when responses are created
2. **Indexed Queries**: All common queries are backed by MongoDB indexes
3. **Concurrent Execution**: Prompts are executed in parallel across LLMs
4. **Caching**: Keyword extraction patterns are compiled once and reused

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
- [ ] Cassandra database support
- [ ] Web dashboard for visualizations
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
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) for database
- [Cron](https://github.com/robfig/cron) for scheduling

## Support

- 📧 Email: jonathan@blocs.fr
- 🐛 Issues: [GitHub Issues](https://github.com/AI2HU/gego/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/AI2HU/gego/discussions)

---

Made with ❤️ for the open-source community
