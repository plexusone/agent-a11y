package specialized

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"

	vibium "github.com/agentplexus/vibium-go"
)

// FocusVisibilityResult contains results of focus visibility testing.
type FocusVisibilityResult struct {
	// AllFocusVisible indicates all focused elements had visible indicators.
	AllFocusVisible bool `json:"allFocusVisible"`

	// ElementsWithoutVisibleFocus lists elements lacking focus indicators.
	ElementsWithoutVisibleFocus []FocusVisibilityIssue `json:"elementsWithoutVisibleFocus,omitempty"`

	// TestedElements is the count of elements tested.
	TestedElements int `json:"testedElements"`
}

// FocusVisibilityIssue represents an element without visible focus.
type FocusVisibilityIssue struct {
	Selector   string `json:"selector"`
	TagName    string `json:"tagName"`
	Role       string `json:"role,omitempty"`
	Screenshot string `json:"screenshot,omitempty"` // Base64 of focused state
}

// TestFocusVisibility tests that focus indicators are visible (WCAG 2.4.7).
// It tabs through elements and compares screenshots to detect focus changes.
func TestFocusVisibility(ctx context.Context, vibe *vibium.Vibe, maxElements int) (*FocusVisibilityResult, error) {
	if maxElements <= 0 {
		maxElements = 50
	}

	result := &FocusVisibilityResult{
		AllFocusVisible: true,
	}

	kb, err := vibe.Keyboard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyboard: %w", err)
	}

	// Focus body first
	_, _ = vibe.Evaluate(ctx, "document.body.focus()")

	var lastScreenshot []byte
	var firstFocusedSelector string

	for i := 0; i < maxElements; i++ {
		// Take screenshot before Tab
		beforeScreenshot, err := vibe.Screenshot(ctx)
		if err != nil {
			continue
		}

		// Press Tab
		if err := kb.Press(ctx, "Tab"); err != nil {
			break
		}

		// Get focused element
		focused, err := getFocusedElement(ctx, vibe, i)
		if err != nil || focused == nil {
			continue
		}

		// Track first focused element for loop detection
		if i == 0 {
			firstFocusedSelector = focused.Selector
		} else if focused.Selector == firstFocusedSelector {
			// We've looped back to the beginning
			break
		}

		// Take screenshot after Tab
		afterScreenshot, err := vibe.Screenshot(ctx)
		if err != nil {
			continue
		}

		result.TestedElements++

		// Compare screenshots to detect visual focus change
		if lastScreenshot != nil {
			hasVisibleChange, err := hasSignificantChange(beforeScreenshot, afterScreenshot)
			if err != nil {
				continue
			}

			if !hasVisibleChange {
				result.AllFocusVisible = false
				result.ElementsWithoutVisibleFocus = append(result.ElementsWithoutVisibleFocus, FocusVisibilityIssue{
					Selector: focused.Selector,
					TagName:  focused.TagName,
					Role:     focused.Role,
				})
			}
		}

		lastScreenshot = afterScreenshot
	}

	return result, nil
}

// hasSignificantChange compares two screenshots to detect visual changes.
// Returns true if there's a noticeable difference (e.g., focus ring).
func hasSignificantChange(before, after []byte) (bool, error) {
	img1, err := png.Decode(bytes.NewReader(before))
	if err != nil {
		return false, err
	}

	img2, err := png.Decode(bytes.NewReader(after))
	if err != nil {
		return false, err
	}

	bounds := img1.Bounds()
	if bounds != img2.Bounds() {
		return true, nil // Different sizes = definitely changed
	}

	// Count significantly different pixels
	differentPixels := 0
	totalPixels := bounds.Dx() * bounds.Dy()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := img1.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()

			// Check if pixel changed significantly
			if colorDiff(r1, r2) > 1000 || colorDiff(g1, g2) > 1000 || colorDiff(b1, b2) > 1000 {
				differentPixels++
			}
		}
	}

	// Consider it a visible change if >0.1% of pixels changed
	// Focus rings typically affect a small but noticeable area
	changePercent := float64(differentPixels) / float64(totalPixels) * 100
	return changePercent > 0.1, nil
}

func colorDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

// Ensure image package is used
var _ image.Image

