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
result, _ := auditor.AuditPage(ctx, url)

// JSON
jsonBytes, _ := result.JSON()

// HTML
htmlBytes, _ := result.HTML()

// Markdown
mdBytes, _ := result.Markdown()

// VPAT
vpatBytes, _ := result.VPAT()

// WCAG-EM
wcagBytes, _ := result.WCAG()
```
