package audit

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

func TestPageResultJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	pr := PageResult{
		URL:       "https://example.com",
		Title:     "Test Page",
		StartTime: now,
		EndTime:   now.Add(time.Minute),
		Duration:  60000,
		LoadTime:  5000,
		IsSPA:     true,
		SPAFramework: "react",
		Findings: []Finding{
			{
				ID:     "f1",
				RuleID: "image-alt",
			},
		},
		Language:  "en",
		DocType:   "html",
		HasSkipNav: true,
		Landmarks: 5,
	}

	data, err := json.Marshal(pr)
	if err != nil {
		t.Fatalf("Failed to marshal PageResult: %v", err)
	}

	var decoded PageResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PageResult: %v", err)
	}

	if decoded.URL != pr.URL {
		t.Errorf("URL mismatch")
	}
	if decoded.IsSPA != pr.IsSPA {
		t.Errorf("IsSPA mismatch")
	}
	if len(decoded.Findings) != 1 {
		t.Errorf("Findings count mismatch")
	}
}

func TestJourneyResultJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	jr := JourneyResult{
		JourneyID:   "journey-1",
		JourneyName: "Checkout Flow",
		Mode:        "deterministic",
		StartTime:   now,
		EndTime:     now.Add(5 * time.Minute),
		Duration:    5 * time.Minute,
		Status:      "success",
		Steps: []StepResult{
			{
				StepIndex: 0,
				StepName:  "Navigate to home",
				Action:    "navigate",
				Status:    "success",
			},
		},
		Findings: []Finding{},
	}

	data, err := json.Marshal(jr)
	if err != nil {
		t.Fatalf("Failed to marshal JourneyResult: %v", err)
	}

	var decoded JourneyResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JourneyResult: %v", err)
	}

	if decoded.JourneyID != jr.JourneyID {
		t.Errorf("JourneyID mismatch")
	}
	if decoded.Status != "success" {
		t.Errorf("Status mismatch")
	}
}

func TestStepResultJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	sr := StepResult{
		StepIndex:      0,
		StepName:       "Click submit",
		Action:         "click",
		Timestamp:      now,
		Duration:       1000,
		Status:         "success",
		PageURL:        "https://example.com/form",
		PageTitle:      "Submit Form",
		AuditTriggered: true,
		Findings: []Finding{
			{ID: "f1"},
		},
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatalf("Failed to marshal StepResult: %v", err)
	}

	var decoded StepResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal StepResult: %v", err)
	}

	if decoded.StepName != sr.StepName {
		t.Errorf("StepName mismatch")
	}
	if !decoded.AuditTriggered {
		t.Error("AuditTriggered should be true")
	}
}

func TestAuditResultJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	ar := AuditResult{
		ID:          "audit-1",
		StartTime:   now,
		EndTime:     now.Add(10 * time.Minute),
		Duration:    600000,
		TargetURL:   "https://example.com",
		WCAGVersion: WCAG22,
		WCAGLevel:   WCAGLevelAA,
		LLMEnabled:  true,
		LLMModel:    "claude-sonnet-4-20250514",
		Pages: []PageResult{
			{URL: "https://example.com"},
		},
		Stats: AuditStats{
			TotalPages:    1,
			TotalFindings: 5,
			Critical:      1,
			Serious:       2,
			Moderate:      1,
			Minor:         1,
		},
		Conformance: ConformanceSummary{
			TargetLevel:   WCAGLevelAA,
			OverallStatus: "Non-Conformant",
		},
	}

	data, err := json.Marshal(ar)
	if err != nil {
		t.Fatalf("Failed to marshal AuditResult: %v", err)
	}

	var decoded AuditResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AuditResult: %v", err)
	}

	if decoded.ID != ar.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.WCAGVersion != WCAG22 {
		t.Errorf("WCAGVersion mismatch")
	}
	if decoded.Stats.TotalFindings != 5 {
		t.Errorf("Stats.TotalFindings mismatch")
	}
}

