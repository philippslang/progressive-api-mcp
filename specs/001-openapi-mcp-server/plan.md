# Implementation Plan: OpenAPI MCP Server

**Branch**: `001-openapi-mcp-server` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-openapi-mcp-server/spec.md`

**Amendment (2026-03-28)**: Core functionality also exposed as a Go library so other servers
and CLIs can embed and use it without running a separate process.

## Summary

Build a Go library + CLI that loads one or more OpenAPI 3.x definitions at startup, validates
them, and exposes six MCP tools to an AI agent: four general-purpose HTTP tools (GET, POST, PUT,
PATCH) that validate requests against the loaded schema before executing them, an exploration
tool for progressive path discovery, and a schema retrieval tool.

The core logic is packaged as an importable Go library (`pkg/`) so that other CLIs and servers
can embed the MCP server functionality directly. The standalone CLI (`cmd/prograpimcp/`) is a
thin consumer of the library. A single YAML configuration file specifies all APIs and server
settings; CLI flags and environment variables override config values following conventional
precedence.

## Technical Context

**Language/Version**: Go 1.22+
**Project Type**: Go library (`pkg/`) + CLI binary (`cmd/prograpimcp/`)
**Primary Dependencies**:
- `github.com/mark3labs/mcp-go` — MCP server with Streamable HTTP + stdio transports
- `github.com/pb33f/libopenapi` — OpenAPI 3.0/3.1 document parsing
- `github.com/pb33f/libopenapi-validator` — Pre-flight request validation with field-level errors
- `github.com/spf13/cobra` — CLI command and flag definition (CLI layer only)
- `github.com/spf13/viper` — Layered config: flags > env vars > config file > defaults (CLI only)
- `github.com/stretchr/testify` — Test assertions

**Storage**: Files only (config YAML, OpenAPI definition files on disk)
**Testing**: `go test ./...` with `net/http/httptest` for integration tests
**Target Platform**: Linux/macOS; single static binary + importable library
**Performance Goals**: Validated HTTP call ≤ 2s (SC-005); startup with 10 APIs ≤ 3s (SC-006)
**Constraints**: No runtime database; no external service dependencies in tests; library API
must not import cobra/viper (those are CLI concerns only)
**Scale/Scope**: Single binary, up to ~10 APIs loaded simultaneously per instance

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | ✅ PASS | Single-responsibility packages; golangci-lint in CI; exported API documented |
| II. Test-First | ✅ PASS | TDD mandatory; tests written before implementation; ≥80% coverage enforced |
| III. Integration Testing | ✅ PASS | `httptest.NewServer` for integration tests; no mocks for owned code |
| IV. Performance | ✅ PASS | SC-005 and SC-006 define measurable gates; benchmarks required for hot paths |
| V. Simplicity | ✅ PASS | Library has no CLI dependencies; CLI is a thin wrapper; YAGNI |

**Complexity Tracking**:

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| Two output artifacts (library + binary) | Other Go programs must be able to `import` and embed the MCP server without spawning a subprocess | A single binary cannot be imported; an `internal/` package cannot be used by external consumers |

## Project Structure

### Documentation (this feature)

```text
specs/001-openapi-mcp-server/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   ├── mcp-tools.md     # MCP tool input/output contracts
│   └── library-api.md   # Go public library API surface
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
pkg/
├── openapimcp/
│   └── server.go             # Public entry point: New(), Server.Start(), Server.Stop()
├── config/
│   └── config.go             # Config struct + LoadFile(); no cobra/viper dependency
├── loader/
│   └── loader.go             # Load OpenAPI file → parsed Document + RequestValidator
├── registry/
│   └── registry.go           # In-memory map[name]APIEntry; Lookup(), ListNames()
├── validator/
│   └── validator.go          # Wraps libopenapi-validator; returns []ValidationError
├── tools/
│   ├── http.go               # http_get/post/put/patch MCP tool handlers
│   ├── explore.go            # explore_api MCP tool handler
│   └── schema.go             # get_schema MCP tool handler
└── httpclient/
    └── client.go             # Outbound HTTP call executor (net/http wrapper)

cmd/
└── prograpimcp/
    └── main.go               # CLI entry point: cobra root cmd + viper config wiring

tests/
├── integration/
│   ├── http_tools_test.go    # End-to-end via httptest.NewServer
│   └── explore_schema_test.go
├── unit/
│   ├── config_test.go
│   ├── loader_test.go
│   ├── validator_test.go
│   └── registry_test.go
├── contract/
│   └── mcp_tools_contract_test.go
└── testdata/
    ├── petstore.yaml          # OpenAPI 3.1 Petstore fixture
    ├── bookstore.yaml         # Second API for multi-API tests
    └── malformed.yaml         # Invalid definition for startup-failure tests

