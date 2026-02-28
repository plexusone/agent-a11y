package remediation

import "github.com/plexusone/agent-a11y/types"

// BuildForFinding creates a Remediation struct for a finding based on its
// success criteria. This attaches standardized reference codes similar to
// how SAST tools attach CWE IDs.
func BuildForFinding(successCriteria []string, axeRuleID string) *types.Remediation {
	if len(successCriteria) == 0 {
		return nil
	}

	// Use the first criterion as primary
	primaryCriterion := successCriteria[0]

	// Get techniques for this criterion
	techs := TechniquesForCriterion(primaryCriterion)
	var techRefs []types.TechniqueRef
	for _, t := range techs {
		techRefs = append(techRefs, types.TechniqueRef{
			ID:    string(t.ID),
			Type:  string(t.Type),
			Title: t.Title,
			URL:   TechniqueURL(t.ID),
		})
	}

	// Build reference URLs
	var refs []types.ReferenceURL

	// W3C Understanding document
	refs = append(refs, types.ReferenceURL{
		Title:  "Understanding " + primaryCriterion,
		URL:    UnderstandingURL(primaryCriterion),
		Source: "w3c",
	})

	// W3C Quick Reference
	refs = append(refs, types.ReferenceURL{
		Title:  "How to Meet " + primaryCriterion,
		URL:    QuickRefURL(primaryCriterion),
		Source: "w3c",
	})

	// WebAIM checklist
	refs = append(refs, types.ReferenceURL{
		Title:  "WebAIM Checklist " + primaryCriterion,
		URL:    WebAIMChecklistURL(primaryCriterion),
		Source: "webaim",
	})

	// axe-core reference if provided
	if axeRuleID != "" {
		refs = append(refs, types.ReferenceURL{
			Title:  "axe-core: " + axeRuleID,
			URL:    AxeRuleURL(axeRuleID),
			Source: "deque",
		})
	}

	// Find ACT rule if available
	actRules := ACTRulesForCriterion(primaryCriterion)
	var actRuleID string
	if len(actRules) > 0 {
		actRuleID = string(actRules[0].ID)
		refs = append(refs, types.ReferenceURL{
			Title:  "ACT Rule: " + actRules[0].Name,
			URL:    ACTRuleURL(actRules[0].ID),
			Source: "w3c",
		})
	}

	// Build summary from sufficient techniques
	summary := buildSummary(primaryCriterion)

	return &types.Remediation{
		Summary:    summary,
		Techniques: techRefs,
		ACTRuleID:  actRuleID,
		AxeRuleID:  axeRuleID,
		References: refs,
	}
}

