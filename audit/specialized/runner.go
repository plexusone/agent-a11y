package specialized

import (
	"context"
	"fmt"
	"log/slog"

	vibium "github.com/agentplexus/vibium-go"
)

// Runner executes specialized accessibility tests.
type Runner struct {
	vibe   *vibium.Vibe
	logger *slog.Logger
}

// NewRunner creates a new specialized test runner.
func NewRunner(vibe *vibium.Vibe, logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.Default()
	}
	return &Runner{
		vibe:   vibe,
		logger: logger,
	}
}

// Finding represents a specialized test finding.
type Finding struct {
	ID              string   `json:"id"`
	RuleID          string   `json:"ruleId"`
	Description     string   `json:"description"`
	Help            string   `json:"help"`
	SuccessCriteria []string `json:"successCriteria"`
	Level           string   `json:"level"`
	Impact          string   `json:"impact"`
	Selector        string   `json:"selector,omitempty"`
	HTML            string   `json:"html,omitempty"`
}

// Results contains all specialized test results.
type Results struct {
	Keyboard      *KeyboardTestResult      `json:"keyboard,omitempty"`
	FocusVisible  *FocusVisibilityResult   `json:"focusVisible,omitempty"`
	FocusOrder    *FocusOrderResult        `json:"focusOrder,omitempty"`
	FocusObscured *FocusObscuredResult     `json:"focusObscured,omitempty"`
	OnFocus       *OnFocusResult           `json:"onFocus,omitempty"`
	Reflow        *ReflowTestResult        `json:"reflow,omitempty"`
	TargetSize    *TargetSizeTestResult    `json:"targetSize,omitempty"`
	TextSpacing   *SpacingTestResult       `json:"textSpacing,omitempty"`
	Findings      []Finding                `json:"findings"`
}

// RunAll executes all specialized tests and returns findings.
func (r *Runner) RunAll(ctx context.Context) (*Results, error) {
	results := &Results{
		Findings: []Finding{},
	}

	// Run keyboard accessibility test (2.1.1, 2.1.2)
	r.logger.Debug("running keyboard accessibility test")
	keyboardResult, err := TestKeyboardAccessibility(ctx, r.vibe, 100)
	if err != nil {
		r.logger.Warn("keyboard test failed", "error", err)
	} else {
		results.Keyboard = keyboardResult
		results.Findings = append(results.Findings, r.keyboardFindings(keyboardResult)...)
	}

	// Run focus visibility test (2.4.7)
	r.logger.Debug("running focus visibility test")
	focusVisResult, err := TestFocusVisibility(ctx, r.vibe, 50)
	if err != nil {
		r.logger.Warn("focus visibility test failed", "error", err)
	} else {
		results.FocusVisible = focusVisResult
		results.Findings = append(results.Findings, r.focusVisibilityFindings(focusVisResult)...)
	}

	// Run focus order test (2.4.3)
	r.logger.Debug("running focus order test")
	focusOrderResult, err := TestFocusOrder(ctx, r.vibe)
	if err != nil {
		r.logger.Warn("focus order test failed", "error", err)
	} else {
		results.FocusOrder = focusOrderResult
		results.Findings = append(results.Findings, r.focusOrderFindings(focusOrderResult)...)
	}

	// Run focus not obscured test (2.4.11)
	r.logger.Debug("running focus not obscured test")
	focusObscuredResult, err := TestFocusNotObscured(ctx, r.vibe)
	if err != nil {
		r.logger.Warn("focus not obscured test failed", "error", err)
	} else {
		results.FocusObscured = focusObscuredResult
		results.Findings = append(results.Findings, r.focusObscuredFindings(focusObscuredResult)...)
	}

	// Run on-focus context change test (3.2.1)
	r.logger.Debug("running on-focus test")
	onFocusResult, err := TestOnFocus(ctx, r.vibe)
	if err != nil {
		r.logger.Warn("on-focus test failed", "error", err)
	} else {
		results.OnFocus = onFocusResult
		results.Findings = append(results.Findings, r.onFocusFindings(onFocusResult)...)
	}

	// Run reflow test (1.4.10)
	r.logger.Debug("running reflow test")
	reflowResult, err := TestReflow(ctx, r.vibe)
	if err != nil {
		r.logger.Warn("reflow test failed", "error", err)
	} else {
		results.Reflow = reflowResult
		results.Findings = append(results.Findings, r.reflowFindings(reflowResult)...)
	}

	// Run target size test (2.5.8)
	r.logger.Debug("running target size test")
	targetResult, err := TestTargetSize(ctx, r.vibe, 24)
	if err != nil {
		r.logger.Warn("target size test failed", "error", err)
	} else {
		results.TargetSize = targetResult
		results.Findings = append(results.Findings, r.targetSizeFindings(targetResult)...)
	}

	// Run text spacing test (1.4.12)
	r.logger.Debug("running text spacing test")
	spacingResult, err := TestTextSpacing(ctx, r.vibe)
	if err != nil {
		r.logger.Warn("text spacing test failed", "error", err)
	} else {
		results.TextSpacing = spacingResult
		results.Findings = append(results.Findings, r.textSpacingFindings(spacingResult)...)
	}

	r.logger.Info("specialized tests complete", "findings", len(results.Findings))
	return results, nil
}

