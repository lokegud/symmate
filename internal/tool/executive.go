package tool

import (
	"context"
	"encoding/json"

	"github.com/loke/symmate/internal/executive"
)

// GetCompanyHealthTool returns a tool that synthesises a cross-functional health digest.
func GetCompanyHealthTool() Tool {
	return Tool{
		Def: toolDef("get_company_health",
			"Returns a cross-functional company health digest for a given period, scoring each department and surfacing top risks and momentum.",
			`{
				"type": "object",
				"properties": {
					"period": {"type": "string", "description": "ISO date range, e.g. \"2026-04-01/2026-04-30\"."}
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
			return jsonString(mockCompanyHealth(p.Period))
		},
		Skill: "- Use get_company_health to produce an executive-level company health digest spanning all departments for a reporting period.",
		Tags:  []string{"executive/primary", "executive/metrics"},
	}
}

// GetOKRGapsTool returns a tool that reports OKR progress and gaps for a department.
func GetOKRGapsTool() Tool {
	return Tool{
		Def: toolDef("get_okr_gaps",
			"Returns OKR progress and gap analysis for a department in a given quarter.",
			`{
				"type": "object",
				"properties": {
					"department": {"type": "string", "description": "Department name, e.g. \"engineering\"."},
					"quarter":    {"type": "string", "description": "Quarter identifier, e.g. \"Q2-2026\"."}
				},
				"required": ["department", "quarter"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Department string `json:"department"`
				Quarter    string `json:"quarter"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockOKRGaps(p.Department, p.Quarter))
		},
		Skill: "- Use get_okr_gaps to identify which OKRs are off-track or at-risk for a department so the executive can intervene early.",
		Tags:  []string{"executive/primary", "executive/metrics"},
	}
}

// GetCrossFunctionalSignalsTool returns a tool that surfaces signals spanning multiple departments.
func GetCrossFunctionalSignalsTool() Tool {
	return Tool{
		Def: toolDef("get_cross_functional_signals",
			"Surfaces notable signals that span multiple departments, such as a sales pipeline drop coinciding with an engineering freeze.",
			`{
				"type": "object",
				"properties": {
					"period": {"type": "string", "description": "ISO date range to scan for signals."}
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
			return jsonString(mockCrossFunctionalSignals(p.Period))
		},
		Skill: "- Use get_cross_functional_signals to detect emerging issues that cut across departments before they compound.",
		Tags:  []string{"executive/primary"},
	}
}

// mockCompanyHealth returns stub company health data.
// TODO: aggregate real scores from each department's metric registry.
func mockCompanyHealth(period string) executive.CompanyHealthDigest {
	_ = period
	return executive.CompanyHealthDigest{
		Period: period,
		DepartmentScores: map[string]int{
			"engineering":     72,
			"sales":           68,
			"customer_success": 81,
			"finance":         90,
			"hr":              75,
		},
		TopRisks: []string{
			"Sales pipeline coverage below 3× for Q3.",
			"Engineering cycle time up 18% week-over-week.",
		},
		TopMomentum: []string{
			"Customer NPS +12 points after recent onboarding revamp.",
			"Finance runway extended to 22 months post close.",
		},
		OverallScore: 77,
	}
}

// mockOKRGaps returns stub OKR gap data.
// TODO: pull from Linear milestones and metric registry once ingestion is live.
func mockOKRGaps(department, quarter string) executive.OKRGapReport {
	_ = quarter
	return executive.OKRGapReport{
		Department: department,
		Quarter:    quarter,
		Objectives: []executive.OKRItem{
			{
				Objective:  "Ship v2 API to 100% of customers",
				Progress:   0.61,
				Status:     "at_risk",
				GapSummary: "Migration tooling delayed by 2 sprints; 39% of customers still on v1.",
			},
			{
				Objective:  "Reduce P95 latency below 200ms",
				Progress:   0.85,
				Status:     "on_track",
				GapSummary: "Current P95 is 214ms; one caching PR in review should close the gap.",
			},
		},
		AtRiskCount:  1,
		OnTrackCount: 1,
	}
}

// mockCrossFunctionalSignals returns stub cross-functional signals.
// TODO: join real events from entity graph and typed event bus.
func mockCrossFunctionalSignals(period string) []executive.CrossFunctionalSignal {
	_ = period
	return []executive.CrossFunctionalSignal{
		{
			Signal:      "Sales close rate declined 14% while engineering entered feature freeze.",
			Departments: []string{"sales", "engineering"},
			Severity:    "warning",
			Recommended: "Coordinate freeze windows with sales cycle; prioritise customer-facing bug fixes during freeze.",
		},
		{
			Signal:      "Support ticket volume up 30% correlating with latest deploy.",
			Departments: []string{"customer_success", "engineering"},
			Severity:    "critical",
			Recommended: "Engineering to triage top-10 ticket drivers; customer success to prepare proactive outreach.",
		},
	}
}
