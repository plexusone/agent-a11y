package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ProjectContext contains resolved project information.
type ProjectContext struct {
	// RepoPath is the absolute path to the project root
	RepoPath string `json:"repoPath"`

	// Language is the detected or configured language
	Language string `json:"language"` // react, vue, svelte, go-templates, etc.

	// Framework is the detected or configured framework
	Framework string `json:"framework"` // next, nuxt, sveltekit, etc.

	// ComponentLibrary is the detected UI library
	ComponentLibrary string `json:"componentLibrary"` // mui, chakra, shadcn, etc.

	// DesignSystem is the path to the design system spec
	DesignSystem string `json:"designSystem,omitempty"`

	// ProjectConfig is the matched project configuration
	ProjectConfig *ProjectConfig `json:"projectConfig,omitempty"`

	// LanguageConfig is the language-specific configuration
	LanguageConfig *LanguageConfig `json:"languageConfig,omitempty"`

	// LibraryConfig is the component library configuration
	LibraryConfig *ComponentLibrary `json:"libraryConfig,omitempty"`
}

// DetectProject detects project context from a directory path.
func DetectProject(dir string) (*ProjectContext, error) {
	// Get absolute path
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// Find project root (look for package.json, go.mod, etc.)
	repoPath := findProjectRoot(absPath)
	if repoPath == "" {
		repoPath = absPath
	}

	ctx := &ProjectContext{
		RepoPath: repoPath,
	}

	// Detect language and framework
	ctx.Language, ctx.Framework = detectLanguageAndFramework(repoPath)

	// Detect component library
	ctx.ComponentLibrary = detectComponentLibrary(repoPath)

	return ctx, nil
}

// ResolveProject resolves full project context including config.
func ResolveProject(dir string, fixesConfig *FixesConfig) (*ProjectContext, error) {
	// Start with detection
	ctx, err := DetectProject(dir)
	if err != nil {
		return nil, err
	}

	if fixesConfig == nil {
		return ctx, nil
	}

	// Match project config
	ctx.ProjectConfig = fixesConfig.MatchProject(ctx.RepoPath)

	// Override with project config if matched
	if ctx.ProjectConfig != nil {
		if ctx.ProjectConfig.Language != "" {
			ctx.Language = ctx.ProjectConfig.Language
		}
		if ctx.ProjectConfig.Framework != "" {
			ctx.Framework = ctx.ProjectConfig.Framework
		}
		if ctx.ProjectConfig.DesignSystem != "" {
			ctx.DesignSystem = expandPath(ctx.ProjectConfig.DesignSystem)
		}
	}

	// Get language config
	ctx.LanguageConfig = fixesConfig.GetLanguageConfig(ctx.Language)

	// Get library config
	ctx.LibraryConfig = fixesConfig.GetLibrary(ctx.ComponentLibrary)

	return ctx, nil
}

// findProjectRoot walks up the directory tree to find a project root.
func findProjectRoot(dir string) string {
	markers := []string{
		"package.json",
		"go.mod",
		"Cargo.toml",
		"pyproject.toml",
		"composer.json",
		".git",
	}

	current := dir
	for {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
				return current
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return ""
}

// detectLanguageAndFramework detects the language and framework from project files.
func detectLanguageAndFramework(repoPath string) (language, framework string) {
	// Check for package.json (JS/TS projects)
	packageJSONPath := filepath.Join(repoPath, "package.json")
	if data, err := os.ReadFile(packageJSONPath); err == nil {
		lang, fw := detectFromPackageJSON(data)
		if lang != "" {
			return lang, fw
		}
	}

	// Check for go.mod (Go projects)
	if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
		// Check for templ files
		if hasFiles(repoPath, "*.templ") {
			return "go-templ", ""
		}
		// Check for html/template usage
		if hasFiles(repoPath, "*.html") || hasFiles(repoPath, "**/templates/*.html") {
			return "go-templates", ""
		}
		return "go", ""
	}

	// Check for Cargo.toml (Rust)
	if _, err := os.Stat(filepath.Join(repoPath, "Cargo.toml")); err == nil {
		return "rust", ""
	}

	// Check for pyproject.toml or requirements.txt (Python)
	if _, err := os.Stat(filepath.Join(repoPath, "pyproject.toml")); err == nil {
		return "python", ""
	}
	if _, err := os.Stat(filepath.Join(repoPath, "requirements.txt")); err == nil {
		return "python", ""
	}

	return "unknown", ""
}

