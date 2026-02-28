package specialized

import (
	"context"
	"encoding/json"
	"fmt"

	vibium "github.com/plexusone/vibium-go"
)

// CharacterShortcutResult contains results of character key shortcut testing (WCAG 2.1.4).
type CharacterShortcutResult struct {
	// HasSingleCharShortcuts indicates single character shortcuts were detected.
	HasSingleCharShortcuts bool `json:"hasSingleCharShortcuts"`

	// DetectedShortcuts lists detected single-key shortcuts.
	DetectedShortcuts []string `json:"detectedShortcuts,omitempty"`

	// HasRemapOption indicates shortcuts can be remapped.
	HasRemapOption bool `json:"hasRemapOption"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestCharacterKeyShortcuts tests for single character key shortcuts (WCAG 2.1.4).
func TestCharacterKeyShortcuts(ctx context.Context, vibe *vibium.Vibe) (*CharacterShortcutResult, error) {
	// This test checks for common patterns that indicate keyboard shortcuts
	script := `
	const result = {
		hasSingleCharShortcuts: false,
		detectedShortcuts: [],
		hasRemapOption: false
	};

	// Check for accesskey attributes (single character shortcuts)
	const accessKeyElements = document.querySelectorAll('[accesskey]');
	accessKeyElements.forEach(el => {
		const key = el.getAttribute('accesskey');
		if (key && key.length === 1) {
			result.hasSingleCharShortcuts = true;
			result.detectedShortcuts.push('accesskey: ' + key);
		}
	});

	// Check for keyboard shortcut documentation/settings
	const shortcutSettings = document.querySelectorAll(
		'[href*="keyboard"], [href*="shortcut"], ' +
		'[aria-label*="keyboard"], [aria-label*="shortcut"], ' +
		'a[href*="settings"], button[aria-label*="settings"]'
	);
	if (shortcutSettings.length > 0) {
		result.hasRemapOption = true;
	}

	// Check for common shortcut libraries patterns in scripts
	const scripts = document.querySelectorAll('script');
	scripts.forEach(script => {
		const text = script.textContent || '';
		if (text.includes('Mousetrap') || text.includes('hotkeys') ||
			text.includes('keyboardShortcuts') || text.includes('accesskey')) {
			result.hasSingleCharShortcuts = true;
		}
	});

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test character shortcuts: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result CharacterShortcutResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no single char shortcuts OR has remap option
	result.PassesTest = !result.HasSingleCharShortcuts || result.HasRemapOption

	return &result, nil
}

// PointerCancellationResult contains results of pointer cancellation testing (WCAG 2.5.2).
type PointerCancellationResult struct {
	// HasMousedownActions indicates elements with mousedown-only actions.
	HasMousedownActions bool `json:"hasMousedownActions"`

	// ProblematicElements lists elements that may have cancellation issues.
	ProblematicElements []string `json:"problematicElements,omitempty"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestPointerCancellation tests that pointer actions can be cancelled (WCAG 2.5.2).
func TestPointerCancellation(ctx context.Context, vibe *vibium.Vibe) (*PointerCancellationResult, error) {
	script := `
	const result = {
		hasMousedownActions: false,
		problematicElements: []
	};

	// Check for elements with onmousedown but no onclick
	const allElements = document.querySelectorAll('*');
	allElements.forEach((el, i) => {
		const hasMousedown = el.hasAttribute('onmousedown') || el.onmousedown;
		const hasClick = el.hasAttribute('onclick') || el.onclick;
		const hasTouchstart = el.hasAttribute('ontouchstart') || el.ontouchstart;

		// Flag if has mousedown/touchstart but no corresponding up/click event
		if ((hasMousedown || hasTouchstart) && !hasClick) {
			result.hasMousedownActions = true;
			const id = el.id ? '#' + el.id : '';
			const classes = el.className && typeof el.className === 'string'
				? '.' + el.className.split(' ').filter(c => c).slice(0, 2).join('.')
				: '';
			result.problematicElements.push(el.tagName.toLowerCase() + id + classes);
		}
	});

	// Limit results
	result.problematicElements = result.problematicElements.slice(0, 10);

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test pointer cancellation: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result PointerCancellationResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no problematic mousedown-only actions
	result.PassesTest = !result.HasMousedownActions

	return &result, nil
}

