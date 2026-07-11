// Package sourcemap provides source map parsing and DOM-to-source mapping.
// This enables coding agents to locate source files for accessibility issues
// found in the rendered DOM.
package sourcemap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/plexusone/agent-a11y/types"
)

// Mapper maps DOM elements to source code locations using source maps.
type Mapper struct {
	// Root directory containing source maps
	sourceMapDir string

	// Loaded source maps indexed by generated file
	sourceMaps map[string]*SourceMap

	// Detected framework
	framework types.Framework

	// Base URL for source map resolution (for dev server)
	baseURL string
}

// SourceMap represents a parsed source map (V3 format).
type SourceMap struct {
	Version        int      `json:"version"`
	File           string   `json:"file"`
	SourceRoot     string   `json:"sourceRoot"`
	Sources        []string `json:"sources"`
	SourcesContent []string `json:"sourcesContent,omitempty"`
	Names          []string `json:"names"`
	Mappings       string   `json:"mappings"`

	// Decoded mappings (populated after parsing)
	segments [][]Segment
}

// Segment represents a single mapping segment.
type Segment struct {
	GeneratedColumn int
	SourceIndex     int
	SourceLine      int
	SourceColumn    int
	NameIndex       int
	HasName         bool
}

// MapperOption configures the Mapper.
type MapperOption func(*Mapper)

// WithBaseURL sets the base URL for source map resolution from dev server.
func WithBaseURL(url string) MapperOption {
	return func(m *Mapper) {
		m.baseURL = url
	}
}

// NewMapper creates a new source map mapper.
func NewMapper(sourceMapDir string, opts ...MapperOption) (*Mapper, error) {
	m := &Mapper{
		sourceMapDir: sourceMapDir,
		sourceMaps:   make(map[string]*SourceMap),
		framework:    types.FrameworkUnknown,
	}

	for _, opt := range opts {
		opt(m)
	}

	if sourceMapDir != "" {
		if err := m.loadSourceMaps(); err != nil {
			return nil, fmt.Errorf("failed to load source maps: %w", err)
		}
	}

	return m, nil
}

// loadSourceMaps loads all source maps from the configured directory.
func (m *Mapper) loadSourceMaps() error {
	if m.sourceMapDir == "" {
		return nil
	}

	return filepath.Walk(m.sourceMapDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".map") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var sm SourceMap
		if err := json.Unmarshal(data, &sm); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Decode VLQ mappings
		sm.segments = decodeMappings(sm.Mappings)

		// Index by generated file name
		genFile := strings.TrimSuffix(filepath.Base(path), ".map")
		m.sourceMaps[genFile] = &sm

		return nil
	})
}

// MapElement maps a DOM element to its source location.
// It uses the element's CSS selector to find the original source.
func (m *Mapper) MapElement(_ context.Context, selector string, html string) (*types.SourceLocation, error) {
	if len(m.sourceMaps) == 0 {
		return nil, nil
	}

	// Extract component info from class names or data attributes
	component := extractComponentFromHTML(html, m.framework)

	// Try to find a matching source location
	loc := m.findSourceLocation(selector, html)
	if loc != nil {
		loc.Component = component
		loc.Framework = m.framework
		return loc, nil
	}

	// If we have a component name, try to find its definition
	if component != "" {
		loc = m.findComponentDefinition(component)
		if loc != nil {
			loc.Framework = m.framework
			return loc, nil
		}
	}

	return nil, nil
}

// findSourceLocation attempts to locate source based on element info.
func (m *Mapper) findSourceLocation(selector string, html string) *types.SourceLocation {
	// Extract identifiable patterns
	patterns := extractIdentifiers(selector, html)

	for _, sm := range m.sourceMaps {
		for i, source := range sm.Sources {
			// Check if source content contains any of our patterns
			if i < len(sm.SourcesContent) {
				content := sm.SourcesContent[i]
				for _, pattern := range patterns {
					if line := findLineContaining(content, pattern); line > 0 {
						return &types.SourceLocation{
							File:          normalizeSourcePath(source, sm.SourceRoot),
							Line:          line,
							SourceMapFile: sm.File,
							Confidence:    0.7,
						}
					}
				}
			}
		}
	}

	return nil
}

