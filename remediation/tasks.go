package remediation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/agentplexus/agent-a11y/types"
)

// TaskPriority maps to Jira/GitHub priority levels.
type TaskPriority string

const (
	PriorityBlocker  TaskPriority = "Blocker"  // P0 - Prevents access
	PriorityCritical TaskPriority = "Critical" // P1 - Serious barrier
	PriorityMajor    TaskPriority = "Major"    // P2 - Moderate barrier
	PriorityMinor    TaskPriority = "Minor"    // P3 - Inconvenience
	PriorityTrivial  TaskPriority = "Trivial"  // P4 - Enhancement
)

// TaskStatus represents remediation task status.
type TaskStatus string

const (
	StatusOpen       TaskStatus = "Open"
	StatusInProgress TaskStatus = "In Progress"
	StatusReview     TaskStatus = "In Review"
	StatusDone       TaskStatus = "Done"
)

// RemediationTask represents a single remediation work item that can be
// converted to a Jira ticket, GitHub issue, or LLM agent prompt.
type RemediationTask struct {
	// Unique identifier (hash of rule + criterion + component pattern)
	ID string `json:"id"`

	// Task metadata
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Priority    TaskPriority `json:"priority"`
	Labels      []string     `json:"labels"`
	Component   string       `json:"component,omitempty"` // UI component or file path

	// WCAG references
	SuccessCriteria []string `json:"successCriteria"`
	WCAGLevel       string   `json:"wcagLevel"`
	RuleID          string   `json:"ruleId"`

	// Affected elements (for grouping)
	AffectedElements []AffectedElement `json:"affectedElements"`
	AffectedCount    int               `json:"affectedCount"`

	// Remediation guidance
	Summary          string             `json:"summary"`
	AcceptanceCriteria []string         `json:"acceptanceCriteria"`
	TechniqueRefs    []types.TechniqueRef `json:"techniqueRefs"`
	References       []types.ReferenceURL `json:"references"`

	// For LLM agents
	SuggestedFix     string   `json:"suggestedFix,omitempty"`
	CodeExamples     []string `json:"codeExamples,omitempty"`
	FilesToModify    []string `json:"filesToModify,omitempty"`

	// Estimation
	StoryPoints int    `json:"storyPoints,omitempty"`
	Complexity  string `json:"complexity,omitempty"` // low, medium, high
}

// AffectedElement represents a single element that needs remediation.
type AffectedElement struct {
	Selector  string `json:"selector"`
	HTML      string `json:"html,omitempty"`
	PageURL   string `json:"pageUrl"`
	PageTitle string `json:"pageTitle,omitempty"`
}

// JiraTicket represents a Jira-compatible issue structure.
type JiraTicket struct {
	Fields JiraFields `json:"fields"`
}

// JiraFields contains Jira issue fields.
type JiraFields struct {
	Project     JiraProject   `json:"project"`
	Summary     string        `json:"summary"`
	Description string        `json:"description"`
	IssueType   JiraIssueType `json:"issuetype"`
	Priority    JiraPriority  `json:"priority,omitempty"`
	Labels      []string      `json:"labels,omitempty"`
	Components  []JiraComponent `json:"components,omitempty"`

	// Custom fields (configurable)
	CustomFields map[string]interface{} `json:"-"`
}

type JiraProject struct {
	Key string `json:"key"`
}

type JiraIssueType struct {
	Name string `json:"name"`
}

type JiraPriority struct {
	Name string `json:"name"`
}

type JiraComponent struct {
	Name string `json:"name"`
}

// GitHubIssue represents a GitHub-compatible issue structure.
type GitHubIssue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
}

