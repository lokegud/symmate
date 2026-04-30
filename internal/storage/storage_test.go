package storage

import (
	"context"
	"testing"
	"time"

	symcrypto "github.com/loke/symmate/internal/crypto"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	key, _, err := symcrypto.DeriveKey([]byte("test-passphrase-loke"), nil)
	if err != nil {
		t.Fatal(err)
	}
	db, err := Open(t.TempDir(), key)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestConversationRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	conv := &Conversation{
		ID:        "01HZ000000000000000TEST01",
		Surface:   "visitor-web",
		StartedAt: time.Now().UnixMilli(),
		Metadata:  map[string]any{"user_agent": "TestAgent/1.0"},
	}
	if err := db.CreateConversation(ctx, conv); err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}

	got, err := db.GetConversation(ctx, conv.ID)
	if err != nil {
		t.Fatalf("GetConversation: %v", err)
	}
	if got.Surface != conv.Surface {
		t.Errorf("surface mismatch: got %q want %q", got.Surface, conv.Surface)
	}
	if ua, ok := got.Metadata["user_agent"].(string); !ok || ua != "TestAgent/1.0" {
		t.Errorf("metadata mismatch: %v", got.Metadata)
	}
}

func TestMessageRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	conv := &Conversation{
		ID:        "01HZ000000000000000TEST02",
		Surface:   "a2a",
		StartedAt: time.Now().UnixMilli(),
	}
	if err := db.CreateConversation(ctx, conv); err != nil {
		t.Fatal(err)
	}

	want := "hello from Loke's Symmate"
	msg := &Message{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        want,
		TS:             time.Now().UnixMilli(),
	}
	if err := db.AppendMessage(ctx, msg); err != nil {
		t.Fatalf("AppendMessage: %v", err)
	}

	msgs, err := db.GetMessages(ctx, conv.ID)
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != want {
		t.Errorf("content mismatch: got %q want %q", msgs[0].Content, want)
	}
}

// TestEncryptedRoundTripAcrossReopen writes a message, closes the DB,
// reopens with the same key, and verifies the plaintext survives.
func TestEncryptedRoundTripAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	passphrase := []byte("loke-test-passphrase")

	key, salt, err := symcrypto.DeriveKey(passphrase, nil)
	if err != nil {
		t.Fatal(err)
	}

	const convID = "01HZ000000000000000REOPEN"
	const wantContent = "secret message that must survive DB close/reopen"

	// Write
	func() {
		db, err := Open(dir, key)
		if err != nil {
			t.Fatalf("Open (write): %v", err)
		}
		defer db.Close()
		ctx := context.Background()
		if err := db.CreateConversation(ctx, &Conversation{
			ID:        convID,
			Surface:   "owner-web",
			StartedAt: time.Now().UnixMilli(),
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.AppendMessage(ctx, &Message{
			ConversationID: convID,
			Role:           "user",
			Content:        wantContent,
			TS:             time.Now().UnixMilli(),
		}); err != nil {
			t.Fatal(err)
		}
	}()

	// Reopen with same derived key
	key2, _, err := symcrypto.DeriveKey(passphrase, salt)
	if err != nil {
		t.Fatal(err)
	}
	db2, err := Open(dir, key2)
	if err != nil {
		t.Fatalf("Open (read): %v", err)
	}
	defer db2.Close()

	msgs, err := db2.GetMessages(context.Background(), convID)
	if err != nil {
		t.Fatalf("GetMessages after reopen: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != wantContent {
		t.Errorf("content mismatch after reopen: got %q", msgs[0].Content)
	}
}

func TestWrongKeyCannotDecrypt(t *testing.T) {
	dir := t.TempDir()

	key1, _, err := symcrypto.DeriveKey([]byte("correct-pass"), nil)
	if err != nil {
		t.Fatal(err)
	}
	key2, _, err := symcrypto.DeriveKey([]byte("wrong-pass"), nil)
	if err != nil {
		t.Fatal(err)
	}

	const convID = "01HZ000000000000000WRONGK"

	func() {
		db, err := Open(dir, key1)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		ctx := context.Background()
		db.CreateConversation(ctx, &Conversation{
			ID: convID, Surface: "visitor-web", StartedAt: time.Now().UnixMilli(),
		})
		db.AppendMessage(ctx, &Message{
			ConversationID: convID, Role: "user", Content: "sensitive", TS: time.Now().UnixMilli(),
		})
	}()

	db2, err := Open(dir, key2)
	if err != nil {
		t.Fatalf("Open with wrong key: %v", err)
	}
	defer db2.Close()

	_, err = db2.GetMessages(context.Background(), convID)
	if err == nil {
		t.Fatal("expected decryption error with wrong key, got nil")
	}
}

func TestFactRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	f := &Fact{
		FactClass:  "preference",
		Content:    "Loke prefers dark mode",
		AuthoredAt: time.Now().UnixMilli(),
		Source:     "manual",
	}
	id, err := db.CreateFact(ctx, f)
	if err != nil {
		t.Fatalf("CreateFact: %v", err)
	}

	got, err := db.GetFact(ctx, id)
	if err != nil {
		t.Fatalf("GetFact: %v", err)
	}
	if got.Content != f.Content {
		t.Errorf("fact content mismatch: got %q want %q", got.Content, f.Content)
	}
}

func TestFactRetract(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	id, err := db.CreateFact(ctx, &Fact{
		FactClass:  "identity",
		Content:    "old fact",
		AuthoredAt: time.Now().UnixMilli(),
		Source:     "manual",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.RetractFact(ctx, id); err != nil {
		t.Fatal(err)
	}

	facts, err := db.ListFacts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range facts {
		if f.ID == id {
			t.Error("retracted fact still appears in ListFacts")
		}
	}
}

func TestSingletonLock(t *testing.T) {
	dir := t.TempDir()
	key, _, err := symcrypto.DeriveKey([]byte("x"), nil)
	if err != nil {
		t.Fatal(err)
	}

	db1, err := Open(dir, key)
	if err != nil {
		t.Fatal(err)
	}
	defer db1.Close()

	_, err = Open(dir, key)
	if err == nil {
		t.Fatal("expected error opening second instance, got nil")
	}
}
