package tool

import (
	"context"
	"encoding/json"

	"github.com/loke/symmate/internal/marketing"
)

var stubCampaigns = []marketing.Campaign{
	{
		ID:        "camp-001",
		Name:      "Q2 Newsletter Series",
		Channel:   "email",
		Status:    marketing.StatusActive,
		Budget:    15000,
		Spend:     9200,
		Leads:     180,
		MQLs:      42,
		StartDate: "2026-04-01",
		EndDate:   "2026-06-30",
	},
	{
		ID:        "camp-002",
		Name:      "Q2 LinkedIn Outreach",
		Channel:   "linkedin",
		Status:    marketing.StatusActive,
		Budget:    25000,
		Spend:     18700,
		Leads:     110,
		MQLs:      33,
		StartDate: "2026-04-15",
		EndDate:   "2026-06-30",
	},
	{
		ID:        "camp-003",
		Name:      "SaaStr Annual Sponsorship",
		Channel:   "events",
		Status:    marketing.StatusCompleted,
		Budget:    40000,
		Spend:     40000,
		Leads:     50,
		MQLs:      12,
		StartDate: "2026-03-10",
		EndDate:   "2026-03-12",
	},
}

var stubContentItems = []marketing.ContentItem{
	{
		ID:          "content-001",
		Title:       "How AI Transforms RevOps",
		Type:        "blog",
		Status:      "published",
		PublishDate: "2026-04-10",
		Views:       1420,
	},
	{
		ID:          "content-002",
		Title:       "Acme Corp Case Study",
		Type:        "case_study",
		Status:      "published",
		PublishDate: "2026-04-22",
		Views:       830,
	},
	{
		ID:          "content-003",
		Title:       "Q3 Product Webinar",
		Type:        "webinar",
		Status:      "draft",
		PublishDate: "2026-07-08",
		Views:       0,
	},
	{
		ID:          "content-004",
		Title:       "AI OS Buyer's Guide",
		Type:        "whitepaper",
		Status:      "draft",
		PublishDate: "2026-06-15",
		Views:       0,
	},
}

// GetCampaignPerformanceTool returns a tool that summarises campaign attribution for a period.
func GetCampaignPerformanceTool() Tool {
	return Tool{
		Def: toolDef("get_campaign_performance",
			"Returns a marketing attribution report with campaign metrics for a given period.",
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
			report := marketing.AttributionReport{
				Period:             p.Period,
				Campaigns:          stubCampaigns,
				TotalLeads:         340,
				TotalMQLs:          87,
				PipelineInfluenced: 1_200_000,
			}
			return jsonString(report)
		},
		Skill: "- Use get_campaign_performance to answer questions about campaign ROI, lead volume, or marketing pipeline influence.",
		Tags:  []string{"marketing/campaign", "marketing/attribution"},
	}
}

// ListContentCalendarTool returns a tool that lists content items filtered by status.
func ListContentCalendarTool() Tool {
	return Tool{
		Def: toolDef("list_content_calendar",
			"Returns the marketing content calendar, optionally filtered by status.",
			`{
				"type": "object",
				"properties": {
					"status": {"type": "string", "description": "Filter by status: draft, published, or all (default all)."}
				},
				"required": []
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Status string `json:"status"`
			}
			// ignore unmarshal error — status is optional
			_ = json.Unmarshal(args, &p)

			var items []marketing.ContentItem
			for _, item := range stubContentItems {
				if p.Status == "" || p.Status == "all" || item.Status == p.Status {
					items = append(items, item)
				}
			}
			if items == nil {
				items = []marketing.ContentItem{}
			}
			return jsonString(items)
		},
		Skill: "- Use list_content_calendar to answer questions about upcoming content, blog posts, or the editorial calendar.",
		Tags:  []string{"marketing/content"},
	}
}

// GetLeadAttributionTool returns a tool that reports which campaign sourced a given lead.
func GetLeadAttributionTool() Tool {
	return Tool{
		Def: toolDef("get_lead_attribution",
			"Returns the source campaign and channel that originally attributed a lead by email address.",
			`{
				"type": "object",
				"properties": {
					"lead_email": {"type": "string", "description": "Email address of the lead to look up."}
				},
				"required": ["lead_email"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				LeadEmail string `json:"lead_email"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			result := map[string]string{
				"lead_email":       p.LeadEmail,
				"source_campaign":  "Q2 LinkedIn Outreach",
				"channel":          "linkedin",
				"attributed_at":    "2026-04-18T14:32:00Z",
			}
			return jsonString(result)
		},
		Skill: "- Use get_lead_attribution to identify which campaign or channel originally sourced a specific lead by email.",
		Tags:  []string{"marketing/attribution", "sales/pipeline"},
	}
}
