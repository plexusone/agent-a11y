# Go Library Reference

## Installation

```bash
go get github.com/plexusone/agent-a11y@latest
```

## Quick Start

```go
import a11y "github.com/plexusone/agent-a11y"

auditor, _ := a11y.New()
defer auditor.Close()

result, _ := auditor.AuditPage(ctx, "https://example.com")
fmt.Println(result.Summary())
```

## Auditor

### Creating an Auditor

```go
auditor, err := a11y.New(
    a11y.WithLevel(a11y.LevelAA),
    a11y.WithVersion(a11y.Version22),
    a11y.WithHeadless(true),
    a11y.WithTimeout(30 * time.Second),
    a11y.WithLLM("anthropic", "claude-sonnet-4-20250514", apiKey),
    a11y.WithLogger(slog.Default()),
)
if err != nil {
    log.Fatal(err)
}
defer auditor.Close()
```

### Options

| Option | Description |
|--------|-------------|
| `WithLevel(level)` | WCAG level (LevelA, LevelAA, LevelAAA) |
| `WithVersion(version)` | WCAG version (Version20, Version21, Version22) |
| `WithHeadless(bool)` | Run browser headless |
| `WithTimeout(duration)` | Browser timeout |
| `WithLLM(provider, model, key)` | Enable LLM evaluation |
| `WithLogger(logger)` | Custom slog logger |

### Audit Methods

```go
// Single page
result, err := auditor.AuditPage(ctx, "https://example.com")

// Site crawl
result, err := auditor.AuditSite(ctx, "https://example.com",
    a11y.WithCrawlDepth(2),
    a11y.WithMaxPages(50),
    a11y.WithCrawlDelay(500 * time.Millisecond),
)

// Journey
result, err := auditor.AuditJourney(ctx, "https://example.com", "journey.yaml")
```

## Result

### Properties

```go
result.URL           // Audited URL
result.Score         // 0-100 score
result.Level         // Target level
result.Version       // WCAG version
result.Findings      // []Finding
result.Pages         // []PageResult (for site audits)
result.Stats         // Stats summary
```

### Methods

```go
// Check conformance
if result.Conformant() {
    // Passed
}

// Get summary
fmt.Println(result.Summary())

// Filter findings
criticalIssues := result.CriticalFindings()
seriousIssues := result.SeriousFindings()
levelAIssues := result.FindingsByLevel(a11y.LevelA)
contrastIssues := result.FindingsByCriterion("1.4.3")

// Generate reports
jsonBytes, _ := result.JSON()
htmlBytes, _ := result.HTML()
mdBytes, _ := result.Markdown()
vpatBytes, _ := result.VPAT()
```

## Multi-Agent Integration

```go
// Convert to AgentResult for multi-agent workflows
agentResult := result.AgentResult()

// Get narrative section for reports
narrative := result.Narrative()

// Get team section for TeamReport
teamSection := result.TeamSection()
```

## Finding

```go
type Finding struct {
    ID              string
    RuleID          string
    Description     string
    Help            string
    SuccessCriteria []string
    Level           string
    Impact          string
    Element         string
    Selector        string
    HTML            string
    PageURL         string
    LLMConfirmed    *bool
    LLMReasoning    string
}
```

## Stats

```go
type Stats struct {
    TotalPages    int
    TotalFindings int
    Critical      int
    Serious       int
    Moderate      int
    Minor         int
    LevelA        int
    LevelAA       int
    LevelAAA      int
}
```

## Constants

```go
// Levels
a11y.LevelA
a11y.LevelAA
a11y.LevelAAA

// Versions
a11y.Version20
a11y.Version21
a11y.Version22
```
