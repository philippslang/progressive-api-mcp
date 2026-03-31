# Contract: API Configuration Schema

**Feature**: 013-api-skip-validation | **Date**: 2026-03-31

## YAML Configuration Contract

The `apis` array in the server config gains one new optional field per entry.

### Updated per-API config block

```yaml
apis:
  - name: my-api                      # required, unique
    definition: ./my-api.yaml         # required, path to OpenAPI spec
    host: https://api.example.com     # optional, override base URL
    base_path: /v1                    # optional, override base path
    skip_validation: true             # optional, default: false
                                      # When true: payload validation is bypassed
                                      # and the request is forwarded as-is to the
                                      # upstream API.
    allow_list:                       # optional
      tools: [http_get, http_post]
      paths:
        http_get: ["/pets", "/pets/{petId}"]
    ignore_headers:                   # optional
      - Authorization
    response_body_only: false         # optional, default: false
```

### Field specification

| Field | Type | Required | Default | Constraint |
|-------|------|----------|---------|------------|
| `skip_validation` | boolean | No | `false` | Must be `true` or `false`; any other YAML value is a parse error |

### Behavior contract

| `skip_validation` value | Validation behaviour |
|------------------------|----------------------|
| `false` (or absent) | Request payload validated against OpenAPI schema before forwarding. Invalid payloads return an MCP tool error; the upstream API is not called. |
| `true` | Payload validation skipped. Request forwarded directly to upstream API. All upstream responses (including upstream validation errors) returned to the caller. |

### Invariants

- Setting `skip_validation: true` does NOT change which tools are available or what schema is exposed to the LLM. Tool input schemas remain derived from the OpenAPI definition.
- Setting `skip_validation: true` does NOT affect other APIs configured in the same server instance.
- Path allow-list and tool allow-list checks still run regardless of `skip_validation`.
- Header suppression (`ignore_headers`) still applies regardless of `skip_validation`.
