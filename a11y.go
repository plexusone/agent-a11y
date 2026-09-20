package a11y

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/agent-a11y/config"
	"github.com/plexusone/agent-a11y/llm"
	"github.com/plexusone/agent-a11y/remediation"
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

// OpenACR returns the result as an OpenACR (Open Accessibility Conformance
// Report) document. Only criteria actually evaluated are reported; unevaluated
// criteria are marked "Not Evaluated" rather than assumed conformant.
func (r *Result) OpenACR() ([]byte, error) {
	w := report.NewWriter(report.FormatOpenACR)
	var buf bytes.Buffer
	if err := w.Write(&buf, r.raw); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CriterionResult is the structured, honest per-criterion conformance verdict.
type CriterionResult struct {
	ID          string
	Name        string
	Level       string
	Conformance string
	Method      string
	Evaluated   bool
	Remarks     string
	IssueCount  int
}

// Conformance returns the per-criterion verdicts for this result. Consumers
// (e.g. the canonical AssessmentRecord) project VPAT/OpenACR/other formats from
// these rather than re-deriving them. Unevaluated criteria are honestly marked.
func (r *Result) Conformance() []CriterionResult {
	crs := report.EvaluateConformance(r.raw)
	out := make([]CriterionResult, len(crs))
	for i, c := range crs {
		out[i] = CriterionResult(c)
	}
	return out
}

// pageContext builds the evidence context for proactive criterion evaluation
// from the captured page evidence (screenshot, HTML, language) when available.
func (r *Result) pageContext() llm.PageContext {
	pc := llm.PageContext{URL: r.raw.TargetURL}
	if len(r.raw.Pages) == 0 {
		return pc
	}
	p := r.raw.Pages[0]
	pc.Title = p.Title
	pc.IsSPA = p.IsSPA
	pc.Framework = p.SPAFramework
	if p.Evidence != nil {
		pc.Language = p.Evidence.Language
		pc.HTMLSnippet = p.Evidence.HTML
		if len(p.Evidence.ScreenshotPNG) > 0 {
			pc.Screenshot = base64.StdEncoding.EncodeToString(p.Evidence.ScreenshotPNG)
		}
	}
	return pc
}

// CriterionQueries returns proactive-evaluation queries for the criteria that
// automation left "Not Evaluated" and that are designed for judgment.
func (r *Result) CriterionQueries() []llm.CriterionQuery {
	return report.CriterionQueriesForUnevaluated(r.raw)
}

// EvaluationBundle exports the local-mode worklist: the shared prompt, the page
// evidence, and per-criterion prompts for an external agent to judge (no API
// key). Feed the agent's verdicts back via ConformanceWithVerdicts.
func (r *Result) EvaluationBundle() llm.EvaluationBundle {
	return llm.BuildBundle(r.CriterionQueries(), r.pageContext())
}

// EvaluateCriteria runs proactive evaluation through the given evaluator (the
// API backend or any CriterionEvaluator) and returns its verdicts.
func (r *Result) EvaluateCriteria(ctx context.Context, evaluator llm.CriterionEvaluator) ([]llm.CriterionVerdict, error) {
	return evaluator.EvaluateCriteria(ctx, r.CriterionQueries(), r.pageContext())
}

// ConformanceWithVerdicts returns per-criterion verdicts with proactive
// verdicts (from either mode) applied over the automated baseline.
func (r *Result) ConformanceWithVerdicts(verdicts []llm.CriterionVerdict) []CriterionResult {
	crs := report.ApplyVerdicts(report.EvaluateConformance(r.raw), verdicts)
	out := make([]CriterionResult, len(crs))
	for i, c := range crs {
		out[i] = CriterionResult(c)
	}
	return out
}

// EvaluationMethods lists the human-readable evaluation methods applied.
func (r *Result) EvaluationMethods() []string {
	return report.EvaluationMethods(r.raw)
}

// ConformanceDisclaimer is the honesty disclaimer describing evaluation scope.
func (r *Result) ConformanceDisclaimer() string {
	return report.ConformanceDisclaimer
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

// AgentOptimizedJSON returns the result as agent-optimized JSON with fix patterns.
// If designSystemPath is provided, it will include token suggestions.
func (r *Result) AgentOptimizedJSON(designSystemPath string) ([]byte, error) {
	agentResult, err := r.AgentOptimized(designSystemPath)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(agentResult, "", "  ")
}

// AgentOptimized returns the result as an agent-optimized struct with fix patterns.
// If designSystemPath is provided, it will include token suggestions.
func (r *Result) AgentOptimized(designSystemPath string) (*types.AgentResult, error) {
	transformer, err := remediation.NewTransformer(designSystemPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create transformer: %w", err)
	}

	// Collect all findings from raw result
	var allFindings []types.Finding
	for _, page := range r.raw.Pages {
		allFindings = append(allFindings, page.Findings...)
	}

	// Calculate duration from raw result
	duration := time.Duration(r.raw.Duration) * time.Millisecond

	// Transform to agent format
	return transformer.TransformResult(r.URL, r.Level, duration, allFindings), nil
}
