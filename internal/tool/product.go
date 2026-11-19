package tool

import (
	"context"
	"encoding/json"

	"github.com/loke/symmate/internal/product"
)

// SynthesizeFeedbackTool returns a tool that clusters raw feedback into themes.
func SynthesizeFeedbackTool() Tool {
	return Tool{
		Def: toolDef("synthesize_feedback",
			"Clusters raw customer feedback items into named themes with sentiment and NPS impact.",
			`{
				"type": "object",
				"properties": {
					"feedback_items": {"type": "array", "items": {"type": "string"}, "description": "List of raw customer feedback strings to synthesize."}
				},
				"required": ["feedback_items"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				FeedbackItems []string `json:"feedback_items"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockSynthesizeFeedback(p.FeedbackItems))
		},
		Skill: "- Use synthesize_feedback to group raw NPS comments or support tickets into actionable themes.",
		Tags:  []string{"product/primary"},
	}
}

// DetectRoadmapGapsTool returns a tool that surfaces customer needs absent from the roadmap.
func DetectRoadmapGapsTool() Tool {
	return Tool{
		Def: toolDef("detect_roadmap_gaps",
			"Identifies customer needs present in feedback clusters that have no matching roadmap item.",
			`{
				"type": "object",
				"properties": {
					"roadmap_items": {"type": "array", "items": {"type": "string"}, "description": "Current roadmap item titles or descriptions."},
					"feedback_themes": {"type": "array", "items": {"type": "string"}, "description": "Synthesized feedback themes to check against the roadmap."}
				},
				"required": ["roadmap_items", "feedback_themes"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				RoadmapItems   []string `json:"roadmap_items"`
				FeedbackThemes []string `json:"feedback_themes"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockDetectRoadmapGaps(p.RoadmapItems, p.FeedbackThemes))
		},
		Skill: "- Use detect_roadmap_gaps to find high-demand customer needs that are missing from the current roadmap.",
		Tags:  []string{"product/primary"},
	}
}

// GetNPSSummaryTool returns a tool that reports NPS scores for a segment over a date range.
func GetNPSSummaryTool() Tool {
	return Tool{
		Def: toolDef("get_nps_summary",
			"Returns NPS score breakdown (promoters, passives, detractors) and top drivers for a customer segment.",
			`{
				"type": "object",
				"properties": {
					"segment":    {"type": "string", "description": "Customer segment, e.g. \"enterprise\" or \"smb\"."},
					"date_range": {"type": "string", "description": "ISO date range like \"2026-04-01/2026-04-30\"."}
				},
				"required": ["segment", "date_range"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Segment   string `json:"segment"`
				DateRange string `json:"date_range"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockNPSSummary(p.Segment, p.DateRange))
		},
		Skill: "- Use get_nps_summary to answer questions about customer satisfaction scores, promoter/detractor ratios, or NPS trends for a segment.",
		Tags:  []string{"product/metrics"},
	}
}

// mockSynthesizeFeedback returns stub clusters.
// TODO: embed and cluster via LLM once feedback ingestion pipeline is live.
func mockSynthesizeFeedback(items []string) []product.FeedbackCluster {
	_ = items // will drive real clustering
	return []product.FeedbackCluster{
		{
			Theme:     "Slow report load times",
			Count:     42,
			Sentiment: "negative",
			Verbatims: []string{
				"Reports take forever to load with large datasets.",
				"Dashboard is unusable when we have >10k rows.",
			},
			NPSImpact: 4.2,
		},
		{
			Theme:     "Missing bulk export",
			Count:     28,
			Sentiment: "negative",
			Verbatims: []string{
				"I need to export everything to CSV, not row by row.",
				"Bulk export is a blocker for our finance team.",
			},
			NPSImpact: 3.1,
		},
		{
			Theme:     "Onboarding experience praised",
			Count:     19,
			Sentiment: "positive",
			Verbatims: []string{
				"Setup was surprisingly smooth.",
				"Best onboarding I've seen in a B2B tool.",
			},
			NPSImpact: -1.5, // removing it would hurt NPS
		},
	}
}

// mockDetectRoadmapGaps returns stub gaps.
// TODO: compare embeddings of roadmap items vs feedback themes using vector store.
func mockDetectRoadmapGaps(roadmapItems, feedbackThemes []string) []product.RoadmapGap {
	_, _ = roadmapItems, feedbackThemes // will drive real gap analysis
	return []product.RoadmapGap{
		{
			Need:          "Bulk CSV export",
			FeedbackCount: 28,
			Priority:      "high",
			Reasoning:     "28 customers requested bulk export; no roadmap item covers it. Finance segment churn risk.",
		},
		{
			Need:          "Report pagination / lazy loading",
			FeedbackCount: 42,
			Priority:      "critical",
			Reasoning:     "Largest feedback cluster with highest NPS impact (4.2 pts). No performance roadmap item matches.",
		},
	}
}

// mockNPSSummary returns stub NPS data.
// TODO: query survey responses from storage once Delighted/Typeform integration is live.
func mockNPSSummary(segment, dateRange string) product.NPSSummary {
	_ = dateRange // will be used in real impl
	return product.NPSSummary{
		Segment:    segment,
		DateRange:  dateRange,
		Score:      38.0,
		Promoters:  61,
		Passives:   22,
		Detractors: 17,
		TopDrivers: []string{"ease of onboarding", "report performance", "missing bulk export"},
	}
}
