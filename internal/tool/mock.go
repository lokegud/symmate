package tool

import (
	"context"
	"encoding/json"
)

// MockDepartmentTools returns the department tools that still run on hardcoded
// sample data (everything except the GitHub-backed engineering tools).
//
// They are NOT registered by default. Set SYMMATE_ENABLE_MOCK_TOOLS=1 to register
// them for demos; every result is then wrapped as demo data so the model can say so
// instead of presenting invented numbers as real.
func MockDepartmentTools() []Tool {
	tools := []Tool{
		// customer success
		GetCustomerHealthTool(), GetOnboardingStatusTool(), GetChurnAlertsTool(),
		// data
		LookupMetricTool(), CheckMetricConsistencyTool(), GetDataQualityReportTool(),
		// entity graph
		GetEntityProfileTool(), ListCompaniesTool(),
		// executive
		GetCompanyHealthTool(), GetOKRGapsTool(), GetCrossFunctionalSignalsTool(),
		// finance
		GetCloseChecklistTool(), DetectSpendAnomalyTool(), GetForecastDeltaTool(),
		// hr
		GetEmployeeOnboardingStatusTool(), FlagEngagementRiskTool(), GetHeadcountGapsTool(),
		// legal
		ReviewContractTool(), GetClauseDeviationTool(), GetComplianceGapsTool(),
		// marketing
		GetCampaignPerformanceTool(), ListContentCalendarTool(), GetLeadAttributionTool(),
		// operations
		GetVendorSLATool(), GetDependencyBlockersTool(), GetOpsHealthTool(),
		// product
		SynthesizeFeedbackTool(), DetectRoadmapGapsTool(), GetNPSSummaryTool(),
		// sales
		GetPipelineForecastTool(), GetDealStatusTool(), GetAccountHealthTool(),
		// security
		TriageAlertTool(), GetCompliancePostureTool(), CorrelateCVETool(),
	}
	for i := range tools {
		tools[i] = markDemo(tools[i])
	}
	return tools
}

// markDemo wraps a tool so its output is labelled as sample data.
func markDemo(t Tool) Tool {
	inner := t.Execute
	t.Execute = func(ctx context.Context, args json.RawMessage) string {
		out := inner(ctx, args)
		var raw json.RawMessage
		if json.Unmarshal([]byte(out), &raw) != nil {
			raw, _ = json.Marshal(out) // plain-text result (e.g. an error string)
		}
		return jsonString(map[string]any{
			"demo_data": true,
			"note":      "Sample data — this integration is not connected yet. Tell the user the figures are not real.",
			"result":    raw,
		})
	}
	t.Def.Description = "[DEMO DATA] " + t.Def.Description
	return t
}
