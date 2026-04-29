package a2a

import (
	"fmt"
	"os"
	"strings"
)

// BuildCard constructs the Agent Card for a Symmate host.
// baseURL is the externally reachable URL for this A2A endpoint, e.g. "https://loke.parlourcraft.com/a2a".
// hostDir is the Symmate's data directory (reads context.md for name/description if present).
func BuildCard(baseURL, hostDir string) AgentCard {
	name, description := loadNameDesc(hostDir)
	return AgentCard{
		SchemaVersion: "1.0",
		Name:          name,
		Description:   description,
		URL:           baseURL,
		Capabilities: CardCapabilities{
			Streaming:              false,
			PushNotifications:      false,
			StateTransitionHistory: false,
		},
		Skills: []CardSkill{
			{
				ID:          "product-discovery",
				Name:        "Product Discovery",
				Description: "Find products or services matching given criteria.",
				Tags:        []string{"shopping", "discovery"},
			},
			{
				ID:          "general-inquiry",
				Name:        "General Inquiry",
				Description: "Answer questions on behalf of the host.",
				Tags:        []string{"general"},
			},
		},
		Authentication: CardAuth{
			Schemes: []string{"none"},
		},
	}
}

// loadNameDesc reads context.md and extracts the name and first meaningful line as description.
func loadNameDesc(hostDir string) (name, description string) {
	name = "Symmate"
	description = "An AI representative."

	data, err := os.ReadFile(hostDir + "/context.md")
	if err != nil {
		return
	}
	text := string(data)

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// "Name: Foo" or "# Name: Foo"
		if after, ok := cutPrefix(line, "Name:"); ok {
			candidate := strings.TrimSpace(after)
			if candidate != "" {
				name = candidate
				continue
			}
		}
		if after, ok := cutPrefix(line, "# Business identity"); ok {
			_ = after
			continue
		}
		// First non-header, non-Name line → description
		if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "Name:") && description == "An AI representative." {
			description = fmt.Sprintf("Symmate for %s", name)
		}
	}
	return
}

// cutPrefix returns (after, true) if s starts with prefix.
func cutPrefix(s, prefix string) (string, bool) {
	stripped := strings.TrimPrefix(s, "# ")
	if strings.HasPrefix(stripped, prefix) {
		return stripped[len(prefix):], true
	}
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):], true
	}
	return "", false
}
