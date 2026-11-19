package llm

import "context"

// Message is a single turn in a conversation.
// Role is one of: "user", "assistant", "system", "tool"
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`       // tool name (role=tool)
	ToolCallID string     `json:"tool_call_id,omitempty"` // result target (role=tool)
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"` // requests from LLM (role=assistant)
}

// Provider sends messages to an LLM and returns the assistant reply.
type Provider interface {
	Complete(ctx context.Context, model string, messages []Message) (string, error)
}

// StreamProvider extends Provider with streaming support.
type StreamProvider interface {
	Provider
	Stream(ctx context.Context, model string, messages []Message, out chan<- string) error
}

// DisclosurePayload records a cloud-LLM call for the audit log (charter §3.6).
type DisclosurePayload struct {
	Provider  string
	Model     string
	Tag       string
	TokensIn  int
	TokensOut int
}

// AuditLogger is the audit interface; noopAudit is used until Wright's storage branch merges.
type AuditLogger interface {
	Disclosure(ctx context.Context, payload DisclosurePayload) error
}

type noopAudit struct{}

func (n noopAudit) Disclosure(_ context.Context, _ DisclosurePayload) error { return nil }

// NoopAudit returns an audit logger that silently drops all events.
func NoopAudit() AuditLogger { return noopAudit{} }
