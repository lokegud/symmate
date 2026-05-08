package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// Model constants — matches runtime.md §21.1.
const (
	ModelSonnet = "claude-sonnet-4-6"
	ModelHaiku  = "claude-haiku-4-5"
	ModelOpus   = "claude-opus-4-7"
)

// Router dispatches LLM calls to the correct provider based on task tag.
type Router struct {
	anthropic    Provider
	hermes       Provider
	hermesModel  string // model name passed to the hermes provider
	audit        AuditLogger
	overrideStore *OverrideStore // optional; owner-configured default provider
}

// NewRouter creates a Router with the default providers drawn from environment.
// ANTHROPIC_API_KEY must be set for Anthropic providers.
// STAK_API_KEY must be set for Hermes.
// HERMES_ENDPOINT overrides the default http://stak:4000/v1.
// HERMES_MODEL overrides the model name sent to the hermes endpoint (default: "hermes").
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
	hermesModel := os.Getenv("HERMES_MODEL")
	if hermesModel == "" {
		hermesModel = "hermes"
	}
	return &Router{
		anthropic:   NewAnthropicProvider(anthropicKey),
		hermes:      NewHermesProvider(hermesEndpoint, stakKey),
		hermesModel: hermesModel,
		audit:       audit,
	}
}

// NewRouterWith creates a Router with explicit providers (for testing).
func NewRouterWith(anthropic, hermes Provider, audit AuditLogger) *Router {
	if audit == nil {
		audit = NoopAudit()
	}
	return &Router{anthropic: anthropic, hermes: hermes, hermesModel: "hermes", audit: audit}
}

// WithOverride attaches an owner-configurable LLM override store to the router.
// Call before any Route calls. Safe for concurrent use after attachment.
func (r *Router) WithOverride(store *OverrideStore) {
	r.overrideStore = store
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
		// Privacy-bound always routes to local Hermes — never to cloud.
		return r.hermes, r.hermesModel, "hermes"
	default: // TagDefault and anything unrecognised
		return r.resolveDefault()
	}
}

// resolveDefault handles the `default` tag with optional owner override.
func (r *Router) resolveDefault() (Provider, string, string) {
	if r.overrideStore == nil {
		return r.hermes, r.hermesModel, "hermes"
	}
	cfg := r.overrideStore.Get()
	if cfg == nil {
		return r.hermes, r.hermesModel, "hermes"
	}
	switch cfg.Provider {
	case "anthropic":
		model := cfg.Model
		if model == "" {
			model = ModelSonnet
		}
		return r.anthropic, model, "anthropic"
	case "openai-compat":
		ep := cfg.CustomEndpoint
		if ep == "" {
			ep = "http://stak:4000/v1"
		}
		model := cfg.Model
		if model == "" {
			model = "default"
		}
		return NewHermesProvider(ep, cfg.CustomAPIKey), model, "openai-compat"
	default: // "hermes" or anything else
		model := cfg.Model
		if model == "" {
			model = r.hermesModel
		}
		return r.hermes, model, "hermes"
	}
}

// RouteWithTools runs an agentic loop: calls CompleteStep, executes tool calls via exec,
// and repeats until the LLM returns a final text reply (or maxIter is exhausted).
// Falls back to Route if the provider doesn't implement ToolProvider.
func (r *Router) RouteWithTools(ctx context.Context, tag string, messages []Message, tools []ToolDef, exec ToolExecutor) (string, error) {
	if len(tools) == 0 {
		return r.Route(ctx, tag, messages)
	}
	provider, model, providerName := r.resolve(tag)
	tp, ok := provider.(ToolProvider)
	if !ok {
		log.Warn().Str("provider", providerName).Str("model", model).Msg("provider does not support tool calls; falling back to plain Route")
		return r.Route(ctx, tag, messages)
	}

	msgs := make([]Message, len(messages))
	copy(msgs, messages)

	const maxIter = 10
	for i := 0; i < maxIter; i++ {
		step, err := tp.CompleteStep(ctx, model, msgs, tools)
		if err != nil {
			return "", fmt.Errorf("llm [%s/%s]: %w", providerName, model, err)
		}
		if len(step.ToolCalls) == 0 {
			_ = r.audit.Disclosure(ctx, DisclosurePayload{Provider: providerName, Model: model, Tag: tag})
			return step.Content, nil
		}
		// Append assistant turn with tool_calls.
		msgs = append(msgs, Message{Role: "assistant", ToolCalls: step.ToolCalls})
		// Execute each tool call and append tool result messages.
		for _, tc := range step.ToolCalls {
			result := exec(ctx, tc.Name, tc.Args)
			msgs = append(msgs, Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
				Name:       tc.Name,
			})
		}
	}
	return "", fmt.Errorf("llm: tool loop exceeded %d iterations", maxIter)
}

// RouteStream routes messages to the correct provider via streaming.
// Tokens are sent to out as they arrive. The caller owns out and must close it.
// Audit disclosure is fired once all tokens have been received.
func (r *Router) RouteStream(ctx context.Context, tag string, messages []Message, out chan<- string) error {
	provider, model, providerName := r.resolve(tag)

	sp, ok := provider.(StreamProvider)
	if !ok {
		// Provider doesn't support streaming — wrap Complete.
		reply, err := provider.Complete(ctx, model, messages)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- reply:
		}
		_ = r.audit.Disclosure(ctx, DisclosurePayload{Provider: providerName, Model: model, Tag: tag})
		return nil
	}

	// buf is owned by sp.Stream (it closes it); RouteStream forwards to out.
	buf := make(chan string, 64)
	errc := make(chan error, 1)
	go func() {
		errc <- sp.Stream(ctx, model, messages, buf)
	}()

	var n int
	for tok := range buf {
		n += len(tok)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- tok:
		}
	}

	err := <-errc
	_ = r.audit.Disclosure(ctx, DisclosurePayload{Provider: providerName, Model: model, Tag: tag, TokensOut: n})
	return err
}
