package tool

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/loke/symmate/internal/llm"
)

// Tool pairs a definition (sent to the LLM) with an executor (run locally).
type Tool struct {
	Def     llm.ToolDef
	Execute func(ctx context.Context, args json.RawMessage) string
	// Skill is appended to the system prompt to teach the LLM how to use this tool.
	Skill string
	// Tags are hierarchical labels (e.g. "net/fetch", "time") used to filter which
	// tools a given satellite may access. An empty slice means the tool is always included.
	Tags []string
}

// allowed reports whether the tool should be visible given allowedTags.
// If allowedTags is empty, all tools pass. Otherwise a tool passes if any of its
// tags equals an allowedTag or starts with allowedTag+"/".
func (t Tool) allowed(allowedTags []string) bool {
	if len(allowedTags) == 0 {
		return true
	}
	for _, allow := range allowedTags {
		for _, tag := range t.Tags {
			if tag == allow || strings.HasPrefix(tag, allow+"/") {
				return true
			}
		}
	}
	return false
}

// Registry holds the tools available to a Symmate instance.
type Registry struct {
	tools []Tool
}

// Register adds a tool to the registry.
func (reg *Registry) Register(t Tool) {
	reg.tools = append(reg.tools, t)
}

// Defs returns the ToolDef slice to send to the LLM.
// Pass allowedTags to restrict to matching tools; omit for all tools.
func (reg *Registry) Defs(allowedTags ...string) []llm.ToolDef {
	var defs []llm.ToolDef
	for _, t := range reg.tools {
		if t.allowed(allowedTags) {
			defs = append(defs, t.Def)
		}
	}
	return defs
}

// Skills returns the full tool-use skill block for the system prompt.
// Pass allowedTags to restrict to matching tools; omit for all tools.
func (reg *Registry) Skills(allowedTags ...string) string {
	var visible []Tool
	for _, t := range reg.tools {
		if t.allowed(allowedTags) {
			visible = append(visible, t)
		}
	}
	if len(visible) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(`

---
## Tools

You have access to tools. Use them — don't guess when a tool can answer.

Activate a tool whenever the question requires:
- The current date, time, or anything time-sensitive
- A live URL, webpage, or external resource
- A web search for current events, facts, prices, or anything that may have changed
- Any information you are uncertain about or that post-dates your training

Do not refuse or hedge with "I don't have real-time access" — call the tool instead.
If a tool call fails, tell the user what went wrong.

Available tools:`)
	for _, t := range visible {
		if t.Skill != "" {
			sb.WriteString("\n")
			sb.WriteString(t.Skill)
		}
	}
	return sb.String()
}

// Executor returns a ToolExecutor that dispatches to the matching tool.
// Pass allowedTags to restrict to matching tools; omit for all tools.
func (reg *Registry) Executor(allowedTags ...string) llm.ToolExecutor {
	return func(ctx context.Context, name string, args json.RawMessage) string {
		for _, t := range reg.tools {
			if t.Def.Name == name && t.allowed(allowedTags) {
				return t.Execute(ctx, args)
			}
		}
		return "error: unknown tool " + name
	}
}

// Len returns the number of registered tools.
func (reg *Registry) Len() int { return len(reg.tools) }
