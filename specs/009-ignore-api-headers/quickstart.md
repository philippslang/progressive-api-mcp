# Quickstart: Ignore Headers for APIs

## What this feature does

Allows you to declare a list of headers per API that the MCP server should treat as non-existent: they will not appear in `get_schema` output and will not be required during request validation. Useful for headers that are injected automatically by infrastructure (API gateways, reverse proxies) and should be invisible to MCP consumers.

## Configuration

Add `ignore_headers` to any API entry in your `config.yaml`:

```yaml
server:
  host: 0.0.0.0
  port: 8080

apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://petstore.example.com
    ignore_headers:
      - Authorization
      - X-Internal-Token
```

Header names are case-insensitive. `Authorization` and `authorization` are treated as the same header.

## What changes

| Behaviour          | Before                                       | After                                                  |
|--------------------|----------------------------------------------|--------------------------------------------------------|
| `get_schema`       | Header params not shown at all               | Non-ignored header params shown; ignored ones excluded |
| `http_get` etc.    | Required headers must be supplied            | Ignored headers not required; optional if supplied     |
| `explore_api`      | No parameters shown                          | No change                                              |

## Example: calling without a required header

**Without `ignore_headers`**, omitting a required `Authorization` header fails:
```json
{ "code": "VALIDATION_FAILED", "message": "request validation failed for GET /pets", "details": [...] }
```

**With `ignore_headers: [Authorization]`**, the same call succeeds:
```json
{ "status_code": 200, "headers": { ... }, "body": [ ... ] }
```

## Supplying an ignored header anyway

If you supply an ignored header explicitly in the `headers` argument, it is forwarded to the upstream API unchanged:

```json
{
  "path": "/pets",
  "headers": { "Authorization": "Bearer my-token" }
}
```

This allows callers to opt in to authentication when they have a token, without being forced to always provide one.
