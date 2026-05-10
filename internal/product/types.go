package product

// FeedbackCluster holds a synthesized theme derived from raw customer feedback.
type FeedbackCluster struct {
	Theme      string   `json:"theme"`
	Count      int      `json:"count"`
	Sentiment  string   `json:"sentiment"` // positive | neutral | negative
	Verbatims  []string `json:"verbatims"`
	NPSImpact  float64  `json:"nps_impact"` // estimated point delta if addressed
}

// RoadmapGap represents a customer need that has no current roadmap coverage.
type RoadmapGap struct {
	Need          string  `json:"need"`
	FeedbackCount int     `json:"feedback_count"`
	Priority      string  `json:"priority"` // low | medium | high | critical
	Reasoning     string  `json:"reasoning"`
}

// NPSSummary holds Net Promoter Score data for a segment over a period.
type NPSSummary struct {
	Segment    string  `json:"segment"`
	DateRange  string  `json:"date_range"`
	Score      float64 `json:"score"`       // -100 to 100
	Promoters  int     `json:"promoters"`   // 9-10 responses
	Passives   int     `json:"passives"`    // 7-8 responses
	Detractors int     `json:"detractors"`  // 0-6 responses
	TopDrivers []string `json:"top_drivers"`
}
