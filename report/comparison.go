package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/agentplexus/agent-a11y/audit"
)

// ComparisonResult holds before/after audit results for comparison.
type ComparisonResult struct {
	// Metadata
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	GeneratedAt time.Time `json:"generatedAt"`

	// Before (inaccessible) version
	Before *audit.AuditResult `json:"before"`

	// After (accessible) version
	After *audit.AuditResult `json:"after"`

	// Computed comparison metrics
	Comparison ComparisonMetrics `json:"comparison"`
}

// ComparisonMetrics shows improvement/regression between versions.
type ComparisonMetrics struct {
	// Issue counts
	BeforeTotalIssues int `json:"beforeTotalIssues"`
	AfterTotalIssues  int `json:"afterTotalIssues"`
	IssuesFixed       int `json:"issuesFixed"`
	IssuesRemaining   int `json:"issuesRemaining"`
	NewIssues         int `json:"newIssues"` // Regressions

	// By severity
	CriticalFixed int `json:"criticalFixed"`
	SeriousFixed  int `json:"seriousFixed"`
	ModerateFixed int `json:"moderateFixed"`
	MinorFixed    int `json:"minorFixed"`

	// Conformance changes
	BeforeConformance string `json:"beforeConformance"`
	AfterConformance  string `json:"afterConformance"`

	// Per-criterion comparison
	CriteriaComparison []CriterionComparison `json:"criteriaComparison"`

	// Improvement score (0-100)
	ImprovementScore float64 `json:"improvementScore"`
}

// CriterionComparison shows before/after for a single criterion.
type CriterionComparison struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Level        string `json:"level"`
	BeforeStatus string `json:"beforeStatus"`
	AfterStatus  string `json:"afterStatus"`
	BeforeIssues int    `json:"beforeIssues"`
	AfterIssues  int    `json:"afterIssues"`
	Status       string `json:"status"` // fixed, improved, unchanged, regressed
}

// NewComparison creates a comparison between before and after audit results.
func NewComparison(name string, before, after *audit.AuditResult) *ComparisonResult {
	result := &ComparisonResult{
		Name:        name,
		GeneratedAt: time.Now(),
		Before:      before,
		After:       after,
	}

	result.Comparison = computeComparison(before, after)
	return result
}

func computeComparison(before, after *audit.AuditResult) ComparisonMetrics {
	metrics := ComparisonMetrics{
		BeforeTotalIssues: before.Stats.TotalFindings,
		AfterTotalIssues:  after.Stats.TotalFindings,
	}

	// Calculate fixed issues
	if before.Stats.TotalFindings > after.Stats.TotalFindings {
		metrics.IssuesFixed = before.Stats.TotalFindings - after.Stats.TotalFindings
	}
	metrics.IssuesRemaining = after.Stats.TotalFindings

	// By severity
	metrics.CriticalFixed = max(0, before.Stats.Critical-after.Stats.Critical)
	metrics.SeriousFixed = max(0, before.Stats.Serious-after.Stats.Serious)
	metrics.ModerateFixed = max(0, before.Stats.Moderate-after.Stats.Moderate)
	metrics.MinorFixed = max(0, before.Stats.Minor-after.Stats.Minor)

	// Check for regressions
	if after.Stats.TotalFindings > before.Stats.TotalFindings {
		metrics.NewIssues = after.Stats.TotalFindings - before.Stats.TotalFindings
	}

	// Conformance status
	metrics.BeforeConformance = before.Conformance.OverallStatus
	metrics.AfterConformance = after.Conformance.OverallStatus

	// Per-criterion comparison
	metrics.CriteriaComparison = compareCriteria(before, after)

	// Calculate improvement score
	if before.Stats.TotalFindings > 0 {
		metrics.ImprovementScore = float64(metrics.IssuesFixed) / float64(before.Stats.TotalFindings) * 100
	} else {
		metrics.ImprovementScore = 100 // No issues in before = 100%
	}

	return metrics
}

