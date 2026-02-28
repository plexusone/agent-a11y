package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Check WCAG defaults
	if cfg.WCAG.Level != "AA" {
		t.Errorf("WCAG.Level: got %q, want %q", cfg.WCAG.Level, "AA")
	}
	if cfg.WCAG.Version != "2.2" {
		t.Errorf("WCAG.Version: got %q, want %q", cfg.WCAG.Version, "2.2")
	}

	// Check Browser defaults
	if !cfg.Browser.Headless {
		t.Error("Browser.Headless should be true by default")
	}
	if cfg.Browser.Timeout.Duration() != 30*time.Second {
		t.Errorf("Browser.Timeout: got %v, want %v", cfg.Browser.Timeout.Duration(), 30*time.Second)
	}

	// Check Output defaults
	if len(cfg.Output.Formats) != 2 {
		t.Errorf("Output.Formats: got %d formats, want 2", len(cfg.Output.Formats))
	}
	if cfg.Output.Directory != "./reports" {
		t.Errorf("Output.Directory: got %q, want %q", cfg.Output.Directory, "./reports")
	}
	if !cfg.Output.Screenshots {
		t.Error("Output.Screenshots should be true by default")
	}

	// Check Crawl defaults
	if cfg.Crawl == nil {
		t.Fatal("Crawl config should not be nil")
	}
	if cfg.Crawl.Depth != 2 {
		t.Errorf("Crawl.Depth: got %d, want 2", cfg.Crawl.Depth)
	}
	if cfg.Crawl.MaxPages != 50 {
		t.Errorf("Crawl.MaxPages: got %d, want 50", cfg.Crawl.MaxPages)
	}
	if !cfg.Crawl.WaitForSPA {
		t.Error("Crawl.WaitForSPA should be true by default")
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
url: https://example.com
wcag:
  level: A
  version: "2.1"
browser:
  headless: false
  timeout: 60s
output:
  formats:
    - json
    - html
    - markdown
  directory: ./output
  screenshots: false
crawl:
  depth: 3
  maxPages: 100
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded values
	if cfg.URL != "https://example.com" {
		t.Errorf("URL: got %q, want %q", cfg.URL, "https://example.com")
	}
	if cfg.WCAG.Level != "A" {
		t.Errorf("WCAG.Level: got %q, want %q", cfg.WCAG.Level, "A")
	}
	if cfg.WCAG.Version != "2.1" {
		t.Errorf("WCAG.Version: got %q, want %q", cfg.WCAG.Version, "2.1")
	}
	if cfg.Browser.Headless {
		t.Error("Browser.Headless should be false")
	}
	if cfg.Browser.Timeout.Duration() != 60*time.Second {
		t.Errorf("Browser.Timeout: got %v, want %v", cfg.Browser.Timeout.Duration(), 60*time.Second)
	}
	if len(cfg.Output.Formats) != 3 {
		t.Errorf("Output.Formats: got %d, want 3", len(cfg.Output.Formats))
	}
	if cfg.Crawl.Depth != 3 {
		t.Errorf("Crawl.Depth: got %d, want 3", cfg.Crawl.Depth)
	}
	if cfg.Crawl.MaxPages != 100 {
		t.Errorf("Crawl.MaxPages: got %d, want 100", cfg.Crawl.MaxPages)
	}
}

func TestLoadConfigWithAuth(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
url: https://example.com
auth:
  type: form
  loginUrl: https://example.com/login
  credentials:
    username: testuser
    password: testpass
  selectors:
    username: "#username"
    password: "#password"
    submit: "#submit"
  successIndicator: ".dashboard"
wcag:
  level: AA
  version: "2.2"
browser:
  headless: true
  timeout: 30s
output:
  formats: [json]
  directory: ./reports
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Auth == nil {
		t.Fatal("Auth config should not be nil")
	}
	if cfg.Auth.Type != "form" {
		t.Errorf("Auth.Type: got %q, want %q", cfg.Auth.Type, "form")
	}
	if cfg.Auth.LoginURL != "https://example.com/login" {
		t.Errorf("Auth.LoginURL: got %q", cfg.Auth.LoginURL)
	}
	if cfg.Auth.Credentials["username"] != "testuser" {
		t.Errorf("Auth.Credentials[username]: got %q", cfg.Auth.Credentials["username"])
	}
	if cfg.Auth.Selectors == nil {
		t.Fatal("Auth.Selectors should not be nil")
	}
	if cfg.Auth.Selectors.Username != "#username" {
		t.Errorf("Auth.Selectors.Username: got %q", cfg.Auth.Selectors.Username)
	}
	if cfg.Auth.SuccessIndicator != ".dashboard" {
		t.Errorf("Auth.SuccessIndicator: got %q", cfg.Auth.SuccessIndicator)
	}
}

