# Contract: Config Schema — ignore_headers field

## YAML Config Contract

```yaml
apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://api.example.com
    ignore_headers:        # optional; list of header names to suppress
      - Authorization
      - X-Internal-Token
```

### Rules

- `ignore_headers` is optional. When absent, behaviour is unchanged.
- Each entry is a header name string (case-insensitive). `Authorization` and `authorization` are equivalent.
- Empty string entries (`""`) are invalid and cause startup to fail with a descriptive error.
- An empty list (`ignore_headers: []`) is equivalent to omitting the field.
- Duplicate names (case-insensitive) are silently collapsed.

---

## MCP Tool Contract Changes

### `get_schema` — response shape (additive change)

Previously, header parameters were silently omitted from the output. After this feature:

- Header parameters that are NOT in `ignore_headers` are included under `header_parameters`.
- Header parameters that ARE in `ignore_headers` are excluded from the output entirely.
- If all header parameters are ignored (or there are none), `header_parameters` is absent from the response.

**Before** (no header params in output):
```json
{
  "path": "/pets",
  "method": "GET",
  "query_parameters": { "limit": { "type": "integer", "required": false } },
  "responses": { "200": { "description": "A list of pets" } }
}
```

**After** (with a non-ignored header `X-Request-ID` and ignored header `Authorization`):
```json
{
  "path": "/pets",
  "method": "GET",
  "header_parameters": {
    "X-Request-ID": { "description": "Trace ID", "required": false, "type": "string" }
  },
  "query_parameters": { "limit": { "type": "integer", "required": false } },
  "responses": { "200": { "description": "A list of pets" } }
}
```

### `http_get` / `http_post` / `http_put` / `http_patch` — validation behaviour

- Ignored headers are no longer required by validation. Requests omitting them pass validation.
- If a caller explicitly provides an ignored header in the `headers` argument, it is forwarded to the upstream API as normal.
- No change to the response shape of HTTP tools.

### `explore_api` — no change

`explore_api` lists paths and HTTP methods only; it does not expose parameters. No behaviour change.
