// Package github is a minimal, read-only GitHub REST client used by the
// engineering tools. It has no third-party dependencies.
//
// Configuration (environment):
//
//	GITHUB_TOKEN    personal access token (fine-grained, read-only is enough).
//	                Optional for public repos, but unauthenticated calls are
//	                limited to 60/hour and cannot see private repos.
//	GITHUB_REPO     default repository as "owner/name", used when a tool call
//	                doesn't name one.
//	GITHUB_API_URL  API base, default https://api.github.com (set for GHES).
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.github.com"

// Client talks to the GitHub REST API.
type Client struct {
	baseURL     string
	token       string
	DefaultRepo string
	http        *http.Client
}

// New creates a client. baseURL "" means api.github.com.
func New(baseURL, token, defaultRepo string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		token:       token,
		DefaultRepo: defaultRepo,
		http:        &http.Client{Timeout: 20 * time.Second},
	}
}

// FromEnv builds a client from GITHUB_TOKEN / GITHUB_REPO / GITHUB_API_URL.
func FromEnv() *Client {
	return New(os.Getenv("GITHUB_API_URL"), os.Getenv("GITHUB_TOKEN"), os.Getenv("GITHUB_REPO"))
}

// Authenticated reports whether a token is configured.
func (c *Client) Authenticated() bool { return c.token != "" }

// APIError is a non-2xx response from GitHub.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("github: HTTP %d: %s", e.Status, e.Message)
}

// ErrNoRepo is returned when no repository was given and no default is configured.
var ErrNoRepo = errors.New("no repository given and GITHUB_REPO is not set")

// Repo is an owner/name pair.
type Repo struct{ Owner, Name string }

func (r Repo) String() string { return r.Owner + "/" + r.Name }

// ResolveRepo parses "owner/name" (or a github.com URL); empty falls back to DefaultRepo.
func (c *Client) ResolveRepo(s string) (Repo, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		s = c.DefaultRepo
	}
	if s == "" {
		return Repo{}, ErrNoRepo
	}
	s = strings.TrimSuffix(s, ".git")
	if i := strings.Index(s, "github.com/"); i >= 0 {
		s = s[i+len("github.com/"):]
	}
	parts := strings.Split(strings.Trim(s, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return Repo{}, fmt.Errorf("invalid repository %q: want owner/name", s)
	}
	return Repo{Owner: parts[0], Name: parts[1]}, nil
}

var prURLRe = regexp.MustCompile(`github\.com/([^/\s]+)/([^/\s]+)/pull/(\d+)`)

// ParsePRRef accepts a PR URL, "owner/name#123", or a bare number (default repo).
func (c *Client) ParsePRRef(ref string) (Repo, int, error) {
	ref = strings.TrimSpace(ref)
	if m := prURLRe.FindStringSubmatch(ref); m != nil {
		n, _ := strconv.Atoi(m[3])
		return Repo{Owner: m[1], Name: m[2]}, n, nil
	}
	repoPart, numPart, hasHash := strings.Cut(ref, "#")
	if !hasHash {
		repoPart, numPart = "", ref
	}
	n, err := strconv.Atoi(strings.TrimSpace(numPart))
	if err != nil || n <= 0 {
		return Repo{}, 0, fmt.Errorf("invalid pull request reference %q: want a PR URL, owner/name#123, or a number", ref)
	}
	repo, err := c.ResolveRepo(repoPart)
	if err != nil {
		return Repo{}, 0, err
	}
	return repo, n, nil
}

// get performs a GET and decodes JSON into out. It returns the Link header's "next" URL, if any.
func (c *Client) get(ctx context.Context, pathOrURL string, out any) (string, error) {
	u := pathOrURL
	if !strings.HasPrefix(u, "http") {
		u = c.baseURL + pathOrURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "symmate-runtime")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("github: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(body, &e)
		msg := e.Message
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		if (resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden) && c.token == "" {
			msg += " (no GITHUB_TOKEN set: private repos are invisible and the rate limit is 60/hour)"
		}
		return "", &APIError{Status: resp.StatusCode, Message: msg}
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return "", fmt.Errorf("github: decode: %w", err)
	}
	return nextLink(resp.Header.Get("Link")), nil
}

var linkNextRe = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func nextLink(h string) string {
	if m := linkNextRe.FindStringSubmatch(h); m != nil {
		return m[1]
	}
	return ""
}

// User is a GitHub account reference.
type User struct {
	Login string `json:"login"`
}

// Label is an issue/PR label.
type Label struct {
	Name string `json:"name"`
}

