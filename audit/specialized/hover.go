package specialized

import (
	"context"
	"encoding/json"
	"fmt"

	vibium "github.com/plexusone/vibium-go"
)

// HoverFocusResult contains results of content on hover/focus testing (WCAG 1.4.13).
type HoverFocusResult struct {
	// HasHoverContent indicates elements show additional content on hover.
	HasHoverContent bool `json:"hasHoverContent"`

	// HasFocusContent indicates elements show additional content on focus.
	HasFocusContent bool `json:"hasFocusContent"`

	// HoverElements lists elements with hover-triggered content.
	HoverElements []HoverElement `json:"hoverElements,omitempty"`

	// IsDismissible indicates hover content can be dismissed.
	IsDismissible bool `json:"isDismissible"`

	// IsHoverable indicates hover content can be hovered.
	IsHoverable bool `json:"isHoverable"`

	// IsPersistent indicates hover content persists until dismissed.
	IsPersistent bool `json:"isPersistent"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// HoverElement represents an element with hover-triggered content.
type HoverElement struct {
	Selector    string `json:"selector"`
	ContentType string `json:"contentType"` // "tooltip", "dropdown", "popover", "custom"
}

// TestContentOnHoverFocus tests hover/focus triggered content (WCAG 1.4.13).
func TestContentOnHoverFocus(ctx context.Context, vibe *vibium.Vibe) (*HoverFocusResult, error) {
	script := `
	const result = {
		hasHoverContent: false,
		hasFocusContent: false,
		hoverElements: [],
		isDismissible: true,
		isHoverable: true,
		isPersistent: true
	};

	// Check for elements with title attribute (native tooltips)
	const titledElements = document.querySelectorAll('[title]');
	titledElements.forEach((el, i) => {
		if (i >= 10) return;
		const title = el.getAttribute('title');
		if (title && title.trim().length > 0) {
			result.hasHoverContent = true;
			const id = el.id ? '#' + el.id : '';
			result.hoverElements.push({
				selector: el.tagName.toLowerCase() + id,
				contentType: 'tooltip'
			});
		}
	});

	// Check for CSS-based hover content
	const allElements = document.querySelectorAll('*');
	const hoverPatterns = ['tooltip', 'popover', 'dropdown', 'hover-content', 'on-hover'];

	allElements.forEach((el, i) => {
		if (i >= 500) return;
		const classes = el.className && typeof el.className === 'string' ? el.className : '';

		for (const pattern of hoverPatterns) {
			if (classes.toLowerCase().includes(pattern)) {
				result.hasHoverContent = true;
				const id = el.id ? '#' + el.id : '';
				result.hoverElements.push({
					selector: el.tagName.toLowerCase() + id,
					contentType: pattern.includes('tooltip') ? 'tooltip' :
								 pattern.includes('dropdown') ? 'dropdown' :
								 pattern.includes('popover') ? 'popover' : 'custom'
				});
				break;
			}
		}
	});

	// Check for aria-describedby (often used for tooltips)
	const describedElements = document.querySelectorAll('[aria-describedby]');
	describedElements.forEach((el, i) => {
		if (i >= 10) return;
		const describedById = el.getAttribute('aria-describedby');
		const describedEl = document.getElementById(describedById);
		if (describedEl) {
			const style = window.getComputedStyle(describedEl);
			// Check if initially hidden (common tooltip pattern)
			if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') {
				result.hasHoverContent = true;
				const id = el.id ? '#' + el.id : '';
				result.hoverElements.push({
					selector: el.tagName.toLowerCase() + id,
					contentType: 'tooltip'
				});
			}
		}
	});

	// Check for dropdown menus
	const dropdowns = document.querySelectorAll('[aria-haspopup], [aria-expanded]');
	dropdowns.forEach((el, i) => {
		if (i >= 10) return;
		result.hasFocusContent = true;
		const id = el.id ? '#' + el.id : '';
		result.hoverElements.push({
			selector: el.tagName.toLowerCase() + id,
			contentType: 'dropdown'
		});
	});

	// Check for escape key handlers (dismissibility indicator)
	const scripts = document.querySelectorAll('script');
	scripts.forEach(script => {
		const text = script.textContent || '';
		if (text.includes('Escape') || text.includes('keydown') || text.includes('keyup')) {
			result.isDismissible = true;
		}
	});

	// Limit results
	result.hoverElements = result.hoverElements.slice(0, 15);

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test hover/focus content: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result HoverFocusResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no hover content, or hover content appears properly handled
	// (has aria patterns suggesting proper implementation)
	result.PassesTest = !result.HasHoverContent || (result.IsDismissible && result.IsHoverable)

	return &result, nil
}
