package specialized

import (
	"context"
	"encoding/json"
	"fmt"

	vibium "github.com/plexusone/vibium-go"
)

// TargetSizeTestResult contains results of target size testing.
type TargetSizeTestResult struct {
	// AllTargetsAdequate indicates all touch targets meet size requirements.
	AllTargetsAdequate bool `json:"allTargetsAdequate"`

	// SmallTargets lists elements below the minimum size.
	SmallTargets []SmallTarget `json:"smallTargets,omitempty"`

	// TestedElements is the count of interactive elements tested.
	TestedElements int `json:"testedElements"`

	// MinimumSize is the required minimum size (24x24 for WCAG 2.5.8).
	MinimumSize int `json:"minimumSize"`
}

// SmallTarget represents an interactive element below minimum size.
type SmallTarget struct {
	Selector string  `json:"selector"`
	TagName  string  `json:"tagName"`
	Role     string  `json:"role,omitempty"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Area     float64 `json:"area"`
}

// TestTargetSize tests that touch targets are at least 24x24 CSS pixels (WCAG 2.5.8).
func TestTargetSize(ctx context.Context, vibe *vibium.Vibe, minimumSize int) (*TargetSizeTestResult, error) {
	if minimumSize <= 0 {
		minimumSize = 24 // WCAG 2.5.8 minimum
	}

	result := &TargetSizeTestResult{
		AllTargetsAdequate: true,
		MinimumSize:        minimumSize,
	}

	script := fmt.Sprintf(`
	const minSize = %d;
	const interactive = document.querySelectorAll(
		'a[href], button, input:not([type="hidden"]), select, textarea, ' +
		'[role="button"], [role="link"], [role="checkbox"], [role="menuitem"], ' +
		'[role="tab"], [role="switch"], [onclick]'
	);

	const results = {
		tested: 0,
		smallTargets: []
	};

	interactive.forEach(el => {
		const rect = el.getBoundingClientRect();
		const style = window.getComputedStyle(el);

		// Skip hidden elements
		if (rect.width === 0 || rect.height === 0 ||
			style.visibility === 'hidden' || style.display === 'none') {
			return;
		}

		results.tested++;

		// Check if target is too small
		if (rect.width < minSize || rect.height < minSize) {
			const id = el.id ? '#' + el.id : '';
			const classes = el.className && typeof el.className === 'string'
				? '.' + el.className.split(' ').filter(c => c).join('.')
				: '';

			results.smallTargets.push({
				selector: el.tagName.toLowerCase() + id + classes,
				tagName: el.tagName.toLowerCase(),
				role: el.getAttribute('role') || '',
				width: rect.width,
				height: rect.height,
				area: rect.width * rect.height
			});
		}
	});

	return JSON.stringify(results);
	`, minimumSize)

	rawResult, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate target sizes: %w", err)
	}

	jsonStr, ok := rawResult.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	var data struct {
		Tested       int           `json:"tested"`
		SmallTargets []SmallTarget `json:"smallTargets"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse target results: %w", err)
	}

	result.TestedElements = data.Tested
	result.SmallTargets = data.SmallTargets

	if len(result.SmallTargets) > 0 {
		result.AllTargetsAdequate = false
	}

	return result, nil
}
