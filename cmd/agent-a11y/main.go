// Package main is the entry point for the agenta11y CLI.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	a11y "github.com/plexusone/agent-a11y"
	"github.com/plexusone/agent-a11y/api"
	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/agent-a11y/config"
	"github.com/plexusone/agent-a11y/mcp"
	"github.com/plexusone/agent-a11y/report"
)

var (
	// Version information (set at build time)
	Version   = "0.1.0"
	GitCommit = "dev"
	BuildDate = "unknown"
)

var (
	// Global flags
	verbose    bool
	configFile string
	headless   bool
	timeout    string
	outputFile string
	format     string

	// LLM flags
	llmProvider string
	llmModel    string
	llmAPIKey   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "agenta11y",
		Short: "WCAG Accessibility Agent",
		Long: `agenta11y is a comprehensive WCAG accessibility auditing tool.

It provides multiple interfaces:
  - CLI for direct auditing
  - MCP server for AI assistant integration (Claude Code, etc.)
  - REST API for programmatic access
  - Go library for embedding in other applications

Features:
  - WCAG 2.0, 2.1, and 2.2 support at A, AA, and AAA levels
  - LLM-as-a-Judge evaluation for nuanced accessibility assessment
  - Journey-based testing for user flow auditing
  - Site crawling for comprehensive audits
  - Multiple report formats (JSON, HTML, Markdown, VPAT)`,
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path")
	rootCmd.PersistentFlags().BoolVar(&headless, "headless", true, "Run browser in headless mode")
	rootCmd.PersistentFlags().StringVar(&timeout, "timeout", "30s", "Default timeout")

	// LLM flags
	rootCmd.PersistentFlags().StringVar(&llmProvider, "llm-provider", "", "LLM provider (anthropic, openai, gemini, ollama, xai)")
	rootCmd.PersistentFlags().StringVar(&llmModel, "llm-model", "", "LLM model name")
	rootCmd.PersistentFlags().StringVar(&llmAPIKey, "llm-api-key", "", "LLM API key (or use env vars)")

	// Add commands
	rootCmd.AddCommand(auditCmd())
	rootCmd.AddCommand(compareCmd())
	rootCmd.AddCommand(demoCmd())
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(mcpCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func auditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit <url>",
		Short: "Run an accessibility audit",
		Long: `Audit a website for WCAG accessibility compliance.

Examples:
  # Basic audit
  agenta11y audit https://example.com

  # With JSON output
  agenta11y audit https://example.com -o report.json

  # Using a config file
  agenta11y audit -c config.yaml

  # With LLM-as-a-Judge
  agenta11y audit https://example.com --llm-provider anthropic --llm-api-key $ANTHROPIC_API_KEY

  # Generate VPAT report
  agenta11y audit https://example.com -f vpat -o vpat-report.html`,
		Args: cobra.MaximumNArgs(1),
		RunE: runAudit,
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format (json, html, markdown, vpat, wcag)")

	return cmd
}

func serveCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		Long: `Start the REST API server for running audits programmatically.

The server provides endpoints for:
- Creating and managing audit jobs
- Retrieving audit results and reports
- Health checks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := setupLogger()

			// Create engine with LLM support
			engineCfg := audit.EngineConfig{
				LLMProvider: getLLMProvider(),
				LLMAPIKey:   getLLMAPIKey(),
				LLMModel:    getLLMModel(),
				Logger:      logger,
			}

			engine, err := audit.NewEngine(engineCfg)
			if err != nil {
				return fmt.Errorf("failed to create audit engine: %w", err)
			}
			defer func() {
				if err := engine.Close(); err != nil {
					logger.Warn("failed to close engine", "error", err)
				}
			}()

			server := api.NewServer(engine, logger, port)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				logger.Info("shutting down server...")
				cancel()
			}()

			return server.Start(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")

	return cmd
}

func mcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
		Long:  `Commands for the Model Context Protocol (MCP) server.`,
	}

	cmd.AddCommand(mcpServeCmd())
	return cmd
}

func mcpServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long: `Start the MCP server for AI assistant integration.

The MCP server allows AI assistants like Claude Code to use agenta11y
as a tool for accessibility auditing. Communication happens over stdin/stdout
using the Model Context Protocol.

Example usage in Claude Code:
  Add to your MCP configuration:
  {
    "mcpServers": {
      "agenta11y": {
        "command": "agenta11y",
        "args": ["mcp", "serve"]
      }
    }
  }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := setupLogger()

			server, err := mcp.NewServer(mcp.ServerConfig{
				Headless: headless,
				Level:    a11y.Level(config.DefaultConfig().WCAG.Level),
				Version:  a11y.Version(config.DefaultConfig().WCAG.Version),
				Logger:   logger,
			})
			if err != nil {
				return fmt.Errorf("failed to create MCP server: %w", err)
			}
			defer func() {
				if err := server.Close(); err != nil {
					logger.Warn("failed to close server", "error", err)
				}
			}()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				logger.Info("shutting down MCP server...")
				cancel()
			}()

			return server.Serve(ctx)
		},
	}
}

