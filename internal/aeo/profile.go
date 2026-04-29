package aeo

import "time"

// HostType identifies the class of Symmate.
type HostType string

const (
	HostTypeMoxPer HostType = "MoxPer"
	HostTypeMoxBus HostType = "MoxBus"
	HostTypeMoxArt HostType = "MoxArt"
	HostTypeMoxNeu HostType = "MoxNeu"
)

// Contact holds public contact fields for a Symmate host.
type Contact struct {
	PublicEmail string `json:"public_email,omitempty"`
	Website     string `json:"website,omitempty"`
	Location    string `json:"location,omitempty"` // coarse only — never precise
}

// Endpoints lists the surface URLs for this Symmate.
type Endpoints struct {
	Visitor string `json:"visitor"`
	A2A     string `json:"a2a,omitempty"`
	MCP     string `json:"mcp,omitempty"`
}

// Profile is the machine-readable AEO identity of a Symmate.
// Consumed by /id, /.well-known/agent-card.json, and MCP get_profile.
type Profile struct {
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	HostType     HostType  `json:"host_type"`
	Description  string    `json:"description"`
	Capabilities []string  `json:"capabilities"`
	Contact      Contact   `json:"contact"`
	Verified     bool      `json:"verified"`
	Endpoints    Endpoints `json:"endpoints"`
	SchemaVer    string    `json:"schema_version"`
	UpdatedAt    time.Time `json:"updated_at"`
}
