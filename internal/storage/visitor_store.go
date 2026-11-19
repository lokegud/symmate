package storage

import (
	"context"
	"time"

	"github.com/loke/symmate/internal/visitor"
)

// VisitorStore adapts *DB to the visitor.ConversationStore interface.
type VisitorStore struct{ db *DB }

// NewVisitorStore wraps a DB as a visitor.ConversationStore.
func NewVisitorStore(db *DB) *VisitorStore { return &VisitorStore{db} }

// EnsureConversation creates the conversation row if it doesn't exist.
func (s *VisitorStore) EnsureConversation(ctx context.Context, id, surface string, handle *string) error {
	_, err := s.db.sqlx.ExecContext(ctx,
		`INSERT OR IGNORE INTO conversations (id, visitor_handle, surface, started_at)
         VALUES (?, ?, ?, ?)`,
		id, handle, surface, time.Now().UnixMilli(),
	)
	return err
}

// RecentMessages returns the most recent limit messages in chronological order.
func (s *VisitorStore) RecentMessages(ctx context.Context, conversationID string, limit int) ([]visitor.Message, error) {
	var rows []struct {
		Role       string `db:"role"`
		ContentEnc []byte `db:"content_enc"`
	}
	// Fetch newest-first, then reverse so caller gets chronological order.
	if err := s.db.sqlx.SelectContext(ctx, &rows,
		`SELECT role, content_enc FROM messages
         WHERE conversation_id=? ORDER BY ts DESC LIMIT ?`,
		conversationID, limit,
	); err != nil {
		return nil, err
	}

	msgs := make([]visitor.Message, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		pt, err := s.db.decrypt(rows[i].ContentEnc)
		if err != nil {
			continue // skip unreadable rows rather than failing the whole request
		}
		msgs = append(msgs, visitor.Message{Role: rows[i].Role, Content: string(pt)})
	}
	return msgs, nil
}

// AppendMessage encrypts and stores a single turn.
func (s *VisitorStore) AppendMessage(ctx context.Context, conversationID, role, content string) error {
	enc, err := s.db.encrypt([]byte(content))
	if err != nil {
		return err
	}
	_, err = s.db.sqlx.ExecContext(ctx,
		`INSERT INTO messages (conversation_id, role, content_enc, ts) VALUES (?, ?, ?, ?)`,
		conversationID, role, enc, time.Now().UnixMilli(),
	)
	return err
}

// compile-time interface check
var _ visitor.ConversationStore = (*VisitorStore)(nil)
