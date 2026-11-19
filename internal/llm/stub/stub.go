// Package stub provides a fake LLM server for integration tests.
package stub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Server is a fake LLM API server backed by httptest.
type Server struct {
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

func New() *Server {
	s := &Server{reply: "stub reply"}
	s.srv = httptest.NewServer(http.HandlerFunc(s.handle))
	s.URL = s.srv.URL
	return s
}

func (s *Server) SetReply(reply string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reply = reply
}

func (s *Server) SetError(statusCode int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextErr = &nextError{statusCode: statusCode, message: message}
}

func (s *Server) SetLatency(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latency = d
}

func (s *Server) Close() { s.srv.Close() }

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	lat := s.latency
	nerr := s.nextErr
	if nerr != nil {
		s.nextErr = nil
	}
	reply := s.reply
	s.mu.Unlock()

	if lat > 0 {
		time.Sleep(lat)
	}

	switch r.URL.Path {
	case "/v1/messages":
		s.handleAnthropic(w, r, nerr, reply)
	case "/v1/chat/completions":
		s.handleOpenAI(w, r, nerr, reply)
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

func (s *Server) handleAnthropic(w http.ResponseWriter, _ *http.Request, nerr *nextError, reply string) {
	w.Header().Set("Content-Type", "application/json")
	if nerr != nil {
		w.WriteHeader(nerr.statusCode)
		var enc anthropicError
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

type openAIStreamChunk struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Choices []sseChoice `json:"choices"`
}

type sseChoice struct {
	Index        int      `json:"index"`
	Delta        sseDelta `json:"delta"`
	FinishReason *string  `json:"finish_reason"`
}

type sseDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content"`
}

type openAIError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type openAIRequest struct {
	Stream bool `json:"stream"`
}

func (s *Server) handleOpenAI(w http.ResponseWriter, r *http.Request, nerr *nextError, reply string) {
	var req openAIRequest
	json.NewDecoder(r.Body).Decode(&req)

	if nerr != nil {
		if req.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(nerr.statusCode)
			errChunk := map[string]interface{}{
				"error": map[string]string{"message": nerr.message},
			}
			data, _ := json.Marshal(errChunk)
			fmt.Fprintf(w, "data: %s\n\ndata: [DONE]\n\n", data)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nerr.statusCode)
		var enc openAIError
		enc.Error.Message = nerr.message
		json.NewEncoder(w).Encode(enc)
		return
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		stop := "stop"
		// Role announcement: role present, no content (proper OpenAI streaming format).
		roleChunk := openAIStreamChunk{
			ID:     "stub-id",
			Object: "chat.completion.chunk",
			Choices: []sseChoice{
				{Index: 0, Delta: sseDelta{Role: "assistant"}},
			},
		}
		data, _ := json.Marshal(roleChunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		// Content chunk: no role.
		contentChunk := openAIStreamChunk{
			ID:     "stub-id",
			Object: "chat.completion.chunk",
			Choices: []sseChoice{
				{Index: 0, Delta: sseDelta{Content: reply}},
			},
		}
		data, _ = json.Marshal(contentChunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		// Stop chunk.
		contentChunk.Choices[0].Delta = sseDelta{}
		contentChunk.Choices[0].FinishReason = &stop
		data, _ = json.Marshal(contentChunk)
		fmt.Fprintf(w, "data: %s\n\ndata: [DONE]\n\n", data)
		return
	}

	type openAIResponse struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(openAIResponse{
		Choices: []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		}{{Message: struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{Role: "assistant", Content: reply}}},
	})
}
