package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Handler is called synchronously for each matching event.
type Handler func(ctx context.Context, e Event)

// Bus is the shared event substrate for all 12 department agents.
type Bus interface {
	// Publish emits an event to all registered handlers for its type.
	Publish(ctx context.Context, e Event) error
	// Subscribe registers h to receive every event of type t.
	Subscribe(t EventType, h Handler)
}

// MemBus is an in-process, synchronous fan-out implementation.
type MemBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
}

// New returns an initialised MemBus.
func New() *MemBus {
	return &MemBus{handlers: make(map[EventType][]Handler)}
}

func (b *MemBus) Subscribe(t EventType, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[t] = append(b.handlers[t], h)
}

func (b *MemBus) Publish(ctx context.Context, e Event) error {
	if e.Type == "" {
		return fmt.Errorf("eventbus: event type required")
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}

	b.mu.RLock()
	hs := make([]Handler, len(b.handlers[e.Type]))
	copy(hs, b.handlers[e.Type])
	b.mu.RUnlock()

	for _, h := range hs {
		h(ctx, e)
	}
	return nil
}
