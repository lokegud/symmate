package audit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Logger is the interface for audit logging
type Logger interface {
	CapabilityGrant(payload *CapabilityGrantPayload, severity Severity, conversationID *string) error
	PolicyChange(payload *PolicyChangePayload, severity Severity, conversationID *string) error
	Escalation(payload *EscalationPayload, severity Severity, conversationID *string) error
	AbuseFlag(payload *AbuseFlagPayload, severity Severity, conversationID *string) error
	Injection(payload *InjectionPayload, severity Severity, conversationID *string) error
	Disclosure(payload *DisclosurePayload, severity Severity, conversationID *string) error
	ToolCall(payload *ToolCallPayload, severity Severity, conversationID *string) error
	FirewallDeny(payload *FirewallDenyPayload, severity Severity, conversationID *string) error
	FirewallAsk(payload *FirewallAskPayload, severity Severity, conversationID *string) error
	Close() error
}

type auditLogger struct {
	store Store
}

// NewLogger creates a new audit logger backed by a store
func NewLogger(store Store) Logger {
	return &auditLogger{
		store: store,
	}
}

func (l *auditLogger) CapabilityGrant(payload *CapabilityGrantPayload, severity Severity, conversationID *string) error {
	return l.log("capability_grant", payload, severity, conversationID)
}

func (l *auditLogger) PolicyChange(payload *PolicyChangePayload, severity Severity, conversationID *string) error {
	return l.log("policy_change", payload, severity, conversationID)
}

func (l *auditLogger) Escalation(payload *EscalationPayload, severity Severity, conversationID *string) error {
	return l.log("escalation", payload, severity, conversationID)
}

func (l *auditLogger) AbuseFlag(payload *AbuseFlagPayload, severity Severity, conversationID *string) error {
	return l.log("abuse_flag", payload, severity, conversationID)
}

func (l *auditLogger) Injection(payload *InjectionPayload, severity Severity, conversationID *string) error {
	return l.log("injection", payload, severity, conversationID)
}

func (l *auditLogger) Disclosure(payload *DisclosurePayload, severity Severity, conversationID *string) error {
	return l.log("disclosure", payload, severity, conversationID)
}

func (l *auditLogger) ToolCall(payload *ToolCallPayload, severity Severity, conversationID *string) error {
	return l.log("tool_call", payload, severity, conversationID)
}

func (l *auditLogger) FirewallDeny(payload *FirewallDenyPayload, severity Severity, conversationID *string) error {
	return l.log("firewall_deny", payload, severity, conversationID)
}

func (l *auditLogger) FirewallAsk(payload *FirewallAskPayload, severity Severity, conversationID *string) error {
	return l.log("firewall_ask", payload, severity, conversationID)
}

func (l *auditLogger) Close() error {
	return l.store.Close()
}

// log is the internal helper that writes to the store
func (l *auditLogger) log(category string, payload interface{}, severity Severity, conversationID *string) error {
	event := &AuditEvent{
		ID:             generateEventID(),
		Timestamp:      time.Now().UTC(),
		Category:       category,
		Severity:       severity,
		ConversationID: conversationID,
		Payload:        payload,
	}
	return l.store.Write(event)
}

// generateEventID creates a unique event ID using crypto/rand
func generateEventID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate event ID: %v", err))
	}
	return hex.EncodeToString(b)
}
