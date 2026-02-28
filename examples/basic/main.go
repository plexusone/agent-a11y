// Example: Basic usage of agent-a11y as a Go library.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	a11y "github.com/plexusone/agent-a11y"
)

func main() {
	// Create an auditor with options
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

	// Audit a single page
	result, err := auditor.AuditPage(ctx, "https://example.com")
	if err != nil {
		log.Fatalf("Audit failed: %v", err)
	}

	// Print summary
	fmt.Println(result.Summary())

	// Check conformance
	if result.Conformant() {
		fmt.Println("Site is conformant!")
	} else {
		fmt.Println("Site has accessibility issues:")
		for _, f := range result.CriticalFindings() {
			fmt.Printf("  - [CRITICAL] %s: %s\n", f.RuleID, f.Description)
		}
		for _, f := range result.SeriousFindings() {
			fmt.Printf("  - [SERIOUS] %s: %s\n", f.RuleID, f.Description)
		}
	}

	// Generate HTML report
	html, err := result.HTML()
	if err != nil {
		log.Fatalf("Failed to generate HTML: %v", err)
	}
	fmt.Printf("\nHTML report generated (%d bytes)\n", len(html))
}