// LLMAgentPrompt contains structured data for an LLM agent to fix issues.
type LLMAgentPrompt struct {
	TaskID          string   `json:"taskId"`
	Objective       string   `json:"objective"`
	Context         string   `json:"context"`
	Requirements    []string `json:"requirements"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	FilesToModify   []string `json:"filesToModify"`
	CodeExamples    []string `json:"codeExamples"`
	References      []string `json:"references"`
	CommitMessage   string   `json:"commitMessage"`
	PRTitle         string   `json:"prTitle"`
	PRDescription   string   `json:"prDescription"`
}

// TaskBuilder creates remediation tasks from audit findings.
type TaskBuilder struct {
	// Grouping configuration
	GroupByRule      bool // Group findings by rule ID
	GroupByComponent bool // Group findings by component/file
	GroupByCriterion bool // Group findings by WCAG criterion

	// Deduplication
	DeduplicateSimilar bool // Combine similar selectors
}

// DefaultTaskBuilder returns a TaskBuilder with sensible defaults.
func DefaultTaskBuilder() *TaskBuilder {
	return &TaskBuilder{
		GroupByRule:        true,
		GroupByComponent:   false,
		GroupByCriterion:   false,
		DeduplicateSimilar: true,
	}
}

// BuildTasks creates remediation tasks from a list of findings.
func (tb *TaskBuilder) BuildTasks(findings []types.Finding) []RemediationTask {
	// Group findings
	groups := tb.groupFindings(findings)

	// Convert groups to tasks
	var tasks []RemediationTask
	for _, group := range groups {
		task := tb.createTask(group)
		tasks = append(tasks, task)
	}

	// Sort by priority
	sort.Slice(tasks, func(i, j int) bool {
		return priorityOrder(tasks[i].Priority) < priorityOrder(tasks[j].Priority)
	})

	return tasks
}

func (tb *TaskBuilder) groupFindings(findings []types.Finding) map[string][]types.Finding {
	groups := make(map[string][]types.Finding)

	for _, f := range findings {
		key := tb.groupKey(f)
		groups[key] = append(groups[key], f)
	}

	return groups
}

func (tb *TaskBuilder) groupKey(f types.Finding) string {
	parts := []string{}

	if tb.GroupByRule {
		parts = append(parts, f.RuleID)
	}
	if tb.GroupByCriterion && len(f.SuccessCriteria) > 0 {
		parts = append(parts, f.SuccessCriteria[0])
	}
	if tb.GroupByComponent && f.Component != "" {
		parts = append(parts, f.Component)
	}

	if len(parts) == 0 {
		parts = append(parts, f.RuleID)
	}

	return strings.Join(parts, ":")
}

func (tb *TaskBuilder) createTask(findings []types.Finding) RemediationTask {
	if len(findings) == 0 {
		return RemediationTask{}
	}

	// Use first finding as representative
	first := findings[0]

	// Collect affected elements
	var affected []AffectedElement
	selectors := make(map[string]bool)

	for _, f := range findings {
		// Deduplicate by selector if enabled
		if tb.DeduplicateSimilar && selectors[f.Selector] {
			continue
		}
		selectors[f.Selector] = true

		affected = append(affected, AffectedElement{
			Selector:  f.Selector,
			HTML:      f.HTML,
			PageURL:   f.PageURL,
			PageTitle: f.PageTitle,
		})
	}

	// Build remediation info
	var techRefs []types.TechniqueRef
	var refs []types.ReferenceURL
	summary := ""

	if first.Remediation != nil {
		techRefs = first.Remediation.Techniques
		refs = first.Remediation.References
		summary = first.Remediation.Summary
	} else if len(first.SuccessCriteria) > 0 {
		// Build remediation on the fly
		rem := BuildForFinding(first.SuccessCriteria, first.RuleID)
		if rem != nil {
			techRefs = rem.Techniques
			refs = rem.References
			summary = rem.Summary
		}
	}

	// Generate task ID
	taskID := generateTaskID(first.RuleID, first.SuccessCriteria, first.Component)

	// Build title
	title := buildTitle(first.RuleID, first.Description, len(affected))

	// Build description
	description := buildDescription(first, affected, summary)

	// Build acceptance criteria
	acceptanceCriteria := buildAcceptanceCriteria(first, techRefs)

	// Determine priority
	priority := mapPriority(first.Impact, first.Severity)

	// Build labels
	labels := buildLabels(first)

	// Estimate complexity
	complexity, storyPoints := estimateComplexity(len(affected), first.RuleID)

	return RemediationTask{
		ID:                 taskID,
		Title:              title,
		Description:        description,
		Priority:           priority,
		Labels:             labels,
		Component:          first.Component,
		SuccessCriteria:    first.SuccessCriteria,
		WCAGLevel:          string(first.Level),
		RuleID:             first.RuleID,
		AffectedElements:   affected,
		AffectedCount:      len(affected),
		Summary:            summary,
		AcceptanceCriteria: acceptanceCriteria,
		TechniqueRefs:      techRefs,
		References:         refs,
		SuggestedFix:       buildSuggestedFix(first),
		CodeExamples:       buildCodeExamples(first.RuleID),
		Complexity:         complexity,
		StoryPoints:        storyPoints,
	}
}

// ToJiraTicket converts a task to Jira API format.
func (t *RemediationTask) ToJiraTicket(projectKey, issueType string) JiraTicket {
	description := t.formatJiraDescription()

	return JiraTicket{
		Fields: JiraFields{
			Project:   JiraProject{Key: projectKey},
			Summary:   t.Title,
			Description: description,
			IssueType: JiraIssueType{Name: issueType},
			Priority:  JiraPriority{Name: string(t.Priority)},
			Labels:    t.Labels,
		},
	}
}

func (t *RemediationTask) formatJiraDescription() string {
	var sb strings.Builder

	sb.WriteString("h2. Summary\n")
	sb.WriteString(t.Summary)
	sb.WriteString("\n\n")

	sb.WriteString("h2. WCAG Criteria\n")
	for _, sc := range t.SuccessCriteria {
		sb.WriteString(fmt.Sprintf("* %s (Level %s)\n", sc, t.WCAGLevel))
	}
	sb.WriteString("\n")

	sb.WriteString("h2. Affected Elements\n")
	sb.WriteString(fmt.Sprintf("*%d elements affected*\n\n", t.AffectedCount))
	for i, el := range t.AffectedElements {
		if i >= 5 {
			sb.WriteString(fmt.Sprintf("... and %d more\n", t.AffectedCount-5))
			break
		}
		sb.WriteString(fmt.Sprintf("* {{%s}} on [%s|%s]\n", el.Selector, el.PageTitle, el.PageURL))
	}
	sb.WriteString("\n")

	sb.WriteString("h2. Acceptance Criteria\n")
	for _, ac := range t.AcceptanceCriteria {
		sb.WriteString(fmt.Sprintf("* %s\n", ac))
	}
	sb.WriteString("\n")

	sb.WriteString("h2. References\n")
	for _, ref := range t.References {
		sb.WriteString(fmt.Sprintf("* [%s|%s]\n", ref.Title, ref.URL))
	}

	if t.SuggestedFix != "" {
		sb.WriteString("\nh2. Suggested Fix\n")
		sb.WriteString("{code}\n")
		sb.WriteString(t.SuggestedFix)
		sb.WriteString("\n{code}\n")
	}

	return sb.String()
}

// ToGitHubIssue converts a task to GitHub API format.
func (t *RemediationTask) ToGitHubIssue() GitHubIssue {
	body := t.formatGitHubBody()

	return GitHubIssue{
		Title:  t.Title,
		Body:   body,
		Labels: t.Labels,
	}
}

func (t *RemediationTask) formatGitHubBody() string {
	var sb strings.Builder

	sb.WriteString("## Summary\n\n")
	sb.WriteString(t.Summary)
	sb.WriteString("\n\n")

	sb.WriteString("## WCAG Criteria\n\n")
	for _, sc := range t.SuccessCriteria {
		sb.WriteString(fmt.Sprintf("- **%s** (Level %s)\n", sc, t.WCAGLevel))
	}
	sb.WriteString("\n")

	sb.WriteString("## Affected Elements\n\n")
	sb.WriteString(fmt.Sprintf("**%d elements affected**\n\n", t.AffectedCount))
	for i, el := range t.AffectedElements {
		if i >= 5 {
			sb.WriteString(fmt.Sprintf("\n... and %d more\n", t.AffectedCount-5))
			break
		}
		sb.WriteString(fmt.Sprintf("- `%s` on [%s](%s)\n", el.Selector, el.PageTitle, el.PageURL))
	}
	sb.WriteString("\n")

	sb.WriteString("## Acceptance Criteria\n\n")
	for _, ac := range t.AcceptanceCriteria {
		sb.WriteString(fmt.Sprintf("- [ ] %s\n", ac))
	}
	sb.WriteString("\n")

	sb.WriteString("## References\n\n")
	for _, ref := range t.References {
		sb.WriteString(fmt.Sprintf("- [%s](%s)\n", ref.Title, ref.URL))
	}

	if t.SuggestedFix != "" {
		sb.WriteString("\n## Suggested Fix\n\n")
		sb.WriteString("```html\n")
		sb.WriteString(t.SuggestedFix)
		sb.WriteString("\n```\n")
	}

	return sb.String()
}

