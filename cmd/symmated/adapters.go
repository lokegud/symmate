package main

import (
	"context"
	"strings"
	"time"

	"github.com/loke/symmate/internal/audit"
	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/symmate/internal/llm"
	"github.com/loke/symmate/internal/mcp"
	"github.com/loke/symmate/internal/storage"
)

// auditStoreBridge wraps *storage.DB to satisfy audit.AuditStoreWriter
// with a no-op Close so main controls DB lifetime independently.
type auditStoreBridge struct{ db *storage.DB }

func (b *auditStoreBridge) InsertAuditEvent(
	eventID string, ts int64, category, severity string,
	payloadJSON []byte, conversationID *string,
) error {
	return b.db.InsertAuditEvent(eventID, ts, category, severity, payloadJSON, conversationID)
}
func (b *auditStoreBridge) Close() error { return nil }

// llmAuditBridge bridges llm.AuditLogger → audit.Logger.
type llmAuditBridge struct{ al audit.Logger }

func (b *llmAuditBridge) Disclosure(_ context.Context, p llm.DisclosurePayload) error {
	return b.al.Disclosure(&audit.DisclosurePayload{
		DataType:     "llm-call",
		SourceAgent:  p.Provider + "/" + p.Model,
		SeverityType: "low",
	}, audit.SeverityInfo, nil)
}

// firewallAuditBridge bridges firewall.AuditLogger → audit.Logger.
type firewallAuditBridge struct{ al audit.Logger }

func (b *firewallAuditBridge) FirewallDeny(_ context.Context, e firewall.FirewallEvent) error {
	return b.al.FirewallDeny(&audit.FirewallDenyPayload{
		RuleName:    "pii-deny",
		RequestType: "outbound-response",
		Reason:      classNames(e.Matches),
	}, audit.SeverityWarning, nil)
}

func (b *firewallAuditBridge) FirewallAsk(_ context.Context, e firewall.FirewallEvent) error {
	return b.al.FirewallAsk(&audit.FirewallAskPayload{
		RequestType: "outbound-response",
		RequestedBy: "firewall",
	}, audit.SeverityNotice, nil)
}

// mcpAuditBridge bridges mcp.AuditLogger → audit.Logger.
type mcpAuditBridge struct{ al audit.Logger }

func (b *mcpAuditBridge) Log(e mcp.AuditEvent) {
	status := "success"
	if !e.Allowed {
		status = "denied"
	}
	_ = b.al.ToolCall(&audit.ToolCallPayload{
		ToolName:     e.ToolName,
		ArgCount:     0,
		ResultStatus: status,
		ErrorMsg:     e.DenyReason,
		ExecutionMs:  time.Since(e.Timestamp).Milliseconds(),
		Metadata:     map[string]string{"session_id": e.SessionID, "caller_ip": e.CallerIP},
	}, audit.SeverityInfo, nil)
}

func classNames(matches []firewall.PIIMatch) string {
	seen := make(map[string]bool)
	var parts []string
	for _, m := range matches {
		if !seen[m.Class] {
			seen[m.Class] = true
			parts = append(parts, m.Class)
		}
	}
	return strings.Join(parts, ",")
}
