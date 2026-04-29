package visitor

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultRatePerMin = 60
	defaultTokenCap   = 5000
	gcInterval        = 5 * time.Minute
	sessionIdleMax    = 30 * time.Minute
)

// Session holds per-IP rate limit and token usage state.
type Session struct {
	limiter    *rate.Limiter
	tokensUsed int
	tokenCap   int
	mu         sync.Mutex
	lastSeen   time.Time
}

// Allow checks the rate limiter without consuming a token.
// Returns false if the IP is over its request rate.
func (s *Session) Allow() bool {
	return s.limiter.Allow()
}

// AddTokens records n tokens of usage.
// Returns false if the session would exceed its cap, true if OK.
func (s *Session) AddTokens(n int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokensUsed+n > s.tokenCap {
		return false
	}
	s.tokensUsed += n
	return true
}

// Sessions manages per-IP visitor sessions.
type Sessions struct {
	mu         sync.Mutex
	sessions   map[string]*Session
	ratePerMin int
	tokenCap   int
}

// NewSessions creates a session store with the given limits.
// ratePerMin <= 0 uses the default (60). tokenCap <= 0 uses the default (5000).
func NewSessions(ratePerMin, tokenCap int) *Sessions {
	if ratePerMin <= 0 {
		ratePerMin = defaultRatePerMin
	}
	if tokenCap <= 0 {
		tokenCap = defaultTokenCap
	}
	s := &Sessions{
		sessions:   make(map[string]*Session),
		ratePerMin: ratePerMin,
		tokenCap:   tokenCap,
	}
	go s.gc()
	return s
}

// Get returns or creates the session for a given IP.
func (s *Sessions) Get(ip string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[ip]
	if !ok {
		r := rate.Every(time.Minute / time.Duration(s.ratePerMin))
		sess = &Session{
			limiter:  rate.NewLimiter(r, s.ratePerMin),
			tokenCap: s.tokenCap,
			lastSeen: time.Now(),
		}
		s.sessions[ip] = sess
	}
	sess.lastSeen = time.Now()
	return sess
}

// gc removes sessions that have been idle longer than sessionIdleMax.
func (s *Sessions) gc() {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-sessionIdleMax)
		s.mu.Lock()
		for ip, sess := range s.sessions {
			if sess.lastSeen.Before(cutoff) {
				delete(s.sessions, ip)
			}
		}
		s.mu.Unlock()
	}
}
