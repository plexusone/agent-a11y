package report

import (
	"fmt"
	"io"
	"time"

	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/openacr-go"
	"github.com/plexusone/openacr-go/catalog"
)

// DefaultOpenACRCatalog is the default catalog ID for generated OpenACR reports.
const DefaultOpenACRCatalog = "2.5-edition-wcag-2.2-508-en"

// writeOpenACR writes the audit result as an OpenACR document.
func (w *Writer) writeOpenACR(out io.Writer, result *audit.AuditResult) error {
	report := convertToOpenACR(result)
	return report.WriteYAML(out)
}

// convertToOpenACR converts an AuditResult to an OpenACR Report.
func convertToOpenACR(result *audit.AuditResult) *openacr.Report {
	report := openacr.NewReport(
		openacr.WithTitle(fmt.Sprintf("%s Accessibility Conformance Report", result.TargetURL)),
		openacr.WithProduct(result.TargetURL, ""),
		openacr.WithAuthor("agent-a11y", ""),
		openacr.WithCatalog(selectCatalog(result.WCAGVersion)),
		openacr.WithReportDate(result.StartTime.Format("2006-01-02")),
		openacr.WithEvaluationMethods("Automated accessibility testing with agent-a11y"),
		openacr.WithLegalDisclaimer("This report is generated automatically and represents the accessibility status at the time of evaluation."),
	)

	// Build chapters from conformance criteria
	report.Chapters = buildChapters(result)

	return report
}

// selectCatalog selects the appropriate OpenACR catalog based on WCAG version.
func selectCatalog(version audit.WCAGVersion) string {
	switch version {
	case audit.WCAG20:
		return "2.5-edition-wcag-2.0-508-en"
	case audit.WCAG21:
		return "2.5-edition-wcag-2.1-508-en"
	case audit.WCAG22:
		return "2.5-edition-wcag-2.2-508-en"
	default:
		return DefaultOpenACRCatalog
	}
}

// buildChapters converts audit conformance results to OpenACR chapters.
func buildChapters(result *audit.AuditResult) map[string]openacr.Chapter {
	chapters := make(map[string]openacr.Chapter)

	// Group criteria by chapter (level)
	levelACriteria := make([]openacr.Criterion, 0)
	levelAACriteria := make([]openacr.Criterion, 0)
	levelAAACriteria := make([]openacr.Criterion, 0)

	for _, cr := range result.Conformance.Criteria {
		criterion := openacr.Criterion{
			Num: cr.ID,
			Components: []openacr.Component{
				{
					Name: openacr.ComponentWeb,
					Adherence: openacr.Adherence{
						Level: mapConformanceStatus(cr.Status),
						Notes: buildCriterionNotes(cr),
					},
				},
			},
		}

		switch cr.Level {
		case "A":
			levelACriteria = append(levelACriteria, criterion)
		case "AA":
			levelAACriteria = append(levelAACriteria, criterion)
		case "AAA":
			levelAAACriteria = append(levelAAACriteria, criterion)
		}
	}

	// Add chapters if they have criteria
	if len(levelACriteria) > 0 {
		chapters["success_criteria_level_a"] = openacr.Chapter{
			Notes:    buildLevelNotes(result.Conformance.LevelA),
			Criteria: levelACriteria,
		}
	}

	if len(levelAACriteria) > 0 {
		chapters["success_criteria_level_aa"] = openacr.Chapter{
			Notes:    buildLevelNotes(result.Conformance.LevelAA),
			Criteria: levelAACriteria,
		}
	}

	if len(levelAAACriteria) > 0 {
		chapters["success_criteria_level_aaa"] = openacr.Chapter{
			Notes:    buildLevelNotes(result.Conformance.LevelAAA),
			Criteria: levelAAACriteria,
		}
	}

	// Add hardware/software chapters as not applicable for web audits
	chapters["hardware"] = openacr.Chapter{
		Notes:    "This is a web application. Hardware accessibility criteria is not applicable.",
		Disabled: true,
	}

	chapters["software"] = openacr.Chapter{
		Notes:    "This is a web application. Software-specific accessibility criteria is not applicable.",
		Disabled: true,
	}

	return chapters
}

// mapConformanceStatus converts audit status to OpenACR adherence level.
func mapConformanceStatus(status string) openacr.AdherenceLevel {
	switch status {
	case "supports", "Supports":
		return openacr.LevelSupports
	case "partially_supports", "Partially Supports", "partially-supports":
		return openacr.LevelPartiallySupports
	case "does_not_support", "Does Not Support", "does-not-support":
		return openacr.LevelDoesNotSupport
	case "not_applicable", "Not Applicable", "not-applicable":
		return openacr.LevelNotApplicable
	case "not_evaluated", "Not Evaluated", "not-evaluated":
		return openacr.LevelNotEvaluated
	default:
		return openacr.LevelNotEvaluated
	}
}

// buildCriterionNotes builds notes for a criterion result.
func buildCriterionNotes(cr audit.CriterionResult) string {
	if cr.Remarks != "" {
		return cr.Remarks
	}
	if cr.IssueCount > 0 {
		return fmt.Sprintf("%d issue(s) found during evaluation.", cr.IssueCount)
	}
	return ""
}

