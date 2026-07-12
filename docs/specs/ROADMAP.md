# Agent A11y Roadmap

This document outlines planned features for achieving fully agentic accessibility identification, fix, and validation.

## Overview: Agentic A11Y Loop

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        AGENTIC A11Y LOOP                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐             │
│  │ IDENTIFY     │────▶│ FIX          │────▶│ VALIDATE     │────┐        │
│  │              │     │              │     │              │    │        │
│  │ agent-a11y   │     │ coding agent │     │ agent-a11y   │    │        │
│  │ v0.4.0 ✅    │     │ needs help   │     │ needs delta  │    │        │
│  └──────────────┘     └──────────────┘     └──────────────┘    │        │
│         ▲                                                       │        │
│         └───────────────────────────────────────────────────────┘        │
│                              (loop until GO)                             │
└─────────────────────────────────────────────────────────────────────────┘
```

## Current Status (v0.4.0)

### Completed

- [x] Agent-optimized JSON output as default
- [x] Fix patterns mapping WCAG techniques to actionable fixes
- [x] Design system integration for token suggestions
- [x] MCP server returns structured agent output
- [x] GO/WARN/NO-GO status for workflow automation

### Gaps

| Gap | Impact | Priority |
|-----|--------|----------|
| Source code mapping | Coding agents can't locate source files | ✅ v0.5.0 |
| Project-specific fixes | Teams can't define custom fix patterns | ✅ v0.5.1 |
| Framework adapters | Fix patterns are generic HTML, not framework-specific | Deprioritized |
| Validation delta | No before/after comparison | High |
| Fix verification | Can't verify specific issues were resolved | Medium |
| Component context | Findings are element-level, fixes are component-level | Medium |

---

## v0.5.0 - Source Code Mapping

**Goal**: Map DOM findings to source file locations so coding agents can apply fixes.

### New Types

```go
// SourceLocation maps a DOM element to its source code location
type SourceLocation struct {
    File       string `json:"file"`       // src/components/Hero.tsx
    Line       int    `json:"line"`       // 42
    Column     int    `json:"column"`     // 8
    Component  string `json:"component"`  // Hero
    Framework  string `json:"framework"`  // react, vue, svelte, vanilla
    SourceMap  string `json:"sourceMap"`  // path to source map if available
}

// Add to ElementContext
type ElementContext struct {
    // ... existing fields ...
    Source *SourceLocation `json:"source,omitempty"`
}
```

### Implementation

| Task | Description | Status |
|------|-------------|--------|
| Source map parser | Parse Webpack/Vite source maps | ✅ Done |
| Framework detection | Detect React, Vue, Svelte from page | ✅ Done |
| Component extraction | Extract component names from devtools data | ✅ Done |
| CLI flag `--source-map` | Path to source map directory | ✅ Done |
| MCP tool `set_source_maps` | Configure source maps dynamically | Pending |

### Source Map Integration

```bash
# CLI with source maps
agent-a11y audit https://localhost:3000 --source-map ./dist/

# Output includes source locations
{
  "element": {
    "selector": "img.hero-image",
    "source": {
      "file": "src/components/Hero.tsx",
      "line": 42,
      "component": "Hero"
    }
  }
}
```

### Framework Detection

| Framework | Detection Method |
|-----------|------------------|
| React | `__REACT_DEVTOOLS_GLOBAL_HOOK__`, `data-reactroot` |
| Vue | `__VUE__`, `data-v-*` attributes |
| Svelte | `__svelte`, `svelte-*` classes |
| Angular | `ng-version`, `_nghost-*` attributes |
| Vanilla | No framework markers detected |

---

## v0.6.0 - Framework Adapters

> **Status: Deprioritized**
>
> Modern coding agents (Claude Code, Kiro CLI, Codex CLI) can translate generic
> fix patterns to framework-specific code using project context. They see imports,
> existing patterns, and framework conventions in the codebase. Pre-computed
> framework fixes add complexity without clear value. Revisit if agent experience
> shows this is needed.
>
> Focus instead on **v0.7.0 Validation Delta** which enables the autonomous
> feedback loop (verify fixes worked, detect regressions).

**Goal**: Generate framework-specific fix code instead of generic HTML.

### Adapter Interface

```go
// FrameworkAdapter transforms generic fix patterns to framework-specific code
type FrameworkAdapter interface {
    Name() string
    Detect(html string) bool
    TransformFix(pattern FixPattern, context ElementContext) FrameworkFix
}

