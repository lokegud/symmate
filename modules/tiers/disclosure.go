package tier

import (
	"fmt"
	"strings"
)

// Line is one piece of owner data a verified caller may hear.
type Line struct {
	Label string
	Value string
}

// Disclosure is the owner data shown at tier 3 and above. Values are plain text.
type Disclosure struct {
	Lines []Line
	// Notes are extra rules for this data, e.g. "Calendar availability: not connected yet."
	Notes []string
}

// PromptBlock returns the block added to the model prompt for a caller at level l.
// It is empty below tier 3, or when there is no disclosure data.
func PromptBlock(l Level, d *Disclosure) string {
	if l < Schedule || d == nil || (len(d.Lines) == 0 && len(d.Notes) == 0) {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\n## Disclosure for this caller\n")
	b.WriteString("This caller has been verified as someone who knows the owner. You may share these details when they ask:\n")
	for _, ln := range d.Lines {
		if ln.Value != "" {
			fmt.Fprintf(&b, "- %s: %s\n", ln.Label, ln.Value)
		}
	}
	for _, n := range d.Notes {
		fmt.Fprintf(&b, "- %s\n", n)
	}
	b.WriteString("Share only what is listed here. Do not reveal anything else about the owner's schedule.\n")
	return b.String()
}
