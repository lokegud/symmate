package tool

import (
	"context"
	"encoding/json"

	"github.com/loke/symmate/internal/hr"
)

// GetEmployeeOnboardingStatusTool returns a tool that reports an employee's onboarding progress.
func GetEmployeeOnboardingStatusTool() Tool {
	return Tool{
		Def: toolDef("get_employee_onboarding_status",
			"Returns the current onboarding status and pending steps for a new employee.",
			`{
				"type": "object",
				"properties": {
					"employee_id": {"type": "string", "description": "The employee's unique identifier."}
				},
				"required": ["employee_id"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				EmployeeID string `json:"employee_id"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockOnboardingStatus(p.EmployeeID))
		},
		Skill: "- Use get_onboarding_status to check whether a new hire is progressing through onboarding or is blocked on a step.",
		Tags:  []string{"hr/primary"},
	}
}

// FlagEngagementRiskTool returns a tool that assesses flight-risk signals for an employee.
func FlagEngagementRiskTool() Tool {
	return Tool{
		Def: toolDef("flag_engagement_risk",
			"Assesses engagement and flight-risk signals for a given employee based on activity patterns.",
			`{
				"type": "object",
				"properties": {
					"employee_id": {"type": "string", "description": "The employee's unique identifier."}
				},
				"required": ["employee_id"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				EmployeeID string `json:"employee_id"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockEngagementRisk(p.EmployeeID))
		},
		Skill: "- Use flag_engagement_risk to surface employees showing disengagement signals before they become attrition events.",
		Tags:  []string{"hr/primary"},
	}
}

// GetHeadcountGapsTool returns a tool that reports headcount gaps across departments.
func GetHeadcountGapsTool() Tool {
	return Tool{
		Def: toolDef("get_headcount_gaps",
			"Reports headcount gaps between current staffing and targets for a given department.",
			`{
				"type": "object",
				"properties": {
					"department": {"type": "string", "description": "Department name, e.g. \"engineering\" or \"sales\". Pass \"all\" for all departments."}
				},
				"required": ["department"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Department string `json:"department"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockHeadcountGaps(p.Department))
		},
		Skill: "- Use get_headcount_gaps to identify which departments are under-staffed relative to their hiring targets.",
		Tags:  []string{"hr/metrics"},
	}
}

// mockOnboardingStatus returns stub onboarding data.
// TODO: fetch from HRIS / onboarding tracker when real APIs are wired.
func mockOnboardingStatus(employeeID string) hr.OnboardingStatus {
	return hr.OnboardingStatus{
		EmployeeID:     employeeID,
		Name:           "Alex Rivera",
		StartDate:      "2026-05-01",
		CompletedSteps: []string{"offer-signed", "background-check", "equipment-shipped", "slack-provisioned"},
		PendingSteps:   []string{"benefits-enrollment", "manager-1on1", "team-intro"},
		OverallStatus:  "on-track",
	}
}

// mockEngagementRisk returns stub engagement risk data.
// TODO: derive from calendar density, PR cadence, and 1:1 gap analysis.
func mockEngagementRisk(employeeID string) hr.EngagementRisk {
	return hr.EngagementRisk{
		EmployeeID: employeeID,
		Name:       "Jordan Kim",
		RiskLevel:  "medium",
		Signals: []string{
			"no 1:1 logged in 3 weeks",
			"PR output down 40% vs prior quarter",
			"missed last two all-hands",
		},
		Reasoning: "Multiple leading indicators of disengagement detected. Recommend manager outreach within 5 business days.",
	}
}

// mockHeadcountGaps returns stub headcount gap data.
// TODO: query HRIS headcount vs approved headcount plan from finance.
func mockHeadcountGaps(department string) hr.HeadcountGap {
	_ = department // will filter by department in real impl
	return hr.HeadcountGap{
		Department:   department,
		CurrentCount: 18,
		TargetCount:  24,
		OpenRoles:    6,
		Urgency:      "high",
	}
}
