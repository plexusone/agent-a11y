// Package report provides report generation for accessibility audits.
package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/agent-a11y/llm"
	"github.com/plexusone/agent-a11y/wcag"
)

// Format represents an output format.
type Format string

const (
	FormatJSON     Format = "json"
	FormatHTML     Format = "html"
	FormatMarkdown Format = "markdown"
	FormatCSV      Format = "csv"
	FormatVPAT     Format = "vpat"
	FormatWCAG     Format = "wcag"
	FormatOpenACR  Format = "openacr"
)

// Writer writes audit results in various formats.
type Writer struct {
	format Format
}

// NewWriter creates a new report writer.
func NewWriter(format Format) *Writer {
	return &Writer{format: format}
}

// Write writes the audit result to the given writer.
func (w *Writer) Write(out io.Writer, result *audit.AuditResult) error {
	switch w.format {
	case FormatJSON:
		return w.writeJSON(out, result)
	case FormatHTML:
		return w.writeHTML(out, result)
	case FormatMarkdown:
		return w.writeMarkdown(out, result)
	case FormatCSV:
		return w.writeCSV(out, result)
	case FormatVPAT:
		return w.writeVPAT(out, result)
	case FormatWCAG:
		return w.writeWCAGReport(out, result)
	case FormatOpenACR:
		return w.writeOpenACR(out, result)
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}
}

func (w *Writer) writeJSON(out io.Writer, result *audit.AuditResult) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func (w *Writer) writeCSV(out io.Writer, result *audit.AuditResult) error {
	csvWriter := csv.NewWriter(out)
	defer csvWriter.Flush()

	// Header row
	header := []string{
		"Page URL",
		"Rule ID",
		"Description",
		"Level",
		"Impact",
		"Success Criteria",
		"Selector",
		"HTML",
		"Help",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Data rows - one per finding
	for _, page := range result.Pages {
		for _, finding := range page.Findings {
			row := []string{
				page.URL,
				finding.RuleID,
				finding.Description,
				string(finding.Level),
				string(finding.Impact),
				strings.Join(finding.SuccessCriteria, "; "),
				finding.Selector,
				truncate(finding.HTML, 200),
				finding.Help,
			}
			if err := csvWriter.Write(row); err != nil {
				return err
			}
		}
	}

	return csvWriter.Error()
}

func (w *Writer) writeMarkdown(out io.Writer, result *audit.AuditResult) error {
	var sb strings.Builder

	// Header
	sb.WriteString("# Accessibility Audit Report\n\n")
	fmt.Fprintf(&sb, "**URL:** %s\n", result.TargetURL)
	fmt.Fprintf(&sb, "**Date:** %s\n", result.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&sb, "**Duration:** %dms\n", result.Duration)
	fmt.Fprintf(&sb, "**WCAG Version:** %s Level %s\n\n", result.WCAGVersion, result.WCAGLevel)

	// Summary
	sb.WriteString("## Summary\n\n")
	fmt.Fprintf(&sb, "- **Total Pages:** %d\n", result.Stats.TotalPages)
	fmt.Fprintf(&sb, "- **Total Findings:** %d\n", result.Stats.TotalFindings)
	fmt.Fprintf(&sb, "  - Critical: %d\n", result.Stats.Critical)
	fmt.Fprintf(&sb, "  - Serious: %d\n", result.Stats.Serious)
	fmt.Fprintf(&sb, "  - Moderate: %d\n", result.Stats.Moderate)
	fmt.Fprintf(&sb, "  - Minor: %d\n", result.Stats.Minor)
	sb.WriteString("\n")

	// Conformance
	sb.WriteString("## Conformance Status\n\n")
	sb.WriteString("| Level | Status | Issues |\n")
	sb.WriteString("|-------|--------|--------|\n")
	fmt.Fprintf(&sb, "| A | %s | %d |\n", result.Conformance.LevelA.Status, result.Conformance.LevelA.TotalIssues)
	fmt.Fprintf(&sb, "| AA | %s | %d |\n", result.Conformance.LevelAA.Status, result.Conformance.LevelAA.TotalIssues)
	fmt.Fprintf(&sb, "| AAA | %s | %d |\n", result.Conformance.LevelAAA.Status, result.Conformance.LevelAAA.TotalIssues)
	sb.WriteString("\n")

	// Findings by page
	sb.WriteString("## Findings\n\n")
	for _, page := range result.Pages {
		fmt.Fprintf(&sb, "### %s\n\n", page.URL)
		fmt.Fprintf(&sb, "**Title:** %s\n", page.Title)
		fmt.Fprintf(&sb, "**Findings:** %d\n\n", len(page.Findings))

		if len(page.Findings) > 0 {
			sb.WriteString("| Rule | Description | Level | Impact |\n")
			sb.WriteString("|------|-------------|-------|--------|\n")
			for _, finding := range page.Findings {
				fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n",
					finding.RuleID,
					truncate(finding.Description, 50),
					finding.Level,
					finding.Impact,
				)
			}
			sb.WriteString("\n")
		}
	}

	_, err := out.Write([]byte(sb.String()))
	return err
}

