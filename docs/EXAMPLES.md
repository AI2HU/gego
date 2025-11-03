# Gego Examples

This document provides practical examples of using Gego for various use cases.

## Example 1: Tracking Streaming Services

### Setup

```bash
# Initialize
gego init
main
# Add OpenAI
gego llm add
# Name: OpenAI GPT-4
# Provider: openai
# Model: gpt-4
# API Key: sk-...

# Create prompts
gego prompt add
# Name: Best Streaming Services
# Template: What are the top 5 streaming services for watching movies and TV shows?

gego prompt add
# Name: Streaming Recommendations
# Template: Which streaming platform would you recommend for someone who loves documentaries?

# Create schedule
gego schedule add
# Name: Daily Streaming Check
# Select prompts: all
# Select LLMs: all
# Cron: 0 9 * * * (every day at 9am)

# Start scheduler
gego run
```

### View Results

```bash
# Top brands
gego stats brands --limit 10

# Specific brand
gego stats brand Netflix

# Compare prompts
gego stats compare prompts
```

## Example 2: Cloud Provider Comparison

### Setup

```bash
# Add multiple LLMs for comparison
gego llm add  # OpenAI GPT-4
gego llm add  # Anthropic Claude
gego llm add  # Ollama llama2

# Create cloud-focused prompts
gego prompt add
# Template: What are the best cloud providers for startups?

gego prompt add
# Template: Compare AWS, Google Cloud, and Azure for machine learning workloads.

# Create schedule to run every 6 hours
gego schedule add
# Cron: 0 */6 * * *
```

### Analyze

```bash
# Which provider is mentioned most?
gego stats brands --limit 5

# Which prompt generates most mentions?
gego stats compare prompts

# Compare LLMs
gego stats compare llms
```

## Example 3: Social Media Tracking

### Custom Brand List

Edit `~/.gego/config.yaml`:

```yaml
brands:
  - Twitter
  - X
  - Facebook
  - Instagram
  - TikTok
  - LinkedIn
  - Reddit
  - Pinterest
  - Snapchat
  - YouTube
```

### Create Prompts

```bash
gego prompt add
# Name: Social Media for Business
# Template: What social media platforms should a small business focus on?

gego prompt add
# Name: Social Media Trends
# Template: What are the most popular social media platforms in 2024?
```

### Run and Analyze

```bash
# Run immediately
gego schedule run <schedule-id>

# View brand statistics
gego stats brand Twitter
gego stats brand TikTok

# Compare
gego stats compare prompts
```

## Example 4: E-commerce Platform Research

### Setup

```bash
# Add e-commerce brands to config
# Edit ~/.gego/config.yaml and add:
# - Amazon
# - Shopify
# - WooCommerce
# - Etsy
# - eBay

# Create prompts
gego prompt add
# Name: Best E-commerce Platform
# Template: What is the best e-commerce platform for a new online store?

gego prompt add
# Name: E-commerce for Artists
# Template: Where should an artist sell their work online?

# Schedule hourly checks
gego schedule add
# Cron: 0 * * * *
```

### Monitor

```bash
# Start daemon
gego run

# In another terminal, monitor results
watch -n 60 'gego stats brands --limit 5'
```

## Example 5: Comparing LLM Biases

### Goal
Understand if different LLMs have different brand preferences.

```bash
# Add multiple LLMs
gego llm add  # OpenAI GPT-4
gego llm add  # OpenAI GPT-3.5-turbo
gego llm add  # Anthropic Claude Sonnet
gego llm add  # Anthropic Claude Opus

# Create neutral prompts
gego prompt add
# Name: Phone Recommendation
# Template: What smartphone would you recommend?

gego prompt add
# Name: Laptop Recommendation
# Template: What laptop is best for programming?

# Run same prompt across all LLMs
gego schedule add
# Select all prompts and all LLMs
# Cron: 0 12 * * * (daily at noon)
```

### Analysis

```bash
# Compare LLMs
gego stats compare llms

# Check individual LLM stats
gego llm list  # Get IDs
gego stats llm <openai-gpt4-id>
gego stats llm <claude-sonnet-id>

# See which LLM mentions Apple most
gego stats brand Apple
```

## Example 6: Local LLM with Ollama

### Setup Ollama

```bash
# Install Ollama
curl https://ollama.ai/install.sh | sh

# Pull a model
ollama pull llama2
```

### Add to Gego

```bash
gego llm add
# Name: Llama2 Local
# Provider: ollama
# Model: llama2
# Base URL: http://localhost:11434

# Create prompts and schedules as usual
gego prompt add
gego schedule add

# Run
gego run
```

## Example 7: Continuous Monitoring

### Setup Background Process

```bash
# Create a systemd service (Linux)
cat > /etc/systemd/system/gego.service <<EOF
[Unit]
Description=Gego GEO Tracker
After=network.target mongod.service

[Service]
Type=simple
User=youruser
ExecStart=/usr/local/bin/gego run
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable gego
sudo systemctl start gego
```

### Monitor Logs

```bash
# View logs
sudo journalctl -u gego -f

# Check stats periodically
gego stats brands
```

## Example 8: Export and Analysis

### View Raw Data

```bash
# Connect to MongoDB
mongosh

use gego

# View recent responses
db.responses.find().sort({created_at: -1}).limit(10)

# Aggregate data
db.brand_stats.aggregate([
  { $sort: { total_mentions: -1 } },
  { $limit: 10 }
])
```

### Export to JSON

```bash
# Export brand stats
mongoexport --db=gego --collection=brand_stats --out=brand_stats.json

# Export responses
mongoexport --db=gego --collection=responses --out=responses.json
```

## Tips and Tricks

### 1. Test Prompts Before Scheduling

```bash
# Run a schedule immediately
gego schedule run <schedule-id>
```

### 2. Disable LLMs Temporarily

```bash
# If an LLM is expensive or rate-limited
gego llm disable <llm-id>

# Enable later
gego llm enable <llm-id>
```

### 3. Update Schedules

```bash
# Disable schedule
gego schedule disable <schedule-id>

# Delete and recreate
gego schedule delete <schedule-id>
gego schedule add
```

### 4. Monitor Costs

Check token usage:

```bash
gego stats llm <llm-id>
# Shows avg_tokens which you can multiply by API pricing
```

### 5. Focus on Specific Brands

Edit your config to only track brands you care about:

```yaml
brands:
  - Netflix
  - Disney+
  - Hulu
```

This reduces noise in your statistics.
