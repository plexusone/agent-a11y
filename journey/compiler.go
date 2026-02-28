// Package journey provides the compiler for converting prompts to deterministic steps.
package journey

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/plexusone/omnillm"
)

// Compiler converts prompt-based journey steps to deterministic steps.
// This enables the "compile once, run many" pattern for reproducible audits.
type Compiler struct {
	client *omnillm.ChatClient
	model  string
	logger *slog.Logger
}

// NewCompiler creates a new journey compiler.
func NewCompiler(client *omnillm.ChatClient, model string, logger *slog.Logger) *Compiler {
	return &Compiler{
		client: client,
		model:  model,
		logger: logger,
	}
}

// CompileJourney compiles all prompt-based steps in a journey to deterministic steps.
func (c *Compiler) CompileJourney(ctx context.Context, def *Definition) (*CompiledJourney, error) {
	result := &CompiledJourney{
		OriginalDefinition: def,
		CompiledAt:         time.Now(),
		Model:              c.model,
		Steps:              make([]CompiledStep, 0),
	}

	// Process each step
	for i, step := range def.Steps {
		if step.Prompt != "" {
			// This step needs compilation
			compiled, err := c.compileStep(ctx, step, def.TestData, i)
			if err != nil {
				return nil, fmt.Errorf("failed to compile step %d: %w", i, err)
			}
			result.Steps = append(result.Steps, *compiled)
		} else {
			// Already deterministic, wrap it
			result.Steps = append(result.Steps, CompiledStep{
				OriginalPrompt: "",
				Steps:          []Step{step},
				CompiledAt:     time.Now(),
				Verified:       true, // Already deterministic
				VerifyStatus:   "deterministic",
			})
		}
	}

	return result, nil
}

// CompiledJourney represents a fully compiled journey.
type CompiledJourney struct {
	OriginalDefinition *Definition    `json:"originalDefinition"`
	CompiledAt         time.Time      `json:"compiledAt"`
	Model              string         `json:"model"`
	Steps              []CompiledStep `json:"steps"`
}

// ToDefinition converts a compiled journey back to a Definition with all deterministic steps.
func (cj *CompiledJourney) ToDefinition() *Definition {
	newDef := *cj.OriginalDefinition
	newDef.Mode = ModeDeterministic
	newDef.Steps = make([]Step, 0)

	for _, compiled := range cj.Steps {
		newDef.Steps = append(newDef.Steps, compiled.Steps...)
	}

	return &newDef
}

// compileStep compiles a single prompt-based step to deterministic steps.
func (c *Compiler) compileStep(ctx context.Context, step Step, testData map[string]any, index int) (*CompiledStep, error) {
	prompt := c.buildCompilationPrompt(step, testData)

	maxTokens := 2048
	temperature := 0.1

	req := &omnillm.ChatCompletionRequest{
		Model: c.model,
		Messages: []omnillm.Message{
			{Role: omnillm.RoleSystem, Content: systemPromptForCompilation},
			{Role: omnillm.RoleUser, Content: prompt},
		},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	response := resp.Choices[0].Message.Content

	// Parse the response into steps
	steps, err := c.parseCompilationResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compilation response: %w", err)
	}

	// Preserve step metadata
	for i := range steps {
		if steps[i].ID == "" {
			steps[i].ID = fmt.Sprintf("compiled-%d-%d", index, i)
		}
		// Inherit audit settings
		if step.Audit && i == len(steps)-1 {
			steps[i].Audit = true
			steps[i].AuditName = step.AuditName
		}
	}

	return &CompiledStep{
		OriginalPrompt: step.Prompt,
		Steps:          steps,
		CompiledAt:     time.Now(),
		Model:          c.model,
		Verified:       false,
	}, nil
}

