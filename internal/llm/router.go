package llm

import (
	"context"
	"fmt"
	"os"
)

// Model constants — matches runtime.md §21.1.
const (
	ModelSonnet = "claude-sonnet-4-6"
	ModelHaiku  = "claude-haiku-4-5-20251001"
	ModelOpus   = "claude-opus-4-7"
)

// Router dispatches LLM calls to the correct provider based on task tag.
type Router struct {
	anthropic Provider
	hermes    Provider
	audit     AuditLogger
}

// NewRouter creates a Router with the default providers drawn from environment.
// ANTHROPIC_API_KEY must be set for Anthropic providers.
// STAK_API_KEY must be set for Hermes.
// HERMES_ENDPOINT overrides the default http://stak:4000/v1.
func NewRouter(audit AuditLogger) *Router {
	if audit == nil {
		audit = NoopAudit()
	}
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	stakKey := os.Getenv("STAK_API_KEY")
	hermesEndpoint := os.Getenv("HERMES_ENDPOINT")
	if hermesEndpoint == "" {
		hermesEndpoint = "http://stak:4000/v1"
	}
	return &Router{
		anthropic: NewAnthropicProvider(anthropicKey),
		hermes:    NewHermesProvider(hermesEndpoint, stakKey),
		audit:     audit,
	}
}

// NewRouterWith creates a Router with explicit providers (for testing).
func NewRouterWith(anthropic, hermes Provider, audit AuditLogger) *Router {
	if audit == nil {
		audit = NoopAudit()
	}
	return &Router{anthropic: anthropic, hermes: hermes, audit: audit}
}

// Route sends messages to the correct provider for the given tag and returns the reply.
func (r *Router) Route(ctx context.Context, tag string, messages []Message) (string, error) {
	provider, model, providerName := r.resolve(tag)
	reply, err := provider.Complete(ctx, model, messages)
	if err != nil {
		return "", fmt.Errorf("llm [%s/%s]: %w", providerName, model, err)
	}
	// Charter §3.6: disclose every cloud call to the audit log.
	_ = r.audit.Disclosure(ctx, DisclosurePayload{
		Provider: providerName,
		Model:    model,
		Tag:      tag,
	})
	return reply, nil
}

// resolve maps a tag to (provider, model, providerName).
func (r *Router) resolve(tag string) (Provider, string, string) {
	switch tag {
	case TagLight:
		return r.anthropic, ModelHaiku, "anthropic"
	case TagHardReason:
		return r.anthropic, ModelOpus, "anthropic"
	case TagPrivacyBound:
		return r.hermes, "hermes", "hermes"
	default: // TagDefault and anything unrecognised
		return r.anthropic, ModelSonnet, "anthropic"
	}
}
