package llm

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDefaultJudgeConfig(t *testing.T) {
	cfg := DefaultJudgeConfig()

	if cfg.ConfidenceThreshold != 0.8 {
		t.Errorf("ConfidenceThreshold: got %v, want 0.8", cfg.ConfidenceThreshold)
	}

	if cfg.Concurrency != 5 {
		t.Errorf("Concurrency: got %d, want 5", cfg.Concurrency)
	}

	expectedCategories := []string{
		"alternative-text",
		"color-contrast",
		"keyboard-access",
		"form-labels",
		"link-purpose",
	}

	if len(cfg.Categories) != len(expectedCategories) {
		t.Errorf("Categories count: got %d, want %d", len(cfg.Categories), len(expectedCategories))
	}

	for i, expected := range expectedCategories {
		if i < len(cfg.Categories) && cfg.Categories[i] != expected {
			t.Errorf("Categories[%d]: got %q, want %q", i, cfg.Categories[i], expected)
		}
	}
}

func TestMatchesCategory(t *testing.T) {
	tests := []struct {
		criterion string
		category  string
		expected  bool
	}{
		// alternative-text
		{"1.1.1", "alternative-text", true},
		{"1.1.2", "alternative-text", false},

		// color-contrast
		{"1.4.3", "color-contrast", true},
		{"1.4.6", "color-contrast", true},
		{"1.4.11", "color-contrast", true},
		{"1.4.1", "color-contrast", false},

		// keyboard-access
		{"2.1.1", "keyboard-access", true},
		{"2.1.2", "keyboard-access", true},
		{"2.1.4", "keyboard-access", true},
		{"2.4.3", "keyboard-access", true},
		{"2.4.7", "keyboard-access", true},
		{"2.2.1", "keyboard-access", false},

		// form-labels
		{"1.3.1", "form-labels", true},
		{"3.3.2", "form-labels", true},
		{"4.1.2", "form-labels", true},
		{"3.3.1", "form-labels", false},

		// link-purpose
		{"2.4.4", "link-purpose", true},
		{"2.4.9", "link-purpose", true},
		{"2.4.1", "link-purpose", false},

		// Unknown category
		{"1.1.1", "unknown-category", false},
	}

	for _, tt := range tests {
		result := matchesCategory(tt.criterion, tt.category)
		if result != tt.expected {
			t.Errorf("matchesCategory(%q, %q): got %v, want %v",
				tt.criterion, tt.category, result, tt.expected)
		}
	}
}

