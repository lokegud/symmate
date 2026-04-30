package aeo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuild_NoContext(t *testing.T) {
	hc, err := Build(BuildConfig{
		Slug:        "test",
		HostDir:     t.TempDir(),
		HostType:    HostTypeMoxPer,
		VisitorBase: "http://localhost:7842",
	})
	if err != nil {
		t.Fatal(err)
	}
	if hc.Profile.Slug != "test" {
		t.Errorf("slug = %q, want %q", hc.Profile.Slug, "test")
	}
	if hc.Profile.Name != "test" {
		t.Errorf("name = %q, want default slug", hc.Profile.Name)
	}
}

func TestBuild_WithContext(t *testing.T) {
	dir := t.TempDir()
	ctx := `name: Alice
description: A test Symmate.
location: San Francisco, CA
website: https://example.com
host_type: MoxBus

## hours
Mon-Fri 9am-5pm

## catalog
Widget A, Widget B

## policies
No refunds after 30 days.
`
	if err := os.WriteFile(filepath.Join(dir, "context.md"), []byte(ctx), 0644); err != nil {
		t.Fatal(err)
	}

	hc, err := Build(BuildConfig{
		Slug:        "alice",
		HostDir:     dir,
		HostType:    HostTypeMoxPer,
		VisitorBase: "http://localhost:7842",
		MCPBase:     "http://localhost:7844",
	})
	if err != nil {
		t.Fatal(err)
	}

	p := hc.Profile
	if p.Name != "Alice" {
		t.Errorf("name = %q, want Alice", p.Name)
	}
	if p.Description != "A test Symmate." {
		t.Errorf("description = %q", p.Description)
	}
	if p.Contact.Location != "San Francisco, CA" {
		t.Errorf("location = %q", p.Contact.Location)
	}
	if p.HostType != HostTypeMoxBus {
		t.Errorf("host_type = %q, want MoxBus", p.HostType)
	}
	if hc.Hours == "" {
		t.Error("hours should be populated")
	}
	if hc.Catalog == "" {
		t.Error("catalog should be populated")
	}
	if hc.Policies == "" {
		t.Error("policies should be populated")
	}

	// Capabilities should include catalog, hours, policies
	caps := make(map[string]bool)
	for _, c := range p.Capabilities {
		caps[c] = true
	}
	for _, want := range []string{"chat", "profile", "catalog", "hours", "policies"} {
		if !caps[want] {
			t.Errorf("missing capability %q", want)
		}
	}
}

func TestFieldValue(t *testing.T) {
	tests := []struct {
		line, key, want string
		ok              bool
	}{
		{"name: Alice", "name", "Alice", true},
		{"Name: Bob", "name", "Bob", true},
		{"description: A long desc.", "description", "A long desc.", true},
		{"other: stuff", "name", "", false},
		{"", "name", "", false},
	}
	for _, tt := range tests {
		got, ok := fieldValue(tt.line, tt.key)
		if ok != tt.ok || got != tt.want {
			t.Errorf("fieldValue(%q, %q) = %q, %v; want %q, %v", tt.line, tt.key, got, ok, tt.want, tt.ok)
		}
	}
}
