package a2a

import (
	"net/http"

	sdk "github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
)

// Handler bundles the SDK's JSON-RPC handler and agent-card handler into a
// single http.Handler. It serves:
//
//   - GET /.well-known/agent-card.json  →  static AgentCard
//   - POST <anything else>              →  JSON-RPC (message/send, tasks/get, ...)
//
// The SDK owns method names, message format, error shape, and version
// negotiation. Symmate's value-add (injection check + privacy firewall)
// lives inside the AgentExecutor, behind this surface.
type Handler struct {
	rpc  http.Handler
	card http.Handler
}

// NewHandler builds the A2A handler from a card and an executor.
func NewHandler(card *sdk.AgentCard, executor a2asrv.AgentExecutor) *Handler {
	rh := a2asrv.NewHandler(executor)
	return &Handler{
		rpc:  a2asrv.NewJSONRPCHandler(rh),
		card: a2asrv.NewStaticAgentCardHandler(card),
	}
}

// ServeHTTP routes by path: agent-card → card handler, else → JSON-RPC.
// Extracts X-A2A-Sender-Card header and injects it into the request context
// so the executor can check Sygil reputation without needing HTTP access.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == a2asrv.WellKnownAgentCardPath {
		h.card.ServeHTTP(w, r)
		return
	}
	if cardURL := r.Header.Get("X-A2A-Sender-Card"); cardURL != "" {
		r = r.WithContext(WithSenderCardURL(r.Context(), cardURL))
	}
	h.rpc.ServeHTTP(w, r)
}
