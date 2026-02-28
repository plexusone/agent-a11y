// Package remediation provides constants and URL builders for accessibility
// remediation references, similar to how net/http provides HTTP status codes.
//
// The package includes:
//   - WCAG Technique IDs (G94, H37, ARIA1, F65, etc.)
//   - ACT Rule IDs (standardized accessibility conformance tests)
//   - WCAG Success Criteria IDs
//   - URL builders for W3C, Deque, and other authoritative sources
//
// # Usage
//
// Reference technique constants:
//
//	import "github.com/agentplexus/agent-a11y/remediation"
//
//	technique := remediation.TechniqueH37 // "H37"
//	url := remediation.TechniqueURL(technique) // W3C URL
//
// Build reference URLs:
//
//	url := remediation.WCAGURL("1.1.1") // Understanding doc
//	url := remediation.ACTRuleURL("23a2a8") // ACT rule page
//
// Get technique metadata:
//
//	info := remediation.GetTechnique(remediation.TechniqueH37)
//	fmt.Println(info.Title) // "Using alt attributes on img elements"
package remediation
