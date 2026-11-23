package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/loke/symmate/internal/firewall"
	"github.com/loke/tiers"
)

// tierDataFile is the owner's disclosure data for visitors at tier 3 and above.
// The file's shape is specific to this host; the tiers module only sees Lines and Notes.
type tierDataFile struct {
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	BarSchedule string   `json:"bar_schedule"`
	Projects    []string `json:"projects"`
}

// loadDisclosure reads <hostDir>/tier-data.json into the tiers module's shape.
// A missing file means no disclosure data, which is not an error.
func loadDisclosure(hostDir string) (*tier.Disclosure, error) {
	data, err := os.ReadFile(filepath.Join(hostDir, "tier-data.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var f tierDataFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse tier-data.json: %w", err)
	}
	d := &tier.Disclosure{
		Lines: []tier.Line{
			{Label: "Email", Value: f.Email},
			{Label: "Phone", Value: f.Phone},
			{Label: "Bar schedule", Value: f.BarSchedule},
		},
		Notes: []string{"Calendar availability: not connected yet. Say you can't see the owner's calendar right now."},
	}
	if len(f.Projects) > 0 {
		v := ""
		for i, p := range f.Projects {
			if i > 0 {
				v += "; "
			}
			v += p
		}
		d.Lines = append(d.Lines, tier.Line{Label: "Recent projects", Value: v})
	}
	return d, nil
}

// firewallRedactor adapts the PII firewall to the tiers module's Redactor.
func firewallRedactor(fw *firewall.Firewall) tier.Redactor {
	return func(ctx context.Context, text string, grants []tier.Grant) string {
		fwGrants := make([]firewall.Grant, 0, len(grants))
		for _, g := range grants {
			fwGrants = append(fwGrants, firewall.Grant{DataClass: g.DataClass, Permission: g.Permission})
		}
		filtered, _ := fw.Filter(ctx, text, fwGrants)
		return filtered
	}
}
