package mcp

// PolicyRule defines whether a tool call is allowed.
type PolicyRule string

const (
	PolicyAllow PolicyRule = "allow"
	PolicyDeny  PolicyRule = "deny"
)

// CapabilityPolicy maps tool names to their allow/deny rule.
// v0: all public read-only tools are allowed for any caller.
// Owner can tighten this via policy.json (future).
type CapabilityPolicy struct {
	rules map[string]PolicyRule
}

// DefaultPolicy returns the v0 default: all defined tools are allowed.
func DefaultPolicy() *CapabilityPolicy {
	return &CapabilityPolicy{
		rules: map[string]PolicyRule{
			"get_profile": PolicyAllow,
			"get_catalog": PolicyAllow,
			"get_hours":   PolicyAllow,
			"get_policies": PolicyAllow,
		},
	}
}

// Check returns (allowed, reason). reason is non-empty only when denied.
func (p *CapabilityPolicy) Check(toolName string) (bool, string) {
	rule, ok := p.rules[toolName]
	if !ok {
		return false, "unknown tool"
	}
	if rule == PolicyDeny {
		return false, "tool denied by policy"
	}
	return true, ""
}