func (w *Writer) writeHTML(out io.Writer, result *audit.AuditResult) error {
	tmpl := template.Must(template.New("report").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"severityClass": func(i audit.Impact) string {
			switch i {
			case audit.ImpactBlocker, audit.ImpactCritical:
				return "critical"
			case audit.ImpactSerious:
				return "serious"
			case audit.ImpactModerate:
				return "moderate"
			default:
				return "minor"
			}
		},
	}).Parse(htmlTemplate))

	return tmpl.Execute(out, result)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Accessibility Audit Report</title>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; max-width: 1200px; margin: 0 auto; padding: 20px; }
        h1, h2, h3 { color: #333; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .stat-card { background: #f5f5f5; padding: 20px; border-radius: 8px; }
        .stat-value { font-size: 2em; font-weight: bold; }
        .critical { color: #d32f2f; }
        .serious { color: #f57c00; }
        .moderate { color: #fbc02d; }
        .minor { color: #388e3c; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f5f5f5; }
        .finding { background: #fff; border: 1px solid #ddd; border-radius: 8px; padding: 15px; margin: 10px 0; }
        .finding-header { display: flex; justify-content: space-between; margin-bottom: 10px; }
        .badge { padding: 4px 8px; border-radius: 4px; font-size: 0.85em; }
        .badge.critical { background: #ffebee; color: #c62828; }
        .badge.serious { background: #fff3e0; color: #e65100; }
        .badge.moderate { background: #fffde7; color: #f9a825; }
        .badge.minor { background: #e8f5e9; color: #2e7d32; }
        code { background: #f5f5f5; padding: 2px 6px; border-radius: 4px; }
        pre { background: #f5f5f5; padding: 15px; border-radius: 8px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>Accessibility Audit Report</h1>

    <div class="meta">
        <p><strong>URL:</strong> {{.TargetURL}}</p>
        <p><strong>Date:</strong> {{formatTime .StartTime}}</p>
        <p><strong>WCAG:</strong> {{.WCAGVersion}} Level {{.WCAGLevel}}</p>
    </div>

    <h2>Summary</h2>
    <div class="summary">
        <div class="stat-card">
            <div class="stat-value">{{.Stats.TotalPages}}</div>
            <div>Pages Audited</div>
        </div>
        <div class="stat-card">
            <div class="stat-value">{{.Stats.TotalFindings}}</div>
            <div>Total Issues</div>
        </div>
        <div class="stat-card">
            <div class="stat-value critical">{{.Stats.Critical}}</div>
            <div>Critical</div>
        </div>
        <div class="stat-card">
            <div class="stat-value serious">{{.Stats.Serious}}</div>
            <div>Serious</div>
        </div>
    </div>

    <h2>Conformance</h2>
    <table>
        <thead>
            <tr><th>Level</th><th>Status</th><th>Issues</th></tr>
        </thead>
        <tbody>
            <tr><td>Level A</td><td>{{.Conformance.LevelA.Status}}</td><td>{{.Conformance.LevelA.TotalIssues}}</td></tr>
            <tr><td>Level AA</td><td>{{.Conformance.LevelAA.Status}}</td><td>{{.Conformance.LevelAA.TotalIssues}}</td></tr>
            <tr><td>Level AAA</td><td>{{.Conformance.LevelAAA.Status}}</td><td>{{.Conformance.LevelAAA.TotalIssues}}</td></tr>
        </tbody>
    </table>

    <h2>Findings</h2>
    {{range .Pages}}
    <h3>{{.URL}}</h3>
    <p><strong>Title:</strong> {{.Title}}</p>
    {{range .Findings}}
    <div class="finding">
        <div class="finding-header">
            <strong>{{.RuleID}}</strong>
            <span class="badge {{severityClass .Impact}}">{{.Impact}}</span>
        </div>
        <p>{{.Description}}</p>
        {{if .Selector}}<p><strong>Selector:</strong> <code>{{.Selector}}</code></p>{{end}}
        {{if .HTML}}<pre><code>{{.HTML}}</code></pre>{{end}}
        {{if .Help}}<p><strong>Help:</strong> {{.Help}}</p>{{end}}
    </div>
    {{end}}
    {{end}}

    <footer>
        <p>Generated by a11y-audit-service</p>
    </footer>
</body>
</html>`

func (w *Writer) writeVPAT(out io.Writer, result *audit.AuditResult) error {
	report := generateVPATReport(result)

	var sb strings.Builder

	// VPAT Header
	sb.WriteString("# VPAT 2.4 Conformance Report\n\n")
	fmt.Fprintf(&sb, "**Product Name:** %s\n", report.ProductName)
	fmt.Fprintf(&sb, "**Report Date:** %s\n", report.ReportDate.Format("2006-01-02"))
	fmt.Fprintf(&sb, "**Evaluation URL:** %s\n\n", report.EvaluationURL)

	// Evaluation Methods
	sb.WriteString("## Evaluation Methods\n\n")
	for _, method := range report.EvaluationMethods {
		fmt.Fprintf(&sb, "- %s\n", method)
	}
	sb.WriteString("\n")

	// Conformance summary — surfaces scope (esp. how much was not evaluated).
	counts := map[string]int{}
	for _, c := range report.WCAGConformance {
		counts[c.Conformance]++
	}
	sb.WriteString("## Conformance Summary\n\n")
	for _, level := range []string{"Supports", "Partially Supports", "Does Not Support", "Not Evaluated"} {
		if n := counts[level]; n > 0 {
			fmt.Fprintf(&sb, "- **%s:** %d criteria\n", level, n)
		}
	}
	sb.WriteString("\n")

	// WCAG Conformance Table
	sb.WriteString("## WCAG 2.2 Conformance\n\n")
	sb.WriteString("| Criterion | Level | Conformance | Remarks |\n")
	sb.WriteString("|-----------|-------|-------------|----------|\n")

	for _, criterion := range report.WCAGConformance {
		fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n",
			criterion.Criterion,
			criterion.Level,
			criterion.Conformance,
			truncate(criterion.Remarks, 120),
		)
	}

	// Legal Disclaimer
	sb.WriteString("\n## Legal Disclaimer\n\n")
	sb.WriteString(report.LegalDisclaimer)
	sb.WriteString("\n")

	_, err := out.Write([]byte(sb.String()))
	return err
}

func (w *Writer) writeWCAGReport(out io.Writer, result *audit.AuditResult) error {
	var sb strings.Builder

	sb.WriteString("# WCAG Conformance Report\n\n")
	fmt.Fprintf(&sb, "**Target:** %s\n", result.TargetURL)
	fmt.Fprintf(&sb, "**Standard:** WCAG %s Level %s\n", result.WCAGVersion, result.WCAGLevel)
	fmt.Fprintf(&sb, "**Date:** %s\n\n", result.StartTime.Format("2006-01-02"))

	// Per-criterion breakdown
	sb.WriteString("## Success Criteria Results\n\n")
	sb.WriteString("| ID | Name | Level | Status | Issues |\n")
	sb.WriteString("|----|------|-------|--------|--------|\n")

	for _, criterion := range result.Conformance.Criteria {
		fmt.Fprintf(&sb, "| %s | %s | %s | %s | %d |\n",
			criterion.ID,
			criterion.Name,
			criterion.Level,
			criterion.Status,
			criterion.IssueCount,
		)
	}

	_, err := out.Write([]byte(sb.String()))
	return err
}

// CriterionResult is the structured, honest per-criterion verdict for an audit.
// It is the shared source both the VPAT renderer and external consumers (e.g.
// the canonical AssessmentRecord) project from.
type CriterionResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Level       string `json:"level"`
	Conformance string `json:"conformance"`
	Method      string `json:"method"`
	Evaluated   bool   `json:"evaluated"`
	Remarks     string `json:"remarks"`
	IssueCount  int    `json:"issueCount"`
}

// EvaluationMethods returns the human-readable evaluation methods applied,
// reflecting whether the LLM judge ran.
func EvaluationMethods(result *audit.AuditResult) []string {
	methods := []string{
		"Automated accessibility testing (axe-core)",
		"Specialized automated testing (keyboard, focus, reflow)",
	}
	if result.LLMEnabled {
		methods = append(methods, "AI-assisted evaluation (LLM-as-a-Judge)")
	}
	return methods
}

// ConformanceDisclaimer is the standard honesty disclaimer for a11y attestations.
const ConformanceDisclaimer = "This report reflects an automated accessibility evaluation. Criteria " +
	"marked \"Not Evaluated\" were not assessed because they require human " +
	"judgment (or AI evaluation that was not enabled) and cannot be verified " +
	"by automated testing alone; they require manual review by a qualified " +
	"evaluator before conformance can be claimed. A \"Supports\" result reflects " +
	"that no issues were detected by the methods listed above, not a guarantee " +
	"of full conformance."

// EvaluateConformance produces the honest per-criterion verdicts for an audit,
// keyed off the WCAG catalog and each criterion's evaluation method.
func EvaluateConformance(result *audit.AuditResult) []CriterionResult {
	// Aggregate findings by success criterion.
	criteriaStatus := make(map[string]string)
	criteriaIssues := make(map[string]int)
	for _, page := range result.Pages {
		for _, finding := range page.Findings {
			for _, sc := range finding.SuccessCriteria {
				criteriaIssues[sc]++
				if finding.Impact == audit.ImpactBlocker || finding.Impact == audit.ImpactCritical {
					criteriaStatus[sc] = string(wcag.ConformanceDoesNotSupport)
				} else if criteriaStatus[sc] != string(wcag.ConformanceDoesNotSupport) {
					criteriaStatus[sc] = string(wcag.ConformancePartiallySupports)
				}
			}
		}
	}

	// The catalog is A/AA; include AA only when the audit targeted AA or above.
	includeAA := result.WCAGLevel == audit.WCAGLevelAA || result.WCAGLevel == audit.WCAGLevelAAA

	var out []CriterionResult
	for _, c := range wcag.WCAG22AA {
		if c.Level == "AA" && !includeAA {
			continue
		}

		var status, remarks string
		if s, ok := criteriaStatus[c.ID]; ok {
			status = s
			remarks = fmt.Sprintf("%d issue(s) detected by automated testing", criteriaIssues[c.ID])
		} else {
			status, remarks = passStatusForMethod(c.Method)
		}

		out = append(out, CriterionResult{
			ID:          c.ID,
			Name:        c.Name,
			Level:       c.Level,
			Conformance: status,
			Method:      string(c.Method),
			Evaluated:   status != string(wcag.ConformanceNotEvaluated),
			Remarks:     remarks,
			IssueCount:  criteriaIssues[c.ID],
		})
	}
	return out
}

// CriterionQueriesForUnevaluated builds proactive-evaluation queries for the
// criteria that automation left "Not Evaluated" and whose method is designed
// for judgment (LLM-Judge or Hybrid). These feed a CriterionEvaluator (API or
// local bundle).
func CriterionQueriesForUnevaluated(result *audit.AuditResult) []llm.CriterionQuery {
	var qs []llm.CriterionQuery
	for _, cr := range EvaluateConformance(result) {
		if cr.Conformance != string(wcag.ConformanceNotEvaluated) {
			continue
		}
		c := wcag.GetCriterion(cr.ID)
		if c == nil || (c.Method != wcag.MethodLLMJudge && c.Method != wcag.MethodHybrid) {
			continue
		}
		qs = append(qs, llm.CriterionQuery{ID: c.ID, Name: c.Name, Level: c.Level, Requirement: c.Description})
	}
	return qs
}

// ApplyVerdicts overlays proactive verdicts onto a base conformance evaluation.
// A verdict only affects criteria automation left "Not Evaluated"; it never
// overrides a finding-backed "Does Not Support"/"Partially Supports".
func ApplyVerdicts(base []CriterionResult, verdicts []llm.CriterionVerdict) []CriterionResult {
	byID := make(map[string]llm.CriterionVerdict, len(verdicts))
	for _, v := range verdicts {
		byID[v.ID] = v
	}
	out := make([]CriterionResult, len(base))
	copy(out, base)
	for i := range out {
		if out[i].Conformance != string(wcag.ConformanceNotEvaluated) {
			continue
		}
		v, ok := byID[out[i].ID]
		if !ok {
			continue
		}
		out[i].Conformance = v.Conformance
		out[i].Evaluated = v.Conformance != llm.VerdictNotEvaluated
		switch {
		case v.Conformance == llm.VerdictNotEvaluated && v.EvidenceGap != "":
			out[i].Remarks = "Not evaluated — needs: " + v.EvidenceGap
		case v.Reasoning != "":
			out[i].Remarks = "AI-judged: " + v.Reasoning
		}
	}
	return out
}

func generateVPATReport(result *audit.AuditResult) *audit.VPATReport {
	report := &audit.VPATReport{
		ProductName:       "Web Application",
		ReportDate:        result.StartTime,
		EvaluationURL:     result.TargetURL,
		EvaluationMethods: EvaluationMethods(result),
		LegalDisclaimer:   ConformanceDisclaimer,
	}
	for _, cr := range EvaluateConformance(result) {
		report.WCAGConformance = append(report.WCAGConformance, audit.VPATCriterion{
			Criterion:   fmt.Sprintf("%s %s", cr.ID, cr.Name),
			Level:       cr.Level,
			Conformance: cr.Conformance,
			Remarks:     cr.Remarks,
		})
	}
	return report
}

// passStatusForMethod returns the honest conformance and remark for a criterion
// that had no findings, based on how the criterion can be evaluated. We only
// claim "Supports" for criteria the automated methods can determine. Criteria
// requiring judgment (LLM-Judge, Hybrid, Manual) are "Not Evaluated" here and
// stay so until a proactive verdict resolves them — enabling an LLM alone does
// not flip them, because nothing has actually evaluated them yet.
func passStatusForMethod(method wcag.EvaluationMethod) (status, remarks string) {
	switch method {
	case wcag.MethodAutomated:
		return string(wcag.ConformanceSupports), "No issues detected by automated testing"
	case wcag.MethodSpecialized:
		return string(wcag.ConformanceSupports), "No issues detected by specialized automated testing"
	case wcag.MethodLLMJudge:
		return string(wcag.ConformanceNotEvaluated), "Requires proactive AI or manual evaluation"
	case wcag.MethodHybrid:
		return string(wcag.ConformanceNotEvaluated), "Automated checks passed; proactive AI or manual review required to confirm"
	case wcag.MethodManual:
		return string(wcag.ConformanceNotEvaluated), "Requires manual review by a qualified evaluator"
	default:
		return string(wcag.ConformanceNotEvaluated), "Not assessed by automated testing"
	}
}

type wcagCriterion struct {
	ID    string
	Name  string
	Level string
}

func getWCAG22Criteria() []wcagCriterion {
	return []wcagCriterion{
		{"1.1.1", "Non-text Content", "A"},
		{"1.2.1", "Audio-only and Video-only", "A"},
		{"1.2.2", "Captions", "A"},
		{"1.2.3", "Audio Description or Media Alternative", "A"},
		{"1.3.1", "Info and Relationships", "A"},
		{"1.3.2", "Meaningful Sequence", "A"},
		{"1.3.3", "Sensory Characteristics", "A"},
		{"1.4.1", "Use of Color", "A"},
		{"1.4.2", "Audio Control", "A"},
		{"1.4.3", "Contrast (Minimum)", "AA"},
		{"1.4.4", "Resize Text", "AA"},
		{"1.4.5", "Images of Text", "AA"},
		{"2.1.1", "Keyboard", "A"},
		{"2.1.2", "No Keyboard Trap", "A"},
		{"2.2.1", "Timing Adjustable", "A"},
		{"2.2.2", "Pause, Stop, Hide", "A"},
		{"2.3.1", "Three Flashes or Below Threshold", "A"},
		{"2.4.1", "Bypass Blocks", "A"},
		{"2.4.2", "Page Titled", "A"},
		{"2.4.3", "Focus Order", "A"},
		{"2.4.4", "Link Purpose (In Context)", "A"},
		{"2.4.5", "Multiple Ways", "AA"},
		{"2.4.6", "Headings and Labels", "AA"},
		{"2.4.7", "Focus Visible", "AA"},
		{"3.1.1", "Language of Page", "A"},
		{"3.1.2", "Language of Parts", "AA"},
		{"3.2.1", "On Focus", "A"},
		{"3.2.2", "On Input", "A"},
		{"3.2.3", "Consistent Navigation", "AA"},
		{"3.2.4", "Consistent Identification", "AA"},
		{"3.3.1", "Error Identification", "A"},
		{"3.3.2", "Labels or Instructions", "A"},
		{"3.3.3", "Error Suggestion", "AA"},
		{"3.3.4", "Error Prevention (Legal, Financial, Data)", "AA"},
		{"4.1.1", "Parsing", "A"},
		{"4.1.2", "Name, Role, Value", "A"},
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
