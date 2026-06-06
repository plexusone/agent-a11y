# Report Formats

agent-a11y supports multiple output formats for different use cases.

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