// Finding conversion helpers

func (r *Runner) keyboardFindings(result *KeyboardTestResult) []Finding {
	var findings []Finding

	if result.TrapDetected {
		findings = append(findings, Finding{
			ID:              "specialized-keyboard-trap",
			RuleID:          "keyboard-trap",
			Description:     fmt.Sprintf("Keyboard trap detected at %s", result.TrapLocation),
			Help:            "Ensure users can navigate away from all elements using keyboard",
			SuccessCriteria: []string{"2.1.2"},
			Level:           "A",
			Impact:          "critical",
			Selector:        result.TrapLocation,
		})
	}

	for _, el := range result.UnreachableElements {
		findings = append(findings, Finding{
			ID:              fmt.Sprintf("specialized-keyboard-unreachable-%s", el),
			RuleID:          "keyboard-unreachable",
			Description:     fmt.Sprintf("Interactive element not reachable via keyboard: %s", el),
			Help:            "Ensure all interactive elements can be reached via Tab navigation",
			SuccessCriteria: []string{"2.1.1"},
			Level:           "A",
			Impact:          "serious",
			Selector:        el,
		})
	}

	return findings
}

func (r *Runner) focusVisibilityFindings(result *FocusVisibilityResult) []Finding {
	var findings []Finding

	for _, el := range result.ElementsWithoutVisibleFocus {
		findings = append(findings, Finding{
			ID:              fmt.Sprintf("specialized-focus-visible-%s", el.Selector),
			RuleID:          "focus-visible",
			Description:     fmt.Sprintf("Focus indicator not visible on %s %s", el.TagName, el.Selector),
			Help:            "Ensure focus is visible when element receives keyboard focus",
			SuccessCriteria: []string{"2.4.7"},
			Level:           "AA",
			Impact:          "serious",
			Selector:        el.Selector,
		})
	}

	return findings
}

func (r *Runner) focusOrderFindings(result *FocusOrderResult) []Finding {
	var findings []Finding

	if result.HasPositiveTabindex {
		for _, el := range result.PositiveTabindexElements {
			findings = append(findings, Finding{
				ID:              fmt.Sprintf("specialized-tabindex-positive-%s", el),
				RuleID:          "tabindex-positive",
				Description:     fmt.Sprintf("Element has positive tabindex: %s", el),
				Help:            "Avoid positive tabindex values as they disrupt natural focus order",
				SuccessCriteria: []string{"2.4.3"},
				Level:           "A",
				Impact:          "moderate",
				Selector:        el,
			})
		}
	}

	if !result.HasLogicalOrder {
		for _, el := range result.OutOfOrderElements {
			findings = append(findings, Finding{
				ID:              fmt.Sprintf("specialized-focus-order-%s", el),
				RuleID:          "focus-order",
				Description:     fmt.Sprintf("Element appears out of visual order: %s", el),
				Help:            "Ensure focus order matches visual reading order",
				SuccessCriteria: []string{"2.4.3"},
				Level:           "A",
				Impact:          "moderate",
				Selector:        el,
			})
		}
	}

	return findings
}

func (r *Runner) focusObscuredFindings(result *FocusObscuredResult) []Finding {
	var findings []Finding

	for _, el := range result.PotentiallyObscured {
		findings = append(findings, Finding{
			ID:              fmt.Sprintf("specialized-focus-obscured-%s", el),
			RuleID:          "focus-obscured",
			Description:     fmt.Sprintf("Focused element may be obscured by sticky/fixed content: %s", el),
			Help:            "Ensure focused elements are not hidden behind sticky headers/footers",
			SuccessCriteria: []string{"2.4.11"},
			Level:           "AA",
			Impact:          "moderate",
			Selector:        el,
		})
	}

	return findings
}

