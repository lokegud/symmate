package tier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMeetingStoreAddListDecide(t *testing.T) {
	s := NewMeetingStore(filepath.Join(t.TempDir(), "meeting-requests.json"))
	req, err := s.Add("Sam", "Tuesday afternoon", "about the avatar")
	if err != nil {
		t.Fatal(err)
	}
	if req.Status != MeetingPending || req.ID == "" {
		t.Fatalf("new request = %+v, want pending with id", req)
	}
	all, err := s.List()
	if err != nil || len(all) != 1 {
		t.Fatalf("list = %d, %v; want 1", len(all), err)
	}
	ok, err := s.Decide(req.ID, MeetingApproved)
	if err != nil || !ok {
		t.Fatalf("decide = %v, %v", ok, err)
	}
	all, _ = s.List()
	if all[0].Status != MeetingApproved {
		t.Fatalf("status after approve = %q", all[0].Status)
	}
	if ok, _ := s.Decide("missing", MeetingDeclined); ok {
		t.Fatal("unknown id must not match")
	}
	if _, err := s.Decide(req.ID, "booked"); err == nil {
		t.Fatal("only approved or declined are valid")
	}
}

func TestMeetingStoreRejectsEmptyAndLong(t *testing.T) {
	s := NewMeetingStore(filepath.Join(t.TempDir(), "m.json"))
	if _, err := s.Add("Sam", "", ""); err == nil {
		t.Error("request with no time or note must be rejected")
	}
	if _, err := s.Add("Sam", strings.Repeat("x", 201), ""); err == nil {
		t.Error("overlong time must be rejected")
	}
}

func TestMeetingStoreFileIsOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m.json")
	s := NewMeetingStore(path)
	if _, err := s.Add("", "Friday", ""); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("file mode = %v, want 0600", info.Mode().Perm())
	}
}
