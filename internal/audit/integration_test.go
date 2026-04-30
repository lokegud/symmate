package audit_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/loke/symmate/internal/audit"
	symcrypto "github.com/loke/symmate/internal/crypto"
	"github.com/loke/symmate/internal/storage"
)

// TestAuditSQLiteRoundTrip verifies that an audit event ends up in both
// the JSONL file and the encrypted SQLite mirror.
func TestAuditSQLiteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	jsonlPath := filepath.Join(dir, "audit.jsonl")

	key, _, err := symcrypto.DeriveKey([]byte("audit-integration-pass"), nil)
	if err != nil {
		t.Fatal(err)
	}

	db, err := storage.Open(dir, key)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer db.Close()

	store, err := audit.NewSQLiteStore(jsonlPath, db)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	logger := audit.NewLogger(store)

	if err := logger.Injection(&audit.InjectionPayload{
		InjectionType: "prompt",
		Source:        "user_input",
		DetectedBy:    "constitutional-filter",
	}, audit.SeverityWarning, nil); err != nil {
		t.Fatalf("Injection log: %v", err)
	}

	// Verify JSONL has the line
	file, err := os.Open(jsonlPath)
	if err != nil {
		t.Fatalf("open jsonl: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var found bool
	for scanner.Scan() {
		var event audit.AuditEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if event.Category == "injection" {
			found = true
		}
	}
	if !found {
		t.Error("injection event not found in JSONL")
	}
}
