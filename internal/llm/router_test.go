package llm

import (
	"context"
	"testing"
)

// mockProvider records what model it was called with.
type mockProvider struct {
	lastModel string
	reply     string
	err       error
}

func (m *mockProvider) Complete(_ context.Context, model string, _ []Message) (string, error) {
	m.lastModel = model
	if m.err != nil {
		return "", m.err
	}
	return m.reply, nil
}

type mockAudit struct {
	calls []DisclosurePayload
}

func (a *mockAudit) Disclosure(_ context.Context, p DisclosurePayload) error {
	a.calls = append(a.calls, p)
	return nil
}

func TestRouterTagRouting(t *testing.T) {
	anthropic := &mockProvider{reply: "anthropic reply"}
	hermes := &mockProvider{reply: "hermes reply"}
	audit := &mockAudit{}
	r := NewRouterWith(anthropic, hermes, audit)

	msgs := []Message{{Role: "user", Content: "hello"}}

	tests := []struct {
		tag           string
		wantProvider  *mockProvider
		wantModel     string
		wantProvName  string
	}{
		{TagDefault, anthropic, ModelSonnet, "anthropic"},
		{TagLight, anthropic, ModelHaiku, "anthropic"},
		{TagHardReason, anthropic, ModelOpus, "anthropic"},
		{TagPrivacyBound, hermes, "hermes", "hermes"},
		{"unknown-tag", anthropic, ModelSonnet, "anthropic"},
	}

	for _, tc := range tests {
		t.Run(tc.tag, func(t *testing.T) {
			anthropic.lastModel = ""
			hermes.lastModel = ""
			audit.calls = nil

			reply, err := r.Route(context.Background(), tc.tag, msgs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if reply != tc.wantProvider.reply {
				t.Errorf("got reply from wrong provider: %q", reply)
			}
			if tc.wantProvider.lastModel != tc.wantModel {
				t.Errorf("got model %q, want %q", tc.wantProvider.lastModel, tc.wantModel)
			}
			if len(audit.calls) != 1 {
				t.Errorf("expected 1 audit disclosure, got %d", len(audit.calls))
			}
			if audit.calls[0].Provider != tc.wantProvName {
				t.Errorf("audit provider %q, want %q", audit.calls[0].Provider, tc.wantProvName)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"summarize my emails from yesterday", TagPrivacyBound},
		{"check my calendar for Tuesday", TagPrivacyBound},
		{"what are the best deals on cameras", TagDefault},
		{"send message to John", TagPrivacyBound},
		{"who is the president of France", TagDefault},
		{"my password is secret123", TagPrivacyBound},
	}
	for _, tc := range tests {
		got := Classify(tc.text)
		if got != tc.want {
			t.Errorf("Classify(%q) = %q, want %q", tc.text, got, tc.want)
		}
	}
}

func TestRouterAuditOnEveryCall(t *testing.T) {
	anthropic := &mockProvider{reply: "ok"}
	hermes := &mockProvider{reply: "ok"}
	audit := &mockAudit{}
	r := NewRouterWith(anthropic, hermes, audit)

	msgs := []Message{{Role: "user", Content: "hi"}}
	for _, tag := range []string{TagDefault, TagLight, TagPrivacyBound} {
		audit.calls = nil
		r.Route(context.Background(), tag, msgs)
		if len(audit.calls) == 0 {
			t.Errorf("tag %s: expected audit disclosure", tag)
		}
	}
}
