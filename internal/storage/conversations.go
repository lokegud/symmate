package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Conversation mirrors the conversations table.
type Conversation struct {
	ID            string  `db:"id"`
	VisitorHandle *string `db:"visitor_handle"`
	Surface       string  `db:"surface"`
	StartedAt     int64   `db:"started_at"`
	EndedAt       *int64  `db:"ended_at"`
	Outcome       *string `db:"outcome"`
	// Metadata is plaintext after decryption; stored encrypted.
	Metadata map[string]any
}

// Message mirrors the messages table. Content is plaintext after decryption.
type Message struct {
	ID             int64   `db:"id"`
	ConversationID string  `db:"conversation_id"`
	Role           string  `db:"role"`
	Content        string  // plaintext; never stored directly
	TS             int64   `db:"ts"`
	LLMProvider    *string `db:"llm_provider"`
	TokensIn       *int64  `db:"tokens_in"`
	TokensOut      *int64  `db:"tokens_out"`
	Flags          *string `db:"flags"`
}

// CreateConversation inserts a new conversation row. metadata may be nil.
func (d *DB) CreateConversation(ctx context.Context, c *Conversation) error {
	var metaEnc []byte
	if c.Metadata != nil {
		raw, err := json.Marshal(c.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
		metaEnc, err = d.encrypt(raw)
		if err != nil {
			return fmt.Errorf("encrypt metadata: %w", err)
		}
	}
	_, err := d.sqlx.ExecContext(ctx,
		`INSERT INTO conversations (id, visitor_handle, surface, started_at, metadata)
         VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.VisitorHandle, c.Surface, c.StartedAt, metaEnc,
	)
	return err
}

// CloseConversation sets ended_at and outcome.
func (d *DB) CloseConversation(ctx context.Context, id, outcome string) error {
	_, err := d.sqlx.ExecContext(ctx,
		`UPDATE conversations SET ended_at=?, outcome=? WHERE id=?`,
		time.Now().UnixMilli(), outcome, id,
	)
	return err
}

// GetConversation fetches and decrypts a conversation by ID.
func (d *DB) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	var row struct {
		ID            string  `db:"id"`
		VisitorHandle *string `db:"visitor_handle"`
		Surface       string  `db:"surface"`
		StartedAt     int64   `db:"started_at"`
		EndedAt       *int64  `db:"ended_at"`
		Outcome       *string `db:"outcome"`
		MetadataEnc   []byte  `db:"metadata"`
	}
	if err := d.sqlx.GetContext(ctx, &row,
		`SELECT id, visitor_handle, surface, started_at, ended_at, outcome, metadata
         FROM conversations WHERE id=?`, id); err != nil {
		return nil, err
	}
	c := &Conversation{
		ID:            row.ID,
		VisitorHandle: row.VisitorHandle,
		Surface:       row.Surface,
		StartedAt:     row.StartedAt,
		EndedAt:       row.EndedAt,
		Outcome:       row.Outcome,
	}
	if len(row.MetadataEnc) > 0 {
		raw, err := d.decrypt(row.MetadataEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt metadata: %w", err)
		}
		if err := json.Unmarshal(raw, &c.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}
	return c, nil
}

// DeleteConversation hard-deletes a conversation and its messages (right-to-delete).
func (d *DB) DeleteConversation(ctx context.Context, id string) error {
	return d.WithTx(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM messages WHERE conversation_id=?`, id); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `DELETE FROM conversations WHERE id=?`, id)
		return err
	})
}

// AppendMessage encrypts content and appends a message to a conversation.
func (d *DB) AppendMessage(ctx context.Context, m *Message) error {
	enc, err := d.encrypt([]byte(m.Content))
	if err != nil {
		return fmt.Errorf("encrypt message: %w", err)
	}
	_, err = d.sqlx.ExecContext(ctx,
		`INSERT INTO messages (conversation_id, role, content_enc, ts, llm_provider, tokens_in, tokens_out, flags)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ConversationID, m.Role, enc, m.TS, m.LLMProvider, m.TokensIn, m.TokensOut, m.Flags,
	)
	return err
}

// ListConversations returns up to limit conversations ordered by started_at DESC.
// Metadata is not decrypted — use GetConversation for that.
func (d *DB) ListConversations(ctx context.Context, limit, offset int) ([]*Conversation, error) {
	var rows []struct {
		ID            string  `db:"id"`
		VisitorHandle *string `db:"visitor_handle"`
		Surface       string  `db:"surface"`
		StartedAt     int64   `db:"started_at"`
		EndedAt       *int64  `db:"ended_at"`
		Outcome       *string `db:"outcome"`
	}
	if err := d.sqlx.SelectContext(ctx, &rows,
		`SELECT id, visitor_handle, surface, started_at, ended_at, outcome
         FROM conversations ORDER BY started_at DESC LIMIT ? OFFSET ?`,
		limit, offset); err != nil {
		return nil, err
	}
	out := make([]*Conversation, 0, len(rows))
	for _, r := range rows {
		out = append(out, &Conversation{
			ID:            r.ID,
			VisitorHandle: r.VisitorHandle,
			Surface:       r.Surface,
			StartedAt:     r.StartedAt,
			EndedAt:       r.EndedAt,
			Outcome:       r.Outcome,
		})
	}
	return out, nil
}

// GetMessages fetches and decrypts all messages for a conversation, ordered by ts.
func (d *DB) GetMessages(ctx context.Context, conversationID string) ([]*Message, error) {
	var rows []struct {
		ID             int64   `db:"id"`
		ConversationID string  `db:"conversation_id"`
		Role           string  `db:"role"`
		ContentEnc     []byte  `db:"content_enc"`
		TS             int64   `db:"ts"`
		LLMProvider    *string `db:"llm_provider"`
		TokensIn       *int64  `db:"tokens_in"`
		TokensOut      *int64  `db:"tokens_out"`
		Flags          *string `db:"flags"`
	}
	if err := d.sqlx.SelectContext(ctx, &rows,
		`SELECT id, conversation_id, role, content_enc, ts, llm_provider, tokens_in, tokens_out, flags
         FROM messages WHERE conversation_id=? ORDER BY ts`, conversationID); err != nil {
		return nil, err
	}
	msgs := make([]*Message, 0, len(rows))
	for _, r := range rows {
		pt, err := d.decrypt(r.ContentEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt message %d: %w", r.ID, err)
		}
		msgs = append(msgs, &Message{
			ID:             r.ID,
			ConversationID: r.ConversationID,
			Role:           r.Role,
			Content:        string(pt),
			TS:             r.TS,
			LLMProvider:    r.LLMProvider,
			TokensIn:       r.TokensIn,
			TokensOut:      r.TokensOut,
			Flags:          r.Flags,
		})
	}
	return msgs, nil
}
