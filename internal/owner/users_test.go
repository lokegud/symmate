package owner

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/loke/symmate/internal/visitor"
)

const testPhrase = "correct horse battery staple"

func newTestServer(t *testing.T) (http.Handler, *visitor.OwnerManager, string) {
	t.Helper()
	dir := t.TempDir()
	m := visitor.NewOwnerManager(dir)
	if err := m.Update(func(c *visitor.OwnerConfig) error { return c.Enroll(testPhrase, "") }); err != nil {
		t.Fatal(err)
	}
	h := New(dir, m, nil, nil).WithToolGroups(func(tags []string) []string {
		// "sales" is blocked whenever an allowlist is set without it.
		if len(tags) > 0 && tags[0] == "engineering" {
			return []string{"engineering"}
		}
		return []string{"engineering", "sales"}
	}).Router()

	w := do(t, h, "", http.MethodPost, "/auth/login", `{"phrase":"`+testPhrase+`"}`)
	var out struct{ Token string }
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if out.Token == "" {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	return h, m, out.Token
}

func do(t *testing.T, h http.Handler, token, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestUsersAPI(t *testing.T) {
	h, m, tok := newTestServer(t)

	if w := do(t, h, "", http.MethodGet, "/users", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated list: %d", w.Code)
	}

	w := do(t, h, tok, http.MethodGet, "/users", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"enabled":false`) ||
		!strings.Contains(w.Body.String(), `"available_groups":["engineering","sales"]`) {
		t.Fatalf("initial list: %d %s", w.Code, w.Body.String())
	}

	if w := do(t, h, tok, http.MethodPost, "/users", `{"name":"Bob","groups":["finance"]}`); w.Code != http.StatusBadRequest {
		t.Errorf("unknown group accepted: %d", w.Code)
	}
	if w := do(t, h, tok, http.MethodPost, "/users", `{"name":""}`); w.Code != http.StatusBadRequest {
		t.Errorf("empty name accepted: %d", w.Code)
	}

	w = do(t, h, tok, http.MethodPost, "/users", `{"name":"Bob","groups":["engineering"]}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		User  userView
		Token string
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if created.Token == "" || created.User.ID == "" || strings.Contains(w.Body.String(), "token_hash") || strings.Contains(w.Body.String(), "last_seen") {
		t.Fatalf("create response wrong: %s", w.Body.String())
	}
	id := created.User.ID

	if w := do(t, h, tok, http.MethodPut, "/users/settings", `{"enabled":true}`); w.Code != http.StatusNoContent {
		t.Fatalf("enable: %d", w.Code)
	}
	cfg, _ := m.Load()
	if cfg.AuthenticateUser(created.Token) == nil {
		t.Fatal("created token does not authenticate")
	}

	w = do(t, h, tok, http.MethodPatch, "/users/"+id, `{"groups":["engineering","sales"],"enabled":false}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"groups":["engineering","sales"]`) || !strings.Contains(w.Body.String(), `"enabled":false`) {
		t.Fatalf("patch: %d %s", w.Code, w.Body.String())
	}
	if w := do(t, h, tok, http.MethodPatch, "/users/nope", `{"enabled":true}`); w.Code != http.StatusNotFound {
		t.Errorf("patch missing user: %d", w.Code)
	}

	w = do(t, h, tok, http.MethodPost, "/users/"+id+"/token", "")
	var rot struct{ Token string }
	_ = json.Unmarshal(w.Body.Bytes(), &rot)
	if w.Code != http.StatusOK || rot.Token == "" || rot.Token == created.Token {
		t.Fatalf("rotate: %d %s", w.Code, w.Body.String())
	}

	if w := do(t, h, tok, http.MethodDelete, "/users/"+id, ""); w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", w.Code)
	}
	if w := do(t, h, tok, http.MethodDelete, "/users/"+id, ""); w.Code != http.StatusNotFound {
		t.Errorf("second delete: %d", w.Code)
	}

	// A Policy allowlist that blocks "sales" removes it from the toggles and from validation.
	_ = m.Update(func(c *visitor.OwnerConfig) error { c.ToolTags = []string{"engineering"}; return nil })
	if w := do(t, h, tok, http.MethodGet, "/users", ""); !strings.Contains(w.Body.String(), `"available_groups":["engineering"]`) {
		t.Errorf("blocked group still offered: %s", w.Body.String())
	}
	if w := do(t, h, tok, http.MethodPost, "/users", `{"name":"Cy","groups":["sales"]}`); w.Code != http.StatusBadRequest {
		t.Errorf("blocked group accepted: %d", w.Code)
	}

	// Owner enrolment survives all the user edits (Update must not clobber other fields).
	cfg, _ = m.Load()
	if !cfg.Enrolled || !cfg.CheckPhrase(testPhrase) || !cfg.UsersEnabled {
		t.Error("unrelated owner.json fields were lost")
	}
}
