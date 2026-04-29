package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Store is the interface for audit log persistence
type Store interface {
	Write(event *AuditEvent) error
	Close() error
}

// FileStore writes audit events to a JSONL file (canonical storage)
type FileStore struct {
	path   string
	file   *os.File
	mu     sync.Mutex
	closed bool
}

// NewFileStore creates a new file-based audit store
// The file is opened in append mode with O_APPEND for atomic writes
func NewFileStore(path string) (*FileStore, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}
	return &FileStore{
		path: path,
		file: file,
	}, nil
}

// Write appends an audit event as a JSON line to the file
func (fs *FileStore) Write(event *AuditEvent) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.closed {
		return fmt.Errorf("audit store is closed")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Append newline to make it a valid JSONL line
	line := append(data, '\n')

	// Write to file; O_APPEND mode ensures atomic append
	if _, err := fs.file.Write(line); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	// Optional: sync to disk for durability (can be made configurable)
	// fs.file.Sync() would go here if strict durability is needed

	return nil
}

// Close closes the audit store
func (fs *FileStore) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.closed {
		return nil
	}

	fs.closed = true
	if fs.file != nil {
		return fs.file.Close()
	}
	return nil
}

// SQLiteStore mirrors audit events to SQLite with encryption
// The payload BLOB is encrypted using Worker B's AEAD encryption
type SQLiteStore struct {
	fileStore     *FileStore
	sqliteWriter  AuditStoreWriter
	mu            sync.Mutex
}

// AuditStoreWriter is the interface for SQLite audit storage.
// Uses primitive types to avoid circular imports with the storage package.
// Encryption of payloadJSON is the responsibility of the implementor.
type AuditStoreWriter interface {
	InsertAuditEvent(eventID string, ts int64, category, severity string, payloadJSON []byte, conversationID *string) error
	Close() error
}

// NewSQLiteStore creates a dual-store: FileStore for canonical JSONL + SQLite for indexed queries.
// sqliteWriter may be nil (JSONL-only mode).
func NewSQLiteStore(jsonlPath string, sqliteWriter AuditStoreWriter) (*SQLiteStore, error) {
	fileStore, err := NewFileStore(jsonlPath)
	if err != nil {
		return nil, err
	}

	return &SQLiteStore{
		fileStore:    fileStore,
		sqliteWriter: sqliteWriter,
	}, nil
}

// Write writes to both the JSONL file (canonical, plaintext) and SQLite (encrypted mirror).
func (ss *SQLiteStore) Write(event *AuditEvent) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// JSONL is canonical — write first.
	if err := ss.fileStore.Write(event); err != nil {
		return err
	}

	if ss.sqliteWriter == nil {
		return nil
	}

	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload for SQLite: %w", err)
	}

	ts := event.Timestamp.UnixMilli()
	if err := ss.sqliteWriter.InsertAuditEvent(event.ID, ts, event.Category, string(event.Severity), payloadBytes, event.ConversationID); err != nil {
		// SQLite mirror failure is non-fatal; JSONL is authoritative.
		fmt.Fprintf(os.Stderr, "warning: SQLite mirror write failed: %v\n", err)
	}

	return nil
}

// Close closes both stores
func (ss *SQLiteStore) Close() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	var errs []error
	if ss.fileStore != nil {
		if err := ss.fileStore.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if ss.sqliteWriter != nil {
		if err := ss.sqliteWriter.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing audit stores: %v", errs)
	}
	return nil
}
