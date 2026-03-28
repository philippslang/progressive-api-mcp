// Package config provides configuration types and loading for the OpenAPI MCP server.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for the MCP server.
// It can be loaded from a YAML file or constructed programmatically.
type Config struct {
	Server ServerConfig `yaml:"server"`
	APIs   []APIConfig  `yaml:"apis"`
}

// ServerConfig controls how the MCP server binds and which transport it uses.
type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Transport string `yaml:"transport"`
}

// APIConfig represents one OpenAPI-defined API to load at startup.
type APIConfig struct {
	Name       string `yaml:"name"`
	Definition string `yaml:"definition"`
	Host       string `yaml:"host"`
	BasePath   string `yaml:"base_path"`
}

// Validate returns an error if the Config is invalid.
func (c Config) Validate() error {
	transport := c.Server.Transport
	if transport == "" {
		transport = "http"
	}
	if transport != "http" && transport != "stdio" {
		return fmt.Errorf("server.transport must be \"http\" or \"stdio\", got %q", transport)
	}
	if transport == "http" {
		port := c.Server.Port
		if port == 0 {
			port = 8080
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port)
		}
	}
	if len(c.APIs) == 0 {
		return fmt.Errorf("at least one API must be configured under apis")
	}
	seen := make(map[string]struct{})
	for i, api := range c.APIs {
		if api.Name == "" {
			return fmt.Errorf("apis[%d].name must not be empty", i)
		}
		if api.Definition == "" {
			return fmt.Errorf("apis[%d].definition must not be empty (API: %q)", i, api.Name)
		}
		lower := strings.ToLower(api.Name)
		if _, ok := seen[lower]; ok {
			return fmt.Errorf("duplicate API name %q (names are case-insensitive)", api.Name)
		}
		seen[lower] = struct{}{}
	}
	return nil
}

// LoadFile reads and parses a YAML config file from the given path.
// Returns a validated Config or an error describing what is wrong.
func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config %q: %w", path, err)
	}
	return cfg, nil
}
