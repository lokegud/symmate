package legal

// ContractReview holds the result of a contract intake and clause analysis.
type ContractReview struct {
	ContractID     string   `json:"contract_id"`
	ContractType   string   `json:"contract_type"` // nda | msa | sow | saas | other
	DeviatedClauses []string `json:"deviated_clauses"`
	RiskLevel      string   `json:"risk_level"` // low | medium | high | critical
	Summary        string   `json:"summary"`
}

// ComplianceGap holds the result of a compliance gap analysis for an entity.
type ComplianceGap struct {
	EntityID   string   `json:"entity_id"`
	Framework  string   `json:"framework"` // gdpr | soc2 | hipaa | ccpa | iso27001
	Gaps       []string `json:"gaps"`
	GapCount   int      `json:"gap_count"`
	Remediation string  `json:"remediation"`
}

// ClauseDeviation describes a single clause that deviates from standard playbook terms.
type ClauseDeviation struct {
	ClauseName   string `json:"clause_name"`
	StandardText string `json:"standard_text"`
	ActualText   string `json:"actual_text"`
	Severity     string `json:"severity"` // low | medium | high | critical
	Guidance     string `json:"guidance"`
}
