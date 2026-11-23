package visitor

import (
	"encoding/json"
	"net/http"

	"github.com/loke/tiers"
)

// verifyHandler checks a visitor's three facts in code and, on a full match,
// raises this session to tier 3. It never tells the caller which fact missed.
func verifyHandler(sessions *Sessions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := sessions.Get(r.RemoteAddr)
		if !sess.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		v := sessions.Verifier()
		if v == nil {
			http.Error(w, "verification unavailable", http.StatusServiceUnavailable)
			return
		}
		if sess.VerifyLocked() {
			http.Error(w, "too many attempts", http.StatusTooManyRequests)
			return
		}

		var f tier.Facts
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&f); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if !v.Verify(f) {
			sess.RecordVerifyFailure()
			writeJSON(w, http.StatusForbidden, map[string]any{"verified": false})
			return
		}

		sess.SetTier(tier.Schedule)
		writeJSON(w, http.StatusOK, map[string]any{"verified": true, "tier": tier.Schedule.String()})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
