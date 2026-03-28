# MCP Tool Contracts

**Branch**: `001-openapi-mcp-server` | **Date**: 2026-03-28

These contracts define the exact input parameters and output structures for each MCP tool
exposed by the server. All tools return JSON-serializable results.

---

## Common Parameter: `api`

All tools accept an optional `api` parameter (string):
- **When a single API is loaded**: `api` may be omitted; the single loaded API is used.
- **When multiple APIs are loaded**: `api` is required. If omitted, the tool returns
  `ToolError{code: "AMBIGUOUS_API", hints: [list of available API names]}`.
- Value MUST match the `name` field from the configuration exactly (case-sensitive).

---

## Tool: `http_get`

Execute a validated HTTP GET request against a loaded API.

### Input

| Parameter    | Type             | Required | Description |
|--------------|------------------|----------|-------------|
| api          | string           | see above | API identifier |
| path         | string           | yes      | Concrete path (e.g., `/pets/42`). Path parameters MUST be substituted by the caller. |
| query_params | object           | no       | Key-value pairs appended as query string parameters |
| headers      | object           | no       | Key-value pairs added as request headers (e.g., `{"Authorization": "Bearer <token>"}`) |

### Output (success)

`HTTPResult` — see data-model.md

### Output (error)

`ToolError` with one of:
- `PATH_NOT_FOUND` — path does not exist in the API definition
- `VALIDATION_FAILED` — request parameters fail schema validation
- `HTTP_ERROR` — network or HTTP-level error during the call

---

## Tool: `http_post`

Execute a validated HTTP POST request against a loaded API.

### Input

| Parameter    | Type             | Required | Description |
|--------------|------------------|----------|-------------|
| api          | string           | see above | API identifier |
| path         | string           | yes      | Concrete path with parameters substituted |
| query_params | object           | no       | Key-value query string parameters |
| headers      | object           | no       | Request headers |
| body         | object \| string | no       | Request body. Serialized as JSON if object; sent as-is if string. Must conform to the endpoint's request body schema. |

### Output (success)

`HTTPResult`

### Output (error)

`ToolError` — same codes as `http_get`, plus:
- `VALIDATION_FAILED` with `MISSING_REQUIRED_FIELD` or `INVALID_FIELD_TYPE` details when body fails schema validation

---

## Tool: `http_put`

Execute a validated HTTP PUT request. Identical contract to `http_post`.

---

## Tool: `http_patch`

Execute a validated HTTP PATCH request. Identical contract to `http_post`.

---

## Tool: `explore_api`

List available API paths for progressive discovery.

### Input

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| api       | string | see above | API identifier |
| prefix    | string | no       | Filter: only return paths that start with this string (e.g., `/pets`). Case-sensitive. If omitted, all paths are returned. |

### Output (success)

Array of `PathInfo` objects sorted lexicographically by path:

```json
[
  {
    "path": "/pets",
    "methods": ["GET", "POST"],
    "description": "List or create pets"
  },
  {
    "path": "/pets/{id}",
    "methods": ["GET", "PUT", "DELETE"],
    "description": "Get, update, or delete a pet by ID"
  }
]
```

### Output (error)

`ToolError` — `AMBIGUOUS_API` if multiple APIs loaded and `api` omitted.

**Note**: This tool never returns `PATH_NOT_FOUND`. If the prefix matches nothing, it returns
an empty array.

---

## Tool: `get_schema`

Return the full schema for one endpoint, enabling the agent to construct a valid request.

### Input

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| api       | string | see above | API identifier |
| path      | string | yes      | OpenAPI path template (e.g., `/pets/{id}`) OR a concrete path; server resolves the template match. |
| method    | string | yes      | HTTP method in any case: `GET`, `POST`, `PUT`, `PATCH` |

### Output (success)

`SchemaResult` example:

```json
{
  "path": "/pets/{id}",
  "method": "GET",
  "path_parameters": {
    "id": {
      "type": "integer",
      "description": "Pet ID",
      "required": true
    }
  },
  "query_parameters": {
    "expand": {
      "type": "string",
      "enum": ["owner", "vaccinations"],
      "required": false
    }
  },
  "responses": {
    "200": {
      "type": "object",
      "properties": {
        "id":   { "type": "integer" },
        "name": { "type": "string" }
      }
    },
    "404": {
      "type": "object",
      "properties": {
        "error": { "type": "string" }
      }
    }
  }
}
```

### Output (error)

`ToolError` with:
- `PATH_NOT_FOUND` — path does not exist or method not defined for that path

---

## Error Format Reference

All errors follow this envelope:

```json
{
  "code": "VALIDATION_FAILED",
  "message": "Request body validation failed for POST /pets",
  "details": [
    {
      "type": "MISSING_REQUIRED_FIELD",
      "field": "name",
      "message": "Field 'name' is required but was not provided"
    },
    {
      "type": "INVALID_FIELD_TYPE",
      "field": "age",
      "message": "Field 'age' must be of type integer, got string"
    }
  ],
  "hints": []
}
```

```json
{
  "code": "PATH_NOT_FOUND",
  "message": "Path '/pet/42' does not exist in API 'petstore'",
  "details": [],
  "hints": ["/pets/{id}", "/pets"]
}
```

```json
{
  "code": "AMBIGUOUS_API",
  "message": "Multiple APIs are loaded; specify the 'api' parameter",
  "details": [],
  "hints": ["petstore", "bookstore"]
}
```
