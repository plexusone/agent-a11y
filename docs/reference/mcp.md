# MCP Server Reference

agent-a11y provides a Model Context Protocol (MCP) server for integration with AI assistants.

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

## Available Tools

### audit_page

Audit a single page for accessibility issues.

**Input:**

```json
{
  "url": "https://example.com",
  "level": "AA",
  "version": "2.2"
}
```

**Output:**

```json
{
  "score": 85,
  "conformant": false,
  "summary": "WCAG 2.2 Level AA: Non-Conformant (Score: 85/100)",
  "findings": [...]
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
