package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/loke/symmate/internal/engineering"
	"github.com/loke/symmate/internal/github"
)

// riskyPathWords are path fragments where mistakes tend to be expensive.
// This is a fixed heuristic, not learned from incident history: no incident
// source (PagerDuty, Linear, ...) is connected yet.
var riskyPathWords = []string{
	"payment", "billing", "checkout", "auth", "session", "webhook",
	"crypto", "secret", "credential", "token", "migration", "schema",
}

// nonCodeExts are never treated as sensitive paths (e.g. design "tokens.css").
var nonCodeExts = map[string]bool{
	".css": true, ".scss": true, ".svg": true, ".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".ico": true, ".icns": true, ".webp": true, ".md": true, ".txt": true,
}

// compareFileCap is GitHub's maximum number of files returned by the compare API.
const compareFileCap = 300

// infraPathSuffixes flag build, dependency, and CI changes.
var infraPathSuffixes = []string{
	"go.mod", "go.sum", "package.json", "bun.lock", "pnpm-lock.yaml", "package-lock.json",
	"cargo.toml", "cargo.lock", "dockerfile", "makefile",
}

const (
	reviewSampleMax  = 25  // PRs whose reviews are fetched for review-wait stats
	cycleTimeMaxPRs  = 300 // closed PRs scanned for a cycle-time window
	listDefaultLimit = 20
	listMaxLimit     = 100
)

func notConfigured(gh *github.Client) string {
	if gh == nil {
		return "error: GitHub integration is not configured on this Symmate"
	}
	return ""
}

