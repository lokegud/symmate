package owner

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/loke/symmate/internal/visitor"
)

// userView is a registered user as returned by the API (never includes the token hash).
type userView struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Enabled   bool       `json:"enabled"`
	Groups    []string   `json:"groups"`
	CreatedAt time.Time  `json:"created_at"`
	LastSeen  *time.Time `json:"last_seen,omitempty"` // nil until first seen
}

func toView(u visitor.RegisteredUser) userView {
	g := u.Groups
	if g == nil {
		g = []string{}
	}
	v := userView{ID: u.ID, Name: u.Name, Enabled: u.Enabled, Groups: g, CreatedAt: u.CreatedAt}
	if !u.LastSeen.IsZero() {
		t := u.LastSeen
		v.LastSeen = &t
	}
	return v
}

const maxUserNameLen = 80

// validateGroups rejects groups that no registered tool uses, so typos don't
// silently grant nothing (or grant something added later).
func (s *Server) validateGroups(groups []string) error {
	known := map[string]bool{}
	for _, g := range s.toolGroups {
		known[g] = true
	}
	for _, g := range groups {
		g = strings.ToLower(strings.TrimSpace(g))
		if g == "" {
			continue
		}
		if !known[g] {
			return errors.New("unknown tool group: " + g)
		}
	}
	return nil
}

// GET /users → master switch, available groups, and all users.
func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.ownerMgr.Load()
	if err != nil {
		jsonErr(w, "load failed", http.StatusInternalServerError)
		return
	}
	users := make([]userView, 0, len(cfg.Users))
	for _, u := range cfg.Users {
		users = append(users, toView(u))
	}
	groups := s.toolGroups
	if groups == nil {
		groups = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled":          cfg.UsersEnabled,
		"available_groups": groups,
		"users":            users,
	})
}

// PUT /users/settings {"enabled": bool} — master switch for registered-user access.
func (s *Server) putUserSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Enabled == nil {
		jsonErr(w, `body must be {"enabled": true|false}`, http.StatusBadRequest)
		return
	}
	if err := s.ownerMgr.Update(func(c *visitor.OwnerConfig) error {
		c.UsersEnabled = *req.Enabled
		return nil
	}); err != nil {
		jsonErr(w, "save failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /users {"name": "...", "groups": [...]} → the user plus its token (shown once).
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string   `json:"name"`
		Groups []string `json:"groups"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid json", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > maxUserNameLen {
		jsonErr(w, "name is required (max 80 characters)", http.StatusBadRequest)
		return
	}
	if err := s.validateGroups(req.Groups); err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	var created visitor.RegisteredUser
	var token string
	if err := s.ownerMgr.Update(func(c *visitor.OwnerConfig) error {
		created, token = c.AddUser(name, req.Groups)
		return nil
	}); err != nil {
		jsonErr(w, "save failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": toView(created), "token": token})
}

// PATCH /users/{id} {"name"?, "enabled"?, "groups"?}
func (s *Server) patchUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name    *string   `json:"name"`
		Enabled *bool     `json:"enabled"`
		Groups  *[]string `json:"groups"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" || len(n) > maxUserNameLen {
			jsonErr(w, "name must be 1-80 characters", http.StatusBadRequest)
			return
		}
	}
	var groups []string
	if req.Groups != nil {
		groups = *req.Groups
		if err := s.validateGroups(groups); err != nil {
			jsonErr(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	var updated visitor.RegisteredUser
	err := s.ownerMgr.Update(func(c *visitor.OwnerConfig) error {
		if err := c.UpdateUser(id, req.Name, req.Enabled, groups, req.Groups != nil); err != nil {
			return err
		}
		updated = *c.User(id)
		return nil
	})
	if errors.Is(err, visitor.ErrUserNotFound) {
		jsonErr(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonErr(w, "save failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toView(updated))
}

// DELETE /users/{id}
func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := s.ownerMgr.Update(func(c *visitor.OwnerConfig) error { return c.RemoveUser(id) })
	if errors.Is(err, visitor.ErrUserNotFound) {
		jsonErr(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonErr(w, "save failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /users/{id}/token → {"token": "..."}; the old token stops working.
func (s *Server) rotateUserToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var token string
	err := s.ownerMgr.Update(func(c *visitor.OwnerConfig) error {
		t, err := c.RotateUserToken(id)
		token = t
		return err
	})
	if errors.Is(err, visitor.ErrUserNotFound) {
		jsonErr(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonErr(w, "save failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
