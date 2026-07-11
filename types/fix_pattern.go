package types

// FixPattern provides actionable guidance for coding agents.
// Maps WCAG techniques to concrete implementation patterns.
type FixPattern struct {
	// Type of change required
	Type FixType `json:"type"`

	// Action to perform
	Action FixAction `json:"action"`

	// What to target (attribute name, CSS property, etc.)
	Target string `json:"target"`

	// Suggested value (generic, may need design system token)
	Value string `json:"value,omitempty"`

	// Pseudo-code example of the fix
	Example string `json:"example"`

	// Priority for this fix (when multiple options exist)
	Priority int `json:"priority,omitempty"`
}

// FixType categorizes the type of code change needed.
type FixType string

const (
	FixTypeAttribute FixType = "attribute" // Add/modify HTML attribute
	FixTypeStyle     FixType = "style"     // CSS style change
	FixTypeStructure FixType = "structure" // DOM structure change
	FixTypeContent   FixType = "content"   // Text content change
	FixTypeARIA      FixType = "aria"      // ARIA attribute change
	FixTypeFocus     FixType = "focus"     // Focus management
	FixTypeOrder     FixType = "order"     // DOM/reading order
	FixTypeSemantic  FixType = "semantic"  // Semantic HTML change
)

// FixAction specifies the action to perform.
type FixAction string

const (
	FixActionAdd     FixAction = "add"     // Add new attribute/element
	FixActionModify  FixAction = "modify"  // Change existing value
	FixActionRemove  FixAction = "remove"  // Remove attribute/element
	FixActionReplace FixAction = "replace" // Replace element entirely
	FixActionWrap    FixAction = "wrap"    // Wrap in new element
	FixActionUnwrap  FixAction = "unwrap"  // Remove wrapper
)

// TokenSuggestion suggests a design system token for a fix.
type TokenSuggestion struct {
	// CSS property this applies to
	Property string `json:"property"`

	// Current value in the code
	CurrentValue string `json:"currentValue"`

	// Suggested design token name
	TokenName string `json:"tokenName"`

	// Resolved token value
	TokenValue string `json:"tokenValue"`

	// Why this token is suggested
	Rationale string `json:"rationale"`

	// Contrast ratio (for color tokens)
	ContrastRatio float64 `json:"contrastRatio,omitempty"`
}

// AgentRemediation extends Remediation with agent-actionable guidance.
type AgentRemediation struct {
	// Inherit standard remediation
	Remediation

	// Actionable fix patterns (ordered by priority)
	FixPatterns []FixPattern `json:"fixPatterns"`

	// Design system token suggestions (if design system loaded)
	TokenSuggestions []TokenSuggestion `json:"tokenSuggestions,omitempty"`

	// Affected component from design system (if detected)
	DesignSystemComponent string `json:"designSystemComponent,omitempty"`

	// Confidence that this fix will resolve the issue (0-1)
	FixConfidence float64 `json:"fixConfidence"`
}

// ElementContext provides detailed context about the affected element.
type ElementContext struct {
	// CSS selector (most specific)
	Selector string `json:"selector"`

	// XPath (for complex selections)
	XPath string `json:"xpath"`

	// HTML snippet
	HTML string `json:"html"`

	// Tag name
	TagName string `json:"tagName"`

	// Computed styles relevant to the issue
	ComputedStyles map[string]string `json:"computedStyles,omitempty"`

	// Parent context (for structural issues)
	ParentHTML string `json:"parentHtml,omitempty"`

	// Bounding box (for visual issues)
	BoundingBox *BoundingBox `json:"boundingBox,omitempty"`

	// Design system component ID (if detected)
	ComponentID string `json:"componentId,omitempty"`

	// Component variant (if detected)
	ComponentVariant string `json:"componentVariant,omitempty"`
}

// BoundingBox represents element position and size.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// AgentFinding is the agent-optimized finding format.
// This is the default output format for coding agents.
type AgentFinding struct {
	// Core finding information
	Finding Finding `json:"finding"`

	// Detailed element context
	Element ElementContext `json:"element"`

	// Agent-actionable remediation
	Remediation AgentRemediation `json:"remediation"`
}

// AgentResult is the agent-optimized audit result.
// This is the default output format.
type AgentResult struct {
	// Audit metadata
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
	Duration  string `json:"duration"`
	Level     string `json:"level"` // Target WCAG level

	// Design system info (if loaded)
	DesignSystem *DesignSystemInfo `json:"designSystem,omitempty"`

	// Agent-optimized findings
	Findings []AgentFinding `json:"findings"`

	// Summary statistics
	Summary AgentSummary `json:"summary"`

	// Overall status for multi-agent workflows
	Status string `json:"status"` // "GO", "WARN", "NO-GO"
}

// DesignSystemInfo contains loaded design system metadata.
type DesignSystemInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

// AgentSummary provides aggregate statistics.
type AgentSummary struct {
	Total      int `json:"total"`
	Critical   int `json:"critical"`
	Serious    int `json:"serious"`
	Moderate   int `json:"moderate"`
	Minor      int `json:"minor"`
	Fixable    int `json:"fixable"`    // Issues with fix patterns
	NeedsToken int `json:"needsToken"` // Issues that need design tokens
}
