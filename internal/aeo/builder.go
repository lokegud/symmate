package aeo

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BuildConfig provides runtime parameters for building a Profile.
type BuildConfig struct {
	Slug        string
	HostDir     string
	HostType    HostType
	VisitorBase string // e.g. "http://localhost:7842"
	A2ABase     string // e.g. "http://localhost:7843"
	MCPBase     string // e.g. "http://localhost:7844"
}

// HostContext holds the full parsed context from context.md.
// Profile is the AEO-facing identity; the extended fields feed MCP tools.
type HostContext struct {
	Profile  *Profile
	Catalog  string
	Hours    string
	Policies string
}

// Build reads <hostDir>/context.md and builds a HostContext.
// If context.md is absent, returns minimal defaults — not an error.
func Build(cfg BuildConfig) (*HostContext, error) {
	p := &Profile{
		Slug:         cfg.Slug,
		HostType:     cfg.HostType,
		SchemaVer:    "0.1",
		UpdatedAt:    time.Now(),
		Name:         cfg.Slug,
		Description:  "A Symmate representative.",
		Capabilities: []string{"chat", "profile"},
		Endpoints: Endpoints{
			Visitor: cfg.VisitorBase,
			A2A:     cfg.A2ABase,
			MCP:     cfg.MCPBase,
		},
	}
	hc := &HostContext{Profile: p}

	ctxPath := filepath.Join(cfg.HostDir, "context.md")
	f, err := os.Open(ctxPath)
	if err != nil {
		return hc, nil
	}
	defer f.Close()

	parseContextMD(f, hc)
	return hc, nil
}

// parseContextMD extracts profile fields and section bodies from context.md.
// Supports:
//   - "key: value" front-matter lines (before any ## section)
//   - "## section-name" headers followed by body text
func parseContextMD(r io.Reader, hc *HostContext) {
	scanner := bufio.NewScanner(r)
	p := hc.Profile

	currentSection := ""
	var sectionLines []string

	flush := func() {
		body := strings.TrimSpace(strings.Join(sectionLines, "\n"))
		switch currentSection {
		case "catalog", "products", "services":
			hc.Catalog = body
		case "hours", "availability":
			hc.Hours = body
		case "policies", "policy":
			hc.Policies = body
		}
		sectionLines = nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			flush()
			currentSection = strings.ToLower(strings.TrimSpace(trimmed[3:]))
			continue
		}

		if currentSection != "" {
			sectionLines = append(sectionLines, line)
			continue
		}

		// Front-matter key-value parsing (before any ## section)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if v, ok := fieldValue(trimmed, "name"); ok {
			p.Name = v
		} else if v, ok := fieldValue(trimmed, "description"); ok {
			p.Description = v
		} else if v, ok := fieldValue(trimmed, "location"); ok {
			p.Contact.Location = v
		} else if v, ok := fieldValue(trimmed, "website"); ok {
			p.Contact.Website = v
		} else if v, ok := fieldValue(trimmed, "email"); ok {
			p.Contact.PublicEmail = v
		} else if v, ok := fieldValue(trimmed, "host_type"); ok {
			p.HostType = HostType(v)
		}
	}
	flush()

	// Expand capabilities based on what context provides.
	if hc.Catalog != "" {
		p.Capabilities = append(p.Capabilities, "catalog")
	}
	if hc.Hours != "" {
		p.Capabilities = append(p.Capabilities, "hours")
	}
	if hc.Policies != "" {
		p.Capabilities = append(p.Capabilities, "policies")
	}
}

// fieldValue returns the value from "key: value" (case-insensitive key match).
func fieldValue(line, key string) (string, bool) {
	prefix := strings.ToLower(key) + ":"
	if !strings.HasPrefix(strings.ToLower(line), prefix) {
		return "", false
	}
	return strings.TrimSpace(line[len(prefix):]), true
}
