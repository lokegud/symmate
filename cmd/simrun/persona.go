package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Persona is the full parsed persona YAML.
type Persona struct {
	Name              string         `yaml:"name"`
	Description       string         `yaml:"description"`
	Demographic       string         `yaml:"demographic"`
	CommunicationStyle []string      `yaml:"communication_style"`
	Vocabulary        yaml.Node      `yaml:"vocabulary"`
	Triggers          []string       `yaml:"triggers"`
	Goals             []string       `yaml:"goals"`
	GoalsInConflict   []string       `yaml:"goals_in_conflict"`
	WillNotDo         []string       `yaml:"will_not_do"`
	SampleUtterances  []string       `yaml:"sample_utterances"`
	Tactics           []yaml.Node    `yaml:"tactics"`
}

// LoadPersona loads a persona YAML by base name from the given directory.
func LoadPersona(dir, name string) (*Persona, error) {
	path := filepath.Join(dir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load persona %q: %w", name, err)
	}
	var p Persona
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("persona %q yaml: %w", name, err)
	}
	return &p, nil
}

// SystemPrompt builds the system prompt used to instantiate a persona agent.
// role is the conversation role (customer, attacker, party_a, etc.).
// scenarioGoal is the goal field from the scenario, scoped to this persona.
func (p *Persona) SystemPrompt(role, scenarioGoal string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "You are playing the role of \"%s\" in a simulated conversation test.\n\n", p.Name)

	if p.Description != "" {
		b.WriteString("## Who you are\n")
		b.WriteString(strings.TrimSpace(p.Description))
		b.WriteString("\n\n")
	}

	if p.Demographic != "" {
		fmt.Fprintf(&b, "**Demographic:** %s\n\n", strings.TrimSpace(p.Demographic))
	}

	if len(p.CommunicationStyle) > 0 {
		b.WriteString("## Communication style\n")
		for _, s := range p.CommunicationStyle {
			fmt.Fprintf(&b, "- %s\n", s)
		}
		b.WriteString("\n")
	}

	goals := p.Goals
	if len(goals) == 0 {
		goals = p.GoalsInConflict
	}
	if len(goals) > 0 {
		b.WriteString("## Your goals in this conversation\n")
		for _, g := range goals {
			fmt.Fprintf(&b, "- %s\n", g)
		}
		b.WriteString("\n")
	}

	if scenarioGoal != "" {
		b.WriteString("## Scenario objective\n")
		b.WriteString(strings.TrimSpace(scenarioGoal))
		b.WriteString("\n\n")
	}

	if len(p.WillNotDo) > 0 {
		b.WriteString("## You will NOT do the following\n")
		for _, w := range p.WillNotDo {
			fmt.Fprintf(&b, "- %s\n", w)
		}
		b.WriteString("\n")
	}

	if len(p.SampleUtterances) > 0 {
		b.WriteString("## Example messages in your style\n")
		for _, u := range p.SampleUtterances {
			fmt.Fprintf(&b, "- \"%s\"\n", u)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Instructions\n")
	b.WriteString("- Stay in character at all times. Do not break character or reveal you are an AI.\n")
	b.WriteString("- Send one short message per turn (1-3 sentences max unless the style calls for more).\n")
	b.WriteString("- If you have achieved your goal or the conversation has reached a natural end, ")
	b.WriteString("output exactly: [END]\n")
	b.WriteString("- Output only your message — no preamble, no labels, no stage directions.\n")

	return b.String()
}
