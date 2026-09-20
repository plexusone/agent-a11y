package wcag

import "testing"

// specializedRunnableCriteria are the success criteria the specialized test
// runner actually executes (see audit/specialized/runner.go RunAll). A
// criterion may only be classified MethodSpecialized if it appears here —
// otherwise a clean audit would falsely report "Supports" for a criterion no
// check ever evaluated. Keep this set in sync with RunAll.
var specializedRunnableCriteria = map[string]bool{
	"1.4.4":  true, // resize text
	"1.4.10": true, // reflow
	"1.4.12": true, // text spacing
	"1.4.13": true, // content on hover/focus
	"2.1.1":  true, // keyboard reachable
	"2.1.2":  true, // no keyboard trap
	"2.4.3":  true, // focus order
	"2.4.7":  true, // focus visible
	"2.4.11": true, // focus not obscured
	"2.5.8":  true, // target size
	"3.2.1":  true, // on focus
}

// TestSpecializedCriteriaAreActuallyRun guards against overclaiming: every
// criterion marked MethodSpecialized must be one the runner truly evaluates.
func TestSpecializedCriteriaAreActuallyRun(t *testing.T) {
	for _, c := range WCAG22AA {
		if c.Method == MethodSpecialized && !specializedRunnableCriteria[c.ID] {
			t.Errorf("%s (%s) is MethodSpecialized but not run by the specialized runner; "+
				"reclassify to LLM-Judge/Manual or wire a check into RunAll", c.ID, c.Name)
		}
	}
}

// TestAutomatedCriteriaHaveAxeRules guards against claiming automated coverage
// for a criterion with no axe rule backing it.
func TestAutomatedCriteriaHaveAxeRules(t *testing.T) {
	for _, c := range WCAG22AA {
		if c.Method == MethodAutomated && len(c.AxeRules) == 0 {
			t.Errorf("%s (%s) is MethodAutomated but has no AxeRules", c.ID, c.Name)
		}
	}
}
