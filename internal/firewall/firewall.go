package firewall

import "context"

// Firewall applies PII extraction, grant lookup, and least-disclosure rewriting.
type Firewall struct {
	audit AuditLogger
}

// New creates a Firewall with the given audit logger.
func New(audit AuditLogger) *Firewall {
	if audit == nil {
		audit = NoopAudit()
	}
	return &Firewall{audit: audit}
}

// Filter inspects text, applies grants, and returns the rewritten text plus an overall decision.
// Decision is the most restrictive verdict across all detected PII classes:
//
//	deny > ask-host > allow
//
// The returned text has denied/asked spans replaced by coarse equivalents.
func (f *Firewall) Filter(ctx context.Context, text string, grants []Grant) (filtered string, decision Decision) {
	matches := Extract(text)

	if len(matches) == 0 {
		return text, DecisionAllow
	}

	grantMap := buildGrantMap(grants)
	decision = DecisionAllow

	for _, m := range matches {
		perm := resolvePermission(m.Class, grantMap)
		switch perm {
		case "deny":
			if decision < DecisionDeny {
				decision = DecisionDeny
			}
		case "ask":
			if decision < DecisionAskHost {
				decision = DecisionAskHost
			}
		}
	}

	filtered = Rewrite(text, matches, grants)

	event := FirewallEvent{
		Decision: decision,
		Matches:  matches,
		Redacted: filtered != text,
	}
	switch decision {
	case DecisionDeny:
		_ = f.audit.FirewallDeny(ctx, event)
	case DecisionAskHost:
		_ = f.audit.FirewallAsk(ctx, event)
	}

	return filtered, decision
}
