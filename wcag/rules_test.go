package wcag

import (
	"context"
	"testing"

	"github.com/plexusone/agent-a11y/types"
	vibium "github.com/plexusone/vibium-go"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry(nil)
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	// Should have rules registered
	rules := registry.All()
	if len(rules) == 0 {
		t.Error("Registry should have builtin rules")
	}
}

func TestRegistryGet(t *testing.T) {
	registry := NewRegistry(nil)

	// Test getting an existing rule
	rule, ok := registry.Get("image-alt")
	if !ok {
		t.Error("Expected to find image-alt rule")
	}
	if rule == nil {
		t.Error("Rule should not be nil")
	}
	if rule.ID() != "image-alt" {
		t.Errorf("Rule ID: got %q, want %q", rule.ID(), "image-alt")
	}

	// Test getting a non-existent rule
	_, ok = registry.Get("non-existent-rule")
	if ok {
		t.Error("Should not find non-existent rule")
	}
}

func TestRegistryGetByLevel(t *testing.T) {
	registry := NewRegistry(nil)

	// Level A rules
	levelARules := registry.GetByLevel(types.WCAGLevelA)
	if len(levelARules) == 0 {
		t.Error("Should have Level A rules")
	}

	// All Level A rules should have Level A or lower
	for _, rule := range levelARules {
		if rule.Level() != types.WCAGLevelA {
			t.Errorf("Rule %s has level %s, expected A", rule.ID(), rule.Level())
		}
	}

	// Level AA should include A and AA rules
	levelAARules := registry.GetByLevel(types.WCAGLevelAA)
	if len(levelAARules) < len(levelARules) {
		t.Error("Level AA should include at least all Level A rules")
	}

	// Level AAA should include all rules
	levelAAARules := registry.GetByLevel(types.WCAGLevelAAA)
	if len(levelAAARules) < len(levelAARules) {
		t.Error("Level AAA should include at least all Level AA rules")
	}
}

func TestRegistryGetByCriterion(t *testing.T) {
	registry := NewRegistry(nil)

	// Get rules for WCAG 1.1.1 (Non-text Content)
	rules := registry.GetByCriterion("1.1.1")
	if len(rules) == 0 {
		t.Error("Should have rules for criterion 1.1.1")
	}

	// Verify all returned rules have the criterion
	for _, rule := range rules {
		found := false
		for _, sc := range rule.SuccessCriteria() {
			if sc == "1.1.1" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Rule %s should have criterion 1.1.1", rule.ID())
		}
	}

	// Non-existent criterion
	emptyRules := registry.GetByCriterion("99.99.99")
	if len(emptyRules) != 0 {
		t.Error("Should not have rules for non-existent criterion")
	}
}

func TestRegistryAll(t *testing.T) {
	registry := NewRegistry(nil)
	rules := registry.All()

	if len(rules) == 0 {
		t.Fatal("Registry should have rules")
	}

	// Check that each rule has required metadata
	for _, rule := range rules {
		if rule.ID() == "" {
			t.Error("Rule should have ID")
		}
		if rule.Name() == "" {
			t.Errorf("Rule %s should have name", rule.ID())
		}
		if rule.Description() == "" {
			t.Errorf("Rule %s should have description", rule.ID())
		}
		if len(rule.SuccessCriteria()) == 0 {
			t.Errorf("Rule %s should have success criteria", rule.ID())
		}
		if rule.Level() == "" {
			t.Errorf("Rule %s should have level", rule.ID())
		}
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry(nil)
	initialCount := len(registry.All())

	// Create a mock rule
	mockRule := &mockRule{
		id:              "mock-rule",
		name:            "Mock Rule",
		description:     "A test rule",
		successCriteria: []string{"1.1.1"},
		level:           types.WCAGLevelA,
	}

	registry.Register(mockRule)

	// Should have one more rule
	if len(registry.All()) != initialCount+1 {
		t.Errorf("Expected %d rules, got %d", initialCount+1, len(registry.All()))
	}

	// Should be able to retrieve it
	rule, ok := registry.Get("mock-rule")
	if !ok {
		t.Error("Should find mock-rule")
	}
	if rule.ID() != "mock-rule" {
		t.Errorf("ID mismatch: got %q", rule.ID())
	}
}

func TestIsLowerLevel(t *testing.T) {
	tests := []struct {
		l1       types.WCAGLevel
		l2       types.WCAGLevel
		expected bool
	}{
		{types.WCAGLevelA, types.WCAGLevelAA, true},
		{types.WCAGLevelA, types.WCAGLevelAAA, true},
		{types.WCAGLevelAA, types.WCAGLevelAAA, true},
		{types.WCAGLevelAA, types.WCAGLevelA, false},
		{types.WCAGLevelAAA, types.WCAGLevelA, false},
		{types.WCAGLevelAAA, types.WCAGLevelAA, false},
		{types.WCAGLevelA, types.WCAGLevelA, false},
		{types.WCAGLevelAA, types.WCAGLevelAA, false},
		{types.WCAGLevelAAA, types.WCAGLevelAAA, false},
	}

	for _, tt := range tests {
		result := isLowerLevel(tt.l1, tt.l2)
		if result != tt.expected {
			t.Errorf("isLowerLevel(%s, %s): got %v, want %v",
				tt.l1, tt.l2, result, tt.expected)
		}
	}
}

func TestImageAltRuleMetadata(t *testing.T) {
	rule := &ImageAltRule{}

	if rule.ID() != "image-alt" {
		t.Errorf("ID: got %q", rule.ID())
	}
	if rule.Name() == "" {
		t.Error("Name should not be empty")
	}
	if rule.Description() == "" {
		t.Error("Description should not be empty")
	}
	if len(rule.SuccessCriteria()) == 0 {
		t.Error("Should have success criteria")
	}
	if rule.Level() != types.WCAGLevelA {
		t.Errorf("Level: got %q, want A", rule.Level())
	}

	// Check that 1.1.1 is in success criteria
	found := false
	for _, sc := range rule.SuccessCriteria() {
		if sc == "1.1.1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should include 1.1.1 in success criteria")
	}
}

