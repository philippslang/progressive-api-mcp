# Implementation Plan: OpenAPI MCP Server

**Branch**: `001-openapi-mcp-server` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-openapi-mcp-server/spec.md`

## Summary

Build a Go CLI that loads one or more OpenAPI 3.x definitions at startup, validates them, and
exposes six MCP tools to an AI agent: four general-purpose HTTP tools (GET, POST, PUT, PATCH)
that validate requests against the loaded schema before executing them, an exploration tool for
progressive path discovery, and a schema retrieval tool. A single YAML configuration file
specifies all APIs and server settings; CLI flags and environment variables override config
values following conventional precedence.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**:
- `github.com/mark3labs/mcp-go` — MCP server with Streamable HTTP + stdio transports
- `github.com/pb33f/libopenapi` — OpenAPI 3.0/3.1 document parsing
- `github.com/pb33f/libopenapi-validator` — Pre-flight request validation with field-level errors
- `github.com/spf13/cobra` — CLI command and flag definition
- `github.com/spf13/viper` — Layered config: flags > env vars > config file > defaults
- `github.com/stretchr/testify` — Test assertions

**Storage**: Files only (config YAML, OpenAPI definition files on disk)
**Testing**: `go test ./...` with `net/http/httptest` for integration tests
**Target Platform**: Linux/macOS server; single static binary
**Performance Goals**: Validated HTTP call ≤ 2s (SC-005); startup with 10 APIs ≤ 3s (SC-006)
**Constraints**: No runtime database; no external service dependencies in tests
**Scale/Scope**: Single binary, up to ~10 APIs loaded simultaneously per instance

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | ✅ PASS | Single-responsibility packages; golangci-lint in CI; no magic numbers |
| II. Test-First | ✅ PASS | TDD mandatory; tests written before implementation; 80%+ coverage enforced |
| III. Integration Testing | ✅ PASS | `httptest.NewServer` used for all integration tests; no mocks for owned code |
| IV. Performance | ✅ PASS | SC-005 and SC-006 define measurable gates; benchmarks required for hot paths |
| V. Simplicity | ✅ PASS | Standard library HTTP client; no abstraction without 2 concrete use cases; YAGNI |

**Complexity Tracking**: No violations — no table needed.

## Project Structure

### Documentation (this feature)

```text
specs/001-openapi-mcp-server/
├── plan.md           # This file
├── research.md       # Phase 0 output
├── data-model.md     # Phase 1 output
├── quickstart.md     # Phase 1 output
├── contracts/
│   └── mcp-tools.md  # Phase 1 output — MCP tool input/output contracts
└── tasks.md          # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/
└── prograpimcp/
    └── main.go                   # Entry point; wires cobra root command

internal/
├── config/
│   └── config.go                 # Config struct, YAML loading, viper/cobra binding
├── loader/
│   └── loader.go                 # Read OpenAPI files; parse + validate with libopenapi
├── registry/
│   └── registry.go               # In-memory map[name]APIEntry; lookup + ambiguity check
├── validator/
│   └── validator.go              # Wraps libopenapi-validator; builds ValidationError slice
├── tools/
│   ├── http.go                   # http_get, http_post, http_put, http_patch tool handlers
│   ├── explore.go                # explore_api tool handler
│   └── schema.go                 # get_schema tool handler
├── httpclient/
│   └── client.go                 # Thin wrapper around net/http; executes outbound calls
└── server/
    └── server.go                 # MCP server init; tool registration; transport selection

tests/
├── integration/
│   ├── http_tools_test.go        # End-to-end: load real definition, call tools via httptest
│   └── explore_schema_test.go    # End-to-end: explore + schema tools
├── unit/
│   ├── config_test.go
│   ├── loader_test.go
│   ├── validator_test.go
│   └── registry_test.go
├── contract/
│   └── mcp_tools_contract_test.go # Tool input/output shape matches contracts/mcp-tools.md
└── testdata/
    ├── petstore.yaml             # OpenAPI 3.1 Petstore fixture
    ├── bookstore.yaml            # Second API for multi-API tests
    └── malformed.yaml            # Invalid definition for startup-failure tests

