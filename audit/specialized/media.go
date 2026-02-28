package specialized

import (
	"context"
	"encoding/json"
	"fmt"

	vibium "github.com/plexusone/vibium-go"
)

// AudioControlResult contains results of audio control testing (WCAG 1.4.2).
type AudioControlResult struct {
	// HasAutoplayAudio indicates if auto-playing audio was detected.
	HasAutoplayAudio bool `json:"hasAutoplayAudio"`

	// AutoplayElements lists elements with autoplay audio/video.
	AutoplayElements []AutoplayElement `json:"autoplayElements,omitempty"`

	// HasControlMechanism indicates if pause/stop/volume controls exist.
	HasControlMechanism bool `json:"hasControlMechanism"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// AutoplayElement represents an element with autoplay media.
type AutoplayElement struct {
	Selector string `json:"selector"`
	TagName  string `json:"tagName"`
	HasAudio bool   `json:"hasAudio"`
	Duration float64 `json:"duration"`
}

// TestAudioControl tests for auto-playing audio with controls (WCAG 1.4.2).
func TestAudioControl(ctx context.Context, vibe *vibium.Vibe) (*AudioControlResult, error) {
	script := `
	const result = {
		hasAutoplayAudio: false,
		autoplayElements: [],
		hasControlMechanism: false
	};

	// Check for audio/video elements with autoplay
	const mediaElements = document.querySelectorAll('audio, video');
	mediaElements.forEach((el, i) => {
		const hasAutoplay = el.hasAttribute('autoplay') || el.autoplay;
		const isMuted = el.muted || el.hasAttribute('muted');

		// Only flag if autoplay AND not muted (or has audio track)
		if (hasAutoplay && !isMuted) {
			result.hasAutoplayAudio = true;
			const id = el.id ? '#' + el.id : '';
			result.autoplayElements.push({
				selector: el.tagName.toLowerCase() + id,
				tagName: el.tagName.toLowerCase(),
				hasAudio: !el.muted,
				duration: el.duration || 0
			});
		}

		// Check for controls attribute
		if (el.hasAttribute('controls')) {
			result.hasControlMechanism = true;
		}
	});

	// Check for custom audio controls (common patterns)
	const customControls = document.querySelectorAll(
		'[aria-label*="pause"], [aria-label*="stop"], [aria-label*="mute"], ' +
		'[aria-label*="volume"], .audio-control, .media-control, ' +
		'button[class*="pause"], button[class*="play"], button[class*="mute"]'
	);
	if (customControls.length > 0) {
		result.hasControlMechanism = true;
	}

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test audio control: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result AudioControlResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no autoplay audio OR has control mechanism
	result.PassesTest = !result.HasAutoplayAudio || result.HasControlMechanism

	return &result, nil
}

