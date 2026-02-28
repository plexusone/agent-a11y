package remediation

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/agentplexus/agent-a11y/types"
)

func TestTaskBuilder_BuildTasks(t *testing.T) {
	findings := []types.Finding{
		{
			ID:              "finding-1",
			RuleID:          "image-alt",
			Description:     "Image missing alt text",
			SuccessCriteria: []string{"1.1.1"},
			Level:           types.WCAGLevelA,
			Impact:          types.ImpactSerious,
			Severity:        types.SeveritySerious,
			Selector:        "img.hero",
			PageURL:         "https://example.com/",
			PageTitle:       "Home",
		},
		{
			ID:              "finding-2",
			RuleID:          "image-alt",
			Description:     "Image missing alt text",
			SuccessCriteria: []string{"1.1.1"},
			Level:           types.WCAGLevelA,
			Impact:          types.ImpactSerious,
			Severity:        types.SeveritySerious,
			Selector:        "img.product",
			PageURL:         "https://example.com/products",
			PageTitle:       "Products",
		},
		{
			ID:              "finding-3",
			RuleID:          "button-name",
			Description:     "Button missing accessible name",
			SuccessCriteria: []string{"4.1.2"},
			Level:           types.WCAGLevelA,
			Impact:          types.ImpactCritical,
			Severity:        types.SeverityCritical,
			Selector:        "button.close",
			PageURL:         "https://example.com/",
			PageTitle:       "Home",
		},
	}

	builder := DefaultTaskBuilder()
	tasks := builder.BuildTasks(findings)

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks (grouped by rule), got %d", len(tasks))
	}

	// Check that image-alt findings are grouped
	var imgTask *RemediationTask
	for i := range tasks {
		if tasks[i].RuleID == "image-alt" {
			imgTask = &tasks[i]
			break
		}
	}

	if imgTask == nil {
		t.Fatal("image-alt task not found")
	}

	if imgTask.AffectedCount != 2 {
		t.Errorf("expected 2 affected elements, got %d", imgTask.AffectedCount)
	}

	if len(imgTask.AcceptanceCriteria) == 0 {
		t.Error("expected acceptance criteria to be populated")
	}
}

func TestRemediationTask_ToJiraTicket(t *testing.T) {
	task := RemediationTask{
		ID:              "a11y-12345678",
		Title:           "Add alt text to images (2 instances)",
		Summary:         "Provide text alternatives for images",
		Priority:        PriorityMajor,
		Labels:          []string{"accessibility", "wcag-a"},
		SuccessCriteria: []string{"1.1.1"},
		WCAGLevel:       "A",
		RuleID:          "image-alt",
		AffectedCount:   2,
		AffectedElements: []AffectedElement{
			{Selector: "img.hero", PageURL: "https://example.com/", PageTitle: "Home"},
		},
		AcceptanceCriteria: []string{
			"All images have alt text",
			"Decorative images have empty alt",
		},
		References: []types.ReferenceURL{
			{Title: "Understanding 1.1.1", URL: "https://w3.org/...", Source: "w3c"},
		},
	}

	ticket := task.ToJiraTicket("A11Y", "Task")

	if ticket.Fields.Project.Key != "A11Y" {
		t.Errorf("expected project key 'A11Y', got %s", ticket.Fields.Project.Key)
	}

	if ticket.Fields.Summary != task.Title {
		t.Errorf("expected summary %q, got %q", task.Title, ticket.Fields.Summary)
	}

	if ticket.Fields.IssueType.Name != "Task" {
		t.Errorf("expected issue type 'Task', got %s", ticket.Fields.IssueType.Name)
	}

	if ticket.Fields.Priority.Name != string(PriorityMajor) {
		t.Errorf("expected priority 'Major', got %s", ticket.Fields.Priority.Name)
	}

	// Verify description contains key sections
	desc := ticket.Fields.Description
	if !strings.Contains(desc, "Summary") {
		t.Error("description should contain Summary section")
	}
	if !strings.Contains(desc, "Acceptance Criteria") {
		t.Error("description should contain Acceptance Criteria section")
	}
}

func TestRemediationTask_ToGitHubIssue(t *testing.T) {
	task := RemediationTask{
		ID:              "a11y-12345678",
		Title:           "Add alt text to images",
		Summary:         "Provide text alternatives for images",
		Labels:          []string{"accessibility", "wcag-1-1-1"},
		SuccessCriteria: []string{"1.1.1"},
		WCAGLevel:       "A",
		AffectedCount:   2,
		AcceptanceCriteria: []string{
			"All images have alt text",
		},
	}

	issue := task.ToGitHubIssue()

	if issue.Title != task.Title {
		t.Errorf("expected title %q, got %q", task.Title, issue.Title)
	}

	if len(issue.Labels) != len(task.Labels) {
		t.Errorf("expected %d labels, got %d", len(task.Labels), len(issue.Labels))
	}

	// Verify body contains markdown checkboxes for acceptance criteria
	if !strings.Contains(issue.Body, "- [ ]") {
		t.Error("body should contain acceptance criteria checkboxes")
	}
}

