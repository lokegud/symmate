package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server ServerConfig `toml:"server"`
}

type ServerConfig struct {
	Port int    `toml:"port"`
	Host string `toml:"host"`
}

// Load loads config from environment or TOML file.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
	}

	// Override from env
	if port := os.Getenv("SYMMATE_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}
	if host := os.Getenv("SYMMATE_HOST"); host != "" {
		cfg.Server.Host = host
	}

	// Load from config.toml if it exists
	if _, err := os.Stat("config.toml"); err == nil {
		if _, err := toml.DecodeFile("config.toml", cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
