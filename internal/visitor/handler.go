package visitor

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/loke/symmate/internal/aeo"
)

//go:embed chat.html
var chatHTML []byte

type AgentCard struct {
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	URL                string            `json:"url"`
	Version            string            `json:"version"`
	Capabilities       agentCapabilities `json:"capabilities"`
	DefaultInputModes  []string          `json:"defaultInputModes"`
	DefaultOutputModes []string          `json:"defaultOutputModes"`
	Skills             []agentSkill      `json:"skills"`
}

type agentCapabilities struct {
	Streaming              bool `json:"streaming"`
	PushNotifications      bool `json:"pushNotifications"`
	StateTransitionHistory bool `json:"stateTransitionHistory"`
}

type agentSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type ChatRequest struct {
	Message       string `json:"message"`
	VisitorHandle string `json:"visitor_handle,omitempty"`
	DeviceToken   string `json:"device_token,omitempty"`
}

const minPhraseLen = 12

// NewRouter builds the visitor API chi router.
// ownerMgr may be nil — when nil, owner recognition is disabled.
// store may be nil — when nil, conversation history is disabled.
// forcefield may be nil — when nil, all messages pass through unchecked.
func NewRouter(hc *aeo.HostContext, brain Brain, sessions *Sessions, ff *Forcefield, store ConversationStore, ownerMgr *OwnerManager) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(chatHTML)
	})
	r.Post("/chat", chatHandler(hc, brain, sessions, ff, store, ownerMgr))
	r.Get("/id", idHandler(hc.Profile))
	r.Get("/llms.txt", llmsTxtHandler(hc))
	r.Get("/.well-known/agent-card.json", agentCardHandler(hc.Profile))
	r.Get("/health", healthHandler)

	return r
}

func chatHandler(hc *aeo.HostContext, brain Brain, sessions *Sessions, ff *Forcefield, store ConversationStore, ownerMgr *OwnerManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := sessions.Get(r.RemoteAddr)
		if !sess.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Message) == "" {
			http.Error(w, "message is required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// --- Owner recognition (only when ownerMgr is configured) ---
		if ownerMgr != nil {
			if ownerAction := handleOwnerRecognition(w, r, req, sess, ownerMgr); ownerAction {
				return
			}
		}

		if ff != nil {
			callerType := "human"
			if r.Header.Get("X-A2A-Sender") != "" || r.Header.Get("Agent-Card") != "" {
				callerType = "agent"
			}
			callerID := r.RemoteAddr
			if handle := req.VisitorHandle; handle != "" {
				callerID = handle
			}
			allow, reportID, _ := ff.Check(r.Context(), callerID, callerType, req.Message,
				fmt.Sprintf("Symmate for %s", hc.Profile.Name))
			if !allow {
				sse := NewSSEWriter(w)
				sse.Send(BlockedMessage(reportID))
				sse.Done()
				return
			}
		}

		convID := sess.ConversationID()

		// Ensure conversation row exists; load history.
		var history []Message
		if store != nil {
			var handle *string
			if req.VisitorHandle != "" {
				h := req.VisitorHandle
				handle = &h
			}
			_ = store.EnsureConversation(ctx, convID, "visitor", handle)
			history, _ = store.RecentMessages(ctx, convID, historyLimit)
		}

		// Append current user turn to history for the LLM, then persist it.
		msgs := append(history, Message{Role: "user", Content: req.Message})
		if store != nil {
			_ = store.AppendMessage(ctx, convID, "user", req.Message)
		}

		// Inject owner mode into context if session is recognised.
		if sess.IsOwner() {
			ctx = WithOwnerMode(ctx, true)
		}

		// Start SSE stream immediately so the browser sees the connection as live
		// while we wait for the LLM (which may have high TTFT on large contexts).
		sse := NewSSEWriter(w)

		ch, err := brain.Chat(ctx, msgs)
		if err != nil {
			sse.Send("(error: brain unavailable)")
			sse.Done()
			return
		}
		var reply strings.Builder
		for token := range ch {
			if ctx.Err() != nil {
				return
			}
			if err := sse.Send(token); err != nil {
				return
			}
			reply.WriteString(token)
			sess.AddTokens(len(token))
		}
		sse.Done()

		if store != nil && reply.Len() > 0 {
			_ = store.AppendMessage(ctx, convID, "assistant", reply.String())
		}
	}
}

