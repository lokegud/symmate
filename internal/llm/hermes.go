package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// HermesProvider calls the LiteLLM proxy at stak (OpenAI-compatible).
// Stak's LiteLLM is configured stream-only; non-streaming requests hang indefinitely.
type HermesProvider struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

func NewHermesProvider(endpoint, apiKey string) *HermesProvider {
	return &HermesProvider{endpoint: endpoint, apiKey: apiKey, client: &http.Client{Timeout: 5 * time.Minute}}
}

// --- request types ---

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Tools    []openAITool    `json:"tools,omitempty"`
}

// openAIMessage uses *string for Content so it serialises as null when absent (tool_calls messages).
type openAIMessage struct {
	Role       string        `json:"role"`
	Content    *string       `json:"content"`
	ToolCalls  []oaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	Name       string        `json:"name,omitempty"`
}

type openAITool struct {
	Type     string `json:"type"` // "function"
	Function struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
	} `json:"function"`
}

type oaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// --- response / stream types ---

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Role      string             `json:"role"`
			Content   string             `json:"content"`
			ToolCalls []oaiToolCallDelta `json:"tool_calls"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type oaiToolCallDelta struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// toOAIMessage converts an llm.Message to the OpenAI wire format.
func toOAIMessage(m Message) openAIMessage {
	msg := openAIMessage{
		Role:       m.Role,
		ToolCallID: m.ToolCallID,
		Name:       m.Name,
	}
	if len(m.ToolCalls) > 0 {
		for _, tc := range m.ToolCalls {
			args := string(tc.Args)
			msg.ToolCalls = append(msg.ToolCalls, oaiToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{Name: tc.Name, Arguments: args},
			})
		}
		// content is nil (null) for tool-call assistant messages
	} else {
		c := m.Content
		msg.Content = &c
	}
	return msg
}

func (p *HermesProvider) newSSERequest(ctx context.Context, req openAIRequest) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	return p.client.Do(httpReq)
}

// Complete calls Stream and accumulates the full reply.
func (p *HermesProvider) Complete(ctx context.Context, model string, messages []Message) (string, error) {
	out := make(chan string, 64)
	errc := make(chan error, 1)
	go func() {
		errc <- p.Stream(ctx, model, messages, out)
	}()
	var sb strings.Builder
	for tok := range out {
		sb.WriteString(tok)
	}
	return sb.String(), <-errc
}

// Stream sends each reply token to out as it arrives from the LiteLLM SSE stream.
// Closes out when the stream ends (or on error).
func (p *HermesProvider) Stream(ctx context.Context, model string, messages []Message, out chan<- string) error {
	defer close(out)

	omsgs := make([]openAIMessage, len(messages))
	for i, m := range messages {
		omsgs[i] = toOAIMessage(m)
	}

	resp, err := p.newSSERequest(ctx, openAIRequest{Model: model, Messages: omsgs, Stream: true})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64<<10), 512<<10)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := line[6:]
		if payload == "[DONE]" {
			break
		}
		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if chunk.Error != nil {
			return fmt.Errorf("hermes error: %s", chunk.Error.Message)
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case out <- chunk.Choices[0].Delta.Content:
			}
		}
	}
	return scanner.Err()
}

// CompleteStep runs one LLM turn with tools available and returns the full result.
// Uses SSE under the hood (LiteLLM is stream-only) but accumulates before returning.
// Implements ToolProvider.
func (p *HermesProvider) CompleteStep(ctx context.Context, model string, messages []Message, tools []ToolDef) (StepResult, error) {
	omsgs := make([]openAIMessage, len(messages))
	for i, m := range messages {
		omsgs[i] = toOAIMessage(m)
	}

	oaiTools := make([]openAITool, len(tools))
	for i, t := range tools {
		oaiTools[i] = openAITool{Type: "function"}
		oaiTools[i].Function.Name = t.Name
		oaiTools[i].Function.Description = t.Description
		oaiTools[i].Function.Parameters = t.Parameters
	}

	resp, err := p.newSSERequest(ctx, openAIRequest{Model: model, Messages: omsgs, Stream: true, Tools: oaiTools})
	if err != nil {
		return StepResult{}, err
	}
	defer resp.Body.Close()

	type pendingCall struct {
		id   string
		name string
		args strings.Builder
	}
	var (
		contentBuf  strings.Builder
		pendingCalls []*pendingCall
	)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64<<10), 512<<10)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := line[6:]
		if payload == "[DONE]" {
			break
		}
		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if chunk.Error != nil {
			return StepResult{}, fmt.Errorf("hermes error: %s", chunk.Error.Message)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta

		if delta.Content != "" {
			contentBuf.WriteString(delta.Content)
		}

		for _, tc := range delta.ToolCalls {
			for tc.Index >= len(pendingCalls) {
				pendingCalls = append(pendingCalls, nil)
			}
			pc := pendingCalls[tc.Index]
			if pc == nil {
				pc = &pendingCall{id: tc.ID, name: tc.Function.Name}
				pendingCalls[tc.Index] = pc
			}
			if tc.ID != "" {
				pc.id = tc.ID
			}
			if tc.Function.Name != "" {
				pc.name = tc.Function.Name
			}
			pc.args.WriteString(tc.Function.Arguments)
		}

		if ctx.Err() != nil {
			return StepResult{}, ctx.Err()
		}
	}
	if err := scanner.Err(); err != nil {
		return StepResult{}, err
	}

	// Return tool calls if any were accumulated
	var calls []ToolCall
	for _, pc := range pendingCalls {
		if pc != nil && pc.name != "" {
			calls = append(calls, ToolCall{
				ID:   pc.id,
				Name: pc.name,
				Args: json.RawMessage(pc.args.String()),
			})
		}
	}
	if len(calls) > 0 {
		return StepResult{ToolCalls: calls}, nil
	}
	return StepResult{Content: contentBuf.String()}, nil
}