func TestFormLabelRuleMetadata(t *testing.T) {
	rule := &FormLabelRule{}

	if rule.ID() != "form-label" {
		t.Errorf("ID: got %q", rule.ID())
	}
	if rule.Level() != types.WCAGLevelA {
		t.Errorf("Level: got %q, want A", rule.Level())
	}

	// Should include relevant success criteria
	criteria := rule.SuccessCriteria()
	expectedCriteria := []string{"1.3.1", "4.1.2"}
	for _, expected := range expectedCriteria {
		found := false
		for _, sc := range criteria {
			if sc == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should include %s in success criteria", expected)
		}
	}
}

func TestPageTitleRuleMetadata(t *testing.T) {
	rule := &PageTitleRule{}

	if rule.ID() != "page-title" {
		t.Errorf("ID: got %q", rule.ID())
	}
	if rule.Level() != types.WCAGLevelA {
		t.Errorf("Level: got %q, want A", rule.Level())
	}

	found := false
	for _, sc := range rule.SuccessCriteria() {
		if sc == "2.4.2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should include 2.4.2 in success criteria")
	}
}

func TestContrastRuleMetadata(t *testing.T) {
	rule := &ContrastRule{}

	if rule.ID() != "color-contrast" {
		t.Errorf("ID: got %q", rule.ID())
	}
	if rule.Level() != types.WCAGLevelAA {
		t.Errorf("Level: got %q, want AA", rule.Level())
	}

	found := false
	for _, sc := range rule.SuccessCriteria() {
		if sc == "1.4.3" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should include 1.4.3 in success criteria")
	}
}

func TestLanguageRuleMetadata(t *testing.T) {
	rule := &LanguageRule{}

	if rule.ID() != "html-lang" {
		t.Errorf("ID: got %q", rule.ID())
	}
	if rule.Level() != types.WCAGLevelA {
		t.Errorf("Level: got %q, want A", rule.Level())
	}

	found := false
	for _, sc := range rule.SuccessCriteria() {
		if sc == "3.1.1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should include 3.1.1 in success criteria")
	}
}

func TestDuplicateIDRuleMetadata(t *testing.T) {
	rule := &DuplicateIDRule{}

	if rule.ID() != "duplicate-id" {
		t.Errorf("ID: got %q", rule.ID())
	}
	if rule.Level() != types.WCAGLevelA {
		t.Errorf("Level: got %q, want A", rule.Level())
	}

	found := false
	for _, sc := range rule.SuccessCriteria() {
		if sc == "4.1.1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should include 4.1.1 in success criteria")
	}
}

// Mock rule for testing
type mockRule struct {
	id              string
	name            string
	description     string
	successCriteria []string
	level           types.WCAGLevel
}

func (r *mockRule) ID() string                { return r.id }
func (r *mockRule) Name() string              { return r.name }
func (r *mockRule) Description() string       { return r.description }
func (r *mockRule) SuccessCriteria() []string { return r.successCriteria }
func (r *mockRule) Level() types.WCAGLevel    { return r.level }
func (r *mockRule) Run(_ context.Context, _ *vibium.Vibe) ([]types.Finding, error) {
	return nil, nil
}
