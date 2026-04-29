package firewall

import (
	"regexp"
	"sort"
	"strings"
)

// reZip matches a US zip code within a larger address string.
var reZip = regexp.MustCompile(`\b\d{5}(?:-\d{4})?\b`)

// Rewrite returns text with denied PII spans replaced by redacted equivalents.
// Matches with an "allow" grant are left in place; all others are coarsened.
func Rewrite(text string, matches []PIIMatch, grants []Grant) string {
	grantMap := buildGrantMap(grants)

	type span struct {
		start, end int
		replacement string
	}
	var spans []span

	for _, m := range matches {
		perm := resolvePermission(m.Class, grantMap)
		if perm == "allow" {
			continue
		}
		spans = append(spans, span{
			start:       m.Start,
			end:         m.End,
			replacement: coarseReplacement(m),
		})
	}

	if len(spans) == 0 {
		return text
	}

	// Process right-to-left so byte offsets stay valid after each substitution.
	sort.Slice(spans, func(i, j int) bool { return spans[i].start > spans[j].start })

	for _, s := range spans {
		text = text[:s.start] + s.replacement + text[s.end:]
	}
	return text
}

// coarseReplacement returns the least-disclosure substitute for a PII match.
// Per runtime.md §10: prefer coarse equivalent over blunt [REDACTED] where possible.
func coarseReplacement(m PIIMatch) string {
	switch m.Class {
	case "address":
		if zip := reZip.FindString(m.Value); zip != "" {
			return "[address in " + zip + "]"
		}
		return "[REDACTED address]"
	case "email":
		parts := strings.SplitN(m.Value, "@", 2)
		if len(parts) == 2 {
			return "[email @" + parts[1] + "]"
		}
		return "[REDACTED email]"
	case "phone":
		return "[REDACTED phone]"
	case "name":
		return "[REDACTED name]"
	default:
		return "[REDACTED]"
	}
}

func buildGrantMap(grants []Grant) map[string]string {
	m := make(map[string]string, len(grants))
	for _, g := range grants {
		m[g.DataClass] = g.Permission
	}
	return m
}

func resolvePermission(dataClass string, grantMap map[string]string) string {
	if perm, ok := grantMap[dataClass]; ok {
		return perm
	}
	return "deny" // conservative default — silent leak is never acceptable
}
