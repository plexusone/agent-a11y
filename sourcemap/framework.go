package sourcemap

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/plexusone/agent-a11y/types"
)

// FrameworkDetector detects the frontend framework from page content.
type FrameworkDetector struct {
	// Detected framework
	framework types.Framework

	// Detection confidence (0-1)
	confidence float64

	// Detection signals found
	signals []string
}

// NewFrameworkDetector creates a new framework detector.
func NewFrameworkDetector() *FrameworkDetector {
	return &FrameworkDetector{
		framework:  types.FrameworkUnknown,
		confidence: 0,
	}
}

// DetectFromHTML detects the framework from HTML content.
func (d *FrameworkDetector) DetectFromHTML(html string) types.Framework {
	d.signals = nil
	d.confidence = 0
	d.framework = types.FrameworkUnknown

	// Check each framework
	scores := make(map[types.Framework]float64)

	// React detection
	scores[types.FrameworkReact] = d.detectReact(html)

	// Vue detection
	scores[types.FrameworkVue] = d.detectVue(html)

	// Svelte detection
	scores[types.FrameworkSvelte] = d.detectSvelte(html)

	// Angular detection
	scores[types.FrameworkAngular] = d.detectAngular(html)

	// Find highest score
	var maxScore float64
	for fw, score := range scores {
		if score > maxScore {
			maxScore = score
			d.framework = fw
			d.confidence = score
		}
	}

	// If no framework detected with high confidence, default to vanilla
	if maxScore < 0.3 {
		d.framework = types.FrameworkVanilla
		d.confidence = 1.0 - maxScore // Higher confidence it's vanilla if no signals
	}

	return d.framework
}

