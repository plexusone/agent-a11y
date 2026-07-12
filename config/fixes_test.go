package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "plexusone/dashboard",
			path:    "/Users/john/plexusone/dashboard",
			want:    true,
		},
		{
			name:    "wildcard prefix",
			pattern: "*/plexusone/dashboard",
			path:    "/Users/john/plexusone/dashboard",
			want:    true,
		},
		{
			name:    "wildcard suffix",
			pattern: "*/plexusone/dashboard-*",
			path:    "/Users/john/plexusone/dashboard-web",
			want:    true,
		},
		{
			name:    "double wildcard",
			pattern: "**",
			path:    "/any/path/here",
			want:    true,
		},
		{
			name:    "no match",
			pattern: "*/plexusone/dashboard-*",
			path:    "/Users/john/other/project",
			want:    false,
		},
		{
			name:    "partial path match",
			pattern: "plexusone/*",
			path:    "/home/user/plexusone/dashboard",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPath(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("matchPath(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestFixesConfigMatchProject(t *testing.T) {
	config := &FixesConfig{
		Projects: []ProjectConfig{
			{
				Match:    "*/plexusone/dashboard-*",
				Language: "react",
			},
			{
				Match:    "*/plexusone/*",
				Language: "vue",
			},
			{
				Match:      "**",
				AutoDetect: true,
			},
		},
	}

	tests := []struct {
		name         string
		repoPath     string
		wantLanguage string
	}{
		{
			name:         "matches first pattern",
			repoPath:     "/Users/john/plexusone/dashboard-web",
			wantLanguage: "react",
		},
		{
			name:         "matches second pattern",
			repoPath:     "/Users/john/plexusone/marketing",
			wantLanguage: "vue",
		},
		{
			name:         "matches fallback",
			repoPath:     "/Users/john/other/project",
			wantLanguage: "", // fallback has no language, uses auto-detect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := config.MatchProject(tt.repoPath)
			if project == nil {
				t.Fatal("expected project match, got nil")
			}
			if project.Language != tt.wantLanguage {
				t.Errorf("got language %q, want %q", project.Language, tt.wantLanguage)
			}
		})
	}
}

func TestLoadFixesConfigMissing(t *testing.T) {
	// Temporarily set HOME to a temp dir
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	config, err := LoadFixesConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	if config.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", config.Version)
	}

	if config.Defaults.Extends != "builtin" {
		t.Errorf("expected extends builtin, got %s", config.Defaults.Extends)
	}
}

func TestSaveAndLoadFixesConfig(t *testing.T) {
	// Create temp dir for config
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create test config
	config := &FixesConfig{
		Version: "1.0",
		Defaults: FixDefaults{
			Extends: "builtin",
			Tokens: map[string][]string{
				"colors": {"primary", "secondary"},
			},
		},
		Languages: map[string]LanguageConfig{
			"react": {
				Patterns: map[string]LanguagePattern{
					"image-alt": {
						Component: "Image",
						Import:    "next/image",
						Example:   "<Image alt=\"...\" />",
					},
				},
			},
		},
		Projects: []ProjectConfig{
			{
				Match:    "*/test/*",
				Language: "react",
			},
		},
	}

	// Save config
	if err := SaveFixesConfig(config); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".plexusone", "a11y", "fixes.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file not created at %s", configPath)
	}

	// Load config
	loaded, err := LoadFixesConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify loaded config
	if loaded.Version != config.Version {
		t.Errorf("version mismatch: got %s, want %s", loaded.Version, config.Version)
	}

	if len(loaded.Languages) != 1 {
		t.Errorf("expected 1 language, got %d", len(loaded.Languages))
	}

	if len(loaded.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(loaded.Projects))
	}
}

func TestGetLanguageConfig(t *testing.T) {
	config := &FixesConfig{
		Languages: map[string]LanguageConfig{
			"react": {
				Conventions: LanguageConventions{
					AriaAttributes: "camelCase",
				},
			},
			"vue": {
				Conventions: LanguageConventions{
					AriaAttributes: "kebab-case",
				},
			},
		},
	}

	// Test existing language
	reactConfig := config.GetLanguageConfig("react")
	if reactConfig == nil {
		t.Fatal("expected react config, got nil")
	}
	if reactConfig.Conventions.AriaAttributes != "camelCase" {
		t.Errorf("expected camelCase, got %s", reactConfig.Conventions.AriaAttributes)
	}

	// Test non-existing language
	goConfig := config.GetLanguageConfig("go")
	if goConfig != nil {
		t.Error("expected nil for non-existing language")
	}
}