// findComponentDefinition searches for a component definition.
func (m *Mapper) findComponentDefinition(component string) *types.SourceLocation {
	// Patterns for component definitions in various frameworks
	patterns := []string{
		fmt.Sprintf("function %s", component),
		fmt.Sprintf("const %s", component),
		fmt.Sprintf("class %s", component),
		fmt.Sprintf("export default %s", component),
		fmt.Sprintf("export function %s", component),
		fmt.Sprintf("name: '%s'", component),
		fmt.Sprintf(`name: "%s"`, component),
	}

	for _, sm := range m.sourceMaps {
		for i, source := range sm.Sources {
			if i < len(sm.SourcesContent) {
				content := sm.SourcesContent[i]
				for _, pattern := range patterns {
					if line := findLineContaining(content, pattern); line > 0 {
						return &types.SourceLocation{
							File:          normalizeSourcePath(source, sm.SourceRoot),
							Line:          line,
							Component:     component,
							SourceMapFile: sm.File,
							Confidence:    0.85,
						}
					}
				}
			}
		}
	}

	return nil
}

// SetFramework sets the detected framework.
func (m *Mapper) SetFramework(fw types.Framework) {
	m.framework = fw
}

// Framework returns the detected framework.
func (m *Mapper) Framework() types.Framework {
	return m.framework
}

// extractComponentFromHTML extracts component name from HTML.
func extractComponentFromHTML(html string, framework types.Framework) string {
	switch framework {
	case types.FrameworkReact:
		// React: data-reactroot, data-testid, or PascalCase class names
		re := regexp.MustCompile(`data-testid="([^"]+)"`)
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			return toPascalCase(m[1])
		}
		// Look for PascalCase class names
		re = regexp.MustCompile(`class="([A-Z][a-zA-Z0-9]+)`)
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			return m[1]
		}

	case types.FrameworkVue:
		// Vue: data-v-* attributes indicate scoped component
		re := regexp.MustCompile(`data-v-([a-f0-9]+)`)
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			// Can't get component name from hash alone
			return ""
		}

	case types.FrameworkSvelte:
		// Svelte: svelte-* class names
		re := regexp.MustCompile(`class="svelte-([a-z0-9]+)`)
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			// Can't get component name from hash alone
			return ""
		}

	case types.FrameworkAngular:
		// Angular: _nghost-*, _ngcontent-* attributes
		re := regexp.MustCompile(`_ng(host|content)-([a-z]+-\d+)`)
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			return ""
		}
	}

	// Generic: look for any PascalCase class or id
	re := regexp.MustCompile(`(?:class|id)="([A-Z][a-zA-Z0-9]+)`)
	if m := re.FindStringSubmatch(html); len(m) > 1 {
		return m[1]
	}

	return ""
}

// extractIdentifiers extracts searchable patterns from element info.
func extractIdentifiers(selector string, html string) []string {
	var patterns []string

	// Extract id
	idRe := regexp.MustCompile(`#([a-zA-Z][a-zA-Z0-9_-]*)`)
	if m := idRe.FindStringSubmatch(selector); len(m) > 1 {
		patterns = append(patterns, fmt.Sprintf(`id="%s"`, m[1]))
		patterns = append(patterns, fmt.Sprintf(`id='%s'`, m[1]))
	}

	// Extract classes
	classRe := regexp.MustCompile(`\.([a-zA-Z][a-zA-Z0-9_-]*)`)
	for _, m := range classRe.FindAllStringSubmatch(selector, -1) {
		if len(m) > 1 {
			patterns = append(patterns, m[1])
		}
	}

	// Extract data attributes
	dataRe := regexp.MustCompile(`data-([a-zA-Z][a-zA-Z0-9-]*)="([^"]*)"`)
	for _, m := range dataRe.FindAllStringSubmatch(html, -1) {
		if len(m) > 2 {
			patterns = append(patterns, fmt.Sprintf(`data-%s="%s"`, m[1], m[2]))
		}
	}

	// Extract aria attributes
	ariaRe := regexp.MustCompile(`aria-([a-zA-Z]+)="([^"]*)"`)
	for _, m := range ariaRe.FindAllStringSubmatch(html, -1) {
		if len(m) > 2 {
			patterns = append(patterns, fmt.Sprintf(`aria-%s="%s"`, m[1], m[2]))
		}
	}

	return patterns
}