type FrameworkFix struct {
    Language   string `json:"language"`   // tsx, vue, svelte
    Code       string `json:"code"`       // Actual code to insert
    ImportAdd  string `json:"importAdd"`  // Imports to add (if any)
    PropsAdd   string `json:"propsAdd"`   // Props to add to component
}
```

### Supported Frameworks

| Framework | Adapter | Example Transformation |
|-----------|---------|------------------------|
| React/TSX | `ReactAdapter` | `<img>` → `<Image alt={...} />` |
| Vue SFC | `VueAdapter` | `<img>` → `<img :alt="altText">` |
| Svelte | `SvelteAdapter` | `<img>` → `<img alt={altText}>` |
| HTML | `HTMLAdapter` | `<img>` → `<img alt="...">` (default) |

### Implementation

| Task | Description | Status |
|------|-------------|--------|
| Adapter interface | Define `FrameworkAdapter` interface | Deprioritized |
| React adapter | TSX/JSX transformations | Deprioritized |
| Vue adapter | Vue SFC transformations | Deprioritized |
| Svelte adapter | Svelte transformations | Deprioritized |
| CLI flag `--framework` | Override framework detection | Deprioritized |
| Agent output field | Add `frameworkFix` to remediation | Deprioritized |

### Example Output

```json
{
  "remediation": {
    "fixPatterns": [
      {
        "type": "attribute",
        "action": "add",
        "target": "alt",
        "example": "<img alt=\"Description\">"
      }
    ],
    "frameworkFix": {
      "framework": "react",
      "language": "tsx",
      "code": "<Image src={heroImage} alt=\"Hero banner showing product\" />",
      "propsAdd": "alt: string"
    }
  }
}
```

---

## v0.7.0 - Validation Delta

**Goal**: Compare before/after audits to track fix progress.

### New Types

```go
// ValidationDelta compares two audit results
type ValidationDelta struct {
    Before      *AgentResult `json:"before"`
    After       *AgentResult `json:"after"`

    // Summary
    BeforeTotal int `json:"beforeTotal"`
    AfterTotal  int `json:"afterTotal"`

    // Changes
    Fixed       []FixedFinding     `json:"fixed"`       // Issues resolved
    Remaining   []AgentFinding     `json:"remaining"`   // Still present
    Regressions []AgentFinding     `json:"regressions"` // New issues

    // Status
    Status      string  `json:"status"`      // "IMPROVED", "FIXED", "REGRESSED", "NO_CHANGE"
    Improvement float64 `json:"improvement"` // Percentage improvement
}

type FixedFinding struct {
    Finding   AgentFinding `json:"finding"`
    FixedBy   string       `json:"fixedBy"`   // Which fix pattern was likely used
    Verified  bool         `json:"verified"`  // Confirmed via element fingerprint
}
```

### CLI Commands

```bash
# Compare two audits
agent-a11y compare https://example.com --baseline ./before.json -o delta.json

# Validate after fixes (uses cached baseline)
agent-a11y validate https://example.com --expect-fixed color-contrast,image-alt

# Continuous validation in CI
agent-a11y audit https://example.com --fail-on-regression --baseline ./baseline.json
```

### Implementation

| Task | Description | Status |
|------|-------------|--------|
| Finding fingerprint | Stable ID for findings across audits | ✅ Done |
| Delta calculation | Compare before/after findings | ✅ Done |
| `compare` command | CLI command for delta comparison | Existing (VPAT) |
| `validate` command | Validate expected fixes | ✅ Done |
| MCP tool `compare_audits` | Compare two audit results | Pending |
| Regression detection | Identify new issues introduced | ✅ Done |

### Finding Fingerprint

Stable identification across audits using:

```go
// Fingerprint generates a stable ID for a finding
func (f *Finding) Fingerprint() string {
    // Combine stable attributes
    data := fmt.Sprintf("%s:%s:%s:%s",
        f.RuleID,
        f.Selector,
        normalizeHTML(f.HTML),
        f.PageURL,
    )
    return sha256(data)[:16]
}
```

---

## v0.8.0 - Fix Verification

**Goal**: Verify that specific fixes resolved specific issues.

### Screenshot Comparison

```go
// ElementScreenshot captures before/after state
type ElementScreenshot struct {
    Selector  string    `json:"selector"`
    Before    string    `json:"before"`    // Base64 PNG
    After     string    `json:"after"`     // Base64 PNG
    Different bool      `json:"different"` // Visual difference detected
}

