package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/loke/symmate/internal/legal"
)

// ReviewContractTool returns a tool that analyzes a contract for clause deviations and risk.
func ReviewContractTool() Tool {
	return Tool{
		Def: toolDef("review_contract",
			"Analyzes a contract for clause deviations from the standard playbook and assigns an overall risk level.",
			`{
				"type": "object",
				"properties": {
					"contract_id":   {"type": "string", "description": "Unique identifier for the contract document."},
					"contract_type": {"type": "string", "description": "Type of contract: nda | msa | sow | saas | other."},
					"clauses":       {"type": "array", "items": {"type": "string"}, "description": "List of clause names to review, e.g. [\"limitation_of_liability\", \"ip_ownership\"]."}
				},
				"required": ["contract_id", "contract_type", "clauses"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				ContractID   string   `json:"contract_id"`
				ContractType string   `json:"contract_type"`
				Clauses      []string `json:"clauses"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockContractReview(p.ContractID, p.ContractType, p.Clauses))
		},
		Skill: "- Use review_contract to flag clause deviations and risk level for any inbound or outbound contract.",
		Tags:  []string{"legal/primary"},
	}
}

// GetClauseDeviationTool returns a tool that surfaces the specific deviation detail for a named clause.
func GetClauseDeviationTool() Tool {
	return Tool{
		Def: toolDef("get_clause_deviation",
			"Returns the standard playbook text vs actual text and severity for a specific contract clause.",
			`{
				"type": "object",
				"properties": {
					"contract_id":  {"type": "string", "description": "Contract to inspect."},
					"clause_name":  {"type": "string", "description": "Clause to retrieve, e.g. \"limitation_of_liability\"."}
				},
				"required": ["contract_id", "clause_name"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				ContractID string `json:"contract_id"`
				ClauseName string `json:"clause_name"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockClauseDeviation(p.ContractID, p.ClauseName))
		},
		Skill: "- Use get_clause_deviation to retrieve exact standard vs actual text and remediation guidance for a specific clause.",
		Tags:  []string{"legal/primary"},
	}
}

// GetComplianceGapsTool returns a tool that reports compliance gaps for an entity against a framework.
func GetComplianceGapsTool() Tool {
	return Tool{
		Def: toolDef("get_compliance_gaps",
			"Reports compliance gaps for an entity against a regulatory framework (gdpr, soc2, hipaa, ccpa, iso27001).",
			`{
				"type": "object",
				"properties": {
					"entity_id":  {"type": "string", "description": "The company or contract entity to assess."},
					"framework":  {"type": "string", "description": "Compliance framework: gdpr | soc2 | hipaa | ccpa | iso27001."}
				},
				"required": ["entity_id", "framework"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				EntityID  string `json:"entity_id"`
				Framework string `json:"framework"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockComplianceGaps(p.EntityID, p.Framework))
		},
		Skill: "- Use get_compliance_gaps to identify regulatory compliance gaps for a customer or vendor against a specific framework.",
		Tags:  []string{"legal/metrics"},
	}
}

// mockContractReview is the Phase 1 stub.
// TODO: parse real contract text via document extraction and compare against clause playbook in storage.
func mockContractReview(contractID, contractType string, clauses []string) legal.ContractReview {
	deviated := []string{}
	riskLevel := "low"
	for _, c := range clauses {
		switch c {
		case "limitation_of_liability", "ip_ownership", "indemnification":
			deviated = append(deviated, c)
			riskLevel = "high"
		case "payment_terms", "termination":
			deviated = append(deviated, c)
			if riskLevel == "low" {
				riskLevel = "medium"
			}
		}
	}
	summary := fmt.Sprintf("%s contract %s reviewed. %d clause(s) deviate from playbook.", contractType, contractID, len(deviated))
	return legal.ContractReview{
		ContractID:      contractID,
		ContractType:    contractType,
		DeviatedClauses: deviated,
		RiskLevel:       riskLevel,
		Summary:         summary,
	}
}

// mockClauseDeviation is the Phase 1 stub.
// TODO: fetch actual clause text from ingested contract document store.
func mockClauseDeviation(contractID, clauseName string) legal.ClauseDeviation {
	_ = contractID // will be used when real storage is wired
	return legal.ClauseDeviation{
		ClauseName:   clauseName,
		StandardText: fmt.Sprintf("Standard playbook text for %s (placeholder).", clauseName),
		ActualText:   fmt.Sprintf("Counterparty text for %s deviates by removing mutual obligations (placeholder).", clauseName),
		Severity:     "high",
		Guidance:     "Request reinstatement of mutual obligations or escalate to General Counsel before signing.",
	}
}

// mockComplianceGaps is the Phase 1 stub.
// TODO: query entity compliance state from the entity graph + policy registry.
func mockComplianceGaps(entityID, framework string) legal.ComplianceGap {
	_ = entityID // will be used when real entity graph is wired
	gaps := []string{
		fmt.Sprintf("%s: data processing agreement not executed", framework),
		fmt.Sprintf("%s: breach notification SLA not contractually defined", framework),
	}
	return legal.ComplianceGap{
		EntityID:    entityID,
		Framework:   framework,
		Gaps:        gaps,
		GapCount:    len(gaps),
		Remediation: "Execute a data processing addendum and insert a 72-hour breach notification clause.",
	}
}
