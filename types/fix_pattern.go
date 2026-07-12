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

	// Project-specific fix patterns (from ~/.plexusone/a11y/fixes.yaml)
	ProjectPatterns []ProjectPattern `json:"projectPatterns,omitempty"`

	// Design system token suggestions (if design system loaded)
	TokenSuggestions []TokenSuggestion `json:"tokenSuggestions,omitempty"`

	// Affected component from design system (if detected)
	DesignSystemComponent string `json:"designSystemComponent,omitempty"`

	// Confidence that this fix will resolve the issue (0-1)
	FixConfidence float64 `json:"fixConfidence"`
}

// ProjectPattern is a fix pattern from project configuration.
type ProjectPattern struct {
	// Source indicates where this pattern came from
	Source string `json:"source"` // "project", "library", "language"

	// Component name (for component-specific fixes)
	Component string `json:"component,omitempty"`

	// Import statement to add (if needed)
	Import string `json:"import,omitempty"`

	// The fix pattern/example
	Pattern string `json:"pattern"`

	// Priority for ordering (higher = more preferred)
	Priority int `json:"priority,omitempty"`
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

	// Source code location (if source maps available)
	Source *SourceLocation `json:"source,omitempty"`
}

// BoundingBox represents element position and size.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// SourceLocation maps a DOM element to its source code location.
// Enables coding agents to locate and fix issues in source files.
type SourceLocation struct {
	// Source file path relative to project root
	File string `json:"file"` // e.g., "src/components/Hero.tsx"

	// Line number in source file (1-indexed)
	Line int `json:"line"`

	// Column number in source file (1-indexed)
	Column int `json:"column,omitempty"`

	// Component name (if framework detected)
	Component string `json:"component,omitempty"` // e.g., "Hero"

	// Detected framework
	Framework Framework `json:"framework,omitempty"`

	// Source map file used for mapping (for debugging)
	SourceMapFile string `json:"sourceMapFile,omitempty"`

	// Confidence that mapping is accurate (0-1)
	Confidence float64 `json:"confidence,omitempty"`
}

// Framework identifies the frontend framework in use.
type Framework string

const (
	FrameworkReact   Framework = "react"
	FrameworkVue     Framework = "vue"
	FrameworkSvelte  Framework = "svelte"
	FrameworkAngular Framework = "angular"
	FrameworkVanilla Framework = "vanilla"
	FrameworkUnknown Framework = "unknown"
)

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
