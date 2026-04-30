package main

import (
	"context"

	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/symmate/internal/llm"
	"github.com/loke/symmate/internal/personality"
	"github.com/loke/symmate/internal/visitor"
)

// Brain implements visitor.Brain by composing the personality engine,
// LLM router, and privacy firewall.
type Brain struct {
	loader   *personality.Loader
	router   *llm.Router
	firewall *firewall.Firewall
	// grants is the active capability policy for outbound responses.
	// Empty = deny-all PII until owner configures grants via Owner API.
	grants []firewall.Grant
}

// Chat satisfies visitor.Brain. It assembles the system prompt, routes the
// conversation through the LLM, and applies the privacy firewall before
// streaming the reply token-by-token.
func (b *Brain) Chat(ctx context.Context, msgs []visitor.Message) (<-chan string, error) {
	sysPrompt := b.loader.Current()
	tag := inferTag(msgs)

	llmMsgs := make([]llm.Message, 0, len(msgs)+1)
	if sysPrompt != "" {
		llmMsgs = append(llmMsgs, llm.Message{Role: "system", Content: sysPrompt})
	}
	for _, m := range msgs {
		llmMsgs = append(llmMsgs, llm.Message{Role: m.Role, Content: m.Content})
	}

	ch := make(chan string, 1)
	go func() {
		defer close(ch)

		reply, err := b.router.Route(ctx, tag, llmMsgs)
		if err != nil {
			ch <- "(error: " + err.Error() + ")"
			return
		}

		filtered, _ := b.firewall.Filter(ctx, reply, b.grants)
		ch <- filtered
	}()

	return ch, nil
}

// inferTag picks the routing tag from the last user message content.
// Falls back to llm.Classify for keyword-based inference.
func inferTag(msgs []visitor.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			return llm.Classify(msgs[i].Content)
		}
	}
	return llm.TagDefault
}
