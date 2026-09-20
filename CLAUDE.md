# CLAUDE.md — agent-a11y

WCAG accessibility auditing toolkit: axe-core + specialized checks (keyboard,
focus, reflow, resize, hover, flash) + optional LLM-as-a-Judge + **proactive
criterion evaluation**.

## Build & test

```bash
go build ./...
go test ./report/... ./wcag/... ./llm/... ./journey/...
```

Full audits need a Chrome/Chromium binary; the packages above test without one.

## Honesty model (load-bearing — don't regress)

- `report.EvaluateConformance` assigns per-criterion conformance. Criteria with
  no finding and a judgment method (LLM-Judge/Hybrid/Manual) are **"Not
  Evaluated"** — enabling an LLM does not flip them; only a real verdict does
  (`report.ApplyVerdicts`).
- `wcag.WCAG22AA` carries each criterion's `Method`. **Only** criteria the
  specialized runner actually executes may be `MethodSpecialized`, and
  `MethodAutomated` requires real axe rules. `wcag/method_coverage_test.go`
  guards this — keep it green when editing the catalog.
- The VPAT never marks an unevaluated criterion "Supports".

## Proactive criterion evaluation

`llm/criteval.go` — two backends sharing one prompt + verdict schema:
`APIEvaluator` (omnillm) and a local bundle (`BuildBundle` → agent judges →
`ParseVerdicts`). See `docs/criterion-evaluation.md`. `Result.EvaluationBundle`,
`Result.EvaluateCriteria`, `Result.ConformanceWithVerdicts` expose the flow.

## Page evidence

The audit engine captures `pagecapture.PageEvidence` (screenshot, HTML,
structure) during the audit and stores it on `PageResult.Evidence`;
`Result.pageContext()` builds the evaluator's `PageContext` from it.

## Local development replace (IMPORTANT)

`go.mod` uses `replace github.com/plexusone/w3pilot => ../w3pilot` to pick up
the `w3pilot/pagecapture` subpackage before it's released. This is
**development-only** — remove it and depend on a tagged w3pilot version before
pushing/releasing.
