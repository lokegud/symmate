package a2a

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type contextKey int

const senderCardURLKey contextKey = 1

// SygilClient queries the Sygil registry for agent identity + reputation signal.
// nil is a valid value — all lookups return Unknown.
//
// Sygil is agent-first: this client returns SIGNAL, not VERDICT. The caller
// (typically the executor) is responsible for surfacing the signal to the
// agent's reasoning layer (LLM + constitution) which makes the actual
// engage/refuse/throttle decision.
type SygilClient struct {
	baseURL string
	client  *http.Client
}

// NewSygilClient returns a client pointing at baseURL (e.g. "http://localhost:7448").
func NewSygilClient(baseURL string) *SygilClient {
	return &SygilClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 2 * time.Second},
	}
}

// SenderSignal is the structured intel Sygil knows about a sender. Empty/zero
// values are meaningful: Known=false means the sender's card URL has never
// been registered in this Sygil instance.
type SenderSignal struct {
	CardURL       string     `json:"card_url"`
	Fingerprint   string     `json:"fingerprint"`
	Known         bool       `json:"known"`
	Name          string     `json:"name,omitempty"`
	Description   string     `json:"description,omitempty"`
	Skills        []string   `json:"skills,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	RegisteredAt  *time.Time `json:"registered_at,omitempty"`
	LastRefreshed *time.Time `json:"last_refreshed,omitempty"`
	Limited       bool       `json:"limited"`
	LimitedUntil  *time.Time `json:"limited_until,omitempty"`
	FlagCount     int        `json:"flag_count"`
}

// Lookup fetches the full Sygil record for cardURL. On any error, network
// failure, or nil receiver, returns SenderSignal{Known:false} (fail-open:
// no Sygil signal, agent decides with whatever else it has).
func (c *SygilClient) Lookup(ctx context.Context, cardURL string) SenderSignal {
	sig := SenderSignal{
		CardURL:     cardURL,
		Fingerprint: agentFingerprint(cardURL),
	}
	if c == nil || c.baseURL == "" {
		return sig
	}
	url := fmt.Sprintf("%s/agents/%s", c.baseURL, sig.Fingerprint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return sig
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return sig
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return sig
	}
	if resp.StatusCode != http.StatusOK {
		return sig
	}
	var body struct {
		Card *struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			SkillIDs    []string `json:"skill_ids"`
			Tags        []string `json:"tags"`
		} `json:"card"`
		RegisteredAt  *time.Time `json:"registered_at"`
		LastRefreshed *time.Time `json:"last_refreshed"`
		Reputation    *struct {
			Stats *struct {
				FlagCount int `json:"flag_count"`
			} `json:"stats"`
			Limited      bool       `json:"limited"`
			LimitedUntil *time.Time `json:"limited_until"`
		} `json:"reputation"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return sig
	}
	sig.Known = true
	if body.Card != nil {
		sig.Name = body.Card.Name
		sig.Description = body.Card.Description
		sig.Skills = body.Card.SkillIDs
		sig.Tags = body.Card.Tags
	}
	sig.RegisteredAt = body.RegisteredAt
	sig.LastRefreshed = body.LastRefreshed
	if body.Reputation != nil {
		sig.Limited = body.Reputation.Limited
		sig.LimitedUntil = body.Reputation.LimitedUntil
		if body.Reputation.Stats != nil {
			sig.FlagCount = body.Reputation.Stats.FlagCount
		}
	}
	return sig
}

// RenderForLLM formats the signal as a system-message block for the LLM.
// Returns "" when the caller should not inject any intel (e.g. no card URL
// supplied; in that case the conversation is treated as anonymous).
//
// Tone: descriptive, not prescriptive. The agent's constitution decides what
// to do with it. The trailing reminder line tells the LLM this is signal,
// not a directive.
func (s SenderSignal) RenderForLLM() string {
	if s.CardURL == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("[Sender Intel — from Sygil registry]\n")
	fmt.Fprintf(&b, "Card URL: %s\n", s.CardURL)
	fmt.Fprintf(&b, "Fingerprint: %s\n", s.Fingerprint)
	if !s.Known {
		b.WriteString("Status: unknown to Sygil (never registered)\n")
	} else {
		b.WriteString("Status: known\n")
		if s.Name != "" {
			fmt.Fprintf(&b, "Name (claimed in card): %s\n", s.Name)
		}
		if s.Description != "" {
			fmt.Fprintf(&b, "Description: %s\n", s.Description)
		}
		if len(s.Skills) > 0 {
			fmt.Fprintf(&b, "Skills: %s\n", strings.Join(s.Skills, ", "))
		}
		if s.RegisteredAt != nil {
			fmt.Fprintf(&b, "Registered at: %s\n", s.RegisteredAt.UTC().Format(time.RFC3339))
		}
		if s.LastRefreshed != nil {
			fmt.Fprintf(&b, "Card last refreshed: %s\n", s.LastRefreshed.UTC().Format(time.RFC3339))
		}
		fmt.Fprintf(&b, "Flag count (90d window): %d\n", s.FlagCount)
		if s.Limited {
			line := "Reputation: currently LIMITED"
			if s.LimitedUntil != nil {
				line += " until " + s.LimitedUntil.UTC().Format(time.RFC3339)
			}
			b.WriteString(line + "\n")
		} else {
			b.WriteString("Reputation: not currently limited\n")
		}
	}
	b.WriteString("This is signal, not a verdict. Apply your own judgment per your constitution.\n")
	return b.String()
}

// IsLimited is a thin compatibility wrapper. New code should use Lookup.
func (c *SygilClient) IsLimited(ctx context.Context, cardURL string) bool {
	return c.Lookup(ctx, cardURL).Limited
}

func agentFingerprint(cardURL string) string {
	h := sha256.Sum256([]byte(cardURL))
	return hex.EncodeToString(h[:])
}

// WithSenderCardURL injects the caller's agent card URL into the context.
func WithSenderCardURL(ctx context.Context, cardURL string) context.Context {
	return context.WithValue(ctx, senderCardURLKey, cardURL)
}

// SenderCardURL extracts the caller's agent card URL from the context.
func SenderCardURL(ctx context.Context) string {
	v, _ := ctx.Value(senderCardURLKey).(string)
	return v
}
