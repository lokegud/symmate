package visitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestUserRegistrationLifecycle(t *testing.T) {
	cfg := &OwnerConfig{}
	u, token := cfg.AddUser("  Bob ", []string{"Engineering", "engineering", " ", "*", "sales"})
	if u.Name != "Bob" || strings.Join(u.Groups, ",") != "engineering,sales" {
		t.Fatalf("normalisation failed: %+v", u)
	}
	if !strings.HasPrefix(token, "smu_") || strings.Contains(u.TokenHash, token) {
		t.Fatal("token should be prefixed and stored only as a hash")
	}

	if cfg.AuthenticateUser(token) != nil {
		t.Error("master switch off must block every user")
	}
	cfg.UsersEnabled = true
	if got := cfg.AuthenticateUser(token); got == nil || got.ID != u.ID {
		t.Fatal("valid token rejected")
	}
	if cfg.AuthenticateUser("smu_wrong") != nil || cfg.AuthenticateUser("") != nil {
		t.Error("bad token accepted")
	}

	off := false
	if err := cfg.UpdateUser(u.ID, nil, &off, nil, false); err != nil {
		t.Fatal(err)
	}
	if cfg.AuthenticateUser(token) != nil {
		t.Error("disabled user accepted")
	}
	on := true
	_ = cfg.UpdateUser(u.ID, nil, &on, nil, true) // setGroups with nil clears groups
	if len(cfg.User(u.ID).Groups) != 0 {
		t.Error("groups not cleared")
	}

	newToken, err := cfg.RotateUserToken(u.ID)
	if err != nil || cfg.AuthenticateUser(token) != nil || cfg.AuthenticateUser(newToken) == nil {
		t.Error("rotation should invalidate the old token only")
	}
	if err := cfg.RemoveUser(u.ID); err != nil || cfg.AuthenticateUser(newToken) != nil {
		t.Error("removed user still authenticates")
	}
	if cfg.RemoveUser(u.ID) != ErrUserNotFound {
		t.Error("expected ErrUserNotFound")
	}
}

func TestOwnerManagerUpdatePersistsAndIsAtomic(t *testing.T) {
	m := NewOwnerManager(t.TempDir())
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.Update(func(c *OwnerConfig) error {
				c.AddUser("u", nil)
				return nil
			})
		}()
	}
	wg.Wait()
	cfg, err := m.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Users) != 20 {
		t.Errorf("lost updates: got %d users, want 20", len(cfg.Users))
	}
}

// ctxBrain records the context of the last Chat call.
type ctxBrain struct {
	mu  sync.Mutex
	ctx context.Context
}

func (b *ctxBrain) Chat(ctx context.Context, _ []Message) (<-chan string, error) {
	b.mu.Lock()
	b.ctx = ctx
	b.mu.Unlock()
	ch := make(chan string, 1)
	ch <- "ok"
	close(ch)
	return ch, nil
}

func TestChatHandlerInjectsRegisteredUser(t *testing.T) {
	m := NewOwnerManager(t.TempDir())
	var token string
	_ = m.Update(func(c *OwnerConfig) error {
		c.UsersEnabled = true
		_, token = c.AddUser("Bob", []string{"engineering"})
		return nil
	})

	brain := &ctxBrain{}
	router := NewRouter(testHostContext(), brain, NewSessions(0, 0), nil, nil, m)

	send := func(body string, header string) (UserInfo, bool) {
		req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(body))
		if header != "" {
			req.Header.Set("X-Symmate-User", header)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d: %s", w.Code, w.Body.String())
		}
		brain.mu.Lock()
		defer brain.mu.Unlock()
		return UserFromCtx(brain.ctx)
	}

	if u, ok := send(`{"message":"hi","user_token":"`+token+`"}`, ""); !ok || u.Name != "Bob" || u.Groups[0] != "engineering" {
		t.Errorf("body token: got %+v %v", u, ok)
	}
	if _, ok := send(`{"message":"hi"}`, token); !ok {
		t.Error("header token not accepted")
	}
	if _, ok := send(`{"message":"hi","user_token":"smu_nope"}`, ""); ok {
		t.Error("bad token produced a user")
	}

	_ = m.Update(func(c *OwnerConfig) error { c.UsersEnabled = false; return nil })
	if _, ok := send(`{"message":"hi","user_token":"`+token+`"}`, ""); ok {
		t.Error("master switch off still produced a user")
	}

	cfg, _ := m.Load()
	if cfg.Users[0].LastSeen.IsZero() {
		t.Error("last_seen not recorded")
	}
}
