package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/loke/symmate/internal/security"
)

// TriageAlertTool returns a tool that scores an incoming security alert.
func TriageAlertTool() Tool {
	return Tool{
		Def: toolDef("triage_alert",
			"Scores an incoming security alert for severity and suppression eligibility.",
			`{
				"type": "object",
				"properties": {
					"alert_id": {"type": "string", "description": "Unique identifier for the alert."},
					"source":   {"type": "string", "description": "Alert origin, e.g. \"crowdstrike\", \"cloudtrail\"."},
					"summary": {"type": "string", "description": "Short human-readable description of the alert."}
				},
				"required": ["alert_id", "source", "summary"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				AlertID string `json:"alert_id"`
				Source  string `json:"source"`
				Summary string `json:"summary"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockTriageAlert(p.AlertID, p.Source, p.Summary))
		},
		Skill: "- Use triage_alert to score an incoming security alert, determine severity, and decide whether to suppress it as noise.",
		Tags:  []string{"security/primary"},
	}
}

// GetCompliancePostureTool returns a tool that reports compliance status for a framework.
func GetCompliancePostureTool() Tool {
	return Tool{
		Def: toolDef("get_compliance_posture",
			"Returns the current compliance posture for a given framework (e.g. SOC2, ISO27001, PCI-DSS).",
			`{
				"type": "object",
				"properties": {
					"framework": {"type": "string", "description": "Compliance framework name, e.g. \"SOC2\"."}
				},
				"required": ["framework"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				Framework string `json:"framework"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockCompliancePosture(p.Framework))
		},
		Skill: "- Use get_compliance_posture to answer questions about the org's standing against a compliance framework and which controls are failing.",
		Tags:  []string{"security/primary", "security/metrics"},
	}
}

// CorrelateCVETool returns a tool that maps a CVE to affected internal services.
func CorrelateCVETool() Tool {
	return Tool{
		Def: toolDef("correlate_cve",
			"Correlates a CVE with internal services to identify affected components and exposure level.",
			`{
				"type": "object",
				"properties": {
					"cve_id": {"type": "string", "description": "CVE identifier, e.g. \"CVE-2024-1234\"."},
					"cvss":   {"type": "number", "description": "CVSS base score (0.0–10.0)."}
				},
				"required": ["cve_id"]
			}`),
		Execute: func(_ context.Context, args json.RawMessage) string {
			var p struct {
				CVEID string  `json:"cve_id"`
				CVSS  float64 `json:"cvss"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			return jsonString(mockCorrelateCVE(p.CVEID, p.CVSS))
		},
		Skill: "- Use correlate_cve to determine which internal services are affected by a CVE and how exposed the org is.",
		Tags:  []string{"security/primary"},
	}
}

// mockTriageAlert is the Phase 1 stub.
// TODO: call real SIEM/EDR APIs and score against MITRE ATT&CK.
func mockTriageAlert(alertID, source, summary string) security.AlertTriage {
	lower := strings.ToLower(summary)
	severity := "low"
	score := 2.0
	tags := []string{"informational"}
	suppress := true
	reasoning := "No high-risk indicators detected in alert summary."

	if strings.ContainsAny(lower, "privilege escalation lateral movement credential dump") {
		severity = "critical"
		score = 9.1
		tags = []string{"privilege-escalation", "lateral-movement"}
		suppress = false
		reasoning = fmt.Sprintf("Alert from %s matches high-risk ATT&CK patterns. Immediate investigation required.", source)
	} else if strings.Contains(lower, "login") || strings.Contains(lower, "auth") {
		severity = "medium"
		score = 5.5
		tags = []string{"authentication"}
		suppress = false
		reasoning = "Authentication-related activity warrants review but is not confirmed malicious."
	}

	return security.AlertTriage{
		AlertID:   alertID,
		Source:    source,
		Severity:  severity,
		Score:     score,
		Tags:      tags,
		Reasoning: reasoning,
		Suppress:  suppress,
	}
}

// mockCompliancePosture is the Phase 1 stub.
// TODO: pull real control evidence from GRC platform / audit log store.
func mockCompliancePosture(framework string) security.CompliancePosture {
	_ = framework // will be used to query framework-specific controls
	return security.CompliancePosture{
		Framework:   framework,
		PassCount:   87,
		FailCount:   13,
		Score:       87.0,
		FailedCtls:  []string{"CC6.1", "CC7.2", "CC8.1"},
		Recommended: []string{"Enable MFA on all admin accounts (CC6.1)", "Review access logs retention policy (CC7.2)"},
	}
}

// mockCorrelateCVE is the Phase 1 stub.
// TODO: cross-reference CVE package list against SBOM data in storage.
func mockCorrelateCVE(cveID string, cvss float64) security.CVECorrelation {
	_ = cvss // will inform severity weighting in real impl
	return security.CVECorrelation{
		CVEID:            cveID,
		CVSS:             cvss,
		AffectedServices: []string{"api-gateway", "auth-service"},
		ExposureLevel:    "internet-facing",
		Reasoning:        fmt.Sprintf("%s affects the ingress stack. Both services are reachable from the public internet. Patch or WAF mitigation recommended within 72 hours.", cveID),
	}
}
