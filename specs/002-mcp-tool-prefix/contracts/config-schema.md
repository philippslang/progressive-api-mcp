# Config Schema: MCP Tool Name Prefix

**Branch**: `002-mcp-tool-prefix` | **Date**: 2026-03-28

This document shows the updated `server` section of the configuration schema,
with the new `tool_prefix` field added.

## Updated `server` Config Block

```yaml
server:
  host: "127.0.0.1"          # PROGRAPIMCP_SERVER_HOST      / --host
  port: 8080                  # PROGRAPIMCP_SERVER_PORT      / --port
  transport: "http"           # PROGRAPIMCP_SERVER_TRANSPORT / --transport
  tool_prefix: "myapi"        # PROGRAPIMCP_SERVER_TOOL_PREFIX / --tool-prefix
                              # Optional. Omit or leave empty for no prefix.
```

## Field: `tool_prefix`

| Property | Value |
|----------|-------|
| YAML key | `server.tool_prefix` |
| CLI flag | `--tool-prefix` |
| Env var | `PROGRAPIMCP_SERVER_TOOL_PREFIX` |
| Type | string |
| Required | No |
| Default | `""` (empty — no prefix applied) |
| Valid values | Alphanumeric + underscore; must start with letter or `_`; not purely numeric |
| Trailing `_` | Stripped automatically before use |

## Precedence (highest to lowest)

1. `--tool-prefix` CLI flag
2. `PROGRAPIMCP_SERVER_TOOL_PREFIX` environment variable
3. `server.tool_prefix` in config file
4. Default: `""` (no prefix)

## Effect on Registered Tool Names

When `tool_prefix` is set to `"myapi"`, the six MCP tools are registered as:

| Original name | Registered name |
|---------------|----------------|
| `http_get` | `myapi_http_get` |
| `http_post` | `myapi_http_post` |
| `http_put` | `myapi_http_put` |
| `http_patch` | `myapi_http_patch` |
| `explore_api` | `myapi_explore_api` |
| `get_schema` | `myapi_get_schema` |

When `tool_prefix` is empty or absent, the original names are used unchanged.

## Validation Errors

| Condition | Error message (example) |
|-----------|------------------------|
| Starts with digit | `server.tool_prefix "123abc" is invalid: must start with a letter or underscore` |
| Contains space | `server.tool_prefix "my api" is invalid: only letters, digits, and underscores are allowed` |
| Contains hyphen | `server.tool_prefix "my-api" is invalid: only letters, digits, and underscores are allowed` |
