package remediation

// ACTRuleID represents a W3C Accessibility Conformance Testing rule identifier.
// ACT Rules are standardized test procedures for evaluating accessibility.
// See: https://www.w3.org/WAI/standards-guidelines/act/rules/
type ACTRuleID string

// ACTRule contains metadata about an ACT rule.
type ACTRule struct {
	ID          ACTRuleID
	Name        string
	Description string
	Criteria    []string // Related WCAG success criteria
	InputAspect string   // DOM Tree, CSS Styling, etc.
}

// ACT Rule IDs for WCAG 2.2 Level A and AA criteria.
// These are the standardized test procedures from W3C WAI.
const (
	// 1.1.1 Non-text Content
	ACTImageHasAccessibleName      ACTRuleID = "23a2a8" // Image has accessible name
	ACTImageButtonHasAccessibleName ACTRuleID = "59796f" // Image button has accessible name
	ACTSVGImageHasAccessibleName   ACTRuleID = "7d6734" // SVG image has accessible name
	ACTObjectHasAccessibleName     ACTRuleID = "8fc3b6" // Object element has accessible name

	// 1.2.1 Audio-only and Video-only (Prerecorded)
	ACTAudioOnlyHasAlternative  ACTRuleID = "afb423" // Audio-only has alternative
	ACTVideoOnlyHasAlternative  ACTRuleID = "c3232f" // Video-only has alternative

	// 1.2.2 Captions (Prerecorded)
	ACTVideoHasCaptions ACTRuleID = "f51b46" // Video element visual content has captions

	// 1.2.5 Audio Description (Prerecorded)
	ACTVideoHasAudioDescription ACTRuleID = "1ea59c" // Video element visual content has audio description

	// 1.3.1 Info and Relationships
	ACTHeadingsAreAccessiblyDescribed ACTRuleID = "b49b2e" // Headings are accessibly described
	ACTTableHeadersReferValidCells    ACTRuleID = "a25f45" // Table headers reference valid cells
	ACTDataTableHasAccessibleName     ACTRuleID = "e7aa44" // Data table has accessible name
	ACTARIARequiredOwned              ACTRuleID = "bc4a75" // Element with role has required owned elements
	ACTARIARequiredContext            ACTRuleID = "ff89c9" // Element with role is in required context

	// 1.3.2 Meaningful Sequence
	ACTVisibleFocusOrder ACTRuleID = "akn7bn" // Visible focus order matches DOM order

	// 1.3.4 Orientation
	ACTContentNotRestrictedToOrientation ACTRuleID = "b33eff" // Content not restricted to orientation

	// 1.3.5 Identify Input Purpose
	ACTAutocompleteValidValue ACTRuleID = "73f2c2" // Autocomplete attribute has valid value

	// 1.4.1 Use of Color
	ACTColorNotOnlyMeans ACTRuleID = "b28dc7" // Text color is not used as only visual means

	// 1.4.2 Audio Control
	ACTAutoPlayingAudioHasControl ACTRuleID = "80f0bf" // Auto-playing audio has control

	// 1.4.3 Contrast (Minimum)
	ACTTextContrastMinimum ACTRuleID = "afw4f7" // Text has minimum contrast

	// 1.4.4 Resize Text
	ACTTextCanBeResized ACTRuleID = "59br37" // Text can be resized

	// 1.4.5 Images of Text
	ACTImageOfTextIsDecorative ACTRuleID = "qt1vmo" // Image of text is decorative

	// 1.4.10 Reflow
	ACTContentReflowsWithoutLoss ACTRuleID = "b4f0c3" // Content reflows without loss

	// 1.4.11 Non-text Contrast
	ACTGraphicsContrastMinimum ACTRuleID = "09o5cg" // Graphics have minimum contrast

	// 1.4.12 Text Spacing
	ACTTextSpacingOverride ACTRuleID = "9e45ec" // Text spacing can be overridden

	// 1.4.13 Content on Hover or Focus
	ACTContentOnHoverDismissible ACTRuleID = "cae760" // Content on hover or focus is dismissible

	// 2.1.1 Keyboard
	ACTElementWithRoleIsKeyboardAccessible ACTRuleID = "0ssw9k" // Element with role is keyboard accessible
	ACTScrollableElementIsKeyboardAccessible ACTRuleID = "0sk6kp" // Scrollable element is keyboard accessible

	// 2.1.2 No Keyboard Trap
	ACTNoKeyboardTrap ACTRuleID = "80af7b" // No keyboard trap

	// 2.1.4 Character Key Shortcuts
	ACTCharacterKeyShortcutModifiable ACTRuleID = "ffbc54" // Character key shortcut modifiable

	// 2.2.1 Timing Adjustable
	ACTTimeoutAdjustable ACTRuleID = "b40fd1" // Timeout is adjustable

	// 2.2.2 Pause, Stop, Hide
	ACTMovingContentCanBePaused ACTRuleID = "b548d9" // Moving content can be paused

	// 2.3.1 Three Flashes or Below Threshold
	ACTNoFlashingContent ACTRuleID = "c249d5" // No flashing content

	// 2.4.1 Bypass Blocks
	ACTBypassBlockMechanism ACTRuleID = "cf77f2" // Page has bypass block mechanism
	ACTSkipLinkIsWorking    ACTRuleID = "ye5d6e" // Skip link is working

	// 2.4.2 Page Titled
	ACTPageHasTitle ACTRuleID = "2779a5" // Page has title

	// 2.4.3 Focus Order
	ACTFocusOrderMeaningful ACTRuleID = "oj04fd" // Focus order is meaningful

	// 2.4.4 Link Purpose (In Context)
	ACTLinkHasAccessibleName ACTRuleID = "c487ae" // Link has accessible name

	// 2.4.5 Multiple Ways
	ACTMultipleWaysToLocatePage ACTRuleID = "b4kp2t" // Multiple ways to locate page

	// 2.4.6 Headings and Labels
	ACTHeadingIsDescriptive ACTRuleID = "b49b2e" // Heading is descriptive

	// 2.4.7 Focus Visible
	ACTFocusIndicatorVisible ACTRuleID = "oj04fd" // Focus indicator is visible

	// 2.4.11 Focus Not Obscured (Minimum)
	ACTFocusNotObscured ACTRuleID = "focus-obscured" // Focus not obscured by content

	// 2.5.1 Pointer Gestures
	ACTPointerGestureAlternative ACTRuleID = "e86pn5" // Pointer gesture has alternative

	// 2.5.2 Pointer Cancellation
	ACTPointerCancellable ACTRuleID = "a1b64e" // Pointer action cancellable

	// 2.5.3 Label in Name
	ACTLabelInName ACTRuleID = "2ee8b8" // Accessible name contains visible label

	// 2.5.4 Motion Actuation
	ACTMotionActuationAlternative ACTRuleID = "c9f1d3" // Motion actuation has alternative

	// 2.5.7 Dragging Movements
	ACTDraggingAlternative ACTRuleID = "drag01" // Dragging has single pointer alternative

	// 2.5.8 Target Size (Minimum)
	ACTTargetSizeMinimum ACTRuleID = "target-size" // Target meets minimum size

	// 3.1.1 Language of Page
	ACTPageHasLang     ACTRuleID = "b5c3f8" // HTML page has lang attribute
	ACTPageLangIsValid ACTRuleID = "bf051a" // HTML page lang attribute is valid

	// 3.1.2 Language of Parts
	ACTLangAttributeValid ACTRuleID = "de46e4" // Element with lang has valid value

	// 3.2.1 On Focus
	ACTNoContextChangeOnFocus ACTRuleID = "focus-change" // No context change on focus

	// 3.2.2 On Input
	ACTNoContextChangeOnInput ACTRuleID = "input-change" // No context change on input

	// 3.2.6 Consistent Help
	ACTHelpLocationConsistent ACTRuleID = "help-consistent" // Help in consistent location

	// 3.3.1 Error Identification
	ACTErrorIdentified ACTRuleID = "36b590" // Error is identified

	// 3.3.2 Labels or Instructions
	ACTFormFieldHasLabel ACTRuleID = "e086e5" // Form field has label

	// 3.3.3 Error Suggestion
	ACTErrorSuggestionProvided ACTRuleID = "error-suggestion" // Error suggestion provided

	// 3.3.7 Redundant Entry
	ACTRedundantEntryMinimized ACTRuleID = "redundant-entry" // Redundant entry minimized

	// 3.3.8 Accessible Authentication
	ACTAuthenticationAccessible ACTRuleID = "auth-accessible" // Authentication is accessible

	// 4.1.1 Parsing
	ACTUniqueIDs               ACTRuleID = "3ea0c8" // ID attributes are unique
	ACTAttributeNotDuplicated ACTRuleID = "e6952f" // Attribute not duplicated

	// 4.1.2 Name, Role, Value
	ACTARIARoleValid           ACTRuleID = "674b10" // ARIA role is valid
	ACTARIAStatePropertyValid  ACTRuleID = "6a7281" // ARIA state/property valid
	ACTARIAHiddenNoFocus       ACTRuleID = "6cfa84" // aria-hidden not on focusable
	ACTButtonHasAccessibleName ACTRuleID = "97a4e1" // Button has accessible name
	ACTFormFieldAccessibleName ACTRuleID = "e086e5" // Form field has accessible name
	ACTFrameHasAccessibleName  ACTRuleID = "cae760" // Frame has accessible name
	ACTIFrameHasAccessibleName ACTRuleID = "4b1c6c" // Iframe has accessible name

	// 4.1.3 Status Messages
	ACTStatusMessageUsingRole ACTRuleID = "status-role" // Status message uses role
)