func TestRemediationTask_ToLLMAgentPrompt(t *testing.T) {
	task := RemediationTask{
		ID:              "a11y-12345678",
		Title:           "Add alt text to images",
		Summary:         "Provide text alternatives for images",
		RuleID:          "image-alt",
		SuccessCriteria: []string{"1.1.1"},
		WCAGLevel:       "A",
		AffectedCount:   2,
		AcceptanceCriteria: []string{
			"All images have alt text",
		},
		TechniqueRefs: []types.TechniqueRef{
			{ID: "H37", Type: "sufficient", Title: "Using alt attributes"},
		},
		References: []types.ReferenceURL{
			{Title: "Understanding 1.1.1", URL: "https://w3.org/..."},
		},
		CodeExamples: []string{
			`<img src="photo.jpg" alt="Description">`,
		},
	}

	prompt := task.ToLLMAgentPrompt()

	if prompt.TaskID != task.ID {
		t.Errorf("expected task ID %q, got %q", task.ID, prompt.TaskID)
	}

	if prompt.Objective != task.Title {
		t.Errorf("expected objective %q, got %q", task.Title, prompt.Objective)
	}

	if len(prompt.Requirements) == 0 {
		t.Error("expected requirements to be populated")
	}

	if len(prompt.AcceptanceCriteria) == 0 {
		t.Error("expected acceptance criteria")
	}

	if !strings.Contains(prompt.CommitMessage, "fix(a11y)") {
		t.Error("commit message should follow conventional commits")
	}

	if !strings.Contains(prompt.PRTitle, "fix(a11y)") {
		t.Error("PR title should follow conventional commits")
	}
}

func TestJiraTicket_JSON(t *testing.T) {
	task := RemediationTask{
		ID:       "a11y-test",
		Title:    "Test task",
		Summary:  "Test summary",
		Priority: PriorityMajor,
	}

	ticket := task.ToJiraTicket("TEST", "Bug")

	jsonBytes, err := json.MarshalIndent(ticket, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal to JSON: %v", err)
	}

	// Verify it's valid JSON that can be sent to Jira API
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	fields, ok := parsed["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'fields' object in JSON")
	}

	if fields["summary"] != "Test task" {
		t.Errorf("expected summary 'Test task', got %v", fields["summary"])
	}
}

func TestPriorityMapping(t *testing.T) {
	tests := []struct {
		impact   types.Impact
		severity types.Severity
		expected TaskPriority
	}{
		{types.ImpactBlocker, types.SeverityCritical, PriorityBlocker},
		{types.ImpactCritical, types.SeverityCritical, PriorityCritical},
		{types.ImpactSerious, types.SeveritySerious, PriorityMajor},
		{types.ImpactModerate, types.SeverityModerate, PriorityMinor},
		{types.ImpactMinor, types.SeverityMinor, PriorityTrivial},
	}

	for _, tt := range tests {
		t.Run(string(tt.impact), func(t *testing.T) {
			got := mapPriority(tt.impact, tt.severity)
			if got != tt.expected {
				t.Errorf("mapPriority(%s, %s) = %s, want %s", tt.impact, tt.severity, got, tt.expected)
			}
		})
	}
}

func TestTasksAreSortedByPriority(t *testing.T) {
	findings := []types.Finding{
		{
			ID:              "1",
			RuleID:          "minor-issue",
			SuccessCriteria: []string{"1.1.1"},
			Impact:          types.ImpactMinor,
		},
		{
			ID:              "2",
			RuleID:          "critical-issue",
			SuccessCriteria: []string{"2.1.2"},
			Impact:          types.ImpactBlocker,
		},
		{
			ID:              "3",
			RuleID:          "moderate-issue",
			SuccessCriteria: []string{"1.4.3"},
			Impact:          types.ImpactModerate,
		},
	}

	builder := DefaultTaskBuilder()
	tasks := builder.BuildTasks(findings)

	// Tasks should be sorted: Blocker, then Moderate, then Minor
	if len(tasks) < 2 {
		t.Skip("not enough tasks for priority test")
	}

	if priorityOrder(tasks[0].Priority) > priorityOrder(tasks[1].Priority) {
		t.Errorf("tasks not sorted by priority: %s should come before %s",
			tasks[0].Priority, tasks[1].Priority)
	}
}
