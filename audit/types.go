// Package audit provides core types for accessibility auditing.
package audit

import (
	"time"

	"github.com/plexusone/agent-a11y/types"
)

// Re-export common types from the types package for convenience.
type (
	Severity      = types.Severity
	Impact        = types.Impact
	WCAGLevel     = types.WCAGLevel
	WCAGVersion   = types.WCAGVersion
	Finding       = types.Finding
	LLMEvaluation = types.LLMEvaluation
)

// Re-export constants
const (
	SeverityCritical = types.SeverityCritical
	SeveritySerious  = types.SeveritySerious
	SeverityModerate = types.SeverityModerate
	SeverityMinor    = types.SeverityMinor

	ImpactBlocker  = types.ImpactBlocker
	ImpactCritical = types.ImpactCritical
	ImpactSerious  = types.ImpactSerious
	ImpactModerate = types.ImpactModerate
	ImpactMinor    = types.ImpactMinor

	WCAGLevelA   = types.WCAGLevelA
	WCAGLevelAA  = types.WCAGLevelAA
	WCAGLevelAAA = types.WCAGLevelAAA

	WCAG20 = types.WCAG20
	WCAG21 = types.WCAG21
	WCAG22 = types.WCAG22
)

// SuccessCriterion represents a WCAG success criterion.
type SuccessCriterion struct {
	ID          string    `json:"id"`          // e.g., "1.1.1"
	Name        string    `json:"name"`        // e.g., "Non-text Content"
	Level       WCAGLevel `json:"level"`       // A, AA, AAA
	Version     string    `json:"version"`     // WCAG version where introduced
	Description string    `json:"description"` // Full description
	URL         string    `json:"url"`         // Link to WCAG docs
}

// PageResult contains audit results for a single page.
type PageResult struct {
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Duration     int64     `json:"durationMs"`
	LoadTime     int64     `json:"loadTimeMs"`
	IsSPA        bool      `json:"isSPA"`
	SPAFramework string    `json:"spaFramework,omitempty"` // react, vue, angular, etc.

	// Findings organized by category
	Findings []Finding `json:"findings"`

	// Screenshots
	FullPageScreenshot string `json:"fullPageScreenshot,omitempty"`

	// Page metadata
	Language   string `json:"language"`
	DocType    string `json:"doctype"`
	HasSkipNav bool   `json:"hasSkipNav"`
	Landmarks  int    `json:"landmarks"`
}

// JourneyResult contains audit results for a user journey.
type JourneyResult struct {
	JourneyID   string        `json:"journeyId"`
	JourneyName string        `json:"journeyName"`
	Mode        string        `json:"mode"` // agentic, deterministic, hybrid
	StartTime   time.Time     `json:"startTime"`
	EndTime     time.Time     `json:"endTime"`
	Duration    time.Duration `json:"duration"`
	Status      string        `json:"status"` // success, failure, partial
	Error       string        `json:"error,omitempty"`

	// Step results
	Steps []StepResult `json:"steps"`

	// Aggregate findings from all steps
	Findings []Finding `json:"findings"`
}

// StepResult contains results for a journey step.
type StepResult struct {
	StepIndex int       `json:"stepIndex"`
	StepName  string    `json:"stepName"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Duration  int64     `json:"durationMs"`
	Status    string    `json:"status"` // success, failure, skipped
	Error     string    `json:"error,omitempty"`

	// Page state after step
	PageURL    string `json:"pageUrl"`
	PageTitle  string `json:"pageTitle"`
	Screenshot string `json:"screenshot,omitempty"`

	// Findings from audit point (if step triggers audit)
	AuditTriggered bool      `json:"auditTriggered"`
	Findings       []Finding `json:"findings,omitempty"`
}

// AuditResult contains the complete audit results.
type AuditResult struct {
	// Metadata
	ID        string    `json:"id"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Duration  int64     `json:"durationMs"`

	// Configuration used
	TargetURL   string      `json:"targetUrl"`
	WCAGVersion WCAGVersion `json:"wcagVersion"`
	WCAGLevel   WCAGLevel   `json:"wcagLevel"`
	LLMEnabled  bool        `json:"llmEnabled"`
	LLMModel    string      `json:"llmModel,omitempty"`

	// Results
	Pages    []PageResult    `json:"pages"`
	Journeys []JourneyResult `json:"journeys,omitempty"`

	// Aggregate statistics
	Stats AuditStats `json:"stats"`

	// Conformance summary
	Conformance ConformanceSummary `json:"conformance"`
}

// AuditStats contains aggregate statistics.
type AuditStats struct {
	TotalPages    int `json:"totalPages"`
	TotalFindings int `json:"totalFindings"`

	// By severity
	Critical int `json:"critical"`
	Serious  int `json:"serious"`
	Moderate int `json:"moderate"`
	Minor    int `json:"minor"`

	// By level
	LevelA   int `json:"levelA"`
	LevelAA  int `json:"levelAA"`
	LevelAAA int `json:"levelAAA"`

	// By category
	ByCategory map[string]int `json:"byCategory"`

	// LLM stats
	LLMEvaluations int `json:"llmEvaluations"`
	ManualReviews  int `json:"manualReviews"`
}

// ConformanceSummary provides WCAG conformance status.
type ConformanceSummary struct {
	// Target level
	TargetLevel   WCAGLevel `json:"targetLevel"`
	Version       string    `json:"version"`
	OverallStatus string    `json:"overallStatus"`

	// Status per level
	LevelA   LevelConformance `json:"levelA"`
	LevelAA  LevelConformance `json:"levelAA"`
	LevelAAA LevelConformance `json:"levelAAA"`

	// Success criteria breakdown
	Criteria []CriterionResult `json:"criteria"`
}

// LevelConformance represents conformance status for a level.
type LevelConformance struct {
	Status         string `json:"status"` // supports, partially_supports, does_not_support, not_applicable
	TotalIssues    int    `json:"totalIssues"`
	BlockingIssues int    `json:"blockingIssues"`
}

// CriterionResult represents the result for a specific success criterion.
type CriterionResult struct {
	ID         string `json:"id"`     // e.g., "1.1.1"
	Name       string `json:"name"`
	Level      string `json:"level"`
	Status     string `json:"status"` // supports, partially_supports, does_not_support, not_applicable
	IssueCount int    `json:"issueCount"`
	Remarks    string `json:"remarks,omitempty"`
}

// VPATReport represents a VPAT 2.4 conformance report.
type VPATReport struct {
	// Report metadata
	ProductName    string    `json:"productName"`
	ProductVersion string    `json:"productVersion"`
	VendorName     string    `json:"vendorName"`
	ContactEmail   string    `json:"contactEmail"`
	ReportDate     time.Time `json:"reportDate"`
	EvaluationURL  string    `json:"evaluationUrl"`

	// Conformance tables
	WCAGConformance []VPATCriterion `json:"wcagConformance"`

	// Evaluation methods
	EvaluationMethods []string `json:"evaluationMethods"`

	// Notes
	LegalDisclaimer string `json:"legalDisclaimer"`
	Notes           string `json:"notes,omitempty"`
}

// VPATCriterion represents a criterion in the VPAT report.
type VPATCriterion struct {
	Criterion   string `json:"criterion"`   // e.g., "1.1.1 Non-text Content"
	Level       string `json:"level"`       // A, AA, AAA
	Conformance string `json:"conformance"` // Supports, Partially Supports, Does Not Support, Not Applicable
	Remarks     string `json:"remarks"`     // Explanation
}
