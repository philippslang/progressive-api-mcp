# Contract: search_api MCP Tool

**Feature**: 014-search-api-tool | **Date**: 2026-04-18

## Tool definition

**Name**: `search_api` (prefixable via `server.tool_prefix` — e.g., `myapi_search_api`)

**Description**: `Search endpoints across APIs by substring match on path or operation description`

## Input parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | yes | — | Substring to search for. Matched case-insensitively against each endpoint's path template, summary, and description. Whitespace-only values are rejected. |
| `api` | string | no | (all APIs) | If set, search only this registered API's endpoints. |

## Output — success

A JSON array (possibly empty) of objects:

```json
[
  {
    "api": "petstore",
    "method": "GET",
    "path": "/pets/{id}",
    "schema": { }
  },
  {
    "api": "petstore",
    "method": "POST",
    "path": "/pets",
    "schema": {
      "required": true,
      "schema": {
        "type": "object",
        "properties": {
          "name": { "type": "string", "required": true },
          "species": { "type": "string", "required": false }
        }
      }
    }
  }
]
```

- `schema` is omitted (key absent) for endpoints with no request body.
- Field presence: `api`, `method`, `path` are always present; `schema` is optional.
- When no endpoint matches, the array is empty (`[]`) with no error.

## Output — error envelope

Standard tool error shape (same as `explore_api`, `get_schema`):

```json
{
  "code": "INVALID_INPUT",
  "message": "query must not be empty",
  "details": null,
  "hints": null
}
```

### Error codes

| Code | When |
|------|------|
| `INVALID_INPUT` | `query` is missing, empty, or whitespace-only |
| `API_NOT_FOUND` | `api` parameter refers to an unregistered API name |

## Match semantics

- Case-insensitive substring match (`strings.Contains` on lower-cased strings).
- Matches if the query substring appears in any of: path template, operation `summary`, operation `description`.
- Examples (query in quotes):
  - `"pet"` matches path `/pets/{id}` ✓
  - `"PET"` matches path `/pets/{id}` ✓ (case-insensitive)
  - `"pet"` matches path `/carpets` ✓ (substring, no word boundary)
  - `"order status"` matches description `"Return the current order status"` ✓
  - `"x/y/z"` — no match if that exact substring is absent from path/summary/description

## Allow-list behaviour

- Tool registers only if at least one API includes `search_api` in `allow_list.tools` (or no API restricts tools — allow-all).
- For each API scanned, paths are filtered by that API's `allow_list.paths["search_api"]` entry (if set). Endpoints whose template is not permitted are excluded.

## Invariants

- No upstream HTTP calls.
- No request validation.
- Result list order is deterministic per server instance (insertion order of APIs × path order × method order).
- Schema field uses the same resolver as `get_schema`'s `request_body.schema` subfield; same depth cap applies.
