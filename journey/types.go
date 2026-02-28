// Package journey provides types and execution for user journey testing.
package journey

import (
	"time"
)

// Mode represents the journey execution mode.
type Mode string

const (
	ModeAgentic       Mode = "agentic"       // LLM navigates based on prompts
	ModeDeterministic Mode = "deterministic" // Fixed steps with selectors
	ModeHybrid        Mode = "hybrid"        // Mix of both
)

// Definition represents a complete journey definition.
type Definition struct {
	// Identification
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string `yaml:"version,omitempty" json:"version,omitempty"`

	// Execution mode
	Mode Mode `yaml:"mode" json:"mode"`

	// For agentic mode: high-level goal
	Goal string `yaml:"goal,omitempty" json:"goal,omitempty"`

	// Journey steps
	Steps []Step `yaml:"steps" json:"steps"`

	// Test data for variable substitution
	TestData map[string]any `yaml:"testData,omitempty" json:"testData,omitempty"`

	// Audit configuration
	AuditPoints []AuditPoint `yaml:"auditPoints,omitempty" json:"auditPoints,omitempty"`

	// Pre/post conditions
	Preconditions  []Condition `yaml:"preconditions,omitempty" json:"preconditions,omitempty"`
	Postconditions []Condition `yaml:"postconditions,omitempty" json:"postconditions,omitempty"`

	// Error handling
	OnError ErrorHandler `yaml:"onError,omitempty" json:"onError,omitempty"`

	// Timeout for entire journey
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Step represents a single step in a journey.
type Step struct {
	// Common fields
	ID   string `yaml:"id,omitempty" json:"id,omitempty"`
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// For deterministic mode
	Action   ActionType `yaml:"action,omitempty" json:"action,omitempty"`
	Selector string     `yaml:"selector,omitempty" json:"selector,omitempty"`
	Value    string     `yaml:"value,omitempty" json:"value,omitempty"`
	URL      string     `yaml:"url,omitempty" json:"url,omitempty"`
	Key      string     `yaml:"key,omitempty" json:"key,omitempty"` // For key press

	// For agentic mode
	Prompt       string `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Instructions string `yaml:"instructions,omitempty" json:"instructions,omitempty"`

	// For file upload/data input
	Data map[string]any `yaml:"data,omitempty" json:"data,omitempty"`
	File string         `yaml:"file,omitempty" json:"file,omitempty"`

	// Wait conditions
	WaitFor      string        `yaml:"waitFor,omitempty" json:"waitFor,omitempty"` // Selector to wait for
	WaitState    string        `yaml:"waitState,omitempty" json:"waitState,omitempty"` // visible, hidden, attached
	WaitTimeout  time.Duration `yaml:"waitTimeout,omitempty" json:"waitTimeout,omitempty"`

	// Audit trigger
	Audit     bool   `yaml:"audit,omitempty" json:"audit,omitempty"`
	AuditName string `yaml:"auditName,omitempty" json:"auditName,omitempty"`

	// Control flow
	Condition string `yaml:"if,omitempty" json:"if,omitempty"`
	UseAuth   bool   `yaml:"auth,omitempty" json:"auth,omitempty"` // Use configured auth

	// Error handling
	ContinueOnError bool          `yaml:"continueOnError,omitempty" json:"continueOnError,omitempty"`
	Retry           *RetryConfig  `yaml:"retry,omitempty" json:"retry,omitempty"`
	Timeout         time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// For storing results
	Store string `yaml:"store,omitempty" json:"store,omitempty"`

	// Sub-steps (for grouped operations)
	Steps []Step `yaml:"steps,omitempty" json:"steps,omitempty"`
}

// ActionType represents deterministic step actions.
type ActionType string

const (
	ActionNavigate   ActionType = "navigate"
	ActionClick      ActionType = "click"
	ActionType_      ActionType = "type" // Underscore to avoid conflict with Go keyword
	ActionFill       ActionType = "fill"
	ActionSelect     ActionType = "select"
	ActionCheck      ActionType = "check"
	ActionUncheck    ActionType = "uncheck"
	ActionUpload     ActionType = "upload"
	ActionScroll     ActionType = "scroll"
	ActionHover      ActionType = "hover"
	ActionPress      ActionType = "press"
	ActionWait       ActionType = "wait"
	ActionScreenshot ActionType = "screenshot"
	ActionAssert     ActionType = "assert"
	ActionExtract    ActionType = "extract"
)

// AuditPoint defines when and what to audit during a journey.
type AuditPoint struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// When to audit
	AfterStep string `yaml:"afterStep,omitempty" json:"afterStep,omitempty"` // After specific step ID
	WaitFor   string `yaml:"waitFor,omitempty" json:"waitFor,omitempty"`     // Wait for selector
	Condition string `yaml:"condition,omitempty" json:"condition,omitempty"` // Conditional expression

	// What to audit
	Categories []string `yaml:"categories,omitempty" json:"categories,omitempty"` // Empty = all
	Rules      []string `yaml:"rules,omitempty" json:"rules,omitempty"`           // Specific rules

	// Options
	FullPage   bool `yaml:"fullPage,omitempty" json:"fullPage,omitempty"`
	Screenshot bool `yaml:"screenshot,omitempty" json:"screenshot,omitempty"`
}

// Condition represents a pre/post condition.
type Condition struct {
	Description string `yaml:"description" json:"description"`
	Check       string `yaml:"check" json:"check"`       // Selector or expression
	Expected    string `yaml:"expected" json:"expected"` // Expected state
}

// RetryConfig specifies retry behavior.
type RetryConfig struct {
	MaxAttempts int           `yaml:"maxAttempts" json:"maxAttempts"`
	Delay       time.Duration `yaml:"delay" json:"delay"`
	Backoff     float64       `yaml:"backoff,omitempty" json:"backoff,omitempty"` // Multiplier
}

// ErrorHandler specifies error handling behavior.
type ErrorHandler struct {
	Screenshot  bool   `yaml:"screenshot" json:"screenshot"`
	ContinueOn  string `yaml:"continueOn,omitempty" json:"continueOn,omitempty"` // error types to continue on
	RecoverySteps []Step `yaml:"recoverySteps,omitempty" json:"recoverySteps,omitempty"`
}

// CompiledStep represents a deterministic step compiled from a prompt.
type CompiledStep struct {
	// Original prompt that generated this
	OriginalPrompt string `json:"originalPrompt"`

	// Compiled deterministic steps
	Steps []Step `json:"steps"`

	// Compilation metadata
	CompiledAt time.Time `json:"compiledAt"`
	Model      string    `json:"model"`

	// Verification status
	Verified     bool      `json:"verified"`
	VerifiedAt   time.Time `json:"verifiedAt,omitempty"`
	VerifyStatus string    `json:"verifyStatus,omitempty"` // success, failure, partial
}

// ExecutionState tracks journey execution state.
type ExecutionState struct {
	// Current position
	CurrentStepIndex int    `json:"currentStepIndex"`
	CurrentStepID    string `json:"currentStepId"`

	// Variables accumulated during execution
	Variables map[string]any `json:"variables"`

	// Page state
	CurrentURL   string `json:"currentUrl"`
	CurrentTitle string `json:"currentTitle"`

	// Execution status
	Status    string        `json:"status"` // running, paused, completed, failed
	Error     string        `json:"error,omitempty"`
	StartTime time.Time     `json:"startTime"`
	Duration  time.Duration `json:"duration"`

	// Step results
	StepResults []StepExecutionResult `json:"stepResults"`
}

// StepExecutionResult contains the result of executing a step.
type StepExecutionResult struct {
	StepIndex int           `json:"stepIndex"`
	StepID    string        `json:"stepId"`
	StepName  string        `json:"stepName"`
	Action    ActionType    `json:"action"`
	Status    string        `json:"status"` // success, failure, skipped
	Error     string        `json:"error,omitempty"`
	StartTime time.Time     `json:"startTime"`
	Duration  time.Duration `json:"duration"`
	Retries   int           `json:"retries"`

	// Output/extracted data
	Output any `json:"output,omitempty"`

	// Page state after step
	PageURL    string `json:"pageUrl"`
	PageTitle  string `json:"pageTitle"`
	Screenshot string `json:"screenshot,omitempty"`

	// If audit was triggered
	AuditResult any `json:"auditResult,omitempty"` // Typed as any to avoid circular import
}
