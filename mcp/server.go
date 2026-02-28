// Package mcp provides a Model Context Protocol server for agent-a11y.
//
// MCP allows AI assistants like Claude Code to use agent-a11y as a tool
// for accessibility auditing. The server exposes tools for:
//   - Auditing single pages
//   - Auditing entire sites
//   - Running journey-based audits
//   - Generating VPAT reports
//   - Checking specific WCAG criteria
package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	a11y "github.com/agentplexus/agent-a11y"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version is the MCP server version.
const Version = "0.1.0"

// Server implements an MCP server for accessibility auditing.
type Server struct {
	auditor   *a11y.Auditor
	mcpServer *mcp.Server
	logger    *slog.Logger
}

// ServerConfig contains configuration for the MCP server.
type ServerConfig struct {
	Headless bool
	Level    a11y.Level
	Version  a11y.Version
	Logger   *slog.Logger
}

// NewServer creates a new MCP server.
func NewServer(cfg ServerConfig) (*Server, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Level == "" {
		cfg.Level = a11y.LevelAA
	}
	if cfg.Version == "" {
		cfg.Version = a11y.Version22
	}

	auditor, err := a11y.New(
		a11y.WithHeadless(cfg.Headless),
		a11y.WithLevel(cfg.Level),
		a11y.WithVersion(cfg.Version),
		a11y.WithLogger(cfg.Logger),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create auditor: %w", err)
	}

	// Create MCP server with implementation info
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "agent-a11y",
		Version: Version,
	}, nil)

	s := &Server{
		auditor:   auditor,
		mcpServer: mcpServer,
		logger:    cfg.Logger,
	}

	// Register tools
	s.registerTools()

	return s, nil
}

// Close releases resources used by the server.
func (s *Server) Close() error {
	if s.auditor != nil {
		return s.auditor.Close()
	}
	return nil
}

// Serve starts the MCP server using stdio transport.
func (s *Server) Serve(ctx context.Context) error {
	s.logger.Info("MCP server started", "name", "agent-a11y", "version", Version)
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// registerTools registers all accessibility auditing tools.
func (s *Server) registerTools() {
	// audit_page - Audit a single web page
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "audit_page",
		Description: "Audit a single web page for WCAG accessibility issues. Returns findings with severity, affected elements, and remediation guidance.",
	}, s.auditPage)

	// audit_site - Audit an entire website
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "audit_site",
		Description: "Audit an entire website by crawling pages. Returns aggregate findings across all pages.",
	}, s.auditSite)

	// check_criterion - Check a specific WCAG criterion
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "check_criterion",
		Description: "Check a specific WCAG success criterion on a page. Returns detailed findings for that criterion only.",
	}, s.checkCriterion)

	// generate_vpat - Generate a VPAT report
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "generate_vpat",
		Description: "Generate a VPAT 2.4 accessibility conformance report for a website.",
	}, s.generateVPAT)
}

// Tool input types

type auditPageArgs struct {
	URL   string `json:"url" jsonschema:"description=The URL of the page to audit,required"`
	Level string `json:"level,omitempty" jsonschema:"description=WCAG conformance level to test against (A, AA, AAA),enum=A,enum=AA,enum=AAA,default=AA"`
}

type auditSiteArgs struct {
	URL      string `json:"url" jsonschema:"description=The starting URL to crawl from,required"`
	MaxPages int    `json:"max_pages,omitempty" jsonschema:"description=Maximum number of pages to audit,default=50"`
	Depth    int    `json:"depth,omitempty" jsonschema:"description=Maximum crawl depth,default=3"`
}

type checkCriterionArgs struct {
	URL       string `json:"url" jsonschema:"description=The URL to check,required"`
	Criterion string `json:"criterion" jsonschema:"description=WCAG success criterion ID (e.g. 1.1.1 or 2.4.7),required"`
}

type generateVPATArgs struct {
	URL         string `json:"url" jsonschema:"description=The URL to audit,required"`
	ProductName string `json:"product_name,omitempty" jsonschema:"description=Name of the product/website"`
	VendorName  string `json:"vendor_name,omitempty" jsonschema:"description=Name of the vendor/organization"`
}

// Tool handlers

func (s *Server) auditPage(ctx context.Context, req *mcp.CallToolRequest, args auditPageArgs) (*mcp.CallToolResult, any, error) {
	s.logger.Info("audit_page called", "url", args.URL, "level", args.Level)

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	result, err := s.auditor.AuditPage(ctx, args.URL)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Format response
	text := result.Summary() + "\n\n"
	if len(result.Findings) > 0 {
		text += "Findings:\n"
		for i, f := range result.Findings {
			if i >= 10 {
				text += fmt.Sprintf("... and %d more findings\n", len(result.Findings)-10)
				break
			}
			text += fmt.Sprintf("- [%s] %s: %s (%s)\n", f.Impact, f.RuleID, f.Description, f.Selector)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, nil, nil
}

func (s *Server) auditSite(ctx context.Context, req *mcp.CallToolRequest, args auditSiteArgs) (*mcp.CallToolResult, any, error) {
	s.logger.Info("audit_site called", "url", args.URL, "max_pages", args.MaxPages, "depth", args.Depth)

	if args.MaxPages == 0 {
		args.MaxPages = 50
	}
	if args.Depth == 0 {
		args.Depth = 3
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	result, err := s.auditor.AuditSite(ctx, args.URL,
		a11y.CrawlMaxPages(args.MaxPages),
		a11y.CrawlDepth(args.Depth),
	)
	if err != nil {
		return errorResult(err), nil, nil
	}

	text := result.Summary() + "\n\n"
	text += fmt.Sprintf("Pages audited: %d\n", result.Stats.TotalPages)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, nil, nil
}

func (s *Server) checkCriterion(ctx context.Context, req *mcp.CallToolRequest, args checkCriterionArgs) (*mcp.CallToolResult, any, error) {
	s.logger.Info("check_criterion called", "url", args.URL, "criterion", args.Criterion)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	result, err := s.auditor.AuditPage(ctx, args.URL)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Filter by criterion
	findings := result.FindingsByCriterion(args.Criterion)

	text := fmt.Sprintf("WCAG %s check for %s:\n\n", args.Criterion, args.URL)
	if len(findings) == 0 {
		text += "No issues found for this criterion."
	} else {
		text += fmt.Sprintf("Found %d issues:\n", len(findings))
		for _, f := range findings {
			text += fmt.Sprintf("- [%s] %s\n  Element: %s\n  Help: %s\n\n",
				f.Impact, f.Description, f.Selector, f.Help)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, nil, nil
}

func (s *Server) generateVPAT(ctx context.Context, req *mcp.CallToolRequest, args generateVPATArgs) (*mcp.CallToolResult, any, error) {
	s.logger.Info("generate_vpat called", "url", args.URL, "product", args.ProductName, "vendor", args.VendorName)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := s.auditor.AuditSite(ctx, args.URL,
		a11y.CrawlMaxPages(20),
		a11y.CrawlDepth(2),
	)
	if err != nil {
		return errorResult(err), nil, nil
	}

	vpat, err := result.VPAT()
	if err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(vpat)},
		},
	}, nil, nil
}

// errorResult creates an error result for tool calls.
func errorResult(err error) *mcp.CallToolResult {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
		},
	}
	result.SetError(err)
	return result
}
