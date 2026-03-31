// Package config provides configuration types and loading for the OpenAPI MCP server.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var validToolPrefix = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// knownToolNames is the fixed set of base MCP tool names.
var knownToolNames = map[string]struct{}{
	"explore_api": {},
	"get_schema":  {},
	"http_get":    {},
	"http_post":   {},
	"http_put":    {},
	"http_patch":  {},
}

const knownToolNamesHint = "explore_api, get_schema, http_get, http_post, http_put, http_patch"

// APIAllowList restricts which MCP tools and paths are available for one API entry.
// Zero value means no restrictions: all tools registered, all paths accessible.
type APIAllowList struct {
	// Tools is the list of base tool names to register for this API.
	// Nil or empty means all 6 tools are allowed.
	// Valid values: "explore_api", "get_schema", "http_get", "http_post", "http_put", "http_patch".
	Tools []string `yaml:"tools"`
	// Paths maps each base tool name to the OpenAPI path templates it may access.
	// Nil or empty map means all paths are allowed for all tools.
	// An empty slice for a tool means the tool is registered but no paths are accessible.
	Paths map[string][]string `yaml:"paths"`
}

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
	// ToolPrefix is an optional prefix prepended to all MCP tool names (e.g. "myapi" → "myapi_http_get").
	// Must start with a letter or underscore; only letters, digits, and underscores are allowed.
	// A trailing underscore is stripped automatically. Empty means no prefix.
	ToolPrefix string `yaml:"tool_prefix"`
}

// APIConfig represents one OpenAPI-defined API to load at startup.
type APIConfig struct {
	Name       string       `yaml:"name"`
	Definition string       `yaml:"definition"`
	Host       string       `yaml:"host"`
	BasePath   string       `yaml:"base_path"`
	// AllowList restricts which tools and paths are available for this API.
	// Zero value means no restrictions (all tools and paths allowed).
	AllowList APIAllowList `yaml:"allow_list"`
	// IgnoreHeaders is a list of header names (case-insensitive) to suppress.
	// Ignored headers are excluded from get_schema output and are not required
	// during request validation for any HTTP tool.
	IgnoreHeaders []string `yaml:"ignore_headers"`
	// ResponseBodyOnly controls the shape of HTTP tool responses.
	// When true, HTTP tools return only the response body instead of the full
	// { status_code, headers, body } envelope. Default: false.
	ResponseBodyOnly bool `yaml:"response_body_only"`
	// SkipValidation disables request payload schema validation for this API.
	// When true, tool calls bypass the MCP server's validator and forward the
	// payload directly to the upstream API. Default: false (validation enabled).
	SkipValidation bool `yaml:"skip_validation"`
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
	prefix := strings.TrimRight(c.Server.ToolPrefix, "_")
	if prefix != "" && !validToolPrefix.MatchString(prefix) {
		return fmt.Errorf("server.tool_prefix %q is invalid: must start with a letter or underscore and contain only letters, digits, and underscores", c.Server.ToolPrefix)
	}
	seen := make(map[string]struct{})
	for i, api := range c.APIs {
		if api.Name == "" {
			return fmt.Errorf("apis[%d].name must not be empty", i)
		}
		if api.Definition == "" {
			return fmt.Errorf("apis[%d].definition must not be empty (API: %q)", i, api.Name)
		}
		for j, tool := range api.AllowList.Tools {
			if _, ok := knownToolNames[tool]; !ok {
				return fmt.Errorf("apis[%d].allow_list.tools[%d]: unknown tool name %q; valid names are: %s", i, j, tool, knownToolNamesHint)
			}
		}
		for tool := range api.AllowList.Paths {
			if _, ok := knownToolNames[tool]; !ok {
				return fmt.Errorf("apis[%d].allow_list.paths: unknown tool name %q; valid names are: %s", i, tool, knownToolNamesHint)
			}
		}
		for j, h := range api.IgnoreHeaders {
			if h == "" {
				return fmt.Errorf("apis[%d].ignore_headers[%d]: header name must not be empty", i, j)
			}
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
