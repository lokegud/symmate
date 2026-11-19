package firewall

import (
	"context"
	"strings"
	"testing"
)

type captureAudit struct {
	denies []FirewallEvent
	asks   []FirewallEvent
}

func (a *captureAudit) FirewallDeny(_ context.Context, e FirewallEvent) error {
	a.denies = append(a.denies, e)
	return nil
}
func (a *captureAudit) FirewallAsk(_ context.Context, e FirewallEvent) error {
	a.asks = append(a.asks, e)
	return nil
}

func TestFilter_NoGrants_PIIRedacted(t *testing.T) {
	audit := &captureAudit{}
	fw := New(audit)

	text := "Please contact me at alice@example.com or 555-867-5309."
	filtered, dec := fw.Filter(context.Background(), text, nil)

	if dec != DecisionDeny {
		t.Errorf("expected DecisionDeny, got %s", dec)
	}
	if strings.Contains(filtered, "alice@example.com") {
		t.Error("email should have been redacted")
	}
	if len(audit.denies) == 0 {
		t.Error("expected audit deny event")
	}
}

func TestFilter_AllowGrant_PIIKept(t *testing.T) {
	fw := New(nil)
	grants := []Grant{
		{DataClass: "email", Permission: "allow"},
	}
	text := "Reach me at bob@example.org"
	filtered, dec := fw.Filter(context.Background(), text, grants)

	if dec != DecisionAllow {
		t.Errorf("expected allow, got %s", dec)
	}
	if !strings.Contains(filtered, "bob@example.org") {
		t.Errorf("allowed email should not be redacted, got: %q", filtered)
	}
}

func TestFilter_AskGrant_AuditEmitted(t *testing.T) {
	audit := &captureAudit{}
	fw := New(audit)
	grants := []Grant{
		{DataClass: "phone", Permission: "ask"},
	}
	text := "Call me at 415-555-1234"
	_, dec := fw.Filter(context.Background(), text, grants)

	if dec != DecisionAskHost {
		t.Errorf("expected ask-owner, got %s", dec)
	}
	if len(audit.asks) == 0 {
		t.Error("expected audit ask event")
	}
}

func TestFilter_NoText(t *testing.T) {
	fw := New(nil)
	filtered, dec := fw.Filter(context.Background(), "no PII here", nil)
	// No PII → allow, unchanged
	if dec != DecisionAllow {
		t.Errorf("expected allow for clean text, got %s", dec)
	}
	if filtered != "no PII here" {
		t.Errorf("text should be unchanged, got %q", filtered)
	}
}

func TestExtract_Email(t *testing.T) {
	matches := Extract("send to test@example.com please")
	if len(matches) == 0 {
		t.Fatal("expected email match")
	}
	if matches[0].Class != "email" {
		t.Errorf("wrong class: %q", matches[0].Class)
	}
	if matches[0].Value != "test@example.com" {
		t.Errorf("wrong value: %q", matches[0].Value)
	}
}

func TestExtract_Phone(t *testing.T) {
	matches := Extract("call 415-555-9876 for info")
	phones := filterClass(matches, "phone")
	if len(phones) == 0 {
		t.Fatal("expected phone match")
	}
}

func TestRewrite_CoarseEmail(t *testing.T) {
	text := "contact user@gmail.com for help"
	matches := []PIIMatch{{Class: "email", Value: "user@gmail.com", Start: 8, End: 22}}
	got := Rewrite(text, matches, nil) // nil grants → deny
	if strings.Contains(got, "user@gmail.com") {
		t.Error("email should be redacted")
	}
	if !strings.Contains(got, "@gmail.com") {
		t.Errorf("expected domain hint, got: %q", got)
	}
}

func TestDecisionOrdering(t *testing.T) {
	if DecisionAllow >= DecisionAskHost {
		t.Error("allow should be less than ask-owner")
	}
	if DecisionAskHost >= DecisionDeny {
		t.Error("ask-owner should be less than deny")
	}
}

func filterClass(matches []PIIMatch, class string) []PIIMatch {
	var out []PIIMatch
	for _, m := range matches {
		if m.Class == class {
			out = append(out, m)
		}
	}
	return out
}
