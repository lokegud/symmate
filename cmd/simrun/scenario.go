package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Scenario is the parsed frontmatter of a scenario markdown file.
type Scenario struct {
	ID                  string          `yaml:"id"`
	Type                string          `yaml:"type"` // "business" | "h2h"
	HostBusiness        string          `yaml:"host_business"`
	Personas            []PersonaRole   `yaml:"personas"`
	Goal                string          `yaml:"goal"`
	ExpectedBehaviors   []string        `yaml:"expected_behaviors"`
	ProhibitedBehaviors []string        `yaml:"prohibited_behaviors"`
	JudgeCriteria       yaml.Node       `yaml:"judge_criteria"`
	Severity            string          `yaml:"severity"`

	// Body is the full markdown body after the frontmatter (setup, opening message, etc.)
	Body string `yaml:"-"`
}

// PersonaRole pairs a conversation role (e.g. "attacker", "customer", "party_a") with a persona file name.
type PersonaRole struct {
	Role    string `yaml:"role"`
	Persona string `yaml:"persona"` // base name, no extension
}

// JudgeCriteriaText renders judge_criteria as a flat string for the judge prompt.
func (s *Scenario) JudgeCriteriaText() string {
	var b strings.Builder
	renderNode(&b, &s.JudgeCriteria, 0)
	return strings.TrimSpace(b.String())
}

func renderNode(b *strings.Builder, n *yaml.Node, depth int) {
	if n == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	switch n.Kind {
	case yaml.DocumentNode:
		for _, c := range n.Content {
			renderNode(b, c, depth)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			key := n.Content[i].Value
			val := n.Content[i+1]
			if val.Kind == yaml.ScalarNode {
				fmt.Fprintf(b, "%s- %s: %s\n", indent, key, val.Value)
			} else {
				fmt.Fprintf(b, "%s- %s:\n", indent, key)
				renderNode(b, val, depth+1)
			}
		}
	case yaml.SequenceNode:
		for _, c := range n.Content {
			renderNode(b, c, depth)
		}
	case yaml.ScalarNode:
		if n.Value != "" {
			fmt.Fprintf(b, "%s- %s\n", indent, n.Value)
		}
	}
}

// LoadScenario parses a scenario markdown file (YAML frontmatter + markdown body).
func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load scenario %q: %w", path, err)
	}

	// Strip leading newlines / BOM
	data = bytes.TrimLeft(data, "\n\r ")

	if !bytes.HasPrefix(data, []byte("---")) {
		return nil, fmt.Errorf("scenario %q has no YAML frontmatter", path)
	}

	// Find closing ---
	rest := data[3:]
	idx := bytes.Index(rest, []byte("\n---"))
	if idx == -1 {
		return nil, fmt.Errorf("scenario %q: unclosed frontmatter", path)
	}

	frontmatter := rest[:idx]
	body := strings.TrimSpace(string(rest[idx+4:]))
	// skip optional newline right after closing ---
	body = strings.TrimPrefix(body, "\n")

	var sc Scenario
	if err := yaml.Unmarshal(frontmatter, &sc); err != nil {
		return nil, fmt.Errorf("scenario %q yaml: %w", path, err)
	}
	sc.Body = body
	return &sc, nil
}
