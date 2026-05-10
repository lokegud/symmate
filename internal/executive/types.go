package executive

// CompanyHealthDigest holds a cross-functional snapshot of company health.
type CompanyHealthDigest struct {
	Period          string            `json:"period"`           // ISO date range, e.g. "2026-04-01/2026-04-30"
	DepartmentScores map[string]int   `json:"department_scores"` // department → 0-100 health score
	TopRisks        []string          `json:"top_risks"`
	TopMomentum     []string          `json:"top_momentum"`
	OverallScore    int               `json:"overall_score"` // 0-100
}

// OKRGapReport holds OKR progress and gap analysis for a department or company-wide.
type OKRGapReport struct {
	Department  string     `json:"department"`
	Quarter     string     `json:"quarter"` // e.g. "Q2-2026"
	Objectives  []OKRItem  `json:"objectives"`
	AtRiskCount int        `json:"at_risk_count"`
	OnTrackCount int       `json:"on_track_count"`
}

// OKRItem is a single objective with its key result summary.
type OKRItem struct {
	Objective  string  `json:"objective"`
	Progress   float64 `json:"progress"`   // 0.0–1.0
	Status     string  `json:"status"`     // on_track | at_risk | off_track
	GapSummary string  `json:"gap_summary"`
}

// CrossFunctionalSignal captures a notable signal that spans multiple departments.
type CrossFunctionalSignal struct {
	Signal      string   `json:"signal"`
	Departments []string `json:"departments"`
	Severity    string   `json:"severity"` // info | warning | critical
	Recommended string   `json:"recommended"`
}
