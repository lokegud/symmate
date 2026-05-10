package entity

import (
	"testing"
)

func TestLookupCompany(t *testing.T) {
	r := NewRegistry()

	// exact match
	c, ok := r.LookupCompany("Acme Corp")
	if !ok {
		t.Fatal("expected to find Acme Corp")
	}
	if c.Name != "Acme Corp" {
		t.Fatalf("got name %q", c.Name)
	}

	// prefix match
	c, ok = r.LookupCompany("Glob")
	if !ok {
		t.Fatal("expected prefix match for 'Glob'")
	}
	if c.Domain != "globex.io" {
		t.Fatalf("got domain %q", c.Domain)
	}

	// missing
	_, ok = r.LookupCompany("NoSuchCo")
	if ok {
		t.Fatal("expected not found for 'NoSuchCo'")
	}
}

func TestContactsFor(t *testing.T) {
	r := NewRegistry()
	contacts := r.ContactsFor("Acme Corp")
	if len(contacts) != 2 {
		t.Fatalf("expected 2 contacts for Acme Corp, got %d", len(contacts))
	}
}

func TestProfile(t *testing.T) {
	r := NewRegistry()

	p, err := r.Profile("Acme Corp")
	if err != nil {
		t.Fatal(err)
	}
	if p.Company.Name != "Acme Corp" {
		t.Fatalf("company name mismatch: %q", p.Company.Name)
	}
	if len(p.Contacts) == 0 {
		t.Fatal("expected non-empty contacts")
	}
	if len(p.DeptSignals) != 4 {
		t.Fatalf("expected 4 dept signals, got %d", len(p.DeptSignals))
	}
	depts := map[string]bool{}
	for _, s := range p.DeptSignals {
		depts[s.Department] = true
	}
	for _, want := range []string{"sales", "cs", "engineering", "marketing"} {
		if !depts[want] {
			t.Fatalf("missing dept signal for %q", want)
		}
	}
}
