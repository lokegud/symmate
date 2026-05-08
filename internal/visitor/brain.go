package visitor

import "context"

// Message is a single conversation turn.
type Message struct {
	Role    string // "user" | "assistant" | "system"
	Content string
}

// Brain is the personality+LLM interface used by the chat handler.
type Brain interface {
	Chat(ctx context.Context, msgs []Message) (<-chan string, error)
}

// ConversationStore persists and retrieves conversation history.
// nil is a valid value — all operations become no-ops and memory is disabled.
type ConversationStore interface {
	// EnsureConversation creates the conversation row if it doesn't exist yet.
	EnsureConversation(ctx context.Context, id, surface string, handle *string) error
	// RecentMessages returns up to limit messages in chronological order.
	RecentMessages(ctx context.Context, conversationID string, limit int) ([]Message, error)
	// AppendMessage stores a single turn.
	AppendMessage(ctx context.Context, conversationID, role, content string) error
}

// BrainStub satisfies Brain before the real implementation is wired.
type BrainStub struct{}

func (BrainStub) Chat(_ context.Context, _ []Message) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- "Symmate brain not yet wired — surfaces functional, brain stub active."
	close(ch)
	return ch, nil
}
