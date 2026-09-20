package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/plexusone/omnillm"
)

// This file adds PROACTIVE criterion evaluation, distinct from the finding
// false-positive review in judge.go. Proactive evaluation asks, for a WCAG
// success criterion that no automated check could resolve, "does this page
// satisfy it?" and returns a conformance verdict.
//
// Two evaluator backends share one contract — the same prompt and the same
// verdict schema — so results are comparable regardless of who runs inference:
//
//   - APIEvaluator: agent-a11y calls an LLM through omnillm (needs an API key).
//   - Local (bundle) mode: agent-a11y exports the exact same prompts as a
//     bundle; an external agent (e.g. Claude Code) produces verdicts against
//     the same schema; agent-a11y imports them. No API key required.

// Conformance verdict values (VPAT/OpenACR vocabulary).
const (
	VerdictSupports          = "Supports"
	VerdictPartiallySupports = "Partially Supports"
	VerdictDoesNotSupport    = "Does Not Support"
	VerdictNotApplicable     = "Not Applicable"
	VerdictNotEvaluated      = "Not Evaluated"
)

// CriterionQuery is a request to proactively evaluate one success criterion.
type CriterionQuery struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Level       string `json:"level"`
	Requirement string `json:"requirement"`
}

// CriterionVerdict is the evaluator's proactive verdict for a criterion. When
// Conformance is "Not Evaluated", EvidenceGap states what evidence is missing.
type CriterionVerdict struct {
	ID          string  `json:"id"`
	Conformance string  `json:"conformance"`
	Confidence  float64 `json:"confidence"`
	Reasoning   string  `json:"reasoning"`
	EvidenceGap string  `json:"evidenceGap,omitempty"`
}

// CriterionEvaluator proactively evaluates criteria against page evidence.
// Both the API backend and any agent-driven local backend satisfy it.
type CriterionEvaluator interface {
	EvaluateCriteria(ctx context.Context, queries []CriterionQuery, page PageContext) ([]CriterionVerdict, error)
}

// CriterionSystemPrompt is the shared instruction used by BOTH modes, so an API
// model and a local agent judge on identical terms.
const CriterionSystemPrompt = `You are an expert accessibility auditor performing a PROACTIVE evaluation of a web page against specific WCAG success criteria.

For each criterion, choose exactly one conformance value:
- "Supports": the page meets the criterion based on the evidence provided.
- "Partially Supports": the page meets the criterion in part but has gaps.
- "Does Not Support": the page fails the criterion.
- "Not Applicable": the page contains no content of the type this criterion governs (e.g. no audio/video for media criteria, no forms for input criteria).
- "Not Evaluated": the provided evidence is insufficient to decide. Use this honestly — never guess "Supports" from absence of evidence. State exactly what evidence is needed in "evidenceGap" (e.g. "screenshot for color use", "raw DOM for alt text", "multi-page crawl", "interaction with the form").

Respond with a JSON array; one object per criterion:
[{"id":"1.4.4","conformance":"Supports","confidence":0.0-1.0,"reasoning":"...","evidenceGap":""}]

Be precise and conservative. Prefer "Not Applicable" or "Not Evaluated" over an unsupported "Supports".`

// BuildCriterionPrompt renders the user prompt for one criterion + page. Both
// modes use this identical prompt.
func BuildCriterionPrompt(q CriterionQuery, page PageContext) string {
	var sb strings.Builder
	sb.WriteString("## Success Criterion\n\n")
	fmt.Fprintf(&sb, "**%s %s** (Level %s)\n", q.ID, q.Name, q.Level)
	if q.Requirement != "" {
		fmt.Fprintf(&sb, "**Requires:** %s\n", q.Requirement)
	}

	sb.WriteString("\n## Page Evidence\n\n")
	fmt.Fprintf(&sb, "**URL:** %s\n", page.URL)
	fmt.Fprintf(&sb, "**Title:** %s\n", page.Title)
	if page.Language != "" {
		fmt.Fprintf(&sb, "**Language:** %s\n", page.Language)
	}
	if page.HTMLSnippet != "" {
		fmt.Fprintf(&sb, "\n**HTML:**\n```html\n%s\n```\n", page.HTMLSnippet)
	}
	if page.Screenshot != "" {
		sb.WriteString("\n_A screenshot of the page is available for visual criteria._\n")
	} else {
		sb.WriteString("\n_No screenshot provided; do not guess visual criteria (color, contrast)._\n")
	}
	return sb.String()
}

