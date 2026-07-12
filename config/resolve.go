package config

import (
	"github.com/plexusone/agent-a11y/types"
)

// ResolvedFix is a fix pattern with source information.
type ResolvedFix struct {
	// Source indicates where this fix came from
	Source string `json:"source"` // "project", "library", "language", "builtin"

	// Component is the component name (for project/library fixes)
	Component string `json:"component,omitempty"`

	// Import is the import statement needed
	Import string `json:"import,omitempty"`

	// Rule is the accessibility rule this fixes
	Rule string `json:"rule"`

	// Pattern is the fix pattern/example
	Pattern string `json:"pattern"`

	// Priority for ordering (higher = more preferred)
	Priority int `json:"priority,omitempty"`
}

// FixResolver resolves fix patterns from multiple sources.
type FixResolver struct {
	fixesConfig *FixesConfig
	projectCtx  *ProjectContext
}

// NewFixResolver creates a new fix resolver.
func NewFixResolver(fixesConfig *FixesConfig, projectCtx *ProjectContext) *FixResolver {
	return &FixResolver{
		fixesConfig: fixesConfig,
		projectCtx:  projectCtx,
	}
}

// ResolveFixes returns all applicable fix patterns for a rule.
// Resolution order: project > library > language > builtin
func (r *FixResolver) ResolveFixes(ruleID string, element *types.ElementContext) []ResolvedFix {
	var fixes []ResolvedFix

	// 1. Project-specific fixes (highest priority)
	if r.projectCtx != nil && r.projectCtx.ProjectConfig != nil {
		projectFixes := r.getProjectFixes(ruleID, element)
		fixes = append(fixes, projectFixes...)
	}

	// 2. Component library fixes
	if r.projectCtx != nil && r.projectCtx.LibraryConfig != nil {
		libraryFixes := r.getLibraryFixes(ruleID, element)
		fixes = append(fixes, libraryFixes...)
	}

	// 3. Language-specific fixes
	if r.projectCtx != nil && r.projectCtx.LanguageConfig != nil {
		langFixes := r.getLanguageFixes(ruleID)
		fixes = append(fixes, langFixes...)
	}

	return fixes
}

// getProjectFixes returns project-specific fixes for a rule.
func (r *FixResolver) getProjectFixes(ruleID string, element *types.ElementContext) []ResolvedFix {
	var fixes []ResolvedFix

	if r.projectCtx.ProjectConfig == nil {
		return fixes
	}

	// Check each component in project config
	for compName, compConfig := range r.projectCtx.ProjectConfig.Components {
		// Check if element matches this component
		if !matchesComponent(element, compConfig.Selectors) {
			continue
		}

		// Find fixes for this rule
		for _, fix := range compConfig.Fixes {
			if fix.Rule == ruleID {
				fixes = append(fixes, ResolvedFix{
					Source:    "project",
					Component: compName,
					Rule:      ruleID,
					Pattern:   fix.Pattern,
					Priority:  100 + fix.Priority, // Project fixes have highest base priority
				})
			}
		}
	}

	return fixes
}

// getLibraryFixes returns component library fixes for a rule.
func (r *FixResolver) getLibraryFixes(ruleID string, element *types.ElementContext) []ResolvedFix {
	var fixes []ResolvedFix

	if r.projectCtx.LibraryConfig == nil {
		return fixes
	}

	// Check each component in library
	for compName, compConfig := range r.projectCtx.LibraryConfig.Components {
		// For library fixes, match by component name in element
		if element != nil && element.ComponentID != "" {
			if element.ComponentID != compName {
				continue
			}
		}

		// Find fixes for this rule
		for _, fix := range compConfig.Fixes {
			if fix.Rule == ruleID {
				fixes = append(fixes, ResolvedFix{
					Source:    "library",
					Component: compName,
					Import:    compConfig.Import,
					Rule:      ruleID,
					Pattern:   fix.Pattern,
					Priority:  50 + fix.Priority, // Library fixes have medium priority
				})
			}
		}
	}

	return fixes
}

