package engineering

// PRRisk holds the result of a PR risk analysis.
type PRRisk struct {
	PRUrl                string   `json:"pr_url"`
	ChangedPaths         []string `json:"changed_paths"`
	IncidentCorrelations []string `json:"incident_correlations"`
	RiskLevel            string   `json:"risk_level"` // low | medium | high | critical
	Reasoning            string   `json:"reasoning"`
}

// CycleTimeMetrics holds cycle time data for a team over a period.
type CycleTimeMetrics struct {
	Team            string  `json:"team"`
	AvgHours        float64 `json:"avg_hours"`
	P95Hours        float64 `json:"p95_hours"`
	BottleneckStage string  `json:"bottleneck_stage"`
}

// DeployRisk holds the result of a deploy risk assessment.
type DeployRisk struct {
	RiskLevel    string   `json:"risk_level"` // low | medium | high | critical
	HotPaths     []string `json:"hot_paths"`
	Reasoning    string   `json:"reasoning"`
}
