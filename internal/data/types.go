package data

// MetricDefinition is a canonical metric registered in the metric registry.
type MetricDefinition struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
	Category    string `json:"category"` // revenue | engagement | ops | quality
}

// MetricConsistencyResult holds the result of a cross-source consistency check for a metric.
type MetricConsistencyResult struct {
	MetricName string            `json:"metric_name"`
	Sources    map[string]string `json:"sources"` // source_name → reported_value
	Consistent bool              `json:"consistent"`
	Discrepancy string           `json:"discrepancy,omitempty"`
	Severity    string           `json:"severity"` // none | low | high | critical
}

// DataQualityReport summarises quality issues found in a dataset or table.
type DataQualityReport struct {
	Dataset     string   `json:"dataset"`
	Table       string   `json:"table"`
	NullRate    float64  `json:"null_rate"`    // 0–1 fraction of null values in key columns
	DuplicateRate float64 `json:"duplicate_rate"` // 0–1 fraction of duplicate rows
	Anomalies   []string `json:"anomalies"`
	Health      string   `json:"health"` // healthy | degraded | critical
}
