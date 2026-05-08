package visitor

import (
	"fmt"
	"net/http"
)

// SSEWriter writes SSE events to an http.ResponseWriter with proper framing and flush.
type SSEWriter struct {
	w http.ResponseWriter
}

// NewSSEWriter sets SSE headers, writes an SSE comment to start the stream immediately
// (keeps long-TTFT models from appearing stuck in the browser), and returns a writer.
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no") // tell nginx not to buffer
	// SSE comment starts the HTTP response body immediately so the browser
	// sees the connection as established rather than pending.
	fmt.Fprint(w, ": connected\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
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