func TestLoadConfigWithJourney(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
url: https://example.com
journey:
  path: ./journey.yaml
wcag:
  level: AA
  version: "2.2"
browser:
  headless: true
  timeout: 30s
output:
  formats: [json]
  directory: ./reports
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Journey == nil {
		t.Fatal("Journey config should not be nil")
	}
	if cfg.Journey.Path != "./journey.yaml" {
		t.Errorf("Journey.Path: got %q, want %q", cfg.Journey.Path, "./journey.yaml")
	}
}

func TestLoadConfigWithLLM(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
url: https://example.com
llm:
  enabled: true
  provider: anthropic
  model: claude-sonnet-4-20250514
  judgeCategories:
    - alternative-text
    - color-contrast
wcag:
  level: AA
  version: "2.2"
browser:
  headless: true
  timeout: 30s
output:
  formats: [json]
  directory: ./reports
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.LLM == nil {
		t.Fatal("LLM config should not be nil")
	}
	if !cfg.LLM.Enabled {
		t.Error("LLM.Enabled should be true")
	}
	if cfg.LLM.Provider != "anthropic" {
		t.Errorf("LLM.Provider: got %q, want %q", cfg.LLM.Provider, "anthropic")
	}
	if cfg.LLM.Model != "claude-sonnet-4-20250514" {
		t.Errorf("LLM.Model: got %q", cfg.LLM.Model)
	}
	if len(cfg.LLM.JudgeCategories) != 2 {
		t.Errorf("LLM.JudgeCategories: got %d, want 2", len(cfg.LLM.JudgeCategories))
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidContent := `
url: https://example.com
wcag:
  level: [invalid yaml
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestDurationUnmarshal(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
url: https://example.com
wcag:
  level: AA
  version: "2.2"
browser:
  headless: true
  timeout: 2m30s
output:
  formats: [json]
  directory: ./reports
crawl:
  depth: 1
  maxPages: 10
  delay: 500ms
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	expectedTimeout := 2*time.Minute + 30*time.Second
	if cfg.Browser.Timeout.Duration() != expectedTimeout {
		t.Errorf("Browser.Timeout: got %v, want %v", cfg.Browser.Timeout.Duration(), expectedTimeout)
	}

	expectedDelay := 500 * time.Millisecond
	if cfg.Crawl.Delay.Duration() != expectedDelay {
		t.Errorf("Crawl.Delay: got %v, want %v", cfg.Crawl.Delay.Duration(), expectedDelay)
	}
}

func TestDurationMethod(t *testing.T) {
	d := Duration(5 * time.Second)
	if d.Duration() != 5*time.Second {
		t.Errorf("Duration(): got %v, want %v", d.Duration(), 5*time.Second)
	}
}

func TestConfigWithViewport(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
url: https://example.com
wcag:
  level: AA
  version: "2.2"
browser:
  headless: true
  timeout: 30s
  viewport:
    width: 1920
    height: 1080
output:
  formats: [json]
  directory: ./reports
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Browser.Viewport == nil {
		t.Fatal("Browser.Viewport should not be nil")
	}
	if cfg.Browser.Viewport.Width != 1920 {
		t.Errorf("Viewport.Width: got %d, want 1920", cfg.Browser.Viewport.Width)
	}
	if cfg.Browser.Viewport.Height != 1080 {
		t.Errorf("Viewport.Height: got %d, want 1080", cfg.Browser.Viewport.Height)
	}
}

func TestCrawlConfigPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
url: https://example.com
wcag:
  level: AA
  version: "2.2"
browser:
  headless: true
  timeout: 30s
output:
  formats: [json]
  directory: ./reports
crawl:
  depth: 2
  maxPages: 50
  includePatterns:
    - "/products/**"
    - "/about/*"
  excludePatterns:
    - "/admin/**"
    - "*.pdf"
  respectRobots: true
  useSitemap: true
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Crawl.IncludePatterns) != 2 {
		t.Errorf("Crawl.IncludePatterns: got %d, want 2", len(cfg.Crawl.IncludePatterns))
	}
	if len(cfg.Crawl.ExcludePatterns) != 2 {
		t.Errorf("Crawl.ExcludePatterns: got %d, want 2", len(cfg.Crawl.ExcludePatterns))
	}
	if !cfg.Crawl.RespectRobots {
		t.Error("Crawl.RespectRobots should be true")
	}
	if !cfg.Crawl.UseSitemap {
		t.Error("Crawl.UseSitemap should be true")
	}
}
