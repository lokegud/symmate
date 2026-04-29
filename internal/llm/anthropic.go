package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const anthropicEndpoint = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

// AnthropicProvider calls the Anthropic Messages API directly.
type AnthropicProvider struct {
	apiKey string
	client *http.Client
}

// NewAnthropicProvider creates an Anthropic provider with the given API key.
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{apiKey: apiKey, client: &http.Client{}}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
		MaxTokens: 4096,
		System:    system,
		Messages:  turns,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicEndpoint, bytes.NewReader(body))
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
	return ar.Content[0].Text, nil
}
