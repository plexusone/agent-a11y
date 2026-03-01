# agent-a11y

Go accessibility auditing toolkit for WCAG 2.0, 2.1, and 2.2 compliance testing.

## Overview

agent-a11y provides comprehensive accessibility testing capabilities:

- **WCAG Compliance** - Test against WCAG 2.0, 2.1, and 2.2 at levels A, AA, and AAA
- **Multiple Audit Modes** - Single page, site crawling, and user journey testing
- **axe-core Integration** - Industry-standard accessibility rules
- **LLM-as-a-Judge** - AI evaluation to reduce false positives
- **Multiple Output Formats** - JSON, HTML, Markdown, VPAT 2.4, WCAG-EM, CSV
- **Flexible Integration** - CLI, HTTP API, MCP server, Go library

## Quick Example

=== "CLI"

    ```bash
    # Audit a page
    agent-a11y audit https://example.com --level AA --version 2.2

    # Output HTML report
    agent-a11y audit https://example.com -o report.html --format html
    ```

=== "Go Library"

    ```go
    auditor, _ := a11y.New(
        a11y.WithLevel(a11y.LevelAA),
        a11y.WithVersion(a11y.Version22),
    )
    defer auditor.Close()

    result, _ := auditor.AuditPage(ctx, "https://example.com")
    fmt.Println(result.Summary())
    ```

## Features

### Audit Types

| Type | Description | Use Case |
|------|-------------|----------|
| **Single Page** | Audit one URL | Quick checks, CI pipelines |
| **Site Crawl** | Crawl and audit multiple pages | Full site audits |
| **Journey** | Follow user journeys | Dynamic content, SPAs |

### WCAG Coverage

- WCAG 2.0, 2.1, 2.2
- Levels A, AA, AAA
- Specialized tests for keyboard, focus, reflow, and more

### Report Formats

- **JSON** - Machine-readable for CI/CD integration
- **HTML** - Interactive reports for stakeholders
- **VPAT 2.4** - Voluntary Product Accessibility Template
- **WCAG-EM** - W3C evaluation methodology format
- **CSV** - Spreadsheet export

## Getting Started

1. [Install agent-a11y](getting-started/installation.md)
2. [Run your first audit](getting-started/quickstart.md)
3. [Configure for your needs](guide/configuration.md)

## Integration Options

- **CLI** - Command-line tool for scripts and CI/CD
- **HTTP API** - REST API for web services
- **MCP Server** - Integration with AI assistants (Claude, etc.)
- **Go Library** - Embed in your Go applications
- **Multi-Agent Spec** - Use in multi-agent workflows