func (c *Compiler) buildCompilationPrompt(step Step, testData map[string]any) string {
	var sb strings.Builder

	sb.WriteString("Convert this user intent into deterministic browser automation steps.\n\n")
	sb.WriteString("User Intent:\n")
	sb.WriteString(step.Prompt)
	sb.WriteString("\n\n")

	if step.Instructions != "" {
		sb.WriteString("Additional Instructions:\n")
		sb.WriteString(step.Instructions)
		sb.WriteString("\n\n")
	}

	if len(testData) > 0 {
		sb.WriteString("Available Test Data:\n")
		for k, v := range testData {
			fmt.Fprintf(&sb, "- %s: %v\n", k, v)
		}
		sb.WriteString("\n")
	}

	if len(step.Data) > 0 {
		sb.WriteString("Step-specific Data:\n")
		for k, v := range step.Data {
			fmt.Fprintf(&sb, "- %s: %v\n", k, v)
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Output a JSON array of steps. Each step should have:
- action: one of [navigate, click, type, fill, select, check, uncheck, upload, scroll, hover, press, wait, screenshot, assert, extract]
- selector: CSS selector (required for element actions)
- value: input value (for type/fill/select)
- url: URL (for navigate)
- key: key name (for press)
- waitFor: selector to wait for after action
- name: human-readable step description

Example output:
[
  {"action": "navigate", "url": "https://example.com/login", "name": "Go to login page"},
  {"action": "fill", "selector": "#username", "value": "${username}", "name": "Enter username"},
  {"action": "fill", "selector": "#password", "value": "${password}", "name": "Enter password"},
  {"action": "click", "selector": "button[type=submit]", "waitFor": ".dashboard", "name": "Submit login form"}
]

Respond with ONLY the JSON array, no explanation.`)

	return sb.String()
}

const systemPromptForCompilation = `You are an expert at converting natural language browser automation instructions into precise, deterministic steps.

Rules:
1. Use specific CSS selectors - prefer IDs, data-testid, or unique attributes
2. Include waitFor conditions after actions that cause navigation or async loading
3. Use ${variable} syntax for test data substitution
4. Break complex actions into atomic steps
5. Add descriptive names to each step
6. Consider SPA behavior - wait for elements to be visible after navigation
7. For forms, fill fields in order and wait for validation
8. Always output valid JSON array

Common selectors:
- IDs: #login-btn
- Test IDs: [data-testid="submit"]
- ARIA: [aria-label="Close"]
- Classes: .btn-primary (less preferred)
- Text: button:has-text("Submit") (for Playwright-style)

Output ONLY the JSON array.`

// parseCompilationResponse parses the LLM response into steps.
func (c *Compiler) parseCompilationResponse(response string) ([]Step, error) {
	// Clean up response - sometimes LLMs add markdown code blocks
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var rawSteps []map[string]any
	if err := json.Unmarshal([]byte(response), &rawSteps); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	steps := make([]Step, 0, len(rawSteps))
	for _, raw := range rawSteps {
		step := Step{}

		if action, ok := raw["action"].(string); ok {
			step.Action = ActionType(action)
		}
		if selector, ok := raw["selector"].(string); ok {
			step.Selector = selector
		}
		if value, ok := raw["value"].(string); ok {
			step.Value = value
		}
		if url, ok := raw["url"].(string); ok {
			step.URL = url
		}
		if key, ok := raw["key"].(string); ok {
			step.Key = key
		}
		if waitFor, ok := raw["waitFor"].(string); ok {
			step.WaitFor = waitFor
		}
		if name, ok := raw["name"].(string); ok {
			step.Name = name
		}

		steps = append(steps, step)
	}

	return steps, nil
}

// VerifyCompiledStep executes a compiled step and compares against expected behavior.
func (c *Compiler) VerifyCompiledStep(ctx context.Context, compiled *CompiledStep, executor StepExecutor) error {
	// Execute the compiled steps
	for _, step := range compiled.Steps {
		if err := executor.ExecuteStep(ctx, step); err != nil {
			compiled.Verified = false
			compiled.VerifyStatus = fmt.Sprintf("failure: %v", err)
			return err
		}
	}

	compiled.Verified = true
	compiled.VerifiedAt = time.Now()
	compiled.VerifyStatus = "success"
	return nil
}

// StepExecutor executes individual steps (implemented by journey executor).
type StepExecutor interface {
	ExecuteStep(ctx context.Context, step Step) error
}
