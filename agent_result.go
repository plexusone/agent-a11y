package a11y

import (
	"fmt"
	"strings"
	"time"

	mas "github.com/plexusone/multi-agent-spec/sdk/go"
)

// AgentResult converts the audit result to a multi-agent-spec AgentResult.
// This allows agent-a11y to be used as part of a multi-agent workflow,
// returning standardized go/no-go decisions and narrative reports.
func (r *Result) AgentResult() *mas.AgentResult {
	tasks := r.toTaskResults()
	blocks := r.toContentBlocks()

	result := &mas.AgentResult{
		Schema:        "https://raw.githubusercontent.com/plexusone/multi-agent-spec/main/schema/report/agent-result.schema.json",
		AgentID:       "a11y",
		StepID:        "accessibility-audit",
		Tasks:         tasks,
		ContentBlocks: blocks,
		ExecutedAt:    time.Now().UTC(),
		Outputs: map[string]interface{}{
			"url":       r.URL,
			"score":     r.Score,
			"level":     r.Level,
			"version":   r.Version,
			"conformant": r.Conformant(),
		},
	}

	result.Status = result.ComputeStatus()
	return result
}

// toTaskResults converts findings to multi-agent-spec TaskResults.
func (r *Result) toTaskResults() []mas.TaskResult {
	tasks := make([]mas.TaskResult, 0)

	// Add overall conformance check
	conformanceStatus := mas.StatusGo
	conformanceDetail := fmt.Sprintf("Score: %d/100", r.Score)
	if !r.Conformant() {
		if r.Stats.Critical > 0 {
			conformanceStatus = mas.StatusNoGo
			conformanceDetail = fmt.Sprintf("Score: %d/100 - %d critical issues", r.Score, r.Stats.Critical)
		} else if r.Stats.Serious > 0 {
			conformanceStatus = mas.StatusNoGo
			conformanceDetail = fmt.Sprintf("Score: %d/100 - %d serious issues", r.Score, r.Stats.Serious)
		} else {
			conformanceStatus = mas.StatusWarn
			conformanceDetail = fmt.Sprintf("Score: %d/100 - %d moderate issues", r.Score, r.Stats.Moderate)
		}
	}

	tasks = append(tasks, mas.TaskResult{
		ID:       "conformance",
		Status:   conformanceStatus,
		Severity: r.conformanceSeverity(),
		Detail:   conformanceDetail,
	})

	// Add task for each WCAG level
	if r.Stats.LevelA > 0 || r.Level == "A" || r.Level == "AA" || r.Level == "AAA" {
		levelAStatus := mas.StatusGo
		if r.Stats.LevelA > 0 {
			levelAStatus = mas.StatusNoGo
		}
		tasks = append(tasks, mas.TaskResult{
			ID:       "wcag-level-a",
			Status:   levelAStatus,
			Severity: severityFromCount(r.Stats.LevelA),
			Detail:   fmt.Sprintf("%d Level A issues", r.Stats.LevelA),
		})
	}

	if r.Stats.LevelAA > 0 || r.Level == "AA" || r.Level == "AAA" {
		levelAAStatus := mas.StatusGo
		if r.Stats.LevelAA > 0 {
			levelAAStatus = mas.StatusNoGo
		}
		tasks = append(tasks, mas.TaskResult{
			ID:       "wcag-level-aa",
			Status:   levelAAStatus,
			Severity: severityFromCount(r.Stats.LevelAA),
			Detail:   fmt.Sprintf("%d Level AA issues", r.Stats.LevelAA),
		})
	}

	if r.Stats.LevelAAA > 0 || r.Level == "AAA" {
		levelAAAStatus := mas.StatusGo
		if r.Stats.LevelAAA > 0 {
			levelAAAStatus = mas.StatusWarn // AAA is advisory
		}
		tasks = append(tasks, mas.TaskResult{
			ID:       "wcag-level-aaa",
			Status:   levelAAAStatus,
			Severity: severityFromCount(r.Stats.LevelAAA),
			Detail:   fmt.Sprintf("%d Level AAA issues", r.Stats.LevelAAA),
		})
	}

	// Add severity breakdown tasks
	if r.Stats.Critical > 0 {
		tasks = append(tasks, mas.TaskResult{
			ID:       "critical-issues",
			Status:   mas.StatusNoGo,
			Severity: "critical",
			Detail:   fmt.Sprintf("%d critical accessibility issues", r.Stats.Critical),
		})
	}

	if r.Stats.Serious > 0 {
		tasks = append(tasks, mas.TaskResult{
			ID:       "serious-issues",
			Status:   mas.StatusNoGo,
			Severity: "high",
			Detail:   fmt.Sprintf("%d serious accessibility issues", r.Stats.Serious),
		})
	}

	if r.Stats.Moderate > 0 {
		tasks = append(tasks, mas.TaskResult{
			ID:       "moderate-issues",
			Status:   mas.StatusWarn,
			Severity: "medium",
			Detail:   fmt.Sprintf("%d moderate accessibility issues", r.Stats.Moderate),
		})
	}

	if r.Stats.Minor > 0 {
		tasks = append(tasks, mas.TaskResult{
			ID:       "minor-issues",
			Status:   mas.StatusWarn,
			Severity: "low",
			Detail:   fmt.Sprintf("%d minor accessibility issues", r.Stats.Minor),
		})
	}

	return tasks
}

