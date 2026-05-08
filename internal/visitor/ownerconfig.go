package visitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// OwnerConfig is the persistent owner identity stored in <hostDir>/owner.json.
type OwnerConfig struct {
	Version    int      `json:"version"`
	Enrolled   bool     `json:"enrolled"`
	PhraseHash string   `json:"phrase_hash"`
	Devices    []Device `json:"devices"`
	// ToolTags is the allowlist of tool tag prefixes for this satellite.
	// Empty means all registered tools are available.
	ToolTags   []string `json:"tool_tags,omitempty"`
}

// Device is an enrolled device fingerprint.
type Device struct {
	Hash       string    `json:"hash"`
	Label      string    `json:"label,omitempty"`
	EnrolledAt time.Time `json:"enrolled_at"`
}

// OwnerManager is a thread-safe reader/writer for owner.json in hostDir.
// nil is a valid value — all operations become no-ops and owner mode is disabled.
type OwnerManager struct {
	mu      sync.Mutex
	hostDir string
}

// NewOwnerManager creates an OwnerManager rooted at hostDir.
func NewOwnerManager(hostDir string) *OwnerManager {
	return &OwnerManager{hostDir: hostDir}
}

// Load reads owner.json; returns an empty unenrolled config if the file doesn't exist.
func (m *OwnerManager) Load() (*OwnerConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return loadOwnerConfig(m.hostDir)
}

// Save atomically writes owner.json with 0600 permissions.
func (m *OwnerManager) Save(cfg *OwnerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.hostDir, "owner.json"), data, 0600)
}

// LoadOwnerConfig reads owner.json from hostDir; returns empty unenrolled config if absent.
func LoadOwnerConfig(hostDir string) (*OwnerConfig, error) {
	return loadOwnerConfig(hostDir)
}

func loadOwnerConfig(hostDir string) (*OwnerConfig, error) {
	data, err := os.ReadFile(filepath.Join(hostDir, "owner.json"))
	if os.IsNotExist(err) {
		return &OwnerConfig{Version: 1}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg OwnerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// CheckPhrase returns true if phrase matches the stored bcrypt hash.
func (c *OwnerConfig) CheckPhrase(phrase string) bool {
	if c.PhraseHash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(c.PhraseHash), []byte(phrase)) == nil
}

// KnownDevice returns true if hash is in the enrolled device list.
func (c *OwnerConfig) KnownDevice(hash string) bool {
	for _, d := range c.Devices {
		if d.Hash == hash {
			return true
		}
	}
	return false
}

// Enroll hashes phrase, marks enrolled, and optionally adds a device fingerprint.
func (c *OwnerConfig) Enroll(phrase, deviceHash string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(phrase), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	c.PhraseHash = string(hash)
	c.Enrolled = true
	if deviceHash != "" {
		c.addDevice(deviceHash)
	}
	return nil
}

// AddDevice enrolls a new device fingerprint if not already known.
func (c *OwnerConfig) AddDevice(hash string) {
	c.addDevice(hash)
}

func (c *OwnerConfig) addDevice(hash string) {
	if !c.KnownDevice(hash) {
		c.Devices = append(c.Devices, Device{Hash: hash, EnrolledAt: time.Now()})
	}
}