// getLanguageFixes returns language-specific fixes for a rule.
func (r *FixResolver) getLanguageFixes(ruleID string) []ResolvedFix {
	var fixes []ResolvedFix

	if r.projectCtx.LanguageConfig == nil {
		return fixes
	}

	// Check if there's a pattern for this rule
	if pattern, ok := r.projectCtx.LanguageConfig.Patterns[ruleID]; ok {
		fixes = append(fixes, ResolvedFix{
			Source:    "language",
			Component: pattern.Component,
			Import:    pattern.Import,
			Rule:      ruleID,
			Pattern:   pattern.Example,
			Priority:  25, // Language fixes have lower priority
		})
	}

	return fixes
}

// matchesComponent checks if an element matches component selectors.
func matchesComponent(element *types.ElementContext, selectors []string) bool {
	if element == nil {
		return false
	}

	for _, selector := range selectors {
		// Simple selector matching
		if matchSelector(element, selector) {
			return true
		}
	}

	return false
}

// matchSelector performs basic selector matching.
func matchSelector(element *types.ElementContext, selector string) bool {
	// Check if selector appears in element's selector
	if element.Selector != "" && containsSelector(element.Selector, selector) {
		return true
	}

	// Check if selector appears in HTML
	if element.HTML != "" && containsSelector(element.HTML, selector) {
		return true
	}

	return false
}

// containsSelector checks if a selector pattern exists in text.
func containsSelector(text, selector string) bool {
	// Handle class selectors
	if len(selector) > 0 && selector[0] == '.' {
		className := selector[1:]
		return containsClass(text, className)
	}

	// Handle data attribute selectors
	if len(selector) > 1 && selector[0] == '[' {
		return containsAttribute(text, selector)
	}

	// Direct string match
	return len(selector) > 0 && len(text) > 0 &&
		(text == selector || containsWord(text, selector))
}

// containsClass checks if text contains a CSS class.
func containsClass(text, className string) bool {
	// Look for class="... className ..."
	patterns := []string{
		`class="` + className + `"`,
		`class="` + className + ` `,
		` ` + className + `"`,
		` ` + className + ` `,
		`className="` + className + `"`,
		`className="` + className + ` `,
	}

	for _, pattern := range patterns {
		if len(text) >= len(pattern) {
			for i := 0; i <= len(text)-len(pattern); i++ {
				if text[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}

	return false
}

// containsAttribute checks if text contains a data attribute.
func containsAttribute(text, selector string) bool {
	// Extract attribute from [attr=value] or [attr]
	if len(selector) < 2 {
		return false
	}

	attr := selector[1 : len(selector)-1] // Remove [ and ]

	// Handle [attr=value]
	if idx := indexOf(attr, "="); idx > 0 {
		attrName := attr[:idx]
		attrValue := attr[idx+1:]
		// Remove quotes from value
		attrValue = trimQuotes(attrValue)

		pattern := attrName + `="` + attrValue + `"`
		return containsWord(text, pattern)
	}

	// Handle [attr]
	return containsWord(text, attr+"=")
}

// containsWord checks if text contains word.
func containsWord(text, word string) bool {
	return len(word) > 0 && len(text) >= len(word) &&
		indexOf(text, word) >= 0
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// trimQuotes removes surrounding quotes from a string.
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// MergeWithBuiltin merges resolved fixes with builtin fix patterns.
func MergeWithBuiltin(resolved []ResolvedFix, builtin []types.FixPattern) []ResolvedFix {
	// Convert builtin patterns to ResolvedFix
	for _, bp := range builtin {
		resolved = append(resolved, ResolvedFix{
			Source:   "builtin",
			Rule:     "", // Builtin patterns don't have rule association here
			Pattern:  bp.Example,
			Priority: 0, // Lowest priority
		})
	}

	// Sort by priority (highest first)
	sortByPriority(resolved)

	return resolved
}

// sortByPriority sorts fixes by priority descending.
func sortByPriority(fixes []ResolvedFix) {
	// Simple bubble sort (small arrays)
	for i := 0; i < len(fixes)-1; i++ {
		for j := 0; j < len(fixes)-i-1; j++ {
			if fixes[j].Priority < fixes[j+1].Priority {
				fixes[j], fixes[j+1] = fixes[j+1], fixes[j]
			}
		}
	}
}
