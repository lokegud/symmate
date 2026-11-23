package tier

import (
	"context"
	"strings"
	"testing"
)

func TestSentenceEndBoundaries(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"Hello there. Next", len("Hello there.")},
		{"Is it? yes", len("Is it?")},
		{"Email me at a@b.com today", -1},
		{"call 707.408.3439 now", -1},
		{"ends with a dot.", -1},
		{"line one\nline two", len("line one\n")},
		{"", -1},
	}
	for _, c := range cases {
		if got := SentenceEnd(c.in); got != c.want {
			t.Errorf("SentenceEnd(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// maskAt replaces any sentence containing "@" when the email grant is denied.
// It stands in for a real PII firewall, so these tests don't depend on one.
func maskAt(_ context.Context, text string, grants []Grant) string {
	for _, g := range grants {
		if g.DataClass == "email" && g.Permission == "deny" && strings.Contains(text, "@") {
			return "[redacted]"
		}
	}
	return text
}

func runFilter(t *testing.T, l Level, tokens []string) string {
	t.Helper()
	in := make(chan string, len(tokens))
	for _, tok := range tokens {
		in <- tok
	}
	close(in)
	out := make(chan string, 64)
	FilterStream(context.Background(), maskAt, GrantsFor(l), in, out)
	close(out)
	var b strings.Builder
	for s := range out {
		b.WriteString(s)
	}
	return b.String()
}

func TestFilterHoldsSplitValueUntilWhole(t *testing.T) {
	got := runFilter(t, Public, []string{"Reach me at loke@justa", "destination.com. Thanks."})
	if strings.Contains(got, "@") {
		t.Fatalf("split email leaked below tier 3: %q", got)
	}
	got = runFilter(t, Schedule, []string{"Reach me at loke@justa", "destination.com. Thanks."})
	if !strings.Contains(got, "loke@justadestination.com") {
		t.Fatalf("tier 3 should see the email: %q", got)
	}
}

func TestFilterFlushesTrailingFragment(t *testing.T) {
	if got := runFilter(t, Public, []string{"no full stop at the end"}); got != "no full stop at the end" {
		t.Fatalf("trailing fragment lost: %q", got)
	}
}

func TestFilterKeepsListLines(t *testing.T) {
	text := "Projects include:\n- One\n- Two\n- Three"
	var toks []string
	for i := 0; i < len(text); i += 5 {
		e := i + 5
		if e > len(text) {
			e = len(text)
		}
		toks = append(toks, text[i:e])
	}
	if got := runFilter(t, Schedule, toks); got != text {
		t.Fatalf("list lost:\nwant %q\n got %q", text, got)
	}
}

func TestGrantsAreCumulative(t *testing.T) {
	perm := func(l Level, class string) string {
		for _, g := range GrantsFor(l) {
			if g.DataClass == class {
				return g.Permission
			}
		}
		return ""
	}
	if perm(Public, "email") != "deny" || perm(Public, "name") != "allow" {
		t.Error("public: name open, email closed")
	}
	if perm(Schedule, "email") != "allow" || perm(Schedule, "address") != "deny" {
		t.Error("schedule: email open, address closed")
	}
	if perm(All, "address") != "allow" {
		t.Error("all: address open")
	}
	if perm(Blocked, "name") != "deny" {
		t.Error("blocked: everything closed")
	}
}

func TestPromptBlockOnlyAtTierThreeAndUp(t *testing.T) {
	d := &Disclosure{
		Lines: []Line{{"Email", "a@b.com"}, {"Bar schedule", "Sat 2-10pm"}},
		Notes: []string{"Calendar: not connected. Say so."},
	}
	if PromptBlock(Public, d) != "" {
		t.Fatal("public tier got a disclosure block")
	}
	got := PromptBlock(Schedule, d)
	if !strings.Contains(got, "Bar schedule: Sat 2-10pm") || !strings.Contains(got, "Calendar: not connected") {
		t.Fatalf("tier 3 block missing data: %q", got)
	}
	if PromptBlock(Schedule, nil) != "" || PromptBlock(Schedule, &Disclosure{}) != "" {
		t.Fatal("empty disclosure must add nothing")
	}
}
