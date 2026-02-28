package specialized

import (
	"context"
	"encoding/json"
	"fmt"

	vibium "github.com/agentplexus/vibium-go"
)

// ReflowTestResult contains results of reflow testing.
type ReflowTestResult struct {
	// NoHorizontalScroll indicates no horizontal scrolling at 320px width.
	NoHorizontalScroll bool `json:"noHorizontalScroll"`

	// HorizontalScrollWidth is the amount of horizontal overflow in pixels.
	HorizontalScrollWidth float64 `json:"horizontalScrollWidth,omitempty"`

	// ContentLoss indicates content was lost or hidden at 320px.
	ContentLoss bool `json:"contentLoss"`

	// LostElements lists elements that became hidden/invisible.
	LostElements []string `json:"lostElements,omitempty"`

	// OverflowingElements lists elements causing horizontal scroll.
	OverflowingElements []OverflowElement `json:"overflowingElements,omitempty"`

	// OriginalViewport is the viewport before testing.
	OriginalViewport Viewport `json:"originalViewport"`

	// TestedViewport is the narrow viewport used for testing.
	TestedViewport Viewport `json:"testedViewport"`
}

// OverflowElement represents an element causing horizontal overflow.
type OverflowElement struct {
	Selector string  `json:"selector"`
	Width    float64 `json:"width"`
	Overflow float64 `json:"overflow"`
}

// Viewport represents viewport dimensions.
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TestReflow tests content reflow at 320px width (WCAG 1.4.10).
func TestReflow(ctx context.Context, vibe *vibium.Vibe) (*ReflowTestResult, error) {
	result := &ReflowTestResult{
		NoHorizontalScroll: true,
		TestedViewport:     Viewport{Width: 320, Height: 480},
	}

	// Get original viewport
	origViewport, err := vibe.GetViewport(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get viewport: %w", err)
	}
	result.OriginalViewport = Viewport{
		Width:  origViewport.Width,
		Height: origViewport.Height,
	}

	// Get elements visible at current viewport
	originalElements, err := getVisibleElements(ctx, vibe)
	if err != nil {
		return nil, fmt.Errorf("failed to get visible elements: %w", err)
	}

	// Set narrow viewport (320px is the WCAG 2.1 requirement)
	if err := vibe.SetViewport(ctx, vibium.Viewport{Width: 320, Height: 480}); err != nil {
		return nil, fmt.Errorf("failed to set viewport: %w", err)
	}

	// Wait for reflow
	_, _ = vibe.Evaluate(ctx, "new Promise(r => setTimeout(r, 500))")

	// Check for horizontal scroll
	scrollInfo, err := getScrollInfo(ctx, vibe)
	if err != nil {
		// Restore viewport before returning error
		_ = vibe.SetViewport(ctx, vibium.Viewport{
			Width:  result.OriginalViewport.Width,
			Height: result.OriginalViewport.Height,
		})
		return nil, fmt.Errorf("failed to get scroll info: %w", err)
	}

	if scrollInfo.HasHorizontalScroll {
		result.NoHorizontalScroll = false
		result.HorizontalScrollWidth = scrollInfo.ScrollWidth - scrollInfo.ClientWidth
	}

	// Find elements causing overflow
	overflowing, err := findOverflowingElements(ctx, vibe)
	if err == nil {
		result.OverflowingElements = overflowing
	}

	// Check for content loss
	narrowElements, err := getVisibleElements(ctx, vibe)
	if err == nil {
		// Find elements that were visible before but not now
		narrowSet := make(map[string]bool)
		for _, el := range narrowElements {
			narrowSet[el] = true
		}

		for _, el := range originalElements {
			if !narrowSet[el] {
				result.ContentLoss = true
				result.LostElements = append(result.LostElements, el)
			}
		}
	}

	// Restore original viewport
	_ = vibe.SetViewport(ctx, vibium.Viewport{
		Width:  result.OriginalViewport.Width,
		Height: result.OriginalViewport.Height,
	})

	return result, nil
}

type scrollInfo struct {
	HasHorizontalScroll bool    `json:"hasHorizontalScroll"`
	ScrollWidth         float64 `json:"scrollWidth"`
	ClientWidth         float64 `json:"clientWidth"`
}

func getScrollInfo(ctx context.Context, vibe *vibium.Vibe) (*scrollInfo, error) {
	script := `
	const body = document.body;
	const html = document.documentElement;
	const scrollWidth = Math.max(body.scrollWidth, html.scrollWidth);
	const clientWidth = Math.max(body.clientWidth, html.clientWidth);
	return JSON.stringify({
		hasHorizontalScroll: scrollWidth > clientWidth,
		scrollWidth: scrollWidth,
		clientWidth: clientWidth
	});
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	jsonStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	info := &scrollInfo{}
	if err := json.Unmarshal([]byte(jsonStr), info); err != nil {
		return nil, fmt.Errorf("failed to parse scroll info: %w", err)
	}

	return info, nil
}

func getVisibleElements(ctx context.Context, vibe *vibium.Vibe) ([]string, error) {
	script := `
	const visible = [];
	const elements = document.querySelectorAll('*');
	elements.forEach((el, i) => {
		const rect = el.getBoundingClientRect();
		const style = window.getComputedStyle(el);
		if (rect.width > 0 && rect.height > 0 &&
			style.visibility !== 'hidden' &&
			style.display !== 'none' &&
			style.opacity !== '0') {
			const id = el.id ? '#' + el.id : '';
			visible.push(el.tagName.toLowerCase() + id);
		}
	});
	return visible.slice(0, 100); // Limit to prevent huge arrays
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	elements, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	visible := make([]string, 0, len(elements))
	for _, el := range elements {
		if s, ok := el.(string); ok {
			visible = append(visible, s)
		}
	}

	return visible, nil
}

func findOverflowingElements(ctx context.Context, vibe *vibium.Vibe) ([]OverflowElement, error) {
	script := `
	const viewportWidth = window.innerWidth;
	const overflowing = [];
	const elements = document.querySelectorAll('*');
	elements.forEach(el => {
		const rect = el.getBoundingClientRect();
		if (rect.right > viewportWidth) {
			const id = el.id ? '#' + el.id : '';
			const classes = el.className && typeof el.className === 'string'
				? '.' + el.className.split(' ').filter(c => c).join('.')
				: '';
			overflowing.push({
				selector: el.tagName.toLowerCase() + id + classes,
				width: rect.width,
				overflow: rect.right - viewportWidth
			});
		}
	});
	return overflowing.slice(0, 20); // Limit results
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	elements, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	overflowing := make([]OverflowElement, 0, len(elements))
	for _, el := range elements {
		if data, ok := el.(map[string]interface{}); ok {
			oe := OverflowElement{}
			if v, ok := data["selector"].(string); ok {
				oe.Selector = v
			}
			if v, ok := data["width"].(float64); ok {
				oe.Width = v
			}
			if v, ok := data["overflow"].(float64); ok {
				oe.Overflow = v
			}
			overflowing = append(overflowing, oe)
		}
	}

	return overflowing, nil
}
