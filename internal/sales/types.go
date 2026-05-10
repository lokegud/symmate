package sales

// PipelineStage represents where a deal sits in the sales funnel.
type PipelineStage string

const (
	StageProspecting  PipelineStage = "prospecting"
	StageQualified    PipelineStage = "qualified"
	StageProposal     PipelineStage = "proposal"
	StageNegotiation  PipelineStage = "negotiation"
	StageClosedWon    PipelineStage = "closed_won"
	StageClosedLost   PipelineStage = "closed_lost"
)

// Deal represents a single sales opportunity.
type Deal struct {
	ID          string        `json:"id"`
	AccountName string        `json:"account_name"`
	OwnerEmail  string        `json:"owner_email"`
	Stage       PipelineStage `json:"stage"`
	ARR         float64       `json:"arr"`
	CloseDate   string        `json:"close_date"`
	Notes       string        `json:"notes"`
}

// ForecastSummary aggregates pipeline metrics for a reporting period.
type ForecastSummary struct {
	Period       string  `json:"period"`
	TotalARR     float64 `json:"total_arr"`
	ClosedWonARR float64 `json:"closed_won_arr"`
	PipelineARR  float64 `json:"pipeline_arr"`
	WinRate      float64 `json:"win_rate"`
	TopDeals     []Deal  `json:"top_deals"`
}

// AccountHealth captures engagement and churn-risk signals for an account.
type AccountHealth struct {
	AccountName        string `json:"account_name"`
	HealthScore        int    `json:"health_score"`
	ChurnRisk          string `json:"churn_risk"` // low | medium | high
	LastActivityDate   string `json:"last_activity_date"`
	OpenOpportunities  int    `json:"open_opportunities"`
}
