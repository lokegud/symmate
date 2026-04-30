package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HermesProvider calls the LiteLLM proxy at stak (OpenAI-compatible).
type HermesProvider struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

// NewHermesProvider creates a Hermes/LiteLLM provider.
// endpoint should be e.g. "http://stak:4000/v1"
func NewHermesProvider(endpoint, apiKey string) *HermesProvider {
	return &HermesProvider{endpoint: endpoint, apiKey: apiKey, client: &http.Client{}}
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends messages to the LiteLLM proxy and returns the assistant reply.
// model is passed through to LiteLLM; use "hermes" to let LiteLLM route to Hermes-4.3-36B.
func (p *HermesProvider) Complete(ctx context.Context, model string, messages []Message) (string, error) {
	var omsgs []openAIMessage
	for _, m := range messages {
		omsgs = append(omsgs, openAIMessage{Role: m.Role, Content: m.Content})
	}

	req := openAIRequest{Model: model, Messages: omsgs}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	url := p.endpoint + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var or openAIResponse
	if err := json.Unmarshal(respBody, &or); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if or.Error != nil {
		return "", fmt.Errorf("hermes error: %s", or.Error.Message)
	}
	if len(or.Choices) == 0 {
		return "", fmt.Errorf("empty choices in response (status %d)", resp.StatusCode)
	}
	return or.Choices[0].Message.Content, nil
}
