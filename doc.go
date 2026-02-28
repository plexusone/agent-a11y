// Package a11y provides comprehensive WCAG accessibility auditing for websites.
//
// agent-a11y is a Go library and CLI tool for automated accessibility testing,
// supporting WCAG 2.0, 2.1, and 2.2 at levels A, AA, and AAA. It includes
// LLM-based evaluation, journey testing, site crawling, and multiple report formats.
//
// # Quick Start
//
// The simplest way to audit a page:
//
//	auditor := a11y.New()
//	result, err := auditor.AuditPage(ctx, "https://example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Score: %d, Issues: %d\n", result.Score, len(result.Findings))
//
// # Configuration
//
// Use functional options to configure the auditor:
//
//	auditor := a11y.New(
//	    a11y.WithLevel(a11y.LevelAA),        // WCAG AA conformance
//	    a11y.WithHeadless(true),              // Run browser headlessly
//	    a11y.WithTimeout(2 * time.Minute),    // Set timeout
//	    a11y.WithLLM("anthropic", "claude-sonnet-4-20250514"), // Enable LLM evaluation
//	)
//
// # Site Crawling
//
// Audit an entire website:
//
//	result, err := auditor.AuditSite(ctx, "https://example.com",
//	    a11y.CrawlDepth(3),
//	    a11y.CrawlMaxPages(100),
//	)
//
// # Journey Testing
//
// Test user flows with journey definitions:
//
//	result, err := auditor.AuditJourney(ctx, "https://example.com", "checkout.yaml")
//
// # Reports
//
// Generate reports in multiple formats:
//
//	// JSON report
//	json, _ := result.JSON()
//
//	// HTML report
//	html, _ := result.HTML()
//
//	// VPAT 2.4 report
//	vpat, _ := result.VPAT()
//
// # Multi-Agent Integration
//
// Convert results to multi-agent-spec format for go/no-go decisions:
//
//	result, _ := auditor.AuditPage(ctx, "https://example.com")
//
//	// Get AgentResult for multi-agent workflows
//	agentResult := result.AgentResult()
//	fmt.Printf("Status: %s\n", agentResult.Status) // GO, WARN, NO-GO
//
//	// Get TeamSection for inclusion in a TeamReport
//	teamSection := result.TeamSection()
//
//	// Get narrative for prose reports
//	narrative := result.Narrative()
//	fmt.Println(narrative.Problem)
//	fmt.Println(narrative.Analysis)
//	fmt.Println(narrative.Recommendation)
//
// # MCP Server
//
// Start an MCP server for AI assistant integration:
//
//	server := a11y.NewMCPServer()
//	server.Serve(ctx)
//
// # CLI Usage
//
// The agent-a11y command provides CLI access:
//
//	# Audit a page
//	agent-a11y audit https://example.com
//
//	# Audit with specific level
//	agent-a11y audit https://example.com --level AA
//
//	# Generate VPAT report
//	agent-a11y vpat https://example.com -o vpat.html
//
//	# Start MCP server
//	agent-a11y mcp serve
package a11y
