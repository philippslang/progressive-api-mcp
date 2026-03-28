# Data Model: MCP Tool Name Prefix

**Branch**: `002-mcp-tool-prefix` | **Date**: 2026-03-28

## Modified Entity: `ServerConfig`

One new field added to the existing `ServerConfig` struct.

| Field | Type | YAML tag | Required | Default | Validation |
|-------|------|----------|----------|---------|------------|
| Host | string | `host` | no | `"127.0.0.1"` | (existing) |
| Port | int | `port` | no | `8080` | (existing) |
| Transport | string | `transport` | no | `"http"` | (existing) |
| **ToolPrefix** | **string** | **`tool_prefix`** | **no** | **`""`** | See below |

### ToolPrefix Validation Rules

Applied in `Config.Validate()` only when `ToolPrefix` is non-empty:

1. Strip trailing `_` from the value → `effective`
2. If `effective` is empty after stripping: treat as no prefix (valid, no error)
3. `effective` MUST match `^[a-zA-Z_][a-zA-Z0-9_]*$`
   - MUST start with a letter (`a-z`, `A-Z`) or underscore `_`
   - Remaining characters MUST be letters, digits, or underscores
   - Purely-numeric values (`"123"`, `"007"`) are rejected
4. If the rule is violated: error message MUST identify the invalid characters

### Tool Name Construction (runtime, in `openapimcp.Start()`)

```
effective_prefix = strings.TrimRight(ServerConfig.ToolPrefix, "_")

if effective_prefix != "" {
    registered_name = effective_prefix + "_" + base_tool_name
} else {
    registered_name = base_tool_name
}
```

**Examples**:

| ToolPrefix config | effective_prefix | base_name | registered_name |
|-------------------|-----------------|-----------|----------------|
| `"myapi"` | `"myapi"` | `"http_get"` | `"myapi_http_get"` |
| `"myapi_"` | `"myapi"` | `"http_get"` | `"myapi_http_get"` |
| `""` | `""` | `"http_get"` | `"http_get"` |
| `"v2_svc"` | `"v2_svc"` | `"explore_api"` | `"v2_svc_explore_api"` |

## No New Entities

This feature does not introduce new runtime entities or data structures. `ToolPrefix` is a
scalar string value consumed once at startup and discarded after tool registration.
