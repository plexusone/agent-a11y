package a11y

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/agent-a11y/config"
	"github.com/plexusone/agent-a11y/report"
	"github.com/plexusone/agent-a11y/types"
)

// Auditor performs WCAG accessibility audits.
type Auditor struct {
	opts   *options
	engine *audit.Engine
}

// New creates a new Auditor with the given options.
func New(opts ...Option) (*Auditor, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	engineCfg := audit.EngineConfig{
		LLMProvider: o.llmProvider,
		LLMModel:    o.llmModel,
		LLMAPIKey:   o.llmAPIKey,
		Logger:      o.logger,
	}

	engine, err := audit.NewEngine(engineCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit engine: %w", err)
	}

	return &Auditor{
		opts:   o,
		engine: engine,
	}, nil
}

// Close releases resources used by the auditor.
func (a *Auditor) Close() error {
	if a.engine != nil {
		return a.engine.Close()
	}
	return nil
}

// AuditPage performs an accessibility audit on a single page.
func (a *Auditor) AuditPage(ctx context.Context, url string) (*Result, error) {
	cfg := a.buildConfig(url)
	cfg.Crawl = nil // Single page only

	auditResult, err := a.engine.RunAudit(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("audit failed: %w", err)
	}

	return newResult(auditResult), nil
}

// AuditSite performs an accessibility audit on an entire website by crawling.
func (a *Auditor) AuditSite(ctx context.Context, url string, crawlOpts ...CrawlOption) (*Result, error) {
	// Apply crawl options
	for _, opt := range crawlOpts {
		opt(a.opts)
	}

	cfg := a.buildConfig(url)
	cfg.Crawl = &config.CrawlConfig{
		Depth:    a.opts.crawlDepth,
		MaxPages: a.opts.crawlMaxPages,
		Delay:    config.Duration(a.opts.crawlDelay),
	}

	auditResult, err := a.engine.RunAudit(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("audit failed: %w", err)
	}

	return newResult(auditResult), nil
}

// AuditJourney performs an accessibility audit using a journey definition file.
func (a *Auditor) AuditJourney(ctx context.Context, url, journeyPath string) (*Result, error) {
	cfg := a.buildConfig(url)
	cfg.Journey = &config.JourneyRef{
		Path: journeyPath,
	}

	auditResult, err := a.engine.RunAudit(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("audit failed: %w", err)
	}

	return newResult(auditResult), nil
}

func (a *Auditor) buildConfig(url string) *config.Config {
	return &config.Config{
		URL: url,
		WCAG: config.WCAGConfig{
			Level:   string(a.opts.level),
			Version: string(a.opts.version),
		},
		Browser: config.BrowserConfig{
			Headless: a.opts.headless,
			Timeout:  config.Duration(a.opts.timeout),
		},
		LLM: &config.LLMConfig{
			Enabled:  a.opts.llmProvider != "",
			Provider: a.opts.llmProvider,
			Model:    a.opts.llmModel,
			APIKey:   a.opts.llmAPIKey,
		},
	}
}

// Result represents the outcome of an accessibility audit.
type Result struct {
	// URL is the audited URL.
	URL string

	// Score is the overall conformance score (0-100).
	Score int

	// Level is the target WCAG level.
	Level string

	// Version is the WCAG version used.
	Version string

	// Findings contains all accessibility issues found.
	Findings []Finding

	// Pages contains per-page results (for site audits).
	Pages []PageResult

	// Stats contains summary statistics.
	Stats Stats

	// raw holds the original audit result for report generation
	raw *audit.AuditResult
}

// Finding represents a single accessibility issue.
type Finding struct {
	// ID is the unique identifier for this finding.
	ID string

	// RuleID identifies the WCAG rule that was violated.
	RuleID string

	// Description explains the issue.
	Description string

	// Help provides guidance on how to fix the issue.
	Help string

	// SuccessCriteria lists the WCAG success criteria affected.
	SuccessCriteria []string

	// Level is the WCAG level (A, AA, AAA).
	Level string

	// Impact indicates the severity (critical, serious, moderate, minor).
	Impact string

	// Element is the HTML element type.
	Element string

	// Selector is the CSS selector to find the element.
	Selector string

	// HTML is a snippet of the problematic HTML.
	HTML string

	// PageURL is the URL where this issue was found.
	PageURL string

	// LLMConfirmed indicates if LLM evaluation confirmed the issue.
	LLMConfirmed *bool

	// LLMReasoning is the LLM's explanation.
	LLMReasoning string
}

// PageResult represents results for a single page.
type PageResult struct {
	URL          string
	Title        string
	FindingCount int
	Score        int
}

// Stats contains summary statistics for the audit.
type Stats struct {
	TotalPages    int
	TotalFindings int
	Critical      int
	Serious       int
	Moderate      int
	Minor         int
	LevelA        int
	LevelAA       int
	LevelAAA      int
}

