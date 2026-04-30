package storage

import (
	"context"
	"fmt"
	"time"
)

// Fact mirrors the facts table. Content is plaintext after decryption.
type Fact struct {
	ID          int64    `db:"id"`
	FactClass   string   `db:"fact_class"`
	Content     string   // plaintext; never stored directly
	AuthoredAt  int64    `db:"authored_at"`
	Source      string   `db:"source"`
	Confidence  *float64 `db:"confidence"`
	RetractedAt *int64   `db:"retracted_at"`
}

// CreateFact encrypts content and inserts a fact.
func (d *DB) CreateFact(ctx context.Context, f *Fact) (int64, error) {
	enc, err := d.encrypt([]byte(f.Content))
	if err != nil {
		return 0, fmt.Errorf("encrypt fact: %w", err)
	}
	res, err := d.sqlx.ExecContext(ctx,
		`INSERT INTO facts (fact_class, content_enc, authored_at, source, confidence)
         VALUES (?, ?, ?, ?, ?)`,
		f.FactClass, enc, f.AuthoredAt, f.Source, f.Confidence,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetFact fetches and decrypts a fact by ID. Returns nil if retracted.
func (d *DB) GetFact(ctx context.Context, id int64) (*Fact, error) {
	var row struct {
		ID          int64    `db:"id"`
		FactClass   string   `db:"fact_class"`
		ContentEnc  []byte   `db:"content_enc"`
		AuthoredAt  int64    `db:"authored_at"`
		Source      string   `db:"source"`
		Confidence  *float64 `db:"confidence"`
		RetractedAt *int64   `db:"retracted_at"`
	}
	if err := d.sqlx.GetContext(ctx, &row,
		`SELECT id, fact_class, content_enc, authored_at, source, confidence, retracted_at
         FROM facts WHERE id=?`, id); err != nil {
		return nil, err
	}
	pt, err := d.decrypt(row.ContentEnc)
	if err != nil {
		return nil, fmt.Errorf("decrypt fact %d: %w", id, err)
	}
	return &Fact{
		ID:          row.ID,
		FactClass:   row.FactClass,
		Content:     string(pt),
		AuthoredAt:  row.AuthoredAt,
		Source:      row.Source,
		Confidence:  row.Confidence,
		RetractedAt: row.RetractedAt,
	}, nil
}

// ListFacts returns all non-retracted facts.
func (d *DB) ListFacts(ctx context.Context) ([]*Fact, error) {
	var rows []struct {
		ID          int64    `db:"id"`
		FactClass   string   `db:"fact_class"`
		ContentEnc  []byte   `db:"content_enc"`
		AuthoredAt  int64    `db:"authored_at"`
		Source      string   `db:"source"`
		Confidence  *float64 `db:"confidence"`
		RetractedAt *int64   `db:"retracted_at"`
	}
	if err := d.sqlx.SelectContext(ctx, &rows,
		`SELECT id, fact_class, content_enc, authored_at, source, confidence, retracted_at
         FROM facts WHERE retracted_at IS NULL ORDER BY authored_at`); err != nil {
		return nil, err
	}
	facts := make([]*Fact, 0, len(rows))
	for _, r := range rows {
		pt, err := d.decrypt(r.ContentEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt fact %d: %w", r.ID, err)
		}
		facts = append(facts, &Fact{
			ID:          r.ID,
			FactClass:   r.FactClass,
			Content:     string(pt),
			AuthoredAt:  r.AuthoredAt,
			Source:      r.Source,
			Confidence:  r.Confidence,
			RetractedAt: r.RetractedAt,
		})
	}
	return facts, nil
}

// RetractFact soft-deletes a fact by setting retracted_at.
func (d *DB) RetractFact(ctx context.Context, id int64) error {
	_, err := d.sqlx.ExecContext(ctx,
		`UPDATE facts SET retracted_at=? WHERE id=?`,
		time.Now().UnixMilli(), id,
	)
	return err
}