// ToLLMAgentPrompt converts a task to an LLM agent prompt.
func (t *RemediationTask) ToLLMAgentPrompt() LLMAgentPrompt {
	// Build requirements
	var requirements []string
	requirements = append(requirements, t.Summary)
	for _, tech := range t.TechniqueRefs {
		if tech.Type == "sufficient" {
			requirements = append(requirements, fmt.Sprintf("Apply technique %s: %s", tech.ID, tech.Title))
		}
	}

	// Build reference URLs
	var refURLs []string
	for _, ref := range t.References {
		refURLs = append(refURLs, ref.URL)
	}

	// Build context
	var contextParts []string
	contextParts = append(contextParts, fmt.Sprintf("Rule: %s", t.RuleID))
	contextParts = append(contextParts, fmt.Sprintf("WCAG: %s (Level %s)", strings.Join(t.SuccessCriteria, ", "), t.WCAGLevel))
	contextParts = append(contextParts, fmt.Sprintf("Affected: %d elements", t.AffectedCount))

	// Build commit message
	commitMsg := fmt.Sprintf("fix(a11y): %s\n\nFixes WCAG %s compliance issue.\nAffected elements: %d\n\nCloses #<issue>",
		t.ruleToCommitSubject(),
		strings.Join(t.SuccessCriteria, ", "),
		t.AffectedCount)

	return LLMAgentPrompt{
		TaskID:             t.ID,
		Objective:          t.Title,
		Context:            strings.Join(contextParts, "\n"),
		Requirements:       requirements,
		AcceptanceCriteria: t.AcceptanceCriteria,
		FilesToModify:      t.FilesToModify,
		CodeExamples:       t.CodeExamples,
		References:         refURLs,
		CommitMessage:      commitMsg,
		PRTitle:            fmt.Sprintf("fix(a11y): %s", t.ruleToCommitSubject()),
		PRDescription:      t.formatGitHubBody(),
	}
}

