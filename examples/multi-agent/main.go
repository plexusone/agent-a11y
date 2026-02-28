// Example: Multi-agent integration with agent-a11y.
//
// This example shows how to use agent-a11y as part of a multi-agent workflow,
// returning standardized go/no-go decisions and narrative reports.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	a11y "github.com/agentplexus/agent-a11y"
	mas "github.com/agentplexus/multi-agent-spec/sdk/go"
)

func main() {
	// Create an auditor
	auditor, err := a11y.New(
		a11y.WithHeadless(true),
		a11y.WithLevel(a11y.LevelAA),
		a11y.WithVersion(a11y.Version22),
		a11y.WithTimeout(2*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create auditor: %v", err)
	}
	defer func() { _ = auditor.Close() }()

	ctx := context.Background()

	// Audit a page
	result, err := auditor.AuditPage(ctx, "https://example.com")
	if err != nil {
		log.Fatalf("Audit failed: %v", err)
	}

	// === Multi-Agent Integration ===

	// Get AgentResult for multi-agent workflows
	agentResult := result.AgentResult()

	fmt.Println("=== Agent Result ===")
	fmt.Printf("Agent ID: %s\n", agentResult.AgentID)
	fmt.Printf("Status: %s %s\n", agentResult.Status.Icon(), agentResult.Status)
	fmt.Printf("Tasks: %d\n", len(agentResult.Tasks))

	// Print tasks
	for _, task := range agentResult.Tasks {
		fmt.Printf("  - %s: %s %s - %s\n",
			task.ID, task.Status.Icon(), task.Status, task.Detail)
	}

	// Get narrative for prose reports
	narrative := result.Narrative()
	fmt.Println("\n=== Narrative ===")
	fmt.Printf("Problem: %s\n\n", narrative.Problem)
	fmt.Printf("Analysis: %s\n\n", narrative.Analysis)
	fmt.Printf("Recommendation: %s\n", narrative.Recommendation)

	// Get TeamSection for inclusion in a TeamReport
	teamSection := result.TeamSection()

	// Create a full TeamReport
	report := &mas.TeamReport{
		Project:     "example-site",
		Version:     "1.0.0",
		Target:      "WCAG 2.2 AA Conformance",
		Phase:       "ACCESSIBILITY AUDIT",
		Teams:       []mas.TeamSection{teamSection},
		GeneratedAt: time.Now().UTC(),
		GeneratedBy: "agent-a11y",
	}
	report.Status = report.ComputeOverallStatus()

	fmt.Println("\n=== Team Report ===")
	fmt.Printf("Overall Status: %s %s\n", report.Status.Icon(), report.Status)
	fmt.Println(report.FinalMessage())

	// Output JSON
	jsonBytes, err := json.MarshalIndent(agentResult, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println("\n=== JSON Output ===")
	fmt.Println(string(jsonBytes))
}
