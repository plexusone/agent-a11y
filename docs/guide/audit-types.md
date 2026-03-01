# Audit Types

agent-a11y supports three types of accessibility audits.

## Single Page Audit

The simplest audit type - test a single URL.

```bash
agent-a11y audit https://example.com
```

**Use cases:**

- Quick accessibility checks
- CI/CD pipeline validation
- Testing specific pages

**Go library:**

```go
result, err := auditor.AuditPage(ctx, "https://example.com")
```

## Site Crawl Audit

Crawl a website and audit multiple pages automatically.

```bash
agent-a11y audit https://example.com --crawl --depth 2 --max-pages 50
```

**Options:**

| Flag | Description | Default |
|------|-------------|---------|
| `--crawl` | Enable crawling | false |
| `--depth` | Maximum link depth | 2 |
| `--max-pages` | Maximum pages to audit | 50 |
| `--delay` | Delay between requests | 500ms |

**Use cases:**

- Full website accessibility audits
- Pre-launch compliance checks
- Ongoing monitoring

**Go library:**

```go
result, err := auditor.AuditSite(ctx, "https://example.com",
    a11y.WithCrawlDepth(2),
    a11y.WithMaxPages(50),
)
```

### SPA Detection

The crawler automatically detects Single Page Applications and:

- Waits for JavaScript hydration
- Handles dynamic content loading
- Follows client-side navigation

## Journey Audit

Test accessibility along specific user journeys defined in YAML.

```bash
agent-a11y audit https://example.com --journey journey.yaml
```

**Journey file example:**

```yaml
name: Login Flow
steps:
  - name: Visit login page
    action: navigate
    url: /login
    audit: true

  - name: Enter credentials
    action: type
    selector: "#username"
    value: "testuser"

  - name: Submit form
    action: click
    selector: "#submit"
    waitFor: "#dashboard"
    audit: true
```

**Use cases:**

- Testing authenticated pages
- Multi-step form validation
- Dynamic content that requires interaction
- SPA navigation flows

**Go library:**

```go
result, err := auditor.AuditJourney(ctx, "https://example.com", "journey.yaml")
```

## Specialized Tests

Beyond axe-core rules, agent-a11y includes specialized tests:

| Test | WCAG Criterion | Description |
|------|----------------|-------------|
| Keyboard | 2.1.1, 2.1.2 | Keyboard navigation and traps |
| Focus | 2.4.7, 2.4.11 | Focus visibility and order |
| Reflow | 1.4.10 | 400% zoom reflow |
| Target Size | 2.5.8 | Touch target dimensions |
| Text Spacing | 1.4.12 | Text spacing override support |
| Hover | 1.4.13 | Hover content accessibility |
| Flash | 2.3.1 | Flashing content detection |

Enable specialized tests:

```bash
agent-a11y audit https://example.com --specialized
```
