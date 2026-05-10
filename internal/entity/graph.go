package entity

import (
	"fmt"
	"strings"
	"time"
)

type Company struct {
	Name     string   `json:"name"`
	Domain   string   `json:"domain"`
	Industry string   `json:"industry"`
	ARR      float64  `json:"arr"`
	Tags     []string `json:"tags"`
}

type Contact struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	CompanyName string `json:"company_name"`
	Role        string `json:"role"`
	Department  string `json:"department"`
}

type EntityProfile struct {
	Company     Company     `json:"company"`
	Contacts    []Contact   `json:"contacts"`
	DeptSignals []DeptSignal `json:"dept_signals"`
}

type DeptSignal struct {
	Department string `json:"department"`
	Signal     string `json:"signal"`
	Severity   string `json:"severity"`
	At         string `json:"at"`
}

type Registry struct {
	companies []Company
	contacts  []Contact
}

func NewRegistry() *Registry {
	r := &Registry{
		companies: []Company{
			{Name: "Acme Corp", Domain: "acme.com", Industry: "SaaS", ARR: 240000, Tags: []string{"enterprise", "fintech"}},
			{Name: "Globex Inc", Domain: "globex.io", Industry: "Manufacturing", ARR: 180000, Tags: []string{"mid-market"}},
			{Name: "Initech LLC", Domain: "initech.net", Industry: "Financial Services", ARR: 95000, Tags: []string{"smb", "at-risk"}},
		},
		contacts: []Contact{
			{Name: "Alice Acme", Email: "alice@acme.com", CompanyName: "Acme Corp", Role: "Champion", Department: "sales"},
			{Name: "Bob Acme", Email: "bob@acme.com", CompanyName: "Acme Corp", Role: "Economic Buyer", Department: "sales"},
			{Name: "Carol Globex", Email: "carol@globex.io", CompanyName: "Globex Inc", Role: "Champion", Department: "sales"},
			{Name: "Dave Globex", Email: "dave@globex.io", CompanyName: "Globex Inc", Role: "Economic Buyer", Department: "cs"},
			{Name: "Eve Initech", Email: "eve@initech.net", CompanyName: "Initech LLC", Role: "Champion", Department: "sales"},
			{Name: "Frank Initech", Email: "frank@initech.net", CompanyName: "Initech LLC", Role: "Economic Buyer", Department: "marketing"},
		},
	}
	return r
}

func (r *Registry) LookupCompany(name string) (Company, bool) {
	lower := strings.ToLower(name)
	for _, c := range r.companies {
		if strings.HasPrefix(strings.ToLower(c.Name), lower) {
			return c, true
		}
	}
	return Company{}, false
}

func (r *Registry) ContactsFor(companyName string) []Contact {
	lower := strings.ToLower(companyName)
	var out []Contact
	for _, c := range r.contacts {
		if strings.ToLower(c.CompanyName) == lower {
			out = append(out, c)
		}
	}
	return out
}

func (r *Registry) Profile(companyName string) (EntityProfile, error) {
	company, ok := r.LookupCompany(companyName)
	if !ok {
		return EntityProfile{}, fmt.Errorf("company not found: %s", companyName)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	signals := []DeptSignal{
		{
			Department: "sales",
			Signal:     fmt.Sprintf("%s deal is in Negotiation stage; legal review in progress.", company.Name),
			Severity:   "info",
			At:         now,
		},
		{
			Department: "cs",
			Signal:     fmt.Sprintf("%s is on Health Tier 2; NPS last scored 34; renewal at risk.", company.Name),
			Severity:   "warning",
			At:         now,
		},
		{
			Department: "engineering",
			Signal:     fmt.Sprintf("Recent deploy to %s tenant introduced p95 latency regression; rollback pending.", company.Domain),
			Severity:   "critical",
			At:         now,
		},
		{
			Department: "marketing",
			Signal:     fmt.Sprintf("%s opened 4 of last 5 campaign emails; attended Q2 webinar.", company.Name),
			Severity:   "info",
			At:         now,
		},
	}
	return EntityProfile{
		Company:     company,
		Contacts:    r.ContactsFor(company.Name),
		DeptSignals: signals,
	}, nil
}

func (r *Registry) AllCompanies() []Company {
	return r.companies
}
