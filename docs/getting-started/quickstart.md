# Quick Start

This guide walks you through running your first accessibility audit.

## Basic Audit

Run an audit on any URL:

```bash
agent-a11y audit https://example.com
```

Output:

```
WCAG 2.2 Level AA: Non-Conformant (Score: 75/100)
Issues: 1 critical, 2 serious, 5 moderate, 3 minor

Critical Issues:
  - [img-alt] Images must have alternate text

Serious Issues:
  - [label] Form elements must have labels
  - [color-contrast] Elements must have sufficient color contrast
```

## Specify WCAG Level and Version

```bash
# Test against WCAG 2.1 Level AAA
agent-a11y audit https://example.com --level AAA --version 2.1
```

## Generate Reports

### HTML Report

```bash
agent-a11y audit https://example.com -o report.html --format html
```

### JSON Report

```bash
agent-a11y audit https://example.com -o report.json --format json
```

### VPAT Report

```bash
agent-a11y audit https://example.com -o vpat.html --format vpat
```

## Site Crawling

Audit an entire website:

```bash
agent-a11y audit https://example.com --crawl --depth 2 --max-pages 50
```

Options:

- `--crawl` - Enable site crawling
- `--depth` - Maximum crawl depth (default: 2)
- `--max-pages` - Maximum pages to audit (default: 50)
- `--delay` - Delay between requests (default: 500ms)

## LLM-Enhanced Evaluation

Reduce false positives with AI evaluation:

```bash
export ANTHROPIC_API_KEY="your-key"

agent-a11y audit https://example.com \
  --llm-provider anthropic \
  --llm-model claude-sonnet-4-20250514
```

The LLM evaluates each finding and provides:

- Confirmation or rejection of the issue
- Reasoning for the decision
- Confidence score

## Using as a Go Library

```go
package main

import (
    "context"
    "fmt"
    "log"

    a11y "github.com/plexusone/agent-a11y"
)

func main() {
    ctx := context.Background()

    // Create auditor with options
    auditor, err := a11y.New(
        a11y.WithLevel(a11y.LevelAA),
        a11y.WithVersion(a11y.Version22),
        a11y.WithHeadless(true),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer auditor.Close()

    // Audit a page
    result, err := auditor.AuditPage(ctx, "https://example.com")
    if err != nil {
        log.Fatal(err)
    }

    // Print summary
    fmt.Println(result.Summary())

    // Check conformance
    if result.Conformant() {
        fmt.Println("Site is conformant!")
    } else {
        fmt.Printf("Found %d critical issues\n", result.Stats.Critical)
    }

    // Generate HTML report
    html, _ := result.HTML()
    os.WriteFile("report.html", html, 0600)
}
```

## Next Steps

- [Learn about different audit types](../guide/audit-types.md)
- [Configure agent-a11y](../guide/configuration.md)
- [Set up LLM integration](../guide/llm-integration.md)
