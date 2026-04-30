package visitor

import (
	"fmt"
	"net/http"
)

// SSEWriter writes SSE events to an http.ResponseWriter with proper framing and flush.
type SSEWriter struct {
	w http.ResponseWriter
}

// NewSSEWriter sets SSE headers and returns a writer.
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no") // tell nginx not to buffer
	return &SSEWriter{w: w}
}

// Send writes a data event and flushes.
func (s *SSEWriter) Send(data string) error {
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return err
	}
	s.flush()
	return nil
}

// SendEvent writes a named event and flushes.
func (s *SSEWriter) SendEvent(event, data string) error {
	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	s.flush()
	return nil
}

// Done sends the [DONE] sentinel used by streaming clients.
func (s *SSEWriter) Done() {
	fmt.Fprintf(s.w, "data: [DONE]\n\n")
	s.flush()
}

func (s *SSEWriter) flush() {
	if f, ok := s.w.(http.Flusher); ok {
		f.Flush()
	}
}
