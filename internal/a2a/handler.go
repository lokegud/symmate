package a2a

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Handler is the HTTP handler for the A2A endpoint (port 7843).
// It serves:
//   - GET /.well-known/agent-card.json
//   - POST /agent/sendMessage (JSON-RPC 2.0)
type Handler struct {
	card    AgentCard
	msgHdlr *MessageHandler
}

// NewHandler creates an A2A Handler.
func NewHandler(card AgentCard, msgHdlr *MessageHandler) *Handler {
	return &Handler{card: card, msgHdlr: msgHdlr}
}

// RegisterRoutes registers the A2A routes on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/.well-known/agent-card.json", h.serveAgentCard)
	mux.HandleFunc("/agent/sendMessage", h.handleRPC)
	mux.HandleFunc("/agent/getCapabilities", h.handleRPC)
	mux.HandleFunc("/agent/getAgentCard", h.handleRPC)
}

// ServeHTTP lets Handler act as an http.Handler directly.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/.well-known/agent-card.json":
		h.serveAgentCard(w, r)
	case "/agent/sendMessage", "/agent/getCapabilities", "/agent/getAgentCard":
		h.handleRPC(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) serveAgentCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, h.card)
}

func (h *Handler) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      interface{}     `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, nil, &RPCError{Code: ErrParse, Message: "parse error: " + err.Error()})
		return
	}
	if req.JSONRPC != "2.0" {
		writeRPCError(w, req.ID, &RPCError{Code: ErrInvalidRequest, Message: "jsonrpc must be '2.0'"})
		return
	}

	switch req.Method {
	case "agent/sendMessage":
		result, rpcErr := h.msgHdlr.Handle(r.Context(), req.Params)
		if rpcErr != nil {
			log.Warn().Int("code", rpcErr.Code).Str("msg", rpcErr.Message).Msg("a2a rpc error")
			writeRPCError(w, req.ID, rpcErr)
			return
		}
		writeRPCResult(w, req.ID, result)

	case "agent/getAgentCard":
		writeRPCResult(w, req.ID, h.card)

	case "agent/getCapabilities":
		writeRPCResult(w, req.ID, h.card.Capabilities)

	default:
		writeRPCError(w, req.ID, &RPCError{
			Code:    ErrMethodNotFound,
			Message: "method not found: " + req.Method,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("json encode")
	}
}

func writeRPCResult(w http.ResponseWriter, id interface{}, result interface{}) {
	writeJSON(w, http.StatusOK, Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	})
}

func writeRPCError(w http.ResponseWriter, id interface{}, rpcErr *RPCError) {
	writeJSON(w, http.StatusOK, Response{
		JSONRPC: "2.0",
		Error:   rpcErr,
		ID:      id,
	})
}
