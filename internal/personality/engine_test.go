package personality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractSectionA(t *testing.T) {
	input := `preamble

## §A. Universal Symmate Constitution

` + "```" + `
You are a Symmate.
Your loyalty is to the host.
` + "```" + `

## §B. Vertical Defaults Template

` + "```" + `
vertical stuff
` + "```"

	got := extractSectionA(input)
	if !strings.Contains(got, "You are a Symmate") {
		t.Errorf("expected §A content, got: %q", got)
	}
	if strings.Contains(got, "vertical stuff") {
		t.Errorf("§B content should not appear in §A extraction, got: %q", got)
	}
}

func TestExtractSectionA_NoMarker(t *testing.T) {
	input := "plain text without section marker"
	got := extractSectionA(input)
	if got != input {
		t.Errorf("expected passthrough, got: %q", got)
	}
}

func TestComposeFromParts(t *testing.T) {
	got := ComposeFromParts("§A content", "§B vertical", "§C context")
	if !strings.Contains(got, "§A content") {
		t.Error("missing §A")
	}
	if !strings.Contains(got, "§B vertical") {
		t.Error("missing §B")
	}
	if !strings.Contains(got, "§C context") {
		t.Error("missing §C")
	}
	// Order: §A first
	aIdx := strings.Index(got, "§A")
	bIdx := strings.Index(got, "§B")
	cIdx := strings.Index(got, "§C")
	if !(aIdx < bIdx && bIdx < cIdx) {
		t.Errorf("wrong order: A=%d B=%d C=%d", aIdx, bIdx, cIdx)
	}
}

func TestComposeFromParts_NoVertical(t *testing.T) {
	got := ComposeFromParts("§A content", "", "§C context")
	if strings.Contains(got, "§B") {
		t.Error("empty vertical should not appear")
	}
	if !strings.Contains(got, "§A content") || !strings.Contains(got, "§C context") {
		t.Error("missing required sections")
	}
}

func TestComposeWith(t *testing.T) {
	// Build a minimal constitution file
	constitutionContent := `preamble

## §A. Universal Symmate Constitution

` + "```" + `
You are a test Symmate.
` + "```" + `

## §B. Vertical Defaults Template
`
	hostDir := t.TempDir()
	constitutionFile := filepath.Join(hostDir, "constitution.md")
	if err := os.WriteFile(constitutionFile, []byte(constitutionContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(hostDir, "context.md"), []byte("# Business Context\nName: Test Co"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ComposeWith(hostDir, constitutionFile)
	if err != nil {
		t.Fatalf("ComposeWith error: %v", err)
	}
	if !strings.Contains(got, "You are a test Symmate") {
		t.Errorf("missing §A content, got: %q", got)
	}
	if !strings.Contains(got, "Test Co") {
		t.Errorf("missing §C content, got: %q", got)
	}
}

func TestComposeWith_WithVertical(t *testing.T) {
	constitutionContent := `## §A. Universal Symmate Constitution

` + "```" + `
Constitution text.
` + "```"

	hostDir := t.TempDir()
	constitutionFile := filepath.Join(hostDir, "constitution.md")
	os.WriteFile(constitutionFile, []byte(constitutionContent), 0644)
	os.WriteFile(filepath.Join(hostDir, "vertical.md"), []byte("# Hospitality vertical"), 0644)
	os.WriteFile(filepath.Join(hostDir, "context.md"), []byte("# Context"), 0644)

	got, err := ComposeWith(hostDir, constitutionFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "Hospitality vertical") {
		t.Errorf("missing §B content, got: %q", got)
	}
}

func TestComposeWith_MissingContext(t *testing.T) {
	constitutionContent := `## §A. Universal Symmate Constitution

` + "```" + `
Constitution text.
` + "```"

	hostDir := t.TempDir()
	constitutionFile := filepath.Join(hostDir, "constitution.md")
	os.WriteFile(constitutionFile, []byte(constitutionContent), 0644)
	// No context.md

	_, err := ComposeWith(hostDir, constitutionFile)
	if err == nil {
		t.Error("expected error for missing context.md")
	}
}
