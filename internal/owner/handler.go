// Package owner implements the owner management API (port 7841).
// Bound to localhost — the SvelteKit dashboard server-side is the only consumer.
// No external auth is required since only local processes can reach this port.
package owner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/loke/symmate/internal/audit"
	"github.com/loke/symmate/internal/storage"
)

// Handler is the owner API HTTP handler.
type Handler struct {
	hostDir   string
	db        *storage.DB
	auditPath string
	// paused controls the kill-switch. When true, the visitor API returns a static offline message.
	paused bool
	mu     sync.RWMutex
	// onGrantsChange is called after any grant CRUD so Brain can reload its live grant list.
	onGrantsChange func()
}

// New creates a new owner API handler.
// onGrantsChange may be nil.
func New(hostDir string, db *storage.DB, onGrantsChange func()) *Handler {
	return &Handler{
		hostDir:        hostDir,
		db:             db,
		auditPath:      filepath.Join(hostDir, "audit.jsonl"),
		onGrantsChange: onGrantsChange,
	}
}

// Router returns the mux for the owner API.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.handleHealth)

	mux.HandleFunc("GET /personality", h.handlePersonalityGet)
	mux.HandleFunc("PUT /personality/{name}", h.handlePersonalityPut)

	mux.HandleFunc("GET /policy/grants", h.handleGrantsList)
	mux.HandleFunc("POST /policy/grants", h.handleGrantsCreate)
	mux.HandleFunc("DELETE /policy/grants/{id}", h.handleGrantsDelete)

	mux.HandleFunc("GET /audit/events", h.handleAuditEvents)
	mux.HandleFunc("GET /audit/stream", h.handleAuditStream)

	mux.HandleFunc("GET /conversations", h.handleConversationsList)
	mux.HandleFunc("GET /conversations/{id}", h.handleConversationsGet)
	mux.HandleFunc("DELETE /conversations/{id}", h.handleConversationsDelete)

	mux.HandleFunc("POST /control/pause", h.handlePause)
	mux.HandleFunc("POST /control/unpause", h.handleUnpause)
	mux.HandleFunc("GET /control/status", h.handleStatus)

	return mux
}

// IsPaused is safe for concurrent use. Visitor API should call this before processing.
func (h *Handler) IsPaused() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.paused
}

// --- personality ---

var allowedPersonalityFiles = map[string]bool{
	"context.md":     true,
	"personality.md": true,
	"values.md":      true,
	"excerpts.md":    true,
}

func (h *Handler) handlePersonalityGet(w http.ResponseWriter, r *http.Request) {
	files := map[string]string{}
	for name := range allowedPersonalityFiles {
		path := filepath.Join(h.hostDir, name)
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				files[name] = ""
				continue
			}
			http.Error(w, fmt.Sprintf("read %s: %v", name, err), http.StatusInternalServerError)
			return
		}
		files[name] = string(b)
	}
	writeJSON(w, files)
}

func (h *Handler) handlePersonalityPut(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !allowedPersonalityFiles[name] {
		http.Error(w, "unknown personality file", http.StatusBadRequest)
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	path := filepath.Join(h.hostDir, name)
	if err := os.WriteFile(path, []byte(body.Content), 0600); err != nil {
		http.Error(w, fmt.Sprintf("write %s: %v", name, err), http.StatusInternalServerError)
		return
	}
	// personality/Loader watches via fsnotify — no explicit reload needed.
	w.WriteHeader(http.StatusNoContent)
}

// --- grants ---

func (h *Handler) handleGrantsList(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	grants, err := h.db.ListGrants(r.Context(), scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, grants)
}

func (h *Handler) handleGrantsCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Scope      string `json:"scope"`
		DataClass  string `json:"data_class"`
		Permission string `json:"permission"`
		ExpiresAt  *int64 `json:"expires_at,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Scope == "" || body.DataClass == "" {
		http.Error(w, "scope and data_class are required", http.StatusBadRequest)
		return
	}
	switch body.Permission {
	case "allow", "deny", "ask":
	default:
		http.Error(w, "permission must be allow, deny, or ask", http.StatusBadRequest)
		return
	}
	g := &storage.Grant{
		Scope:      body.Scope,
		DataClass:  body.DataClass,
		Permission: body.Permission,
		GrantedAt:  time.Now().UnixMilli(),
		GrantedBy:  "owner",
		ExpiresAt:  body.ExpiresAt,
	}
	id, err := h.db.UpsertGrant(r.Context(), g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	g.ID = id
	if h.onGrantsChange != nil {
		h.onGrantsChange()
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, g)
}

func (h *Handler) handleGrantsDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.db.DeleteGrant(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.onGrantsChange != nil {
		h.onGrantsChange()
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- audit ---

func (h *Handler) handleAuditEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := audit.QueryFilter{
		Severity: q.Get("severity"),
		Category: q.Get("category"),
	}
	if s := q.Get("since"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			f.Since = t
		}
	}
	if l := q.Get("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err == nil && n > 0 {
			f.Limit = n
		}
	}
	events, err := audit.QueryJSONL(h.auditPath, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if events == nil {
		events = []*audit.AuditEvent{}
	}
	writeJSON(w, events)
}

// handleAuditStream tails audit.jsonl as an SSE stream.
// It sends existing events (up to 50) on connect, then polls every 500ms for new lines.
func (h *Handler) handleAuditStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Seed with last 50 events.
	recent, _ := audit.QueryJSONL(h.auditPath, audit.QueryFilter{Limit: 50})
	for i := len(recent) - 1; i >= 0; i-- {
		b, _ := json.Marshal(recent[i])
		fmt.Fprintf(w, "data: %s\n\n", b)
	}
	flusher.Flush()

	// Track file offset for tailing.
	fi, err := os.Open(h.auditPath)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	var offset int64
	if err == nil {
		offset, _ = fi.Seek(0, 2) // seek to end
		fi.Close()
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fi, err := os.Open(h.auditPath)
			if err != nil {
				continue
			}
			fi.Seek(offset, 0)
			scanner := newLineScanner(fi)
			for scanner.Scan() {
				line := scanner.Bytes()
				if len(strings.TrimSpace(string(line))) == 0 {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", line)
			}
			offset, _ = fi.Seek(0, 1) // current position after scan
			fi.Close()
			flusher.Flush()
		}
	}
}

// --- conversations ---

func (h *Handler) handleConversationsList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	convs, err := h.db.ListConversations(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if convs == nil {
		convs = []*storage.Conversation{}
	}
	writeJSON(w, convs)
}

func (h *Handler) handleConversationsGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	conv, err := h.db.GetConversation(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	msgs, err := h.db.GetMessages(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"conversation": conv,
		"messages":     msgs,
	})
}

func (h *Handler) handleConversationsDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.db.DeleteConversation(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- control ---

func (h *Handler) handlePause(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	h.paused = true
	h.mu.Unlock()
	writeJSON(w, map[string]bool{"paused": true})
}

func (h *Handler) handleUnpause(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	h.paused = false
	h.mu.Unlock()
	writeJSON(w, map[string]bool{"paused": false})
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	paused := h.paused
	h.mu.RUnlock()
	writeJSON(w, map[string]bool{"paused": paused})
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "ok\n")
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func newLineScanner(f *os.File) *bufio.Scanner {
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 1<<20), 1<<20)
	return s
}
