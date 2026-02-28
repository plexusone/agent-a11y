package remediation

import (
	"strings"
	"testing"
)

func TestTechniqueURL(t *testing.T) {
	tests := []struct {
		id       TechniqueID
		expected string
	}{
		{TechniqueH37, "https://www.w3.org/WAI/WCAG22/Techniques/html/H37"},
		{TechniqueG94, "https://www.w3.org/WAI/WCAG22/Techniques/general/G94"},
		{TechniqueARIA6, "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA6"},
		{TechniqueC15, "https://www.w3.org/WAI/WCAG22/Techniques/css/C15"},
		{TechniqueF65, "https://www.w3.org/WAI/WCAG22/Techniques/failures/F65"},
		{TechniqueSCR2, "https://www.w3.org/WAI/WCAG22/Techniques/client-side-script/SCR2"},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			got := TechniqueURL(tt.id)
			if got != tt.expected {
				t.Errorf("TechniqueURL(%s) = %s, want %s", tt.id, got, tt.expected)
			}
		})
	}
}

func TestUnderstandingURL(t *testing.T) {
	tests := []struct {
		criterion string
		expected  string
	}{
		{"1.1.1", "https://www.w3.org/WAI/WCAG22/Understanding/non-text-content.html"},
		{"2.4.7", "https://www.w3.org/WAI/WCAG22/Understanding/focus-visible.html"},
		{"1.4.3", "https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html"},
		{"2.1.2", "https://www.w3.org/WAI/WCAG22/Understanding/no-keyboard-trap.html"},
	}

	for _, tt := range tests {
		t.Run(tt.criterion, func(t *testing.T) {
			got := UnderstandingURL(tt.criterion)
			if got != tt.expected {
				t.Errorf("UnderstandingURL(%s) = %s, want %s", tt.criterion, got, tt.expected)
			}
		})
	}
}

func TestACTRuleURL(t *testing.T) {
	tests := []struct {
		id       ACTRuleID
		expected string
	}{
		{ACTImageHasAccessibleName, "https://www.w3.org/WAI/standards-guidelines/act/rules/23a2a8/"},
		{ACTLinkHasAccessibleName, "https://www.w3.org/WAI/standards-guidelines/act/rules/c487ae/"},
		{ACTPageHasTitle, "https://www.w3.org/WAI/standards-guidelines/act/rules/2779a5/"},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			got := ACTRuleURL(tt.id)
			if got != tt.expected {
				t.Errorf("ACTRuleURL(%s) = %s, want %s", tt.id, got, tt.expected)
			}
		})
	}
}

func TestAxeRuleURL(t *testing.T) {
	url := AxeRuleURL("image-alt")
	expected := "https://dequeuniversity.com/rules/axe/4.10/image-alt"
	if url != expected {
		t.Errorf("AxeRuleURL(image-alt) = %s, want %s", url, expected)
	}
}

func TestGetTechnique(t *testing.T) {
	tech, ok := GetTechnique(TechniqueH37)
	if !ok {
		t.Fatal("GetTechnique(H37) returned not found")
	}
	if tech.ID != TechniqueH37 {
		t.Errorf("tech.ID = %s, want %s", tech.ID, TechniqueH37)
	}
	if tech.Category != CategoryHTML {
		t.Errorf("tech.Category = %s, want %s", tech.Category, CategoryHTML)
	}
	if !strings.Contains(tech.Title, "alt") {
		t.Errorf("tech.Title should contain 'alt', got %s", tech.Title)
	}
}

func TestGetACTRule(t *testing.T) {
	rule, ok := GetACTRule(ACTImageHasAccessibleName)
	if !ok {
		t.Fatal("GetACTRule(23a2a8) returned not found")
	}
	if rule.ID != ACTImageHasAccessibleName {
		t.Errorf("rule.ID = %s, want %s", rule.ID, ACTImageHasAccessibleName)
	}
	if len(rule.Criteria) == 0 {
		t.Error("rule.Criteria should not be empty")
	}
}

func TestTechniquesForCriterion(t *testing.T) {
	techs := TechniquesForCriterion("1.1.1")
	if len(techs) == 0 {
		t.Fatal("TechniquesForCriterion(1.1.1) returned no techniques")
	}

	// Should include H37
	found := false
	for _, tech := range techs {
		if tech.ID == TechniqueH37 {
			found = true
			break
		}
	}
	if !found {
		t.Error("TechniquesForCriterion(1.1.1) should include H37")
	}
}

func TestSufficientTechniques(t *testing.T) {
	techs := SufficientTechniques("1.1.1")
	if len(techs) == 0 {
		t.Fatal("SufficientTechniques(1.1.1) returned no techniques")
	}

	for _, tech := range techs {
		if tech.Type != TechniqueTypeSufficient {
			t.Errorf("SufficientTechniques returned non-sufficient technique: %s", tech.ID)
		}
	}
}

func TestFailureTechniques(t *testing.T) {
	techs := FailureTechniques("1.1.1")
	if len(techs) == 0 {
		t.Fatal("FailureTechniques(1.1.1) returned no techniques")
	}

	for _, tech := range techs {
		if tech.Type != TechniqueTypeFailure {
			t.Errorf("FailureTechniques returned non-failure technique: %s", tech.ID)
		}
	}
}

func TestBuildReference(t *testing.T) {
	ref := BuildReference(
		"1.1.1",
		[]TechniqueID{TechniqueH37, TechniqueG94},
		ACTImageHasAccessibleName,
		"image-alt",
	)

	if ref.CriterionID != "1.1.1" {
		t.Errorf("ref.CriterionID = %s, want 1.1.1", ref.CriterionID)
	}
	if !strings.Contains(ref.UnderstandingURL, "non-text-content") {
		t.Errorf("ref.UnderstandingURL should contain 'non-text-content'")
	}
	if len(ref.TechniqueURLs) != 2 {
		t.Errorf("ref.TechniqueURLs length = %d, want 2", len(ref.TechniqueURLs))
	}
	if ref.ACTRuleID != ACTImageHasAccessibleName {
		t.Errorf("ref.ACTRuleID = %s, want %s", ref.ACTRuleID, ACTImageHasAccessibleName)
	}
	if !strings.Contains(ref.AxeRuleURL, "image-alt") {
		t.Errorf("ref.AxeRuleURL should contain 'image-alt'")
	}
}

func TestACTRulesForCriterion(t *testing.T) {
	rules := ACTRulesForCriterion("1.1.1")
	if len(rules) == 0 {
		t.Fatal("ACTRulesForCriterion(1.1.1) returned no rules")
	}

	// Should include image-alt rule
	found := false
	for _, rule := range rules {
		if rule.ID == ACTImageHasAccessibleName {
			found = true
			break
		}
	}
	if !found {
		t.Error("ACTRulesForCriterion(1.1.1) should include 23a2a8")
	}
}