// detectFromPackageJSON detects language and framework from package.json.
func detectFromPackageJSON(data []byte) (language, framework string) {
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", ""
	}

	// Merge dependencies
	deps := make(map[string]bool)
	for k := range pkg.Dependencies {
		deps[k] = true
	}
	for k := range pkg.DevDependencies {
		deps[k] = true
	}

	// Detect framework first (more specific)
	if deps["next"] {
		return "react", "next"
	}
	if deps["gatsby"] {
		return "react", "gatsby"
	}
	if deps["remix"] {
		return "react", "remix"
	}
	if deps["nuxt"] {
		return "vue", "nuxt"
	}
	if deps["@sveltejs/kit"] {
		return "svelte", "sveltekit"
	}
	if deps["astro"] {
		return "astro", "astro"
	}

	// Detect base framework
	if deps["react"] || deps["react-dom"] {
		return "react", ""
	}
	if deps["vue"] {
		return "vue", ""
	}
	if deps["svelte"] {
		return "svelte", ""
	}
	if deps["@angular/core"] {
		return "angular", ""
	}
	if deps["solid-js"] {
		return "solid", ""
	}
	if deps["preact"] {
		return "preact", ""
	}

	// Default to generic JS/TS
	if deps["typescript"] {
		return "typescript", ""
	}

	return "javascript", ""
}

// detectComponentLibrary detects the UI component library from package.json.
func detectComponentLibrary(repoPath string) string {
	packageJSONPath := filepath.Join(repoPath, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return ""
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}

	// Merge dependencies
	deps := make(map[string]bool)
	for k := range pkg.Dependencies {
		deps[k] = true
	}
	for k := range pkg.DevDependencies {
		deps[k] = true
	}

	// Check for common component libraries
	// Order matters: more specific (multi-dep) libraries first
	// matchAll: true means ALL deps must be present, false means ANY dep matches
	libraries := []struct {
		name     string
		deps     []string
		matchAll bool
	}{
		// shadcn requires BOTH radix AND cva - check first before radix
		{"shadcn", []string{"class-variance-authority"}, true}, // cva is the distinguishing factor
		// Single-dep or any-match libraries
		{"mui", []string{"@mui/material", "@mui/core"}, false},
		{"chakra", []string{"@chakra-ui/react"}, false},
		{"mantine", []string{"@mantine/core"}, false},
		{"antd", []string{"antd"}, false},
		{"radix", []string{"@radix-ui/react-dialog", "@radix-ui/themes"}, false},
		{"headless-ui", []string{"@headlessui/react"}, false},
		{"bootstrap", []string{"react-bootstrap", "bootstrap"}, false},
		{"vuetify", []string{"vuetify"}, false},
		{"quasar", []string{"quasar"}, false},
		{"element-plus", []string{"element-plus"}, false},
		{"naive-ui", []string{"naive-ui"}, false},
	}

	for _, lib := range libraries {
		if lib.matchAll {
			// All deps must be present
			allMatch := true
			for _, dep := range lib.deps {
				if !deps[dep] {
					allMatch = false
					break
				}
			}
			if allMatch {
				return lib.name
			}
		} else {
			// Any dep matches
			for _, dep := range lib.deps {
				if deps[dep] {
					return lib.name
				}
			}
		}
	}

	return ""
}

// hasFiles checks if any files matching the pattern exist in the directory.
func hasFiles(dir, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return false
	}
	return len(matches) > 0
}

// expandPath expands ~ in paths.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
