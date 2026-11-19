package tool

import (
	"context"
	"encoding/json"
	"sort"
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
	// The first segment of each tag ("engineering" in "engineering/pr") is the tool's group.
	Tags []string
	// Public tools are available to every caller, including anonymous visitors.
	// Non-public tools are restricted: only the owner, or a registered user who has
	// been granted every group the tool belongs to, can see and call them.
	Public bool
}

// AllGroups is the Access.Groups wildcard granting every restricted group (owner).
const AllGroups = "*"

// Access describes which tools a caller may use.
type Access struct {
	// Tags is the satellite-wide tag-prefix allowlist (owner.json tool_tags).
	// Empty means no satellite restriction.
	Tags []string
	// Groups lists the restricted tool groups this caller has been granted.
	// AllGroups ("*") grants every group. Empty means public tools only.
	Groups []string
}

// OwnerAccess is full access, subject only to the satellite tag allowlist.
func OwnerAccess(tags []string) Access { return Access{Tags: tags, Groups: []string{AllGroups}} }

// PublicAccess is the access level of an anonymous visitor.
func PublicAccess(tags []string) Access { return Access{Tags: tags} }

// Groups returns the distinct groups (first tag segment) a tool belongs to.
func (t Tool) Groups() []string {
	seen := map[string]bool{}
	var out []string
	for _, tag := range t.Tags {
		g, _, _ := strings.Cut(tag, "/")
		if g != "" && !seen[g] {
			seen[g] = true
			out = append(out, g)
		}
	}
	return out
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

// permits reports whether a caller with access a may use t.
// A restricted tool spanning several groups (e.g. sales + cs) requires all of them,
// so granting one department never leaks another department's data.
func (t Tool) permits(a Access) bool {
	if !t.allowed(a.Tags) {
		return false
	}
	if t.Public {
		return true
	}
	granted := map[string]bool{}
	for _, g := range a.Groups {
		if g == AllGroups {
			return true
		}
		granted[g] = true
	}
	groups := t.Groups()
	if len(groups) == 0 {
		return false // restricted tool with no group: owner only
	}
	for _, g := range groups {
		if !granted[g] {
			return false
		}
	}
	return true
}

// Registry holds the tools available to a Symmate instance.
type Registry struct {
	tools []Tool
}

// Register adds a tool to the registry.
func (reg *Registry) Register(t Tool) {
	reg.tools = append(reg.tools, t)
}

// DefsFor returns the ToolDefs visible to a caller with access a.
func (reg *Registry) DefsFor(a Access) []llm.ToolDef {
	var defs []llm.ToolDef
	for _, t := range reg.tools {
		if t.permits(a) {
			defs = append(defs, t.Def)
		}
	}
	return defs
}

// SkillsFor returns the tool-use skill block for a caller with access a.
func (reg *Registry) SkillsFor(a Access) string {
	var visible []Tool
	for _, t := range reg.tools {
		if t.permits(a) {
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

// ExecutorFor returns a ToolExecutor that refuses tools outside access a.
// The check is repeated at execution time so a model cannot call a tool it was not shown.
func (reg *Registry) ExecutorFor(a Access) llm.ToolExecutor {
	return func(ctx context.Context, name string, args json.RawMessage) string {
		for _, t := range reg.tools {
			if t.Def.Name == name && t.permits(a) {
				return t.Execute(ctx, args)
			}
		}
		return "error: unknown tool " + name
	}
}

// RestrictedGroups returns the sorted set of groups used by registered non-public tools.
// These are the toggles the owner can grant to registered users.
func (reg *Registry) RestrictedGroups() []string {
	seen := map[string]bool{}
	for _, t := range reg.tools {
		if t.Public {
			continue
		}
		for _, g := range t.Groups() {
			seen[g] = true
		}
	}
	out := make([]string, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}

// Len returns the number of registered tools.
func (reg *Registry) Len() int { return len(reg.tools) }