// handleOwnerRecognition processes enrollment and phrase-check logic.
// Returns true if the request was fully handled (response written) and no further processing needed.
func handleOwnerRecognition(w http.ResponseWriter, r *http.Request, req ChatRequest, sess *Session, ownerMgr *OwnerManager) bool {
	cfg, err := ownerMgr.Load()
	if err != nil {
		return false // config unreadable — skip recognition, don't break visitors
	}

	// Check device fingerprint before processing the message.
	if req.DeviceToken != "" && cfg.Enrolled && cfg.KnownDevice(req.DeviceToken) {
		sess.AddDeviceMatch()
	}

	// Enrollment: only when not yet enrolled and message starts with "ENROLL ".
	if !cfg.Enrolled {
		const enrollPrefix = "ENROLL "
		if strings.HasPrefix(req.Message, enrollPrefix) {
			phrase := strings.TrimPrefix(req.Message, enrollPrefix)
			phrase = strings.TrimSpace(phrase)
			if len(phrase) >= minPhraseLen {
				if err := cfg.Enroll(phrase, req.DeviceToken); err == nil {
					_ = ownerMgr.Save(cfg)
					sse := NewSSEWriter(w)
					sse.Send("Enrollment complete. Trigger phrase set and this device registered. I'll recognize you next time.")
					sse.Done()
					return true
				}
			}
		}
		// Not enrolled yet — no recognition possible
		return false
	}

	// Phrase check: only when not yet in owner mode for this session.
	if !sess.IsOwner() {
		candidate := strings.TrimSpace(req.Message)
		if cfg.CheckPhrase(candidate) {
			sess.AddPhraseMatch()
			// Auto-enroll new device when phrase matches
			if req.DeviceToken != "" && !cfg.KnownDevice(req.DeviceToken) {
				cfg.AddDevice(req.DeviceToken)
				_ = ownerMgr.Save(cfg)
			}
			sse := NewSSEWriter(w)
			sse.Send("Welcome back. Owner mode active for this session.")
			sse.Done()
			return true
		}
	}

	return false
}

func idHandler(p *aeo.Profile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		json.NewEncoder(w).Encode(p)
	}
}

func llmsTxtHandler(hc *aeo.HostContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := hc.Profile
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=300")

		var sb strings.Builder
		fmt.Fprintf(&sb, "# %s\n\n", p.Name)
		fmt.Fprintf(&sb, "> %s\n\n", p.Description)
		if p.Contact.Location != "" {
			fmt.Fprintf(&sb, "Location: %s\n", p.Contact.Location)
		}
		if p.Contact.Website != "" {
			fmt.Fprintf(&sb, "Website: %s\n", p.Contact.Website)
		}
		fmt.Fprintf(&sb, "Host type: %s\n", p.HostType)
		fmt.Fprintf(&sb, "Schema version: %s\n\n", p.SchemaVer)
		fmt.Fprintf(&sb, "## Capabilities\n\n")
		for _, c := range p.Capabilities {
			fmt.Fprintf(&sb, "- %s\n", c)
		}
		fmt.Fprintln(&sb)
		if hc.Hours != "" {
			fmt.Fprintf(&sb, "## Availability\n\n%s\n\n", hc.Hours)
		}
		if hc.Catalog != "" {
			fmt.Fprintf(&sb, "## Catalog\n\n%s\n\n", hc.Catalog)
		}
		if hc.Policies != "" {
			fmt.Fprintf(&sb, "## Policies\n\n%s\n\n", hc.Policies)
		}
		fmt.Fprintf(&sb, "## Agent Endpoints\n\n")
		fmt.Fprintf(&sb, "Visitor API: %s\n", p.Endpoints.Visitor)
		if p.Endpoints.MCP != "" {
			fmt.Fprintf(&sb, "MCP server: %s\n", p.Endpoints.MCP)
		}
		if p.Endpoints.A2A != "" {
			fmt.Fprintf(&sb, "A2A endpoint: %s\n", p.Endpoints.A2A)
		}
		fmt.Fprintf(&sb, "\nGenerated: %s\n", time.Now().UTC().Format(time.RFC3339))
		fmt.Fprint(w, sb.String())
	}
}

func agentCardHandler(p *aeo.Profile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		json.NewEncoder(w).Encode(buildAgentCard(p))
	}
}

func buildAgentCard(p *aeo.Profile) AgentCard {
	a2aURL := p.Endpoints.A2A
	if a2aURL == "" {
		a2aURL = p.Endpoints.Visitor
	}
	return AgentCard{
		Name:        p.Name,
		Description: p.Description,
		URL:         a2aURL,
		Version:     "0.1",
		Capabilities: agentCapabilities{
			Streaming: true,
		},
		DefaultInputModes:  []string{"text/plain"},
		DefaultOutputModes: []string{"text/plain"},
		Skills: []agentSkill{{
			ID:          "conversation",
			Name:        "Conversation",
			Description: "Conversational interface for visitors.",
			Tags:        []string{"conversation", "general"},
		}},
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok\n")
}
