package report

import (
	"testing"
	"time"

	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/agent-a11y/llm"
	"github.com/plexusone/agent-a11y/wcag"
)

// findConformance returns the conformance status for a criterion ID prefix.
func findConformance(rep *audit.VPATReport, idPrefix string) string {
	for _, c := range rep.WCAGConformance {
		if len(c.Criterion) >= len(idPrefix) && c.Criterion[:len(idPrefix)] == idPrefix {
			return c.Conformance
		}
	}
	return ""
}

func TestVPATDoesNotOverclaimUnevaluatedCriteria(t *testing.T) {
	// A clean audit (no findings), automated only (LLM off), targeting AA.
	result := &audit.AuditResult{
		TargetURL:  "https://example.com",
		StartTime:  time.Now(),
		WCAGLevel:  audit.WCAGLevelAA,
		LLMEnabled: false,
		Pages:      []audit.PageResult{{URL: "https://example.com"}},
	}

	rep := generateVPATReport(result)

	// 1.3.1 is automated (axe rules) → clean pass is honestly "Supports".
	if got := findConformance(rep, "1.3.1"); got != string(wcag.ConformanceSupports) {
		t.Errorf("1.3.1 (automated) = %q, want Supports", got)
	}

	// 1.2.1 requires AI/manual judgment (LLM-Judge) and the judge did not run →
	// must be "Not Evaluated", never "Supports".
	if got := findConformance(rep, "1.2.1"); got != string(wcag.ConformanceNotEvaluated) {
		t.Errorf("1.2.1 (LLM-judge, LLM off) = %q, want Not Evaluated", got)
	}

	// 1.1.1 is hybrid; with LLM off, the judgment portion is unverified →
	// "Not Evaluated", not an assumed "Supports".
	if got := findConformance(rep, "1.1.1"); got != string(wcag.ConformanceNotEvaluated) {
		t.Errorf("1.1.1 (hybrid, LLM off) = %q, want Not Evaluated", got)
	}
}

func TestLLMEnabledAloneDoesNotFlipJudgmentCriteria(t *testing.T) {
	// Enabling an LLM does not, by itself, resolve judgment criteria: nothing
	// has proactively evaluated them, so they must stay "Not Evaluated".
	result := &audit.AuditResult{
		TargetURL:  "https://example.com",
		StartTime:  time.Now(),
		WCAGLevel:  audit.WCAGLevelAA,
		LLMEnabled: true,
		Pages:      []audit.PageResult{{URL: "https://example.com"}},
	}
	rep := generateVPATReport(result)
	if got := findConformance(rep, "1.2.1"); got != string(wcag.ConformanceNotEvaluated) {
		t.Errorf("1.2.1 (LLM-judge, LLM on, no verdict) = %q, want Not Evaluated", got)
	}
}

func TestApplyVerdictsResolvesUnevaluatedCriteria(t *testing.T) {
	result := &audit.AuditResult{
		TargetURL:  "https://example.com",
		StartTime:  time.Now(),
		WCAGLevel:  audit.WCAGLevelAA,
		LLMEnabled: false,
		Pages:      []audit.PageResult{{URL: "https://example.com"}},
	}
	base := EvaluateConformance(result)

	// A proactive verdict resolves a previously Not-Evaluated criterion...
	verdicts := []llm.CriterionVerdict{
		{ID: "1.2.1", Conformance: llm.VerdictNotApplicable, Reasoning: "no media present"},
		{ID: "1.4.1", Conformance: llm.VerdictNotEvaluated, EvidenceGap: "screenshot for color use"},
	}
	got := ApplyVerdicts(base, verdicts)

	find := func(id string) CriterionResult {
		for _, c := range got {
			if c.ID == id {
				return c
			}
		}
		return CriterionResult{}
	}
	if c := find("1.2.1"); c.Conformance != llm.VerdictNotApplicable || !c.Evaluated {
		t.Errorf("1.2.1 after verdict = %q evaluated=%v, want Not Applicable/true", c.Conformance, c.Evaluated)
	}
	// ...and an honest "Not Evaluated" verdict records the evidence gap.
	if c := find("1.4.1"); c.Conformance != llm.VerdictNotEvaluated || c.Evaluated {
		t.Errorf("1.4.1 after verdict = %q evaluated=%v, want Not Evaluated/false", c.Conformance, c.Evaluated)
	}
}

func TestVPATReportsFindingsAsNonConforming(t *testing.T) {
	result := &audit.AuditResult{
		TargetURL:  "https://example.com",
		StartTime:  time.Now(),
		WCAGLevel:  audit.WCAGLevelAA,
		LLMEnabled: false,
		Pages: []audit.PageResult{{
			URL: "https://example.com",
			Findings: []audit.Finding{{
				RuleID:          "keyboard-trap",
				Impact:          audit.ImpactCritical,
				SuccessCriteria: []string{"2.1.2"},
			}},
		}},
	}
	rep := generateVPATReport(result)

	if got := findConformance(rep, "2.1.2"); got != string(wcag.ConformanceDoesNotSupport) {
		t.Errorf("2.1.2 (critical finding) = %q, want Does Not Support", got)
	}
}