// AnalyzePRRiskTool assesses a pull request using its real diff from GitHub.
func AnalyzePRRiskTool(gh *github.Client) Tool {
	return Tool{
		Def: toolDef("analyze_pr_risk",
			"Fetches a GitHub pull request and scores its risk from the real diff: size, risky paths (auth, payments, migrations, ...), dependency/CI changes, and test coverage of the change.",
			`{
				"type": "object",
				"properties": {
					"pr": {"type": "string", "description": "PR URL (https://github.com/owner/repo/pull/123), \"owner/repo#123\", or just a number for the default repo."}
				},
				"required": ["pr"]
			}`),
		Execute: func(ctx context.Context, args json.RawMessage) string {
			if e := notConfigured(gh); e != "" {
				return e
			}
			var p struct {
				PR    string `json:"pr"`
				PRUrl string `json:"pr_url"` // legacy argument name
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			if p.PR == "" {
				p.PR = p.PRUrl
			}
			repo, num, err := gh.ParsePRRef(p.PR)
			if err != nil {
				return "error: " + err.Error()
			}
			pr, err := gh.GetPR(ctx, repo, num)
			if err != nil {
				return "error: " + err.Error()
			}
			files, err := gh.ListPRFiles(ctx, repo, num)
			if err != nil {
				return "error: " + err.Error()
			}
			return jsonString(analyzePRRisk(pr, files))
		},
		Skill: "- Use analyze_pr_risk to review a GitHub pull request's risk before merge. Report the signals it returns; don't invent incident history.",
		Tags:  []string{"engineering/pr", "engineering/risk"},
	}
}

// GetDeployRiskTool scores a change set given as a file list or a GitHub base...head comparison.
func GetDeployRiskTool(gh *github.Client) Tool {
	return Tool{
		Def: toolDef("get_deploy_risk",
			"Scores the risk of deploying a change. Give either changed_files, or base and head (branch, tag or SHA) to compare on GitHub, e.g. base=last release tag, head=main.",
			`{
				"type": "object",
				"properties": {
					"changed_files": {"type": "array", "items": {"type": "string"}, "description": "File paths in the change."},
					"repo": {"type": "string", "description": "owner/repo; defaults to the configured repo."},
					"base": {"type": "string", "description": "Base ref for a GitHub comparison."},
					"head": {"type": "string", "description": "Head ref for a GitHub comparison."}
				}
			}`),
		Execute: func(ctx context.Context, args json.RawMessage) string {
			var p struct {
				ChangedFiles []string `json:"changed_files"`
				Repo         string   `json:"repo"`
				Base         string   `json:"base"`
				Head         string   `json:"head"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			if p.Base != "" || p.Head != "" {
				if e := notConfigured(gh); e != "" {
					return e
				}
				if p.Base == "" || p.Head == "" {
					return "error: both base and head are required for a comparison"
				}
				repo, err := gh.ResolveRepo(p.Repo)
				if err != nil {
					return "error: " + err.Error()
				}
				files, err := gh.Compare(ctx, repo, p.Base, p.Head)
				if err != nil {
					return "error: " + err.Error()
				}
				r := scoreFiles(files)
				if len(files) >= compareFileCap {
					r.Reasoning += fmt.Sprintf(" GitHub returns at most %d files for a comparison, so the real change may be larger.", compareFileCap)
				}
				return jsonString(r)
			}
			if len(p.ChangedFiles) == 0 {
				return "error: give changed_files, or base and head"
			}
			files := make([]github.File, len(p.ChangedFiles))
			for i, f := range p.ChangedFiles {
				files[i] = github.File{Filename: f, Status: "modified"}
			}
			return jsonString(scoreFiles(files))
		},
		Skill: "- Use get_deploy_risk to flag whether a deploy touches risky paths, dependencies or CI. For \"what's going out\", compare the last release tag to main.",
		Tags:  []string{"engineering/risk", "engineering/deploy"},
	}
}

// GetCycleTimeTool reports PR cycle time for a repository from GitHub history.
func GetCycleTimeTool(gh *github.Client) Tool {
	return Tool{
		Def: toolDef("get_cycle_time",
			"Returns pull-request cycle time (opened → merged) for a GitHub repo over a date range, plus time to first review and where PRs wait longest.",
			`{
				"type": "object",
				"properties": {
					"repo":       {"type": "string", "description": "owner/repo; defaults to the configured repo."},
					"date_range": {"type": "string", "description": "ISO range \"2026-04-01/2026-04-30\". Defaults to the last 30 days."}
				}
			}`),
		Execute: func(ctx context.Context, args json.RawMessage) string {
			if e := notConfigured(gh); e != "" {
				return e
			}
			var p struct {
				Repo      string `json:"repo"`
				DateRange string `json:"date_range"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			repo, err := gh.ResolveRepo(p.Repo)
			if err != nil {
				return "error: " + err.Error()
			}
			from, to, err := parseDateRange(p.DateRange, time.Now().UTC())
			if err != nil {
				return "error: " + err.Error()
			}
			m, err := cycleTime(ctx, gh, repo, from, to)
			if err != nil {
				return "error: " + err.Error()
			}
			return jsonString(m)
		},
		Skill: "- Use get_cycle_time for questions about delivery speed, review delays or throughput on a GitHub repo.",
		Tags:  []string{"engineering/metrics", "engineering/cycle-time"},
	}
}

// ListPullRequestsTool lists PRs on a GitHub repo.
func ListPullRequestsTool(gh *github.Client) Tool {
	return Tool{
		Def: toolDef("list_pull_requests",
			"Lists pull requests on a GitHub repo, most recently updated first.",
			`{
				"type": "object",
				"properties": {
					"repo":  {"type": "string", "description": "owner/repo; defaults to the configured repo."},
					"state": {"type": "string", "enum": ["open", "closed", "all"], "description": "Default open."},
					"limit": {"type": "integer", "description": "Max results, default 20, max 100."}
				}
			}`),
		Execute: func(ctx context.Context, args json.RawMessage) string {
			if e := notConfigured(gh); e != "" {
				return e
			}
			var p struct {
				Repo  string `json:"repo"`
				State string `json:"state"`
				Limit int    `json:"limit"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			repo, err := gh.ResolveRepo(p.Repo)
			if err != nil {
				return "error: " + err.Error()
			}
			prs, err := gh.ListPRs(ctx, repo, p.State, clampLimit(p.Limit))
			if err != nil {
				return "error: " + err.Error()
			}
			type row struct {
				Number  int      `json:"number"`
				Title   string   `json:"title"`
				State   string   `json:"state"`
				Draft   bool     `json:"draft,omitempty"`
				Merged  bool     `json:"merged,omitempty"`
				Author  string   `json:"author"`
				Labels  []string `json:"labels,omitempty"`
				Head    string   `json:"head"`
				Base    string   `json:"base"`
				Updated string   `json:"updated"`
				URL     string   `json:"url"`
			}
			out := make([]row, 0, len(prs))
			for _, pr := range prs {
				out = append(out, row{
					Number: pr.Number, Title: pr.Title, State: pr.State, Draft: pr.Draft,
					Merged: pr.MergedAt != nil, Author: pr.User.Login, Labels: labelNames(pr.Labels),
					Head: pr.Head.Ref, Base: pr.Base.Ref, Updated: pr.UpdatedAt.Format(time.RFC3339), URL: pr.HTMLURL,
				})
			}
			return jsonString(map[string]any{"repo": repo.String(), "count": len(out), "pull_requests": out})
		},
		Skill: "- Use list_pull_requests to see what's open, waiting, or recently merged on a GitHub repo.",
		Tags:  []string{"engineering/pr"},
	}
}

// ListIssuesTool lists issues on a GitHub repo.
func ListIssuesTool(gh *github.Client) Tool {
	return Tool{
		Def: toolDef("list_issues",
			"Lists GitHub issues (not PRs) on a repo, most recently updated first.",
			`{
				"type": "object",
				"properties": {
					"repo":   {"type": "string", "description": "owner/repo; defaults to the configured repo."},
					"state":  {"type": "string", "enum": ["open", "closed", "all"], "description": "Default open."},
					"labels": {"type": "string", "description": "Comma-separated label filter, e.g. \"bug,urgent\"."},
					"limit":  {"type": "integer", "description": "Max results, default 20, max 100."}
				}
			}`),
		Execute: func(ctx context.Context, args json.RawMessage) string {
			if e := notConfigured(gh); e != "" {
				return e
			}
			var p struct {
				Repo   string `json:"repo"`
				State  string `json:"state"`
				Labels string `json:"labels"`
				Limit  int    `json:"limit"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			repo, err := gh.ResolveRepo(p.Repo)
			if err != nil {
				return "error: " + err.Error()
			}
			issues, err := gh.ListIssues(ctx, repo, p.State, p.Labels, clampLimit(p.Limit))
			if err != nil {
				return "error: " + err.Error()
			}
			type row struct {
				Number    int      `json:"number"`
				Title     string   `json:"title"`
				State     string   `json:"state"`
				Author    string   `json:"author"`
				Assignees []string `json:"assignees,omitempty"`
				Labels    []string `json:"labels,omitempty"`
				Comments  int      `json:"comments"`
				Updated   string   `json:"updated"`
				URL       string   `json:"url"`
			}
			out := make([]row, 0, len(issues))
			for _, is := range issues {
				var assignees []string
				for _, a := range is.Assignees {
					assignees = append(assignees, a.Login)
				}
				out = append(out, row{
					Number: is.Number, Title: is.Title, State: is.State, Author: is.User.Login,
					Assignees: assignees, Labels: labelNames(is.Labels), Comments: is.Comments,
					Updated: is.UpdatedAt.Format(time.RFC3339), URL: is.HTMLURL,
				})
			}
			return jsonString(map[string]any{"repo": repo.String(), "count": len(out), "issues": out})
		},
		Skill: "- Use list_issues to check bugs, open work, or anything labelled on a GitHub repo.",
		Tags:  []string{"engineering/issues"},
	}
}

// GitHubTools returns every GitHub-backed engineering tool.
func GitHubTools(gh *github.Client) []Tool {
	return []Tool{
		AnalyzePRRiskTool(gh),
		GetDeployRiskTool(gh),
		GetCycleTimeTool(gh),
		ListPullRequestsTool(gh),
		ListIssuesTool(gh),
	}
}

// --- scoring ---

func analyzePRRisk(pr *github.PullRequest, files []github.File) engineering.PRRisk {
	r := scoreFiles(files)
	// Prefer GitHub's own totals (the file list is capped at 1000 entries).
	add, del := pr.Additions, pr.Deletions
	if add == 0 && del == 0 {
		for _, f := range files {
			add += f.Additions
			del += f.Deletions
		}
	}
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Filename
	}
	return engineering.PRRisk{
		PRUrl:        pr.HTMLURL,
		Title:        pr.Title,
		Author:       pr.User.Login,
		State:        prState(pr),
		Additions:    add,
		Deletions:    del,
		ChangedPaths: paths,
		RiskLevel:    r.RiskLevel,
		Signals:      r.Signals,
		Reasoning:    r.Reasoning,
		Source:       "github",
	}
}

func prState(pr *github.PullRequest) string {
	switch {
	case pr.MergedAt != nil:
		return "merged"
	case pr.Draft:
		return "draft"
	default:
		return pr.State
	}
}

func isTestPath(p string) bool {
	l := strings.ToLower(p)
	base := path.Base(l)
	return strings.HasSuffix(base, "_test.go") || strings.Contains(base, ".test.") ||
		strings.Contains(base, ".spec.") || strings.HasPrefix(base, "test_") ||
		strings.Contains(l, "/test/") || strings.Contains(l, "/tests/") || strings.HasPrefix(l, "test/") || strings.HasPrefix(l, "tests/")
}

func isDocPath(p string) bool {
	l := strings.ToLower(p)
	return strings.HasSuffix(l, ".md") || strings.HasPrefix(l, "docs/") || strings.Contains(l, "/docs/")
}

// scoreFiles is a transparent points-based heuristic. Each signal is reported so the
// owner can see why a level was assigned.
func scoreFiles(files []github.File) engineering.DeployRisk {
	var hot, signals []string
	score := 0
	lines := 0
	codeChanged, testsChanged := false, false
	var testsDeleted, infra, ci []string

	for _, f := range files {
		l := strings.ToLower(f.Filename)
		lines += f.Additions + f.Deletions
		test := isTestPath(f.Filename)
		switch {
		case test:
			testsChanged = true
			if f.Status == "removed" {
				testsDeleted = append(testsDeleted, f.Filename)
			}
		case !isDocPath(f.Filename):
			codeChanged = true
		}
		if strings.HasPrefix(l, ".github/workflows/") {
			ci = append(ci, f.Filename)
		}
		for _, suf := range infraPathSuffixes {
			if path.Base(l) == suf {
				infra = append(infra, f.Filename)
				break
			}
		}
		if test || nonCodeExts[path.Ext(l)] {
			continue
		}
		for _, w := range riskyPathWords {
			if strings.Contains(l, w) {
				hot = append(hot, f.Filename)
				break
			}
		}
	}

	if len(hot) > 0 {
		score += 3
		signals = append(signals, fmt.Sprintf("touches %d file(s) on sensitive paths (auth, payments, secrets, migrations...): %s", len(hot), truncList(hot, 8)))
	}
	if len(ci) > 0 {
		score += 2
		signals = append(signals, "changes CI workflows: "+truncList(ci, 5))
	}
	if len(infra) > 0 {
		score++
		signals = append(signals, "changes build or dependency files: "+truncList(infra, 5))
	}
	switch {
	case lines > 1000:
		score += 2
		signals = append(signals, fmt.Sprintf("large change: %d lines added/removed", lines))
	case lines > 400:
		score++
		signals = append(signals, fmt.Sprintf("medium-sized change: %d lines added/removed", lines))
	}
	if len(files) > 30 {
		score++
		signals = append(signals, fmt.Sprintf("wide change: %d files", len(files)))
	}
	if len(testsDeleted) > 0 {
		score += 2
		signals = append(signals, "deletes test files: "+truncList(testsDeleted, 5))
	}
	if codeChanged && !testsChanged {
		score++
		signals = append(signals, "code changed with no test changes")
	}

	level := "low"
	switch {
	case score >= 5:
		level = "high"
	case score >= 2:
		level = "medium"
	}
	reasoning := "No risk signals found."
	if len(signals) > 0 {
		reasoning = fmt.Sprintf("Heuristic score %d (low <2, medium 2-4, high ≥5) from: %s.", score, strings.Join(signals, "; "))
		if level == "high" {
			reasoning += " Suggest an extra reviewer and watching the deploy closely."
		}
	}
	if hot == nil {
		hot = []string{}
	}
	if signals == nil {
		signals = []string{}
	}
	return engineering.DeployRisk{RiskLevel: level, HotPaths: hot, Signals: signals, Reasoning: reasoning}
}

func truncList(xs []string, n int) string {
	if len(xs) <= n {
		return strings.Join(xs, ", ")
	}
	return strings.Join(xs[:n], ", ") + fmt.Sprintf(" (+%d more)", len(xs)-n)
}

// --- cycle time ---

// parseDateRange parses "YYYY-MM-DD/YYYY-MM-DD"; empty means the 30 days up to now.
// The end date is inclusive.
func parseDateRange(s string, now time.Time) (time.Time, time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return now.AddDate(0, 0, -30), now, nil
	}
	a, b, ok := strings.Cut(s, "/")
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date_range %q: want YYYY-MM-DD/YYYY-MM-DD", s)
	}
	from, err := time.Parse("2006-01-02", strings.TrimSpace(a))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
	}
	to, err := time.Parse("2006-01-02", strings.TrimSpace(b))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
	}
	to = to.Add(24*time.Hour - time.Nanosecond)
	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("date_range ends before it starts")
	}
	return from, to, nil
}

