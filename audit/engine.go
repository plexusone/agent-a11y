// Package audit provides the core accessibility audit engine.
package audit

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/plexusone/omnillm"
	vibium "github.com/plexusone/vibium-go"

	"github.com/plexusone/agent-a11y/audit/specialized"
	"github.com/plexusone/agent-a11y/auth"
	"github.com/plexusone/agent-a11y/config"
	"github.com/plexusone/agent-a11y/crawler"
	"github.com/plexusone/agent-a11y/journey"
	"github.com/plexusone/agent-a11y/llm"
	"github.com/plexusone/agent-a11y/wcag"
)

// Engine is the main accessibility audit engine.
type Engine struct {
	logger    *slog.Logger
	llmClient *omnillm.ChatClient
	mu        sync.Mutex
}

// EngineConfig configures the audit engine.
type EngineConfig struct {
	// LLM provider configuration
	LLMProvider string // anthropic, openai, gemini, ollama, xai
	LLMAPIKey   string
	LLMModel    string
	LLMBaseURL  string // For ollama or custom endpoints

	// Logger
	Logger *slog.Logger
}

// NewEngine creates a new audit engine.
func NewEngine(cfg EngineConfig) (*Engine, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	// Create omnillm client if LLM is configured
	var llmClient *omnillm.ChatClient
	if cfg.LLMAPIKey != "" || cfg.LLMProvider == "ollama" {
		providerConfig := omnillm.ProviderConfig{
			Provider: omnillm.ProviderName(cfg.LLMProvider),
			APIKey:   cfg.LLMAPIKey,
		}
		if cfg.LLMBaseURL != "" {
			providerConfig.BaseURL = cfg.LLMBaseURL
		}

		client, err := omnillm.NewClient(omnillm.ClientConfig{
			Providers: []omnillm.ProviderConfig{providerConfig},
			Logger:    logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create LLM client: %w", err)
		}
		llmClient = client
	}

	return &Engine{
		logger:    logger,
		llmClient: llmClient,
	}, nil
}

// Close releases resources held by the engine.
func (e *Engine) Close() error {
	if e.llmClient != nil {
		return e.llmClient.Close()
	}
	return nil
}

// RunAudit executes a full accessibility audit.
func (e *Engine) RunAudit(ctx context.Context, cfg *config.Config) (*AuditResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	startTime := time.Now()
	e.logger.Info("starting audit", "url", cfg.URL)

	result := &AuditResult{
		ID:          fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		StartTime:   startTime,
		TargetURL:   cfg.URL,
		WCAGVersion: WCAGVersion(cfg.WCAG.Version),
		WCAGLevel:   WCAGLevel(cfg.WCAG.Level),
		Pages:       []PageResult{},
		Conformance: ConformanceSummary{
			TargetLevel: WCAGLevel(cfg.WCAG.Level),
			Version:     cfg.WCAG.Version,
		},
	}

	// Initialize browser
	vibe, err := e.initBrowser(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize browser: %w", err)
	}
	defer func() {
		if err := vibe.Quit(ctx); err != nil {
			e.logger.Warn("failed to close browser", "error", err)
		}
	}()

	// Handle authentication if configured
	if cfg.Auth != nil && cfg.Auth.Type != "" {
		if err := e.authenticate(ctx, vibe, cfg); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Get pages to audit
	var pagesToAudit []string
	if cfg.Journey != nil && cfg.Journey.Path != "" {
		// Execute journey and collect pages
		journeyPages, err := e.executeJourney(ctx, vibe, cfg)
		if err != nil {
			e.logger.Warn("journey execution failed", "error", err)
		}
		pagesToAudit = journeyPages
	} else if cfg.Crawl != nil {
		// Crawl the site
		crawledPages, err := e.crawlSite(ctx, vibe, cfg)
		if err != nil {
			e.logger.Warn("crawling failed", "error", err)
		}
		pagesToAudit = crawledPages
	} else {
		// Just audit the target URL
		pagesToAudit = []string{cfg.URL}
	}

	// Initialize WCAG rules
	rules := wcag.NewRules(vibe, e.logger)

	// Initialize LLM judge if configured
	var judge *llm.Judge
	if e.llmClient != nil {
		model := cfg.LLM.Model
		if model == "" {
			model = "claude-sonnet-4-20250514"
		}
		judge = llm.NewJudge(e.llmClient, model, e.logger, llm.DefaultJudgeConfig())
	}

	// Audit each page
	for _, pageURL := range pagesToAudit {
		pageResult, err := e.auditPage(ctx, vibe, rules, judge, pageURL, cfg)
		if err != nil {
			e.logger.Warn("failed to audit page", "url", pageURL, "error", err)
			continue
		}
		result.Pages = append(result.Pages, *pageResult)
	}

	// Aggregate results
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime).Milliseconds()
	e.aggregateStats(result)
	e.computeConformance(result)

	e.logger.Info("audit completed",
		"pages", len(result.Pages),
		"findings", result.Stats.TotalFindings,
		"duration", result.Duration,
	)

	return result, nil
}

func (e *Engine) initBrowser(ctx context.Context, cfg *config.Config) (*vibium.Vibe, error) {
	var vibe *vibium.Vibe
	var err error

	if cfg.Browser.Headless {
		vibe, err = vibium.LaunchHeadless(ctx)
	} else {
		vibe, err = vibium.Launch(ctx)
	}
	if err != nil {
		return nil, err
	}

	return vibe, nil
}

func (e *Engine) authenticate(ctx context.Context, vibe *vibium.Vibe, cfg *config.Config) error {
	// Convert config.AuthSelectors to auth.FormSelectors
	var formSelectors *auth.FormSelectors
	if cfg.Auth.Selectors != nil {
		formSelectors = &auth.FormSelectors{
			Username: cfg.Auth.Selectors.Username,
			Password: cfg.Auth.Selectors.Password,
			Submit:   cfg.Auth.Selectors.Submit,
			MFA:      cfg.Auth.Selectors.MFA,
		}
	}

	authCfg := &auth.Config{
		Type:             auth.AuthType(cfg.Auth.Type),
		LoginURL:         cfg.Auth.LoginURL,
		Credentials:      cfg.Auth.Credentials,
		Selectors:        formSelectors,
		Headers:          cfg.Auth.Headers,
		Cookies:          cfg.Auth.Cookies,
		SuccessIndicator: cfg.Auth.SuccessIndicator,
	}

	handler := auth.NewHandler(vibe, e.logger, authCfg)
	return handler.Authenticate(ctx)
}

func (e *Engine) executeJourney(ctx context.Context, vibe *vibium.Vibe, cfg *config.Config) ([]string, error) {
	parser := journey.NewParser()
	def, err := parser.ParseFile(cfg.Journey.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse journey file: %w", err)
	}

	// Create compiler if we have LLM client (for agentic steps)
	var compiler *journey.Compiler
	if e.llmClient != nil {
		model := cfg.LLM.Model
		if model == "" {
			model = "claude-sonnet-4-20250514"
		}
		compiler = journey.NewCompiler(e.llmClient, model, e.logger)
	}

	executor := journey.NewExecutor(vibe, compiler, e.logger)
	state, err := executor.Execute(ctx, def)
	if err != nil {
		return nil, err
	}

	// Collect unique pages visited during journey
	pageSet := make(map[string]bool)
	for _, stepResult := range state.StepResults {
		if stepResult.PageURL != "" {
			pageSet[stepResult.PageURL] = true
		}
	}

	pages := make([]string, 0, len(pageSet))
	for page := range pageSet {
		pages = append(pages, page)
	}

	return pages, nil
}

func (e *Engine) crawlSite(ctx context.Context, vibe *vibium.Vibe, cfg *config.Config) ([]string, error) {
	crawlerCfg := crawler.Config{
		MaxDepth:        cfg.Crawl.Depth,
		MaxPages:        cfg.Crawl.MaxPages,
		IncludePatterns: cfg.Crawl.IncludePatterns,
		ExcludePatterns: cfg.Crawl.ExcludePatterns,
		WaitForSPA:      cfg.Crawl.WaitForSPA,
		SPAIndicators:   cfg.Crawl.SPAIndicators,
		RespectRobots:   cfg.Crawl.RespectRobots,
		UseSitemap:      cfg.Crawl.UseSitemap,
		Delay:           cfg.Crawl.Delay.Duration(),
	}

	c := crawler.NewCrawler(vibe, e.logger, crawlerCfg)
	result, err := c.Crawl(ctx, cfg.URL)
	if err != nil {
		return nil, err
	}

	// Extract URLs from crawl results
	urls := make([]string, len(result.Pages))
	for i, page := range result.Pages {
		urls[i] = page.URL
	}

	return urls, nil
}

func (e *Engine) auditPage(ctx context.Context, vibe *vibium.Vibe, rules *wcag.Rules, judge *llm.Judge, pageURL string, cfg *config.Config) (*PageResult, error) {
	e.logger.Debug("auditing page", "url", pageURL)

	// Navigate to page
	if err := vibe.Go(ctx, pageURL); err != nil {
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	// Wait for page to load
	if err := vibe.WaitForLoad(ctx, "load", time.Duration(cfg.Browser.Timeout)); err != nil {
		e.logger.Warn("timeout waiting for page load", "url", pageURL)
	}

	// Get page info
	title, _ := vibe.Evaluate(ctx, "document.title")
	titleStr, _ := title.(string)

	currentURL, _ := vibe.URL(ctx)

	pageResult := &PageResult{
		URL:       currentURL,
		Title:     titleStr,
		StartTime: time.Now(),
		Findings:  []Finding{},
	}

	// Run WCAG rules
	targetLevel := WCAGLevel(cfg.WCAG.Level)
	findings, err := rules.RunAll(ctx, string(targetLevel))
	if err != nil {
		e.logger.Warn("error running WCAG rules", "url", pageURL, "error", err)
	}
	pageResult.Findings = findings

	// Run specialized tests (keyboard, focus, reflow, etc.)
	e.logger.Debug("running specialized tests", "url", pageURL)
	specRunner := specialized.NewRunner(vibe, e.logger)
	specResults, err := specRunner.RunAll(ctx)
	if err != nil {
		e.logger.Warn("error running specialized tests", "url", pageURL, "error", err)
	} else if specResults != nil {
		// Convert specialized findings to audit findings
		for _, sf := range specResults.Findings {
			pageResult.Findings = append(pageResult.Findings, fromSpecializedFinding(sf, currentURL, titleStr))
		}
		e.logger.Debug("specialized tests complete", "url", pageURL, "findings", len(specResults.Findings))
	}

	// Detect SPA framework
	framework, isSPA := e.detectSPAFramework(ctx, vibe)

	// Get page language
	lang, _ := vibe.Evaluate(ctx, "document.documentElement.lang || ''")
	langStr, _ := lang.(string)

	// Run LLM evaluation if judge is available
	if judge != nil && len(findings) > 0 {
		pageContext := llm.PageContext{
			URL:       currentURL,
			Title:     titleStr,
			IsSPA:     isSPA,
			Framework: framework,
			Language:  langStr,
		}

		// Convert and filter findings that should be evaluated
		var toEvaluate []llm.Finding
		for _, f := range findings {
			llmFinding := toLLMFinding(f)
			if judge.ShouldEvaluate(llmFinding) {
				toEvaluate = append(toEvaluate, llmFinding)
			}
		}

		if len(toEvaluate) > 0 {
			evaluations, err := judge.EvaluateFindings(ctx, toEvaluate, pageContext)
			if err != nil {
				e.logger.Warn("LLM evaluation failed", "error", err)
			} else {
				// Attach evaluations to findings
				evalMap := make(map[string]llm.Evaluation)
				for _, eval := range evaluations {
					evalMap[eval.FindingID] = eval
				}

				for i := range pageResult.Findings {
					if eval, ok := evalMap[pageResult.Findings[i].ID]; ok {
						pageResult.Findings[i].LLMEvaluation = &LLMEvaluation{
							Confirmed:         eval.Confirmed,
							Confidence:        eval.Confidence,
							Severity:          eval.Severity,
							Reasoning:         eval.Reasoning,
							Remediation:       eval.Remediation,
							NeedsManualReview: eval.NeedsManualReview,
							ReviewGuidance:    eval.ReviewGuidance,
							Model:             eval.Model,
							EvalTime:          eval.EvalTime,
						}
					}
				}
			}
		}
	}

	pageResult.EndTime = time.Now()
	pageResult.Duration = pageResult.EndTime.Sub(pageResult.StartTime).Milliseconds()

	return pageResult, nil
}

func (e *Engine) detectSPAFramework(ctx context.Context, vibe *vibium.Vibe) (string, bool) {
	script := `
		(function() {
			if (window.__NEXT_DATA__) return 'nextjs';
			if (window.__NUXT__) return 'nuxt';
			if (document.querySelector('[ng-version]')) return 'angular';
			if (document.querySelector('[data-reactroot]') || document.querySelector('#__next')) return 'react';
			if (window.__VUE__) return 'vue';
			return '';
		})()
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return "", false
	}

	framework, ok := result.(string)
	if !ok || framework == "" {
		return "", false
	}

	return framework, true
}

func (e *Engine) aggregateStats(result *AuditResult) {
	stats := AuditStats{
		TotalPages: len(result.Pages),
	}

	for _, page := range result.Pages {
		for _, finding := range page.Findings {
			stats.TotalFindings++

			switch finding.Severity {
			case SeverityCritical:
				stats.Critical++
			case SeveritySerious:
				stats.Serious++
			case SeverityModerate:
				stats.Moderate++
			case SeverityMinor:
				stats.Minor++
			}
		}
	}

	result.Stats = stats
}

func (e *Engine) computeConformance(result *AuditResult) {
	// Count issues by level
	levelAIssues := 0
	levelAAIssues := 0
	levelAAAIssues := 0

	for _, page := range result.Pages {
		for _, finding := range page.Findings {
			switch finding.Level {
			case "A":
				levelAIssues++
			case "AA":
				levelAAIssues++
			case "AAA":
				levelAAAIssues++
			}
		}
	}

	// Determine conformance status
	result.Conformance.LevelA = LevelConformance{
		TotalIssues: levelAIssues,
		Status:      conformanceStatus(levelAIssues),
	}

	result.Conformance.LevelAA = LevelConformance{
		TotalIssues: levelAAIssues,
		Status:      conformanceStatus(levelAAIssues),
	}

	result.Conformance.LevelAAA = LevelConformance{
		TotalIssues: levelAAAIssues,
		Status:      conformanceStatus(levelAAAIssues),
	}

	// Overall conformance
	if levelAIssues == 0 && levelAAIssues == 0 && result.Conformance.TargetLevel == "AA" {
		result.Conformance.OverallStatus = "Conformant"
	} else if levelAIssues == 0 && result.Conformance.TargetLevel == "A" {
		result.Conformance.OverallStatus = "Conformant"
	} else {
		result.Conformance.OverallStatus = "Non-Conformant"
	}
}

func conformanceStatus(issues int) string {
	if issues == 0 {
		return "Supports"
	} else if issues <= 3 {
		return "Partially Supports"
	}
	return "Does Not Support"
}

// toLLMFinding converts an audit.Finding to llm.Finding for LLM evaluation.
func toLLMFinding(f Finding) llm.Finding {
	return llm.Finding{
		ID:              f.ID,
		RuleID:          f.RuleID,
		Description:     f.Description,
		SuccessCriteria: f.SuccessCriteria,
		Level:           string(f.Level),
		Impact:          string(f.Impact),
		Selector:        f.Selector,
		HTML:            f.HTML,
		Help:            f.Help,
	}
}

// fromSpecializedFinding converts a specialized.Finding to audit.Finding.
func fromSpecializedFinding(f specialized.Finding, pageURL, pageTitle string) Finding {
	// Convert string level to WCAGLevel
	var level WCAGLevel
	switch f.Level {
	case "A":
		level = WCAGLevelA
	case "AA":
		level = WCAGLevelAA
	case "AAA":
		level = WCAGLevelAAA
	default:
		level = WCAGLevelA
	}

	// Convert string impact to Impact
	var impact Impact
	switch f.Impact {
	case "critical":
		impact = ImpactCritical
	case "serious":
		impact = ImpactSerious
	case "moderate":
		impact = ImpactModerate
	case "minor":
		impact = ImpactMinor
	default:
		impact = ImpactModerate
	}

	// Map impact to severity
	var severity Severity
	switch impact {
	case ImpactCritical:
		severity = SeverityCritical
	case ImpactSerious:
		severity = SeveritySerious
	case ImpactModerate:
		severity = SeverityModerate
	case ImpactMinor:
		severity = SeverityMinor
	default:
		severity = SeverityModerate
	}

	return Finding{
		ID:              f.ID,
		RuleID:          f.RuleID,
		Description:     f.Description,
		Help:            f.Help,
		SuccessCriteria: f.SuccessCriteria,
		Level:           level,
		Impact:          impact,
		Severity:        severity,
		Selector:        f.Selector,
		HTML:            f.HTML,
		PageURL:         pageURL,
		PageTitle:       pageTitle,
	}
}
