package firewall

import "regexp"

// PII regex patterns for v0. Flag + ask-owner beats silent leak.
var (
	reEmail = regexp.MustCompile(
		`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`,
	)

	rePhone = regexp.MustCompile(
		`\b(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}\b`,
	)

	// Street address: number + street name + type abbreviation
	reAddress = regexp.MustCompile(
		`\b\d{1,5}\s+[A-Za-z0-9\s]{1,40}(?:Street|St|Avenue|Ave|Road|Rd|Drive|Dr|` +
			`Lane|Ln|Boulevard|Blvd|Way|Court|Ct|Place|Pl|Circle|Cir|Terrace|Ter)\b`,
	)

	// Capitalized full-name-looking strings adjacent to name indicators
	reName = regexp.MustCompile(
		`(?:my name is|I am|I'm|contact|from|owner|signed|for)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)+)`,
	)
)

// Extract returns all PII matches found in text.
func Extract(text string) []PIIMatch {
	var matches []PIIMatch

	for _, loc := range reEmail.FindAllStringIndex(text, -1) {
		matches = append(matches, PIIMatch{
			Class: "email",
			Value: text[loc[0]:loc[1]],
			Start: loc[0],
			End:   loc[1],
		})
	}

	for _, loc := range rePhone.FindAllStringIndex(text, -1) {
		matches = append(matches, PIIMatch{
			Class: "phone",
			Value: text[loc[0]:loc[1]],
			Start: loc[0],
			End:   loc[1],
		})
	}

	for _, loc := range reAddress.FindAllStringIndex(text, -1) {
		matches = append(matches, PIIMatch{
			Class: "address",
			Value: text[loc[0]:loc[1]],
			Start: loc[0],
			End:   loc[1],
		})
	}

	for _, sub := range reName.FindAllStringSubmatchIndex(text, -1) {
		if len(sub) >= 4 {
			matches = append(matches, PIIMatch{
				Class: "name",
				Value: text[sub[2]:sub[3]],
				Start: sub[2],
				End:   sub[3],
			})
		}
	}

	return matches
}