func compareCmd() *cobra.Command {
	var (
		name      string
		fromFiles bool
	)

	cmd := &cobra.Command{
		Use:   "compare <before-url> <after-url>",
		Short: "Compare before/after accessibility audits",
		Long: `Compare two versions of a page to measure accessibility improvements.

This is useful for:
  - Testing remediation efforts
  - Demonstrating accessibility fixes
  - Generating before/after VPAT reports
  - CI/CD regression testing

Examples:
  # Compare two URLs (runs audits)
  agenta11y compare https://site.com/old https://site.com/new

  # Compare from existing VPAT JSON files (no audit needed)
  agenta11y compare before.json after.json --from-files

  # With name and output
  agenta11y compare --name "Homepage Redesign" \
    https://old.site.com https://new.site.com \
    -o comparison.html -f html

  # Generate VPAT comparison
  agenta11y compare --name "Q4 Remediation" \
    https://before.example.com https://after.example.com \
    -f vpat -o vpat-comparison.md`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromFiles {
				return runComparisonFromFiles(args[0], args[1], name)
			}
			return runComparison(args[0], args[1], name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Comparison name (default: URL-based)")
	cmd.Flags().BoolVar(&fromFiles, "from-files", false, "Compare from existing VPAT JSON files instead of URLs")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&format, "format", "f", "markdown", "Output format (json, html, markdown, vpat)")

	return cmd
}

func demoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Demo site commands",
		Long:  `Commands for working with accessibility demo sites.`,
	}

	cmd.AddCommand(demoListCmd())
	cmd.AddCommand(demoRunCmd())
	cmd.AddCommand(demoGenerateCmd())

	return cmd
}

func demoListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List known accessibility demo sites",
		Long: `List known accessibility demo sites with before/after versions.

These sites are designed for learning and testing accessibility tools.`,
		Run: func(cmd *cobra.Command, args []string) {
			demos := report.KnownDemoSites()
			fmt.Println("Known Accessibility Demo Sites:")
			fmt.Println()
			for i, demo := range demos {
				fmt.Printf("%d. %s\n", i+1, demo.Name)
				fmt.Printf("   Source: %s\n", demo.Source)
				fmt.Printf("   Description: %s\n", demo.Description)
				fmt.Printf("   Before: %s\n", demo.BeforeURL)
				fmt.Printf("   After:  %s\n", demo.AfterURL)
				fmt.Println()
			}
		},
	}
}

func demoRunCmd() *cobra.Command {
	var demoIndex int

	cmd := &cobra.Command{
		Use:   "run [name]",
		Short: "Run a comparison on a demo site",
		Long: `Run a before/after comparison on a known demo site.

Examples:
  # List available demos
  agenta11y demo list

  # Run by index (from list)
  agenta11y demo run --index 1

  # Run by name (partial match)
  agenta11y demo run "W3C BAD"

  # Generate HTML comparison report
  agenta11y demo run "W3C BAD" -f html -o w3c-bad-comparison.html`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			demos := report.KnownDemoSites()

			var demo report.DemoSite
			found := false

			if demoIndex > 0 {
				// Use index
				if demoIndex > len(demos) {
					return fmt.Errorf("invalid demo index %d (max: %d)", demoIndex, len(demos))
				}
				demo = demos[demoIndex-1]
				found = true
			} else if len(args) > 0 {
				// Search by name
				search := args[0]
				for _, d := range demos {
					if containsIgnoreCase(d.Name, search) {
						demo = d
						found = true
						break
					}
				}
			}

			if !found {
				return fmt.Errorf("demo not found. Use 'agenta11y demo list' to see available demos")
			}

			fmt.Printf("Running comparison: %s\n", demo.Name)
			fmt.Printf("Before: %s\n", demo.BeforeURL)
			fmt.Printf("After:  %s\n\n", demo.AfterURL)

			return runComparison(demo.BeforeURL, demo.AfterURL, demo.Name)
		},
	}

	cmd.Flags().IntVar(&demoIndex, "index", 0, "Demo index from list")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&format, "format", "f", "markdown", "Output format (json, html, markdown, vpat)")

	return cmd
}