// FocusOrderResult contains results of focus order testing (WCAG 2.4.3).
type FocusOrderResult struct {
	// TotalFocusable is the count of focusable elements.
	TotalFocusable int `json:"totalFocusable"`

	// HasPositiveTabindex indicates elements with tabindex > 0 exist.
	HasPositiveTabindex bool `json:"hasPositiveTabindex"`

	// PositiveTabindexElements lists elements with tabindex > 0.
	PositiveTabindexElements []string `json:"positiveTabindexElements,omitempty"`

	// HasLogicalOrder indicates focus order follows visual/DOM order.
	HasLogicalOrder bool `json:"hasLogicalOrder"`

	// OutOfOrderElements lists elements that break logical order.
	OutOfOrderElements []string `json:"outOfOrderElements,omitempty"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestFocusOrder tests that focus order preserves meaning (WCAG 2.4.3).
func TestFocusOrder(ctx context.Context, vibe *vibium.Vibe) (*FocusOrderResult, error) {
	script := `
	const result = {
		totalFocusable: 0,
		hasPositiveTabindex: false,
		positiveTabindexElements: [],
		hasLogicalOrder: true,
		outOfOrderElements: []
	};

	// Get all focusable elements
	const focusable = document.querySelectorAll(
		'a[href], button, input:not([type="hidden"]), select, textarea, ' +
		'[tabindex]:not([tabindex="-1"]), [contenteditable="true"]'
	);

	result.totalFocusable = focusable.length;

	// Check for positive tabindex (disrupts natural order)
	focusable.forEach((el, i) => {
		const tabindex = el.getAttribute('tabindex');
		if (tabindex && parseInt(tabindex, 10) > 0) {
			result.hasPositiveTabindex = true;
			const id = el.id ? '#' + el.id : '';
			result.positiveTabindexElements.push(
				el.tagName.toLowerCase() + id + '[tabindex=' + tabindex + ']'
			);
		}
	});

	// Check if visual order matches DOM order
	// Get positions of focusable elements
	const positions = [];
	focusable.forEach((el, i) => {
		if (i >= 50) return;
		const rect = el.getBoundingClientRect();
		if (rect.width > 0 && rect.height > 0) {
			positions.push({
				index: i,
				top: rect.top,
				left: rect.left,
				el: el
			});
		}
	});

	// Sort by visual position (top to bottom, left to right)
	const visualOrder = [...positions].sort((a, b) => {
		const rowDiff = Math.floor(a.top / 50) - Math.floor(b.top / 50);
		if (rowDiff !== 0) return rowDiff;
		return a.left - b.left;
	});

	// Compare visual order to DOM order
	let orderIssues = 0;
	for (let i = 0; i < visualOrder.length - 1; i++) {
		const current = visualOrder[i];
		const next = visualOrder[i + 1];

		// If visual next comes before current in DOM, it's out of order
		if (next.index < current.index) {
			orderIssues++;
			if (result.outOfOrderElements.length < 5) {
				const id = current.el.id ? '#' + current.el.id : '';
				result.outOfOrderElements.push(current.el.tagName.toLowerCase() + id);
			}
		}
	}

	// Allow some variance (CSS layouts can legitimately reorder)
	result.hasLogicalOrder = orderIssues <= positions.length * 0.1;

	// Limit results
	result.positiveTabindexElements = result.positiveTabindexElements.slice(0, 10);
	result.outOfOrderElements = result.outOfOrderElements.slice(0, 10);

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test focus order: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result FocusOrderResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no positive tabindex and logical order is maintained
	result.PassesTest = !result.HasPositiveTabindex && result.HasLogicalOrder

	return &result, nil
}

