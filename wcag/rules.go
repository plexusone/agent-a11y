// Package wcag provides WCAG accessibility testing rules.
package wcag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	vibium "github.com/plexusone/vibium-go"
	"github.com/plexusone/agent-a11y/types"
)

// Rule represents a WCAG accessibility test rule.
type Rule interface {
	// ID returns the rule identifier
	ID() string

	// Name returns a human-readable name
	Name() string

	// Description returns the rule description
	Description() string

	// SuccessCriteria returns the WCAG success criteria this rule tests
	SuccessCriteria() []string

	// Level returns the WCAG level (A, AA, AAA)
	Level() types.WCAGLevel

	// Run executes the rule and returns findings
	Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error)
}

// Registry holds all available rules.
type Registry struct {
	rules  map[string]Rule
	logger *slog.Logger
}

// NewRegistry creates a new rule registry with all built-in rules.
func NewRegistry(logger *slog.Logger) *Registry {
	r := &Registry{
		rules:  make(map[string]Rule),
		logger: logger,
	}

	// Register all built-in rules
	r.registerBuiltinRules()

	return r
}

// Register adds a rule to the registry.
func (r *Registry) Register(rule Rule) {
	r.rules[rule.ID()] = rule
}

// Get returns a rule by ID.
func (r *Registry) Get(id string) (Rule, bool) {
	rule, ok := r.rules[id]
	return rule, ok
}

// GetByLevel returns all rules for a specific WCAG level.
func (r *Registry) GetByLevel(level types.WCAGLevel) []Rule {
	var result []Rule
	for _, rule := range r.rules {
		if rule.Level() == level || isLowerLevel(rule.Level(), level) {
			result = append(result, rule)
		}
	}
	return result
}

// GetByCriterion returns all rules that test a specific success criterion.
func (r *Registry) GetByCriterion(criterion string) []Rule {
	var result []Rule
	for _, rule := range r.rules {
		for _, sc := range rule.SuccessCriteria() {
			if sc == criterion {
				result = append(result, rule)
				break
			}
		}
	}
	return result
}

// All returns all registered rules.
func (r *Registry) All() []Rule {
	result := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		result = append(result, rule)
	}
	return result
}

// isLowerLevel checks if l1 is a lower level than l2.
func isLowerLevel(l1, l2 types.WCAGLevel) bool {
	levels := map[types.WCAGLevel]int{
		types.WCAGLevelA:   1,
		types.WCAGLevelAA:  2,
		types.WCAGLevelAAA: 3,
	}
	return levels[l1] < levels[l2]
}

// registerBuiltinRules registers all built-in WCAG rules.
func (r *Registry) registerBuiltinRules() {
	// WCAG 1.1.1 - Non-text Content
	r.Register(&ImageAltRule{})
	r.Register(&ImageButtonAltRule{})

	// WCAG 1.3.1 - Info and Relationships
	r.Register(&FormLabelRule{})
	r.Register(&HeadingStructureRule{})
	r.Register(&TableHeaderRule{})
	r.Register(&LandmarkRule{})

	// WCAG 1.4.1 - Use of Color
	r.Register(&LinkDistinguishableRule{})

	// WCAG 1.4.3 - Contrast (Minimum)
	r.Register(&ContrastRule{})

	// WCAG 2.1.1 - Keyboard
	r.Register(&KeyboardAccessRule{})
	r.Register(&FocusVisibleRule{})

	// WCAG 2.4.1 - Bypass Blocks
	r.Register(&SkipLinkRule{})

	// WCAG 2.4.2 - Page Titled
	r.Register(&PageTitleRule{})

	// WCAG 2.4.4 - Link Purpose
	r.Register(&LinkPurposeRule{})

	// WCAG 2.4.6 - Headings and Labels
	r.Register(&DescriptiveHeadingRule{})

	// WCAG 3.1.1 - Language of Page
	r.Register(&LanguageRule{})

	// WCAG 4.1.1 - Parsing
	r.Register(&DuplicateIDRule{})

	// WCAG 4.1.2 - Name, Role, Value
	r.Register(&AriaLabelRule{})
	r.Register(&ButtonNameRule{})
	r.Register(&FormInputNameRule{})
}

// BaseRule provides common rule functionality.
type BaseRule struct {
	id          string
	name        string
	description string
	criteria    []string
	level       types.WCAGLevel
}

func (r *BaseRule) ID() string                 { return r.id }
func (r *BaseRule) Name() string               { return r.name }
func (r *BaseRule) Description() string        { return r.description }
func (r *BaseRule) SuccessCriteria() []string  { return r.criteria }
func (r *BaseRule) Level() types.WCAGLevel     { return r.level }

// ImageAltRule checks for missing alt text on images.
type ImageAltRule struct {
	BaseRule
}

