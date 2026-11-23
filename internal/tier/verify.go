// Package tier holds visitor access tiers and the code-side verification that
// moves a visitor up a tier. Verification runs here, in code, never in the
// model: the twin's prompt never sees the facts or the rule.
package tier

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Level is a visitor access tier. Tiers are cumulative: each includes the ones below.
type Level int

const (
	// Blocked gets nothing at all.
	Blocked Level = 1
	// Public is the default for anonymous visitors: whatever is already public.
	Public Level = 2
	// Schedule adds free/busy, meeting requests, email, phone, bar schedule and project names.
	Schedule Level = 3
	// All is assigned only by the owner, never by verification.
	All Level = 4
)

// DefaultAnonymous is the tier an unverified visitor gets.
const DefaultAnonymous = Public

// String returns a short label for logs.
func (l Level) String() string {
	switch l {
	case Blocked:
		return "blocked"
	case Public:
		return "public"
	case Schedule:
		return "schedule"
	case All:
		return "all"
	default:
		return fmt.Sprintf("unknown(%d)", int(l))
	}
}

// Facts is what a visitor submits to be verified.
type Facts struct {
	Birthday string `json:"birthday"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// storedFile is the on-disk shape of tier-verify.json. It holds hashes only.
type storedFile struct {
	Tier3Verify struct {
		Match string `json:"match"`
		Facts struct {
			BirthdayMD  string `json:"birthday_md"`
			PhoneDigits string `json:"phone_digits"`
			EmailLower  string `json:"email_lower"`
		} `json:"facts"`
	} `json:"tier3_verify"`
}

// Verifier checks submitted facts against stored hashes.
type Verifier struct {
	birthday string
	phone    string
	email    string
}

// ErrNotConfigured means no verification facts are stored, so nobody can reach tier 3.
var ErrNotConfigured = errors.New("tier verification not configured")

// LoadVerifier reads the stored hashes. A missing file disables verification.
func LoadVerifier(path string) (*Verifier, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotConfigured
		}
		return nil, err
	}
	var f storedFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if f.Tier3Verify.Match != "all" {
		return nil, fmt.Errorf("%s: only match=\"all\" is supported", path)
	}
	v := &Verifier{
		birthday: f.Tier3Verify.Facts.BirthdayMD,
		phone:    f.Tier3Verify.Facts.PhoneDigits,
		email:    f.Tier3Verify.Facts.EmailLower,
	}
	if v.birthday == "" || v.phone == "" || v.email == "" {
		return nil, fmt.Errorf("%s: all three facts must be set", path)
	}
	return v, nil
}

// Verify reports whether every submitted fact matches. Every fact must match:
// one correct answer is not enough. A fact that fails to normalize counts as a miss.
func (v *Verifier) Verify(f Facts) bool {
	if v == nil {
		return false
	}
	bday, okB := NormalizeBirthday(f.Birthday)
	phone, okP := NormalizePhone(f.Phone)
	email, okE := NormalizeEmail(f.Email)
	if !okB || !okP || !okE {
		return false
	}
	// Compare every fact, even after a miss, so timing doesn't reveal which one failed.
	matchB := hashEqual(bday, v.birthday)
	matchP := hashEqual(phone, v.phone)
	matchE := hashEqual(email, v.email)
	return matchB && matchP && matchE
}

func hashEqual(normalized, storedHex string) bool {
	sum := sha256.Sum256([]byte(normalized))
	got := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(got), []byte(storedHex)) == 1
}

// HashFact returns the stored form of a normalized fact. Used by the setup command.
func HashFact(normalized string) string {
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

var (
	ordinalRE = regexp.MustCompile(`(\d)(st|nd|rd|th)\b`)
	spaceRE   = regexp.MustCompile(`[\s,]+`)
	numericRE = regexp.MustCompile(`^(\d{1,2})[/.\-](\d{1,2})$`)
	isoRE     = regexp.MustCompile(`^(\d{4})[/.\-](\d{1,2})[/.\-](\d{1,2})$`)
)

// NormalizeBirthday turns the many ways of writing a month and day into "MM-DD".
// Accepts "June 2nd", "June 2", "2 June", "Jun 2", "6/2", "06-02", "2026-06-02".
// The year is ignored, since the stored birthday has none.
func NormalizeBirthday(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", false
	}
	s = ordinalRE.ReplaceAllString(s, "$1")
	s = strings.TrimSpace(spaceRE.ReplaceAllString(s, " "))

	if m := isoRE.FindStringSubmatch(s); m != nil {
		return monthDay(m[2], m[3])
	}
	if m := numericRE.FindStringSubmatch(s); m != nil {
		return monthDay(m[1], m[2])
	}
	for _, layout := range []string{"January 2", "Jan 2", "2 January", "2 Jan", "January 2 2006", "2 January 2006"} {
		t, err := time.Parse(layout, titleCase(s))
		if err == nil {
			return fmt.Sprintf("%02d-%02d", int(t.Month()), t.Day()), true
		}
	}
	return "", false
}

func monthDay(mStr, dStr string) (string, bool) {
	var m, d int
	if _, err := fmt.Sscanf(mStr, "%d", &m); err != nil {
		return "", false
	}
	if _, err := fmt.Sscanf(dStr, "%d", &d); err != nil {
		return "", false
	}
	// Validate against a leap year so Feb 29 is accepted.
	if _, err := time.Parse("2006-01-02", fmt.Sprintf("2000-%02d-%02d", m, d)); err != nil {
		return "", false
	}
	return fmt.Sprintf("%02d-%02d", m, d), true
}

// titleCase capitalizes words so time.Parse's month names match.
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 && w[0] >= 'a' && w[0] <= 'z' {
			words[i] = string(w[0]-'a'+'A') + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// NormalizePhone keeps only the digits, and drops a leading US country code 1.
func NormalizePhone(s string) (string, bool) {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	if len(d) == 11 && d[0] == '1' {
		d = d[1:]
	}
	if len(d) != 10 {
		return "", false
	}
	return d, true
}

// NormalizeEmail trims and lowercases an address.
func NormalizeEmail(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if !strings.Contains(s, "@") || strings.ContainsAny(s, " \t") {
		return "", false
	}
	return s, true
}

type levelKey struct{}

// WithLevel stores the visitor's tier in ctx. The chat handler sets it per request;
// the brain reads it to decide what context it may use.
func WithLevel(ctx context.Context, l Level) context.Context {
	return context.WithValue(ctx, levelKey{}, l)
}

// LevelFrom returns the tier in ctx, or DefaultAnonymous when none is set.
// Missing means least access, so a code path that forgets to set it gets Public.
func LevelFrom(ctx context.Context) Level {
	if l, ok := ctx.Value(levelKey{}).(Level); ok && l != 0 {
		return l
	}
	return DefaultAnonymous
}
