package visitor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/loke/symmate/internal/aeo"
)

func testHostContext() *aeo.HostContext {
	return &aeo.HostContext{
		Profile: &aeo.Profile{
			Name:         "TestHost",
			Slug:         "test",
			HostType:     aeo.HostTypeMoxPer,
			Description:  "A test Symmate.",
			Capabilities: []string{"chat", "profile"},
			Contact: aeo.Contact{
				Location: "Testville",
				Website:  "https://example.com",
			},
			Endpoints: aeo.Endpoints{
				Visitor: "http://localhost:7842",
				MCP:     "http://localhost:7844",
			},
			SchemaVer: "0.1",
			UpdatedAt: time.Now(),
		},
		Hours:    "Mon-Fri 9am-5pm",
		Catalog:  "Test Widget",
		Policies: "No returns.",
	}
}

func TestIDHandler(t *testing.T) {
	hc := testHostContext()
	router := NewRouter(hc, BrainStub{}, NewSessions(0, 0), nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("content-type = %q, want application/json", ct)
	}

	var p aeo.Profile
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Name != "TestHost" {
		t.Errorf("name = %q, want TestHost", p.Name)
	}
	if p.Endpoints.Visitor == "" {
		t.Error("visitor endpoint should be set")
	}
}

func TestAgentCardHandler(t *testing.T) {
	hc := testHostContext()
	router := NewRouter(hc, BrainStub{}, NewSessions(0, 0), nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-card.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var card AgentCard
	if err := json.Unmarshal(w.Body.Bytes(), &card); err != nil {
		t.Fatalf("unmarshal agent card: %v", err)
	}
	if card.Name != "TestHost" {
		t.Errorf("name = %q, want TestHost", card.Name)
	}
	if !card.Capabilities.Streaming {
		t.Error("streaming should be true")
	}
	if len(card.Skills) == 0 {
		t.Error("skills should be non-empty")
	}
}

func TestLLMsTxtHandler(t *testing.T) {
	hc := testHostContext()
	router := NewRouter(hc, BrainStub{}, NewSessions(0, 0), nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/llms.txt", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "# TestHost") {
		t.Error("llms.txt should contain # TestHost heading")
	}
	if !strings.Contains(body, "Mon-Fri") {
		t.Error("llms.txt should contain hours")
	}
}

func TestChatHandler_Streams(t *testing.T) {
	hc := testHostContext()
	router := NewRouter(hc, BrainStub{}, NewSessions(0, 0), nil, nil, nil)

	body := strings.NewReader(`{"message":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/chat", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Errorf("content-type = %q, want text/event-stream", ct)
	}

	// Must include at least one data: line and the [DONE] sentinel.
	var gotData, gotDone bool
	scanner := bufio.NewScanner(strings.NewReader(w.Body.String()))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			payload := line[6:]
			if payload == "[DONE]" {
				gotDone = true
			} else {
				gotData = true
			}
		}
	}
	if !gotData {
		t.Error("expected at least one data event")
	}
	if !gotDone {
		t.Error("expected [DONE] sentinel")
	}
}

func TestChatHandler_RateLimit(t *testing.T) {
	hc := testHostContext()
	// 1 req/min → second request should be rate-limited.
	sessions := NewSessions(1, 0)
	router := NewRouter(hc, BrainStub{}, sessions, nil, nil, nil)

	sendChat := func() int {
		body := strings.NewReader(`{"message":"hello"}`)
		req := httptest.NewRequest(http.MethodPost, "/chat", body)
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "10.0.0.1:9999"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	if code := sendChat(); code != http.StatusOK {
		t.Fatalf("first request: status = %d, want 200", code)
	}
	if code := sendChat(); code != http.StatusTooManyRequests {
		t.Fatalf("second request (over limit): status = %d, want 429", code)
	}
}

func TestChatHandler_EmptyMessage(t *testing.T) {
	hc := testHostContext()
	router := NewRouter(hc, BrainStub{}, NewSessions(0, 0), nil, nil, nil)

	body := strings.NewReader(`{"message":""}`)
	req := httptest.NewRequest(http.MethodPost, "/chat", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// errorBrain always returns an error from Chat.
type errorBrain struct{}

func (errorBrain) Chat(_ context.Context, _ []Message) (<-chan string, error) {
	return nil, fmt.Errorf("brain failed")
}

// recordingBrain records each call's message slice.
type recordingBrain struct {
	mu    sync.Mutex
	calls [][]Message
}

func (b *recordingBrain) Chat(_ context.Context, msgs []Message) (<-chan string, error) {
	b.mu.Lock()
	cp := make([]Message, len(msgs))
	copy(cp, msgs)
	b.calls = append(b.calls, cp)
	b.mu.Unlock()
	ch := make(chan string, 1)
	ch <- "ok"
	close(ch)
	return ch, nil
}

// memStore is an in-memory ConversationStore for tests.
type memStore struct {
	mu   sync.Mutex
	msgs map[string][]Message
}

func newMemStore() *memStore {
	return &memStore{msgs: make(map[string][]Message)}
}

func (m *memStore) EnsureConversation(_ context.Context, id, _ string, _ *string) error {
	m.mu.Lock()
	if _, ok := m.msgs[id]; !ok {
		m.msgs[id] = nil
	}
	m.mu.Unlock()
	return nil
}

func (m *memStore) RecentMessages(_ context.Context, convID string, limit int) ([]Message, error) {
	m.mu.Lock()
	all := m.msgs[convID]
	m.mu.Unlock()
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	out := make([]Message, len(all))
	copy(out, all)
	return out, nil
}

func (m *memStore) AppendMessage(_ context.Context, convID, role, content string) error {
	m.mu.Lock()
	m.msgs[convID] = append(m.msgs[convID], Message{Role: role, Content: content})
	m.mu.Unlock()
	return nil
}

func TestChatHandler_Memory(t *testing.T) {
	hc := testHostContext()
	brain := &recordingBrain{}
	store := newMemStore()
	router := NewRouter(hc, brain, NewSessions(0, 0), nil, store, nil)

	send := func(msg string) {
		body := strings.NewReader(`{"message":` + `"` + msg + `"` + `}`)
		req := httptest.NewRequest(http.MethodPost, "/chat", body)
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "10.0.0.1:9999"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("send %q: status = %d, want 200", msg, w.Code)
		}
	}

	send("hello, my name is alice")
	send("what is my name?")

	brain.mu.Lock()
	defer brain.mu.Unlock()

	if len(brain.calls) < 2 {
		t.Fatalf("expected 2 brain calls, got %d", len(brain.calls))
	}
	// Second call must include prior turns: user→"hello", assistant→"ok", user→"what"
	second := brain.calls[1]
	if len(second) < 3 {
		t.Fatalf("second brain call: want ≥3 messages (history+current), got %d: %v", len(second), second)
	}
	if !strings.Contains(second[0].Content, "alice") {
		t.Errorf("second call history[0] should contain 'alice', got %q", second[0].Content)
	}
	if second[1].Role != "assistant" {
		t.Errorf("second call history[1] should be assistant turn, got role %q", second[1].Role)
	}
}
