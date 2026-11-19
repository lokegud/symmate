package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/loke/symmate/internal/aeo"
	"github.com/rs/zerolog/log"
)

const mcpProtocolVersion = "2024-11-05"

// JSON-RPC 2.0 types.

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP initialize result.
type initResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    mcpCaps      `json:"capabilities"`
	ServerInfo      mcpInfo      `json:"serverInfo"`
}

type mcpCaps struct {
	Tools *struct{} `json:"tools,omitempty"`
}

type mcpInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MCP tools/list result.
type toolsListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

// MCP tools/call params.
type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Server is the MCP HTTP+SSE server.
// HTTP+SSE transport only — no stdio (OX Security RCE advisory, April 2026).
type Server struct {
	tools        *ToolRegistry
	policy       *CapabilityPolicy
	audit        AuditLogger
	publicMsgURL string // absolute message URL prefix when SYMMATE_MCP_PUBLIC_URL is set
}

// NewServer creates a MCP server for the given host context.
func NewServer(hc *aeo.HostContext, audit AuditLogger) *Server {
	if audit == nil {
		audit = NoopAuditLogger{}
	}
	var publicMsgURL string
	if base := os.Getenv("SYMMATE_MCP_PUBLIC_URL"); base != "" {
		publicMsgURL = base + "/message"
	}
	return &Server{
		tools:        newToolRegistry(hc),
		policy:       DefaultPolicy(),
		audit:        audit,
		publicMsgURL: publicMsgURL,
	}
}

// Router returns the HTTP handler for the MCP server.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/sse", s.handleSSE)
	r.Post("/message", s.handleMessage)
	r.Get("/.well-known/mcp.json", s.handleWellKnown)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok\n")
	})

	return r
}

// handleSSE handles the SSE subscription endpoint.
// Sends the endpoint event, then holds the connection open with periodic pings
// until the client disconnects.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	sessionID := newSessionID()

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	msgBase := "/message"
	if s.publicMsgURL != "" {
		msgBase = s.publicMsgURL
	}
	msgURL := fmt.Sprintf("%s?sessionId=%s", msgBase, sessionID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", msgURL)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Hold the connection, sending a comment ping every 30s to keep proxies alive.
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// handleMessage accepts JSON-RPC 2.0 requests from MCP clients.
// Responds directly via HTTP (not via SSE) — valid per MCP spec §4.2.
func (s *Server) handleMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	callerIP := r.RemoteAddr

	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, nil, -32700, "parse error")
		return
	}
	if req.JSONRPC != "2.0" {
		writeRPCError(w, req.ID, -32600, "invalid request: jsonrpc must be 2.0")
		return
	}

	log.Debug().Str("session", sessionID).Str("method", req.Method).Msg("mcp request")

	switch req.Method {
	case "initialize":
		writeRPCResult(w, req.ID, initResult{
			ProtocolVersion: mcpProtocolVersion,
			Capabilities:    mcpCaps{Tools: &struct{}{}},
			ServerInfo:      mcpInfo{Name: "symmate-mcp", Version: "0.1.0"},
		})

	case "initialized":
		// Client notification — no response required.
		w.WriteHeader(http.StatusAccepted)

	case "tools/list":
		writeRPCResult(w, req.ID, toolsListResult{Tools: s.tools.Definitions()})

	case "tools/call":
		s.handleToolCall(w, r, req, sessionID, callerIP)

	case "ping":
		writeRPCResult(w, req.ID, struct{}{})

	default:
		writeRPCError(w, req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *Server) handleToolCall(w http.ResponseWriter, r *http.Request, req rpcRequest, sessionID, callerIP string) {
	var params toolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeRPCError(w, req.ID, -32602, "invalid params")
		return
	}

	allowed, reason := s.policy.Check(params.Name)

	s.audit.Log(AuditEvent{
		Timestamp:  time.Now(),
		SessionID:  sessionID,
		ToolName:   params.Name,
		CallerIP:   callerIP,
		Allowed:    allowed,
		DenyReason: reason,
	})

	if !allowed {
		writeRPCError(w, req.ID, -32603, fmt.Sprintf("tool denied: %s", reason))
		return
	}

	result, err := s.tools.Call(params.Name, params.Arguments)
	if err != nil {
		writeRPCError(w, req.ID, -32603, err.Error())
		return
	}

	writeRPCResult(w, req.ID, result)
}

func writeRPCResult(w http.ResponseWriter, id json.RawMessage, result interface{}) {
	resp := rpcResponse{JSONRPC: "2.0", ID: id, Result: result}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeRPCError(w http.ResponseWriter, id json.RawMessage, code int, msg string) {
	resp := rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors are still 200
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleWellKnown(w http.ResponseWriter, r *http.Request) {
	type mcpDiscovery struct {
		Name            string `json:"name"`
		Version         string `json:"version"`
		ProtocolVersion string `json:"protocolVersion"`
		Transport       string `json:"transport"`
		URL             string `json:"url,omitempty"`
		Capabilities    struct {
			Tools bool `json:"tools"`
		} `json:"capabilities"`
	}
	d := mcpDiscovery{
		Name:            s.tools.hc.Profile.Name,
		Version:         "0.1.0",
		ProtocolVersion: mcpProtocolVersion,
		Transport:       "http+sse",
		URL:             s.tools.hc.Profile.Endpoints.MCP,
	}
	d.Capabilities.Tools = true
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

func newSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