func cycleTime(ctx context.Context, gh *github.Client, repo github.Repo, from, to time.Time) (engineering.CycleTimeMetrics, error) {
	m := engineering.CycleTimeMetrics{
		Repo: repo.String(), From: from.Format("2006-01-02"), To: to.Format("2006-01-02"),
		BottleneckStage: "unknown",
	}
	prs, err := gh.ListPRs(ctx, repo, "closed", cycleTimeMaxPRs)
	if err != nil {
		return m, err
	}
	var merged []github.PullRequest
	for _, pr := range prs {
		if pr.MergedAt != nil && !pr.MergedAt.Before(from) && !pr.MergedAt.After(to) {
			merged = append(merged, pr)
		}
	}
	m.PRsMerged = len(merged)
	notes := []string{"Measured from PR creation, so time spent as a draft counts."}
	if len(prs) == cycleTimeMaxPRs {
		notes = append(notes, fmt.Sprintf("Only the %d most recently updated closed PRs were scanned.", cycleTimeMaxPRs))
	}
	if len(merged) == 0 {
		m.Notes = strings.Join(append(notes, "No PRs were merged in this range."), " ")
		return m, nil
	}

	hours := make([]float64, len(merged))
	for i, pr := range merged {
		hours[i] = pr.MergedAt.Sub(pr.CreatedAt).Hours()
	}
	m.AvgHours = round1(mean(hours))
	m.MedianHours = round1(percentile(hours, 50))
	m.P95Hours = round1(percentile(hours, 95))

	// Review timing on a sample of the most recent merges (one API call each).
	sort.Slice(merged, func(i, j int) bool { return merged[i].MergedAt.After(*merged[j].MergedAt) })
	sample := merged
	if len(sample) > reviewSampleMax {
		sample = sample[:reviewSampleMax]
	}
	var toReview, reviewToMerge []float64
	unreviewed := 0
	for _, pr := range sample {
		reviews, err := gh.ListReviews(ctx, repo, pr.Number)
		if err != nil {
			return m, err
		}
		var first *time.Time
		for _, rv := range reviews {
			if rv.User.Login == pr.User.Login || rv.SubmittedAt.IsZero() {
				continue
			}
			t := rv.SubmittedAt
			if first == nil || t.Before(*first) {
				first = &t
			}
		}
		if first == nil {
			unreviewed++
			continue
		}
		toReview = append(toReview, first.Sub(pr.CreatedAt).Hours())
		reviewToMerge = append(reviewToMerge, pr.MergedAt.Sub(*first).Hours())
	}
	m.ReviewSampleSize = len(toReview)
	if len(toReview) > 0 {
		m.AvgHoursToFirstReview = round1(mean(toReview))
		m.AvgHoursReviewToMerge = round1(mean(reviewToMerge))
		if m.AvgHoursToFirstReview >= m.AvgHoursReviewToMerge {
			m.BottleneckStage = "waiting-for-review"
		} else {
			m.BottleneckStage = "review-to-merge"
		}
	}
	if unreviewed > 0 {
		notes = append(notes, fmt.Sprintf("%d of the %d sampled PRs were merged without a review from someone other than the author.", unreviewed, len(sample)))
	}
	m.Notes = strings.Join(notes, " ")
	return m, nil
}

func mean(xs []float64) float64 {
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

// percentile uses the nearest-rank method on a copy of xs.
func percentile(xs []float64, p float64) float64 {
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	rank := int(math.Ceil(p/100*float64(len(s)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(s) {
		rank = len(s) - 1
	}
	return s[rank]
}

func round1(x float64) float64 { return math.Round(x*10) / 10 }

func clampLimit(n int) int {
	if n <= 0 {
		return listDefaultLimit
	}
	if n > listMaxLimit {
		return listMaxLimit
	}
	return n
}

func labelNames(ls []github.Label) []string {
	var out []string
	for _, l := range ls {
		out = append(out, l.Name)
	}
	return out
}

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"marshal failed"}`
	}
	return string(b)
}