func TestAuditStatsJSONSerialization(t *testing.T) {
	stats := AuditStats{
		TotalPages:     10,
		TotalFindings:  50,
		Critical:       5,
		Serious:        15,
		Moderate:       20,
		Minor:          10,
		LevelA:         25,
		LevelAA:        20,
		LevelAAA:       5,
		ByCategory:     map[string]int{"color-contrast": 15, "alternative-text": 10},
		LLMEvaluations: 30,
		ManualReviews:  5,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal AuditStats: %v", err)
	}

	var decoded AuditStats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AuditStats: %v", err)
	}

	if decoded.TotalPages != stats.TotalPages {
		t.Errorf("TotalPages mismatch")
	}
	if decoded.ByCategory["color-contrast"] != 15 {
		t.Errorf("ByCategory[color-contrast] mismatch")
	}
}

func TestConformanceSummaryJSONSerialization(t *testing.T) {
	cs := ConformanceSummary{
		TargetLevel:   WCAGLevelAA,
		Version:       "2.2",
		OverallStatus: "Partially Conformant",
		LevelA: LevelConformance{
			Status:         "Supports",
			TotalIssues:    0,
			BlockingIssues: 0,
		},
		LevelAA: LevelConformance{
			Status:         "Partially Supports",
			TotalIssues:    5,
			BlockingIssues: 1,
		},
		LevelAAA: LevelConformance{
			Status:         "Does Not Support",
			TotalIssues:    10,
			BlockingIssues: 3,
		},
		Criteria: []CriterionResult{
			{
				ID:         "1.1.1",
				Name:       "Non-text Content",
				Level:      "A",
				Status:     "Supports",
				IssueCount: 0,
			},
		},
	}

	data, err := json.Marshal(cs)
	if err != nil {
		t.Fatalf("Failed to marshal ConformanceSummary: %v", err)
	}

	var decoded ConformanceSummary
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ConformanceSummary: %v", err)
	}

	if decoded.TargetLevel != WCAGLevelAA {
		t.Errorf("TargetLevel mismatch")
	}
	if decoded.LevelAA.TotalIssues != 5 {
		t.Errorf("LevelAA.TotalIssues mismatch")
	}
}

func TestVPATReportJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	vr := VPATReport{
		ProductName:    "Test Application",
		ProductVersion: "1.0.0",
		VendorName:     "Test Vendor",
		ContactEmail:   "a11y@example.com",
		ReportDate:     now,
		EvaluationURL:  "https://example.com",
		WCAGConformance: []VPATCriterion{
			{
				Criterion:   "1.1.1 Non-text Content",
				Level:       "A",
				Conformance: "Supports",
				Remarks:     "No issues found",
			},
		},
		EvaluationMethods: []string{"Automated testing", "Manual review"},
		LegalDisclaimer:   "This report is provided as-is.",
		Notes:             "Additional notes",
	}

	data, err := json.Marshal(vr)
	if err != nil {
		t.Fatalf("Failed to marshal VPATReport: %v", err)
	}

	var decoded VPATReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal VPATReport: %v", err)
	}

	if decoded.ProductName != vr.ProductName {
		t.Errorf("ProductName mismatch")
	}
	if len(decoded.WCAGConformance) != 1 {
		t.Errorf("WCAGConformance count mismatch")
	}
}

func TestSuccessCriterionJSONSerialization(t *testing.T) {
	sc := SuccessCriterion{
		ID:          "1.4.3",
		Name:        "Contrast (Minimum)",
		Level:       WCAGLevelAA,
		Version:     "2.0",
		Description: "Text has sufficient contrast ratio",
		URL:         "https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html",
	}

	data, err := json.Marshal(sc)
	if err != nil {
		t.Fatalf("Failed to marshal SuccessCriterion: %v", err)
	}

	var decoded SuccessCriterion
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SuccessCriterion: %v", err)
	}

	if decoded.ID != sc.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.Level != WCAGLevelAA {
		t.Errorf("Level mismatch")
	}
}
