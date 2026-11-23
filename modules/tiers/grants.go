package tier

// Grant says what a data class may do for a caller. Permission is "allow", "deny" or "ask".
type Grant struct {
	DataClass  string
	Permission string
}

// GrantsFor returns the grants for a tier. Tiers are cumulative.
// Name is public at tier 2. Email and phone open at tier 3. Address stays closed
// unless the owner assigns tier 4.
func GrantsFor(l Level) []Grant {
	deny := func(class string) Grant { return Grant{DataClass: class, Permission: "deny"} }
	allow := func(class string) Grant { return Grant{DataClass: class, Permission: "allow"} }
	switch {
	case l >= All:
		return []Grant{allow("name"), allow("email"), allow("phone"), allow("address")}
	case l >= Schedule:
		return []Grant{allow("name"), allow("email"), allow("phone"), deny("address")}
	case l >= Public:
		return []Grant{allow("name"), deny("email"), deny("phone"), deny("address")}
	default:
		return []Grant{deny("name"), deny("email"), deny("phone"), deny("address")}
	}
}