// toContentBlocks generates rich content blocks for the report.
func (r *Result) toContentBlocks() []mas.ContentBlock {
	blocks := make([]mas.ContentBlock, 0)

	// Summary metrics
	blocks = append(blocks, mas.ContentBlock{
		Type:  mas.ContentBlockKVPairs,
		Title: "Audit Summary",
		Pairs: []mas.KVPair{
			{Key: "URL", Value: r.URL},
			{Key: "WCAG Version", Value: r.Version},
			{Key: "Target Level", Value: r.Level},
			{Key: "Score", Value: fmt.Sprintf("%d/100", r.Score)},
			{Key: "Status", Value: r.conformanceVerdict(), Icon: r.conformanceIcon()},
			{Key: "Total Findings", Value: fmt.Sprintf("%d", r.Stats.TotalFindings)},
			{Key: "Pages Audited", Value: fmt.Sprintf("%d", r.Stats.TotalPages)},
		},
	})

	// Severity breakdown
	if r.Stats.TotalFindings > 0 {
		blocks = append(blocks, mas.ContentBlock{
			Type:  mas.ContentBlockKVPairs,
			Title: "Issues by Severity",
			Pairs: []mas.KVPair{
				{Key: "Critical", Value: fmt.Sprintf("%d", r.Stats.Critical), Icon: severityIcon("critical")},
				{Key: "Serious", Value: fmt.Sprintf("%d", r.Stats.Serious), Icon: severityIcon("serious")},
				{Key: "Moderate", Value: fmt.Sprintf("%d", r.Stats.Moderate), Icon: severityIcon("moderate")},
				{Key: "Minor", Value: fmt.Sprintf("%d", r.Stats.Minor), Icon: severityIcon("minor")},
			},
		})
	}

	// Top findings list (limit to 10)
	if len(r.Findings) > 0 {
		items := make([]mas.ListItem, 0, 10)
		for i, f := range r.Findings {
			if i >= 10 {
				items = append(items, mas.ListItem{
					Text: fmt.Sprintf("... and %d more findings", len(r.Findings)-10),
					Icon: "...",
				})
				break
			}
			items = append(items, mas.ListItem{
				Text:   fmt.Sprintf("[%s] %s: %s", strings.ToUpper(f.Impact), f.RuleID, f.Description),
				Icon:   severityIcon(f.Impact),
				Status: impactToStatus(f.Impact),
			})
		}
		blocks = append(blocks, mas.ContentBlock{
			Type:  mas.ContentBlockList,
			Title: "Top Findings",
			Items: items,
		})
	}

	return blocks
}

