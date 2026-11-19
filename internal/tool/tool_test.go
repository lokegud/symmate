package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/loke/symmate/internal/github"
	"github.com/loke/symmate/internal/llm"
)

func fakeTool(name string, public bool, tags ...string) Tool {
	return Tool{
		Def:     llm.ToolDef{Name: name},
		Execute: func(context.Context, json.RawMessage) string { return "ran " + name },
		Tags:    tags,
		Public:  public,
	}
}

func testRegistry() *Registry {
	r := &Registry{}
	r.Register(fakeTool("datetime", true, "time"))
	r.Register(fakeTool("list_prs", false, "engineering/pr"))
	r.Register(fakeTool("deal_status", false, "sales/deal"))
	r.Register(fakeTool("profile_360", false, "entity/company", "sales/account", "engineering/risk"))
	r.Register(fakeTool("untagged_restricted", false))
	return r
}

func names(defs []llm.ToolDef) string {
	var n []string
	for _, d := range defs {
		n = append(n, d.Name)
	}
	return strings.Join(n, ",")
}

func TestAccessTiers(t *testing.T) {
	r := testRegistry()
	cases := []struct {
		name string
		acc  Access
		want string
	}{
		{"anonymous", PublicAccess(nil), "datetime"},
		{"owner", OwnerAccess(nil), "datetime,list_prs,deal_status,profile_360,untagged_restricted"},
		{"user engineering", Access{Groups: []string{"engineering"}}, "datetime,list_prs"},
		// profile_360 spans entity+sales+engineering: needs all three.
		{"user eng+sales", Access{Groups: []string{"engineering", "sales"}}, "datetime,list_prs,deal_status"},
		{"user all three", Access{Groups: []string{"engineering", "sales", "entity"}}, "datetime,list_prs,deal_status,profile_360"},
		{"satellite allowlist narrows owner", OwnerAccess([]string{"engineering"}), "list_prs,profile_360"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := names(r.DefsFor(c.acc)); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestExecutorRefusesHiddenTools(t *testing.T) {
	r := testRegistry()
	exec := r.ExecutorFor(Access{Groups: []string{"engineering"}})
	if got := exec(context.Background(), "deal_status", nil); !strings.HasPrefix(got, "error: unknown tool") {
		t.Errorf("user without sales ran deal_status: %q", got)
	}
	if got := exec(context.Background(), "list_prs", nil); got != "ran list_prs" {
		t.Errorf("got %q", got)
	}
}

func TestRestrictedGroups(t *testing.T) {
	got := strings.Join(testRegistry().RestrictedGroups(), ",")
	if got != "engineering,entity,sales" {
		t.Errorf("got %q", got)
	}
}

func TestSkillsForAnonymousExcludesRestricted(t *testing.T) {
	r := &Registry{}
	pub := fakeTool("datetime", true, "time")
	pub.Skill = "- datetime skill"
	priv := fakeTool("list_prs", false, "engineering/pr")
	priv.Skill = "- secret skill"
	r.Register(pub)
	r.Register(priv)
	s := r.SkillsFor(PublicAccess(nil))
	if !strings.Contains(s, "datetime skill") || strings.Contains(s, "secret skill") {
		t.Errorf("unexpected skills block: %q", s)
	}
}

func TestBuiltinsArePublicDepartmentToolsAreNot(t *testing.T) {
	for _, tl := range []Tool{DateTimeTool(), FetchURLTool(), SearchTool("http://x")} {
		if !tl.Public {
			t.Errorf("%s should be public", tl.Def.Name)
		}
	}
	tools := append(MockDepartmentTools(), GitHubTools(github.New("", "", ""))...)
	for _, tl := range tools {
		if tl.Public {
			t.Errorf("%s must not be public", tl.Def.Name)
		}
		if len(tl.Groups()) == 0 {
			t.Errorf("%s has no group", tl.Def.Name)
		}
	}
}

func TestMockToolsAreLabelled(t *testing.T) {
	tl := MockDepartmentTools()[0]
	if !strings.HasPrefix(tl.Def.Description, "[DEMO DATA]") {
		t.Errorf("description not labelled: %q", tl.Def.Description)
	}
	out := tl.Execute(context.Background(), json.RawMessage(`{"account_id":"a"}`))
	var v struct {
		Demo bool `json:"demo_data"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil || !v.Demo {
		t.Errorf("output not marked demo: %s", out)
	}
}

func TestScoreFiles(t *testing.T) {
	low := scoreFiles([]github.File{
		{Filename: "internal/foo/bar.go", Status: "modified", Additions: 10},
		{Filename: "internal/foo/bar_test.go", Status: "modified", Additions: 10},
	})
	if low.RiskLevel != "low" || len(low.Signals) != 0 {
		t.Errorf("expected low/no signals, got %+v", low)
	}

	high := scoreFiles([]github.File{
		{Filename: "internal/auth/session.go", Status: "modified", Additions: 900, Deletions: 300},
		{Filename: ".github/workflows/ci.yml", Status: "modified", Additions: 2},
		{Filename: "internal/auth/session_test.go", Status: "removed", Deletions: 50},
	})
	if high.RiskLevel != "high" {
		t.Errorf("expected high, got %+v", high)
	}
	if len(high.HotPaths) != 1 || high.HotPaths[0] != "internal/auth/session.go" {
		t.Errorf("hot paths should exclude tests: %v", high.HotPaths)
	}

	noTests := scoreFiles([]github.File{{Filename: "main.go", Status: "modified", Additions: 5}})
	if len(noTests.Signals) != 1 || !strings.Contains(noTests.Signals[0], "no test changes") {
		t.Errorf("expected missing-tests signal, got %v", noTests.Signals)
	}

	assets := scoreFiles([]github.File{{Filename: "web/styles/tokens.css", Status: "modified"}})
	if len(assets.HotPaths) != 0 {
		t.Errorf("stylesheets must not count as sensitive: %v", assets.HotPaths)
	}

	docs := scoreFiles([]github.File{{Filename: "README.md", Status: "modified", Additions: 5}})
	if docs.RiskLevel != "low" || len(docs.Signals) != 0 {
		t.Errorf("docs-only change should be clean, got %+v", docs)
	}
}

func TestParseDateRange(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	from, to, err := parseDateRange("", now)
	if err != nil || !to.Equal(now) || !from.Equal(now.AddDate(0, 0, -30)) {
		t.Errorf("default range wrong: %v %v %v", from, to, err)
	}
	from, to, err = parseDateRange("2026-04-01/2026-04-30", now)
	if err != nil || from.Format(time.RFC3339) != "2026-04-01T00:00:00Z" || to.Format("2006-01-02T15:04") != "2026-04-30T23:59" {
		t.Errorf("explicit range wrong: %v %v %v", from, to, err)
	}
	for _, bad := range []string{"2026-04-01", "x/y", "2026-05-01/2026-04-01"} {
		if _, _, err := parseDateRange(bad, now); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}

func TestPercentile(t *testing.T) {
	xs := []float64{5, 1, 4, 2, 3}
	if p := percentile(xs, 50); p != 3 {
		t.Errorf("p50 = %v", p)
	}
	if p := percentile(xs, 95); p != 5 {
		t.Errorf("p95 = %v", p)
	}
	if xs[0] != 5 {
		t.Error("percentile mutated its input")
	}
}
