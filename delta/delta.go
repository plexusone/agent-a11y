// Package delta provides before/after comparison for accessibility audits.
// This enables the autonomous fix loop by verifying whether fixes worked.
package delta

import (
	"time"

	"github.com/plexusone/agent-a11y/types"
)

// Calculator compares audit results to produce validation deltas.
type Calculator struct {
	// IncludeResults controls whether full before/after results are included.
	// Set to false to reduce output size.
	IncludeResults bool
}

// NewCalculator creates a delta calculator with default options.
func NewCalculator() *Calculator {
	return &Calculator{
		IncludeResults: false, // Default to compact output
	}
}

// Compare compares before and after audit results.
func (c *Calculator) Compare(before, after *types.AgentResult) *types.ValidationDelta {
	delta := &types.ValidationDelta{
		BeforeTotal: len(before.Findings),
		AfterTotal:  len(after.Findings),
		Timestamp:   time.Now().Format(time.RFC3339),
		BaselineURL: before.URL,
	}

	if c.IncludeResults {
		delta.Before = before
		delta.After = after
	}

	// Build fingerprint maps
	beforeMap := make(map[string]types.AgentFinding)
	for _, f := range before.Findings {
		fp := f.Finding.Fingerprint()
		beforeMap[fp] = f
	}

	afterMap := make(map[string]types.AgentFinding)
	for _, f := range after.Findings {
		fp := f.Finding.Fingerprint()
		afterMap[fp] = f
	}

	// Find fixed issues (in before, not in after)
	for fp, finding := range beforeMap {
		if _, exists := afterMap[fp]; !exists {
			delta.Fixed = append(delta.Fixed, types.FixedFinding{
				Finding:     finding,
				Fingerprint: fp,
				Verified:    true, // Verified via fingerprint
				FixedBy:     inferFixPattern(finding),
			})
		}
	}

	// Find remaining issues (in both before and after)
	for fp, finding := range afterMap {
		if _, exists := beforeMap[fp]; exists {
			delta.Remaining = append(delta.Remaining, finding)
		}
	}

	// Find regressions (in after, not in before)
	for fp, finding := range afterMap {
		if _, exists := beforeMap[fp]; !exists {
			delta.Regressions = append(delta.Regressions, finding)
		}
	}

	// Calculate improvement percentage
	if delta.BeforeTotal > 0 {
		delta.Improvement = float64(len(delta.Fixed)-len(delta.Regressions)) / float64(delta.BeforeTotal) * 100
	}

	// Determine overall status
	delta.Status = calculateStatus(delta)

	return delta
}

// CompareWithExpected compares results and checks expected fixes.
func (c *Calculator) CompareWithExpected(before, after *types.AgentResult, expectedFixed []string) *types.ValidationDelta {
	delta := c.Compare(before, after)

	// Check if expected rules were fixed
	fixedRules := make(map[string]bool)
	for _, f := range delta.Fixed {
		fixedRules[f.Finding.Finding.RuleID] = true
	}

	// Add unmet expectations to metadata (could extend delta type)
	// For now, this is informational

	return delta
}

// calculateStatus determines the overall delta status.
func calculateStatus(delta *types.ValidationDelta) types.DeltaStatus {
	hasFixed := len(delta.Fixed) > 0
	hasRegressions := len(delta.Regressions) > 0
	hasRemaining := len(delta.Remaining) > 0

	switch {
	case delta.AfterTotal == 0:
		return types.DeltaStatusFixed
	case hasFixed && hasRegressions:
		return types.DeltaStatusMixed
	case hasRegressions:
		return types.DeltaStatusRegressed
	case hasFixed && !hasRemaining:
		return types.DeltaStatusFixed
	case hasFixed:
		return types.DeltaStatusImproved
	default:
		return types.DeltaStatusNoChange
	}
}

// inferFixPattern tries to determine which fix pattern resolved an issue.
func inferFixPattern(finding types.AgentFinding) string {
	// Look at the remediation patterns to infer what was likely used
	if len(finding.Remediation.FixPatterns) > 0 {
		// Return the highest priority pattern
		pattern := finding.Remediation.FixPatterns[0]
		return string(pattern.Type) + ":" + string(pattern.Action) + ":" + pattern.Target
	}
	return ""
}

// LoadResult loads an AgentResult from a JSON file.
func LoadResult(path string) (*types.AgentResult, error) {
	// Implementation will be in a separate file to avoid circular imports
	// For now, this is a placeholder
	return nil, nil
}
