package audit

import "time"

// Severity levels for audit events
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityNotice  Severity = "notice"
	SeverityWarning Severity = "warning"
	SeverityAlert   Severity = "alert"
)

// AuditEvent is the base structure for all audit log entries
type AuditEvent struct {
	ID             string        `json:"id"`
	Timestamp      time.Time     `json:"timestamp"`
	Category       string        `json:"category"`
	Severity       Severity      `json:"severity"`
	ConversationID *string       `json:"conversation_id,omitempty"`
	Payload        interface{}   `json:"payload"`
}

// CapabilityGrantPayload records when a capability is granted to an agent
type CapabilityGrantPayload struct {
	AgentID      string   `json:"agent_id"`
	CapabilityID string   `json:"capability_id"`
	GrantedBy    string   `json:"granted_by"`
	TTL          *int     `json:"ttl_seconds,omitempty"`
	Constraints  []string `json:"constraints,omitempty"`
}

// PolicyChangePayload records changes to runtime policies
type PolicyChangePayload struct {
	PolicyID  string      `json:"policy_id"`
	Operation string      `json:"operation"` // "create", "update", "delete"
	ChangedBy string      `json:"changed_by"`
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty"`
}

// EscalationPayload records permission escalations or privilege increases
type EscalationPayload struct {
	AgentID           string `json:"agent_id"`
	FromCapability    string `json:"from_capability"`
	ToCapability      string `json:"to_capability"`
	Reason            string `json:"reason,omitempty"`
	ApprovedBy        string `json:"approved_by,omitempty"`
	ApprovalMethod    string `json:"approval_method,omitempty"` // "admin", "policy", "auto"
}

// AbuseFlagPayload records potential abuse or policy violations
type AbuseFlagPayload struct {
	AgentID       string   `json:"agent_id"`
	ViolationType string   `json:"violation_type"` // "rate_limit", "resource_abuse", "policy_violation", etc.
	Evidence      []string `json:"evidence,omitempty"`
	Severity      string   `json:"severity"` // "low", "medium", "high"
	Action        string   `json:"action,omitempty"` // "throttle", "block", "alert"
}

// InjectionPayload records attempted injections or invalid input
type InjectionPayload struct {
	InjectionType string `json:"injection_type"` // "prompt", "sql", "os_command", "template", etc.
	Source        string `json:"source"`         // "user_input", "api", "config", etc.
	DetectedBy    string `json:"detected_by"`    // filter/validator name
}

// DisclosurePayload records data disclosures or information leaks
type DisclosurePayload struct {
	DataType     string `json:"data_type"` // "credential", "token", "pii", "secret", etc.
	SourceAgent  string `json:"source_agent,omitempty"`
	TargetAgent  string `json:"target_agent,omitempty"`
	SeverityType string `json:"severity_type"` // "low", "high"
}

// ToolCallPayload records tool/capability invocations
type ToolCallPayload struct {
	AgentID      string            `json:"agent_id"`
	ToolName     string            `json:"tool_name"`
	ToolVersion  string            `json:"tool_version,omitempty"`
	ArgCount     int               `json:"arg_count"`
	ResultStatus string            `json:"result_status"` // "success", "error", "timeout"
	ErrorMsg     string            `json:"error_msg,omitempty"`
	ExecutionMs  int64             `json:"execution_ms,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// FirewallDenyPayload records firewall/policy rejections
type FirewallDenyPayload struct {
	AgentID     string `json:"agent_id"`
	RuleID      string `json:"rule_id"`
	RuleName    string `json:"rule_name"`
	RequestType string `json:"request_type"` // "capability", "resource", "message", etc.
	Reason      string `json:"reason"`
}

// FirewallAskPayload records firewall/policy escalations for human approval
type FirewallAskPayload struct {
	AgentID      string   `json:"agent_id"`
	RequestType  string   `json:"request_type"`
	RequestedBy  string   `json:"requested_by"`
	ApprovalPath []string `json:"approval_path,omitempty"`
	Timeout      int      `json:"timeout_seconds,omitempty"`
}
