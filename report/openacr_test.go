package report

import (
	"bytes"
	"testing"
	"time"

	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/openacr-go"
)

func TestMapConformanceStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected openacr.AdherenceLevel
	}{
		{"supports", openacr.LevelSupports},
		{"Supports", openacr.LevelSupports},
		{"partially_supports", openacr.LevelPartiallySupports},
		{"Partially Supports", openacr.LevelPartiallySupports},
		{"partially-supports", openacr.LevelPartiallySupports},
		{"does_not_support", openacr.LevelDoesNotSupport},
		{"Does Not Support", openacr.LevelDoesNotSupport},
		{"does-not-support", openacr.LevelDoesNotSupport},
		{"not_applicable", openacr.LevelNotApplicable},
		{"Not Applicable", openacr.LevelNotApplicable},
		{"not-applicable", openacr.LevelNotApplicable},
		{"not_evaluated", openacr.LevelNotEvaluated},
		{"Not Evaluated", openacr.LevelNotEvaluated},
		{"not-evaluated", openacr.LevelNotEvaluated},
		{"unknown", openacr.LevelNotEvaluated},
		{"", openacr.LevelNotEvaluated},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapConformanceStatus(tt.input)
			if got != tt.expected {
				t.Errorf("mapConformanceStatus(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildCriterionNotes(t *testing.T) {
	tests := []struct {
		name     string
		cr       audit.CriterionResult
		expected string
	}{
		{
			name: "with remarks",
			cr: audit.CriterionResult{
				Remarks:    "Custom remarks",
				IssueCount: 5,
			},
			expected: "Custom remarks",
		},
		{
			name: "with issues no remarks",
			cr: audit.CriterionResult{
				IssueCount: 3,
			},
			expected: "3 issue(s) found during evaluation.",
		},
		{
			name: "no issues no remarks",
			cr: audit.CriterionResult{
				IssueCount: 0,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCriterionNotes(tt.cr)
			if got != tt.expected {
				t.Errorf("buildCriterionNotes() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBuildLevelNotes(t *testing.T) {
	tests := []struct {
		name     string
		level    audit.LevelConformance
		expected string
	}{
		{
			name: "no issues",
			level: audit.LevelConformance{
				TotalIssues:    0,
				BlockingIssues: 0,
			},
			expected: "No issues found at this level.",
		},
		{
			name: "issues with blocking",
			level: audit.LevelConformance{
				TotalIssues:    5,
				BlockingIssues: 2,
			},
			expected: "5 total issues, 2 blocking issues.",
		},
		{
			name: "issues without blocking",
			level: audit.LevelConformance{
				TotalIssues:    3,
				BlockingIssues: 0,
			},
			expected: "3 issue(s) found at this level.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildLevelNotes(tt.level)
			if got != tt.expected {
				t.Errorf("buildLevelNotes() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSelectCatalog(t *testing.T) {
	tests := []struct {
		version  audit.WCAGVersion
		expected string
	}{
		{audit.WCAG20, "2.5-edition-wcag-2.0-508-en"},
		{audit.WCAG21, "2.5-edition-wcag-2.1-508-en"},
		{audit.WCAG22, "2.5-edition-wcag-2.2-508-en"},
		{"unknown", DefaultOpenACRCatalog},
	}

	for _, tt := range tests {
		t.Run(string(tt.version), func(t *testing.T) {
			got := selectCatalog(tt.version)
			if got != tt.expected {
				t.Errorf("selectCatalog(%q) = %q, want %q", tt.version, got, tt.expected)
			}
		})
	}
}

func TestConvertToOpenACR(t *testing.T) {
	result := &audit.AuditResult{
		ID:          "test-audit-1",
		TargetURL:   "https://example.com",
		WCAGVersion: audit.WCAG22,
		WCAGLevel:   audit.WCAGLevelAA,
		StartTime:   time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		Conformance: audit.ConformanceSummary{
			TargetLevel: audit.WCAGLevelAA,
			Version:     "2.2",
			LevelA: audit.LevelConformance{
				Status:      "supports",
				TotalIssues: 0,
			},
			LevelAA: audit.LevelConformance{
				Status:      "partially_supports",
				TotalIssues: 3,
			},
			Criteria: []audit.CriterionResult{
				{
					ID:         "1.1.1",
					Name:       "Non-text Content",
					Level:      "A",
					Status:     "supports",
					IssueCount: 0,
				},
				{
					ID:         "1.4.3",
					Name:       "Contrast (Minimum)",
					Level:      "AA",
					Status:     "partially_supports",
					IssueCount: 3,
					Remarks:    "Some elements have insufficient contrast.",
				},
			},
		},
	}

	report := convertToOpenACR(result)

	// Check basic fields
	if report.Product.Name != "https://example.com" {
		t.Errorf("Product.Name = %q, want %q", report.Product.Name, "https://example.com")
	}

	if report.Catalog != "2.5-edition-wcag-2.2-508-en" {
		t.Errorf("Catalog = %q, want %q", report.Catalog, "2.5-edition-wcag-2.2-508-en")
	}

	if report.ReportDate != "2025-01-15" {
		t.Errorf("ReportDate = %q, want %q", report.ReportDate, "2025-01-15")
	}

	// Check chapters
	if len(report.Chapters) == 0 {
		t.Error("expected chapters to be populated")
	}

	// Check Level A chapter
	levelA, ok := report.Chapters["success_criteria_level_a"]
	if !ok {
		t.Error("expected success_criteria_level_a chapter")
	} else {
		if len(levelA.Criteria) != 1 {
			t.Errorf("Level A criteria count = %d, want 1", len(levelA.Criteria))
		}
		if levelA.Criteria[0].Num != "1.1.1" {
			t.Errorf("Level A criterion num = %q, want %q", levelA.Criteria[0].Num, "1.1.1")
		}
	}

	// Check Level AA chapter
	levelAA, ok := report.Chapters["success_criteria_level_aa"]
	if !ok {
		t.Error("expected success_criteria_level_aa chapter")
	} else {
		if len(levelAA.Criteria) != 1 {
			t.Errorf("Level AA criteria count = %d, want 1", len(levelAA.Criteria))
		}
		if levelAA.Criteria[0].Components[0].Adherence.Level != openacr.LevelPartiallySupports {
			t.Errorf("Level AA adherence = %q, want %q",
				levelAA.Criteria[0].Components[0].Adherence.Level, openacr.LevelPartiallySupports)
		}
	}

	// Check hardware/software chapters are disabled
	hw, ok := report.Chapters["hardware"]
	if !ok {
		t.Error("expected hardware chapter")
	} else if !hw.Disabled {
		t.Error("expected hardware chapter to be disabled")
	}

	sw, ok := report.Chapters["software"]
	if !ok {
		t.Error("expected software chapter")
	} else if !sw.Disabled {
		t.Error("expected software chapter to be disabled")
	}
}

func TestWriteOpenACR(t *testing.T) {
	result := &audit.AuditResult{
		ID:          "test-audit-1",
		TargetURL:   "https://example.com",
		WCAGVersion: audit.WCAG22,
		WCAGLevel:   audit.WCAGLevelAA,
		StartTime:   time.Now(),
		Conformance: audit.ConformanceSummary{
			Criteria: []audit.CriterionResult{
				{
					ID:     "1.1.1",
					Level:  "A",
					Status: "supports",
				},
			},
		},
	}

	var buf bytes.Buffer
	writer := NewWriter(FormatOpenACR)
	err := writer.Write(&buf, result)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()

	// Check YAML output contains expected fields
	if !bytes.Contains(buf.Bytes(), []byte("title:")) {
		t.Error("output missing 'title' field")
	}
	if !bytes.Contains(buf.Bytes(), []byte("product:")) {
		t.Error("output missing 'product' field")
	}
	if !bytes.Contains(buf.Bytes(), []byte("catalog:")) {
		t.Error("output missing 'catalog' field")
	}
	if !bytes.Contains(buf.Bytes(), []byte("chapters:")) {
		t.Error("output missing 'chapters' field")
	}

	// Should be valid YAML that can be loaded
	_, err = openacr.LoadYAMLBytes([]byte(output))
	if err != nil {
		t.Errorf("output is not valid OpenACR YAML: %v", err)
	}
}

func TestGenerateOpenACRWithOptions(t *testing.T) {
	result := &audit.AuditResult{
		ID:          "test-audit-1",
		TargetURL:   "https://example.com",
		WCAGVersion: audit.WCAG22,
		StartTime:   time.Now(),
		Conformance: audit.ConformanceSummary{
			Criteria: []audit.CriterionResult{
				{ID: "1.1.1", Level: "A", Status: "supports"},
			},
		},
	}

	opts := OpenACROptions{
		ProductName:    "My Product",
		ProductVersion: "1.0.0",
		AuthorName:     "Jane Doe",
		AuthorEmail:    "jane@example.com",
		VendorName:     "Acme Corp",
		VendorEmail:    "contact@acme.com",
		CatalogID:      "2.5-edition-wcag-2.1-508-en",
	}

	report, err := GenerateOpenACR(result, opts)
	if err != nil {
		t.Fatalf("GenerateOpenACR() error = %v", err)
	}

	if report.Product.Name != "My Product" {
		t.Errorf("Product.Name = %q, want %q", report.Product.Name, "My Product")
	}
	if report.Product.Version != "1.0.0" {
		t.Errorf("Product.Version = %q, want %q", report.Product.Version, "1.0.0")
	}
	if report.Author.Name != "Jane Doe" {
		t.Errorf("Author.Name = %q, want %q", report.Author.Name, "Jane Doe")
	}
	if report.Author.Email != "jane@example.com" {
		t.Errorf("Author.Email = %q, want %q", report.Author.Email, "jane@example.com")
	}
	if report.Vendor == nil {
		t.Fatal("expected Vendor to be set")
	}
	if report.Vendor.CompanyName != "Acme Corp" {
		t.Errorf("Vendor.CompanyName = %q, want %q", report.Vendor.CompanyName, "Acme Corp")
	}
	if report.Catalog != "2.5-edition-wcag-2.1-508-en" {
		t.Errorf("Catalog = %q, want %q", report.Catalog, "2.5-edition-wcag-2.1-508-en")
	}
}

func TestValidateOpenACRReport(t *testing.T) {
	result := &audit.AuditResult{
		ID:          "test-audit-1",
		TargetURL:   "https://example.com",
		WCAGVersion: audit.WCAG22,
		StartTime:   time.Now(),
		Conformance: audit.ConformanceSummary{
			Criteria: []audit.CriterionResult{
				{ID: "1.1.1", Level: "A", Status: "supports"},
			},
		},
	}

	report := convertToOpenACR(result)

	// The converted report should have validation errors because
	// we don't set author email by default
	errs := ValidateOpenACRReport(report)

	// Check that we get the expected validation error
	hasEmailError := false
	for _, err := range errs {
		if err.Field == "author.email" {
			hasEmailError = true
			break
		}
	}

	if !hasEmailError {
		t.Error("expected validation error for missing author.email")
	}
}

func TestBuildChapters_EmptyConformance(t *testing.T) {
	result := &audit.AuditResult{
		Conformance: audit.ConformanceSummary{
			Criteria: []audit.CriterionResult{},
		},
	}

	chapters := buildChapters(result)

	// Should still have hardware and software chapters
	if _, ok := chapters["hardware"]; !ok {
		t.Error("expected hardware chapter even with empty conformance")
	}
	if _, ok := chapters["software"]; !ok {
		t.Error("expected software chapter even with empty conformance")
	}

	// Should not have WCAG level chapters
	if _, ok := chapters["success_criteria_level_a"]; ok {
		t.Error("should not have level A chapter with no criteria")
	}
}
