# CLI Reference

## Commands

### agent-a11y audit

Run an accessibility audit.

```bash
agent-a11y audit <url> [flags]
```

**Arguments:**

- `<url>` - URL to audit (required)

**Flags:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--level` | | WCAG level (A, AA, AAA) | AA |
| `--version` | | WCAG version (2.0, 2.1, 2.2) | 2.2 |
| `--format` | `-f` | Output format | json |
| `--output` | `-o` | Output file path | stdout |
| `--crawl` | | Enable site crawling | false |
| `--depth` | | Crawl depth | 2 |
| `--max-pages` | | Maximum pages | 50 |
| `--delay` | | Request delay | 500ms |
| `--journey` | | Journey file path | |
| `--specialized` | | Run specialized tests | false |
| `--llm-provider` | | LLM provider | |
| `--llm-model` | | LLM model | |
| `--headless` | | Headless browser | true |
| `--timeout` | | Browser timeout | 30s |

**Examples:**

```bash
# Basic audit
agent-a11y audit https://example.com

# WCAG 2.1 Level AAA
agent-a11y audit https://example.com --level AAA --version 2.1

# Site crawl with HTML report
agent-a11y audit https://example.com --crawl -o report.html -f html

# With LLM evaluation
agent-a11y audit https://example.com --llm-provider anthropic --llm-model claude-sonnet-4-20250514
```

### agent-a11y serve

Start the HTTP API server.

```bash
agent-a11y serve [flags]
```

**Flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--port` | Server port | 8080 |
| `--host` | Server host | localhost |

### agent-a11y mcp serve

Start the MCP server for AI assistant integration.

```bash
agent-a11y mcp serve
```

Communicates via stdio for MCP protocol.

### agent-a11y compare

Compare accessibility between two URLs.

```bash
agent-a11y compare <before-url> <after-url> [flags]
```

Shows:

- Issues fixed
- Issues introduced
- Issues unchanged

### agent-a11y demo

Run demo audits on sample sites.

```bash
# List available demos
agent-a11y demo list

# Run a specific demo
agent-a11y demo run w3c-bad

# Generate a demo site
agent-a11y demo generate my-demo
```

### agent-a11y version

Show version information.

```bash
agent-a11y version
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path |
| `--verbose` | Verbose output |
| `--help` | Show help |

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success (conformant) |
| 1 | Non-conformant (issues found) |
| 2 | Error (audit failed) |
