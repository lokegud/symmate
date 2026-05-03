package main

import (
	"context"
	"sync"

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
	grants   []firewall.Grant
	mu       sync.RWMutex
}

// SetGrants replaces the live grant list. Called by the owner API after CRUD.
func (b *Brain) SetGrants(grants []firewall.Grant) {
	b.mu.Lock()
	b.grants = grants
	b.mu.Unlock()
}

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

		b.mu.RLock()
		grants := b.grants
		b.mu.RUnlock()
		filtered, _ := b.firewall.Filter(ctx, reply, grants)
		ch <- filtered
	}()

	return ch, nil
}

func inferTag(msgs []visitor.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			return llm.Classify(msgs[i].Content)
		}
	}
	return llm.TagDefault
}
