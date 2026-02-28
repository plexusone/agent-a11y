// Package llm provides LLM-as-a-Judge functionality for accessibility evaluation.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/agentplexus/omnillm"
)

// Judge evaluates accessibility findings using LLM reasoning.
type Judge struct {
	client *omnillm.ChatClient
	model  string
	logger *slog.Logger
	config JudgeConfig
}

// JudgeConfig configures the LLM judge.
type JudgeConfig struct {
	// Categories to evaluate
	Categories []string

	// Confidence threshold for automated decisions
	ConfidenceThreshold float64

	// Maximum concurrent evaluations
	Concurrency int

	// System prompt override
	SystemPrompt string
}

// DefaultJudgeConfig returns default configuration.
func DefaultJudgeConfig() JudgeConfig {
	return JudgeConfig{
		Categories: []string{
			"alternative-text",
			"color-contrast",
			"keyboard-access",
			"form-labels",
			"link-purpose",
		},
		ConfidenceThreshold: 0.8,
		Concurrency:         5,
	}
}

// Finding represents an accessibility finding to be evaluated.
// This is a simplified type to avoid import cycles with the audit package.
type Finding struct {
	ID              string
	RuleID          string
	Description     string
	SuccessCriteria []string
	Level           string
	Impact          string
	Selector        string
	HTML            string
	Help            string
}

// NewJudge creates a new LLM judge using an omnillm client.
func NewJudge(client *omnillm.ChatClient, model string, logger *slog.Logger, config JudgeConfig) *Judge {
	return &Judge{
		client: client,
		model:  model,
		logger: logger,
		config: config,
	}
}

// Evaluation represents the judge's evaluation of a finding.
type Evaluation struct {
	FindingID   string  `json:"findingId"`
	Confirmed   bool    `json:"confirmed"`
	Confidence  float64 `json:"confidence"`
	Severity    string  `json:"severity"`
	Reasoning   string  `json:"reasoning"`
	Remediation string  `json:"remediation"`

	// For items needing manual review
	NeedsManualReview bool   `json:"needsManualReview"`
	ReviewGuidance    string `json:"reviewGuidance,omitempty"`

	// Metadata
	Model     string    `json:"model"`
	EvalTime  time.Time `json:"evalTime"`
	TokensIn  int       `json:"tokensIn"`
	TokensOut int       `json:"tokensOut"`
}

// EvaluateFinding evaluates a single accessibility finding.
func (j *Judge) EvaluateFinding(ctx context.Context, finding Finding, pageContext PageContext) (*Evaluation, error) {
	prompt := j.buildEvaluationPrompt(finding, pageContext)

	maxTokens := 1024
	temperature := 0.1

	req := &omnillm.ChatCompletionRequest{
		Model: j.model,
		Messages: []omnillm.Message{
			{Role: omnillm.RoleSystem, Content: j.getSystemPrompt()},
			{Role: omnillm.RoleUser, Content: prompt},
		},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	resp, err := j.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	content := resp.Choices[0].Message.Content
	eval, err := j.parseEvaluation(content, finding.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse evaluation: %w", err)
	}

	eval.Model = j.model
	eval.EvalTime = time.Now()
	eval.TokensIn = resp.Usage.PromptTokens
	eval.TokensOut = resp.Usage.CompletionTokens

	return eval, nil
}

// EvaluateFindings evaluates multiple findings concurrently.
func (j *Judge) EvaluateFindings(ctx context.Context, findings []Finding, pageContext PageContext) ([]Evaluation, error) {
	results := make([]Evaluation, len(findings))

	// Simple sequential for now - can be parallelized later
	for i, finding := range findings {
		eval, err := j.EvaluateFinding(ctx, finding, pageContext)
		if err != nil {
			j.logger.Warn("failed to evaluate finding", "findingId", finding.ID, "error", err)
			results[i] = Evaluation{
				FindingID:         finding.ID,
				NeedsManualReview: true,
				ReviewGuidance:    fmt.Sprintf("LLM evaluation failed: %v", err),
			}
			continue
		}
		results[i] = *eval
	}

	return results, nil
}

// PageContext provides context about the page being evaluated.
type PageContext struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	IsSPA       bool   `json:"isSPA"`
	Framework   string `json:"framework,omitempty"`
	Language    string `json:"language"`
	Screenshot  string `json:"screenshot,omitempty"` // Base64
	HTMLSnippet string `json:"htmlSnippet,omitempty"`
}

