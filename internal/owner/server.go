// Package owner provides the authenticated owner HTTP API on :7841.
// All endpoints require Authorization: Bearer <session-token>.
// Tokens are minted by POST /auth/login and stored in memory (TTL 24h).
package owner

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/loke/symmate/internal/llm"
	"github.com/loke/symmate/internal/storage"
	"github.com/loke/symmate/internal/visitor"
	"github.com/rs/zerolog/log"
)

const tokenTTL = 24 * time.Hour

var validProviders = map[string]bool{
	"anthropic":     true,
	"hermes":        true,
	"openai-compat": true,
}

// Server is the owner API HTTP handler.
type Server struct {
	hostDir       string
	ownerMgr      *visitor.OwnerManager
	db            *storage.DB
	llmStore      *llm.OverrideStore
	allowedOrigin string
	toolGroups    []string // restricted tool groups that can be granted to users

	mu       sync.Mutex
	sessions map[string]time.Time // token → expiry
}

// New creates the owner API server.
func New(hostDir string, ownerMgr *visitor.OwnerManager, db *storage.DB, llmStore *llm.OverrideStore) *Server {
	s := &Server{
		hostDir:       hostDir,
		ownerMgr:      ownerMgr,
		db:            db,
		llmStore:      llmStore,
		allowedOrigin: os.Getenv("OWNER_CORS_ORIGIN"),
		sessions:      make(map[string]time.Time),
	}
	go func() {
		t := time.NewTicker(time.Hour)
		for range t.C {
			s.mu.Lock()
			s.gc()
			s.mu.Unlock()
		}
	}()
	return s
}

// WithToolGroups sets the restricted tool groups the owner can toggle per user.
func (s *Server) WithToolGroups(groups []string) *Server {
	s.toolGroups = append([]string(nil), groups...)
	return s
}

// Router returns the chi router for the owner API.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(s.cors)

	// Auth — no token required.
	r.Post("/auth/login", s.login)
	r.Post("/auth/logout", s.logout)

	// All other routes require a valid session token.
	r.Group(func(r chi.Router) {
		r.Use(s.requireToken)

		r.Get("/summary", s.summary)

		r.Get("/personality", s.getPersonality)
		r.Put("/personality", s.putPersonality)

		r.Get("/policy", s.getPolicy)
		r.Put("/policy", s.putPolicy)

		r.Get("/audit", s.getAudit)

		r.Get("/conversations", s.listConversations)
		r.Get("/conversations/{id}", s.getConversation)

		r.Get("/llm", s.getLLM)
		r.Put("/llm", s.putLLM)

		r.Get("/users", s.listUsers)
		r.Put("/users/settings", s.putUserSettings)
		r.Post("/users", s.createUser)
		r.Patch("/users/{id}", s.patchUser)
		r.Delete("/users/{id}", s.deleteUser)
		r.Post("/users/{id}/token", s.rotateUserToken)
	})

	return r
}

