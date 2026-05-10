package customersuccess

// HealthTier is a traffic-light classification for customer health.
type HealthTier string

const (
	HealthGreen  HealthTier = "green"
	HealthYellow HealthTier = "yellow"
	HealthRed    HealthTier = "red"
)

// CustomerRecord holds the full CS profile for an account.
type CustomerRecord struct {
	AccountName         string     `json:"account_name"`
	CSMID               string     `json:"csm_id"`
	HealthTier          HealthTier `json:"health_tier"`
	NPS                 int        `json:"nps"`
	ARR                 float64    `json:"arr"`
	ContractRenewalDate string     `json:"contract_renewal_date"`
	OpenTickets         int        `json:"open_tickets"`
	LastQBRDate         string     `json:"last_qbr_date"`
}

// OnboardingStatus tracks where a customer is in the implementation funnel.
type OnboardingStatus struct {
	AccountName     string `json:"account_name"`
	Stage           string `json:"stage"` // kickoff | implementation | training | go_live | complete
	PercentComplete int    `json:"percent_complete"`
	BlockedReason   string `json:"blocked_reason"`
}

// ChurnAlert surfaces an at-risk account with risk signals and a recommended action.
type ChurnAlert struct {
	AccountName         string   `json:"account_name"`
	RiskScore           int      `json:"risk_score"` // 0-100
	Signals             []string `json:"signals"`
	RecommendedAction   string   `json:"recommended_action"`
}
