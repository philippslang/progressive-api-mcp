# Data Model: OpenAPI MCP Server

**Branch**: `001-openapi-mcp-server` | **Date**: 2026-03-28

## Entities

### Config

Top-level configuration loaded from the YAML config file at startup.

| Field     | Type          | Required | Description |
|-----------|---------------|----------|-------------|
| server    | ServerConfig  | yes      | MCP server binding settings |
| apis      | []APIConfig   | yes      | One or more API definitions to load |

**Validation rules**:
- `apis` MUST contain at least one entry
- All `name` values in `apis` MUST be unique (case-insensitive)

---

### ServerConfig

Settings for how the MCP server binds and which transport it uses.

| Field     | Type   | Required | Default     | Description |
|-----------|--------|----------|-------------|-------------|
| host      | string | no       | `127.0.0.1` | IP/hostname to bind the HTTP transport |
| port      | int    | no       | `8080`      | TCP port to bind the HTTP transport |
| transport | string | no       | `http`      | `http` (Streamable HTTP) or `stdio` |

**Override precedence** (highest to lowest):
1. CLI flag: `--host`, `--port`, `--transport`
2. Environment variable: `PROGRAPIMCP_SERVER_HOST`, `PROGRAPIMCP_SERVER_PORT`,
   `PROGRAPIMCP_SERVER_TRANSPORT`
3. Config file value
4. Default

**Validation rules**:
- `transport` MUST be one of `http` or `stdio`
- `port` MUST be in range 1–65535
- When `transport` is `stdio`, `host` and `port` are ignored (no warning needed)

---

### APIConfig

Configuration for a single OpenAPI-defined API loaded at startup.

| Field      | Type   | Required | Description |
|------------|--------|----------|-------------|
| name       | string | yes      | Unique identifier used as the `api` parameter in MCP tools |
| definition | string | yes      | Path to the OpenAPI 3.x definition file (YAML or JSON); relative to config file directory |
| host       | string | no       | Base host URL (e.g., `https://api.example.com`); overrides `servers[0].url` in the definition |
| base_path  | string | no       | Base path to prepend to all API paths (e.g., `/v2`); overrides the path component of `servers[0].url` |

**URL resolution**:
- Final base URL = `host` + `base_path` if both provided
- Final base URL = `host` if only `host` provided (no base path)
- Final base URL = `servers[0].url` from OpenAPI definition if neither `host` nor `base_path` provided
- If `base_path` is provided without `host`, it is appended to the host from `servers[0].url`

**Validation rules**:
- `name` MUST be a non-empty string matching `[a-zA-Z0-9_-]+`
- `definition` MUST resolve to a readable file at startup
- `host` MUST be a valid URL with scheme (`http://` or `https://`) if provided
- The referenced definition file MUST be valid OpenAPI 3.x; startup fails otherwise

---

### APIEntry (runtime, in-memory)

Represents a fully loaded and validated API, held in the registry at runtime.

| Field      | Type                | Description |
|------------|---------------------|-------------|
| Name       | string              | Matches `APIConfig.name` |
| Config     | APIConfig           | Original config entry |
| Document   | libopenapi.Document | Parsed and validated OpenAPI document |
| Validator  | RequestValidator    | Pre-built request validator for this definition |
| BaseURL    | string              | Resolved base URL (scheme + host + base_path) |

---

### ValidationError (tool output)

Structured error returned when path or schema validation fails before any HTTP call.

| Field   | Type   | Description |
|---------|--------|-------------|
| type    | string | Error category: `PATH_NOT_FOUND`, `MISSING_REQUIRED_PARAM`, `INVALID_PARAM_TYPE`, `MISSING_REQUIRED_FIELD`, `INVALID_FIELD_TYPE`, `ADDITIONAL_PROPERTY`, `SCHEMA_VIOLATION` |
| field   | string | The affected field, parameter name, or path segment (empty for `PATH_NOT_FOUND`) |
| message | string | Human-readable description suitable for agent self-correction |

---

### ToolError (tool output)

Top-level error envelope returned by any MCP tool on failure.

| Field   | Type              | Description |
|---------|-------------------|-------------|
| code    | string            | Error code: `VALIDATION_FAILED`, `PATH_NOT_FOUND`, `AMBIGUOUS_API`, `HTTP_ERROR`, `INTERNAL_ERROR` |
| message | string            | Summary message |
| details | []ValidationError | Field-level validation errors (present only when `code` is `VALIDATION_FAILED`) |
| hints   | []string          | Suggestions for next steps (e.g., list of valid paths when `PATH_NOT_FOUND`) |

---

### PathInfo (tool output — explore_api)

One entry in the exploration tool's result list.

| Field       | Type     | Description |
|-------------|----------|-------------|
| path        | string   | The OpenAPI path template (e.g., `/pets/{id}`) |
| methods     | []string | Supported HTTP methods in uppercase (e.g., `["GET", "PUT", "DELETE"]`) |
| description | string   | `summary` from the OpenAPI operation, or empty string if not defined |

---

### SchemaResult (tool output — get_schema)

Full schema information for one endpoint, returned by the schema tool.

| Field            | Type   | Description |
|------------------|--------|-------------|
| path             | string | The OpenAPI path template |
| method           | string | HTTP method in uppercase |
| path_parameters  | object | JSON Schema for path parameters (omitted if none) |
| query_parameters | object | JSON Schema for query parameters (omitted if none) |
| request_body     | object | JSON Schema for the request body (omitted if not defined) |
| responses        | object | Map of status code → JSON Schema for response (e.g., `{"200": {...}, "404": {...}}`) |

---

### HTTPResult (tool output — http_get / http_post / http_put / http_patch)

The result of a successfully validated and executed HTTP call.

| Field       | Type              | Description |
|-------------|-------------------|-------------|
| status_code | int               | HTTP response status code |
| headers     | map[string]string | Response headers (first value only per header name) |
| body        | any               | Parsed JSON body if `Content-Type` is `application/json`; otherwise raw string |

---

## State Transitions

### Startup Sequence

```
Config file loaded
  → Each APIConfig parsed
    → OpenAPI definition file read from disk
      → Definition parsed by libopenapi
        → Definition validated as OpenAPI 3.x
          → RequestValidator built
            → APIEntry stored in registry
→ All APIs loaded successfully
  → MCP server starts, tools registered
→ Any definition fails → startup aborted with error listing which file failed and why
```

### Request Flow (HTTP Tools)

```
Agent calls tool (e.g., http_post)
  → API identifier resolved (error if ambiguous or unknown)
  → Path matched against OpenAPI path templates
    → No match → ToolError{PATH_NOT_FOUND, hints: [similar paths]}
    → Match found
      → Request parameters + body validated against schema
        → Validation errors → ToolError{VALIDATION_FAILED, details: [ValidationError...]}
        → Validation passes
          → HTTP call executed to BaseURL + path
            → Response received → HTTPResult returned
            → Network error → ToolError{HTTP_ERROR}
```
