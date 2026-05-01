// Package stub provides a fake LLM server for integration tests.
// It speaks both the Anthropic Messages API (/v1/messages) and the
// OpenAI-compatible API (/v1/chat/completions) so tests can exercise
// AnthropicProvider and HermesProvider without hitting the real network.
package stub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Server is a fake LLM API server backed by httptest.
type Server struct {
	// URL is the base URL (e.g. "http://127.0.0.1:PORT").
	// For Anthropic tests pass Server.URL+"/v1/messages" as the endpoint.
	// For OpenAI-compat tests pass Server.URL+"/v1" as the endpoint.
	URL string

	srv *httptest.Server
	mu  sync.Mutex

	reply   string
	nextErr *nextError
	latency time.Duration
}

type nextError struct {
	statusCode int
	message    string
}

// New starts a stub server and returns it. Call Close when done.
func New() *Server {
	s := &Server{reply: "stub reply"}
	s.srv = httptest.NewServer(http.HandlerFunc(s.handle))
	s.URL = s.srv.URL
	return s
}

// SetReply sets the text the stub will return for the next (and subsequent) calls.
func (s *Server) SetReply(reply string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reply = reply
}

// SetError causes the next call to return an API-level error.
// statusCode is the HTTP status (e.g. 429, 500).
// message is the error text embedded in the provider's native error format.
// The error is consumed after one use; subsequent calls return the normal reply.
func (s *Server) SetError(statusCode int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextErr = &nextError{statusCode: statusCode, message: message}
}

// SetLatency adds a fixed delay before every response. Pass 0 to disable.
func (s *Server) SetLatency(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latency = d
}

// Close shuts down the stub server.
func (s *Server) Close() {
	s.srv.Close()
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	lat := s.latency
	err := s.nextErr
	if err != nil {
		s.nextErr = nil
	}
	reply := s.reply
	s.mu.Unlock()

	if lat > 0 {
		time.Sleep(lat)
	}

	switch r.URL.Path {
	case "/v1/messages":
		s.handleAnthropic(w, r, err, reply)
	case "/v1/chat/completions":
		s.handleOpenAI(w, r, err, reply)
	default:
		http.NotFound(w, r)
	}
}

// --- Anthropic format ---

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicError struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (s *Server) handleAnthropic(w http.ResponseWriter, r *http.Request, nerr *nextError, reply string) {
	w.Header().Set("Content-Type", "application/json")
	if nerr != nil {
		w.WriteHeader(nerr.statusCode)
		enc := anthropicError{}
		enc.Error.Type = "error"
		enc.Error.Message = nerr.message
		json.NewEncoder(w).Encode(enc)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(anthropicResponse{
		Content: []anthropicContent{{Type: "text", Text: reply}},
	})
}

// --- OpenAI-compat format ---

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *Server) handleOpenAI(w http.ResponseWriter, r *http.Request, nerr *nextError, reply string) {
	w.Header().Set("Content-Type", "application/json")
	if nerr != nil {
		w.WriteHeader(nerr.statusCode)
		enc := openAIError{}
		enc.Error.Message = nerr.message
		json.NewEncoder(w).Encode(enc)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(openAIResponse{
		Choices: []openAIChoice{
			{Message: openAIMessage{Role: "assistant", Content: reply}},
		},
	})
}
