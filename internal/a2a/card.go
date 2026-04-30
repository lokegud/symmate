package a2a

// BuildCard constructs the Agent Card for a Symmate host.
// baseURL is the externally reachable URL for this A2A endpoint.
// name and description come from the already-parsed AEO profile (avoids re-parsing context.md).
// Pass empty strings to fall back to generic defaults.
func BuildCard(baseURL, name, description string) AgentCard {
	if name == "" {
		name = "Symmate"
	}
	if description == "" {
		description = "An AI representative."
	}
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

