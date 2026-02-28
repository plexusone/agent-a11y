// Package config provides configuration types for the accessibility audit service.
package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete audit configuration.
type Config struct {
	// URL is the starting URL for the audit.
	URL string `yaml:"url" json:"url"`

	// Auth contains authentication configuration.
	Auth *AuthConfig `yaml:"auth,omitempty" json:"auth,omitempty"`

	// Crawl contains crawling configuration.
	Crawl *CrawlConfig `yaml:"crawl,omitempty" json:"crawl,omitempty"`

	// Journey specifies a user journey to execute.
	Journey *JourneyRef `yaml:"journey,omitempty" json:"journey,omitempty"`

	// WCAG contains WCAG testing configuration.
	WCAG WCAGConfig `yaml:"wcag" json:"wcag"`

	// LLM contains LLM-as-a-Judge configuration.
	LLM *LLMConfig `yaml:"llm,omitempty" json:"llm,omitempty"`

	// Output contains output configuration.
	Output OutputConfig `yaml:"output" json:"output"`

	// Browser contains browser configuration.
	Browser BrowserConfig `yaml:"browser" json:"browser"`
}

// AuthConfig contains authentication settings.
type AuthConfig struct {
	// Type is the authentication type: form, oauth, basic, bearer.
	Type string `yaml:"type" json:"type"`

	// LoginURL is the URL of the login page.
	LoginURL string `yaml:"loginUrl,omitempty" json:"loginUrl,omitempty"`

	// Credentials contains login credentials.
	Credentials map[string]string `yaml:"credentials,omitempty" json:"credentials,omitempty"`

	// Selectors contains CSS selectors for form elements.
	Selectors *AuthSelectors `yaml:"selectors,omitempty" json:"selectors,omitempty"`

	// Headers contains authentication headers (for bearer/basic).
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// Cookies contains authentication cookies.
	Cookies map[string]string `yaml:"cookies,omitempty" json:"cookies,omitempty"`

	// SuccessIndicator is a selector that indicates successful login.
	SuccessIndicator string `yaml:"successIndicator,omitempty" json:"successIndicator,omitempty"`
}

// AuthSelectors contains selectors for form-based authentication.
type AuthSelectors struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Submit   string `yaml:"submit" json:"submit"`
	MFA      string `yaml:"mfa,omitempty" json:"mfa,omitempty"`
}

// CrawlConfig contains crawling settings.
type CrawlConfig struct {
	// Depth is the maximum link depth to crawl.
	Depth int `yaml:"depth" json:"depth"`

	// MaxPages is the maximum number of pages to audit.
	MaxPages int `yaml:"maxPages" json:"maxPages"`

	// IncludePatterns are glob patterns for URLs to include.
	IncludePatterns []string `yaml:"includePatterns,omitempty" json:"includePatterns,omitempty"`

	// ExcludePatterns are glob patterns for URLs to exclude.
	ExcludePatterns []string `yaml:"excludePatterns,omitempty" json:"excludePatterns,omitempty"`

	// WaitForSPA enables waiting for SPA hydration.
	WaitForSPA bool `yaml:"waitForSPA" json:"waitForSPA"`

	// SPAIndicators are selectors that indicate SPA readiness.
	SPAIndicators []string `yaml:"spaIndicators,omitempty" json:"spaIndicators,omitempty"`

	// RespectRobots respects robots.txt rules.
	RespectRobots bool `yaml:"respectRobots" json:"respectRobots"`

	// UseSitemap uses sitemap.xml for discovery.
	UseSitemap bool `yaml:"useSitemap" json:"useSitemap"`

	// Delay between page requests.
	Delay Duration `yaml:"delay,omitempty" json:"delay,omitempty"`
}

// JourneyRef references a journey definition.
type JourneyRef struct {
	// ID is the journey ID (for stored journeys).
	ID string `yaml:"id,omitempty" json:"id,omitempty"`

	// Path is a file path to a journey definition.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Inline contains an inline journey definition.
	Inline *JourneyInline `yaml:"inline,omitempty" json:"inline,omitempty"`
}

// JourneyInline contains an inline journey definition.
type JourneyInline struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Mode        string         `yaml:"mode" json:"mode"` // agentic, deterministic, hybrid
	Steps       []JourneyStep  `yaml:"steps,omitempty" json:"steps,omitempty"`
	Goal        string         `yaml:"goal,omitempty" json:"goal,omitempty"`
	TestData    map[string]any `yaml:"testData,omitempty" json:"testData,omitempty"`
	AuditPoints []AuditPoint   `yaml:"auditPoints,omitempty" json:"auditPoints,omitempty"`
}

