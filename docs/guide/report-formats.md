# Report Formats

agent-a11y supports multiple output formats for different use cases.

## Agent Output (Default)

The default output format is **agent-optimized JSON**, designed for consumption by coding agents and AI assistants. This format includes actionable fix patterns and design system token suggestions.

```bash
# Default: agent-optimized JSON
agent-a11y audit https://example.com

# Explicitly request agent format
agent-a11y audit https://example.com -o report.json

# With design system integration
agent-a11y audit https://example.com --design-system /path/to/design-system/

# With source mapping (maps findings to source files)
agent-a11y audit https://localhost:3000 --source-map ./dist/
```

**Structure:**

```json
{
  "url": "https://example.com",
  "timestamp": "2025-01-15T10:30:00Z",
  "duration": "5.2s",
  "level": "AA",
  "status": "WARN",
  "summary": {
    "total": 5,
    "critical": 0,
    "serious": 2,
    "moderate": 2,
    "minor": 1,
    "fixable": 4,
    "needsToken": 1
  },
  "findings": [
    {
      "finding": {
        "ruleId": "image-alt",
        "description": "Images must have alternate text",
        "impact": "critical",
        "selector": "img.hero-image",
        "html": "<img src=\"hero.jpg\" class=\"hero-image\">"
      },
      "element": {
        "selector": "img.hero-image",
        "xpath": "/html/body/main/section/img",
        "tagName": "img",
        "componentId": "HeroImage",
        "source": {
          "file": "src/components/Hero.tsx",
          "line": 42,
          "component": "Hero",
          "framework": "react"
        }
      },
      "remediation": {
        "summary": "Add descriptive alt text to the image",
        "fixPatterns": [
          {
            "type": "attribute",
            "action": "add",
            "target": "alt",
            "value": "[descriptive text]",
            "example": "<img src=\"hero.jpg\" alt=\"Description of image content\">"
          }
        ],
        "fixConfidence": 0.85
      }
    }
  ]
}
```

**Key Fields:**

| Field | Description |
|-------|-------------|
| `status` | Overall status: `GO` (no critical/serious), `WARN` (serious issues), `NO-GO` (critical issues) |
| `summary.fixable` | Number of issues with available fix patterns |
| `element.source` | Source code location (when `--source-map` is used) |
| `fixPatterns` | Actionable fix instructions with type, action, target, and example |
| `fixConfidence` | Confidence score (0-1) that the fix will resolve the issue |
| `tokenSuggestions` | Design system token recommendations (when `--design-system` is used) |

**Source Location Fields:**

| Field | Description |
|-------|-------------|
| `source.file` | Path to source file (e.g., `src/components/Hero.tsx`) |
| `source.line` | Line number in source file |
| `source.component` | Component name (if detected) |
| `source.framework` | Detected framework (`react`, `vue`, `svelte`, `angular`, `vanilla`) |

**Fix Pattern Types:**

| Type | Description |
|------|-------------|
| `attribute` | Add/modify HTML attribute |
| `aria` | ARIA attribute change |
| `style` | CSS style change |
| `structure` | DOM structure change |
| `content` | Text content change |
| `semantic` | Semantic HTML change |
| `focus` | Focus management |
| `order` | DOM/reading order |

### Human-Readable Output

Use `--human` flag to get human-readable output instead of agent JSON:

```bash
agent-a11y audit https://example.com --human
agent-a11y audit https://example.com --human --format html -o report.html
```

## JSON

Machine-readable format for CI/CD integration and programmatic processing.

```bash
agent-a11y audit https://example.com --format json -o report.json
```

**Structure:**

```json
{
  "url": "https://example.com",
  "wcagVersion": "2.2",
  "wcagLevel": "AA",
  "score": 85,
  "conformant": false,
  "stats": {
    "totalPages": 1,
    "totalFindings": 5,
    "critical": 0,
    "serious": 2,
    "moderate": 2,
    "minor": 1
  },
  "pages": [...],
  "findings": [...]
}
```

## HTML

Interactive HTML report for stakeholders and manual review.

```bash
agent-a11y audit https://example.com --format html -o report.html
```

**Features:**

- Sortable/filterable findings table
- Expandable issue details
- Screenshot annotations
- Summary dashboard

## Markdown

Text-based format for documentation and version control.

```bash
agent-a11y audit https://example.com --format markdown -o report.md
```

## VPAT 2.4

Voluntary Product Accessibility Template for procurement and compliance documentation.

```bash
agent-a11y audit https://example.com --format vpat -o vpat.html
```

**Includes:**

- Product information
- WCAG 2.x conformance table
- Success criteria evaluation
- Remarks and explanations

## WCAG-EM

W3C Website Accessibility Conformance Evaluation Methodology format.

```bash
agent-a11y audit https://example.com --format wcag -o wcag-em.json
```

**Follows:**

- WCAG-EM 1.0 specification
- Evaluation scope definition
- Sample selection documentation
- Audit results

## OpenACR

[OpenACR](https://github.com/GSA/openacr) is a machine-readable format for accessibility conformance reports developed by the GSA.

```bash
agent-a11y audit https://example.com --format openacr -o report.yaml
```

**Features:**

- Machine-readable YAML/JSON format
- Based on VPAT structure
- Supports multiple WCAG versions and catalogs
- Compatible with GSA accessibility reporting tools

**Structure:**

```yaml
title: Example Site Accessibility Conformance Report
product:
  name: https://example.com
catalog: 2.5-edition-wcag-2.2-508-en
author:
  name: agent-a11y
report_date: "2025-01-15"
chapters:
  success_criteria_level_a:
    criteria:
      - num: "1.1.1"
        components:
          - name: web
            adherence:
              level: supports
              notes: All images have appropriate alt text.
```

**Available Catalogs:**

| Catalog ID | Description |
|------------|-------------|
| `2.5-edition-wcag-2.2-508-en` | WCAG 2.2 + Section 508 (default) |
| `2.5-edition-wcag-2.1-508-en` | WCAG 2.1 + Section 508 |
| `2.5-edition-wcag-2.0-508-en` | WCAG 2.0 + Section 508 |

## CSV

Spreadsheet-compatible format for data analysis.

```bash
agent-a11y audit https://example.com --format csv -o findings.csv
```

**Columns:**

- URL, Rule ID, Description, Impact, Level
- Element, Selector, HTML snippet
- Success Criteria, Help URL

## Go Library

```go
import "github.com/plexusone/agent-a11y/report"

result, _ := auditor.AuditPage(ctx, url)

// Using report.Writer for any format
writer := report.NewWriter(report.FormatJSON)
writer.Write(outputFile, result)

// Available formats
report.FormatJSON     // JSON
report.FormatHTML     // HTML
report.FormatMarkdown // Markdown
report.FormatVPAT     // VPAT 2.4
report.FormatWCAG     // WCAG-EM
report.FormatOpenACR  // OpenACR
report.FormatCSV      // CSV
```

### OpenACR with Custom Options

```go
import "github.com/plexusone/agent-a11y/report"

// Generate OpenACR with custom metadata
openACRReport, err := report.GenerateOpenACR(result, report.OpenACROptions{
    ProductName:    "My Application",
    ProductVersion: "1.0.0",
    AuthorName:     "Accessibility Team",
    AuthorEmail:    "a11y@example.com",
    VendorName:     "My Company",
    CatalogID:      "2.5-edition-wcag-2.2-508-en",
})

// Write as YAML
openACRReport.WriteYAML(file)

// Or as JSON
jsonBytes, _ := openACRReport.JSON()
```
