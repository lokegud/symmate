package llm

import "strings"

// Known tags.
const (
	TagDefault      = "default"
	TagLight        = "light"
	TagHardReason   = "hard-reasoning"
	TagPrivacyBound = "privacy-bound"
)

// privacyKeywords trigger local Hermes routing — these touch host-private data.
var privacyKeywords = []string{
	"email", "e-mail", "inbox",
	"calendar", "meeting", "appointment", "schedule",
	"message", "sms", "text message", "whatsapp",
	"contact list", "phone book",
	"password", "credential", "secret",
	"address", "location", "home",
}

// Classify returns a routing tag based on content.
// Callers may pass an explicit tag; this function is for when no tag is specified.
func Classify(text string) string {
	lower := strings.ToLower(text)
	for _, kw := range privacyKeywords {
		if strings.Contains(lower, kw) {
			return TagPrivacyBound
		}
	}
	return TagDefault
}
