package a2a

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	sdk "github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/symmate/internal/llm"
)

// PersonalitySource provides the current system prompt.
type PersonalitySource interface {
	Current() string
}

// LLMRouter routes messages to an LLM provider.
type LLMRouter interface {
	Route(ctx context.Context, tag string, messages []llm.Message) (string, error)
}

// SymmateExecutor implements a2asrv.AgentExecutor.
//
// Sygil is agent-first: instead of pre-deciding refusals based on the registry,
// the executor surfaces the registry's signal to the LLM as a system message.
// The agent's constitution + reasoning decide whether to engage, with what
// posture, and what to disclose. Sygil informs; the agent decides.
//
// Layering:
//  1. Sygil signal injection — fetch sender intel, prepend as system message
//  2. inbound prompt-injection check (fail closed with a polite reply)
//  3. outbound privacy firewall (PII redaction)
type SymmateExecutor struct {
	personality   PersonalitySource
	router        LLMRouter
	firewall      *firewall.Firewall
	sygil         *SygilClient // nil = disabled
	defaultGrants []firewall.Grant
}

// NewSymmateExecutor wires the executor with its dependencies.
func NewSymmateExecutor(p PersonalitySource, r LLMRouter, fw *firewall.Firewall) *SymmateExecutor {
	return &SymmateExecutor{
		personality:   p,
		router:        r,
		firewall:      fw,
		defaultGrants: []firewall.Grant{},
	}
}

// WithSygil attaches a Sygil client. When attached, the executor injects a
// "Sender Intel" system message ahead of each request so the agent's LLM can
// factor identity + reputation signal into its reply.
func (e *SymmateExecutor) WithSygil(sc *SygilClient) *SymmateExecutor {
	e.sygil = sc
	return e
}

// Compile-time interface check.
var _ a2asrv.AgentExecutor = (*SymmateExecutor)(nil)

// Execute is invoked by the SDK for each inbound message/send.
func (e *SymmateExecutor) Execute(ctx context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[sdk.Event, error] {
	return func(yield func(sdk.Event, error) bool) {
		if execCtx.Message == nil {
			yield(nil, fmt.Errorf("no message in execution context"))
			return
		}

		// Layer 1: Sygil signal injection (agent-first — signal in, decision out).
		// The registry's view of the sender is rendered as a system message
		// alongside the agent's constitution; the LLM decides what to do.
		var senderIntel string
		if senderCard := SenderCardURL(ctx); senderCard != "" {
			sig := e.sygil.Lookup(ctx, senderCard)
			senderIntel = sig.RenderForLLM()
		}

		text := extractText(execCtx.Message)
		if text == "" {
			yield(nil, fmt.Errorf("message contains no text"))
			return
		}

		// Layer 2: injection check before LLM call.
		if pattern := detectInjection(text); pattern != "" {
			yield(sdk.NewMessage(
				sdk.MessageRoleAgent,
				sdk.NewTextPart("I noticed something in that message I'm not going to act on."),
			), nil)
			return
		}

		sysPrompt := e.personality.Current()
		tag := llm.Classify(text)
		messages := []llm.Message{
			{Role: "system", Content: sysPrompt},
		}
		if senderIntel != "" {
			messages = append(messages, llm.Message{Role: "system", Content: senderIntel})
		}
		messages = append(messages, llm.Message{Role: "user", Content: text})

		llmCtx, llmCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer llmCancel()

		reply, err := e.router.Route(llmCtx, tag, messages)
		if err != nil {
			yield(nil, fmt.Errorf("llm error: %w", err))
			return
		}

		// Layer 3: privacy firewall before responding.
		filtered, _ := e.firewall.Filter(ctx, reply, e.defaultGrants)

		yield(sdk.NewMessage(sdk.MessageRoleAgent, sdk.NewTextPart(filtered)), nil)
	}
}

// Cancel is required by the AgentExecutor interface. We don't run long tasks,
// so cancel is a no-op for now.
func (e *SymmateExecutor) Cancel(ctx context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[sdk.Event, error] {
	return func(yield func(sdk.Event, error) bool) {}
}

// extractText concatenates all text parts of a message into one string.
func extractText(m *sdk.Message) string {
	var sb strings.Builder
	for _, p := range m.Parts {
		if p == nil {
			continue
		}
		if t, ok := p.Content.(sdk.Text); ok {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(string(t))
		}
	}
	return sb.String()
}

// injectionPatterns are obvious prompt-injection signals.
var injectionPatterns = []string{
	"ignore previous instructions",
	"ignore all previous",
	"disregard your instructions",
	"forget your instructions",
	"your new instructions are",
	"[[system]]",
	"[[prompt]]",
	"<|system|>",
	"system override",
	"you are now a",
	"pretend you have no",
	"act as if you have no restrictions",
	"jailbreak",
	"do anything now",
	"dan mode",
}

func detectInjection(text string) string {
	lower := strings.ToLower(text)
	for _, p := range injectionPatterns {
		if strings.Contains(lower, p) {
			return p
		}
	}
	return ""
}
