package visitor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/loke/tiers"
)

// levelSeen sends one chat as addr, optionally with a registered user's token,
// and reports the tier the brain received, or false if the brain was never called.
func levelSeen(t *testing.T, sessions *Sessions, m *OwnerManager, addr, token string) (tier.Level, bool, string) {
	t.Helper()
	brain := &ctxBrain{}
	router := NewRouter(testHostContext(), brain, sessions, nil, nil, m)
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"hello"}`))
	req.RemoteAddr = addr
	if token != "" {
		req.Header.Set("X-Symmate-User", token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	brain.mu.Lock()
	defer brain.mu.Unlock()
	if brain.ctx == nil {
		return 0, false, w.Body.String()
	}
	return tier.LevelFrom(brain.ctx), true, w.Body.String()
}

func TestTierScenarios(t *testing.T) {
	m := NewOwnerManager(t.TempDir())
	var blockedTok, fullTok, plainTok string
	_ = m.Update(func(c *OwnerConfig) error {
		c.UsersEnabled = true
		u1, tok1 := c.AddUser("Blocked", nil)
		_ = c.SetUserTier(u1.ID, 1)
		_, tok2 := c.AddUser("Full", nil)
		u2 := c.Users[len(c.Users)-1]
		_ = c.SetUserTier(u2.ID, 4)
		_, tok3 := c.AddUser("Plain", nil)
		blockedTok, fullTok, plainTok = tok1, tok2, tok3
		return nil
	})

	// Tier 1: blocked. Refused, and the brain is never called.
	sessions := NewSessions(0, 0)
	if _, called, body := levelSeen(t, sessions, m, "10.1.0.1:1", blockedTok); called || !strings.Contains(body, "can't help") {
		t.Fatalf("tier 1 must be refused before the brain: called=%v body=%q", called, body)
	}

	// Tier 2: anonymous visitor gets the public default.
	if level, called, _ := levelSeen(t, NewSessions(0, 0), m, "10.1.0.2:1", ""); !called || level != tier.Public {
		t.Fatalf("tier 2 anonymous: level=%v called=%v", level, called)
	}

	// Tier 3: a visitor verified on this session gets schedule.
	verifiedSessions := NewSessions(0, 0)
	verifiedSessions.Get("10.1.0.3:1").SetTier(tier.Schedule)
	if level, called, _ := levelSeen(t, verifiedSessions, m, "10.1.0.3:1", ""); !called || level != tier.Schedule {
		t.Fatalf("tier 3 verified: level=%v called=%v", level, called)
	}

	// Tier 4: a registered user with tier 4 gets all, even from a fresh session.
	if level, called, _ := levelSeen(t, NewSessions(0, 0), m, "10.1.0.4:1", fullTok); !called || level != tier.All {
		t.Fatalf("tier 4 user: level=%v called=%v", level, called)
	}

	// A registered user with no tier override keeps the session's tier.
	if level, called, _ := levelSeen(t, NewSessions(0, 0), m, "10.1.0.5:1", plainTok); !called || level != tier.Public {
		t.Fatalf("plain user: level=%v called=%v", level, called)
	}
}
