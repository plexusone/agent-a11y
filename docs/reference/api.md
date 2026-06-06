# HTTP API Reference

## Starting the Server

```bash
agent-a11y serve --port 8080
```

## Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/audits` | Create a new audit job |
| GET | `/api/v1/audits` | List all audit jobs |
| GET | `/api/v1/audits/{id}` | Get audit job status and results |
| DELETE | `/api/v1/audits/{id}` | Cancel an audit job |
| GET | `/api/v1/audits/{id}/report` | Get audit report |
| GET | `/api/v1/audits/{id}/openacr` | Get OpenACR report |
| GET | `/api/v1/health` | Health check |

## Endpoint Details

### POST /api/v1/audits

Create a new audit job.

**Request:**

```json
{
  "url": "https://example.com",
  "config": {
    "level": "AA",
    "version": "2.2"
  }
}
```

**Response:**

```json
{
  "id": "audit-1234567890",
  "status": "pending"
}
```

### GET /api/v1/audits/{id}

Get audit job status and results.

**Response:**

```json
{
  "id": "audit-1234567890",
  "status": "completed",
  "result": {
    "targetUrl": "https://example.com",
    "wcagVersion": "2.2",
    "wcagLevel": "AA",
    "stats": {
      "totalPages": 1,
      "totalFindings": 5
    }
  }
}
```

### GET /api/v1/audits/{id}/report

Get audit report in specified format.

**Query Parameters:**

| Parameter | Default | Description |
|-----------|---------|-------------|
| `format` | `json` | Output format: `json`, `html` |

### GET /api/v1/audits/{id}/openacr

Get OpenACR accessibility conformance report.

**Query Parameters:**

| Parameter | Default | Description |
|-----------|---------|-------------|
| `format` | `yaml` | Output format: `yaml`, `json` |
| `product_name` | URL | Product name override |
| `product_version` | - | Product version |
| `author_name` | `agent-a11y` | Author name |
| `author_email` | - | Author email |
| `vendor_name` | - | Vendor company name |
| `vendor_email` | - | Vendor email |
| `catalog` | auto | Catalog ID (e.g., `2.5-edition-wcag-2.2-508-en`) |

**Example:**

```bash
# Get OpenACR as YAML (default)
curl http://localhost:8080/api/v1/audits/{id}/openacr

# Get as JSON with custom metadata
curl "http://localhost:8080/api/v1/audits/{id}/openacr?format=json&product_name=MyApp&author_email=a11y@example.com"
```

**Response (YAML):**

```yaml
title: https://example.com Accessibility Conformance Report
product:
  name: https://example.com
catalog: 2.5-edition-wcag-2.2-508-en
author:
  name: agent-a11y
report_date: "2025-01-15"
chapters:
  success_criteria_level_a:
    notes: No issues found at this level.
    criteria:
      - num: "1.1.1"
        components:
          - name: web
            adherence:
              level: supports
```

---

## Legacy Endpoints

### POST /audit

Run an accessibility audit.

**Request:**

```json
{
  "url": "https://example.com",
  "level": "AA",
  "version": "2.2",
  "crawl": {
    "enabled": false,
    "depth": 2,
    "maxPages": 50
  },
  "llm": {
    "enabled": true,
    "provider": "anthropic",
    "model": "claude-sonnet-4-20250514"
  }
}
```

**Response:**

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
  "findings": [
    {
      "id": "finding-1",
      "ruleId": "color-contrast",
      "description": "Elements must have sufficient color contrast",
      "impact": "serious",
      "level": "AA",
      "element": "p",
      "selector": "#content > p:nth-child(2)",
      "html": "<p style=\"color: #777\">...</p>",
      "successCriteria": ["1.4.3"],
      "help": "https://dequeuniversity.com/rules/axe/4.x/color-contrast"
    }
  ]
}
```

**Status Codes:**

| Code | Description |
|------|-------------|
| 200 | Audit completed |
| 400 | Invalid request |
| 500 | Audit failed |

### GET /health

Health check endpoint.

**Response:**

```json
{
  "status": "healthy",
  "version": "0.1.0"
}
```

## Examples

### cURL

```bash
curl -X POST http://localhost:8080/audit \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "level": "AA",
    "version": "2.2"
  }'
```

### Python

```python
import requests

response = requests.post("http://localhost:8080/audit", json={
    "url": "https://example.com",
    "level": "AA",
    "version": "2.2"
})

result = response.json()
print(f"Score: {result['score']}/100")
```

### JavaScript

```javascript
const response = await fetch("http://localhost:8080/audit", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    url: "https://example.com",
    level: "AA",
    version: "2.2"
  })
});

const result = await response.json();
console.log(`Score: ${result.score}/100`);
```