// buildSummary generates a brief summary for a criterion.
func buildSummary(criterionID string) string {
	summaries := map[string]string{
		"1.1.1":  "Provide a text alternative that serves the same purpose as the non-text content.",
		"1.2.1":  "Provide an alternative for time-based media (transcript for audio, description for video).",
		"1.2.2":  "Provide captions for all prerecorded audio content in synchronized media.",
		"1.2.3":  "Provide audio description or a media alternative for prerecorded video.",
		"1.2.5":  "Provide audio description for all prerecorded video content.",
		"1.3.1":  "Use semantic markup to convey structure and relationships programmatically.",
		"1.3.2":  "Present content in a meaningful sequence that can be programmatically determined.",
		"1.3.3":  "Provide instructions that don't rely solely on sensory characteristics.",
		"1.3.4":  "Don't restrict content display to a single orientation unless essential.",
		"1.3.5":  "Use autocomplete attributes to identify the purpose of input fields.",
		"1.4.1":  "Don't use color as the only means of conveying information.",
		"1.4.2":  "Provide a mechanism to pause, stop, or control audio that plays automatically.",
		"1.4.3":  "Ensure text has a contrast ratio of at least 4.5:1 (3:1 for large text).",
		"1.4.4":  "Allow text to be resized up to 200% without loss of content or functionality.",
		"1.4.5":  "Use text instead of images of text whenever possible.",
		"1.4.10": "Content reflows at 320px width without horizontal scrolling.",
		"1.4.11": "UI components and graphics have a contrast ratio of at least 3:1.",
		"1.4.12": "Allow text spacing adjustments without loss of content or functionality.",
		"1.4.13": "Content triggered by hover or focus is dismissible, hoverable, and persistent.",
		"2.1.1":  "Ensure all functionality is available via keyboard.",
		"2.1.2":  "Ensure keyboard focus can be moved away from any component.",
		"2.1.4":  "Allow users to remap or disable character key shortcuts.",
		"2.2.1":  "Allow users to adjust, extend, or disable time limits.",
		"2.2.2":  "Provide controls to pause, stop, or hide moving, blinking, or auto-updating content.",
		"2.3.1":  "Ensure content doesn't flash more than 3 times per second.",
		"2.4.1":  "Provide a mechanism to bypass repeated blocks of content.",
		"2.4.2":  "Provide descriptive page titles.",
		"2.4.3":  "Ensure focus order preserves meaning and operability.",
		"2.4.4":  "Make link purpose determinable from the link text or context.",
		"2.4.5":  "Provide multiple ways to locate pages within a set of pages.",
		"2.4.6":  "Provide headings and labels that describe topic or purpose.",
		"2.4.7":  "Ensure keyboard focus is visible on interactive elements.",
		"2.4.11": "Ensure focused elements are not entirely hidden by other content.",
		"2.5.1":  "Provide single-pointer alternatives for path-based gestures.",
		"2.5.2":  "Ensure pointer actions can be cancelled.",
		"2.5.3":  "Ensure the accessible name contains the visible label text.",
		"2.5.4":  "Provide alternatives for motion-triggered functionality.",
		"2.5.7":  "Provide single-pointer alternatives for drag operations.",
		"2.5.8":  "Ensure touch targets are at least 24x24 CSS pixels.",
		"3.1.1":  "Specify the language of the page using the lang attribute.",
		"3.1.2":  "Identify changes in language within the content.",
		"3.2.1":  "Don't change context when an element receives focus.",
		"3.2.2":  "Don't change context when an input receives input.",
		"3.2.3":  "Present navigation consistently across pages.",
		"3.2.4":  "Identify components with the same functionality consistently.",
		"3.2.6":  "Place help mechanisms in consistent locations.",
		"3.3.1":  "Identify and describe input errors in text.",
		"3.3.2":  "Provide labels or instructions for user input.",
		"3.3.3":  "Provide suggestions for correcting input errors when known.",
		"3.3.4":  "Make legal, financial, or data submissions reversible, checked, or confirmed.",
		"3.3.7":  "Don't require users to re-enter previously provided information.",
		"3.3.8":  "Don't require cognitive tests for authentication.",
		"4.1.1":  "Ensure markup is well-formed (unique IDs, no duplicate attributes).",
		"4.1.2":  "Expose name, role, and value for all UI components.",
		"4.1.3":  "Provide status messages that can be announced by assistive technology.",
	}

	if summary, ok := summaries[criterionID]; ok {
		return summary
	}
	return "Review and fix this accessibility issue."
}

// AttachRemediation adds remediation information to a finding in place.
func AttachRemediation(finding *types.Finding) {
	if finding.Remediation != nil {
		return // Already has remediation
	}

	axeRuleID := ""
	// Map common rule IDs to axe-core rule IDs
	axeMap := map[string]string{
		"image-alt":           "image-alt",
		"button-name":         "button-name",
		"link-name":           "link-name",
		"label":               "label",
		"color-contrast":      "color-contrast",
		"focus-visible":       "focus-order-semantics",
		"keyboard-trap":       "focus-trap",
		"keyboard-unreachable": "focusable-element",
		"tabindex-positive":   "tabindex",
		"reflow-horizontal-scroll": "scrollable-region-focusable",
		"target-size-minimum": "target-size",
		"text-spacing-loss":   "avoid-inline-spacing",
	}

	if mapped, ok := axeMap[finding.RuleID]; ok {
		axeRuleID = mapped
	}

	finding.Remediation = BuildForFinding(finding.SuccessCriteria, axeRuleID)
}
