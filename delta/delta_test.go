package delta

import (
	"testing"
	"time"

	"github.com/plexusone/agent-a11y/types"
)

func TestCompare_AllFixed(t *testing.T) {
	before := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("color-contrast", "button.primary", "<button class='primary'>"),
			makeFinding("image-alt", "img.hero", "<img src='hero.png'>"),
		},
	}

	after := &types.AgentResult{
		URL:      "https://example.com",
		Findings: []types.AgentFinding{},
	}

	calc := NewCalculator()
	delta := calc.Compare(before, after)

	if delta.Status != types.DeltaStatusFixed {
		t.Errorf("expected FIXED, got %s", delta.Status)
	}

	if len(delta.Fixed) != 2 {
		t.Errorf("expected 2 fixed, got %d", len(delta.Fixed))
	}

	if len(delta.Regressions) != 0 {
		t.Errorf("expected 0 regressions, got %d", len(delta.Regressions))
	}

	if delta.Improvement != 100 {
		t.Errorf("expected 100%% improvement, got %.1f%%", delta.Improvement)
	}
}

func TestCompare_SomeFixed(t *testing.T) {
	before := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("color-contrast", "button.primary", "<button class='primary'>"),
			makeFinding("image-alt", "img.hero", "<img src='hero.png'>"),
		},
	}

	after := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("image-alt", "img.hero", "<img src='hero.png'>"), // Still present
		},
	}

	calc := NewCalculator()
	delta := calc.Compare(before, after)

	if delta.Status != types.DeltaStatusImproved {
		t.Errorf("expected IMPROVED, got %s", delta.Status)
	}

	if len(delta.Fixed) != 1 {
		t.Errorf("expected 1 fixed, got %d", len(delta.Fixed))
	}

	if len(delta.Remaining) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(delta.Remaining))
	}

	if delta.Improvement != 50 {
		t.Errorf("expected 50%% improvement, got %.1f%%", delta.Improvement)
	}
}

func TestCompare_Regression(t *testing.T) {
	before := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("color-contrast", "button.primary", "<button class='primary'>"),
		},
	}

	after := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("color-contrast", "button.primary", "<button class='primary'>"), // Still present
			makeFinding("link-name", "a.nav", "<a href='/about'>"),                       // New issue!
		},
	}

	calc := NewCalculator()
	delta := calc.Compare(before, after)

	if delta.Status != types.DeltaStatusRegressed {
		t.Errorf("expected REGRESSED, got %s", delta.Status)
	}

	if len(delta.Regressions) != 1 {
		t.Errorf("expected 1 regression, got %d", len(delta.Regressions))
	}

	if !delta.HasRegressions() {
		t.Error("expected HasRegressions() to return true")
	}

	if delta.IsPassing() {
		t.Error("expected IsPassing() to return false")
	}
}

func TestCompare_Mixed(t *testing.T) {
	before := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("color-contrast", "button.primary", "<button class='primary'>"),
			makeFinding("image-alt", "img.hero", "<img src='hero.png'>"),
		},
	}

	after := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("image-alt", "img.hero", "<img src='hero.png'>"), // Still present
			makeFinding("link-name", "a.nav", "<a href='/about'>"),       // New issue!
		},
	}

	calc := NewCalculator()
	delta := calc.Compare(before, after)

	if delta.Status != types.DeltaStatusMixed {
		t.Errorf("expected MIXED, got %s", delta.Status)
	}

	if len(delta.Fixed) != 1 {
		t.Errorf("expected 1 fixed, got %d", len(delta.Fixed))
	}

	if len(delta.Regressions) != 1 {
		t.Errorf("expected 1 regression, got %d", len(delta.Regressions))
	}
}

func TestCompare_NoChange(t *testing.T) {
	before := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("color-contrast", "button.primary", "<button class='primary'>"),
		},
	}

	after := &types.AgentResult{
		URL: "https://example.com",
		Findings: []types.AgentFinding{
			makeFinding("color-contrast", "button.primary", "<button class='primary'>"),
		},
	}

	calc := NewCalculator()
	delta := calc.Compare(before, after)

	if delta.Status != types.DeltaStatusNoChange {
		t.Errorf("expected NO_CHANGE, got %s", delta.Status)
	}

	if len(delta.Remaining) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(delta.Remaining))
	}
}

func TestFingerprint_Stable(t *testing.T) {
	f1 := types.Finding{
		RuleID:   "color-contrast",
		Selector: "button.primary",
		HTML:     "<button class='primary'>Click</button>",
		PageURL:  "https://example.com",
	}

	f2 := types.Finding{
		RuleID:   "color-contrast",
		Selector: "button.primary",
		HTML:     "<button class='primary'>Click</button>",
		PageURL:  "https://example.com",
	}

	if f1.Fingerprint() != f2.Fingerprint() {
		t.Error("identical findings should have same fingerprint")
	}
}

func TestFingerprint_IgnoresDataAttrs(t *testing.T) {
	f1 := types.Finding{
		RuleID:   "color-contrast",
		Selector: "button.primary",
		HTML:     "<button class='primary' data-testid='btn'>Click</button>",
		PageURL:  "https://example.com",
	}

	f2 := types.Finding{
		RuleID:   "color-contrast",
		Selector: "button.primary",
		HTML:     "<button class='primary' data-testid='other'>Click</button>",
		PageURL:  "https://example.com",
	}

	if f1.Fingerprint() != f2.Fingerprint() {
		t.Error("data-* attributes should be ignored in fingerprint")
	}
}

func TestDeltaSummary(t *testing.T) {
	delta := &types.ValidationDelta{
		BeforeTotal: 10,
		AfterTotal:  5,
		Fixed:       make([]types.FixedFinding, 6),
		Remaining:   make([]types.AgentFinding, 4),
		Regressions: make([]types.AgentFinding, 1),
		Status:      types.DeltaStatusMixed,
		Improvement: 50,
	}

	summary := delta.Summary()

	if summary.TotalBefore != 10 {
		t.Errorf("expected TotalBefore 10, got %d", summary.TotalBefore)
	}

	if summary.FixedCount != 6 {
		t.Errorf("expected FixedCount 6, got %d", summary.FixedCount)
	}

	if summary.NewCount != 1 {
		t.Errorf("expected NewCount 1, got %d", summary.NewCount)
	}

	if summary.StatusEmoji != "⚠️" {
		t.Errorf("expected ⚠️ emoji, got %s", summary.StatusEmoji)
	}
}

// Helper to create test findings
func makeFinding(ruleID, selector, html string) types.AgentFinding {
	return types.AgentFinding{
		Finding: types.Finding{
			RuleID:   ruleID,
			Selector: selector,
			HTML:     html,
			PageURL:  "https://example.com",
			FoundAt:  time.Now(),
		},
	}
}
