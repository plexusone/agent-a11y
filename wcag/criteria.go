// Package wcag provides WCAG 2.2 criteria definitions and evaluation methods.
package wcag

// EvaluationMethod indicates how a criterion is evaluated.
type EvaluationMethod string

const (
	// MethodAutomated uses deterministic automated testing (axe-core).
	MethodAutomated EvaluationMethod = "Automated"

	// MethodSpecialized uses specialized automation (keyboard, focus, reflow).
	MethodSpecialized EvaluationMethod = "Specialized"

	// MethodLLMJudge uses LLM-as-a-Judge for semantic evaluation.
	MethodLLMJudge EvaluationMethod = "LLM-Judge"

	// MethodHybrid combines multiple methods.
	MethodHybrid EvaluationMethod = "Hybrid"

	// MethodManual requires human evaluation.
	MethodManual EvaluationMethod = "Manual"
)

// Conformance represents the conformance level for a criterion.
type Conformance string

const (
	ConformanceSupports          Conformance = "Supports"
	ConformancePartiallySupports Conformance = "Partially Supports"
	ConformanceDoesNotSupport    Conformance = "Does Not Support"
	ConformanceNotApplicable     Conformance = "Not Applicable"
	ConformanceNotEvaluated      Conformance = "Not Evaluated"
)

// Criterion defines a WCAG success criterion.
type Criterion struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Level       string           `json:"level"`
	Principle   string           `json:"principle"`
	Guideline   string           `json:"guideline"`
	Description string           `json:"description"`
	Method      EvaluationMethod `json:"method"`
	AxeRules    []string         `json:"axeRules,omitempty"`
}

