package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/loke/symmate/internal/finance"
)

// GetCloseChecklistTool returns a tool that retrieves the monthly close checklist status.
func GetCloseChecklistTool() Tool {
	return Tool{
		Def: toolDef("get_close_checklist",
			"Returns the current monthly financial close checklist with item statuses and owners.",
			`{
				"type": "object",
				"properties": {
					"period": {"type": "string", "description": "Accounting period in YYYY-MM format, e.g. \"2026-04\"."}
				},
				"required": ["period"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Period string `json:"period"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockCloseChecklist(p.Period))
		},
		Skill: "- Use get_close_checklist to review close progress, find blocked items, or report status for a given accounting period.",
		Tags:  []string{"finance/primary"},
	}
}

// DetectSpendAnomalyTool returns a tool that surfaces spend deviations from expected baselines.
func DetectSpendAnomalyTool() Tool {
	return Tool{
		Def: toolDef("detect_spend_anomaly",
			"Detects spend anomalies by comparing actual vendor/category spend against the expected baseline for a period.",
			`{
				"type": "object",
				"properties": {
					"period":    {"type": "string", "description": "Accounting period in YYYY-MM format."},
					"threshold": {"type": "number", "description": "Minimum deviation percentage to flag, e.g. 15.0 for 15%."}
				},
				"required": ["period"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Period    string  `json:"period"`
				Threshold float64 `json:"threshold"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			if p.Threshold == 0 {
				p.Threshold = 20.0
			}
			return jsonString(mockSpendAnomalies(p.Period, p.Threshold))
		},
		Skill: "- Use detect_spend_anomaly to find unexpected cost spikes or drops across vendors and categories for a given period.",
		Tags:  []string{"finance/primary", "finance/metrics"},
	}
}

// GetForecastDeltaTool returns a tool that reports forecast-vs-actual variance for a metric.
func GetForecastDeltaTool() Tool {
	return Tool{
		Def: toolDef("get_forecast_delta",
			"Returns the forecast-vs-actual variance for a financial metric over a given period.",
			`{
				"type": "object",
				"properties": {
					"period": {"type": "string", "description": "Accounting period in YYYY-MM format."},
					"metric": {"type": "string", "description": "Financial metric to report: \"revenue\", \"opex\", or \"gross_margin\"."}
				},
				"required": ["period", "metric"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Period string `json:"period"`
				Metric string `json:"metric"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockForecastDelta(p.Period, p.Metric))
		},
		Skill: "- Use get_forecast_delta to explain why actuals diverged from forecast for revenue, opex, or gross margin in any period.",
		Tags:  []string{"finance/metrics"},
	}
}

// mockCloseChecklist returns stub close checklist data.
// TODO: pull from accounting system (QuickBooks, NetSuite) when APIs are wired.
func mockCloseChecklist(period string) []finance.CloseChecklistItem {
	due := period + "-10"
	return []finance.CloseChecklistItem{
		{ID: "close-01", Name: "Bank reconciliation", Status: "complete", Owner: "finance@", DueDate: due},
		{ID: "close-02", Name: "AP invoice sweep", Status: "in_progress", Owner: "ap@", DueDate: due},
		{ID: "close-03", Name: "Revenue recognition review", Status: "open", Owner: "controller@", DueDate: due},
		{ID: "close-04", Name: "Payroll accrual posting", Status: "complete", Owner: "hr@", DueDate: due},
		{ID: "close-05", Name: "Intercompany eliminations", Status: "blocked", Owner: "controller@", DueDate: due,
			Notes: "Waiting on subsidiary trial balance from EMEA entity."},
	}
}

// mockSpendAnomalies returns stub anomaly data filtered by threshold.
// TODO: query ledger transactions from storage once accounting events are ingested.
func mockSpendAnomalies(period string, thresholdPct float64) []finance.SpendAnomaly {
	candidates := []finance.SpendAnomaly{
		{
			Category: "Cloud Infrastructure", Vendor: "AWS",
			ExpectedUSD: 42000, ActualUSD: 61800, DeviationPct: 47.1,
			Severity:  "high",
			Reasoning: fmt.Sprintf("AWS spend in %s exceeded baseline by 47%%. Spike correlates with data migration job kicked off on the 8th.", period),
		},
		{
			Category: "SaaS Tools", Vendor: "Figma",
			ExpectedUSD: 3200, ActualUSD: 3840, DeviationPct: 20.0,
			Severity:  "medium",
			Reasoning: "Figma seats increased from 16 to 20. Matches headcount plan but not reflected in forecast.",
		},
		{
			Category: "Travel & Expense", Vendor: "Expensify",
			ExpectedUSD: 8500, ActualUSD: 8670, DeviationPct: 2.0,
			Severity:  "low",
			Reasoning: "Within normal range.",
		},
	}
	var out []finance.SpendAnomaly
	for _, a := range candidates {
		if a.DeviationPct >= thresholdPct {
			out = append(out, a)
		}
	}
	return out
}

// mockForecastDelta returns stub forecast-vs-actual variance.
// TODO: replace with real plan vs. ledger pull from storage.
func mockForecastDelta(period, metric string) finance.ForecastDelta {
	stubs := map[string]finance.ForecastDelta{
		"revenue": {
			Period: period, Metric: "revenue",
			ForecastUSD: 1_250_000, ActualUSD: 1_198_000,
			DeltaUSD: -52_000, DeltaPct: -4.16,
			Commentary: "Revenue missed by $52k. Two enterprise deals slipped to the following month; expansion MRR tracked on plan.",
		},
		"opex": {
			Period: period, Metric: "opex",
			ForecastUSD: 920_000, ActualUSD: 974_000,
			DeltaUSD: 54_000, DeltaPct: 5.87,
			Commentary: "OpEx over by $54k, driven by unplanned AWS spend and one-time legal fee for IP filing.",
		},
		"gross_margin": {
			Period: period, Metric: "gross_margin",
			ForecastUSD: 875_000, ActualUSD: 838_600,
			DeltaUSD: -36_400, DeltaPct: -4.16,
			Commentary: "Gross margin compressed 4pp vs. plan, primarily from cloud cost overrun reducing contribution margin.",
		},
	}
	if d, ok := stubs[metric]; ok {
		return d
	}
	return finance.ForecastDelta{
		Period: period, Metric: metric,
		Commentary: fmt.Sprintf("No forecast data available for metric %q in %s.", metric, period),
	}
}
