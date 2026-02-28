package journey

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser parses journey definitions from various formats.
type Parser struct{}

// NewParser creates a new journey parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a journey definition from a file.
// Supports .yaml, .yml, and .json extensions.
func (p *Parser) ParseFile(path string) (*Definition, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open journey file: %w", err)
	}
	defer func() { _ = f.Close() }()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return p.ParseYAML(f)
	case ".json":
		return p.ParseJSON(f)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// ParseYAML parses a journey definition from YAML.
func (p *Parser) ParseYAML(r io.Reader) (*Definition, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML: %w", err)
	}

	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := p.validate(&def); err != nil {
		return nil, err
	}

	return &def, nil
}

// ParseJSON parses a journey definition from JSON.
func (p *Parser) ParseJSON(r io.Reader) (*Definition, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON: %w", err)
	}

	var def Definition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := p.validate(&def); err != nil {
		return nil, err
	}

	return &def, nil
}

// ParseBytes parses a journey definition from bytes with format detection.
func (p *Parser) ParseBytes(data []byte) (*Definition, error) {
	// Try YAML first (JSON is valid YAML)
	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		// Try JSON explicitly
		if jsonErr := json.Unmarshal(data, &def); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse as YAML or JSON: yaml=%v, json=%v", err, jsonErr)
		}
	}

	if err := p.validate(&def); err != nil {
		return nil, err
	}

	return &def, nil
}

// validate checks the journey definition for errors.
func (p *Parser) validate(def *Definition) error {
	if def.Name == "" {
		return fmt.Errorf("journey name is required")
	}

	if def.Mode == "" {
		// Default to deterministic if not specified
		def.Mode = ModeDeterministic
	}

	// Validate mode
	switch def.Mode {
	case ModeAgentic, ModeDeterministic, ModeHybrid:
		// Valid
	default:
		return fmt.Errorf("invalid mode: %s (must be agentic, deterministic, or hybrid)", def.Mode)
	}

	// Agentic mode requires goal or prompts
	if def.Mode == ModeAgentic {
		if def.Goal == "" && len(def.Steps) == 0 {
			return fmt.Errorf("agentic mode requires a goal or prompt-based steps")
		}
	}

	// Deterministic mode requires steps with actions
	if def.Mode == ModeDeterministic {
		if len(def.Steps) == 0 {
			return fmt.Errorf("deterministic mode requires steps")
		}
		for i, step := range def.Steps {
			if step.Action == "" && step.Prompt == "" {
				return fmt.Errorf("step %d: requires action or prompt", i)
			}
		}
	}

	// Assign IDs to steps if not provided
	for i := range def.Steps {
		if def.Steps[i].ID == "" {
			def.Steps[i].ID = fmt.Sprintf("step-%d", i+1)
		}
	}

	return nil
}

// ValidateSteps validates a list of steps.
func (p *Parser) ValidateSteps(steps []Step) []ValidationError {
	var errors []ValidationError

	for i, step := range steps {
		stepErrors := p.validateStep(i, step)
		errors = append(errors, stepErrors...)
	}

	return errors
}

// ValidationError represents a validation error.
type ValidationError struct {
	StepIndex int    `json:"stepIndex"`
	StepID    string `json:"stepId"`
	Field     string `json:"field"`
	Message   string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("step %d (%s): %s - %s", e.StepIndex, e.StepID, e.Field, e.Message)
}

func (p *Parser) validateStep(index int, step Step) []ValidationError {
	var errors []ValidationError
	stepID := step.ID
	if stepID == "" {
		stepID = fmt.Sprintf("step-%d", index)
	}

	// Check for valid action or prompt
	if step.Action == "" && step.Prompt == "" && len(step.Steps) == 0 {
		errors = append(errors, ValidationError{
			StepIndex: index,
			StepID:    stepID,
			Field:     "action/prompt",
			Message:   "step must have an action, prompt, or sub-steps",
		})
	}

	// Validate action-specific requirements
	switch step.Action {
	case ActionNavigate:
		if step.URL == "" {
			errors = append(errors, ValidationError{
				StepIndex: index,
				StepID:    stepID,
				Field:     "url",
				Message:   "navigate action requires url",
			})
		}
	case ActionClick, ActionSelect, ActionCheck, ActionUncheck, ActionHover:
		if step.Selector == "" {
			errors = append(errors, ValidationError{
				StepIndex: index,
				StepID:    stepID,
				Field:     "selector",
				Message:   fmt.Sprintf("%s action requires selector", step.Action),
			})
		}
	case ActionType_, ActionFill:
		if step.Selector == "" {
			errors = append(errors, ValidationError{
				StepIndex: index,
				StepID:    stepID,
				Field:     "selector",
				Message:   fmt.Sprintf("%s action requires selector", step.Action),
			})
		}
		if step.Value == "" {
			errors = append(errors, ValidationError{
				StepIndex: index,
				StepID:    stepID,
				Field:     "value",
				Message:   fmt.Sprintf("%s action requires value", step.Action),
			})
		}
	case ActionUpload:
		if step.File == "" {
			errors = append(errors, ValidationError{
				StepIndex: index,
				StepID:    stepID,
				Field:     "file",
				Message:   "upload action requires file",
			})
		}
	case ActionPress:
		if step.Key == "" {
			errors = append(errors, ValidationError{
				StepIndex: index,
				StepID:    stepID,
				Field:     "key",
				Message:   "press action requires key",
			})
		}
	}

	// Recursively validate sub-steps
	for i, subStep := range step.Steps {
		subErrors := p.validateStep(i, subStep)
		for _, e := range subErrors {
			e.StepID = fmt.Sprintf("%s.%s", stepID, e.StepID)
			errors = append(errors, e)
		}
	}

	return errors
}