// JourneyStep represents a step in a journey.
type JourneyStep struct {
	// For deterministic mode
	Action   string `yaml:"action,omitempty" json:"action,omitempty"`
	Selector string `yaml:"selector,omitempty" json:"selector,omitempty"`
	Value    string `yaml:"value,omitempty" json:"value,omitempty"`
	URL      string `yaml:"url,omitempty" json:"url,omitempty"`
	WaitFor  string `yaml:"waitFor,omitempty" json:"waitFor,omitempty"`

	// For agentic/hybrid mode
	Prompt       string         `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Instructions string         `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	Data         map[string]any `yaml:"data,omitempty" json:"data,omitempty"`
	File         string         `yaml:"file,omitempty" json:"file,omitempty"`

	// Audit trigger
	Audit     bool   `yaml:"audit,omitempty" json:"audit,omitempty"`
	AuditName string `yaml:"auditName,omitempty" json:"auditName,omitempty"`

	// Flow control
	UseAuth bool `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// AuditPoint defines when to audit during a journey.
type AuditPoint struct {
	Description string `yaml:"description" json:"description"`
	WaitFor     string `yaml:"waitFor,omitempty" json:"waitFor,omitempty"`
	Condition   string `yaml:"condition,omitempty" json:"condition,omitempty"`
}

// WCAGConfig contains WCAG testing settings.
type WCAGConfig struct {
	// Level is the conformance level: A, AA, AAA.
	Level string `yaml:"level" json:"level"`

	// Version is the WCAG version: 2.0, 2.1, 2.2.
	Version string `yaml:"version" json:"version"`

	// Categories to test (empty = all).
	Categories []string `yaml:"categories,omitempty" json:"categories,omitempty"`

	// Rules to skip.
	SkipRules []string `yaml:"skipRules,omitempty" json:"skipRules,omitempty"`

	// Rules to include (empty = all applicable).
	IncludeRules []string `yaml:"includeRules,omitempty" json:"includeRules,omitempty"`
}

// LLMConfig contains LLM-as-a-Judge settings.
type LLMConfig struct {
	// Enabled enables LLM-based evaluation.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Provider is the LLM provider: anthropic, openai.
	Provider string `yaml:"provider" json:"provider"`

	// Model is the model ID.
	Model string `yaml:"model" json:"model"`

	// APIKey is the API key (or use env var).
	APIKey string `yaml:"apiKey,omitempty" json:"apiKey,omitempty"`

	// JudgeCategories are the categories to evaluate with LLM.
	JudgeCategories []string `yaml:"judgeCategories,omitempty" json:"judgeCategories,omitempty"`

	// CompileJourneys enables journey prompt compilation.
	CompileJourneys bool `yaml:"compileJourneys,omitempty" json:"compileJourneys,omitempty"`
}

// OutputConfig contains output settings.
type OutputConfig struct {
	// Formats are the output formats to generate.
	Formats []string `yaml:"formats" json:"formats"` // json, html, vpat, wcag, markdown

	// Directory is the output directory.
	Directory string `yaml:"directory" json:"directory"`

	// Screenshots enables screenshot capture.
	Screenshots bool `yaml:"screenshots" json:"screenshots"`

	// WebhookURL for progress notifications.
	WebhookURL string `yaml:"webhookUrl,omitempty" json:"webhookUrl,omitempty"`
}

// BrowserConfig contains browser settings.
type BrowserConfig struct {
	// Headless runs in headless mode.
	Headless bool `yaml:"headless" json:"headless"`

	// Timeout is the default page timeout.
	Timeout Duration `yaml:"timeout" json:"timeout"`

	// Viewport sets the viewport size.
	Viewport *Viewport `yaml:"viewport,omitempty" json:"viewport,omitempty"`

	// UserAgent overrides the user agent.
	UserAgent string `yaml:"userAgent,omitempty" json:"userAgent,omitempty"`
}

// Viewport represents browser viewport dimensions.
type Viewport struct {
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
}

// Duration is a time.Duration that can be unmarshaled from YAML/JSON.
type Duration time.Duration

func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		WCAG: WCAGConfig{
			Level:   "AA",
			Version: "2.2",
		},
		Browser: BrowserConfig{
			Headless: true,
			Timeout:  Duration(30 * time.Second),
		},
		Output: OutputConfig{
			Formats:     []string{"json", "html"},
			Directory:   "./reports",
			Screenshots: true,
		},
		Crawl: &CrawlConfig{
			Depth:      2,
			MaxPages:   50,
			WaitForSPA: true,
			SPAIndicators: []string{
				"[data-reactroot]",
				"#__next",
				"[data-v-]",
				"[ng-app]",
			},
		},
	}
}
