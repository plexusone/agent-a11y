package specialized

import (
	"context"
	"encoding/json"
	"fmt"

	vibium "github.com/agentplexus/vibium-go"
)

// ResizeTextResult contains results of text resize testing (WCAG 1.4.4).
type ResizeTextResult struct {
	// OriginalFontSizes captures font sizes before zoom.
	OriginalFontSizes int `json:"originalFontSizes"`

	// HasContentLoss indicates if content is lost at 200% zoom.
	HasContentLoss bool `json:"hasContentLoss"`

	// HasOverflow indicates if text overflows at 200% zoom.
	HasOverflow bool `json:"hasOverflow"`

	// ClippedElements lists elements with clipped text.
	ClippedElements []string `json:"clippedElements,omitempty"`

	// OverlappingElements lists elements that overlap at 200%.
	OverlappingElements []string `json:"overlappingElements,omitempty"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestResizeText tests that text can be resized up to 200% without loss (WCAG 1.4.4).
func TestResizeText(ctx context.Context, vibe *vibium.Vibe) (*ResizeTextResult, error) {
	// First, get baseline measurements at normal zoom
	baselineScript := `
	const textElements = document.querySelectorAll('p, span, div, li, td, th, h1, h2, h3, h4, h5, h6, label, a');
	const baseline = [];

	textElements.forEach((el, i) => {
		if (i >= 100) return; // Limit elements
		const rect = el.getBoundingClientRect();
		const style = window.getComputedStyle(el);
		const text = el.textContent?.trim() || '';

		if (text.length > 0 && rect.width > 0 && rect.height > 0) {
			baseline.push({
				index: i,
				width: rect.width,
				height: rect.height,
				fontSize: parseFloat(style.fontSize),
				overflow: style.overflow,
				textOverflow: style.textOverflow
			});
		}
	});

	return JSON.stringify({
		count: baseline.length,
		elements: baseline
	});
	`

	baselineRaw, err := vibe.Evaluate(ctx, baselineScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline: %w", err)
	}

	baselineStr, ok := baselineRaw.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected baseline type: %T", baselineRaw)
	}

	var baseline struct {
		Count    int `json:"count"`
		Elements []struct {
			Index        int     `json:"index"`
			Width        float64 `json:"width"`
			Height       float64 `json:"height"`
			FontSize     float64 `json:"fontSize"`
			Overflow     string  `json:"overflow"`
			TextOverflow string  `json:"textOverflow"`
		} `json:"elements"`
	}
	if err := json.Unmarshal([]byte(baselineStr), &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse baseline: %w", err)
	}

	// Apply 200% zoom using CSS transform on document element
	zoomScript := `
	document.documentElement.style.fontSize = '200%';

	// Force synchronous reflow by reading offsetHeight
	void document.body.offsetHeight;

	const textElements = document.querySelectorAll('p, span, div, li, td, th, h1, h2, h3, h4, h5, h6, label, a');
	const issues = {
		clipped: [],
		overlapping: [],
		hasOverflow: false,
		hasContentLoss: false
	};

	const rects = [];
	textElements.forEach((el, i) => {
		if (i >= 100) return;
		const rect = el.getBoundingClientRect();
		const style = window.getComputedStyle(el);
		const text = el.textContent?.trim() || '';

		if (text.length > 0 && rect.width > 0) {
			rects.push({ el, rect, index: i });

			// Check for clipping
			if (style.overflow === 'hidden' || style.textOverflow === 'ellipsis') {
				if (el.scrollWidth > el.clientWidth || el.scrollHeight > el.clientHeight) {
					issues.hasOverflow = true;
					const id = el.id ? '#' + el.id : '';
					issues.clipped.push(el.tagName.toLowerCase() + id);
				}
			}
		}
	});

	// Check for overlapping elements
	for (let i = 0; i < rects.length && i < 50; i++) {
		for (let j = i + 1; j < rects.length && j < 50; j++) {
			const a = rects[i].rect;
			const b = rects[j].rect;

			// Check if rectangles overlap significantly
			const overlapX = Math.max(0, Math.min(a.right, b.right) - Math.max(a.left, b.left));
			const overlapY = Math.max(0, Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top));
			const overlapArea = overlapX * overlapY;
			const minArea = Math.min(a.width * a.height, b.width * b.height);

			if (overlapArea > minArea * 0.5 && minArea > 100) {
				issues.overlapping.push(rects[i].el.tagName.toLowerCase());
				break;
			}
		}
	}

	// Limit results
	issues.clipped = issues.clipped.slice(0, 10);
	issues.overlapping = issues.overlapping.slice(0, 10);

	// Reset zoom
	document.documentElement.style.fontSize = '';

	return JSON.stringify(issues);
	`

	zoomRaw, err := vibe.Evaluate(ctx, zoomScript)
	if err != nil {
		return nil, fmt.Errorf("failed to test zoom: %w", err)
	}

	zoomStr, ok := zoomRaw.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected zoom result type: %T", zoomRaw)
	}

	var issues struct {
		Clipped        []string `json:"clipped"`
		Overlapping    []string `json:"overlapping"`
		HasOverflow    bool     `json:"hasOverflow"`
		HasContentLoss bool     `json:"hasContentLoss"`
	}
	if err := json.Unmarshal([]byte(zoomStr), &issues); err != nil {
		return nil, fmt.Errorf("failed to parse zoom results: %w", err)
	}

	result := &ResizeTextResult{
		OriginalFontSizes:   baseline.Count,
		HasContentLoss:      issues.HasContentLoss || len(issues.Clipped) > 5,
		HasOverflow:         issues.HasOverflow,
		ClippedElements:     issues.Clipped,
		OverlappingElements: issues.Overlapping,
	}

	// Passes if minimal clipping and no significant overlap
	result.PassesTest = len(issues.Clipped) <= 2 && len(issues.Overlapping) <= 2

	return result, nil
}
