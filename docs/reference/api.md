# HTTP API Reference

## Starting the Server

```bash
agent-a11y serve --port 8080
```

## Endpoints

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