// findLineContaining finds the line number containing a pattern.
func findLineContaining(content, pattern string) int {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			return i + 1 // 1-indexed
		}
	}
	return 0
}

// normalizeSourcePath normalizes a source path relative to source root.
func normalizeSourcePath(source, sourceRoot string) string {
	// Remove webpack:// or similar protocol prefixes
	source = regexp.MustCompile(`^[a-z]+://`).ReplaceAllString(source, "")

	// Remove leading ./ or ../
	source = strings.TrimPrefix(source, "./")

	// Combine with source root if present
	if sourceRoot != "" && !strings.HasPrefix(source, sourceRoot) {
		source = filepath.Join(sourceRoot, source)
	}

	return source
}

// toPascalCase converts a string to PascalCase.
func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})

	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}

	return result.String()
}

// decodeMappings decodes VLQ-encoded source map mappings.
func decodeMappings(mappings string) [][]Segment {
	if mappings == "" {
		return nil
	}

	var result [][]Segment
	var currentLine []Segment

	// State for relative decoding
	var (
		generatedColumn int
		sourceIndex     int
		sourceLine      int
		sourceColumn    int
		nameIndex       int
	)

	for _, line := range strings.Split(mappings, ";") {
		generatedColumn = 0 // Reset for each line
		currentLine = nil

		if line == "" {
			result = append(result, currentLine)
			continue
		}

		for _, segment := range strings.Split(line, ",") {
			if segment == "" {
				continue
			}

			values := decodeVLQ(segment)
			if len(values) == 0 {
				continue
			}

			seg := Segment{
				GeneratedColumn: generatedColumn + values[0],
			}
			generatedColumn = seg.GeneratedColumn

			if len(values) >= 4 {
				sourceIndex += values[1]
				sourceLine += values[2]
				sourceColumn += values[3]
				seg.SourceIndex = sourceIndex
				seg.SourceLine = sourceLine
				seg.SourceColumn = sourceColumn
			}

			if len(values) >= 5 {
				nameIndex += values[4]
				seg.NameIndex = nameIndex
				seg.HasName = true
			}

			currentLine = append(currentLine, seg)
		}

		result = append(result, currentLine)
	}

	return result
}

// decodeVLQ decodes a VLQ-encoded segment.
func decodeVLQ(segment string) []int {
	var values []int
	var value, shift int
	var negative bool

	for _, c := range segment {
		digit := vlqCharToInt(c)
		if digit == -1 {
			continue
		}

		value += (digit & 31) << shift
		shift += 5

		if (digit & 32) == 0 {
			// Check sign bit
			negative = (value & 1) == 1
			value >>= 1
			if negative {
				value = -value
			}
			values = append(values, value)
			value = 0
			shift = 0
		}
	}

	return values
}

var vlqChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// vlqCharToInt converts a VLQ character to its integer value.
func vlqCharToInt(c rune) int {
	for i, ch := range vlqChars {
		if ch == c {
			return i
		}
	}
	return -1
}

// SourceFiles returns all source files in loaded source maps.
func (m *Mapper) SourceFiles() []string {
	seen := make(map[string]bool)
	var files []string

	for _, sm := range m.sourceMaps {
		for _, source := range sm.Sources {
			normalized := normalizeSourcePath(source, sm.SourceRoot)
			if !seen[normalized] {
				seen[normalized] = true
				files = append(files, normalized)
			}
		}
	}

	sort.Strings(files)
	return files
}
