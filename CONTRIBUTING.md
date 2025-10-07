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
- MongoDB (or Docker to run MongoDB)
- Make (optional but recommended)

### Setup

```bash
# Clone the repository
git clone https://github.com/AI2HU/gego.git
cd gego

# Install dependencies
go mod download

# Build
make build

# Run
./build/gego --help
```

### Running MongoDB for Development

```bash
# Using Docker
docker run -d -p 27017:27017 --name gego-mongo mongo:latest

# Or install MongoDB locally
# https://www.mongodb.com/docs/manual/installation/
```

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