// WCAG22AA contains all 50 WCAG 2.2 Level A and AA success criteria.
var WCAG22AA = []Criterion{
	// Principle 1: Perceivable
	// Guideline 1.1: Text Alternatives
	{ID: "1.1.1", Name: "Non-text Content", Level: "A", Principle: "Perceivable", Guideline: "Text Alternatives",
		Description: "All non-text content has a text alternative that serves the equivalent purpose.",
		Method: MethodHybrid, AxeRules: []string{"image-alt", "input-image-alt", "area-alt", "object-alt", "svg-img-alt"}},

	// Guideline 1.2: Time-based Media
	{ID: "1.2.1", Name: "Audio-only and Video-only (Prerecorded)", Level: "A", Principle: "Perceivable", Guideline: "Time-based Media",
		Description: "Prerecorded audio-only and video-only media have alternatives.",
		Method: MethodLLMJudge},
	{ID: "1.2.2", Name: "Captions (Prerecorded)", Level: "A", Principle: "Perceivable", Guideline: "Time-based Media",
		Description: "Captions are provided for all prerecorded audio content in synchronized media.",
		Method: MethodHybrid, AxeRules: []string{"video-caption"}},
	{ID: "1.2.3", Name: "Audio Description or Media Alternative (Prerecorded)", Level: "A", Principle: "Perceivable", Guideline: "Time-based Media",
		Description: "An alternative for time-based media or audio description is provided.",
		Method: MethodLLMJudge},
	{ID: "1.2.4", Name: "Captions (Live)", Level: "AA", Principle: "Perceivable", Guideline: "Time-based Media",
		Description: "Captions are provided for all live audio content in synchronized media.",
		Method: MethodSpecialized},
	{ID: "1.2.5", Name: "Audio Description (Prerecorded)", Level: "AA", Principle: "Perceivable", Guideline: "Time-based Media",
		Description: "Audio description is provided for all prerecorded video content.",
		Method: MethodLLMJudge},

	// Guideline 1.3: Adaptable
	{ID: "1.3.1", Name: "Info and Relationships", Level: "A", Principle: "Perceivable", Guideline: "Adaptable",
		Description: "Information, structure, and relationships conveyed through presentation can be programmatically determined.",
		Method: MethodAutomated, AxeRules: []string{"definition-list", "dlitem", "list", "listitem", "table-duplicate-name", "td-headers-attr", "th-has-data-cells"}},
	{ID: "1.3.2", Name: "Meaningful Sequence", Level: "A", Principle: "Perceivable", Guideline: "Adaptable",
		Description: "Correct reading sequence can be programmatically determined when sequence affects meaning.",
		Method: MethodLLMJudge},
	{ID: "1.3.3", Name: "Sensory Characteristics", Level: "A", Principle: "Perceivable", Guideline: "Adaptable",
		Description: "Instructions do not rely solely on sensory characteristics of components.",
		Method: MethodLLMJudge},
	{ID: "1.3.4", Name: "Orientation", Level: "AA", Principle: "Perceivable", Guideline: "Adaptable",
		Description: "Content does not restrict view and operation to a single display orientation.",
		Method: MethodAutomated, AxeRules: []string{"css-orientation-lock"}},
	{ID: "1.3.5", Name: "Identify Input Purpose", Level: "AA", Principle: "Perceivable", Guideline: "Adaptable",
		Description: "Input fields collecting user information have programmatically determinable purpose.",
		Method: MethodAutomated, AxeRules: []string{"autocomplete-valid"}},

	// Guideline 1.4: Distinguishable
	{ID: "1.4.1", Name: "Use of Color", Level: "A", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "Color is not the only visual means of conveying information.",
		Method: MethodLLMJudge, AxeRules: []string{"link-in-text-block"}},
	{ID: "1.4.2", Name: "Audio Control", Level: "A", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "Mechanism to pause, stop, or control volume of audio that plays automatically.",
		Method: MethodSpecialized, AxeRules: []string{"no-autoplay-audio"}},
	{ID: "1.4.3", Name: "Contrast (Minimum)", Level: "AA", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "Text has a contrast ratio of at least 4.5:1 (3:1 for large text).",
		Method: MethodAutomated, AxeRules: []string{"color-contrast"}},
	{ID: "1.4.4", Name: "Resize Text", Level: "AA", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "Text can be resized up to 200% without loss of content or functionality.",
		Method: MethodSpecialized, AxeRules: []string{"meta-viewport"}},
	{ID: "1.4.5", Name: "Images of Text", Level: "AA", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "If technologies can achieve visual presentation, text is used rather than images of text.",
		Method: MethodLLMJudge},
	{ID: "1.4.10", Name: "Reflow", Level: "AA", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "Content reflows without requiring scrolling in two dimensions at 320 CSS pixels.",
		Method: MethodSpecialized},
	{ID: "1.4.11", Name: "Non-text Contrast", Level: "AA", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "UI components and graphical objects have a contrast ratio of at least 3:1.",
		Method: MethodAutomated, AxeRules: []string{"color-contrast-enhanced"}},
	{ID: "1.4.12", Name: "Text Spacing", Level: "AA", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "No loss of content when text spacing is adjusted within WCAG requirements.",
		Method: MethodSpecialized},
	{ID: "1.4.13", Name: "Content on Hover or Focus", Level: "AA", Principle: "Perceivable", Guideline: "Distinguishable",
		Description: "Additional content triggered by hover/focus can be dismissed, hovered, and is persistent.",
		Method: MethodSpecialized},

	// Principle 2: Operable
	// Guideline 2.1: Keyboard Accessible
	{ID: "2.1.1", Name: "Keyboard", Level: "A", Principle: "Operable", Guideline: "Keyboard Accessible",
		Description: "All functionality is operable through a keyboard interface.",
		Method: MethodSpecialized, AxeRules: []string{"scrollable-region-focusable"}},
	{ID: "2.1.2", Name: "No Keyboard Trap", Level: "A", Principle: "Operable", Guideline: "Keyboard Accessible",
		Description: "Keyboard focus can be moved away from any component using only a keyboard.",
		Method: MethodSpecialized},
	{ID: "2.1.4", Name: "Character Key Shortcuts", Level: "A", Principle: "Operable", Guideline: "Keyboard Accessible",
		Description: "Single character key shortcuts can be turned off, remapped, or activated only on focus.",
		Method: MethodSpecialized},

	// Guideline 2.2: Enough Time
	{ID: "2.2.1", Name: "Timing Adjustable", Level: "A", Principle: "Operable", Guideline: "Enough Time",
		Description: "Users can turn off, adjust, or extend time limits.",
		Method: MethodSpecialized, AxeRules: []string{"meta-refresh"}},
	{ID: "2.2.2", Name: "Pause, Stop, Hide", Level: "A", Principle: "Operable", Guideline: "Enough Time",
		Description: "Moving, blinking, scrolling, or auto-updating content can be paused, stopped, or hidden.",
		Method: MethodSpecialized, AxeRules: []string{"blink", "marquee"}},

	// Guideline 2.3: Seizures and Physical Reactions
	{ID: "2.3.1", Name: "Three Flashes or Below Threshold", Level: "A", Principle: "Operable", Guideline: "Seizures",
		Description: "Content does not contain anything that flashes more than three times in one second.",
		Method: MethodSpecialized},

	// Guideline 2.4: Navigable
	{ID: "2.4.1", Name: "Bypass Blocks", Level: "A", Principle: "Operable", Guideline: "Navigable",
		Description: "A mechanism is available to bypass blocks of content repeated on multiple pages.",
		Method: MethodAutomated, AxeRules: []string{"bypass", "skip-link"}},
	{ID: "2.4.2", Name: "Page Titled", Level: "A", Principle: "Operable", Guideline: "Navigable",
		Description: "Web pages have titles that describe topic or purpose.",
		Method: MethodAutomated, AxeRules: []string{"document-title"}},
	{ID: "2.4.3", Name: "Focus Order", Level: "A", Principle: "Operable", Guideline: "Navigable",
		Description: "Focusable components receive focus in a sequence that preserves meaning.",
		Method: MethodSpecialized, AxeRules: []string{"tabindex"}},
	{ID: "2.4.4", Name: "Link Purpose (In Context)", Level: "A", Principle: "Operable", Guideline: "Navigable",
		Description: "Link purpose can be determined from link text or its context.",
		Method: MethodLLMJudge, AxeRules: []string{"link-name"}},
	{ID: "2.4.5", Name: "Multiple Ways", Level: "AA", Principle: "Operable", Guideline: "Navigable",
		Description: "More than one way to locate a page within a set of pages.",
		Method: MethodLLMJudge},
	{ID: "2.4.6", Name: "Headings and Labels", Level: "AA", Principle: "Operable", Guideline: "Navigable",
		Description: "Headings and labels describe topic or purpose.",
		Method: MethodLLMJudge, AxeRules: []string{"empty-heading"}},
	{ID: "2.4.7", Name: "Focus Visible", Level: "AA", Principle: "Operable", Guideline: "Navigable",
		Description: "Keyboard focus indicator is visible.",
		Method: MethodSpecialized},
	{ID: "2.4.11", Name: "Focus Not Obscured (Minimum)", Level: "AA", Principle: "Operable", Guideline: "Navigable",
		Description: "Focused component is not entirely hidden by author-created content.",
		Method: MethodSpecialized},

	// Guideline 2.5: Input Modalities
	{ID: "2.5.1", Name: "Pointer Gestures", Level: "A", Principle: "Operable", Guideline: "Input Modalities",
		Description: "All functionality using multipoint/path-based gestures can be operated with single pointer.",
		Method: MethodSpecialized},
	{ID: "2.5.2", Name: "Pointer Cancellation", Level: "A", Principle: "Operable", Guideline: "Input Modalities",
		Description: "Functions triggered by single pointer can be cancelled.",
		Method: MethodSpecialized},
	{ID: "2.5.3", Name: "Label in Name", Level: "A", Principle: "Operable", Guideline: "Input Modalities",
		Description: "Visible text label is contained in accessible name.",
		Method: MethodAutomated, AxeRules: []string{"label-content-name-mismatch"}},
	{ID: "2.5.4", Name: "Motion Actuation", Level: "A", Principle: "Operable", Guideline: "Input Modalities",
		Description: "Functions operable by device motion can be operated by UI and motion can be disabled.",
		Method: MethodSpecialized},
	{ID: "2.5.7", Name: "Dragging Movements", Level: "AA", Principle: "Operable", Guideline: "Input Modalities",
		Description: "Functionality using dragging can be achieved by single pointer without dragging.",
		Method: MethodLLMJudge},
	{ID: "2.5.8", Name: "Target Size (Minimum)", Level: "AA", Principle: "Operable", Guideline: "Input Modalities",
		Description: "Target size for pointer inputs is at least 24x24 CSS pixels.",
		Method: MethodSpecialized},

	// Principle 3: Understandable
	// Guideline 3.1: Readable
	{ID: "3.1.1", Name: "Language of Page", Level: "A", Principle: "Understandable", Guideline: "Readable",
		Description: "Default human language of each page can be programmatically determined.",
		Method: MethodAutomated, AxeRules: []string{"html-has-lang", "html-lang-valid"}},
	{ID: "3.1.2", Name: "Language of Parts", Level: "AA", Principle: "Understandable", Guideline: "Readable",
		Description: "Language of each passage or phrase can be programmatically determined.",
		Method: MethodAutomated, AxeRules: []string{"valid-lang"}},

	// Guideline 3.2: Predictable
	{ID: "3.2.1", Name: "On Focus", Level: "A", Principle: "Understandable", Guideline: "Predictable",
		Description: "Receiving focus does not initiate a change of context.",
		Method: MethodSpecialized},
	{ID: "3.2.2", Name: "On Input", Level: "A", Principle: "Understandable", Guideline: "Predictable",
		Description: "Changing a UI component does not automatically cause a change of context.",
		Method: MethodSpecialized, AxeRules: []string{"select-name"}},
	{ID: "3.2.3", Name: "Consistent Navigation", Level: "AA", Principle: "Understandable", Guideline: "Predictable",
		Description: "Navigation mechanisms that appear on multiple pages occur in the same relative order.",
		Method: MethodLLMJudge},
	{ID: "3.2.4", Name: "Consistent Identification", Level: "AA", Principle: "Understandable", Guideline: "Predictable",
		Description: "Components with same functionality within a set of pages are identified consistently.",
		Method: MethodLLMJudge},
	{ID: "3.2.6", Name: "Consistent Help", Level: "A", Principle: "Understandable", Guideline: "Predictable",
		Description: "Help mechanisms occur in same relative order across pages.",
		Method: MethodLLMJudge},

	// Guideline 3.3: Input Assistance
	{ID: "3.3.1", Name: "Error Identification", Level: "A", Principle: "Understandable", Guideline: "Input Assistance",
		Description: "If an input error is detected, the item and error are described in text.",
		Method: MethodLLMJudge},
	{ID: "3.3.2", Name: "Labels or Instructions", Level: "A", Principle: "Understandable", Guideline: "Input Assistance",
		Description: "Labels or instructions are provided when content requires user input.",
		Method: MethodAutomated, AxeRules: []string{"label", "form-field-multiple-labels"}},
	{ID: "3.3.3", Name: "Error Suggestion", Level: "AA", Principle: "Understandable", Guideline: "Input Assistance",
		Description: "If an input error is detected, suggestions are provided.",
		Method: MethodLLMJudge},
	{ID: "3.3.4", Name: "Error Prevention (Legal, Financial, Data)", Level: "AA", Principle: "Understandable", Guideline: "Input Assistance",
		Description: "Submissions that cause legal commitments can be reversed, checked, or confirmed.",
		Method: MethodLLMJudge},
	{ID: "3.3.7", Name: "Redundant Entry", Level: "A", Principle: "Understandable", Guideline: "Input Assistance",
		Description: "Information previously entered is auto-populated or available for selection.",
		Method: MethodSpecialized},
	{ID: "3.3.8", Name: "Accessible Authentication (Minimum)", Level: "AA", Principle: "Understandable", Guideline: "Input Assistance",
		Description: "Cognitive function test is not required for authentication unless alternative provided.",
		Method: MethodLLMJudge},

	// Principle 4: Robust
	// Guideline 4.1: Compatible
	{ID: "4.1.2", Name: "Name, Role, Value", Level: "A", Principle: "Robust", Guideline: "Compatible",
		Description: "Name and role can be programmatically determined; states and values can be set programmatically.",
		Method: MethodAutomated, AxeRules: []string{"aria-allowed-attr", "aria-hidden-body", "aria-required-attr", "aria-roles", "aria-valid-attr-value", "aria-valid-attr", "button-name", "input-button-name", "role-img-alt"}},
	{ID: "4.1.3", Name: "Status Messages", Level: "AA", Principle: "Robust", Guideline: "Compatible",
		Description: "Status messages can be programmatically determined through role or properties.",
		Method: MethodAutomated, AxeRules: []string{"aria-live-region-text"}},
}

// GetCriterion returns a criterion by ID.
func GetCriterion(id string) *Criterion {
	for i := range WCAG22AA {
		if WCAG22AA[i].ID == id {
			return &WCAG22AA[i]
		}
	}
	return nil
}

// GetCriteriaByMethod returns criteria filtered by evaluation method.
func GetCriteriaByMethod(method EvaluationMethod) []Criterion {
	var result []Criterion
	for _, c := range WCAG22AA {
		if c.Method == method {
			result = append(result, c)
		}
	}
	return result
}

// GetCriteriaByLevel returns criteria filtered by WCAG level.
func GetCriteriaByLevel(level string) []Criterion {
	var result []Criterion
	for _, c := range WCAG22AA {
		if c.Level == level {
			result = append(result, c)
		}
	}
	return result
}

// MethodCounts returns the count of criteria by evaluation method.
func MethodCounts() map[EvaluationMethod]int {
	counts := make(map[EvaluationMethod]int)
	for _, c := range WCAG22AA {
		counts[c.Method]++
	}
	return counts
}

// TotalCriteria returns the total number of WCAG 2.2 AA criteria.
func TotalCriteria() int {
	return len(WCAG22AA)
}
