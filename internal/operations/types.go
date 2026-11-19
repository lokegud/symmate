package operations

// VendorSLAStatus holds the current SLA health for a vendor.
type VendorSLAStatus struct {
	VendorName     string  `json:"vendor_name"`
	SLATarget      float64 `json:"sla_target"`       // e.g. 99.9 (percent uptime)
	CurrentUptime  float64 `json:"current_uptime"`   // trailing 30d percent
	BreachCount    int     `json:"breach_count"`     // SLA breaches in period
	Status         string  `json:"status"`           // ok | at_risk | breached
	NextReviewDate string  `json:"next_review_date"` // ISO date
}

// DependencyBlocker represents a cross-team dependency that is blocking progress.
type DependencyBlocker struct {
	BlockerID      string `json:"blocker_id"`
	Title          string `json:"title"`
	OwnerTeam      string `json:"owner_team"`
	BlockedTeam    string `json:"blocked_team"`
	Severity       string `json:"severity"` // low | medium | high | critical
	DaysBlocked    int    `json:"days_blocked"`
	Resolution     string `json:"resolution"` // empty until resolved
}

// OperationsHealthSummary gives an aggregate view of ops health for a time window.
type OperationsHealthSummary struct {
	Period          string              `json:"period"`           // ISO date range
	ActiveBlockers  []DependencyBlocker `json:"active_blockers"`
	VendorsAtRisk   []string            `json:"vendors_at_risk"`
	OverallStatus   string              `json:"overall_status"`   // healthy | degraded | critical
	TopRisk         string              `json:"top_risk"`
}