// FocusObscuredResult contains results of focus obscured testing (WCAG 2.4.11).
type FocusObscuredResult struct {
	// TestedElements is the count of focusable elements tested.
	TestedElements int `json:"testedElements"`

	// HasStickyHeaders indicates sticky/fixed headers exist.
	HasStickyHeaders bool `json:"hasStickyHeaders"`

	// HasStickyFooters indicates sticky/fixed footers exist.
	HasStickyFooters bool `json:"hasStickyFooters"`

	// StickyElements lists sticky/fixed positioned elements.
	StickyElements []string `json:"stickyElements,omitempty"`

	// PotentiallyObscured lists elements that might be obscured.
	PotentiallyObscured []string `json:"potentiallyObscured,omitempty"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestFocusNotObscured tests that focused elements aren't hidden (WCAG 2.4.11).
func TestFocusNotObscured(ctx context.Context, vibe *vibium.Vibe) (*FocusObscuredResult, error) {
	script := `
	const result = {
		testedElements: 0,
		hasStickyHeaders: false,
		hasStickyFooters: false,
		stickyElements: [],
		potentiallyObscured: []
	};

	// Find sticky/fixed elements that could obscure focus
	const allElements = document.querySelectorAll('*');
	const stickyFixedElements = [];

	allElements.forEach(el => {
		const style = window.getComputedStyle(el);
		if (style.position === 'fixed' || style.position === 'sticky') {
			const rect = el.getBoundingClientRect();
			if (rect.width > 0 && rect.height > 0) {
				stickyFixedElements.push({
					el: el,
					rect: rect,
					position: style.position
				});

				const id = el.id ? '#' + el.id : '';
				const tag = el.tagName.toLowerCase();
				result.stickyElements.push(tag + id + ' (' + style.position + ')');

				// Determine if header or footer based on position
				if (rect.top < window.innerHeight / 2) {
					result.hasStickyHeaders = true;
				} else {
					result.hasStickyFooters = true;
				}
			}
		}
	});

	// Get focusable elements
	const focusable = document.querySelectorAll(
		'a[href], button, input:not([type="hidden"]), select, textarea, ' +
		'[tabindex]:not([tabindex="-1"])'
	);

	result.testedElements = Math.min(focusable.length, 30);

	// Check if any focusable elements could be obscured
	focusable.forEach((el, i) => {
		if (i >= 30) return;
		const rect = el.getBoundingClientRect();
		if (rect.width === 0 || rect.height === 0) return;

		// Check against each sticky/fixed element
		for (const sticky of stickyFixedElements) {
			const stickyRect = sticky.rect;

			// Check for vertical overlap
			const verticalOverlap =
				rect.top < stickyRect.bottom &&
				rect.bottom > stickyRect.top;

			// Check for horizontal overlap
			const horizontalOverlap =
				rect.left < stickyRect.right &&
				rect.right > stickyRect.left;

			if (verticalOverlap && horizontalOverlap) {
				// Element could be obscured when focused and scrolled
				const id = el.id ? '#' + el.id : '';
				result.potentiallyObscured.push(el.tagName.toLowerCase() + id);
				break;
			}
		}
	});

	// Limit results
	result.stickyElements = result.stickyElements.slice(0, 10);
	result.potentiallyObscured = result.potentiallyObscured.slice(0, 10);

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test focus obscured: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result FocusObscuredResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no sticky elements or minimal potential obscuring
	// Note: Having sticky elements doesn't automatically fail -
	// browsers typically scroll to show focused elements
	result.PassesTest = len(result.PotentiallyObscured) < result.TestedElements/4

	return &result, nil
}

// OnFocusResult contains results of on-focus context change testing (WCAG 3.2.1).
type OnFocusResult struct {
	// HasOnFocusHandlers indicates elements with onfocus handlers exist.
	HasOnFocusHandlers bool `json:"hasOnFocusHandlers"`

	// ProblematicElements lists elements that may cause context changes on focus.
	ProblematicElements []string `json:"problematicElements,omitempty"`

	// HasAutoFocus indicates autofocus attribute is used.
	HasAutoFocus bool `json:"hasAutoFocus"`

	// AutoFocusElement is the element with autofocus.
	AutoFocusElement string `json:"autoFocusElement,omitempty"`

	// PassesTest indicates the criterion is met.
	PassesTest bool `json:"passesTest"`
}

// TestOnFocus tests that focus doesn't cause context changes (WCAG 3.2.1).
func TestOnFocus(ctx context.Context, vibe *vibium.Vibe) (*OnFocusResult, error) {
	script := `
	const result = {
		hasOnFocusHandlers: false,
		problematicElements: [],
		hasAutoFocus: false,
		autoFocusElement: ''
	};

	// Check for autofocus attribute
	const autofocusEl = document.querySelector('[autofocus]');
	if (autofocusEl) {
		result.hasAutoFocus = true;
		const id = autofocusEl.id ? '#' + autofocusEl.id : '';
		result.autoFocusElement = autofocusEl.tagName.toLowerCase() + id;
	}

	// Check for onfocus handlers that might cause context changes
	const allElements = document.querySelectorAll('*');
	allElements.forEach((el, i) => {
		if (i >= 500) return;

		const onfocus = el.getAttribute('onfocus') || '';

		// Check for patterns that indicate context change
		const problematicPatterns = [
			'location',
			'href',
			'navigate',
			'submit',
			'window.open',
			'document.write'
		];

		for (const pattern of problematicPatterns) {
			if (onfocus.toLowerCase().includes(pattern)) {
				result.hasOnFocusHandlers = true;
				const id = el.id ? '#' + el.id : '';
				result.problematicElements.push(
					el.tagName.toLowerCase() + id + ' (onfocus contains "' + pattern + '")'
				);
				break;
			}
		}
	});

	// Check for scripts that add focus listeners with navigation
	const scripts = document.querySelectorAll('script');
	scripts.forEach(script => {
		const text = script.textContent || '';
		// Look for focus event listeners combined with navigation
		if ((text.includes('focus') || text.includes('Focus')) &&
			(text.includes('location') || text.includes('navigate') || text.includes('href'))) {
			// This is a heuristic - may have false positives
			result.hasOnFocusHandlers = true;
		}
	});

	// Limit results
	result.problematicElements = result.problematicElements.slice(0, 10);

	return JSON.stringify(result);
	`

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to test on-focus: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var result OnFocusResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Passes if no problematic onfocus handlers detected
	result.PassesTest = len(result.ProblematicElements) == 0

	return &result, nil
}
