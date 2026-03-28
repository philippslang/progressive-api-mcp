# Go Library API Contract

**Branch**: `001-openapi-mcp-server` | **Date**: 2026-03-28

This document defines the public Go API surface exported by the `pkg/` packages.
All types and functions listed here are part of the stable public API.

---

## Package: `pkg/config`

Import path: `github.com/your-org/prograpimcp/pkg/config`

### Types

```go
// Config is the top-level configuration for the MCP server.
// It can be loaded from a YAML file or constructed programmatically.
type Config struct {
    Server ServerConfig  `yaml:"server"`
    APIs   []APIConfig   `yaml:"apis"`
}

// ServerConfig controls how the MCP server binds and which transport it uses.
type ServerConfig struct {
    Host      string `yaml:"host"`      // Default: "127.0.0.1"
    Port      int    `yaml:"port"`      // Default: 8080
    Transport string `yaml:"transport"` // "http" or "stdio". Default: "http"
}

// APIConfig represents one OpenAPI-defined API to load at startup.
type APIConfig struct {
    Name       string `yaml:"name"`        // Required. Unique identifier for this API.
    Definition string `yaml:"definition"`  // Required. Path to OpenAPI 3.x file (YAML/JSON).
    Host       string `yaml:"host"`        // Optional. Overrides servers[0].url host.
    BasePath   string `yaml:"base_path"`   // Optional. Overrides servers[0].url path component.
}

// Validate returns an error if the Config is invalid.
// Called internally by openapimcp.New(); callers may invoke it for early validation.
func (c Config) Validate() error
```

### Functions

```go
// LoadFile reads and parses a YAML config file from the given path.
// Returns a validated Config or an error describing what is wrong.
func LoadFile(path string) (Config, error)
```

---

## Package: `pkg/openapimcp`

Import path: `github.com/your-org/prograpimcp/pkg/openapimcp`

This is the primary entry point for library consumers.

### Types

```go
// Server is the MCP server instance. Create with New(); start with Start().
// It is safe to call Stop() from a different goroutine than Start().
type Server struct { /* unexported fields */ }
```

### Functions

```go
// New creates a new Server from the given Config.
// It does NOT load or validate the OpenAPI definitions yet.
// Returns an error only if Config.Validate() fails.
func New(cfg config.Config) (*Server, error)

// Start loads all API definitions, registers MCP tools, and starts the transport.
// Blocks until ctx is cancelled or Stop() is called.
// Returns an error if any API definition fails to load or the transport fails to bind.
func (s *Server) Start(ctx context.Context) error

// Stop signals the server to shut down gracefully.
// Returns after all in-flight requests have completed or the deadline is reached.
func (s *Server) Stop() error

// APIs returns the names of all successfully loaded API definitions.
// Only valid after Start() has returned without error.
func (s *Server) APIs() []string
```

### Usage Example

```go
package main

import (
    "context"
    "log"

    "github.com/your-org/prograpimcp/pkg/config"
    "github.com/your-org/prograpimcp/pkg/openapimcp"
)

func main() {
    cfg := config.Config{
        Server: config.ServerConfig{
            Host:      "0.0.0.0",
            Port:      8080,
            Transport: "http",
        },
        APIs: []config.APIConfig{
            {
                Name:       "petstore",
                Definition: "./petstore.yaml",
                Host:       "https://api.example.com",
                BasePath:   "/v2",
            },
        },
    }

    srv, err := openapimcp.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    if err := srv.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

---

## Package: `pkg/validator`

Import path: `github.com/your-org/prograpimcp/pkg/validator`

Exposed so callers can use the validation logic standalone — e.g., inside an existing
HTTP server to validate requests without running a full MCP server.

### Types

```go
// ValidationError describes a single field-level schema violation.
type ValidationError struct {
    Type    string // PATH_NOT_FOUND | MISSING_REQUIRED_PARAM | INVALID_PARAM_TYPE |
                   // MISSING_REQUIRED_FIELD | INVALID_FIELD_TYPE |
                   // ADDITIONAL_PROPERTY | SCHEMA_VIOLATION
    Field   string // Affected parameter or field name; empty for PATH_NOT_FOUND
    Message string // Human-readable description for agent self-correction
}

// Result holds the outcome of validating a single request.
type Result struct {
    Valid  bool
    Errors []ValidationError
}

// Validator validates HTTP requests against a loaded OpenAPI document.
type Validator struct { /* unexported fields */ }
```

### Functions

```go
// New creates a Validator from a libopenapi Document.
// Panics if doc is nil.
func New(doc libopenapi.Document) *Validator

// Validate checks whether the given http.Request conforms to the OpenAPI schema.
// Returns a Result with Valid=true and empty Errors on success.
func (v *Validator) Validate(r *http.Request) Result
```

---

## Package: `pkg/registry`

Import path: `github.com/your-org/prograpimcp/pkg/registry`

Exposed for callers that want to load and query API definitions without the full server.

### Types

```go
// APIEntry is a fully loaded and validated API.
type APIEntry struct {
    Name      string
    Config    config.APIConfig
    BaseURL   string
    Validator *validator.Validator
}

// Registry holds all loaded API entries.
type Registry struct { /* unexported fields */ }
```

### Functions

```go
// New creates an empty Registry.
func New() *Registry

// Load parses, validates, and registers an APIConfig.
// Returns an error if the definition file is missing or fails OpenAPI validation.
func (r *Registry) Load(cfg config.APIConfig) error

// Lookup returns the APIEntry for the given name (case-sensitive).
// Returns false if not found.
func (r *Registry) Lookup(name string) (APIEntry, bool)

// ListNames returns all registered API names in insertion order.
func (r *Registry) ListNames() []string

// Len returns the number of registered APIs.
func (r *Registry) Len() int
```

---

## Stability Policy

- All types and functions in `pkg/config`, `pkg/openapimcp`, `pkg/validator`, and
  `pkg/registry` are part of the **stable public API** as of v1.0.0.
- The `pkg/tools/`, `pkg/loader/`, and `pkg/httpclient/` packages are exported but
  considered **internal implementation detail**; they may change between minor versions.
  Callers should prefer the `pkg/openapimcp` entry point.
- Breaking changes to the stable API require a MAJOR version bump (Go module `v2` path).
