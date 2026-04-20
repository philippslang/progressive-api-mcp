# Contract: `list_api` MCP Tool

**Feature**: 015-rename-explore-to-list | **Date**: 2026-04-18

This replaces the tool previously known as `explore_api`. Behaviour is identical; only the name changes.

## Tool definition

**Name**: `list_api` (prefixable via `server.tool_prefix` — e.g. `store_list_api`)

**Description**: `List available API paths for progressive discovery`

## Input parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `api` | string | no (required when multiple APIs loaded) | — | Identifier of the registered API to list. Required only when more than one API is registered and disambiguation is necessary. |
| `prefix` | string | no | `""` | Only return paths whose template starts with this substring. |

## Output — success

A JSON array of path entries, sorted ascending by `path`:

```json
[
  { "path": "/pets",       "methods": ["GET", "POST"], "description": "List all pets" },
  { "path": "/pets/{id}",  "methods": ["GET", "PUT", "PATCH", "DELETE"], "description": "Get a pet by ID" }
]
```

- `methods` preserves the order GET, POST, PUT, PATCH, DELETE.
- `description` is `""` when neither GET nor POST on the path has a `summary`.
- `path` uses the OpenAPI template form (e.g. `/pets/{id}`, not a concrete URL).

## Output — error envelope

Standard tool error shape:

```json
{ "code": "AMBIGUOUS_API", "message": "...", "details": null, "hints": ["petstore", "bookstore"] }
```

### Error codes

| Code | When |
|------|------|
| `AMBIGUOUS_API` | `api` omitted while multiple APIs are registered. `hints` lists the registered names. |
| `API_NOT_FOUND` | `api` refers to an unregistered API name. |
| `INTERNAL_ERROR` | OpenAPI document failed to build v3 model. |

## Allow-list behaviour

- Tool registers only if at least one API's `allow_list.tools` includes `list_api`, or if no API restricts tools (allow-all).
- For the API being listed, paths are filtered by `allow_list.paths["list_api"]` when set. Paths not matching any permitted template are omitted from the result.
- The `allow_list.tools` validator rejects `"explore_api"` — startup fails with an error that lists `list_api` among the valid names.

## Invariants

- No upstream HTTP calls.
- No request validation.
- Result list is sorted deterministically by `path`.
- Renaming does not change the tool's handler logic, input schema, output schema, error envelope, or allow-list behaviour.