func (t *RemediationTask) ruleToCommitSubject() string {
	subjects := map[string]string{
		"image-alt":              "add missing alt text to images",
		"button-name":            "add accessible names to buttons",
		"link-name":              "add accessible names to links",
		"label":                  "add labels to form controls",
		"color-contrast":         "improve color contrast",
		"focus-visible":          "add visible focus indicators",
		"keyboard-trap":          "fix keyboard trap",
		"keyboard-unreachable":   "make elements keyboard accessible",
		"tabindex-positive":      "remove positive tabindex values",
		"reflow-horizontal-scroll": "fix horizontal scroll at narrow widths",
		"target-size-minimum":    "increase touch target size",
		"text-spacing-loss":      "fix text spacing overflow",
		"focus-order":            "fix focus order",
		"focus-obscured":         "prevent focus obscuration",
	}

	if subject, ok := subjects[t.RuleID]; ok {
		return subject
	}
	return fmt.Sprintf("fix %s accessibility issue", t.RuleID)
}

// Helper functions

func generateTaskID(ruleID string, criteria []string, component string) string {
	data := ruleID + strings.Join(criteria, ",") + component
	hash := sha256.Sum256([]byte(data))
	return "a11y-" + hex.EncodeToString(hash[:])[:8]
}

func buildTitle(ruleID, description string, count int) string {
	titles := map[string]string{
		"image-alt":              "Add alt text to images",
		"button-name":            "Add accessible names to buttons",
		"link-name":              "Add accessible names to links",
		"label":                  "Add labels to form controls",
		"color-contrast":         "Fix color contrast issues",
		"focus-visible":          "Add visible focus indicators",
		"keyboard-trap":          "Fix keyboard trap",
		"keyboard-unreachable":   "Make elements keyboard accessible",
		"tabindex-positive":      "Remove positive tabindex values",
		"reflow-horizontal-scroll": "Fix horizontal scrolling at 320px",
		"target-size-minimum":    "Increase touch target size to 24x24px",
		"text-spacing-loss":      "Fix text clipping with increased spacing",
		"focus-order":            "Fix focus order issues",
		"focus-obscured":         "Prevent focused elements from being obscured",
	}

	title := titles[ruleID]
	if title == "" {
		// Use description, truncated
		title = description
		if len(title) > 60 {
			title = title[:57] + "..."
		}
	}

	if count > 1 {
		title = fmt.Sprintf("%s (%d instances)", title, count)
	}

	return title
}

func buildDescription(f types.Finding, affected []AffectedElement, summary string) string {
	var sb strings.Builder

	sb.WriteString(f.Description)
	sb.WriteString("\n\n")
	sb.WriteString("**Remediation:** ")
	sb.WriteString(summary)

	return sb.String()
}

func buildAcceptanceCriteria(f types.Finding, techs []types.TechniqueRef) []string {
	criteria := []string{}

	// Generic criteria based on rule
	acMap := map[string][]string{
		"image-alt": {
			"All img elements have descriptive alt attributes",
			"Decorative images have empty alt=\"\"",
			"Alt text conveys the same information as the image",
		},
		"button-name": {
			"All buttons have accessible names",
			"Button names describe the action",
			"Screen readers announce button purpose",
		},
		"link-name": {
			"All links have accessible names",
			"Link text describes the destination",
			"No generic 'click here' or 'read more' links",
		},
		"color-contrast": {
			"Text contrast ratio is at least 4.5:1",
			"Large text (18pt+) contrast ratio is at least 3:1",
			"Contrast verified with browser dev tools",
		},
		"focus-visible": {
			"All interactive elements have visible focus indicator",
			"Focus indicator has sufficient contrast",
			"Custom focus styles don't remove default outline",
		},
		"keyboard-trap": {
			"Users can tab through all focusable elements",
			"Focus can exit all components using keyboard",
			"No infinite focus loops exist",
		},
	}

	if ac, ok := acMap[f.RuleID]; ok {
		criteria = append(criteria, ac...)
	} else {
		criteria = append(criteria, fmt.Sprintf("Issue resolved for all %d affected elements", len(f.SuccessCriteria)))
		criteria = append(criteria, "Verified with accessibility testing tools")
		criteria = append(criteria, "Screen reader announces content correctly")
	}

	return criteria
}

