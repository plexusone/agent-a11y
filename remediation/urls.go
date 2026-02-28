package remediation

import (
	"fmt"
	"strings"
)

// Base URLs for reference documentation.
const (
	// W3C WCAG 2.2 documentation
	W3CBaseURL           = "https://www.w3.org"
	WCAGTechniquesBase   = W3CBaseURL + "/WAI/WCAG22/Techniques"
	WCAGUnderstandingBase = W3CBaseURL + "/WAI/WCAG22/Understanding"
	WCAGQuickRefBase     = W3CBaseURL + "/WAI/WCAG22/quickref"
	ACTRulesBase         = W3CBaseURL + "/WAI/standards-guidelines/act/rules"

	// Deque axe-core documentation
	DequeBaseURL    = "https://dequeuniversity.com"
	AxeRulesBase    = DequeBaseURL + "/rules/axe/4.10"
	DequeTipsBase   = DequeBaseURL + "/tips"

	// WebAIM documentation
	WebAIMBaseURL = "https://webaim.org"
)

// TechniqueURL returns the W3C documentation URL for a WCAG technique.
//
// Example:
//
//	url := TechniqueURL(TechniqueH37) // https://www.w3.org/WAI/WCAG22/Techniques/html/H37
func TechniqueURL(id TechniqueID) string {
	category := techniqueCategory(id)
	return fmt.Sprintf("%s/%s/%s", WCAGTechniquesBase, category, id)
}

// TechniquesURL returns URLs for multiple techniques.
func TechniquesURL(ids ...TechniqueID) []string {
	urls := make([]string, len(ids))
	for i, id := range ids {
		urls[i] = TechniqueURL(id)
	}
	return urls
}

// UnderstandingURL returns the W3C "Understanding" document URL for a WCAG criterion.
//
// Example:
//
//	url := UnderstandingURL("1.1.1") // https://www.w3.org/WAI/WCAG22/Understanding/non-text-content.html
func UnderstandingURL(criterionID string) string {
	slug := criterionSlug(criterionID)
	return fmt.Sprintf("%s/%s.html", WCAGUnderstandingBase, slug)
}

// QuickRefURL returns the W3C Quick Reference URL for a WCAG criterion.
//
// Example:
//
//	url := QuickRefURL("1.1.1") // https://www.w3.org/WAI/WCAG22/quickref/#non-text-content
func QuickRefURL(criterionID string) string {
	slug := criterionSlug(criterionID)
	return fmt.Sprintf("%s/#%s", WCAGQuickRefBase, slug)
}

// ACTRuleURL returns the W3C ACT Rules documentation URL.
//
// Example:
//
//	url := ACTRuleURL(ACTImageHasAccessibleName) // https://www.w3.org/WAI/standards-guidelines/act/rules/23a2a8/
func ACTRuleURL(id ACTRuleID) string {
	return fmt.Sprintf("%s/%s/", ACTRulesBase, id)
}

// AxeRuleURL returns the Deque axe-core documentation URL for a rule.
//
// Example:
//
//	url := AxeRuleURL("image-alt") // https://dequeuniversity.com/rules/axe/4.10/image-alt
func AxeRuleURL(ruleID string) string {
	return fmt.Sprintf("%s/%s", AxeRulesBase, ruleID)
}

// WebAIMChecklistURL returns the WebAIM WCAG checklist URL for a criterion.
//
// Example:
//
//	url := WebAIMChecklistURL("1.1.1") // https://webaim.org/standards/wcag/checklist#sc1.1.1
func WebAIMChecklistURL(criterionID string) string {
	// WebAIM uses format like "sc1.1.1"
	return fmt.Sprintf("%s/standards/wcag/checklist#sc%s", WebAIMBaseURL, criterionID)
}

// Reference contains all reference URLs for a finding.
type Reference struct {
	// WCAG criterion references
	CriterionID    string `json:"criterionId"`
	UnderstandingURL string `json:"understandingUrl"`
	QuickRefURL    string `json:"quickRefUrl"`
	WebAIMURL      string `json:"webaimUrl"`

	// Technique references
	TechniqueURLs []string `json:"techniqueUrls,omitempty"`

	// ACT Rule reference
	ACTRuleID  ACTRuleID `json:"actRuleId,omitempty"`
	ACTRuleURL string    `json:"actRuleUrl,omitempty"`

	// axe-core reference
	AxeRuleID  string `json:"axeRuleId,omitempty"`
	AxeRuleURL string `json:"axeRuleUrl,omitempty"`
}

// BuildReference creates a complete Reference for a WCAG criterion.
func BuildReference(criterionID string, techniques []TechniqueID, actRuleID ACTRuleID, axeRuleID string) Reference {
	ref := Reference{
		CriterionID:      criterionID,
		UnderstandingURL: UnderstandingURL(criterionID),
		QuickRefURL:      QuickRefURL(criterionID),
		WebAIMURL:        WebAIMChecklistURL(criterionID),
	}

	if len(techniques) > 0 {
		ref.TechniqueURLs = TechniquesURL(techniques...)
	}

	if actRuleID != "" {
		ref.ACTRuleID = actRuleID
		ref.ACTRuleURL = ACTRuleURL(actRuleID)
	}

	if axeRuleID != "" {
		ref.AxeRuleID = axeRuleID
		ref.AxeRuleURL = AxeRuleURL(axeRuleID)
	}

	return ref
}

