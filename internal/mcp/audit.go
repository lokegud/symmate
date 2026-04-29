package mcp

import "time"

// AuditEvent records a single MCP tool invocation.
type AuditEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	SessionID  string    `json:"session_id"`
	ToolName   string    `json:"tool_name"`
	CallerIP   string    `json:"caller_ip"`
	Allowed    bool      `json:"allowed"`
	DenyReason string    `json:"deny_reason,omitempty"`
}

// AuditLogger is the interface for recording MCP tool calls.
// The stub logs to zerolog; a real implementation writes to the storage layer
// once Wright's branch merges.
type AuditLogger interface {
	Log(ev AuditEvent)
}

// NoopAuditLogger satisfies AuditLogger without writing anything.
// Used in tests.
type NoopAuditLogger struct{}

func (NoopAuditLogger) Log(_ AuditEvent) {}
