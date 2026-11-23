package visitor

import (
	"encoding/json"
	"net/http"

	"github.com/loke/symmate/internal/tier"
)

// meetingRequestHandler lets a tier 3 visitor ask for time. The request waits for
// the owner's approval; nothing is booked here.
func meetingRequestHandler(sessions *Sessions, store *tier.MeetingStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := sessions.Get(r.RemoteAddr)
		if !sess.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		if store == nil || sess.Tier() < tier.Schedule {
			http.Error(w, "not available", http.StatusForbidden)
			return
		}
		var body struct {
			Name string `json:"name"`
			When string `json:"when"`
			Note string `json:"note"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		req, err := store.Add(body.Name, body.When, body.Note)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"id": req.ID, "status": req.Status})
	}
}
