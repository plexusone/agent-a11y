// Package remediation provides fix pattern mappings for accessibility issues.
// Maps WCAG techniques and common rule IDs to actionable fix patterns.
package remediation

import "github.com/plexusone/agent-a11y/types"

// PatternRegistry maps rule IDs to fix patterns.
type PatternRegistry struct {
	patterns map[string][]types.FixPattern
}

// NewPatternRegistry creates a registry with all known patterns.
func NewPatternRegistry() *PatternRegistry {
	r := &PatternRegistry{
		patterns: make(map[string][]types.FixPattern),
	}
	r.registerAllPatterns()
	return r
}

// GetPatterns returns fix patterns for a rule ID.
func (r *PatternRegistry) GetPatterns(ruleID string) []types.FixPattern {
	if patterns, ok := r.patterns[ruleID]; ok {
		return patterns
	}
	return nil
}

// GetPatternsForTechnique returns fix patterns for a WCAG technique.
func (r *PatternRegistry) GetPatternsForTechnique(techniqueID string) []types.FixPattern {
	key := "technique:" + techniqueID
	if patterns, ok := r.patterns[key]; ok {
		return patterns
	}
	return nil
}

func (r *PatternRegistry) registerAllPatterns() {
	// Color contrast issues
	r.registerColorContrast()

	// Image accessibility
	r.registerImageAlt()

	// Form accessibility
	r.registerFormLabels()

	// Link accessibility
	r.registerLinks()

	// Heading structure
	r.registerHeadings()

	// ARIA patterns
	r.registerARIA()

	// Keyboard accessibility
	r.registerKeyboard()

	// Focus management
	r.registerFocus()

	// Language
	r.registerLanguage()

	// Tables
	r.registerTables()

	// Media
	r.registerMedia()
}

func (r *PatternRegistry) registerColorContrast() {
	// axe-core: color-contrast
	r.patterns["color-contrast"] = []types.FixPattern{
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionModify,
			Target:   "color",
			Value:    "darker text color meeting 4.5:1 ratio",
			Example:  `color: var(--color-text-primary); /* Use design token */`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionModify,
			Target:   "background-color",
			Value:    "lighter background meeting 4.5:1 ratio",
			Example:  `background-color: var(--color-bg-primary); /* Use design token */`,
			Priority: 2,
		},
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionModify,
			Target:   "font-weight",
			Value:    "bolder weight (700+) allows 3:1 ratio for large text",
			Example:  `font-weight: 700;`,
			Priority: 3,
		},
	}

	// WCAG Technique G18: Ensuring contrast ratio of 4.5:1
	r.patterns["technique:G18"] = r.patterns["color-contrast"]

	// WCAG Technique G145: Ensuring contrast ratio of 3:1 for large text
	r.patterns["technique:G145"] = []types.FixPattern{
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionModify,
			Target:   "font-size",
			Value:    "18pt+ (24px) or 14pt+ bold for large text exception",
			Example:  `font-size: 1.5rem; /* 24px qualifies as large text */`,
			Priority: 1,
		},
	}

	// Non-text contrast (UI components)
	r.patterns["color-contrast-enhanced"] = []types.FixPattern{
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionModify,
			Target:   "border-color",
			Value:    "border meeting 3:1 ratio against background",
			Example:  `border: 2px solid var(--color-border-strong);`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionAdd,
			Target:   "outline",
			Value:    "visible outline for focus states",
			Example:  `outline: 2px solid var(--color-focus);`,
			Priority: 2,
		},
	}
}

func (r *PatternRegistry) registerImageAlt() {
	// axe-core: image-alt
	r.patterns["image-alt"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "alt",
			Value:    "descriptive alternative text",
			Example:  `<img src="photo.jpg" alt="Description of image content">`,
			Priority: 1,
		},
	}

	// Decorative images
	r.patterns["image-decorative"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "alt",
			Value:    "empty alt for decorative images",
			Example:  `<img src="decorative.svg" alt="" role="presentation">`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "role",
			Value:    "presentation",
			Example:  `role="presentation"`,
			Priority: 2,
		},
	}

	// WCAG Technique H37: Using alt attributes on img elements
	r.patterns["technique:H37"] = r.patterns["image-alt"]

	// WCAG Technique H67: Empty alt for decorative images
	r.patterns["technique:H67"] = r.patterns["image-decorative"]

	// SVG accessibility
	r.patterns["svg-img-alt"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "role",
			Value:    "img",
			Example:  `<svg role="img" aria-label="Description">`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-label",
			Value:    "descriptive label for SVG",
			Example:  `aria-label="Icon description"`,
			Priority: 2,
		},
	}
}