func (r *ImageAltRule) ID() string { return "image-alt" }
func (r *ImageAltRule) Name() string { return "Image Alt Text" }
func (r *ImageAltRule) Description() string { return "Images must have alt text" }
func (r *ImageAltRule) SuccessCriteria() []string { return []string{"1.1.1"} }
func (r *ImageAltRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *ImageAltRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		Array.from(document.querySelectorAll('img:not([role="presentation"]):not([role="none"])')).map(img => ({
			selector: getSelector(img),
			html: img.outerHTML.substring(0, 200),
			hasAlt: img.hasAttribute('alt'),
			alt: img.getAttribute('alt'),
			src: img.src
		}));

		function getSelector(el) {
			if (el.id) return '#' + el.id;
			if (el.className) return el.tagName.toLowerCase() + '.' + el.className.split(' ').join('.');
			return el.tagName.toLowerCase();
		}
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate script: %w", err)
	}

	findings := make([]types.Finding, 0)
	if images, ok := result.([]any); ok {
		for _, item := range images {
			img, ok := item.(map[string]any)
			if !ok {
				continue
			}

			hasAlt, _ := img["hasAlt"].(bool)
			alt, _ := img["alt"].(string)

			if !hasAlt {
				findings = append(findings, types.Finding{
					RuleID:          r.ID(),
					Description:     "Image is missing alt attribute",
					Help:            "Add an alt attribute to describe the image content",
					SuccessCriteria: r.SuccessCriteria(),
					Level:           r.Level(),
					Impact:          types.ImpactCritical,
					Selector:        getString(img, "selector"),
					HTML:            getString(img, "html"),
					Element:         "img",
				})
			} else if alt == "" {
				// Empty alt is OK for decorative images, but flag for review
				findings = append(findings, types.Finding{
					RuleID:          r.ID(),
					Description:     "Image has empty alt attribute - verify it's decorative",
					Help:            "If the image is decorative, this is correct. Otherwise, add descriptive text.",
					SuccessCriteria: r.SuccessCriteria(),
					Level:           r.Level(),
					Impact:          types.ImpactMinor,
					Selector:        getString(img, "selector"),
					HTML:            getString(img, "html"),
					Element:         "img",
				})
			}
		}
	}

	return findings, nil
}

// ImageButtonAltRule checks for missing alt text on image buttons.
type ImageButtonAltRule struct{ BaseRule }

func (r *ImageButtonAltRule) ID() string { return "image-button-alt" }
func (r *ImageButtonAltRule) Name() string { return "Image Button Alt Text" }
func (r *ImageButtonAltRule) Description() string { return "Image buttons must have alt text" }
func (r *ImageButtonAltRule) SuccessCriteria() []string { return []string{"1.1.1"} }
func (r *ImageButtonAltRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *ImageButtonAltRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		Array.from(document.querySelectorAll('input[type="image"]')).map(img => ({
			selector: img.id ? '#' + img.id : 'input[type="image"]',
			html: img.outerHTML,
			hasAlt: img.hasAttribute('alt'),
			alt: img.getAttribute('alt')
		}));
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if buttons, ok := result.([]any); ok {
		for _, item := range buttons {
			btn, ok := item.(map[string]any)
			if !ok {
				continue
			}

			hasAlt, _ := btn["hasAlt"].(bool)
			alt, _ := btn["alt"].(string)

			if !hasAlt || alt == "" {
				findings = append(findings, types.Finding{
					RuleID:          r.ID(),
					Description:     "Image button is missing alt attribute",
					Help:            "Add alt text describing the button action",
					SuccessCriteria: r.SuccessCriteria(),
					Level:           r.Level(),
					Impact:          types.ImpactCritical,
					Selector:        getString(btn, "selector"),
					HTML:            getString(btn, "html"),
					Element:         "input",
				})
			}
		}
	}

	return findings, nil
}

// FormLabelRule checks that form inputs have labels.
type FormLabelRule struct{ BaseRule }

func (r *FormLabelRule) ID() string { return "form-label" }
func (r *FormLabelRule) Name() string { return "Form Labels" }
func (r *FormLabelRule) Description() string { return "Form inputs must have labels" }
func (r *FormLabelRule) SuccessCriteria() []string { return []string{"1.3.1", "4.1.2"} }
func (r *FormLabelRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *FormLabelRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const inputs = document.querySelectorAll('input:not([type="hidden"]):not([type="submit"]):not([type="button"]):not([type="reset"]):not([type="image"]), select, textarea');
		Array.from(inputs).map(input => {
			const id = input.id;
			const hasLabel = id && document.querySelector('label[for="' + id + '"]');
			const hasAriaLabel = input.hasAttribute('aria-label');
			const hasAriaLabelledBy = input.hasAttribute('aria-labelledby');
			const hasTitle = input.hasAttribute('title');
			const isLabelled = hasLabel || hasAriaLabel || hasAriaLabelledBy || hasTitle;

			return {
				selector: id ? '#' + id : input.tagName.toLowerCase() + (input.name ? '[name="' + input.name + '"]' : ''),
				html: input.outerHTML.substring(0, 200),
				type: input.type || input.tagName.toLowerCase(),
				hasLabel: isLabelled
			};
		});
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if inputs, ok := result.([]any); ok {
		for _, item := range inputs {
			input, ok := item.(map[string]any)
			if !ok {
				continue
			}

			hasLabel, _ := input["hasLabel"].(bool)
			if !hasLabel {
				findings = append(findings, types.Finding{
					RuleID:          r.ID(),
					Description:     "Form input is missing a label",
					Help:            "Add a <label> element or aria-label attribute",
					SuccessCriteria: r.SuccessCriteria(),
					Level:           r.Level(),
					Impact:          types.ImpactSerious,
					Selector:        getString(input, "selector"),
					HTML:            getString(input, "html"),
					Element:         getString(input, "type"),
				})
			}
		}
	}

	return findings, nil
}

// HeadingStructureRule checks heading hierarchy.
type HeadingStructureRule struct{ BaseRule }

