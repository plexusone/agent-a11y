// Package types provides shared types used across the accessibility audit service.
// This package exists to break import cycles between packages.
package types

import "time"

// Severity represents the severity of an accessibility issue.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeveritySerious  Severity = "serious"
	SeverityModerate Severity = "moderate"
	SeverityMinor    Severity = "minor"
)

// Impact represents the WCAG impact level.
type Impact string

const (
	ImpactBlocker  Impact = "blocker"  // Prevents access entirely
	ImpactCritical Impact = "critical" // Significant barrier
	ImpactSerious  Impact = "serious"  // Moderate barrier
	ImpactModerate Impact = "moderate" // Minor barrier
	ImpactMinor    Impact = "minor"    // Inconvenience
)

// WCAGLevel represents WCAG conformance levels.
type WCAGLevel string

const (
	WCAGLevelA   WCAGLevel = "A"
	WCAGLevelAA  WCAGLevel = "AA"
	WCAGLevelAAA WCAGLevel = "AAA"
)

// WCAGVersion represents WCAG version.
type WCAGVersion string

const (
	WCAG20 WCAGVersion = "2.0"
	WCAG21 WCAGVersion = "2.1"
	WCAG22 WCAGVersion = "2.2"
)

// Finding represents an individual accessibility finding.
type Finding struct {
	// Core identification
	ID          string `json:"id"`
	RuleID      string `json:"ruleId"`      // Rule that found this issue
	Description string `json:"description"` // Human-readable description
	Help        string `json:"help"`        // Help text for fixing

	// WCAG mapping
	SuccessCriteria []string  `json:"successCriteria"` // e.g., ["1.1.1", "4.1.2"]
	Level           WCAGLevel `json:"level"`           // Highest level affected
	Impact          Impact    `json:"impact"`

	// Element information
	Selector   string `json:"selector"`   // CSS selector
	XPath      string `json:"xpath"`      // XPath
	HTML       string `json:"html"`       // HTML snippet
	Element    string `json:"element"`    // Element tag name
	PageURL    string `json:"pageUrl"`    // URL where found
	PageTitle  string `json:"pageTitle"`  // Page title
	Screenshot string `json:"screenshot"` // Base64 screenshot (optional)

	// Context
	JourneyStep string `json:"journeyStep,omitempty"` // Which journey step
	Component   string `json:"component,omitempty"`   // UI component name

	// Severity (for easier sorting/filtering)
	Severity Severity `json:"severity,omitempty"`

	// Remediation guidance
	Remediation *Remediation `json:"remediation,omitempty"`

	// LLM evaluation (if enabled)
	LLMEvaluation *LLMEvaluation `json:"llmEvaluation,omitempty"`

	// Timestamps
	FoundAt time.Time `json:"foundAt"`
}

// Remediation contains standardized remediation references.
// Uses WCAG techniques and ACT rules similar to how SAST tools use CWE IDs.
type Remediation struct {
	// Brief inline summary (1-2 sentences)
	Summary string `json:"summary"`

	// WCAG Technique references (e.g., H37, G94, ARIA6)
	Techniques []TechniqueRef `json:"techniques,omitempty"`

	// ACT Rule ID (W3C standardized test methodology)
	ACTRuleID string `json:"actRuleId,omitempty"`

	// axe-core rule ID (Deque)
	AxeRuleID string `json:"axeRuleId,omitempty"`

	// Reference URLs
	References []ReferenceURL `json:"references,omitempty"`

	// LLM-generated context-specific guidance (optional)
	ContextualFix string `json:"contextualFix,omitempty"`
}

// TechniqueRef references a WCAG technique.
type TechniqueRef struct {
	ID    string `json:"id"`    // e.g., "H37"
	Type  string `json:"type"`  // "sufficient", "advisory", "failure"
	Title string `json:"title"` // e.g., "Using alt attributes on img elements"
	URL   string `json:"url"`   // W3C documentation URL
}

// ReferenceURL contains a reference link with metadata.
type ReferenceURL struct {
	Title  string `json:"title"`  // e.g., "WCAG Understanding 1.1.1"
	URL    string `json:"url"`    // Full URL
	Source string `json:"source"` // "w3c", "deque", "webaim"
}

// LLMEvaluation contains LLM-as-a-Judge evaluation results.
type LLMEvaluation struct {
	// Evaluation result
	Confirmed   bool    `json:"confirmed"`   // LLM confirms the issue
	Confidence  float64 `json:"confidence"`  // 0.0-1.0
	Reasoning   string  `json:"reasoning"`   // Why the LLM made this decision
	Severity    string  `json:"severity"`    // LLM-assessed severity
	Remediation string  `json:"remediation"` // Suggested fix

	// For manual review items
	NeedsManualReview bool   `json:"needsManualReview"`
	ReviewGuidance    string `json:"reviewGuidance,omitempty"`

	// Metadata
	Model     string    `json:"model"`
	EvalTime  time.Time `json:"evalTime"`
	TokensIn  int       `json:"tokensIn"`
	TokensOut int       `json:"tokensOut"`
}
