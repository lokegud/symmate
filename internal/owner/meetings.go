package owner

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/loke/tiers"
)

// WithMeetings attaches the meeting-request store. Nil disables the routes.
func (s *Server) WithMeetings(m *tier.MeetingStore) *Server {
	s.meetings = m
	return s
}

// listMeetings returns every visitor meeting request, oldest first.
func (s *Server) listMeetings(w http.ResponseWriter, r *http.Request) {
	if s.meetings == nil {
		jsonErr(w, "meeting requests not configured", http.StatusServiceUnavailable)
		return
	}
	all, err := s.meetings.List()
	if err != nil {
		jsonErr(w, "load failed", http.StatusInternalServerError)
		return
	}
	if all == nil {
		all = []tier.MeetingRequest{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"requests": all})
}

// decideMeeting approves or declines a request. Approval does not book anything;
// the owner does that in their own calendar.
func (s *Server) decideMeeting(w http.ResponseWriter, r *http.Request) {
	if s.meetings == nil {
		jsonErr(w, "meeting requests not configured", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body); err != nil {
		jsonErr(w, "invalid body", http.StatusBadRequest)
		return
	}
	ok, err := s.meetings.Decide(chi.URLParam(r, "id"), body.Status)
	if err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		jsonErr(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": body.Status})
}