// OnInputResult contains results of on-input context change testing (WCAG 3.2.2).
type OnInputResult struct {
	// HasAutoSubmit indicates forms with auto-submit on input change.
	HasAutoSubmit bool `json:"hasAutoSubmit"`

	// ProblematicFields lists fields that may cause context changes.
	ProblematicFields []string `json:"problematicFields,omitempty"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestOnInput tests that input changes don't cause unexpected context changes (WCAG 3.2.2).
func TestOnInput(ctx context.Context, vibe *vibium.Vibe) (*OnInputResult, error) {
	script := `
	const result = {
		hasAutoSubmit: false,
		problematicFields: []
	};

	// Check for select elements with onchange that might submit
	const selects = document.querySelectorAll('select');
	selects.forEach((el, i) => {
		const onchange = el.getAttribute('onchange') || '';
		if (onchange.includes('submit') || onchange.includes('location') ||
			onchange.includes('href') || onchange.includes('navigate')) {
			result.hasAutoSubmit = true;
			const id = el.id ? '#' + el.id : '';
			const name = el.name ? '[name=' + el.name + ']' : '';
			result.problematicFields.push('select' + id + name);
		}
	});

	// Check for inputs with onchange/oninput that trigger navigation
	const inputs = document.querySelectorAll('input, textarea');
	inputs.forEach((el, i) => {
		const onchange = el.getAttribute('onchange') || '';
		const oninput = el.getAttribute('oninput') || '';
		const combined = onchange + oninput;

		if (combined.includes('submit') || combined.includes('location') ||
			combined.includes('href') || combined.includes('navigate')) {
			result.hasAutoSubmit = true;
			const id = el.id ? '#' + el.id : '';
			const name = el.name ? '[name=' + el.name + ']' : '';
			result.problematicFields.push(el.tagName.toLowerCase() + id + name);
		}
	});

	// Limit results
	result.problematicFields = result.problematicFields.slice(0, 10);

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test on-input: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result OnInputResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no auto-submit on input change
	result.PassesTest = !result.HasAutoSubmit

	return &result, nil
}

// RedundantEntryResult contains results of redundant entry testing (WCAG 3.3.7).
type RedundantEntryResult struct {
	// HasMultiStepForm indicates multi-step forms exist.
	HasMultiStepForm bool `json:"hasMultiStepForm"`

	// HasAutocomplete indicates autocomplete is enabled.
	HasAutocomplete bool `json:"hasAutocomplete"`

	// FormsWithoutAutocomplete lists forms missing autocomplete.
	FormsWithoutAutocomplete int `json:"formsWithoutAutocomplete"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestRedundantEntry tests that previously entered info is auto-populated (WCAG 3.3.7).
func TestRedundantEntry(ctx context.Context, vibe *vibium.Vibe) (*RedundantEntryResult, error) {
	script := `
	const result = {
		hasMultiStepForm: false,
		hasAutocomplete: false,
		formsWithoutAutocomplete: 0
	};

	// Check for multi-step form indicators
	const stepIndicators = document.querySelectorAll(
		'[class*="step"], [class*="wizard"], [class*="progress"], ' +
		'[data-step], [aria-label*="step"]'
	);
	if (stepIndicators.length > 0) {
		result.hasMultiStepForm = true;
	}

	// Check forms for autocomplete
	const forms = document.querySelectorAll('form');
	forms.forEach(form => {
		const inputs = form.querySelectorAll('input:not([type="hidden"]):not([type="submit"]):not([type="button"])');
		let hasAnyAutocomplete = false;

		inputs.forEach(input => {
			const autocomplete = input.getAttribute('autocomplete');
			if (autocomplete && autocomplete !== 'off') {
				hasAnyAutocomplete = true;
				result.hasAutocomplete = true;
			}
		});

		if (inputs.length > 0 && !hasAnyAutocomplete) {
			result.formsWithoutAutocomplete++;
		}
	});

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test redundant entry: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result RedundantEntryResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if has autocomplete enabled or no multi-step forms needing it
	result.PassesTest = result.HasAutocomplete || !result.HasMultiStepForm

	return &result, nil
}
