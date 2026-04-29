package storage

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	symcrypto "github.com/loke/symmate/internal/crypto"
)

//go:embed migrations/*.sql
var migrations embed.FS

// DB is the per-host storage handle.
type DB struct {
	sqlx     *sqlx.DB
	key      []byte // AEAD key; in memory only, never written to disk
	lockFile *os.File
}

// Open acquires the singleton lock, runs migrations, and returns a DB.
// hostDir is the per-host data directory (e.g. /var/lib/symmate/loke/).
// key is the 32-byte AEAD master key (caller derives via crypto.DeriveKey).
func Open(hostDir string, key []byte) (*DB, error) {
	if err := os.MkdirAll(hostDir, 0700); err != nil {
		return nil, fmt.Errorf("mkdir hostDir: %w", err)
	}

	lf, err := acquireLock(hostDir)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("file:%s/symmate.db?_journal_mode=WAL&_foreign_keys=on", hostDir)
	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		lf.Close()
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite WAL; single writer

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		db.Close()
		lf.Close()
		return nil, fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(db.DB, "migrations"); err != nil {
		db.Close()
		lf.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return &DB{sqlx: db, key: key, lockFile: lf}, nil
}

// Close releases the DB and the singleton lock.
func (d *DB) Close() error {
	err := d.sqlx.Close()
	d.lockFile.Close()
	return err
}

// WithTx runs fn inside a transaction, rolling back on error.
func (d *DB) WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := d.sqlx.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// encrypt is a helper that AEAD-encrypts plaintext JSON using the DB key.
func (d *DB) encrypt(plaintext []byte) ([]byte, error) {
	return symcrypto.Encrypt(d.key, plaintext, nil)
}

// decrypt is a helper that AEAD-decrypts a stored blob using the DB key.
func (d *DB) decrypt(ciphertext []byte) ([]byte, error) {
	return symcrypto.Decrypt(d.key, ciphertext, nil)
}

// InsertAuditEvent stores an encrypted audit event blob in the SQLite mirror.
// Implements audit.AuditStoreWriter; payloadJSON is encrypted before storage.
func (d *DB) InsertAuditEvent(eventID string, ts int64, category, severity string, payloadJSON []byte, conversationID *string) error {
	enc, err := d.encrypt(payloadJSON)
	if err != nil {
		return fmt.Errorf("encrypt audit payload: %w", err)
	}
	_, err = d.sqlx.Exec(
		`INSERT INTO audit (event_id, ts, category, severity, payload, conversation_id)
         VALUES (?, ?, ?, ?, ?, ?)`,
		eventID, ts, category, severity, enc, conversationID,
	)
	return err
}
