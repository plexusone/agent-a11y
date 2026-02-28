package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/plexusone/agent-a11y/audit"
)

func createTestAuditResult() *audit.AuditResult {
	now := time.Now()
	return &audit.AuditResult{
		ID:          "test-audit-1",
		StartTime:   now,
		EndTime:     now.Add(5 * time.Minute),
		Duration:    300000,
		TargetURL:   "https://example.com",
		WCAGVersion: "2.2",
		WCAGLevel:   "AA",
		Pages: []audit.PageResult{
			{
				URL:       "https://example.com",
				Title:     "Example Page",
				StartTime: now,
				EndTime:   now.Add(30 * time.Second),
				Duration:  30000,
				Findings: []audit.Finding{
					{
						ID:              "finding-1",
						RuleID:          "image-alt",
						Description:     "Image missing alternative text",
						Help:            "Add alt attribute to img element",
						SuccessCriteria: []string{"1.1.1"},
						Level:           "A",
						Impact:          "critical",
						Selector:        "img.hero",
						HTML:            `<img src="hero.jpg">`,
						Severity:        "critical",
					},
					{
						ID:              "finding-2",
						RuleID:          "link-purpose",
						Description:     "Link text is not descriptive",
						Help:            "Use descriptive link text",
						SuccessCriteria: []string{"2.4.4"},
						Level:           "A",
						Impact:          "moderate",
						Selector:        "a.readmore",
						HTML:            `<a href="/more">Click here</a>`,
						Severity:        "moderate",
					},
				},
			},
			{
				URL:       "https://example.com/about",
				Title:     "About Us",
				StartTime: now.Add(30 * time.Second),
				EndTime:   now.Add(60 * time.Second),
				Duration:  30000,
				Findings:  []audit.Finding{},
			},
		},
		Stats: audit.AuditStats{
			TotalPages:    2,
			TotalFindings: 2,
			Critical:      1,
			Serious:       0,
			Moderate:      1,
			Minor:         0,
			LevelA:        2,
			LevelAA:       0,
			LevelAAA:      0,
		},
		Conformance: audit.ConformanceSummary{
			TargetLevel:   "AA",
			Version:       "2.2",
			OverallStatus: "Non-Conformant",
			LevelA: audit.LevelConformance{
				Status:      "Partially Supports",
				TotalIssues: 2,
			},
			LevelAA: audit.LevelConformance{
				Status:      "Supports",
				TotalIssues: 0,
			},
			LevelAAA: audit.LevelConformance{
				Status:      "Supports",
				TotalIssues: 0,
			},
		},
	}
}

func TestNewWriter(t *testing.T) {
	tests := []struct {
		format   Format
		expected Format
	}{
		{FormatJSON, FormatJSON},
		{FormatHTML, FormatHTML},
		{FormatMarkdown, FormatMarkdown},
		{FormatVPAT, FormatVPAT},
		{FormatWCAG, FormatWCAG},
	}

	for _, tt := range tests {
		w := NewWriter(tt.format)
		if w == nil {
			t.Errorf("NewWriter(%q) returned nil", tt.format)
			continue
		}
		if w.format != tt.expected {
			t.Errorf("NewWriter(%q).format: got %q, want %q", tt.format, w.format, tt.expected)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	result := createTestAuditResult()
	w := NewWriter(FormatJSON)

	var buf bytes.Buffer
	if err := w.Write(&buf, result); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify it's valid JSON
	var decoded audit.AuditResult
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify key fields
	if decoded.ID != result.ID {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, result.ID)
	}
	if decoded.TargetURL != result.TargetURL {
		t.Errorf("TargetURL mismatch: got %q, want %q", decoded.TargetURL, result.TargetURL)
	}
	if len(decoded.Pages) != 2 {
		t.Errorf("Pages count: got %d, want 2", len(decoded.Pages))
	}
}

func TestWriteMarkdown(t *testing.T) {
	result := createTestAuditResult()
	w := NewWriter(FormatMarkdown)

	var buf bytes.Buffer
	if err := w.Write(&buf, result); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Check for expected markdown elements
	expectedStrings := []string{
		"# Accessibility Audit Report",
		"**URL:** https://example.com",
		"**Total Pages:** 2",
		"**Total Findings:** 2",
		"Critical: 1",
		"Moderate: 1",
		"## Conformance Status",
		"| Level | Status | Issues |",
		"| A | Partially Supports | 2 |",
		"## Findings",
		"### https://example.com",
		"### https://example.com/about",
		"image-alt",
		"link-purpose",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Markdown output missing: %q", expected)
		}
	}
}

func TestWriteHTML(t *testing.T) {
	result := createTestAuditResult()
	w := NewWriter(FormatHTML)

	var buf bytes.Buffer
	if err := w.Write(&buf, result); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Check for expected HTML elements
	expectedStrings := []string{
		"<!DOCTYPE html>",
		"<html lang=\"en\">",
		"<title>Accessibility Audit Report</title>",
		"<h1>Accessibility Audit Report</h1>",
		"https://example.com",
		"WCAG:",
		"2.2 Level AA",
		"Pages Audited",
		"Total Issues",
		"Critical",
		"Conformance",
		"Level A",
		"image-alt",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("HTML output missing: %q", expected)
		}
	}
}

