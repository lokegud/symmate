package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const anthropicDefaultEndpoint = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

// AnthropicProvider calls the Anthropic Messages API directly.
type AnthropicProvider struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

// NewAnthropicProvider creates an Anthropic provider with the given API key.
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{endpoint: anthropicDefaultEndpoint, apiKey: apiKey, client: &http.Client{}}
}

// NewAnthropicProviderAt creates an Anthropic provider pointing at a custom endpoint.
// Used in tests to point at a stub server instead of api.anthropic.com.
func NewAnthropicProviderAt(endpoint, apiKey string) *AnthropicProvider {
	return &AnthropicProvider{endpoint: endpoint, apiKey: apiKey, client: &http.Client{}}
}

type anthropicRequest struct {
	Model        string             `json:"model"`
	MaxTokens    int                `json:"max_tokens"`
	System       string             `json:"system,omitempty"`
	Messages     []anthropicMessage `json:"messages"`
	Thinking     *thinkingConfig    `json:"thinking,omitempty"`
	OutputConfig *outputConfigParam `json:"output_config,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// thinkingConfig is the request-side thinking knob. On Opus 4.7 only "adaptive"
// and "disabled" are accepted — the legacy {type:"enabled", budget_tokens:N}
// form returns 400.
type thinkingConfig struct {
	Type string `json:"type"`
}

// outputConfigParam carries the GA effort knob. Supported on Opus 4.5+,
// Opus 4.6, Opus 4.7, and Sonnet 4.6. Errors on Haiku 4.5.
type outputConfigParam struct {
	Effort string `json:"effort"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete sends a message list to the Anthropic API and returns the assistant reply.
// System messages are extracted and sent as the top-level system field.
func (p *AnthropicProvider) Complete(ctx context.Context, model string, messages []Message) (string, error) {
	var system string
	var turns []anthropicMessage
	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		turns = append(turns, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	req := anthropicRequest{
		Model:     model,
		MaxTokens: 16000,
		System:    system,
		Messages:  turns,
	}
	// Opus 4.x and Sonnet 4.6 accept adaptive thinking + the GA effort knob.
	// Haiku 4.5 returns 400 on `effort`, so skip both there.
	if !strings.Contains(model, "haiku") {
		req.Thinking = &thinkingConfig{Type: "adaptive"}
		req.OutputConfig = &outputConfigParam{Effort: "high"}
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ar anthropicResponse
	if err := json.Unmarshal(respBody, &ar); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if ar.Error != nil {
		return "", fmt.Errorf("anthropic error %s: %s", ar.Error.Type, ar.Error.Message)
	}
	if len(ar.Content) == 0 {
		return "", fmt.Errorf("empty content in response (status %d)", resp.StatusCode)
	}
	// With adaptive thinking enabled the first block can be a `thinking` block
	// (text empty by default on Opus 4.7); pick the first `text` block.
	for _, c := range ar.Content {
		if c.Type == "text" {
			return c.Text, nil
		}
	}
	return "", fmt.Errorf("no text block in response (status %d)", resp.StatusCode)
}

// Stream calls Complete and sends the full reply as a single token.
// TODO: implement native Anthropic streaming when needed.
func (p *AnthropicProvider) Stream(ctx context.Context, model string, messages []Message, out chan<- string) error {
	defer close(out)
	reply, err := p.Complete(ctx, model, messages)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case out <- reply:
	}
	return nil
}
