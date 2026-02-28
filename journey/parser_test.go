package journey

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser returned nil")
	}
}

func TestParseYAML(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Test Journey
description: A test journey
mode: deterministic
steps:
  - id: step-1
    name: Navigate to home
    action: navigate
    url: https://example.com
  - id: step-2
    name: Click button
    action: click
    selector: "#submit"
`

	def, err := parser.ParseYAML(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if def.Name != "Test Journey" {
		t.Errorf("Name: got %q, want %q", def.Name, "Test Journey")
	}
	if def.Mode != ModeDeterministic {
		t.Errorf("Mode: got %q, want %q", def.Mode, ModeDeterministic)
	}
	if len(def.Steps) != 2 {
		t.Errorf("Steps count: got %d, want 2", len(def.Steps))
	}
	if def.Steps[0].Action != ActionNavigate {
		t.Errorf("Steps[0].Action: got %q, want %q", def.Steps[0].Action, ActionNavigate)
	}
}

func TestParseJSON(t *testing.T) {
	parser := NewParser()

	jsonData := `{
		"name": "Test Journey",
		"mode": "deterministic",
		"steps": [
			{
				"id": "step-1",
				"action": "navigate",
				"url": "https://example.com"
			}
		]
	}`

	def, err := parser.ParseJSON(strings.NewReader(jsonData))
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	if def.Name != "Test Journey" {
		t.Errorf("Name: got %q, want %q", def.Name, "Test Journey")
	}
	if len(def.Steps) != 1 {
		t.Errorf("Steps count: got %d, want 1", len(def.Steps))
	}
}

func TestParseBytes(t *testing.T) {
	parser := NewParser()

	// Test YAML
	yamlData := []byte(`
name: Bytes Journey
mode: deterministic
steps:
  - action: navigate
    url: https://example.com
`)

	def, err := parser.ParseBytes(yamlData)
	if err != nil {
		t.Fatalf("ParseBytes(YAML) failed: %v", err)
	}
	if def.Name != "Bytes Journey" {
		t.Errorf("Name: got %q, want %q", def.Name, "Bytes Journey")
	}

	// Test JSON
	jsonData := []byte(`{"name": "JSON Journey", "mode": "deterministic", "steps": [{"action": "navigate", "url": "https://example.com"}]}`)
	def, err = parser.ParseBytes(jsonData)
	if err != nil {
		t.Fatalf("ParseBytes(JSON) failed: %v", err)
	}
	if def.Name != "JSON Journey" {
		t.Errorf("Name: got %q, want %q", def.Name, "JSON Journey")
	}
}

func TestParseFile(t *testing.T) {
	parser := NewParser()
	tmpDir := t.TempDir()

	// Test YAML file
	yamlPath := filepath.Join(tmpDir, "journey.yaml")
	yamlContent := `
name: File Journey
mode: deterministic
steps:
  - action: navigate
    url: https://example.com
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	def, err := parser.ParseFile(yamlPath)
	if err != nil {
		t.Fatalf("ParseFile(YAML) failed: %v", err)
	}
	if def.Name != "File Journey" {
		t.Errorf("Name: got %q, want %q", def.Name, "File Journey")
	}

	// Test JSON file
	jsonPath := filepath.Join(tmpDir, "journey.json")
	jsonContent := `{"name": "JSON File Journey", "mode": "deterministic", "steps": [{"action": "navigate", "url": "https://example.com"}]}`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}

	def, err = parser.ParseFile(jsonPath)
	if err != nil {
		t.Fatalf("ParseFile(JSON) failed: %v", err)
	}
	if def.Name != "JSON File Journey" {
		t.Errorf("Name: got %q, want %q", def.Name, "JSON File Journey")
	}
}