// --- Auth ---

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phrase string `json:"phrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid json", http.StatusBadRequest)
		return
	}
	cfg, err := s.ownerMgr.Load()
	if err != nil || !cfg.Enrolled {
		jsonErr(w, "owner not enrolled", http.StatusUnauthorized)
		return
	}
	if !cfg.CheckPhrase(req.Phrase) {
		jsonErr(w, "invalid phrase", http.StatusUnauthorized)
		return
	}
	token := mintToken()
	s.mu.Lock()
	s.gc()
	s.sessions[token] = time.Now().Add(tokenTTL)
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token != "" {
		s.mu.Lock()
		delete(s.sessions, token)
		s.mu.Unlock()
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Summary ---

func (s *Server) summary(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.ownerMgr.Load()
	ctx := r.Context()
	convs, _ := s.db.ListConversations(ctx, 1000, 0)
	convCount := len(convs)

	name := filepath.Base(s.hostDir)
	if data, err := os.ReadFile(filepath.Join(s.hostDir, "context.md")); err == nil {
		for _, line := range strings.SplitN(string(data), "\n", 5) {
			if strings.HasPrefix(line, "# ") {
				name = strings.TrimPrefix(line, "# ")
				break
			}
		}
	}

	devices := 0
	if cfg != nil {
		devices = len(cfg.Devices)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"name":               name,
		"enrolled":           cfg != nil && cfg.Enrolled,
		"conversation_count": convCount,
		"devices":            devices,
	})
}

// --- Personality ---

type personalityResp struct {
	Context  string `json:"context"`
	Vertical string `json:"vertical"`
}

func (s *Server) getPersonality(w http.ResponseWriter, r *http.Request) {
	ctx, _ := readFileOrEmpty(filepath.Join(s.hostDir, "context.md"))
	vert, _ := readFileOrEmpty(filepath.Join(s.hostDir, "vertical.md"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(personalityResp{Context: ctx, Vertical: vert})
}

func (s *Server) putPersonality(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Context  *string `json:"context"`
		Vertical *string `json:"vertical"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Context != nil {
		if err := os.WriteFile(filepath.Join(s.hostDir, "context.md"), []byte(*req.Context), 0644); err != nil {
			log.Error().Err(err).Msg("owner: write context.md")
			jsonErr(w, "write failed", http.StatusInternalServerError)
			return
		}
	}
	if req.Vertical != nil {
		if err := os.WriteFile(filepath.Join(s.hostDir, "vertical.md"), []byte(*req.Vertical), 0644); err != nil {
			log.Error().Err(err).Msg("owner: write vertical.md")
			jsonErr(w, "write failed", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Policy ---

func (s *Server) getPolicy(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.ownerMgr.Load()
	if err != nil {
		jsonErr(w, "load failed", http.StatusInternalServerError)
		return
	}
	tags := cfg.ToolTags
	if tags == nil {
		tags = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"tool_tags": tags,
	})
}

func (s *Server) putPolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ToolTags []string `json:"tool_tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid json", http.StatusBadRequest)
		return
	}
	for _, tag := range req.ToolTags {
		if strings.TrimSpace(tag) == "" {
			jsonErr(w, "tool_tags must not contain empty strings", http.StatusBadRequest)
			return
		}
	}
	if err := s.ownerMgr.Update(func(cfg *visitor.OwnerConfig) error {
		cfg.ToolTags = req.ToolTags
		return nil
	}); err != nil {
		jsonErr(w, "save failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Audit ---

func (s *Server) getAudit(w http.ResponseWriter, r *http.Request) {
	// TODO: stream instead of loading all into memory when audit.jsonl grows large;
	// track in symmate-build.md decisions log under getAudit scalability.
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	const limit = 50
	skip := (page - 1) * limit

	path := filepath.Join(s.hostDir, "audit.jsonl")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entries": []any{}, "page": page, "total": 0, "pages": 0})
		return
	}
	if err != nil {
		jsonErr(w, "read failed", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	var all []json.RawMessage
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) > 0 {
			cp := make([]byte, len(line))
			copy(cp, line)
			all = append(all, cp)
		}
	}

	// Reverse for newest-first display.
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}

	total := len(all)
	end := skip + limit
	if end > total {
		end = total
	}
	pageItems := []json.RawMessage{}
	if skip < total {
		pageItems = all[skip:end]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"entries": pageItems,
		"page":    page,
		"total":   total,
		"pages":   (total + limit - 1) / limit,
	})
}

// --- Conversations ---

func (s *Server) listConversations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	const limit = 50
	offset := (page - 1) * limit

	convs, err := s.db.ListConversations(r.Context(), limit, offset)
	if err != nil {
		jsonErr(w, "db error", http.StatusInternalServerError)
		return
	}

	out := make([]item, 0, len(convs))
	for _, c := range convs {
		out = append(out, item{
			ID:            c.ID,
			VisitorHandle: c.VisitorHandle,
			Surface:       c.Surface,
			StartedAt:     c.StartedAt,
			EndedAt:       c.EndedAt,
			Outcome:       c.Outcome,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"conversations": out, "page": page})
}

func (s *Server) getConversation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	conv, err := s.db.GetConversation(ctx, id)
	if err != nil {
		jsonErr(w, "not found", http.StatusNotFound)
		return
	}
	msgs, err := s.db.GetMessages(ctx, id)
	if err != nil {
		jsonErr(w, "db error", http.StatusInternalServerError)
		return
	}

	out := make([]msgItem, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, msgItem{Role: m.Role, Content: m.Content, TS: m.TS})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":             conv.ID,
		"visitor_handle": conv.VisitorHandle,
		"surface":        conv.Surface,
		"started_at":     conv.StartedAt,
		"ended_at":       conv.EndedAt,
		"outcome":        conv.Outcome,
		"messages":       out,
	})
}

// --- LLM ---

func (s *Server) getLLM(w http.ResponseWriter, r *http.Request) {
	cfg := s.llmStore.Get()
	if cfg == nil {
		cfg = &llm.OverrideConfig{Provider: "hermes", Model: "hermes"}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"provider":        cfg.Provider,
		"model":           cfg.Model,
		"custom_endpoint": cfg.CustomEndpoint,
		"has_custom_key":  cfg.CustomAPIKey != "",
	})
}

func (s *Server) putLLM(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider       string `json:"provider"`
		Model          string `json:"model"`
		CustomEndpoint string `json:"custom_endpoint"`
		CustomAPIKey   string `json:"custom_api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid json", http.StatusBadRequest)
		return
	}
	if !validProviders[req.Provider] {
		jsonErr(w, "invalid provider: must be anthropic, hermes, or openai-compat", http.StatusBadRequest)
		return
	}
	cfg := &llm.OverrideConfig{
		Provider:       req.Provider,
		Model:          req.Model,
		CustomEndpoint: req.CustomEndpoint,
		CustomAPIKey:   req.CustomAPIKey,
	}
	// Preserve existing key if not supplied.
	if cfg.CustomAPIKey == "" {
		if existing := s.llmStore.Get(); existing != nil {
			cfg.CustomAPIKey = existing.CustomAPIKey
		}
	}
	if err := s.llmStore.Save(cfg); err != nil {
		jsonErr(w, "save failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Helpers ---

func (s *Server) requireToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			jsonErr(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		s.mu.Lock()
		exp, ok := s.sessions[token]
		s.mu.Unlock()
		if !ok || time.Now().After(exp) {
			jsonErr(w, "session expired", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) gc() {
	now := time.Now()
	for t, exp := range s.sessions {
		if now.After(exp) {
			delete(s.sessions, t)
		}
	}
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", s.allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}

func mintToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(data), err
}
