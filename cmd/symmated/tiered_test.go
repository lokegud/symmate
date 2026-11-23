package main

import (
	"context"
	"strings"
	"testing"

	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/symmate/internal/tier"
)

func TestSentenceEndBoundaries(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"Hello there. Next", len("Hello there.")},
		{"Is it? yes", len("Is it?")},
		{"Email me at loke@justadestination.com today", -1},
		{"call 707.408.3439 now", -1},
		{"ends with a dot.", -1}, // wait for the next token before deciding
		{"line one\nline two", len("line one\n")},
		{"", -1},
	}
	for _, c := range cases {
		if got := sentenceEnd(c.in); got != c.want {
			t.Errorf("sentenceEnd(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func runFilter(t *testing.T, l tier.Level, tokens []string) string {
	t.Helper()
	in := make(chan string, len(tokens))
	for _, tok := range tokens {
		in <- tok
	}
	close(in)
	out := make(chan string, 64)
	filterStream(context.Background(), firewall.New(nil), tierGrants(l), in, out)
	close(out)
	var b strings.Builder
	for s := range out {
		b.WriteString(s)
	}
	return b.String()
}

func TestFilterRedactsEmailBelowTierThree(t *testing.T) {
	// The address is split across tokens, so it must be judged as a whole sentence.
	got := runFilter(t, tier.Public, []string{"Reach me at loke@justa", "destination.com. Thanks."})
	if strings.Contains(got, "loke@justadestination.com") {
		t.Fatalf("public tier leaked the email: %q", got)
	}
}

func TestFilterAllowsEmailAtTierThree(t *testing.T) {
	got := runFilter(t, tier.Schedule, []string{"Reach me at loke@justa", "destination.com. Thanks."})
	if !strings.Contains(got, "loke@justadestination.com") {
		t.Fatalf("tier 3 should see the email: %q", got)
	}
}

func TestFilterFlushesTrailingFragment(t *testing.T) {
	got := runFilter(t, tier.Public, []string{"no full stop at the end"})
	if got != "no full stop at the end" {
		t.Fatalf("trailing fragment lost: %q", got)
	}
}

func TestTierGrantsAreCumulative(t *testing.T) {
	perm := func(l tier.Level, class string) string {
		for _, g := range tierGrants(l) {
			if g.DataClass == class {
				return g.Permission
			}
		}
		return ""
	}
	if perm(tier.Public, "email") != "deny" || perm(tier.Public, "name") != "allow" {
		t.Error("public tier: name open, email closed")
	}
	if perm(tier.Schedule, "email") != "allow" || perm(tier.Schedule, "address") != "deny" {
		t.Error("schedule tier: email open, address closed")
	}
	if perm(tier.All, "address") != "allow" {
		t.Error("all tier: address open")
	}
}

func TestAddendumOnlyAtTierThreeAndUp(t *testing.T) {
	d := &tierData{Email: "loke@justadestination.com", Phone: "7074083439", BarSchedule: "Sat and Sun 2-10pm at @shadyoak"}
	if got := tierPromptAddendum(tier.Public, d); got != "" {
		t.Fatalf("public tier got a disclosure block: %q", got)
	}
	got := tierPromptAddendum(tier.Schedule, d)
	if !strings.Contains(got, "Bar schedule") || !strings.Contains(got, "loke@justadestination.com") {
		t.Fatalf("tier 3 block missing data: %q", got)
	}
	if tierPromptAddendum(tier.Schedule, nil) != "" {
		t.Fatal("nil tier data must add nothing")
	}
}
