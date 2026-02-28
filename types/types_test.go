package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSeverityConstants(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityCritical, "critical"},
		{SeveritySerious, "serious"},
		{SeverityModerate, "moderate"},
		{SeverityMinor, "minor"},
	}

	for _, tt := range tests {
		if string(tt.severity) != tt.expected {
			t.Errorf("Severity %v: got %q, want %q", tt.severity, string(tt.severity), tt.expected)
		}
	}
}

func TestImpactConstants(t *testing.T) {
	tests := []struct {
		impact   Impact
		expected string
	}{
		{ImpactBlocker, "blocker"},
		{ImpactCritical, "critical"},
		{ImpactSerious, "serious"},
		{ImpactModerate, "moderate"},
		{ImpactMinor, "minor"},
	}

	for _, tt := range tests {
		if string(tt.impact) != tt.expected {
			t.Errorf("Impact %v: got %q, want %q", tt.impact, string(tt.impact), tt.expected)
		}
	}
}

func TestWCAGLevelConstants(t *testing.T) {
	tests := []struct {
		level    WCAGLevel
		expected string
	}{
		{WCAGLevelA, "A"},
		{WCAGLevelAA, "AA"},
		{WCAGLevelAAA, "AAA"},
	}

	for _, tt := range tests {
		if string(tt.level) != tt.expected {
			t.Errorf("WCAGLevel %v: got %q, want %q", tt.level, string(tt.level), tt.expected)
		}
	}
}

func TestWCAGVersionConstants(t *testing.T) {
	tests := []struct {
		version  WCAGVersion
		expected string
	}{
		{WCAG20, "2.0"},
		{WCAG21, "2.1"},
		{WCAG22, "2.2"},
	}

	for _, tt := range tests {
		if string(tt.version) != tt.expected {
			t.Errorf("WCAGVersion %v: got %q, want %q", tt.version, string(tt.version), tt.expected)
		}
	}
}

func TestFindingJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	finding := Finding{
		ID:              "finding-1",
		RuleID:          "image-alt",
		Description:     "Image missing alt text",
		Help:            "Add alt attribute to img element",
		SuccessCriteria: []string{"1.1.1"},
		Level:           WCAGLevelA,
		Impact:          ImpactCritical,
		Selector:        "img.hero",
		XPath:           "/html/body/img[1]",
		HTML:            `<img src="hero.jpg">`,
		Element:         "img",
		PageURL:         "https://example.com",
		PageTitle:       "Example Page",
		Severity:        SeverityCritical,
		FoundAt:         now,
	}

	// Marshal to JSON
	data, err := json.Marshal(finding)
	if err != nil {
		t.Fatalf("Failed to marshal Finding: %v", err)
	}

	// Unmarshal back
	var decoded Finding
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Finding: %v", err)
	}

	// Verify fields
	if decoded.ID != finding.ID {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, finding.ID)
	}
	if decoded.RuleID != finding.RuleID {
		t.Errorf("RuleID mismatch: got %q, want %q", decoded.RuleID, finding.RuleID)
	}
	if decoded.Level != finding.Level {
		t.Errorf("Level mismatch: got %q, want %q", decoded.Level, finding.Level)
	}
	if decoded.Impact != finding.Impact {
		t.Errorf("Impact mismatch: got %q, want %q", decoded.Impact, finding.Impact)
	}
	if len(decoded.SuccessCriteria) != 1 || decoded.SuccessCriteria[0] != "1.1.1" {
		t.Errorf("SuccessCriteria mismatch: got %v, want [1.1.1]", decoded.SuccessCriteria)
	}
}

func TestFindingWithLLMEvaluation(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	finding := Finding{
		ID:     "finding-1",
		RuleID: "image-alt",
		LLMEvaluation: &LLMEvaluation{
			Confirmed:         true,
			Confidence:        0.95,
			Reasoning:         "Image lacks alternative text",
			Severity:          "critical",
			Remediation:       "Add descriptive alt attribute",
			NeedsManualReview: false,
			Model:             "claude-sonnet-4-20250514",
			EvalTime:          now,
			TokensIn:          100,
			TokensOut:         50,
		},
		FoundAt: now,
	}

	data, err := json.Marshal(finding)
	if err != nil {
		t.Fatalf("Failed to marshal Finding with LLMEvaluation: %v", err)
	}

	var decoded Finding
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Finding: %v", err)
	}

	if decoded.LLMEvaluation == nil {
		t.Fatal("LLMEvaluation should not be nil")
	}
	if !decoded.LLMEvaluation.Confirmed {
		t.Error("LLMEvaluation.Confirmed should be true")
	}
	if decoded.LLMEvaluation.Confidence != 0.95 {
		t.Errorf("Confidence mismatch: got %v, want 0.95", decoded.LLMEvaluation.Confidence)
	}
	if decoded.LLMEvaluation.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model mismatch: got %q", decoded.LLMEvaluation.Model)
	}
}

func TestLLMEvaluationJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	eval := LLMEvaluation{
		Confirmed:         true,
		Confidence:        0.87,
		Reasoning:         "The issue is confirmed",
		Severity:          "serious",
		Remediation:       "Fix by adding labels",
		NeedsManualReview: true,
		ReviewGuidance:    "Check if label is semantically correct",
		Model:             "gpt-4o",
		EvalTime:          now,
		TokensIn:          200,
		TokensOut:         100,
	}

	data, err := json.Marshal(eval)
	if err != nil {
		t.Fatalf("Failed to marshal LLMEvaluation: %v", err)
	}

	var decoded LLMEvaluation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal LLMEvaluation: %v", err)
	}

	if decoded.Confirmed != eval.Confirmed {
		t.Errorf("Confirmed mismatch")
	}
	if decoded.Confidence != eval.Confidence {
		t.Errorf("Confidence mismatch: got %v, want %v", decoded.Confidence, eval.Confidence)
	}
	if decoded.NeedsManualReview != eval.NeedsManualReview {
		t.Errorf("NeedsManualReview mismatch")
	}
	if decoded.ReviewGuidance != eval.ReviewGuidance {
		t.Errorf("ReviewGuidance mismatch: got %q, want %q", decoded.ReviewGuidance, eval.ReviewGuidance)
	}
	if decoded.TokensIn != eval.TokensIn {
		t.Errorf("TokensIn mismatch: got %d, want %d", decoded.TokensIn, eval.TokensIn)
	}
}

func TestFindingOmitsNilLLMEvaluation(t *testing.T) {
	finding := Finding{
		ID:            "finding-1",
		RuleID:        "test-rule",
		LLMEvaluation: nil,
	}

	data, err := json.Marshal(finding)
	if err != nil {
		t.Fatalf("Failed to marshal Finding: %v", err)
	}

	// Check that llmEvaluation is omitted
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, exists := raw["llmEvaluation"]; exists {
		t.Error("llmEvaluation should be omitted when nil")
	}
}