func TestParseFileUnsupportedFormat(t *testing.T) {
	parser := NewParser()
	tmpDir := t.TempDir()

	txtPath := filepath.Join(tmpDir, "journey.txt")
	if err := os.WriteFile(txtPath, []byte("name: test"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err := parser.ParseFile(txtPath)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported file format") {
		t.Errorf("Error should mention unsupported format: %v", err)
	}
}

func TestParseFileNotFound(t *testing.T) {
	parser := NewParser()

	_, err := parser.ParseFile("/nonexistent/path/journey.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestValidateMissingName(t *testing.T) {
	parser := NewParser()

	yaml := `
mode: deterministic
steps:
  - action: navigate
    url: https://example.com
`
	_, err := parser.ParseYAML(strings.NewReader(yaml))
	if err == nil {
		t.Error("Expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("Error should mention name is required: %v", err)
	}
}

func TestValidateInvalidMode(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Test
mode: invalid
steps:
  - action: navigate
    url: https://example.com
`
	_, err := parser.ParseYAML(strings.NewReader(yaml))
	if err == nil {
		t.Error("Expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("Error should mention invalid mode: %v", err)
	}
}

func TestValidateDefaultMode(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Test
steps:
  - action: navigate
    url: https://example.com
`
	def, err := parser.ParseYAML(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	// Mode should default to deterministic
	if def.Mode != ModeDeterministic {
		t.Errorf("Mode should default to deterministic: got %q", def.Mode)
	}
}

func TestValidateAgenticMode(t *testing.T) {
	parser := NewParser()

	// Agentic mode without goal or steps should fail
	yaml := `
name: Test
mode: agentic
`
	_, err := parser.ParseYAML(strings.NewReader(yaml))
	if err == nil {
		t.Error("Expected error for agentic mode without goal")
	}

	// Agentic mode with goal should pass
	yamlWithGoal := `
name: Test
mode: agentic
goal: Complete the checkout process
`
	def, err := parser.ParseYAML(strings.NewReader(yamlWithGoal))
	if err != nil {
		t.Fatalf("ParseYAML with goal failed: %v", err)
	}
	if def.Goal != "Complete the checkout process" {
		t.Errorf("Goal mismatch: got %q", def.Goal)
	}
}

func TestValidateDeterministicMode(t *testing.T) {
	parser := NewParser()

	// Deterministic mode without steps should fail
	yaml := `
name: Test
mode: deterministic
`
	_, err := parser.ParseYAML(strings.NewReader(yaml))
	if err == nil {
		t.Error("Expected error for deterministic mode without steps")
	}
	if !strings.Contains(err.Error(), "requires steps") {
		t.Errorf("Error should mention requires steps: %v", err)
	}
}

func TestValidateStepWithoutAction(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Test
mode: deterministic
steps:
  - selector: "#test"
`
	_, err := parser.ParseYAML(strings.NewReader(yaml))
	if err == nil {
		t.Error("Expected error for step without action")
	}
	if !strings.Contains(err.Error(), "requires action or prompt") {
		t.Errorf("Error should mention requires action: %v", err)
	}
}

func TestAutoAssignStepIDs(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Test
mode: deterministic
steps:
  - action: navigate
    url: https://example.com
  - action: click
    selector: "#submit"
`
	def, err := parser.ParseYAML(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	// IDs should be auto-assigned
	if def.Steps[0].ID != "step-1" {
		t.Errorf("Steps[0].ID: got %q, want %q", def.Steps[0].ID, "step-1")
	}
	if def.Steps[1].ID != "step-2" {
		t.Errorf("Steps[1].ID: got %q, want %q", def.Steps[1].ID, "step-2")
	}
}

func TestValidateSteps(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		steps       []Step
		expectError bool
		errorField  string
	}{
		{
			name: "valid navigate step",
			steps: []Step{
				{Action: ActionNavigate, URL: "https://example.com"},
			},
			expectError: false,
		},
		{
			name: "navigate without url",
			steps: []Step{
				{Action: ActionNavigate},
			},
			expectError: true,
			errorField:  "url",
		},
		{
			name: "valid click step",
			steps: []Step{
				{Action: ActionClick, Selector: "#button"},
			},
			expectError: false,
		},
		{
			name: "click without selector",
			steps: []Step{
				{Action: ActionClick},
			},
			expectError: true,
			errorField:  "selector",
		},
		{
			name: "valid type step",
			steps: []Step{
				{Action: ActionType_, Selector: "#input", Value: "test"},
			},
			expectError: false,
		},
		{
			name: "type without selector",
			steps: []Step{
				{Action: ActionType_, Value: "test"},
			},
			expectError: true,
			errorField:  "selector",
		},
		{
			name: "type without value",
			steps: []Step{
				{Action: ActionType_, Selector: "#input"},
			},
			expectError: true,
			errorField:  "value",
		},
		{
			name: "upload without file",
			steps: []Step{
				{Action: ActionUpload, Selector: "#file"},
			},
			expectError: true,
			errorField:  "file",
		},
		{
			name: "press without key",
			steps: []Step{
				{Action: ActionPress},
			},
			expectError: true,
			errorField:  "key",
		},
		{
			name: "step without action or prompt",
			steps: []Step{
				{Selector: "#test"},
			},
			expectError: true,
			errorField:  "action/prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := parser.ValidateSteps(tt.steps)
			if tt.expectError {
				if len(errors) == 0 {
					t.Error("Expected validation error")
				} else if tt.errorField != "" {
					found := false
					for _, e := range errors {
						if strings.Contains(e.Field, tt.errorField) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for field %q, got: %v", tt.errorField, errors)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("Unexpected errors: %v", errors)
				}
			}
		})
	}
}

func TestValidationErrorString(t *testing.T) {
	err := ValidationError{
		StepIndex: 0,
		StepID:    "step-1",
		Field:     "url",
		Message:   "navigate action requires url",
	}

	expected := "step 0 (step-1): url - navigate action requires url"
	if err.Error() != expected {
		t.Errorf("Error string: got %q, want %q", err.Error(), expected)
	}
}

func TestParseHybridMode(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Hybrid Journey
mode: hybrid
goal: Complete user registration
steps:
  - action: navigate
    url: https://example.com/register
  - prompt: Fill out the registration form with test user data
  - action: click
    selector: "#submit"
`
	def, err := parser.ParseYAML(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if def.Mode != ModeHybrid {
		t.Errorf("Mode: got %q, want %q", def.Mode, ModeHybrid)
	}
	if len(def.Steps) != 3 {
		t.Errorf("Steps count: got %d, want 3", len(def.Steps))
	}
	if def.Steps[1].Prompt == "" {
		t.Error("Step 2 should have a prompt")
	}
}

func TestParseWithTestData(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Test with Data
mode: deterministic
testData:
  username: testuser
  password: testpass
  email: test@example.com
steps:
  - action: navigate
    url: https://example.com
`
	def, err := parser.ParseYAML(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if def.TestData == nil {
		t.Fatal("TestData should not be nil")
	}
	if def.TestData["username"] != "testuser" {
		t.Errorf("TestData[username]: got %v", def.TestData["username"])
	}
}

func TestParseWithAuditPoints(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Journey with Audits
mode: deterministic
steps:
  - action: navigate
    url: https://example.com
auditPoints:
  - name: Home Page Audit
    afterStep: step-1
    screenshot: true
    categories:
      - color-contrast
      - alternative-text
`
	def, err := parser.ParseYAML(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if len(def.AuditPoints) != 1 {
		t.Errorf("AuditPoints count: got %d, want 1", len(def.AuditPoints))
	}
	if def.AuditPoints[0].Name != "Home Page Audit" {
		t.Errorf("AuditPoints[0].Name: got %q", def.AuditPoints[0].Name)
	}
	if !def.AuditPoints[0].Screenshot {
		t.Error("AuditPoints[0].Screenshot should be true")
	}
}

func TestParseWithRetryConfig(t *testing.T) {
	parser := NewParser()

	yaml := `
name: Journey with Retry
mode: deterministic
steps:
  - action: click
    selector: "#submit"
    retry:
      maxAttempts: 3
      delay: 1s
      backoff: 2.0
`
	def, err := parser.ParseYAML(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if def.Steps[0].Retry == nil {
		t.Fatal("Retry config should not be nil")
	}
	if def.Steps[0].Retry.MaxAttempts != 3 {
		t.Errorf("Retry.MaxAttempts: got %d, want 3", def.Steps[0].Retry.MaxAttempts)
	}
	if def.Steps[0].Retry.Backoff != 2.0 {
		t.Errorf("Retry.Backoff: got %v, want 2.0", def.Steps[0].Retry.Backoff)
	}
}

func TestParseInvalidYAML(t *testing.T) {
	parser := NewParser()

	invalidYAML := `
name: Test
steps:
  - action: [invalid
`
	_, err := parser.ParseYAML(strings.NewReader(invalidYAML))
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	parser := NewParser()

	invalidJSON := `{"name": "Test", "steps": [}`
	_, err := parser.ParseJSON(strings.NewReader(invalidJSON))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
