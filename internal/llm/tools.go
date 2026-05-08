package llm

import (
	"context"
	"encoding/json"
)

// ToolDef is the schema sent to the LLM describing an available tool.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema object
}

// ToolCall is what the LLM requests to execute.
type ToolCall struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

// StepResult is the outcome of a single LLM turn when tools are available.
// Either Content is set (final answer) or ToolCalls is set (need execution).
type StepResult struct {
	Content   string
	ToolCalls []ToolCall
}

// ToolProvider extends Provider with single-turn tool-calling support.
type ToolProvider interface {
	Provider
	CompleteStep(ctx context.Context, model string, messages []Message, tools []ToolDef) (StepResult, error)
}

// ToolExecutor executes a named tool call and returns its result as a string.
type ToolExecutor func(ctx context.Context, name string, args json.RawMessage) string
