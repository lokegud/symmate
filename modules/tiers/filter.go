package tier

import (
	"context"
	"strings"
)

// Redactor rewrites text so that only what the grants allow gets through.
// Hosts plug in their PII firewall here.
type Redactor func(ctx context.Context, text string, grants []Grant) string

// SentenceEnd returns the index just past the first sentence boundary in s, or -1.
// A newline always ends a sentence. . ! ? end one when followed by whitespace.
// The end of the text is not a boundary, because more tokens may follow.
// A dot inside an email, phone or number is not followed by a space, so it is skipped.
func SentenceEnd(s string) int {
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

// FilterStream forwards tokens from in to out one complete sentence at a time,
// after running each sentence through redact. Partial sentences are held back, so a
// value split across tokens is still seen whole. It returns when in closes or ctx ends.
func FilterStream(ctx context.Context, redact Redactor, grants []Grant, in <-chan string, out chan<- string) {
	var buf strings.Builder
	send := func(text string) bool {
		if text == "" {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case out <- redact(ctx, text, grants):
			return true
		}
	}
	for token := range in {
		buf.WriteString(token)
		for {
			s := buf.String()
			end := SentenceEnd(s)
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
