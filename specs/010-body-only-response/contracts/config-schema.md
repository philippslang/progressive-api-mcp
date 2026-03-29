# Contract: Config Schema — response_body_only field

## YAML Config Contract

```yaml
apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://api.example.com
    response_body_only: true   # optional boolean; default false
```

### Rules

- `response_body_only` is optional. When absent or `false`, behaviour is unchanged.
- Only `true` and `false` are valid values (standard YAML boolean).

---

## MCP Tool Response Contract Changes

### `http_get` / `http_post` / `http_put` / `http_patch` — response shape

**Default (`response_body_only: false` or absent)**:
```json
{
  "status_code": 200,
  "headers": { "Content-Type": "application/json" },
  "body": [ { "id": 1, "name": "Cat" } ]
}
```

**Body-only (`response_body_only: true`)**:
```json
[ { "id": 1, "name": "Cat" } ]
```

For a non-JSON upstream response:

**Default**:
```json
{ "status_code": 200, "headers": { ... }, "body": "OK" }
```

**Body-only**:
```json
"OK"
```

### Rules

- The body-only value is exactly the value that would have appeared in the `body` field of the full response.
- If the upstream returns an empty body, the result is `null` or `""` depending on content type.
- No change to `explore_api` or `get_schema` tool responses.
