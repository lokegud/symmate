package tier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeBirthdayAcceptsVariants(t *testing.T) {
	cases := []string{
		"June 2nd", "June 2", "june 2nd", "2 June", "Jun 2", "2 Jun",
		"6/2", "06/02", "06-02", "6-2", "2026-06-02", "June 2, 1990",
		" June   2nd ",
	}
	for _, c := range cases {
		got, ok := NormalizeBirthday(c)
		if !ok || got != "06-02" {
			t.Errorf("NormalizeBirthday(%q) = %q, %v; want 06-02", c, got, ok)
		}
	}
}

func TestNormalizeBirthdayRejects(t *testing.T) {
	for _, c := range []string{"", "not a date", "13/01", "February 30", "June", "2"} {
		if got, ok := NormalizeBirthday(c); ok {
			t.Errorf("NormalizeBirthday(%q) = %q, want rejection", c, got)
		}
	}
}

func TestNormalizePhone(t *testing.T) {
	for _, c := range []string{"707-408-3439", "(707) 408 3439", "7074083439", "+1 707.408.3439", "1-707-408-3439"} {
		got, ok := NormalizePhone(c)
		if !ok || got != "7074083439" {
			t.Errorf("NormalizePhone(%q) = %q, %v; want 7074083439", c, got, ok)
		}
	}
	if _, ok := NormalizePhone("707-408"); ok {
		t.Error("short phone should be rejected")
	}
}

func TestNormalizeEmail(t *testing.T) {
	got, ok := NormalizeEmail("  Loke@JustaDestination.com ")
	if !ok || got != "loke@justadestination.com" {
		t.Errorf("NormalizeEmail = %q, %v", got, ok)
	}
	if _, ok := NormalizeEmail("not an email"); ok {
		t.Error("non-email should be rejected")
	}
}

func newTestVerifier(t *testing.T) *Verifier {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tier-verify.json")
	body := `{"tier3_verify":{"match":"all","facts":{` +
		`"birthday_md":"` + HashFact("06-02") + `",` +
		`"phone_digits":"` + HashFact("7074083439") + `",` +
		`"email_lower":"` + HashFact("loke@justadestination.com") + `"}}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	v, err := LoadVerifier(path)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestVerifyRequiresAllThree(t *testing.T) {
	v := newTestVerifier(t)
	good := Facts{Birthday: "June 2nd", Phone: "707-408-3439", Email: "loke@justadestination.com"}
	if !v.Verify(good) {
		t.Fatal("all three correct should verify")
	}
	wrongOne := []Facts{
		{Birthday: "June 3rd", Phone: "707-408-3439", Email: "loke@justadestination.com"},
		{Birthday: "June 2nd", Phone: "707-408-0000", Email: "loke@justadestination.com"},
		{Birthday: "June 2nd", Phone: "707-408-3439", Email: "someone@example.com"},
	}
	for _, f := range wrongOne {
		if v.Verify(f) {
			t.Errorf("one wrong fact must fail: %+v", f)
		}
	}
	if v.Verify(Facts{}) {
		t.Error("empty facts must fail")
	}
}

func TestVerifyNilVerifierFails(t *testing.T) {
	var v *Verifier
	if v.Verify(Facts{Birthday: "June 2nd", Phone: "7074083439", Email: "loke@justadestination.com"}) {
		t.Error("nil verifier must never verify")
	}
}

func TestLoadVerifierMissingFile(t *testing.T) {
	if _, err := LoadVerifier(filepath.Join(t.TempDir(), "nope.json")); err != ErrNotConfigured {
		t.Errorf("missing file: got %v, want ErrNotConfigured", err)
	}
}

func TestLevelOrdering(t *testing.T) {
	if !(Blocked < Public && Public < Schedule && Schedule < All) {
		t.Error("tiers must be ordered for cumulative checks")
	}
	if DefaultAnonymous != Public {
		t.Error("anonymous visitors must default to public")
	}
}