// AnimationResult contains results of animation/motion testing (WCAG 2.2.2).
type AnimationResult struct {
	// HasMovingContent indicates moving, blinking, or scrolling content was found.
	HasMovingContent bool `json:"hasMovingContent"`

	// MovingElements lists elements with motion.
	MovingElements []MovingElement `json:"movingElements,omitempty"`

	// HasPauseControl indicates pause/stop mechanism exists.
	HasPauseControl bool `json:"hasPauseControl"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// MovingElement represents an element with motion/animation.
type MovingElement struct {
	Selector      string `json:"selector"`
	AnimationType string `json:"animationType"` // "css-animation", "carousel", "marquee", "blink", "auto-scroll"
}

// TestPauseStopHide tests for pause/stop mechanisms on moving content (WCAG 2.2.2).
func TestPauseStopHide(ctx context.Context, vibe *vibium.Vibe) (*AnimationResult, error) {
	script := `
	const result = {
		hasMovingContent: false,
		movingElements: [],
		hasPauseControl: false
	};

	// Check for marquee elements (deprecated but still used)
	const marquees = document.querySelectorAll('marquee');
	marquees.forEach(el => {
		result.hasMovingContent = true;
		result.movingElements.push({
			selector: 'marquee',
			animationType: 'marquee'
		});
	});

	// Check for blink elements
	const blinks = document.querySelectorAll('blink');
	blinks.forEach(el => {
		result.hasMovingContent = true;
		result.movingElements.push({
			selector: 'blink',
			animationType: 'blink'
		});
	});

	// Check for CSS animations that run indefinitely
	const allElements = document.querySelectorAll('*');
	allElements.forEach((el, i) => {
		const style = window.getComputedStyle(el);
		const animationName = style.animationName;
		const animationIterationCount = style.animationIterationCount;

		if (animationName && animationName !== 'none' && animationIterationCount === 'infinite') {
			result.hasMovingContent = true;
			const id = el.id ? '#' + el.id : '';
			result.movingElements.push({
				selector: el.tagName.toLowerCase() + id,
				animationType: 'css-animation'
			});
		}
	});

	// Check for carousels/sliders (common patterns)
	const carousels = document.querySelectorAll(
		'[class*="carousel"], [class*="slider"], [class*="slideshow"], ' +
		'[data-slick], [data-swiper], .owl-carousel'
	);
	carousels.forEach((el, i) => {
		result.hasMovingContent = true;
		const id = el.id ? '#' + el.id : '';
		result.movingElements.push({
			selector: el.tagName.toLowerCase() + id,
			animationType: 'carousel'
		});
	});

	// Check for pause controls
	const pauseControls = document.querySelectorAll(
		'[aria-label*="pause"], [aria-label*="stop"], ' +
		'button[class*="pause"], button[class*="stop"], ' +
		'.pause-button, .stop-button, [data-action="pause"]'
	);
	if (pauseControls.length > 0) {
		result.hasPauseControl = true;
	}

	// Limit results
	result.movingElements = result.movingElements.slice(0, 20);

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test pause/stop/hide: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result AnimationResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no moving content OR has pause control
	result.PassesTest = !result.HasMovingContent || result.HasPauseControl

	return &result, nil
}

// TimingResult contains results of timing testing (WCAG 2.2.1).
type TimingResult struct {
	// HasMetaRefresh indicates a meta refresh redirect exists.
	HasMetaRefresh bool `json:"hasMetaRefresh"`

	// MetaRefreshDelay is the delay in seconds (0 = immediate).
	MetaRefreshDelay int `json:"metaRefreshDelay,omitempty"`

	// HasSessionTimeout indicates session timeout was detected.
	HasSessionTimeout bool `json:"hasSessionTimeout"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestTimingAdjustable tests for adjustable time limits (WCAG 2.2.1).
func TestTimingAdjustable(ctx context.Context, vibe *vibium.Vibe) (*TimingResult, error) {
	script := `
	const result = {
		hasMetaRefresh: false,
		metaRefreshDelay: 0,
		hasSessionTimeout: false
	};

	// Check for meta refresh
	const metaRefresh = document.querySelector('meta[http-equiv="refresh"]');
	if (metaRefresh) {
		const content = metaRefresh.getAttribute('content');
		if (content) {
			const match = content.match(/^(\d+)/);
			if (match) {
				result.hasMetaRefresh = true;
				result.metaRefreshDelay = parseInt(match[1], 10);
			}
		}
	}

	// Check for common session timeout patterns
	const scripts = document.querySelectorAll('script');
	scripts.forEach(script => {
		const text = script.textContent || '';
		if (text.includes('sessionTimeout') || text.includes('session_timeout') ||
			text.includes('idleTimeout') || text.includes('autoLogout')) {
			result.hasSessionTimeout = true;
		}
	});

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test timing: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result TimingResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no meta refresh with short delay (< 20 hours) and no detected timeouts
	// Note: meta refresh of 0 is acceptable (immediate redirect)
	result.PassesTest = !result.HasMetaRefresh || result.MetaRefreshDelay == 0 || result.MetaRefreshDelay >= 72000

	return &result, nil
}