func (r *HeadingStructureRule) ID() string { return "heading-structure" }
func (r *HeadingStructureRule) Name() string { return "Heading Structure" }
func (r *HeadingStructureRule) Description() string { return "Headings should follow logical order" }
func (r *HeadingStructureRule) SuccessCriteria() []string { return []string{"1.3.1"} }
func (r *HeadingStructureRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *HeadingStructureRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
		const levels = Array.from(headings).map(h => ({
			level: parseInt(h.tagName.substring(1)),
			text: h.textContent.trim().substring(0, 50),
			selector: h.id ? '#' + h.id : h.tagName.toLowerCase(),
			html: h.outerHTML.substring(0, 100)
		}));

		// Check for missing h1
		const hasH1 = levels.some(h => h.level === 1);

		// Check for skipped levels
		const skipped = [];
		for (let i = 1; i < levels.length; i++) {
			if (levels[i].level > levels[i-1].level + 1) {
				skipped.push({
					...levels[i],
					previousLevel: levels[i-1].level
				});
			}
		}

		return { hasH1, skipped, count: levels.length };
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if data, ok := result.(map[string]any); ok {
		hasH1, _ := data["hasH1"].(bool)
		if !hasH1 {
			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     "Page is missing an h1 heading",
				Help:            "Add an h1 heading that describes the page content",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactModerate,
				Element:         "h1",
			})
		}

		if skipped, ok := data["skipped"].([]any); ok {
			for _, item := range skipped {
				skip, ok := item.(map[string]any)
				if !ok {
					continue
				}
				level, _ := skip["level"].(float64)
				prevLevel, _ := skip["previousLevel"].(float64)
				findings = append(findings, types.Finding{
					RuleID:          r.ID(),
					Description:     fmt.Sprintf("Heading level skipped from h%d to h%d", int(prevLevel), int(level)),
					Help:            "Don't skip heading levels - use headings sequentially",
					SuccessCriteria: r.SuccessCriteria(),
					Level:           r.Level(),
					Impact:          types.ImpactModerate,
					Selector:        getString(skip, "selector"),
					HTML:            getString(skip, "html"),
					Element:         fmt.Sprintf("h%d", int(level)),
				})
			}
		}
	}

	return findings, nil
}

// TableHeaderRule checks that tables have headers.
type TableHeaderRule struct{ BaseRule }

func (r *TableHeaderRule) ID() string { return "table-header" }
func (r *TableHeaderRule) Name() string { return "Table Headers" }
func (r *TableHeaderRule) Description() string { return "Tables must have headers" }
func (r *TableHeaderRule) SuccessCriteria() []string { return []string{"1.3.1"} }
func (r *TableHeaderRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *TableHeaderRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		Array.from(document.querySelectorAll('table')).filter(table => {
			// Skip layout tables
			if (table.getAttribute('role') === 'presentation') return false;
			return true;
		}).map(table => ({
			selector: table.id ? '#' + table.id : 'table',
			html: table.outerHTML.substring(0, 200),
			hasTh: table.querySelectorAll('th').length > 0,
			hasCaption: table.querySelector('caption') !== null
		}));
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if tables, ok := result.([]any); ok {
		for _, item := range tables {
			table, ok := item.(map[string]any)
			if !ok {
				continue
			}

			hasTh, _ := table["hasTh"].(bool)
			if !hasTh {
				findings = append(findings, types.Finding{
					RuleID:          r.ID(),
					Description:     "Table is missing header cells (th)",
					Help:            "Add th elements to identify column/row headers",
					SuccessCriteria: r.SuccessCriteria(),
					Level:           r.Level(),
					Impact:          types.ImpactSerious,
					Selector:        getString(table, "selector"),
					HTML:            getString(table, "html"),
					Element:         "table",
				})
			}
		}
	}

	return findings, nil
}

// LandmarkRule checks for ARIA landmarks.
type LandmarkRule struct{ BaseRule }

func (r *LandmarkRule) ID() string { return "landmark-regions" }
func (r *LandmarkRule) Name() string { return "Landmark Regions" }
func (r *LandmarkRule) Description() string { return "Page should use landmark regions" }
func (r *LandmarkRule) SuccessCriteria() []string { return []string{"1.3.1"} }
func (r *LandmarkRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *LandmarkRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const landmarks = {
			main: document.querySelector('main, [role="main"]'),
			nav: document.querySelector('nav, [role="navigation"]'),
			header: document.querySelector('header, [role="banner"]'),
			footer: document.querySelector('footer, [role="contentinfo"]')
		};
		return landmarks;
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if landmarks, ok := result.(map[string]any); ok {
		if landmarks["main"] == nil {
			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     "Page is missing a main landmark",
				Help:            "Add a <main> element or role=\"main\"",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactModerate,
				Element:         "main",
			})
		}
	}

	return findings, nil
}

// ContrastRule checks color contrast.
type ContrastRule struct{ BaseRule }

func (r *ContrastRule) ID() string { return "color-contrast" }
func (r *ContrastRule) Name() string { return "Color Contrast" }
func (r *ContrastRule) Description() string { return "Text must have sufficient color contrast" }
func (r *ContrastRule) SuccessCriteria() []string { return []string{"1.4.3"} }
func (r *ContrastRule) Level() types.WCAGLevel { return types.WCAGLevelAA }