// ParseVerdicts parses a JSON array of verdicts (as produced by either mode).
// It tolerates a leading/trailing markdown code fence.
func ParseVerdicts(data []byte) ([]CriterionVerdict, error) {
	s := strings.TrimSpace(string(data))
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	var verdicts []CriterionVerdict
	if err := json.Unmarshal([]byte(s), &verdicts); err != nil {
		return nil, fmt.Errorf("invalid verdicts JSON: %w", err)
	}
	return verdicts, nil
}

// --- Local (bundle) mode ---

// EvaluationBundle is the exported worklist for local mode. It contains the
// shared system prompt, the page evidence, and a per-criterion prompt — exactly
// what an external agent needs to produce verdicts on the same terms as the API.
type EvaluationBundle struct {
	SystemPrompt string       `json:"systemPrompt"`
	Page         PageContext  `json:"page"`
	Items        []BundleItem `json:"items"`
}

// BundleItem pairs a query with its rendered prompt.
type BundleItem struct {
	Query  CriterionQuery `json:"query"`
	Prompt string         `json:"prompt"`
}

// BuildBundle constructs the local-mode export from queries and page evidence.
func BuildBundle(queries []CriterionQuery, page PageContext) EvaluationBundle {
	b := EvaluationBundle{SystemPrompt: CriterionSystemPrompt, Page: page}
	for _, q := range queries {
		b.Items = append(b.Items, BundleItem{Query: q, Prompt: BuildCriterionPrompt(q, page)})
	}
	return b
}

// --- API mode ---

// APIEvaluator evaluates criteria by calling an LLM through omnillm.
type APIEvaluator struct {
	client *omnillm.ChatClient
	model  string
}

// NewAPIEvaluator constructs an API-backed criterion evaluator.
func NewAPIEvaluator(client *omnillm.ChatClient, model string) *APIEvaluator {
	return &APIEvaluator{client: client, model: model}
}

// EvaluateCriteria implements CriterionEvaluator via the LLM API. It evaluates
// each criterion with the shared prompt and parses the returned verdict.
func (e *APIEvaluator) EvaluateCriteria(ctx context.Context, queries []CriterionQuery, page PageContext) ([]CriterionVerdict, error) {
	verdicts := make([]CriterionVerdict, 0, len(queries))
	maxTokens := 800
	temperature := 0.1

	for _, q := range queries {
		req := &omnillm.ChatCompletionRequest{
			Model: e.model,
			Messages: []omnillm.Message{
				{Role: omnillm.RoleSystem, Content: CriterionSystemPrompt},
				{Role: omnillm.RoleUser, Content: BuildCriterionPrompt(q, page)},
			},
			MaxTokens:   &maxTokens,
			Temperature: &temperature,
		}
		resp, err := e.client.CreateChatCompletion(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("evaluate %s: %w", q.ID, err)
		}
		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("evaluate %s: no choices returned", q.ID)
		}
		// The single-criterion prompt yields either a bare object or a
		// one-element array; normalize by wrapping bare objects.
		content := strings.TrimSpace(resp.Choices[0].Message.Content)
		vs, err := parseVerdictFlexible(content, q.ID)
		if err != nil {
			return nil, fmt.Errorf("evaluate %s: %w", q.ID, err)
		}
		verdicts = append(verdicts, vs...)
	}
	return verdicts, nil
}

// parseVerdictFlexible accepts either a JSON array or a single object.
func parseVerdictFlexible(content, id string) ([]CriterionVerdict, error) {
	s := strings.TrimSpace(content)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "[") {
		return ParseVerdicts([]byte(s))
	}
	var v CriterionVerdict
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, fmt.Errorf("invalid verdict JSON: %w", err)
	}
	if v.ID == "" {
		v.ID = id
	}
	return []CriterionVerdict{v}, nil
}
