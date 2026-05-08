package a2a

import (
	sdk "github.com/a2aproject/a2a-go/v2/a2a"
)

// BuildCard constructs the public A2A AgentCard for a Symmate.
//
// baseURL is the externally reachable URL of this A2A endpoint
// (the JSON-RPC POST endpoint that handles message/send and friends).
// name and description come from the parsed AEO profile.
//
// Authentication is "none" by default — Symmates run as public-facing
// representatives and rely on the diffuser/Forcefield, not auth, to filter
// adversarial callers. That decision is up to the host; if the host has
// configured a real auth scheme, override SecuritySchemes after this returns.
func BuildCard(baseURL, name, description string) *sdk.AgentCard {
	if name == "" {
		name = "Symmate"
	}
	if description == "" {
		description = "An AI representative."
	}

	return &sdk.AgentCard{
		Name:        name,
		Description: description,
		Version:     "0.1.0",
		SupportedInterfaces: []*sdk.AgentInterface{
			sdk.NewAgentInterface(baseURL, sdk.TransportProtocolJSONRPC),
		},
		Capabilities: sdk.AgentCapabilities{
			Streaming:         false,
			PushNotifications: false,
		},
		DefaultInputModes:  []string{"text/plain"},
		DefaultOutputModes: []string{"text/plain"},
		Skills: []sdk.AgentSkill{
			{
				ID:          "general-inquiry",
				Name:        "General Inquiry",
				Description: "Answer questions on behalf of the host.",
				Tags:        []string{"general"},
				Examples:    []string{"hours of operation", "is X available", "what does the host do"},
			},
			{
				ID:          "product-discovery",
				Name:        "Product Discovery",
				Description: "Find products or services matching given criteria.",
				Tags:        []string{"shopping", "discovery"},
			},
		},
	}
}
