package tool

import (
	"context"
	"encoding/json"

	"github.com/loke/symmate/internal/data"
)

// LookupMetricTool returns a tool that retrieves a canonical metric definition by name.
func LookupMetricTool() Tool {
	return Tool{
		Def: toolDef("lookup_metric",
			"Looks up a canonical metric definition from the company metric registry.",
			`{
				"type": "object",
				"properties": {
					"metric_name": {"type": "string", "description": "Canonical name of the metric, e.g. \"monthly_active_users\"."}
				},
				"required": ["metric_name"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				MetricName string `json:"metric_name"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockMetricDefinition(p.MetricName))
		},
		Skill: "- Use lookup_metric to retrieve the canonical definition, owner, unit, and category of any company metric before using it in analysis.",
		Tags:  []string{"data/primary", "data/metrics"},
	}
}

// CheckMetricConsistencyTool returns a tool that checks whether a metric reports the same value across data sources.
func CheckMetricConsistencyTool() Tool {
	return Tool{
		Def: toolDef("check_metric_consistency",
			"Checks whether a metric reports consistent values across multiple data sources and flags discrepancies.",
			`{
				"type": "object",
				"properties": {
					"metric_name": {"type": "string", "description": "Canonical name of the metric to check."},
					"sources":     {"type": "array", "items": {"type": "string"}, "description": "Data source names to compare, e.g. [\"warehouse\", \"crm\", \"billing\"]."}
				},
				"required": ["metric_name", "sources"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				MetricName string   `json:"metric_name"`
				Sources    []string `json:"sources"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockConsistencyCheck(p.MetricName, p.Sources))
		},
		Skill: "- Use check_metric_consistency to detect when the same metric shows different values across warehouse, CRM, or billing sources.",
		Tags:  []string{"data/primary", "data/metrics"},
	}
}

// GetDataQualityReportTool returns a tool that reports quality health for a given dataset and table.
func GetDataQualityReportTool() Tool {
	return Tool{
		Def: toolDef("get_data_quality_report",
			"Returns a data quality report (null rates, duplicate rates, anomalies) for a specific dataset and table.",
			`{
				"type": "object",
				"properties": {
					"dataset": {"type": "string", "description": "Dataset name, e.g. \"analytics\"."},
					"table":   {"type": "string", "description": "Table name within the dataset, e.g. \"events\"."}
				},
				"required": ["dataset", "table"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Dataset string `json:"dataset"`
				Table   string `json:"table"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockDataQuality(p.Dataset, p.Table))
		},
		Skill: "- Use get_data_quality_report to surface null rates, duplicate rows, and anomalies in a warehouse table before trusting it for analysis.",
		Tags:  []string{"data/primary", "data/metrics"},
	}
}

// mockMetricDefinition returns a stub canonical metric.
// TODO: query the canonical metric registry from storage once wired.
func mockMetricDefinition(name string) data.MetricDefinition {
	return data.MetricDefinition{
		Name:        name,
		Owner:       "data@company.com",
		Description: "Monthly count of unique users who performed at least one action.",
		Unit:        "users",
		Category:    "engagement",
	}
}

// mockConsistencyCheck returns a stub consistency result.
// TODO: query each source via their respective adapters and compare.
func mockConsistencyCheck(metricName string, sources []string) data.MetricConsistencyResult {
	sourceValues := make(map[string]string, len(sources))
	for _, s := range sources {
		sourceValues[s] = "12,450"
	}
	// Simulate a discrepancy on the "crm" source to produce a realistic mock.
	if _, ok := sourceValues["crm"]; ok {
		sourceValues["crm"] = "11,980"
	}
	consistent := true
	discrepancy := ""
	severity := "none"
	if sourceValues["crm"] != "" && len(sources) > 1 {
		consistent = false
		discrepancy = "crm reports 11,980 vs warehouse 12,450 — delta of 470 (3.8%)"
		severity = "low"
	}
	return data.MetricConsistencyResult{
		MetricName:  metricName,
		Sources:     sourceValues,
		Consistent:  consistent,
		Discrepancy: discrepancy,
		Severity:    severity,
	}
}

// mockDataQuality returns a stub quality report.
// TODO: run real profiling queries against the warehouse.
func mockDataQuality(dataset, table string) data.DataQualityReport {
	return data.DataQualityReport{
		Dataset:       dataset,
		Table:         table,
		NullRate:      0.03,
		DuplicateRate: 0.001,
		Anomalies:     []string{"spike in event_count on 2026-04-15 (+340% vs 7-day avg)"},
		Health:        "degraded",
	}
}