func compareCriteria(before, after *audit.AuditResult) []CriterionComparison {
	// Build issue counts per criterion for before
	beforeIssues := countIssuesByCriterion(before)
	afterIssues := countIssuesByCriterion(after)

	// Get all WCAG criteria
	criteria := getWCAG22Criteria()

	var comparisons []CriterionComparison
	for _, c := range criteria {
		beforeCount := beforeIssues[c.ID]
		afterCount := afterIssues[c.ID]

		beforeStatus := conformanceStatus(beforeCount)
		afterStatus := conformanceStatus(afterCount)

		status := "unchanged"
		if beforeCount > 0 && afterCount == 0 {
			status = "fixed"
		} else if beforeCount > afterCount {
			status = "improved"
		} else if afterCount > beforeCount {
			status = "regressed"
		}

		comparisons = append(comparisons, CriterionComparison{
			ID:           c.ID,
			Name:         c.Name,
			Level:        c.Level,
			BeforeStatus: beforeStatus,
			AfterStatus:  afterStatus,
			BeforeIssues: beforeCount,
			AfterIssues:  afterCount,
			Status:       status,
		})
	}

	return comparisons
}

func countIssuesByCriterion(result *audit.AuditResult) map[string]int {
	counts := make(map[string]int)
	for _, page := range result.Pages {
		for _, finding := range page.Findings {
			for _, sc := range finding.SuccessCriteria {
				counts[sc]++
			}
		}
	}
	return counts
}

func conformanceStatus(issueCount int) string {
	switch {
	case issueCount == 0:
		return "Supports"
	case issueCount <= 3:
		return "Partially Supports"
	default:
		return "Does Not Support"
	}
}

// WriteComparison writes a comparison report.
func (w *Writer) WriteComparison(out io.Writer, result *ComparisonResult) error {
	switch w.format {
	case FormatJSON:
		return w.writeComparisonJSON(out, result)
	case FormatMarkdown:
		return w.writeComparisonMarkdown(out, result)
	case FormatHTML:
		return w.writeComparisonHTML(out, result)
	case FormatVPAT:
		return w.writeComparisonVPAT(out, result)
	default:
		return w.writeComparisonMarkdown(out, result)
	}
}

