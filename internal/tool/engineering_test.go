package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/loke/symmate/internal/engineering"
	"github.com/loke/symmate/internal/github"
)

// fakeGitHub serves a tiny slice of the GitHub API for repo acme/app.
func fakeGitHub(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/app/pulls/7", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("missing auth header")
		}
		fmt.Fprint(w, `{"number":7,"title":"Rework login","state":"open","html_url":"https://github.com/acme/app/pull/7",
			"user":{"login":"dev"},"additions":1200,"deletions":40,"changed_files":2,
			"created_at":"2026-05-01T00:00:00Z","updated_at":"2026-05-02T00:00:00Z"}`)
	})
	mux.HandleFunc("/repos/acme/app/pulls/7/files", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "" {
			// Page 1 links to page 2, exercising pagination.
			w.Header().Set("Link", fmt.Sprintf(`<%s/repos/acme/app/pulls/7/files?per_page=100&page=2>; rel="next"`, "http://"+r.Host))
			fmt.Fprint(w, `[{"filename":"internal/auth/login.go","status":"modified","additions":1100,"deletions":40}]`)
			return
		}
		fmt.Fprint(w, `[{"filename":"go.mod","status":"modified","additions":100}]`)
	})
	mux.HandleFunc("/repos/acme/app/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") == "closed" {
			fmt.Fprint(w, `[
				{"number":1,"user":{"login":"a"},"created_at":"2026-04-01T00:00:00Z","merged_at":"2026-04-01T10:00:00Z"},
				{"number":2,"user":{"login":"b"},"created_at":"2026-04-02T00:00:00Z","merged_at":"2026-04-03T06:00:00Z"},
				{"number":3,"user":{"login":"c"},"created_at":"2026-04-03T00:00:00Z","merged_at":null},
				{"number":4,"user":{"login":"d"},"created_at":"2026-03-01T00:00:00Z","merged_at":"2026-03-02T00:00:00Z"}
			]`)
			return
		}
		fmt.Fprint(w, `[{"number":9,"title":"WIP","state":"open","draft":true,"user":{"login":"x"},
			"labels":[{"name":"bug"}],"head":{"ref":"feat"},"base":{"ref":"main"},"updated_at":"2026-05-01T00:00:00Z","html_url":"u"}]`)
	})
	mux.HandleFunc("/repos/acme/app/pulls/1/reviews", func(w http.ResponseWriter, r *http.Request) {
		// Author's own review is ignored; first other review at +2h.
		fmt.Fprint(w, `[{"user":{"login":"a"},"state":"COMMENTED","submitted_at":"2026-04-01T01:00:00Z"},
			{"user":{"login":"rev"},"state":"APPROVED","submitted_at":"2026-04-01T02:00:00Z"}]`)
	})
	mux.HandleFunc("/repos/acme/app/pulls/2/reviews", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[]`) // merged without review
	})
	mux.HandleFunc("/repos/acme/app/issues", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"number":5,"title":"Crash","state":"open","user":{"login":"u"},"labels":[{"name":"bug"}],"assignees":[{"login":"dev"}],"comments":2,"updated_at":"2026-05-01T00:00:00Z","html_url":"i5"},
			{"number":9,"title":"WIP","state":"open","user":{"login":"x"},"pull_request":{},"updated_at":"2026-05-01T00:00:00Z"}]`)
	})
	mux.HandleFunc("/repos/acme/app/compare/v1.0...main", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"files":[{"filename":".github/workflows/deploy.yml","status":"modified","additions":3}]}`)
	})
	mux.HandleFunc("/repos/acme/private/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Not Found"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func run(t *testing.T, tl Tool, args string) string {
	t.Helper()
	return tl.Execute(context.Background(), json.RawMessage(args))
}

func TestAnalyzePRRiskFromGitHub(t *testing.T) {
	srv := fakeGitHub(t)
	gh := github.New(srv.URL, "tok", "acme/app")
	out := run(t, AnalyzePRRiskTool(gh), `{"pr":"https://github.com/acme/app/pull/7"}`)
	var r engineering.PRRisk
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("bad output %q: %v", out, err)
	}
	if r.Title != "Rework login" || r.Author != "dev" || r.Additions != 1200 || r.Source != "github" {
		t.Errorf("metadata wrong: %+v", r)
	}
	if len(r.ChangedPaths) != 2 {
		t.Errorf("pagination: want 2 files, got %v", r.ChangedPaths)
	}
	if r.RiskLevel != "high" {
		t.Errorf("auth + large + go.mod + no tests should be high, got %s (%v)", r.RiskLevel, r.Signals)
	}
	// Bare number uses the default repo.
	if out := run(t, AnalyzePRRiskTool(gh), `{"pr":"7"}`); strings.HasPrefix(out, "error") {
		t.Errorf("bare number failed: %s", out)
	}
}

func TestCycleTimeFromGitHub(t *testing.T) {
	srv := fakeGitHub(t)
	gh := github.New(srv.URL, "tok", "acme/app")
	out := run(t, GetCycleTimeTool(gh), `{"date_range":"2026-04-01/2026-04-30"}`)
	var m engineering.CycleTimeMetrics
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("bad output %q: %v", out, err)
	}
	// PR1: 10h, PR2: 30h. PR3 unmerged, PR4 outside range.
	if m.PRsMerged != 2 || m.AvgHours != 20 || m.MedianHours != 10 || m.P95Hours != 30 {
		t.Errorf("cycle stats wrong: %+v", m)
	}
	// Only PR1 has a non-author review: 2h to review, 8h review→merge.
	if m.ReviewSampleSize != 1 || m.AvgHoursToFirstReview != 2 || m.AvgHoursReviewToMerge != 8 || m.BottleneckStage != "review-to-merge" {
		t.Errorf("review stats wrong: %+v", m)
	}
	if !strings.Contains(m.Notes, "without a review") {
		t.Errorf("notes should flag unreviewed merge: %q", m.Notes)
	}
}

func TestListToolsFromGitHub(t *testing.T) {
	srv := fakeGitHub(t)
	gh := github.New(srv.URL, "tok", "acme/app")

	out := run(t, ListPullRequestsTool(gh), `{}`)
	if !strings.Contains(out, `"number":9`) || !strings.Contains(out, `"draft":true`) {
		t.Errorf("list_pull_requests: %s", out)
	}
	out = run(t, ListIssuesTool(gh), `{"repo":"acme/app"}`)
	if !strings.Contains(out, `"count":1`) || strings.Contains(out, `"number":9`) {
		t.Errorf("list_issues must drop PRs: %s", out)
	}
}

func TestDeployRiskCompareAndFileList(t *testing.T) {
	srv := fakeGitHub(t)
	gh := github.New(srv.URL, "tok", "acme/app")
	out := run(t, GetDeployRiskTool(gh), `{"base":"v1.0","head":"main"}`)
	if !strings.Contains(out, "CI workflows") {
		t.Errorf("compare: %s", out)
	}
	// File-list mode works even without GitHub configured.
	out = run(t, GetDeployRiskTool(nil), `{"changed_files":["billing/charge.go"]}`)
	if !strings.Contains(out, "billing/charge.go") {
		t.Errorf("file list: %s", out)
	}
	if out := run(t, GetDeployRiskTool(gh), `{"base":"v1.0"}`); !strings.HasPrefix(out, "error") {
		t.Errorf("half a comparison should error: %s", out)
	}
}

func TestGitHubErrors(t *testing.T) {
	srv := fakeGitHub(t)
	anon := github.New(srv.URL, "", "")
	if out := run(t, AnalyzePRRiskTool(anon), `{"pr":"acme/private#1"}`); !strings.Contains(out, "no GITHUB_TOKEN") {
		t.Errorf("404 without token should hint at the token: %s", out)
	}
	if out := run(t, ListIssuesTool(anon), `{}`); !strings.Contains(out, "GITHUB_REPO") {
		t.Errorf("missing repo should say so: %s", out)
	}
	if out := run(t, ListIssuesTool(nil), `{}`); !strings.Contains(out, "not configured") {
		t.Errorf("nil client: %s", out)
	}
}

func TestParsePRRef(t *testing.T) {
	gh := github.New("", "", "acme/app")
	cases := map[string]string{
		"https://github.com/o/r/pull/12/files": "o/r#12",
		"o/r#3":                                "o/r#3",
		"42":                                   "acme/app#42",
	}
	for in, want := range cases {
		repo, n, err := gh.ParsePRRef(in)
		if err != nil || fmt.Sprintf("%s#%d", repo, n) != want {
			t.Errorf("%q → %s#%d, %v; want %s", in, repo, n, err, want)
		}
	}
	if _, _, err := gh.ParsePRRef("not a pr"); err == nil {
		t.Error("expected error")
	}
}
