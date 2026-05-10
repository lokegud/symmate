package finance

// CloseChecklistItem represents a single line item in the monthly close checklist.
type CloseChecklistItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`   // open | in_progress | complete | blocked
	Owner    string `json:"owner"`
	DueDate  string `json:"due_date"` // ISO 8601
	Notes    string `json:"notes,omitempty"`
}

// SpendAnomaly represents a detected deviation in spend vs. expected baseline.
type SpendAnomaly struct {
	Category      string  `json:"category"`
	Vendor        string  `json:"vendor"`
	ExpectedUSD   float64 `json:"expected_usd"`
	ActualUSD     float64 `json:"actual_usd"`
	DeviationPct  float64 `json:"deviation_pct"`
	Severity      string  `json:"severity"` // low | medium | high
	Reasoning     string  `json:"reasoning"`
}

// ForecastDelta captures the variance between a financial forecast and actuals for a period.
type ForecastDelta struct {
	Period       string  `json:"period"`        // e.g. "2026-04"
	Metric       string  `json:"metric"`        // e.g. "revenue" | "opex" | "gross_margin"
	ForecastUSD  float64 `json:"forecast_usd"`
	ActualUSD    float64 `json:"actual_usd"`
	DeltaUSD     float64 `json:"delta_usd"`
	DeltaPct     float64 `json:"delta_pct"`
	Commentary   string  `json:"commentary"`
}
