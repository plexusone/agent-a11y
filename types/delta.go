package types

// ValidationDelta compares two audit results to track fix progress.
// This enables the autonomous fix loop: audit → fix → validate → repeat.
type ValidationDelta struct {
	// Original audit results
	Before *AgentResult `json:"before,omitempty"`
	After  *AgentResult `json:"after,omitempty"`

	// Summary counts
	BeforeTotal int `json:"beforeTotal"`
	AfterTotal  int `json:"afterTotal"`

	// Categorized changes
	Fixed       []FixedFinding   `json:"fixed"`       // Issues that were resolved
	Remaining   []AgentFinding   `json:"remaining"`   // Issues still present
	Regressions []AgentFinding   `json:"regressions"` // New issues introduced

	// Overall status
	Status      DeltaStatus `json:"status"`      // Overall change status
	Improvement float64     `json:"improvement"` // Percentage improvement (can be negative)

	// Metadata
	BaselineURL string `json:"baselineUrl,omitempty"` // URL that was compared
	Timestamp   string `json:"timestamp"`             // When comparison was made
}

// DeltaStatus represents the overall change status.
type DeltaStatus string

const (
	// DeltaStatusFixed means all issues were resolved.
	DeltaStatusFixed DeltaStatus = "FIXED"

	// DeltaStatusImproved means some issues were fixed, none regressed.
	DeltaStatusImproved DeltaStatus = "IMPROVED"

	// DeltaStatusNoChange means nothing changed.
	DeltaStatusNoChange DeltaStatus = "NO_CHANGE"

	// DeltaStatusRegressed means new issues were introduced.
	DeltaStatusRegressed DeltaStatus = "REGRESSED"

	// DeltaStatusMixed means some fixed, some regressed.
	DeltaStatusMixed DeltaStatus = "MIXED"
)

// FixedFinding represents an issue that was resolved.
type FixedFinding struct {
	// The original finding that was fixed
	Finding AgentFinding `json:"finding"`

	// Which fix pattern likely resolved this (if determinable)
	FixedBy string `json:"fixedBy,omitempty"`

	// Whether fix was verified via element fingerprint match
	Verified bool `json:"verified"`

	// Fingerprint used for matching
	Fingerprint string `json:"fingerprint"`
}

// DeltaSummary provides a human-readable summary of changes.
type DeltaSummary struct {
	TotalBefore  int     `json:"totalBefore"`
	TotalAfter   int     `json:"totalAfter"`
	FixedCount   int     `json:"fixedCount"`
	NewCount     int     `json:"newCount"`     // Regressions
	Improvement  float64 `json:"improvement"`  // Percentage
	StatusEmoji  string  `json:"statusEmoji"`  // GO/WARN/NO-GO style
	StatusText   string  `json:"statusText"`   // Human readable
}

// Summary generates a human-readable summary of the delta.
func (d *ValidationDelta) Summary() DeltaSummary {
	s := DeltaSummary{
		TotalBefore: d.BeforeTotal,
		TotalAfter:  d.AfterTotal,
		FixedCount:  len(d.Fixed),
		NewCount:    len(d.Regressions),
		Improvement: d.Improvement,
	}

	switch d.Status {
	case DeltaStatusFixed:
		s.StatusEmoji = "✅"
		s.StatusText = "All issues fixed"
	case DeltaStatusImproved:
		s.StatusEmoji = "📈"
		s.StatusText = "Improved"
	case DeltaStatusNoChange:
		s.StatusEmoji = "➡️"
		s.StatusText = "No change"
	case DeltaStatusRegressed:
		s.StatusEmoji = "🔴"
		s.StatusText = "Regressed"
	case DeltaStatusMixed:
		s.StatusEmoji = "⚠️"
		s.StatusText = "Mixed results"
	}

	return s
}

// HasRegressions returns true if new issues were introduced.
func (d *ValidationDelta) HasRegressions() bool {
	return len(d.Regressions) > 0
}

// IsImproved returns true if the situation is better (or same).
func (d *ValidationDelta) IsImproved() bool {
	return d.Status == DeltaStatusFixed ||
		d.Status == DeltaStatusImproved ||
		d.Status == DeltaStatusNoChange
}

// IsPassing returns true if suitable for CI pass (no regressions).
func (d *ValidationDelta) IsPassing() bool {
	return !d.HasRegressions()
}