go.mod
go.sum
config.yaml.example
```

**Structure Decision**: All reusable logic lives in `pkg/` (exportable). The CLI (`cmd/`) uses
cobra and viper to wire configuration into the library's `Config` struct, then calls
`openapimcp.New(cfg).Start()`. No `internal/` packages — the library is the product.

## Phase 0: Research Findings

See [research.md](research.md) for full rationale. Key decisions:

- **MCP library**: `mark3labs/mcp-go` — full HTTP transport (Streamable HTTP + stdio)
- **OpenAPI validation**: `pb33f/libopenapi` + `pb33f/libopenapi-validator` — OpenAPI 3.1,
  field-level errors
- **Config/CLI**: `cobra` + `viper` in CLI layer only; library uses plain structs
- **Transport**: Streamable HTTP default; stdio via `--transport stdio` flag
- **HTTP client**: `net/http` standard library — no external dependency needed
- **Library pattern**: Thin public API in `pkg/openapimcp/`; implementation details in sibling
  `pkg/` packages; CLI is ~50 lines wrapping the library

**Amendment research** — Go library design:
- Decision: Single `pkg/openapimcp` package provides `New(Config) *Server`, `Server.Start(ctx)`,
  `Server.Stop()`. All other `pkg/` packages are importable but callers can also use them
  individually (e.g., embed only the validator logic without the full server).
- Rationale: Flat `pkg/` layout matches idiomatic Go; no deep abstraction; callers import
  exactly what they need. An `openapimcp.Server` struct composes the registry, validator, and
  MCP server internally.
- Alternatives considered: Single-package library (`package openapimcp` for everything) —
  rejected because it would bundle MCP server concerns with config loading and validation,
  making unit testing of individual pieces harder.

## Phase 1: Design

### Data Model

See [data-model.md](data-model.md) for full entity definitions.

Key types (unchanged from original plan, now in `pkg/`):
- `Config` → `ServerConfig` + `[]APIConfig`
- `APIEntry` (runtime) → loaded document + validator + resolved base URL
- `ValidationError` → type, field, message
- `ToolError` → code, message, details, hints
- `PathInfo`, `SchemaResult`, `HTTPResult` → tool outputs

### Public Library API Surface

See [contracts/library-api.md](contracts/library-api.md) for full Go API.

Minimum viable API:

```go
import "github.com/your-org/prograpimcp/pkg/openapimcp"
import "github.com/your-org/prograpimcp/pkg/config"

cfg, err := config.LoadFile("config.yaml")
srv := openapimcp.New(cfg)
err = srv.Start(ctx)   // blocks until ctx cancelled or Stop() called
srv.Stop()
```

Programmatic config (no file):

```go
cfg := config.Config{
    Server: config.ServerConfig{Host: "0.0.0.0", Port: 9090, Transport: "http"},
    APIs: []config.APIConfig{
        {Name: "petstore", Definition: "./petstore.yaml", Host: "https://api.example.com"},
    },
}
srv := openapimcp.New(cfg)
```

### MCP Tool Contracts

See [contracts/mcp-tools.md](contracts/mcp-tools.md) for full input/output schemas.

Tools: `http_get`, `http_post`, `http_put`, `http_patch`, `explore_api`, `get_schema`.
Unchanged from original plan.

### Configuration

**File format**: YAML (loaded via `config.LoadFile`)
**Default path**: `./config.yaml` (overridable via `--config` flag in CLI)
**Env prefix in CLI**: `PROGRAPIMCP_` (cobra+viper wiring in `cmd/` only)
**Library consumers**: pass `config.Config` struct directly; no file required

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  transport: "http"   # "http" or "stdio"

apis:
  - name: petstore
    definition: "./petstore.yaml"
    host: "https://api.example.com"   # optional
    base_path: "/v2"                   # optional
```

### Startup Sequence (unchanged, now in `openapimcp.Start`)

1. Validate `Config` struct (no cobra/viper involved)
2. For each `APIConfig`: read definition file → parse with libopenapi → validate structure →
   build request validator → store `APIEntry` in registry
3. If any definition fails: return descriptive error
4. Register 6 MCP tools with `mcp-go` server
5. Start transport (Streamable HTTP or stdio)

### Request Validation Flow (unchanged)

For `http_post path=/pets body={species:"dog"}`:
1. Resolve API from `api` param
2. Match `/pets` against path templates → `POST /pets`
3. Build synthetic `http.Request`; call `libopenapi-validator.ValidateHttpRequest`
4. Validation errors → return `ToolError{VALIDATION_FAILED, details:[...]}`
5. On pass: execute `net/http` POST → return `HTTPResult`

### Testing Strategy

- **Unit tests**: Each `pkg/` package tested independently; table-driven tests
- **Integration tests**: `testdata/petstore.yaml` + `httptest.NewServer`; cover both
  library-level calls and CLI-invoked calls
- **Contract tests**: Tool input/output shapes; library API shapes
- **Library embedding test**: A `tests/integration/embedding_test.go` that imports
  `pkg/openapimcp` programmatically (no CLI) and runs a full tool invocation — validates
  the library is importable and functional independently of the CLI
- **Benchmark**: `BenchmarkValidatedHTTPCall` guards SC-005

### Post-Phase-1 Constitution Re-Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | ✅ PASS | All exported symbols documented; cyclomatic complexity ≤10 per func |
| II. Test-First | ✅ PASS | Library embedding test written before implementation |
| III. Integration Testing | ✅ PASS | Real httptest server; no mocks for owned code |
| IV. Performance | ✅ PASS | Benchmarks guard SC-005; startup test guards SC-006 |
| V. Simplicity | ✅ PASS | Library has zero CLI dependencies; CLI is ~50 lines |
