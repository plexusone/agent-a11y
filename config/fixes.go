// Package config provides configuration loading and management for agent-a11y.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FixesConfig is the root configuration for project-specific fixes.
// Loaded from ~/.plexusone/a11y/fixes.yaml
type FixesConfig struct {
	Version   string                      `yaml:"version" json:"version"`
	Defaults  FixDefaults                 `yaml:"defaults" json:"defaults"`
	Languages map[string]LanguageConfig   `yaml:"languages" json:"languages"`
	Projects  []ProjectConfig             `yaml:"projects" json:"projects"`
	Libraries map[string]ComponentLibrary `yaml:"component-libraries" json:"componentLibraries"`
}

// FixDefaults contains global default settings.
type FixDefaults struct {
	Extends string            `yaml:"extends" json:"extends"` // "builtin" or path to base config
	Tokens  map[string][]string `yaml:"tokens" json:"tokens"`
}

// LanguageConfig contains language/framework-specific patterns.
type LanguageConfig struct {
	Patterns    map[string]LanguagePattern `yaml:"patterns" json:"patterns"`
	Conventions LanguageConventions        `yaml:"conventions" json:"conventions"`
}

// LanguagePattern is a fix pattern for a specific language.
type LanguagePattern struct {
	Component string `yaml:"component" json:"component,omitempty"`
	Import    string `yaml:"import" json:"import,omitempty"`
	Example   string `yaml:"example" json:"example"`
}

// LanguageConventions describes language-specific conventions.
type LanguageConventions struct {
	AriaAttributes string `yaml:"aria-attributes" json:"ariaAttributes"` // "camelCase" or "kebab-case"
	EventHandlers  string `yaml:"event-handlers" json:"eventHandlers"`   // "onKeyDown" or "@keydown"
}

// ProjectConfig contains project-specific overrides.
type ProjectConfig struct {
	Match        string                        `yaml:"match" json:"match"`                 // Glob pattern for repo path
	Language     string                        `yaml:"language" json:"language"`           // react, vue, svelte, go-templates
	Framework    string                        `yaml:"framework" json:"framework"`         // next, nuxt, sveltekit
	DesignSystem string                        `yaml:"design-system" json:"designSystem"`  // Path to design system spec
	AutoDetect   bool                          `yaml:"auto-detect" json:"autoDetect"`      // Auto-detect language
	Components   map[string]ComponentFixConfig `yaml:"components" json:"components"`
}

// ComponentFixConfig contains fix patterns for a specific component.
type ComponentFixConfig struct {
	Selectors []string        `yaml:"selectors" json:"selectors"`
	Fixes     []ComponentFix  `yaml:"fixes" json:"fixes"`
}

// ComponentFix is a fix pattern for a component.
type ComponentFix struct {
	Rule     string `yaml:"rule" json:"rule"`
	Pattern  string `yaml:"pattern" json:"pattern"`
	Priority int    `yaml:"priority" json:"priority,omitempty"`
}

// ComponentLibrary contains patterns for a UI component library.
type ComponentLibrary struct {
	Components map[string]LibraryComponent `yaml:",inline"`
}

// LibraryComponent is a component from a UI library.
type LibraryComponent struct {
	Import string          `yaml:"import" json:"import"`
	Fixes  []ComponentFix  `yaml:"fixes" json:"fixes"`
}

// PlexusOneDir returns the path to ~/.plexusone/
func PlexusOneDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".plexusone"), nil
}

// A11yConfigDir returns the path to ~/.plexusone/a11y/
func A11yConfigDir() (string, error) {
	plexusDir, err := PlexusOneDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(plexusDir, "a11y"), nil
}

// FixesConfigPath returns the path to ~/.plexusone/a11y/fixes.yaml
func FixesConfigPath() (string, error) {
	a11yDir, err := A11yConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(a11yDir, "fixes.yaml"), nil
}

// LoadFixesConfig loads the fixes configuration from ~/.plexusone/a11y/fixes.yaml
func LoadFixesConfig() (*FixesConfig, error) {
	configPath, err := FixesConfigPath()
	if err != nil {
		return nil, err
	}

	// Return empty config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &FixesConfig{
			Version: "1.0",
			Defaults: FixDefaults{
				Extends: "builtin",
			},
			Languages: make(map[string]LanguageConfig),
			Projects:  nil,
			Libraries: make(map[string]ComponentLibrary),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixes config: %w", err)
	}

	var config FixesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse fixes config: %w", err)
	}

	return &config, nil
}