// detectReact checks for React signals in HTML.
func (d *FrameworkDetector) detectReact(html string) float64 {
	var score float64

	// Check for React-specific attributes
	if strings.Contains(html, "data-reactroot") {
		d.signals = append(d.signals, "data-reactroot")
		score += 0.4
	}

	// Check for React DevTools hook
	if strings.Contains(html, "__REACT_DEVTOOLS_GLOBAL_HOOK__") {
		d.signals = append(d.signals, "__REACT_DEVTOOLS_GLOBAL_HOOK__")
		score += 0.3
	}

	// Check for React ID pattern (older React)
	reactIDPattern := regexp.MustCompile(`data-reactid="[^"]*"`)
	if reactIDPattern.MatchString(html) {
		d.signals = append(d.signals, "data-reactid")
		score += 0.3
	}

	// Check for data-react-helmet (common in React apps)
	if strings.Contains(html, "data-react-helmet") {
		d.signals = append(d.signals, "data-react-helmet")
		score += 0.2
	}

	// Check for _reactRootContainer in script
	if strings.Contains(html, "_reactRootContainer") {
		d.signals = append(d.signals, "_reactRootContainer")
		score += 0.3
	}

	// Check for Next.js
	if strings.Contains(html, "__NEXT_DATA__") {
		d.signals = append(d.signals, "__NEXT_DATA__")
		score += 0.3
	}

	// Check for Gatsby
	if strings.Contains(html, "___gatsby") {
		d.signals = append(d.signals, "___gatsby")
		score += 0.3
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// detectVue checks for Vue signals in HTML.
func (d *FrameworkDetector) detectVue(html string) float64 {
	var score float64

	// Check for Vue DevTools hook
	if strings.Contains(html, "__VUE__") {
		d.signals = append(d.signals, "__VUE__")
		score += 0.4
	}

	// Check for Vue scoped style attributes
	vueScopedPattern := regexp.MustCompile(`data-v-[a-f0-9]+`)
	if vueScopedPattern.MatchString(html) {
		d.signals = append(d.signals, "data-v-* (scoped styles)")
		score += 0.3
	}

	// Check for v-* directives in attributes
	vDirectivePattern := regexp.MustCompile(`v-[a-z]+`)
	if vDirectivePattern.MatchString(html) {
		d.signals = append(d.signals, "v-* directives")
		score += 0.2
	}

	// Check for Vue-specific app mounting
	if strings.Contains(html, "Vue.createApp") || strings.Contains(html, "createApp(") {
		d.signals = append(d.signals, "createApp")
		score += 0.3
	}

	// Check for Nuxt.js
	if strings.Contains(html, "__NUXT__") {
		d.signals = append(d.signals, "__NUXT__")
		score += 0.3
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// detectSvelte checks for Svelte signals in HTML.
func (d *FrameworkDetector) detectSvelte(html string) float64 {
	var score float64

	// Check for Svelte component hash classes
	svelteClassPattern := regexp.MustCompile(`svelte-[a-z0-9]+`)
	if svelteClassPattern.MatchString(html) {
		d.signals = append(d.signals, "svelte-* classes")
		score += 0.4
	}

	// Check for SvelteKit
	if strings.Contains(html, "__sveltekit") {
		d.signals = append(d.signals, "__sveltekit")
		score += 0.3
	}

	// Check for svelte hydration markers
	if strings.Contains(html, "data-svelte") {
		d.signals = append(d.signals, "data-svelte")
		score += 0.3
	}

	// Check for __svelte global
	if strings.Contains(html, "__svelte") {
		d.signals = append(d.signals, "__svelte")
		score += 0.2
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// detectAngular checks for Angular signals in HTML.
func (d *FrameworkDetector) detectAngular(html string) float64 {
	var score float64

	// Check for ng-version attribute
	if strings.Contains(html, "ng-version") {
		d.signals = append(d.signals, "ng-version")
		score += 0.4
	}

	// Check for _nghost-* or _ngcontent-* attributes
	ngAttributePattern := regexp.MustCompile(`_ng(host|content)-[a-z]+-\d+`)
	if ngAttributePattern.MatchString(html) {
		d.signals = append(d.signals, "_nghost-*/_ngcontent-*")
		score += 0.3
	}

	// Check for Angular app root
	if strings.Contains(html, "ng-app") || strings.Contains(html, "app-root") {
		d.signals = append(d.signals, "ng-app/app-root")
		score += 0.2
	}

	// Check for Angular Universal
	if strings.Contains(html, "ng-state") {
		d.signals = append(d.signals, "ng-state")
		score += 0.2
	}

	// Check for zone.js
	if strings.Contains(html, "Zone.") || strings.Contains(html, "__zone_symbol__") {
		d.signals = append(d.signals, "zone.js")
		score += 0.2
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// Framework returns the detected framework.
func (d *FrameworkDetector) Framework() types.Framework {
	return d.framework
}

// Confidence returns the detection confidence.
func (d *FrameworkDetector) Confidence() float64 {
	return d.confidence
}

// Signals returns the detection signals found.
func (d *FrameworkDetector) Signals() []string {
	return d.signals
}

// DetectFromPage detects framework from a live page using JavaScript execution.
// This is used when the HTML alone doesn't provide enough signals.
// The execJS function should execute JavaScript and return the result.
func (d *FrameworkDetector) DetectFromPage(_ context.Context, execJS func(string) (string, error)) (types.Framework, error) {
	// JavaScript to detect framework at runtime
	detectScript := `
		(function() {
			var fw = { name: 'unknown', confidence: 0, signals: [] };

			// Check React
			if (window.__REACT_DEVTOOLS_GLOBAL_HOOK__ ||
			    document.querySelector('[data-reactroot]') ||
			    document.querySelector('[data-reactid]') ||
			    window.__NEXT_DATA__ ||
			    document.getElementById('___gatsby')) {
				fw.name = 'react';
				fw.confidence = 0.9;
				fw.signals.push('react-devtools');
			}

			// Check Vue
			else if (window.__VUE__ ||
			         document.querySelector('[data-v-]') ||
			         window.__NUXT__) {
				fw.name = 'vue';
				fw.confidence = 0.9;
				fw.signals.push('vue-devtools');
			}

			// Check Svelte
			else if (document.querySelector('[class*="svelte-"]') ||
			         window.__sveltekit) {
				fw.name = 'svelte';
				fw.confidence = 0.9;
				fw.signals.push('svelte-classes');
			}

			// Check Angular
			else if (document.querySelector('[ng-version]') ||
			         document.querySelector('app-root') ||
			         window.getAllAngularRootElements) {
				fw.name = 'angular';
				fw.confidence = 0.9;
				fw.signals.push('angular-root');
			}

			// Default to vanilla
			else {
				fw.name = 'vanilla';
				fw.confidence = 1.0;
			}

			return JSON.stringify(fw);
		})()
	`

	result, err := execJS(detectScript)
	if err != nil {
		// Fall back to HTML-based detection
		return types.FrameworkUnknown, err
	}

	// Parse result
	var detection struct {
		Name       string   `json:"name"`
		Confidence float64  `json:"confidence"`
		Signals    []string `json:"signals"`
	}
	if err := parseJSON(result, &detection); err != nil {
		return types.FrameworkUnknown, err
	}

	d.framework = frameworkFromString(detection.Name)
	d.confidence = detection.Confidence
	d.signals = detection.Signals

	return d.framework, nil
}

// frameworkFromString converts a string to Framework type.
func frameworkFromString(s string) types.Framework {
	switch strings.ToLower(s) {
	case "react":
		return types.FrameworkReact
	case "vue":
		return types.FrameworkVue
	case "svelte":
		return types.FrameworkSvelte
	case "angular":
		return types.FrameworkAngular
	case "vanilla":
		return types.FrameworkVanilla
	default:
		return types.FrameworkUnknown
	}
}

// parseJSON parses JSON into a target struct.
func parseJSON(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}
