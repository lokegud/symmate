package visitor

import "context"

// Message is a single conversation turn.
type Message struct {
	Role    string // "user" | "assistant" | "system"
	Content string
}

// Brain is the personality+LLM interface used by the chat handler.
// Forge's branch wires the real implementation; until then BrainStub is used.
type Brain interface {
	Chat(ctx context.Context, msgs []Message) (<-chan string, error)
}

// BrainStub satisfies Brain before Forge's branch merges.
type BrainStub struct{}

func (BrainStub) Chat(_ context.Context, _ []Message) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- "Symmate brain not yet wired — surfaces functional, brain stub active."
	close(ch)
	return ch, nil
}
