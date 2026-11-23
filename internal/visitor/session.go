package visitor

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/loke/symmate/internal/tier"
)

const (
	defaultRatePerMin = 60
	defaultTokenCap   = 5000
	gcInterval        = 5 * time.Minute
	sessionIdleMax    = 30 * time.Minute
	historyLimit      = 60   // 30 turns × 2 messages; trimmed further by historyTokenCap
	historyTokenCap   = 8000 // approximate token budget for history (chars/4)
)

// Session holds per-IP rate limit, token usage, conversation state, and recognition state.
type Session struct {
	limiter        *rate.Limiter
	tokensUsed     int
	tokenCap       int
	mu             sync.Mutex
	lastSeen       time.Time
	conversationID string
	recognition    RecognitionState
	tier           tier.Level // 0 means unset, which resolves to DefaultAnonymous
	verifyFails    int
}

// Tier returns the session's access tier, defaulting to tier.DefaultAnonymous.
func (s *Session) Tier() tier.Level {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tier == 0 {
		return tier.DefaultAnonymous
	}
	return s.tier
}

// SetTier records a tier granted to this session.
func (s *Session) SetTier(l tier.Level) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tier = l
}

// verifyLockoutAfter is how many failed verification attempts a session gets
// before it is locked out of verification for its lifetime.
const verifyLockoutAfter = 5

// VerifyLocked reports whether this session has used up its verification attempts.
func (s *Session) VerifyLocked() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.verifyFails >= verifyLockoutAfter
}

// RecordVerifyFailure counts a failed attempt.
func (s *Session) RecordVerifyFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verifyFails++
}

// Allow checks the rate limiter without consuming a token.
func (s *Session) Allow() bool {
	return s.limiter.Allow()
}

// AddTokens records n tokens of usage.
// Returns false if the session would exceed its cap.
func (s *Session) AddTokens(n int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokensUsed+n > s.tokenCap {
		return false
	}
	s.tokensUsed += n
	return true
}

// ConversationID returns the conversation ID for this session, generating one
// on first call. The caller is responsible for creating the DB row.
func (s *Session) ConversationID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conversationID == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			panic("visitor: rand failed: " + err.Error())
		}
		s.conversationID = "conv_" + hex.EncodeToString(b)
	}
	return s.conversationID
}

// AddPhraseMatch records a successful phrase authentication for this session.
func (s *Session) AddPhraseMatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recognition.AddPhrase()
}

// AddDeviceMatch records a known-device confirmation for this session.
func (s *Session) AddDeviceMatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recognition.AddDevice()
}

// AddEpisodicMatch records an episodic probe outcome for this session.
// confirmed=true means the visitor correctly recalled a real anchor.
// confirmed=false means the visitor accepted a fabricated memory (impersonation flag).
func (s *Session) AddEpisodicMatch(confirmed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recognition.AddEpisodic(confirmed)
}

// IsOwner returns true if recognition confidence exceeds the owner threshold.
func (s *Session) IsOwner() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recognition.IsOwner()
}

// RecognitionLevel returns the human-readable confidence tier for this session.
func (s *Session) RecognitionLevel() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recognition.ConfidenceLevel()
}

// Sessions manages per-IP visitor sessions.
type Sessions struct {
	mu         sync.Mutex
	sessions   map[string]*Session
	ratePerMin int
	tokenCap   int
	verifier   *tier.Verifier // nil means tier 3 verification is off
	meetings   *tier.MeetingStore
}

// SetMeetingStore enables tier 3 meeting requests. Nil disables them.
func (s *Sessions) SetMeetingStore(m *tier.MeetingStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.meetings = m
}

// MeetingStore returns the meeting store, or nil when requests are off.
func (s *Sessions) MeetingStore() *tier.MeetingStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.meetings
}

// SetVerifier enables tier 3 verification with the given facts. Nil disables it.
func (s *Sessions) SetVerifier(v *tier.Verifier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verifier = v
}

// Verifier returns the configured verifier, or nil when verification is off.
func (s *Sessions) Verifier() *tier.Verifier {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.verifier
}

// NewSessions creates a session store with the given limits.
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
func (s *Sessions) Get(addr string) *Session {
	// Key by host only. The client port changes with every new connection, so keying
	// on ip:port would drop the session, and its tier, whenever the browser reconnects.
	ip := addr
	if host, _, err := net.SplitHostPort(addr); err == nil {
		ip = host
	}
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

// gc removes sessions idle longer than sessionIdleMax.
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
