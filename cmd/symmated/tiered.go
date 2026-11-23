package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/symmate/internal/tier"
)

// tierData is the owner's disclosure data. Everything here is shown only to
// visitors at tier 3 or above, and only in the prompt built for that caller.
type tierData struct {
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	BarSchedule string   `json:"bar_schedule"`
	Projects    []string `json:"projects"`
}

// loadTierData reads <hostDir>/tier-data.json. A missing file means no tier 3
// disclosure data, which is not an error.
func loadTierData(hostDir string) (*tierData, error) {
	data, err := os.ReadFile(filepath.Join(hostDir, "tier-data.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var d tierData
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parse tier-data.json: %w", err)
	}
	return &d, nil
}

// tierGrants returns the firewall grants for a tier. Tiers are cumulative.
// Name is public at tier 2. Email and phone open at tier 3. Address stays closed
// unless the owner assigns tier 4.
func tierGrants(l tier.Level) []firewall.Grant {
	deny := func(class string) firewall.Grant { return firewall.Grant{DataClass: class, Permission: "deny"} }
	allow := func(class string) firewall.Grant { return firewall.Grant{DataClass: class, Permission: "allow"} }
	switch {
	case l >= tier.All:
		return []firewall.Grant{allow("name"), allow("email"), allow("phone"), allow("address")}
	case l >= tier.Schedule:
		return []firewall.Grant{allow("name"), allow("email"), allow("phone"), deny("address")}
	case l >= tier.Public:
		return []firewall.Grant{allow("name"), deny("email"), deny("phone"), deny("address")}
	default:
		return []firewall.Grant{deny("name"), deny("email"), deny("phone"), deny("address")}
	}
}

// tierPromptAddendum returns the disclosure block for a caller. It is added only
// to the prompt for that caller, and is empty below tier 3.
func tierPromptAddendum(l tier.Level, d *tierData) string {
	if l < tier.Schedule || d == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\n## Disclosure for this caller\n")
	b.WriteString("This caller has been verified as someone who knows Loke. You may share these details when they ask:\n")
	if d.Email != "" {
		fmt.Fprintf(&b, "- Email: %s\n", d.Email)
	}
	if d.Phone != "" {
		fmt.Fprintf(&b, "- Phone: %s\n", d.Phone)
	}
	if d.BarSchedule != "" {
		fmt.Fprintf(&b, "- Bar schedule: %s\n", d.BarSchedule)
	}
	if len(d.Projects) > 0 {
		fmt.Fprintf(&b, "- Recent projects: %s\n", strings.Join(d.Projects, "; "))
	}
	b.WriteString("- Calendar availability: not connected yet. Say you can't see Loke's calendar right now.\n")
	b.WriteString("Share only what is listed here. Do not reveal anything else about Loke's schedule.\n")
	return b.String()
}

// sentenceEnd returns the index just past the first sentence boundary in s, or -1.
// A newline always ends a sentence. . ! ? end one when followed by whitespace.
// The end of the text is not a boundary, because more tokens may follow.
// A dot inside an email, phone or number is not followed by a space, so it is skipped.
func sentenceEnd(s string) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\n' {
			return i + 1
		}
		if c != '.' && c != '!' && c != '?' {
			continue
		}
		if i+1 >= len(s) {
			return -1
		}
		next := s[i+1]
		if next == ' ' || next == '\t' || next == '\r' || next == '\n' {
			return i + 1
		}
	}
	return -1
}

// filterStream forwards tokens from in to out, one complete sentence at a time,
// after running each sentence through the firewall with the caller's grants.
// Partial sentences are held back, so a value split across tokens is still seen whole.
// It returns when in closes or ctx ends.
func filterStream(ctx context.Context, fw *firewall.Firewall, grants []firewall.Grant, in <-chan string, out chan<- string) {
	var buf strings.Builder
	send := func(text string) bool {
		if text == "" {
			return true
		}
		filtered, _ := fw.Filter(ctx, text, grants)
		select {
		case <-ctx.Done():
			return false
		case out <- filtered:
			return true
		}
	}
	for token := range in {
		buf.WriteString(token)
		for {
			s := buf.String()
			end := sentenceEnd(s)
			if end < 0 {
				break
			}
			if !send(s[:end]) {
				return
			}
			buf.Reset()
			buf.WriteString(s[end:])
		}
	}
	send(buf.String())
}
