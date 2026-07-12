package delta

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plexusone/agent-a11y/types"
)

// LoadAgentResult loads an AgentResult from a JSON file.
func LoadAgentResult(path string) (*types.AgentResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result types.AgentResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// SaveDelta saves a ValidationDelta to a JSON file.
func SaveDelta(delta *types.ValidationDelta, path string) error {
	data, err := json.MarshalIndent(delta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal delta: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SaveAgentResult saves an AgentResult to a JSON file.
func SaveAgentResult(result *types.AgentResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