// Narrative returns a NarrativeSection for prose reports.
func (r *Result) Narrative() *mas.NarrativeSection {
	problem := fmt.Sprintf(
		"Accessibility audit of %s targeting WCAG %s Level %s conformance.",
		r.URL, r.Version, r.Level,
	)

	var analysisBuilder strings.Builder
	fmt.Fprintf(&analysisBuilder,
		"The audit identified %d accessibility issues across %d page(s). ",
		r.Stats.TotalFindings, r.Stats.TotalPages,
	)

	if r.Stats.TotalFindings > 0 {
		fmt.Fprintf(&analysisBuilder,
			"Severity breakdown: %d critical, %d serious, %d moderate, %d minor. ",
			r.Stats.Critical, r.Stats.Serious, r.Stats.Moderate, r.Stats.Minor,
		)
	}

	fmt.Fprintf(&analysisBuilder, "Overall conformance score: %d/100.", r.Score)

	var recommendation string
	if r.Conformant() {
		recommendation = "The site meets WCAG conformance requirements at the target level. Continue monitoring for accessibility as content changes."
	} else if r.Stats.Critical > 0 {
		recommendation = fmt.Sprintf(
			"CRITICAL: Address %d critical issues immediately. These represent barriers that prevent users from accessing content.",
			r.Stats.Critical,
		)
	} else if r.Stats.Serious > 0 {
		recommendation = fmt.Sprintf(
			"HIGH PRIORITY: Address %d serious issues. These significantly impact user experience for people with disabilities.",
			r.Stats.Serious,
		)
	} else {
		recommendation = fmt.Sprintf(
			"Address %d moderate/minor issues to improve accessibility. Consider prioritizing by impact and affected user groups.",
			r.Stats.Moderate+r.Stats.Minor,
		)
	}

	return &mas.NarrativeSection{
		Problem:        problem,
		Analysis:       analysisBuilder.String(),
		Recommendation: recommendation,
	}
}

// TeamSection returns a TeamSection for inclusion in a TeamReport.
func (r *Result) TeamSection() mas.TeamSection {
	ar := r.AgentResult()
	section := ar.ToTeamSection()
	section.Verdict = r.conformanceVerdict()
	section.Narrative = r.Narrative()
	return section
}

// Helper functions

func (r *Result) conformanceSeverity() string {
	if r.Conformant() {
		return "info"
	}
	if r.Stats.Critical > 0 {
		return "critical"
	}
	if r.Stats.Serious > 0 {
		return "high"
	}
	if r.Stats.Moderate > 0 {
		return "medium"
	}
	return "low"
}

func (r *Result) conformanceVerdict() string {
	if r.Conformant() {
		return "CONFORMANT"
	}
	if r.Stats.Critical > 0 {
		return "NON_CONFORMANT_CRITICAL"
	}
	if r.Stats.Serious > 0 {
		return "NON_CONFORMANT_SERIOUS"
	}
	return "NON_CONFORMANT"
}

func (r *Result) conformanceIcon() string {
	if r.Conformant() {
		return "\U0001F7E2" // 🟢
	}
	if r.Stats.Critical > 0 || r.Stats.Serious > 0 {
		return "\U0001F534" // 🔴
	}
	return "\U0001F7E1" // 🟡
}

func severityFromCount(count int) string {
	if count == 0 {
		return "info"
	}
	if count >= 10 {
		return "critical"
	}
	if count >= 5 {
		return "high"
	}
	if count >= 2 {
		return "medium"
	}
	return "low"
}

func severityIcon(impact string) string {
	switch strings.ToLower(impact) {
	case "critical":
		return "\U0001F534" // 🔴
	case "serious", "high":
		return "\U0001F7E0" // 🟠
	case "moderate", "medium":
		return "\U0001F7E1" // 🟡
	case "minor", "low":
		return "\U0001F535" // 🔵
	default:
		return "\u26AA" // ⚪
	}
}

func impactToStatus(impact string) mas.Status {
	switch strings.ToLower(impact) {
	case "critical", "serious":
		return mas.StatusNoGo
	case "moderate", "minor":
		return mas.StatusWarn
	default:
		return mas.StatusGo
	}
}