// buildLevelNotes builds notes for a conformance level.
func buildLevelNotes(level audit.LevelConformance) string {
	if level.TotalIssues == 0 {
		return "No issues found at this level."
	}
	if level.BlockingIssues > 0 {
		return fmt.Sprintf("%d total issues, %d blocking issues.", level.TotalIssues, level.BlockingIssues)
	}
	return fmt.Sprintf("%d issue(s) found at this level.", level.TotalIssues)
}

// OpenACROptions provides options for OpenACR generation.
type OpenACROptions struct {
	// ProductName overrides the default product name (target URL).
	ProductName string

	// ProductVersion sets the product version.
	ProductVersion string

	// AuthorName sets the author name.
	AuthorName string

	// AuthorEmail sets the author email.
	AuthorEmail string

	// VendorName sets the vendor company name.
	VendorName string

	// VendorEmail sets the vendor email.
	VendorEmail string

	// CatalogID overrides the default catalog selection.
	CatalogID string

	// IncludeAllCriteria includes all catalog criteria, not just evaluated ones.
	IncludeAllCriteria bool
}

// GenerateOpenACR generates an OpenACR report with custom options.
func GenerateOpenACR(result *audit.AuditResult, opts OpenACROptions) (*openacr.Report, error) {
	// Start with basic conversion
	report := convertToOpenACR(result)

	// Apply options
	if opts.ProductName != "" {
		report.Product.Name = opts.ProductName
	}
	if opts.ProductVersion != "" {
		report.Product.Version = opts.ProductVersion
	}
	if opts.AuthorName != "" || opts.AuthorEmail != "" {
		report.Author.Name = opts.AuthorName
		report.Author.Email = opts.AuthorEmail
	}
	if opts.VendorName != "" || opts.VendorEmail != "" {
		report.Vendor = &openacr.Contact{
			CompanyName: opts.VendorName,
			Email:       opts.VendorEmail,
		}
	}
	if opts.CatalogID != "" {
		report.Catalog = opts.CatalogID
	}

	// Include all criteria from catalog if requested
	if opts.IncludeAllCriteria {
		cat, err := catalog.Get(report.Catalog)
		if err != nil {
			return nil, fmt.Errorf("loading catalog: %w", err)
		}
		report.Chapters = buildChaptersFromCatalog(result, cat)
	}

	return report, nil
}

// buildChaptersFromCatalog builds chapters with all criteria from the catalog.
func buildChaptersFromCatalog(result *audit.AuditResult, cat *catalog.Catalog) map[string]openacr.Chapter {
	chapters := make(map[string]openacr.Chapter)

	// Build lookup for evaluated criteria
	evaluated := make(map[string]audit.CriterionResult)
	for _, cr := range result.Conformance.Criteria {
		evaluated[cr.ID] = cr
	}

	// Process each catalog chapter
	for _, chDef := range cat.Chapters {
		var criteria []openacr.Criterion

		for _, crDef := range chDef.Criteria {
			var adherence openacr.Adherence

			if cr, ok := evaluated[crDef.ID]; ok {
				// Criterion was evaluated
				adherence = openacr.Adherence{
					Level: mapConformanceStatus(cr.Status),
					Notes: buildCriterionNotes(cr),
				}
			} else {
				// Criterion not evaluated
				adherence = openacr.Adherence{
					Level: openacr.LevelNotEvaluated,
				}
			}

			// Determine applicable components from catalog
			var components []openacr.Component
			for _, compID := range crDef.Components {
				compName := openacr.ComponentName(compID)
				if compName == openacr.ComponentWeb {
					components = append(components, openacr.Component{
						Name:      compName,
						Adherence: adherence,
					})
				} else {
					// Non-web components are not applicable for web audits
					components = append(components, openacr.Component{
						Name: compName,
						Adherence: openacr.Adherence{
							Level: openacr.LevelNotApplicable,
						},
					})
				}
			}

			criteria = append(criteria, openacr.Criterion{
				Num:        crDef.ID,
				Components: components,
			})
		}

		if len(criteria) > 0 {
			chapters[chDef.ID] = openacr.Chapter{
				Criteria: criteria,
			}
		}
	}

	return chapters
}

// ValidateOpenACRReport validates a generated OpenACR report.
func ValidateOpenACRReport(report *openacr.Report) []openacr.ValidationError {
	return report.Validate()
}

// ValidateOpenACRAgainstCatalog validates a report against its catalog.
func ValidateOpenACRAgainstCatalog(report *openacr.Report) ([]openacr.ValidationError, error) {
	cat, err := catalog.Get(report.Catalog)
	if err != nil {
		return nil, fmt.Errorf("loading catalog %s: %w", report.Catalog, err)
	}
	return report.ValidateAgainstCatalog(cat), nil
}

// OpenACRReportDate returns the report date for audit results.
func OpenACRReportDate(result *audit.AuditResult) time.Time {
	return result.StartTime
}