func newResult(ar *audit.AuditResult) *Result {
	r := &Result{
		URL:     ar.TargetURL,
		Level:   string(ar.WCAGLevel),
		Version: string(ar.WCAGVersion),
		raw:     ar,
		Stats: Stats{
			TotalPages:    ar.Stats.TotalPages,
			TotalFindings: ar.Stats.TotalFindings,
			Critical:      ar.Stats.Critical,
			Serious:       ar.Stats.Serious,
			Moderate:      ar.Stats.Moderate,
			Minor:         ar.Stats.Minor,
			LevelA:        ar.Stats.LevelA,
			LevelAA:       ar.Stats.LevelAA,
			LevelAAA:      ar.Stats.LevelAAA,
		},
	}

	// Calculate score based on findings
	r.Score = calculateScore(ar.Stats)

	// Convert pages
	for _, p := range ar.Pages {
		r.Pages = append(r.Pages, PageResult{
			URL:          p.URL,
			Title:        p.Title,
			FindingCount: len(p.Findings),
		})

		// Collect all findings
		for _, f := range p.Findings {
			finding := convertFinding(f, p.URL)
			r.Findings = append(r.Findings, finding)
		}
	}

	return r
}

func convertFinding(f audit.Finding, pageURL string) Finding {
	finding := Finding{
		ID:              f.ID,
		RuleID:          f.RuleID,
		Description:     f.Description,
		Help:            f.Help,
		SuccessCriteria: f.SuccessCriteria,
		Level:           string(f.Level),
		Impact:          string(f.Impact),
		Element:         f.Element,
		Selector:        f.Selector,
		HTML:            f.HTML,
		PageURL:         pageURL,
	}

	if f.LLMEvaluation != nil {
		finding.LLMConfirmed = &f.LLMEvaluation.Confirmed
		finding.LLMReasoning = f.LLMEvaluation.Reasoning
	}

	return finding
}

func calculateScore(stats audit.AuditStats) int {
	if stats.TotalFindings == 0 {
		return 100
	}

	// Deduct points based on severity
	deductions := stats.Critical*20 + stats.Serious*10 + stats.Moderate*5 + stats.Minor*2
	score := 100 - deductions

	if score < 0 {
		score = 0
	}
	return score
}

// JSON returns the result as JSON bytes.
func (r *Result) JSON() ([]byte, error) {
	return json.MarshalIndent(r.raw, "", "  ")
}

// HTML returns the result as an HTML report.
func (r *Result) HTML() ([]byte, error) {
	w := report.NewWriter(report.FormatHTML)
	var buf bytes.Buffer
	if err := w.Write(&buf, r.raw); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Markdown returns the result as a Markdown report.
func (r *Result) Markdown() ([]byte, error) {
	w := report.NewWriter(report.FormatMarkdown)
	var buf bytes.Buffer
	if err := w.Write(&buf, r.raw); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// VPAT returns the result as a VPAT 2.4 report.
func (r *Result) VPAT() ([]byte, error) {
	w := report.NewWriter(report.FormatVPAT)
	var buf bytes.Buffer
	if err := w.Write(&buf, r.raw); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WCAG returns the result in WCAG-EM format.
func (r *Result) WCAG() ([]byte, error) {
	w := report.NewWriter(report.FormatWCAG)
	var buf bytes.Buffer
	if err := w.Write(&buf, r.raw); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Conformant returns true if the audit passed at the target level.
func (r *Result) Conformant() bool {
	// Conformant if no critical or serious issues at target level
	switch Level(r.Level) {
	case LevelA:
		return r.Stats.Critical == 0 && r.Stats.Serious == 0
	case LevelAA:
		return r.Stats.Critical == 0 && r.Stats.Serious == 0
	case LevelAAA:
		return r.Stats.Critical == 0 && r.Stats.Serious == 0 && r.Stats.Moderate == 0
	}
	return false
}

// Summary returns a brief text summary of the audit.
func (r *Result) Summary() string {
	status := "Non-Conformant"
	if r.Conformant() {
		status = "Conformant"
	}

	return fmt.Sprintf(
		"WCAG %s Level %s: %s (Score: %d/100, Issues: %d critical, %d serious, %d moderate, %d minor)",
		r.Version, r.Level, status,
		r.Score,
		r.Stats.Critical, r.Stats.Serious, r.Stats.Moderate, r.Stats.Minor,
	)
}

// FindingsByLevel returns findings filtered by WCAG level.
func (r *Result) FindingsByLevel(level Level) []Finding {
	var filtered []Finding
	for _, f := range r.Findings {
		if Level(f.Level) == level {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FindingsByCriterion returns findings for a specific success criterion.
func (r *Result) FindingsByCriterion(criterion string) []Finding {
	var filtered []Finding
	for _, f := range r.Findings {
		for _, sc := range f.SuccessCriteria {
			if sc == criterion {
				filtered = append(filtered, f)
				break
			}
		}
	}
	return filtered
}

// CriticalFindings returns only critical severity findings.
func (r *Result) CriticalFindings() []Finding {
	return r.findingsByImpact(string(types.ImpactCritical))
}

// SeriousFindings returns only serious severity findings.
func (r *Result) SeriousFindings() []Finding {
	return r.findingsByImpact(string(types.ImpactSerious))
}

func (r *Result) findingsByImpact(impact string) []Finding {
	var filtered []Finding
	for _, f := range r.Findings {
		if f.Impact == impact {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
