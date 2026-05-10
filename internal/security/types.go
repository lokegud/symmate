package security

// AlertTriage holds the result of scoring a security alert.
type AlertTriage struct {
	AlertID   string   `json:"alert_id"`
	Source    string   `json:"source"`             // e.g. "crowdstrike", "cloudtrail", "siem"
	Severity  string   `json:"severity"`           // low | medium | high | critical
	Score     float64  `json:"score"`              // 0.0–10.0
	Tags      []string `json:"tags"`               // e.g. ["lateral-movement", "privilege-escalation"]
	Reasoning string   `json:"reasoning"`
	Suppress  bool     `json:"suppress"`           // true if likely noise / auto-closeable
}

// CompliancePosture holds the compliance status for an org against a framework.
type CompliancePosture struct {
	Framework   string   `json:"framework"`    // e.g. "SOC2", "ISO27001", "PCI-DSS"
	PassCount   int      `json:"pass_count"`
	FailCount   int      `json:"fail_count"`
	Score       float64  `json:"score"`        // 0.0–100.0 percent passing
	FailedCtls  []string `json:"failed_ctls"`  // control IDs that are failing
	Recommended []string `json:"recommended"`  // highest-priority remediation steps
}

// CVECorrelation maps a CVE to affected internal services.
type CVECorrelation struct {
	CVEID            string   `json:"cve_id"`
	CVSS             float64  `json:"cvss"`
	AffectedServices []string `json:"affected_services"`
	ExposureLevel    string   `json:"exposure_level"` // none | internal | internet-facing
	Reasoning        string   `json:"reasoning"`
}
