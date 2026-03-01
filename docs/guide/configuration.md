# Configuration

agent-a11y can be configured via command-line flags, environment variables, or a configuration file.

## Configuration File

Create `~/.agent-a11y.yaml` or specify with `--config`:

```yaml
# WCAG settings
wcag:
  level: AA          # A, AA, or AAA
  version: "2.2"     # 2.0, 2.1, or 2.2

# Browser settings
browser:
  headless: true
  timeout: 30s

# LLM settings (optional)
llm:
  enabled: true
  provider: anthropic
  model: claude-sonnet-4-20250514
  # API key from environment: ANTHROPIC_API_KEY

# Crawling settings
crawl:
  depth: 2
  maxPages: 50
  delay: 500ms

# Output settings
output:
  format: json       # json, html, markdown, vpat, wcag, csv
  verbose: false
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENAI_API_KEY` | OpenAI API key |
| `GOOGLE_API_KEY` | Google AI API key |
| `XAI_API_KEY` | xAI API key |
| `VIBIUM_HEADLESS` | Run browser headless (0/1) |
| `VIBIUM_DEBUG` | Enable debug logging (0/1) |

## CLI Flags

### Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Path to config file |
| `--verbose` | Enable verbose output |
| `--headless` | Run browser in headless mode |
| `--timeout` | Browser timeout |

### Audit Flags

| Flag | Description |
|------|-------------|
| `--level` | WCAG level (A, AA, AAA) |
| `--version` | WCAG version (2.0, 2.1, 2.2) |
| `--format` | Output format |
| `-o, --output` | Output file path |
| `--crawl` | Enable site crawling |
| `--depth` | Crawl depth |
| `--max-pages` | Maximum pages to crawl |
| `--delay` | Delay between requests |
| `--journey` | Journey definition file |
| `--specialized` | Run specialized tests |

### LLM Flags

| Flag | Description |
|------|-------------|
| `--llm-provider` | LLM provider (anthropic, openai, google, xai, ollama) |
| `--llm-model` | Model name |
| `--llm-api-key` | API key (or use environment variable) |

## Go Library Configuration

```go
auditor, err := a11y.New(
    // WCAG settings
    a11y.WithLevel(a11y.LevelAA),
    a11y.WithVersion(a11y.Version22),

    // Browser settings
    a11y.WithHeadless(true),
    a11y.WithTimeout(30 * time.Second),

    // LLM settings
    a11y.WithLLM("anthropic", "claude-sonnet-4-20250514", os.Getenv("ANTHROPIC_API_KEY")),

    // Logging
    a11y.WithLogger(slog.Default()),
)
```
