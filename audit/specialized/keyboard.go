// Package automated provides specialized automated accessibility tests
// beyond what axe-core can detect.
package specialized

import (
	"context"
	"fmt"

	vibium "github.com/agentplexus/vibium-go"
)

// KeyboardTestResult contains results of keyboard accessibility testing.
type KeyboardTestResult struct {
	// AllElementsReachable indicates all interactive elements can be tabbed to.
	AllElementsReachable bool `json:"allElementsReachable"`

	// TrapDetected indicates a keyboard trap was found.
	TrapDetected bool `json:"trapDetected"`

	// TrapLocation is the selector where the trap was detected.
	TrapLocation string `json:"trapLocation,omitempty"`

	// UnreachableElements lists elements that couldn't be reached via Tab.
	UnreachableElements []string `json:"unreachableElements,omitempty"`

	// FocusOrder records the order elements received focus.
	FocusOrder []FocusedElement `json:"focusOrder"`

	// TabCount is the number of Tab presses needed to traverse the page.
	TabCount int `json:"tabCount"`
}

// FocusedElement records an element that received focus.
type FocusedElement struct {
	Index     int    `json:"index"`
	Selector  string `json:"selector"`
	TagName   string `json:"tagName"`
	Role      string `json:"role,omitempty"`
	Label     string `json:"label,omitempty"`
	TabIndex  int    `json:"tabIndex"`
	IsVisible bool   `json:"isVisible"`
}

// TestKeyboardAccessibility tests keyboard navigation (WCAG 2.1.1, 2.1.2).
func TestKeyboardAccessibility(ctx context.Context, vibe *vibium.Vibe, maxTabs int) (*KeyboardTestResult, error) {
	if maxTabs <= 0 {
		maxTabs = 100 // Default limit to prevent infinite loops
	}

	result := &KeyboardTestResult{
		AllElementsReachable: true,
		FocusOrder:           make([]FocusedElement, 0),
	}

	// Get all interactive elements on the page
	interactiveElements, err := getInteractiveElements(ctx, vibe)
	if err != nil {
		return nil, fmt.Errorf("failed to get interactive elements: %w", err)
	}

	// Track which elements we've seen
	seenSelectors := make(map[string]bool)
	var lastSelector string
	stuckCount := 0

	// Get keyboard controller
	kb, err := vibe.Keyboard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyboard: %w", err)
	}

	// Start by focusing the body
	_, err = vibe.Evaluate(ctx, "document.body.focus()")
	if err != nil {
		return nil, fmt.Errorf("failed to focus body: %w", err)
	}

	// Tab through the page
	for i := 0; i < maxTabs; i++ {
		// Press Tab
		if err := kb.Press(ctx, "Tab"); err != nil {
			return nil, fmt.Errorf("failed to press Tab: %w", err)
		}

		// Get currently focused element
		focused, err := getFocusedElement(ctx, vibe, i)
		if err != nil {
			continue // Skip if we can't get focus info
		}

		result.FocusOrder = append(result.FocusOrder, *focused)
		result.TabCount = i + 1

		// Check for keyboard trap
		if focused.Selector == lastSelector {
			stuckCount++
			if stuckCount >= 3 {
				result.TrapDetected = true
				result.TrapLocation = focused.Selector
				break
			}
		} else {
			stuckCount = 0
		}
		lastSelector = focused.Selector

		// Track seen elements
		seenSelectors[focused.Selector] = true

		// Check if we've looped back to the beginning
		if i > 0 && focused.Selector == result.FocusOrder[0].Selector {
			break // Completed full cycle
		}
	}

	// Check for unreachable elements
	for _, el := range interactiveElements {
		if !seenSelectors[el] {
			result.UnreachableElements = append(result.UnreachableElements, el)
			result.AllElementsReachable = false
		}
	}

	return result, nil
}

// getInteractiveElements returns selectors for all interactive elements.
func getInteractiveElements(ctx context.Context, vibe *vibium.Vibe) ([]string, error) {
	script := `
	const interactive = document.querySelectorAll(
		'a[href], button, input:not([type="hidden"]), select, textarea, ' +
		'[tabindex]:not([tabindex="-1"]), [contenteditable="true"], ' +
		'[role="button"], [role="link"], [role="checkbox"], [role="menuitem"]'
	);
	return Array.from(interactive).map((el, i) => {
		const id = el.id ? '#' + el.id : '';
		const classes = el.className ? '.' + el.className.split(' ').join('.') : '';
		return el.tagName.toLowerCase() + id + classes || '[index=' + i + ']';
	});
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	// Convert result to string slice
	elements, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	selectors := make([]string, 0, len(elements))
	for _, el := range elements {
		if s, ok := el.(string); ok {
			selectors = append(selectors, s)
		}
	}

	return selectors, nil
}

// getFocusedElement returns information about the currently focused element.
func getFocusedElement(ctx context.Context, vibe *vibium.Vibe, index int) (*FocusedElement, error) {
	script := `
	const el = document.activeElement;
	if (!el || el === document.body) return null;

	const rect = el.getBoundingClientRect();
	const isVisible = rect.width > 0 && rect.height > 0 &&
		window.getComputedStyle(el).visibility !== 'hidden' &&
		window.getComputedStyle(el).display !== 'none';

	const id = el.id ? '#' + el.id : '';
	const classes = el.className && typeof el.className === 'string'
		? '.' + el.className.split(' ').filter(c => c).join('.')
		: '';

	return {
		selector: el.tagName.toLowerCase() + id + classes,
		tagName: el.tagName.toLowerCase(),
		role: el.getAttribute('role') || '',
		label: el.getAttribute('aria-label') || el.innerText?.substring(0, 50) || '',
		tabIndex: el.tabIndex,
		isVisible: isVisible
	};
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("no element focused")
	}

	// Parse result
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	focused := &FocusedElement{Index: index}

	if v, ok := data["selector"].(string); ok {
		focused.Selector = v
	}
	if v, ok := data["tagName"].(string); ok {
		focused.TagName = v
	}
	if v, ok := data["role"].(string); ok {
		focused.Role = v
	}
	if v, ok := data["label"].(string); ok {
		focused.Label = v
	}
	if v, ok := data["tabIndex"].(float64); ok {
		focused.TabIndex = int(v)
	}
	if v, ok := data["isVisible"].(bool); ok {
		focused.IsVisible = v
	}

	return focused, nil
}
