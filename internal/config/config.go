package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server  ServerConfig  `toml:"server"`
	Host    HostConfig    `toml:"host"`
	Visitor VisitorConfig `toml:"visitor"`
	MCP     MCPConfig     `toml:"mcp"`
	A2A     A2AConfig     `toml:"a2a"`
}

// ServerConfig is the legacy health/version server (kept for compatibility).
type ServerConfig struct {
	Port int    `toml:"port"`
	Host string `toml:"host"`
}

// HostConfig describes the Symmate host identity.
type HostConfig struct {
	Slug    string `toml:"slug"`    // e.g. "loke"
	HostDir string `toml:"host_dir"` // e.g. "/var/lib/symmate/loke"
	BaseURL string `toml:"base_url"` // e.g. "http://localhost" (no port)
}

// VisitorConfig configures the visitor API listener on :7842.
type VisitorConfig struct {
	Port       int `toml:"port"`
	RatePerMin int `toml:"rate_per_min"` // per-IP request rate limit
	TokenCap   int `toml:"token_cap"`    // per-session token cap
}

// MCPConfig configures the MCP server listener on :7844.
type MCPConfig struct {
	Port int `toml:"port"`
}

// A2AConfig configures the A2A endpoint.
type A2AConfig struct {
	Port    int    `toml:"port"`
	YggAddr string `toml:"ygg_addr"` // optional extra listener on Yggdrasil interface, e.g. "[200:xxxx::1]:7843"
}

// Load loads config from environment or TOML file, applying defaults.
func Load() (*Config, error) {
	cfg := defaults()

	// Override from env
	if port := os.Getenv("SYMMATE_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}
	if host := os.Getenv("SYMMATE_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if slug := os.Getenv("SYMMATE_SLUG"); slug != "" {
		cfg.Host.Slug = slug
	}
	if dir := os.Getenv("SYMMATE_HOST_DIR"); dir != "" {
		cfg.Host.HostDir = dir
	}
	if base := os.Getenv("SYMMATE_BASE_URL"); base != "" {
		cfg.Host.BaseURL = base
	}
	if port := os.Getenv("SYMMATE_VISITOR_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Visitor.Port)
	}
	if port := os.Getenv("SYMMATE_MCP_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.MCP.Port)
	}
	if port := os.Getenv("SYMMATE_A2A_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.A2A.Port)
	}
	if addr := os.Getenv("SYMMATE_A2A_YGG_ADDR"); addr != "" {
		cfg.A2A.YggAddr = addr
	}

	// Load from config.toml if present
	if _, err := os.Stat("config.toml"); err == nil {
		if _, err := toml.DecodeFile("config.toml", cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Host: HostConfig{
			Slug:    "default",
			HostDir: "/var/lib/symmate/default",
			BaseURL: "http://localhost",
		},
		Visitor: VisitorConfig{
			Port:       7842,
			RatePerMin: 60,
			TokenCap:   5000,
		},
		MCP: MCPConfig{
			Port: 7844,
		},
		A2A: A2AConfig{
			Port: 7843,
		},
	}
}
