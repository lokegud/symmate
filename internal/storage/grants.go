package storage

import (
	"context"
	"time"
)

// Grant mirrors the capability_grants table.
type Grant struct {
	ID         int64   `db:"id"`
	Scope      string  `db:"scope"`
	DataClass  string  `db:"data_class"`
	Permission string  `db:"permission"` // "allow", "deny", "ask"
	GrantedAt  int64   `db:"granted_at"`
	GrantedBy  string  `db:"granted_by"`
	ExpiresAt  *int64  `db:"expires_at"`
}

// ListGrants returns all non-expired grants, optionally filtered by scope.
// Pass scope="" to get all grants.
func (d *DB) ListGrants(ctx context.Context, scope string) ([]*Grant, error) {
	now := time.Now().UnixMilli()
	var rows []*Grant
	var err error
	if scope == "" {
		err = d.sqlx.SelectContext(ctx, &rows,
			`SELECT id, scope, data_class, permission, granted_at, granted_by, expires_at
             FROM capability_grants
             WHERE expires_at IS NULL OR expires_at > ?
             ORDER BY scope, data_class`, now)
	} else {
		err = d.sqlx.SelectContext(ctx, &rows,
			`SELECT id, scope, data_class, permission, granted_at, granted_by, expires_at
             FROM capability_grants
             WHERE scope=? AND (expires_at IS NULL OR expires_at > ?)
             ORDER BY data_class`, scope, now)
	}
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// UpsertGrant inserts a grant or updates it if (scope, data_class) already exists.
func (d *DB) UpsertGrant(ctx context.Context, g *Grant) (int64, error) {
	if g.GrantedAt == 0 {
		g.GrantedAt = time.Now().UnixMilli()
	}
	var existing Grant
	err := d.sqlx.GetContext(ctx, &existing,
		`SELECT id FROM capability_grants WHERE scope=? AND data_class=?`,
		g.Scope, g.DataClass)
	if err == nil {
		// Update existing.
		_, err = d.sqlx.ExecContext(ctx,
			`UPDATE capability_grants SET permission=?, granted_at=?, granted_by=?, expires_at=? WHERE id=?`,
			g.Permission, g.GrantedAt, g.GrantedBy, g.ExpiresAt, existing.ID)
		return existing.ID, err
	}
	// Insert new.
	res, err := d.sqlx.ExecContext(ctx,
		`INSERT INTO capability_grants (scope, data_class, permission, granted_at, granted_by, expires_at)
         VALUES (?, ?, ?, ?, ?, ?)`,
		g.Scope, g.DataClass, g.Permission, g.GrantedAt, g.GrantedBy, g.ExpiresAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// DeleteGrant hard-deletes a grant by ID.
func (d *DB) DeleteGrant(ctx context.Context, id int64) error {
	_, err := d.sqlx.ExecContext(ctx,
		`DELETE FROM capability_grants WHERE id=?`, id)
	return err
}
