package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCapabilityGrantLogging(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "audit.jsonl")
	store, err := NewFileStore(tmpFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	logger := NewLogger(store)

	payload := &CapabilityGrantPayload{
		AgentID:      "agent-123",
		CapabilityID: "cap-read-files",
		GrantedBy:    "system-admin",
		TTL:          intPtr(3600),
		Constraints:  []string{"path_prefix:/tmp", "readonly"},
	}

	convID := "conv-456"
	if err := logger.CapabilityGrant(payload, SeverityNotice, &convID); err != nil {
		t.Fatalf("failed to log capability grant: %v", err)
	}

	event := readLastJSONLLine(t, tmpFile)
	if event.Category != "capability_grant" {
		t.Errorf("expected category 'capability_grant', got %q", event.Category)
	}
	if event.Severity != SeverityNotice {
		t.Errorf("expected severity 'notice', got %q", event.Severity)
	}
	if event.ConversationID == nil || *event.ConversationID != convID {
		t.Errorf("conversation ID mismatch")
	}

	payloadBytes, _ := json.Marshal(event.Payload)
	var loggedPayload CapabilityGrantPayload
	json.Unmarshal(payloadBytes, &loggedPayload)
	if loggedPayload.AgentID != "agent-123" {
		t.Errorf("agent ID mismatch: %q", loggedPayload.AgentID)
	}
}

func TestMultipleEventTypes(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "audit.jsonl")
	store, err := NewFileStore(tmpFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	logger := NewLogger(store)

	logger.PolicyChange(&PolicyChangePayload{
		PolicyID:  "pol-123",
		Operation: "update",
		ChangedBy: "admin",
	}, SeverityWarning, nil)

	logger.Escalation(&EscalationPayload{
		AgentID:        "agent-456",
		FromCapability: "read",
		ToCapability:   "write",
		Reason:         "user_requested",
	}, SeverityAlert, nil)

	logger.ToolCall(&ToolCallPayload{
		AgentID:      "agent-789",
		ToolName:     "file_read",
		ResultStatus: "success",
		ExecutionMs:  42,
	}, SeverityInfo, nil)

	if lineCount := countJSONLLines(t, tmpFile); lineCount != 3 {
		t.Errorf("expected 3 lines, got %d", lineCount)
	}
}

func TestFileStoreAtomicAppend(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "audit.jsonl")
	store, err := NewFileStore(tmpFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	logger := NewLogger(store)

	for i := 0; i < 10; i++ {
		logger.ToolCall(&ToolCallPayload{
			AgentID:      "agent-123",
			ToolName:     "test_tool",
			ResultStatus: "success",
			ExecutionMs:  int64(i),
		}, SeverityInfo, nil)
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		var event AuditEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Errorf("invalid JSON on line %d: %v", count+1, err)
		}
		count++
	}
	if count != 10 {
		t.Errorf("expected 10 lines, got %d", count)
	}
}

func TestEventIDUniqueness(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "audit.jsonl")
	store, err := NewFileStore(tmpFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	logger := NewLogger(store)

	for i := 0; i < 5; i++ {
		logger.ToolCall(&ToolCallPayload{
			AgentID:      "agent-123",
			ToolName:     "test",
			ResultStatus: "success",
		}, SeverityInfo, nil)
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	ids := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event AuditEvent
		json.Unmarshal(scanner.Bytes(), &event)
		if ids[event.ID] {
			t.Errorf("duplicate event ID: %s", event.ID)
		}
		ids[event.ID] = true
	}
	if len(ids) != 5 {
		t.Errorf("expected 5 unique IDs, got %d", len(ids))
	}
}

func TestSeverityLevels(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "audit.jsonl")
	store, err := NewFileStore(tmpFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	logger := NewLogger(store)

	for _, sev := range []Severity{SeverityInfo, SeverityNotice, SeverityWarning, SeverityAlert} {
		logger.ToolCall(&ToolCallPayload{
			AgentID:      "agent-123",
			ToolName:     "test",
			ResultStatus: "success",
		}, sev, nil)
	}

	if lineCount := countJSONLLines(t, tmpFile); lineCount != 4 {
		t.Errorf("expected 4 lines, got %d", lineCount)
	}
}

// Helper functions

func readLastJSONLLine(t *testing.T, path string) *AuditEvent {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	var lastEvent *AuditEvent
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event AuditEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		lastEvent = &event
	}

	if lastEvent == nil {
		t.Fatalf("no events found in file")
	}
	return lastEvent
}

func countJSONLLines(t *testing.T, path string) int {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}

func intPtr(v int) *int {
	return &v
}
