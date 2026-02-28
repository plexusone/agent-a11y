package journey

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	vibium "github.com/plexusone/vibium-go"
)

// Executor runs journey definitions using vibium for browser automation.
type Executor struct {
	vibe     *vibium.Vibe
	logger   *slog.Logger
	compiler *Compiler
	config   ExecutorConfig
}

// ExecutorConfig configures the executor.
type ExecutorConfig struct {
	// Default timeout for steps
	StepTimeout time.Duration

	// Screenshot on error
	ScreenshotOnError bool

	// Compile prompts before execution (for hybrid/agentic modes)
	CompilePrompts bool
}

// NewExecutor creates a new journey executor.
// The compiler parameter is optional - pass nil if LLM compilation is not needed.
func NewExecutor(vibe *vibium.Vibe, compiler *Compiler, logger *slog.Logger) *Executor {
	return &Executor{
		vibe:     vibe,
		logger:   logger,
		compiler: compiler,
		config: ExecutorConfig{
			StepTimeout:       30 * time.Second,
			ScreenshotOnError: true,
		},
	}
}

// Execute runs a journey definition and returns the execution state.
func (e *Executor) Execute(ctx context.Context, def *Definition) (*ExecutionState, error) {
	state := &ExecutionState{
		Variables:   make(map[string]any),
		StartTime:   time.Now(),
		Status:      "running",
		StepResults: make([]StepExecutionResult, 0),
	}

	// Copy test data to variables
	for k, v := range def.TestData {
		state.Variables[k] = v
	}

	// For agentic/hybrid mode, compile prompts first if enabled
	var stepsToExecute []Step
	if def.Mode == ModeDeterministic {
		stepsToExecute = def.Steps
	} else if e.config.CompilePrompts && e.compiler != nil {
		compiled, err := e.compiler.CompileJourney(ctx, def)
		if err != nil {
			state.Status = "failed"
			state.Error = fmt.Sprintf("compilation failed: %v", err)
			return state, err
		}
		stepsToExecute = compiled.ToDefinition().Steps
	} else if def.Mode == ModeAgentic && def.Goal != "" {
		// Execute goal directly with LLM guidance
		return e.executeAgentic(ctx, def, state)
	} else {
		stepsToExecute = def.Steps
	}

	// Execute steps
	for i, step := range stepsToExecute {
		select {
		case <-ctx.Done():
			state.Status = "cancelled"
			state.Error = ctx.Err().Error()
			return state, ctx.Err()
		default:
		}

		state.CurrentStepIndex = i
		state.CurrentStepID = step.ID

		result, err := e.executeStep(ctx, step, state)
		state.StepResults = append(state.StepResults, result)

		if err != nil {
			if step.ContinueOnError {
				e.logger.Warn("step failed but continuing", "step", step.ID, "error", err)
				continue
			}

			// Handle with onError handler if available
			if def.OnError.Screenshot && e.config.ScreenshotOnError {
				if screenshot, screenshotErr := e.takeScreenshot(ctx); screenshotErr == nil {
					result.Screenshot = screenshot
				}
			}

			// Execute recovery steps
			if len(def.OnError.RecoverySteps) > 0 {
				e.logger.Info("executing recovery steps")
				for _, recoveryStep := range def.OnError.RecoverySteps {
					if _, recoveryErr := e.executeStep(ctx, recoveryStep, state); recoveryErr != nil {
						e.logger.Error("recovery step failed", "error", recoveryErr)
					}
				}
			}

			state.Status = "failed"
			state.Error = err.Error()
			return state, err
		}

		// Update page state
		state.CurrentURL = result.PageURL
		state.CurrentTitle = result.PageTitle
	}

	state.Status = "completed"
	state.Duration = time.Since(state.StartTime)
	return state, nil
}