func (j *Judge) getSystemPrompt() string {
	if j.config.SystemPrompt != "" {
		return j.config.SystemPrompt
	}

	return `You are an expert accessibility auditor evaluating WCAG compliance issues.

Your role is to:
1. Confirm or refute automated accessibility findings
2. Assess the severity and impact of confirmed issues
3. Provide clear remediation guidance
4. Identify cases that need manual review

For each finding, analyze:
- Whether the issue is a genuine accessibility barrier
- The impact on users with disabilities
- The WCAG success criterion being violated
- Practical steps to fix the issue

Respond in JSON format with the following structure:
{
  "confirmed": boolean,
  "confidence": float (0.0-1.0),
  "severity": "critical|serious|moderate|minor",
  "reasoning": "Brief explanation of your assessment",
  "remediation": "Specific steps to fix the issue",
  "needsManualReview": boolean,
  "reviewGuidance": "Why manual review is needed (if applicable)"
}

Be precise and objective. If uncertain, mark for manual review rather than making assumptions.`
}

func (j *Judge) buildEvaluationPrompt(finding Finding, pageContext PageContext) string {
	var sb strings.Builder

	sb.WriteString("## Accessibility Finding to Evaluate\n\n")
	fmt.Fprintf(&sb, "**Rule ID:** %s\n", finding.RuleID)
	fmt.Fprintf(&sb, "**Description:** %s\n", finding.Description)
	fmt.Fprintf(&sb, "**WCAG Criteria:** %s\n", strings.Join(finding.SuccessCriteria, ", "))
	fmt.Fprintf(&sb, "**Level:** %s\n", finding.Level)
	fmt.Fprintf(&sb, "**Impact:** %s\n", finding.Impact)

	if finding.Selector != "" {
		fmt.Fprintf(&sb, "**Element Selector:** `%s`\n", finding.Selector)
	}
	if finding.HTML != "" {
		fmt.Fprintf(&sb, "**HTML Snippet:**\n```html\n%s\n```\n", finding.HTML)
	}
	if finding.Help != "" {
		fmt.Fprintf(&sb, "**Help Text:** %s\n", finding.Help)
	}

	sb.WriteString("\n## Page Context\n\n")
	fmt.Fprintf(&sb, "**URL:** %s\n", pageContext.URL)
	fmt.Fprintf(&sb, "**Title:** %s\n", pageContext.Title)
	if pageContext.IsSPA {
		fmt.Fprintf(&sb, "**SPA Framework:** %s\n", pageContext.Framework)
	}
	fmt.Fprintf(&sb, "**Language:** %s\n", pageContext.Language)

	if pageContext.HTMLSnippet != "" {
		fmt.Fprintf(&sb, "\n**Page HTML Snippet:**\n```html\n%s\n```\n", pageContext.HTMLSnippet)
	}

	sb.WriteString("\n## Evaluation Request\n\n")
	sb.WriteString("Please evaluate this finding and respond with your assessment in JSON format.")

	return sb.String()
}

func (j *Judge) parseEvaluation(content string, findingID string) (*Evaluation, error) {
	// Clean up response - sometimes LLMs add markdown code blocks
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var eval Evaluation
	if err := json.Unmarshal([]byte(content), &eval); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	eval.FindingID = findingID
	return &eval, nil
}

// ReviewCategories returns the categories this judge evaluates.
func (j *Judge) ReviewCategories() []string {
	return j.config.Categories
}

// ShouldEvaluate checks if a finding should be evaluated by the LLM.
func (j *Judge) ShouldEvaluate(finding Finding) bool {
	// Check if the finding's criteria match our categories
	for _, criterion := range finding.SuccessCriteria {
		for _, category := range j.config.Categories {
			if matchesCategory(criterion, category) {
				return true
			}
		}
	}
	return false
}

func matchesCategory(criterion, category string) bool {
	categoryMap := map[string][]string{
		"alternative-text": {"1.1.1"},
		"color-contrast":   {"1.4.3", "1.4.6", "1.4.11"},
		"keyboard-access":  {"2.1.1", "2.1.2", "2.1.4", "2.4.3", "2.4.7"},
		"form-labels":      {"1.3.1", "3.3.2", "4.1.2"},
		"link-purpose":     {"2.4.4", "2.4.9"},
	}

	if criteria, ok := categoryMap[category]; ok {
		for _, c := range criteria {
			if criterion == c {
				return true
			}
		}
	}
	return false
}
