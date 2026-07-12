package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFromPackageJSON(t *testing.T) {
	tests := []struct {
		name          string
		packageJSON   string
		wantLanguage  string
		wantFramework string
	}{
		{
			name: "Next.js project",
			packageJSON: `{
				"dependencies": {
					"next": "14.0.0",
					"react": "18.0.0",
					"react-dom": "18.0.0"
				}
			}`,
			wantLanguage:  "react",
			wantFramework: "next",
		},
		{
			name: "Plain React project",
			packageJSON: `{
				"dependencies": {
					"react": "18.0.0",
					"react-dom": "18.0.0"
				}
			}`,
			wantLanguage:  "react",
			wantFramework: "",
		},
		{
			name: "Nuxt project",
			packageJSON: `{
				"dependencies": {
					"nuxt": "3.0.0",
					"vue": "3.0.0"
				}
			}`,
			wantLanguage:  "vue",
			wantFramework: "nuxt",
		},
		{
			name: "SvelteKit project",
			packageJSON: `{
				"devDependencies": {
					"@sveltejs/kit": "2.0.0",
					"svelte": "4.0.0"
				}
			}`,
			wantLanguage:  "svelte",
			wantFramework: "sveltekit",
		},
		{
			name: "Angular project",
			packageJSON: `{
				"dependencies": {
					"@angular/core": "17.0.0"
				}
			}`,
			wantLanguage:  "angular",
			wantFramework: "",
		},
		{
			name: "TypeScript project",
			packageJSON: `{
				"devDependencies": {
					"typescript": "5.0.0"
				}
			}`,
			wantLanguage:  "typescript",
			wantFramework: "",
		},
		{
			name: "Plain JavaScript",
			packageJSON: `{
				"name": "my-project"
			}`,
			wantLanguage:  "javascript",
			wantFramework: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang, fw := detectFromPackageJSON([]byte(tt.packageJSON))
			if lang != tt.wantLanguage {
				t.Errorf("language: got %q, want %q", lang, tt.wantLanguage)
			}
			if fw != tt.wantFramework {
				t.Errorf("framework: got %q, want %q", fw, tt.wantFramework)
			}
		})
	}
}

func TestDetectComponentLibrary(t *testing.T) {
	tests := []struct {
		name        string
		packageJSON string
		want        string
	}{
		{
			name: "MUI project",
			packageJSON: `{
				"dependencies": {
					"@mui/material": "5.0.0"
				}
			}`,
			want: "mui",
		},
		{
			name: "Chakra project",
			packageJSON: `{
				"dependencies": {
					"@chakra-ui/react": "2.0.0"
				}
			}`,
			want: "chakra",
		},
		{
			name: "Ant Design project",
			packageJSON: `{
				"dependencies": {
					"antd": "5.0.0"
				}
			}`,
			want: "antd",
		},
		{
			name: "shadcn project (radix + cva)",
			packageJSON: `{
				"dependencies": {
					"@radix-ui/react-dialog": "1.0.0",
					"class-variance-authority": "0.7.0"
				}
			}`,
			want: "shadcn",
		},
		{
			name: "No component library",
			packageJSON: `{
				"dependencies": {
					"react": "18.0.0"
				}
			}`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir with package.json
			tmpDir := t.TempDir()
			pkgPath := filepath.Join(tmpDir, "package.json")
			if err := os.WriteFile(pkgPath, []byte(tt.packageJSON), 0644); err != nil {
				t.Fatalf("failed to write package.json: %v", err)
			}

			got := detectComponentLibrary(tmpDir)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create project structure
	projectDir := filepath.Join(tmpDir, "projects", "my-app")
	srcDir := filepath.Join(projectDir, "src", "components")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	// Create package.json at project root
	pkgPath := filepath.Join(projectDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"name": "my-app"}`), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Test from subdirectory
	root := findProjectRoot(srcDir)
	if root != projectDir {
		t.Errorf("got %q, want %q", root, projectDir)
	}

	// Test from project root
	root = findProjectRoot(projectDir)
	if root != projectDir {
		t.Errorf("got %q, want %q", root, projectDir)
	}
}

func TestDetectProject(t *testing.T) {
	// Create temp Next.js project
	tmpDir := t.TempDir()

	// Create package.json
	pkgJSON := `{
		"name": "test-app",
		"dependencies": {
			"next": "14.0.0",
			"react": "18.0.0",
			"@chakra-ui/react": "2.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	ctx, err := DetectProject(tmpDir)
	if err != nil {
		t.Fatalf("DetectProject failed: %v", err)
	}

	if ctx.Language != "react" {
		t.Errorf("language: got %q, want %q", ctx.Language, "react")
	}

	if ctx.Framework != "next" {
		t.Errorf("framework: got %q, want %q", ctx.Framework, "next")
	}

	if ctx.ComponentLibrary != "chakra" {
		t.Errorf("componentLibrary: got %q, want %q", ctx.ComponentLibrary, "chakra")
	}
}

func TestResolveProject(t *testing.T) {
	// Create temp project
	tmpDir := t.TempDir()

	// Create package.json
	pkgJSON := `{
		"dependencies": {
			"react": "18.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Create fixes config with override
	fixesConfig := &FixesConfig{
		Projects: []ProjectConfig{
			{
				Match:     "**",
				Language:  "react",
				Framework: "custom-framework",
				Components: map[string]ComponentFixConfig{
					"Button": {
						Selectors: []string{".btn"},
						Fixes: []ComponentFix{
							{
								Rule:    "button-name",
								Pattern: "<Button aria-label=\"...\">",
							},
						},
					},
				},
			},
		},
		Languages: map[string]LanguageConfig{
			"react": {
				Conventions: LanguageConventions{
					AriaAttributes: "camelCase",
				},
			},
		},
	}

	ctx, err := ResolveProject(tmpDir, fixesConfig)
	if err != nil {
		t.Fatalf("ResolveProject failed: %v", err)
	}

	// Check project config was matched
	if ctx.ProjectConfig == nil {
		t.Fatal("expected project config match")
	}

	// Check framework override
	if ctx.Framework != "custom-framework" {
		t.Errorf("framework: got %q, want %q", ctx.Framework, "custom-framework")
	}

	// Check language config was loaded
	if ctx.LanguageConfig == nil {
		t.Fatal("expected language config")
	}
	if ctx.LanguageConfig.Conventions.AriaAttributes != "camelCase" {
		t.Errorf("conventions: got %q, want %q",
			ctx.LanguageConfig.Conventions.AriaAttributes, "camelCase")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/config", filepath.Join(home, "config")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := expandPath(tt.input)
		if got != tt.want {
			t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