func (r *PatternRegistry) registerFormLabels() {
	// axe-core: label
	r.patterns["label"] = []types.FixPattern{
		{
			Type:     types.FixTypeStructure,
			Action:   types.FixActionWrap,
			Target:   "label",
			Value:    "wrap input in label or use for attribute",
			Example:  `<label for="input-id">Label text</label>\n<input id="input-id">`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-label",
			Value:    "accessible name via aria-label",
			Example:  `<input aria-label="Field description">`,
			Priority: 2,
		},
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-labelledby",
			Value:    "reference existing visible label",
			Example:  `<input aria-labelledby="label-id">`,
			Priority: 3,
		},
	}

	// WCAG Technique H44: Using label elements
	r.patterns["technique:H44"] = r.patterns["label"]

	// Required field indication
	r.patterns["aria-required-attr"] = []types.FixPattern{
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-required",
			Value:    "true",
			Example:  `<input aria-required="true">`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "required",
			Value:    "HTML5 required attribute",
			Example:  `<input required>`,
			Priority: 2,
		},
	}

	// Form error identification
	r.patterns["aria-input-field-name"] = []types.FixPattern{
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-describedby",
			Value:    "reference error message element",
			Example:  `<input aria-describedby="error-msg">\n<span id="error-msg">Error text</span>`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-invalid",
			Value:    "true when field has error",
			Example:  `<input aria-invalid="true">`,
			Priority: 2,
		},
	}

	// Autocomplete
	r.patterns["autocomplete-valid"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "autocomplete",
			Value:    "appropriate autocomplete token",
			Example:  `<input type="email" autocomplete="email">`,
			Priority: 1,
		},
	}
}

func (r *PatternRegistry) registerLinks() {
	// axe-core: link-name
	r.patterns["link-name"] = []types.FixPattern{
		{
			Type:     types.FixTypeContent,
			Action:   types.FixActionAdd,
			Target:   "text content",
			Value:    "descriptive link text",
			Example:  `<a href="/page">Descriptive link text</a>`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-label",
			Value:    "accessible name for icon-only links",
			Example:  `<a href="/page" aria-label="Go to page">\n  <svg>...</svg>\n</a>`,
			Priority: 2,
		},
	}

	// Link in context
	r.patterns["link-in-text-block"] = []types.FixPattern{
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionAdd,
			Target:   "text-decoration",
			Value:    "underline to distinguish from surrounding text",
			Example:  `a { text-decoration: underline; }`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionModify,
			Target:   "color",
			Value:    "distinct color with 3:1 contrast from surrounding text",
			Example:  `a { color: var(--color-link); }`,
			Priority: 2,
		},
	}

	// WCAG Technique H30: Providing link text
	r.patterns["technique:H30"] = r.patterns["link-name"]
}

func (r *PatternRegistry) registerHeadings() {
	// Heading order
	r.patterns["heading-order"] = []types.FixPattern{
		{
			Type:     types.FixTypeSemantic,
			Action:   types.FixActionModify,
			Target:   "heading level",
			Value:    "use sequential heading levels (h1 → h2 → h3)",
			Example:  `<!-- Change h4 to h2 if it follows h1 -->\n<h1>Title</h1>\n<h2>Section</h2>`,
			Priority: 1,
		},
	}

	// Empty heading
	r.patterns["empty-heading"] = []types.FixPattern{
		{
			Type:     types.FixTypeContent,
			Action:   types.FixActionAdd,
			Target:   "text content",
			Value:    "add heading text or remove empty heading",
			Example:  `<h2>Section Title</h2>`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeStructure,
			Action:   types.FixActionRemove,
			Target:   "element",
			Value:    "remove if heading is not needed",
			Example:  `<!-- Remove empty <h2></h2> -->`,
			Priority: 2,
		},
	}

	// Page has h1
	r.patterns["page-has-heading-one"] = []types.FixPattern{
		{
			Type:     types.FixTypeStructure,
			Action:   types.FixActionAdd,
			Target:   "h1",
			Value:    "add h1 as main page heading",
			Example:  `<main>\n  <h1>Page Title</h1>\n  ...\n</main>`,
			Priority: 1,
		},
	}

	// WCAG Technique G141: Organizing a page using headings
	r.patterns["technique:G141"] = r.patterns["heading-order"]
}

