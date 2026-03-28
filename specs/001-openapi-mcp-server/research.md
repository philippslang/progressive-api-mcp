# Research: OpenAPI MCP Server

**Branch**: `001-openapi-mcp-server` | **Date**: 2026-03-28

## 1. MCP Library for Go

**Decision**: `github.com/mark3labs/mcp-go`

**Rationale**: Provides first-class support for all three MCP transports (stdio, HTTP SSE,
Streamable HTTP) out of the box. This project requires configurable host/port, which demands
HTTP transport support. The official `modelcontextprotocol/go-sdk` (v1.0.0) is stdio/command
focused and requires external wiring for HTTP transport, making it less practical for this use
case today.

**Alternatives considered**:
- `github.com/modelcontextprotocol/go-sdk` — official Anthropic/Google SDK, v1.0.0 stable,
  guaranteed spec alignment; rejected because HTTP transport is not built-in and must be
  manually bridged. Revisit when official HTTP support matures.
- `github.com/metoro-io/mcp-golang` — type-safe, custom transport support; rejected due to
  lower adoption and unclear long-term maintenance.

---

## 2. OpenAPI 3 Parsing and Request Validation

**Decision**: `github.com/pb33f/libopenapi` + `github.com/pb33f/libopenapi-validator`

**Rationale**: The only Go library combination that satisfies all requirements:
- Full OpenAPI 3.0 and 3.1 parsing (not limited to 3.0 like kin-openapi)
- Pre-flight request validation: path parameters, query parameters, headers, request body
- Path template resolution: maps `/pets/42` → `/pets/{id}` correctly
- Structured error output: `ValidationError` includes field name, message, type, sub-type,
  spec line, and "how to fix" guidance — sufficient for an AI agent to self-correct

**Alternatives considered**:
- `github.com/getkin/kin-openapi` with `openapi3filter` — well-established, simpler API,
  wider adoption; rejected because it is limited to OpenAPI 3.0 (not 3.1) and its validation
  errors lack field-level detail needed for agent self-correction feedback (SC-003).

---

## 3. CLI Configuration Management

**Decision**: `github.com/spf13/cobra` + `github.com/spf13/viper`

**Rationale**: The industry-standard Go combination for CLIs requiring layered config
precedence (flags > env vars > config file > defaults). Used by Kubernetes, Docker CLI, Hugo,
GitHub CLI. Viper natively enforces the exact precedence the spec requires; Cobra auto-generates
`--help`, shell completions, and man pages. The two libraries are designed as companions and
integrate with a single `BindPFlag` call.

**Alternatives considered**:
- `urfave/cli` — lighter weight, struct-based; rejected because it doesn't natively enforce
  the four-level precedence without custom code.
- `peterbourgon/ff` — intentionally minimal; rejected because it lacks the config file
  support needed here.
- `alecthomas/kong` — modern and ergonomic; viable alternative but less established than
  cobra+viper for this combination of requirements.

---

## 4. MCP Transport for Configurable Host/Port

**Decision**: Streamable HTTP as default production transport; stdio available as a flag option
for local development and Claude Desktop integration.

**Rationale**: MCP spec 2025-03-26 designates Streamable HTTP as the current standard for
remote deployments. SSE-based HTTP transport (the previous standard) is now deprecated. Since
the spec requires a configurable host/port (FR-002 companion config, and server config section),
Streamable HTTP is the correct choice. Stdio remains valuable for single-developer local testing
and Claude Desktop plugin mode.

**Transport behavior**:
- `--transport http` (default): binds to configured `host:port`, serves Streamable HTTP
- `--transport stdio`: ignores host/port, communicates over stdin/stdout (for Claude Desktop,
  local agent testing)

**Alternatives considered**:
- SSE transport — deprecated in MCP 2025-03-26; rejected.
- HTTP SSE via `mark3labs/mcp-go` SSE adapter — functionally equivalent to Streamable HTTP
  for this use case but carries deprecation risk.

---

## 5. Configuration File Format

**Decision**: YAML, loaded from a path specified by `--config` flag (default: `./config.yaml`)

**Rationale**: YAML is the most readable format for configuration files that include lists
of API definitions with nested host/path settings. Viper supports YAML natively. TOML and
JSON are viable but less ergonomic for multi-API configs with nested structures.

**Environment variable prefix**: `PROGRAPIMCP_` (e.g., `PROGRAPIMCP_SERVER_HOST`,
`PROGRAPIMCP_SERVER_PORT`). Viper handles prefix stripping automatically.

---

## 6. HTTP Client for Outbound Calls

**Decision**: Standard library `net/http` with a configured `http.Client` (timeout set,
no redirect following by default)

**Rationale**: No additional dependency needed. The standard library client is production-ready,
supports all required HTTP methods, and allows per-request header injection (for agent-provided
auth headers). External HTTP client libraries (resty, go-resty) add dependencies without
meaningful benefit here.

---

## 7. Testing Strategy

**Decision**: `go test` with `net/http/httptest` for integration tests; `testify` for assertions

**Rationale**: `httptest.NewServer` provides real HTTP servers for integration tests without
requiring external services, satisfying the constitution's requirement for real-dependency
integration tests. `testify/assert` and `testify/require` are the Go standard for readable
test assertions. The Petstore OpenAPI 3.1 example spec will be used as a standard test fixture.

No database or external service mocking needed — the only external dependency is the target
HTTP API, which is replaced with `httptest.NewServer` in tests.
