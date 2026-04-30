package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/loke/symmate/internal/aeo"
)

// ToolDefinition describes a tool for tools/list responses.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolResult is the content returned by tools/call.
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent is a single content block in a tool result.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var emptySchema = json.RawMessage(`{"type":"object","properties":{},"required":[]}`)

// ToolRegistry holds the v0 tool set and executes them.
type ToolRegistry struct {
	hc   *aeo.HostContext
	defs []ToolDefinition
}

func newToolRegistry(hc *aeo.HostContext) *ToolRegistry {
	defs := []ToolDefinition{
		{
			Name:        "get_profile",
			Description: "Returns the Symmate's public AEO profile (name, description, capabilities, endpoints).",
			InputSchema: emptySchema,
		},
		{
			Name:        "get_catalog",
			Description: "Returns the host's product or service catalog.",
			InputSchema: emptySchema,
		},
		{
			Name:        "get_hours",
			Description: "Returns the host's availability or business hours.",
			InputSchema: emptySchema,
		},
		{
			Name:        "get_policies",
			Description: "Returns the host's contact, collaboration, and service policies.",
			InputSchema: emptySchema,
		},
	}
	return &ToolRegistry{hc: hc, defs: defs}
}

// Definitions returns the list of tools for tools/list.
func (r *ToolRegistry) Definitions() []ToolDefinition {
	return r.defs
}

// Call executes the named tool and returns its result.
func (r *ToolRegistry) Call(name string, _ json.RawMessage) (*ToolResult, error) {
	switch name {
	case "get_profile":
		data, err := json.MarshalIndent(r.hc.Profile, "", "  ")
		if err != nil {
			return nil, err
		}
		return textResult(string(data)), nil

	case "get_catalog":
		if r.hc.Catalog == "" {
			return textResult("No catalog information available."), nil
		}
		return textResult(r.hc.Catalog), nil

	case "get_hours":
		if r.hc.Hours == "" {
			return textResult("No availability information published."), nil
		}
		return textResult(r.hc.Hours), nil

	case "get_policies":
		if r.hc.Policies == "" {
			return textResult("No policy information published."), nil
		}
		return textResult(r.hc.Policies), nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func textResult(text string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{Type: "text", Text: text}},
	}
}
