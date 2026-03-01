# Architecture Overview

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         agent-a11y                               │
├─────────────────────────────────────────────────────────────────┤
│  CLI          HTTP API         MCP Server         Go Library    │
│   │              │                 │                  │         │
│   └──────────────┴────────┬────────┴──────────────────┘         │
│                           │                                      │
│                    ┌──────▼──────┐                              │
│                    │ Audit Engine │                              │
│                    └──────┬──────┘                              │
│           ┌───────────────┼───────────────┐                     │
│           │               │               │                      │
│    ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐             │
│    │   axe-core   │ │ Specialized │ │ LLM Judge   │             │
│    │   Rules      │ │   Tests     │ │             │             │
│    └──────────────┘ └─────────────┘ └─────────────┘             │
│                           │                                      │
│                    ┌──────▼──────┐                              │
│                    │   Crawler    │                              │
│                    └──────┬──────┘                              │
│                           │                                      │
│                    ┌──────▼──────┐                              │
│                    │  vibium-go   │                              │
│                    │  (browser)   │                              │
│                    └─────────────┘                              │
└─────────────────────────────────────────────────────────────────┘
```

## Core Packages

### `a11y` (root)

Public API and result types.

- `Auditor` - Main entry point
- `Result` - Audit results
- `Finding` - Individual accessibility issue

### `audit`

Core audit engine.

- `Engine` - Orchestrates audit process
- `AuditResult` - Internal result type
- `Finding` - Internal finding type

### `audit/axe`

axe-core integration.

- Injects axe-core into pages
- Parses axe-core results
- Maps to WCAG criteria

### `audit/specialized`

Browser-based specialized tests.

- Keyboard navigation
- Focus management
- Reflow testing
- Target size
- Text spacing

### `audit/rules`

WCAG rule definitions.

- Rule registry
- Criteria mapping
- Impact classification

### `crawler`

Website crawler.

- Link extraction
- SPA detection
- Depth/limit management
- Robots.txt respect

### `journey`

Journey-based testing.

- YAML journey compiler
- Step executor
- State management

### `llm`

LLM-as-a-Judge integration.

- Provider abstraction
- Prompt templates
- Result parsing

### `report`

Report generation.

- JSON, HTML, Markdown
- VPAT 2.4
- WCAG-EM
- CSV

### `remediation`

Remediation task generation.

- Issue analysis
- Fix suggestions
- Priority assignment

### `mcp`

MCP server implementation.

- Tool definitions
- Resource handlers
- Session management

### `api`

HTTP API server.

- REST endpoints
- Request validation
- Response formatting

### `wcag`

WCAG criteria database.

- Success criteria
- Levels and versions
- Techniques

### `types`

Shared types.

- Impact levels
- WCAG levels
- Common interfaces

### `config`

Configuration management.

- File loading
- Environment variables
- Defaults

## Data Flow

1. **Input** - URL from CLI, API, MCP, or library
2. **Crawling** (optional) - Discover pages via crawler
3. **Browser** - Load page via vibium-go
4. **axe-core** - Run accessibility rules
5. **Specialized** - Run browser-based tests
6. **LLM** (optional) - Evaluate findings
7. **Aggregation** - Combine results
8. **Reporting** - Generate output format
9. **Output** - Return to caller

## Dependencies

- **vibium-go** - Browser automation
- **omnillm** - LLM provider abstraction
- **multi-agent-spec** - Multi-agent integration
- **cobra** - CLI framework
- **MCP SDK** - Model Context Protocol