func TestShouldEvaluate(t *testing.T) {
	cfg := DefaultJudgeConfig()
	// NewJudge requires omnillm client, but we can test ShouldEvaluate logic
	// by creating a Judge with nil client (won't be used for this test)
	judge := &Judge{
		config: cfg,
	}

	tests := []struct {
		name     string
		finding  Finding
		expected bool
	}{
		{
			name: "matches alternative-text",
			finding: Finding{
				ID:              "f1",
				SuccessCriteria: []string{"1.1.1"},
			},
			expected: true,
		},
		{
			name: "matches color-contrast",
			finding: Finding{
				ID:              "f2",
				SuccessCriteria: []string{"1.4.3"},
			},
			expected: true,
		},
		{
			name: "matches keyboard-access",
			finding: Finding{
				ID:              "f3",
				SuccessCriteria: []string{"2.1.1"},
			},
			expected: true,
		},
		{
			name: "no matching category",
			finding: Finding{
				ID:              "f4",
				SuccessCriteria: []string{"2.2.1"},
			},
			expected: false,
		},
		{
			name: "multiple criteria with one match",
			finding: Finding{
				ID:              "f5",
				SuccessCriteria: []string{"2.2.1", "1.1.1"},
			},
			expected: true,
		},
		{
			name: "empty criteria",
			finding: Finding{
				ID:              "f6",
				SuccessCriteria: []string{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := judge.ShouldEvaluate(tt.finding)
			if result != tt.expected {
				t.Errorf("ShouldEvaluate: got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReviewCategories(t *testing.T) {
	cfg := JudgeConfig{
		Categories: []string{"cat1", "cat2", "cat3"},
	}
	judge := &Judge{config: cfg}

	categories := judge.ReviewCategories()
	if len(categories) != 3 {
		t.Errorf("ReviewCategories: got %d categories, want 3", len(categories))
	}
	if categories[0] != "cat1" {
		t.Errorf("ReviewCategories[0]: got %q, want %q", categories[0], "cat1")
	}
}

func TestBuildEvaluationPrompt(t *testing.T) {
	judge := &Judge{}

	finding := Finding{
		ID:              "finding-1",
		RuleID:          "image-alt",
		Description:     "Image missing alternative text",
		SuccessCriteria: []string{"1.1.1"},
		Level:           "A",
		Impact:          "critical",
		Selector:        "img.hero",
		HTML:            `<img src="hero.jpg">`,
		Help:            "Add alt attribute to describe the image",
	}

	pageContext := PageContext{
		URL:       "https://example.com",
		Title:     "Example Page",
		IsSPA:     true,
		Framework: "react",
		Language:  "en",
	}

	prompt := judge.buildEvaluationPrompt(finding, pageContext)

	// Check for expected content
	expectedStrings := []string{
		"## Accessibility Finding to Evaluate",
		"**Rule ID:** image-alt",
		"**Description:** Image missing alternative text",
		"**WCAG Criteria:** 1.1.1",
		"**Level:** A",
		"**Impact:** critical",
		"**Element Selector:** `img.hero`",
		`<img src="hero.jpg">`,
		"**Help Text:** Add alt attribute",
		"## Page Context",
		"**URL:** https://example.com",
		"**Title:** Example Page",
		"**SPA Framework:** react",
		"**Language:** en",
		"## Evaluation Request",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(prompt, expected) {
			t.Errorf("Prompt missing: %q", expected)
		}
	}
}

func TestBuildEvaluationPromptMinimal(t *testing.T) {
	judge := &Judge{}

	finding := Finding{
		ID:              "finding-1",
		RuleID:          "test-rule",
		Description:     "Test description",
		SuccessCriteria: []string{},
		Level:           "A",
		Impact:          "minor",
		// No selector, HTML, or help
	}

	pageContext := PageContext{
		URL:      "https://example.com",
		Title:    "Test",
		IsSPA:    false,
		Language: "en",
	}

	prompt := judge.buildEvaluationPrompt(finding, pageContext)

	// Should still have basic structure
	if !strings.Contains(prompt, "## Accessibility Finding to Evaluate") {
		t.Error("Prompt should contain finding header")
	}
	if !strings.Contains(prompt, "## Page Context") {
		t.Error("Prompt should contain page context header")
	}
	// Should NOT contain SPA Framework when not SPA
	if strings.Contains(prompt, "SPA Framework") {
		t.Error("Prompt should not contain SPA Framework when IsSPA is false")
	}
}

func TestGetSystemPrompt(t *testing.T) {
	// Test default system prompt
	judge := &Judge{
		config: JudgeConfig{},
	}

	prompt := judge.getSystemPrompt()
	if prompt == "" {
		t.Error("Default system prompt should not be empty")
	}
	if !strings.Contains(prompt, "accessibility auditor") {
		t.Error("System prompt should mention accessibility auditor")
	}
	if !strings.Contains(prompt, "WCAG") {
		t.Error("System prompt should mention WCAG")
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("System prompt should mention JSON response format")
	}

	// Test custom system prompt
	customPrompt := "Custom evaluation instructions"
	judgeCustom := &Judge{
		config: JudgeConfig{
			SystemPrompt: customPrompt,
		},
	}

	if judgeCustom.getSystemPrompt() != customPrompt {
		t.Errorf("Should use custom system prompt: got %q", judgeCustom.getSystemPrompt())
	}
}

func TestParseEvaluation(t *testing.T) {
	judge := &Judge{}

	tests := []struct {
		name      string
		content   string
		findingID string
		wantErr   bool
		check     func(*Evaluation) error
	}{
		{
			name: "valid JSON",
			content: `{
				"confirmed": true,
				"confidence": 0.95,
				"severity": "critical",
				"reasoning": "Image lacks alt text",
				"remediation": "Add descriptive alt attribute",
				"needsManualReview": false
			}`,
			findingID: "f1",
			wantErr:   false,
			check: func(e *Evaluation) error {
				if !e.Confirmed {
					return errFromString("Confirmed should be true")
				}
				if e.Confidence != 0.95 {
					return errFromString("Confidence mismatch")
				}
				if e.FindingID != "f1" {
					return errFromString("FindingID mismatch")
				}
				return nil
			},
		},
		{
			name: "JSON with markdown code block",
			content: "```json\n{\"confirmed\": true, \"confidence\": 0.8, \"severity\": \"moderate\", \"reasoning\": \"test\", \"remediation\": \"fix it\", \"needsManualReview\": false}\n```",
			findingID: "f2",
			wantErr:   false,
			check: func(e *Evaluation) error {
				if !e.Confirmed {
					return errFromString("Confirmed should be true")
				}
				return nil
			},
		},
		{
			name: "JSON with plain code block",
			content: "```\n{\"confirmed\": false, \"confidence\": 0.6, \"severity\": \"minor\", \"reasoning\": \"not an issue\", \"remediation\": \"none\", \"needsManualReview\": true}\n```",
			findingID: "f3",
			wantErr:   false,
			check: func(e *Evaluation) error {
				if e.Confirmed {
					return errFromString("Confirmed should be false")
				}
				if !e.NeedsManualReview {
					return errFromString("NeedsManualReview should be true")
				}
				return nil
			},
		},
		{
			name:      "invalid JSON",
			content:   `{invalid json}`,
			findingID: "f4",
			wantErr:   true,
		},
		{
			name:      "empty content",
			content:   "",
			findingID: "f5",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eval, err := judge.parseEvaluation(tt.content, tt.findingID)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if tt.check != nil {
				if checkErr := tt.check(eval); checkErr != nil {
					t.Errorf("Check failed: %v", checkErr)
				}
			}
		})
	}
}

// Helper to create error from string
type stringError string

func (e stringError) Error() string { return string(e) }

func errFromString(s string) error { return stringError(s) }

func TestEvaluationJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	eval := Evaluation{
		FindingID:         "finding-1",
		Confirmed:         true,
		Confidence:        0.92,
		Severity:          "serious",
		Reasoning:         "The element lacks proper labeling",
		Remediation:       "Add aria-label attribute",
		NeedsManualReview: false,
		Model:             "claude-sonnet-4-20250514",
		EvalTime:          now,
		TokensIn:          150,
		TokensOut:         75,
	}

	data, err := json.Marshal(eval)
	if err != nil {
		t.Fatalf("Failed to marshal Evaluation: %v", err)
	}

	var decoded Evaluation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Evaluation: %v", err)
	}

	if decoded.FindingID != eval.FindingID {
		t.Errorf("FindingID mismatch")
	}
	if decoded.Confirmed != eval.Confirmed {
		t.Errorf("Confirmed mismatch")
	}
	if decoded.Confidence != eval.Confidence {
		t.Errorf("Confidence mismatch")
	}
	if decoded.TokensIn != eval.TokensIn {
		t.Errorf("TokensIn mismatch")
	}
}

func TestFindingType(t *testing.T) {
	finding := Finding{
		ID:              "test-finding",
		RuleID:          "test-rule",
		Description:     "Test description",
		SuccessCriteria: []string{"1.1.1", "4.1.2"},
		Level:           "A",
		Impact:          "critical",
		Selector:        "#test",
		HTML:            "<div id='test'>",
		Help:            "Test help",
	}

	// Verify JSON serialization
	data, err := json.Marshal(finding)
	if err != nil {
		t.Fatalf("Failed to marshal Finding: %v", err)
	}

	var decoded Finding
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Finding: %v", err)
	}

	if decoded.ID != finding.ID {
		t.Errorf("ID mismatch")
	}
	if len(decoded.SuccessCriteria) != 2 {
		t.Errorf("SuccessCriteria length mismatch: got %d", len(decoded.SuccessCriteria))
	}
}

func TestPageContextType(t *testing.T) {
	ctx := PageContext{
		URL:         "https://example.com",
		Title:       "Test Page",
		IsSPA:       true,
		Framework:   "react",
		Language:    "en",
		Screenshot:  "base64data",
		HTMLSnippet: "<html>...</html>",
	}

	data, err := json.Marshal(ctx)
	if err != nil {
		t.Fatalf("Failed to marshal PageContext: %v", err)
	}

	var decoded PageContext
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PageContext: %v", err)
	}

	if decoded.URL != ctx.URL {
		t.Errorf("URL mismatch")
	}
	if decoded.IsSPA != ctx.IsSPA {
		t.Errorf("IsSPA mismatch")
	}
	if decoded.Framework != ctx.Framework {
		t.Errorf("Framework mismatch")
	}
}
