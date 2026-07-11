package sourcemap

import (
	"testing"

	"github.com/plexusone/agent-a11y/types"
)

func TestFrameworkDetection(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected types.Framework
	}{
		{
			name:     "React with data-reactroot",
			html:     `<div data-reactroot><span>Hello</span></div>`,
			expected: types.FrameworkReact,
		},
		{
			name:     "React with Next.js",
			html:     `<script id="__NEXT_DATA__" type="application/json">{}</script>`,
			expected: types.FrameworkReact,
		},
		{
			name:     "React with Gatsby",
			html:     `<div id="___gatsby"></div>`,
			expected: types.FrameworkReact,
		},
		{
			name:     "Vue with scoped styles",
			html:     `<div data-v-abc123><span data-v-abc123>Hello</span></div>`,
			expected: types.FrameworkVue,
		},
		{
			name:     "Vue with Nuxt",
			html:     `<script>window.__NUXT__={}</script>`,
			expected: types.FrameworkVue,
		},
		{
			name:     "Svelte with class hash",
			html:     `<div class="svelte-xyz123"><span class="svelte-xyz123">Hello</span></div>`,
			expected: types.FrameworkSvelte,
		},
		{
			name:     "Angular with ng-version",
			html:     `<app-root ng-version="14.0.0"></app-root>`,
			expected: types.FrameworkAngular,
		},
		{
			name:     "Angular with _nghost",
			html:     `<div _nghost-abc-123><span _ngcontent-abc-123>Hello</span></div>`,
			expected: types.FrameworkAngular,
		},
		{
			name:     "Vanilla HTML",
			html:     `<div class="container"><span>Hello</span></div>`,
			expected: types.FrameworkVanilla,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewFrameworkDetector()
			result := detector.DetectFromHTML(tt.html)

			if result != tt.expected {
				t.Errorf("expected %s, got %s (confidence: %.2f, signals: %v)",
					tt.expected, result, detector.Confidence(), detector.Signals())
			}

			// Verify confidence is reasonable
			if detector.Confidence() <= 0 {
				t.Error("expected positive confidence")
			}
		})
	}
}

func TestFrameworkDetectorSignals(t *testing.T) {
	detector := NewFrameworkDetector()

	// Test React detection signals
	_ = detector.DetectFromHTML(`<div data-reactroot data-react-helmet="true"></div>`)
	signals := detector.Signals()

	if len(signals) == 0 {
		t.Error("expected detection signals")
	}

	// Should include both signals
	found := make(map[string]bool)
	for _, s := range signals {
		found[s] = true
	}

	if !found["data-reactroot"] {
		t.Error("expected data-reactroot signal")
	}
	if !found["data-react-helmet"] {
		t.Error("expected data-react-helmet signal")
	}
}

func TestFrameworkDetectorConfidence(t *testing.T) {
	detector := NewFrameworkDetector()

	// More signals should increase confidence
	// Single signal
	_ = detector.DetectFromHTML(`<div data-reactroot></div>`)
	singleConfidence := detector.Confidence()

	// Multiple signals
	_ = detector.DetectFromHTML(`<div data-reactroot data-react-helmet="true"><script>window.__NEXT_DATA__={}</script></div>`)
	multiConfidence := detector.Confidence()

	if multiConfidence <= singleConfidence {
		t.Errorf("multiple signals should increase confidence: single=%.2f, multi=%.2f",
			singleConfidence, multiConfidence)
	}
}

func TestFrameworkFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Framework
	}{
		{"react", types.FrameworkReact},
		{"React", types.FrameworkReact},
		{"REACT", types.FrameworkReact},
		{"vue", types.FrameworkVue},
		{"svelte", types.FrameworkSvelte},
		{"angular", types.FrameworkAngular},
		{"vanilla", types.FrameworkVanilla},
		{"unknown", types.FrameworkUnknown},
		{"invalid", types.FrameworkUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := frameworkFromString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetectorReuse(t *testing.T) {
	detector := NewFrameworkDetector()

	// First detection
	fw1 := detector.DetectFromHTML(`<div data-reactroot></div>`)
	signals1 := detector.Signals()
	conf1 := detector.Confidence()

	// Second detection (should reset state)
	fw2 := detector.DetectFromHTML(`<div data-v-abc123></div>`)
	signals2 := detector.Signals()
	conf2 := detector.Confidence()

	if fw1 == fw2 {
		t.Error("expected different frameworks")
	}

	// Signals should be different
	if len(signals1) == len(signals2) {
		same := true
		for i := range signals1 {
			if i < len(signals2) && signals1[i] != signals2[i] {
				same = false
				break
			}
		}
		if same && len(signals1) > 0 {
			t.Error("signals should be reset between detections")
		}
	}

	// Just verify confidence is set
	if conf1 == 0 || conf2 == 0 {
		t.Error("expected non-zero confidence")
	}
}
