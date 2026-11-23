package tier

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Meeting request states. A request is never booked automatically; the owner decides.
const (
	MeetingPending  = "pending"
	MeetingApproved = "approved"
	MeetingDeclined = "declined"
)

// MeetingRequest is a visitor's request for time with the owner.
type MeetingRequest struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	When      string    `json:"when"`
	Note      string    `json:"note"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// MeetingStore keeps requests in one JSON file, written atomically.
type MeetingStore struct {
	mu   sync.Mutex
	path string
}

// NewMeetingStore uses the given file. The file is created on first write.
func NewMeetingStore(path string) *MeetingStore {
	return &MeetingStore{path: path}
}

const (
	maxNameLen = 100
	maxWhenLen = 200
	maxNoteLen = 1000
)

// Add validates and stores a new pending request.
func (s *MeetingStore) Add(name, when, note string) (MeetingRequest, error) {
	name = strings.TrimSpace(name)
	when = strings.TrimSpace(when)
	note = strings.TrimSpace(note)
	if when == "" && note == "" {
		return MeetingRequest{}, errors.New("give a time or a note")
	}
	if len(name) > maxNameLen || len(when) > maxWhenLen || len(note) > maxNoteLen {
		return MeetingRequest{}, errors.New("request is too long")
	}
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return MeetingRequest{}, err
	}
	req := MeetingRequest{
		ID: hex.EncodeToString(b), Name: name, When: when, Note: note,
		Status: MeetingPending, CreatedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.readLocked()
	if err != nil {
		return MeetingRequest{}, err
	}
	all = append(all, req)
	if err := s.writeLocked(all); err != nil {
		return MeetingRequest{}, err
	}
	return req, nil
}

// List returns every request, oldest first.
func (s *MeetingStore) List() ([]MeetingRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

// Decide sets a pending request to approved or declined. It reports whether the id exists.
func (s *MeetingStore) Decide(id, status string) (bool, error) {
	if status != MeetingApproved && status != MeetingDeclined {
		return false, fmt.Errorf("status must be %q or %q", MeetingApproved, MeetingDeclined)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.readLocked()
	if err != nil {
		return false, err
	}
	for i := range all {
		if all[i].ID == id {
			all[i].Status = status
			return true, s.writeLocked(all)
		}
	}
	return false, nil
}

func (s *MeetingStore) readLocked() ([]MeetingRequest, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var all []MeetingRequest
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	return all, nil
}

func (s *MeetingStore) writeLocked(all []MeetingRequest) error {
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".meetings-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}