// techniqueCategory returns the URL path category for a technique ID.
func techniqueCategory(id TechniqueID) string {
	s := string(id)
	switch {
	case strings.HasPrefix(s, "G"):
		return "general"
	case strings.HasPrefix(s, "H"):
		return "html"
	case strings.HasPrefix(s, "C"):
		return "css"
	case strings.HasPrefix(s, "ARIA"):
		return "aria"
	case strings.HasPrefix(s, "SCR"):
		return "client-side-script"
	case strings.HasPrefix(s, "SVR"):
		return "server-side-script"
	case strings.HasPrefix(s, "SM"):
		return "smil"
	case strings.HasPrefix(s, "T"):
		return "text"
	case strings.HasPrefix(s, "PDF"):
		return "pdf"
	case strings.HasPrefix(s, "F"):
		return "failures"
	default:
		return "general"
	}
}

// criterionSlug returns the URL slug for a WCAG criterion.
var criterionSlugs = map[string]string{
	// Principle 1: Perceivable
	"1.1.1": "non-text-content",
	"1.2.1": "audio-only-and-video-only-prerecorded",
	"1.2.2": "captions-prerecorded",
	"1.2.3": "audio-description-or-media-alternative-prerecorded",
	"1.2.4": "captions-live",
	"1.2.5": "audio-description-prerecorded",
	"1.2.6": "sign-language-prerecorded",
	"1.2.7": "extended-audio-description-prerecorded",
	"1.2.8": "media-alternative-prerecorded",
	"1.2.9": "audio-only-live",
	"1.3.1": "info-and-relationships",
	"1.3.2": "meaningful-sequence",
	"1.3.3": "sensory-characteristics",
	"1.3.4": "orientation",
	"1.3.5": "identify-input-purpose",
	"1.3.6": "identify-purpose",
	"1.4.1": "use-of-color",
	"1.4.2": "audio-control",
	"1.4.3": "contrast-minimum",
	"1.4.4": "resize-text",
	"1.4.5": "images-of-text",
	"1.4.6": "contrast-enhanced",
	"1.4.7": "low-or-no-background-audio",
	"1.4.8": "visual-presentation",
	"1.4.9": "images-of-text-no-exception",
	"1.4.10": "reflow",
	"1.4.11": "non-text-contrast",
	"1.4.12": "text-spacing",
	"1.4.13": "content-on-hover-or-focus",

	// Principle 2: Operable
	"2.1.1": "keyboard",
	"2.1.2": "no-keyboard-trap",
	"2.1.3": "keyboard-no-exception",
	"2.1.4": "character-key-shortcuts",
	"2.2.1": "timing-adjustable",
	"2.2.2": "pause-stop-hide",
	"2.2.3": "no-timing",
	"2.2.4": "interruptions",
	"2.2.5": "re-authenticating",
	"2.2.6": "timeouts",
	"2.3.1": "three-flashes-or-below-threshold",
	"2.3.2": "three-flashes",
	"2.3.3": "animation-from-interactions",
	"2.4.1": "bypass-blocks",
	"2.4.2": "page-titled",
	"2.4.3": "focus-order",
	"2.4.4": "link-purpose-in-context",
	"2.4.5": "multiple-ways",
	"2.4.6": "headings-and-labels",
	"2.4.7": "focus-visible",
	"2.4.8": "location",
	"2.4.9": "link-purpose-link-only",
	"2.4.10": "section-headings",
	"2.4.11": "focus-not-obscured-minimum",
	"2.4.12": "focus-not-obscured-enhanced",
	"2.4.13": "focus-appearance",
	"2.5.1": "pointer-gestures",
	"2.5.2": "pointer-cancellation",
	"2.5.3": "label-in-name",
	"2.5.4": "motion-actuation",
	"2.5.5": "target-size-enhanced",
	"2.5.6": "concurrent-input-mechanisms",
	"2.5.7": "dragging-movements",
	"2.5.8": "target-size-minimum",

	// Principle 3: Understandable
	"3.1.1": "language-of-page",
	"3.1.2": "language-of-parts",
	"3.1.3": "unusual-words",
	"3.1.4": "abbreviations",
	"3.1.5": "reading-level",
	"3.1.6": "pronunciation",
	"3.2.1": "on-focus",
	"3.2.2": "on-input",
	"3.2.3": "consistent-navigation",
	"3.2.4": "consistent-identification",
	"3.2.5": "change-on-request",
	"3.2.6": "consistent-help",
	"3.3.1": "error-identification",
	"3.3.2": "labels-or-instructions",
	"3.3.3": "error-suggestion",
	"3.3.4": "error-prevention-legal-financial-data",
	"3.3.5": "help",
	"3.3.6": "error-prevention-all",
	"3.3.7": "redundant-entry",
	"3.3.8": "accessible-authentication-minimum",
	"3.3.9": "accessible-authentication-enhanced",

	// Principle 4: Robust
	"4.1.1": "parsing",
	"4.1.2": "name-role-value",
	"4.1.3": "status-messages",
}

func criterionSlug(criterionID string) string {
	if slug, ok := criterionSlugs[criterionID]; ok {
		return slug
	}
	// Fallback: return the criterion ID
	return criterionID
}
