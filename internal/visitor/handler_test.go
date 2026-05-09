package visitor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
