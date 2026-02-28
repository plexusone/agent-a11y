// Package specialized provides specialized automated accessibility tests
// that go beyond what axe-core can detect.
//
// These tests use browser automation to validate WCAG criteria that require
// dynamic interaction or visual comparison:
//
//   - Keyboard accessibility (2.1.1, 2.1.2): Tab navigation, keyboard traps
//   - Focus visibility (2.4.7, 2.4.3, 2.4.11, 3.2.1): Focus indicators, order, context
//   - Content reflow (1.4.10): 320px viewport behavior
//   - Target size (2.5.8): Touch target minimum dimensions
//   - Text spacing (1.4.12): Content adaptation to increased spacing
//   - Hover/focus content (1.4.13): Tooltips, dropdowns, dismissibility
//   - Character shortcuts (2.1.4): Single-character keyboard shortcuts
//   - Media (1.2.2, 1.2.4, 1.2.5, 1.4.2): Captions, audio controls
//   - Flashing content (2.3.1): Flash rate detection
//   - Responsive behavior (1.4.4, 1.3.4): Zoom, orientation
package specialized
