package crawler

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxDepth != 3 {
		t.Errorf("MaxDepth: got %d, want 3", cfg.MaxDepth)
	}
	if cfg.MaxPages != 100 {
		t.Errorf("MaxPages: got %d, want 100", cfg.MaxPages)
	}
	if !cfg.WaitForSPA {
		t.Error("WaitForSPA should be true")
	}
	if !cfg.RespectRobots {
		t.Error("RespectRobots should be true")
	}
	if !cfg.UseSitemap {
		t.Error("UseSitemap should be true")
	}
	if cfg.Delay != 500*time.Millisecond {
		t.Errorf("Delay: got %v, want 500ms", cfg.Delay)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout: got %v, want 30s", cfg.Timeout)
	}
	if cfg.Concurrency != 3 {
		t.Errorf("Concurrency: got %d, want 3", cfg.Concurrency)
	}

	expectedIndicators := []string{
		"[data-reactroot]",
		"#__next",
		"[data-v-]",
		"#app[data-server-rendered]",
		"[ng-app]",
		"[ng-controller]",
	}
	if len(cfg.SPAIndicators) != len(expectedIndicators) {
		t.Errorf("SPAIndicators count: got %d, want %d", len(cfg.SPAIndicators), len(expectedIndicators))
	}
}

func TestNormalizeURL(t *testing.T) {
	c := &Crawler{}

	tests := []struct {
		input    string
		expected string
	}{
		// Remove fragments
		{"https://example.com/page#section", "https://example.com/page"},
		{"https://example.com/#top", "https://example.com"},

		// Remove trailing slash
		{"https://example.com/page/", "https://example.com/page"},

		// Normalize case
		{"HTTPS://EXAMPLE.COM/Page", "https://example.com/Page"},

		// Keep query strings
		{"https://example.com/page?id=1", "https://example.com/page?id=1"},

		// Handle already normalized
		{"https://example.com/page", "https://example.com/page"},

		// Handle root - trailing slash is removed
		{"https://example.com/", "https://example.com"},
	}

	for _, tt := range tests {
		result := c.normalizeURL(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeURL(%q): got %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMatchesPatterns(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		includePatterns []string
		excludePatterns []string
		expected        bool
	}{
		{
			name:            "no patterns - include all",
			url:             "https://example.com/any/page",
			includePatterns: nil,
			excludePatterns: nil,
			expected:        true,
		},
		{
			name:            "matches include pattern with full URL glob",
			url:             "https://example.com/products/item",
			includePatterns: []string{"**/products/**"},
			excludePatterns: nil,
			expected:        true,
		},
		{
			name:            "does not match include pattern",
			url:             "https://example.com/about",
			includePatterns: []string{"**/products/**"},
			excludePatterns: nil,
			expected:        false,
		},
		{
			name:            "excluded by pattern",
			url:             "https://example.com/admin/dashboard",
			includePatterns: nil,
			excludePatterns: []string{"**/admin/**"},
			expected:        false,
		},
		{
			name:            "exclude takes precedence",
			url:             "https://example.com/products/admin/edit",
			includePatterns: []string{"**/products/**"},
			excludePatterns: []string{"**/admin/**"},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Crawler{
				config: Config{
					IncludePatterns: tt.includePatterns,
					ExcludePatterns: tt.excludePatterns,
				},
			}
			result := c.matchesPatterns(tt.url)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMatchGlob(t *testing.T) {
	c := &Crawler{}

	tests := []struct {
		s        string
		pattern  string
		expected bool
	}{
		// Single wildcard
		{"/products/item", "/products/*", true},
		{"/products/item/detail", "/products/*", false},

		// Double wildcard - note: ** requires chars after it in regex
		{"/products/item/detail", "/products/**", true},
		{"/products/", "/products/**", true},
		{"/admin/dashboard", "/products/**", false},

		// Exact match
		{"/about", "/about", true},
		{"/about-us", "/about", false},

		// File extension
		{"file.pdf", "*.pdf", true},
		{"file.doc", "*.pdf", false},

		// Full URL patterns
		{"https://example.com/products/item", "**/products/**", true},
		{"https://example.com/about", "**/about", true},
	}

	for _, tt := range tests {
		result := c.matchGlob(tt.s, tt.pattern)
		if result != tt.expected {
			t.Errorf("matchGlob(%q, %q): got %v, want %v", tt.s, tt.pattern, result, tt.expected)
		}
	}
}

func TestPageStruct(t *testing.T) {
	page := Page{
		URL:            "https://example.com/page",
		Title:          "Test Page",
		Depth:          2,
		DiscoveredFrom: "https://example.com",
		IsSPA:          true,
		SPAFramework:   "react",
		LoadTime:       500 * time.Millisecond,
		Links:          []string{"https://example.com/other"},
		Error:          "",
	}

	if page.URL != "https://example.com/page" {
		t.Errorf("URL mismatch")
	}
	if !page.IsSPA {
		t.Error("IsSPA should be true")
	}
	if page.SPAFramework != "react" {
		t.Errorf("SPAFramework: got %q", page.SPAFramework)
	}
}

func TestResultStruct(t *testing.T) {
	result := Result{
		StartURL:   "https://example.com",
		Pages:      []Page{{URL: "https://example.com"}},
		TotalPages: 1,
		Duration:   5 * time.Second,
		RobotsTxt:  "User-agent: *\nDisallow: /admin",
		SitemapURLs: []string{
			"https://example.com/page1",
			"https://example.com/page2",
		},
	}

	if result.StartURL != "https://example.com" {
		t.Errorf("StartURL mismatch")
	}
	if result.TotalPages != 1 {
		t.Errorf("TotalPages: got %d", result.TotalPages)
	}
	if len(result.SitemapURLs) != 2 {
		t.Errorf("SitemapURLs count: got %d", len(result.SitemapURLs))
	}
}

func TestConfigValidation(t *testing.T) {
	// Test that config with zero values doesn't cause issues
	cfg := Config{
		MaxDepth:    0, // Should allow 0 (no recursion)
		MaxPages:    1,
		WaitForSPA:  false,
		Concurrency: 0, // Should handle gracefully
	}

	// These are just struct values, no validation logic to test
	// but we verify the struct can be created
	if cfg.MaxDepth != 0 {
		t.Errorf("MaxDepth should be 0")
	}
}

func TestNewCrawler(t *testing.T) {
	cfg := DefaultConfig()

	// Note: NewCrawler requires a vibium.Vibe instance, which we can't create in unit tests
	// This test just verifies the function signature and nil handling
	c := NewCrawler(nil, nil, cfg)
	if c == nil {
		t.Fatal("NewCrawler returned nil")
	}
	if c.config.MaxDepth != cfg.MaxDepth {
		t.Errorf("Config not stored correctly")
	}
}

func TestPageWithError(t *testing.T) {
	page := Page{
		URL:            "https://example.com/error",
		Depth:          1,
		DiscoveredFrom: "https://example.com",
		Error:          "timeout waiting for page load",
	}

	if page.Error == "" {
		t.Error("Error should be set")
	}
	if page.Title != "" {
		t.Error("Title should be empty for error page")
	}
}

func TestQueueItemStruct(t *testing.T) {
	item := queueItem{
		url:   "https://example.com/page",
		depth: 2,
		from:  "https://example.com",
	}

	if item.url != "https://example.com/page" {
		t.Errorf("url mismatch")
	}
	if item.depth != 2 {
		t.Errorf("depth: got %d", item.depth)
	}
	if item.from != "https://example.com" {
		t.Errorf("from mismatch")
	}
}
