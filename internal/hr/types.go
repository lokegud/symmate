package hr

// OnboardingStatus holds the state of an employee's onboarding flow.
type OnboardingStatus struct {
	EmployeeID     string   `json:"employee_id"`
	Name           string   `json:"name"`
	StartDate      string   `json:"start_date"`
	CompletedSteps []string `json:"completed_steps"`
	PendingSteps   []string `json:"pending_steps"`
	OverallStatus  string   `json:"overall_status"` // on-track | at-risk | blocked
}

// EngagementRisk holds the result of an engagement risk assessment for an employee.
type EngagementRisk struct {
	EmployeeID  string   `json:"employee_id"`
	Name        string   `json:"name"`
	RiskLevel   string   `json:"risk_level"` // low | medium | high | critical
	Signals     []string `json:"signals"`
	Reasoning   string   `json:"reasoning"`
}

// HeadcountGap holds an identified gap between current headcount and open roles.
type HeadcountGap struct {
	Department   string `json:"department"`
	CurrentCount int    `json:"current_count"`
	TargetCount  int    `json:"target_count"`
	OpenRoles    int    `json:"open_roles"`
	Urgency      string `json:"urgency"` // low | medium | high
}