func (w *Writer) writeComparisonJSON(out io.Writer, result *ComparisonResult) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func (w *Writer) writeComparisonMarkdown(out io.Writer, result *ComparisonResult) error {
	var sb strings.Builder

	// Header
	sb.WriteString("# Accessibility Comparison Report\n\n")
	fmt.Fprintf(&sb, "**Comparison:** %s\n", result.Name)
	fmt.Fprintf(&sb, "**Generated:** %s\n\n", result.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Overview
	sb.WriteString("## Overview\n\n")
	sb.WriteString("| Version | URL | Total Issues | Conformance |\n")
	sb.WriteString("|---------|-----|--------------|-------------|\n")
	fmt.Fprintf(&sb, "| Before | %s | %d | %s |\n",
		result.Before.TargetURL,
		result.Comparison.BeforeTotalIssues,
		result.Comparison.BeforeConformance)
	fmt.Fprintf(&sb, "| After | %s | %d | %s |\n",
		result.After.TargetURL,
		result.Comparison.AfterTotalIssues,
		result.Comparison.AfterConformance)
	sb.WriteString("\n")

	// Improvement Summary
	sb.WriteString("## Improvement Summary\n\n")
	fmt.Fprintf(&sb, "**Improvement Score:** %.1f%%\n\n", result.Comparison.ImprovementScore)
	fmt.Fprintf(&sb, "- Issues Fixed: **%d**\n", result.Comparison.IssuesFixed)
	fmt.Fprintf(&sb, "- Issues Remaining: %d\n", result.Comparison.IssuesRemaining)
	if result.Comparison.NewIssues > 0 {
		fmt.Fprintf(&sb, "- ⚠️ New Issues (Regressions): **%d**\n", result.Comparison.NewIssues)
	}
	sb.WriteString("\n")

	// By Severity
	sb.WriteString("### Fixed by Severity\n\n")
	sb.WriteString("| Severity | Fixed |\n")
	sb.WriteString("|----------|-------|\n")
	fmt.Fprintf(&sb, "| Critical | %d |\n", result.Comparison.CriticalFixed)
	fmt.Fprintf(&sb, "| Serious | %d |\n", result.Comparison.SeriousFixed)
	fmt.Fprintf(&sb, "| Moderate | %d |\n", result.Comparison.ModerateFixed)
	fmt.Fprintf(&sb, "| Minor | %d |\n", result.Comparison.MinorFixed)
	sb.WriteString("\n")

	// Criterion-by-Criterion
	sb.WriteString("## Criterion Comparison\n\n")
	sb.WriteString("| Criterion | Level | Before | After | Status |\n")
	sb.WriteString("|-----------|-------|--------|-------|--------|\n")

	for _, c := range result.Comparison.CriteriaComparison {
		statusEmoji := "➖"
		switch c.Status {
		case "fixed":
			statusEmoji = "✅"
		case "improved":
			statusEmoji = "📈"
		case "regressed":
			statusEmoji = "⚠️"
		}

		fmt.Fprintf(&sb, "| %s %s | %s | %s (%d) | %s (%d) | %s %s |\n",
			c.ID, c.Name, c.Level,
			c.BeforeStatus, c.BeforeIssues,
			c.AfterStatus, c.AfterIssues,
			statusEmoji, c.Status)
	}
	sb.WriteString("\n")

	// Legend
	sb.WriteString("### Legend\n\n")
	sb.WriteString("- ✅ Fixed: All issues resolved\n")
	sb.WriteString("- 📈 Improved: Fewer issues than before\n")
	sb.WriteString("- ➖ Unchanged: Same number of issues\n")
	sb.WriteString("- ⚠️ Regressed: More issues than before\n")

	_, err := out.Write([]byte(sb.String()))
	return err
}

func (w *Writer) writeComparisonHTML(out io.Writer, result *ComparisonResult) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Accessibility Comparison: %s</title>
    <style>
        body { font-family: system-ui, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        .overview { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin: 20px 0; }
        .card { background: #f5f5f5; padding: 20px; border-radius: 8px; }
        .before { border-left: 4px solid #d32f2f; }
        .after { border-left: 4px solid #388e3c; }
        .score { font-size: 3em; font-weight: bold; text-align: center; margin: 20px 0; }
        .score.good { color: #388e3c; }
        .score.moderate { color: #f57c00; }
        .score.poor { color: #d32f2f; }
        table { width: 100%%; border-collapse: collapse; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f5f5f5; }
        .fixed { color: #388e3c; }
        .improved { color: #1976d2; }
        .regressed { color: #d32f2f; }
        .unchanged { color: #757575; }
    </style>
</head>
<body>
    <h1>Accessibility Comparison: %s</h1>
    <p>Generated: %s</p>

    <div class="score %s">%.0f%% Improvement</div>

    <div class="overview">
        <div class="card before">
            <h2>Before (Inaccessible)</h2>
            <p><strong>URL:</strong> %s</p>
            <p><strong>Issues:</strong> %d</p>
            <p><strong>Status:</strong> %s</p>
        </div>
        <div class="card after">
            <h2>After (Accessible)</h2>
            <p><strong>URL:</strong> %s</p>
            <p><strong>Issues:</strong> %d</p>
            <p><strong>Status:</strong> %s</p>
        </div>
    </div>

    <h2>Summary</h2>
    <ul>
        <li><strong>Issues Fixed:</strong> %d</li>
        <li><strong>Issues Remaining:</strong> %d</li>
        <li><strong>Regressions:</strong> %d</li>
    </ul>

    <h2>Criterion Comparison</h2>
    <table>
        <thead>
            <tr><th>Criterion</th><th>Level</th><th>Before</th><th>After</th><th>Status</th></tr>
        </thead>
        <tbody>
            %s
        </tbody>
    </table>
</body>
</html>`,
		result.Name,
		result.Name,
		result.GeneratedAt.Format("2006-01-02 15:04:05"),
		scoreClass(result.Comparison.ImprovementScore),
		result.Comparison.ImprovementScore,
		result.Before.TargetURL,
		result.Comparison.BeforeTotalIssues,
		result.Comparison.BeforeConformance,
		result.After.TargetURL,
		result.Comparison.AfterTotalIssues,
		result.Comparison.AfterConformance,
		result.Comparison.IssuesFixed,
		result.Comparison.IssuesRemaining,
		result.Comparison.NewIssues,
		generateCriteriaRows(result.Comparison.CriteriaComparison),
	)

	_, err := out.Write([]byte(html))
	return err
}

func scoreClass(score float64) string {
	switch {
	case score >= 80:
		return "good"
	case score >= 50:
		return "moderate"
	default:
		return "poor"
	}
}

func generateCriteriaRows(criteria []CriterionComparison) string {
	var sb strings.Builder
	for _, c := range criteria {
		fmt.Fprintf(&sb, `<tr>
            <td>%s %s</td>
            <td>%s</td>
            <td>%s (%d)</td>
            <td>%s (%d)</td>
            <td class="%s">%s</td>
        </tr>`,
			c.ID, c.Name, c.Level,
			c.BeforeStatus, c.BeforeIssues,
			c.AfterStatus, c.AfterIssues,
			c.Status, c.Status)
	}
	return sb.String()
}

func (w *Writer) writeComparisonVPAT(out io.Writer, result *ComparisonResult) error {
	var sb strings.Builder

	// Header
	sb.WriteString("# VPAT Comparison Report\n\n")
	fmt.Fprintf(&sb, "**Comparison:** %s\n", result.Name)
	fmt.Fprintf(&sb, "**Generated:** %s\n\n", result.GeneratedAt.Format("2006-01-02"))

	// Side-by-side VPAT table
	sb.WriteString("## WCAG 2.2 Conformance Comparison\n\n")
	sb.WriteString("| Criterion | Level | Before | After | Change |\n")
	sb.WriteString("|-----------|-------|--------|-------|--------|\n")

	for _, c := range result.Comparison.CriteriaComparison {
		change := "→"
		if c.Status == "fixed" || c.Status == "improved" {
			change = "✓ Improved"
		} else if c.Status == "regressed" {
			change = "✗ Regressed"
		}

		fmt.Fprintf(&sb, "| %s %s | %s | %s | %s | %s |\n",
			c.ID, c.Name, c.Level,
			c.BeforeStatus,
			c.AfterStatus,
			change)
	}

	sb.WriteString("\n## Summary\n\n")
	fmt.Fprintf(&sb, "- Overall improvement: %.1f%%\n", result.Comparison.ImprovementScore)
	fmt.Fprintf(&sb, "- Issues fixed: %d\n", result.Comparison.IssuesFixed)
	fmt.Fprintf(&sb, "- Before conformance: %s\n", result.Comparison.BeforeConformance)
	fmt.Fprintf(&sb, "- After conformance: %s\n", result.Comparison.AfterConformance)

	_, err := out.Write([]byte(sb.String()))
	return err
}

// DemoSite represents a known accessibility demo site with before/after URLs.
type DemoSite struct {
	Name        string
	Slug        string // Directory-safe name for output
	Description string
	BeforeURL   string
	AfterURL    string
	Source      string // Organization that created it
}

// KnownDemoSites returns a list of known accessibility demo sites.
func KnownDemoSites() []DemoSite {
	return []DemoSite{
		{
			Name:        "W3C BAD (Before-After Demo)",
			Slug:        "w3c-bad",
			Description: "W3C's official before-and-after demonstration site",
			BeforeURL:   "https://www.w3.org/WAI/demos/bad/before/home.html",
			AfterURL:    "https://www.w3.org/WAI/demos/bad/after/home.html",
			Source:      "W3C WAI",
		},
		{
			Name:        "AccessComputing Demo",
			Slug:        "accesscomputing",
			Description: "University of Washington AccessComputing demonstration",
			BeforeURL:   "https://projects.accesscomputing.uw.edu/au/before.html",
			AfterURL:    "https://projects.accesscomputing.uw.edu/au/after.html",
			Source:      "University of Washington",
		},
		{
			Name:        "A11yQuest Demo - Forms",
			Slug:        "a11yquest-forms",
			Description: "A11yQuest accessible forms demonstration",
			BeforeURL:   "https://www.a11yquest.com/demos/forms/before",
			AfterURL:    "https://www.a11yquest.com/demos/forms/after",
			Source:      "A11yQuest",
		},
		{
			Name:        "A11yQuest Demo - Navigation",
			Slug:        "a11yquest-navigation",
			Description: "A11yQuest accessible navigation demonstration",
			BeforeURL:   "https://www.a11yquest.com/demos/navigation/before",
			AfterURL:    "https://www.a11yquest.com/demos/navigation/after",
			Source:      "A11yQuest",
		},
	}
}