func (r *PatternRegistry) registerARIA() {
	// aria-hidden on focusable
	r.patterns["aria-hidden-focus"] = []types.FixPattern{
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionRemove,
			Target:   "aria-hidden",
			Value:    "remove aria-hidden from focusable elements",
			Example:  `<!-- Remove aria-hidden="true" from button -->\n<button>Click me</button>`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "tabindex",
			Value:    "-1 to remove from focus order if truly hidden",
			Example:  `<div aria-hidden="true" tabindex="-1">Hidden content</div>`,
			Priority: 2,
		},
	}

	// aria-valid-attr-value
	r.patterns["aria-valid-attr-value"] = []types.FixPattern{
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionModify,
			Target:   "aria-*",
			Value:    "correct value per ARIA specification",
			Example:  `aria-expanded="true" <!-- not "yes" -->`,
			Priority: 1,
		},
	}

	// aria-roles
	r.patterns["aria-roles"] = []types.FixPattern{
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionModify,
			Target:   "role",
			Value:    "valid ARIA role",
			Example:  `role="button" <!-- valid role -->`,
			Priority: 1,
		},
	}

	// aria-required-children
	r.patterns["aria-required-children"] = []types.FixPattern{
		{
			Type:     types.FixTypeStructure,
			Action:   types.FixActionAdd,
			Target:   "child element",
			Value:    "required child role (e.g., listitem for list)",
			Example:  `<ul role="list">\n  <li role="listitem">Item</li>\n</ul>`,
			Priority: 1,
		},
	}

	// WCAG Technique ARIA6: Using aria-label
	r.patterns["technique:ARIA6"] = []types.FixPattern{
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-label",
			Value:    "accessible name for element",
			Example:  `<button aria-label="Close dialog">×</button>`,
			Priority: 1,
		},
	}

	// WCAG Technique ARIA11: Using aria-labelledby
	r.patterns["technique:ARIA11"] = []types.FixPattern{
		{
			Type:     types.FixTypeARIA,
			Action:   types.FixActionAdd,
			Target:   "aria-labelledby",
			Value:    "reference to labelling element ID",
			Example:  `<h2 id="section-title">Section</h2>\n<div aria-labelledby="section-title">...</div>`,
			Priority: 1,
		},
	}
}

func (r *PatternRegistry) registerKeyboard() {
	// Keyboard focusable
	r.patterns["focus-order-semantics"] = []types.FixPattern{
		{
			Type:     types.FixTypeSemantic,
			Action:   types.FixActionReplace,
			Target:   "element",
			Value:    "use native interactive element",
			Example:  `<!-- Replace div with button -->\n<button onclick="...">Click</button>`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "tabindex",
			Value:    "0 to add to focus order",
			Example:  `<div role="button" tabindex="0">Click</div>`,
			Priority: 2,
		},
	}

	// Scrollable region focusable
	r.patterns["scrollable-region-focusable"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "tabindex",
			Value:    "0 for scrollable regions",
			Example:  `<div class="scrollable" tabindex="0">...</div>`,
			Priority: 1,
		},
	}

	// Keyboard trap
	r.patterns["no-trap-focus"] = []types.FixPattern{
		{
			Type:     types.FixTypeFocus,
			Action:   types.FixActionModify,
			Target:   "focus management",
			Value:    "ensure Escape key or other mechanism exits",
			Example:  `dialog.addEventListener('keydown', (e) => {\n  if (e.key === 'Escape') dialog.close();\n});`,
			Priority: 1,
		},
	}

	// WCAG Technique G202: Ensuring keyboard control
	r.patterns["technique:G202"] = r.patterns["focus-order-semantics"]
}