func TestWriteVPAT(t *testing.T) {
	result := createTestAuditResult()
	w := NewWriter(FormatVPAT)

	var buf bytes.Buffer
	if err := w.Write(&buf, result); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Check for expected VPAT elements
	expectedStrings := []string{
		"# VPAT 2.4 Conformance Report",
		"**Product Name:**",
		"**Report Date:**",
		"**Evaluation URL:** https://example.com",
		"## Evaluation Methods",
		"Automated accessibility testing",
		"## WCAG 2.2 Conformance",
		"| Criterion | Level | Conformance | Remarks |",
		"1.1.1 Non-text Content",
		"1.4.3 Contrast (Minimum)",
		"## Legal Disclaimer",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("VPAT output missing: %q", expected)
		}
	}
}

func TestWriteWCAGReport(t *testing.T) {
	result := createTestAuditResult()
	result.Conformance.Criteria = []audit.CriterionResult{
		{
			ID:         "1.1.1",
			Name:       "Non-text Content",
			Level:      "A",
			Status:     "Does Not Support",
			IssueCount: 1,
		},
		{
			ID:         "2.4.4",
			Name:       "Link Purpose",
			Level:      "A",
			Status:     "Partially Supports",
			IssueCount: 1,
		},
	}

	w := NewWriter(FormatWCAG)

	var buf bytes.Buffer
	if err := w.Write(&buf, result); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	expectedStrings := []string{
		"# WCAG Conformance Report",
		"**Target:** https://example.com",
		"**Standard:** WCAG 2.2 Level AA",
		"## Success Criteria Results",
		"| ID | Name | Level | Status | Issues |",
		"| 1.1.1 | Non-text Content | A | Does Not Support | 1 |",
		"| 2.4.4 | Link Purpose | A | Partially Supports | 1 |",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("WCAG report output missing: %q", expected)
		}
	}
}

func TestWriteUnsupportedFormat(t *testing.T) {
	result := createTestAuditResult()
	w := &Writer{format: "invalid"}

	var buf bytes.Buffer
	err := w.Write(&buf, result)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Error message should mention unsupported format: %v", err)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d): got %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestFormatConstants(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{FormatJSON, "json"},
		{FormatHTML, "html"},
		{FormatMarkdown, "markdown"},
		{FormatVPAT, "vpat"},
		{FormatWCAG, "wcag"},
	}

	for _, tt := range tests {
		if string(tt.format) != tt.expected {
			t.Errorf("Format %v: got %q, want %q", tt.format, string(tt.format), tt.expected)
		}
	}
}

func TestWriteEmptyResult(t *testing.T) {
	result := &audit.AuditResult{
		ID:          "empty-audit",
		TargetURL:   "https://example.com",
		WCAGVersion: "2.2",
		WCAGLevel:   "AA",
		Pages:       []audit.PageResult{},
		Stats:       audit.AuditStats{},
		Conformance: audit.ConformanceSummary{
			TargetLevel: "AA",
		},
	}

	formats := []Format{FormatJSON, FormatHTML, FormatMarkdown, FormatVPAT, FormatWCAG}

	for _, format := range formats {
		w := NewWriter(format)
		var buf bytes.Buffer
		if err := w.Write(&buf, result); err != nil {
			t.Errorf("Write(%q) with empty result failed: %v", format, err)
		}
		if buf.Len() == 0 {
			t.Errorf("Write(%q) produced empty output", format)
		}
	}
}

func TestGenerateVPATReport(t *testing.T) {
	result := createTestAuditResult()
	report := generateVPATReport(result)

	if report == nil {
		t.Fatal("generateVPATReport returned nil")
	}

	if report.EvaluationURL != result.TargetURL {
		t.Errorf("EvaluationURL mismatch: got %q, want %q", report.EvaluationURL, result.TargetURL)
	}

	if len(report.EvaluationMethods) == 0 {
		t.Error("EvaluationMethods should not be empty")
	}

	if len(report.WCAGConformance) == 0 {
		t.Error("WCAGConformance should not be empty")
	}

	if report.LegalDisclaimer == "" {
		t.Error("LegalDisclaimer should not be empty")
	}

	// Check that criteria with issues are marked correctly
	foundIssue := false
	for _, criterion := range report.WCAGConformance {
		if strings.Contains(criterion.Criterion, "1.1.1") {
			foundIssue = true
			if criterion.Conformance == "Supports" {
				t.Error("1.1.1 should not be 'Supports' when issues exist")
			}
		}
	}
	if !foundIssue {
		t.Error("Should have found criterion 1.1.1")
	}
}

func TestGetWCAG22Criteria(t *testing.T) {
	criteria := getWCAG22Criteria()

	if len(criteria) == 0 {
		t.Error("getWCAG22Criteria returned empty slice")
	}

	// Check for some expected criteria
	expectedIDs := []string{"1.1.1", "1.4.3", "2.1.1", "2.4.7", "4.1.2"}
	for _, expectedID := range expectedIDs {
		found := false
		for _, c := range criteria {
			if c.ID == expectedID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected criterion: %s", expectedID)
		}
	}

	// Verify all criteria have required fields
	for _, c := range criteria {
		if c.ID == "" {
			t.Error("Criterion has empty ID")
		}
		if c.Name == "" {
			t.Error("Criterion has empty Name")
		}
		if c.Level != "A" && c.Level != "AA" && c.Level != "AAA" {
			t.Errorf("Invalid level for criterion %s: %s", c.ID, c.Level)
		}
	}
}
