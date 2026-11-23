package visitor

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/loke/symmate/internal/tier"
)

func verifierForTest(t *testing.T) *tier.Verifier {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tier-verify.json")
	body := `{"tier3_verify":{"match":"all","facts":{` +
		`"birthday_md":"` + tier.HashFact("06-02") + `",` +
		`"phone_digits":"` + tier.HashFact("7074083439") + `",` +
		`"email_lower":"` + tier.HashFact("loke@justadestination.com") + `"}}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	v, err := tier.LoadVerifier(path)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func postVerify(h http.Handler, addr, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewBufferString(body))
	req.RemoteAddr = addr
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestVerifyRaisesTierOnFullMatch(t *testing.T) {
	sessions := NewSessions(0, 0)
	sessions.SetVerifier(verifierForTest(t))
	h := verifyHandler(sessions)

	good := `{"birthday":"June 2nd","phone":"707-408-3439","email":"Loke@JustaDestination.com"}`
	rec := postVerify(h, "10.0.0.1:1234", good)
	if rec.Code != http.StatusOK {
		t.Fatalf("full match: status %d, want 200", rec.Code)
	}
	if got := sessions.Get("10.0.0.1:1234").Tier(); got != tier.Schedule {
		t.Fatalf("tier after verify = %v, want schedule", got)
	}
}

func TestVerifyWrongAnswerDoesNotRevealWhichFact(t *testing.T) {
	sessions := NewSessions(0, 0)
	sessions.SetVerifier(verifierForTest(t))
	h := verifyHandler(sessions)

	rec := postVerify(h, "10.0.0.2:1", `{"birthday":"June 2nd","phone":"707-408-0000","email":"loke@justadestination.com"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("partial match: status %d, want 403", rec.Code)
	}
	if got := sessions.Get("10.0.0.2:1").Tier(); got != tier.DefaultAnonymous {
		t.Fatalf("failed verify raised tier to %v", got)
	}
}

func TestVerifyLocksOutAfterRepeatedFailures(t *testing.T) {
	sessions := NewSessions(0, 0)
	sessions.SetVerifier(verifierForTest(t))
	h := verifyHandler(sessions)
	bad := `{"birthday":"June 3rd","phone":"7074083439","email":"loke@justadestination.com"}`
	for i := 0; i < verifyLockoutAfter; i++ {
		postVerify(h, "10.0.0.3:1", bad)
	}
	good := `{"birthday":"June 2nd","phone":"707-408-3439","email":"loke@justadestination.com"}`
	if rec := postVerify(h, "10.0.0.3:1", good); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("after lockout: status %d, want 429 even with correct facts", rec.Code)
	}
}

func TestVerifyUnavailableWithoutFacts(t *testing.T) {
	sessions := NewSessions(0, 0)
	h := verifyHandler(sessions)
	rec := postVerify(h, "10.0.0.4:1", `{"birthday":"June 2nd","phone":"7074083439","email":"loke@justadestination.com"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("no verifier: status %d, want 503", rec.Code)
	}
}

func TestMeetingRequestNeedsTierThree(t *testing.T) {
	sessions := NewSessions(0, 0)
	sessions.SetMeetingStore(tier.NewMeetingStore(filepath.Join(t.TempDir(), "m.json")))
	h := meetingRequestHandler(sessions, sessions.MeetingStore())
	body := `{"when":"Tuesday 3pm","note":"about the project"}`

	if rec := postVerify(h, "10.0.1.1:1", body); rec.Code != http.StatusForbidden {
		t.Fatalf("public visitor: status %d, want 403", rec.Code)
	}

	sessions.Get("10.0.1.2:1").SetTier(tier.Schedule)
	if rec := postVerify(h, "10.0.1.2:1", body); rec.Code != http.StatusAccepted {
		t.Fatalf("tier 3 visitor: status %d, want 202", rec.Code)
	}
}

func TestSessionSurvivesNewConnectionFromSameIP(t *testing.T) {
	sessions := NewSessions(0, 0)
	sessions.SetVerifier(verifierForTest(t))
	h := verifyHandler(sessions)
	good := `{"birthday":"June 2nd","phone":"707-408-3439","email":"loke@justadestination.com"}`
	if rec := postVerify(h, "10.0.2.1:50001", good); rec.Code != http.StatusOK {
		t.Fatalf("verify: status %d", rec.Code)
	}
	// A later request arrives on a new connection, so a new client port.
	if got := sessions.Get("10.0.2.1:50977").Tier(); got != tier.Schedule {
		t.Fatalf("tier lost on new connection: got %v, want schedule", got)
	}
}
