package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/loke/symmate/internal/entity"
)

var entityRegistry = entity.NewRegistry()

// GetEntityProfileTool returns a tool that builds a 360-degree cross-department
// profile for a named company.
func GetEntityProfileTool() Tool {
	return Tool{
		Def: toolDef("get_entity_profile",
			"Returns a 360-degree cross-department profile for a named company, including sales deal status, CS health tier, engineering risk signals, and marketing engagement.",
			`{
				"type": "object",
				"properties": {
					"company_name": {"type": "string", "description": "The name of the company to profile."}
				},
				"required": ["company_name"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				CompanyName string `json:"company_name"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			profile, err := entityRegistry.Profile(p.CompanyName)
			if err != nil {
				return fmt.Sprintf(`{"error":"company not found: %s"}`, p.CompanyName)
			}
			return jsonString(profile)
		},
		Skill: "- Use get_entity_profile when asked for a 360-degree view of a company across departments — it returns sales deal status, CS health tier, engineering risk signals, and marketing engagement in one call.",
		Tags:  []string{"entity/company", "sales/account", "cs/health", "engineering/risk"},
	}
}

// ListCompaniesTool returns a tool that lists all companies in the entity registry.
func ListCompaniesTool() Tool {
	return Tool{
		Def: toolDef("list_companies",
			"Returns a list of all companies (customers/accounts) known to the system.",
			`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		Execute: func(_ context.Context, _ json.RawMessage) string {
			return jsonString(entityRegistry.AllCompanies())
		},
		Skill: "- Use list_companies when asked to list customers, accounts, or companies in the system.",
		Tags:  []string{"entity/company"},
	}
}