func (r *PatternRegistry) registerFocus() {
	// Focus visible
	r.patterns["focus-visible"] = []types.FixPattern{
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionAdd,
			Target:   "outline",
			Value:    "visible focus indicator",
			Example:  `:focus-visible {\n  outline: 2px solid var(--color-focus);\n  outline-offset: 2px;\n}`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionRemove,
			Target:   "outline: none",
			Value:    "remove outline suppression",
			Example:  `/* Remove: outline: none; */`,
			Priority: 2,
		},
	}

	// Focus order
	r.patterns["tabindex"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionModify,
			Target:   "tabindex",
			Value:    "0 or -1 only (avoid positive values)",
			Example:  `<!-- Use tabindex="0" not tabindex="5" -->\n<button tabindex="0">Click</button>`,
			Priority: 1,
		},
	}

	// WCAG Technique G149: Focus visible on user interface components
	r.patterns["technique:G149"] = r.patterns["focus-visible"]

	// WCAG Technique C15: Using CSS to change presentation on focus
	r.patterns["technique:C15"] = []types.FixPattern{
		{
			Type:     types.FixTypeStyle,
			Action:   types.FixActionAdd,
			Target:   ":focus",
			Value:    "CSS focus styles",
			Example:  `button:focus {\n  background: var(--color-bg-focus);\n  outline: 2px solid var(--color-focus);\n}`,
			Priority: 1,
		},
	}
}

func (r *PatternRegistry) registerLanguage() {
	// html-has-lang
	r.patterns["html-has-lang"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "lang",
			Value:    "valid language code",
			Example:  `<html lang="en">`,
			Priority: 1,
		},
	}

	// html-lang-valid
	r.patterns["html-lang-valid"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionModify,
			Target:   "lang",
			Value:    "valid BCP 47 language tag",
			Example:  `<html lang="en-US"> <!-- or "en", "es", "fr", etc. -->`,
			Priority: 1,
		},
	}

	// WCAG Technique H57: Using the language attribute
	r.patterns["technique:H57"] = r.patterns["html-has-lang"]
}

func (r *PatternRegistry) registerTables() {
	// Table headers
	r.patterns["th-has-data-cells"] = []types.FixPattern{
		{
			Type:     types.FixTypeSemantic,
			Action:   types.FixActionAdd,
			Target:   "th",
			Value:    "use th for header cells",
			Example:  `<table>\n  <thead>\n    <tr><th>Header 1</th><th>Header 2</th></tr>\n  </thead>\n  <tbody>\n    <tr><td>Data 1</td><td>Data 2</td></tr>\n  </tbody>\n</table>`,
			Priority: 1,
		},
	}

	// Table scope
	r.patterns["td-headers-attr"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "scope",
			Value:    "col or row for header cells",
			Example:  `<th scope="col">Column Header</th>`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "headers",
			Value:    "reference header IDs for complex tables",
			Example:  `<td headers="header1 header2">Data</td>`,
			Priority: 2,
		},
	}

	// WCAG Technique H43: Using id and headers for data cells
	r.patterns["technique:H43"] = r.patterns["td-headers-attr"]

	// WCAG Technique H63: Using scope attribute
	r.patterns["technique:H63"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "scope",
			Value:    "col, row, colgroup, or rowgroup",
			Example:  `<th scope="col">Column</th>\n<th scope="row">Row</th>`,
			Priority: 1,
		},
	}
}

func (r *PatternRegistry) registerMedia() {
	// Video captions
	r.patterns["video-caption"] = []types.FixPattern{
		{
			Type:     types.FixTypeStructure,
			Action:   types.FixActionAdd,
			Target:   "track",
			Value:    "captions track element",
			Example:  `<video>\n  <source src="video.mp4">\n  <track kind="captions" src="captions.vtt" srclang="en" label="English">\n</video>`,
			Priority: 1,
		},
	}

	// Audio description
	r.patterns["video-description"] = []types.FixPattern{
		{
			Type:     types.FixTypeStructure,
			Action:   types.FixActionAdd,
			Target:   "track",
			Value:    "descriptions track element",
			Example:  `<video>\n  <track kind="descriptions" src="descriptions.vtt" srclang="en">\n</video>`,
			Priority: 1,
		},
	}

	// Autoplay
	r.patterns["no-autoplay-audio"] = []types.FixPattern{
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionRemove,
			Target:   "autoplay",
			Value:    "remove autoplay or add muted",
			Example:  `<video muted autoplay> <!-- muted makes autoplay acceptable -->`,
			Priority: 1,
		},
		{
			Type:     types.FixTypeAttribute,
			Action:   types.FixActionAdd,
			Target:   "muted",
			Value:    "mute autoplay media",
			Example:  `<video autoplay muted>`,
			Priority: 2,
		},
	}

	// WCAG Technique G87: Providing closed captions
	r.patterns["technique:G87"] = r.patterns["video-caption"]

	// WCAG Technique G78: Providing audio description
	r.patterns["technique:G78"] = r.patterns["video-description"]
}