// executeStep executes a single step and returns the result.
func (e *Executor) executeStep(ctx context.Context, step Step, state *ExecutionState) (StepExecutionResult, error) {
	result := StepExecutionResult{
		StepIndex: state.CurrentStepIndex,
		StepID:    step.ID,
		StepName:  step.Name,
		Action:    step.Action,
		StartTime: time.Now(),
		Status:    "running",
	}

	// Check condition
	if step.Condition != "" {
		condMet, err := e.evaluateCondition(step.Condition, state.Variables)
		if err != nil {
			result.Status = "failure"
			result.Error = fmt.Sprintf("condition evaluation failed: %v", err)
			return result, err
		}
		if !condMet {
			result.Status = "skipped"
			result.Duration = time.Since(result.StartTime)
			return result, nil
		}
	}

	// Resolve variables in step parameters
	resolvedStep := e.resolveVariables(step, state.Variables)

	// Set step timeout
	stepCtx := ctx
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	} else if e.config.StepTimeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, e.config.StepTimeout)
		defer cancel()
	}

	// Execute with retry if configured
	var execErr error
	maxAttempts := 1
	if step.Retry != nil {
		maxAttempts = step.Retry.MaxAttempts
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.Retries = attempt - 1

		execErr = e.doExecuteAction(stepCtx, resolvedStep)
		if execErr == nil {
			break
		}

		if attempt < maxAttempts {
			delay := step.Retry.Delay
			if step.Retry.Backoff > 0 {
				delay = time.Duration(float64(delay) * step.Retry.Backoff * float64(attempt))
			}
			e.logger.Debug("step failed, retrying", "step", step.ID, "attempt", attempt, "delay", delay)
			time.Sleep(delay)
		}
	}

	if execErr != nil {
		result.Status = "failure"
		result.Error = execErr.Error()
		result.Duration = time.Since(result.StartTime)

		if e.config.ScreenshotOnError {
			if screenshot, err := e.takeScreenshot(ctx); err == nil {
				result.Screenshot = screenshot
			}
		}

		return result, execErr
	}

	// Wait for element if specified
	if resolvedStep.WaitFor != "" {
		waitTimeout := resolvedStep.WaitTimeout
		if waitTimeout == 0 {
			waitTimeout = 10 * time.Second
		}
		if err := e.waitForSelector(stepCtx, resolvedStep.WaitFor, resolvedStep.WaitState, waitTimeout); err != nil {
			result.Status = "failure"
			result.Error = fmt.Sprintf("wait failed: %v", err)
			result.Duration = time.Since(result.StartTime)
			return result, err
		}
	}

	// Store result if configured
	if step.Store != "" && result.Output != nil {
		state.Variables[step.Store] = result.Output
	}

	// Get current page state
	if pageURL, err := e.vibe.URL(ctx); err == nil {
		result.PageURL = pageURL
	}
	result.PageTitle = e.getPageTitle(ctx)
	result.Status = "success"
	result.Duration = time.Since(result.StartTime)

	return result, nil
}

// doExecuteAction performs the actual browser action.
func (e *Executor) doExecuteAction(ctx context.Context, step Step) error {
	switch step.Action {
	case ActionNavigate:
		return e.vibe.Go(ctx, step.URL)

	case ActionClick:
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("element not found: %s: %w", step.Selector, err)
		}
		return el.Click(ctx, nil)

	case ActionType_, ActionFill:
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("element not found: %s: %w", step.Selector, err)
		}
		return el.Fill(ctx, step.Value, nil)

	case ActionSelect:
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("element not found: %s: %w", step.Selector, err)
		}
		return el.SelectOption(ctx, vibium.SelectOptionValues{Values: []string{step.Value}}, nil)

	case ActionCheck:
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("element not found: %s: %w", step.Selector, err)
		}
		return el.Check(ctx, nil)

	case ActionUncheck:
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("element not found: %s: %w", step.Selector, err)
		}
		return el.Uncheck(ctx, nil)

	case ActionHover:
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("element not found: %s: %w", step.Selector, err)
		}
		return el.Hover(ctx, nil)

	case ActionScroll:
		// Scroll to element
		if step.Selector != "" {
			el, err := e.vibe.Find(ctx, step.Selector, nil)
			if err != nil {
				return fmt.Errorf("element not found: %s: %w", step.Selector, err)
			}
			return el.ScrollIntoView(ctx, nil)
		}
		// Scroll by amount (use JavaScript)
		return nil

	case ActionPress:
		keyboard, err := e.vibe.Keyboard(ctx)
		if err != nil {
			return fmt.Errorf("failed to get keyboard: %w", err)
		}
		return keyboard.Press(ctx, step.Key)

	case ActionWait:
		if step.WaitFor != "" {
			return e.waitForSelector(ctx, step.WaitFor, step.WaitState, step.WaitTimeout)
		}
		// Fixed delay
		if step.WaitTimeout > 0 {
			time.Sleep(step.WaitTimeout)
		}
		return nil

	case ActionScreenshot:
		_, err := e.takeScreenshot(ctx)
		return err

	case ActionUpload:
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("element not found: %s: %w", step.Selector, err)
		}
		return el.SetFiles(ctx, []string{step.File}, nil)

	case ActionExtract:
		// Extract text from element
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("element not found: %s: %w", step.Selector, err)
		}
		text, err := el.Text(ctx)
		if err != nil {
			return err
		}
		e.logger.Debug("extracted text", "selector", step.Selector, "text", text)
		return nil

	case ActionAssert:
		el, err := e.vibe.Find(ctx, step.Selector, nil)
		if err != nil {
			return fmt.Errorf("assertion failed: element not found: %s", step.Selector)
		}
		if step.Value != "" {
			text, _ := el.Text(ctx)
			if !strings.Contains(text, step.Value) {
				return fmt.Errorf("assertion failed: expected text '%s' not found in '%s'", step.Value, text)
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown action: %s", step.Action)
	}
}

