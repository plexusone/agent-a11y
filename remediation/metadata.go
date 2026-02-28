package remediation

// techniques maps technique IDs to their metadata.
// This is a subset of the most commonly referenced techniques.
var techniques = map[TechniqueID]Technique{
	// Non-text Content (1.1.1)
	TechniqueG94: {
		ID:          TechniqueG94,
		Category:    CategoryGeneral,
		Title:       "Providing short text alternative for non-text content",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.1.1"},
		Description: "Provide a short text alternative that serves the same purpose as the non-text content.",
	},
	TechniqueG95: {
		ID:          TechniqueG95,
		Category:    CategoryGeneral,
		Title:       "Providing short text alternatives that provide a brief description of the non-text content",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.1.1"},
		Description: "Provide a brief description for non-text content that cannot be fully described in a short text alternative.",
	},
	TechniqueH37: {
		ID:          TechniqueH37,
		Category:    CategoryHTML,
		Title:       "Using alt attributes on img elements",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.1.1"},
		Description: "Use the alt attribute on img elements to provide text alternatives for images.",
	},
	TechniqueH67: {
		ID:          TechniqueH67,
		Category:    CategoryHTML,
		Title:       "Using null alt text and no title attribute on img elements for decorative images",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.1.1"},
		Description: "Mark decorative images with alt=\"\" so assistive technologies ignore them.",
	},
	TechniqueARIA6: {
		ID:          TechniqueARIA6,
		Category:    CategoryARIA,
		Title:       "Using aria-label to provide labels for objects",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.1.1", "4.1.2"},
		Description: "Use aria-label to provide an accessible name for elements.",
	},
	TechniqueARIA10: {
		ID:          TechniqueARIA10,
		Category:    CategoryARIA,
		Title:       "Using aria-labelledby to provide a text alternative for non-text content",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.1.1"},
		Description: "Use aria-labelledby to reference visible text as the alternative for non-text content.",
	},
	TechniqueF65: {
		ID:          TechniqueF65,
		Category:    CategoryFailure,
		Title:       "Failure of Success Criterion 1.1.1 due to omitting the alt attribute on img elements",
		Type:        TechniqueTypeFailure,
		Criteria:    []string{"1.1.1"},
		Description: "Images must have an alt attribute to provide a text alternative.",
	},

	// Captions (1.2.2)
	TechniqueG87: {
		ID:          TechniqueG87,
		Category:    CategoryGeneral,
		Title:       "Providing closed captions",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.2.2"},
		Description: "Provide closed captions for all prerecorded audio content in synchronized media.",
	},
	TechniqueG93: {
		ID:          TechniqueG93,
		Category:    CategoryGeneral,
		Title:       "Providing open (always visible) captions",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.2.2"},
		Description: "Provide open captions that are always visible as part of the video.",
	},
	TechniqueH95: {
		ID:          TechniqueH95,
		Category:    CategoryHTML,
		Title:       "Using the track element to provide captions",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.2.2"},
		Description: "Use the HTML5 track element with kind=\"captions\" to provide captions for video.",
	},

	// Contrast (1.4.3)
	TechniqueG18: {
		ID:          TechniqueG18,
		Category:    CategoryGeneral,
		Title:       "Ensuring that a contrast ratio of at least 4.5:1 exists between text and background",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.4.3"},
		Description: "Ensure text has a contrast ratio of at least 4.5:1 against its background.",
	},
	TechniqueG145: {
		ID:          TechniqueG145,
		Category:    CategoryGeneral,
		Title:       "Ensuring that a contrast ratio of at least 3:1 exists for large text",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.4.3"},
		Description: "Large text (18pt or 14pt bold) requires a contrast ratio of at least 3:1.",
	},

	// Reflow (1.4.10)
	TechniqueC31: {
		ID:          TechniqueC31,
		Category:    CategoryCSS,
		Title:       "Using CSS Flexbox to reflow content",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.4.10"},
		Description: "Use CSS Flexbox to create layouts that reflow at 320px width.",
	},
	TechniqueC32: {
		ID:          TechniqueC32,
		Category:    CategoryCSS,
		Title:       "Using media queries and grid CSS to reflow columns",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.4.10"},
		Description: "Use CSS Grid and media queries to reflow multi-column layouts.",
	},

	// Text Spacing (1.4.12)
	TechniqueC35: {
		ID:          TechniqueC35,
		Category:    CategoryCSS,
		Title:       "Allowing for text spacing without wrapping",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.4.12"},
		Description: "Ensure content adapts to increased text spacing without loss of content or functionality.",
	},
	TechniqueC36: {
		ID:          TechniqueC36,
		Category:    CategoryCSS,
		Title:       "Allowing for text spacing override",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.4.12"},
		Description: "Allow users to override text spacing properties without loss of content.",
	},

	// Keyboard (2.1.1)
	TechniqueG202: {
		ID:          TechniqueG202,
		Category:    CategoryGeneral,
		Title:       "Ensuring keyboard control for all functionality",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.1.1"},
		Description: "Ensure all functionality is operable through a keyboard interface.",
	},
	TechniqueG90: {
		ID:          TechniqueG90,
		Category:    CategoryGeneral,
		Title:       "Providing keyboard-triggered event handlers",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.1.1"},
		Description: "Provide keyboard event handlers in addition to mouse event handlers.",
	},
	TechniqueH91: {
		ID:          TechniqueH91,
		Category:    CategoryHTML,
		Title:       "Using HTML form controls and links",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.1.1", "4.1.2"},
		Description: "Use native HTML form controls and links for keyboard accessibility.",
	},

	// No Keyboard Trap (2.1.2)
	TechniqueG21: {
		ID:          TechniqueG21,
		Category:    CategoryGeneral,
		Title:       "Ensuring that users are not trapped in content",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.1.2"},
		Description: "Ensure users can navigate away from all content using standard keyboard navigation.",
	},

	// Focus Order (2.4.3)
	TechniqueG59: {
		ID:          TechniqueG59,
		Category:    CategoryGeneral,
		Title:       "Placing the interactive elements in an order that follows sequences and relationships within the content",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.4.3"},
		Description: "Organize interactive elements so focus order matches the logical reading order.",
	},
	TechniqueH4: {
		ID:          TechniqueH4,
		Category:    CategoryHTML,
		Title:       "Creating a logical tab order through links, form controls, and objects",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.4.3"},
		Description: "Ensure the DOM order creates a logical tab order without positive tabindex values.",
	},
	TechniqueC27: {
		ID:          TechniqueC27,
		Category:    CategoryCSS,
		Title:       "Making the DOM order match the visual order",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.4.3"},
		Description: "Ensure CSS positioning doesn't create a visual order that differs from DOM order.",
	},

	// Focus Visible (2.4.7)
	TechniqueG149: {
		ID:          TechniqueG149,
		Category:    CategoryGeneral,
		Title:       "Using user interface components that are highlighted by the user agent when they receive focus",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.4.7"},
		Description: "Use native controls that have built-in visible focus indicators.",
	},
	TechniqueG165: {
		ID:          TechniqueG165,
		Category:    CategoryGeneral,
		Title:       "Using the default focus indicator for the platform",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.4.7"},
		Description: "Allow the platform's default focus indicator to remain visible.",
	},
	TechniqueG195: {
		ID:          TechniqueG195,
		Category:    CategoryGeneral,
		Title:       "Using an author-supplied, highly visible focus indicator",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.4.7"},
		Description: "Provide a custom focus indicator that is highly visible.",
	},
	TechniqueC15: {
		ID:          TechniqueC15,
		Category:    CategoryCSS,
		Title:       "Using CSS to change the presentation of a user interface component when it receives focus",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.4.7"},
		Description: "Use CSS :focus to provide a visible focus indicator.",
	},
	TechniqueC40: {
		ID:          TechniqueC40,
		Category:    CategoryCSS,
		Title:       "Creating a two-color focus indicator to ensure sufficient contrast",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.4.7", "2.4.13"},
		Description: "Use a two-color focus indicator (outline + shadow) for visibility on any background.",
	},
	TechniqueF78: {
		ID:          TechniqueF78,
		Category:    CategoryFailure,
		Title:       "Failure due to styling element outlines and borders in a way that removes or renders non-visible the visual focus indicator",
		Type:        TechniqueTypeFailure,
		Criteria:    []string{"2.4.7"},
		Description: "Do not use outline:none or outline:0 without providing an alternative focus indicator.",
	},

	// Target Size (2.5.8)
	TechniqueC42: {
		ID:          TechniqueC42,
		Category:    CategoryCSS,
		Title:       "Using min-height and min-width to ensure sufficient target spacing",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"2.5.8"},
		Description: "Use CSS min-height and min-width to ensure touch targets are at least 24x24 CSS pixels.",
	},

	// On Focus (3.2.1)
	TechniqueG107: {
		ID:          TechniqueG107,
		Category:    CategoryGeneral,
		Title:       "Using \"activate\" rather than \"focus\" as a trigger for changes of context",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"3.2.1"},
		Description: "Do not trigger navigation or form submission when an element receives focus.",
	},
	TechniqueF9: {
		ID:          TechniqueF9,
		Category:    CategoryFailure,
		Title:       "Failure due to changing the context when the user removes focus from a form element",
		Type:        TechniqueTypeFailure,
		Criteria:    []string{"3.2.1"},
		Description: "Avoid triggering context changes when focus moves away from an element.",
	},

	// Error Identification (3.3.1)
	TechniqueG83: {
		ID:          TechniqueG83,
		Category:    CategoryGeneral,
		Title:       "Providing text descriptions to identify required fields that were not completed",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"3.3.1"},
		Description: "Provide clear text descriptions identifying which required fields are incomplete.",
	},
	TechniqueARIA21: {
		ID:          TechniqueARIA21,
		Category:    CategoryARIA,
		Title:       "Using aria-invalid to indicate an error field",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"3.3.1"},
		Description: "Use aria-invalid=\"true\" to indicate fields with errors.",
	},
	TechniqueARIA19: {
		ID:          TechniqueARIA19,
		Category:    CategoryARIA,
		Title:       "Using ARIA role=alert or live regions to identify errors",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"3.3.1"},
		Description: "Use role=\"alert\" or aria-live regions to announce errors to screen readers.",
	},

	// Labels (3.3.2)
	TechniqueG131: {
		ID:          TechniqueG131,
		Category:    CategoryGeneral,
		Title:       "Providing descriptive labels",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"3.3.2"},
		Description: "Provide visible, descriptive labels for form controls.",
	},
	TechniqueH44: {
		ID:          TechniqueH44,
		Category:    CategoryHTML,
		Title:       "Using label elements to associate text labels with form controls",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.3.1", "3.3.2", "4.1.2"},
		Description: "Use <label> elements with for/id to associate labels with form controls.",
	},
	TechniqueARIA16: {
		ID:          TechniqueARIA16,
		Category:    CategoryARIA,
		Title:       "Using aria-labelledby to provide a name for user interface controls",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"1.3.1", "4.1.2"},
		Description: "Use aria-labelledby to reference visible text as the accessible name.",
	},

	// Name, Role, Value (4.1.2)
	TechniqueG108: {
		ID:          TechniqueG108,
		Category:    CategoryGeneral,
		Title:       "Using markup features to expose the name and role",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"4.1.2"},
		Description: "Use standard markup to expose name and role to assistive technologies.",
	},
	TechniqueARIA4: {
		ID:          TechniqueARIA4,
		Category:    CategoryARIA,
		Title:       "Using a WAI-ARIA role to expose the role of a user interface component",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"4.1.2"},
		Description: "Use ARIA roles to indicate the purpose of custom UI components.",
	},
	TechniqueARIA5: {
		ID:          TechniqueARIA5,
		Category:    CategoryARIA,
		Title:       "Using WAI-ARIA state and property attributes to expose the state of a user interface component",
		Type:        TechniqueTypeSufficient,
		Criteria:    []string{"4.1.2"},
		Description: "Use ARIA states (aria-expanded, aria-checked, etc.) to convey component state.",
	},
}

// GetTechnique returns metadata for a technique ID.
func GetTechnique(id TechniqueID) (Technique, bool) {
	t, ok := techniques[id]
	return t, ok
}

// TechniquesForCriterion returns all techniques that apply to a specific WCAG criterion.
func TechniquesForCriterion(criterionID string) []Technique {
	var result []Technique
	for _, t := range techniques {
		for _, c := range t.Criteria {
			if c == criterionID {
				result = append(result, t)
				break
			}
		}
	}
	return result
}

// SufficientTechniques returns sufficient techniques for a criterion.
func SufficientTechniques(criterionID string) []Technique {
	var result []Technique
	for _, t := range TechniquesForCriterion(criterionID) {
		if t.Type == TechniqueTypeSufficient {
			result = append(result, t)
		}
	}
	return result
}

// FailureTechniques returns failure techniques for a criterion.
func FailureTechniques(criterionID string) []Technique {
	var result []Technique
	for _, t := range TechniquesForCriterion(criterionID) {
		if t.Type == TechniqueTypeFailure {
			result = append(result, t)
		}
	}
	return result
}
