# MCP Server Reference

agent-a11y provides a Model Context Protocol (MCP) server for integration with AI assistants like Claude Code.

## Starting the Server

```bash
agent-a11y mcp serve
```

The server communicates via stdio using the MCP protocol.

## Claude Desktop Integration

Add to your Claude Desktop configuration (`~/.config/claude/config.json`):

```json
{
  "mcpServers": {
    "a11y": {
      "command": "agent-a11y",
      "args": ["mcp", "serve"]
    }
  }
}
```

## Agent-Optimized Output

All audit tools return **agent-optimized JSON** with actionable fix patterns that coding agents can use to implement fixes directly. The output includes:

- **Fix patterns**: Specific instructions with type, action, target, and example code
- **Token suggestions**: Design system token recommendations (when `--design-system` is configured)
- **Fix confidence**: Confidence score (0-1) that the suggested fix will resolve the issue
- **Status**: Overall GO/WARN/NO-GO status for workflow automation

## Available Tools

### audit_page

Audit a single page for accessibility issues. Returns agent-optimized JSON with fix patterns.

**Input:**

```json
{
  "url": "https://example.com",
  "level": "AA"
}
```

**Output:**

```json
{
  "url": "https://example.com",
  "status": "WARN",
  "summary": {
    "total": 5,
    "critical": 0,
    "serious": 2,
    "fixable": 4
  },
  "findings": [
    {
      "finding": {
        "ruleId": "image-alt",
        "impact": "critical",
        "selector": "img.hero"
      },
      "remediation": {
        "fixPatterns": [
          {
            "type": "attribute",
            "action": "add",
            "target": "alt",
            "example": "<img src=\"...\" alt=\"Description\">"
          }
        ],
        "fixConfidence": 0.85
      }
    }
  ]
}
```

### audit_site

Audit an entire website by crawling.

**Input:**

```json
{
  "url": "https://example.com",
  "level": "AA",
  "depth": 2,
  "maxPages": 50
}
```

### get_finding_details

Get detailed information about a specific finding.

**Input:**

```json
{
  "findingId": "finding-123"
}
```

### generate_report

Generate a formatted report from audit results.

**Input:**

```json
{
  "auditId": "audit-456",
  "format": "html"
}
```

### get_remediation

Get remediation suggestions for findings.

**Input:**

```json
{
  "findingId": "finding-123"
}
```

## Example Conversation

**User:** Check the accessibility of example.com

**Assistant:** I'll audit example.com for accessibility issues.

*[Calls audit_page tool]*

The audit found 5 accessibility issues:

**Score:** 85/100 (Non-Conformant)

**Critical Issues:**
- None

**Serious Issues:**
1. Color contrast insufficient on navigation links
2. Form inputs missing labels

**Recommendations:**
- Increase contrast ratio on nav links to at least 4.5:1
- Add `<label>` elements for all form inputs

Would you like me to generate a detailed HTML report?

## Resources

The MCP server exposes these resources:

### wcag_criteria

Access WCAG success criteria information.

```
wcag://criteria/1.4.3
```

### audit_results

Access previous audit results.

```
audit://results/{audit-id}
```