// PullRequest is the subset of PR fields Symmate uses.
type PullRequest struct {
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	State        string     `json:"state"`
	Draft        bool       `json:"draft"`
	HTMLURL      string     `json:"html_url"`
	User         User       `json:"user"`
	Labels       []Label    `json:"labels"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	MergedAt     *time.Time `json:"merged_at"`
	ClosedAt     *time.Time `json:"closed_at"`
	Additions    int        `json:"additions"`     // only on single-PR fetch
	Deletions    int        `json:"deletions"`     // only on single-PR fetch
	ChangedFiles int        `json:"changed_files"` // only on single-PR fetch
	Base         struct {
		Ref string `json:"ref"`
	} `json:"base"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
}

// File is one changed file in a PR or comparison.
type File struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"` // added | removed | modified | renamed | ...
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// Review is a PR review.
type Review struct {
	User        User      `json:"user"`
	State       string    `json:"state"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Issue is an issue (PRs are filtered out by ListIssues).
type Issue struct {
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	State       string     `json:"state"`
	HTMLURL     string     `json:"html_url"`
	User        User       `json:"user"`
	Labels      []Label    `json:"labels"`
	Assignees   []User     `json:"assignees"`
	Comments    int        `json:"comments"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	PullRequest *struct{}  `json:"pull_request,omitempty"`
}

// maxFilePages caps file listing at 10×100 files (GitHub's own limit is 3000).
const maxFilePages = 10

// GetPR fetches a single pull request (includes additions/deletions/changed_files).
func (c *Client) GetPR(ctx context.Context, r Repo, number int) (*PullRequest, error) {
	var pr PullRequest
	_, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", url.PathEscape(r.Owner), url.PathEscape(r.Name), number), &pr)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

// ListPRFiles lists the files changed in a PR (up to 1000).
func (c *Client) ListPRFiles(ctx context.Context, r Repo, number int) ([]File, error) {
	var all []File
	next := fmt.Sprintf("/repos/%s/%s/pulls/%d/files?per_page=100", url.PathEscape(r.Owner), url.PathEscape(r.Name), number)
	for page := 0; next != "" && page < maxFilePages; page++ {
		var batch []File
		n, err := c.get(ctx, next, &batch)
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		next = n
	}
	return all, nil
}

// ListReviews lists reviews on a PR (first page, up to 100).
func (c *Client) ListReviews(ctx context.Context, r Repo, number int) ([]Review, error) {
	var out []Review
	_, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews?per_page=100", url.PathEscape(r.Owner), url.PathEscape(r.Name), number), &out)
	return out, err
}

// ListPRs lists pull requests. state: open | closed | all. At most limit results.
func (c *Client) ListPRs(ctx context.Context, r Repo, state string, limit int) ([]PullRequest, error) {
	if state == "" {
		state = "open"
	}
	q := url.Values{"state": {state}, "sort": {"updated"}, "direction": {"desc"}, "per_page": {"100"}}
	next := fmt.Sprintf("/repos/%s/%s/pulls?%s", url.PathEscape(r.Owner), url.PathEscape(r.Name), q.Encode())
	var all []PullRequest
	for next != "" && len(all) < limit {
		var batch []PullRequest
		n, err := c.get(ctx, next, &batch)
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		next = n
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

// ListIssues lists issues (excluding PRs). state: open | closed | all.
func (c *Client) ListIssues(ctx context.Context, r Repo, state, labels string, limit int) ([]Issue, error) {
	if state == "" {
		state = "open"
	}
	q := url.Values{"state": {state}, "sort": {"updated"}, "direction": {"desc"}, "per_page": {"100"}}
	if labels != "" {
		q.Set("labels", labels)
	}
	next := fmt.Sprintf("/repos/%s/%s/issues?%s", url.PathEscape(r.Owner), url.PathEscape(r.Name), q.Encode())
	var all []Issue
	for next != "" && len(all) < limit {
		var batch []Issue
		n, err := c.get(ctx, next, &batch)
		if err != nil {
			return nil, err
		}
		for _, is := range batch {
			if is.PullRequest == nil {
				all = append(all, is)
			}
		}
		next = n
		if len(batch) == 0 {
			break
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

// Compare lists files changed between base and head (branch, tag, or SHA).
func (c *Client) Compare(ctx context.Context, r Repo, base, head string) ([]File, error) {
	var out struct {
		Files []File `json:"files"`
	}
	_, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/compare/%s...%s",
		url.PathEscape(r.Owner), url.PathEscape(r.Name), url.PathEscape(base), url.PathEscape(head)), &out)
	return out.Files, err
}
