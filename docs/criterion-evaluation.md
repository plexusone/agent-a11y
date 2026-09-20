# Proactive Criterion Evaluation

Automated and specialized checks can only resolve a subset of WCAG success
criteria. The rest — captions, use of color, link purpose, consistent
navigation, and so on — require **judgment**. Rather than silently marking those
"Supports" (an overclaim) or leaving them permanently "Not Evaluated", agent-a11y
can **proactively evaluate** them: ask, for each unresolved criterion, "does this
page satisfy it?" and record an honest verdict.

There are **two modes**, and they operate identically except for *who runs the
inference*. Both use the **same system prompt** (`llm.CriterionSystemPrompt`),
the **same per-criterion prompt** (`llm.BuildCriterionPrompt`), and the **same
verdict schema** (`llm.CriterionVerdict`), so results are comparable.

| | Local (agent) mode | API mode |
|---|---|---|
| Who judges | An external agent (e.g. Claude Code) | An LLM via the API |
| API key | Not required | Required |
| Entry point | `Result.EvaluationBundle()` → verdicts JSON → `Result.ConformanceWithVerdicts()` | `Result.EvaluateCriteria(ctx, apiEvaluator)` |
| Use when | Validating, offline, or already inside an agent loop | Unattended/CI runs |

## The verdict vocabulary

Every verdict is one of: `Supports`, `Partially Supports`, `Does Not Support`,
`Not Applicable`, or `Not Evaluated`. The last two matter for honesty:

- **Not Applicable** — the page has no content of the type the criterion governs
  (e.g. no audio/video for caption criteria). This is a legitimate, common
  outcome and is what content-level evidence most often supports.
- **Not Evaluated** — the evidence is insufficient. The verdict must then state
  what evidence is needed (`evidenceGap`), e.g. "screenshot for color use",
  "raw DOM for alt text", "multi-page crawl", "interaction with the form".

A verdict never overrides a finding-backed `Does Not Support`/`Partially
Supports`; it only resolves criteria automation left `Not Evaluated`. Enabling an
LLM does **not** by itself flip anything — only an actual verdict does.

## Local (agent) mode

No API key. The agent that is already running the tool acts as the judge.

```go
res, _ := auditor.AuditPage(ctx, url)

bundle := res.EvaluationBundle()      // shared system prompt + page evidence + per-criterion prompts
// hand `bundle` to the agent; it returns verdicts on the same schema:
//   [{"id":"1.2.1","conformance":"Not Applicable","confidence":0.9,"reasoning":"no media","evidenceGap":""}]
verdicts, _ := llm.ParseVerdicts(agentOutput)

criteria := res.ConformanceWithVerdicts(verdicts)   // baseline + verdicts applied
```

The bundle is plain JSON, so the export/import can also cross a process boundary
(write bundle to a file, agent writes verdicts to a file, import them).

## API mode

Needs an LLM API key (via omnillm). Same prompts and schema, run automatically.

```go
client := omnillm.NewChatClient(/* provider, key */)
evaluator := llm.NewAPIEvaluator(client, "claude-...")

verdicts, _ := res.EvaluateCriteria(ctx, evaluator)   // calls the model per criterion
criteria := res.ConformanceWithVerdicts(verdicts)
```

Because `APIEvaluator` implements `llm.CriterionEvaluator`, any alternative
backend (a different provider, a cached evaluator, a test stub) can be dropped in
without changing callers.

## Evidence the judge needs

Content alone (title, text, links, headings) resolves media (→ Not Applicable),
link purpose, headings/labels, and language. To resolve the rest, provide richer
evidence in `PageContext`:

- **Screenshot** → use of color (1.4.1), non-text contrast (1.4.11)
- **Raw DOM** → non-text content / alt quality (1.1.1)
- **Multi-page crawl** → consistent navigation/identification/help (3.2.3/3.2.4/3.2.6), multiple ways (2.4.5)
- **Interaction** → on-input behavior (3.2.2), form error handling (3.3.1/3.3.3), authentication (3.3.8)

When evidence is missing, the honest verdict is `Not Evaluated` with the gap
named — never a guessed `Supports`.
