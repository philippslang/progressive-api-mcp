# Data Model: search_api Tool

**Branch**: `014-search-api-tool` | **Date**: 2026-04-18

## New Entities

### SearchResult (Go struct, `pkg/tools/search.go`)

One match in the response array.

| Field | Go type | JSON tag | Description |
|-------|---------|----------|-------------|
| API | string | `api` | API name (as registered in `registry`) |
| Method | string | `method` | HTTP method in upper case: `GET`, `POST`, `PUT`, `PATCH`, `DELETE` |
| Path | string | `path` | OpenAPI path template (e.g., `/pets/{id}`) |
| Schema | map[string]any | `schema,omitempty` | Request-body schema (from `schemaToMap`); omitted when operation has no request body |

### Tool input schema (MCP)

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `query` | string | yes | Substring to search for (case-insensitive). Must be non-empty after trim. |
| `api` | string | no | If present, restrict search to the API with this name. If absent, search all registered APIs. |

### Tool output shape

- Success: JSON array of `SearchResult` objects — possibly empty.
- Error: standard tool error envelope via `toolErrorResult(code, message, details, hints)`.

---

## Changed Entities

### `knownToolNames` (in `pkg/config/config.go`)

Add `"search_api": {}`. Also update `knownToolNamesHint` to include `search_api` in the comma-separated list.

No struct field changes. No YAML schema changes beyond acceptance of `"search_api"` in existing `allow_list.tools` and `allow_list.paths` maps.

---

## Call Flow

```
MCP tool call: search_api(query, api?)
  → RegisterSearchTools handler
    → validate query non-empty (trim) → [INVALID_INPUT] on fail
    → if api present:
         lookup registry → [API_NOT_FOUND] on fail
         entries = [one entry]
       else:
         entries = registry.All()
    → for entry in entries:
         build v3 model
         for (path, item) in model.Paths.PathItems:
           if allow_list.paths["search_api"] set and path not permitted: skip
           for method in [GET, POST, PUT, PATCH, DELETE]:
             op = item.<Method>
             if op == nil: continue
             haystack = lower(path) + " " + lower(op.Summary) + " " + lower(op.Description)
             if contains(haystack, lower(query)):
               results = append(results, {api, method, path, schemaToMap(requestBody)})
    → return JSON(results)
```

## Invariants

- Result ordering: entries iteration order from the registry (insertion order), then path iteration order from `PathItems.FromOldest()`, then method order GET → POST → PUT → PATCH → DELETE. Deterministic per server instance.
- No network calls; no validation; no side effects.
- Schema resolution depth is capped by existing `maxSchemaDepth = 10` in `schemaToMap`.

---

## Backward Compatibility

- Fully additive. No existing config or API response shapes change.
- Configs that reference `"search_api"` in `allow_list.tools`/`allow_list.paths` now validate instead of being rejected — this is a relaxation, not a break.
