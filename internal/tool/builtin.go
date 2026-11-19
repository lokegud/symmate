package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/loke/symmate/internal/llm"
	"time"
)

// DateTimeTool returns the current UTC date and time.
func DateTimeTool() Tool {
	return Tool{
		Def: toolDef("get_datetime", "Returns the current UTC date and time.", `{
			"type": "object",
			"properties": {},
			"required": []
		}`),
		Execute: func(_ context.Context, _ json.RawMessage) string {
			return time.Now().UTC().Format(time.RFC3339)
		},
		Skill:  "- Use get_datetime to answer questions about the current date or time.",
		Tags:   []string{"time"},
		Public: true,
	}
}

// FetchURLTool returns a tool that fetches a URL and returns its body (truncated).
func FetchURLTool() Tool {
	client := &http.Client{Timeout: 15 * time.Second}
	return Tool{
		Def: toolDef("fetch_url", "Fetches the content of a URL and returns the text body.", `{
			"type": "object",
			"properties": {
				"url": {"type": "string", "description": "The URL to fetch."}
			},
			"required": ["url"]
		}`),
		Execute: func(ctx context.Context, args json.RawMessage) string {
			var p struct {
				URL string `json:"url"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
			if err != nil {
				return "error: " + err.Error()
			}
			req.Header.Set("User-Agent", "Symmate/1.0")
			resp, err := client.Do(req)
			if err != nil {
				return "error: " + err.Error()
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
			if err != nil {
				return "error: " + err.Error()
			}
			return string(body)
		},
		Skill:  "- Use fetch_url to retrieve the content of a web page or API endpoint.",
		Tags:   []string{"net/fetch"},
		Public: true,
	}
}

// SearchTool returns a tool that queries a search endpoint (SearXNG-compatible).
// endpoint is the base URL, e.g. "http://stak:8080".
func SearchTool(endpoint string) Tool {
	client := &http.Client{Timeout: 15 * time.Second}
	return Tool{
		Def: toolDef("web_search", "Searches the web and returns a list of results.", `{
			"type": "object",
			"properties": {
				"query": {"type": "string", "description": "The search query."}
			},
			"required": ["query"]
		}`),
		Execute: func(ctx context.Context, args json.RawMessage) string {
			var p struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return "error: " + err.Error()
			}
			searchURL := fmt.Sprintf("%s/search?q=%s&format=json", endpoint, url.QueryEscape(p.Query))
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
			if err != nil {
				return "error: " + err.Error()
			}
			resp, err := client.Do(req)
			if err != nil {
				return "error: " + err.Error()
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
			if err != nil {
				return "error: " + err.Error()
			}
			// SearXNG returns {"results":[{"title":"...","url":"...","content":"..."},...]}
			// Summarise to a readable list.
			var result struct {
				Results []struct {
					Title   string `json:"title"`
					URL     string `json:"url"`
					Content string `json:"content"`
				} `json:"results"`
			}
			if err := json.Unmarshal(body, &result); err != nil {
				return string(body) // return raw if not parseable
			}
			var sb strings.Builder
			for i, r := range result.Results {
				if i >= 5 {
					break
				}
				fmt.Fprintf(&sb, "%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Content)
			}
			if sb.Len() == 0 {
				return "no results"
			}
			return sb.String()
		},
		Skill:  "- Use web_search to find current information, news, or facts about any topic.",
		Tags:   []string{"net/search", "research/find"},
		Public: true,
	}
}

func toolDef(name, description, parametersJSON string) llm.ToolDef {
	return llm.ToolDef{
		Name:        name,
		Description: description,
		Parameters:  json.RawMessage(parametersJSON),
	}
}
