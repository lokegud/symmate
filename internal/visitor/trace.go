package visitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// TraceEntry is one JSONL record emitted per chat request.
type TraceEntry struct {
	Time         string   `json:"time"`                    // RFC3339
	Provider     string   `json:"provider"`                // "anthropic" | "hermes" | "openai-compat"
	Model        string   `json:"model"`                   // model ID
	Tag          string   `json:"tag"`                     // LLM routing tag
	OwnerMode    bool     `json:"owner_mode"`              // owner session
	HistoryLen   int      `json:"history_len"`             // messages in context before current turn
	SysPromptLen int      `json:"sys_prompt_len"`          // byte length of system prompt
	ToolsCalled  []string `json:"tools_called,omitempty"` // tool names invoked
	ResponseMS   int64    `json:"response_ms"`             // wall time: LLM start → channel close
	Error        string   `json:"error,omitempty"`
}

// TraceLogger appends TraceEntry records as JSONL to a file.
// A nil *TraceLogger is valid — all calls become no-ops.
type TraceLogger struct {
	mu   sync.Mutex
	file *os.File
}

// NewTraceLogger opens (or creates) path for append-only JSONL writes.
func NewTraceLogger(path string) (*TraceLogger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &TraceLogger{file: f}, nil
}

func (tl *TraceLogger) Log(e *TraceEntry) {
	if tl == nil {
		return
	}
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	tl.mu.Lock()
	defer tl.mu.Unlock()
	_, _ = tl.file.Write(append(data, '\n'))
}

func (tl *TraceLogger) Close() error {
	if tl == nil {
		return nil
	}
	return tl.file.Close()
}
