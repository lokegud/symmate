package personality

import (
	"fmt"
	"os"
	"strings"
)

// DefaultConstitutionPath is the canonical §A source on this host.
const DefaultConstitutionPath = "/mnt/KB/claude/_protocol/symmate-prompts/symmate-v1.md"

// Compose assembles the three-layer system prompt from hostDir.
// Reads §A from DefaultConstitutionPath, §B from <hostDir>/vertical.md (optional),
// §C from <hostDir>/context.md (required).
func Compose(hostDir string) (string, error) {
	constitution, err := LoadConstitution(DefaultConstitutionPath)
	if err != nil {
		return "", fmt.Errorf("load constitution: %w", err)
	}
	return compose(hostDir, constitution)
}

// ComposeWith is Compose with an explicit constitutionPath (for testing / custom installs).
func ComposeWith(hostDir, constitutionPath string) (string, error) {
	constitution, err := LoadConstitution(constitutionPath)
	if err != nil {
		return "", fmt.Errorf("load constitution: %w", err)
	}
	return compose(hostDir, constitution)
}

// ComposeFromParts assembles directly from pre-loaded strings (used by Loader after caching §A).
func ComposeFromParts(constitution, vertical, context string) string {
	var sb strings.Builder
	sb.WriteString(constitution)
	if vertical != "" {
		sb.WriteString("\n\n")
		sb.WriteString(vertical)
	}
	sb.WriteString("\n\n")
	sb.WriteString(context)
	return sb.String()
}

func compose(hostDir, constitution string) (string, error) {
	vertical, err := readOptional(hostDir + "/vertical.md")
	if err != nil {
		return "", fmt.Errorf("read vertical.md: %w", err)
	}

	context, err := readRequired(hostDir + "/context.md")
	if err != nil {
		return "", fmt.Errorf("read context.md: %w", err)
	}

	return ComposeFromParts(constitution, vertical, context), nil
}

// LoadConstitution reads a symmate-v1.md file and extracts the §A block.
func LoadConstitution(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return extractSectionA(string(data)), nil
}

// extractSectionA pulls the content of the first code block under "## §A".
func extractSectionA(content string) string {
	const marker = "## §A. Universal Symmate Constitution"
	idx := strings.Index(content, marker)
	if idx == -1 {
		// No marker found — return the full content as-is
		return strings.TrimSpace(content)
	}
	tail := content[idx+len(marker):]

	// Find opening ```
	open := strings.Index(tail, "```")
	if open == -1 {
		return strings.TrimSpace(tail)
	}
	inner := tail[open+3:]
	// Skip newline after opening fence
	if len(inner) > 0 && inner[0] == '\n' {
		inner = inner[1:]
	}

	// Find closing ```
	close := strings.Index(inner, "\n```")
	if close == -1 {
		return strings.TrimSpace(inner)
	}
	return strings.TrimSpace(inner[:close])
}

func readOptional(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func readRequired(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
