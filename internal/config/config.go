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

type ServerConfig struct {
	Port int    `toml:"port"`
	Host string `toml:"host"`
}

type HostConfig struct {
	Slug    string `toml:"slug"`
	HostDir string `toml:"host_dir"`
	BaseURL string `toml:"base_url"`

	// Per-surface public URL overrides. When set, these are advertised in the
	// agent card verbatim (no `:port` suffix is appended). Production
	// deployments need HTTPS URLs here per A2A spec §7.1; without overrides,
	// URLs fall back to BaseURL + ":" + port (typically plain HTTP and
	// not spec-compliant).
	A2APublicURL     string `toml:"a2a_public_url"`
	VisitorPublicURL string `toml:"visitor_public_url"`
	MCPPublicURL     string `toml:"mcp_public_url"`
}

type VisitorConfig struct {
	Port          int    `toml:"port"`
	RatePerMin    int    `toml:"rate_per_min"`
	TokenCap      int    `toml:"token_cap"`
	PokerEndpoint string `toml:"poker_endpoint"` // e.g. "http://localhost:7437"; empty = forcefield disabled
}

type MCPConfig struct {
	Port int `toml:"port"`
}

type A2AConfig struct {
	Port    int    `toml:"port"`
	YggAddr string `toml:"ygg_addr"`
}

func Load() (*Config, error) {
	cfg := defaults()

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
	if u := os.Getenv("SYMMATE_A2A_PUBLIC_URL"); u != "" {
		cfg.Host.A2APublicURL = u
	}
	if u := os.Getenv("SYMMATE_VISITOR_PUBLIC_URL"); u != "" {
		cfg.Host.VisitorPublicURL = u
	}
	if u := os.Getenv("SYMMATE_MCP_PUBLIC_URL"); u != "" {
		cfg.Host.MCPPublicURL = u
	}
	if port := os.Getenv("SYMMATE_VISITOR_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Visitor.Port)
	}
	if ep := os.Getenv("POKER_ENDPOINT"); ep != "" {
		cfg.Visitor.PokerEndpoint = ep
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
			Port:          7842,
			RatePerMin:    60,
			TokenCap:      5000,
			PokerEndpoint: "http://localhost:7437",
		},
		MCP: MCPConfig{
			Port: 7844,
		},
		A2A: A2AConfig{
			Port: 7843,
		},
	}
}
