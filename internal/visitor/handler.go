package visitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/loke/symmate/internal/aeo"
)

// AgentCard is the A2A Agent Card served at /.well-known/agent-card.json.
// Read-only copy; the authoritative one is on the A2A port (7843).
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

// ChatRequest is the body accepted by POST /chat.
type ChatRequest struct {
	Message       string `json:"message"`
	VisitorHandle string `json:"visitor_handle,omitempty"`
}

// NewRouter builds the visitor API chi router.
// profile is pre-built; brain is stubbed until Forge merges.
func NewRouter(hc *aeo.HostContext, brain Brain, sessions *Sessions) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Post("/chat", chatHandler(hc, brain, sessions))
	r.Get("/id", idHandler(hc.Profile))
	r.Get("/llms.txt", llmsTxtHandler(hc))
	r.Get("/.well-known/agent-card.json", agentCardHandler(hc.Profile))
	r.Get("/health", healthHandler)

	return r
}

func chatHandler(hc *aeo.HostContext, brain Brain, sessions *Sessions) http.HandlerFunc {
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

		msgs := []Message{
			{Role: "user", Content: req.Message},
		}

		ctx := r.Context()
		ch, err := brain.Chat(ctx, msgs)
		if err != nil {
			http.Error(w, "brain error", http.StatusInternalServerError)
			return
		}

		sse := NewSSEWriter(w)
		for token := range ch {
			if ctx.Err() != nil {
				return
			}
			if err := sse.Send(token); err != nil {
				return
			}
			sess.AddTokens(len(token))
		}
		sse.Done()
	}
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
		card := buildAgentCard(p)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		json.NewEncoder(w).Encode(card)
	}
}

func buildAgentCard(p *aeo.Profile) AgentCard {
	skills := []agentSkill{
		{
			ID:          "conversation",
			Name:        "Conversation",
			Description: "Conversational interface for visitors.",
			Tags:        []string{"conversation", "general"},
		},
	}

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
			Streaming:              true,
			PushNotifications:      false,
			StateTransitionHistory: false,
		},
		DefaultInputModes:  []string{"text/plain"},
		DefaultOutputModes: []string{"text/plain"},
		Skills:             skills,
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok\n")
}