// SaveFixesConfig saves the fixes configuration to ~/.plexusone/a11y/fixes.yaml
func SaveFixesConfig(config *FixesConfig) error {
	configPath, err := FixesConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// MatchProject finds the best matching project config for a repo path.
func (c *FixesConfig) MatchProject(repoPath string) *ProjectConfig {
	// Normalize path
	repoPath = filepath.Clean(repoPath)

	// Try each project config in order
	for i := range c.Projects {
		project := &c.Projects[i]
		if matchPath(project.Match, repoPath) {
			return project
		}
	}

	return nil
}

// matchPath matches a repo path against a glob pattern.
// Supports wildcards: * matches any single path segment, ** matches any number of segments.
// Glob patterns in individual segments (like dashboard-*) are also supported.
func matchPath(pattern, path string) bool {
	// Handle ** for any path
	if pattern == "**" {
		return true
	}

	// Normalize separators
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	path = filepath.ToSlash(path)

	// Split pattern and path into parts, filtering empty strings
	patternParts := splitNonEmpty(pattern, "/")
	pathParts := splitNonEmpty(path, "/")

	// Try to find the pattern as a subsequence in the path
	return matchPartsSliding(patternParts, pathParts)
}

// splitNonEmpty splits a string and filters out empty parts.
func splitNonEmpty(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// matchPartsSliding finds the pattern parts as a subsequence in path parts.
// The pattern must match contiguously, but can start at any position in the path.
func matchPartsSliding(pattern, path []string) bool {
	if len(pattern) == 0 {
		return true
	}
	if len(path) == 0 {
		return false
	}

	// Try starting the match at each position in the path
	for startIdx := 0; startIdx <= len(path)-len(pattern); startIdx++ {
		if matchFromPosition(pattern, path, startIdx) {
			return true
		}
	}

	// Also try matching with ** semantics where pattern can span non-contiguous parts
	return matchWithDoubleWildcard(pattern, path, 0, 0)
}

// matchFromPosition checks if pattern matches path starting at startIdx.
func matchFromPosition(pattern, path []string, startIdx int) bool {
	pi := 0
	for i := startIdx; i < len(path) && pi < len(pattern); i++ {
		part := pattern[pi]

		if part == "*" {
			// Single wildcard matches any single part
			pi++
			continue
		}

		if part == "**" {
			// Double wildcard - try matching rest from each remaining position
			for j := i; j <= len(path); j++ {
				if pi+1 >= len(pattern) {
					return true
				}
				if matchFromPosition(pattern[pi+1:], path, j) {
					return true
				}
			}
			return false
		}

		// Check for glob patterns in the part
		matched, err := filepath.Match(part, path[i])
		if err != nil || !matched {
			return false
		}
		pi++
	}

	return pi == len(pattern)
}

// matchWithDoubleWildcard handles ** anywhere in the pattern.
func matchWithDoubleWildcard(pattern, path []string, pi, pathIdx int) bool {
	for pi < len(pattern) {
		if pathIdx > len(path) {
			return false
		}

		part := pattern[pi]

		if part == "**" {
			// Try matching rest from each remaining position
			for j := pathIdx; j <= len(path); j++ {
				if matchWithDoubleWildcard(pattern, path, pi+1, j) {
					return true
				}
			}
			return false
		}

		if pathIdx >= len(path) {
			return false
		}

		if part == "*" {
			pi++
			pathIdx++
			continue
		}

		matched, err := filepath.Match(part, path[pathIdx])
		if err != nil || !matched {
			return false
		}
		pi++
		pathIdx++
	}

	return pi == len(pattern)
}

// GetLanguageConfig returns the language config for a given language.
func (c *FixesConfig) GetLanguageConfig(language string) *LanguageConfig {
	if config, ok := c.Languages[language]; ok {
		return &config
	}
	return nil
}

// GetLibrary returns a component library config by name.
func (c *FixesConfig) GetLibrary(name string) *ComponentLibrary {
	if lib, ok := c.Libraries[name]; ok {
		return &lib
	}
	return nil
}