func (r *ContrastRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	// Color contrast checking with luminance calculation
	script := `
		function getLuminance(r, g, b) {
			const [rs, gs, bs] = [r, g, b].map(c => {
				c = c / 255;
				return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
			});
			return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs;
		}

		function parseColor(color) {
			if (!color || color === 'transparent' || color === 'rgba(0, 0, 0, 0)') return null;
			const match = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
			if (match) return [parseInt(match[1]), parseInt(match[2]), parseInt(match[3])];
			return null;
		}

		function getContrastRatio(l1, l2) {
			const lighter = Math.max(l1, l2);
			const darker = Math.min(l1, l2);
			return (lighter + 0.05) / (darker + 0.05);
		}

		const textElements = document.querySelectorAll('p, span, a, li, td, th, label, h1, h2, h3, h4, h5, h6, button');
		const issues = [];

		textElements.forEach(el => {
			const style = getComputedStyle(el);
			const text = el.textContent.trim();
			if (!text || text.length === 0) return;

			const fg = parseColor(style.color);
			const bg = parseColor(style.backgroundColor);

			if (!fg || !bg) return; // Skip if colors can't be determined

			const fgLum = getLuminance(fg[0], fg[1], fg[2]);
			const bgLum = getLuminance(bg[0], bg[1], bg[2]);
			const ratio = getContrastRatio(fgLum, bgLum);

			const fontSize = parseFloat(style.fontSize);
			const isBold = parseInt(style.fontWeight) >= 700;
			const isLargeText = fontSize >= 18 || (fontSize >= 14 && isBold);
			const requiredRatio = isLargeText ? 3 : 4.5;

			if (ratio < requiredRatio) {
				issues.push({
					selector: el.id ? '#' + el.id : el.tagName.toLowerCase(),
					html: el.outerHTML.substring(0, 150),
					ratio: ratio.toFixed(2),
					required: requiredRatio,
					isLargeText: isLargeText
				});
			}
		});

		return issues.slice(0, 20); // Limit to 20 findings
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if issues, ok := result.([]any); ok {
		for _, item := range issues {
			issue, ok := item.(map[string]any)
			if !ok {
				continue
			}

			ratio := getString(issue, "ratio")
			required, _ := issue["required"].(float64)
			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     fmt.Sprintf("Contrast ratio is %s:1, requires %v:1", ratio, required),
				Help:            "Increase color contrast between text and background",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactSerious,
				Selector:        getString(issue, "selector"),
				HTML:            getString(issue, "html"),
			})
		}
	}

	return findings, nil
}

// Additional rule stubs - implement as needed

type LinkDistinguishableRule struct{ BaseRule }
func (r *LinkDistinguishableRule) ID() string { return "link-distinguishable" }
func (r *LinkDistinguishableRule) Name() string { return "Link Distinguishable" }
func (r *LinkDistinguishableRule) Description() string { return "Links must be distinguishable from surrounding text" }
func (r *LinkDistinguishableRule) SuccessCriteria() []string { return []string{"1.4.1"} }
func (r *LinkDistinguishableRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *LinkDistinguishableRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const links = document.querySelectorAll('a');
		const issues = [];

		links.forEach(link => {
			const style = getComputedStyle(link);
			const parent = link.parentElement;
			if (!parent) return;

			const parentStyle = getComputedStyle(parent);

			// Check if link has underline or is visually distinct
			const hasUnderline = style.textDecoration.includes('underline');
			const hasDifferentColor = style.color !== parentStyle.color;
			const hasBorder = style.borderBottomWidth !== '0px' && style.borderBottomStyle !== 'none';
			const hasBold = parseInt(style.fontWeight) > parseInt(parentStyle.fontWeight);

			// Link should be distinguishable by more than just color
			if (hasDifferentColor && !hasUnderline && !hasBorder && !hasBold) {
				issues.push({
					selector: link.id ? '#' + link.id : 'a',
					html: link.outerHTML.substring(0, 150),
					text: link.textContent.trim().substring(0, 50)
				});
			}
		});

		return issues.slice(0, 15);
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if issues, ok := result.([]any); ok {
		for _, item := range issues {
			issue, ok := item.(map[string]any)
			if !ok {
				continue
			}

			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     "Link relies on color alone to be distinguishable",
				Help:            "Add underline, bold, or other non-color indicator",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactModerate,
				Selector:        getString(issue, "selector"),
				HTML:            getString(issue, "html"),
				Element:         "a",
			})
		}
	}

	return findings, nil
}

type KeyboardAccessRule struct{ BaseRule }
func (r *KeyboardAccessRule) ID() string { return "keyboard-access" }
func (r *KeyboardAccessRule) Name() string { return "Keyboard Access" }
func (r *KeyboardAccessRule) Description() string { return "All functionality must be keyboard accessible" }
func (r *KeyboardAccessRule) SuccessCriteria() []string { return []string{"2.1.1"} }
func (r *KeyboardAccessRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *KeyboardAccessRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const issues = [];

		// Check for click handlers on non-focusable elements
		const clickableElements = document.querySelectorAll('[onclick], [onmousedown], [onmouseup]');
		clickableElements.forEach(el => {
			const tag = el.tagName.toLowerCase();
			const isFocusable = ['a', 'button', 'input', 'select', 'textarea'].includes(tag) ||
				el.hasAttribute('tabindex') ||
				el.getAttribute('role') === 'button' ||
				el.getAttribute('role') === 'link';

			if (!isFocusable) {
				issues.push({
					type: 'non-focusable-click',
					selector: el.id ? '#' + el.id : tag,
					html: el.outerHTML.substring(0, 150)
				});
			}
		});

		// Check for elements with role but missing tabindex
		const interactiveRoles = ['button', 'link', 'checkbox', 'radio', 'slider', 'tab', 'menuitem'];
		interactiveRoles.forEach(role => {
			document.querySelectorAll('[role="' + role + '"]').forEach(el => {
				const tag = el.tagName.toLowerCase();
				const isNativelyFocusable = ['a', 'button', 'input', 'select', 'textarea'].includes(tag);
				const hasTabindex = el.hasAttribute('tabindex');

				if (!isNativelyFocusable && !hasTabindex) {
					issues.push({
						type: 'missing-tabindex',
						selector: el.id ? '#' + el.id : '[role="' + role + '"]',
						html: el.outerHTML.substring(0, 150),
						role: role
					});
				}
			});
		});

		// Check for negative tabindex on interactive elements (can trap keyboard users)
		document.querySelectorAll('[tabindex]').forEach(el => {
			const tabindex = parseInt(el.getAttribute('tabindex'));
			if (tabindex < -1) {
				issues.push({
					type: 'invalid-tabindex',
					selector: el.id ? '#' + el.id : el.tagName.toLowerCase(),
					html: el.outerHTML.substring(0, 150),
					tabindex: tabindex
				});
			}
		});

		return issues.slice(0, 20);
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if issues, ok := result.([]any); ok {
		for _, item := range issues {
			issue, ok := item.(map[string]any)
			if !ok {
				continue
			}

			issueType := getString(issue, "type")
			var desc, help string

			switch issueType {
			case "non-focusable-click":
				desc = "Click handler on non-focusable element"
				help = "Use a <button> or add tabindex and keyboard handlers"
			case "missing-tabindex":
				role := getString(issue, "role")
				desc = fmt.Sprintf("Element with role='%s' is not focusable", role)
				help = "Add tabindex='0' to make the element keyboard accessible"
			case "invalid-tabindex":
				desc = "Element has invalid tabindex value"
				help = "Use tabindex='0' for focusable or tabindex='-1' for programmatic focus only"
			default:
				desc = "Keyboard accessibility issue"
				help = "Ensure element is keyboard accessible"
			}

			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     desc,
				Help:            help,
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactSerious,
				Selector:        getString(issue, "selector"),
				HTML:            getString(issue, "html"),
			})
		}
	}

	return findings, nil
}

type FocusVisibleRule struct{ BaseRule }
func (r *FocusVisibleRule) ID() string { return "focus-visible" }
func (r *FocusVisibleRule) Name() string { return "Focus Visible" }
func (r *FocusVisibleRule) Description() string { return "Focus indicator must be visible" }
func (r *FocusVisibleRule) SuccessCriteria() []string { return []string{"2.4.7"} }
func (r *FocusVisibleRule) Level() types.WCAGLevel { return types.WCAGLevelAA }

func (r *FocusVisibleRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const issues = [];

		// Check for elements that suppress focus outline
		const focusableElements = document.querySelectorAll('a, button, input, select, textarea, [tabindex]');

		focusableElements.forEach(el => {
			const style = getComputedStyle(el);

			// Check if outline is explicitly removed
			const outlineNone = style.outline === 'none' || style.outline === '0px none' ||
				style.outlineStyle === 'none' || style.outlineWidth === '0px';

			if (outlineNone) {
				// Check if there's an alternative focus indicator (box-shadow, border change, etc.)
				// This is a basic check - true focus visibility testing requires visual comparison
				const selector = el.id ? '#' + el.id : el.tagName.toLowerCase();
				issues.push({
					selector: selector,
					html: el.outerHTML.substring(0, 150),
					reason: 'outline-none'
				});
			}
		});

		// Check for CSS that might hide focus globally
		const stylesheets = Array.from(document.styleSheets);
		let hasFocusSuppressionRule = false;

		try {
			stylesheets.forEach(sheet => {
				if (!sheet.cssRules) return;
				Array.from(sheet.cssRules).forEach(rule => {
					if (rule.cssText && rule.cssText.includes(':focus') &&
						rule.cssText.includes('outline') &&
						(rule.cssText.includes('none') || rule.cssText.includes('0'))) {
						hasFocusSuppressionRule = true;
					}
				});
			});
		} catch(e) {
			// Cross-origin stylesheets may throw
		}

		return {
			elements: issues.slice(0, 15),
			hasFocusSuppressionRule: hasFocusSuppressionRule
		};
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if data, ok := result.(map[string]any); ok {
		// Check for global focus suppression
		if hasSuppression, _ := data["hasFocusSuppressionRule"].(bool); hasSuppression {
			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     "CSS rule may suppress focus indicators globally",
				Help:            "Ensure :focus styles provide visible indicators, not just outline:none",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactSerious,
				Element:         "style",
			})
		}

		// Check individual elements
		if elements, ok := data["elements"].([]any); ok {
			for _, item := range elements {
				el, ok := item.(map[string]any)
				if !ok {
					continue
				}

				findings = append(findings, types.Finding{
					RuleID:          r.ID(),
					Description:     "Element has no visible focus indicator",
					Help:            "Add visible focus styles (outline, box-shadow, or border)",
					SuccessCriteria: r.SuccessCriteria(),
					Level:           r.Level(),
					Impact:          types.ImpactSerious,
					Selector:        getString(el, "selector"),
					HTML:            getString(el, "html"),
				})
			}
		}
	}

	return findings, nil
}

type SkipLinkRule struct{ BaseRule }
func (r *SkipLinkRule) ID() string { return "skip-link" }
func (r *SkipLinkRule) Name() string { return "Skip Link" }
func (r *SkipLinkRule) Description() string { return "Page should have a skip link" }
func (r *SkipLinkRule) SuccessCriteria() []string { return []string{"2.4.1"} }
func (r *SkipLinkRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *SkipLinkRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const firstLink = document.querySelector('a[href^="#"]');
		const hasSkipLink = firstLink && (
			firstLink.textContent.toLowerCase().includes('skip') ||
			firstLink.textContent.toLowerCase().includes('jump') ||
			firstLink.getAttribute('href') === '#main' ||
			firstLink.getAttribute('href') === '#content'
		);
		return { hasSkipLink };
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if data, ok := result.(map[string]any); ok {
		hasSkipLink, _ := data["hasSkipLink"].(bool)
		if !hasSkipLink {
			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     "Page is missing a skip navigation link",
				Help:            "Add a skip link as the first focusable element",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactModerate,
				Element:         "a",
			})
		}
	}

	return findings, nil
}

type PageTitleRule struct{ BaseRule }
func (r *PageTitleRule) ID() string { return "page-title" }
func (r *PageTitleRule) Name() string { return "Page Title" }
func (r *PageTitleRule) Description() string { return "Page must have a title" }
func (r *PageTitleRule) SuccessCriteria() []string { return []string{"2.4.2"} }
func (r *PageTitleRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *PageTitleRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	title, err := vibe.Title(ctx)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if strings.TrimSpace(title) == "" {
		findings = append(findings, types.Finding{
			RuleID:          r.ID(),
			Description:     "Page is missing a title",
			Help:            "Add a descriptive <title> element",
			SuccessCriteria: r.SuccessCriteria(),
			Level:           r.Level(),
			Impact:          types.ImpactSerious,
			Element:         "title",
		})
	}

	return findings, nil
}

type LinkPurposeRule struct{ BaseRule }
func (r *LinkPurposeRule) ID() string { return "link-purpose" }
func (r *LinkPurposeRule) Name() string { return "Link Purpose" }
func (r *LinkPurposeRule) Description() string { return "Link purpose must be clear" }
func (r *LinkPurposeRule) SuccessCriteria() []string { return []string{"2.4.4"} }
func (r *LinkPurposeRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *LinkPurposeRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const vagueTexts = ['click here', 'here', 'read more', 'learn more', 'more', 'link'];
		Array.from(document.querySelectorAll('a')).filter(a => {
			const text = a.textContent.trim().toLowerCase();
			return vagueTexts.includes(text) && !a.hasAttribute('aria-label');
		}).map(a => ({
			selector: a.id ? '#' + a.id : 'a[href="' + a.getAttribute('href') + '"]',
			html: a.outerHTML.substring(0, 100),
			text: a.textContent.trim()
		}));
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if links, ok := result.([]any); ok {
		for _, item := range links {
			link, ok := item.(map[string]any)
			if !ok {
				continue
			}

			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     fmt.Sprintf("Link text '%s' is not descriptive", getString(link, "text")),
				Help:            "Use descriptive link text or add aria-label",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactModerate,
				Selector:        getString(link, "selector"),
				HTML:            getString(link, "html"),
				Element:         "a",
			})
		}
	}

	return findings, nil
}

type DescriptiveHeadingRule struct{ BaseRule }
func (r *DescriptiveHeadingRule) ID() string { return "descriptive-headings" }
func (r *DescriptiveHeadingRule) Name() string { return "Descriptive Headings" }
func (r *DescriptiveHeadingRule) Description() string { return "Headings should be descriptive" }
func (r *DescriptiveHeadingRule) SuccessCriteria() []string { return []string{"2.4.6"} }
func (r *DescriptiveHeadingRule) Level() types.WCAGLevel { return types.WCAGLevelAA }

func (r *DescriptiveHeadingRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const vaguePhrases = [
			'click here', 'read more', 'learn more', 'more', 'section', 'content',
			'header', 'title', 'heading', 'untitled', 'test', 'temp', 'placeholder'
		];
		const issues = [];

		document.querySelectorAll('h1, h2, h3, h4, h5, h6').forEach(heading => {
			const text = heading.textContent.trim().toLowerCase();

			// Check for empty or very short headings
			if (text.length === 0) {
				issues.push({
					selector: heading.id ? '#' + heading.id : heading.tagName.toLowerCase(),
					html: heading.outerHTML.substring(0, 100),
					issue: 'empty',
					text: ''
				});
				return;
			}

			if (text.length < 3) {
				issues.push({
					selector: heading.id ? '#' + heading.id : heading.tagName.toLowerCase(),
					html: heading.outerHTML.substring(0, 100),
					issue: 'too-short',
					text: heading.textContent.trim()
				});
				return;
			}

			// Check for vague headings
			if (vaguePhrases.some(phrase => text === phrase)) {
				issues.push({
					selector: heading.id ? '#' + heading.id : heading.tagName.toLowerCase(),
					html: heading.outerHTML.substring(0, 100),
					issue: 'vague',
					text: heading.textContent.trim()
				});
			}

			// Check for duplicate headings at same level
			const sameLevel = document.querySelectorAll(heading.tagName);
			let duplicateCount = 0;
			sameLevel.forEach(other => {
				if (other !== heading && other.textContent.trim().toLowerCase() === text) {
					duplicateCount++;
				}
			});

			if (duplicateCount > 0) {
				issues.push({
					selector: heading.id ? '#' + heading.id : heading.tagName.toLowerCase(),
					html: heading.outerHTML.substring(0, 100),
					issue: 'duplicate',
					text: heading.textContent.trim(),
					count: duplicateCount + 1
				});
			}
		});

		return issues.slice(0, 20);
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if issues, ok := result.([]any); ok {
		for _, item := range issues {
			issue, ok := item.(map[string]any)
			if !ok {
				continue
			}

			issueType := getString(issue, "issue")
			text := getString(issue, "text")
			var desc, help string

			switch issueType {
			case "empty":
				desc = "Heading is empty"
				help = "Add descriptive text to the heading"
			case "too-short":
				desc = fmt.Sprintf("Heading text '%s' is too short", text)
				help = "Use more descriptive heading text"
			case "vague":
				desc = fmt.Sprintf("Heading '%s' is not descriptive", text)
				help = "Use specific text that describes the section content"
			case "duplicate":
				count, _ := issue["count"].(float64)
				desc = fmt.Sprintf("Heading '%s' appears %d times at same level", text, int(count))
				help = "Use unique headings to differentiate sections"
			default:
				desc = "Heading may not be descriptive"
				help = "Ensure heading text describes the section content"
			}

			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     desc,
				Help:            help,
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactMinor,
				Selector:        getString(issue, "selector"),
				HTML:            getString(issue, "html"),
			})
		}
	}

	return findings, nil
}

type LanguageRule struct{ BaseRule }
func (r *LanguageRule) ID() string { return "html-lang" }
func (r *LanguageRule) Name() string { return "HTML Language" }
func (r *LanguageRule) Description() string { return "Page must have a lang attribute" }
func (r *LanguageRule) SuccessCriteria() []string { return []string{"3.1.1"} }
func (r *LanguageRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *LanguageRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `document.documentElement.getAttribute('lang')`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	lang, _ := result.(string)
	if strings.TrimSpace(lang) == "" {
		findings = append(findings, types.Finding{
			RuleID:          r.ID(),
			Description:     "HTML element is missing lang attribute",
			Help:            "Add lang attribute to <html> element",
			SuccessCriteria: r.SuccessCriteria(),
			Level:           r.Level(),
			Impact:          types.ImpactSerious,
			Element:         "html",
		})
	}

	return findings, nil
}

type DuplicateIDRule struct{ BaseRule }
func (r *DuplicateIDRule) ID() string { return "duplicate-id" }
func (r *DuplicateIDRule) Name() string { return "Duplicate IDs" }
func (r *DuplicateIDRule) Description() string { return "IDs must be unique" }
func (r *DuplicateIDRule) SuccessCriteria() []string { return []string{"4.1.1"} }
func (r *DuplicateIDRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *DuplicateIDRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const ids = {};
		const duplicates = [];
		document.querySelectorAll('[id]').forEach(el => {
			const id = el.id;
			if (ids[id]) {
				duplicates.push({
					id: id,
					selector: '#' + id,
					html: el.outerHTML.substring(0, 100)
				});
			}
			ids[id] = true;
		});
		return duplicates;
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if dups, ok := result.([]any); ok {
		for _, item := range dups {
			dup, ok := item.(map[string]any)
			if !ok {
				continue
			}

			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     fmt.Sprintf("Duplicate ID: %s", getString(dup, "id")),
				Help:            "Ensure all IDs are unique on the page",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactModerate,
				Selector:        getString(dup, "selector"),
				HTML:            getString(dup, "html"),
			})
		}
	}

	return findings, nil
}

type AriaLabelRule struct{ BaseRule }
func (r *AriaLabelRule) ID() string { return "aria-label" }
func (r *AriaLabelRule) Name() string { return "ARIA Labels" }
func (r *AriaLabelRule) Description() string { return "ARIA labels must be valid" }
func (r *AriaLabelRule) SuccessCriteria() []string { return []string{"4.1.2"} }
func (r *AriaLabelRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *AriaLabelRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const issues = [];

		// Check for empty aria-label
		document.querySelectorAll('[aria-label]').forEach(el => {
			const label = el.getAttribute('aria-label').trim();
			if (label === '') {
				issues.push({
					selector: el.id ? '#' + el.id : el.tagName.toLowerCase(),
					html: el.outerHTML.substring(0, 150),
					issue: 'empty-label'
				});
			}
		});

		// Check for aria-labelledby referencing non-existent IDs
		document.querySelectorAll('[aria-labelledby]').forEach(el => {
			const ids = el.getAttribute('aria-labelledby').split(/\s+/);
			const missingIds = ids.filter(id => !document.getElementById(id));

			if (missingIds.length > 0) {
				issues.push({
					selector: el.id ? '#' + el.id : el.tagName.toLowerCase(),
					html: el.outerHTML.substring(0, 150),
					issue: 'invalid-labelledby',
					missingIds: missingIds
				});
			}
		});

		// Check for aria-describedby referencing non-existent IDs
		document.querySelectorAll('[aria-describedby]').forEach(el => {
			const ids = el.getAttribute('aria-describedby').split(/\s+/);
			const missingIds = ids.filter(id => !document.getElementById(id));

			if (missingIds.length > 0) {
				issues.push({
					selector: el.id ? '#' + el.id : el.tagName.toLowerCase(),
					html: el.outerHTML.substring(0, 150),
					issue: 'invalid-describedby',
					missingIds: missingIds
				});
			}
		});

		// Check for invalid ARIA roles
		const validRoles = [
			'alert', 'alertdialog', 'application', 'article', 'banner', 'button',
			'cell', 'checkbox', 'columnheader', 'combobox', 'complementary',
			'contentinfo', 'definition', 'dialog', 'directory', 'document',
			'feed', 'figure', 'form', 'grid', 'gridcell', 'group', 'heading',
			'img', 'link', 'list', 'listbox', 'listitem', 'log', 'main',
			'marquee', 'math', 'menu', 'menubar', 'menuitem', 'menuitemcheckbox',
			'menuitemradio', 'navigation', 'none', 'note', 'option', 'presentation',
			'progressbar', 'radio', 'radiogroup', 'region', 'row', 'rowgroup',
			'rowheader', 'scrollbar', 'search', 'searchbox', 'separator', 'slider',
			'spinbutton', 'status', 'switch', 'tab', 'table', 'tablist', 'tabpanel',
			'term', 'textbox', 'timer', 'toolbar', 'tooltip', 'tree', 'treegrid', 'treeitem'
		];

		document.querySelectorAll('[role]').forEach(el => {
			const role = el.getAttribute('role').toLowerCase();
			if (!validRoles.includes(role)) {
				issues.push({
					selector: el.id ? '#' + el.id : el.tagName.toLowerCase(),
					html: el.outerHTML.substring(0, 150),
					issue: 'invalid-role',
					role: role
				});
			}
		});

		return issues.slice(0, 20);
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if issues, ok := result.([]any); ok {
		for _, item := range issues {
			issue, ok := item.(map[string]any)
			if !ok {
				continue
			}

			issueType := getString(issue, "issue")
			var desc, help string

			switch issueType {
			case "empty-label":
				desc = "aria-label attribute is empty"
				help = "Provide descriptive text or remove the attribute"
			case "invalid-labelledby":
				desc = "aria-labelledby references non-existent ID(s)"
				help = "Ensure referenced IDs exist in the document"
			case "invalid-describedby":
				desc = "aria-describedby references non-existent ID(s)"
				help = "Ensure referenced IDs exist in the document"
			case "invalid-role":
				role := getString(issue, "role")
				desc = fmt.Sprintf("Invalid ARIA role: %s", role)
				help = "Use a valid ARIA role from the WAI-ARIA specification"
			default:
				desc = "ARIA attribute issue"
				help = "Check ARIA attributes are valid and properly used"
			}

			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     desc,
				Help:            help,
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactSerious,
				Selector:        getString(issue, "selector"),
				HTML:            getString(issue, "html"),
			})
		}
	}

	return findings, nil
}

type ButtonNameRule struct{ BaseRule }
func (r *ButtonNameRule) ID() string { return "button-name" }
func (r *ButtonNameRule) Name() string { return "Button Name" }
func (r *ButtonNameRule) Description() string { return "Buttons must have accessible names" }
func (r *ButtonNameRule) SuccessCriteria() []string { return []string{"4.1.2"} }
func (r *ButtonNameRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *ButtonNameRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		Array.from(document.querySelectorAll('button, [role="button"]')).filter(btn => {
			const text = btn.textContent.trim();
			const ariaLabel = btn.getAttribute('aria-label');
			const title = btn.getAttribute('title');
			return !text && !ariaLabel && !title;
		}).map(btn => ({
			selector: btn.id ? '#' + btn.id : 'button',
			html: btn.outerHTML.substring(0, 100)
		}));
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if buttons, ok := result.([]any); ok {
		for _, item := range buttons {
			btn, ok := item.(map[string]any)
			if !ok {
				continue
			}

			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     "Button is missing an accessible name",
				Help:            "Add text content, aria-label, or title",
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactCritical,
				Selector:        getString(btn, "selector"),
				HTML:            getString(btn, "html"),
				Element:         "button",
			})
		}
	}

	return findings, nil
}

type FormInputNameRule struct{ BaseRule }
func (r *FormInputNameRule) ID() string { return "form-input-name" }
func (r *FormInputNameRule) Name() string { return "Form Input Name" }
func (r *FormInputNameRule) Description() string { return "Form inputs must have accessible names" }
func (r *FormInputNameRule) SuccessCriteria() []string { return []string{"4.1.2"} }
func (r *FormInputNameRule) Level() types.WCAGLevel { return types.WCAGLevelA }

func (r *FormInputNameRule) Run(ctx context.Context, vibe *vibium.Vibe) ([]types.Finding, error) {
	script := `
		const inputs = document.querySelectorAll(
			'input:not([type="hidden"]):not([type="submit"]):not([type="reset"]):not([type="button"]):not([type="image"]), ' +
			'select, textarea'
		);
		const issues = [];

		inputs.forEach(input => {
			// Check all possible sources of accessible name
			const id = input.id;
			const hasLabel = id && document.querySelector('label[for="' + id + '"]');
			const hasAriaLabel = input.hasAttribute('aria-label') && input.getAttribute('aria-label').trim();
			const hasAriaLabelledBy = input.hasAttribute('aria-labelledby');
			const hasTitle = input.hasAttribute('title') && input.getAttribute('title').trim();
			const hasPlaceholder = input.hasAttribute('placeholder') && input.getAttribute('placeholder').trim();

			// Check aria-labelledby validity
			let ariaLabelledByValid = false;
			if (hasAriaLabelledBy) {
				const ids = input.getAttribute('aria-labelledby').split(/\s+/);
				ariaLabelledByValid = ids.every(refId => document.getElementById(refId));
			}

			// Placeholder alone is not sufficient for accessible name
			const hasAccessibleName = hasLabel || hasAriaLabel || (hasAriaLabelledBy && ariaLabelledByValid) || hasTitle;

			if (!hasAccessibleName) {
				let issue = 'no-name';
				if (hasPlaceholder && !hasLabel && !hasAriaLabel && !hasTitle) {
					issue = 'placeholder-only';
				}

				issues.push({
					selector: id ? '#' + id : input.tagName.toLowerCase() + (input.name ? '[name="' + input.name + '"]' : ''),
					html: input.outerHTML.substring(0, 150),
					type: input.type || input.tagName.toLowerCase(),
					issue: issue
				});
			}
		});

		return issues.slice(0, 20);
	`

	result, err := vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	findings := make([]types.Finding, 0)
	if issues, ok := result.([]any); ok {
		for _, item := range issues {
			issue, ok := item.(map[string]any)
			if !ok {
				continue
			}

			issueType := getString(issue, "issue")
			inputType := getString(issue, "type")
			var desc, help string

			switch issueType {
			case "placeholder-only":
				desc = fmt.Sprintf("Form %s relies only on placeholder for accessible name", inputType)
				help = "Add a <label>, aria-label, or title attribute"
			default:
				desc = fmt.Sprintf("Form %s has no accessible name", inputType)
				help = "Add a <label> element or aria-label attribute"
			}

			findings = append(findings, types.Finding{
				RuleID:          r.ID(),
				Description:     desc,
				Help:            help,
				SuccessCriteria: r.SuccessCriteria(),
				Level:           r.Level(),
				Impact:          types.ImpactSerious,
				Selector:        getString(issue, "selector"),
				HTML:            getString(issue, "html"),
				Element:         inputType,
			})
		}
	}

	return findings, nil
}

// Rules provides a simplified interface for running WCAG accessibility rules.
type Rules struct {
	registry *Registry
	vibe     *vibium.Vibe
	logger   *slog.Logger
}

// NewRules creates a new Rules instance with all built-in rules.
func NewRules(vibe *vibium.Vibe, logger *slog.Logger) *Rules {
	return &Rules{
		registry: NewRegistry(logger),
		vibe:     vibe,
		logger:   logger,
	}
}

// RunAll runs all rules for the specified WCAG level and returns findings.
func (r *Rules) RunAll(ctx context.Context, level string) ([]types.Finding, error) {
	wcagLevel := types.WCAGLevel(level)
	rules := r.registry.GetByLevel(wcagLevel)

	var allFindings []types.Finding
	for _, rule := range rules {
		findings, err := rule.Run(ctx, r.vibe)
		if err != nil {
			r.logger.Warn("rule execution failed", "ruleId", rule.ID(), "error", err)
			continue
		}

		// Add finding IDs
		for i := range findings {
			findings[i].ID = fmt.Sprintf("%s-%d", rule.ID(), i)
		}

		allFindings = append(allFindings, findings...)
	}

	return allFindings, nil
}

// Helper functions

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
