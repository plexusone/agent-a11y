package specialized

import (
	"context"
	"encoding/json"
	"fmt"

	vibium "github.com/plexusone/vibium-go"
)

// SpacingTestResult contains results of text spacing testing.
type SpacingTestResult struct {
	// PassesSpacingTest indicates content remains readable with increased spacing.
	PassesSpacingTest bool `json:"passesSpacingTest"`

	// ContentLoss indicates text was cut off or hidden.
	ContentLoss bool `json:"contentLoss"`

	// OverlappingElements lists elements that overlap after spacing change.
	OverlappingElements []OverlapIssue `json:"overlappingElements,omitempty"`

	// ClippedElements lists elements with text clipped after spacing change.
	ClippedElements []ClipIssue `json:"clippedElements,omitempty"`

	// TestedElements is the count of text elements tested.
	TestedElements int `json:"testedElements"`
}

// OverlapIssue represents elements that overlap after spacing change.
type OverlapIssue struct {
	Selector1 string `json:"selector1"`
	Selector2 string `json:"selector2"`
}

// ClipIssue represents an element with clipped text.
type ClipIssue struct {
	Selector       string `json:"selector"`
	OriginalHeight float64 `json:"originalHeight"`
	NewHeight      float64 `json:"newHeight"`
	HasOverflow    bool    `json:"hasOverflow"`
}

// WCAG 1.4.12 required spacing values
const (
	LineHeight    = 1.5   // 1.5 times the font size
	ParagraphSpacing = 2.0 // 2 times the font size
	LetterSpacing = 0.12  // 0.12 times the font size
	WordSpacing   = 0.16  // 0.16 times the font size
)

// TestTextSpacing tests that content adapts to increased text spacing (WCAG 1.4.12).
func TestTextSpacing(ctx context.Context, vibe *vibium.Vibe) (*SpacingTestResult, error) {
	result := &SpacingTestResult{
		PassesSpacingTest: true,
	}

	// Get original state of text elements
	originalState, err := getTextElementsState(ctx, vibe)
	if err != nil {
		return nil, fmt.Errorf("failed to get original state: %w", err)
	}

	result.TestedElements = len(originalState)

	// Apply WCAG 1.4.12 spacing
	script := fmt.Sprintf(`
	const style = document.createElement('style');
	style.id = 'wcag-spacing-test';
	style.textContent = %q;
	document.head.appendChild(style);
	return true;
	`, fmt.Sprintf(`
		* {
			line-height: %f !important;
			letter-spacing: %fem !important;
			word-spacing: %fem !important;
		}
		p {
			margin-bottom: %fem !important;
		}
	`, LineHeight, LetterSpacing, WordSpacing, ParagraphSpacing))

	if _, err := vibe.Evaluate(ctx, script); err != nil {
		return nil, fmt.Errorf("failed to apply spacing: %w", err)
	}

	// Wait for reflow
	_, _ = vibe.Evaluate(ctx, "new Promise(r => setTimeout(r, 300))")

	// Get state after spacing change
	newState, err := getTextElementsState(ctx, vibe)
	if err != nil {
		// Clean up before returning error
		_, _ = vibe.Evaluate(ctx, "document.getElementById('wcag-spacing-test')?.remove()")
		return nil, fmt.Errorf("failed to get new state: %w", err)
	}

	// Check for issues
	for selector, orig := range originalState {
		if newEl, ok := newState[selector]; ok {
			// Check for overflow clipping
			if newEl.HasOverflow && !orig.HasOverflow {
				result.PassesSpacingTest = false
				result.ContentLoss = true
				result.ClippedElements = append(result.ClippedElements, ClipIssue{
					Selector:       selector,
					OriginalHeight: orig.Height,
					NewHeight:      newEl.Height,
					HasOverflow:    true,
				})
			}

			// Check for significant height reduction (text getting cut off)
			if newEl.Height < orig.Height*0.8 {
				result.PassesSpacingTest = false
				result.ContentLoss = true
				result.ClippedElements = append(result.ClippedElements, ClipIssue{
					Selector:       selector,
					OriginalHeight: orig.Height,
					NewHeight:      newEl.Height,
					HasOverflow:    newEl.HasOverflow,
				})
			}
		}
	}

	// Check for overlapping elements
	overlaps, err := checkForOverlaps(ctx, vibe)
	if err == nil && len(overlaps) > 0 {
		result.PassesSpacingTest = false
		result.OverlappingElements = overlaps
	}

	// Clean up - remove test styles
	_, _ = vibe.Evaluate(ctx, "document.getElementById('wcag-spacing-test')?.remove()")

	return result, nil
}

type elementState struct {
	Height      float64 `json:"height"`
	HasOverflow bool    `json:"hasOverflow"`
}

func getTextElementsState(ctx context.Context, vibe *vibium.Vibe) (map[string]elementState, error) {
	script := `
	const textElements = document.querySelectorAll('p, h1, h2, h3, h4, h5, h6, li, td, th, span, a, label, button');
	const states = {};

	textElements.forEach((el, i) => {
		const rect = el.getBoundingClientRect();
		const style = window.getComputedStyle(el);

		// Skip hidden elements
		if (rect.width === 0 || rect.height === 0 ||
			style.visibility === 'hidden' || style.display === 'none') {
			return;
		}

		const id = el.id ? '#' + el.id : '';
		const classes = el.className && typeof el.className === 'string'
			? '.' + el.className.split(' ').filter(c => c).join('.')
			: '';
		const selector = el.tagName.toLowerCase() + id + classes + '[' + i + ']';

		// Check for overflow
		const hasOverflow = el.scrollHeight > el.clientHeight ||
			el.scrollWidth > el.clientWidth ||
			style.overflow === 'hidden' ||
			style.textOverflow === 'ellipsis';

		states[selector] = {
			height: rect.height,
			hasOverflow: hasOverflow
		};
	});

	return JSON.stringify(states);
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	jsonStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	var data map[string]elementState
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse element states: %w", err)
	}

	return data, nil
}

func checkForOverlaps(ctx context.Context, vibe *vibium.Vibe) ([]OverlapIssue, error) {
	script := `
	const elements = document.querySelectorAll('p, h1, h2, h3, h4, h5, h6, li');
	const overlaps = [];

	for (let i = 0; i < elements.length; i++) {
		const rect1 = elements[i].getBoundingClientRect();
		if (rect1.width === 0 || rect1.height === 0) continue;

		for (let j = i + 1; j < elements.length; j++) {
			const rect2 = elements[j].getBoundingClientRect();
			if (rect2.width === 0 || rect2.height === 0) continue;

			// Check for overlap
			if (rect1.right > rect2.left && rect1.left < rect2.right &&
				rect1.bottom > rect2.top && rect1.top < rect2.bottom) {

				const id1 = elements[i].id ? '#' + elements[i].id : '';
				const id2 = elements[j].id ? '#' + elements[j].id : '';

				overlaps.push({
					selector1: elements[i].tagName.toLowerCase() + id1,
					selector2: elements[j].tagName.toLowerCase() + id2
				});
			}
		}

		// Limit to first 10 overlaps
		if (overlaps.length >= 10) break;
	}

	return JSON.stringify(overlaps);
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	jsonStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	var overlaps []OverlapIssue
	if err := json.Unmarshal([]byte(jsonStr), &overlaps); err != nil {
		return nil, fmt.Errorf("failed to parse overlaps: %w", err)
	}

	return overlaps, nil
}
