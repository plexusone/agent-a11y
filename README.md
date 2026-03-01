# agent-a11y

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

Go accessibility auditing toolkit for WCAG 2.0, 2.1, and 2.2 compliance testing.

## Features

- ✅ **WCAG Compliance Testing** - Supports WCAG 2.0, 2.1, and 2.2 at levels A, AA, and AAA
- 🔍 **Multiple Audit Modes** - Single page, site crawling, and user journey testing
- 🪓 **axe-core Integration** - Industry-standard accessibility testing via axe-core
- 🤖 **LLM-as-a-Judge** - Optional AI evaluation to reduce false positives
- 📊 **Report Formats** - JSON, HTML, Markdown, VPAT 2.4, WCAG-EM, CSV
- 🔌 **MCP Server** - Model Context Protocol integration for AI assistants
- 🌐 **HTTP API** - REST API for programmatic access
- 🤝 **Multi-Agent Spec** - Integration with multi-agent workflows

## Installation

```bash
go install github.com/plexusone/agent-a11y/cmd/agent-a11y@latest
```

### Prerequisites

- Go 1.25+
- Chrome/Chromium browser (for browser-based testing)

## Quick Start

### CLI Usage

```bash
# Audit a single page
agent-a11y audit https://example.com

# Audit with site crawling
agent-a11y audit https://example.com --crawl --depth 2

# Specify WCAG level and version
agent-a11y audit https://example.com --level AA --version 2.2

# Output to file with format
agent-a11y audit https://example.com -o report.html --format html

# Enable LLM evaluation (reduces false positives)
agent-a11y audit https://example.com --llm-provider anthropic --llm-model claude-sonnet-4-20250514
```

### Go Library

```go
package main

import (
    "context"
    "fmt"
    "log"

    a11y "github.com/plexusone/agent-a11y"
)

func main() {
    // Create auditor
    auditor, err := a11y.New(
        a11y.WithLevel(a11y.LevelAA),
        a11y.WithVersion(a11y.Version22),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer auditor.Close()

    // Audit a page
    result, err := auditor.AuditPage(context.Background(), "https://example.com")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Summary())
    // WCAG 2.2 Level AA: Conformant (Score: 95/100, Issues: 0 critical, 0 serious, 1 moderate, 2 minor)
}
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `audit <url>` | Run accessibility audit on a URL |
| `serve` | Start HTTP API server |
| `mcp serve` | Start MCP server for AI assistants |
| `compare <before> <after>` | Compare accessibility between two URLs |
| `demo list` | List available demo sites |
| `demo run [name]` | Run demo audit on sample sites |
| `version` | Show version information |

## Output Formats

| Format | Flag | Description |
|--------|------|-------------|
| JSON | `--format json` | Machine-readable JSON output |
| HTML | `--format html` | Interactive HTML report |
| Markdown | `--format markdown` | Markdown report |
| VPAT | `--format vpat` | VPAT 2.4 accessibility conformance report |
| WCAG-EM | `--format wcag` | WCAG-EM evaluation report |
| CSV | `--format csv` | CSV export of findings |

## Configuration

Create a config file at `~/.agent-a11y.yaml` or specify with `--config`:

```yaml
wcag:
  level: AA
  version: "2.2"

browser:
  headless: true
  timeout: 30s

llm:
  enabled: true
  provider: anthropic
  model: claude-sonnet-4-20250514

crawl:
  depth: 2
  maxPages: 50
  delay: 500ms
```

## LLM Providers

agent-a11y supports multiple LLM providers for AI-assisted evaluation:

| Provider | Environment Variable |
|----------|---------------------|
| Anthropic | `ANTHROPIC_API_KEY` |
| OpenAI | `OPENAI_API_KEY` |
| Google | `GOOGLE_API_KEY` |
| xAI | `XAI_API_KEY` |
| Ollama | (local, no key required) |

## MCP Server

Run as an MCP server for integration with AI assistants:

```bash
agent-a11y mcp serve
```

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "a11y": {
      "command": "agent-a11y",
      "args": ["mcp", "serve"]
    }
  }
}
```

## HTTP API

Start the API server:

```bash
agent-a11y serve --port 8080
```

Endpoints:

- `POST /audit` - Run an accessibility audit
- `GET /health` - Health check

## Multi-Agent Integration

agent-a11y integrates with the [multi-agent-spec](https://github.com/plexusone/multi-agent-spec) for use in multi-agent workflows:

```go
result, _ := auditor.AuditPage(ctx, url)

// Convert to AgentResult for multi-agent workflows
agentResult := result.AgentResult()

// Get narrative for reports
narrative := result.Narrative()
```

## Examples

See the [examples](./examples) directory:

- [basic](./examples/basic) - Simple single-page audit
- [multi-agent](./examples/multi-agent) - Multi-agent workflow integration

## Development

```bash
# Run tests
go test -v ./...

# Run linter
golangci-lint run

# Build CLI
go build -o agent-a11y ./cmd/agent-a11y
```

## License

MIT

 [go-ci-svg]: https://github.com/plexusone/agent-a11y/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/agent-a11y/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/agent-a11y/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/agent-a11y/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/agent-a11y/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/agent-a11y/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/agent-a11y
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/agent-a11y
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/agent-a11y
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/agent-a11y
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=grokify%2Fcoreforge
 [loc-svg]: https://tokei.rs/b1/github/plexusone/agent-a11y
 [repo-url]: https://github.com/plexusone/agent-a11y
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/agent-a11y/blob/master/LICENSE
