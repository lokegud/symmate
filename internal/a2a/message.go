package a2a

import (
	"context"
	"encoding/json"
	"strings"

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

// MessageHandler processes inbound A2A sendMessage requests.
type MessageHandler struct {
	personality PersonalitySource
	router      LLMRouter
	firewall    *firewall.Firewall
	// defaultGrants applied to all A2A responses (conservative; owner configures via policy).
	defaultGrants []firewall.Grant
}

// NewMessageHandler creates a MessageHandler with the given dependencies.
func NewMessageHandler(p PersonalitySource, r LLMRouter, fw *firewall.Firewall) *MessageHandler {
	return &MessageHandler{
		personality: p,
		router:      r,
		firewall:    fw,
		// Default: deny all PII disclosure to anonymous agents.
		defaultGrants: []firewall.Grant{},
	}
}

// Handle processes a sendMessage JSON-RPC call and returns the result or an error.
func (h *MessageHandler) Handle(ctx context.Context, rawParams json.RawMessage) (SendMessageResult, *RPCError) {
	var params SendMessageParams
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return SendMessageResult{}, &RPCError{Code: ErrInvalidParams, Message: "invalid params: " + err.Error()}
	}

	text := params.Text
	if text == "" && params.Intent != "" {
		text = params.Intent
	}
	if text == "" {
		return SendMessageResult{}, &RPCError{Code: ErrInvalidParams, Message: "text or intent required"}
	}

	// All inbound A2A traffic is untrusted — detect injection first.
	if inj := detectInjection(text); inj != "" {
		return SendMessageResult{}, &RPCError{Code: ErrInjection, Message: "injection attempt detected: " + inj}
	}

	sysPrompt := h.personality.Current()
	tag := llm.Classify(text)

	messages := []llm.Message{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: text},
	}

	reply, err := h.router.Route(ctx, tag, messages)
	if err != nil {
		return SendMessageResult{}, &RPCError{Code: ErrInternal, Message: "llm error: " + err.Error()}
	}

	// Apply privacy firewall to outbound response.
	filtered, dec := h.firewall.Filter(ctx, reply, h.defaultGrants)
	wasFiltered := dec != firewall.DecisionAllow || filtered != reply

	return SendMessageResult{Text: filtered, Filtered: wasFiltered}, nil
}

// injectionPatterns are common prompt-injection signals.
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

// detectInjection returns the matched pattern or "" if clean.
func detectInjection(text string) string {
	lower := strings.ToLower(text)
	for _, p := range injectionPatterns {
		if strings.Contains(lower, p) {
			return p
		}
	}
	return ""
}
