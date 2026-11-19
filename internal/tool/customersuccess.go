package tool

import (
	"context"
	"encoding/json"

	"github.com/loke/symmate/internal/customersuccess"
)

// GetCustomerHealthTool returns a tool that reports health tier, NPS, and renewal info for an account.
func GetCustomerHealthTool() Tool {
	return Tool{
		Def: toolDef("get_customer_health",
			"Returns health tier, NPS, ARR, renewal date, and CSM assignment for a named customer account.",
			`{
				"type": "object",
				"properties": {
					"account_name": {"type": "string", "description": "The customer account to look up."}
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
			record := customersuccess.CustomerRecord{
				AccountName:         p.AccountName,
				CSMID:               "csm-007",
				HealthTier:          customersuccess.HealthGreen,
				NPS:                 52,
				ARR:                 120_000,
				ContractRenewalDate: "2027-01-15",
				OpenTickets:         2,
				LastQBRDate:         "2026-03-10",
			}
			return jsonString(record)
		},
		Skill: "- Use get_customer_health to answer questions about a customer's health score, NPS, renewal date, or CSM assignment.",
		Tags:  []string{"cs/health", "sales/account"},
	}
}

// GetOnboardingStatusTool returns a tool that reports onboarding progress for a customer account.
func GetOnboardingStatusTool() Tool {
	return Tool{
		Def: toolDef("get_onboarding_status",
			"Returns the current onboarding stage and completion percentage for a named customer account.",
			`{
				"type": "object",
				"properties": {
					"account_name": {"type": "string", "description": "The customer account to look up."}
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
			status := customersuccess.OnboardingStatus{
				AccountName:     p.AccountName,
				Stage:           "implementation",
				PercentComplete: 60,
				BlockedReason:   "",
			}
			return jsonString(status)
		},
		Skill: "- Use get_onboarding_status to answer questions about a customer's onboarding progress or go-live timeline.",
		Tags:  []string{"cs/onboarding"},
	}
}

// GetChurnAlertsTool returns a tool that lists at-risk accounts above a minimum risk score.
func GetChurnAlertsTool() Tool {
	return Tool{
		Def: toolDef("get_churn_alerts",
			"Returns a list of at-risk customer accounts with churn risk scores above a minimum threshold.",
			`{
				"type": "object",
				"properties": {
					"min_risk_score": {"type": "integer", "description": "Minimum risk score (0-100). Defaults to 50."}
				},
				"required": []
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				MinRiskScore *int `json:"min_risk_score"`
			}
			// ignore unmarshal error — min_risk_score is optional
			_ = json.Unmarshal(args, &p)

			minScore := 50
			if p.MinRiskScore != nil {
				minScore = *p.MinRiskScore
			}

			allAlerts := []customersuccess.ChurnAlert{
				{
					AccountName: "Umbrella Corp",
					RiskScore:   78,
					Signals: []string{
						"no login in 30 days",
						"support ticket spike",
						"champion left company",
					},
					RecommendedAction: "Schedule executive sponsor call within 5 business days.",
				},
				{
					AccountName: "Soylent Systems",
					RiskScore:   62,
					Signals: []string{
						"feature adoption below 40%",
						"renewal 90 days out",
					},
					RecommendedAction: "Assign CSM for an enablement session focused on unused features.",
				},
			}

			var filtered []customersuccess.ChurnAlert
			for _, alert := range allAlerts {
				if alert.RiskScore >= minScore {
					filtered = append(filtered, alert)
				}
			}
			if filtered == nil {
				filtered = []customersuccess.ChurnAlert{}
			}
			return jsonString(filtered)
		},
		Skill: "- Use get_churn_alerts to identify at-risk customers, review churn signals, or find accounts needing intervention.",
		Tags:  []string{"cs/churn", "cs/health"},
	}
}