go.mod
go.sum
config.yaml.example
```

**Structure Decision**: Single-project layout with `internal/` packages for isolation and
`cmd/prograpimcp/` as the binary entry point. This is idiomatic Go for a CLI tool with
multiple internal concerns. No multi-module workspace needed.

## Phase 0: Research Findings

See [research.md](research.md) for full rationale. Key decisions:

- **MCP library**: `mark3labs/mcp-go` — full HTTP transport (Streamable HTTP + stdio)
- **OpenAPI validation**: `pb33f/libopenapi` + `pb33f/libopenapi-validator` — OpenAPI 3.1,
  field-level errors
- **Config/CLI**: `cobra` + `viper` — flags > env > config file > defaults
- **Transport**: Streamable HTTP default; stdio via `--transport stdio` flag
- **HTTP client**: `net/http` standard library — no external dependency needed

## Phase 1: Design

### Data Model

See [data-model.md](data-model.md) for full entity definitions.

Key types:
- `Config` → `ServerConfig` + `[]APIConfig`
- `APIEntry` (runtime) → loaded document + validator + resolved base URL
- `ValidationError` → type, field, message
- `ToolError` → code, message, details, hints
- `PathInfo`, `SchemaResult`, `HTTPResult` → tool outputs

### MCP Tool Contracts

See [contracts/mcp-tools.md](contracts/mcp-tools.md) for full input/output schemas.

Tools: `http_get`, `http_post`, `http_put`, `http_patch`, `explore_api`, `get_schema`

All tools share the common `api` parameter behavior (required when multiple APIs loaded;
omittable when exactly one API is loaded; error with hints when ambiguous).

### Configuration

**File format**: YAML
**Default path**: `./config.yaml` (overridable via `--config` flag)
**Env prefix**: `PROGRAPIMCP_`

```yaml
server:
  host: "127.0.0.1"          # PROGRAPIMCP_SERVER_HOST / --host
  port: 8080                  # PROGRAPIMCP_SERVER_PORT / --port
  transport: "http"           # PROGRAPIMCP_SERVER_TRANSPORT / --transport

apis:
  - name: petstore
    definition: "./petstore.yaml"
    host: "https://api.example.com"   # optional
    base_path: "/v2"                   # optional
```

### Startup Sequence

1. Parse CLI flags (cobra)
2. Load config file via viper; bind flags to viper keys
3. For each `APIConfig`: read definition file → parse with libopenapi → validate structure →
   build request validator → store `APIEntry` in registry
4. If any definition fails: print error with file name and reason → exit 1
5. Register 6 MCP tools with the server
6. Start transport (Streamable HTTP or stdio)

### Request Validation Flow

For `http_post path=/pets body={species:"dog"}`:

1. Resolve API from `api` param (or single loaded API)
2. Match `/pets` against path templates → found: `POST /pets`
3. Build synthetic `http.Request` from tool inputs
4. Call `libopenapi-validator.ValidateHttpRequest(req)` against the API document
5. Collect validation errors → return `ToolError{VALIDATION_FAILED, details:[...]}`
6. On pass: execute `net/http` POST to `BaseURL + /pets` with provided body and headers
7. Return `HTTPResult{status_code, headers, body}`

### Testing Strategy

- **Unit tests**: Each `internal/` package tested in isolation; table-driven tests for
  validation edge cases (missing field, wrong type, extra property, path template matching)
- **Integration tests**: Load real `testdata/petstore.yaml`; start `httptest.NewServer` as
  target; call tools; assert validation behavior and HTTP round-trips
- **Contract tests**: For each tool in `contracts/mcp-tools.md`, assert that valid inputs
  produce `HTTPResult` shape and invalid inputs produce `ToolError` shape
- **Benchmark**: `BenchmarkValidatedHTTPCall` to track SC-005 regression gate

### Post-Phase-1 Constitution Re-Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | ✅ PASS | 7 single-purpose packages; no cross-cutting state |
| II. Test-First | ✅ PASS | Contract + integration tests defined before implementation starts |
| III. Integration Testing | ✅ PASS | `httptest.NewServer` replaces real API; no mocks for owned code |
| IV. Performance | ✅ PASS | `BenchmarkValidatedHTTPCall` guards SC-005; startup test guards SC-006 |
| V. Simplicity | ✅ PASS | 6 packages, 1 binary, no abstraction without 2 concrete use cases |
