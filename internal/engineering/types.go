package engineering

// PRRisk holds the result of a PR risk analysis.
type PRRisk struct {
	PRUrl        string   `json:"pr_url"`
	Title        string   `json:"title,omitempty"`
	Author       string   `json:"author,omitempty"`
	State        string   `json:"state,omitempty"`
	Additions    int      `json:"additions"`
	Deletions    int      `json:"deletions"`
	ChangedPaths []string `json:"changed_paths"`
	RiskLevel    string   `json:"risk_level"` // low | medium | high
	Signals      []string `json:"signals"`    // the individual reasons behind RiskLevel
	Reasoning    string   `json:"reasoning"`
	Source       string   `json:"source"` // "github" or "provided-file-list"
}

// CycleTimeMetrics holds PR cycle time for a repository over a period.
type CycleTimeMetrics struct {
	Repo                  string  `json:"repo"`
	From                  string  `json:"from"`
	To                    string  `json:"to"`
	PRsMerged             int     `json:"prs_merged"`
	AvgHours              float64 `json:"avg_hours"`    // PR opened → merged
	MedianHours           float64 `json:"median_hours"` // PR opened → merged
	P95Hours              float64 `json:"p95_hours"`    // PR opened → merged
	AvgHoursToFirstReview float64 `json:"avg_hours_to_first_review,omitempty"`
	AvgHoursReviewToMerge float64 `json:"avg_hours_review_to_merge,omitempty"`
	ReviewSampleSize      int     `json:"review_sample_size"`
	BottleneckStage       string  `json:"bottleneck_stage"` // waiting-for-review | review-to-merge | unknown
	Notes                 string  `json:"notes,omitempty"`
}

// DeployRisk holds the result of a deploy risk assessment.
type DeployRisk struct {
	RiskLevel string   `json:"risk_level"` // low | medium | high
	HotPaths  []string `json:"hot_paths"`
	Signals   []string `json:"signals"`
	Reasoning string   `json:"reasoning"`
}
