# Accessibility Demo Sites

This directory contains VPAT reports and comparison analyses for well-known accessibility demonstration sites. These sites are designed with intentional "before" (inaccessible) and "after" (accessible) versions to demonstrate accessibility improvements.

## Demo Sites

| Site | Source | Description |
|------|--------|-------------|
| [w3c-bad](./w3c-bad/) | W3C WAI | W3C's official Before-After Demonstration |
| [accesscomputing](./accesscomputing/) | University of Washington | AccessComputing demonstration site |
| [a11yquest-forms](./a11yquest-forms/) | A11yQuest | Accessible forms demonstration |

## Report Formats

Each demo site includes reports in multiple formats:

- **JSON** (`vpat.json`) - Machine-readable, suitable for CI/CD and programmatic access
- **Markdown** (`vpat.md`) - Human-readable, GitHub-friendly
- **PDF** (`vpat.pdf`) - Generated via Pandoc, suitable for formal documentation

## Directory Structure

```
demo-sites/
├── w3c-bad/
│   ├── README.md           # Site-specific documentation
│   ├── before/
│   │   ├── vpat.json       # VPAT in JSON format
│   │   ├── vpat.md         # VPAT in Markdown format
│   │   └── vpat.pdf        # VPAT in PDF format (via Pandoc)
│   ├── after/
│   │   ├── vpat.json
│   │   ├── vpat.md
│   │   └── vpat.pdf
│   └── comparison/
│       ├── comparison.json # Before/after comparison
│       ├── comparison.md
│       └── comparison.pdf
├── accesscomputing/
│   └── ...
└── a11yquest-forms/
    └── ...
```

## Generating Reports

### Prerequisites

- Go 1.21+
- Pandoc (for PDF generation)
- Chrome/Chromium (for browser automation)

### Generate All Reports

```bash
# Generate all demo site reports
make demo-reports

# Or use the script directly
./script/generate-demo-reports.sh
```

### Generate Reports for a Single Site

```bash
# W3C BAD demo
make demo-w3c-bad

# Or manually:
agenta11y audit https://www.w3.org/WAI/demos/bad/before/home.html \
  -f json -o examples/demo-sites/w3c-bad/before/vpat.json

agenta11y audit https://www.w3.org/WAI/demos/bad/before/home.html \
  -f vpat -o examples/demo-sites/w3c-bad/before/vpat.md

pandoc examples/demo-sites/w3c-bad/before/vpat.md \
  -o examples/demo-sites/w3c-bad/before/vpat.pdf \
  --pdf-engine=xelatex
```

### Generate Comparison Reports

```bash
agenta11y compare \
  https://www.w3.org/WAI/demos/bad/before/home.html \
  https://www.w3.org/WAI/demos/bad/after/home.html \
  --name "W3C BAD Demo" \
  -f json -o examples/demo-sites/w3c-bad/comparison/comparison.json
```

## Use Cases

### 1. Learning Accessibility

Compare the before/after VPATs to understand:

- What issues exist in inaccessible sites
- How those issues are remediated
- WCAG success criteria in practice

### 2. Tool Validation

Use these demo sites to validate accessibility testing tools:

- The "before" site should have many findings
- The "after" site should have significantly fewer
- Compare your tool's results with these baseline reports

### 3. Training Material

Use the comparison reports for:

- Developer training sessions
- Accessibility workshops
- Demonstrating remediation impact

### 4. CI/CD Integration

Use the JSON reports to:

- Set baseline metrics
- Validate remediation in CI/CD pipelines
- Track accessibility improvements over time

## Notes

- Reports are generated against live demo sites; results may vary if sites are updated
- PDF generation requires Pandoc with a LaTeX engine (xelatex, pdflatex, or lualatex)
- Some demo sites may require specific browser configurations or have rate limiting