// Add to ValidationDelta
type ValidationDelta struct {
    // ... existing fields ...
    Screenshots []ElementScreenshot `json:"screenshots,omitempty"`
}
```

### Implementation

| Task | Description | Status |
|------|-------------|--------|
| Element screenshot | Capture element before fix | Pending |
| Visual diff | Compare before/after screenshots | Pending |
| Accessibility tree diff | Compare a11y tree structure | Pending |
| Fix attribution | Link resolved findings to fix patterns | Pending |

---

## Integration Points

### design-system-spec Integration

| Feature | Description | Status |
|---------|-------------|--------|
| Token lookup | Query design system for token values | v0.4.0 ✅ |
| Component matching | Match DOM to design system components | v0.4.0 ✅ |
| Token suggestions | Suggest compliant tokens for fixes | v0.4.0 ✅ |
| Component props | Get required props for accessibility | Pending |
| Anti-patterns | Get accessibility anti-patterns to avoid | Pending |

### w3pilot Integration

| Feature | Description | Status |
|---------|-------------|--------|
| Browser automation | Navigate, interact, screenshot | v0.3.0 ✅ |
| Accessibility tree | Get full a11y tree snapshot | v0.3.0 ✅ |
| Source map loading | Load source maps from directory | v0.5.0 ✅ |
| Framework detection | Detect framework from HTML | v0.5.0 ✅ |
| Element fingerprinting | Generate stable element IDs | Pending |

---

## Multi-Agent Workflow

### Full Loop Example

```yaml
# Agent workflow definition
agents:
  - id: a11y-audit
    tool: agent-a11y
    action: audit
    output: findings.json

  - id: code-fixer
    tool: claude-code
    input: findings.json
    action: apply-fixes
    output: changes.patch

  - id: a11y-validate
    tool: agent-a11y
    action: validate
    input:
      baseline: findings.json
      expect_fixed: ["color-contrast", "image-alt"]
    output: validation.json

  - id: loop-or-done
    condition: validation.status != "FIXED"
    goto: code-fixer
```

### MCP Orchestration

```json
{
  "mcpServers": {
    "a11y": {
      "command": "agent-a11y",
      "args": ["mcp", "serve", "--design-system", "./design-system/"]
    },
    "design-system": {
      "command": "dss-mcp",
      "args": ["--spec", "./design-system/"]
    }
  }
}
```

---

## Success Criteria

### v0.5.0 (Source Mapping)

- [ ] Source locations included for 80%+ of findings when source maps available
- [ ] Framework correctly detected for React, Vue, Svelte projects
- [ ] Component names extracted from React/Vue devtools data

### v0.6.0 (Framework Adapters)

- [ ] React adapter generates valid TSX fixes
- [ ] Vue adapter generates valid SFC fixes
- [ ] Coding agents can apply fixes without manual translation

### v0.7.0 (Validation Delta)

- [ ] Delta correctly identifies fixed vs remaining issues
- [ ] Regressions detected and reported
- [ ] CI can fail on regressions

### v0.8.0 (Fix Verification)

- [ ] Visual diff shows element changes
- [ ] Fix attribution links findings to applied fixes
- [ ] Full loop completes autonomously for simple fixes

---

## Related Documents

- [ASDM Integration](../architecture/asdm-integration.md) - How agent-a11y enables ASDM Level 6
- [Report Formats](../guide/report-formats.md) - Agent output format documentation
- [MCP Reference](../reference/mcp.md) - MCP server tools
