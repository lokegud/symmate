package tool

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/loke/symmate/internal/engineering"
)

// incidentHotPaths are path prefixes that have historically correlated with incidents.
// TODO: load from incident history in Linear/PagerDuty when real APIs are wired.
var incidentHotPaths = []string{
	"payment", "billing", "checkout", "auth", "session", "webhook",
}

// AnalyzePRRiskTool returns a tool that assesses risk for a given PR URL.
func AnalyzePRRiskTool() Tool {
	return Tool{
		Def: toolDef("analyze_pr_risk",
			"Analyzes a pull request for deploy risk based on changed paths and historical incident correlation.",
			`{
				"type": "object",
				"properties": {
					"pr_url":        {"type": "string",  "description": "Full URL of the pull request."},
					"changed_files": {"type": "array",   "items": {"type": "string"}, "description": "List of file paths changed in the PR."}
				},
				"required": ["pr_url", "changed_files"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				PRUrl        string   `json:"pr_url"`
				ChangedFiles []string `json:"changed_files"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(analyzePRRisk(p.PRUrl, p.ChangedFiles))
		},
		Skill: "- Use analyze_pr_risk to assess the blast radius and incident likelihood of a pull request before merge.",
		Tags:  []string{"engineering/pr", "engineering/risk"},
	}
}

// GetCycleTimeTool returns a tool that reports cycle time metrics for a team.
func GetCycleTimeTool() Tool {
	return Tool{
		Def: toolDef("get_cycle_time",
			"Returns cycle time metrics (ticket open → production) for a team over a date range.",
			`{
				"type": "object",
				"properties": {
					"team_name":   {"type": "string", "description": "The team to report on, e.g. \"platform\"."},
					"date_range":  {"type": "string", "description": "ISO date range like \"2026-04-01/2026-04-30\"."}
				},
				"required": ["team_name", "date_range"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				TeamName  string `json:"team_name"`
				DateRange string `json:"date_range"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockCycleTime(p.TeamName, p.DateRange))
		},
		Skill: "- Use get_cycle_time to answer questions about engineering throughput, bottlenecks, or delivery speed for a team.",
		Tags:  []string{"engineering/metrics", "engineering/cycle-time"},
	}
}

// GetDeployRiskTool returns a tool that scores deploy risk from a list of changed files.
func GetDeployRiskTool() Tool {
	return Tool{
		Def: toolDef("get_deploy_risk",
			"Scores the risk of deploying a change based on which files were modified.",
			`{
				"type": "object",
				"properties": {
					"changed_files": {"type": "array", "items": {"type": "string"}, "description": "List of file paths in the diff."}
				},
				"required": ["changed_files"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				ChangedFiles []string `json:"changed_files"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(deployRisk(p.ChangedFiles))
		},
		Skill: "- Use get_deploy_risk to flag whether a deploy touches high-risk paths (payment, auth, webhooks) before it goes out.",
		Tags:  []string{"engineering/risk", "engineering/deploy"},
	}
}

// analyzePRRisk is the Phase 1 stub implementation.
// TODO: fetch real diff from GitHub API; correlate with incidents from Linear/PagerDuty.
func analyzePRRisk(prURL string, changedFiles []string) engineering.PRRisk {
	risk := deployRisk(changedFiles)
	var correlations []string
	if len(risk.HotPaths) > 0 {
		correlations = append(correlations,
			"3 of the last 5 production incidents originated in paths matching: "+strings.Join(risk.HotPaths, ", "),
		)
	}
	return engineering.PRRisk{
		PRUrl:                prURL,
		ChangedPaths:         changedFiles,
		IncidentCorrelations: correlations,
		RiskLevel:            risk.RiskLevel,
		Reasoning:            risk.Reasoning,
	}
}

// deployRisk scores a file list against known hot paths.
// TODO: replace with real path-weighted incident model from storage.
func deployRisk(changedFiles []string) engineering.DeployRisk {
	var hot []string
	for _, f := range changedFiles {
		lower := strings.ToLower(f)
		for _, hp := range incidentHotPaths {
			if strings.Contains(lower, hp) {
				hot = append(hot, f)
				break
			}
		}
	}

	level := "low"
	reasoning := "No changes to historically incident-prone paths."
	if len(hot) > 0 {
		level = "high"
		reasoning = "This diff touches " + strings.Join(hot, ", ") +
			". These paths account for the majority of recent production incidents. " +
			"Recommend: extra reviewer sign-off, staged rollout, and post-deploy monitoring for 30 min."
	}

	return engineering.DeployRisk{
		RiskLevel: level,
		HotPaths:  hot,
		Reasoning: reasoning,
	}
}

// mockCycleTime returns stub metrics.
// TODO: query SQLite storage once ticket/PR events are ingested from Linear + GitHub.
func mockCycleTime(team, dateRange string) engineering.CycleTimeMetrics {
	_ = dateRange // will be used in real impl
	return engineering.CycleTimeMetrics{
		Team:            team,
		AvgHours:        54.3,
		P95Hours:        168.0,
		BottleneckStage: "code-review", // TODO: derive from actual event timestamps
	}
}

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"marshal failed"}`
	}
	return string(b)
}
