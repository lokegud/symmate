package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// OverrideConfig is the owner-configurable LLM routing override.
// Stored as <hostDir>/llm.json. When present, overrides the default provider/model
// for non-privacy-bound requests.
type OverrideConfig struct {
	// Provider: "anthropic" | "hermes" | "openai-compat"
	Provider string `json:"provider"`
	// Model name sent to the provider (e.g. "claude-sonnet-4-6", "hermes", custom).
	Model string `json:"model"`
	// CustomEndpoint is the base URL for openai-compat providers.
	CustomEndpoint string `json:"custom_endpoint,omitempty"`
	// CustomAPIKey for openai-compat providers. NOT logged or stored in audit.
	CustomAPIKey string `json:"custom_api_key,omitempty"`
}

// OverrideStore is a thread-safe reader/writer for llm.json in hostDir.
type OverrideStore struct {
	mu      sync.RWMutex
	hostDir string
	current *OverrideConfig
}

// NewOverrideStore creates an OverrideStore and does an initial load.
func NewOverrideStore(hostDir string) *OverrideStore {
	s := &OverrideStore{hostDir: hostDir}
	cfg, _ := s.load()
	s.current = cfg
	return s
}

func (s *OverrideStore) Get() *OverrideConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Save writes cfg to disk and updates the in-memory copy.
func (s *OverrideStore) Save(cfg *OverrideConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.hostDir, "llm.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	s.mu.Lock()
	s.current = cfg
	s.mu.Unlock()
	return nil
}

func (s *OverrideStore) load() (*OverrideConfig, error) {
	data, err := os.ReadFile(filepath.Join(s.hostDir, "llm.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg OverrideConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
