package specialized

import (
	"context"
	"encoding/json"
	"fmt"

	vibium "github.com/agentplexus/vibium-go"
)

// FlashTestResult contains results of flashing content testing (WCAG 2.3.1).
type FlashTestResult struct {
	// HasPotentialFlashing indicates content that may flash was detected.
	HasPotentialFlashing bool `json:"hasPotentialFlashing"`

	// FlashingElements lists elements with potential flashing.
	FlashingElements []FlashingElement `json:"flashingElements,omitempty"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// FlashingElement represents an element with potential flashing content.
type FlashingElement struct {
	Selector  string `json:"selector"`
	Type      string `json:"type"` // "gif", "video", "css-animation", "blink"
	RiskLevel string `json:"riskLevel"` // "high", "medium", "low"
}

// TestThreeFlashes tests for content that flashes more than 3 times per second (WCAG 2.3.1).
func TestThreeFlashes(ctx context.Context, vibe *vibium.Vibe) (*FlashTestResult, error) {
	script := `
	const result = {
		hasPotentialFlashing: false,
		flashingElements: []
	};

	// Check for blink elements (deprecated but dangerous)
	const blinks = document.querySelectorAll('blink');
	blinks.forEach(el => {
		result.hasPotentialFlashing = true;
		result.flashingElements.push({
			selector: 'blink',
			type: 'blink',
			riskLevel: 'high'
		});
	});

	// Check for animated GIFs (potential flashing)
	const images = document.querySelectorAll('img');
	images.forEach((img, i) => {
		const src = img.src || '';
		if (src.toLowerCase().endsWith('.gif')) {
			// GIFs could have flashing - flag for review
			const id = img.id ? '#' + img.id : '';
			result.flashingElements.push({
				selector: 'img' + id,
				type: 'gif',
				riskLevel: 'medium'
			});
		}
	});

	// Check for CSS animations with very fast iteration
	const allElements = document.querySelectorAll('*');
	allElements.forEach((el, i) => {
		const style = window.getComputedStyle(el);
		const animationDuration = parseFloat(style.animationDuration) || 0;
		const animationName = style.animationName;

		// Flag animations faster than 333ms (3 per second)
		if (animationName && animationName !== 'none' && animationDuration > 0 && animationDuration < 0.333) {
			result.hasPotentialFlashing = true;
			const id = el.id ? '#' + el.id : '';
			result.flashingElements.push({
				selector: el.tagName.toLowerCase() + id,
				type: 'css-animation',
				riskLevel: 'high'
			});
		}
	});

	// Check for video elements (could contain flashing)
	const videos = document.querySelectorAll('video');
	videos.forEach((video, i) => {
		const id = video.id ? '#' + video.id : '';
		result.flashingElements.push({
			selector: 'video' + id,
			type: 'video',
			riskLevel: 'low'
		});
	});

	// Limit results
	result.flashingElements = result.flashingElements.slice(0, 20);

	// Only flag as having potential flashing if high-risk items found
	result.hasPotentialFlashing = result.flashingElements.some(el => el.riskLevel === 'high');

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test flashing: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result FlashTestResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no high-risk flashing content detected
	result.PassesTest = !result.HasPotentialFlashing

	return &result, nil
}