// waitForSelector waits for an element to reach a specific state.
func (e *Executor) waitForSelector(ctx context.Context, selector, state string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	// Use Find with timeout for waiting
	opts := &vibium.FindOptions{
		Timeout: timeout,
	}

	switch state {
	case "hidden":
		// For hidden state, poll until element is not found
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			_, err := e.vibe.Find(ctx, selector, nil)
			if err != nil {
				// Element not found = hidden
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
		return fmt.Errorf("element still visible after timeout")
	default:
		// For visible/attached states, wait for element to appear
		_, err := e.vibe.Find(ctx, selector, opts)
		return err
	}
}

// resolveVariables substitutes ${var} patterns in step parameters.
func (e *Executor) resolveVariables(step Step, vars map[string]any) Step {
	resolved := step

	// Pattern for ${varName} or ${varName.path}
	pattern := regexp.MustCompile(`\$\{([^}]+)\}`)

	resolveStr := func(s string) string {
		return pattern.ReplaceAllStringFunc(s, func(match string) string {
			varPath := match[2 : len(match)-1] // Remove ${ and }
			value := e.getVariableValue(varPath, vars)
			if value != nil {
				return fmt.Sprintf("%v", value)
			}
			return match // Keep original if not found
		})
	}

	resolved.Selector = resolveStr(step.Selector)
	resolved.Value = resolveStr(step.Value)
	resolved.URL = resolveStr(step.URL)
	resolved.WaitFor = resolveStr(step.WaitFor)
	resolved.File = resolveStr(step.File)

	return resolved
}

// getVariableValue retrieves a variable value, supporting dot notation.
func (e *Executor) getVariableValue(path string, vars map[string]any) any {
	parts := strings.Split(path, ".")
	current := any(vars)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			if val, ok := v[part]; ok {
				current = val
			} else {
				return nil
			}
		case map[string]string:
			if val, ok := v[part]; ok {
				return val
			}
			return nil
		default:
			return nil
		}
	}

	return current
}

// evaluateCondition evaluates a simple condition expression.
func (e *Executor) evaluateCondition(condition string, vars map[string]any) (bool, error) {
	// Simple implementation - expand as needed
	condition = strings.TrimSpace(condition)

	// Handle negation
	if strings.HasPrefix(condition, "!") {
		result, err := e.evaluateCondition(condition[1:], vars)
		return !result, err
	}

	// Variable existence check: ${varName}
	if strings.HasPrefix(condition, "${") && strings.HasSuffix(condition, "}") {
		varPath := condition[2 : len(condition)-1]
		value := e.getVariableValue(varPath, vars)
		if value == nil {
			return false, nil
		}
		// Truthy check
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return v != "", nil
		case int, int64, float64:
			return v != 0, nil
		default:
			return true, nil
		}
	}

	// TODO: Add comparison operators (==, !=, >, <, etc.)

	return true, nil
}

// executeAgentic executes a journey in agentic mode using LLM guidance.
func (e *Executor) executeAgentic(ctx context.Context, def *Definition, state *ExecutionState) (*ExecutionState, error) {
	if e.compiler == nil {
		state.Status = "failed"
		state.Error = "agentic mode requires LLM compiler"
		return state, fmt.Errorf("agentic mode requires LLM compiler")
	}

	// Build prompt for LLM to generate next action
	// This is a simplified implementation - a full agentic executor would
	// include screenshot analysis, DOM inspection, and iterative planning

	e.logger.Info("executing in agentic mode", "goal", def.Goal)

	// Compile the goal into steps
	agenticStep := Step{
		Prompt: def.Goal,
		Data:   def.TestData,
	}

	compiled, err := e.compiler.compileStep(ctx, agenticStep, def.TestData, 0)
	if err != nil {
		state.Status = "failed"
		state.Error = fmt.Sprintf("failed to compile goal: %v", err)
		return state, err
	}

	// Execute the compiled steps
	for i, step := range compiled.Steps {
		select {
		case <-ctx.Done():
			state.Status = "cancelled"
			return state, ctx.Err()
		default:
		}

		state.CurrentStepIndex = i
		state.CurrentStepID = step.ID

		result, err := e.executeStep(ctx, step, state)
		state.StepResults = append(state.StepResults, result)

		if err != nil {
			state.Status = "failed"
			state.Error = err.Error()
			return state, err
		}
	}

	state.Status = "completed"
	state.Duration = time.Since(state.StartTime)
	return state, nil
}

// takeScreenshot captures a screenshot of the current page.
func (e *Executor) takeScreenshot(ctx context.Context) (string, error) {
	data, err := e.vibe.Screenshot(ctx)
	if err != nil {
		return "", err
	}
	// Return as base64
	return string(data), nil
}

// getPageTitle returns the current page title.
func (e *Executor) getPageTitle(ctx context.Context) string {
	title, err := e.vibe.Title(ctx)
	if err != nil {
		return ""
	}
	return title
}

// ExecuteStep implements StepExecutor interface.
func (e *Executor) ExecuteStep(ctx context.Context, step Step) error {
	state := &ExecutionState{
		Variables: make(map[string]any),
	}
	_, err := e.executeStep(ctx, step, state)
	return err
}