// actRules maps ACT rule IDs to their metadata.
var actRules = map[ACTRuleID]ACTRule{
	ACTImageHasAccessibleName: {
		ID:          ACTImageHasAccessibleName,
		Name:        "Image has accessible name",
		Description: "This rule checks that each image that is not marked as decorative has an accessible name.",
		Criteria:    []string{"1.1.1"},
		InputAspect: "DOM Tree",
	},
	ACTImageButtonHasAccessibleName: {
		ID:          ACTImageButtonHasAccessibleName,
		Name:        "Image button has accessible name",
		Description: "This rule checks that each image button element has an accessible name.",
		Criteria:    []string{"1.1.1", "4.1.2"},
		InputAspect: "DOM Tree",
	},
	ACTLinkHasAccessibleName: {
		ID:          ACTLinkHasAccessibleName,
		Name:        "Link has accessible name",
		Description: "This rule checks that each link has an accessible name.",
		Criteria:    []string{"2.4.4", "4.1.2"},
		InputAspect: "DOM Tree",
	},
	ACTButtonHasAccessibleName: {
		ID:          ACTButtonHasAccessibleName,
		Name:        "Button has accessible name",
		Description: "This rule checks that each button element has an accessible name.",
		Criteria:    []string{"4.1.2"},
		InputAspect: "DOM Tree",
	},
	ACTPageHasTitle: {
		ID:          ACTPageHasTitle,
		Name:        "Page has title",
		Description: "This rule checks that an HTML page has a non-empty title.",
		Criteria:    []string{"2.4.2"},
		InputAspect: "DOM Tree",
	},
	ACTPageHasLang: {
		ID:          ACTPageHasLang,
		Name:        "HTML page has lang attribute",
		Description: "This rule checks that an HTML page has a lang attribute.",
		Criteria:    []string{"3.1.1"},
		InputAspect: "DOM Tree",
	},
	ACTNoKeyboardTrap: {
		ID:          ACTNoKeyboardTrap,
		Name:        "No keyboard trap",
		Description: "This rule checks that there is no keyboard trap.",
		Criteria:    []string{"2.1.2"},
		InputAspect: "DOM Tree",
	},
	ACTFocusIndicatorVisible: {
		ID:          ACTFocusIndicatorVisible,
		Name:        "Focus indicator is visible",
		Description: "This rule checks that focusable elements have a visible focus indicator.",
		Criteria:    []string{"2.4.7"},
		InputAspect: "CSS Styling",
	},
	ACTTextContrastMinimum: {
		ID:          ACTTextContrastMinimum,
		Name:        "Text has minimum contrast",
		Description: "This rule checks that text has sufficient contrast against its background.",
		Criteria:    []string{"1.4.3"},
		InputAspect: "CSS Styling",
	},
	ACTFormFieldHasLabel: {
		ID:          ACTFormFieldHasLabel,
		Name:        "Form field has label",
		Description: "This rule checks that each form field has a label.",
		Criteria:    []string{"1.3.1", "3.3.2", "4.1.2"},
		InputAspect: "DOM Tree",
	},
	ACTUniqueIDs: {
		ID:          ACTUniqueIDs,
		Name:        "ID attributes are unique",
		Description: "This rule checks that all id attribute values on a page are unique.",
		Criteria:    []string{"4.1.1"},
		InputAspect: "DOM Tree",
	},
	ACTLabelInName: {
		ID:          ACTLabelInName,
		Name:        "Accessible name contains visible label",
		Description: "This rule checks that the accessible name contains the visible label text.",
		Criteria:    []string{"2.5.3"},
		InputAspect: "DOM Tree",
	},
	ACTBypassBlockMechanism: {
		ID:          ACTBypassBlockMechanism,
		Name:        "Page has bypass block mechanism",
		Description: "This rule checks that pages provide a mechanism to bypass repeated blocks of content.",
		Criteria:    []string{"2.4.1"},
		InputAspect: "DOM Tree",
	},
}

// GetACTRule returns metadata for an ACT rule ID.
func GetACTRule(id ACTRuleID) (ACTRule, bool) {
	rule, ok := actRules[id]
	return rule, ok
}

// ACTRulesForCriterion returns all ACT rules that test a specific WCAG criterion.
func ACTRulesForCriterion(criterionID string) []ACTRule {
	var rules []ACTRule
	for _, rule := range actRules {
		for _, c := range rule.Criteria {
			if c == criterionID {
				rules = append(rules, rule)
				break
			}
		}
	}
	return rules
}
