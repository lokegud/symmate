package tool

import (
	"context"
	"encoding/json"

	"github.com/loke/symmate/internal/operations"
)

// GetVendorSLATool returns a tool that reports SLA health for a named vendor.
func GetVendorSLATool() Tool {
	return Tool{
		Def: toolDef("get_vendor_sla",
			"Returns current SLA health and breach history for a vendor.",
			`{
				"type": "object",
				"properties": {
					"vendor_name": {"type": "string", "description": "Name of the vendor, e.g. \"Stripe\"."},
					"date_range":  {"type": "string", "description": "ISO date range like \"2026-04-01/2026-04-30\"."}
				},
				"required": ["vendor_name", "date_range"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				VendorName string `json:"vendor_name"`
				DateRange  string `json:"date_range"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockVendorSLA(p.VendorName, p.DateRange))
		},
		Skill: "- Use get_vendor_sla to check whether a vendor is meeting their contractual SLA and how many breaches have occurred in a period.",
		Tags:  []string{"operations/primary", "operations/metrics"},
	}
}

// GetDependencyBlockersTool returns a tool that surfaces active cross-team blockers.
func GetDependencyBlockersTool() Tool {
	return Tool{
		Def: toolDef("get_dependency_blockers",
			"Lists active cross-team dependency blockers filtered by team or severity.",
			`{
				"type": "object",
				"properties": {
					"team":     {"type": "string", "description": "Filter by owner or blocked team name. Omit for all teams."},
					"severity": {"type": "string", "description": "Minimum severity: low | medium | high | critical."}
				},
				"required": []
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Team     string `json:"team"`
				Severity string `json:"severity"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockDependencyBlockers(p.Team, p.Severity))
		},
		Skill: "- Use get_dependency_blockers to find which cross-team dependencies are stalled and who owns resolution.",
		Tags:  []string{"operations/primary"},
	}
}

// GetOpsHealthTool returns a tool that summarises overall operations health for a period.
func GetOpsHealthTool() Tool {
	return Tool{
		Def: toolDef("get_ops_health",
			"Returns an aggregate operations health summary: active blockers, at-risk vendors, and overall status.",
			`{
				"type": "object",
				"properties": {
					"period": {"type": "string", "description": "ISO date range like \"2026-04-01/2026-04-30\"."}
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
			return jsonString(mockOpsHealth(p.Period))
		},
		Skill: "- Use get_ops_health for a top-level operations status check covering vendor SLAs, cross-team blockers, and overall risk level.",
		Tags:  []string{"operations/primary", "operations/metrics"},
	}
}

// mockVendorSLA returns stub SLA data.
// TODO: query real vendor status pages / contract DB when APIs are wired.
func mockVendorSLA(vendor, dateRange string) operations.VendorSLAStatus {
	_ = dateRange
	return operations.VendorSLAStatus{
		VendorName:     vendor,
		SLATarget:      99.9,
		CurrentUptime:  99.72,
		BreachCount:    2,
		Status:         "at_risk",
		NextReviewDate: "2026-06-01",
	}
}

// mockDependencyBlockers returns stub blocker list.
// TODO: ingest from Linear / Jira dependency fields when real APIs are wired.
func mockDependencyBlockers(team, severity string) []operations.DependencyBlocker {
	_, _ = team, severity
	return []operations.DependencyBlocker{
		{
			BlockerID:   "BLK-001",
			Title:       "Auth service rate-limit config not delivered to platform team",
			OwnerTeam:   "auth",
			BlockedTeam: "platform",
			Severity:    "high",
			DaysBlocked: 6,
			Resolution:  "",
		},
		{
			BlockerID:   "BLK-002",
			Title:       "Data warehouse schema migration pending sign-off from analytics",
			OwnerTeam:   "analytics",
			BlockedTeam: "data",
			Severity:    "medium",
			DaysBlocked: 3,
			Resolution:  "",
		},
	}
}

// mockOpsHealth returns a stub health summary.
// TODO: aggregate from live SLA + blocker data once storage is wired.
func mockOpsHealth(period string) operations.OperationsHealthSummary {
	blockers := mockDependencyBlockers("", "")
	return operations.OperationsHealthSummary{
		Period:         period,
		ActiveBlockers: blockers,
		VendorsAtRisk:  []string{"Stripe", "Twilio"},
		OverallStatus:  "degraded",
		TopRisk:        "Auth rate-limit blocker has been open 6 days and is delaying the platform Q2 release.",
	}
}
