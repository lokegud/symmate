package tool

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/loke/symmate/internal/sales"
)

var stubDeals = []sales.Deal{
	{
		ID:          "deal-001",
		AccountName: "Acme Corp",
		OwnerEmail:  "alice@symmate.io",
		Stage:       sales.StageNegotiation,
		ARR:         180000,
		CloseDate:   "2026-06-30",
		Notes:       "Legal review in progress.",
	},
	{
		ID:          "deal-002",
		AccountName: "Globex Industries",
		OwnerEmail:  "bob@symmate.io",
		Stage:       sales.StageProposal,
		ARR:         240000,
		CloseDate:   "2026-07-15",
		Notes:       "Champion confirmed; awaiting procurement.",
	},
	{
		ID:          "deal-003",
		AccountName: "Initech Solutions",
		OwnerEmail:  "carol@symmate.io",
		Stage:       sales.StageQualified,
		ARR:         95000,
		CloseDate:   "2026-08-01",
		Notes:       "Technical evaluation underway.",
	},
}

// GetPipelineForecastTool returns a tool that summarises ARR forecast for a quarter.
func GetPipelineForecastTool() Tool {
	return Tool{
		Def: toolDef("get_pipeline_forecast",
			"Returns an ARR forecast summary for the sales pipeline for a given reporting period.",
			`{
				"type": "object",
				"properties": {
					"period": {"type": "string", "description": "The reporting period, e.g. \"Q2-2026\"."}
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
			summary := sales.ForecastSummary{
				Period:       p.Period,
				TotalARR:     2_400_000,
				ClosedWonARR: 840_000,
				PipelineARR:  1_560_000,
				WinRate:      0.35,
				TopDeals:     stubDeals,
			}
			return jsonString(summary)
		},
		Skill: "- Use get_pipeline_forecast to answer questions about revenue pipeline, ARR forecast, or close rates for a quarter.",
		Tags:  []string{"sales/forecast", "sales/pipeline"},
	}
}

// GetDealStatusTool returns a tool that looks up the current stage and ARR for a named account.
func GetDealStatusTool() Tool {
	return Tool{
		Def: toolDef("get_deal_status",
			"Looks up the current deal stage and ARR for a named account.",
			`{
				"type": "object",
				"properties": {
					"account_name": {"type": "string", "description": "The name of the account to look up."}
				},
				"required": ["account_name"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				AccountName string `json:"account_name"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			lower := strings.ToLower(p.AccountName)
			for _, d := range stubDeals {
				if strings.ToLower(d.AccountName) == lower {
					return jsonString(d)
				}
			}
			return `{"error":"deal not found for account: ` + p.AccountName + `"}`
		},
		Skill: "- Use get_deal_status to look up the current deal stage and ARR for a named account.",
		Tags:  []string{"sales/deal", "sales/pipeline"},
	}
}

// GetAccountHealthTool returns a tool that reports churn risk and engagement score for an account.
func GetAccountHealthTool() Tool {
	return Tool{
		Def: toolDef("get_account_health",
			"Returns churn risk classification and engagement health score for a named account.",
			`{
				"type": "object",
				"properties": {
					"account_name": {"type": "string", "description": "The name of the account to assess."}
				},
				"required": ["account_name"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				AccountName string `json:"account_name"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			health := sales.AccountHealth{
				AccountName:       p.AccountName,
				HealthScore:       72,
				ChurnRisk:         "medium",
				LastActivityDate:  "2026-04-28",
				OpenOpportunities: 2,
			}
			return jsonString(health)
		},
		Skill: "- Use get_account_health to retrieve churn risk classification and engagement score for a named account.",
		Tags:  []string{"sales/account", "cs/health"},
	}
}
