# Quickstart: Body-Only HTTP Responses

## What this feature does

Adds a per-API `response_body_only` flag. When `true`, HTTP tool calls return only the response body instead of the full `{ status_code, headers, body }` envelope. Useful when callers only need the data and the extra metadata adds noise.

## Configuration

```yaml
apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://petstore.example.com
    response_body_only: true
```

## Before / After

**Before** (default):
```json
{
  "status_code": 200,
  "headers": { "Content-Type": "application/json" },
  "body": [ { "id": 1, "name": "Cat" } ]
}
```

**After** (`response_body_only: true`):
```json
[ { "id": 1, "name": "Cat" } ]
```

## Notes

- Applies to all four HTTP tools: `http_get`, `http_post`, `http_put`, `http_patch`.
- Does not affect `explore_api` or `get_schema`.
- When the upstream returns a non-JSON body, the raw text string is returned.
- Error detection (non-2xx) is the caller's responsibility in body-only mode; the status code is not included in the response.