func buildLabels(f types.Finding) []string {
	labels := []string{"accessibility", "a11y"}

	// Add WCAG level
	labels = append(labels, fmt.Sprintf("wcag-%s", strings.ToLower(string(f.Level))))

	// Add criteria
	for _, sc := range f.SuccessCriteria {
		labels = append(labels, fmt.Sprintf("wcag-%s", strings.ReplaceAll(sc, ".", "-")))
	}

	// Add severity
	if f.Severity != "" {
		labels = append(labels, string(f.Severity))
	}

	return labels
}

func mapPriority(impact types.Impact, severity types.Severity) TaskPriority {
	switch impact {
	case types.ImpactBlocker:
		return PriorityBlocker
	case types.ImpactCritical:
		return PriorityCritical
	case types.ImpactSerious:
		return PriorityMajor
	case types.ImpactModerate:
		return PriorityMinor
	default:
		return PriorityTrivial
	}
}

func priorityOrder(p TaskPriority) int {
	order := map[TaskPriority]int{
		PriorityBlocker:  0,
		PriorityCritical: 1,
		PriorityMajor:    2,
		PriorityMinor:    3,
		PriorityTrivial:  4,
	}
	return order[p]
}

func estimateComplexity(elementCount int, ruleID string) (string, int) {
	// Simple heuristics
	complexRules := map[string]bool{
		"color-contrast":         true,
		"reflow-horizontal-scroll": true,
		"keyboard-trap":          true,
	}

	if complexRules[ruleID] {
		if elementCount > 10 {
			return "high", 8
		}
		return "medium", 5
	}

	if elementCount > 20 {
		return "medium", 3
	}
	if elementCount > 5 {
		return "low", 2
	}
	return "low", 1
}

func buildSuggestedFix(f types.Finding) string {
	fixes := map[string]string{
		"image-alt": `<!-- Before -->
<img src="photo.jpg">

<!-- After -->
<img src="photo.jpg" alt="Description of the image content">

<!-- For decorative images -->
<img src="decoration.jpg" alt="" role="presentation">`,

		"button-name": `<!-- Before -->
<button><svg>...</svg></button>

<!-- After: Add aria-label -->
<button aria-label="Close dialog"><svg>...</svg></button>

<!-- Or: Add visually hidden text -->
<button>
  <span class="sr-only">Close dialog</span>
  <svg>...</svg>
</button>`,

		"focus-visible": `/* Add visible focus styles */
:focus-visible {
  outline: 2px solid #005fcc;
  outline-offset: 2px;
}

/* Two-color focus for visibility on any background */
:focus-visible {
  outline: 2px solid white;
  box-shadow: 0 0 0 4px #005fcc;
}`,

		"color-contrast": `/* Ensure 4.5:1 contrast ratio */
/* Tools: WebAIM Contrast Checker, Chrome DevTools */

/* Before: Low contrast */
.text { color: #767676; } /* 4.48:1 - fails */

/* After: Sufficient contrast */
.text { color: #595959; } /* 7:1 - passes */`,

		"target-size-minimum": `/* Ensure 24x24px minimum target size */
.button, .link {
  min-width: 24px;
  min-height: 24px;
  padding: 8px;
}`,
	}

	if fix, ok := fixes[f.RuleID]; ok {
		return fix
	}
	return ""
}

func buildCodeExamples(ruleID string) []string {
	examples := map[string][]string{
		"image-alt": {
			`<img src="chart.png" alt="Bar chart showing Q4 sales increased 25%">`,
			`<img src="logo.png" alt="Company Name">`,
			`<img src="divider.png" alt="" role="presentation">`,
		},
		"button-name": {
			`<button aria-label="Close">×</button>`,
			`<button type="submit">Submit Form</button>`,
		},
		"label": {
			`<label for="email">Email address</label>
<input type="email" id="email" name="email">`,
		},
	}

	return examples[ruleID]
}
