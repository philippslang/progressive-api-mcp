# Research: MCP Tool Name Prefix

**Branch**: `002-mcp-tool-prefix` | **Date**: 2026-03-28

No external research required. All technology decisions inherit from `001-openapi-mcp-server`.

## 1. Where to add the new field

**Decision**: Add `ToolPrefix string` to `ServerConfig` in `pkg/config/config.go`.

**Rationale**: `ServerConfig` already holds the four peer settings (`Host`, `Port`,
`Transport`). Adding `ToolPrefix` there keeps all server-level settings together and allows
it to be driven by the same Viper/Cobra precedence mechanism with minimal new code.

**Alternatives considered**:
- Top-level `Config` field — rejected; it is logically a server setting, not a global one.
- Per-`APIConfig` field — rejected; spec says server-wide (FR-008).

## 2. Where to apply the prefix

**Decision**: Apply in `pkg/openapimcp/server.go` inside `Start()`, wrapping the tool name
strings passed to `tools.Register*` functions.

**Rationale**: Tool registration is centralized in `Start()`. The simplest change is to pass
the effective prefix into each `Register*` function and have it prepend the prefix before
calling `mcp-go`'s `AddTool`. No new package, no new abstraction.

**Implementation sketch**:
```go
prefix := applyPrefix(s.cfg.Server.ToolPrefix)
tools.RegisterHTTPTools(s.mcpSrv, s.registry, s.client, prefix)
tools.RegisterExploreTools(s.mcpSrv, s.registry, prefix)
tools.RegisterSchemaTools(s.mcpSrv, s.registry, prefix)
```

Each `Register*` function gains a `prefix string` parameter and prepends it to the name
passed to `mcp.NewTool(name, ...)`.

## 3. Validation rule

**Decision**: After stripping a trailing `_`, the prefix MUST match `^[a-zA-Z_][a-zA-Z0-9_]*$`.

**Rationale**: This mirrors standard identifier rules in most languages and ensures
`prefix_toolname` is itself a valid identifier. It rejects spaces, hyphens, and purely-numeric
prefixes. Starting with a letter or `_` (not a digit) is the conventional identifier rule.

## 4. Trailing underscore handling

**Decision**: Strip any trailing `_` from the provided prefix before use, silently.

**Rationale**: Users may naturally write `tool_prefix: "myapi_"` expecting the tool to be
`myapi_http_get`. Stripping the trailing `_` before joining is a zero-friction fix that
avoids surprising `myapi__http_get` output. FR-007 explicitly requires this behavior.
