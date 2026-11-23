package visitor

import (
	"testing"
)

func TestSetUserTierOnlyZeroOrFour(t *testing.T) {
	c := &OwnerConfig{Users: []RegisteredUser{{ID: "u1", Name: "Sam", Enabled: true}}}
	if err := c.SetUserTier("u1", 4); err != nil {
		t.Fatalf("tier 4: %v", err)
	}
	if c.User("u1").Tier != 4 {
		t.Fatal("tier 4 not stored")
	}
	if err := c.SetUserTier("u1", 3); err == nil {
		t.Fatal("tier 3 must not be grantable by hand")
	}
	if err := c.SetUserTier("u1", 0); err != nil || c.User("u1").Tier != 0 {
		t.Fatal("tier 0 should clear the override")
	}
	if err := c.SetUserTier("nobody", 4); err != ErrUserNotFound {
		t.Fatalf("unknown user: %v", err)
	}
}