func demoGenerateCmd() *cobra.Command {
	var (
		outputDir   string
		siteName    string
		skipPDF     bool
		formats     []string
		defaultDir  = "examples/demo-sites"
	)

	cmd := &cobra.Command{
		Use:   "generate [site-name]",
		Short: "Generate VPAT reports for demo sites",
		Long: `Generate VPAT reports for known accessibility demo sites.

This command generates before/after VPATs and comparison reports in multiple
formats (JSON, Markdown, PDF) for all demo sites or a specific site.

PDF generation requires Pandoc to be installed. If Pandoc is not found,
PDF generation is skipped automatically.

Examples:
  # Generate reports for all demo sites
  agenta11y demo generate

  # Generate reports for a specific site
  agenta11y demo generate w3c-bad

  # Generate to a custom directory
  agenta11y demo generate --output-dir ./reports

  # Skip PDF generation
  agenta11y demo generate --skip-pdf

  # Generate only specific formats
  agenta11y demo generate --formats json,markdown`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := setupLogger()

			// Get site filter
			if len(args) > 0 {
				siteName = args[0]
			}

			// Check for pandoc
			pandocPath, err := exec.LookPath("pandoc")
			hasPandoc := err == nil && !skipPDF
			if !hasPandoc && !skipPDF {
				logger.Warn("pandoc not found, PDF generation will be skipped")
			}

			// Default formats
			if len(formats) == 0 {
				formats = []string{"json", "markdown"}
				if hasPandoc {
					formats = append(formats, "pdf")
				}
			}

			// Get demo sites
			demos := report.KnownDemoSites()

			for _, demo := range demos {
				// Filter by name or slug if specified
				if siteName != "" && !containsIgnoreCase(demo.Name, siteName) && !containsIgnoreCase(demo.Slug, siteName) {
					continue
				}

				// Use slug for directory name
				siteDir := filepath.Join(outputDir, demo.Slug)

				logger.Info("generating reports", "site", demo.Name, "dir", siteDir)

				// Generate reports for this site
				if err := generateSiteReports(cmd.Context(), logger, demo, siteDir, formats, pandocPath); err != nil {
					logger.Error("failed to generate reports", "site", demo.Name, "error", err)
					// Continue with other sites
				}
			}

			fmt.Println("Report generation complete")
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", defaultDir, "Output directory for reports")
	cmd.Flags().BoolVar(&skipPDF, "skip-pdf", false, "Skip PDF generation even if Pandoc is available")
	cmd.Flags().StringSliceVar(&formats, "formats", nil, "Output formats (json, markdown, pdf)")

	return cmd
}

// generateSiteReports generates all reports for a single demo site.
func generateSiteReports(ctx context.Context, logger *slog.Logger, demo report.DemoSite, siteDir string, formats []string, pandocPath string) error {
	// Create directories
	beforeDir := filepath.Join(siteDir, "before")
	afterDir := filepath.Join(siteDir, "after")
	comparisonDir := filepath.Join(siteDir, "comparison")

	for _, dir := range []string{beforeDir, afterDir, comparisonDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create audit engine
	engineCfg := audit.EngineConfig{
		LLMProvider: getLLMProvider(),
		LLMAPIKey:   getLLMAPIKey(),
		LLMModel:    getLLMModel(),
		Logger:      logger,
	}

	engine, err := audit.NewEngine(engineCfg)
	if err != nil {
		return fmt.Errorf("failed to create audit engine: %w", err)
	}
	defer func() {
		if err := engine.Close(); err != nil {
			logger.Warn("failed to close engine", "error", err)
		}
	}()

	// Create audit config with longer timeout for demo sites
	cfg := config.DefaultConfig()
	cfg.Browser.Headless = headless
	cfg.Browser.Timeout = config.Duration(2 * time.Minute) // Longer timeout for external sites

	// Audit "before" URL
	logger.Info("auditing before version", "url", demo.BeforeURL)
	cfg.URL = demo.BeforeURL
	beforeResult, err := engine.RunAudit(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to audit before URL: %w", err)
	}

	// Write before reports
	if err := writeReportsForResult(beforeResult, beforeDir, "vpat", formats, pandocPath, logger); err != nil {
		return fmt.Errorf("failed to write before reports: %w", err)
	}

	// Audit "after" URL
	logger.Info("auditing after version", "url", demo.AfterURL)
	cfg.URL = demo.AfterURL
	afterResult, err := engine.RunAudit(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to audit after URL: %w", err)
	}

	// Write after reports
	if err := writeReportsForResult(afterResult, afterDir, "vpat", formats, pandocPath, logger); err != nil {
		return fmt.Errorf("failed to write after reports: %w", err)
	}

	// Create and write comparison
	comparison := report.NewComparison(demo.Name, beforeResult, afterResult)
	if err := writeComparisonReports(comparison, comparisonDir, formats, pandocPath, logger); err != nil {
		return fmt.Errorf("failed to write comparison reports: %w", err)
	}

	logger.Info("site reports complete",
		"site", demo.Name,
		"beforeIssues", beforeResult.Stats.TotalFindings,
		"afterIssues", afterResult.Stats.TotalFindings,
		"improvement", fmt.Sprintf("%.1f%%", comparison.Comparison.ImprovementScore))

	return nil
}

// writeReportsForResult writes audit results in multiple formats.
func writeReportsForResult(result *audit.AuditResult, dir, baseName string, formats []string, pandocPath string, logger *slog.Logger) error {
	for _, format := range formats {
		switch format {
		case "json":
			writer := report.NewWriter(report.FormatJSON)
			path := filepath.Join(dir, baseName+".json")
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			if err := writer.Write(f, result); err != nil {
				_ = f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
			logger.Debug("wrote report", "path", path)

		case "markdown":
			writer := report.NewWriter(report.FormatVPAT)
			path := filepath.Join(dir, baseName+".md")
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			if err := writer.Write(f, result); err != nil {
				_ = f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
			logger.Debug("wrote report", "path", path)

		case "pdf":
			if pandocPath == "" {
				continue
			}
			// Generate PDF from markdown
			mdPath := filepath.Join(dir, baseName+".md")
			pdfPath := filepath.Join(dir, baseName+".pdf")

			// Ensure markdown exists
			if _, err := os.Stat(mdPath); os.IsNotExist(err) {
				// Generate markdown first
				writer := report.NewWriter(report.FormatVPAT)
				f, err := os.Create(mdPath)
				if err != nil {
					return err
				}
				if err := writer.Write(f, result); err != nil {
					_ = f.Close()
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}
			}

			if err := runPandoc(pandocPath, mdPath, pdfPath); err != nil {
				logger.Warn("failed to generate PDF", "error", err)
			} else {
				logger.Debug("wrote report", "path", pdfPath)
			}
		}
	}
	return nil
}

// writeComparisonReports writes comparison results in multiple formats.
func writeComparisonReports(result *report.ComparisonResult, dir string, formats []string, pandocPath string, logger *slog.Logger) error {
	for _, format := range formats {
		switch format {
		case "json":
			writer := report.NewWriter(report.FormatJSON)
			path := filepath.Join(dir, "comparison.json")
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			if err := writer.WriteComparison(f, result); err != nil {
				_ = f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
			logger.Debug("wrote report", "path", path)

		case "markdown":
			writer := report.NewWriter(report.FormatMarkdown)
			path := filepath.Join(dir, "comparison.md")
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			if err := writer.WriteComparison(f, result); err != nil {
				_ = f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
			logger.Debug("wrote report", "path", path)

		case "pdf":
			if pandocPath == "" {
				continue
			}
			mdPath := filepath.Join(dir, "comparison.md")
			pdfPath := filepath.Join(dir, "comparison.pdf")

			// Ensure markdown exists
			if _, err := os.Stat(mdPath); os.IsNotExist(err) {
				writer := report.NewWriter(report.FormatMarkdown)
				f, err := os.Create(mdPath)
				if err != nil {
					return err
				}
				if err := writer.WriteComparison(f, result); err != nil {
					_ = f.Close()
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}
			}

			if err := runPandoc(pandocPath, mdPath, pdfPath); err != nil {
				logger.Warn("failed to generate PDF", "error", err)
			} else {
				logger.Debug("wrote report", "path", pdfPath)
			}
		}
	}
	return nil
}

// runPandoc runs pandoc to convert markdown to PDF.
func runPandoc(pandocPath, inputPath, outputPath string) error {
	// Try xelatex first, then pdflatex
	engines := []string{"xelatex", "pdflatex", "lualatex"}

	var lastErr error
	for _, engine := range engines {
		cmd := exec.Command(pandocPath, inputPath, "-o", outputPath, "--pdf-engine="+engine)
		if err := cmd.Run(); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	// Try without specifying engine (uses pandoc default)
	cmd := exec.Command(pandocPath, inputPath, "-o", outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pandoc failed: %w (last engine error: %v)", err, lastErr)
	}
	return nil
}

func runComparison(beforeURL, afterURL, name string) error {
	logger := setupLogger()

	if name == "" {
		name = "Before vs After Comparison"
	}

	// Create audit config
	cfg := config.DefaultConfig()
	cfg.Browser.Headless = headless
	if timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.Browser.Timeout = config.Duration(d)
		}
	}

	// Create audit engine
	engineCfg := audit.EngineConfig{
		LLMProvider: getLLMProvider(),
		LLMAPIKey:   getLLMAPIKey(),
		LLMModel:    getLLMModel(),
		Logger:      logger,
	}

	engine, err := audit.NewEngine(engineCfg)
	if err != nil {
		return fmt.Errorf("failed to create audit engine: %w", err)
	}
	defer func() {
		if err := engine.Close(); err != nil {
			logger.Warn("failed to close engine", "error", err)
		}
	}()

	ctx := context.Background()

	// Audit "before" URL
	logger.Info("auditing before version", "url", beforeURL)
	cfg.URL = beforeURL
	beforeResult, err := engine.RunAudit(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to audit before URL: %w", err)
	}

	// Audit "after" URL
	logger.Info("auditing after version", "url", afterURL)
	cfg.URL = afterURL
	afterResult, err := engine.RunAudit(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to audit after URL: %w", err)
	}

	// Create comparison
	comparison := report.NewComparison(name, beforeResult, afterResult)

	// Write output
	writer := report.NewWriter(report.Format(format))

	var out *os.File
	if outputFile != "" {
		out, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			if cerr := out.Close(); cerr != nil {
				logger.Warn("failed to close output file", "error", cerr)
			}
		}()
	} else {
		out = os.Stdout
	}

	if err := writer.WriteComparison(out, comparison); err != nil {
		return fmt.Errorf("failed to write comparison report: %w", err)
	}

	// Print summary
	if outputFile != "" {
		logger.Info("comparison saved", "file", outputFile)
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Before: %d issues\n", comparison.Comparison.BeforeTotalIssues)
	fmt.Printf("After:  %d issues\n", comparison.Comparison.AfterTotalIssues)
	fmt.Printf("Fixed:  %d issues\n", comparison.Comparison.IssuesFixed)
	fmt.Printf("Improvement: %.1f%%\n", comparison.Comparison.ImprovementScore)

	return nil
}

// runComparisonFromFiles compares two existing VPAT JSON files without running new audits.
func runComparisonFromFiles(beforeFile, afterFile, name string) error {
	logger := setupLogger()

	if name == "" {
		name = "Before vs After Comparison"
	}

	// Read before file
	beforeData, err := os.ReadFile(beforeFile)
	if err != nil {
		return fmt.Errorf("failed to read before file: %w", err)
	}

	var beforeResult audit.AuditResult
	if err := json.Unmarshal(beforeData, &beforeResult); err != nil {
		return fmt.Errorf("failed to parse before file: %w", err)
	}

	// Read after file
	afterData, err := os.ReadFile(afterFile)
	if err != nil {
		return fmt.Errorf("failed to read after file: %w", err)
	}

	var afterResult audit.AuditResult
	if err := json.Unmarshal(afterData, &afterResult); err != nil {
		return fmt.Errorf("failed to parse after file: %w", err)
	}

	// Create comparison
	comparison := report.NewComparison(name, &beforeResult, &afterResult)

	// Write output
	writer := report.NewWriter(report.Format(format))

	var out *os.File
	if outputFile != "" {
		out, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			if cerr := out.Close(); cerr != nil {
				logger.Warn("failed to close output file", "error", cerr)
			}
		}()
	} else {
		out = os.Stdout
	}

	if err := writer.WriteComparison(out, comparison); err != nil {
		return fmt.Errorf("failed to write comparison report: %w", err)
	}

	// Print summary
	if outputFile != "" {
		logger.Info("comparison saved", "file", outputFile)
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Before: %d issues\n", comparison.Comparison.BeforeTotalIssues)
	fmt.Printf("After:  %d issues\n", comparison.Comparison.AfterTotalIssues)
	fmt.Printf("Fixed:  %d issues\n", comparison.Comparison.IssuesFixed)
	fmt.Printf("Improvement: %.1f%%\n", comparison.Comparison.ImprovementScore)

	return nil
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("agenta11y version %s\n", Version)
			fmt.Printf("  Git commit: %s\n", GitCommit)
			fmt.Printf("  Build date: %s\n", BuildDate)
		},
	}
}

func runAudit(cmd *cobra.Command, args []string) error {
	logger := setupLogger()

	// Load config
	var cfg *config.Config
	if configFile != "" {
		var err error
		cfg, err = config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// Override with CLI args
	if len(args) > 0 {
		cfg.URL = args[0]
	}

	if cfg.URL == "" {
		return fmt.Errorf("URL is required")
	}

	cfg.Browser.Headless = headless
	if timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.Browser.Timeout = config.Duration(d)
		}
	}

	// Ensure LLM config exists before overriding
	if cfg.LLM == nil {
		cfg.LLM = &config.LLMConfig{}
	}

	// Override LLM config from flags
	if getLLMProvider() != "" {
		cfg.LLM.Provider = getLLMProvider()
	}
	if getLLMModel() != "" {
		cfg.LLM.Model = getLLMModel()
	}
	if getLLMAPIKey() != "" {
		cfg.LLM.APIKey = getLLMAPIKey()
	}

	// Create audit engine
	engineCfg := audit.EngineConfig{
		LLMProvider: cfg.LLM.Provider,
		LLMAPIKey:   cfg.LLM.APIKey,
		LLMModel:    cfg.LLM.Model,
		Logger:      logger,
	}

	engine, err := audit.NewEngine(engineCfg)
	if err != nil {
		return fmt.Errorf("failed to create audit engine: %w", err)
	}
	defer func() {
		if err := engine.Close(); err != nil {
			logger.Warn("failed to close engine", "error", err)
		}
	}()

	// Run the audit
	ctx := context.Background()
	result, err := engine.RunAudit(ctx, cfg)
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	// Write output
	writer := report.NewWriter(report.Format(format))

	var out *os.File
	if outputFile != "" {
		out, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			if cerr := out.Close(); cerr != nil {
				logger.Warn("failed to close output file", "error", cerr)
			}
		}()
	} else {
		out = os.Stdout
	}

	if err := writer.Write(out, result); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	if outputFile != "" {
		logger.Info("report saved", "file", outputFile)
	}

	return nil
}

func setupLogger() *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

// getLLMProvider returns the LLM provider from flags or env vars.
func getLLMProvider() string {
	if llmProvider != "" {
		return llmProvider
	}
	// Check env vars
	if provider := os.Getenv("LLM_PROVIDER"); provider != "" {
		return provider
	}
	// Auto-detect from API key env vars
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "anthropic"
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "openai"
	}
	if os.Getenv("GEMINI_API_KEY") != "" {
		return "gemini"
	}
	return ""
}

// getLLMAPIKey returns the LLM API key from flags or env vars.
func getLLMAPIKey() string {
	if llmAPIKey != "" {
		return llmAPIKey
	}
	// Check provider-specific env vars
	provider := getLLMProvider()
	switch provider {
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "gemini":
		return os.Getenv("GEMINI_API_KEY")
	case "xai":
		return os.Getenv("XAI_API_KEY")
	}
	return os.Getenv("LLM_API_KEY")
}

// getLLMModel returns the LLM model from flags or defaults.
func getLLMModel() string {
	if llmModel != "" {
		return llmModel
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		return model
	}
	// Default models per provider
	provider := getLLMProvider()
	switch provider {
	case "anthropic":
		return "claude-sonnet-4-20250514"
	case "openai":
		return "gpt-4o"
	case "gemini":
		return "gemini-1.5-pro"
	default:
		return ""
	}
}
