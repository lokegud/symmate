package visitor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// RegisteredUser is someone other than the owner who has been given an access token
// and a set of tool-group toggles. Only the token's SHA-256 is stored.
type RegisteredUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TokenHash string    `json:"token_hash"`
	Enabled   bool      `json:"enabled"`
	Groups    []string  `json:"groups"` // restricted tool groups granted, e.g. ["engineering"]
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen,omitempty"`
}

// ErrUserNotFound is returned when a user ID does not exist.
var ErrUserNotFound = errors.New("user not found")

// UserInfo is the identity of an authenticated registered user, carried in request context.
type UserInfo struct {
	ID     string
	Name   string
	Groups []string
}

type userCtxKey struct{}

// WithUser injects an authenticated registered user into ctx.
func WithUser(ctx context.Context, u UserInfo) context.Context {
	return context.WithValue(ctx, userCtxKey{}, u)
}

// UserFromCtx returns the registered user in ctx, if any.
func UserFromCtx(ctx context.Context) (UserInfo, bool) {
	u, ok := ctx.Value(userCtxKey{}).(UserInfo)
	return u, ok
}

func newSecret(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// normalizeGroups trims, lowercases, de-duplicates and drops empty group names.
func normalizeGroups(groups []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, g := range groups {
		g = strings.ToLower(strings.TrimSpace(g))
		if g == "" || g == "*" || seen[g] {
			continue
		}
		seen[g] = true
		out = append(out, g)
	}
	return out
}

// AddUser registers a new user and returns it with its plaintext token.
// The token is only available here; afterwards only its hash is kept.
func (c *OwnerConfig) AddUser(name string, groups []string) (RegisteredUser, string) {
	token := "smu_" + newSecret(24)
	u := RegisteredUser{
		ID:        newSecret(6),
		Name:      strings.TrimSpace(name),
		TokenHash: hashToken(token),
		Enabled:   true,
		Groups:    normalizeGroups(groups),
		CreatedAt: time.Now().UTC(),
	}
	c.Users = append(c.Users, u)
	return u, token
}

// User returns a pointer to the user with id, or nil.
func (c *OwnerConfig) User(id string) *RegisteredUser {
	for i := range c.Users {
		if c.Users[i].ID == id {
			return &c.Users[i]
		}
	}
	return nil
}

// UpdateUser applies optional changes to a user. Nil fields are left unchanged.
func (c *OwnerConfig) UpdateUser(id string, name *string, enabled *bool, groups []string, setGroups bool) error {
	u := c.User(id)
	if u == nil {
		return ErrUserNotFound
	}
	if name != nil {
		u.Name = strings.TrimSpace(*name)
	}
	if enabled != nil {
		u.Enabled = *enabled
	}
	if setGroups {
		u.Groups = normalizeGroups(groups)
	}
	return nil
}

// RotateUserToken issues a new token for a user, invalidating the old one.
func (c *OwnerConfig) RotateUserToken(id string) (string, error) {
	u := c.User(id)
	if u == nil {
		return "", ErrUserNotFound
	}
	token := "smu_" + newSecret(24)
	u.TokenHash = hashToken(token)
	return token, nil
}

// RemoveUser deletes a user.
func (c *OwnerConfig) RemoveUser(id string) error {
	for i := range c.Users {
		if c.Users[i].ID == id {
			c.Users = append(c.Users[:i], c.Users[i+1:]...)
			return nil
		}
	}
	return ErrUserNotFound
}

// AuthenticateUser resolves a plaintext token to an active registered user.
// Returns nil when user access is switched off, the token is unknown, or the user is disabled.
func (c *OwnerConfig) AuthenticateUser(token string) *RegisteredUser {
	token = strings.TrimSpace(token)
	if !c.UsersEnabled || token == "" {
		return nil
	}
	h := []byte(hashToken(token))
	var match *RegisteredUser
	for i := range c.Users {
		// Compare every entry in constant time; don't short-circuit on the first hit.
		if subtle.ConstantTimeCompare(h, []byte(c.Users[i].TokenHash)) == 1 {
			match = &c.Users[i]
		}
	}
	if match == nil || !match.Enabled {
		return nil
	}
	return match
}
