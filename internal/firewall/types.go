package firewall

import "context"

// Decision is the firewall's verdict on a piece of text.
type Decision int

const (
	DecisionAllow    Decision = iota
	DecisionAskOwner          // default-deny on timeout; owner must be notified
	DecisionDeny
)

func (d Decision) String() string {
	switch d {
	case DecisionAllow:
		return "allow"
	case DecisionAskOwner:
		return "ask-owner"
	case DecisionDeny:
		return "deny"
	default:
		return "unknown"
	}
}

// Grant represents a capability-grant row: what data class this scope may receive.
type Grant struct {
	DataClass  string // "email", "phone", "address", "name", "location-coarse", etc.
	Permission string // "allow", "deny", "ask"
}

// PIIMatch is a detected PII span within a text.
type PIIMatch struct {
	Class string // "email", "phone", "address", "name"
	Value string // the matched text
	Start int
	End   int
}

// FirewallEvent records the filter decision for the audit log.
type FirewallEvent struct {
	Decision Decision
	Matches  []PIIMatch
	Redacted bool
}

// AuditLogger records firewall events; stubbed until Wright's storage branch merges.
type AuditLogger interface {
	FirewallDeny(ctx context.Context, event FirewallEvent) error
	FirewallAsk(ctx context.Context, event FirewallEvent) error
}

type noopAudit struct{}

func (n noopAudit) FirewallDeny(_ context.Context, _ FirewallEvent) error { return nil }
func (n noopAudit) FirewallAsk(_ context.Context, _ FirewallEvent) error  { return nil }

// NoopAudit returns a no-op audit logger.
func NoopAudit() AuditLogger { return noopAudit{} }
