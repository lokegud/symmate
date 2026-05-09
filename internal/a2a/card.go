package a2a

import (
	"encoding/json"
	"net/http"

	sdk "github.com/a2aproject/a2a-go/v2/a2a"
)

// Card extends the SDK AgentCard with a top-level "url" field required by the
// current A2A spec. SDK v2.2.1 omits this field; we add it here.
//
// url is the canonical discovery URL for this agent — typically the visitor
// website root (e.g. https://loke.symmate.com), not the JSON-RPC endpoint.
// External clients should prefer SupportedInterfaces for the RPC endpoint.
type Card struct {
	*sdk.AgentCard
	URL string `json:"url"`
}

// BuildCard constructs the public A2A AgentCard for a Symmate.
//
// a2aEndpointURL is the externally reachable JSON-RPC endpoint.
// discoveryURL is the canonical agent website URL (added as top-level "url").
// name and description come from the AEO profile.
func BuildCard(a2aEndpointURL, discoveryURL, name, description string) *Card {
	if name == "" {
		name = "Symmate"
	}
	if description == "" {
		description = "An AI representative."
	}

	return &Card{
		AgentCard: &sdk.AgentCard{
			Name:        name,
			Description: description,
			Version:     "0.1.0",
			SupportedInterfaces: []*sdk.AgentInterface{
				sdk.NewAgentInterface(a2aEndpointURL, sdk.TransportProtocolJSONRPC),
			},
			// Streaming: true — executor yields events over the iter.Seq2 channel;
			// supports incremental delivery per A2A spec §streaming.
			Capabilities: sdk.AgentCapabilities{
				Streaming:         true,
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
		},
		URL: discoveryURL,
	}
}

// staticCardHandler returns an http.Handler that serves the card as JSON.
// We own the serialization so the top-level "url" field is included.
func staticCardHandler(card *Card) http.Handler {
	payload, _ := json.Marshal(card)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(payload)
	})
}
