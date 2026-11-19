package aeo

import (
	"os"
	"path/filepath"
	"strings"
)

// UpdateContextMD updates a single field in <hostDir>/context.md.
// For section keys (hours, catalog, policies and their aliases), it replaces
// the section body under the matching ## header.
// For front-matter keys (name, description, location, website, email, host_type),
// it replaces or inserts the "key: value" line before the first ## header.
func UpdateContextMD(hostDir, key, value string) error {
	path := filepath.Join(hostDir, "context.md")
	current, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	updated := applyContextMDUpdate(string(current), strings.ToLower(strings.TrimSpace(key)), value)
	return os.WriteFile(path, []byte(updated), 0644)
}

var sectionAliases = map[string]bool{
	"hours": true, "availability": true,
	"catalog": true, "products": true, "services": true,
	"policies": true, "policy": true,
}

var frontMatterKeys = map[string]bool{
	"name": true, "description": true, "location": true,
	"website": true, "email": true, "host_type": true,
}

func applyContextMDUpdate(content, key, value string) string {
	if sectionAliases[key] {
		return updateSection(content, key, value)
	}
	if frontMatterKeys[key] {
		return updateFrontMatter(content, key, value)
	}
	return content
}

// updateFrontMatter replaces an existing "key: ..." line or inserts before the
// first ## header. Preserves ordering and all other lines.
func updateFrontMatter(content, key, value string) string {
	lines := strings.Split(content, "\n")
	newLine := key + ": " + value
	prefix := key + ":"

	// Replace existing
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			break // past front-matter
		}
		if strings.HasPrefix(strings.ToLower(trimmed), prefix) {
			lines[i] = newLine
			return strings.Join(lines, "\n")
		}
	}

	// Insert before first ## header
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "## ") {
			result := make([]string, 0, len(lines)+1)
			result = append(result, lines[:i]...)
			result = append(result, newLine)
			result = append(result, lines[i:]...)
			return strings.Join(result, "\n")
		}
	}

	// No sections — append
	if content == "" {
		return newLine + "\n"
	}
	return strings.TrimRight(content, "\n") + "\n" + newLine + "\n"
}

// updateSection replaces the body of a ## section, or appends a new section.
func updateSection(content, key, value string) string {
	lines := strings.Split(content, "\n")

	// Find the section header matching key (exact or alias)
	headerIdx := -1
	var headerLine string
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			sectionName := strings.ToLower(strings.TrimSpace(trimmed[3:]))
			if sectionName == key {
				headerIdx = i
				headerLine = line
				break
			}
		}
	}

	if headerIdx == -1 {
		// Append new section
		suffix := "\n\n## " + key + "\n" + value
		return strings.TrimRight(content, "\n") + suffix + "\n"
	}

	// Find end of this section (next ## header or EOF)
	endIdx := len(lines)
	for i := headerIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			endIdx = i
			break
		}
	}

	result := make([]string, 0, len(lines))
	result = append(result, lines[:headerIdx]...)
	result = append(result, headerLine)
	result = append(result, value)
	result = append(result, lines[endIdx:]...)
	return strings.Join(result, "\n")
}
