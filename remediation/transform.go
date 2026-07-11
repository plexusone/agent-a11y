package remediation

import (
	"strings"
	"time"

	"github.com/plexusone/agent-a11y/types"
)

// Transformer converts standard findings to agent-optimized format.
type Transformer struct {
	patterns     *PatternRegistry
	designSystem *DesignSystemLoader
}

// NewTransformer creates a new transformer.
func NewTransformer(designSystemPath string) (*Transformer, error) {
	t := &Transformer{
		patterns: NewPatternRegistry(),
	}

	// Load design system if path provided
	if designSystemPath != "" {
		ds, err := LoadDesignSystem(designSystemPath)
		if err != nil {
			// Non-fatal: continue without design system
			// Log warning in production
		} else {
			t.designSystem = ds
		}
	}

	return t, nil
}

// TransformFindings converts findings to agent-optimized format.
func (t *Transformer) TransformFindings(findings []types.Finding) []types.AgentFinding {
	result := make([]types.AgentFinding, 0, len(findings))

	for _, f := range findings {
		af := t.transformFinding(f)
		result = append(result, af)
	}

	return result
}

// TransformResult converts a full result to agent format.
func (t *Transformer) TransformResult(
	url string,
	level string,
	duration time.Duration,
	findings []types.Finding,
) *types.AgentResult {
	agentFindings := t.TransformFindings(findings)

	result := &types.AgentResult{
		URL:       url,
		Timestamp: time.Now().Format(time.RFC3339),
		Duration:  duration.String(),
		Level:     level,
		Findings:  agentFindings,
		Summary:   t.calculateSummary(agentFindings),
		Status:    t.calculateStatus(agentFindings),
	}

	if t.designSystem != nil {
		result.DesignSystem = t.designSystem.Info()
	}

	return result
}

func (t *Transformer) transformFinding(f types.Finding) types.AgentFinding {
	// Build element context
	element := types.ElementContext{
		Selector: f.Selector,
		XPath:    f.XPath,
		HTML:     f.HTML,
		TagName:  f.Element,
	}

	// Try to detect design system component
	if t.designSystem != nil {
		classes := extractClasses(f.HTML)
		compID, variant := t.designSystem.DetectComponent(f.HTML, f.Element, classes)
		if compID != "" {
			element.ComponentID = compID
			element.ComponentVariant = variant
		}
	}

	// Build agent remediation
	remediation := t.buildAgentRemediation(f)

	return types.AgentFinding{
		Finding:     f,
		Element:     element,
		Remediation: remediation,
	}
}

func (t *Transformer) buildAgentRemediation(f types.Finding) types.AgentRemediation {
	ar := types.AgentRemediation{
		FixConfidence: 0.8, // Default confidence
	}

	// Copy standard remediation if present
	if f.Remediation != nil {
		ar.Remediation = *f.Remediation
	}

	// Get fix patterns for the rule
	patterns := t.patterns.GetPatterns(f.RuleID)
	if patterns != nil {
		ar.FixPatterns = patterns
	}

	// Also check technique-based patterns
	if f.Remediation != nil {
		for _, tech := range f.Remediation.Techniques {
			techPatterns := t.patterns.GetPatternsForTechnique(tech.ID)
			// Append unique patterns (range over nil is safe)
			for _, tp := range techPatterns {
				if !containsPattern(ar.FixPatterns, tp) {
					ar.FixPatterns = append(ar.FixPatterns, tp)
				}
			}
		}
	}

	// Add design system token suggestions
	if t.designSystem != nil {
		ar.TokenSuggestions = t.suggestTokens(f)
	}

	// Calculate fix confidence based on available guidance
	ar.FixConfidence = t.calculateFixConfidence(ar)

	return ar
}

func (t *Transformer) suggestTokens(f types.Finding) []types.TokenSuggestion {
	var suggestions []types.TokenSuggestion

	// Color contrast issues need color token suggestions
	if f.RuleID == "color-contrast" || f.RuleID == "color-contrast-enhanced" {
		// Try to extract colors from the finding
		// This would need actual computed styles from the audit
		// For now, provide generic guidance
		suggestion := t.designSystem.SuggestColorToken(
			"#999999",        // Would come from computed styles
			"color",          // Property
			"#FFFFFF",        // Background (would come from computed styles)
			4.5,              // AA contrast requirement
		)
		if suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	return suggestions
}

func (t *Transformer) calculateSummary(findings []types.AgentFinding) types.AgentSummary {
	summary := types.AgentSummary{
		Total: len(findings),
	}

	for _, f := range findings {
		// Count by impact
		switch f.Finding.Impact {
		case types.ImpactCritical, types.ImpactBlocker:
			summary.Critical++
		case types.ImpactSerious:
			summary.Serious++
		case types.ImpactModerate:
			summary.Moderate++
		case types.ImpactMinor:
			summary.Minor++
		}

		// Count fixable (has fix patterns)
		if len(f.Remediation.FixPatterns) > 0 {
			summary.Fixable++
		}

		// Count those needing tokens
		if len(f.Remediation.TokenSuggestions) > 0 {
			summary.NeedsToken++
		}
	}

	return summary
}

func (t *Transformer) calculateStatus(findings []types.AgentFinding) string {
	var critical, serious int

	for _, f := range findings {
		switch f.Finding.Impact {
		case types.ImpactCritical, types.ImpactBlocker:
			critical++
		case types.ImpactSerious:
			serious++
		}
	}

	if critical > 0 {
		return "NO-GO"
	}
	if serious > 0 {
		return "WARN"
	}
	return "GO"
}

func (t *Transformer) calculateFixConfidence(ar types.AgentRemediation) float64 {
	confidence := 0.5 // Base confidence

	// More fix patterns = higher confidence
	if len(ar.FixPatterns) > 0 {
		confidence += 0.2
	}
	if len(ar.FixPatterns) > 2 {
		confidence += 0.1
	}

	// Token suggestions help
	if len(ar.TokenSuggestions) > 0 {
		confidence += 0.1
	}

	// WCAG technique references help
	if len(ar.Techniques) > 0 {
		confidence += 0.1
	}

	// Cap at 0.95
	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
}

// Helper functions

func extractClasses(html string) []string {
	// Simple class extraction from HTML
	var classes []string

	// Find class="..." or class='...'
	for _, quote := range []string{`"`, `'`} {
		prefix := `class=` + quote
		idx := strings.Index(html, prefix)
		if idx >= 0 {
			start := idx + len(prefix)
			end := strings.Index(html[start:], quote)
			if end > 0 {
				classStr := html[start : start+end]
				classes = append(classes, strings.Fields(classStr)...)
			}
		}
	}

	return classes
}

func containsPattern(patterns []types.FixPattern, p types.FixPattern) bool {
	for _, existing := range patterns {
		if existing.Type == p.Type && existing.Target == p.Target && existing.Action == p.Action {
			return true
		}
	}
	return false
}
