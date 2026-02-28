package a11y

import (
	"log/slog"
	"time"
)

// Level represents WCAG conformance levels.
type Level string

const (
	// LevelA is WCAG Level A (minimum conformance).
	LevelA Level = "A"
	// LevelAA is WCAG Level AA (standard conformance, most common target).
	LevelAA Level = "AA"
	// LevelAAA is WCAG Level AAA (enhanced conformance).
	LevelAAA Level = "AAA"
)

// Version represents WCAG specification versions.
type Version string

const (
	// Version20 is WCAG 2.0.
	Version20 Version = "2.0"
	// Version21 is WCAG 2.1.
	Version21 Version = "2.1"
	// Version22 is WCAG 2.2.
	Version22 Version = "2.2"
)

// Option configures an Auditor.
type Option func(*options)

type options struct {
	// Browser settings
	headless bool
	timeout  time.Duration

	// WCAG settings
	level   Level
	version Version

	// LLM settings
	llmProvider string
	llmModel    string
	llmAPIKey   string

	// Crawl settings
	crawlDepth    int
	crawlMaxPages int
	crawlDelay    time.Duration

	// Logging
	logger *slog.Logger
}

func defaultOptions() *options {
	return &options{
		headless:      false,
		timeout:       2 * time.Minute,
		level:         LevelAA,
		version:       Version22,
		crawlDepth:    3,
		crawlMaxPages: 100,
		crawlDelay:    500 * time.Millisecond,
		logger:        slog.Default(),
	}
}

// WithHeadless configures the browser to run in headless mode.
func WithHeadless(headless bool) Option {
	return func(o *options) {
		o.headless = headless
	}
}

// WithTimeout sets the maximum duration for audit operations.
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

// WithLevel sets the target WCAG conformance level.
func WithLevel(level Level) Option {
	return func(o *options) {
		o.level = level
	}
}

// WithVersion sets the WCAG specification version.
func WithVersion(version Version) Option {
	return func(o *options) {
		o.version = version
	}
}

// WithLLM enables LLM-based evaluation of accessibility issues.
// Provider can be "anthropic", "openai", "gemini", "ollama", or "xai".
func WithLLM(provider, model string) Option {
	return func(o *options) {
		o.llmProvider = provider
		o.llmModel = model
	}
}

// WithLLMAPIKey sets the API key for LLM provider.
func WithLLMAPIKey(apiKey string) Option {
	return func(o *options) {
		o.llmAPIKey = apiKey
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// CrawlOption configures site crawling behavior.
type CrawlOption func(*options)

// CrawlDepth sets the maximum crawl depth.
func CrawlDepth(depth int) CrawlOption {
	return func(o *options) {
		o.crawlDepth = depth
	}
}

// CrawlMaxPages sets the maximum number of pages to crawl.
func CrawlMaxPages(maxPages int) CrawlOption {
	return func(o *options) {
		o.crawlMaxPages = maxPages
	}
}

// CrawlDelay sets the delay between page requests.
func CrawlDelay(delay time.Duration) CrawlOption {
	return func(o *options) {
		o.crawlDelay = delay
	}
}
