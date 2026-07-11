# Agent-A11y: ASDM Integration

This document describes how agent-a11y enables organizations to achieve [ASDM (Autonomous Software Delivery Model)](https://productbuildershq.github.io/frameworks/software-delivery-autonomy/) Level 6 for accessibility compliance.

## ASDM Context

The Software Delivery Autonomy Model defines 7 levels of increasing autonomy:

| Level | Name | Human Role | Key Characteristic |
|-------|------|------------|-------------------|
| 5 | Agentic Engineering | Batch reviewer | Human review bottleneck |
| 6 | Autonomous Coding & Review | Specification owner | Scenario-based validation |
| 7 | Autonomous Operations | Governor | Production telemetry validation |

**Key insight**: Level 6 requires replacing human code review with **specification-based validation**. Agent-a11y provides the accessibility validation layer.

## Agent-A11y's Role in Level 6

Agent-a11y is a **validation component** for ASDM Level 6. It provides:

```
┌─────────────────────────────────────────────────────────────────┐
│                    ASDM LEVEL 6 STACK                            │
├─────────────────────────────────────────────────────────────────┤
│  Specification Layer                                             │
│  ├── design-system-spec  → Visual specifications                │
│  ├── WCAG 2.x            → Accessibility specifications         │
│  └── API schemas         → Functional specifications            │
├─────────────────────────────────────────────────────────────────┤
│  Validation Layer                                                │
│  ├── design-system-spec  → Visual regression tests              │
│  ├── agent-a11y          → WCAG compliance audits    ◀── HERE   │
│  └── integration tests   → Functional validation                │
├─────────────────────────────────────────────────────────────────┤
│  Decision Layer                                                  │
│  └── multi-agent-spec    → GO/WARN/NO-GO orchestration          │
└─────────────────────────────────────────────────────────────────┘
```

Agent-a11y validates that implementations conform to **WCAG accessibility specifications**, enabling autonomous merge decisions for accessibility compliance.

## ASDM Capabilities

### 1. Specification-Based Validation

Agent-a11y maps every finding to WCAG success criteria:

```go
type Finding struct {
    RuleID            string      // e.g., "color-contrast"
    SuccessCriteria   []string    // e.g., ["1.4.3", "1.4.11"]
    Level             WCAGLevel   // A, AA, or AAA
    Impact            Impact      // blocker/critical/serious/moderate/minor
}
```

**WCAG as specification**: The WCAG guidelines serve as machine-readable specifications. Agent-a11y's rule registry maps rules to success criteria, enabling automated compliance determination.

### 2. Multi-Agent Integration

Results export to multi-agent-spec format with GO/WARN/NO-GO decisions:

```go
// AgentResult provides multi-agent workflow integration
result := audit.AgentResult()
// Returns:
// - Status: GO (conformant) | WARN (minor issues) | NO-GO (critical issues)
// - TaskResults: Per-criterion compliance status
```

This enables agent-a11y to participate in autonomous decision gates:

```yaml
# Multi-agent workflow
steps:
  - agent: code-generator
    action: implement_component

  - agent: agent-a11y
    action: audit_accessibility
    gate:
      - status: GO → continue
      - status: NO-GO → reject

  - agent: design-system-spec
    action: visual_regression
    gate:
      - status: PASS → merge
      - status: FAIL → iterate
```

### 3. Standardized Remediation References

Every finding includes actionable fix guidance:

```go
type Remediation struct {
    Summary      string           // Brief fix description
    Techniques   []TechniqueRef   // WCAG Techniques (H37, G94, ARIA6, etc.)
    ACTRuleID    string           // W3C ACT Rule reference
    AxeRuleID    string           // axe-core rule mapping
}
```

**Why this matters for Level 6**: Code generation agents can use these standardized references to implement fixes without human guidance. A Technique reference like "G18" (luminosity contrast ratio) provides a deterministic fix pattern.

### 4. LLM-as-a-Judge

Optional AI evaluation reduces false positives:

```bash
agent-a11y audit https://example.com --llm anthropic
```

Each finding includes:

```go
type LLMEvaluation struct {
    Confirmed   bool    // LLM agrees this is a real issue
    Confidence  float64 // Confidence level (0-1)
    Reasoning   string  // Explanation
}
```

This enables filtering findings before agent action, reducing wasted remediation cycles.

### 5. MCP Server Integration

Agent-a11y exposes MCP tools for AI assistant integration:

```json
{
  "tools": [
    {"name": "audit_page", "description": "Audit single page for WCAG compliance"},
    {"name": "audit_site", "description": "Crawl and audit entire site"},
    {"name": "check_criterion", "description": "Check specific WCAG criterion"},
    {"name": "generate_vpat", "description": "Generate VPAT 2.4 report"}
  ]
}
```

AI assistants (Claude, etc.) can invoke agent-a11y as part of autonomous workflows.

## Level 6 Workflow with Agent-A11y

### Autonomous Accessibility Compliance

```
┌─────────────────────────────────────────────────────────────────┐
│                    AUTONOMOUS LOOP                               │
│                                                                  │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌─────────┐  │
│   │ Generate │───▶│  Audit   │───▶│ Decide   │───▶│  Merge  │  │
│   │   Code   │    │  A11y    │    │  GO/NO   │    │   or    │  │
│   └──────────┘    └──────────┘    └──────────┘    │ Iterate │  │
│        ▲                                          └────┬────┘  │
│        │                                               │       │
│        └───────────────────────────────────────────────┘       │
│                       (iterate if NO-GO)                        │
└─────────────────────────────────────────────────────────────────┘
```

### CI/CD Integration

```yaml
name: Autonomous A11y Compliance

on: pull_request

jobs:
  accessibility:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build and deploy preview
        run: npm run build && npm run preview &

      - name: Audit accessibility
        id: audit
        run: |
          agent-a11y audit http://localhost:3000 \
            --level AA \
            --format json \
            --output audit-result.json

          # Check conformance
          CONFORMANT=$(jq '.conformant' audit-result.json)
          echo "conformant=$CONFORMANT" >> $GITHUB_OUTPUT

      - name: Auto-merge if conformant
        if: steps.audit.outputs.conformant == 'true'
        run: gh pr merge --auto --squash

      - name: Block if non-conformant
        if: steps.audit.outputs.conformant == 'false'
        run: |
          gh pr comment --body "$(cat <<EOF
          ## Accessibility Audit Failed

          WCAG Level AA conformance not met. See findings:

          $(jq -r '.findings[] | "- [\(.impact)] \(.ruleId): \(.message)"' audit-result.json)

          Escalating to accessibility specification owner.
          EOF
          )"
          gh pr edit --add-label "a11y-review-needed"
          exit 1
```

### Combining with Design-System-Spec

For complete Level 6 UI validation, combine both validators:

```yaml
jobs:
  validate:
    steps:
      # Visual validation
      - name: Visual regression
        run: dss visual test --baseline v2.0.0

      # Accessibility validation
      - name: Accessibility audit
        run: agent-a11y audit $PREVIEW_URL --level AA

      # Both must pass for auto-merge
      - name: Auto-merge
        if: success()
        run: gh pr merge --auto --squash
```

## Conformance Levels as Specification Tiers

Agent-a11y supports tiered validation:

| WCAG Level | Conformance Rule | Use Case |
|------------|------------------|----------|
| A | No critical or serious issues | Minimum legal compliance |
| AA | No critical, serious, or major issues | Industry standard |
| AAA | No critical, serious, major, or moderate issues | Maximum accessibility |

Configure validation stringency:

```bash
# Strict AA compliance (most organizations)
agent-a11y audit $URL --level AA

# Lenient A-only compliance (legacy systems)
agent-a11y audit $URL --level A

# Maximum AAA compliance (public sector)
agent-a11y audit $URL --level AAA
```

## What Agent-A11y Does NOT Provide

To achieve full Level 6, agent-a11y must be combined with:

| Capability | Agent-A11y | Required Addition |
|------------|------------|-------------------|
| Issue detection | ✅ Full | - |
| Fix suggestion | ✅ WCAG Techniques | - |
| Code generation | ❌ None | Code generation agent |
| Fix verification | ⚠️ Re-audit only | Automated loop |
| Visual validation | ❌ None | design-system-spec |

### Complete Level 6 Stack

```
┌─────────────────────────────────────────────────────────────────┐
│  Code Generation Agent                                           │
│  - Receives findings from agent-a11y                             │
│  - Uses WCAG Techniques for fix patterns                         │
│  - References design-system-spec for component patterns          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Validation Layer                                                │
│  ├── agent-a11y: WCAG compliance ✓                              │
│  └── design-system-spec: Visual regression ✓                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Decision Gate                                                   │
│  - Both validators pass → Auto-merge                            │
│  - Either fails → Iterate or escalate                           │
└─────────────────────────────────────────────────────────────────┘
```

## Extending to Level 7: Autonomous Operations

At Level 7, accessibility monitoring extends to production:

```yaml
# Production accessibility monitoring
monitors:
  - name: homepage-a11y
    url: "https://app.example.com/"
    schedule: "0 */6 * * *"  # Every 6 hours
    level: AA

    on_regression:
      - alert: pagerduty
      - action: rollback_deployment
      - action: create_incident
```

**Production telemetry**: If live site accessibility degrades (e.g., new deployment introduces color contrast issues), autonomous systems can:
1. Detect the regression via scheduled audits
2. Roll back the deployment
3. Create an incident for investigation
4. Trigger fix generation in development environment

## ASDM Practice Mapping

| ASDM Practice | Agent-A11y Implementation |
|---------------|---------------------------|
| Specification-first | WCAG 2.x as accessibility specification |
| Scenario-based validation | WCAG success criteria as scenarios |
| Probabilistic satisfaction | Impact levels (critical/serious/moderate/minor) |
| LLM-as-a-Judge | Optional AI evaluation for false positive reduction |
| Zero human code review | Conformance determination enables auto-merge |
| Full provenance | Findings trace to WCAG criteria, ACT rules, axe rules |
| Governance-by-policy | Level (A/AA/AAA) configuration |

## Summary

Agent-a11y enables ASDM Level 6 for accessibility by providing:

1. **WCAG as specification**: Machine-readable accessibility requirements
2. **Automated validation**: Conformance determination without human review
3. **Multi-agent integration**: GO/WARN/NO-GO decision outputs
4. **Standardized remediation**: WCAG Techniques for automated fix generation
5. **LLM evaluation**: AI-assisted false positive reduction

Combined with design-system-spec (visual validation) and a code generation agent, agent-a11y completes the validation layer for autonomous UI delivery.

## References

- [ASDM Framework](https://productbuildershq.github.io/frameworks/software-delivery-autonomy/)
- [design-system-spec ASDM Integration](https://github.com/plexusone/design-system-spec/docs/specs/features/agentic-delivery/ASDM-INTEGRATION.md)
- [WCAG 2.2 Specification](https://www.w3.org/TR/WCAG22/)
- [W3C ACT Rules](https://www.w3.org/WAI/standards-guidelines/act/rules/)
