package sourcemap

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/plexusone/agent-a11y/types"
)

func TestDecodeMappings(t *testing.T) {
	// Test basic VLQ decoding
	// "AAAA" = [0,0,0,0] - first segment at column 0, source 0, line 0, column 0
	segments := decodeMappings("AAAA")
	if len(segments) != 1 || len(segments[0]) != 1 {
		t.Errorf("expected 1 line with 1 segment, got %d lines", len(segments))
	}
	if len(segments) > 0 && len(segments[0]) > 0 {
		seg := segments[0][0]
		if seg.GeneratedColumn != 0 || seg.SourceLine != 0 || seg.SourceColumn != 0 {
			t.Errorf("unexpected segment values: %+v", seg)
		}
	}
}

func TestDecodeVLQ(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"A", []int{0}},
		{"C", []int{1}},
		{"D", []int{-1}},
		{"AAAA", []int{0, 0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := decodeVLQ(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d values, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d: expected %d, got %d", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestExtractComponentFromHTML(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		framework types.Framework
		expected  string
	}{
		{
			name:      "React data-testid",
			html:      `<button data-testid="submit-button">Submit</button>`,
			framework: types.FrameworkReact,
			expected:  "SubmitButton",
		},
		{
			name:      "React PascalCase class",
			html:      `<div class="HeroSection">Content</div>`,
			framework: types.FrameworkReact,
			expected:  "HeroSection",
		},
		{
			name:      "Generic PascalCase class",
			html:      `<div class="UserProfile">Content</div>`,
			framework: types.FrameworkVanilla,
			expected:  "UserProfile",
		},
		{
			name:      "No component found",
			html:      `<div class="container">Content</div>`,
			framework: types.FrameworkVanilla,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractComponentFromHTML(tt.html, tt.framework)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractIdentifiers(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		html     string
		minCount int
	}{
		{
			name:     "Extract ID",
			selector: "#main-content",
			html:     `<div id="main-content"></div>`,
			minCount: 1,
		},
		{
			name:     "Extract classes",
			selector: ".hero.active",
			html:     `<div class="hero active"></div>`,
			minCount: 2,
		},
		{
			name:     "Extract data attributes",
			selector: ".btn",
			html:     `<button class="btn" data-action="submit" data-form="login">Submit</button>`,
			minCount: 2,
		},
		{
			name:     "Extract aria attributes",
			selector: ".alert",
			html:     `<div class="alert" aria-live="polite" aria-atomic="true">Alert</div>`,
			minCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIdentifiers(tt.selector, tt.html)
			if len(result) < tt.minCount {
				t.Errorf("expected at least %d patterns, got %d: %v", tt.minCount, len(result), result)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"submit-button", "SubmitButton"},
		{"user_profile", "UserProfile"},
		{"hello world", "HelloWorld"},
		{"already-Pascal-case", "AlreadyPascalCase"},
		{"single", "Single"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNormalizeSourcePath(t *testing.T) {
	tests := []struct {
		source     string
		sourceRoot string
		expected   string
	}{
		{"./src/App.tsx", "", "src/App.tsx"},
		{"webpack://./src/App.tsx", "", "src/App.tsx"},
		{"src/components/Button.tsx", "src", "src/components/Button.tsx"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := normalizeSourcePath(tt.source, tt.sourceRoot)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMapperWithSourceMaps(t *testing.T) {
	// Create a temporary directory with a test source map
	tmpDir := t.TempDir()

	// Create a minimal source map
	sourceMap := `{
		"version": 3,
		"file": "bundle.js",
		"sourceRoot": "",
		"sources": ["src/components/Hero.tsx"],
		"sourcesContent": ["import React from 'react';\n\nexport function Hero() {\n  return <div id=\"hero\">Hero Content</div>;\n}\n"],
		"names": [],
		"mappings": "AAAA"
	}`

	mapFile := filepath.Join(tmpDir, "bundle.js.map")
	if err := os.WriteFile(mapFile, []byte(sourceMap), 0644); err != nil {
		t.Fatalf("failed to write test source map: %v", err)
	}

	// Create mapper
	mapper, err := NewMapper(tmpDir)
	if err != nil {
		t.Fatalf("failed to create mapper: %v", err)
	}

	// Test source files listing
	files := mapper.SourceFiles()
	if len(files) == 0 {
		t.Error("expected at least one source file")
	}

	// Test mapping with matching pattern
	loc, err := mapper.MapElement(context.Background(), "#hero", `<div id="hero">Hero Content</div>`)
	if err != nil {
		t.Fatalf("MapElement failed: %v", err)
	}

	if loc != nil {
		t.Logf("Found source location: %+v", loc)
		if loc.File == "" {
			t.Error("expected non-empty file path")
		}
	}
}

func TestMapperWithoutSourceMaps(t *testing.T) {
	// Create mapper without source maps
	mapper, err := NewMapper("")
	if err != nil {
		t.Fatalf("failed to create mapper: %v", err)
	}

	// Should return nil location when no source maps loaded
	loc, err := mapper.MapElement(context.Background(), ".test", "<div class='test'></div>")
	if err != nil {
		t.Fatalf("MapElement failed: %v", err)
	}
	if loc != nil {
		t.Error("expected nil location when no source maps loaded")
	}
}