func (r *Runner) onFocusFindings(result *OnFocusResult) []Finding {
	var findings []Finding

	for _, el := range result.ProblematicElements {
		findings = append(findings, Finding{
			ID:              fmt.Sprintf("specialized-onfocus-context-%s", el),
			RuleID:          "onfocus-context-change",
			Description:     fmt.Sprintf("Element may cause context change on focus: %s", el),
			Help:            "Focus should not automatically trigger navigation or form submission",
			SuccessCriteria: []string{"3.2.1"},
			Level:           "A",
			Impact:          "serious",
			Selector:        el,
		})
	}

	return findings
}

func (r *Runner) reflowFindings(result *ReflowTestResult) []Finding {
	var findings []Finding

	if !result.NoHorizontalScroll {
		findings = append(findings, Finding{
			ID:              "specialized-reflow-horizontal-scroll",
			RuleID:          "reflow-horizontal-scroll",
			Description:     fmt.Sprintf("Horizontal scrolling required at 320px width (overflow: %.0fpx)", result.HorizontalScrollWidth),
			Help:            "Content should reflow without requiring horizontal scrolling at 320px width",
			SuccessCriteria: []string{"1.4.10"},
			Level:           "AA",
			Impact:          "serious",
		})

		for _, el := range result.OverflowingElements {
			findings = append(findings, Finding{
				ID:              fmt.Sprintf("specialized-reflow-overflow-%s", el.Selector),
				RuleID:          "reflow-overflow",
				Description:     fmt.Sprintf("Element causes horizontal overflow: %s (width: %.0fpx, overflow: %.0fpx)", el.Selector, el.Width, el.Overflow),
				Help:            "Ensure elements resize or wrap at narrow viewport widths",
				SuccessCriteria: []string{"1.4.10"},
				Level:           "AA",
				Impact:          "moderate",
				Selector:        el.Selector,
			})
		}
	}

	if result.ContentLoss {
		for _, el := range result.LostElements {
			findings = append(findings, Finding{
				ID:              fmt.Sprintf("specialized-reflow-content-loss-%s", el),
				RuleID:          "reflow-content-loss",
				Description:     fmt.Sprintf("Content hidden at narrow viewport: %s", el),
				Help:            "Ensure all content remains accessible at 320px width",
				SuccessCriteria: []string{"1.4.10"},
				Level:           "AA",
				Impact:          "serious",
				Selector:        el,
			})
		}
	}

	return findings
}

func (r *Runner) targetSizeFindings(result *TargetSizeTestResult) []Finding {
	var findings []Finding

	for _, target := range result.SmallTargets {
		findings = append(findings, Finding{
			ID:              fmt.Sprintf("specialized-target-size-%s", target.Selector),
			RuleID:          "target-size-minimum",
			Description:     fmt.Sprintf("Touch target too small: %s (%.0fx%.0fpx, minimum 24x24px)", target.Selector, target.Width, target.Height),
			Help:            "Interactive elements should be at least 24x24 CSS pixels",
			SuccessCriteria: []string{"2.5.8"},
			Level:           "AA",
			Impact:          "moderate",
			Selector:        target.Selector,
		})
	}

	return findings
}

func (r *Runner) textSpacingFindings(result *SpacingTestResult) []Finding {
	var findings []Finding

	if result.ContentLoss {
		for _, el := range result.ClippedElements {
			findings = append(findings, Finding{
				ID:              fmt.Sprintf("specialized-text-spacing-%s", el.Selector),
				RuleID:          "text-spacing-loss",
				Description:     fmt.Sprintf("Text clipped when spacing increased: %s", el.Selector),
				Help:            "Content should adapt to increased text spacing without loss",
				SuccessCriteria: []string{"1.4.12"},
				Level:           "AA",
				Impact:          "serious",
				Selector:        el.Selector,
			})
		}
	}

	for _, overlap := range result.OverlappingElements {
		findings = append(findings, Finding{
			ID:              fmt.Sprintf("specialized-text-spacing-overlap-%s", overlap.Selector1),
			RuleID:          "text-spacing-overlap",
			Description:     fmt.Sprintf("Elements overlap when spacing increased: %s and %s", overlap.Selector1, overlap.Selector2),
			Help:            "Content should not overlap when text spacing is increased",
			SuccessCriteria: []string{"1.4.12"},
			Level:           "AA",
			Impact:          "moderate",
		})
	}

	return findings
}
