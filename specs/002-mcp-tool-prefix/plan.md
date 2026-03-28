# Implementation Plan: MCP Tool Name Prefix

**Branch**: `002-mcp-tool-prefix` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-mcp-tool-prefix/spec.md`

## Summary

Add an optional `tool_prefix` setting to the server configuration that prepends a
namespacing prefix (e.g., `"myapi"`) to every MCP tool name at registration time.
When set, `http_get` becomes `myapi_http_get`, etc. The prefix follows the same
CLI flag > environment variable > config file > default precedence as all other server
settings. Changes are confined to `pkg/config`, `pkg/openapimcp`, and `cmd/prograpimcp`.

## Technical Context

**Language/Version**: Go 1.22+ (inherits from `001-openapi-mcp-server`)
**Project Type**: Amendment to existing library + CLI
**Primary Dependencies**: No new dependencies
**Storage**: N/A
**Testing**: `go test ./...` — existing test framework; new unit + integration tests added
**Target Platform**: Same as `001`
**Performance Goals**: Startup time unchanged (prefix applied in-memory at registration)
**Constraints**: Prefix validation at startup; no runtime changes
**Scale/Scope**: One new field + one registration-time transformation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | ✅ PASS | One new field, one helper function; complexity unaffected |
| II. Test-First | ✅ PASS | Unit + integration tests written before implementation |
| III. Integration Testing | ✅ PASS | Existing integration test infrastructure reused |
| IV. Performance | ✅ PASS | No performance-sensitive path touched |
| V. Simplicity | ✅ PASS | ~20 lines of new code across 3 files; no abstraction needed |

**Complexity Tracking**: No violations.

## Project Structure

### Documentation (this feature)

```text
specs/002-mcp-tool-prefix/
├── plan.md
├── research.md
├── data-model.md
├── contracts/
│   └── config-schema.md
└── tasks.md
```

### Source Code Changes (repository root)

Only these files change — no new source files needed:

```text
pkg/config/config.go          # Add ToolPrefix to ServerConfig; add validation
tests/unit/config_test.go     # Add ToolPrefix validation tests
pkg/openapimcp/server.go      # Apply prefix at tool registration time
tests/integration/            # Add prefix integration test (new test function)
cmd/prograpimcp/main.go       # Add --tool-prefix flag + viper binding
config.yaml.example           # Document new tool_prefix field
```

## Phase 0: Research Findings

See [research.md](research.md). No external research required — all decisions inherit
from `001-openapi-mcp-server`.

Key decisions:
- `ToolPrefix` added to `ServerConfig` (co-located with `Host`, `Port`, `Transport`)
- Applied in `openapimcp.Start()` via a single helper `applyPrefix(prefix, name string)`
- Validation: must match `^[a-zA-Z_][a-zA-Z0-9_]*$` after stripping trailing `_`
- Trailing `_` stripped before concatenation to prevent double-underscore

## Phase 1: Design

### Data Model

See [data-model.md](data-model.md).

`ServerConfig` gets one new field:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| ToolPrefix | string | no | `""` | Prefix prepended to all MCP tool names. Empty = no prefix. |

**Validation** (added to `Config.Validate()`):
- If non-empty after stripping trailing `_`: MUST match `^[a-zA-Z_][a-zA-Z0-9_]*$`
- Purely-numeric values (e.g., `"123"`) are rejected as invalid

**Tool name construction**:
```
effective := strings.TrimRight(cfg.Server.ToolPrefix, "_")
if effective != "" {
    toolName = effective + "_" + baseName
} else {
    toolName = baseName
}
```

### CLI / Environment Contract

| Layer | Key |
|-------|-----|
| CLI flag | `--tool-prefix` |
| Environment variable | `PROGRAPIMCP_SERVER_TOOL_PREFIX` |
| Config file | `server.tool_prefix` |
| Default | `""` (no prefix) |

### Updated Config Schema

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  transport: "http"
  tool_prefix: "myapi"   # optional; omit or leave empty for no prefix
```

### Post-Phase-1 Constitution Re-Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | ✅ PASS | All exported fields documented; helper is ≤5 lines |
| II. Test-First | ✅ PASS | Tests written first, confirmed failing before implementation |
| III. Integration Testing | ✅ PASS | InProcess transport verifies prefix applied correctly |
| IV. Performance | ✅ PASS | No hot path touched |
| V. Simplicity | ✅ PASS | No abstraction beyond what the task requires |
