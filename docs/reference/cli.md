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
| `--human` | | Output human-readable format | false |
| `--design-system` | | Path to design system spec | |
| `--source-map` | | Path to source map directory | |
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
# Basic audit (agent-optimized JSON output)
agent-a11y audit https://example.com

# Human-readable output
agent-a11y audit https://example.com --human

# With design system token suggestions
agent-a11y audit https://example.com --design-system ./design-system/

# With source mapping (maps findings to source files)
agent-a11y audit https://localhost:3000 --source-map ./dist/

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

Compare accessibility between two URLs (VPAT comparison).

```bash
agent-a11y compare <before-url> <after-url> [flags]
```

**Flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--name` | Comparison name | URL-based |
| `--from-files` | Compare from existing JSON files | false |
| `--format` | Output format (json, html, markdown, vpat) | markdown |
| `--output` | Output file path | stdout |

Shows:

- Issues fixed
- Issues introduced
- Issues unchanged

### agent-a11y validate

Validate accessibility fixes against a baseline. Enables the autonomous fix loop.

```bash
agent-a11y validate <url> [flags]
```

**Flags:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--baseline` | `-b` | Baseline audit file to compare against | |
| `--expect-fixed` | | Rule IDs expected to be fixed (comma-separated) | |
| `--fail-on-regression` | | Exit with code 1 if regressions detected | false |
| `--save-baseline` | | Save this audit as baseline (no comparison) | false |
| `--include-results` | | Include full before/after results in delta | false |
| `--output` | `-o` | Output file path | stdout |

**Examples:**

```bash
# Create baseline from first audit
agent-a11y validate https://example.com --save-baseline -o baseline.json

# Validate after fixes
agent-a11y validate https://example.com --baseline baseline.json

# Check specific rules were fixed
agent-a11y validate https://example.com --baseline baseline.json \
  --expect-fixed color-contrast,image-alt

# CI mode: fail on any regressions
agent-a11y validate https://example.com --baseline baseline.json \
  --fail-on-regression
```

**Output:**

The validate command outputs a `ValidationDelta` JSON with:

- `fixed`: Issues that were resolved
- `remaining`: Issues still present
- `regressions`: New issues introduced
- `status`: FIXED, IMPROVED, NO_CHANGE, REGRESSED, or MIXED
- `improvement`: Percentage change

### agent-a11y config

Manage project-specific fix configuration.

```bash
agent-a11y config <subcommand>
```

**Subcommands:**

#### config show

Show resolved configuration for the current project.

```bash
agent-a11y config show
```

Displays:

- Detected language and framework
- Component library (MUI, Chakra, shadcn, etc.)
- Matched project configuration
- Config file location

#### config init

Create a default configuration file.

```bash
agent-a11y config init
```

Creates `~/.plexusone/a11y/fixes.yaml` with example configuration.

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
