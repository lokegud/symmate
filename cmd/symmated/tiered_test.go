package main

import (
	"context"
	"strings"
	"testing"

	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/tiers"
)

// Firewall-backed checks: the tiers module is tested on its own with a fake redactor.
func runFirewallFilter(t *testing.T, l tier.Level, tokens []string) string {
	t.Helper()
	in := make(chan string, len(tokens))
	for _, tok := range tokens {
		in <- tok
	}
	close(in)
	out := make(chan string, 64)
	tier.FilterStream(context.Background(), firewallRedactor(firewall.New(nil)), tier.GrantsFor(l), in, out)
	close(out)
	var b strings.Builder
	for s := range out {
		b.WriteString(s)
	}
	return b.String()
}

func TestFirewallRedactsEmailBelowTierThree(t *testing.T) {
	got := runFirewallFilter(t, tier.Public, []string{"Reach me at loke@justa", "destination.com. Thanks."})
	if strings.Contains(got, "loke@justadestination.com") {
		t.Fatalf("public tier leaked the email: %q", got)
	}
}

func TestFirewallAllowsEmailAtTierThree(t *testing.T) {
	got := runFirewallFilter(t, tier.Schedule, []string{"Reach me at loke@justa", "destination.com. Thanks."})
	if !strings.Contains(got, "loke@justadestination.com") {
		t.Fatalf("tier 3 should see the email: %q", got)
	}
}
